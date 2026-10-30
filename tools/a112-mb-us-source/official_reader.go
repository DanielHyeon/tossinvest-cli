package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudgetruntime"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// officialReader is the only production reader. It intentionally calls only
// the accepted, single-GET M-B0 seam and has no retry, OAuth, account or
// configuration capability of its own.
type officialReader struct{ client *official.Client }

func openProductionDependencies(ctx context.Context, cfg config) (dependencies, *accountbudgetruntime.Factory, error) {
	mode, modeSet := os.LookupEnv("TOSSOS_ACCOUNT_BUDGET_MODE")
	manifest, manifestSet := os.LookupEnv("TOSSOS_ACCOUNT_BUDGET_MANIFEST")
	conf, err := accountbudgetruntime.ParseLaunchConfig(mode, modeSet, manifest, manifestSet)
	if err != nil {
		return dependencies{}, nil, err
	}
	conf.CleanupError = func(error) { fmt.Fprintln(os.Stderr, "budget cleanup failed") }
	factory, err := accountbudgetruntime.Open(ctx, conf)
	if err != nil {
		return dependencies{}, nil, err
	}
	client, err := factory.NewClient(official.Credentials{}, cfg.tokenCache)
	if err != nil {
		return dependencies{}, nil, errors.Join(err, factory.Close())
	}
	return dependencies{reader: officialReader{client: client}, now: time.Now, identity: systemIdentity{}, builder: systemBuildRunner{}}, factory, nil
}

func (r officialReader) Candle(ctx context.Context, before []byte) (sourceResult, error) {
	result, err := official.A112MBUSCandle(accountbudget.WithPriority(ctx, accountbudget.Auxiliary), r.client, before)
	return resultFromOfficial(result), err
}

func (r officialReader) Orderbook(ctx context.Context) (sourceResult, error) {
	result, err := official.A112MBUSOrderbook(accountbudget.WithPriority(ctx, accountbudget.Auxiliary), r.client)
	return resultFromOfficial(result), err
}

func (r officialReader) Calendar(ctx context.Context, date string) (sourceResult, error) {
	result, err := official.A112MBUSCalendar(accountbudget.WithPriority(ctx, accountbudget.Auxiliary), r.client, date)
	return resultFromOfficial(result), err
}

func resultFromOfficial(result official.A112MBUSResult) sourceResult {
	return sourceResult{method: result.Method(), path: result.Path(), query: result.CanonicalQuery(), statusCode: result.StatusCode(), body: result.Body(), cursorJSON: result.CursorJSON(), cursorValue: result.CursorValue(), rateHeaders: result.RateHeaders()}
}
