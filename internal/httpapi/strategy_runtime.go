package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

type StrategyRuntimeReader interface {
	Read(context.Context) (strategyprojection.Snapshot, error)
}

func StrategyRuntimeSnapshotFunc(reader StrategyRuntimeReader, now func() time.Time) SnapshotFunc {
	return func(ctx context.Context) ([]byte, error) {
		if reader == nil || now == nil {
			return nil, errors.New("httpapi: strategy runtime stream reader unavailable")
		}
		snapshot, err := reader.Read(ctx)
		if err != nil {
			return nil, err
		}
		if err := strategyprojection.Validate(snapshot); err != nil {
			return nil, err
		}
		return json.Marshal(Envelope{SchemaVersion: SchemaVersion, Resource: "strategy-runtime",
			GeneratedAt: now().UTC(), Data: strategyprojection.Clone(snapshot)})
	}
}
