package httpserver

import (
	"context"
	"net/http"

	"github.com/gopherd/core/component"
	"github.com/gopherd/example/components/httpserver/httpserverapi"
)

const name = "github.com/gopherd/example/components/httpserver"

type httpserverComponent struct {
	component.BaseComponent[struct {
		Addr string
	}]
	server *http.Server
}

// 断言 httpserverComponent 实现了接口 httpserverapi.Component
var _ httpserverapi.Component = (*httpserverComponent)(nil)

func init() {
	component.Register(name, func() component.Component {
		return new(httpserverComponent)
	})
}

func (h *httpserverComponent) Init(ctx context.Context) error {
	addr := h.Options().Addr
	if addr == "" {
		addr = ":http"
	}
	h.server = &http.Server{Addr: addr}
	return nil
}

func (h *httpserverComponent) Start(ctx context.Context) error {
	h.Logger().Info("Starting HTTP server", "addr", h.server.Addr)
	go func() {
		if err := h.server.ListenAndServe(); err != http.ErrServerClosed {
			h.Logger().Error("HTTP server error", "error", err)
		}
	}()
	return nil
}

func (h *httpserverComponent) Shutdown(ctx context.Context) error {
	h.Logger().Info("Shutting down HTTP server")
	return h.server.Shutdown(ctx)
}

func (h *httpserverComponent) HandleFunc(pattern string, handler http.HandlerFunc) {
	http.HandleFunc(pattern, handler)
}
