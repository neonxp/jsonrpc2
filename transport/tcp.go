package transport

import (
	"context"
	"net"
)

type TCP struct {
	Bind string
}

func (t *TCP) Run(ctx context.Context, resolver Resolver) error {
	ln, _ := net.Listen("tcp", t.Bind)

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go resolver.Resolve(ctx, conn, conn)
	}
}
