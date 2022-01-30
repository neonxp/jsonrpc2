package jsonrpc2

import "fmt"

const (
	ErrCodeParseError     = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternalError  = -32603
	ErrUser               = -32000
)

var errorMap = map[int]string{
	-32700: "Parse error",      // Invalid JSON was received by the server. An error occurred on the server while parsing the JSON text.
	-32600: "Invalid Request",  // The JSON sent is not a valid Request object.
	-32601: "Method not found", // The method does not exist / is not available.
	-32602: "Invalid params",   // Invalid method parameter(s).
	-32603: "Internal error",   // Internal JSON-RPC error.
	-32000: "Other error",
}

//-32000 to -32099 	Server error 	Reserved for implementation-defined server-errors.

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return fmt.Sprintf("jsonrpc2 error: code: %d message: %s", e.Code, e.Message)
}

func NewError(code int) Error {
	if _, ok := errorMap[code]; ok {
		return Error{
			Code:    code,
			Message: errorMap[code],
		}
	}
	return Error{Code: code}
}
