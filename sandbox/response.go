package sandbox

import (
	"encoding/json"
	"io"
)

type Response struct {
	StatusCode int
	body       []byte
}

// NewResponse create a new response instance, if fail to read body, return an error
func NewResponse(statusCode int, body io.ReadCloser) (r *Response, err error) {
	b, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return &Response{
		StatusCode: statusCode,
		body:       b,
	}, nil
}

// GetStringBody return the body as string
func (r *Response) GetStringBody() string {
	return string(r.body)
}

// To decode the body to the given dest
func (r *Response) To(dest any) error {
	return json.Unmarshal(r.body, dest)
}
