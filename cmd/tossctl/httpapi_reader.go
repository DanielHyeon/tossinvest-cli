package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/operatorview"
	"github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

const httpAPIPositionFreshness = 30 * time.Second

type httpAPIAccountRef func() (string, error)
type httpAPIAdoptionSettingsRead func() (config.Adoption, string, error)

type httpAPIManagementRuntimeReader interface {
	Runtime(context.Context) (positionpolicy.ManagementRuntime, error)
}

// httpAPIReader is the production adapter for the seven transport reads. Every
// field is a narrow read interface or SELECT-only handle; the API package never
// receives a broker, journal writer, engine controller, or configuration writer.
type httpAPIReader struct {
	now            func() time.Time
	engineMarker   string
	holdings       httpAPIHoldingsReader
	orders         httpAPIOrdersReader
	signals        httpAPISignalsReader
	accountRef     httpAPIAccountRef
	journal        *journal.ReadOnly
	journalErr     error
	performance    performanceDashboardReader
	performanceErr error
	// performanceSource pins the production lineage bridge to journal.ReadOnly.
	// The current REST dashboard reads the already-built performance database;
	// retaining this SELECT-only seam makes future rebuild status observable
	// without ever widening the daemon to a journal writer.
	performanceSource performance.JournalLineageReader
	optimization      *optimization.Store
	adoptionDesired   httpAPIAdoptionSettingsRead
	managementRuntime httpAPIManagementRuntimeReader
	strategyRuntime   httpapi.StrategyRuntimeReader
	positionsCache    httpAPITimedReadCache[httpapi.PositionsResource]
	ordersCache       httpAPITimedReadCache[httpapi.OrdersResource]
}

func (r *httpAPIReader) Engine(context.Context) (httpapi.EngineResource, error) {
	if r == nil || strings.TrimSpace(r.engineMarker) == "" {
		return httpapi.EngineResource{}, errors.New("httpapi: engine marker is unavailable")
	}
	now := r.clockNow()
	status := enginelock.Read(r.engineMarker, now)
	out := httpapi.EngineResource{Status: "stopped", Running: status.Running, Source: "engine-marker"}
	if !status.Running {
		return out, nil
	}
	out.Status = "running"
	out.PID = status.Marker.PID
	if !status.Marker.StartedAt.IsZero() {
		at := status.Marker.StartedAt.UTC()
		out.StartedAt = &at
	}
	if !status.RefreshedAt.IsZero() {
		at := status.RefreshedAt.UTC()
		out.RefreshedAt = &at
	}
	if status.Marker.Binary.Known() {
		at := status.Marker.Binary.ModTime.UTC()
		out.BuildAt = &at
	}
	return out, nil
}

func (r *httpAPIReader) Positions(ctx context.Context) (httpapi.PositionsResource, error) {
	return r.positionsCache.Get(ctx, r.clockNow(), httpAPIBrokerReadTTL, httpAPIBrokerReadErrorTTL, r.readPositions)
}

