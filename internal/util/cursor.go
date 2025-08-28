package util

import (
	"strings"

	"github.com/rs/zerolog/log"
	pkgpbv1 "github.com/skyrocket-qy/protos/gen/pkgpb/v1"
	"gorm.io/gorm"
)

func ApplyCursor(c *pkgpbv1.CursorData) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if c == nil || len(c.GetFields()) == 0 {
			return db
		}

		var (
			orConditions []string
			orArgs       [][]any
		)

		for i := range c.GetFields() {
			var (
				condParts []string
				args      []any
			)

			// Equal conditions for previous fields

			for j := range i {
				condParts = append(condParts, c.GetFields()[j].GetCol()+" = ?")
				args = append(args, c.GetFields()[j].GetVal())
			}

			// Comparison condition for current field
			op := ">"
			if !c.GetFields()[i].GetAsc() {
				op = "<"
			}

			condParts = append(condParts, c.GetFields()[i].GetCol()+" "+op+" ?")
			args = append(args, c.GetFields()[i].GetVal())

			orConditions = append(orConditions, "("+strings.Join(condParts, " AND ")+")")
			orArgs = append(orArgs, args)
		}

		// Combine all OR conditions
		fullWhere := strings.Join(orConditions, " OR ")

		var allArgs []any
		for _, args := range orArgs {
			allArgs = append(allArgs, args...)
		}

		log.Print(fullWhere)
		log.Print(allArgs)

		return db.Where(fullWhere, allArgs...)
	}
}
