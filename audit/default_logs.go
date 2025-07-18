package audit

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy"
	"github.com/uozi-tech/cosy/logger"
	cModel "github.com/uozi-tech/cosy/model"
	cSettings "github.com/uozi-tech/cosy/settings"
)

func GetDefaultLogs(c *gin.Context) {
	if !cSettings.SLSSettings.Enable() {
		c.JSON(http.StatusOK, cModel.DataList{})
		return
	}
	s := logger.NewSessionLogger(c)
	var query struct {
		Page     int64  `form:"page"`
		PageSize int64  `form:"page_size"`
		From     int64  `form:"from"`
		To       int64  `form:"to"`
		Level    string `form:"level"`
		Msg      string `form:"msg"`
		Caller   string `form:"caller"`
		Server   string `form:"__source__"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		cosy.ErrHandler(c, err)
		return
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = int64(cSettings.AppSettings.PageSize)
	}

	queryExp := "*"
	filter := make([]string, 0)
	if query.Level != "" {
		filter = append(filter, fmt.Sprintf("level:%s", query.Level))
	}
	if query.Msg != "" {
		filter = append(filter, fmt.Sprintf("msg:%s*", query.Msg))
	}
	if query.Caller != "" {
		filter = append(filter, fmt.Sprintf("caller:%s*", query.Caller))
	}
	if query.Server != "" {
		filter = append(filter, fmt.Sprintf("__source__:%s*", query.Server))
	}
	if len(filter) > 0 {
		queryExp = strings.Join(filter, " and ")
	}

	s.Info("[SLS Query Exp]", queryExp)

	audit := NewAuditClient()

	var offset int64
	if query.Page == 1 {
		offset = 0
	} else {
		offset = (query.Page - 1) * query.PageSize
	}

	audit.SetQueryParams(cSettings.SLSSettings.DefaultLogStoreName, "", query.From, query.To, offset, query.PageSize, queryExp)

	histogramsResp, err := audit.GetHistograms()
	if err != nil {
		cosy.ErrHandler(c, err)
		return
	}

	logResp, err := audit.GetLogs(c)
	if err != nil {
		cosy.ErrHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, cModel.DataList{
		Data: logResp.Logs,
		Pagination: cModel.Pagination{
			Total:       histogramsResp.Count,
			TotalPages:  cModel.TotalPage(histogramsResp.Count, int(query.PageSize)),
			CurrentPage: int(query.Page),
			PerPage:     int(query.PageSize),
		},
	})
}
