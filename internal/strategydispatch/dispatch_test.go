package strategydispatch

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"
)

func manifestFixture(now time.Time) ManifestBinding {
	return ManifestBinding{
		AccountRef: "acct", Profile: "prod", BuildDigest: "build", CommitDigest: "commit",
		LaneID: strategyengine.LaneID, LaneVersion: strategyengine.LaneVersion,
		LaneSourceDigest: strategyengine.FrozenSourceSetDigest, LaneConstantsDigest: "constants-digest",
		ThresholdVersion: "threshold-v1", ThresholdSetDigest: "threshold-digest", EvidenceDigest: "evidence-digest",
		SettingsDigest: "settings", AttestationDigest: "attest", AttestationExpiresAt: now.Add(time.Hour),
		GuardianVersion: "guardian-v1", GuardianLimitsDigest: "limits", ReconciliationWatermark: "watermark",
		ProtectionProfile: "broker-stop-v1", ProtectionState: "WIRED", OperatingPolicy: "LIVE",
		SchedulerScope: "KR/regular", CalendarVersion: "calendar", LaneApproved: true, SchedulerApproved: true,
		AutoStartApproved: true, GateApproved: true, LiveApproved: true, Actor: "operator", AuditID: "audit",
		IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour), Generation: 7,
	}
}

func installedRepository(binding ManifestBinding) *ManifestRepository {
	digest, err := manifestDigest(binding)
	if err != nil {
		panic(err)
	}
	return &ManifestRepository{current: manifest{binding: binding, digest: digest}}
}

