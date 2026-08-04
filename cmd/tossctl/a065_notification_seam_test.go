package main

// a065_notification_seam_test.go pins the wiring end: what the button actually
// writes, what it writes only once, and what never reaches the audit log.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// notificationSeamFixture gives the seam a real config file and a real audit
// directory, and a publish stub so no test message leaves the machine.
func notificationSeamFixture(t *testing.T) (*consoleNotifications, string, string) {
	t.Helper()
	cfgDir := t.TempDir()
	dataDir := t.TempDir()
	t.Setenv("TOSSOS_DATA_DIR", dataDir)

	body := `{
  "schema_version": 5,
  "trading": {"note": "keep me"},
  "engine": {"automation_gate": {"enabled": false}}
}`
	path := filepath.Join(cfgDir, "config.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	seam := newNotificationSeam(&rootOptions{configDir: cfgDir})
	if seam == nil {
		t.Fatal("the seam did not resolve a --config-dir profile")
	}
	seam.publish = func(context.Context, config.Notifications, string) error { return nil }
	seam.getenv = func(string) string { return "" }
	return seam, path, filepath.Join(dataDir, "audit.log")
}

// TestEnablingCreatesAChannelAndWritesTheThreeKeys.
func TestEnablingCreatesAChannelAndWritesTheThreeKeys(t *testing.T) {
	seam, path, _ := notificationSeamFixture(t)

	block, err := seam.Enable()
	if err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if !block.Enabled {
		t.Error("Enable returned a block that is not enabled")
	}
	if !strings.HasPrefix(block.Topic, "tossos-") || len(block.Topic) < 30 {
		t.Errorf("Enable returned the channel %q; it must be the generated one", block.Topic)
	}
	if block.BaseURL == "" {
		t.Error("Enable left the server address empty; the file should say where alerts go")
	}

	// It is on disk, and the neighbours survived the surgical write.
	after, _ := os.ReadFile(path)
	for _, want := range []string{block.Topic, `"notifications"`, "keep me", `"automation_gate"`} {
		if !strings.Contains(string(after), want) {
			t.Errorf("the file does not carry %q after Enable:\n%s", want, after)
		}
	}
}

// TestEnablingAgainKeepsTheSameChannel.
//
// The operator who turns alerts off for an afternoon and back on must not have to
// re-subscribe every device. A second Enable that minted a new channel would make
// every existing subscription silently stop receiving — the worst failure this
// feature can have, because it looks exactly like working.
func TestEnablingAgainKeepsTheSameChannel(t *testing.T) {
	seam, _, _ := notificationSeamFixture(t)

	first, err := seam.Enable()
	if err != nil {
		t.Fatalf("first Enable: %v", err)
	}
	if err := seam.Disable(); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	second, err := seam.Enable()
	if err != nil {
		t.Fatalf("second Enable: %v", err)
	}
	if second.Topic != first.Topic {
		t.Errorf("re-enabling minted a new channel (%q → %q); every subscribed device "+
			"would go quiet without saying so", first.Topic, second.Topic)
	}
}

// TestDisablingKeepsTheChannelBytes.
func TestDisablingKeepsTheChannelBytes(t *testing.T) {
	seam, path, _ := notificationSeamFixture(t)
	block, err := seam.Enable()
	if err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if err := seam.Disable(); err != nil {
		t.Fatalf("Disable: %v", err)
	}

	after, _ := os.ReadFile(path)
	if !strings.Contains(string(after), block.Topic) {
		t.Errorf("Disable erased the channel from the file:\n%s", after)
	}
	current, err := seam.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if current.Enabled {
		t.Error("Disable left the switch on")
	}
	if current.Topic != block.Topic {
		t.Errorf("Disable changed the channel to %q", current.Topic)
	}
}

// TestTheAuditTrailCarriesNoChannelValue.
//
// a064 set the contract — the audit line says whether a channel exists, never
// which one — and a065 is the first change that exercises it several times a
// session. The channel is the access control on a public ntfy topic, so an
// append-only operations log holding it is that credential written to a log.
func TestTheAuditTrailCarriesNoChannelValue(t *testing.T) {
	seam, _, auditPath := notificationSeamFixture(t)
	block, err := seam.Enable()
	if err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if err := seam.Disable(); err != nil {
		t.Fatalf("Disable: %v", err)
	}

	data, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("no audit entry was appended: %v", err)
	}
	trail := string(data)
	if strings.Contains(trail, block.Topic) {
		t.Errorf("the audit log carries the channel value:\n%s", trail)
	}
	for _, want := range []string{
		"engine.notifications.enabled",
		"engine.notifications.base_url",
		"engine.notifications.topic_configured",
	} {
		if !strings.Contains(trail, want) {
			t.Errorf("the audit log does not record %q:\n%s", want, trail)
		}
	}
	// The record is a fact about existence, so it reads as a boolean.
	if !strings.Contains(trail, `"topic_configured"`) && !strings.Contains(trail, "topic_configured") {
		t.Error("topic_configured is missing from the trail")
	}
}

