package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SignalContext returns a child context that is cancelled when the process
// receives SIGINT or SIGTERM.
//
// The returned stop function must be called to release resources associated
// with signal delivery.

func SignalContext(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(
		parent,
		os.Interrupt,
		syscall.SIGTERM,
	)
}
