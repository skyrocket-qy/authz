package pkg

import (
	pkgpbv1 "github.com/skyrocket-qy/protos/gen/pkgpb/v1"
	"gorm.io/gorm"
)

func ApplyPager(pager *pkgpbv1.Pager) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if pager == nil {
			return db
		}

		size := int(pager.Size)
		number := int(pager.Number)
		return db.
			Offset(size * (number - 1)).
			Limit(size)
	}
}
