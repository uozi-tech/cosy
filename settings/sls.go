package settings

import (
	"github.com/uozi-tech/cosy/sls"
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

func (s *SLS) GetCredentials() sls.Credentials {
	return sls.Credentials{
		AccessKeyID:     s.AccessKeyId,
		AccessKeySecret: s.AccessKeySecret,
	}
}