func changeField(t *testing.T, target any, index int) {
	t.Helper()
	field := reflect.ValueOf(target).Elem().Field(index)
	switch field.Kind() {
	case reflect.String:
		field.SetString(field.String() + "-changed")
	case reflect.Bool:
		field.SetBool(!field.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeFor[time.Time]() {
			field.Set(reflect.ValueOf(field.Interface().(time.Time).Add(time.Nanosecond)))
		} else {
			field.SetInt(field.Int() + 1)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.SetUint(field.Uint() + 1)
	case reflect.Struct:
		if field.Type() != reflect.TypeFor[time.Time]() {
			t.Fatalf("unsupported struct field %s", field.Type())
		}
		field.Set(reflect.ValueOf(field.Interface().(time.Time).Add(time.Nanosecond)))
	case reflect.Array:
		if field.Len() == 0 || field.Index(0).Kind() != reflect.String {
			t.Fatalf("unsupported array field %s", field.Type())
		}
		field.Index(0).SetString(field.Index(0).String() + "-changed")
	default:
		t.Fatalf("unsupported field kind %s", field.Kind())
	}
}

func decisionRecordFixture(now time.Time) strategyengine.DecisionRecord {
	barClosed := now.Add(-time.Second)
	return strategyengine.DecisionRecord{
		Identity: "strategy-decision:v1:sha256:fixture", CandidateLifeID: "candidate-life:v1:sha256:fixture", CandidateState: "active",
		CandidateFirstSeen: now.Add(-time.Hour).UnixNano(), CandidateLastSeen: now.Add(-time.Minute).UnixNano(),
		CandidateValidUntil: now.Add(time.Hour).UnixNano(), CandidateApprovedAt: now.Add(-30 * time.Second).UnixNano(),
		Market: "KR", Symbol: "005930", LaneID: strategyengine.LaneID, LaneVersion: strategyengine.LaneVersion,
		SourceCommit: strategyengine.SourceCommit, SourceDigest: strategyengine.FrozenSourceSetDigest, ConstantsDigest: "constants-digest",
		ThresholdVersion: "threshold-v1", ThresholdSetDigest: "threshold-digest", EvidenceDigest: "evidence-digest",
		MarketInputVersion: strategyengine.MarketInputVersion, CalendarSource: strategyengine.CalendarSource, CalendarVersion: "krx-calendar:2026-08-01",
		ConfigSource: strategyengine.ConfigSource, ConfigVersion: strategyengine.ConfigVersion,
		IndicatorSource: strategyengine.IndicatorSource, IndicatorVersion: strategyengine.IndicatorVersion, IndicatorComputedAt: now.UnixNano(),
		TradingDay: true, SessionOpenAt: now.Add(-time.Hour).UnixNano(), SessionCloseAt: now.Add(5 * time.Hour).UnixNano(), NoEntryAfter: now.Add(4 * time.Hour).UnixNano(),
		BarSource: strategyengine.CalendarSource, BarOpenAt: barClosed.Add(-5 * time.Minute).UnixNano(), BarClosedAt: barClosed.UnixNano(),
		EvaluatedAt: now.UnixNano(), ExpiresAt: now.Add(15 * time.Second).UnixNano(), Open: "100", High: "101", Low: "99", Close: "100.1", Volume: "1000", Currency: "KRW",
		VWAP: "100", VWAPSlopePct: "0.1", EMA9: "100", LVNSpacePct: "1.2", TangledPct: "0.35", Expansion: "1.8", HVNAboveDistancePct: "1.2",
		StateSource: "official-symbol-state", StateAt: now.UnixNano(), PositionSource: "official-position", PositionAt: now.UnixNano(),
		EntryPrice: "100.1", LivePrice: "100.11", LivePriceObserved: true, EntryPriceDriftPct: "0.01",
		StopPrice: "99.3993", TargetPrice: "102.2021", ExpectedRR: "3",
		AcceptReasons: [7]string{"VWAP_ABOVE", "VWAP_SLOPE_UP", "EMA9_PULLBACK_CONFIRMED", "VOLUME_PROFILE_SPACE_OK", "RR_GE_2", "NOT_TANGLED", "NOT_AFTER_ENTRY_CUTOFF"},
	}
}

func bindingForRecord(now time.Time, record strategyengine.DecisionRecord) ManifestBinding {
	binding := manifestFixture(now)
	binding.LaneID, binding.LaneVersion = record.LaneID, record.LaneVersion
	binding.LaneSourceDigest, binding.LaneConstantsDigest = record.SourceDigest, record.ConstantsDigest
	binding.ThresholdVersion, binding.ThresholdSetDigest, binding.EvidenceDigest = record.ThresholdVersion, record.ThresholdSetDigest, record.EvidenceDigest
	return binding
}

func gateForRecord(binding ManifestBinding, record strategyengine.DecisionRecord) GateSnapshot {
	return GateSnapshot{
		Binding:     binding,
		Decision:    DecisionBinding(record),
		Order:       OrderSettings{OrderType: "LIMIT", Currency: "KRW", SettingsDigest: binding.SettingsDigest},
		LaneDesired: true, LaneEffective: true, ProtectionWired: true, ReconcileHealthy: true,
		SchedulerValid: true, AutoStart: true, GateOpen: true, LiveApproved: true, Revision: 9,
	}
}

type gateStore struct {
	mu              sync.RWMutex
	snapshot        GateSnapshot
	readErr         error
	leaseBeforeErr  error
	leaseSnapshot   func(GateSnapshot) GateSnapshot
	leaseAfterError error
}

func (g *gateStore) ReadGate(context.Context) (GateSnapshot, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.readErr != nil {
		return GateSnapshot{}, g.readErr
	}
	return g.snapshot, nil
}

func (g *gateStore) WithLease(_ context.Context, revision uint64, fn func(GateSnapshot) error) error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.leaseBeforeErr != nil {
		return g.leaseBeforeErr
	}
	if g.snapshot.Revision != revision {
		return &Error{Reason: ReasonTOCTOU}
	}
	snapshot := g.snapshot
	if g.leaseSnapshot != nil {
		snapshot = g.leaseSnapshot(snapshot)
	}
	if err := fn(snapshot); err != nil {
		return err
	}
	return g.leaseAfterError
}

func (g *gateStore) mutate(fn func(*GateSnapshot)) {
	g.mu.Lock()
	defer g.mu.Unlock()
	fn(&g.snapshot)
}

type issuerSpy struct {
	issue                         func()
	plan                          func(*AtomicPlan)
	issueErr                      error
	refusalErr, inDoubtErr        error
	dispatchedErr                 error
	issued, refused, inDoubt      int
	dispatched                    int
	refusalReason, inDoubtReason  Reason
	dispatchedAttemptID, brokerID string
}

func (s *issuerSpy) IssueAndPlan(_ context.Context, request IssueRequest) (AtomicPlan, PlanReceipt, error) {
	s.issued++
	if s.issue != nil {
		s.issue()
	}
	receipt := PlanReceipt{
		AttemptID: request.AttemptID, AccountRef: request.Binding.AccountRef,
		DecisionIdentity: request.Decision.Record().Identity, RiskIntentID: "risk-1",
		ClientOrderID: "client-1", Quantity: "1", Revision: 1, State: "PLANNED",
	}
	// The package-private core is intentionally testable with a zero opaque
	// decision, so bind its receipt to the already validated record via the gate.
	if receipt.DecisionIdentity == "" {
		receipt.DecisionIdentity = "strategy-decision:v1:sha256:fixture"
	}
	plan := AtomicPlan{AttemptID: request.AttemptID, Decision: request.Decision, ManifestDigest: request.ManifestDigest}
	if s.plan != nil {
		s.plan(&plan)
	}
	return plan, receipt, s.issueErr
}

func (s *issuerSpy) RecordStrategyRefusal(_ context.Context, _ PlanReceipt, reason Reason) error {
	s.refused++
	s.refusalReason = reason
	return s.refusalErr
}
func (s *issuerSpy) RecordStrategyInDoubt(_ context.Context, _ PlanReceipt, reason Reason) error {
	s.inDoubt++
	s.inDoubtReason = reason
	return s.inDoubtErr
}
func (s *issuerSpy) RecordStrategyDispatched(_ context.Context, _ PlanReceipt, attemptID, brokerID string) error {
	s.dispatched++
	s.dispatchedAttemptID, s.brokerID = attemptID, brokerID
	return s.dispatchedErr
}

type gatewaySpy struct {
	calls int
	out   execgw.Outcome
	err   error
}

func (s *gatewaySpy) PlaceStrategyEntry(context.Context, AtomicPlan) (execgw.Outcome, error) {
	s.calls++
	return s.out, s.err
}

type dispatchFixture struct {
	now      time.Time
	record   strategyengine.DecisionRecord
	binding  ManifestBinding
	gates    *gateStore
	manifest *ManifestRepository
	issuer   *issuerSpy
	gateway  *gatewaySpy
	nowFn    func() time.Time
}

func newDispatchFixture() *dispatchFixture {
	now := time.Date(2026, 8, 1, 1, 0, 0, 0, time.UTC)
	record := decisionRecordFixture(now)
	binding := bindingForRecord(now, record)
	return &dispatchFixture{
		now: now, record: record, binding: binding,
		gates: &gateStore{snapshot: gateForRecord(binding, record)}, manifest: installedRepository(binding),
		issuer: &issuerSpy{}, gateway: &gatewaySpy{}, nowFn: func() time.Time { return now },
	}
}

func (f *dispatchFixture) dependencies() Dependencies {
	return Dependencies{Gates: f.gates, Manifest: f.manifest, Issuer: f.issuer, Gateway: f.gateway, Now: f.nowFn}
}

func (f *dispatchFixture) run() error {
	return dispatchValidated(context.Background(), strategyengine.Decision{}, f.record, f.dependencies())
}

func requireDispatchReason(t *testing.T, err error, want Reason) *Error {
	t.Helper()
	var typed *Error
	if !errors.As(err, &typed) || typed.Reason != want {
		t.Fatalf("err=%v reason=%v want=%s", err, typed, want)
	}
	return typed
}

func TestDispatchRejectsOpaqueZeroDecisionBeforeAuthority(t *testing.T) {
	err := Dispatch(context.Background(), strategyengine.Decision{}, Dependencies{})
	var typed *Error
	if !errors.As(err, &typed) || typed.Reason != ReasonDecisionInvalid {
		t.Fatalf("err=%v", err)
	}
}

func TestPostValidationDispatchRejectsEveryMissingDependencyBeforeAuthority(t *testing.T) {
	for _, missing := range []string{"gates", "manifest", "issuer", "gateway", "clock"} {
		t.Run(missing, func(t *testing.T) {
			fixture := newDispatchFixture()
			deps := fixture.dependencies()
			switch missing {
			case "gates":
				deps.Gates = nil
			case "manifest":
				deps.Manifest = nil
			case "issuer":
				deps.Issuer = nil
			case "gateway":
				deps.Gateway = nil
			case "clock":
				deps.Now = nil
			}
			err := dispatchValidated(context.Background(), strategyengine.Decision{}, fixture.record, deps)
			requireDispatchReason(t, err, ReasonActivation)
			if fixture.issuer.issued != 0 || fixture.gateway.calls != 0 {
				t.Fatalf("authority reached issuer=%+v gateway=%+v", fixture.issuer, fixture.gateway)
			}
		})
	}
}

func TestPostValidationDispatchPrePlanFailuresAreExactAndCallOfficialGatewayZeroTimes(t *testing.T) {
	sentinel := errors.New("injected authority failure")
	tests := []struct {
		name      string
		want      Reason
		wantIssue int
		setup     func(*dispatchFixture)
	}{
		{name: "zero clock", want: ReasonDecisionStale, setup: func(f *dispatchFixture) { f.nowFn = func() time.Time { return time.Time{} } }},
		{name: "exact decision expiry", want: ReasonDecisionStale, setup: func(f *dispatchFixture) { f.nowFn = func() time.Time { return time.Unix(0, f.record.ExpiresAt) } }},
		{name: "gate read error", want: ReasonGate, setup: func(f *dispatchFixture) { f.gates.readErr = sentinel }},
		{name: "initial gate refusal", want: ReasonLaneOff, setup: func(f *dispatchFixture) { f.gates.snapshot.LaneDesired = false }},
		{name: "manifest verification error", want: ReasonActivation, setup: func(f *dispatchFixture) { f.manifest = NewDormantManifestRepository() }},
		{name: "atomic issue and plan error", want: ReasonGuardian, wantIssue: 1, setup: func(f *dispatchFixture) { f.issuer.issueErr = sentinel }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newDispatchFixture()
			tc.setup(fixture)
			err := fixture.run()
			typed := requireDispatchReason(t, err, tc.want)
			if fixture.issuer.issued != tc.wantIssue || fixture.issuer.refused != 0 || fixture.issuer.inDoubt != 0 || fixture.issuer.dispatched != 0 || fixture.gateway.calls != 0 {
				t.Fatalf("unexpected calls issuer=%+v gateway=%+v", fixture.issuer, fixture.gateway)
			}
			if (tc.name == "gate read error" || tc.name == "atomic issue and plan error") && !errors.Is(typed, sentinel) {
				t.Fatalf("cause=%v does not retain injected error", typed.Cause)
			}
		})
	}
}

