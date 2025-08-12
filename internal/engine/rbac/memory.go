package rbac

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"authz/internal/entity"
	"authz/internal/schema"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

// TODO: prebuild the graph to reduce loading time

// zanzibar in memory
type ZanzibarEngine interface {
	Check(c context.Context, sbj *entity.Instance, rel string, obj *entity.Instance) (bool, error)
	Lookup(c context.Context, sbj *entity.Instance, rel string) ([]*entity.Instance, error)
	Expand(c context.Context, rel string, obj *entity.Instance) ([]*entity.Instance, error)
}

var _ ZanzibarEngine = (*ZanzibarEngineImpl)(nil)

type ZanzibarEngineImpl struct {
	schema *schema.Schema
	kafkaR *kafka.Reader
	db     *gorm.DB
	rds    *redis.Client
	graph  map[entity.Instance]map[string]map[entity.Instance]struct{}

	mutex sync.RWMutex
}

func NewZanzibarEngine(c context.Context, db *gorm.DB, rds *redis.Client, s *schema.Schema,
	kafkaR *kafka.Reader) (*ZanzibarEngineImpl, error,
) {

	engine := ZanzibarEngineImpl{
		kafkaR: kafkaR,
		db:     db,
		rds:    rds,
		graph:  make(map[entity.Instance]map[string]map[entity.Instance]struct{}),
	}

	if err := engine.build(c); err != nil {
		return nil, err
	}

	go engine.sync()
	engine.schema = s

	return &engine, nil
}

func (e *ZanzibarEngineImpl) Check(c context.Context, user *entity.Instance, perm string,
	obj *entity.Instance) (bool, error,
) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	ns, ok := e.schema.Namespaces[obj.Ns]
	if !ok {
		return false, fmt.Errorf("unknown namespace: %s", obj.Ns)
	}
	relation, ok := ns.Relations[perm]
	if !ok {
		return false, fmt.Errorf("unknown relation: %s", perm)
	}

	if relation == nil {
		return false, nil
	}

	return e.evalUsersetRewrite(c, &relation.UsersetRewrite, user, obj), nil
}

func (e *ZanzibarEngineImpl) evalUsersetRewrite(c context.Context, rewrite *schema.UsersetRewrite,
	user *entity.Instance, obj *entity.Instance) bool {
	switch {
	case rewrite.ComputedUserSet != nil:
		return e.hasDirectTuple(user, rewrite.ComputedUserSet.Relation, obj)

	case rewrite.TupleToUserset != nil:
		relSbj, ok := e.graph[*obj]
		if !ok {
			return false
		}
		sbjs, ok := relSbj[rewrite.TupleToUserset.Tupleset.Relation]
		if !ok {
			return false
		}

		for sbj := range sbjs {
			ok, _ := e.Check(c, user, rewrite.TupleToUserset.ComputedUserset.Relation, &sbj)
			if ok {
				return true
			}
		}

		return false

	case rewrite.Union != nil:
		for _, r := range rewrite.Union {
			if e.evalUsersetRewrite(c, r, user, obj) {
				return true
			}
		}
		return false

	case rewrite.Intersection != nil:
		for _, r := range rewrite.Intersection {
			if !e.evalUsersetRewrite(c, r, user, obj) {
				return false
			}
		}
		return true

	case rewrite.Exclusion != nil:
		return e.evalUsersetRewrite(c, rewrite.Exclusion.Base, user, obj) &&
			!e.evalUsersetRewrite(c, rewrite.Exclusion.Subtract, user, obj)
	}

	return false
}

func (e *ZanzibarEngineImpl) hasDirectTuple(user *entity.Instance, rel string, obj *entity.Instance) bool {
	if _, ok := e.graph[*obj][rel][*user]; ok {
		return true
	}
	return false
}

func (e *ZanzibarEngineImpl) Lookup(c context.Context, sbj *entity.Instance, rel string) (
	[]*entity.Instance, error) {
	return nil, nil
}

func (e *ZanzibarEngineImpl) Expand(c context.Context, rel string, obj *entity.Instance) (
	sbjs []*entity.Instance, err error,
) {

	return nil, nil
}

func (e *ZanzibarEngineImpl) build(c context.Context) error {
	r := e.kafkaR
	e.graph = make(map[entity.Instance]map[string]map[entity.Instance]struct{})

	// Reset offset to earliest to read all messages from start
	if err := r.SetOffset(kafka.FirstOffset); err != nil {
		return fmt.Errorf("failed to set offset to earliest: %w", err)
	}

	for {
		readCtx, cancel := context.WithTimeout(c, 3*time.Second)
		defer cancel()
		m, err := r.ReadMessage(readCtx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return nil
			}

			return fmt.Errorf("read message error: %w", err)
		}

		if err := e.applyMessage(m); err != nil {
			return err
		}
	}
}

