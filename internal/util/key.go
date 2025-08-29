package util

import (
	"context"

	"gorm.io/gorm"
)

type DBCtxKey struct{}

func WithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, DBCtxKey{}, db)
}
