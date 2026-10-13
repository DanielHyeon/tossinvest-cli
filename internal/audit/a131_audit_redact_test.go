package audit_test

// a131_audit_redact_test.go — 감사 기록은 계좌 텍스트만 치환하고 의미(행위·설정·이전/이후 값·시각·주체)는 보존(design D3-a).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
)

const a131AuditAccount = "987-65-AUDITSENTINEL"

func init() { redact.Register(a131AuditAccount) }

func TestA131AuditRecordReplacesOnlyAccountText(t *testing.T) {
	path := filepath.Join(t.TempDir(), audit.FileName)
	log, err := audit.Open(audit.Options{Path: path, Subject: "operator"})
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	if err := log.Record(audit.Entry{At: at, Action: audit.ActionGateAccepted, Setting: "engine.automation_gate.enabled",
		Old: "false", New: "true", Detail: "account=" + a131AuditAccount + " attestation_expires=2026-10-23 masked=*******9999"}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	if strings.Contains(string(raw), a131AuditAccount) || strings.Contains(string(raw), "*******9999") {
		t.Fatalf("audit line leaked the account: %s", raw)
	}
	entries, err := log.Entries()
	if err != nil || len(entries) != 1 {
		t.Fatalf("Entries = %v, %v", entries, err)
	}
	e := entries[0]
	if e.Action != audit.ActionGateAccepted || e.Setting != "engine.automation_gate.enabled" || e.Old != "false" ||
		e.New != "true" || e.Subject != "operator" || !e.At.Equal(at) {
		t.Fatalf("audit meaning changed: %+v", e)
	}
	if !strings.Contains(e.Detail, "attestation_expires=2026-10-23") || !strings.Contains(e.Detail, "account=[account]") {
		t.Fatalf("detail lost its non-account content or the placeholder: %q", e.Detail)
	}
}

func TestA131RecordChangeStillDeduplicates(t *testing.T) {
	log, err := audit.Open(audit.Options{Path: filepath.Join(t.TempDir(), audit.FileName), Subject: "operator"})
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []bool{true, false, true} {
		value := "100"
		if i == 2 {
			value = "200"
		}
		wrote, err := log.RecordChange(audit.ActionLimitChange, "engine.automation_gate.max_order_quantity", value,
			"account="+a131AuditAccount)
		if err != nil || wrote != want {
			t.Fatalf("RecordChange #%d = %v, %v; want %v", i, wrote, err, want)
		}
	}
}

func TestA131RedactEntryIsTheRecordRule(t *testing.T) {
	e := audit.RedactEntry(audit.Entry{Setting: "s", Old: a131AuditAccount, New: "n", Detail: "d " + a131AuditAccount, Subject: "u"})
	if strings.Contains(e.Old+e.Detail, a131AuditAccount) || e.Setting != "s" || e.New != "n" || e.Subject != "u" {
		t.Fatalf("RedactEntry = %+v", e)
	}
}