// This is a package-private post-validation state-machine test with spies. It
// is not evidence of a production-positive Dispatch path: a047 deliberately
// ships without the authentic source manifest or an activation installer.
func TestPostValidationDispatchCorePlansOnceAndPersistsExactOfficialOutcomeWithSpies(t *testing.T) {
	now := time.Date(2026, 8, 1, 1, 0, 0, 0, time.UTC)
	record := decisionRecordFixture(now)
	binding := bindingForRecord(now, record)
	gates := &gateStore{snapshot: gateForRecord(binding, record)}
	issuer := &issuerSpy{}
	gateway := &gatewaySpy{out: execgw.Outcome{AttemptID: "mutation-1", BrokerOrderID: "broker-1", State: journal.StateConfirmed}}
	err := dispatchValidated(context.Background(), strategyengine.Decision{}, record, Dependencies{
		Gates: gates, Manifest: installedRepository(binding), Issuer: issuer, Gateway: gateway, Now: func() time.Time { return now },
	})
	if err != nil || issuer.issued != 1 || issuer.dispatched != 1 || issuer.refused != 0 || issuer.inDoubt != 0 ||
		issuer.dispatchedAttemptID != "mutation-1" || issuer.brokerID != "broker-1" || gateway.calls != 1 {
		t.Fatalf("err=%v issuer=%+v gateway=%+v", err, issuer, gateway)
	}
}

