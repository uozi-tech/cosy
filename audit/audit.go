package audit

import (
	sls "github.com/aliyun/aliyun-log-go-sdk"
	cSettings "github.com/uozi-tech/cosy/settings"
)

type AuditClient struct {
	client       sls.ClientInterface
	logStoreName string
	topic        string
	from         int64
	to           int64
	offset       int64
	pageSize     int64
	queryExp     string
	logsHandler  func(logs []map[string]string)
}

func NewAuditClient() *AuditClient {
	endpoint := cSettings.SLSSettings.EndPoint
	provider := cSettings.SLSSettings.GetCredentialsProvider()

	client := sls.CreateNormalInterfaceV2(endpoint, provider)

	return &AuditClient{client: client}
}

func (a *AuditClient) SetQueryParams(logStoreName string, topic string, from int64, to int64, offset int64, pageSize int64, queryExp string) *AuditClient {
	a.logStoreName = logStoreName
	a.topic = topic
	a.from = from
	a.to = to
	a.offset = offset
	a.pageSize = pageSize
	a.queryExp = queryExp
	return a
}

func (a *AuditClient) SetLogsHandler(logsHandler func(logs []map[string]string)) *AuditClient {
	a.logsHandler = logsHandler
	return a
}
