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
	"time"

	"go.neonxp.dev/jsonrpc2/rpc"
)

func Logger(logger rpc.Logger) rpc.Middleware {
	return func(handler rpc.RpcHandler) rpc.RpcHandler {
		return func(ctx context.Context, req *rpc.RpcRequest) *rpc.RpcResponse {
			t1 := time.Now().UnixMicro()
			resp := handler(ctx, req)
			t2 := time.Now().UnixMicro()
			var params any
			if err := json.Unmarshal(req.Params, &params); req.Params != nil && err != nil {
				params = fmt.Sprintf("<invalid body: %s>", err.Error())
			}
			if req.Params == nil {
				params = "<empty body>"
			}
			logger.Logf("rpc call=%s, args=%+v, take=%dμs", req.Method, params, (t2 - t1))
			return resp
		}
	}
}
