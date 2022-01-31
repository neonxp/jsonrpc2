# JSON-RPC 2.0

Golang implementation of JSON-RPC 2.0 server with generics.

Go 1.18+ required

## Features:

- [x] Batch request and responses
- [ ] WebSocket transport

## Usage (http transport)

1. Create JSON-RPC/HTTP server:
    ```go
    import "github.com/neonxp/jsonrpc2/http"
    ...
    s := http.New()
    ```
2. Write handler:
    ```go
    func Multiply(ctx context.Context, args *Args) (int, error) {
        return args.A * args.B, nil
    }
    ```
   Handler must have exact two arguments (context and input of any json serializable type) and exact two return values (output of any json serializable type and error)
3. Wrap handler with `rpc.Wrap` method and register it in server:
    ```go
    s.Register("multiply", rpc.Wrap(Multiply))
    ```
4. Use server as common http handler:
    ```go
    http.ListenAndServe(":8000", s)
    ```

## Custom transport

See [http/server.go](/http/server.go) for example of transport implementation.

## Complete example

[Full code](/examples/http)

```go
package main

import (
   "context"
   "net/http"

   httpRPC "github.com/neonxp/jsonrpc2/http"
   "github.com/neonxp/jsonrpc2/rpc"
)

func main() {
   s := httpRPC.New()
   s.Register("multiply", rpc.Wrap(Multiply))
   s.Register("divide", rpc.Wrap(Divide))

   http.ListenAndServe(":8000", s)
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