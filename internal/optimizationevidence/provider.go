// Package optimizationevidence translates the a049 performance dashboard into
// the narrow, transport-neutral evidence contract consumed by a050. It owns no
// performance collector, journal, broker, setting writer, lane, gate, or LIVE
// capability.
package optimizationevidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	strategyopt "github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
)

const (
	MissingCompleteLineage      = "complete-lineage"
	MissingLink                 = "link_missing"
	MissingNotMeasured          = "not_measured"
	MissingMinimumSample        = "minimum-sample"
	MissingRequiredMetricPrefix = "required-metric:"
	MissingDashboardError       = "performance-dashboard-error"
	MissingQueryMismatch        = "performance-query-mismatch"
	MissingSourceTime           = "source-observed-at"
	MissingStale                = "stale-source"
	missingSemanticsVersion     = "semantics-version"
	MaxEvidenceAge              = 72 * time.Hour
)

var requiredMetricKeys = []string{
	"net_pnl",
	"average_r",
	"win_rate",
	"profit_factor",
	"max_drawdown",
	"slippage",
	"mfe",
	"mae",
	"markout_5",
	"markout_15",
	"markout_30",
}

// DashboardReader is the complete authority this adapter accepts. In
// particular it cannot collect, prune, write the journal, or change trading
// state.
type DashboardReader interface {
	Dashboard(context.Context, performance.Query) (performance.DashboardView, error)
}

// Provider applies the server-owned recommendation evidence policy to a fixed
// a049 dashboard query.
type Provider struct {
	reader DashboardReader
	now    func() time.Time
}

func New(reader DashboardReader, now func() time.Time) *Provider {
	return &Provider{reader: reader, now: now}
}

func RequiredMetricKeys() []string {
	return append([]string(nil), requiredMetricKeys...)
}

func (p *Provider) ReadEvidence(ctx context.Context) (strategyopt.Evidence, error) {
	if p == nil || p.reader == nil || p.now == nil {
		return strategyopt.Evidence{
			Status:  strategyopt.EvidenceUnavailable,
			Missing: []string{MissingDashboardError},
		}, errors.New("optimization evidence: dashboard reader and clock are required")
	}
	asOf := p.now().UTC()
	if asOf.IsZero() {
		return strategyopt.Evidence{
			Status:  strategyopt.EvidenceUnavailable,
			Missing: []string{MissingDashboardError},
		}, errors.New("optimization evidence: observation time is required")
	}
	query := performance.DefaultQuery(asOf)
	// A recommendation gate must see link_missing rows in order to refuse them.
	// The operator performance screen remains complete-only; this separate fixed
	// server query cannot be changed by URL or form input.
	query.CompleteOnly = false
	view, err := p.reader.Dashboard(ctx, query)
	if err != nil {
		return strategyopt.Evidence{
			Status: strategyopt.EvidenceUnavailable, ObservedAt: asOf,
			Missing: []string{MissingDashboardError},
		}, fmt.Errorf("optimization evidence: reading fixed performance dashboard: %w", err)
	}
	if !samePerformanceQuery(view.Query, query) {
		return strategyopt.Evidence{
			Status:  strategyopt.EvidenceUnavailable,
			Missing: []string{MissingQueryMismatch},
		}, errors.New("optimization evidence: performance dashboard did not echo the fixed query")
	}

	digest, err := digestDashboard(query, view)
	if err != nil {
		return strategyopt.Evidence{
			Status: strategyopt.EvidenceUnavailable, ObservedAt: asOf,
			Missing: []string{MissingDashboardError},
		}, err
	}
	missing := missingEvidence(view)
	observedAt := view.NewestSourceAt.UTC()
	status := strategyopt.EvidenceComplete
	temporalMissing := ""
	switch {
	case observedAt.IsZero(), observedAt.After(asOf):
		temporalMissing = MissingSourceTime
	case asOf.Sub(observedAt) > MaxEvidenceAge:
		temporalMissing = MissingStale
	}
	if temporalMissing != "" {
		missing = appendMissing(missing, temporalMissing)
		status = strategyopt.EvidenceStale
	} else if len(missing) > 0 {
		status = strategyopt.EvidenceInsufficient
	}
	return strategyopt.Evidence{
		Status: status, Digest: digest, ObservedAt: observedAt, Missing: missing,
	}, nil
}

