package obs_test

// obs_test.go covers structured logging, the ntfy transport, alert grading, the
// durable critical outbox and the heartbeat (task 4.3).
//
// Every network interaction is httptest. There is no code path in these tests
// that can produce a request to a real host.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

var obsNow = time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)

func openJournal(t *testing.T, clk clock.Clock) *journal.Journal {
	t.Helper()
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(t.TempDir(), "journal.db"),
		Clock:    clk,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	return j
}

// --- structured logging -----------------------------------------------------

// TestLogLinesAreCountable is the "셀 수 있는 이벤트 규약" test: every line
// carries the stable event name, its subject and its grade as *fields*, and the
// variable parts are fields too rather than being interpolated into prose.
func TestLogLinesAreCountable(t *testing.T) {
	var buf bytes.Buffer
	log := obs.NewLogger(obs.LogOptions{Writer: &buf, JSON: true, Clock: clock.NewFake(obsNow)})

	log.OrderTransition(obs.EventOrderConfirmed, "i-1", "a-1", "005930", "ACKED", "CONFIRMED",
		obs.FieldOrderID, "O-7")

	line := decodeLine(t, buf.String())
	if line[obs.FieldEvent] != string(obs.EventOrderConfirmed) {
		t.Errorf("event = %v", line[obs.FieldEvent])
	}
	if line[obs.FieldSubject] != "order" {
		t.Errorf("subject = %v, want order", line[obs.FieldSubject])
	}
	if line[obs.FieldSeverity] != string(obs.SeverityNormal) {
		t.Errorf("severity = %v", line[obs.FieldSeverity])
	}
	for field, want := range map[string]string{
		obs.FieldIntentID:  "i-1",
		obs.FieldAttemptID: "a-1",
		obs.FieldSymbol:    "005930",
		obs.FieldFromState: "ACKED",
		obs.FieldToState:   "CONFIRMED",
		obs.FieldOrderID:   "O-7",
	} {
		if line[field] != want {
			t.Errorf("%s = %v, want %q", field, line[field], want)
		}
	}
	// The injected clock has to reach the emitted line, or a test cannot assert
	// on timing and a replay cannot be ordered.
	if ts, _ := line["time"].(string); !strings.HasPrefix(ts, "2026-07-26T10:00:00") {
		t.Errorf("time = %v, want the injected clock's instant", line["time"])
	}
}

func TestErrorLinesCarryTheErrorAsAField(t *testing.T) {
	var buf bytes.Buffer
	log := obs.NewLogger(obs.LogOptions{Writer: &buf, JSON: true, Clock: clock.NewFake(obsNow),
		Level: slog.LevelDebug})

	log.Error(obs.EventBrokerStateUnknown, errors.New("filled quantity went backwards"),
		obs.FieldSymbol, "AAPL")

	line := decodeLine(t, buf.String())
	if line[obs.FieldError] != "filled quantity went backwards" {
		t.Errorf("error = %v", line[obs.FieldError])
	}
	if line[obs.FieldSeverity] != string(obs.SeverityCritical) {
		t.Errorf("severity = %v, want critical", line[obs.FieldSeverity])
	}
}

func TestReconcileLineCounts(t *testing.T) {
	var buf bytes.Buffer
	log := obs.NewLogger(obs.LogOptions{Writer: &buf, JSON: true, Clock: clock.NewFake(obsNow)})

	log.Reconcile(obs.EventReconcileMismatch, 8, 2, 1500*time.Millisecond)

	line := decodeLine(t, buf.String())
	if line["matched"] != float64(8) || line["mismatched"] != float64(2) {
		t.Errorf("matched/mismatched = %v/%v", line["matched"], line["mismatched"])
	}
	if line[obs.FieldCount] != float64(10) {
		t.Errorf("count = %v, want 10", line[obs.FieldCount])
	}
	if line[obs.FieldDurationMS] != float64(1500) {
		t.Errorf("duration = %v", line[obs.FieldDurationMS])
	}
}

// TestNilLoggerIsSafe: a component with no logger wired must not panic. Alerting
// code runs while something is already wrong.
func TestNilLoggerIsSafe(t *testing.T) {
	var log *obs.Logger
	log.Event(obs.EventEngineStarted)
	log.Warn(obs.EventEntryBlocked)
	log.Error(obs.EventOrderInDoubt, errors.New("boom"))
	if log.Slog() == nil {
		t.Error("Slog() must return a usable logger even from nil")
	}
}

