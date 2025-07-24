package model

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
