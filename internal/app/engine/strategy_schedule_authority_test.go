package engine

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestPairedProductionScheduleAuthorityLoadsExactKRUSAtOneFrozenClock(t *testing.T) {
	fixture := newStrategyScheduleFixture(t)
	countingClock := &countingStrategyScheduleClock{Clock: clock.NewFake(fixture.now)}
	loader := newStrategyScheduleAuthorityLoader(fixture.dir, countingClock, fixture.reader, fixture.getenv)
	authority := loader.collect(context.Background())
	snapshot := authority.Snapshot()
	if !snapshot.ObservedAt.Equal(fixture.now) || countingClock.calls.Load() != 1 {
		t.Fatalf("observed=%s clockCalls=%d", snapshot.ObservedAt, countingClock.calls.Load())
	}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		got := snapshot.For(market)
		if !got.DesiredEnabled || !got.DesiredAutostart || !got.Ready || got.Reason != scheduler.ResumeExactManifest ||
			got.CalendarVersion != fixture.calendarVersion(market) || got.ActivationManifestDigest != fixture.manifestDigest(market) ||
			got.ConfigDigest != strategyRuntimeConfigDigest() {
			t.Fatalf("market=%s snapshot=%+v", market, got)
		}
		private := authority.forMarket(market)
		if private.restore.Activation == nil || private.calendar.Version != got.CalendarVersion {
			t.Fatalf("market=%s private authority incomplete", market)
		}
	}
}

func TestPairedProductionScheduleAuthorityStartsKRUSCalendarReadsConcurrently(t *testing.T) {
	fixture := newStrategyScheduleFixture(t)
	started := make(chan string, 2)
	usCompleted := make(chan struct{})
	releaseKR := make(chan struct{})
	fixture.reader.started = started
	fixture.reader.releaseKR = releaseKR
	fixture.reader.usCompleted = usCompleted
	loader := newStrategyScheduleAuthorityLoader(fixture.dir, clock.NewFake(fixture.now), fixture.reader, fixture.getenv)
	done := make(chan strategyScheduleAuthorityPair, 1)
	go func() { done <- loader.collect(context.Background()) }()
	seen := map[string]bool{}
	for len(seen) < 2 {
		select {
		case market := <-started:
			seen[market] = true
		case <-time.After(time.Second):
			t.Fatalf("paired calendar reads did not start: %v", seen)
		}
	}
	select {
	case <-usCompleted:
	case <-time.After(time.Second):
		t.Fatal("US calendar did not complete while KR was blocked")
	}
	select {
	case <-done:
		t.Fatal("loader returned before KR calendar was released")
	default:
	}
	close(releaseKR)
	authority := <-done
	if !authority.Snapshot().KR.Ready || !authority.Snapshot().US.Ready {
		t.Fatalf("paired snapshot=%+v", authority.Snapshot())
	}
}

func TestFinalScheduleAuthorityReadsOnlyTargetMarketKRUS(t *testing.T) {
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		t.Run(string(market), func(t *testing.T) {
			fixture := newStrategyScheduleFixture(t)
			blocked := make(chan struct{})
			if market == StrategyMarketKR {
				fixture.reader.releaseUS = blocked
			} else {
				fixture.reader.releaseKR = blocked
			}
			loader := newStrategyScheduleAuthorityLoader(fixture.dir, clock.NewFake(fixture.now), fixture.reader, fixture.getenv)
			done := make(chan strategyScheduleMarketAuthority, 1)
			go func() { done <- loader.collectMarket(context.Background(), market) }()
			select {
			case authority := <-done:
				if !authority.snapshot.Ready || fixture.reader.totalCalls() != 1 {
					t.Fatalf("market=%s authority=%+v calls=%d", market, authority.snapshot, fixture.reader.totalCalls())
				}
			case <-time.After(time.Second):
				t.Fatalf("market=%s final validation waited for blocked peer", market)
			}
		})
	}
}

func TestProductionScheduleAuthorityKRRefusalDoesNotContaminateUS(t *testing.T) {
	fixture := newStrategyScheduleFixture(t)
	fixture.env[strategyActivationManifestDigestEnv(StrategyMarketKR)] = "sha256:" + strings.Repeat("0", 64)
	loader := newStrategyScheduleAuthorityLoader(fixture.dir, clock.NewFake(fixture.now), fixture.reader, fixture.getenv)
	snapshot := loader.collect(context.Background()).Snapshot()
	if snapshot.KR.Ready || snapshot.KR.Reason != scheduler.ResumeManifestMismatch {
		t.Fatalf("KR=%+v", snapshot.KR)
	}
	if !snapshot.US.Ready || snapshot.US.Reason != scheduler.ResumeExactManifest {
		t.Fatalf("US=%+v", snapshot.US)
	}
}

