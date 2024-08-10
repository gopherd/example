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
