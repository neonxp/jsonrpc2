package main

import (
	"context"
	"errors"
	"net/http"

	httpRPC "go.neonxp.dev/jsonrpc2/http"
	"go.neonxp.dev/jsonrpc2/rpc"
)

func main() {
	s := httpRPC.New()

	s.Register("multiply", rpc.Wrap(Multiply))
	s.Register("divide", rpc.Wrap(Divide))

	http.ListenAndServe(":8000", s)
}

func Multiply(ctx context.Context, args *Args) (int, error) {
	return args.A * args.B, nil
}

func Divide(ctx context.Context, args *Args) (*Quotient, error) {
	if args.B == 0 {
		return nil, errors.New("divide by zero")
	}
	quo := new(Quotient)
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	return quo, nil
}

type Args struct {
	A int `json:"a"`
	B int `json:"b"`
}

type Quotient struct {
	Quo int `json:"quo"`
	Rem int `json:"rem"`
}
