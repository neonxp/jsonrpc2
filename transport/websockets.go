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

package transport

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	websocket "github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type WebSocket struct {
	Bind                        string
	TLS                         *tls.Config
	CORSOrigin                  string
	Parallel                    bool
	ReadDeadline, WriteDeadline time.Duration //Set custom timeout for future read and write calls
}

func (ws *WebSocket) WithReadDealine() bool  { return ws.ReadDeadline != 0 }
func (ws *WebSocket) WithWriteDealine() bool { return ws.WriteDeadline != 0 }

func (ws *WebSocket) Run(ctx context.Context, resolver Resolver) error {
	srv := http.Server{
		Addr: ws.Bind,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsconn, _, _, err := websocket.UpgradeHTTP(r, w)
			if err != nil {
				return
			}

			defer func() {
				wsconn.Close()
			}()

			if ws.WithReadDealine() {
				wsconn.SetReadDeadline(time.Now().Add(ws.ReadDeadline * time.Second))
			}

			if ws.WithWriteDealine() {
				wsconn.SetWriteDeadline(time.Now().Add(ws.WriteDeadline * time.Second))
			}

			for {

				// read message from connection
				_, reader, err := wsutil.NextReader(wsconn, websocket.StateServerSide)
				if err != nil {
					return
				}

				// create writer object that implements io.WriterCloser interface
				writer := wsutil.NewWriter(wsconn, websocket.StateServerSide, websocket.OpText)

				resolver.Resolve(ctx, reader, writer, ws.Parallel)

				if err := writer.Flush(); err != nil {
					return
				}

			}

		}),

		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
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
