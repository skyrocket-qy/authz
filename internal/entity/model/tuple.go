package model

// Don't use fkey for performance consideration
// type Tuple struct {
// 	SbjId      uint64 `gorm:"uniqueIndex:idx_tuple;not null"`
// 	RelationId uint64 `gorm:"uniqueIndex:idx_tuple;not null"`
// 	ObjId      uint64 `gorm:"uniqueIndex:idx_tuple;not null"`

// 	Sbj      *Vertex `gorm:"foreignKey:SbjId"`
// 	Relation *Edge   `gorm:"foreignKey:RelationId"`
// 	Obj      *Vertex `gorm:"foreignKey:ObjId"`
// }

// type Vertex struct {
// 	Id        uint64  `gorm:"primaryKey"`
// 	Namespace string  `gorm:"uniqueIndex:idx_ns_name_relation;not null"`
// 	Name      string  `gorm:"uniqueIndex:idx_ns_name_relation;not null"`
// 	Relation  *string `gorm:"uniqueIndex:idx_ns_name_relation"`
// }

// type Edge struct {
// 	Id       uint64 `gorm:"primaryKey"`
// 	Relation string `gorm:"uniqueIndex;not null"`
// }

type Tuple struct {
	SbjNs    string `gorm:"uniqueIndex:idx_tuple;not null"`
	SbjId    string `gorm:"uniqueIndex:idx_tuple;not null"`
	Relation string `gorm:"uniqueIndex:idx_tuple;not null"`
	ObjNs    string `gorm:"uniqueIndex:idx_tuple;not null"`
	ObjId    string `gorm:"uniqueIndex:idx_tuple;not null"`
}
