package engine_test

// a064 tasks 5 and 7: the transport reaches the Notifier, and a delivered
// critical alert is actually marked delivered.
//
// The operational ledger on 2026-08-04 held six critical alerts, all PENDING,
// all with attempts=0. Zero attempts is the whole diagnosis: Notifier.deliver
// breaks out of its retry loop before recording an attempt when Publisher is
// nil, so nothing had failed to send — nothing had been sent.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// writeNotificationConfig writes the standard engine config plus a notifications
// block. It is separate from writeEngineConfig so a064 adds no field to the
// landed fixture.
func writeNotificationConfig(t *testing.T, dir string, notifications config.Notifications) {
	t.Helper()
	cfg := config.DefaultFile()
	cfg.OpenAPI = config.OpenAPI{Enabled: false, Prefer: "wts", Fallback: true}
	cfg.Trading = config.Trading{
		Place: true, Sell: true, Fractional: true, Cancel: true, Amend: true,
		Conditional: true, AllowLiveOrderActions: true,
	}
	cfg.Engine.Notifications = notifications
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func startEngineWithNotifications(t *testing.T, notifications config.Notifications,
	env map[string]string, decorate func(*engine.Options)) *engine.Context {
	t.Helper()
	dir := isolate(t)
	writeNotificationConfig(t, dir, notifications)
	writeCredentials(t, dir, "key", "secret")
	srv, _ := engineStub(t, "18301045921")

	opts := engine.Options{ConfigDir: dir}
	opts.OfficialOptions = []official.Option{
		official.WithBaseURL(srv.URL), official.WithHTTPClient(srv.Client()),
	}
	opts.SetJournalProberForTest(journal.FixedFSProber(journal.FSInfo{
		Name: "ext4", Magic: journal.MagicExt,
	}))
	if env != nil {
		opts.Getenv = func(key string) string { return env[key] }
	}
	if decorate != nil {
		decorate(&opts)
	}
	eng, err := engine.New(opts)
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	t.Cleanup(func() { _ = eng.Close() })
	return eng
}

func TestStartupWiresNoPublisherWhenNotificationsAreOff(t *testing.T) {
	// §0.2: this is every config anyone has, and it must produce exactly the
	// build's pre-a064 behaviour.
	eng := startEngineWithNotifications(t, config.Notifications{}, nil, nil)
	if eng.Notifier == nil {
		t.Fatal("no notifier was built")
	}
	if eng.Notifier.Publisher != nil {
		t.Fatalf("an unconfigured engine wired a transport: %#v", eng.Notifier.Publisher)
	}
}

func TestStartupWiresTheConfiguredPublisher(t *testing.T) {
	eng := startEngineWithNotifications(t, config.Notifications{
		Enabled: true, BaseURL: "https://ntfy.example.internal", Topic: "tossos-alerts",
	}, map[string]string{"TOSSCTL_NTFY_TOKEN": "s3cret"}, nil)

	ntfy, ok := eng.Notifier.Publisher.(*obs.Ntfy)
	if !ok {
		t.Fatalf("publisher = %#v, want an *obs.Ntfy", eng.Notifier.Publisher)
	}
	if ntfy.Topic != "tossos-alerts" {
		t.Fatalf("topic = %q", ntfy.Topic)
	}
}

func TestARefusedNotificationBlockDoesNotBlockStartup(t *testing.T) {
	// Design D7. Alert delivery is not a protection path: an engine that refuses
	// to start over a typo in a topic is an engine whose stop evaluation has also
	// stopped, which is strictly worse than an engine with no alerts.
	eng := startEngineWithNotifications(t, config.Notifications{
		Enabled: true, BaseURL: "ftp://elsewhere", Topic: "tossos-alerts",
	}, nil, nil)

	if eng.Notifier.Publisher != nil {
		t.Fatalf("a refused block produced a transport: %#v", eng.Notifier.Publisher)
	}
	entries, err := eng.Audit.Entries()
	if err != nil {
		t.Fatalf("Entries: %v", err)
	}
	found := false
	for _, e := range entries {
		if e.Setting == "engine.notifications.enabled" && strings.Contains(e.Detail, "refused") {
			found = true
		}
	}
	if !found {
		t.Fatal("the refusal left no reason in the audit trail")
	}
}

func TestAnInjectedPublisherWins(t *testing.T) {
	// The package-test seam: a test that injected a transport must not have the
	// config file take it away, or these tests become a function of a file.
	injected := &countingPublisher{}
	eng := startEngineWithNotifications(t, config.Notifications{
		Enabled: true, Topic: "tossos-alerts",
	}, nil, func(o *engine.Options) { o.Publisher = injected })

	if eng.Notifier.Publisher != obs.Publisher(injected) {
		t.Fatalf("publisher = %#v, want the injected one", eng.Notifier.Publisher)
	}
}

// countingPublisher records what it was asked to send.
type countingPublisher struct {
	mu   sync.Mutex
	sent []obs.Notification
	fail error
}

func (p *countingPublisher) Publish(_ context.Context, n obs.Notification) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.fail != nil {
		return p.fail
	}
	p.sent = append(p.sent, n)
	return nil
}

func (p *countingPublisher) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.sent)
}

func TestAWiredPublisherDeliversTheCriticalAlert(t *testing.T) {
	// Task 7.1: outbox first, then send, then delivered. The row is what a
	// restart still knows about; the send is what a human reads.
	publisher := &countingPublisher{}
	eng := startEngineWithNotifications(t, config.Notifications{}, nil,
		func(o *engine.Options) { o.Publisher = publisher })

	ctx := context.Background()
	if err := eng.Notifier.Notify(ctx, obs.Event{
		Type:  obs.EventExitSnapshotQuarantined,
		Key:   "exit.snapshot_quarantined|pos-1|1|1",
		Title: "005930 is no longer being judged",
		Body:  "ambiguous_recovery",
	}); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	if publisher.count() != 1 {
		t.Fatalf("publishes = %d, want one", publisher.count())
	}
	remaining, err := eng.Journal.UndeliveredCount(ctx)
	if err != nil {
		t.Fatalf("UndeliveredCount: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("undelivered = %d, want the alert marked delivered", remaining)
	}
}

func TestTheNotificationBodyCarriesNoToken(t *testing.T) {
	// §0.8. The token belongs in the Authorization header the transport builds,
	// and nowhere a message body or a log line can reach.
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		seen = append(seen, string(body), r.Header.Get("Title"), r.Header.Get("Tags"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ntfy := &obs.Ntfy{BaseURL: srv.URL, Topic: "t", Token: "top-secret-token", HTTPClient: srv.Client()}
	if err := ntfy.Publish(context.Background(), obs.Notification{
		Type: obs.EventExitSnapshotQuarantined, Severity: obs.SeverityCritical,
		Title: "005930", Body: "ambiguous_recovery",
	}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	for _, part := range seen {
		if strings.Contains(part, "top-secret-token") {
			t.Fatalf("the token reached the message: %q", part)
		}
	}
}