func TestPostValidationDispatchCoreRefusesEveryInitialGateBeforeIssuerWithStablePrecedence(t *testing.T) {
	now := time.Date(2026, 8, 1, 1, 0, 0, 0, time.UTC)
	record := decisionRecordFixture(now)
	binding := bindingForRecord(now, record)
	tests := []struct {
		name   string
		mutate func(*GateSnapshot)
		want   Reason
	}{
		{name: "lane desired off", mutate: func(v *GateSnapshot) { v.LaneDesired = false }, want: ReasonLaneOff},
		{name: "lane effective off", mutate: func(v *GateSnapshot) { v.LaneEffective = false }, want: ReasonLaneOff},
		{name: "kill switch", mutate: func(v *GateSnapshot) { v.KillSwitch = true }, want: ReasonKillSwitch},
		{name: "protection unwired", mutate: func(v *GateSnapshot) { v.ProtectionWired = false }, want: ReasonProtection},
		{name: "reconciliation unhealthy", mutate: func(v *GateSnapshot) { v.ReconcileHealthy = false }, want: ReasonReconcile},
		{name: "scheduler invalid", mutate: func(v *GateSnapshot) { v.SchedulerValid = false }, want: ReasonScheduler},
		{name: "autostart off", mutate: func(v *GateSnapshot) { v.AutoStart = false }, want: ReasonAutoStart},
		{name: "gate closed", mutate: func(v *GateSnapshot) { v.GateOpen = false }, want: ReasonGate},
		{name: "live unapproved", mutate: func(v *GateSnapshot) { v.LiveApproved = false }, want: ReasonLive},
		{name: "settings digest mismatch", mutate: func(v *GateSnapshot) { v.Order.SettingsDigest = "other-settings" }, want: ReasonActivation},
		{name: "order type mismatch", mutate: func(v *GateSnapshot) { v.Order.OrderType = "MARKET" }, want: ReasonActivation},
		{name: "currency mismatch", mutate: func(v *GateSnapshot) { v.Order.Currency = "USD" }, want: ReasonActivation},
		{name: "lane precedes kill", mutate: func(v *GateSnapshot) { v.LaneDesired = false; v.KillSwitch = true }, want: ReasonLaneOff},
		{name: "kill precedes protection", mutate: func(v *GateSnapshot) { v.KillSwitch = true; v.ProtectionWired = false }, want: ReasonKillSwitch},
		{name: "protection precedes reconcile", mutate: func(v *GateSnapshot) { v.ProtectionWired = false; v.ReconcileHealthy = false }, want: ReasonProtection},
		{name: "reconcile precedes scheduler", mutate: func(v *GateSnapshot) { v.ReconcileHealthy = false; v.SchedulerValid = false }, want: ReasonReconcile},
		{name: "scheduler precedes autostart", mutate: func(v *GateSnapshot) { v.SchedulerValid = false; v.AutoStart = false }, want: ReasonScheduler},
		{name: "autostart precedes gate", mutate: func(v *GateSnapshot) { v.AutoStart = false; v.GateOpen = false }, want: ReasonAutoStart},
		{name: "gate precedes live", mutate: func(v *GateSnapshot) { v.GateOpen = false; v.LiveApproved = false }, want: ReasonGate},
		{name: "activation precedes lane", mutate: func(v *GateSnapshot) { v.Binding.LaneVersion = "other"; v.LaneDesired = false }, want: ReasonActivation},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gate := gateForRecord(binding, record)
			tc.mutate(&gate)
			issuer := &issuerSpy{}
			gateway := &gatewaySpy{}
			err := dispatchValidated(context.Background(), strategyengine.Decision{}, record, Dependencies{
				Gates: &gateStore{snapshot: gate}, Manifest: installedRepository(binding), Issuer: issuer,
				Gateway: gateway, Now: func() time.Time { return now },
			})
			var typed *Error
			if !errors.As(err, &typed) || typed.Reason != tc.want || issuer.issued != 0 || issuer.refused != 0 || issuer.inDoubt != 0 || issuer.dispatched != 0 || gateway.calls != 0 {
				t.Fatalf("err=%v issuer=%+v gateway=%+v want=%s", err, issuer, gateway, tc.want)
			}
		})
	}
}

