package cosy

import (
	"strings"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/uozi-tech/cosy/model"
	"gorm.io/gorm"
)

type Ctx[T any] struct {
	// Place potentially large/aligned generics first to minimize padding
	Model               T
	OriginModel         T
	ResultData          any
	DefaultResponseData any

	// Pointer-heavy fields
	*gin.Context
	Tx          *gorm.DB
	nextHandler *gin.HandlerFunc
	listService *ListService[T]

	// Function pointers
	scan            func(tx *gorm.DB) any
	transformer     func(*T) any
	responseBuilder func(ctx *Ctx[T])

	// Fixed-size and map headers
	ID                      model.IDType
	rules                   gin.H
	Payload                 map[string]any
	selectedFields          map[string]bool
	selectedFieldCandidates map[string]bool
	columnWhiteList         map[string]bool
	columnMapping           map[string]string

	// Strings (16B) grouped
	table   string
	itemKey string

	// Slice headers (24B) grouped
	BatchEffectedIDs      []string
	tableArgs             []any
	prepareHookFunc       []func(ctx *Ctx[T])
	beforeDecodeHookFunc  []func(ctx *Ctx[T])
	beforeExecuteHookFunc []func(ctx *Ctx[T])
	executedHookFunc      []func(ctx *Ctx[T])
	gormScopes            []func(tx *gorm.DB) *gorm.DB
	preloads              []string
	joins                 []string
	unique                []string

	// Packed bools at the end to avoid repeated padding
	useTransaction           bool
	abort                    bool
	skipAssociationsOnCreate bool
	permanentlyDelete        bool
}

func Core[T any](c *gin.Context) *Ctx[T] {
	ctx := &Ctx[T]{
		Context:                  c,
		Tx:                       model.UseDB(c),
		rules:                    make(gin.H),
		gormScopes:               make([]func(tx *gorm.DB) *gorm.DB, 0),
		prepareHookFunc:          make([]func(ctx *Ctx[T]), 0),
		beforeExecuteHookFunc:    make([]func(ctx *Ctx[T]), 0),
		beforeDecodeHookFunc:     make([]func(ctx *Ctx[T]), 0),
		executedHookFunc:         make([]func(ctx *Ctx[T]), 0),
		itemKey:                  "id",
		skipAssociationsOnCreate: true,
		columnWhiteList:          make(map[string]bool),
		columnMapping:            make(map[string]string),
		selectedFields:           make(map[string]bool),
		selectedFieldCandidates:  make(map[string]bool),
	}

	ctx.listService = &ListService[T]{
		ctx:           ctx,
		customFilters: orderedmap.NewOrderedMap[string, string](),
	}

	return ctx
}

func (c *Ctx[T]) SetTable(table string, args ...any) *Ctx[T] {
	c.table = table
	c.tableArgs = args
	return c
}

func (c *Ctx[T]) SetItemKey(key string) *Ctx[T] {
	c.itemKey = key
	return c
}

func (c *Ctx[T]) SetValidRules(rules gin.H) *Ctx[T] {
	for k, rule := range rules {
		c.rules[k] = rule
		rule := cast.ToString(rule)
		if strings.Contains(rule, "db_unique") {
			c.unique = append(c.unique, k)
			rules[k] = strings.ReplaceAll(rule, "db_unique", "")
			rules[k] = strings.TrimRight(rule, ",")
		}
	}

	return c
}

func (c *Ctx[T]) SetUnique(keys ...string) {
	c.unique = append(c.unique, keys...)
}

func (c *Ctx[T]) SetPreloads(args ...string) *Ctx[T] {
	c.preloads = append(c.preloads, args...)
	return c
}

func (c *Ctx[T]) SetJoins(args ...string) *Ctx[T] {
	c.joins = append(c.joins, args...)
	return c
}

func (c *Ctx[T]) SetScan(scan func(tx *gorm.DB) any) *Ctx[T] {
	c.scan = scan
	return c
}

func (c *Ctx[T]) SetTransformer(t func(m *T) any) *Ctx[T] {
	c.transformer = t
	return c
}

func (c *Ctx[T]) AddColWhiteList(cols ...string) *Ctx[T] {
	for _, col := range cols {
		c.columnWhiteList[col] = true
	}
	return c
}

func (c *Ctx[T]) resolveColumn(queryKey string) string {
	if column, ok := c.columnMapping[queryKey]; ok && column != "" {
		return column
	}

	return queryKey
}

func (c *Ctx[T]) AddSelectedFields(fields ...string) *Ctx[T] {
	selectedFields := c.selectedFields
	if c.Payload == nil {
		selectedFields = c.selectedFieldCandidates
	}

	for _, field := range fields {
		selectedFields[field] = true
	}
	return c
}

func (c *Ctx[T]) GetSelectedFields() []string {
	selectedFields := make(map[string]bool, len(c.selectedFields)+len(c.selectedFieldCandidates))
	for field := range c.selectedFields {
		selectedFields[field] = true
	}
	for field := range c.selectedFieldCandidates {
		if c.payloadContainsSelectedField(field) {
			selectedFields[field] = true
		}
	}

	var fields []string
	for field := range selectedFields {
		fields = append(fields, field)
	}
	return fields
}

func (c *Ctx[T]) payloadContainsSelectedField(field string) bool {
	if c.Payload == nil {
		return false
	}

	if c.payloadMapContainsSelectedField(c.Payload, field) {
		return true
	}

	data, ok := c.Payload["data"].(map[string]any)
	return ok && c.payloadMapContainsSelectedField(data, field)
}

func (c *Ctx[T]) payloadMapContainsSelectedField(payload map[string]any, field string) bool {
	if _, ok := payload[field]; ok {
		return true
	}

	for key := range payload {
		if c.resolveColumn(key) == field {
			return true
		}
	}

	return false
}

// WithoutSortOrder disable sort order for "get list"
func (c *Ctx[T]) WithoutSortOrder() *Ctx[T] {
	c.listService.disableSortOrder = true
	return c
}

// WithTransaction use transaction for "create" and "update"
func (c *Ctx[T]) WithTransaction() *Ctx[T] {
	c.useTransaction = true
	c.Tx = c.Tx.Begin()
	return c
}