func (r *httpAPIReader) readPositions(ctx context.Context) (httpapi.PositionsResource, error) {
	if r == nil || r.holdings == nil || r.accountRef == nil {
		return httpapi.PositionsResource{}, errors.New("httpapi: position read is unavailable")
	}
	accountRef, err := r.accountRef()
	if err != nil {
		return httpapi.PositionsResource{}, err
	}
	brokerRows, err := r.holdings.Holdings(ctx, "")
	if err != nil {
		return httpapi.PositionsResource{}, err
	}

	journalRows := []journal.PositionExit{}
	policyByPosition := map[string]positionpolicy.State{}
	journalReadable := r.journal != nil && r.journalErr == nil
	if journalReadable {
		journalRows, err = r.journal.LivePositionExits(ctx, accountRef)
		if err != nil {
			return httpapi.PositionsResource{}, err
		}
		policyStates, policyErr := r.journal.PositionPolicies(ctx)
		if policyErr != nil {
			return httpapi.PositionsResource{}, policyErr
		}
		for _, state := range policyStates {
			policyByPosition[strings.TrimSpace(state.PositionID)] = state
		}
	}
	byKey := make(map[string]journal.PositionExit, len(journalRows))
	for _, row := range journalRows {
		byKey[positionKey(row.Position.Market, row.Position.Symbol)] = row
	}
	runtime := r.readManagementRuntime(ctx)

	now := r.clockNow()
	items := make([]httpapi.Position, 0, len(brokerRows)+len(journalRows))
	seen := make(map[string]struct{}, len(brokerRows))
	for _, broker := range brokerRows {
		key := positionKey(positionMarket(broker), broker.Symbol)
		seen[key] = struct{}{}
		projected := positionFromBroker(attest.Mask(accountRef), broker)
		if stored, ok := byKey[key]; ok {
			applyStoredPosition(&projected, stored, now)
			state, lifecycleKnown := policyByPosition[strings.TrimSpace(stored.Position.ID)]
			managed, released := lifecycleFlags(state, lifecycleKnown)
			if released {
				applyReleasedExitTruth(&projected, stored)
			}
			applyManagementProjection(&projected, lifecycleKnown, managed, released, runtime)
			applyExitLineReference(&projected, &stored, state, lifecycleKnown, runtime)
		} else {
			why := "position is not present in the read-only journal"
			if !journalReadable {
				why = "read-only journal is unavailable"
			}
			projected.ExitLine = httpapi.ExitLineFrom(operatorview.BuildExitLine(operatorview.Source{UnknownReason: why}))
			applyManagementProjection(&projected, journalReadable, false, false, runtime)
			applyExitLineReference(&projected, nil, positionpolicy.State{}, false, runtime)
		}
		items = append(items, projected)
	}
	for _, stored := range journalRows {
		if _, ok := seen[positionKey(stored.Position.Market, stored.Position.Symbol)]; ok {
			continue
		}
		projected := httpapi.Position{AccountLabel: attest.Mask(accountRef), PositionID: stored.Position.ID,
			Market: strings.ToUpper(stored.Position.Market), Symbol: stored.Position.Symbol,
			Quantity: stored.Position.Quantity, AveragePrice: stored.Position.AvgPrice, InJournal: true,
			Eligible: stored.Position.ExitEligible()}
		applyStoredPosition(&projected, stored, now)
		state, lifecycleKnown := policyByPosition[strings.TrimSpace(stored.Position.ID)]
		managed, released := lifecycleFlags(state, lifecycleKnown)
		if released {
			applyReleasedExitTruth(&projected, stored)
		}
		applyManagementProjection(&projected, lifecycleKnown, managed, released, runtime)
		applyExitLineReference(&projected, &stored, state, lifecycleKnown, runtime)
		items = append(items, projected)
	}
	return httpapi.PositionsResource{ObservedAt: pointerTo(now), Source: "official+journal-read-only", Items: items}, nil
}

func (r *httpAPIReader) readManagementRuntime(ctx context.Context) positionpolicy.ManagementRuntime {
	if r == nil || r.managementRuntime == nil {
		return positionpolicy.ManagementRuntime{}
	}
	runtime, err := r.managementRuntime.Runtime(ctx)
	if err != nil {
		return positionpolicy.ManagementRuntime{}
	}
	return runtime
}

func lifecycleFlags(state positionpolicy.State, known bool) (managed, released bool) {
	if !known {
		return false, false
	}
	return state.Status == positionpolicy.StatusManaged && state.ExitEligible,
		state.Status == positionpolicy.StatusReleased && state.Version > 0
}

func applyManagementProjection(out *httpapi.Position, journalKnown, managed, released bool,
	runtime positionpolicy.ManagementRuntime) {
	projection := positionpolicy.ProjectManagement(positionpolicy.ManagementInput{
		Market: out.Market, Symbol: out.Symbol, JournalKnown: journalKnown,
		Managed: managed, Released: released, Runtime: runtime,
	})
	out.AdoptionStatus = httpapi.AdoptionStatus(projection.Status)
	out.StatusKnown = projection.StatusKnown
	out.AdoptionLabel = projection.Label
	out.AdoptionReason = httpapi.AdoptionReason(projection.Reason)
	out.Included, out.Excluded, out.Candidate = projection.Included, projection.Excluded, projection.Candidate
	out.DesignationKnown = projection.DesignationKnown
	out.Eligible = projection.Status == positionpolicy.ManagementStatusManaged
	out.ManagementStatus = strings.ToLower(strings.ReplaceAll(string(projection.Status), "_", "-"))
	out.CoveringBlock = reconcileBlockFrom(projection.Block)
}

func applyReleasedExitTruth(out *httpapi.Position, stored journal.PositionExit) {
	out.ExitLine = httpapi.ExitLineFrom(operatorview.BuildExitLine(operatorview.Source{
		UnknownReason: "operator_released_lifecycle",
	}))
	if hasStoredExitEvidence(stored.Exit) {
		out.StoredExitEvidence = &httpapi.StoredExitEvidence{
			EntryPrice: stored.Exit.EntryPrice, InitialStop: stored.Exit.InitialStop,
			Baseline: stored.Exit.Baseline, HighWater: stored.Exit.HighWater,
			EffectiveKnown: false, Label: "원장 기록 · 실효 미확인",
		}
	}
}