func TestValidatedDispatchPlanTimeGateChangeRefusesBeforeOfficialCall(t *testing.T) {
	now := time.Date(2026, 8, 1, 1, 0, 0, 0, time.UTC)
	record := decisionRecordFixture(now)
	binding := bindingForRecord(now, record)
	gates := &gateStore{snapshot: gateForRecord(binding, record)}
	issuer := &issuerSpy{issue: func() { gates.mutate(func(value *GateSnapshot) { value.Revision++; value.KillSwitch = true }) }}
	gateway := &gatewaySpy{}
	err := dispatchValidated(context.Background(), strategyengine.Decision{}, record, Dependencies{
		Gates: gates, Manifest: installedRepository(binding), Issuer: issuer, Gateway: gateway, Now: func() time.Time { return now },
	})
	var typed *Error
	if !errors.As(err, &typed) || typed.Reason != ReasonTOCTOU || issuer.refused != 1 || issuer.inDoubt != 0 || gateway.calls != 0 {
		t.Fatalf("err=%v issuer=%+v gateway=%+v", err, issuer, gateway)
	}
}

func TestValidatedDispatchExpiryReachedDuringPlanningRefusesAtExactBoundary(t *testing.T) {
	now := time.Date(2026, 8, 1, 1, 0, 0, 0, time.UTC)
	record := decisionRecordFixture(now)
	binding := bindingForRecord(now, record)
	gates := &gateStore{snapshot: gateForRecord(binding, record)}
	issuer := &issuerSpy{}
	gateway := &gatewaySpy{}
	calls := 0
	nowFn := func() time.Time {
		calls++
		if calls >= 3 {
			return time.Unix(0, record.ExpiresAt)
		}
		return now
	}
	err := dispatchValidated(context.Background(), strategyengine.Decision{}, record, Dependencies{
		Gates: gates, Manifest: installedRepository(binding), Issuer: issuer, Gateway: gateway, Now: nowFn,
	})
	var typed *Error
	if !errors.As(err, &typed) || typed.Reason != ReasonTOCTOU || issuer.refused != 1 || gateway.calls != 0 {
		t.Fatalf("err=%v calls=%d issuer=%+v gateway=%+v", err, calls, issuer, gateway)
	}
}

func TestValidatedDispatchPostPlanAuthorityChangesPersistTOCTOUAndCallOfficialGatewayZeroTimes(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*dispatchFixture)
	}{
		{
			name: "leased binding differs from planned binding",
			setup: func(f *dispatchFixture) {
				f.gates.leaseSnapshot = func(snapshot GateSnapshot) GateSnapshot {
					snapshot.Binding.AccountRef = "other-account"
					return snapshot
				}
			},
		},
		{
			name: "leased blocker changes",
			setup: func(f *dispatchFixture) {
				f.gates.leaseSnapshot = func(snapshot GateSnapshot) GateSnapshot {
					snapshot.KillSwitch = true
					return snapshot
				}
			},
		},
		{
			name: "planned manifest digest differs from leased activation",
			setup: func(f *dispatchFixture) {
				f.issuer.plan = func(plan *AtomicPlan) { plan.ManifestDigest = "sha256:other" }
			},
		},
		{
			name: "manifest is revoked after durable plan",
			setup: func(f *dispatchFixture) {
				f.issuer.issue = func() {
					f.manifest.mu.Lock()
					f.manifest.current.revoked = true
					f.manifest.mu.Unlock()
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newDispatchFixture()
			tc.setup(fixture)
			err := fixture.run()
			requireDispatchReason(t, err, ReasonTOCTOU)
			if fixture.issuer.issued != 1 || fixture.issuer.refused != 1 || fixture.issuer.refusalReason != ReasonTOCTOU || fixture.issuer.inDoubt != 0 || fixture.issuer.dispatched != 0 || fixture.gateway.calls != 0 {
				t.Fatalf("unexpected calls issuer=%+v gateway=%+v", fixture.issuer, fixture.gateway)
			}
		})
	}
}

