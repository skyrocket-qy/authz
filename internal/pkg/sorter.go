package pkg

import (
	"strings"

	pkgpbv1 "github.com/skyrocket-qy/protos/gen/pkgpb/v1"
	"gorm.io/gorm"
)

func ApplySorter(seqSorters []*pkgpbv1.Sorter, dfSort ...*pkgpbv1.Sorter) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(seqSorters) == 0 {
			if len(dfSort) == 0 {
				return db
			}

			df := dfSort[0]
			expr := df.Field
			if !df.Asc {
				expr += " DESC"
			}
			return db.Order(expr)
		}

		for _, sorter := range seqSorters {
			expr := ToPascalCase(sorter.Field)
			if !sorter.Asc {
				expr += " DESC"
			}
			db = db.Order(expr)
		}
		return db
	}
}

func ToPascalCase(input string) string {
	if len(input) == 0 {
		return ""
	}

	// Capitalize the first character
	result := strings.ToUpper(string(input[0])) + input[1:]

	return result
}
