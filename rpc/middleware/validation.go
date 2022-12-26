//Package middleware provides middlewares for rpc server
//
//Copyright (C) 2022 Alexander Kiryukhin <i@neonxp.dev>
//
//This file is part of go.neonxp.dev/jsonrpc2 project.
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

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qri-io/jsonschema"

	"github.com/salemzii/jsonrpc2/rpc"
)

type ServiceSchema map[string]MethodSchema

func MustSchema(schema string) ServiceSchema {
	ss := new(ServiceSchema)
	if err := json.Unmarshal([]byte(schema), ss); err != nil {
		panic(err)
	}
	return *ss
}

type MethodSchema struct {
	Request  *jsonschema.Schema `json:"request"`
	Response *jsonschema.Schema `json:"response"`
}

func Validation(serviceSchema ServiceSchema) (rpc.Middleware, error) {
	return func(handler rpc.RpcHandler) rpc.RpcHandler {
		return func(ctx context.Context, req *rpc.RpcRequest) *rpc.RpcResponse {
			rs, hasSchema := serviceSchema[strings.ToLower(req.Method)]
			if hasSchema && rs.Request != nil {
				if errResp := formatError(ctx, req.Id, *rs.Request, req.Params); errResp != nil {
					return errResp
				}
			}
			resp := handler(ctx, req)
			if hasSchema && rs.Response != nil {
				if errResp := formatError(ctx, req.Id, *rs.Response, resp.Result); errResp != nil {
					return errResp
				}
			}
			return resp
		}
	}, nil
}

func formatError(ctx context.Context, requestId any, schema jsonschema.Schema, data json.RawMessage) *rpc.RpcResponse {
	errs, err := schema.ValidateBytes(ctx, data)
	if err != nil {
		return rpc.ErrorResponse(requestId, err)
	}
	if errs != nil && len(errs) > 0 {
		messages := []string{}
		for _, msg := range errs {
			messages = append(messages, fmt.Sprintf("%s: %s", msg.PropertyPath, msg.Message))
		}
		return rpc.ErrorResponse(requestId, rpc.Error{
			Code:    rpc.ErrCodeInvalidParams,
			Message: strings.Join(messages, "\n"),
		})
	}
	return nil
}