// TestTheTestSendCarriesTheEnvironmentToken.
//
// The self-hosted path a064 built stays whole: an operator behind a token gets a
// working test button without the console ever offering a field for it.
func TestTheTestSendCarriesTheEnvironmentToken(t *testing.T) {
	seam, _, _ := notificationSeamFixture(t)
	if _, err := seam.Enable(); err != nil {
		t.Fatalf("Enable: %v", err)
	}

	var sawToken string
	var sawTopic string
	seam.publish = func(_ context.Context, block config.Notifications, token string) error {
		sawToken, sawTopic = token, block.Topic
		return nil
	}
	seam.getenv = func(name string) string {
		if name == envNtfyToken {
			return "tk_secret"
		}
		return ""
	}
	if err := seam.Test(context.Background()); err != nil {
		t.Fatalf("Test: %v", err)
	}
	if sawToken != "tk_secret" {
		t.Errorf("the test send carried the token %q, want the environment's", sawToken)
	}
	if !strings.HasPrefix(sawTopic, "tossos-") {
		t.Errorf("the test send used the channel %q", sawTopic)
	}
}

// TestTheTestButtonRefusesWhenThereIsNoChannel.
func TestTheTestButtonRefusesWhenThereIsNoChannel(t *testing.T) {
	seam, _, _ := notificationSeamFixture(t)
	seam.publish = nil
	err := seam.Test(context.Background())
	if err == nil {
		t.Fatal("Test with no channel configured returned no error")
	}
	if !errors.Is(err, errNoNotificationChannel) {
		t.Errorf("Test refused with %v, want the no-channel reason", err)
	}
}

// TestAFailedTestDoesNotUnwindTheSetting.
//
// Asserted at the seam as well as at the screen: there is no rollback path here
// to accidentally grow, because Test does not write.
func TestAFailedTestDoesNotUnwindTheSetting(t *testing.T) {
	seam, _, _ := notificationSeamFixture(t)
	block, err := seam.Enable()
	if err != nil {
		t.Fatalf("Enable: %v", err)
	}
	seam.publish = func(context.Context, config.Notifications, string) error {
		return errors.New("dial tcp: connection refused")
	}
	if err := seam.Test(context.Background()); err == nil {
		t.Fatal("the failing publish was reported as success")
	}
	current, err := seam.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !current.Enabled || current.Topic != block.Topic {
		t.Errorf("a failed test changed the setting to %+v", current)
	}
}

// TestTheSeamIsNilRatherThanATypedNilWhenThereIsNoProfile.
func TestTheNotificationSeamIsAbsentRatherThanTypedNil(t *testing.T) {
	if seam := consoleNotificationSeam(&rootOptions{configDir: ""}); seam != nil {
		if _, ok := seam.(*consoleNotifications); ok {
			t.Skip("this environment resolves a default profile; the typed-nil path is not reachable here")
		}
		t.Errorf("consoleNotificationSeam returned %T for an unresolvable profile", seam)
	}
}
