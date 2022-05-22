//Package transport provides transports for rpc server
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

package transport

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
)

type HTTP struct {
	Bind       string
	TLS        *tls.Config
	CORSOrigin string
	Parallel   bool
}

func (h *HTTP) Run(ctx context.Context, resolver Resolver) error {
	srv := http.Server{
		Addr: h.Bind,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions && h.CORSOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", h.CORSOrigin)
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.WriteHeader(http.StatusOK)
				return
			}
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if h.CORSOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", h.CORSOrigin)
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			}
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			resolver.Resolve(ctx, r.Body, w, h.Parallel)
		}),
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
		TLSConfig: h.TLS,
	}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
