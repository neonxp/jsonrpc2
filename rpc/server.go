//Package rpc provides abstract rpc server
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

package rpc

import (
	"context"
	"encoding/json"
	"io"
	"sync"

	"golang.org/x/sync/errgroup"

	"go.neonxp.dev/jsonrpc2/transport"
)

const version = "2.0"

type RpcServer struct {
	logger      Logger
	handlers    map[string]Handler
	middlewares []Middleware
	transports  []transport.Transport
	mu          sync.RWMutex
}

func New(opts ...Option) *RpcServer {
	s := &RpcServer{
		logger:     nopLogger{},
		handlers:   map[string]Handler{},
		transports: []transport.Transport{},
		mu:         sync.RWMutex{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (r *RpcServer) Use(opts ...Option) {
	for _, opt := range opts {
		opt(r)
	}
}

func (r *RpcServer) Register(method string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[method] = handler
}

func (r *RpcServer) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	for _, t := range r.transports {
		eg.Go(func(t transport.Transport) func() error {
			return func() error { return t.Run(ctx, r) }
		}(t))
	}
	return eg.Wait()
}

func (r *RpcServer) Resolve(ctx context.Context, rd io.Reader, w io.Writer, parallel bool) {
	dec := json.NewDecoder(rd)
	enc := json.NewEncoder(w)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	for {
		req := new(RpcRequest)
		if err := dec.Decode(req); err != nil {
			break
		}
		exec := func() {
			h := r.callMethod
			for _, m := range r.middlewares {
				h = m(h)
			}
			resp := h(ctx, req)
			if req.Id == nil {
				// notification request
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if err := enc.Encode(resp); err != nil {
				r.logger.Logf("Can't write response: %v", err)
				WriteError(ErrCodeInternalError, enc)
			}
			if w, canFlush := w.(Flusher); canFlush {
				w.Flush()
			}
		}
		if parallel {
			wg.Add(1)
			go func(req *RpcRequest) {
				defer wg.Done()
				exec()
			}(req)
		} else {
			exec()
		}
	}
	if parallel {
		wg.Wait()
	}
}

func (r *RpcServer) callMethod(ctx context.Context, req *RpcRequest) *RpcResponse {
	r.mu.RLock()
	h, ok := r.handlers[req.Method]
	r.mu.RUnlock()
	if !ok {
		return &RpcResponse{
			Jsonrpc: version,
			Error:   ErrorFromCode(ErrCodeMethodNotFound),
			Id:      req.Id,
		}
	}
	resp, err := h(ctx, req.Params)
	if err != nil {
		r.logger.Logf("User error %v", err)
		return &RpcResponse{
			Jsonrpc: version,
			Error:   err,
			Id:      req.Id,
		}
	}
	return &RpcResponse{
		Jsonrpc: version,
		Result:  resp,
		Id:      req.Id,
	}
}

func WriteError(code int, enc *json.Encoder) {
	enc.Encode(RpcResponse{
		Jsonrpc: version,
		Error:   ErrorFromCode(code),
	})
}

type RpcRequest struct {
	Jsonrpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	Id      any             `json:"id"`
}

type RpcResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   error           `json:"error,omitempty"`
	Id      any             `json:"id,omitempty"`
}

type Flusher interface {
	// Flush sends any buffered data to the client.
	Flush()
}
