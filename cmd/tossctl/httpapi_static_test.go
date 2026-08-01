package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHTTPAPIDaemonSourceHasNoEngineLifecycleOrTradingAuthority(t *testing.T) {
	for _, name := range []string{"httpapi.go", "httpapi_reader.go"} {
		body, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		code := string(body)
		for _, forbidden := range []string{
			"runConfiguredEngineAutostart(", "startEngine(", "stopEngine(",
			"PlaceOrder(", "CancelOrder(", "ModifyOrder(",
		} {
			if strings.Contains(code, forbidden) {
				t.Errorf("%s contains forbidden authority %q", name, forbidden)
			}
		}
	}
}

func TestHTTPAPIProductionPerformanceLineageIsPinnedToReadOnlyJournal(t *testing.T) {
	body, err := os.ReadFile("httpapi.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "performancejournal.New(journalReader)") {
		t.Fatal("production HTTP API does not pin performance lineage to journal.ReadOnly")
	}
	reader, err := os.ReadFile("httpapi_reader.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(reader), "performanceSource performance.JournalLineageReader") {
		t.Fatal("production reader lost the read-only performance lineage interface")
	}
}

func TestHTTPAPIAddsNoArbitraryInputUI(t *testing.T) {
	for _, root := range []string{"../../internal/httpapi", "."} {
		entries, err := os.ReadDir(root)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasPrefix(entry.Name(), "httpapi") || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}
			body, err := os.ReadFile(filepath.Join(root, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			lower := strings.ToLower(string(body))
			for _, forbidden := range []string{"<input", "<textarea", "contenteditable", "type=\"number\"", "type=\"range\""} {
				if strings.Contains(lower, forbidden) {
					t.Errorf("%s introduces arbitrary-input UI %q", entry.Name(), forbidden)
				}
			}
		}
	}
}

func TestComposeDeploysHTTPAPISeparatelyOnVPNWithoutAutostart(t *testing.T) {
	body, err := os.ReadFile("../../compose.yaml")
	if err != nil {
		t.Fatal(err)
	}
	compose := string(body)
	for _, required := range []string{
		"  httpapi:\n", "- httpapi\n", "${TOSSOS_VPN_BIND_IP:?set the host VPN IP}:${TOSSOS_API_PORT:-37086}:37086",
		"Host: $$api_host", "https://${TOSSOS_HTTPAPI_CONTAINER_IP:-172.30.85.86}:37086/api/v1/engine", "TOSSOS_API_PUBLIC_URL",
	} {
		if !strings.Contains(compose, required) {
			t.Errorf("compose lacks %q", required)
		}
	}
	httpapiSection := strings.SplitN(compose, "  httpapi:\n", 2)[1]
	if !strings.Contains(httpapiSection, "--bind\n      - ${TOSSOS_HTTPAPI_CONTAINER_IP:-172.30.85.86}") {
		t.Fatal("httpapi service does not bind its explicit private container address")
	}
	for _, forbidden := range []string{"console", "engine run", "engine-start", "--trusted-network"} {
		if strings.Contains(httpapiSection, forbidden) {
			t.Errorf("httpapi service contains %q", forbidden)
		}
	}
	entrypoint, err := os.ReadFile("../../deploy/container-entrypoint.sh")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(entrypoint), "/run/secrets/remote-token") {
		t.Fatal("entrypoint still requires the retired remote-token secret")
	}
}
