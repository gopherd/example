# Gopherd/core 框架使用指南

## 1. 介绍

[Gopherd/core](https://github.com/gopherd/core) 是一个现代化的 Go 语言框架，专为构建可扩展、模块化的后端服务而设计。它充分利用了 Go 的泛型特性，提供了一种类型安全、灵活且高效的方式来开发应用程序。

本指南将通过一个逐步构建的 Web 服务器项目来展示 Gopherd/core 的各个特性和使用方法。我们的示例项目将包含 HTTP 服务器、事件系统、身份验证（auth）和用户管理（users）等功能。

### 1.1 主要特性

- **基于泛型的架构**：利用 Go 的泛型实现类型安全的组件创建和管理
- **灵活的配置**：通过类型安全的方式从文件、URL 或标准输入加载配置
- **模板处理**：在组件配置中使用 Go 模板实现动态设置
- **多格式支持**：通过编码器和解码器处理 JSON、TOML、YAML 和其他任意配置格式
- **自动依赖注入**：通过内置的依赖解析和注入简化组件集成

## 2. 快速开始

### 2.1 安装

确保你的 Go 版本至少是 1.21。然后，在你的项目中安装 Gopherd/core：

```bash
go get github.com/gopherd/core
```

### 2.2 项目结构

我们的示例项目结构如下：

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

### 2.3 创建主程序

让我们从创建一个最小的主程序开始。在 `main.go` 中：

```go
package main

import (
	"github.com/gopherd/core/service"
)

func main() {
	service.Run()
}
```

这个简单的主程序调用 `service.Run()` 来启动应用程序。现在，它还不会做任何事情，因为我们还没有注册任何组件。

我们首先实现一个最简单的 `blockexit` 组件，这个组件就是保持程序一直运行不退出，使用 `Ctrl-C` 关闭。

创建 `components/blockexit/blockexit.go`：

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

然后修改 `main.go`

```go
package main

import (
	"github.com/gopherd/core/service"

	// 引入组件，组件包的 init 方法会注册组件
	_ "github.com/gopherd/example/components/blockexit"
)

func main() {
	service.Run()
}
```

main 函数没有变化，永远都只需要这样，我们修改的是增加了一个 import 将实现的组件引入。现在可以在命令行中执行了：

```sh
echo '{"Components":[{"Name":"github.com/gopherd/example/components/blockexit"}]}' | go run main.go -
```

运行后你将看到输出 `Starting blockExitComponent`，并且程序不退出，`Ctrl-C` 即可退出。

*注*：这里需要注意
* 组件的名称推荐使用包名，这样可以避免不小心名称重复
* 这里的运行没有使用配置文件，而是通过标准输入读取的配置信息，配置中只配置了一个 blockexit 组件。
* 我们还支持从文件和http服务中获取配置，使用 `-h` 可以查看使用帮助。

## 3. 基本组件实现

刚才的组件过于简单，接下来让我们从实现一个基本的 HTTP 服务器组件开始，以理解组件的基本结构和注册过程。

### 3.1 HTTP 服务器组件

首先，我们创建 HTTP 服务器组件的 API 定义。在 `components/httpserver/httpserverapi/httpserverapi.go` 中：

```go
package httpserverapi

import "net/http"

type Component interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
}
```

然后，我们实现 HTTP 服务器组件。在 `components/httpserver/httpserver.go` 中：

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

// 断言 httpserverComponent 实现了接口 httpserverapi.Component
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
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

### 3.2 基本配置

创建一个基本的 `config.json` 文件：

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

### 3.3 更新主程序

现在，我们更新 `main.go` 来导入 HTTP 服务器组件：

```go
package main

import (
	"github.com/gopherd/core/service"

	// 引入组件，组件包的 init 方法会注册组件
	_ "github.com/gopherd/example/components/blockexit"
	_ "github.com/gopherd/example/components/httpserver"
)

func main() {
	service.Run()
}
```

现在，你可以运行你的应用：

```bash
go run main.go config.json
```

这将启动一个基本的 HTTP 服务器，监听在 8080 端口。启动后你将看到如下的输出

```
Starting HTTP server addr :8080
Starting blockExitComponent
```

按 `Ctrl-C` 后程序将会关闭，将输出

```
Received interrupt signal
Shutting down HTTP server
```

这里我们可以基本说明一下了：

**组件的开发**: 嵌入一个 component.BaseComponent[T]，范型参数 T 是组件的配置，代码中通过组件的 Options() 获取到，然后可以选择实现组件的 `Init`，`Start`，`Shutdown`（还有这里没有出现的 `Uninit`）等生命周期的函数。

**组件的组册**: 在 init 函数中调用 component.Register 根据组件的名称注册了一个构造函数用户创建出我们实现的组件对象，这个函数我们不需要去初始化组件的任何数据，任何组件都只需要 new 创建就可以了。最后在 main 中引入这个包即可完成注册。

**组件的配置**: 注册的组件并不会自己就运行，需要在配置文件中的 Components 数组下增加这个组件的配置。

**组件的顺序**: Components 中的顺序就是 Init 和 Start 函数的调用顺序，Shutdown 和 Uninit 则是反过来的，根据我们当前的配置，执行的两个组件的生命周期函数顺序如下

```
blockexit.Init      -> httpserver.Init    ->
blockexit.Start     -> httpserver.Start   ->
httpserver.Shutdown -> blockexit.Shutdown ->
httpserver.Uninit   -> blockexit.Uninit   -> 程序退出
```

## 4. 配置管理和模板特性

Gopherd/core 提供了灵活的配置管理机制，包括模板特性。支持 `json`，`toml`，`yaml` 三种配置格式，默认使用 `json`，让我们深入了解如何使用这些功能。

### 4.1 配置文件结构

配置文件通常是一个 JSON 文件，包含以下主要部分：

- `Context`: 全局上下文，可以在模板中使用
- `Components`: 组件列表，每个组件包含必需的 `Name` 和可选的 `UUID`、`Refs`、`Options`

### 4.2 模板语法和使用

配置文件支持 Go 的模板语法。你可以使用 `{{.}}` 来引用上下文中的值。例如：

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

在这个例子中：

- `{{.R.HTTPServer}}` 会被替换为 "httpserver"
- `{{add 8000 .ID}}` 会被计算为 9001（8000 + 1001）

要使用模板，在运行程序是需要加上 `-T` 参数，表示启用模板，默认是不启用的。

### 4.3 JSON 配置支持最简单的行注释

对于配置使用 `JSON` 格式，考虑到配置中可能需要有一些说明，所以支持了以 `//` 行首（可以有前导空白符）的行注释，不支持 `/* ... */` 这种块注释。

合法的注释例子：

```jsonc
{
	// 合法的注释
	// 还是合法的注释
	"Context": {
		// 也是合法的注释
		"ID": 1001
	},
	"Components": [
		// 仍然是合法的注释
	]
}
```

不合法的注释例子：

```json
{
	/* 不合法的注释 */
	"Context": { // 不合法的注释
		"ID": 1001, // 也是不合法的注释
	},
	"Components": [
		/*
		不合法的注释
		*/
	]
}
```

### 4.4 如何使用 TOML 或 YAML 格式的配置，以及其他任意格式的配置

首先需要修改 main 函数。

对于 `TOML` 格式：

```go
func main() {
	service.Run(service.WithEncoder(toml.Marshal), service.WithDecoder(toml.Unmarshal))
}
```

对于 `YAML` 格式：

```go
func main() {
	service.Run(service.WithEncoder(yaml.Marshal), service.WithDecoder(yaml.Unmarshal))
}
```

其中的 `toml` 包和 `yaml` 包可以自行选择使用。比如以下几个流行的库均支持：

* [github.com/BurntSushi/toml](https://github.com/BurntSushi/toml)
* [github.com/pelletier/go-toml/v2](https://github.com/pelletier/go-toml)
* [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml)
* [github.com/goccy/go-yaml](https://github.com/goccy/go-yaml)

*注*: 需要说明的是，目前对 `toml` 和 `yaml` 的支持是通过一次转换的来的，即现将读取的配置文件根据传入的 Decoder 解析到一个 `map[string]any` 中，然后再将其编码成 json，后面程序中就会继续使用 `json` 了。

**参照以上方式通过传递编码器，解码器参数，你可以支持任意格式的配置。**

## 5. 实现核心组件

现在，让我们实现其他核心组件，包括 EventSystem，Auth 和 Users 组件。

### 5.1 事件系统组件实现

首先，让我们实现事件系统组件。



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

// 我们没有定义一个单独的 eventsystemapi 的包，直接使用了 event.Dispatcher 作为组件的导出接口
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

### 5.2 Auth 组件

Auth 组件处理用户认证。在 `components/auth/auth.go` 中：

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
	// 简单的认证逻辑，实际应用中应该更加安全
	if username != "" {
		a.Refs().EventSystem.Component().DispatchEvent(context.Background(), &authapi.LoginEvent{Username: username})
		w.Write([]byte("Login successful"))
	} else {
		http.Error(w, "Invalid username", http.StatusBadRequest)
	}
}
```

在 `components/auth/authapi/authapi.go` 中：

```go
package authapi

import (
	"context"
	"reflect"

	"github.com/gopherd/core/event"
)

type Component interface {
	// 如果有的话，可以在这里定义 Auth 组件的公共方法
	// 如果没有，则可以不定义这个接口
}

// 事件也可以定义在这里，或者项目中可以集中定义事件
type LoginEvent struct {
	Username string
}

// 以下关于事件的类型和 Listener 代码可以使用 github.com/gopherd/tools/cmd/eventer 工具生成，通过 go generate
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

### 5.3 Users 组件

Users 组件处理用户相关的功能。在 `components/users/users.go`


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

现在我们已经实现了核心组件，让我们继续完善我们的应用程序。

## 6. 组件依赖和引用

### 6.1 组件间依赖关系

在我们的示例中，Auth 和 Users 组件都依赖于 HTTPServer 和 EventSystem 组件。这种依赖关系通过 `Refs` 字段和配置文件中的 UUID 来建立。

例如，在 Auth 组件中：

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

这里，`HTTPServer` 和 `EventSystem` 字段定义了 Auth 组件对这两个组件的依赖。

在配置文件中，我们通过 `Refs` 字段来指定这些依赖，依赖的值是对应组件的 UUID：

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

*注*：可能有人会问，为什么不用组件名称作为依赖的依据，而需要一个 UUID 呢？因为考虑到某些组件在一个服务中可能存在多个，比如 DB，连接不同库的产生了多个 DB 组件示例，在引用时根据组件名是区分不出来的。

### 6.2 API包的作用和实现

API 包（如 `httpserverapi`、`authapi`）在我们的架构中扮演着关键角色。它们定义了每个组件对外暴露的接口，使得其他组件可以依赖这些接口而不是具体实现。

这种设计有以下几个好处：

1. 解耦：组件之间通过接口而不是具体实现进行交互，降低了耦合度。
2. 灵活性：可以轻松替换组件的实现，只要新的实现满足接口定义。
3. 避免循环依赖：通过将接口定义和实现分离，可以有效避免包级别的循环依赖问题。

例如，`httpserverapi.Component` 接口定义了 HTTP 服务器组件应该提供的功能，而不涉及具体实现细节：

```go
type Component interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
}
```

其他组件（如 Auth 和 Users）可以依赖这个接口，而不需要知道 HTTP 服务器的具体实现细节。

让我们扩展我们的配置文件，添加的新开发的组件：

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

然后更新 `main.go`

```go
package main

import (
	"github.com/gopherd/core/service"

	// 引入组件，组件包的 init 方法会注册组件
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

然后我们就可以运行了，注意因为要使用模板，启动时增加了 `-T` 参数。

```sh
go run main.go -T config.json
```

此时可以看到一下输出：

```
Starting Auth component
Starting Users component
Starting HTTP server addr :9001
Starting blockExitComponent
```

我们在浏览器访问一下地址 [http://localhost:9001/login?username=xiaowang]() 去登录一下，也可以使用 curl 命令：

```sh
curl http://localhost:9001/login?username=xiaowang
```

一切正常的话，访问会收到 `Login successful` 的返回，控制输出

```
User logged in username xiaowang
```

这是 users 组件监听的 LoginEvent 事件的输出。到现在，功能都正常运作了，我们所开发的就是这样一个一个的组件，同作依赖注入实现相互调用，也可通过事件系统通信。接下来我们开发一个日志组件来处理日志。


## 7. 日志系统

Gopherd/core 使用 Go 标准库的 `log/slog` 包进行日志记录。每个组件都可以使用其 `Logger()` 方法获取带有组件上下文的 Logger 来输出日志，这也是建议的方式，它追踪了日志源于那个组件。

### 7.1 使用 log/slog 包

在组件中使用日志的示例：

```go
func (c *myComponent) doSomething() {
	c.Logger().Info("Doing something", "key", "value")
	// 或者
	c.Logger().Info("Doing something", slog.String("key", "value"))
}
```

`Gopherd/core` 只使用 `log/slog` 日志工具，关于 `log/slog` 可以参见这个文章 [https://betterstack.com/community/guides/logging/logging-in-go/]() 的介绍说明以及官方文档。总之在 `log/slog` 出来之后已经不再推荐自己开发或者使用其他的开源日志系统了，其他各种关于日志的自定义处理都可以通过实现 `slog.Handler` 来完成。组件的 `Logger()` 方法则是获取一个含有组件基本上下文信息的 `*slog.Logger`。

截至目前，我们还没有输出过日志，而使用了 fmt.Println，接下来我们将日志的初始化等操作也包装成一个组件并配置到 Components 中。

### 7.2 实现一个 logger 组件

我们增加 `components/logger/logger.go`

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
		JSON   bool       // 是否使用 json 格式输出
		Level  slog.Level // 日志等级
		Output string     // 日志输出到哪里，这里简单的实现了 stderr, stdout, discard
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

然后更新 `main.go`

```go
package main

import (
	"github.com/gopherd/core/service"

	// 引入组件，组件包的 init 方法会注册组件
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

配置文件 `config.json` 中也许加入 logger，考虑到大家都应该要使用日志，logger 组件应该最先初始化，所以 logger 要作为第一个组件（在配置组件时，有时候需要考虑一下哪些组件在前，哪些在后，比如之前 blockexit 总是最后一个， httpserver 放在倒数第二则是希望其他组件都准备好了之后才启动 http 接口给客户端访问）。最新的 `config.json` 如下：

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

把之前使用 `fmt.Println` 输出的地方都改成组件的 `Logger().Info` 函数。比如 `httpserver` 组件的 Start 函数：

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

其他的也依次替换成日志，然后运行程序：

```sh
go run main.go -T config.json
```

输出大概像这样了，比较冗长：

```
time=2024-08-11T01:10:34.156+08:00 level=INFO msg="component initialized" component=github.com/gopherd/example/components/logger
time=2024-08-11T01:10:34.157+08:00 level=INFO msg="initializing component" component=github.com/gopherd/example/components/eventsystem#eventsystem
time=2024-08-11T01:10:34.157+08:00 level=INFO msg="component initialized" component=github.com/gopherd/example/components/eventsystem#eventsystem
....... 其他省略 .......
```

事实上，框架在所有组件的 4 个生命周期函数调用前后都有日志，所以通常组件实现时生命周开始的地方不用在打日志了。


好了，到现在，我们的基本的 app 结构都实现了，需要更多的功能就不同开发组件，配置组件就可以了。每个组件完成自己的工作。我们使用的其他的包括 `DB`，`redis` 都应该包装成组件以供使用，任何东西就可以作为组件提供使用。但有时后有一些辅助功能，函数，仍然提供为相应的包，组件是管理有生命周期的各类资源，功能等。

最后我们在看一下当前的目录结构：

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

这个示例项目代码托管在 [https://github.com/gopherd/example]()。


## 8. 高级主题

### 8.1 组件依赖机制详解

Gopherd/core 的组件依赖机制基于以下几个关键概念：

1. UUID：每个可被依赖的组件都有一个唯一标识符（UUID）。
2. Refs：组件通过 Refs 字段声明对其他组件的依赖。
3. 配置文件：在配置文件中，通过 Refs 字段将依赖组件的 UUID 与被依赖组件关联起来。

框架在初始化时会自动解析这些依赖关系，并将正确的组件实例注入到依赖它们的组件中。

### 8.2 API包设计原则

设计 API 包时，应遵循以下原则：

1. 只定义接口，不包含实现细节。
2. 接口应该是小而精的，只包含必要的方法。
3. 使用通用的类型，避免引入特定实现的类型。
4. 考虑未来的扩展性，但不过度设计。

### 8.3 避免循环依赖

为了避免循环依赖，可以采取以下策略：

1. 使用依赖注入：通过 Refs 字段注入依赖，而不是直接导入。
2. 分离接口和实现：将接口定义放在单独的 API 包中。
3. 重新设计：如果出现循环依赖，可能意味着需要重新考虑组件的职责划分。

### 8.4 命令行参数

Gopherd/core 提供了几个有用的命令行参数：

- `-v`: 打印版本信息并退出
- `-p`: 打印解析后的配置并退出
- `-t`: 测试配置有效性并退出
- `-T`: 启用配置模板处理

使用示例：

```bash
# 启用模板处理并运行应用
go run main.go -T config.json

# 打印解析后的配置，不带 -T 输出的就是原来的 config.json，而带有 -T 的输出的就是经过模版处理后的配置
go run main.go -p config.json
go run main.go -p -T config.json

# 测试配置有效性
go run main.go -t config.json
go run main.go -t -T config.json
```

### 8.5 测试策略

Gopherd/core 的组件设计使得测试变得简单。

1. 为每个组件编写单元测试
2. 使用模拟（mock）对象来模拟依赖组件
3. 编写集成测试来测试组件间的交互


### 8.6 资源管理

1. 在 Init 方法中分配资源，在 Uninit 方法中释放资源。
2. 使用 defer 语句确保资源被正确释放。
3. 在 Shutdown 方法中优雅地关闭长时间运行的操作。
4. Init 方法应该只负责初始化自己，这个时候不应该去使用依赖的组件，因为他可能还没有完成初始化，如果需要使用依赖的组件执行初始化，那么这部分初始化应该放在 Start 中。切记！！


## 9. 总结

Gopherd/core 框架提供了一种强大而灵活的方式来构建模块化的 Go 应用程序。通过使用组件系统、依赖注入和事件机制，可以创建出高度可维护和可扩展的应用。

本指南涵盖了 Gopherd/core 的核心概念和使用方法，从基本的组件实现到高级特性和最佳实践。通过遵循这些指导原则，开发者可以充分利用 Gopherd/core 框架的优势，构建出健壮、高效的后端服务。

随着你对框架的深入使用，你会发现更多的可能性和用法。不断实践和探索，将帮助你更好地掌握 Gopherd/core，并在实际项目中发挥其最大潜力。