// --- grading ----------------------------------------------------------------

// TestGradingMatchesTheSpec pins every event engine-safety names as critical.
func TestGradingMatchesTheSpec(t *testing.T) {
	critical := []obs.EventType{
		obs.EventOrderInDoubt,       // IN_DOUBT
		obs.EventOrderUnresolved,    // UNRESOLVED_IN_DOUBT
		obs.EventCredentialExpiring, // 자격증명 만료 임박
		obs.EventReconcilePermanent, // 영구 불일치
		obs.EventBrokerStateUnknown, // UNKNOWN_BROKER_STATE
	}
	for _, e := range critical {
		if got := obs.SeverityOf(e); got != obs.SeverityCritical {
			t.Errorf("SeverityOf(%s) = %s, want critical", e, got)
		}
	}

	for _, e := range []obs.EventType{
		obs.EventFillObserved, obs.EventOrderConfirmed, obs.EventReconcileClean,
		obs.EventEngineHeartbeat,
	} {
		if got := obs.SeverityOf(e); got != obs.SeverityNormal {
			t.Errorf("SeverityOf(%s) = %s, want normal", e, got)
		}
	}
}

// TestUnknownEventIsNotCritical documents the deliberate direction: an
// unrecognised event type comes from a typo or a newer component, and neither is
// a reason to durably enqueue an unactionable alert and stop the engine trading.
func TestUnknownEventIsNotCritical(t *testing.T) {
	if got := obs.SeverityOf(obs.EventType("something.new")); got != obs.SeverityNormal {
		t.Errorf("SeverityOf(unknown) = %s, want normal", got)
	}
}

// TestPhase2EventsAreReserved: the enum is stable before the features exist.
func TestPhase2EventsAreReserved(t *testing.T) {
	if obs.EventKillSwitch == "" || obs.EventOperatingMode == "" {
		t.Error("the Phase 2 event types must be declared so consumers do not have to change")
	}
}

// --- ntfy transport ---------------------------------------------------------

type capturedRequest struct {
	path    string
	body    string
	headers http.Header
}

