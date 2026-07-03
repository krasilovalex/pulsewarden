package lifecycle

import (
	"context"
	"testing"
	"time"
)

func TestSignalContextCancelledWhenParentCancelled(t *testing.T) {
	t.Parallel()

	parent, cancelParent := context.WithCancel(context.Background())
	ctx, stop := SignalContext(parent)
	defer stop()

	cancelParent()

	select {
	case <-ctx.Done():
		if err := ctx.Err(); err != context.Canceled {
			t.Fatalf("unexpected context error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("context was not cancelled")
	}
}
