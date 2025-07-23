package zanzibarx

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}


func (r *Repository) Create