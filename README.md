# Gopherd/core: A Comprehensive Guide

## 1. Introduction

[Gopherd/core](https://github.com/gopherd/core) is a modern Go framework designed for building scalable, modular backend services. It leverages Go's generics to provide a type-safe, flexible, and efficient way to develop applications.

This guide will demonstrate the features and usage of Gopherd/core through a step-by-step web server project. Our example project will include functionalities such as an HTTP server, event system, authentication (auth), and user management.

### 1.1 Key Features

- Generic-based component system
- Flexible configuration management
- Lifecycle management
- Dependency injection
- Event system

## 2. Quick Start

### 2.1 Installation

Ensure you have Go version 1.21 or later. Then, install Gopherd/core in your project:

```bash
go get github.com/gopherd/core
```

### 2.2 Project Structure

Our example project structure is as follows:

```
example/
├── components
│   ├── auth
│   │   ├── auth.go
│   │   └── authapi
│   │       └── authapi.go
│   ├── blockexit
│   │   └── blockexit.go
│   ├── eventsystem
│   │   └── eventsystem.go
│   ├── httpserver
│   │   ├── httpserver.go
│   │   └── httpserverapi
│   │       └── httpserverapi.go
│   ├── logger
│   │   └── logger.go
│   └── users
│       └── users.go
├── config.json
└── main.go
```

### 2.3 Creating the Main Program

Let's start by creating a minimal main program. In `main.go`:

```go
package main

import (
	"github.com/gopherd/core/service"
)

func main() {
	service.Run()
}
```

This simple main program calls `service.Run()` to start the application. At this point, it won't do anything because we haven't registered any components.

Let's first implement a simple `blockexit` component, which keeps the program running and allows it to be closed using `Ctrl-C`.

Create `components/blockexit/blockexit.go`:

```go
package blockexit

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/gopherd/core/component"
)

const name = "github.com/gopherd/example/components/blockexit"

func init() {
	component.Register(name, func() component.Component {
		return new(blockExitComponent)
	})
}

type blockExitComponent struct {
	component.BaseComponent[struct{}]
}

func (b *blockExitComponent) Start(ctx context.Context) error {
	fmt.Println("Starting blockExitComponent")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("Received interrupt signal")
	return nil
}
```

Then modify `main.go`:

```go
package main

import (
	"github.com/gopherd/core/service"

	// Import the component, the init method of the component package will register the component
	_ "github.com/gopherd/example/components/blockexit"
)

func main() {
	service.Run()
}
```

The main function hasn't changed; it will always remain this way. We modified it by adding an import to include the implemented component. Now you can execute it in the command line:

```sh
echo '{"Components":[{"Name":"github.com/gopherd/example/components/blockexit"}]}' | go run main.go -
```

After running, you'll see the output `Starting blockExitComponent`, and the program won't exit. Use `Ctrl-C` to exit.

*Note*: Pay attention to the following:
* It's recommended to use the package name as the component name to avoid accidental name duplication.
* This run doesn't use a configuration file but reads configuration information through standard input, which only configures one blockexit component.
* We also support obtaining configurations from files and HTTP services. Use `-h` to view the usage help.

## 3. Basic Component Implementation

The previous component was too simple. Let's start by implementing a basic HTTP server component to understand the basic structure and registration process of components.

### 3.1 HTTP Server Component

First, let's create the API definition for the HTTP server component. In `components/httpserver/httpserverapi/httpserverapi.go`:

```go
package httpserverapi

import "net/http"

type Component interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
}
```

Then, we implement the HTTP server component. In `components/httpserver/httpserver.go`:

```go
package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gopherd/core/component"
	"github.com/gopherd/example/components/httpserver/httpserverapi"
)

const name = "github.com/gopherd/example/components/httpserver"

func init() {
	component.Register(name, func() component.Component {
		return new(httpserverComponent)
	})
}

// Assert that httpserverComponent implements the httpserverapi.Component interface
var _ httpserverapi.Component = (*httpserverComponent)(nil)

type httpserverComponent struct {
	component.BaseComponent[struct {
		Addr string
	}]
	server *http.Server
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
	fmt.Println("Starting HTTP server", "addr", h.server.Addr)
	go func() {
		if err := h.server.ListenAndServe(); err != http.ErrServerClosed {
			h.Logger().Error("HTTP server error", "error", err)
		}
	}()
	return nil
}

func (h *httpserverComponent) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down HTTP server")
	return h.server.Shutdown(ctx)
}

func (h *httpserverComponent) HandleFunc(pattern string, handler http.HandlerFunc) {
	http.HandleFunc(pattern, handler)
}
```

### 3.2 Basic Configuration

Create a basic `config.json` file:

```json
{
    "Components": [
        {
            "Name": "github.com/gopherd/example/components/httpserver",
            "UUID": "httpserver",
            "Options": {
                "Addr": ":8080"
            }
        },
        {
            "Name": "github.com/gopherd/example/components/blockexit"
        }
    ]
}
```

### 3.3 Updating the Main Program

Now, let's update `main.go` to import the HTTP server component:

```go
package main

import (
	"github.com/gopherd/core/service"

	// Import components, the init method of the component package will register the components
	_ "github.com/gopherd/example/components/blockexit"
	_ "github.com/gopherd/example/components/httpserver"
)

func main() {
	service.Run()
}
```

Now you can run your application:

```bash
go run main.go config.json
```

This will start a basic HTTP server listening on port 8080. After starting, you'll see output like this:

```
Starting HTTP server addr :8080
Starting blockExitComponent
```

After pressing `Ctrl-C`, the program will close and output:

```
Received interrupt signal
Shutting down HTTP server
```

Here we can basically explain:

**Component Development**: Embed a component.BaseComponent[T], where the generic parameter T is the component's configuration, accessed through the component's Options() in the code. Then you can choose to implement the component's lifecycle functions such as `Init`, `Start`, `Shutdown` (and `Uninit` which hasn't appeared here).

