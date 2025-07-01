package settings

import (
	sls "github.com/aliyun/aliyun-log-go-sdk"
)

type SLS struct {
	AccessKeyId         string
	AccessKeySecret     string
	EndPoint            string
	ProjectName         string
	APILogStoreName     string
	DefaultLogStoreName string
	Source              string
}

var SLSSettings = &SLS{}

func (s *SLS) Enable() bool {
	return s.AccessKeyId != "" &&
		s.AccessKeySecret != "" &&
		s.EndPoint != "" &&
		s.ProjectName != "" &&
		s.APILogStoreName != "" &&
		s.DefaultLogStoreName != "" &&
		s.Source != ""
}

func (s *SLS) GetCredentialsProvider() *sls.StaticCredentialsProvider {
	return sls.NewStaticCredentialsProvider(s.AccessKeyId, s.AccessKeySecret, "")
}
