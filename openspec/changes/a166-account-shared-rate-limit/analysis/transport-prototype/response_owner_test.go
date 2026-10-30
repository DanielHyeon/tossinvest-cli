//go:build linux

package prototype

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"
)

type racedBody struct {
	entered  chan struct{}
	closed   chan struct{}
	closes   atomic.Int32
	closeErr error
}

func (b *racedBody) Read([]byte) (int, error) { close(b.entered); <-b.closed; return 0, io.EOF }
func (b *racedBody) Close() error {
	if b.closes.Add(1) == 1 {
		close(b.closed)
	}
	return b.closeErr
}
func TestResponseOwnerConcurrentReadCloseCancelExactlyOnce(t *testing.T) {
	for _, closeErr := range []error{nil, errors.New("fixture-close-error")} {
		t.Run(func() string {
			if closeErr != nil {
				return "close-error"
			}
			return "normal"
		}(), func(t *testing.T) {
			body := &racedBody{entered: make(chan struct{}), closed: make(chan struct{}), closeErr: closeErr}
			var finishes atomic.Int32
			owner := &responseOwner{body: body, done: make(chan struct{}), finish: func() { finishes.Add(1) }}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var wg sync.WaitGroup
			wg.Add(3)
			errs := make(chan error, 2)
			go func() {
				defer wg.Done()
				_, err := owner.Read(make([]byte, 1))
				if !errors.Is(err, io.EOF) {
					t.Errorf("Read: %v", err)
				}
			}()
			<-body.entered
			start := make(chan struct{})
			go func() { defer wg.Done(); <-start; errs <- owner.Close() }()
			go func() { defer wg.Done(); <-start; cancel(); <-ctx.Done(); errs <- owner.Close() }()
			close(start)
			wg.Wait()
			close(errs)
			for err := range errs {
				if !errors.Is(err, closeErr) {
					t.Errorf("Close error=%v want=%v", err, closeErr)
				}
			}
			if body.closes.Load() != 1 || finishes.Load() != 1 {
				t.Fatalf("underlying Close=%d finish=%d", body.closes.Load(), finishes.Load())
			}
			if err := owner.Close(); !errors.Is(err, closeErr) {
				t.Errorf("later Close=%v", err)
			}
		})
	}
}