func samePerformanceQuery(got, want performance.Query) bool {
	return got.AsOf.Equal(want.AsOf) && got.PeriodDays == want.PeriodDays && got.Market == want.Market &&
		got.Lane == want.Lane && got.CompleteOnly == want.CompleteOnly && got.MinimumSample == want.MinimumSample
}

func appendMissing(missing []string, value string) []string {
	for _, existing := range missing {
		if existing == value {
			return missing
		}
	}
	missing = append(missing, value)
	sort.Strings(missing)
	return missing
}

func missingEvidence(view performance.DashboardView) []string {
	missing := make(map[string]struct{})
	add := func(value string) { missing[value] = struct{}{} }
	if view.States.Complete == 0 {
		add(MissingCompleteLineage)
	}
	if view.States.Complete < performance.DefaultMinimumSample ||
		view.States.InsufficientSample > 0 || len(view.Aggregates) == 0 {
		add(MissingMinimumSample)
	}
	if view.States.LinkMissing > 0 {
		add(MissingLink)
	}
	if view.States.NotMeasured > 0 {
		add(MissingNotMeasured)
	}
	for _, aggregate := range view.Aggregates {
		if strings.TrimSpace(aggregate.Market) == "" || strings.TrimSpace(aggregate.LaneID) == "" ||
			strings.TrimSpace(aggregate.LaneVersion) == "" || strings.TrimSpace(aggregate.PolicyID) == "" ||
			strings.TrimSpace(aggregate.PolicyVersion) == "" {
			add(MissingCompleteLineage)
		}
		if aggregate.Status != performance.StatusComplete || aggregate.Samples < performance.DefaultMinimumSample {
			add(MissingMinimumSample)
		}
		if aggregate.SemanticsVersion != performance.SemanticsVersion {
			add(missingSemanticsVersion)
		}
		metrics := make(map[string]performance.MetricSummary, len(aggregate.Metrics))
		duplicates := make(map[string]bool)
		for _, metric := range aggregate.Metrics {
			if _, exists := metrics[metric.Key]; exists {
				duplicates[metric.Key] = true
			}
			metrics[metric.Key] = metric
		}
		for _, key := range requiredMetricKeys {
			metric, exists := metrics[key]
			if !exists || duplicates[key] || metric.Status != performance.StatusComplete ||
				metric.Samples < performance.DefaultMinimumSample || strings.TrimSpace(metric.Value) == "" {
				add(MissingRequiredMetricPrefix + key)
			}
		}
	}
	out := make([]string, 0, len(missing))
	for value := range missing {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

type canonicalDashboard struct {
	Contract       string                  `json:"contract"`
	Query          performance.Query       `json:"query"`
	NewestSourceAt time.Time               `json:"newest_source_at"`
	States         performance.StateCounts `json:"states"`
	Aggregates     []performance.Aggregate `json:"aggregates"`
}

func digestDashboard(query performance.Query, view performance.DashboardView) (string, error) {
	aggregates := append([]performance.Aggregate(nil), view.Aggregates...)
	for i := range aggregates {
		aggregates[i].Metrics = append([]performance.MetricSummary(nil), aggregates[i].Metrics...)
		sort.Slice(aggregates[i].Metrics, func(left, right int) bool {
			a, b := aggregates[i].Metrics[left], aggregates[i].Metrics[right]
			return strings.Join([]string{a.Key, string(a.Status), a.Value, a.Provenance}, "\x00") <
				strings.Join([]string{b.Key, string(b.Status), b.Value, b.Provenance}, "\x00")
		})
	}
	sort.Slice(aggregates, func(left, right int) bool {
		a, b := aggregates[left], aggregates[right]
		return strings.Join([]string{a.Market, a.LaneID, a.LaneVersion, a.PolicyID, a.PolicyVersion}, "\x00") <
			strings.Join([]string{b.Market, b.LaneID, b.LaneVersion, b.PolicyID, b.PolicyVersion}, "\x00")
	})
	raw, err := json.Marshal(canonicalDashboard{
		Contract: "a050-optimization-evidence/v1", Query: query,
		NewestSourceAt: view.NewestSourceAt.UTC(), States: view.States, Aggregates: aggregates,
	})
	if err != nil {
		return "", fmt.Errorf("optimization evidence: canonical dashboard: %w", err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

var _ strategyopt.EvidenceProvider = (*Provider)(nil)