**Component Registration**: In the init function, component.Register is called to register a constructor function that creates our implemented component object based on the component's name. This function doesn't need to initialize any data for the component; for any component, just using new to create is sufficient. Finally, importing this package in main completes the registration.

**Component Configuration**: Registered components won't run by themselves; they need to be added to the Components array in the configuration file.

**Component Order**: The order in Components is the calling order of Init and Start functions, while Shutdown and Uninit are in reverse order. Based on our current configuration, the lifecycle function execution order of the two components is as follows:

```
blockexit.Init      -> httpserver.Init    ->
blockexit.Start     -> httpserver.Start   ->
httpserver.Shutdown -> blockexit.Shutdown ->
httpserver.Uninit   -> blockexit.Uninit   -> Program exits
```

## 4. Configuration Management and Template Features

Gopherd/core provides a flexible configuration management mechanism, including template features. Let's delve into how to use these functionalities.

### 4.1 Configuration File Structure

The configuration file is typically a JSON file containing the following main sections:

- `Context`: Global context, which can be used in templates
- `Components`: List of components, each containing a required `Name` and optional `UUID`, `Refs`, `Options`

### 4.2 Template Syntax and Usage

The configuration file supports Go's template syntax. You can use `{{.}}` to reference values in the context. For example:

```json
{
    "Context": {
        "Namespace": "example",
        "Name": "example-server",
        "ID": 1001,
        "R": {
            "HTTPServer": "httpserver"
        }
    },
    "Components": [
        {
            "Name": "github.com/gopherd/example/components/httpserver",
            "UUID": "{{.R.HTTPServer}}",
            "Options": {
                "Addr": ":{{add 8000 .ID}}"
            }
        },
        {
            "Name": "github.com/gopherd/example/components/blockexit"
        }
    ]
}
```

In this example:

