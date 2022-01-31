//Package rpc provides abstract rpc server
//
//Copyright (C) 2022 Alexander Kiryukhin <i@neonxp.dev>
//
//This file is part of github.com/neonxp/jsonrpc2 project.
//
//This program is free software: you can redistribute it and/or modify
//it under the terms of the GNU General Public License as published by
//the Free Software Foundation, either version 3 of the License, or
//(at your option) any later version.
//
//This program is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//GNU General Public License for more details.
//
//You should have received a copy of the GNU General Public License
//along with this program.  If not, see <https://www.gnu.org/licenses/>.

package rpc

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

//-32000 to -32099 	RpcServer error 	Reserved for implementation-defined server-errors.

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
