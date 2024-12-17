package filter

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func QueriesToInSearch(c *gin.Context, db *gorm.DB, keys ...string) *gorm.DB {
	for _, v := range keys {
		QueryToInSearch(c, db, v)
	}
	return db
}

func QueryToInSearch(c *gin.Context, db *gorm.DB, value string, key ...string) *gorm.DB {
	queryArray := c.QueryArray(value + "[]")
	if len(queryArray) == 0 {
		queryArray = c.QueryArray(value)
	}
	if len(queryArray) == 1 && queryArray[0] == "" {
		return db
	}

	if len(queryArray) >= 1 {
		var builder strings.Builder
		stmt := db.Statement

		column := value
		if len(key) != 0 {
			column = key[0]
		}

		stmt.QuoteTo(&builder, clause.Column{Table: stmt.Table, Name: column})
		builder.WriteString(" IN ?")

		return db.Where(builder.String(), queryArray)
	}
	return db
}

func QueryToOrInSearch(c *gin.Context, db *gorm.DB, keys ...string) *gorm.DB {
	for _, v := range keys {
		queryArray := c.QueryArray(v + "[]")
		if len(queryArray) == 0 {
			queryArray = c.QueryArray(v)
		}
		if len(queryArray) == 1 && queryArray[0] == "" {
			continue
		}
		if len(queryArray) >= 1 {
			var sb strings.Builder
			stmt := db.Statement

			stmt.QuoteTo(&sb, clause.Column{Table: stmt.Table, Name: v})
			sb.WriteString(" IN ?")

			db = db.Or(sb.String(), queryArray)
		}
	}
	return db
}
