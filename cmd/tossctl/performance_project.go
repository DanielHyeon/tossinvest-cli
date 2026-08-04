package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performancejournal"
	"github.com/spf13/cobra"
)

const (
	defaultAttributionProjectionInterval = 5 * time.Minute
	minimumAttributionProjectionInterval = time.Minute
	maximumAttributionProjectionInterval = 24 * time.Hour
)

type attributionProjectionResult struct {
	Accounts     int
	EvidenceRows int
	CalculatedAt time.Time
}

func newPerformanceCmd(root *rootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "performance", Short: "Operate the derived lane-performance read model"}
	cmd.AddCommand(newPerformanceProjectAttributionCmd(root))
	return cmd
}

func newPerformanceProjectAttributionCmd(root *rootOptions) *cobra.Command {
	var interval time.Duration
	var once bool
	cmd := &cobra.Command{
		Use:   "project-attribution",
		Short: "Project exact journal lineage into the derived performance database",
		Long: strings.TrimSpace(`
Read exact closed-trade and campaign identifiers from the query-only trading
journal and atomically rebuild the separate performance.db attribution head.
Missing fill, fee, tax or FX evidence remains link_missing/not_measured. This
command has no broker, order, operating-toggle, activation or LIVE capability.`),
		Annotations:  map[string]string{"source": "official", "mutating": "false"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if once {
				interval = 0
			}
			return runPerformanceAttributionProjector(cmd.Context(), cmd.OutOrStdout(), root, interval, time.Now)
		},
	}
	cmd.Flags().DurationVar(&interval, "interval", defaultAttributionProjectionInterval,
		"Projection interval (1m..24h); use --once for one rebuild")
	cmd.Flags().BoolVar(&once, "once", false, "Project once and exit")
	return cmd
}

func runPerformanceAttributionProjector(ctx context.Context, out io.Writer, root *rootOptions, interval time.Duration,
	now func() time.Time,
) error {
	if ctx == nil || out == nil || now == nil {
		return errors.New("performance: attribution projector is unavailable")
	}
	if interval != 0 && (interval < minimumAttributionProjectionInterval || interval > maximumAttributionProjectionInterval) {
		return fmt.Errorf("performance: projection interval %s is outside %s..%s", interval,
			minimumAttributionProjectionInterval, maximumAttributionProjectionInterval)
	}
	journalPath, err := consoleJournalPath(root)
	if err != nil {
		return err
	}
	project := func() error {
		result, err := projectLaneAttributionOnce(ctx, journalPath, now().UTC())
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "lane attribution projected accounts=%d evidence_rows=%d calculated_at=%s\n",
			result.Accounts, result.EvidenceRows, result.CalculatedAt.Format(time.RFC3339Nano))
		return err
	}
	if err := project(); err != nil || interval == 0 {
		return err
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := project(); err != nil {
				return err
			}
		}
	}
}

func projectLaneAttributionOnce(ctx context.Context, journalPath string, calculatedAt time.Time) (attributionProjectionResult, error) {
	result := attributionProjectionResult{CalculatedAt: calculatedAt.UTC()}
	if ctx == nil || strings.TrimSpace(journalPath) == "" || calculatedAt.IsZero() {
		return result, errors.New("performance: attribution projection requires context, journal path and time")
	}
	reader, err := journal.OpenReadOnly(ctx, journal.ReadOnlyOptions{Path: journalPath})
	if err != nil {
		return result, err
	}
	defer reader.Close()
	accounts, err := reader.AccountRefs(ctx)
	if err != nil {
		return result, err
	}
	if len(accounts) == 0 {
		return result, nil
	}
	store, err := performance.Open(filepath.Join(filepath.Dir(journalPath), consolePerformanceDatabaseFileName))
	if err != nil {
		return result, err
	}
	defer store.Close()
	adapter := performancejournal.New(reader)
	windowStart := time.Unix(1, 0).UTC()
	for _, accountRef := range accounts {
		rebuild, err := adapter.AttributionRebuild(ctx, performance.AttributionEvidenceWindow{
			AccountRef: accountRef, ClosedAfter: windowStart, ClosedAtOrBefore: result.CalculatedAt,
		}, attributionProjectionID(accountRef, result.CalculatedAt), result.CalculatedAt)
		if err != nil {
			return result, err
		}
		if err := store.PersistAttributionRebuild(ctx, rebuild); err != nil {
			return result, err
		}
		result.Accounts++
		result.EvidenceRows += len(rebuild.Positions) + len(rebuild.FillDeltas) + len(rebuild.Unavailable)
	}
	return result, nil
}

func attributionProjectionID(accountRef string, calculatedAt time.Time) string {
	sum := sha256.Sum256([]byte("lane-attribution-projector/v1\x00" + strings.TrimSpace(accountRef) + "\x00" +
		calculatedAt.UTC().Format(time.RFC3339Nano)))
	return "lane-attribution:" + hex.EncodeToString(sum[:])
}