- `{{.R.HTTPServer}}` will be replaced with "httpserver"
- `{{add 8000 .ID}}` will be calculated as 9001 (8000 + 1001)

To use templates, you need to add the `-T` parameter when running the program, indicating that templates are enabled. By default, they are not enabled.

### 4.3 Configure support for the simplest line comments

For configurations using the `JSON` format, we've taken into account that you might need to include some explanatory notes. Therefore, we support line comments that begin with `//` at the start of a line (allowing for leading whitespace). However, block comments using `/* ... */` are not supported.

This approach allows you to add comments to your JSON configuration files for better readability and maintenance, while still maintaining compatibility with standard JSON parsers when the comments are stripped out.

Examples of valid comments:

```jsonc
{
    // Valid comment
    // Still a valid comment
    "Context": {
        // Also a valid comment
        "ID": 1001
    },
    "Components": [
        // Still a valid comment
    ]
}
```

Examples of invalid comments:

```json
{
    /* Invalid comment */
    "Context": { // Invalid comment
        "ID": 1001, // Also an invalid comment
    },
    "Components": [
        /*
        Invalid comment
        */
    ]
}
```

### 4.4 How to Use TOML or YAML Format for Configuration

First, you need to modify the main function.

For the `TOML` format:

```go
func main() {
    service.Run(service.WithEncoder(toml.Marshal), service.WithDecoder(toml.Unmarshal))
}
```

For the `YAML` format:

```go
func main() {
    service.Run(service.WithEncoder(yaml.Marshal), service.WithDecoder(yaml.Unmarshal))
}
```

You can choose from several popular libraries for the `toml` and `yaml` packages. For example, the following libraries are widely supported:

