package filter

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func QueriesToBetweenSearch(c *gin.Context, db *gorm.DB, keys ...string) *gorm.DB {
	for _, v := range keys {
		db = QueryToBetweenSearch(c, db, v)
	}
	return db
}

func QueryToBetweenSearch(c *gin.Context, db *gorm.DB, value string, key ...string) *gorm.DB {
	queryArray := c.QueryArray(value + "[]")
	if len(queryArray) == 0 {
		queryArray = c.QueryArray(value)
	}
	if len(queryArray) <= 1 {
		return db
	}

	if len(queryArray) == 2 && queryArray[0] != "" && queryArray[1] != "" {
		var builder strings.Builder
		stmt := db.Statement

		column := value
		if len(key) != 0 {
			column = key[0]
		}

		stmt.QuoteTo(&builder, clause.Column{Table: stmt.Table, Name: column})
		builder.WriteString(" BETWEEN ? AND ?")

		return db.Where(builder.String(), queryArray[0], queryArray[1])
	}
	return db
}
