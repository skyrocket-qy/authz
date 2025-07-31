package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	Id        uint64 `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Email            string `gorm:"unique;not null"`
	IsEmailConfirmed bool
	Password         *string `gorm:"type:varchar(255)"`
	Name             string  `gorm:"unique;type:varchar(255);not null"`
	IsActive         bool    `gorm:"default:true"`

	Orgs      []*Org      `gorm:"many2many:user_orgs;"`
	UserAuths []*UserAuth `gorm:"foreignKey:UserId"`
}

type UserAuth struct {
	Id         uint64  `gorm:"primaryKey"`
	AuthType   string  `gorm:"type:varchar(20);not null"` // "password", "google", "github", etc.
	ProviderId *string `gorm:"type:varchar(255)"`         // OAuth user ID / sub claim

	UserId uint `gorm:"not null"`
}

type Org struct {
	gorm.Model
	Name  string  `gorm:"unique;type:varchar(255);not null"`
	Users []*User `gorm:"many2many:user_orgs;"`
}

// metadata
type Role struct {
	Id   uint64 `gorm:"primarykey"`
	Name string `gorm:"unique;type:varchar(255);not null"`
}
