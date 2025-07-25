package model

import (
	"gorm.io/gorm"
)

// Don't use fkey for performance consideration
type Tuple struct {
	SbjNsId       uint64  `gorm:"uniqueIndex:idx_tuple;not null"`
	SbjId         uint64  `gorm:"uniqueIndex:idx_tuple;not null"`
	SbjRelationId *uint64 `gorm:"uniqueIndex:idx_tuple"`
	RelationId    uint64  `gorm:"uniqueIndex:idx_tuple;not null"`
	ObjNsId       uint64  `gorm:"uniqueIndex:idx_tuple;not null"`
	ObjId         uint64  `gorm:"uniqueIndex:idx_tuple;not null"`

	SbjNs       Namespace    `gorm:"foreignKey:SbjNsId"`
	SbjInstance Instance     `gorm:"foreignKey:SbjId"`
	SbjRelation *RelationDef `gorm:"foreignKey:SbjRelationId"`
	RelationDef RelationDef  `gorm:"foreignKey:RelationId"`
	ObjNs       Namespace    `gorm:"foreignKey:ObjNsId"`
	ObjInstance Instance     `gorm:"foreignKey:ObjId"`
}

func (t *Tuple) ToQuery() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if t.SbjNsId != 0 {
			db = db.Where("sbj_ns_id = ?", t.SbjNsId)
		}
		if t.SbjId != 0 {
			db = db.Where("sbj_id = ?", t.SbjId)
		}
		if t.SbjRelationId != nil {
			db = db.Where("sbj_relation_id = ?", *t.SbjRelationId)
		}
		if t.RelationId != 0 {
			db = db.Where("relation_id = ?", t.RelationId)
		}
		if t.ObjNsId != 0 {
			db = db.Where("obj_ns_id = ?", t.ObjNsId)
		}
		if t.ObjId != 0 {
			db = db.Where("obj_id = ?", t.ObjId)
		}
		return db
	}
}

type Namespace struct {
	Id   uint64 `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex"`
}
type RelationDef struct {
	Id   uint64 `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex"`
}

type Instance struct {
	Id   uint64 `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex"`
}
