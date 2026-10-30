package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudgetruntime"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
)

type rootBudgetOwner struct {
	gate      chan struct{}
	cfg       accountbudgetruntime.Config
	configErr error
	factory   *accountbudgetruntime.Factory
	closing   bool
	cleanup   io.Writer
}

func parseBudgetEnv(mode string, modeSet bool, manifest string, manifestSet bool) (accountbudgetruntime.Config, error) {
	if !modeSet && !manifestSet {
		return accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeOff}, nil
	}
	switch {
	case modeSet && mode == "off" && !manifestSet:
		return accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeOff}, nil
	case modeSet && mode == "required" && manifestSet && manifest != "":
		return accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: manifest}, nil
	default:
		return accountbudgetruntime.Config{}, fmt.Errorf("%w: invalid budget launch mode", accountbudget.ErrUnavailable)
	}
}
func newRootBudgetOwner(lookup func(string) (string, bool), cleanup io.Writer) *rootBudgetOwner {
	mode, modeSet := lookup("TOSSOS_ACCOUNT_BUDGET_MODE")
	manifest, manifestSet := lookup("TOSSOS_ACCOUNT_BUDGET_MANIFEST")
	cfg, err := parseBudgetEnv(mode, modeSet, manifest, manifestSet)
	owner := &rootBudgetOwner{gate: make(chan struct{}, 1), cfg: cfg, configErr: err, cleanup: cleanup}
	owner.gate <- struct{}{}
	owner.cfg.CleanupError = func(e error) {
		if owner.cleanup != nil {
			fmt.Fprintln(redact.Writer(owner.cleanup), e)
		}
	}
	return owner
}
func (o *rootBudgetOwner) get(ctx context.Context) (*accountbudgetruntime.Factory, error) {
	if o == nil {
		return nil, accountbudget.ErrUnavailable
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-o.gate:
	}
	defer func() { o.gate <- struct{}{} }()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if o.closing {
		return nil, accountbudget.ErrUnavailable
	}
	if o.configErr != nil {
		return nil, o.configErr
	}
	if o.factory != nil {
		return o.factory, nil
	}
	f, e := accountbudgetruntime.Open(ctx, o.cfg)
	if e != nil {
		return nil, e
	}
	if f == nil {
		return nil, accountbudget.ErrUnavailable
	}
	o.factory = f
	return f, nil
}
func (o *rootBudgetOwner) Close() error {
	if o == nil {
		return nil
	}
	<-o.gate
	defer func() { o.gate <- struct{}{} }()
	o.closing = true
	if o.factory == nil {
		return nil
	}
	if e := o.factory.Close(); e != nil {
		return e
	}
	o.factory = nil
	return nil
}

var errMonitorFailed = errors.New("monitor api failed")

// newRootOfficialClient borrows the root's factory; an absent owner preserves legacy OFF callers.
func newRootOfficialClient(ctx context.Context, root *rootOptions, creds official.Credentials, tokenFile string, opts ...official.Option) (*official.Client, error) {
	if root == nil || root.budgetOwner == nil {
		return official.New(creds, tokenFile, opts...), nil
	}
	if ctx == nil {
		return nil, accountbudget.ErrUnavailable
	}
	factory, err := root.budgetOwner.get(ctx)
	if err != nil {
		return nil, err
	}
	return factory.NewClient(creds, tokenFile, opts...)
}
