
# JSON-RPC 2.0

Golang implementation of JSON-RPC 2.0 server with generics.

Go 1.18+ required

## Features:

- [x] HTTP/HTTPS transport
- [x] TCP transport
- [ ] WebSocket transport

## Usage (http transport)

1. Create JSON-RPC server with options:
```go
    import "go.neonxp.dev/jsonrpc2/rpc"
    ...
    s := rpc.New(
        rpc.WithTransport(&transport.HTTP{
            Bind: ":8000",      // Port to bind
            CORSOrigin: "*",    // CORS origin
            TLS: &tls.Config{}, // Optional TLS config (default nil)
            Parallel: true,     // Allow parallel run batch methods (default false)
        }),
        //Other options like transports/middlewares...
    )
```

2. Add required transport(s):
```go
    import "go.neonxp.dev/jsonrpc2/transport"
    ...
    s.Use(
        rpc.WithTransport(&transport.TCP{Bind: ":3000"}),
        //...
    )
```

3. Write handler:
```go
    func Multiply(ctx context.Context, args *Args) (int, error) {
        return args.A * args.B, nil
    }
```

   Handler must have exact two arguments (context and input of any json serializable type) and exact two return values (output of any json serializable type and error)

4. Wrap handler with `rpc.H` method and register it in server:
```go
    s.Register("multiply", rpc.H(Multiply))
```

5. Run RPC server:
```go
    s.Run(ctx)
```

## Custom transport

Any transport must implement simple interface `transport.Transport`:

```go
type Transport interface {
	Run(ctx context.Context, resolver Resolver) error
}
```

## Complete example

[Full code](/example)

```go
package main

import (
   "context"

   "go.neonxp.dev/jsonrpc2/rpc"
   "go.neonxp.dev/jsonrpc2/transport"
)

func main() {
    s := rpc.New(
        rpc.WithLogger(rpc.StdLogger), // Optional logger
        rpc.WithTransport(&transport.HTTP{Bind: ":8000"}), // HTTP transport
    )

    // Set options after constructor
    s.Use(
        rpc.WithTransport(&transport.TCP{Bind: ":3000"}), // TCP transport
        rpc.WithMiddleware(rpc.LoggerMiddleware(rpc.StdLogger)), // Logger middleware
    )

   s.Register("multiply", rpc.H(Multiply))
   s.Register("divide", rpc.H(Divide))

   s.Run(context.Background())
}

func Multiply(ctx context.Context, args *Args) (int, error) {
    //...
}

func Divide(ctx context.Context, args *Args) (*Quotient, error) {
    //...
}

type Args struct {
	A int `json:"a"`
	B int `json:"b"`
}

type Quotient struct {
	Quo int `json:"quo"`
	Rem int `json:"rem"`
}

```

## Author

Alexander Kiryukhin <i@neonxp.dev>

## License

![GPL v3](https://www.gnu.org/graphics/gplv3-with-text-136x68.png)


