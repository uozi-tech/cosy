package audit

import (
	sls "github.com/aliyun/aliyun-log-go-sdk"
	cSettings "github.com/uozi-tech/cosy/settings"
)

func (a *AuditClient) GetHistograms() (resp *sls.GetHistogramsResponse, err error) {
	resp, err = a.client.GetHistograms(
		cSettings.SLSSettings.ProjectName,
		a.logStoreName,
		a.topic,
		a.from,
		a.to,
		a.queryExp)
	return
}
