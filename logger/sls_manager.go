package logger

import (
	"fmt"
	"log"

	"github.com/uozi-tech/cosy/settings"
	"github.com/uozi-tech/cosy/sls"
)

// SLSManager handles LogStore and index management
type SLSManager struct {
	client *sls.Client
}

// NewSLSManager creates a new SLS manager instance
func NewSLSManager() (*SLSManager, error) {
	slsSettings := settings.SLSSettings
	if !slsSettings.Enable() {
		return nil, fmt.Errorf("SLS settings not enabled")
	}

	client := sls.NewClient(slsSettings.EndPoint, slsSettings.GetCredentials())
	return &SLSManager{client: client}, nil
}

// EnsureLogStore checks if LogStore exists, creates it if not
func (m *SLSManager) EnsureLogStore(projectName, logStoreName string) error {
	err := m.client.GetLogStore(projectName, logStoreName)
	if err == nil {
		return nil
	}
	if slsErr, ok := err.(*sls.Error); ok && slsErr.Code == "LogStoreNotExist" {
		return m.createLogStore(projectName, logStoreName)
	}
	return fmt.Errorf("failed to check LogStore existence: %w", err)
}

// createLogStore creates a new LogStore with default settings
func (m *SLSManager) createLogStore(projectName, logStoreName string) error {
	err := m.client.CreateLogStore(projectName, logStoreName, 180, 2, true, 64)
	if err != nil {
		return fmt.Errorf("failed to create LogStore %s: %w", logStoreName, err)
	}
	return nil
}

// EnsureLogStoreIndex checks if index exists for LogStore, creates it if not, or updates it if the configuration differs
func (m *SLSManager) EnsureLogStoreIndex(projectName, logStoreName string) error {
	expectedIndex := m.indexForLogStore(logStoreName)

	existingIndex, err := m.client.GetIndex(projectName, logStoreName)
	if err != nil {
		if slsErr, ok := err.(*sls.Error); ok && slsErr.Code == "IndexConfigNotExist" {
			return m.createLogStoreIndex(projectName, logStoreName)
		}
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if m.needsIndexUpdate(existingIndex, expectedIndex) {
		err = m.client.UpdateIndex(projectName, logStoreName, *expectedIndex)
		if err != nil {
			return fmt.Errorf("failed to update index for LogStore %s: %w", logStoreName, err)
		}
		log.Printf("Updated index for LogStore %s with new fields", logStoreName)
	}
	return nil
}

func (m *SLSManager) indexForLogStore(logStoreName string) *sls.Index {
	if logStoreName == settings.SLSSettings.APILogStoreName {
		return m.createAPILogStoreIndex()
	}
	return m.createDefaultLogStoreIndex()
}

// needsIndexUpdate compares existing and expected index configurations
func (m *SLSManager) needsIndexUpdate(existing *sls.Index, expected *sls.Index) bool {
	if existing == nil || expected == nil {
		return true
	}
	if len(existing.Keys) != len(expected.Keys) {
		return true
	}
	for expectedKey, expectedConfig := range expected.Keys {
		existingConfig, exists := existing.Keys[expectedKey]
		if !exists {
			return true
		}
		if !m.compareIndexKey(existingConfig, expectedConfig) {
			return true
		}
	}
	if existing.Line != nil && expected.Line != nil {
		if !m.compareIndexLine(*existing.Line, *expected.Line) {
			return true
		}
	} else if (existing.Line == nil) != (expected.Line == nil) {
		return true
	}
	return false
}

func stringSetEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[string]struct{}, len(a))
	for _, s := range a {
		set[s] = struct{}{}
	}
	for _, s := range b {
		if _, ok := set[s]; !ok {
			return false
		}
	}
	return true
}

func (m *SLSManager) compareIndexKey(existing, expected sls.IndexKey) bool {
	return existing.Type == expected.Type &&
		existing.CaseSensitive == expected.CaseSensitive &&
		stringSetEqual(existing.Token, expected.Token)
}

func (m *SLSManager) compareIndexLine(existing, expected sls.IndexLine) bool {
	return existing.CaseSensitive == expected.CaseSensitive &&
		stringSetEqual(existing.Token, expected.Token) &&
		stringSetEqual(existing.IncludeKeys, expected.IncludeKeys) &&
		stringSetEqual(existing.ExcludeKeys, expected.ExcludeKeys)
}

func (m *SLSManager) createLogStoreIndex(projectName, logStoreName string) error {
	idx := m.indexForLogStore(logStoreName)
	if err := m.client.CreateIndex(projectName, logStoreName, *idx); err != nil {
		return fmt.Errorf("failed to create index for LogStore %s: %w", logStoreName, err)
	}
	return nil
}

