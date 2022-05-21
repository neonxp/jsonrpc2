package transport

import (
	"context"
	"io"
)

type Transport interface {
	Run(ctx context.Context, resolver Resolver) error
}

type Resolver interface {
	Resolve(context.Context, io.Reader, io.Writer)
}
