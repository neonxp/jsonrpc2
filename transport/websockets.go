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
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline           = []byte{'\n'}
	space             = []byte{' '}
	errUpgradingConn  = errors.New("encountered error upgrading connection to websocket protocol")
	errStartingServer = errors.New("encountered error starting http server")
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WebSocket struct {
	Bind       string
	TLS        *tls.Config
	CORSOrigin string
	Parallel   bool
}

func (ws *WebSocket) Run(ctx context.Context, resolver Resolver) error {
	srv := http.Server{
		Addr: ws.Bind,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			wsconn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println(err)
				return
			}

			log.Println("successfully upgraded connection")

			defer func() {
				wsconn.Close()
			}()

			wsconn.SetReadLimit(maxMessageSize)
			wsconn.SetReadDeadline(time.Now().Add(pongWait))
			wsconn.SetPongHandler(func(string) error {
				wsconn.SetReadDeadline(time.Now().Add(pongWait))
				return nil
			})

			for {
				// read message from connection
				messageType, message, err := wsconn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("error: %v", err)
					}
					break
				}

				message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

				wsconn.SetWriteDeadline(time.Now().Add(writeWait))

				// create writer object that implements io.WriterCloser interface
				// messageType is same as the messageType recieved from the connection
				w, err := wsconn.NextWriter(messageType)
				if err != nil {
					return
				}

				resolver.Resolve(ctx, bytes.NewBuffer(message), w, ws.Parallel)
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
		log.Println(err)
		return err
	}
	return nil
}
