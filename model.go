package zanzibarx

type RelationTuple struct {
	SbjNs       uint `gorm:"uniqueIndex:idx_tuple"`
	SbjId       uint `gorm:"uniqueIndex:idx_tuple"`
	SbjRelation uint `gorm:"uniqueIndex:idx_tuple"`
	ObjNs       uint `gorm:"uniqueIndex:idx_tuple"`
	ObjId       uint `gorm:"uniqueIndex:idx_tuple"`
	ObjRelation uint `gorm:"uniqueIndex:idx_tuple"`

	SNs              Namespace `gorm:"foreignKey:NamespaceID;references:Id"`
	S                Instance  `gorm:"foreignKey:ObjectID;references:Id"`
	Relation         Relation  `gorm:"foreignKey:RelationID;references:Id"`
	SubjectNamespace Namespace `gorm:"foreignKey:SubjectNsID;references:Id"`
}

type Namespace struct {
	Id   uint   `gorm:"primarykey"`
	Name string `gorm:"unique"`
}

type Instance struct {
	Id   uint   `gorm:"primarykey"`
	Name string `gorm:"unique"`
}

type Relation struct {
	Id   uint   `gorm:"primarykey"`
	Name string `gorm:"unique"`
}
