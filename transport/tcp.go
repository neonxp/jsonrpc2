//Package transport provides transports for rpc server
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

package transport

import (
	"context"
	"net"
)

type TCP struct {
	Bind     string
	Parallel bool
}

func (t *TCP) Run(ctx context.Context, resolver Resolver) error {
	ln, err := net.Listen("tcp", t.Bind)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		ln.Close()
	}()
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go resolver.Resolve(ctx, conn, conn, t.Parallel)
	}
}
