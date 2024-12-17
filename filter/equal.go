package filter

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func QueryToEqualSearch(c *gin.Context, db *gorm.DB, keys ...string) *gorm.DB {
	for _, v := range keys {
		if c.Query(v) != "" {
			var sb strings.Builder
			stmt := db.Statement

			stmt.QuoteTo(&sb, clause.Column{Table: stmt.Table, Name: v})
			sb.WriteString(" = ?")

			db = db.Where(sb.String(), c.Query(v))
		}
	}
	return db
}

func QueryToOrEqualSearch(c *gin.Context, db *gorm.DB, keys ...string) *gorm.DB {
	for _, v := range keys {
		if c.Query(v) != "" {
			var sb strings.Builder
			stmt := db.Statement

			stmt.QuoteTo(&sb, clause.Column{Table: stmt.Table, Name: v})
			sb.WriteString(" = ?")

			db = db.Or(sb.String(), c.Query(v))
		}
	}
	return db
}
