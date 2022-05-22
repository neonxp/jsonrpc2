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
	Logger              Logger
	IgnoreNotifications bool
	handlers            map[string]Handler
	transports          []transport.Transport
	mu                  sync.RWMutex
}

func New() *RpcServer {
	return &RpcServer{
		Logger:              nopLogger{},
		IgnoreNotifications: true,
		handlers:            map[string]Handler{},
		transports:          []transport.Transport{},
		mu:                  sync.RWMutex{},
	}
}

func (r *RpcServer) Register(method string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[method] = handler
}

func (r *RpcServer) AddTransport(transport transport.Transport) {
	r.transports = append(r.transports, transport)
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

func (r *RpcServer) Resolve(ctx context.Context, rd io.Reader, w io.Writer) {
	dec := json.NewDecoder(rd)
	enc := json.NewEncoder(w)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	for {
		req := new(rpcRequest)
		if err := dec.Decode(req); err != nil {
			if err == io.EOF {
				break
			}
			r.Logger.Logf("Can't read body: %v", err)
			WriteError(ErrCodeParseError, enc)
			break
		}
		wg.Add(1)
		go func(req *rpcRequest) {
			defer wg.Done()
			resp := r.callMethod(ctx, req)
			if req.Id == nil {
				// notification request
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if err := enc.Encode(resp); err != nil {
				r.Logger.Logf("Can't write response: %v", err)
				WriteError(ErrCodeInternalError, enc)
			}
			if w, canFlush := w.(Flusher); canFlush {
				w.Flush()
			}
		}(req)
	}
	wg.Wait()
}

func (r *RpcServer) callMethod(ctx context.Context, req *rpcRequest) *rpcResponse {
	r.mu.RLock()
	h, ok := r.handlers[req.Method]
	r.mu.RUnlock()
	if !ok {
		return &rpcResponse{
			Jsonrpc: version,
			Error:   ErrorFromCode(ErrCodeMethodNotFound),
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

func WriteError(code int, enc *json.Encoder) {
	enc.Encode(rpcResponse{
		Jsonrpc: version,
		Error:   ErrorFromCode(code),
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

type Flusher interface {
	// Flush sends any buffered data to the client.
	Flush()
}
