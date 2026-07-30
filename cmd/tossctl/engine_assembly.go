package main

import (
	"context"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

type engineOptionsDecorator func(*engine.Options)

func assembleEngine(ctx context.Context, root *rootOptions, clk clock.Clock,
	logger *obs.Logger, decorate engineOptionsDecorator, officialOptions ...official.Option,
) (*engine.Context, error) {
	configDir := ""
	if root != nil {
		configDir = root.configDir
	}
	opts := engine.Options{
		ConfigDir:       configDir,
		Clock:           clk,
		Logger:          logger,
		Operator:        engineOperator(),
		OfficialOptions: officialOptions,
		// Guardian is deliberately absent: engine.NewContext owns the resolved
		// account and the one writable journal, so it constructs the production
		// RiskGuardian from those exact instances.
		//
		// Publisher: a nil transport makes every critical alert undeliverable, the
		// outbox row stays PENDING, the entry gate latches and sustained failure
		// tightens the mode. That is the specified direction (risk-management) and
		// it is recorded in exitwiring.go's newNotifier; configuring a transport is
		// an operational setting with an audit trail, which is a change of its own.
	}
	if decorate != nil {
		decorate(&opts)
	}
	return engine.NewContext(ctx, opts)
}
