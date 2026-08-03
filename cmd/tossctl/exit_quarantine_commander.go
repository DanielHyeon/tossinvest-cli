package main

// exit_quarantine_commander.go extends the console's position-policy commander
// with change a063's quarantine release, without widening
// positionPolicyLifecycleClient.
//
// The lifecycle client is an interface the policy tests fake in several places,
// none of which is about quarantines; widening it would break all of them to
// express a capability they do not have. Asserting for it instead means an
// engine that predates a063 — or any fake that does not implement it — yields
// ErrUnwired, which the console draws as "no release button" rather than as a
// failure.

import (
	"context"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitquarantine"
)

type exitQuarantineClient interface {
	Quarantines(context.Context) ([]exitquarantine.Row, error)
	PreviewQuarantineRelease(context.Context, exitquarantine.Request) (exitquarantine.Preview, error)
	ReleaseQuarantine(context.Context, exitquarantine.ApplyRequest) (exitquarantine.Result, error)
}

func (c *consolePositionPolicyCommander) quarantineClient() (exitQuarantineClient, error) {
	client, ok := c.lifecycle.(exitQuarantineClient)
	if !ok {
		return nil, exitquarantine.ErrUnwired
	}
	return client, nil
}

func (c *consolePositionPolicyCommander) Quarantines(ctx context.Context) ([]exitquarantine.Row, error) {
	client, err := c.quarantineClient()
	if err != nil {
		return nil, err
	}
	return client.Quarantines(ctx)
}

func (c *consolePositionPolicyCommander) PreviewQuarantineRelease(ctx context.Context,
	req exitquarantine.Request) (exitquarantine.Preview, error) {
	client, err := c.quarantineClient()
	if err != nil {
		return exitquarantine.Preview{}, err
	}
	return client.PreviewQuarantineRelease(ctx, req)
}

func (c *consolePositionPolicyCommander) ReleaseQuarantine(ctx context.Context,
	req exitquarantine.ApplyRequest) (exitquarantine.Result, error) {
	client, err := c.quarantineClient()
	if err != nil {
		return exitquarantine.Result{}, err
	}
	return client.ReleaseQuarantine(ctx, req)
}
