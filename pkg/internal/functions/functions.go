package functions

import (
	jsonParser "encoding/json"
	"github.com/pkg/errors"
	"io"
)

// Invoke godoc
/* Invokes a function

functionName: the name of the function to invoke

Usage:
	Invoke("Hello-Function", &FunctionInvokeOptions{
		Body			io.Reader
		ResponseType	string
	})
*/
func (c *Client) Invoke(functionName string, options FunctionInvokeOptions) FunctionResponse {
	var responseType string
	if len(options.ResponseType) > 0 {
		responseType = options.ResponseType
	} else {
		responseType = Json
	}
	response, _ := c.session.Post(c.clientTransport.baseUrl.String()+"/"+functionName, responseType, options.Body)

	isRelayError := response.Header.Get("x-relay-error")
	if len(isRelayError) > 0 && isRelayError == "true" {
		return FunctionResponse{
			Error:  errors.Errorf("error: %s", response.Header.Get("x-relay-error-message")),
			Status: response.StatusCode,
		}
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return FunctionResponse{
			Error: err,
		}
	}

	defer response.Body.Close()

	var data interface{}
	switch responseType {
	case Json:
		logger.Debug("response", "body", string(body))
		err = jsonParser.Unmarshal(body, &data)
		if err != nil {
			return FunctionResponse{
				Error: err,
			}
		}
	case ArrayBuffer, Blob:
		data = body
	case Text:
		data = string(body)
	default:
		return FunctionResponse{
			Error: errors.Errorf("invalid response type: %s", responseType),
		}
	}

	return FunctionResponse{
		Data:   data,
		Status: response.StatusCode,
	}
}
