package filter

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	fuzzyLikeClause = " LIKE ?"
	fuzzyLikeOnce   sync.Once
)

// resolveFuzzyClause picks the fuzzy-match clause for a given dialect name.
// PostgreSQL's LIKE is case-sensitive, so ILIKE keeps behavior consistent with
// MySQL / SQLite (whose LIKE is case-insensitive by default).
func resolveFuzzyClause(dialect string) string {
	if dialect == "postgres" {
		return " ILIKE ?"
	}
	return " LIKE ?"
}

// fuzzyLike returns the fuzzy match clause for the active dialect, resolved
// once per process so callers pay no per-query cost.
func fuzzyLike() string {
	fuzzyLikeOnce.Do(func() {
		fuzzyLikeClause = resolveFuzzyClause(model.DialectName())
	})
	return fuzzyLikeClause
}

func QueryToFussySearch(c *gin.Context, db *gorm.DB, col Column) *gorm.DB {
	if qArr := c.QueryArray(col.QueryKey + "[]"); qArr != nil {
		db = applyFuzzyCondition(c, db, col.DBColumn, qArr)
	} else if q := c.Query(col.QueryKey); q != "" {
		db = applyFuzzyCondition(c, db, col.DBColumn, []string{q})
	}
	return db
}

func applyFuzzyCondition(c *gin.Context, tx *gorm.DB, column string, values []string) *gorm.DB {
	stmt := tx.Statement

	// build column name (column LIKE ?)
	var colBuilder strings.Builder
	stmt.QuoteTo(&colBuilder, clause.Column{Table: stmt.Table, Name: column})
	colBuilder.WriteString(fuzzyLike())

	db := model.UseDB(c)
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

func QueryToFussyKeysSearch(c *gin.Context, tx *gorm.DB, cols ...Column) *gorm.DB {
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

	likeClause := fuzzyLike()
	db := model.UseDB(c)
	var colBuilder strings.Builder
	stmt := tx.Statement

	for _, col := range cols {
		// build column name (column LIKE ?)
		colBuilder.Reset()
		stmt.QuoteTo(&colBuilder, clause.Column{Table: stmt.Table, Name: col.DBColumn})
		colBuilder.WriteString(likeClause)

		db = db.Or(colBuilder.String(), likeValue)
	}

	return tx.Where(db)
}

func QueryToOrFussySearch(c *gin.Context, db *gorm.DB, cols ...Column) *gorm.DB {
	likeClause := fuzzyLike()
	for _, col := range cols {
		if c.Query(col.QueryKey) != "" {
			var sb strings.Builder
			stmt := db.Statement

			stmt.QuoteTo(&sb, clause.Column{Table: stmt.Table, Name: col.DBColumn})

			sb.WriteString(likeClause)

			var sbValue strings.Builder

			_, err := fmt.Fprintf(&sbValue, "%%%s%%", c.Query(col.QueryKey))

			if err != nil {
				logger.Error(err)
				continue
			}

			db = db.Or(sb.String(), sbValue.String())
		}
	}
	return db
}
