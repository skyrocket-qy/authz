package rbac

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"authz/internal/config"
	"authz/internal/entity"
	"authz/internal/schema"
	"authz/internal/util"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/skyrocket-qy/erx"
	"gorm.io/gorm"
)

// zanzibar in memory.
type ZanzibarMemory interface {
	Check(c context.Context, sbj entity.Instance, rel string, obj entity.Instance) (bool, error)
	// Lookup(c context.Context, sbj *entity.Instance, rel string) ([]*entity.Instance, error)
	// Expand(c context.Context, rel string, obj *entity.Instance) ([]*entity.Instance, error)
}

var _ ZanzibarMemory = (*ZanzibarMemoryImpl)(nil)

type ZanzibarMemoryImpl struct {
	schema *schema.Schema
	kafkaR *kafka.Reader
	db     *gorm.DB

	graph  *Graph
	Offest int64
	cancel context.CancelFunc
	mutex  sync.RWMutex
}

func NewZanzibarMemory(c context.Context, lc *util.LifecycleParallel, db *gorm.DB, s *schema.Schema,
	kafkaR *kafka.Reader) (*ZanzibarMemoryImpl, error,
) {
	engine := ZanzibarMemoryImpl{
		kafkaR: kafkaR,
		db:     db,
		graph:  NewGraph(),
	}

	st := time.Now()

	if err := engine.build(c); err != nil {
		return nil, err
	}

	log.Info().
		Int64("offset", engine.Offest).
		Int64("took ms", time.Since(st).Milliseconds()).Msg("build rbac graph")

	cc, cancel := context.WithCancel(c)
	engine.cancel = cancel

	if err := engine.kafkaR.SetOffset(engine.Offest + 1); err != nil {
		return nil, erx.W(err, "failed to set offset to earliest")
	}

	go engine.sync(cc)

	engine.schema = s

	lc.Add(&engine, engine.Close, db, kafkaR)

	return &engine, nil
}

func (e *ZanzibarMemoryImpl) Close(c context.Context) error {
	e.cancel()

	return nil
}

func (e *ZanzibarMemoryImpl) Check(c context.Context, user entity.Instance, perm string,
	obj entity.Instance) (bool, error,
) {
	return e.check(user, perm, obj, map[entity.Instance]struct{}{})
}

func (e *ZanzibarMemoryImpl) check(user entity.Instance, perm string,
	obj entity.Instance, visited map[entity.Instance]struct{}) (bool, error,
) {
	ns, ok := e.schema.Namespaces[obj.Ns]
	if !ok {
		return false, erx.Newf(util.ErrBadRequest, "unknown namespace: %s", obj.Ns)
	}

	relation, ok := ns.Relations[perm]
	if !ok {
		return false, erx.Newf(util.ErrBadRequest, "unknown relation: %s", perm)
	}

	if relation == nil {
		return false, nil
	}

	return e.evalUsersetRewrite(&relation.UsersetRewrite, user, obj, visited), nil
}

func (e *ZanzibarMemoryImpl) evalUsersetRewrite(rewrite *schema.UsersetRewrite,
	user, obj entity.Instance, visited map[entity.Instance]struct{},
) bool {
	if _, ok := visited[obj]; ok || (len(visited) >= config.Conf.MaxCheckNodes) {
		// the return boolean will not tell caller to stop checking, but it will stop recursion
		return false
	}

	visited[obj] = struct{}{}

	switch {
	case rewrite.ComputedUserSet != nil:
		return e.graph.Exist(obj, rewrite.ComputedUserSet.Relation, user)

	case rewrite.TupleToUserset != nil:
		return e.evalTupleToUserset(rewrite.TupleToUserset, user, obj, visited)

	case rewrite.Union != nil:
		for _, r := range rewrite.Union {
			if e.evalUsersetRewrite(r, user, obj, visited) {
				return true
			}
		}

		return false

	case rewrite.Intersection != nil:
		for _, r := range rewrite.Intersection {
			if !e.evalUsersetRewrite(r, user, obj, visited) {
				return false
			}
		}

		return true

	case rewrite.Exclusion != nil:
		return e.evalUsersetRewrite(rewrite.Exclusion.Base, user, obj, visited) &&
			!e.evalUsersetRewrite(rewrite.Exclusion.Subtract, user, obj, visited)
	}

	return false
}

