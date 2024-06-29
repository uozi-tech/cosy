package sandbox

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Client return a sandbox client
func (t *Instance) Client() *Client {
	return &Client{
		Header: make(map[string]string),
	}
}

// AddHeader add header
func (c *Client) AddHeader(key, value string) {
	c.Header[key] = value
}

// attachHeader attach header to the given request
func (c *Client) attachHeader(req *http.Response) {
	for k, v := range c.Header {
		req.Header.Set(k, v)
	}
}

// Get send a get request
func (c *Client) Get(uri string) (r *Response, err error) {
	return c.Request(http.MethodGet, uri, nil)
}

// Post send a post request
func (c *Client) Post(uri string, body gin.H) (r *Response, err error) {
	return c.Request(http.MethodPost, uri, body)
}

// Put send a put request
func (c *Client) Put(uri string, body gin.H) (r *Response, err error) {
	return c.Request(http.MethodPut, uri, body)
}

// Patch send a patch request
func (c *Client) Patch(uri string, body gin.H) (r *Response, err error) {
	return c.Request(http.MethodPatch, uri, body)
}

// Delete send a delete request
func (c *Client) Delete(uri string, body gin.H) (r *Response, err error) {
	return c.Request(http.MethodDelete, uri, body)
}

// Option send a option request
func (c *Client) Option(uri string, body gin.H) (r *Response, err error) {
	return c.Request(http.MethodOptions, uri, body)
}