func reconcileBlockFrom(value *positionpolicy.ReconcileBlock) *httpapi.ReconcileBlock {
	if value == nil {
		return nil
	}
	out := &httpapi.ReconcileBlock{Scope: strings.ToUpper(string(value.Scope)),
		Market: strings.ToUpper(strings.TrimSpace(value.Market)), Symbol: strings.ToUpper(strings.TrimSpace(value.Symbol)),
		Reason: value.Reason}
	if !value.StartedAt.IsZero() {
		at := value.StartedAt.UTC()
		out.StartedAt = &at
	}
	return out
}

func positionFromBroker(accountLabel string, value domain.Position) httpapi.Position {
	return httpapi.Position{AccountLabel: accountLabel, Market: positionMarket(value), Symbol: value.Symbol, Name: value.Name,
		Quantity: decimal(value.Quantity), AveragePrice: decimal(value.AveragePrice), LastPrice: decimal(value.CurrentPrice),
		MarketValue: decimal(value.MarketValue), UnrealizedPnL: decimal(value.UnrealizedPnL),
		ProfitRate: decimal(value.ProfitRate), InBroker: true}
}

func applyStoredPosition(out *httpapi.Position, stored journal.PositionExit, asOf time.Time) {
	out.InJournal = true
	out.PositionID = stored.Position.ID
	out.Eligible = stored.Position.ExitEligible()
	if out.Eligible {
		out.ManagementStatus = "managed"
	} else {
		out.ManagementStatus = "unmanaged"
	}
	source := operatorview.Source{UnknownReason: "persisted exit snapshot is absent",
		RemainingQuantity: stored.Position.Quantity, EffectiveSource: "persisted effective snapshot"}
	if stored.HasExit {
		view := stored.Exit.Snapshot.WithFreshness(asOf, httpAPIPositionFreshness)
		source.UnknownReason, source.StaleReason = view.UnknownReason, view.StaleReason
		if view.Snapshot != nil {
			source.Snapshot = &view.Snapshot.Line
			source.ObservationSource = view.Snapshot.ObservationSource
			source.ObservedAt = view.Snapshot.ObservedAt
		} else if hasStoredExitEvidence(stored.Exit) {
			out.StoredExitEvidence = &httpapi.StoredExitEvidence{
				EntryPrice: stored.Exit.EntryPrice, InitialStop: stored.Exit.InitialStop,
				Baseline: stored.Exit.Baseline, HighWater: stored.Exit.HighWater,
				EffectiveKnown: false, Label: "원장 기록 · 실효 미확인",
			}
		}
	}
	out.ExitLine = httpapi.ExitLineFrom(operatorview.BuildExitLine(source))
}

func hasStoredExitEvidence(state journal.ExitState) bool {
	return strings.TrimSpace(state.EntryPrice) != "" || strings.TrimSpace(state.InitialStop) != "" ||
		strings.TrimSpace(state.Baseline) != "" || strings.TrimSpace(state.HighWater) != ""
}

func positionMarket(value domain.Position) string {
	for _, raw := range []string{value.MarketCode, value.MarketType} {
		if trimmed := strings.ToUpper(strings.TrimSpace(raw)); trimmed != "" {
			return trimmed
		}
	}
	return "UNKNOWN"
}

func positionKey(market, symbol string) string {
	return strings.ToUpper(strings.TrimSpace(market)) + ":" + strings.ToUpper(strings.TrimSpace(symbol))
}