func TestProductionScheduleAuthorityMarketLocalInputFailuresPreservePeer(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*strategyScheduleFixture) error
		want scheduler.ResumeReason
	}{
		{name: "missing KR pins", want: scheduler.ResumeManifestUnavailable, edit: func(f *strategyScheduleFixture) error {
			delete(f.env, strategyActivationManifestDigestEnv(StrategyMarketKR))
			delete(f.env, strategyActivationKeyIDEnv(StrategyMarketKR))
			delete(f.env, strategyActivationPublicKeyEnv(StrategyMarketKR))
			return nil
		}},
		{name: "malformed KR desired", want: scheduler.ResumeVerificationFailed, edit: func(f *strategyScheduleFixture) error {
			return os.WriteFile(filepath.Join(f.dir, strategyDesiredFileName(StrategyMarketKR)), []byte(`{"enabled":`), 0o600)
		}},
		{name: "KR calendar unavailable", want: scheduler.ResumeVerificationFailed, edit: func(f *strategyScheduleFixture) error {
			f.reader.errors = map[string]error{"KR": errors.New("KR calendar unavailable")}
			return nil
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newStrategyScheduleFixture(t)
			if err := tc.edit(fixture); err != nil {
				t.Fatal(err)
			}
			snapshot := newStrategyScheduleAuthorityLoader(fixture.dir, clock.NewFake(fixture.now), fixture.reader, fixture.getenv).
				collect(context.Background()).Snapshot()
			if snapshot.KR.Ready || snapshot.KR.Reason != tc.want {
				t.Fatalf("KR=%+v", snapshot.KR)
			}
			if !snapshot.US.Ready || snapshot.US.Reason != scheduler.ResumeExactManifest {
				t.Fatalf("US=%+v", snapshot.US)
			}
		})
	}
}

func TestProductionScheduleAuthorityDormantBaselineMakesZeroCalendarReads(t *testing.T) {
	reader := &strategyScheduleCalendarFixture{calls: make(map[string]int)}
	now := time.Date(2026, 8, 4, 5, 0, 0, 0, time.UTC)
	loader := newStrategyScheduleAuthorityLoader(t.TempDir(), clock.NewFake(now), reader, func(string) string { return "" })
	authority := loader.collect(context.Background())
	if reader.totalCalls() != 0 {
		t.Fatalf("dormant calendar calls=%d", reader.totalCalls())
	}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		got := authority.Snapshot().For(market)
		if got.Ready || got.DesiredEnabled || got.DesiredAutostart || got.Reason != scheduler.ResumeAutoStartOff ||
			got.CalendarVersion != "" || got.ActivationManifestDigest != "" {
			t.Fatalf("market=%s dormant=%+v", market, got)
		}
	}
}

func TestProductionSchedulePublicSnapshotHasNoAuthorityOrMutationCapability(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeOf(StrategyScheduleMarketSnapshot{}), reflect.TypeOf(PairedStrategyScheduleSnapshot{})} {
		for index := 0; index < typ.NumField(); index++ {
			field := typ.Field(index)
			identity := field.Type.String()
			for _, forbidden := range []string{"Activation", "CalendarSnapshot", "Gateway", "Journal", "Guardian", "Trigger", "Cycle", "Writer", "Toggle", "Broker", "Order"} {
				if strings.Contains(identity, forbidden) {
					t.Fatalf("public snapshot %s contains forbidden capability %q", identity, forbidden)
				}
			}
		}
	}
}

