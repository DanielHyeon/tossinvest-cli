package config

// a064 task 4: the alert transport becomes a config block.
//
// The rule every test here defends is §0.2: off is the default and off is
// indistinguishable from the build that had no block at all. The second rule is
// §0.8 — there is no field in this struct that can hold a secret, and that is a
// design choice rather than an omission.

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestAConfigWithNoNotificationsBlockIsUnchanged(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":4,"trading":{}}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got := cfg.Engine.Notifications
	if got != (Notifications{}) {
		t.Fatalf("notifications default = %+v, want the zero value", got)
	}
	if got.Enabled {
		t.Fatal("notifications must default to off")
	}
}

func TestNotificationsAreMerged(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":4,"trading":{},"engine":{"notifications":{
		"enabled":true,"base_url":"https://ntfy.example.internal","topic":"tossos-alerts"}}}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got := cfg.Engine.Notifications
	if !got.Enabled || got.BaseURL != "https://ntfy.example.internal" || got.Topic != "tossos-alerts" {
		t.Fatalf("notifications = %+v", got)
	}
	if got.Rejected != "" {
		t.Fatalf("a well-formed block was refused: %s", got.Rejected)
	}
}

func TestNotificationsRefuseANonHTTPBaseURL(t *testing.T) {
	svc := writeConfig(t, `{"schema_version":4,"trading":{},"engine":{"notifications":{
		"enabled":true,"base_url":"ftp://elsewhere","topic":"tossos-alerts"}}}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got := cfg.Engine.Notifications
	if got.Rejected == "" {
		t.Fatal("a non-HTTP base url was accepted")
	}
	// A refused block is zeroed, not partially kept — the same rule adoption has.
	// Half a transport configuration is not a safer transport configuration.
	if got.Enabled || got.BaseURL != "" || got.Topic != "" {
		t.Fatalf("a refused block was partially kept: %+v", got)
	}
}

func TestNotificationsMergeWithoutAnAutomationGate(t *testing.T) {
	// mergeEngine returns early when there is no automation_gate. The notification
	// branch has to run before that return, or an operator with the gate off gets
	// no alerts and no explanation.
	svc := writeConfig(t, `{"schema_version":4,"trading":{},"engine":{"notifications":{
		"enabled":true,"topic":"tossos-alerts"}}}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Engine.Notifications.Enabled || cfg.Engine.Notifications.Topic != "tossos-alerts" {
		t.Fatalf("notifications = %+v, want them merged without an automation gate",
			cfg.Engine.Notifications)
	}
}

func TestNotificationsHaveNoFieldForASecret(t *testing.T) {
	// §0.8, stated as a structural fact rather than a convention: a token written
	// into the config file has nowhere to land, so it cannot be read from there.
	svc := writeConfig(t, `{"schema_version":4,"trading":{},"engine":{"notifications":{
		"enabled":true,"topic":"tossos-alerts","token":"secret-value","auth":"secret-value"}}}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if strings.Contains(fmt.Sprintf("%+v", cfg.Engine.Notifications), "secret-value") {
		t.Fatalf("a token key in the file reached the parsed block: %+v", cfg.Engine.Notifications)
	}
	for i := 0; i < reflect.TypeOf(Notifications{}).NumField(); i++ {
		name := strings.ToLower(reflect.TypeOf(Notifications{}).Field(i).Name)
		if strings.Contains(name, "token") || strings.Contains(name, "secret") ||
			strings.Contains(name, "password") {
			t.Fatalf("Notifications has a field a secret could be written into: %s", name)
		}
	}
}

func TestConfigReadsNoEnvironment(t *testing.T) {
	// The resolution of file ⊕ environment belongs to the assembly point, so this
	// package's tests never depend on the process environment (design D4).
	t.Setenv("TOSSCTL_NTFY_TOPIC", "leaked-from-env")
	t.Setenv("TOSSCTL_NTFY_TOKEN", "leaked-from-env")
	svc := writeConfig(t, `{"schema_version":4,"trading":{},"engine":{"notifications":{
		"enabled":true,"topic":"from-the-file"}}}`)

	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Engine.Notifications.Topic != "from-the-file" {
		t.Fatalf("topic = %q, want the file's value untouched by the environment",
			cfg.Engine.Notifications.Topic)
	}
}