func (e *ZanzibarEngineImpl) applyMessage(m kafka.Message) error {
	type Val struct {
		Tuple
		Op string `json:"__op"`
	}
	var val Val

	if err := json.Unmarshal(m.Value, &val); err != nil {
		return err
	}

	sbj := entity.Instance{Ns: val.SbjNs, Id: val.SbjId}
	rel := val.Relation
	obj := entity.Instance{Ns: val.ObjNs, Id: val.ObjId}

	e.mutex.Lock()
	defer e.mutex.Unlock()
	if _, exists := e.graph[obj]; !exists {
		e.graph[obj] = make(map[string]map[entity.Instance]struct{})
	}
	if _, exists := e.graph[obj][rel]; !exists {
		e.graph[obj][rel] = make(map[entity.Instance]struct{})
	}
	e.graph[obj][rel][sbj] = struct{}{}

	// Apply operation type from val.Op ("c"=create, "d"=delete, etc.)
	switch val.Op {
	case "c": // create/add edge
		if _, exists := e.graph[obj]; !exists {
			e.graph[obj] = make(map[string]map[entity.Instance]struct{})
		}
		if _, exists := e.graph[obj][rel]; !exists {
			e.graph[obj][rel] = make(map[entity.Instance]struct{})
		}
		e.graph[obj][rel][sbj] = struct{}{}
	case "d": // delete/remove edge
		delete(e.graph[obj][rel], sbj)
	default:
		// handle other ops if any
	}

	return nil
}

func (e *ZanzibarEngineImpl) sync() {
	for {
		m, err := e.kafkaR.ReadMessage(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("Failed to read message")
			continue
		}

		if err := e.applyMessage(m); err != nil {
			log.Error().Err(err).Msg("Failed to apply message")
			continue
		}
	}
}

// if kafka message.version not fit, need to catchup from db first
// func (e *ZanzibarEngineImpl) CatchUp(c context.Context) {
// 	for {
// 		select {
// 		case <-c.Done():
// 			log.Println("Closing Kafka reader...")
// 			if err := e.kafkaR.Close(); err != nil {
// 				log.Printf("Error closing reader: %v\n", err)
// 			}
// 			log.Println("Kafka reader closed. Exiting.")
// 		default:
// 			m, err := e.kafkaR.FetchMessage(c)
// 			if err != nil {
// 				if c.Err() != nil {
// 					continue
// 				}
// 				log.Printf("Error fetching message: %v\n", err)
// 				time.Sleep(1 * time.Second)
// 				continue
// 			}

// 			fmt.Printf("Received message at offset %d: key=%s, value=%s\n",
// 				m.Offset, string(m.Key), string(m.Value))

// 			update := &authzpbv1.GraphUpdate{}
// 			if err := proto.Unmarshal(m.Value, update); err != nil {
// 				log.Printf("Failed to unmarshal message at offset %d: %v\n", m.Offset, err)
// 				continue
// 			}

// 			if update.Version > e.version {
// 				if update.Version != e.version+1 {
// 					if err := e.CatchUpFromDb(c); err != nil {
// 						log.Printf("Failed to build graph at offset %d: %v\n", m.Offset, err)

// 						if err := e.rebuild(c); err != nil {
// 							log.Printf("Failed to build graph at offset %d: %v\n", m.Offset, err)
// 							continue
// 						}
// 					}
// 				}

// 				if err := e.updateGraph(update); err != nil {
// 					log.Printf("Failed to update graph at offset %d: %v\n", m.Offset, err)
// 					continue
// 				}
// 			}

// 			if err := e.kafkaR.CommitMessages(c, m); err != nil {
// 				log.Printf("Failed to commit offset for message at offset %d: %v\n", m.Offset, err)
// 				continue
// 			}
// 		}
// 	}
// }

// func (e *ZanzibarEngineImpl) CatchUpFromDb(c context.Context) error {
// 	changeLogs := []*ChangeLog{}
// 	if err := e.db.Order("id").Where("version > ?", e.version).Find(&changeLogs).Error; err != nil {
// 		return err
// 	}

// 	e.mutex.Lock()
// 	defer e.mutex.Unlock()

// 	d := &authzpbv1.GraphUpdate{}
// 	for _, changeLog := range changeLogs {
// 		if err := proto.Unmarshal(changeLog.Data, d); err != nil {
// 			return err
// 		}

