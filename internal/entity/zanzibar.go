package entity

import (
	"authz/internal/entity/model"
	"context"

	"gorm.io/gorm"
)

type Tuple struct {
	Sbj    *Instance `validate:"required"`
	SbjRel *string
	Rel    string    `validate:"required"`
	Obj    *Instance `validate:"required"`
}

func (t *Tuple) ToFilterModel(ctx context.Context, db *gorm.DB) (*model.Tuple, error) {
	var (
		sbjNs, objNs model.Namespace
		sbj, obj     model.Instance
		rel, sbjRel  model.RelationDef
	)

	tupleModel := &model.Tuple{}

	// Subject
	if t.Sbj != nil {
		if t.Sbj.Ns != "" {
			if err := db.WithContext(ctx).Where("name = ?", t.Sbj.Ns).Take(&sbjNs).Error; err != nil {
				return nil, err
			}
			tupleModel.SbjNsId = sbjNs.Id
		}
		if t.Sbj.Id != "" {
			if err := db.WithContext(ctx).Where("name = ?", t.Sbj.Id).Take(&sbj).Error; err != nil {
				return nil, err
			}
			tupleModel.SbjId = sbj.Id
		}
	}

	if t.SbjRel != nil && *t.SbjRel != "" {
		if err := db.WithContext(ctx).Where("name = ?", *t.SbjRel).Take(&sbjRel).Error; err != nil {
			return nil, err
		}
		tupleModel.SbjRelationId = &sbjRel.Id
	}

	// Relation
	if t.Rel != "" {
		if err := db.WithContext(ctx).Where("name = ?", t.Rel).Take(&rel).Error; err != nil {
			return nil, err
		}
		tupleModel.RelationId = rel.Id
	}

	// Object
	if t.Obj != nil {
		if t.Obj.Ns != "" {
			if err := db.WithContext(ctx).Where("name = ?", t.Obj.Ns).Take(&objNs).Error; err != nil {
				return nil, err
			}
			tupleModel.ObjNsId = objNs.Id
		}
		if t.Obj.Id != "" {
			if err := db.WithContext(ctx).Where("name = ?", t.Obj.Id).Take(&obj).Error; err != nil {
				return nil, err
			}
			tupleModel.ObjId = obj.Id
		}
	}

	return tupleModel, nil
}

type Instance struct {
	Ns  string `validate:"required"`
	Id  string `validate:"required"`
	Rel string
}

type User struct {
	Ns string `validate:"required"`
	Id string `validate:"required"`
}

// type TreeNode struct {
// 	Root        *Instance `validate:"required"`
// 	RelChildren map[string]*TreeNode
// }
