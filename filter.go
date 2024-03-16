package cosy

import (
	"fmt"
	"git.uozi.org/uozi/cosy/logger"
	"git.uozi.org/uozi/cosy/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

func (c *Ctx[T]) SetFussy(keys ...string) *Ctx[T] {
	c.fussy = append(c.fussy, keys...)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return QueryToFussySearch(c.Context, tx, keys...)
	})
	return c
}

func (c *Ctx[T]) SetSearchFussyKeys(keys ...string) *Ctx[T] {
	c.search = append(c.search, keys...)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return QueryToFussyKeysSearch(c.Context, tx, keys...)
	})
	return c
}

func (c *Ctx[T]) SetEqual(keys ...string) *Ctx[T] {
	c.eq = append(c.eq, keys...)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return QueryToEqualSearch(c.Context, tx, keys...)
	})
	return c
}

func (c *Ctx[T]) SetIn(keys ...string) *Ctx[T] {
	c.in = append(c.in, keys...)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return QueryToInSearch(c.Context, tx, keys...)
	})
	return c
}

func (c *Ctx[T]) SetOrFussy(keys ...string) *Ctx[T] {
	c.orFussy = append(c.orFussy, keys...)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return QueryToOrFussySearch(c.Context, tx, keys...)
	})
	return c
}

func (c *Ctx[T]) SetOrEqual(keys ...string) *Ctx[T] {
	c.orEq = append(c.orEq, keys...)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return QueryToOrEqualSearch(c.Context, tx, keys...)
	})
	return c
}

func (c *Ctx[T]) SetBetween(keys ...string) *Ctx[T] {
	c.between = append(c.between, keys...)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return QueryToBetweenSearch(c.Context, tx, keys...)
	})
	return c
}

func (c *Ctx[T]) SetOrIn(keys ...string) *Ctx[T] {
	c.orIn = append(c.orIn, keys...)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return QueryToOrInSearch(c.Context, tx, keys...)
	})
	return c
}

func QueryToInSearch(c *gin.Context, db *gorm.DB, keys ...string) *gorm.DB {
	for _, v := range keys {
		queryArray := c.QueryArray(v + "[]")
		if len(queryArray) == 0 {
			queryArray = c.QueryArray(v)
		}
		if len(queryArray) == 1 && queryArray[0] == "" {
			continue
		}
		if len(queryArray) >= 1 {
			var builder strings.Builder
			stmt := db.Statement

			stmt.QuoteTo(&builder, clause.Column{Table: stmt.Table, Name: v})
			builder.WriteString(" IN ?")

			db = db.Where(builder.String(), queryArray)
		}
	}
	return db
}

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

func QueryToFussySearch(c *gin.Context, db *gorm.DB, keys ...string) *gorm.DB {
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

			db = db.Where(sb.String(), sbValue.String())
		}
	}
	return db
}

func QueryToFussyKeysSearch(c *gin.Context, tx *gorm.DB, keys ...string) *gorm.DB {
	value := c.Query("search")
	if value == "" {
		return tx
	}

	var condition *gorm.DB
	db := model.UseDB()
	for i, v := range keys {
		sb := v + " LIKE ?"
		sv := "%" + value + "%"

		switch i {
		case 0:
			condition = db.Where(db.Where(sb, sv))
		default:
			condition = condition.Or(sb, sv)
		}
	}

	return tx.Where(condition)
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

func QueryToBetweenSearch(c *gin.Context, db *gorm.DB, keys ...string) *gorm.DB {
	for _, v := range keys {
		queryArray := c.QueryArray(v + "[]")
		if len(queryArray) == 0 {
			queryArray = c.QueryArray(v)
		}
		if len(queryArray) <= 1 {
			continue
		}
		if len(queryArray) == 2 && queryArray[0] != "" && queryArray[1] != "" {
			var builder strings.Builder
			stmt := db.Statement

			stmt.QuoteTo(&builder, clause.Column{Table: stmt.Table, Name: v})
			builder.WriteString(" BETWEEN ? AND ?")

			db = db.Where(builder.String(), queryArray[0], queryArray[1])
		}
	}
	return db
}
