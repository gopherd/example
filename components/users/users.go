package users

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gopherd/core/component"
	"github.com/gopherd/core/event"
	"github.com/gopherd/example/components/auth/authapi"
	"github.com/gopherd/example/components/httpserver/httpserverapi"
)

const name = "github.com/gopherd/example/components/users"

func init() {
	component.Register(name, func() component.Component {
		return new(usersComponent)
	})
}

type usersComponent struct {
	component.BaseComponentWithRefs[struct {
		MaxUsers int
	}, struct {
		HTTPServer  component.Reference[httpserverapi.Component]
		EventSystem component.Reference[event.Dispatcher[reflect.Type]]
	}]
	loggedInUsers map[string]bool
}

func (u *usersComponent) Init(ctx context.Context) error {
	u.loggedInUsers = make(map[string]bool)
	return nil
}

func (u *usersComponent) Start(ctx context.Context) error {
	u.Logger().Info("Starting Users component")
	u.Refs().HTTPServer.Component().HandleFunc("/profile", u.handleProfile)
	u.Refs().EventSystem.Component().AddListener(authapi.LoginEventListener(u.onLoginEvent))
	return nil
}

func (u *usersComponent) handleProfile(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	if u.loggedInUsers[username] {
		fmt.Fprintf(w, "Profile for user: %s", username)
	} else {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
	}
}

func (u *usersComponent) onLoginEvent(ctx context.Context, e *authapi.LoginEvent) error {
	u.Logger().Info("User logged in", "username", e.Username)
	u.loggedInUsers[e.Username] = true
	if len(u.loggedInUsers) > u.Options().MaxUsers {
		u.Logger().Warn("Warning: Too many users logged")
	}
	return nil
}
