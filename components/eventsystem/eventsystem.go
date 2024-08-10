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
