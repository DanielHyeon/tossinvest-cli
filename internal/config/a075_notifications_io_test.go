package config

// a075_notifications_io_test.go holds the two closed member lists to their
// shapes, and the surgical write to everything it must not touch.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// notificationFixture writes a config with something in every other block, so a
// save that reached outside its own three keys has somewhere to be caught.
func notificationFixture(t *testing.T) (*Service, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	body := `{
  "schema_version": 5,
  "$unknown": {"kept": true},
  "trading": {"place": true, "note": "keep me"},
  "engine": {
    "automation_gate": {"enabled": true, "max_order_notional": 500000},
    "adoption": {"default_stop_pct": 0.05}
  }
}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return &Service{path: path}, path
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// TestSavingNotificationsWritesTheThreeKeys.
func TestSavingNotificationsWritesTheThreeKeys(t *testing.T) {
	svc, path := notificationFixture(t)
	if err := svc.SaveNotifications(Notifications{
		Enabled: true, BaseURL: "https://ntfy.sh", Topic: "tossos-abcdef",
	}); err != nil {
		t.Fatalf("SaveNotifications: %v", err)
	}

	after := readFile(t, path)
	for _, want := range []string{`"notifications"`, `"enabled": true`, `"base_url"`, `"topic"`,
		"https://ntfy.sh", "tossos-abcdef"} {
		if !strings.Contains(after, want) {
			t.Errorf("the save did not write %q:\n%s", want, after)
		}
	}

	block, err := svc.LoadRawNotifications()
	if err != nil {
		t.Fatalf("LoadRawNotifications: %v", err)
	}
	if !block.Enabled || block.Topic != "tossos-abcdef" || block.BaseURL != "https://ntfy.sh" {
		t.Errorf("the block did not round-trip: %+v", block)
	}
}

// TestANotificationSaveLeavesEveryOtherBlockAlone.
//
// This is the property limits_io.go's header is about: a save that re-emitted its
// neighbours would carry a value read outside the file lock, and on this file the
// neighbours include the automation gate's switch.
func TestANotificationSaveLeavesEveryOtherBlockAlone(t *testing.T) {
	svc, path := notificationFixture(t)
	if err := svc.SaveNotifications(Notifications{
		Enabled: true, BaseURL: "https://ntfy.sh", Topic: "tossos-abcdef",
	}); err != nil {
		t.Fatalf("SaveNotifications: %v", err)
	}
	after := readFile(t, path)
	for _, want := range []string{`"$unknown"`, "keep me", `"automation_gate"`, "500000",
		`"adoption"`, "0.05", `"schema_version"`} {
		if !strings.Contains(after, want) {
			t.Errorf("the notification save dropped %q from the file:\n%s", want, after)
		}
	}
}

// TestTurningNotificationsOffWritesOnlyTheSwitch.
//
// The channel byte-for-byte survives, because the operator's phone is subscribed
// to it and a mute must not cost a re-subscription.
func TestTurningNotificationsOffWritesOnlyTheSwitch(t *testing.T) {
	svc, path := notificationFixture(t)
	if err := svc.SaveNotifications(Notifications{
		Enabled: true, BaseURL: "https://ntfy.example", Topic: "tossos-keepme",
	}); err != nil {
		t.Fatalf("SaveNotifications: %v", err)
	}
	if err := svc.SaveNotificationsEnabled(false); err != nil {
		t.Fatalf("SaveNotificationsEnabled: %v", err)
	}

	after := readFile(t, path)
	if !strings.Contains(after, "tossos-keepme") {
		t.Errorf("turning alerts off erased the channel; re-enabling would break every "+
			"existing subscription:\n%s", after)
	}
	if !strings.Contains(after, "https://ntfy.example") {
		t.Errorf("turning alerts off erased the server address:\n%s", after)
	}
	block, err := svc.LoadRawNotifications()
	if err != nil {
		t.Fatalf("LoadRawNotifications: %v", err)
	}
	if block.Enabled {
		t.Error("the switch is still on after SaveNotificationsEnabled(false)")
	}
	if block.Topic != "tossos-keepme" {
		t.Errorf("the channel became %q; the off path must write one key", block.Topic)
	}
}

// TestTheOffPathEmitsExactlyOneKey and TestTheOnPathEmitsExactlyThree.
//
// Asserted on the member lists rather than on a rendered file, for the reason
// operating_io.go's header gives: a reviewer checks the list, not every caller.
func TestTheNotificationMemberListsAreClosed(t *testing.T) {
	off := notificationSwitch(false)
	if len(off) != 1 || off[0].key != "enabled" {
		t.Errorf("the off path emits %d key(s) %v; it must emit `enabled` and nothing else",
			len(off), keysOf(off))
	}

	on, err := notificationMembersOf(Notifications{Enabled: true, BaseURL: "x", Topic: "y"})
	if err != nil {
		t.Fatalf("notificationMembersOf: %v", err)
	}
	want := map[string]bool{"enabled": true, "base_url": true, "topic": true}
	if len(on) != len(want) {
		t.Errorf("the on path emits %v, want exactly %d keys", keysOf(on), len(want))
	}
	for _, m := range on {
		if !want[m.key] {
			t.Errorf("the on path emits %q, which is outside its declared three keys", m.key)
		}
	}
	// Two keys that must never appear: a token has no config home at all (a074),
	// and `rejected` is this parser's own diagnostic — emitting it would let the
	// next load read an opinion back as configuration.
	for _, m := range on {
		if m.key == "token" || m.key == "rejected" {
			t.Errorf("the on path emits %q", m.key)
		}
	}
}

func keysOf(members []gateMember) []string {
	out := make([]string, 0, len(members))
	for _, m := range members {
		out = append(out, m.key)
	}
	return out
}

// TestNotificationsAreReadableFromAFileThatHasNoBlock, and from no file at all.
func TestNotificationsAreReadableWhenNothingIsConfigured(t *testing.T) {
	svc, _ := notificationFixture(t)
	block, err := svc.LoadRawNotifications()
	if err != nil {
		t.Fatalf("LoadRawNotifications on a file with no block: %v", err)
	}
	if block.Enabled || block.Topic != "" {
		t.Errorf("a file with no notifications block read as %+v", block)
	}

	missing := &Service{path: filepath.Join(t.TempDir(), "absent.json")}
	block, err = missing.LoadRawNotifications()
	if err != nil {
		t.Fatalf("a missing file is 꺼짐, not an error the screen cannot act on: %v", err)
	}
	if block.Enabled {
		t.Error("a missing file read as enabled")
	}
}

// TestANotificationSaveCreatesTheBlockAndTheEngineBlock.
func TestANotificationSaveCreatesTheBlocksItNeeds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"schema_version": 5}`), 0o600); err != nil {
		t.Fatal(err)
	}
	svc := &Service{path: path}
	if err := svc.SaveNotifications(Notifications{
		Enabled: true, BaseURL: "https://ntfy.sh", Topic: "tossos-fresh",
	}); err != nil {
		t.Fatalf("SaveNotifications into a file with no engine block: %v", err)
	}
	block, err := svc.LoadRawNotifications()
	if err != nil {
		t.Fatalf("LoadRawNotifications: %v", err)
	}
	if !block.Enabled || block.Topic != "tossos-fresh" {
		t.Errorf("the created block reads as %+v", block)
	}
	if !strings.Contains(readFile(t, path), `"schema_version"`) {
		t.Error("creating the block dropped schema_version")
	}
}
