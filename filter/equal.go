package filter

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func QueryToEqualSearch(c *gin.Context, db *gorm.DB, cols ...Column) *gorm.DB {
	for _, col := range cols {
		if c.Query(col.QueryKey) != "" {
			var sb strings.Builder
			stmt := db.Statement

			stmt.QuoteTo(&sb, clause.Column{Table: stmt.Table, Name: col.DBColumn})
			sb.WriteString(" = ?")

			db = db.Where(sb.String(), c.Query(col.QueryKey))
		}
	}
	return db
}

func QueryToOrEqualSearch(c *gin.Context, db *gorm.DB, cols ...Column) *gorm.DB {
	for _, col := range cols {
		if c.Query(col.QueryKey) != "" {
			var sb strings.Builder
			stmt := db.Statement

			stmt.QuoteTo(&sb, clause.Column{Table: stmt.Table, Name: col.DBColumn})
			sb.WriteString(" = ?")

			db = db.Or(sb.String(), c.Query(col.QueryKey))
		}
	}
	return db
}
