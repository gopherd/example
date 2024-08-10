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

func LoginEventListener(f func(context.Context, *LoginEvent) error) event.Listener[reflect.Type] {
	return event.Listen(loginEventType, f)
}
