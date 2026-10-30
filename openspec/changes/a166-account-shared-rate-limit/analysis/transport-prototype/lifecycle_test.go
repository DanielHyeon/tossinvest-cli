package prototype

import (
	"errors"
	"io"
	"sync"
	"testing"
	"time"
)

type fakeConn struct {
	entered chan struct{}
	release chan struct{}
	mu      sync.Mutex
	writes  int
	closed  bool
}

func (c *fakeConn) Write(p []byte) (int, error) {
	c.mu.Lock()
	c.writes++
	c.mu.Unlock()
	close(c.entered)
	<-c.release
	return len(p), nil
}
func (c *fakeConn) Close() error { c.mu.Lock(); defer c.mu.Unlock(); c.closed = true; return nil }

type gate struct {
	mu      sync.Mutex
	revoked bool
	conn    *fakeConn
}

func (g *gate) register(c *fakeConn) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.revoked {
		_ = c.Close()
		return false
	}
	g.conn = c
	return true
}
func (g *gate) write(p []byte) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.revoked || g.conn == nil {
		return 0, io.ErrClosedPipe
	}
	return g.conn.Write(p)
}
func (g *gate) revoke() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.revoked = true
	if g.conn != nil {
		_ = g.conn.Close()
	}
}

func TestRevokeWaitsForWriteAndBlocksLaterWrites(t *testing.T) {
	c := &fakeConn{entered: make(chan struct{}), release: make(chan struct{})}
	g := &gate{}
	g.register(c)
	written := make(chan struct{})
	go func() { _, _ = g.write([]byte("x")); close(written) }()
	<-c.entered
	started := make(chan struct{})
	revoked := make(chan struct{})
	lockReleased := make(chan struct{})
	go func() { close(started); g.revoke(); close(revoked); close(lockReleased) }()
	<-started
	select {
	case <-revoked:
		t.Fatal("revoke returned during write")
	case <-time.After(30 * time.Millisecond):
	}
	select {
	case <-lockReleased:
		t.Fatal("lock released before revoke")
	default:
	}
	close(c.release)
	<-written
	<-revoked
	<-lockReleased
	if _, err := g.write([]byte("later")); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("late write: %v", err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.writes != 1 || !c.closed {
		t.Fatalf("writes=%d closed=%v", c.writes, c.closed)
	}
}
func TestRegistrationAfterRevokeClosesConnection(t *testing.T) {
	g := &gate{}
	g.revoke()
	c := &fakeConn{entered: make(chan struct{}), release: make(chan struct{})}
	if g.register(c) {
		t.Fatal("registered after revoke")
	}
	if _, err := g.write([]byte("x")); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("late write: %v", err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed || c.writes != 0 {
		t.Fatalf("closed=%v writes=%d", c.closed, c.writes)
	}
}