* [github.com/BurntSushi/toml](https://github.com/BurntSushi/toml)
* [github.com/pelletier/go-toml/v2](https://github.com/pelletier/go-toml)
* [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml)
* [github.com/goccy/go-yaml](https://github.com/goccy/go-yaml)

*Note*: It's important to mention that the current support for `toml` and `yaml` is implemented through a conversion process. The configuration file is first parsed into a `map[string]any` using the provided Decoder, then encoded into JSON. The rest of the program will continue to use `json` thereafter.

**By passing encoder and decoder parameters as demonstrated above, you can support configuration in any arbitrary format.**

## 5. Implementing Core Components

Now, let's implement other core components, including EventSystem, Auth, and Users components.

### 5.1 Event System Component Implementation

First, let's implement the event system component.

```go
package eventsystem

import (
	"context"
	"reflect"

	"github.com/gopherd/core/component"
	"github.com/gopherd/core/event"
)

const name = "github.com/gopherd/example/components/eventsystem"

func init() {
	component.Register(name, func() component.Component {
		return new(eventsystemComponent)
	})
}

// We didn't define a separate eventsystemapi package, directly using event.Dispatcher as the component's exported interface
var _ event.Dispatcher[reflect.Type] = (*eventsystemComponent)(nil)

type eventsystemComponent struct {
	component.BaseComponent[struct {
		Ordered *bool
	}]
	event.Dispatcher[reflect.Type]
}

func (com *eventsystemComponent) Init(ctx context.Context) error {
	ordered := true
	if com.Options().Ordered != nil {
		ordered = *com.Options().Ordered
	}
	com.Dispatcher = event.NewDispatcher[reflect.Type](ordered)
	return nil
}
```

### 5.2 Auth Component

The Auth component handles user authentication. In `components/auth/auth.go`:

```go
package auth

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
	fmt.Println("Starting Auth component")
	a.Refs().HTTPServer.Component().HandleFunc("/login", a.handleLogin)
	return nil
}

func (a *authComponent) handleLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	// Simple authentication logic, should be more secure in actual applications
	if username != "" {
		a.Refs().EventSystem.Component().DispatchEvent(context.Background(), &authapi.LoginEvent{Username: username})
		w.Write([]byte("Login successful"))
	} else {
		http.Error(w, "Invalid username", http.StatusBadRequest)
	}
}
```

In `components/auth/authapi/authapi.go`:

```go
package authapi

import (
	"context"
	"reflect"

	"github.com/gopherd/core/event"
)

type Component interface {
	// Define public methods for the Auth component here, if any
	// If there are none, this interface can be omitted
}

// Events can also be defined here, or events can be centrally defined in the project
type LoginEvent struct {
	Username string
}

// The following code about event types and Listeners can be generated using the github.com/gopherd/tools/cmd/eventer tool, via go generate
var loginEventType = reflect.TypeOf((*LoginEvent)(nil))

func (e *LoginEvent) Typeof() reflect.Type {
	return loginEventType
}

func init() {
	event.Register(new(LoginEvent))
}

func LoginEventListener(f func(context.Context, *LoginEvent) error) event.Listener[reflect.Type] {
	return event.Listen(loginEventType, f)
}
```

### 5.3 Users Component

The Users component handles user-related functionality. In `components/users/users.go`:

```go
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
	fmt.Println("Starting Users component")
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
	fmt.Println("User logged in", "username", e.Username)
	u.loggedInUsers[e.Username] = true
	if len(u.loggedInUsers) > u.Options().MaxUsers {
		fmt.Println("Warning: Too many users logged")
	}
	return nil
}
```

Now that we have implemented the core components, let's continue to refine our application.

## 6. Component Dependencies and References

### 6.1 Inter-component Dependencies

In our example, both the Auth and Users components depend on the HTTPServer and EventSystem components. These dependencies are established through the `Refs` field and the UUID in the configuration file.

For example, in the Auth component:

```go
type authComponent struct {
	component.BaseComponentWithRefs[struct{
		Secret string
	}, struct{
		HTTPServer  component.Reference[httpserverapi.Component]
		EventSystem component.Reference[event.Dispatcher[reflect.Type]]
	}]
}
```

Here, the `HTTPServer` and `EventSystem` fields define the Auth component's dependencies on these two components.

In the configuration file, we specify these dependencies through the `Refs` field, where the value is the UUID of the corresponding component:

```json
{
    "Name": "auth",
    "Refs": {
        "HTTPServer": "{{.R.HTTPServer}}",
        "EventSystem": "{{.R.EventSystem}}"
    },
    "Options": {
        "Secret": "{{.Namespace}}-secret"
    }
}
```

*Note*: Some might ask, why not use the component name as the basis for dependencies, and instead need a UUID? This is because some components may exist multiple times in a service, such as DB, connecting to different databases resulting in multiple DB component instances. When referencing, it's not possible to distinguish based on the component name.

### 6.2 Role and Implementation of API Packages

API packages (such as `httpserverapi`, `authapi`) play a crucial role in our architecture. They define the interfaces exposed by each component, allowing other components to depend on these interfaces rather than specific implementations.

This design has several benefits:

1. Decoupling: Components interact through interfaces rather than specific implementations, reducing coupling.
2. Flexibility: It's easy to replace component implementations as long as the new implementation satisfies the interface definition.
3. Avoiding Circular Dependencies: By separating interface definitions and implementations, package-level circular dependencies can be effectively avoided.

For example, the `httpserverapi.Component` interface defines the functionality that the HTTP server component should provide, without involving specific implementation details:

```go
type Component interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
}
```

Other components (such as Auth and Users) can depend on this interface without needing to know the specific implementation details of the HTTP server.

Let's expand our configuration file to add the newly developed components:

```json
{
    "Context": {
        "Namespace": "example",
        "Name": "example-server",
        "ID": 1001,
        "R": {
            "HTTPServer": "httpserver",
            "EventSystem": "eventsystem"
        }
    },
    "Components": [
        {
            "Name": "github.com/gopherd/example/components/eventsystem",
            "UUID": "{{.R.EventSystem}}",
            "Options": {
                "Ordered": true
            }
        },
        {
            "Name": "github.com/gopherd/example/components/auth",
            "Refs": {
                "HTTPServer": "{{.R.HTTPServer}}",
                "EventSystem": "{{.R.EventSystem}}"
            },
            "Options": {
                "Secret": "{{.Namespace}}-secret"
            }
        },
        {
            "Name": "github.com/gopherd/example/components/users",
            "Refs": {
                "HTTPServer": "{{.R.HTTPServer}}",
                "EventSystem": "{{.R.EventSystem}}"
            },
            "Options": {
                "MaxUsers": 1000
            }
        },
        {
            "Name": "github.com/gopherd/example/components/httpserver",
            "UUID": "{{.R.HTTPServer}}",
            "Options": {
                "Addr": ":{{add 8000 .ID}}"
            }
        },
        {
            "Name": "github.com/gopherd/example/components/blockexit"
        }
    ]
}
```

Then update `main.go`:

```go
package main

import (
	"github.com/gopherd/core/service"

	// Import components, the init method of the component package will register components
	_ "github.com/gopherd/example/components/auth"
	_ "github.com/gopherd/example/components/blockexit"
	_ "github.com/gopherd/example/components/eventsystem"
	_ "github.com/gopherd/example/components/httpserver"
	_ "github.com/gopherd/example/components/users"
)

func main() {
	service.Run()
}
```

Now we can run it, noting that we add the `-T` parameter when starting to use templates.

```sh
go run main.go -T config.json
```

At this point, you should see the following output:

```
Starting Auth component
Starting Users component
Starting HTTP server addr :9001
Starting blockExitComponent
```

Let's try logging in by visiting [http://localhost:9001/login?username=xiaowang]() in a browser, or using the curl command:

```sh
curl http://localhost:9001/login?username=xiaowang
```

If everything is normal, the access will receive a `Login successful` response, and the control output will show:

```
User logged in username xiaowang
```

This is the output of the LoginEvent event that the users component is listening to. Now, everything is functioning normally. What we have developed is a series of components like this, implementing mutual calls through dependency injection, and also communicating through the event system. Next, let's develop a logger component to handle logging.

## 7. Logging System

Gopherd/core uses the `log/slog` package from the Go standard library for logging. Each component can use its `Logger()` method to obtain a Logger with component context to output logs, which is the recommended approach as it tracks which component the log originates from.

### 7.1 Using the log/slog Package

Example of using logs in a component:

```go
func (c *myComponent) doSomething() {
	c.Logger().Info("Doing something", "key", "value")
	// or
	c.Logger().Info("Doing something", slog.String("key", "value"))
}
```

`Gopherd/core` only uses the `log/slog` logging tool. For information about `log/slog`, you can refer to this article [https://betterstack.com/community/guides/logging/logging-in-go/]() for an introduction and explanation, as well as the official documentation. In short, after the introduction of `log/slog`, it is no longer recommended to develop your own or use other open-source logging systems. Any other custom handling of logs can be done by implementing `slog.Handler`. The `Logger()` method of components is to obtain a `*slog.Logger` containing basic context information of the component.

Up to now, we haven't output any logs, but have been using fmt.Println. Next, we will package the initialization of logs into a component and configure it in Components.

### 7.2 Implementing a Logger Component

Let's add `components/logger/logger.go`:

```go
package logger

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"

	"github.com/gopherd/core/component"
)

const name = "github.com/gopherd/example/components/logger"

func init() {
	component.Register(name, func() component.Component {
		return new(loggerComponent)
	})
}

type loggerComponent struct {
	component.BaseComponent[struct {
		JSON   bool       // Whether to use JSON format output
		Level  slog.Level // Log level
		Output string     // Where to output logs, here we simply implemented stderr, stdout, discard
	}]
}

func (com *loggerComponent) Init(ctx context.Context) error {
	output, err := com.createOutput()
	if err != nil {
		return err
	}

	opts := &slog.HandlerOptions{
		Level: com.Options().Level,
	}
	var handler slog.Handler
	if com.Options().JSON {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}
	slog.SetDefault(slog.New(handler))
	return nil
}

func (com *loggerComponent) createOutput() (io.Writer, error) {
	switch com.Options().Output {
	case "stderr":
		return os.Stderr, nil
	case "stdout":
		return os.Stdout, nil
	case "":
		return io.Discard, nil
	default:
		return nil, errors.New("unsupported output")
	}
}
```

Then update `main.go`:

```go
package main

import (
	"github.com/gopherd/core/service"

	// Import components, the init method of the component package will register components
	_ "github.com/gopherd/example/components/auth"
	_ "github.com/gopherd/example/components/blockexit"
	_ "github.com/gopherd/example/components/eventsystem"
	_ "github.com/gopherd/example/components/httpserver"
	_ "github.com/gopherd/example/components/logger"
	_ "github.com/gopherd/example/components/users"
)

func main() {
	service.Run()
}
```

In the `config.json` configuration file, also add logger. Considering that everyone should use logs, the logger component should be initialized first, so logger should be the first component (when configuring components, sometimes you need to consider which components go first and which go last, for example, blockexit is always the last one, and httpserver is placed second to last because we want other components to be ready before starting the HTTP interface for client access). The latest `config.json` is as follows:

```json
{
    "Context": {
        "Namespace": "example",
        "Name": "example-server",
        "ID": 1001,
        "R": {
            "HTTPServer": "httpserver",
            "EventSystem": "eventsystem"
        }
    },
    "Components": [
        {
            "Name": "github.com/gopherd/example/components/logger",
                "Options": {
                "Level": "DEBUG",
                "Output": "stdout"
            }
        },
        {
            "Name": "github.com/gopherd/example/components/eventsystem",
            "UUID": "{{.R.EventSystem}}",
            "Options": {
                "Ordered": true
            }
        },
        {
            "Name": "github.com/gopherd/example/components/auth",
            "Refs": {
                "HTTPServer": "{{.R.HTTPServer}}",
                "EventSystem": "{{.R.EventSystem}}"
            },
            "Options": {
                "Secret": "{{.Namespace}}-secret"
            }
        },
        {
            "Name": "github.com/gopherd/example/components/users",
            "Refs": {
                "HTTPServer": "{{.R.HTTPServer}}",
                "EventSystem": "{{.R.EventSystem}}"
            },
            "Options": {
                "MaxUsers": 1000
            }
        },
        {
            "Name": "github.com/gopherd/example/components/httpserver",
            "UUID": "{{.R.HTTPServer}}",
            "Options": {
                "Addr": ":{{add 8000 .ID}}"
            }
        },
        {
            "Name": "github.com/gopherd/example/components/blockexit"
        }
    ]
}
```

Change all the places where `fmt.Println` was used for output to the component's `Logger().Info` function. For example, the Start function of the `httpserver` component:

```go
func (h *httpserverComponent) Start(ctx context.Context) error {
	h.Logger().Info("Starting HTTP server", "addr", h.server.Addr)
	go func() {
		if err := h.server.ListenAndServe(); err != http.ErrServerClosed {
			h.Logger().Error("HTTP server error", "error", err)
		}
	}()
	return nil
}
```

Replace others with logs as well, then run the program:

```sh
go run main.go -T config.json
```

The output should look something like this, quite verbose:

```
time=2024-08-11T01:10:34.156+08:00 level=INFO msg="component initialized" component=github.com/gopherd/example/components/logger
time=2024-08-11T01:10:34.157+08:00 level=INFO msg="initializing component" component=github.com/gopherd/example/components/eventsystem#eventsystem
time=2024-08-11T01:10:34.157+08:00 level=INFO msg="component initialized" component=github.com/gopherd/example/components/eventsystem#eventsystem
....... other omitted .......
```

In fact, the framework logs before and after calling all 4 lifecycle functions of the components, so usually when implementing components, there's no need to log again at the beginning of the lifecycle.

Alright, by now, we have implemented the basic structure of our app. For more functionality, we just need to develop components and configure them. Each component completes its own work. Other things we use, including `DB` and `redis`, should be wrapped as components for use. Anything can be provided as a component for use. But sometimes there are some auxiliary functions that are still provided as corresponding packages. Components are for managing resources and functionalities with lifecycles.

Finally, let's take a look at the current directory structure:

```
example/
├── components
│   ├── auth
│   │   ├── auth.go
│   │   └── authapi
│   │       └── authapi.go
│   ├── blockexit
│   │   └── blockexit.go
│   ├── eventsystem
│   │   └── eventsystem.go
│   ├── httpserver
│   │   ├── httpserver.go
│   │   └── httpserverapi
│   │       └── httpserverapi.go
│   ├── logger
│   │   └── logger.go
│   └── users
│       └── users.go
├── config.json
└── main.go
```

The code for this example project is hosted at [https://github.com/gopherd/example]().

## 8. Advanced Topics

### 8.1 Detailed Explanation of Component Dependency Mechanism

The component dependency mechanism in Gopherd/core is based on several key concepts:

1. UUID: Each component that can be depended upon has a unique identifier (UUID).
2. Refs: Components declare their dependencies on other components through the Refs field.
3. Configuration File: In the configuration file, the UUID of the dependent component is associated with the components that depend on it through the Refs field.

The framework automatically parses these dependency relationships during initialization and injects the correct component instances into the components that depend on them.

### 8.2 API Package Design Principles

When designing API packages, the following principles should be followed:

1. Only define interfaces, do not include implementation details.
2. Interfaces should be small and focused, only including necessary methods.
3. Use generic types, avoid introducing types specific to a particular implementation.
4. Consider future extensibility, but don't over-design.

### 8.3 Avoiding Circular Dependencies

To avoid circular dependencies, the following strategies can be adopted:

1. Use dependency injection: Inject dependencies through the Refs field instead of direct imports.
2. Separate interfaces and implementations: Put interface definitions in separate API packages.
3. Redesign: If circular dependencies occur, it may indicate a need to reconsider the division of component responsibilities.

### 8.4 Command Line Arguments

Gopherd/core provides several useful command line arguments:

- `-v`: Print version information and exit
- `-p`: Print parsed configuration and exit
- `-t`: Test configuration validity and exit
- `-T`: Enable configuration template processing

Usage examples:

```bash
# Enable template processing and run the application
go run main.go -T config.json

# Print parsed configuration. Without -T, it outputs the original config.json, while with -T, it outputs the configuration after template processing
go run main.go -p config.json
go run main.go -p -T config.json

# Test configuration validity
go run main.go -t config.json
go run main.go -t -T config.json
```

### 8.5 Testing Strategy

The component design of Gopherd/core makes testing simple.

1. Write unit tests for each component
2. Use mock objects to simulate dependent components
3. Write integration tests to test interactions between components

### 8.6 Resource Management

1. Allocate resources in the Init method, release resources in the Uninit method.
2. Use defer statements to ensure resources are properly released.
3. Gracefully close long-running operations in the Shutdown method.
4. The Init method should only be responsible for initializing itself. At this point, it should not use dependent components as they may not have completed initialization. If you need to use dependent components for initialization, that part of initialization should be placed in Start. Remember this!

## 9. Conclusion

The Gopherd/core framework provides a powerful and flexible way to build modular Go applications. By using the component system, dependency injection, and event mechanism, highly maintainable and extensible applications can be created.

This guide covers the core concepts and usage methods of Gopherd/core, from basic component implementation to advanced features and best practices. By following these guidelines, developers can fully leverage the advantages of the Gopherd/core framework to build robust, efficient backend services.

As you delve deeper into using the framework, you'll discover more possibilities and uses. Continuous practice and exploration will help you better master Gopherd/core and maximize its potential in real projects.