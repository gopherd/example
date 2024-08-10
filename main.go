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