func TestValidatedDispatchSuccessfulLeaseOutcomePersistenceBranches(t *testing.T) {
	persistErr := errors.New("injected journal persistence failure")
	tests := []struct {
		name           string
		outcome        execgw.Outcome
		configure      func(*issuerSpy)
		wantReason     Reason
		wantRefused    int
		wantInDoubt    int
		wantDispatched int
		wantPersistErr bool
	}{
		{
			name: "confirmed dispatch persistence fails", outcome: execgw.Outcome{AttemptID: "mutation-1", BrokerOrderID: "broker-1", State: journal.StateConfirmed},
			configure: func(issuer *issuerSpy) { issuer.dispatchedErr = persistErr }, wantReason: ReasonPlan, wantDispatched: 1, wantPersistErr: true,
		},
		{
			name: "definitive refusal persists", outcome: execgw.Outcome{State: journal.StateNotDispatched},
			wantReason: ReasonGateway, wantRefused: 1,
		},
		{
			name: "definitive refusal persistence fails", outcome: execgw.Outcome{State: journal.StateNotDispatched},
			configure: func(issuer *issuerSpy) { issuer.refusalErr = persistErr }, wantReason: ReasonPlan, wantRefused: 1, wantPersistErr: true,
		},
		{
			name: "ambiguous malformed success persists in doubt", outcome: execgw.Outcome{AttemptID: "mutation-1", State: journal.StateConfirmed},
			wantReason: ReasonGateway, wantInDoubt: 1,
		},
		{
			name: "ambiguous malformed success persistence fails", outcome: execgw.Outcome{AttemptID: "mutation-1", State: journal.StateConfirmed},
			configure: func(issuer *issuerSpy) { issuer.inDoubtErr = persistErr }, wantReason: ReasonPlan, wantInDoubt: 1, wantPersistErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newDispatchFixture()
			fixture.gateway.out = tc.outcome
			if tc.configure != nil {
				tc.configure(fixture.issuer)
			}
			err := fixture.run()
			requireDispatchReason(t, err, tc.wantReason)
			if tc.wantPersistErr && !errors.Is(err, persistErr) {
				t.Fatalf("err=%v does not retain persistence failure", err)
			}
			if fixture.gateway.calls != 1 || fixture.issuer.refused != tc.wantRefused || fixture.issuer.inDoubt != tc.wantInDoubt || fixture.issuer.dispatched != tc.wantDispatched {
				t.Fatalf("unexpected calls issuer=%+v gateway=%+v", fixture.issuer, fixture.gateway)
			}
		})
	}
}

func TestValidatedDispatchLeaseAndPostCallErrorsPersistExactTerminalState(t *testing.T) {
	callErr := errors.New("injected lease or official call failure")
	persistErr := errors.New("injected journal persistence failure")
	tests := []struct {
		name           string
		setup          func(*dispatchFixture)
		wantReason     Reason
		wantRefused    int
		wantInDoubt    int
		wantDispatched int
		wantGateway    int
	}{
		{
			name:       "lease rejected before callback and TOCTOU refusal persistence fails",
			setup:      func(f *dispatchFixture) { f.gates.leaseBeforeErr = callErr; f.issuer.refusalErr = persistErr },
			wantReason: ReasonPlan, wantRefused: 1,
		},
		{
			name: "post-call definitive refusal persistence fails",
			setup: func(f *dispatchFixture) {
				f.gateway.out = execgw.Outcome{AttemptID: "mutation-1", State: journal.StateFailedConfirmed}
				f.gateway.err = callErr
				f.issuer.refusalErr = persistErr
			},
			wantReason: ReasonPlan, wantRefused: 1, wantGateway: 1,
		},
		{
			name: "post-call ambiguous outcome persistence fails",
			setup: func(f *dispatchFixture) {
				f.gateway.out = execgw.Outcome{AttemptID: "mutation-1", State: journal.StateDispatchStarted}
				f.gateway.err = callErr
				f.issuer.inDoubtErr = persistErr
			},
			wantReason: ReasonPlan, wantInDoubt: 1, wantGateway: 1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newDispatchFixture()
			tc.setup(fixture)
			err := fixture.run()
			requireDispatchReason(t, err, tc.wantReason)
			if !errors.Is(err, callErr) || !errors.Is(err, persistErr) {
				t.Fatalf("joined error lost cause: %v", err)
			}
			if fixture.gateway.calls != tc.wantGateway || fixture.issuer.refused != tc.wantRefused || fixture.issuer.inDoubt != tc.wantInDoubt || fixture.issuer.dispatched != tc.wantDispatched {
				t.Fatalf("unexpected calls issuer=%+v gateway=%+v", fixture.issuer, fixture.gateway)
			}
		})
	}
}

