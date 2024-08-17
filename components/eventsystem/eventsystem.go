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

var _ event.EventSystem[reflect.Type] = (*eventsystemComponent)(nil)

type eventsystemComponent struct {
	component.BaseComponent[struct {
		Ordered *bool
	}]
	event.EventSystem[reflect.Type]
}

func (com *eventsystemComponent) Init(ctx context.Context) error {
	ordered := true
	if com.Options().Ordered != nil {
		ordered = *com.Options().Ordered
	}
	com.EventSystem = event.NewEventSystem[reflect.Type](ordered)
	return nil
}