func (e *ZanzibarMemoryImpl) evalTupleToUserset(tupleToUserset *schema.TupleToUserset,
	sbj, obj entity.Instance, visited map[entity.Instance]struct{},
) bool {
	reWriteObjs := e.getSbjs(
		[]entity.Instance{obj, {Ns: obj.Ns, Id: "*"}},
		tupleToUserset.ComputedUserset.Relation,
	)

	for nextObj := range reWriteObjs {
		ok, _ := e.check(sbj, tupleToUserset.ComputedUserset.Relation, nextObj, visited)
		if ok {
			return true
		}
	}

	return false
}

func (e *ZanzibarMemoryImpl) getSbjs(
	objs []entity.Instance,
	rel string,
) map[entity.Instance]struct{} {
	sbjs := make(map[entity.Instance]struct{})

	for _, obj := range objs {
		s := e.graph.getShard(obj)
		s.mu.RLock()
		objEntry, ok := s.Graph[obj]
		s.mu.RUnlock()

		if !ok {
			continue
		}

		objEntry.mu.RLock()

		sbjEntry, ok := objEntry.Relations[rel]
		if ok {
			for sbj := range sbjEntry {
				sbjs[sbj] = struct{}{}
			}
		}

		objEntry.mu.RUnlock()
	}

	return sbjs
}

func (e *ZanzibarMemoryImpl) build(c context.Context) error {
	log.Info().Msg("build rbac graph")

	r := e.kafkaR

	cp := GraphCheckpoint{}

	tx := e.db.WithContext(c).Take(&cp)
	if err := tx.Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return erx.W(err, "failed to get graph checkpoint")
		}
	}

	if tx.RowsAffected > 0 {
		if err := util.DecodeGob(cp.Data, e.graph); err != nil {
			return erx.W(err, "failed to decode graph")
		}

		e.Offest = cp.LastOffset

		return nil
	}

	// Build from all messages
	if err := r.SetOffset(kafka.FirstOffset); err != nil {
		return erx.W(err, "failed to set offset to earliest")
	}

	for {
		select {
		case <-c.Done():
			return nil
		default:
		}

		readCtx, cancel := context.WithTimeout(c, 3*time.Second)
		defer cancel()

		m, err := r.ReadMessage(readCtx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return nil
			}

			return erx.W(err, "read message error")
		}

		if err := e.applyMessage(m); err != nil {
			return err
		}
	}
}

func (e *ZanzibarMemoryImpl) applyMessage(m kafka.Message) error {
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

	e.Offest = m.Offset // TODO: it only use on job service

	switch val.Op {
	case "c":
		e.graph.Create(obj, rel, sbj)
	case "d":
		e.graph.Delete(obj, rel, sbj)
	default:
	}

	return nil
}

func (e *ZanzibarMemoryImpl) sync(c context.Context) {
	for {
		select {
		case <-c.Done():
			return
		default:
			m, err := e.kafkaR.ReadMessage(c)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}

				log.Error().Err(err).Msg("Failed to read message")

				continue
			}

			log.Debug().Msg("read kafk amessage")

			if err := e.applyMessage(m); err != nil {
				log.Error().Err(err).Msg("Failed to apply message")

				continue
			}
		}
	}
}

func (e *ZanzibarMemoryImpl) SyncGraphCheckpoint(c context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-c.Done():
			return
		case <-ticker.C:
			cp := GraphCheckpoint{}

			tx := e.db.WithContext(c).Take(&cp)
			if err := tx.Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error().Err(err).Msg("Failed to get graph checkpoint")

					continue
				}
			}

			if tx.RowsAffected == 0 {
				e.mutex.RLock()

				bytes, err := util.EncodeGob(&e.graph)
				if err != nil {
					log.Warn().Err(err).Msg("Failed to encode graph")
					e.mutex.RUnlock()

					continue
				}

				cp.LastOffset = e.Offest
				e.mutex.RUnlock()

				cp.Data = bytes

				if err := e.db.WithContext(c).Create(&cp).Error; err != nil {
					log.Error().Err(err).Msg("Failed to create graph checkpoint")

					continue
				}
			} else if e.Offest-cp.LastOffset > 1000 {
				e.mutex.RLock()

				bytes, err := util.EncodeGob(&e.graph)
				if err != nil {
					log.Warn().Err(err).Msg("Failed to encode graph")
					e.mutex.RUnlock()

					continue
				}

				cp.LastOffset = e.Offest
				e.mutex.RUnlock()

				cp.Data = bytes

				if err := e.db.WithContext(c).Save(&cp).Error; err != nil {
					log.Error().Err(err).Msg("Failed to update graph checkpoint")

					continue
				}
			}
		}
	}
}
