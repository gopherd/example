package auth

import (
	"context"
	"net/http"
	"reflect"

	"github.com/gopherd/core/component"
	"github.com/gopherd/core/event"
	"github.com/gopherd/example/components/auth/authapi"
	"github.com/gopherd/example/components/httpserver/httpserverapi"
)

const name = "github.com/gopherd/example/components/auth"

var _ authapi.Component = (*authComponent)(nil)

func init() {
	component.Register(name, func() component.Component {
		return new(authComponent)
	})
}

type authComponent struct {
	component.BaseComponentWithRefs[struct {
		Secret string
	}, struct {
		HTTPServer  component.Reference[httpserverapi.Component]
		EventSystem component.Reference[event.Dispatcher[reflect.Type]]
	}]
}

func (a *authComponent) Start(ctx context.Context) error {
	a.Logger().Info("Starting Auth component")
	a.Refs().HTTPServer.Component().HandleFunc("/login", a.handleLogin)
	return nil
}

func (a *authComponent) handleLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	// 简单的认证逻辑，实际应用中应该更加安全
	if username != "" {
		a.Refs().EventSystem.Component().DispatchEvent(context.Background(), &authapi.LoginEvent{Username: username})
		w.Write([]byte("Login successful"))
	} else {
		http.Error(w, "Invalid username", http.StatusBadRequest)
	}
}
