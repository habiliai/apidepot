package functions

import "io"

const (
	Json        = "json"
	Text        = "text"
	ArrayBuffer = "arrayBuffer"
	Blob        = "blob"
)

type FunctionInvokeOptions struct {
	Body         io.Reader `json:"body,omitempty"`
	ResponseType string    `json:"responseType,omitempty"`
}

type FunctionResponse struct {
	Data   interface{}
	Error  error
	Status int
}
