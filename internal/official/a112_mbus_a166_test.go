//go:build linux

package official

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestA166A112CloneSharesBudget(t *testing.T) {
	b := a166Budget(t)
	lease, e := b.Acquire(context.Background(), accountbudget.Critical)
	if e != nil {
		t.Fatal(e)
	}
	defer lease.Release()
	cache := filepath.Join(t.TempDir(), "token.json")
	data, e := json.Marshal(cachedToken{AccessToken: "cached", ExpiresAt: time.Now().Add(time.Hour)})
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(cache, data, 0600); e != nil {
		t.Fatal(e)
	}
	c, e := NewWithAccountBudget(Credentials{}, cache, b, func(error) {})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("dialtls=%v dial=%v next=%v tls=%v", c.authorityTransport.DialTLSContext != nil, c.authorityTransport.Dial != nil, c.authorityTransport.TLSNextProto != nil, c.authorityTransport.TLSClientConfig != nil)
	c.authorityTransport.DialContext = func(context.Context, string, string) (net.Conn, error) { return nil, errors.New("test dial refused") }
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, e = A112MBUSCandle(ctx, c, nil)
	if e == nil || ctx.Err() == nil {
		t.Fatalf("clone escaped shared budget: %v", e)
	}
}
