package filter

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func QueriesToInSearch(c *gin.Context, db *gorm.DB, cols ...Column) *gorm.DB {
	for _, col := range cols {
		db = QueryToInSearch(c, db, col)
	}
	return db
}

func QueryToInSearch(c *gin.Context, db *gorm.DB, col Column) *gorm.DB {
	queryArray := c.QueryArray(col.QueryKey + "[]")
	if len(queryArray) == 0 {
		queryArray = c.QueryArray(col.QueryKey)
	}
	if len(queryArray) == 1 && queryArray[0] == "" {
		return db
	}

	if len(queryArray) >= 1 {
		var builder strings.Builder
		stmt := db.Statement

		stmt.QuoteTo(&builder, clause.Column{Table: stmt.Table, Name: col.DBColumn})
		builder.WriteString(" IN ?")

		return db.Where(builder.String(), queryArray)
	}
	return db
}

func QueryToOrInSearch(c *gin.Context, db *gorm.DB, cols ...Column) *gorm.DB {
	for _, col := range cols {
		queryArray := c.QueryArray(col.QueryKey + "[]")
		if len(queryArray) == 0 {
			queryArray = c.QueryArray(col.QueryKey)
		}
		if len(queryArray) == 1 && queryArray[0] == "" {
			continue
		}
		if len(queryArray) >= 1 {
			var sb strings.Builder
			stmt := db.Statement

			stmt.QuoteTo(&sb, clause.Column{Table: stmt.Table, Name: col.DBColumn})
			sb.WriteString(" IN ?")

			db = db.Or(sb.String(), queryArray)
		}
	}
	return db
}
