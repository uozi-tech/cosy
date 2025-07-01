package audit

import (
	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uozi-tech/cosy/geoip"
	cSettings "github.com/uozi-tech/cosy/settings"
)

func (a *AuditClient) GetLogs(c *gin.Context) (resp *sls.GetLogsResponse, err error) {
	resp, err = a.client.GetLogs(cSettings.SLSSettings.ProjectName,
		a.logStoreName,
		a.topic,
		a.from,
		a.to,
		a.queryExp,
		a.pageSize,
		a.offset,
		true)

	if err != nil {
		return
	}

	for _, log := range resp.Logs {
		// backward compatibility, old logs don't have request_id
		// so we generate a new one for them temporarily
		uuidStr := uuid.New().String()
		if log["id"] == "" {
			log["id"] = uuidStr
		}

		if log["request_id"] == "" {
			log["request_id"] = uuidStr
		}

		// geoip
		if ip, ok := log["ip"]; ok {
			log["geoip"] = geoip.ParseIP(ip)
		}
	}

	if a.logsHandler != nil {
		a.logsHandler(resp.Logs)
	}

	return
}