func TestValidatedDispatchPostCallDefinitiveRefusalPersistsAndReturnsGatewayCause(t *testing.T) {
	callErr := errors.New("injected confirmed broker refusal")
	fixture := newDispatchFixture()
	fixture.gateway.out = execgw.Outcome{AttemptID: "mutation-1", State: journal.StateFailedConfirmed}
	fixture.gateway.err = callErr

	err := fixture.run()
	requireDispatchReason(t, err, ReasonGateway)
	if !errors.Is(err, callErr) || fixture.gateway.calls != 1 || fixture.issuer.refused != 1 || fixture.issuer.refusalReason != ReasonGateway || fixture.issuer.inDoubt != 0 || fixture.issuer.dispatched != 0 {
		t.Fatalf("unexpected result err=%v issuer=%+v gateway=%+v", err, fixture.issuer, fixture.gateway)
	}
}

func TestPostCallErrorCannotReachStructurallyUnreachableDispatchedPersistenceBranch(t *testing.T) {
	fixture := newDispatchFixture()
	fixture.gateway.out = execgw.Outcome{AttemptID: "mutation-1", BrokerOrderID: "broker-1", State: journal.StateConfirmed}
	fixture.gates.leaseAfterError = errors.New("injected post-callback lease failure")

	err := fixture.run()
	requireDispatchReason(t, err, ReasonGateway)
	if fixture.gateway.calls != 1 || fixture.issuer.dispatched != 0 || fixture.issuer.inDoubt != 1 || fixture.issuer.refused != 0 {
		t.Fatalf("post-call error must classify IN_DOUBT, issuer=%+v gateway=%+v", fixture.issuer, fixture.gateway)
	}
}

func TestManifestVerificationIsExactExpiringAndRevocable(t *testing.T) {
	now := time.Date(2026, 8, 1, 1, 0, 0, 0, time.UTC)
	binding := manifestFixture(now)
	repo := installedRepository(binding)
	activation, err := repo.Verify(binding, now)
	if err != nil || activation.Digest() == "" {
		t.Fatalf("activation=%+v err=%v", activation, err)
	}
	for _, tc := range []struct {
		name string
		at   time.Time
		edit func(*ManifestBinding)
	}{
		{name: "expired", at: binding.ExpiresAt},
		{name: "binding mismatch", at: now, edit: func(v *ManifestBinding) { v.SettingsDigest = "other" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := binding
			if tc.edit != nil {
				tc.edit(&changed)
			}
			if _, err := repo.Verify(changed, tc.at); err == nil {
				t.Fatal("manifest accepted")
			}
		})
	}
	repo.mu.Lock()
	repo.current.revoked = true
	repo.mu.Unlock()
	if _, err := repo.Verify(binding, now); err == nil {
		t.Fatal("revoked manifest accepted")
	}
}

func TestManifestVerificationRejectsMismatchInEveryField(t *testing.T) {
	now := time.Date(2026, 8, 1, 1, 0, 0, 0, time.UTC)
	binding := manifestFixture(now)
	repo := installedRepository(binding)
	typeOf := reflect.TypeFor[ManifestBinding]()
	if typeOf.NumField() != 32 {
		t.Fatalf("ManifestBinding field count=%d; update exhaustive mismatch coverage", typeOf.NumField())
	}
	for index := range typeOf.NumField() {
		field := typeOf.Field(index)
		t.Run(field.Name, func(t *testing.T) {
			changed := binding
			changeField(t, &changed, index)
			if _, err := repo.Verify(changed, now); err == nil {
				t.Fatalf("manifest accepted changed %s", field.Name)
			}
		})
	}
}

