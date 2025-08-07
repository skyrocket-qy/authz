package pkg

import (
	"fmt"
	"strings"

	pkgpbv1 "github.com/skyrocket-qy/protos/gen/pkgpb/v1"
	"gorm.io/gorm"
)

func ApplyCursor(c *pkgpbv1.CursorData) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(c.Fields) == 0 {
			return db
		}

		var orConditions []string
		var orArgs [][]any

		for i := range c.Fields {
			var condParts []string
			var args []any

			// Equal conditions for previous fields
			for j := 0; j < i; j++ {
				condParts = append(condParts, c.Fields[j].Col+" = ?")
				args = append(args, c.Fields[j].Val)
			}

			// Comparison condition for current field
			op := ">"
			if !c.Fields[i].Asc {
				op = "<"
			}
			condParts = append(condParts, c.Fields[i].Col+" "+op+" ?")
			args = append(args, c.Fields[i].Val)

			orConditions = append(orConditions, "("+strings.Join(condParts, " AND ")+")")
			orArgs = append(orArgs, args)
		}

		// Combine all OR conditions
		fullWhere := strings.Join(orConditions, " OR ")

		var allArgs []any
		for _, args := range orArgs {
			allArgs = append(allArgs, args...)
		}

		fmt.Println(fullWhere)
		fmt.Println(allArgs)
		return db.Where(fullWhere, allArgs...)
	}
}
