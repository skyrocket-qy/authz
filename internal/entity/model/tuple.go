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
	Id       uint64 `gorm:"primaryKey"`
	SbjNs    string `gorm:"uniqueIndex:idx_tuple;not null"`
	SbjId    string `gorm:"uniqueIndex:idx_tuple;not null"`
	Relation string `gorm:"uniqueIndex:idx_tuple;not null"`
	ObjNs    string `gorm:"uniqueIndex:idx_tuple;not null"`
	ObjId    string `gorm:"uniqueIndex:idx_tuple;not null"`
}

type ChangeLog struct {
	Id    uint64 `gorm:"primaryKey"`
	Tuple []byte // marshal tuple into byte data to reduce storage
}

// only one latest one record
type GraphCheckpoint struct {
	LastChangeLogID uint64 `gorm:"primaryKey"` // up to which changelog was applied
	Data            []byte // marshaled in-memory graph
}

/*
1. Every successful write (create/delete of a tuple) will append a ChangeLog entry.
2. Periodically (based on time or ChangeLog size), create a GraphCheckpoint by serializing the current in-memory graph.
3. On service restart or graph desync:
   a. First attempt to replay ChangeLog entries from the latest checkpoint.
   b. If that fails (e.g., checkpoint corrupted), rebuild from scratch using the full ChangeLog.
4. Periodically purge old ChangeLog entries that are already reflected in a persisted checkpoint.
*/
