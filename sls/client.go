package sls

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pierrec/lz4/v4"
)

// Client is an SLS HTTP API client for management and data operations.
type Client struct {
	endpoint   string
	creds      Credentials
	httpClient *http.Client
}

// NewClient creates a new SLS API client.
func NewClient(endpoint string, creds Credentials) *Client {
	endpoint = strings.TrimRight(endpoint, "/")
	return &Client{
		endpoint: endpoint,
		creds:    creds,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// baseURL returns the full scheme+host for a project.
func (c *Client) baseURL(project string) string {
	ep := c.endpoint
	if !strings.HasPrefix(ep, "http://") && !strings.HasPrefix(ep, "https://") {
		ep = "https://" + ep
	}
	// insert project as subdomain: https://project.cn-hangzhou.log.aliyuncs.com
	if project != "" {
		scheme, host := splitSchemeHost(ep)
		ep = scheme + project + "." + host
	}
	return ep
}

func splitSchemeHost(u string) (scheme, host string) {
	if idx := strings.Index(u, "://"); idx >= 0 {
		return u[:idx+3], u[idx+3:]
	}
	return "https://", u
}

// do executes a signed HTTP request and returns the response body.
// On non-200 status, an *Error is returned.
func (c *Client) do(method, project, uri string, headers map[string]string, body []byte) ([]byte, error) {
	base := c.baseURL(project)

	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers[headerBodyRawSize]; !ok {
		headers[headerBodyRawSize] = strconv.Itoa(len(body))
	}
	sign(method, uri, headers, body, c.creds)

	req, err := http.NewRequest(method, base+uri, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("sls: create request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sls: http request: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		slsErr := &Error{HTTPCode: int32(resp.StatusCode)}
		if json.Unmarshal(respBody, slsErr) != nil {
			slsErr.Code = "Unknown"
			slsErr.Message = string(respBody)
		}
		slsErr.RequestID = resp.Header.Get("x-log-requestid")
		return nil, slsErr
	}
	return respBody, nil
}

// ---------- Management APIs ----------

// GetLogStore checks if a LogStore exists. Returns nil on success.
func (c *Client) GetLogStore(project, logstore string) error {
	h := map[string]string{"x-log-bodyrawsize": "0"}
	_, err := c.do("GET", project, "/logstores/"+logstore, h, nil)
	return err
}

// CreateLogStore creates a new LogStore.
func (c *Client) CreateLogStore(project, logstore string, ttl, shardCount int, autoSplit bool, maxSplitShard int) error {
	payload := struct {
		Name          string `json:"logstoreName"`
		TTL           int    `json:"ttl"`
		ShardCount    int    `json:"shardCount"`
		AutoSplit     bool   `json:"autoSplit"`
		MaxSplitShard int    `json:"maxSplitShard"`
	}{logstore, ttl, shardCount, autoSplit, maxSplitShard}
	body, _ := json.Marshal(payload)

	h := map[string]string{
		"x-log-bodyrawsize": strconv.Itoa(len(body)),
		"Content-Type":      "application/json",
	}
	_, err := c.do("POST", project, "/logstores", h, body)
	return err
}

// GetIndex retrieves the index configuration for a LogStore.
func (c *Client) GetIndex(project, logstore string) (*Index, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
		"Content-Type":      "application/json",
	}
	data, err := c.do("GET", project, fmt.Sprintf("/logstores/%s/index", logstore), h, nil)
	if err != nil {
		return nil, err
	}
	idx := &Index{}
	if err := json.Unmarshal(data, idx); err != nil {
		return nil, fmt.Errorf("sls: decode index: %w", err)
	}
	return idx, nil
}

// CreateIndex creates an index for a LogStore.
func (c *Client) CreateIndex(project, logstore string, index Index) error {
	body, _ := json.Marshal(index)
	h := map[string]string{
		"x-log-bodyrawsize": strconv.Itoa(len(body)),
		"Content-Type":      "application/json",
	}
	_, err := c.do("POST", project, fmt.Sprintf("/logstores/%s/index", logstore), h, body)
	return err
}

// UpdateIndex updates the index configuration for a LogStore.
func (c *Client) UpdateIndex(project, logstore string, index Index) error {
	body, _ := json.Marshal(index)
	h := map[string]string{
		"x-log-bodyrawsize": strconv.Itoa(len(body)),
		"Content-Type":      "application/json",
	}
	_, err := c.do("PUT", project, fmt.Sprintf("/logstores/%s/index", logstore), h, body)
	return err
}

// ---------- Data API ----------

// PutLogs sends a LogGroup to SLS via load-balanced shards.
func (c *Client) PutLogs(project, logstore string, lg *LogGroup) error {
	raw := MarshalLogGroup(lg)
	if len(raw) == 0 {
		return nil
	}

	compressed, rawSize, err := compressLZ4(raw)
	if err != nil {
		return fmt.Errorf("sls: lz4 compress: %w", err)
	}

	h := map[string]string{
		"x-log-compresstype": "lz4",
		"x-log-bodyrawsize":  strconv.Itoa(rawSize),
		"Content-Type":       "application/x-protobuf",
	}
	_, err = c.do("POST", project, fmt.Sprintf("/logstores/%s/shards/lb", logstore), h, compressed)
	return err
}

func compressLZ4(data []byte) (compressed []byte, rawSize int, err error) {
	rawSize = len(data)
	out := make([]byte, lz4.CompressBlockBound(rawSize))
	var ht [1 << 16]int
	n, err := lz4.CompressBlock(data, out, ht[:])
	if err != nil {
		return nil, rawSize, err
	}
	if n == 0 {
		// incompressible: wrap as lz4 literal-only block
		n = copyLiteralOnly(data, out)
	}
	return out[:n], rawSize, nil
}

// copyLiteralOnly encodes src as an lz4 literal-only block (no matches).
func copyLiteralOnly(src, dst []byte) int {
	lLen := len(src)
	di := 0
	if lLen < 0xF {
		dst[di] = byte(lLen << 4)
	} else {
		dst[di] = 0xF0
		di++
		remaining := lLen - 0xF
		for remaining >= 0xFF {
			dst[di] = 0xFF
			di++
			remaining -= 0xFF
		}
		dst[di] = byte(remaining)
	}
	di++
	di += copy(dst[di:], src)
	return di
}
