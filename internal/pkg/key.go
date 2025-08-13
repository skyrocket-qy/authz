package pkg

import (
	"context"

	"gorm.io/gorm"
)

type DbCtxKey struct{}

func WithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, DbCtxKey{}, db)
}
