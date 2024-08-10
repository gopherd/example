package blockexit

import (
	"context"
	"os"
	"os/signal"

	"github.com/gopherd/core/component"
)

const Name = "github.com/gopherd/example/components/blockexit"

type blockExitComponent struct {
	component.BaseComponent[struct{}]
}

func init() {
	component.Register(Name, func() component.Component {
		return new(blockExitComponent)
	})
}

func (b *blockExitComponent) Start(ctx context.Context) error {
	b.Logger().Info("Starting blockExitComponent")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	b.Logger().Info("Received interrupt signal")
	return nil
}
