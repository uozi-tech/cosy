package settings

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSLSGetSourceUsesHostnameWithoutMutatingConfig(t *testing.T) {
	s := &SLS{}

	source := s.GetSource()
	hostname, err := os.Hostname()
	if err == nil {
		assert.Equal(t, hostname, source)
	}
	assert.Empty(t, s.Source)
}

func TestSLSGetSourcePrefersConfiguredSource(t *testing.T) {
	s := &SLS{Source: "configured-source"}

	assert.Equal(t, "configured-source", s.GetSource())
}
