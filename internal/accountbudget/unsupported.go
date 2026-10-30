//go:build !linux

// Package accountbudget leaves enforcement unavailable outside Linux.
package accountbudget

import (
	"context"
	"errors"
	"net/http"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

var ErrUnavailable = errors.New("account budget unavailable")
var ErrNotSent = errors.New("account budget: request not sent")

type Priority uint8

const (
	Normal Priority = iota
	Critical
	Auxiliary
)

type Config struct {
	Root     string
	Scope    string
	Instance string
	Clock    clock.Clock
}
type Budget struct{}
type Lease struct{}
type Transport struct{}

func Provision(string, string, string) error                      { return ErrUnavailable }
func OpenContext(context.Context, Config) (*Budget, error)        { return nil, ErrUnavailable }
func Open(Config) (*Budget, error)                                { return nil, ErrUnavailable }
func (*Budget) Close() error                                      { return nil }
func (*Budget) Acquire(context.Context, Priority) (*Lease, error) { return nil, ErrUnavailable }
func (*Lease) Release() error                                     { return nil }

type priorityKey struct{}

func WithPriority(ctx context.Context, p Priority) context.Context {
	return context.WithValue(ctx, priorityKey{}, p)
}
func PriorityFromContext(ctx context.Context) Priority {
	p, ok := ctx.Value(priorityKey{}).(Priority)
	if !ok {
		return Normal
	}
	return p
}
func NewTransport(*Budget, *http.Transport, func(error)) (*Transport, error) {
	return nil, ErrUnavailable
}
func (*Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		return nil, errors.Join(ErrNotSent, ErrUnavailable, req.Body.Close())
	}
	return nil, errors.Join(ErrNotSent, ErrUnavailable)
}
