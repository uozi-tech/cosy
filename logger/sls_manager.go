package logger

import (
	"fmt"
	"log"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/uozi-tech/cosy/settings"
)

// SLSManager handles LogStore and index management
type SLSManager struct {
	client sls.ClientInterface
}

// NewSLSManager creates a new SLS manager instance
func NewSLSManager() (*SLSManager, error) {
	slsSettings := settings.SLSSettings
	if !slsSettings.Enable() {
		return nil, fmt.Errorf("SLS settings not enabled")
	}

	provider := slsSettings.GetCredentialsProvider()
	client := sls.CreateNormalInterfaceV2(slsSettings.EndPoint, provider)

	return &SLSManager{
		client: client,
	}, nil
}

// EnsureLogStore checks if LogStore exists, creates it if not
func (m *SLSManager) EnsureLogStore(projectName, logStoreName string) error {
	// Check if LogStore exists
	_, err := m.client.GetLogStore(projectName, logStoreName)
	if err == nil {
		// LogStore exists
		return nil
	}

	// Check if error is "LogStoreNotExist"
	if slsErr, ok := err.(*sls.Error); ok && slsErr.Code == "LogStoreNotExist" {
		// LogStore doesn't exist, create it
		return m.createLogStore(projectName, logStoreName)
	}

	// Other error occurred
	return fmt.Errorf("failed to check LogStore existence: %w", err)
}

// createLogStore creates a new LogStore with default settings
func (m *SLSManager) createLogStore(projectName, logStoreName string) error {
	err := m.client.CreateLogStore(
		projectName,
		logStoreName,
		180,  // TTL: 180 days retention
		2,    // ShardCount: 2 shards
		true, // AutoSplit: enable auto split
		64,   // MaxSplitShard: max 64 shards
	)
	if err != nil {
		return fmt.Errorf("failed to create LogStore %s: %w", logStoreName, err)
	}

	return nil
}

