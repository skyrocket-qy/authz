package rbac

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

	UserAuths []*UserAuth `gorm:"foreignKey:UserId"`
}

type UserAuth struct {
	Id         uint64  `gorm:"primaryKey"`
	Type       string  `gorm:"type:varchar(20);not null"` // "password", "google", "github", etc.
	ProviderId *string `gorm:"type:varchar(255)"`         // OAuth user ID / sub claim

	UserId uint `gorm:"not null"`
}

// metadata
type Role struct {
	Id   uint64 `gorm:"primarykey"`
	Name string `gorm:"unique;type:varchar(255);not null"`
}

// metadata
type Resource struct {
	Id   uint64 `gorm:"primarykey"`
	Ns   string `gorm:"uniqueindex:idx_ns_name;type:varchar(255);not null"`
	Name string `gorm:"uniqueindex:idx_ns_name;type:varchar(255);not null"`
}
