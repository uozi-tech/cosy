package filter

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func QueriesToBetweenSearch(c *gin.Context, db *gorm.DB, cols ...Column) *gorm.DB {
	for _, col := range cols {
		db = QueryToBetweenSearch(c, db, col)
	}
	return db
}

func QueryToBetweenSearch(c *gin.Context, db *gorm.DB, col Column) *gorm.DB {
	queryArray := c.QueryArray(col.QueryKey + "[]")
	if len(queryArray) == 0 {
		queryArray = c.QueryArray(col.QueryKey)
	}
	if len(queryArray) <= 1 {
		return db
	}

	if len(queryArray) == 2 && queryArray[0] != "" && queryArray[1] != "" {
		var builder strings.Builder
		stmt := db.Statement

		stmt.QuoteTo(&builder, clause.Column{Table: stmt.Table, Name: col.DBColumn})
		builder.WriteString(" BETWEEN ? AND ?")

		return db.Where(builder.String(), queryArray[0], queryArray[1])
	}
	return db
}