func TestStrategyRuntimeConfigDigestIsPairedStableAndReleaseBound(t *testing.T) {
	first, second := strategyRuntimeConfigDigest(), strategyRuntimeConfigDigest()
	if first != second || !strings.HasPrefix(first, "sha256:") || len(first) != len("sha256:")+64 {
		t.Fatalf("config digests=(%q,%q)", first, second)
	}
	descriptors := strategyflow.Descriptors()
	descriptors[0].Release += "-drift"
	if drifted := strategyRuntimeConfigDigestFor(strategyrouter.RouterID, strategyrouter.RouterRelease, descriptors); drifted == first {
		t.Fatal("lane release drift did not change compiled strategy config digest")
	}
	if drifted := strategyRuntimeConfigDigestFor(strategyrouter.RouterID, strategyrouter.RouterRelease+"-drift", strategyflow.Descriptors()); drifted == first {
		t.Fatal("router release drift did not change compiled strategy config digest")
	}
}

type strategyScheduleFixture struct {
	dir       string
	now       time.Time
	public    ed25519.PublicKey
	private   ed25519.PrivateKey
	env       map[string]string
	reader    *strategyScheduleCalendarFixture
	calendars map[StrategyMarket]scheduler.CalendarSnapshot
	manifests map[StrategyMarket]string
}

func newStrategyScheduleFixture(t *testing.T) *strategyScheduleFixture {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 4, 5, 0, 0, 0, time.UTC)
	reader := &strategyScheduleCalendarFixture{payloads: strategyCalendarPayloads(now), calls: make(map[string]int)}
	fixture := &strategyScheduleFixture{dir: t.TempDir(), now: now, public: public, private: private,
		env: make(map[string]string), reader: reader, calendars: make(map[StrategyMarket]scheduler.CalendarSnapshot, 2), manifests: make(map[StrategyMarket]string, 2)}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		clockMarket := strategyClockMarket(market)
		calendar, err := scheduler.AdaptOfficialCalendar(clockMarket, reader.payloads[string(market)], now)
		if err != nil {
			t.Fatalf("calendar %s: %v", market, err)
		}
		fixture.calendars[market] = calendar
		desired := scheduler.DesiredState{Revision: 1, Version: scheduler.SchedulerVersion, Enabled: true, AutoStart: true,
			Market: strategySchedulerMarket(market), Session: scheduler.SessionRegular, Actor: "human-approver",
			ApprovedAt: now.Add(-2 * time.Minute), CalendarVersion: calendar.Version, ConfigVersion: strategyRuntimeConfigDigest()}
		desiredData, err := json.Marshal(desired)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(fixture.dir, strategyDesiredFileName(market)), desiredData, 0o600); err != nil {
			t.Fatal(err)
		}
		body := engineActivationManifestBody{SchemaVersion: "strategy-activation-manifest:v1",
			Domain: "TossOS/strategy-scheduler-activation/ed25519/v1", SignatureAlgorithm: "Ed25519", KeyID: "key-" + string(market), Generation: 1,
			SchedulerVersion: scheduler.SchedulerVersion, DesiredRevision: desired.Revision, CalendarVersion: calendar.Version,
			Market: desired.Market, Session: desired.Session, ConfigVersion: desired.ConfigVersion, BuildDigest: strategyRuntimeBuildDigest(),
			Actor: desired.Actor, ApprovedAt: desired.ApprovedAt.UTC().Format(time.RFC3339Nano),
			IssuedAt: now.Add(-time.Minute).Format(time.RFC3339Nano), ExpiresAt: now.Add(time.Hour).Format(time.RFC3339Nano)}
		canonical, _ := json.Marshal(body)
		manifestData, _ := json.Marshal(engineActivationManifest{engineActivationManifestBody: body,
			Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, canonical))})
		path := filepath.Join(fixture.dir, scheduler.ProductionActivationFileName(desired.Market))
		if err := os.WriteFile(path, manifestData, 0o400); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, 0o400); err != nil {
			t.Fatal(err)
		}
		digest := strategyTestDigest(manifestData)
		fixture.manifests[market] = digest
		fixture.env[strategyActivationManifestDigestEnv(market)] = digest
		fixture.env[strategyActivationKeyIDEnv(market)] = body.KeyID
		fixture.env[strategyActivationPublicKeyEnv(market)] = base64.StdEncoding.EncodeToString(public)
	}
	return fixture
}

func (fixture *strategyScheduleFixture) getenv(key string) string { return fixture.env[key] }

func (fixture *strategyScheduleFixture) calendarVersion(market StrategyMarket) string {
	return fixture.calendars[market].Version
}

func (fixture *strategyScheduleFixture) manifestDigest(market StrategyMarket) string {
	return fixture.manifests[market]
}

