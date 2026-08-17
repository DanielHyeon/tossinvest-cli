package main

import (
	"context"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// officialReader is the only production reader. It intentionally calls only
// the accepted, single-GET M-B0 seam and has no retry, OAuth, account or
// configuration capability of its own.
type officialReader struct{ client *official.Client }

func newProductionDependencies(cfg config) dependencies {
	client := official.New(official.Credentials{}, cfg.tokenCache)
	return dependencies{reader: officialReader{client: client}, now: time.Now, identity: systemIdentity{}, builder: systemBuildRunner{}}
}

func (r officialReader) Candle(ctx context.Context, before []byte) (sourceResult, error) {
	result, err := official.A112MBUSCandle(ctx, r.client, before)
	return resultFromOfficial(result), err
}

func (r officialReader) Orderbook(ctx context.Context) (sourceResult, error) {
	result, err := official.A112MBUSOrderbook(ctx, r.client)
	return resultFromOfficial(result), err
}

func (r officialReader) Calendar(ctx context.Context, date string) (sourceResult, error) {
	result, err := official.A112MBUSCalendar(ctx, r.client, date)
	return resultFromOfficial(result), err
}

func resultFromOfficial(result official.A112MBUSResult) sourceResult {
	return sourceResult{method: result.Method(), path: result.Path(), query: result.CanonicalQuery(), statusCode: result.StatusCode(), body: result.Body(), cursorJSON: result.CursorJSON(), cursorValue: result.CursorValue(), rateHeaders: result.RateHeaders()}
}