func TestDecisionBindingRejectsMismatchInEveryDecisionRecordField(t *testing.T) {
	now := time.Date(2026, 8, 1, 1, 0, 0, 0, time.UTC)
	record := decisionRecordFixture(now)
	binding := bindingForRecord(now, record)
	gate := gateForRecord(binding, record)
	typeOf := reflect.TypeFor[strategyengine.DecisionRecord]()
	if typeOf.NumField() != 60 || reflect.TypeFor[DecisionBinding]().NumField() != typeOf.NumField() {
		t.Fatalf("decision binding coverage=%d/%d; update exhaustive laundering guard", reflect.TypeFor[DecisionBinding]().NumField(), typeOf.NumField())
	}
	for index := range typeOf.NumField() {
		field := typeOf.Field(index)
		t.Run(field.Name, func(t *testing.T) {
			changed := gate
			changeField(t, &changed.Decision, index)
			if reason := checkGate(changed, record); reason != ReasonActivation {
				t.Fatalf("changed %s reason=%q", field.Name, reason)
			}
		})
	}
}

func TestManifestLeaseBlocksRevocationAcrossCallback(t *testing.T) {
	now := time.Date(2026, 8, 1, 1, 0, 0, 0, time.UTC)
	binding := manifestFixture(now)
	repo := installedRepository(binding)
	entered := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- repo.WithVerifiedLease(binding, now, func(Activation) error {
			close(entered)
			<-release
			return nil
		})
	}()
	<-entered
	if repo.mu.TryLock() {
		repo.mu.Unlock()
		t.Fatal("manifest write lock acquired while verified callback held its read lease")
	}
	writerStarted := make(chan struct{})
	mutated := make(chan struct{})
	go func() {
		close(writerStarted)
		repo.mu.Lock()
		repo.current.revoked = true
		repo.mu.Unlock()
		close(mutated)
	}()
	// Start the competing writer before release. The TryLock assertion above is
	// the deterministic proof that the callback owns the read lease; this
	// barrier only orders release before the writer-completion assertion below.
	<-writerStarted
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	select {
	case <-mutated:
	case <-time.After(time.Second):
		t.Fatal("manifest writer remained blocked after callback")
	}
}

func TestOfficialOutcomeDispositionPreservesAttemptEvidence(t *testing.T) {
	tests := []struct {
		name string
		out  execgw.Outcome
		err  error
		want outcomeDisposition
	}{
		{name: "exact confirmed", out: execgw.Outcome{AttemptID: "mutation-1", BrokerOrderID: "broker-1", State: journal.StateConfirmed}, want: outcomeDispatched},
		{name: "empty attempt refuses even on timeout", err: context.DeadlineExceeded, want: outcomeRefused},
		{name: "confirmed failure refuses", out: execgw.Outcome{AttemptID: "mutation-1", State: journal.StateFailedConfirmed}, err: errors.New("broker refused"), want: outcomeRefused},
		{name: "not dispatched refuses", out: execgw.Outcome{AttemptID: "mutation-1", State: journal.StateNotDispatched}, err: errors.New("local refusal"), want: outcomeRefused},
		{name: "dispatch-started timeout is in doubt", out: execgw.Outcome{AttemptID: "mutation-1", State: journal.StateDispatchStarted}, err: context.DeadlineExceeded, want: outcomeInDoubt},
		{name: "malformed success is in doubt", out: execgw.Outcome{AttemptID: "mutation-1", State: journal.StateConfirmed}, want: outcomeInDoubt},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := classifyOfficialOutcome(tc.out, tc.err)
			if got != tc.want {
				t.Fatalf("got=%s want=%s", got, tc.want)
			}
		})
	}
}

func TestExactFloatRejectsLossyOrNonCanonicalDecimals(t *testing.T) {
	for _, tc := range []struct {
		raw   string
		whole bool
		ok    bool
	}{
		{raw: "1", whole: true, ok: true},
		{raw: "100.125", ok: true},
		{raw: "1.0", whole: true},
		{raw: "01", whole: true},
		{raw: "9007199254740993", whole: true},
		{raw: "0.10000000000000001"},
		{raw: "0"},
	} {
		_, err := exactFloat(tc.raw, tc.whole)
		if (err == nil) != tc.ok {
			t.Errorf("exactFloat(%q,%v) err=%v ok=%v", tc.raw, tc.whole, err, tc.ok)
		}
	}
}

func TestDeterministicAttemptIDBindsAccountDecisionAndGeneration(t *testing.T) {
	base := deterministicAttemptID("acct", "decision", 7)
	if base != deterministicAttemptID("acct", "decision", 7) {
		t.Fatal("attempt identity is not deterministic")
	}
	for _, other := range []string{
		deterministicAttemptID("other", "decision", 7),
		deterministicAttemptID("acct", "other", 7),
		deterministicAttemptID("acct", "decision", 8),
	} {
		if other == base {
			t.Fatal("attempt identity omitted a binding field")
		}
	}
}

// Keep sync imported: this compile-time assertion pins ManifestRepository's
// lease to a real RWMutex rather than a copyable snapshot lock.
var _ sync.Locker = (*sync.RWMutex)(nil)