func ntfyServer(t *testing.T, status int) (*obs.Ntfy, func() []capturedRequest) {
	t.Helper()
	var mu sync.Mutex
	var seen []capturedRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		mu.Lock()
		seen = append(seen, capturedRequest{path: r.URL.Path, body: string(body), headers: r.Header.Clone()})
		mu.Unlock()
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"id":"x"}`))
	}))
	t.Cleanup(srv.Close)

	return &obs.Ntfy{BaseURL: srv.URL, Topic: "tossos-test", HTTPClient: srv.Client()},
		func() []capturedRequest {
			mu.Lock()
			defer mu.Unlock()
			return append([]capturedRequest(nil), seen...)
		}
}

func TestNtfyPublishesToTheTopic(t *testing.T) {
	pub, requests := ntfyServer(t, http.StatusOK)

	err := pub.Publish(context.Background(), obs.Notification{
		Type:     obs.EventOrderInDoubt,
		Severity: obs.SeverityCritical,
		Title:    "IN_DOUBT on 005930",
		Body:     "attempt a-1 could not be classified",
		Tags:     []string{"warning"},
	})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}

	got := requests()
	if len(got) != 1 {
		t.Fatalf("requests = %d, want 1", len(got))
	}
	if got[0].path != "/tossos-test" {
		t.Errorf("path = %q, want /tossos-test", got[0].path)
	}
	if got[0].body != "attempt a-1 could not be classified" {
		t.Errorf("body = %q", got[0].body)
	}
	if title := got[0].headers.Get("Title"); title != "IN_DOUBT on 005930" {
		t.Errorf("Title = %q", title)
	}
	if tags := got[0].headers.Get("Tags"); !strings.Contains(tags, string(obs.EventOrderInDoubt)) {
		t.Errorf("Tags = %q, want the event type in it", tags)
	}
	// Critical must bypass a phone's do-not-disturb.
	if prio := got[0].headers.Get("Priority"); prio != "5" {
		t.Errorf("Priority = %q, want 5 for critical", prio)
	}
}

func TestNtfyNormalPriority(t *testing.T) {
	pub, requests := ntfyServer(t, http.StatusOK)

	if err := pub.Publish(context.Background(), obs.Notification{
		Type: obs.EventFillObserved, Severity: obs.SeverityNormal, Body: "filled",
	}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if prio := requests()[0].headers.Get("Priority"); prio != "3" {
		t.Errorf("Priority = %q, want 3", prio)
	}
}

func TestNtfySendsTheBearerToken(t *testing.T) {
	pub, requests := ntfyServer(t, http.StatusOK)
	pub.Token = "tk_secret"

	if err := pub.Publish(context.Background(), obs.Notification{Type: obs.EventEngineStarted}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if auth := requests()[0].headers.Get("Authorization"); auth != "Bearer tk_secret" {
		t.Errorf("Authorization = %q", auth)
	}
}

func TestNtfyReportsARefusal(t *testing.T) {
	pub, _ := ntfyServer(t, http.StatusForbidden)

	err := pub.Publish(context.Background(), obs.Notification{Type: obs.EventOrderInDoubt})
	if err == nil {
		t.Fatal("a 403 must be reported: the caller decides what an undelivered alert means")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("err = %v, want the status in it", err)
	}
}

func TestNtfyWithoutATopicIsNotConfigured(t *testing.T) {
	var pub obs.Ntfy
	if err := pub.Publish(context.Background(), obs.Notification{}); !errors.Is(err, obs.ErrNtfyNotConfigured) {
		t.Fatalf("err = %v, want ErrNtfyNotConfigured", err)
	}
}

// TestNtfyPublicServiceIsFlagged: on ntfy.sh the topic name is the only access
// control, and account events on a guessable topic are a leak.
func TestNtfyPublicServiceIsFlagged(t *testing.T) {
	if !(&obs.Ntfy{Topic: "t"}).UsesPublicService() {
		t.Error("the default base URL is the public service and must be reported as such")
	}
	if (&obs.Ntfy{BaseURL: "https://ntfy.internal", Topic: "t"}).UsesPublicService() {
		t.Error("a self-hosted instance must not be flagged")
	}
}

// --- the critical outbox ----------------------------------------------------

// failingPublisher fails until it is told to stop.
type failingPublisher struct {
	mu       sync.Mutex
	fail     bool
	calls    int
	messages []obs.Notification
}

func (p *failingPublisher) Publish(_ context.Context, n obs.Notification) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	if p.fail {
		return errors.New("transport is down")
	}
	p.messages = append(p.messages, n)
	return nil
}

func (p *failingPublisher) setFail(v bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.fail = v
}

func (p *failingPublisher) callCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

func newNotifier(t *testing.T, pub obs.Publisher) (*obs.Notifier, *journal.Journal, *execgw.EntryGate) {
	t.Helper()
	clk := clock.NewFake(obsNow)
	j := openJournal(t, clk)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	return &obs.Notifier{
		Log:       obs.NewLogger(obs.LogOptions{Writer: newDiscard(), JSON: true, Clock: clk}),
		Publisher: pub,
		Journal:   j,
		Gate:      gate,
		// The retry wait runs on a real clock with a real (tiny) delay. Driving it
		// from the fake would mean the test has to advance time from outside the
		// call it is inside, which is a deadlock rather than a test.
		Clock:      clock.System(),
		Attempts:   3,
		RetryDelay: time.Millisecond,
	}, j, gate
}

// TestCriticalAlertIsDurableBeforeItIsSent is the ordering the outbox exists for:
// the record is on disk before anybody tries the network.
func TestCriticalAlertIsDurableBeforeItIsSent(t *testing.T) {
	pub := &failingPublisher{}
	n, j, gate := newNotifier(t, pub)
	ctx := context.Background()

	if err := n.Notify(ctx, obs.Event{
		Type:   obs.EventOrderInDoubt,
		Title:  "IN_DOUBT",
		Body:   "attempt a-1",
		Fields: map[string]any{obs.FieldSymbol: "005930", obs.FieldAttemptID: "a-1"},
	}); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	remaining, err := j.UndeliveredCount(ctx)
	if err != nil {
		t.Fatalf("UndeliveredCount: %v", err)
	}
	if remaining != 0 {
		t.Errorf("undelivered = %d, want 0 after a successful send", remaining)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Errorf("a delivered alert must not block entries, got %v", rejected)
	}
	if pub.callCount() != 1 {
		t.Errorf("publish calls = %d, want 1", pub.callCount())
	}
}

// TestPersistentDeliveryFailureBlocksEntries is engine-safety's scenario
// "critical 알림 전달 실패 지속".
func TestPersistentDeliveryFailureBlocksEntries(t *testing.T) {
	pub := &failingPublisher{fail: true}
	n, j, gate := newNotifier(t, pub)
	ctx := context.Background()

	if err := n.Notify(ctx, obs.Event{
		Type:   obs.EventOrderUnresolved,
		Title:  "UNRESOLVED_IN_DOUBT",
		Fields: map[string]any{obs.FieldAttemptID: "a-9"},
	}); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	if pub.callCount() != 3 {
		t.Errorf("publish attempts = %d, want the configured 3", pub.callCount())
	}

	rejected := gate.CheckEntry()
	if rejected == nil || rejected.Reason != execgw.ReasonAlertUndelivered {
		t.Fatalf("entries must be blocked, got %v", rejected)
	}

	// Preserved, not abandoned.
	pending, err := j.PendingAlerts(ctx, 0)
	if err != nil {
		t.Fatalf("PendingAlerts: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("pending = %d, want the alert preserved as undelivered", len(pending))
	}
	if pending[0].State != journal.AlertPending {
		t.Errorf("state = %q, want PENDING", pending[0].State)
	}
	if pending[0].Attempts != 3 {
		t.Errorf("attempts recorded = %d, want 3", pending[0].Attempts)
	}
	if !strings.Contains(pending[0].LastError, "transport is down") {
		t.Errorf("last error = %q", pending[0].LastError)
	}
}

// TestRecoveredDeliveryDoesNotReleaseTheGateByItself: "the network came back" is
// not evidence anybody read the alert.
func TestRecoveredDeliveryDoesNotReleaseTheGateByItself(t *testing.T) {
	pub := &failingPublisher{fail: true}
	n, j, gate := newNotifier(t, pub)
	ctx := context.Background()

	if err := n.Notify(ctx, obs.Event{Type: obs.EventReconcilePermanent, Key: "perm-1"}); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if gate.CheckEntry() == nil {
		t.Fatal("entries must be blocked after the failures")
	}

	pub.setFail(false)
	delivered, remaining, err := n.Flush(ctx)
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if delivered != 1 || remaining != 0 {
		t.Fatalf("flush = %d delivered, %d remaining; want 1, 0", delivered, remaining)
	}
	if gate.CheckEntry() == nil {
		t.Fatal("delivery recovering must NOT release the gate on its own")
	}

	if err := n.Acknowledge(ctx, "operator"); err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("an operator acknowledgement must release the gate, got %v", rejected)
	}
	_ = j
}

// TestAcknowledgeRequiresAnIdentity, like every other operator override.
func TestAcknowledgeRequiresAnIdentity(t *testing.T) {
	n, _, _ := newNotifier(t, &failingPublisher{})
	if err := n.Acknowledge(context.Background(), "  "); err == nil {
		t.Fatal("acknowledging without an identity must fail")
	}
}

// TestAcknowledgeWhileStillPendingKeepsTheBlock: acknowledging one alert out of
// two must not reopen the gate.
func TestAcknowledgeWhileStillPendingKeepsTheBlock(t *testing.T) {
	pub := &failingPublisher{fail: true}
	n, j, gate := newNotifier(t, pub)
	ctx := context.Background()

	for _, key := range []string{"one", "two"} {
		if err := n.Notify(ctx, obs.Event{Type: obs.EventBrokerStateUnknown, Key: key}); err != nil {
			t.Fatalf("Notify(%s): %v", key, err)
		}
	}
	pending, err := j.PendingAlerts(ctx, 0)
	if err != nil || len(pending) != 2 {
		t.Fatalf("pending = %+v (%v), want 2", pending, err)
	}

	if err := n.Acknowledge(ctx, "operator", pending[0].ID); err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}
	if gate.CheckEntry() == nil {
		t.Fatal("one alert is still undelivered; the gate must stay shut")
	}

	if err := n.Acknowledge(ctx, "operator", pending[1].ID); err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("gate = %v, want released once nothing is pending", rejected)
	}
}

// TestTheSameConditionEnqueuesOnce: three consecutive polls observing one
// contradiction is one alert, not a pager storm.
func TestTheSameConditionEnqueuesOnce(t *testing.T) {
	pub := &failingPublisher{fail: true}
	n, j, _ := newNotifier(t, pub)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if err := n.Notify(ctx, obs.Event{
			Type:   obs.EventBrokerStateUnknown,
			Fields: map[string]any{obs.FieldSymbol: "AAPL"},
		}); err != nil {
			t.Fatalf("Notify: %v", err)
		}
	}

	pending, err := j.PendingAlerts(ctx, 0)
	if err != nil {
		t.Fatalf("PendingAlerts: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("outbox rows = %d, want 1 — the same condition is one alert", len(pending))
	}
}

// TestOrdinaryAlertsAreBestEffort: a failed send must not enqueue anything, must
// not block entries, and must not fail the caller.
func TestOrdinaryAlertsAreBestEffort(t *testing.T) {
	pub := &failingPublisher{fail: true}
	n, j, gate := newNotifier(t, pub)
	ctx := context.Background()

	if err := n.Notify(ctx, obs.Event{
		Type:   obs.EventFillObserved,
		Fields: map[string]any{obs.FieldSymbol: "AAPL"},
	}); err != nil {
		t.Fatalf("an ordinary alert must never fail its caller: %v", err)
	}
	if count, _ := j.UndeliveredCount(ctx); count != 0 {
		t.Errorf("outbox rows = %d, want 0 — ordinary alerts are not durable", count)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Errorf("an ordinary alert must never block entries, got %v", rejected)
	}
	if pub.callCount() != 1 {
		t.Errorf("publish calls = %d, want exactly 1 — best-effort means no retries", pub.callCount())
	}
}

// TestCriticalAlertSurvivesAProcessRestart: the outbox is on disk, so a crash
// between the enqueue and the send does not lose the alert.
func TestCriticalAlertSurvivesAProcessRestart(t *testing.T) {
	clk := clock.NewFake(obsNow)
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	open := func() *journal.Journal {
		j, err := journal.Open(context.Background(), journal.Options{
			Path: path, Clock: clk,
			FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
		})
		if err != nil {
			t.Fatalf("journal.Open: %v", err)
		}
		return j
	}
	ctx := context.Background()

	first := open()
	pub := &failingPublisher{fail: true}
	n1 := &obs.Notifier{Journal: first, Publisher: pub, Clock: clock.System(), Attempts: 1,
		RetryDelay: time.Millisecond}
	if err := n1.Notify(ctx, obs.Event{Type: obs.EventOrderInDoubt, Key: "a-1"}); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	_ = first.Close()

	// New process, same file.
	second := open()
	defer second.Close()
	pending, err := second.PendingAlerts(ctx, 0)
	if err != nil {
		t.Fatalf("PendingAlerts: %v", err)
	}
	if len(pending) != 1 || pending[0].EventKey != "a-1" {
		t.Fatalf("pending after restart = %+v, want the enqueued alert", pending)
	}

	recovered := &failingPublisher{}
	n2 := &obs.Notifier{Journal: second, Publisher: recovered, Clock: clock.System(), Attempts: 1,
		RetryDelay: time.Millisecond}
	delivered, remaining, err := n2.Flush(ctx)
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if delivered != 1 || remaining != 0 {
		t.Fatalf("flush after restart = %d/%d, want 1/0", delivered, remaining)
	}
}

// TestCriticalWithoutAJournalIsLoudRatherThanSilent.
func TestCriticalWithoutAJournalIsLoudRatherThanSilent(t *testing.T) {
	var buf bytes.Buffer
	pub := &failingPublisher{}
	n := &obs.Notifier{
		Log:       obs.NewLogger(obs.LogOptions{Writer: &buf, JSON: true, Clock: clock.NewFake(obsNow)}),
		Publisher: pub,
	}
	if err := n.Notify(context.Background(), obs.Event{Type: obs.EventOrderInDoubt}); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if !strings.Contains(buf.String(), "not durable") {
		t.Errorf("log = %q, want a warning that the alert is not durable", buf.String())
	}
	if pub.callCount() != 1 {
		t.Errorf("publish calls = %d, want the alert still attempted", pub.callCount())
	}
}

// TestNotifierIsConcurrencySafe covers the -race requirement: fill detection,
// reconciliation and the gateway all notify from different goroutines.
func TestNotifierIsConcurrencySafe(t *testing.T) {
	pub := &failingPublisher{}
	n, _, _ := newNotifier(t, pub)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for k := 0; k < 20; k++ {
				_ = n.Notify(ctx, obs.Event{
					Type:   obs.EventOrderInDoubt,
					Key:    "key-" + string(rune('a'+i)),
					Fields: map[string]any{obs.FieldSymbol: "AAPL"},
				})
				_ = n.Notify(ctx, obs.Event{Type: obs.EventFillObserved})
			}
		}(i)
	}
	wg.Wait()
}

// --- heartbeat --------------------------------------------------------------

// TestHeartbeatPublishesOverHTTP is the httptest half.
func TestHeartbeatPublishesOverHTTP(t *testing.T) {
	pub, requests := ntfyServer(t, http.StatusOK)
	hb := &obs.Heartbeat{
		Publisher: pub,
		Clock:     clock.NewFake(obsNow),
		Status:    func() string { return "gate=off positions=0" },
	}

	if err := hb.Beat(context.Background()); err != nil {
		t.Fatalf("Beat: %v", err)
	}
	got := requests()
	if len(got) != 1 {
		t.Fatalf("requests = %d, want 1", len(got))
	}
	if got[0].body != "gate=off positions=0" {
		t.Errorf("body = %q", got[0].body)
	}
	// A heartbeat must arrive silently and be noticed by its absence.
	if prio := got[0].headers.Get("Priority"); prio != "1" {
		t.Errorf("Priority = %q, want 1", prio)
	}
}

// TestHeartbeatRunBeatsImmediatelyAndThenPeriodically. The first beat cannot wait
// for a full interval: an engine that starts and dies inside it would look
// identical to one that never started.
func TestHeartbeatRunBeatsImmediatelyAndThenPeriodically(t *testing.T) {
	pub, requests := ntfyServer(t, http.StatusOK)
	clk := clock.NewFake(obsNow)
	hb := &obs.Heartbeat{Publisher: pub, Clock: clk, Interval: time.Minute}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = hb.Run(ctx)
		close(done)
	}()

	// The first beat lands before any time passes.
	if !clk.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("the heartbeat never reached its first sleep")
	}
	if len(requests()) != 1 {
		t.Fatalf("beats before the first interval = %d, want 1", len(requests()))
	}

	clk.Advance(time.Minute)
	if !clk.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("the heartbeat never reached its second sleep")
	}
	if got := len(requests()); got != 2 {
		t.Fatalf("beats after one interval = %d, want 2", got)
	}

	cancel()
	<-done
}

// TestHeartbeatFailureIsNotAnIncident: a missed beat is normal, and escalating it
// would make the liveness signal the noisiest thing in the system.
func TestHeartbeatFailureIsNotAnIncident(t *testing.T) {
	pub, _ := ntfyServer(t, http.StatusInternalServerError)
	var buf bytes.Buffer
	hb := &obs.Heartbeat{
		Publisher: pub,
		Log:       obs.NewLogger(obs.LogOptions{Writer: &buf, JSON: true, Clock: clock.NewFake(obsNow)}),
		Clock:     clock.NewFake(obsNow),
	}

	if err := hb.Beat(context.Background()); err == nil {
		t.Fatal("Beat must report the failure to a caller that asks")
	}
	line := decodeLine(t, buf.String())
	if line[obs.FieldSeverity] != string(obs.SeverityNormal) {
		t.Errorf("severity = %v, want normal — a missed beat is not critical", line[obs.FieldSeverity])
	}
}

// --- helpers ----------------------------------------------------------------

func decodeLine(t *testing.T, out string) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("no log line was written")
	}
	var line map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &line); err != nil {
		t.Fatalf("parsing log line %q: %v", lines[0], err)
	}
	return line
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

func newDiscard() *discardWriter { return &discardWriter{} }
