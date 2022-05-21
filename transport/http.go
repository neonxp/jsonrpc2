package transport

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
)

type HTTP struct {
	Bind string
	TLS  *tls.Config
}

func (h *HTTP) Run(ctx context.Context, resolver Resolver) error {
	srv := http.Server{
		Addr: h.Bind,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			resolver.Resolve(ctx, r.Body, w)
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
