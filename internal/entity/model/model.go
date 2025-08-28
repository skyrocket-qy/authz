package model

// TODO: use uint64 replace string for best performance but consider convert overhead.
type Tuple struct {
	Id       uint64 `gorm:"primaryKey"                     json:"id"`
	SbjNs    string `gorm:"uniqueIndex:idx_tuple;not null" json:"sbj_ns"`
	SbjId    string `gorm:"uniqueIndex:idx_tuple;not null" json:"sbj_id"`
	Relation string `gorm:"uniqueIndex:idx_tuple;not null" json:"relation"`
	ObjNs    string `gorm:"uniqueIndex:idx_tuple;not null" json:"obj_ns"`
	ObjId    string `gorm:"uniqueIndex:idx_tuple;not null" json:"obj_id"`
}

type GraphCheckpoint struct {
	LastOffset int64  `gorm:"primaryKey"` // up to which changelog was applied
	Data       []byte // marshaled in-memory graph
}