type strategyScheduleCalendarFixture struct {
	mu          sync.Mutex
	payloads    map[string]official.MarketCalendarResponse
	calls       map[string]int
	errors      map[string]error
	started     chan<- string
	releaseKR   <-chan struct{}
	releaseUS   <-chan struct{}
	usCompleted chan<- struct{}
}

func (reader *strategyScheduleCalendarFixture) TypedMarketCalendar(ctx context.Context, country, _ string) (official.MarketCalendarResponse, error) {
	reader.mu.Lock()
	reader.calls[country]++
	readErr := reader.errors[country]
	reader.mu.Unlock()
	if reader.started != nil {
		reader.started <- country
	}
	if country == "KR" && reader.releaseKR != nil {
		select {
		case <-reader.releaseKR:
		case <-ctx.Done():
			return official.MarketCalendarResponse{}, ctx.Err()
		}
	}
	if country == "US" && reader.releaseUS != nil {
		select {
		case <-reader.releaseUS:
		case <-ctx.Done():
			return official.MarketCalendarResponse{}, ctx.Err()
		}
	}
	if readErr != nil {
		return official.MarketCalendarResponse{}, readErr
	}
	payload, ok := reader.payloads[country]
	if !ok {
		return official.MarketCalendarResponse{}, errors.New("calendar unavailable")
	}
	if country == "US" && reader.usCompleted != nil {
		close(reader.usCompleted)
	}
	return payload, nil
}

func (reader *strategyScheduleCalendarFixture) totalCalls() int {
	reader.mu.Lock()
	defer reader.mu.Unlock()
	total := 0
	for _, count := range reader.calls {
		total += count
	}
	return total
}

func strategyCalendarPayloads(now time.Time) map[string]official.MarketCalendarResponse {
	makeDay := func(market StrategyMarket, date string) official.MarketCalendarDay {
		location, _ := strategyClockMarket(market).Location()
		day, _ := time.ParseInLocation("2006-01-02", date, location)
		open := day.Add(9 * time.Hour)
		if market == StrategyMarketUS {
			open = day.Add(9*time.Hour + 30*time.Minute)
		}
		session := &official.MarketCalendarSession{StartTime: open, EndTime: open.Add(6*time.Hour + 30*time.Minute)}
		if market == StrategyMarketKR {
			return official.MarketCalendarDay{Date: date, Integrated: &official.MarketCalendarSessions{RegularMarket: session}}
		}
		return official.MarketCalendarDay{Date: date, RegularMarket: session}
	}
	return map[string]official.MarketCalendarResponse{
		"KR": {PreviousBusinessDay: makeDay(StrategyMarketKR, "2026-08-03"), Today: makeDay(StrategyMarketKR, "2026-08-04"), NextBusinessDay: makeDay(StrategyMarketKR, "2026-08-05")},
		"US": {PreviousBusinessDay: makeDay(StrategyMarketUS, "2026-08-02"), Today: makeDay(StrategyMarketUS, "2026-08-03"), NextBusinessDay: makeDay(StrategyMarketUS, "2026-08-04")},
	}
}

type countingStrategyScheduleClock struct {
	clock.Clock
	calls atomic.Int32
}

func (value *countingStrategyScheduleClock) Now() time.Time {
	value.calls.Add(1)
	return value.Clock.Now()
}

type engineActivationManifestBody struct {
	SchemaVersion      string                 `json:"schema_version"`
	Domain             string                 `json:"domain"`
	SignatureAlgorithm string                 `json:"signature_algorithm"`
	KeyID              string                 `json:"key_id"`
	Generation         uint64                 `json:"generation"`
	SchedulerVersion   string                 `json:"scheduler_version"`
	DesiredRevision    uint64                 `json:"desired_revision"`
	CalendarVersion    string                 `json:"calendar_version"`
	Market             scheduler.MarketScope  `json:"market"`
	Session            scheduler.SessionScope `json:"session"`
	ConfigVersion      string                 `json:"config_version"`
	BuildDigest        string                 `json:"build_digest"`
	Actor              string                 `json:"actor"`
	ApprovedAt         string                 `json:"approved_at"`
	IssuedAt           string                 `json:"issued_at"`
	ExpiresAt          string                 `json:"expires_at"`
	Revoked            bool                   `json:"revoked"`
}

type engineActivationManifest struct {
	engineActivationManifestBody
	Signature string `json:"signature"`
}

func strategyTestDigest(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}
