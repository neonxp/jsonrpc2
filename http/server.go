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
