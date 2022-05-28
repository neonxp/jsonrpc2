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
	"go.neonxp.dev/jsonrpc2/transport"
)

type Option func(s *RpcServer)

func WithTransport(transport transport.Transport) Option {
	return func(s *RpcServer) {
		s.transports = append(s.transports, transport)
	}
}

func WithMiddleware(mw Middleware) Option {
	return func(s *RpcServer) {
		s.middlewares = append(s.middlewares, mw)
	}
}

func WithLogger(l Logger) Option {
	return func(s *RpcServer) {
		s.logger = l
	}
}