// createAPILogStoreIndex creates index optimized for API request logs
func (m *SLSManager) createAPILogStoreIndex() *sls.Index {
	return &sls.Index{
		Keys: map[string]sls.IndexKey{
			"request_id": {
				Token:         []string{"-"},
				CaseSensitive: false,
				Type:          "text",
			},
			"ip": {
				Token:         []string{".", ":"},
				CaseSensitive: false,
				Type:          "text",
			},
			"req_method": {
				Token:         []string{" "},
				CaseSensitive: false,
				Type:          "text",
			},
			"req_url": {
				Token:         []string{"/", "?", "&", "="},
				CaseSensitive: false,
				Type:          "text",
				Chn:           true,
			},
			"resp_status_code": {
				Type: "long",
			},
			"latency": {
				Token:         []string{" ", ".", "µ", "m", "s"},
				CaseSensitive: false,
				Type:          "text",
			},
			"is_websocket": {
				Token:         []string{" ", "\t", "\r", "\n"},
				CaseSensitive: false,
				Type:          "text",
			},
			"req_body": {
				Token:         []string{" ", "\t", "\r", "\n", ":", ",", "{", "}", "[", "]"},
				CaseSensitive: false,
				Type:          "text",
				Chn:           true,
			},
			"resp_body": {
				Token:         []string{" ", "\t", "\r", "\n", ":", ",", "{", "}", "[", "]"},
				CaseSensitive: false,
				Type:          "text",
				Chn:           true,
			},
			"session_logs": {
				Token:         []string{" ", "\t", "\r", "\n", ":", ",", "{", "}", "[", "]"},
				CaseSensitive: false,
				Type:          "text",
				Chn:           true,
			},
			"__source__": {
				Token:         []string{" ", "\t", "\r", "\n"},
				CaseSensitive: false,
				Type:          "text",
				Chn:           true,
			},
			"__tag__:type": {
				Token:         []string{" ", "\t", "\r", "\n"},
				CaseSensitive: false,
				Type:          "text",
				Chn:           true,
			},
			"__topic__": {
				Token:         []string{" ", "\t", "\r", "\n"},
				CaseSensitive: false,
				Type:          "text",
				Chn:           true,
			},
		},
		Line: &sls.IndexLine{
			Token:         []string{" ", "\t", "\r", "\n"},
			CaseSensitive: false,
			IncludeKeys:   []string{},
			ExcludeKeys:   []string{},
		},
	}
}

// createDefaultLogStoreIndex creates index optimized for application logs
func (m *SLSManager) createDefaultLogStoreIndex() *sls.Index {
	return &sls.Index{
		Keys: map[string]sls.IndexKey{
			"level": {
				Token:         []string{" ", "\t", "\r", "\n"},
				CaseSensitive: false,
				Type:          "text",
			},
			"time": {
				Type: "long",
			},
			"msg": {
				Token:         []string{" ", "\t", "\r", "\n"},
				CaseSensitive: false,
				Type:          "text",
			},
			"message": {
				Token:         []string{" ", "\t", "\r", "\n"},
				CaseSensitive: false,
				Type:          "text",
			},
			"caller": {
				Token:         []string{" ", "\t", "\r", "\n", "/", ":"},
				CaseSensitive: false,
				Type:          "text",
			},
			"logger": {
				Token:         []string{" ", "\t", "\r", "\n"},
				CaseSensitive: false,
				Type:          "text",
			},
			"error": {
				Token:         []string{" ", "\t", "\r", "\n", ":", ";"},
				CaseSensitive: false,
				Type:          "text",
			},
			"stacktrace": {
				Token:         []string{" ", "\t", "\r", "\n", "/", ":"},
				CaseSensitive: false,
				Type:          "text",
			},
			"func_name": {
				Token:         []string{" ", "\t", "\r", "\n", ".", "(", ")"},
				CaseSensitive: false,
				Type:          "text",
			},
			"module": {
				Token:         []string{" ", "\t", "\r", "\n", "/", "."},
				CaseSensitive: false,
				Type:          "text",
			},
			"line_no": {
				Type: "long",
			},
			"__source__": {
				Token:         []string{" ", "\t", "\r", "\n"},
				CaseSensitive: false,
				Type:          "text",
			},
			"__tag__:type": {
				Token:         []string{" ", "\t", "\r", "\n"},
				CaseSensitive: false,
				Type:          "text",
			},
			"__topic__": {
				Token:         []string{" ", "\t", "\r", "\n"},
				CaseSensitive: false,
				Type:          "text",
			},
		},
		Line: &sls.IndexLine{
			Token:         []string{" ", "\t", "\r", "\n"},
			CaseSensitive: false,
			IncludeKeys:   []string{},
			ExcludeKeys:   []string{},
		},
	}
}

// InitializeLogStores initializes all LogStores and their indexes
func (m *SLSManager) InitializeLogStores() error {
	slsSettings := settings.SLSSettings
	projectName := slsSettings.ProjectName

	if err := m.EnsureLogStore(projectName, slsSettings.APILogStoreName); err != nil {
		return fmt.Errorf("failed to ensure API LogStore: %w", err)
	}
	if err := m.EnsureLogStoreIndex(projectName, slsSettings.APILogStoreName); err != nil {
		return fmt.Errorf("failed to ensure API LogStore index: %w", err)
	}
	if err := m.EnsureLogStore(projectName, slsSettings.DefaultLogStoreName); err != nil {
		return fmt.Errorf("failed to ensure Default LogStore: %w", err)
	}
	if err := m.EnsureLogStoreIndex(projectName, slsSettings.DefaultLogStoreName); err != nil {
		return fmt.Errorf("failed to ensure Default LogStore index: %w", err)
	}
	return nil
}

// InitializeSLS initializes SLS LogStores and indexes
func InitializeSLS() error {
	manager, err := NewSLSManager()
	if err != nil {
		return fmt.Errorf("failed to create SLS manager: %w", err)
	}
	err = manager.InitializeLogStores()
	if err != nil {
		return fmt.Errorf("failed to initialize LogStores: %w", err)
	}
	log.Println("SLS LogStores and indexes initialized successfully")
	return nil
}
