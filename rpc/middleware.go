//Package rpc provides abstract rpc server
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

package rpc

import (
	"context"
	"strings"
	"time"
)

type Middleware func(handler RpcHandler) RpcHandler

type RpcHandler func(ctx context.Context, req *RpcRequest) *RpcResponse

func LoggerMiddleware(logger Logger) Middleware {
	return func(handler RpcHandler) RpcHandler {
		return func(ctx context.Context, req *RpcRequest) *RpcResponse {
			t1 := time.Now().UnixMicro()
			resp := handler(ctx, req)
			t2 := time.Now().UnixMicro()
			args := strings.ReplaceAll(string(req.Params), "\n", "")
			logger.Logf("rpc call=%s, args=%s, take=%dμs", req.Method, args, (t2 - t1))
			return resp
		}
	}
}
