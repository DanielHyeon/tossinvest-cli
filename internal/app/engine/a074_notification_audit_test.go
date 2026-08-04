package engine

// a074 task 6: the notification settings become auditable.
//
// engine.go's Publisher field said outright why the transport had no config
// block: "nothing in this change audits notification settings, and an
// operational setting with no audit trail is the thing §0.5 exists to prevent."
// These tests are the other half of that sentence being satisfied — and the
// constraint on how, which is §0.8: the trail records that a channel and a
// credential exist, never what they are.

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

func auditFixture(t *testing.T) *audit.Log {
	t.Helper()
	log, err := audit.Open(audit.Options{
		Path:    filepath.Join(t.TempDir(), "audit.log"),
		Clock:   clock.NewFake(time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)),
		Subject: "test-operator",
	})
	if err != nil {
		t.Fatalf("audit.Open: %v", err)
	}
	return log
}

func settingsFrom(t *testing.T, log *audit.Log) map[string]audit.Entry {
	t.Helper()
	entries, err := log.Entries()
	if err != nil {
		t.Fatalf("Entries: %v", err)
	}
	out := map[string]audit.Entry{}
	for _, e := range entries {
		out[e.Setting] = e
	}
	return out
}

func TestNotificationSettingsAreAudited(t *testing.T) {
	log := auditFixture(t)
	_, resolution := resolveNotificationPublisher(config.Notifications{
		Enabled: true, BaseURL: "https://ntfy.example.internal", Topic: "tossos-alerts",
	}, fixedEnv(map[string]string{envNtfyToken: "s3cret"}))

	if err := recordGateSettings(log, config.AutomationGate{}, config.Adoption{},
		resolution, "/tmp/attestation.json"); err != nil {
		t.Fatalf("recordGateSettings: %v", err)
	}

	settings := settingsFrom(t, log)
	for setting, want := range map[string]string{
		"engine.notifications.enabled":          "true",
		"engine.notifications.base_url":         "https://ntfy.example.internal",
		"engine.notifications.topic_configured": "true",
		"engine.notifications.token_configured": "true",
	} {
		entry, ok := settings[setting]
		if !ok {
			t.Errorf("%s was not audited", setting)
			continue
		}
		if entry.New != want {
			t.Errorf("%s = %q, want %q", setting, entry.New, want)
		}
	}
}

func TestTheAuditTrailCarriesNoNotificationSecret(t *testing.T) {
	log := auditFixture(t)
	_, resolution := resolveNotificationPublisher(config.Notifications{
		Enabled: true, Topic: "top-secret-topic",
	}, fixedEnv(map[string]string{envNtfyToken: "top-secret-token"}))

	if err := recordGateSettings(log, config.AutomationGate{}, config.Adoption{},
		resolution, ""); err != nil {
		t.Fatalf("recordGateSettings: %v", err)
	}

	entries, err := log.Entries()
	if err != nil {
		t.Fatalf("Entries: %v", err)
	}
	for _, e := range entries {
		line := strings.ToLower(e.Setting + "|" + e.Old + "|" + e.New + "|" + e.Detail)
		if strings.Contains(line, "top-secret") {
			t.Fatalf("the audit trail carries a notification secret: %+v", e)
		}
	}
}

func TestARefusedNotificationBlockIsAudited(t *testing.T) {
	log := auditFixture(t)
	_, resolution := resolveNotificationPublisher(config.Notifications{Enabled: true}, fixedEnv(nil))

	if err := recordGateSettings(log, config.AutomationGate{}, config.Adoption{},
		resolution, ""); err != nil {
		t.Fatalf("recordGateSettings: %v", err)
	}

	entry, ok := settingsFrom(t, log)["engine.notifications.enabled"]
	if !ok {
		t.Fatal("a refused block left no audit entry at all")
	}
	if !strings.Contains(entry.Detail, "no topic") {
		t.Fatalf("the refusal reason is not in the trail: %q", entry.Detail)
	}
}

func TestExistingAuditEntriesKeepTheirOrder(t *testing.T) {
	// The trail's value is that two engines' startups can be read side by side.
	// New settings append; they never reorder or rename what is already there.
	log := auditFixture(t)
	_, resolution := resolveNotificationPublisher(config.Notifications{}, fixedEnv(nil))
	if err := recordGateSettings(log, config.AutomationGate{LimitCurrency: "KRW"},
		config.Adoption{}, resolution, "/tmp/attestation.json"); err != nil {
		t.Fatalf("recordGateSettings: %v", err)
	}

	entries, err := log.Entries()
	if err != nil {
		t.Fatalf("Entries: %v", err)
	}
	want := []string{
		"engine.automation_gate.enabled",
		"engine.adoption.enabled",
		"engine.adoption.default_stop_pct",
		"engine.adoption.exclude_symbols",
		"engine.adoption.include_symbols",
		"engine.automation_gate.max_order_quantity",
		"engine.automation_gate.max_order_notional",
		"engine.automation_gate.max_total_exposure",
		"engine.automation_gate.max_daily_loss_amount",
		"engine.automation_gate.max_daily_loss_ratio",
		"engine.automation_gate.limit_currency",
	}
	if len(entries) < len(want) {
		t.Fatalf("entries = %d, want at least the %d landed settings", len(entries), len(want))
	}
	for i, setting := range want {
		if entries[i].Setting != setting {
			t.Fatalf("entry %d = %q, want %q — the landed order moved", i, entries[i].Setting, setting)
		}
	}
}
