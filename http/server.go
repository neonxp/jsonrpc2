//Package http provides HTTP transport for JSON-RPC 2.0 server
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

package http

import (
	"bufio"
	"net/http"

	"github.com/neonxp/jsonrpc2/rpc"
)

type Server struct {
	*rpc.RpcServer
}

func New() *Server {
	return &Server{RpcServer: rpc.New()}
}

func (r *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	reader := bufio.NewReader(request.Body)
	defer request.Body.Close()
	firstByte, err := reader.Peek(1)
	if err != nil {
		r.Logger.Logf("Can't read body: %v", err)
		rpc.WriteError(rpc.ErrCodeParseError, writer)
		return
	}
	if string(firstByte) == "[" {
		r.BatchRequest(request.Context(), reader, writer)
		return
	}
	r.SingleRequest(request.Context(), reader, writer)
}
