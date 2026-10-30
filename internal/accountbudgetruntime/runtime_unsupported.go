//go:build !linux

package accountbudgetruntime

import (
	"context"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"sync"
)

type Mode uint8

const (
	ModeOff Mode = iota + 1
	ModeRequired
)

type Config struct {
	Mode         Mode
	Manifest     string
	CleanupError func(error)
}
type Factory struct {
	mode   Mode
	mu     sync.Mutex
	closed bool
}

func FromEnv(lookup func(string) (string, bool), key string, cleanup func(error)) (Config, error) {
	if lookup == nil || key == "" {
		return Config{}, accountbudget.ErrUnavailable
	}
	path, ok := lookup(key)
	if !ok {
		return Config{Mode: ModeOff}, nil
	}
	if path == "" {
		return Config{}, accountbudget.ErrUnavailable
	}
	return Config{Mode: ModeRequired, Manifest: path, CleanupError: cleanup}, nil
}
func Open(ctx context.Context, cfg Config) (*Factory, error) {
	if ctx == nil {
		return nil, accountbudget.ErrUnavailable
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if cfg.Mode == ModeOff && cfg.Manifest == "" {
		return &Factory{mode: ModeOff}, nil
	}
	return nil, accountbudget.ErrUnavailable
}
func (f *Factory) NewClient(creds official.Credentials, cache string, opts ...official.Option) (*official.Client, error) {
	if f == nil {
		return nil, accountbudget.ErrUnavailable
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed || f.mode != ModeOff {
		return nil, accountbudget.ErrUnavailable
	}
	return official.New(creds, cache, opts...), nil
}
func (f *Factory) Close() error {
	if f == nil {
		return accountbudget.ErrUnavailable
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mode != ModeOff {
		return accountbudget.ErrUnavailable
	}
	f.closed = true
	return nil
}

// Budget returns a borrowed binding; Factory owns Close. Nil denotes OFF.
func (f *Factory) Budget() (*accountbudget.Budget, error) {
	if f == nil {
		return nil, accountbudget.ErrUnavailable
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed || f.mode != ModeOff {
		return nil, accountbudget.ErrUnavailable
	}
	return nil, nil
}
