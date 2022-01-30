package jsonrpc2

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

const version = "2.0"

type Server struct {
	Logger              Logger
	IgnoreNotifications bool
	handlers            map[string]Handler
	mu                  sync.RWMutex
}

func (r *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	buf := bufio.NewReader(request.Body)
	defer request.Body.Close()
	firstByte, err := buf.Peek(1)
	if err != nil {
		r.Logger.Logf("Can't read body: %v", err)
		writeError(ErrCodeParseError, writer)
		return
	}
	if string(firstByte) == "[" {
		r.batchRequest(writer, request, buf)
		return
	}
	r.singleRequest(writer, request, buf)
}

func New() *Server {
	return &Server{
		Logger:              nopLogger{},
		IgnoreNotifications: true,
		handlers:            map[string]Handler{},
		mu:                  sync.RWMutex{},
	}
}

func (r *Server) Register(method string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[method] = handler
}

func (r *Server) singleRequest(writer http.ResponseWriter, request *http.Request, buf *bufio.Reader) {
	req := new(rpcRequest)
	if err := json.NewDecoder(buf).Decode(req); err != nil {
		r.Logger.Logf("Can't read body: %v", err)
		writeError(ErrCodeParseError, writer)
		return
	}
	resp := r.callMethod(request.Context(), req)
	if req.Id == nil && r.IgnoreNotifications {
		// notification request
		return
	}
	if err := json.NewEncoder(writer).Encode(resp); err != nil {
		r.Logger.Logf("Can't write response: %v", err)
		writeError(ErrCodeInternalError, writer)
		return
	}
}

func (r *Server) batchRequest(writer http.ResponseWriter, request *http.Request, buf *bufio.Reader) {
	var req []rpcRequest
	if err := json.NewDecoder(buf).Decode(&req); err != nil {
		r.Logger.Logf("Can't read body: %v", err)
		writeError(ErrCodeParseError, writer)
		return
	}
	var responses []*rpcResponse
	wg := sync.WaitGroup{}
	wg.Add(len(req))
	for _, j := range req {
		go func(req rpcRequest) {
			defer wg.Done()
			resp := r.callMethod(request.Context(), &req)
			if req.Id == nil && r.IgnoreNotifications {
				// notification request
				return
			}
			responses = append(responses, resp)
		}(j)
	}
	wg.Wait()
	if err := json.NewEncoder(writer).Encode(responses); err != nil {
		r.Logger.Logf("Can't write response: %v", err)
		writeError(ErrCodeInternalError, writer)
	}
}

func (r *Server) callMethod(ctx context.Context, req *rpcRequest) *rpcResponse {
	r.mu.RLock()
	h, ok := r.handlers[req.Method]
	r.mu.RUnlock()
	if !ok {
		return &rpcResponse{
			Jsonrpc: version,
			Error:   NewError(ErrCodeMethodNotFound),
			Id:      req.Id,
		}
	}
	resp, err := h(ctx, req.Params)
	if err != nil {
		r.Logger.Logf("User error %v", err)
		return &rpcResponse{
			Jsonrpc: version,
			Error:   err,
			Id:      req.Id,
		}
	}
	return &rpcResponse{
		Jsonrpc: version,
		Result:  resp,
		Id:      req.Id,
	}
}

func writeError(code int, w io.Writer) {
	_ = json.NewEncoder(w).Encode(rpcResponse{
		Jsonrpc: version,
		Error:   NewError(code),
	})
}

type rpcRequest struct {
	Jsonrpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	Id      any             `json:"id"`
}

type rpcResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   error           `json:"error,omitempty"`
	Id      any             `json:"id,omitempty"`
}
