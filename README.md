# JSON-RPC 2.0

Golang implementation of JSON-RPC 2.0 server with generics.

Go 1.18+ required

## Features:

- [x] Batch request and responses
- [ ] WebSockets 

## Usage

1. Create JSON-RPC 2.0 server:
    ```go
    s := jsonrpc2.New()
    ```
2. Write handler:
    ```go
    func Multiply(ctx context.Context, args *Args) (int, error) {
        return args.A * args.B, nil
    }
    ```
   Handler must have exact two arguments (context and input of any json serializable type) and exact two return values (output of any json serializable type and error)
3. Wrap handler with `jsonrpc2.Wrap` method and register it in server:
    ```go
    s.Register("multiply", jsonrpc2.Wrap(Multiply))
    ```
4. Use server as common http handler:
    ```go
    http.ListenAndServe(":8000", s)
    ```

## Complete example

[Full code](/examples/http)

```go
package main

import (
	"context"
	"net/http"

	"github.com/neonxp/rpc"
)

func main() {
	s := jsonrpc2.New()
	s.Register("multiply", jsonrpc2.Wrap(Multiply)) // Register handlers
	s.Register("divide", jsonrpc2.Wrap(Divide))

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