// 		switch d.Operation {
// 		case authzpbv1.Operation_CREATE:
// 			for _, tuple := range d.Tuples {
// 				sbj := entity.Instance{
// 					Ns: tuple.SbjNs,
// 					Id: tuple.SbjId,
// 				}
// 				obj := entity.Instance{
// 					Ns: tuple.ObjNs,
// 					Id: tuple.ObjId,
// 				}
// 				rel := tuple.Rel
// 				if _, exists := e.graph[obj]; !exists {
// 					e.graph[obj] = make(map[string]map[entity.Instance]struct{})
// 				}
// 				if _, exists := e.graph[obj][rel]; !exists {
// 					e.graph[obj][rel] = make(map[entity.Instance]struct{})
// 				}
// 				e.graph[obj][rel][sbj] = struct{}{}
// 			}

// 		case authzpbv1.Operation_DELETE:
// 			for _, tuple := range d.Tuples {
// 				obj := entity.Instance{
// 					Ns: tuple.ObjNs,
// 					Id: tuple.ObjId,
// 				}
// 				rel := tuple.Rel
// 				if _, exists := e.graph[obj]; !exists {
// 					continue
// 				}
// 				if _, exists := e.graph[obj][rel]; !exists {
// 					continue
// 				}
// 				delete(e.graph[obj][rel], entity.Instance{
// 					Ns: tuple.SbjNs,
// 					Id: tuple.SbjId,
// 				})
// 			}
// 		}
// 	}

// 	e.version = changeLogs[len(changeLogs)-1].Id

// 	return nil
// }

// func (e *ZanzibarEngineImpl) buildFromChangeLog(c context.Context) error {
// 	changeLogs := []*ChangeLog{}
// 	if err := e.db.Order("id").Find(&changeLogs).Error; err != nil {
// 		return err
// 	}

// 	e.mutex.Lock()
// 	defer e.mutex.Unlock()

// 	d := &authzpbv1.GraphUpdate{}
// 	for _, changeLog := range changeLogs {
// 		if err := proto.Unmarshal(changeLog.Data, d); err != nil {
// 			return err
// 		}

// 		switch d.Operation {
// 		case authzpbv1.Operation_CREATE:
// 			for _, tuple := range d.Tuples {
// 				sbj := entity.Instance{
// 					Ns: tuple.SbjNs,
// 					Id: tuple.SbjId,
// 				}
// 				obj := entity.Instance{
// 					Ns: tuple.ObjNs,
// 					Id: tuple.ObjId,
// 				}
// 				rel := tuple.Rel
// 				if _, exists := e.graph[obj]; !exists {
// 					e.graph[obj] = make(map[string]map[entity.Instance]struct{})
// 				}
// 				if _, exists := e.graph[obj][rel]; !exists {
// 					e.graph[obj][rel] = make(map[entity.Instance]struct{})
// 				}
// 				e.graph[obj][rel][sbj] = struct{}{}
// 			}

// 		case authzpbv1.Operation_DELETE:
// 			for _, tuple := range d.Tuples {
// 				obj := entity.Instance{
// 					Ns: tuple.ObjNs,
// 					Id: tuple.ObjId,
// 				}
// 				rel := tuple.Rel
// 				if _, exists := e.graph[obj]; !exists {
// 					continue
// 				}
// 				if _, exists := e.graph[obj][rel]; !exists {
// 					continue
// 				}
// 				delete(e.graph[obj][rel], entity.Instance{
// 					Ns: tuple.SbjNs,
// 					Id: tuple.SbjId,
// 				})
// 			}
// 		}
// 	}

// 	e.version = changeLogs[len(changeLogs)-1].Id

// 	return nil
// }

// func (e *ZanzibarEngineImpl) rebuild(c context.Context) error {
// 	latestGraph := GraphCheckpoint{}
// 	if err := e.db.Limit(1).Order("last_change_log_id desc").Take(&latestGraph).Error; err != nil {
// 		return e.buildFromChangeLog(c)
// 	}

// 	e.mutex.Lock()
// 	defer e.mutex.Unlock()

// 	dec := gob.NewDecoder(bytes.NewReader(latestGraph.Data))
// 	if err := dec.Decode(&e.graph); err != nil {
// 		return err
// 	}
// 	e.version = latestGraph.LastChangeLogId

// 	return nil
// }

// func (e *ZanzibarEngineImpl) updateGraph(update *authzpbv1.GraphUpdate) error {
// 	e.mutex.Lock()
// 	defer e.mutex.Unlock()

// 	for _, tuple := range update.Tuples {
// 		obj := entity.Instance{
// 			Ns: tuple.ObjNs,
// 			Id: tuple.ObjId,
// 		}
// 		sbj := entity.Instance{
// 			Ns: tuple.SbjNs,
// 			Id: tuple.SbjId,
// 		}
// 		rel := tuple.Rel

// 		delete(e.graph[obj][rel], sbj)
// 	}

// 	return nil
// }