// EnsureLogStoreIndex checks if index exists for LogStore, creates it if not, or updates it if the configuration differs
func (m *SLSManager) EnsureLogStoreIndex(projectName, logStoreName string) error {
	// Get expected index configuration
	expectedIndex := m.getExpectedIndex(logStoreName)

	// Check if index exists
	existingIndex, err := m.client.GetIndex(projectName, logStoreName)
	if err != nil {
		// Check if error is "IndexConfigNotExist"
		if slsErr, ok := err.(*sls.Error); ok && slsErr.Code == "IndexConfigNotExist" {
			// Index doesn't exist, create it
			return m.createLogStoreIndex(projectName, logStoreName)
		}
		// Other error occurred
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	// Index exists, check if it needs to be updated
	if m.needsIndexUpdate(existingIndex, expectedIndex) {
		err = m.client.UpdateIndex(projectName, logStoreName, *expectedIndex)
		if err != nil {
			return fmt.Errorf("failed to update index for LogStore %s: %w", logStoreName, err)
		}
		log.Printf("Updated index for LogStore %s with new fields", logStoreName)
	}

	return nil
}

// getExpectedIndex returns the expected index configuration for the given LogStore
func (m *SLSManager) getExpectedIndex(logStoreName string) *sls.Index {
	slsSettings := settings.SLSSettings

	// Create different indexes based on LogStore type
	switch logStoreName {
	case slsSettings.APILogStoreName:
		return m.createAPILogStoreIndex()
	case slsSettings.DefaultLogStoreName, "":
		return m.createDefaultLogStoreIndex()
	default:
		// For custom LogStores, use default application log index
		return m.createDefaultLogStoreIndex()
	}
}

// needsIndexUpdate compares existing and expected index configurations
func (m *SLSManager) needsIndexUpdate(existing *sls.Index, expected *sls.Index) bool {
	if existing == nil || expected == nil {
		return true
	}

	// Compare the number of keys
	if len(existing.Keys) != len(expected.Keys) {
		return true
	}

	// Check if all expected keys exist in the existing index
	for expectedKey, expectedConfig := range expected.Keys {
		existingConfig, exists := existing.Keys[expectedKey]
		if !exists {
			// Key doesn't exist in current index
			return true
		}

		// Compare key configurations
		if !m.compareIndexKey(existingConfig, expectedConfig) {
			return true
		}
	}

	// Compare line configurations if both exist
	if existing.Line != nil && expected.Line != nil {
		if !m.compareIndexLine(*existing.Line, *expected.Line) {
			return true
		}
	} else if (existing.Line == nil) != (expected.Line == nil) {
		// One has line config, the other doesn't
		return true
	}

	return false
}

// compareIndexKey compares two IndexKey configurations
func (m *SLSManager) compareIndexKey(existing, expected sls.IndexKey) bool {
	// Compare type
	if existing.Type != expected.Type {
		return false
	}

	// Compare case sensitivity
	if existing.CaseSensitive != expected.CaseSensitive {
		return false
	}

	// Compare tokens
	if len(existing.Token) != len(expected.Token) {
		return false
	}

	// Create maps for token comparison (order doesn't matter)
	existingTokens := make(map[string]bool)
	for _, token := range existing.Token {
		existingTokens[token] = true
	}

	for _, token := range expected.Token {
		if !existingTokens[token] {
			return false
		}
	}

	return true
}

// compareIndexLine compares two IndexLine configurations
func (m *SLSManager) compareIndexLine(existing, expected sls.IndexLine) bool {
	// Compare case sensitivity
	if existing.CaseSensitive != expected.CaseSensitive {
		return false
	}

	// Compare tokens
	if len(existing.Token) != len(expected.Token) {
		return false
	}

	// Create maps for token comparison
	existingTokens := make(map[string]bool)
	for _, token := range existing.Token {
		existingTokens[token] = true
	}

	for _, token := range expected.Token {
		if !existingTokens[token] {
			return false
		}
	}

	// Compare include keys
	if len(existing.IncludeKeys) != len(expected.IncludeKeys) {
		return false
	}

	existingInclude := make(map[string]bool)
	for _, key := range existing.IncludeKeys {
		existingInclude[key] = true
	}

	for _, key := range expected.IncludeKeys {
		if !existingInclude[key] {
			return false
		}
	}

	// Compare exclude keys
	if len(existing.ExcludeKeys) != len(expected.ExcludeKeys) {
		return false
	}

	existingExclude := make(map[string]bool)
	for _, key := range existing.ExcludeKeys {
		existingExclude[key] = true
	}

	for _, key := range expected.ExcludeKeys {
		if !existingExclude[key] {
			return false
		}
	}

	return true
}

// createLogStoreIndex creates index for LogStore with fields specific to its purpose
func (m *SLSManager) createLogStoreIndex(projectName, logStoreName string) error {
	slsSettings := settings.SLSSettings

	var index *sls.Index

	// Create different indexes based on LogStore type
	switch logStoreName {
	case slsSettings.APILogStoreName:
		index = m.createAPILogStoreIndex()
	case slsSettings.DefaultLogStoreName, "":
		index = m.createDefaultLogStoreIndex()
	default:
		// For custom LogStores, use default application log index
		index = m.createDefaultLogStoreIndex()
	}

	err := m.client.CreateIndex(projectName, logStoreName, *index)
	if err != nil {
		return fmt.Errorf("failed to create index for LogStore %s: %w", logStoreName, err)
	}

	return nil
}

// createAPILogStoreIndex creates index optimized for API request logs
func (m *SLSManager) createAPILogStoreIndex() *sls.Index {
	return &sls.Index{
		Keys: map[string]sls.IndexKey{
			// Request identification
			"request_id": {
				Token:         []string{"-"},
				CaseSensitive: false,
				Type:          "text",
			},
			// Network and client info
			"ip": {
				Token:         []string{".", ":"},
				CaseSensitive: false,
				Type:          "text",
			},
			// HTTP request info
			"req_method": {
				Token:         []string{" "},
				CaseSensitive: false,
				Type:          "text",
			},
			"req_url": {
				Token:         []string{"/", "?", "&", "="},
				CaseSensitive: false,
				Type:          "text",
			},
			// HTTP response info
			"resp_status_code": {
				Type: "long",
			},
			// Performance metrics
			"latency": {
				Token:         []string{" ", ".", "µ", "m", "s"},
				CaseSensitive: false,
				Type:          "text",
			},
			// WebSocket indicator
			"is_websocket": {
				Token:         []string{" ", "\t", "\r", "\n"},
				CaseSensitive: false,
				Type:          "text",
			},
			// Request/Response content (for search but not detailed analysis)
			"req_body": {
				Token:         []string{" ", "\t", "\r", "\n", ":", ",", "{", "}", "[", "]"},
				CaseSensitive: false,
				Type:          "text",
			},
			"resp_body": {
				Token:         []string{" ", "\t", "\r", "\n", ":", ",", "{", "}", "[", "]"},
				CaseSensitive: false,
				Type:          "text",
			},
			"session_logs": {
				Token:         []string{" ", "\t", "\r", "\n", ":", ",", "{", "}", "[", "]"},
				CaseSensitive: false,
				Type:          "text",
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

// createDefaultLogStoreIndex creates index optimized for application logs
func (m *SLSManager) createDefaultLogStoreIndex() *sls.Index {
	return &sls.Index{
		Keys: map[string]sls.IndexKey{
			// Log level
			"level": {
				Token:         []string{" ", "\t", "\r", "\n"},
				CaseSensitive: false,
				Type:          "text",
			},
			// Timestamp
			"time": {
				Type: "long",
			},
			// Log message
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
			// Source location
			"caller": {
				Token:         []string{" ", "\t", "\r", "\n", "/", ":"},
				CaseSensitive: false,
				Type:          "text",
			},
			// Logger name
			"logger": {
				Token:         []string{" ", "\t", "\r", "\n"},
				CaseSensitive: false,
				Type:          "text",
			},
			// Error information
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
			// Function and module info
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
			// Line number
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

	// Initialize API LogStore
	if err := m.EnsureLogStore(projectName, slsSettings.APILogStoreName); err != nil {
		return fmt.Errorf("failed to ensure API LogStore: %w", err)
	}

	// Create index for API LogStore
	if err := m.EnsureLogStoreIndex(projectName, slsSettings.APILogStoreName); err != nil {
		return fmt.Errorf("failed to ensure API LogStore index: %w", err)
	}

	// Initialize Default LogStore
	if err := m.EnsureLogStore(projectName, slsSettings.DefaultLogStoreName); err != nil {
		return fmt.Errorf("failed to ensure Default LogStore: %w", err)
	}

	// Create index for Default LogStore
	if err := m.EnsureLogStoreIndex(projectName, slsSettings.DefaultLogStoreName); err != nil {
		return fmt.Errorf("failed to ensure Default LogStore index: %w", err)
	}

	return nil
}

// WaitForLogStoreReady waits for LogStore to be ready after creation
func (m *SLSManager) WaitForLogStoreReady(projectName, logStoreName string, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		_, err := m.client.GetLogStore(projectName, logStoreName)
		if err == nil {
			return nil
		}

		if slsErr, ok := err.(*sls.Error); ok && slsErr.Code != "LogStoreNotExist" {
			return fmt.Errorf("failed to check LogStore readiness: %w", err)
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for LogStore %s to be ready", logStoreName)
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
