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
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/salemzii/jsonrpc2/transport"
)

const version = "2.0"

type RpcServer struct {
	logger      Logger
	handlers    map[string]HandlerFunc
	middlewares []Middleware
	transports  []transport.Transport
	mu          sync.RWMutex
}

func New(opts ...Option) *RpcServer {
	s := &RpcServer{
		logger:     nopLogger{},
		handlers:   map[string]HandlerFunc{},
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

func (r *RpcServer) Register(method string, handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Logf("Register method %s", method)
	r.handlers[strings.ToLower(method)] = handler
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
				enc.Encode(ErrorResponse(req.Id, ErrorFromCode(ErrCodeInternalError)))
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
	h, ok := r.handlers[strings.ToLower(req.Method)]
	r.mu.RUnlock()
	if !ok {
		return ErrorResponse(req.Id, ErrorFromCode(ErrCodeMethodNotFound))
	}
	resp, err := h(ctx, req.Params)
	if err != nil {
		r.logger.Logf("User error %v", err)
		return ErrorResponse(req.Id, err)
	}

	return ResultResponse(req.Id, resp)
}

func ResultResponse(id any, resp json.RawMessage) *RpcResponse {
	return &RpcResponse{
		Jsonrpc: version,
		Result:  resp,
		Id:      id,
	}
}

func ErrorResponse(id any, err error) *RpcResponse {
	return &RpcResponse{
		Jsonrpc: version,
		Error:   err,
		Id:      id,
	}
}
