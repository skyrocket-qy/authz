package zanzibarx

import "gorm.io/gorm"

type RelationTupleModel struct {
	SbjNsId       uint `gorm:"uniqueIndex:idx_tuple"`
	SbjId         uint `gorm:"uniqueIndex:idx_tuple"`
	RelationId    uint `gorm:"uniqueIndex:idx_tuple"`
	ObjNsId       uint `gorm:"uniqueIndex:idx_tuple"`
	ObjId         uint `gorm:"uniqueIndex:idx_tuple"`
	ObjRelationId uint `gorm:"uniqueIndex:idx_tuple"`

	SbjNs       Namespace `gorm:"foreignKey:SbjNsId;references:Id"`
	Sbj         Instance  `gorm:"foreignKey:SbjId;references:Id"`
	Relation    Relation  `gorm:"foreignKey:RelationId;references:Id"`
	ObjNs       Namespace `gorm:"foreignKey:ObjNsId;references:Id"`
	Obj         Instance  `gorm:"foreignKey:ObjId;references:Id"`
	ObjRelation Relation  `gorm:"foreignKey:ObjRelationId;references:Id"`
}

type RelationTuple struct {
	SbjNs       string
	Sbj         string
	Relation    string
	ObjNs       string
	Obj         string
	ObjRelation string
}

func (r *RelationTuple) Create(db *gorm.DB) error {
	var sbjNs, objNs Namespace
	var sbj, obj Instance
	var rel, objRel Relation

	// Upsert (find or create) each string reference
	if err := db.FirstOrCreate(&sbjNs, Namespace{Name: r.SbjNs}).Error; err != nil {
		return err
	}
	if err := db.FirstOrCreate(&objNs, Namespace{Name: r.ObjNs}).Error; err != nil {
		return err
	}
	if err := db.FirstOrCreate(&sbj, Instance{Name: r.Sbj}).Error; err != nil {
		return err
	}
	if err := db.FirstOrCreate(&obj, Instance{Name: r.Obj}).Error; err != nil {
		return err
	}
	if err := db.FirstOrCreate(&rel, Relation{Name: r.Relation}).Error; err != nil {
		return err
	}
	if err := db.FirstOrCreate(&objRel, Relation{Name: r.ObjRelation}).Error; err != nil {
		return err
	}

	// Compose and insert the RelationTuple
	tuple := RelationTupleModel{
		SbjNsId:       sbjNs.Id,
		SbjId:         sbj.Id,
		RelationId:    rel.Id,
		ObjNsId:       objNs.Id,
		ObjId:         obj.Id,
		ObjRelationId: objRel.Id,
	}

	return db.Create(&tuple).Error
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
