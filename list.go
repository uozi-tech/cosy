package cosy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/model"
	"github.com/uozi-tech/cosy/settings"
	"gorm.io/gorm"
)

type ListService[T any] struct {
	ctx              *Ctx[T]
	disableSortOrder bool
	in               []string
	eq               []string
	fussy            []string
	orIn             []string
	orEq             []string
	orFussy          []string
	search           []string
	between          []string
	customFilters    *orderedmap.OrderedMap[string, string]
}

// GetPagingParams get paging params
func GetPagingParams(c *gin.Context) (page, offset, pageSize int) {
	page = cast.ToInt(c.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize = cast.ToInt(settings.AppSettings.PageSize)
	reqPageSize := c.Query("page_size")
	if reqPageSize != "" {
		pageSize = cast.ToInt(reqPageSize)
	}
	offset = (page - 1) * pageSize
	return
}

// combineStdSelectorRequest combine std selector request
func (c *Ctx[T]) combineStdSelectorRequest() {
	StdSelectorInitID := c.QueryArray("id[]")

	if len(StdSelectorInitID) == 0 {
		return
	}

	c.GormScope(func(tx *gorm.DB) *gorm.DB {
		var sb strings.Builder
		_, err := fmt.Fprintf(&sb, "%s IN ?", c.itemKey)
		if err != nil {
			logger.Error(err)
			return tx
		}
		return tx.Where(sb.String(), StdSelectorInitID)
	})
}

// result get result
func (c *Ctx[T]) result() *gorm.DB {
	var dbModel T
	result := model.UseDB(c.Context)

	if cast.ToBool(c.Query("trash")) {
		tableName := c.table
		if c.table == "" {
			stmt := &gorm.Statement{DB: model.UseDB(c.Context)}
			err := stmt.Parse(&dbModel)
			if err != nil {
				c.AbortWithError(err)
				return nil
			}
			tableName = stmt.Schema.Table
		}

		resolvedModel := model.GetResolvedModel[T]()
		if deletedAt, ok := resolvedModel.Fields["DeletedAt"]; !ok ||
			(deletedAt.DefaultValue == "" || deletedAt.DefaultValue == "null") {
			result = result.Unscoped().Where(tableName + ".deleted_at IS NOT NULL")
		} else {
			result = result.Unscoped().Where(tableName + ".deleted_at != 0")
		}
	}

	result = result.Model(&dbModel)

	c.handleTable()

	c.combineStdSelectorRequest()

	c.applyGormScopes(result)

	return result
}

// resolveData resolve data
func (c *Ctx[T]) resolveData(result *gorm.DB) (data any) {
	// has scanner
	if c.scan != nil {
		return c.scan(result)
	}

	models := make([]*T, 0)
	result.Find(&models)

	// no transformer
	if c.transformer == nil {
		return models
	}

	// use transformer
	transformed := make([]any, 0)
	for k := range models {
		transformed = append(transformed, c.transformer(models[k]))
	}
	return transformed
}

// ListAllData return list all data
func (c *Ctx[T]) ListAllData() (data any) {
	result := c.result()
	data = c.resolveData(result)
	return data
}

// PagingListData return paging list data
func (c *Ctx[T]) PagingListData() *model.DataList {
	result := c.result()

	scopesResult := result.Scopes(c.paginate)

	data := &model.DataList{}
	data.Data = c.resolveData(scopesResult)

	var totalRecords int64
	delete(result.Statement.Clauses, "ORDER BY")
	delete(result.Statement.Clauses, "LIMIT")
	result.Count(&totalRecords)

	page := cast.ToInt(c.Query("page"))
	if page == 0 {
		page = 1
	}

	pageSize := settings.AppSettings.PageSize
	if reqPageSize := c.Query("page_size"); reqPageSize != "" {
		pageSize = cast.ToInt(reqPageSize)
	}

	data.Pagination = model.Pagination{
		Total:       totalRecords,
		PerPage:     pageSize,
		CurrentPage: page,
		TotalPages:  model.TotalPage(totalRecords, pageSize),
	}
	return data
}

func (c *Ctx[T]) prepareListHook(ctx *Ctx[T]) {
	getListHook[T]()(ctx)
	ctx.resolvePreloadWithScope()
	ctx.resolveJoinsWithScopes()
	prepareHook(ctx)
}

// PagingList return paging list
func (c *Ctx[T]) PagingList() {
	NewProcessChain(c).
		SetPrepare(c.prepareListHook).
		SetBeforeExecute(beforeExecuteHook[T]).
		SetExecuted(executedHook[T]).
		SetResponse(func(ctx *Ctx[T]) {
			data := c.PagingListData()
			c.JSON(http.StatusOK, data)
		}).
		GetOrGetList()
}

// EmptyPagingList return empty list
func (c *Ctx[T]) EmptyPagingList() {
	pageSize := settings.AppSettings.PageSize
	if reqPageSize := c.Query("page_size"); reqPageSize != "" {
		pageSize = cast.ToInt(reqPageSize)
	}

	data := &model.DataList{Data: make([]any, 0)}
	data.Pagination.PerPage = pageSize
	c.JSON(http.StatusOK, data)
}

// List return all list data
func (c *Ctx[T]) List() {
	NewProcessChain(c).
		SetPrepare(c.prepareListHook).
		SetBeforeExecute(beforeExecuteHook[T]).
		SetExecuted(executedHook[T]).
		SetResponse(func(ctx *Ctx[T]) {
			c.JSON(http.StatusOK,
				model.DataList{
					Data: c.ListAllData(),
				})
		}).
		GetOrGetList()
}
