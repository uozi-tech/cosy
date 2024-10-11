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
    defer body.Close()
    return &Response{
        StatusCode: statusCode,
        body:       b,
    }, nil
}

// To decode the body to the given dest
func (r *Response) To(dest any) error {
    return json.Unmarshal(r.body, dest)
}

// BodyText return the body in string
func (r *Response) BodyText() string {
    return string(r.body)
}
