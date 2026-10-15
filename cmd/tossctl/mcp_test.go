package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/session"
)

func TestMCPOfficialEligibilityMatchesCLI(t *testing.T) {
	creds := &official.Credentials{APIKey: "k", SecretKey: "s"}
	for _, tc := range []struct {
		name    string
		creds   *official.Credentials
		enabled bool
		prefer  string
		want    bool
	}{
		{name: "enabled auto", creds: creds, enabled: true, prefer: "auto", want: true},
		{name: "enabled openapi", creds: creds, enabled: true, prefer: "openapi", want: true},
		{name: "wts denies stored credentials", creds: creds, enabled: true, prefer: "wts"},
		{name: "disabled denies stored credentials", creds: creds, enabled: false, prefer: "openapi"},
		{name: "missing credentials", enabled: true, prefer: "auto"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := mcpOfficialEligible(tc.creds, config.OpenAPI{Enabled: tc.enabled}, tc.prefer)
			if got != tc.want {
				t.Fatalf("mcpOfficialEligible() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMCPCmdWTSDoesNotWireOfficialDependencies(t *testing.T) {
	dir := isolate(t)
	cfg := config.DefaultFile()
	cfg.OpenAPI = config.OpenAPI{Enabled: true, Prefer: "wts"}
	writeConfigFile(t, dir, cfg)
	writeCredentials(t, dir)
	if err := session.WriteFile(filepath.Join(dir, "session.json"), &session.Session{Cookies: map[string]string{"SESSION": "fixture"}}); err != nil {
		t.Fatal(err)
	}

	lines := []string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"call_operation","arguments":{"operation":"auth_status"}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"call_operation","arguments":{"operation":"accounts"}}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"call_operation","arguments":{"operation":"place_order","params":{"symbol":"AAPL","side":"buy","market":"us","quantity":1,"price":1,"currency_mode":"USD"}}}}`,
	}
	cmd := newMCPCmd(&rootOptions{configDir: dir})
	cmd.SetIn(strings.NewReader(strings.Join(lines, "\n") + "\n"))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mcp command: %v", err)
	}

	var responses []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		var response map[string]any
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			t.Fatal(err)
		}
		responses = append(responses, response)
	}
	if len(responses) != 3 {
		t.Fatalf("responses = %d, want 3", len(responses))
	}
	text := func(t *testing.T, response map[string]any) (string, bool) {
		t.Helper()
		result := response["result"].(map[string]any)
		return result["content"].([]any)[0].(map[string]any)["text"].(string), result["isError"].(bool)
	}
	statusText, statusError := text(t, responses[0])
	var status struct {
		WTS      struct{ Connected bool } `json:"wts"`
		Official struct{ Connected bool } `json:"official"`
	}
	if err := json.Unmarshal([]byte(statusText), &status); err != nil {
		t.Fatal(err)
	}
	if statusError || !status.WTS.Connected || status.Official.Connected {
		t.Fatalf("auth status = %s, want WTS only", statusText)
	}
	for _, response := range responses[1:] {
		if _, isError := text(t, response); !isError {
			t.Fatalf("official operation unexpectedly available: %v", response)
		}
	}
}
