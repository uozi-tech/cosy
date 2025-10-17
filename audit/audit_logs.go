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

// GetAuditLogs retrieves audit logs
func GetAuditLogs(c *gin.Context, logsHandler func(logs []map[string]string)) {
	if !cSettings.SLSSettings.Enable() {
		c.JSON(http.StatusOK, cModel.DataList{})
		return
	}
	s := logger.NewSessionLogger(c)
	var query struct {
		Page           int64  `form:"page"`
		PageSize       int64  `form:"page_size"`
		From           int64  `form:"from"`
		To             int64  `form:"to"`
		IP             string `form:"ip"`
		ReqMethod      string `form:"req_method"`
		ReqUrl         string `form:"req_url"`
		RespStatusCode string `form:"resp_status_code"`
		UserID         string `form:"user_id"`
		Server         string `form:"__source__"`
		SessionContent string `form:"session_content"`
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
	if query.IP != "" {
		if fieldQuery := BuildFieldQuery(query.IP, "ip"); fieldQuery != "" {
			filter = append(filter, fieldQuery)
		}
	}
	if query.ReqMethod != "" {
		filter = append(filter, fmt.Sprintf("req_method:%s", query.ReqMethod))
	}
	if query.ReqUrl != "" {
		if fieldQuery := BuildFieldQuery(query.ReqUrl, "req_url"); fieldQuery != "" {
			filter = append(filter, fieldQuery)
		}
	}
	if query.RespStatusCode != "" {
		filter = append(filter, fmt.Sprintf("resp_status_code:%s", query.RespStatusCode))
	}
	if query.UserID != "" {
		filter = append(filter, fmt.Sprintf("user_id:%s", query.UserID))
	}
	if query.Server != "" {
		if fieldQuery := BuildFieldQuery(query.Server, "__source__"); fieldQuery != "" {
			filter = append(filter, fieldQuery)
		}
	}
	if query.SessionContent != "" {
		if fieldQuery := BuildFieldQuery(query.SessionContent, "session_logs"); fieldQuery != "" {
			filter = append(filter, fieldQuery)
		}
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

	audit.SetQueryParams(cSettings.SLSSettings.APILogStoreName, logger.Topic, query.From, query.To, offset, query.PageSize, queryExp)

	audit.SetLogsHandler(logsHandler)

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
