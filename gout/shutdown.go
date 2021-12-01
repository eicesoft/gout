package gout

import (
	"os"
	"os/signal"
	"syscall"
)

var _ Hook = (*hook)(nil)

// Hook a graceful shutdown hook, default with signals of SIGINT and SIGTERM
type Hook interface {
	// WithSignals add more signals into hook
	WithSignals(signals ...syscall.Signal) Hook

	// Close register shutdown handles
	Close(funcs ...func())
}

type hook struct {
	quit chan os.Signal
}

// ExitHook create a Hook instance
func ExitHook(funcs ...func()) Hook {
	hook := &hook{
		quit: make(chan os.Signal, 1),
	}

	return hook.WithSignals(syscall.SIGINT, syscall.SIGTERM)
}

func (h *hook) WithSignals(signals ...syscall.Signal) Hook {
	for _, s := range signals {
		signal.Notify(h.quit, s)
	}

	return h
}

func (h *hook) Close(funcs ...func()) {
	select {
	case <-h.quit:
	}
	signal.Stop(h.quit)

	for _, f := range funcs {
		f()
	}
}
