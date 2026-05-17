package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type GomonHandler struct {
	w io.Writer
}

func NewGomonHandler(w io.Writer) *GomonHandler {
	return &GomonHandler{w: w}
}

func (h *GomonHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *GomonHandler) Handle(_ context.Context, r slog.Record) error {
	timeStr := Colorize(COLORGRAY, r.Time.Format("15:04:05"))

	var levelStr string
	switch r.Level {
	case slog.LevelDebug:
		levelStr = Colorize(COLORDEBUG, "DEBUG ")
	case slog.LevelInfo:
		levelStr = Colorize(COLORINFO, "INFO ")
	case slog.LevelWarn:
		levelStr = Colorize(COLORWARN, "WARN ")
	case slog.LevelError:
		levelStr = Colorize(COLORERROR, "ERROR ")
	default:
		levelStr = r.Level.String()
	}

	attrs := ""
	r.Attrs(func(a slog.Attr) bool {
		attrs += fmt.Sprintf(" %s%s=%v%s", COLORGRAY, a.Key, a.Value.Any(), COLORRESET)
		return true
	})

	prefix := Colorize(COLORCYAN, "gomon")
	_, err := fmt.Fprintf(h.w, "%s [%s] %s %s%s\n", 
		timeStr, 
		prefix, 
		levelStr, 
		r.Message, 
		attrs,
	)
	return err
}

func (h *GomonHandler) WithAttrs(attrs []slog.Attr) slog.Handler { 
	return h 
}


func (h *GomonHandler) WithGroup(name string) slog.Handler    { 
	return h
}

func InitDefault() {
	handler := NewGomonHandler(os.Stdout)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
