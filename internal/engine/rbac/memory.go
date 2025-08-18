package rbac

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"authz/internal/cfg"
	"authz/internal/entity"
	"authz/internal/pkg"
	"authz/internal/schema"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

// zanzibar in memory
type ZanzibarMemory interface {
	Check(c context.Context, sbj *entity.Instance, rel string, obj *entity.Instance) (bool, error)
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

func NewZanzibarMemory(c context.Context, lc *pkg.LifecycleParallel, db *gorm.DB, s *schema.Schema,
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
		return nil, fmt.Errorf("failed to set offset to earliest: %w", err)
	}
	go engine.sync(cc)
	engine.schema = s

	lc.Add(&engine, engine.Close, db, kafkaR)

	return &engine, nil
}

func (e *ZanzibarMemoryImpl) Close() error {
	e.cancel()
	return nil
}

func (e *ZanzibarMemoryImpl) Check(c context.Context, user *entity.Instance, perm string,
	obj *entity.Instance) (bool, error,
) {
	return e.check(c, user, perm, obj, map[entity.Instance]struct{}{})
}

func (e *ZanzibarMemoryImpl) check(c context.Context, user *entity.Instance, perm string,
	obj *entity.Instance, visited map[entity.Instance]struct{}) (bool, error,
) {
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

	return e.evalUsersetRewrite(c, &relation.UsersetRewrite, user, obj, visited), nil
}

func (e *ZanzibarMemoryImpl) evalUsersetRewrite(c context.Context, rewrite *schema.UsersetRewrite,
	user *entity.Instance, obj *entity.Instance, visited map[entity.Instance]struct{}) bool {
	if _, ok := visited[*obj]; ok || !(len(visited) < cfg.Cfg.MaxCheckNodes) {
		// the return boolean will not tell caller to stop checking, but it will stop recursion
		return false
	}
	visited[*obj] = struct{}{}

	switch {
	case rewrite.ComputedUserSet != nil:
		return e.graph.Exist(obj, rewrite.ComputedUserSet.Relation, user)

	case rewrite.TupleToUserset != nil:
		s := e.graph.getShard(obj)
		s.mu.RLock()
		objEntry := s.graph[*obj]
		s.mu.RUnlock()

		objEntry.mu.RLock()
		sbjs, ok := objEntry.Relations[rewrite.TupleToUserset.Tupleset.Relation]
		if !ok {
			return false
		}

		sbjArr := make([]*entity.Instance, 0, len(sbjs))
		for sbj := range sbjs {
			sbjArr = append(sbjArr, &sbj)
		}
		objEntry.mu.RUnlock()

		for _, sbj := range sbjArr {
			ok, _ := e.check(c, user, rewrite.TupleToUserset.ComputedUserset.Relation, sbj, visited)
			if ok {
				return true
			}
		}

		return false

	case rewrite.Union != nil:
		for _, r := range rewrite.Union {
			if e.evalUsersetRewrite(c, r, user, obj, visited) {
				return true
			}
		}
		return false

	case rewrite.Intersection != nil:
		for _, r := range rewrite.Intersection {
			if !e.evalUsersetRewrite(c, r, user, obj, visited) {
				return false
			}
		}
		return true

	case rewrite.Exclusion != nil:
		return e.evalUsersetRewrite(c, rewrite.Exclusion.Base, user, obj, visited) &&
			!e.evalUsersetRewrite(c, rewrite.Exclusion.Subtract, user, obj, visited)
	}

	return false
}

func (e *ZanzibarMemoryImpl) build(c context.Context) error {
	log.Info().Msg("build rbac graph")
	r := e.kafkaR

	cp := GraphCheckpoint{}
	tx := e.db.WithContext(c).Take(&cp)
	if err := tx.Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get graph checkpoint: %w", err)
		}
	}

	if tx.RowsAffected > 0 {
		if err := pkg.DecodeGob(cp.Data, &e.graph); err != nil {
			return fmt.Errorf("failed to decode graph: %w", err)
		}
		e.Offest = cp.LastOffset
		return nil
	}

	// Build from all messages
	if err := r.SetOffset(kafka.FirstOffset); err != nil {
		return fmt.Errorf("failed to set offset to earliest: %w", err)
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
			return fmt.Errorf("read message error: %w", err)
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
		e.graph.Create(&obj, rel, &sbj)
	case "d":
		e.graph.Delete(&obj, rel, &sbj)
	default:
	}

	return nil
}

func (e *ZanzibarMemoryImpl) sync(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			m, err := e.kafkaR.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				log.Error().Err(err).Msg("Failed to read message")
				continue
			}
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
				bytes, err := pkg.EncodeGob(&e.graph)
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
			} else {
				if e.Offest-cp.LastOffset > 1000 {
					e.mutex.RLock()
					bytes, err := pkg.EncodeGob(&e.graph)
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
}
