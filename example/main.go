package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"

	"go.neonxp.dev/jsonrpc2/rpc"
	"go.neonxp.dev/jsonrpc2/rpc/middleware"
	"go.neonxp.dev/jsonrpc2/transport"
)

func main() {
	s := rpc.New(
		rpc.WithLogger(rpc.StdLogger),
		rpc.WithTransport(&transport.HTTP{Bind: ":8000", CORSOrigin: "*"}),
	)

	// Set options after constructor
	serviceSchema := `
	{
		"divide": {
			"request": {
				"type": "object",
				"properties": {
					"a": {
						"type": "integer"
					},
					"b": {
						"type": "integer",
						"not":{"const":0}
					}
				},
				"required": ["a", "b"]
			},
			"response": {
				"type": "object",
				"properties": {
					"quo": {
						"type": "integer"
					},
					"rem": {
						"type": "integer"
					}
				},
				"required": ["quo", "rem"]
			}
		},
		"multiply": {
			"request": {
				"type": "object",
				"properties": {
					"a": {
						"type": "integer"
					},
					"b": {
						"type": "integer"
					}
				},
				"required": ["a", "b"]
			},
			"response": {
				"type": "integer"
			}
		}
	}`

	validation, err := middleware.Validation(middleware.MustSchema(serviceSchema))
	if err != nil {
		log.Fatal(err)
	}
	s.Use(
		rpc.WithTransport(&transport.TCP{Bind: ":3000"}),
		rpc.WithMiddleware(middleware.Logger(rpc.StdLogger)),
		rpc.WithMiddleware(validation),
	)

	s.Register("multiply", rpc.H(Multiply))
	s.Register("divide", rpc.H(Divide))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if err := s.Run(ctx); err != nil {
		log.Fatal(err)
	}
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
