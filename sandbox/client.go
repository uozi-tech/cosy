package sandbox

import (
	"bytes"
	"encoding/json"
	"github.com/uozi-tech/cosy/router"
	"net/http"
	"net/http/httptest"
)

type Client struct {
	Header map[string]string
}

// NewClient create a new client instance
func newClient() *Client {
	return &Client{
		Header: make(map[string]string),
	}
}

// Request send a request and get response
func (c *Client) Request(method string, uri string, body any) (r *Response, err error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return
	}

	var req *http.Request
	if body == nil {
		req = httptest.NewRequest(method, uri, nil)
	} else {
		req = httptest.NewRequest(method, uri, bytes.NewBuffer(buf))
	}
	c.attachHeader(req)

	w := httptest.NewRecorder()
	router.GetEngine().ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	return NewResponse(resp.StatusCode, resp.Body)
}