func decimal(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func (r *httpAPIReader) Orders(ctx context.Context) (httpapi.OrdersResource, error) {
	return r.ordersCache.Get(ctx, r.clockNow(), httpAPIBrokerReadTTL, httpAPIBrokerReadErrorTTL, r.readOrders)
}

func (r *httpAPIReader) readOrders(ctx context.Context) (httpapi.OrdersResource, error) {
	if r == nil || r.orders == nil {
		return httpapi.OrdersResource{}, errors.New("httpapi: order read is unavailable")
	}
	reading, err := r.orders.Orders(ctx)
	if err != nil {
		return httpapi.OrdersResource{}, err
	}
	if reading.OpenError != "" || reading.ClosedError != "" || reading.ConditionalError != "" {
		return httpapi.OrdersResource{}, errors.New("httpapi: one or more bounded order reads are unavailable")
	}
	items := make([]httpapi.Order, 0, len(reading.Open)+len(reading.Closed)+len(reading.Conditional))
	for _, row := range slices.Concat(reading.Open, reading.Closed) {
		items = append(items, httpapi.Order{ID: row.ID, AccountLabel: attest.Mask(reading.AccountRef), Market: row.Market,
			Symbol: row.Symbol, Side: row.Side, Kind: row.Kind, Status: row.Status, Currency: row.Currency,
			Quantity: row.Quantity, Price: row.Price, FilledQuantity: row.FilledQuantity,
			AverageFilledPrice: row.AverageFilledPrice, OrderedAt: row.OrderedAt, CanceledAt: row.CanceledAt,
			Origin: "plain"})
	}
	for _, row := range reading.Conditional {
		items = append(items, httpapi.Order{ID: row.ID, AccountLabel: attest.Mask(reading.AccountRef), Market: row.Market,
			Symbol: row.Symbol, Kind: row.Kind + "/" + row.ConditionKind, Status: row.Status,
			Quantity: row.Quantity, Price: row.OrderPrice, OrderedAt: row.CreatedAt, Origin: "conditional"})
	}
	now := r.clockNow()
	return httpapi.OrdersResource{ObservedAt: &now, Source: "official-bounded", Items: items,
		Stale: reading.OpenTruncated || reading.ConditionalTruncated}, nil
}

func (r *httpAPIReader) Candidates(ctx context.Context) (httpapi.CandidatesResource, error) {
	if r == nil || r.signals == nil {
		return httpapi.CandidatesResource{}, errors.New("httpapi: candidate read is unavailable")
	}
	reading, err := r.signals.Signals(ctx)
	if err != nil {
		return httpapi.CandidatesResource{}, err
	}
	items := []httpapi.Candidate{}
	degraded := false
	for _, market := range reading.Markets {
		if strings.TrimSpace(market.Why) != "" {
			return httpapi.CandidatesResource{}, errors.New("httpapi: candidate market assessment is unavailable")
		}
		degraded = degraded || market.Panel.Degraded
		for _, verdict := range market.Verdicts {
			state := "unmeasured"
			switch {
			case verdict.Chase.Vetoed():
				state = "vetoed"
			case verdict.Chase.Passed():
				state = "eligible"
			}
			reasons := make([]string, 0, len(verdict.Chase.Raised())+len(verdict.Chase.NotMeasured()))
			for _, code := range verdict.Chase.Raised() {
				reasons = append(reasons, "raised:"+string(code))
			}
			for _, code := range verdict.Chase.NotMeasured() {
				reasons = append(reasons, "unmeasured:"+string(code))
			}
			first, last := verdict.Summary.FirstSeenAt.UTC(), verdict.Summary.LastSeenAt.UTC()
			items = append(items, httpapi.Candidate{Market: verdict.Summary.Market, Symbol: verdict.Summary.Symbol,
				Verdict: state, ReasonCodes: reasons, FirstSeenAt: &first, LastObservedAt: &last,
				ThresholdVersion: "unapproved"})
		}
	}
	at := reading.At.UTC()
	return httpapi.CandidatesResource{ObservedAt: &at, Stale: degraded, Source: "candidate-store-assessment", Items: items}, nil
}

func (r *httpAPIReader) Performance(ctx context.Context) (performance.DashboardView, error) {
	if r == nil || r.performance == nil {
		if r != nil && r.performanceErr != nil {
			return performance.DashboardView{}, r.performanceErr
		}
		return performance.DashboardView{}, errors.New("httpapi: performance read is unavailable")
	}
	view, err := r.performance.Dashboard(ctx, performance.DefaultQuery(r.clockNow()))
	if err != nil {
		return performance.DashboardView{}, err
	}
	if r.accountRef == nil {
		return view, nil
	}
	accountRef, err := r.accountRef()
	if err != nil {
		return performance.DashboardView{}, err
	}
	for _, market := range []string{"KR", "US"} {
		rows, readErr := r.performance.AttributionRows(ctx, accountRef,
			performance.AttributionQuery{Market: market, IncludeLinkMissing: true}, performance.MaxAttributionQueryRows/2)
		if errors.Is(readErr, performance.ErrAttributionUnavailable) {
			continue
		}
		if readErr != nil {
			return performance.DashboardView{}, readErr
		}
		view.Attributions = append(view.Attributions, rows...)
	}
	return view, nil
}

func (r *httpAPIReader) Settings(ctx context.Context) (httpapi.SettingsResource, error) {
	view, err := r.Optimization(ctx)
	if err != nil {
		return httpapi.SettingsResource{}, err
	}
	projection := httpapi.OptimizationFrom(view)
	items := make([]httpapi.Setting, 0, len(projection.Fields))
	for _, field := range projection.Fields {
		items = append(items, httpapi.Setting{Key: field.Key, Label: field.Label, Description: field.Description,
			Unit: field.Unit, Default: field.Default, Desired: field.Desired, Effective: field.Effective,
			ApplyTiming: field.ApplyTiming, Safety: field.SafetyDirection, Provenance: field.Provenance})
	}
	return httpapi.SettingsResource{Version: projection.Version, EffectiveVersion: projection.EffectiveVersion, Items: items}, nil
}

func (r *httpAPIReader) Optimization(ctx context.Context) (httpapi.OptimizationRead, error) {
	if r == nil || r.optimization == nil {
		return httpapi.OptimizationRead{}, errors.New("httpapi: optimization read is unavailable")
	}
	base, err := r.optimization.Read(ctx)
	if err != nil {
		return httpapi.OptimizationRead{}, err
	}
	if r.adoptionDesired == nil {
		return httpapi.OptimizationRead{}, errors.New("httpapi: desired adoption settings are unavailable")
	}
	desired, rejected, err := r.adoptionDesired()
	if err != nil {
		return httpapi.OptimizationRead{}, err
	}
	desiredSettings := positionpolicy.NewAdoptionSettings(desired.Enabled, desired.DefaultStopPct,
		desired.IncludeSymbols, desired.ExcludeSymbols, rejected)
	runtime := r.readManagementRuntime(ctx)
	actual := httpapi.PositionManagementActual{Desired: adoptionSettingsFrom(desiredSettings),
		EffectiveKnown: runtime.EffectiveKnown, BlockSource: runtime.BlockSource}
	if runtime.EffectiveKnown {
		effective := adoptionSettingsFrom(runtime.Effective)
		actual.Effective = &effective
	}
	return httpapi.OptimizationRead{Core: base, PositionManagement: actual}, nil
}

func adoptionSettingsFrom(value positionpolicy.AdoptionSettings) httpapi.AdoptionSettings {
	return httpapi.AdoptionSettings{Enabled: value.Enabled, DefaultStopPct: value.DefaultStopPct,
		IncludeSymbols: append([]string(nil), value.IncludeSymbols...),
		ExcludeSymbols: append([]string(nil), value.ExcludeSymbols...), Rejected: value.Rejected}
}

func (r *httpAPIReader) Snapshot(ctx context.Context) ([]byte, error) {
	engine, err := r.Engine(ctx)
	if err != nil {
		return nil, err
	}
	positions, err := r.Positions(ctx)
	if err != nil {
		return nil, err
	}
	orders, err := r.Orders(ctx)
	if err != nil {
		return nil, err
	}
	candidates, err := r.Candidates(ctx)
	if err != nil {
		return nil, err
	}
	performanceView, err := r.Performance(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := r.Settings(ctx)
	if err != nil {
		return nil, err
	}
	optimizationView, err := r.Optimization(ctx)
	if err != nil {
		return nil, err
	}
	strategyRuntime := strategyprojection.DormantSnapshot(r.clockNow())
	if r.strategyRuntime != nil {
		strategyRuntime, err = r.strategyRuntime.Read(ctx)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(struct {
		SchemaVersion   string                       `json:"schemaVersion"`
		GeneratedAt     time.Time                    `json:"generatedAt"`
		Engine          httpapi.EngineResource       `json:"engine"`
		Positions       httpapi.PositionsResource    `json:"positions"`
		Orders          httpapi.OrdersResource       `json:"orders"`
		Candidates      httpapi.CandidatesResource   `json:"candidates"`
		Performance     httpapi.PerformanceResource  `json:"performance"`
		Settings        httpapi.SettingsResource     `json:"settings"`
		Optimization    httpapi.OptimizationResource `json:"optimization"`
		StrategyRuntime strategyprojection.Snapshot  `json:"strategyRuntime"`
	}{httpapi.SchemaVersion, r.clockNow(), engine, positions, orders, candidates,
		httpapi.PerformanceFrom(performanceView), settings, httpapi.OptimizationFrom(optimizationView), strategyRuntime})
}

func (r *httpAPIReader) clockNow() time.Time {
	if r != nil && r.now != nil {
		return r.now().UTC()
	}
	return time.Now().UTC()
}

func pointerTo(value time.Time) *time.Time {
	value = value.UTC()
	return &value
}

func (r *httpAPIReader) validate() error {
	if r == nil || r.holdings == nil || r.orders == nil || r.signals == nil || r.accountRef == nil ||
		r.optimization == nil || r.adoptionDesired == nil {
		return fmt.Errorf("httpapi: incomplete production read adapter")
	}
	return nil
}
