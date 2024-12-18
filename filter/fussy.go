package filter

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func QueryToFussySearch(c *gin.Context, db *gorm.DB, key string) *gorm.DB {
	if qArr := c.QueryArray(key + "[]"); qArr != nil {
		db = applyFuzzyCondition(db, key, qArr)
	} else if q := c.Query(key); q != "" {
		db = applyFuzzyCondition(db, key, []string{q})
	}
	return db
}

func applyFuzzyCondition(tx *gorm.DB, column string, values []string) *gorm.DB {
	stmt := tx.Statement

	// build column name (column LIKE ?)
	var colBuilder strings.Builder
	stmt.QuoteTo(&colBuilder, clause.Column{Table: stmt.Table, Name: column})
	colBuilder.WriteString(" LIKE ?")

	db := model.UseDB()
	var valueBuilder strings.Builder

	for _, value := range values {
		// build value for query (%value%)
		valueBuilder.Reset()
		valueBuilder.WriteString("%")
		valueBuilder.WriteString(value)
		valueBuilder.WriteString("%")

		db = db.Or(colBuilder.String(), valueBuilder.String())
	}

	return tx.Where(db)
}

func QueryToFussyKeysSearch(c *gin.Context, tx *gorm.DB, keys ...string) *gorm.DB {
	value := c.Query("search")
	if value == "" {
		return tx
	}

	// build value for query (%value%)
	var valueBuilder strings.Builder
	valueBuilder.WriteString("%")
	valueBuilder.WriteString(value)
	valueBuilder.WriteString("%")
	likeValue := valueBuilder.String()

	db := model.UseDB()
	var colBuilder strings.Builder

	for _, v := range keys {
		// build column name (column LIKE ?)
		colBuilder.Reset()
		colBuilder.WriteString(v)
		colBuilder.WriteString(" LIKE ?")

		db = db.Or(colBuilder.String(), likeValue)
	}

	return tx.Where(db)
}

func QueryToOrFussySearch(c *gin.Context, db *gorm.DB, keys ...string) *gorm.DB {
	for _, v := range keys {
		if c.Query(v) != "" {
			var sb strings.Builder
			stmt := db.Statement

			stmt.QuoteTo(&sb, clause.Column{Table: stmt.Table, Name: v})

			sb.WriteString(" LIKE ?")

			var sbValue strings.Builder

			_, err := fmt.Fprintf(&sbValue, "%%%s%%", c.Query(v))

			if err != nil {
				logger.Error(err)
				continue
			}

			db = db.Or(sb.String(), sbValue.String())
		}
	}
	return db
}
