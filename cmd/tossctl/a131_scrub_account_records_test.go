package main

// a131_scrub_account_records_test.go — 일회성 정리 명령(design D5)과 soak attest 출력(L10). 모든 파일은 testenv 격리 디렉터리 —
// 운영 파일에 쓰지 않음.

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

const a131ScrubAccount = "98765432155"

var a131ScrubMask = attest.Mask(a131ScrubAccount)

type a131Fixture struct {
	configDir, soakPath, verifyPath, auditPath, attestPath string
}

// a131LegacyFiles 는 a131 이전 판본이 쓴 모양의 파일을 만듦 — 평문 계좌(soak) · 마스크(verify) · detail 의 계좌(audit) · attestation.
func a131LegacyFiles(t *testing.T) a131Fixture {
	t.Helper()
	configDir := testenv.Isolate(t)
	f := a131Fixture{
		configDir:  configDir,
		soakPath:   filepath.Join(configDir, soak.FileName),
		verifyPath: filepath.Join(configDir, verifylive.RecordFileName(verifylive.MarketKR)),
		auditPath:  filepath.Join(testenv.DataDir(t), audit.FileName),
		attestPath: filepath.Join(configDir, attest.FileName),
	}
	at := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	var soakLines []byte
	for i := 0; i < 3; i++ {
		c := soak.Cycle{FormatVersion: soak.RecordFormatVersion, Kind: "cycle", StartedAt: at.Add(time.Duration(i) * time.Hour),
			FinishedAt: at, AccountRef: a131ScrubAccount}
		line, _ := json.Marshal(c)
		soakLines = append(append(soakLines, line...), '\n')
	}
	clean, _ := json.Marshal(soak.Cycle{FormatVersion: soak.RecordFormatVersion, Kind: "cycle", StartedAt: at})
	soakLines = append(append(soakLines, clean...), '\n')
	writeFile(t, f.soakPath, soakLines)

	var verifyLines []byte
	for i := 0; i < 2; i++ {
		line, _ := json.Marshal(verifylive.Entry{FormatVersion: verifylive.RecordFormatVersion, Kind: verifylive.KindStep,
			RunID: "r", StepID: "s", StartedAt: at, FinishedAt: at, AccountRef: a131ScrubMask, Verdict: verifylive.VerdictPass})
		verifyLines = append(append(verifyLines, line...), '\n')
	}
	writeFile(t, f.verifyPath, verifyLines)

	var auditLines []byte
	for _, e := range []audit.Entry{
		{At: at, Subject: "op", Action: audit.ActionGateAccepted, Setting: "engine.automation_gate.enabled", Old: "false", New: "true",
			Detail: "account=" + a131ScrubAccount + " attestation_expires=2026-10-23"},
		{At: at, Subject: "op", Action: audit.ActionLimitChange, Setting: "engine.automation_gate.max_order_quantity", Old: "1", New: "2"},
	} {
		line, _ := json.Marshal(e)
		auditLines = append(append(auditLines, line...), '\n')
	}
	writeFile(t, f.auditPath, auditLines)

	writeFile(t, f.attestPath, []byte(`{"format_version":1,"account_ref":"`+a131ScrubAccount+`"}`+"\n"))
	return f
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

type fileState struct {
	data  string
	mtime time.Time
}

func snapshot(t *testing.T, dirs ...string) map[string]fileState {
	t.Helper()
	out := map[string]fileState{}
	for _, dir := range dirs {
		_ = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				b, _ := os.ReadFile(p)
				out[p] = fileState{string(b), info.ModTime()}
			}
			return nil
		})
	}
	return out
}

func TestA131ScrubPreviewWritesNothing(t *testing.T) {
	f := a131LegacyFiles(t)
	before := snapshot(t, f.configDir, filepath.Dir(f.auditPath))
	out, _, err := runCLI(t, "--config-dir", f.configDir, "scrub-account-records")
	if err != nil {
		t.Fatalf("preview: %v\n%s", err, out)
	}
	after := snapshot(t, f.configDir, filepath.Dir(f.auditPath))
	if len(after) != len(before) {
		t.Fatalf("preview created or removed files: %d → %d", len(before), len(after))
	}
	for p, st := range before {
		if after[p] != st {
			t.Fatalf("preview changed %s", p)
		}
	}
	if _, err := os.Stat(filepath.Join(f.configDir, redact.KeyFileName)); !os.IsNotExist(err) {
		t.Fatalf("preview created the seal key: %v", err)
	}
	for _, want := range []string{f.soakPath + ": 4 line(s), 3 to rewrite", f.verifyPath + ": 2 line(s), 2 to rewrite",
		f.auditPath + ": 2 line(s), 1 to rewrite", "preview"} {
		if !strings.Contains(out, want) {
			t.Fatalf("preview output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, a131ScrubAccount) || strings.Contains(out, a131ScrubMask) {
		t.Fatalf("preview output carries the account:\n%s", out)
	}
}

func TestA131ScrubApplyReplacesAtomically(t *testing.T) {
	f := a131LegacyFiles(t)
	out, _, err := runCLI(t, "--config-dir", f.configDir, "scrub-account-records", "--apply")
	if err != nil {
		t.Fatalf("apply: %v\n%s", err, out)
	}
	for _, p := range []string{f.soakPath, f.verifyPath, f.auditPath} {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), a131ScrubAccount) || strings.Contains(string(raw), a131ScrubMask) {
			t.Fatalf("%s still carries the account:\n%s", p, raw)
		}
		info, _ := os.Stat(p)
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("%s mode %v, want 0600", p, info.Mode().Perm())
		}
	}
	// 읽는 쪽은 종전 값을 봄.
	cycles, err := soak.LoadCycles(f.soakPath)
	if err != nil || len(cycles) != 4 || cycles[0].AccountRef != a131ScrubAccount || cycles[3].AccountRef != "" {
		t.Fatalf("soak after scrub: %d, %v, %+v", len(cycles), err, cycles)
	}
	entries, err := verifylive.LoadEntries(f.verifyPath)
	if err != nil || len(entries) != 2 || entries[0].AccountRef != a131ScrubMask {
		t.Fatalf("verify after scrub: %+v, %v", entries, err)
	}
	log, _ := audit.Open(audit.Options{Path: f.auditPath})
	got, err := log.Entries()
	if err != nil || len(got) != 2 || got[0].Action != audit.ActionGateAccepted || got[0].New != "true" ||
		got[1].Setting != "engine.automation_gate.max_order_quantity" || !strings.Contains(got[0].Detail, "attestation_expires=2026-10-23") {
		t.Fatalf("audit meaning changed: %+v, %v", got, err)
	}
	// 두 번째 미리보기는 바꿀 것이 없음(멱등).
	again, _, err := runCLI(t, "--config-dir", f.configDir, "scrub-account-records")
	if err != nil || strings.Contains(again, "to rewrite") && !strings.Contains(again, "0 to rewrite") {
		t.Fatalf("second preview: %v\n%s", err, again)
	}
	for _, p := range []string{f.soakPath, f.verifyPath, f.auditPath} {
		if !strings.Contains(again, p+": ") || !strings.Contains(again, "0 to rewrite") {
			t.Fatalf("second preview should report 0 for %s:\n%s", p, again)
		}
	}
	// 임시 파일이 남지 않음.
	leftovers, _ := filepath.Glob(filepath.Join(f.configDir, ".*a131*"))
	if len(leftovers) != 0 {
		t.Fatalf("temporary files left behind: %v", leftovers)
	}
}

func TestA131ScrubApplyAbortsWhenTheFileMoved(t *testing.T) {
	f := a131LegacyFiles(t)
	previous := scrubBeforeReplace
	scrubBeforeReplace = func(path string) {
		if path == f.soakPath {
			fh, _ := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
			_, _ = fh.WriteString("{}\n")
			_ = fh.Close()
		}
	}
	t.Cleanup(func() { scrubBeforeReplace = previous })
	before, _ := os.ReadFile(f.soakPath)
	_, _, err := runCLI(t, "--config-dir", f.configDir, "scrub-account-records", "--apply")
	if err == nil || !strings.Contains(err.Error(), "changed while") {
		t.Fatalf("apply on a moving file = %v; want a refusal", err)
	}
	after, _ := os.ReadFile(f.soakPath)
	if string(after) != string(before)+"{}\n" {
		t.Fatalf("the moving file was replaced anyway")
	}
}

func TestA131ScrubApplyRefusesWhileAVerificationRuns(t *testing.T) {
	f := a131LegacyFiles(t)
	lock, err := runlock.Acquire(filepath.Join(filepath.Dir(f.verifyPath), runlock.FileName), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Release()
	before, _ := os.ReadFile(f.verifyPath)
	_, _, err = runCLI(t, "--config-dir", f.configDir, "scrub-account-records", "--apply")
	if err == nil || !strings.Contains(err.Error(), "verification") {
		t.Fatalf("apply during a verification = %v; want a refusal", err)
	}
	after, _ := os.ReadFile(f.verifyPath)
	if string(after) != string(before) {
		t.Fatalf("the record was rewritten during a verification")
	}
}

func TestA131SoakAttestOutputCarriesNoAccount(t *testing.T) {
	configDir := testenv.Isolate(t)
	seedQualifyingRecord(t, filepath.Join(configDir, soak.FileName))
	out, _, err := runCLI(t, "--config-dir", configDir, "soak", "attest", "--verified-by", "tester")
	if err != nil {
		t.Fatalf("soak attest: %v", err)
	}
	if strings.Contains(out, attest.Mask("123-45-678901")) || strings.Contains(out, "  account ") {
		t.Fatalf("attest output carries the account:\n%s", out)
	}
}

// TestA131CobraErrorLineIsRedacted 는 cobra 가 실패 시 루트 stderr 에 쓰는 `Error:` 줄이 main 의 배선(루트 SetErr 감쌈)으로
// 계좌 규칙을 지남을 잼. 양성 대조: 감싸지 않으면 같은 줄에 계좌가 나옴(cobra 가 정말 그 줄을 쓴다는 증거).
func TestA131CobraErrorLineIsRedacted(t *testing.T) {
	testenv.Isolate(t)
	const acct = "98765432166"
	redact.Register(acct)
	dir := filepath.Join(t.TempDir(), acct)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	run := func(wrap bool) string {
		root := newRootCmd()
		var errOut strings.Builder
		if wrap {
			root.SetErr(redact.Writer(&errOut))
		} else {
			root.SetErr(&errOut)
		}
		root.SetOut(io.Discard)
		root.SetArgs([]string{"scrub-account-records", "--file", dir})
		if err := root.Execute(); err == nil {
			t.Fatal("scrubbing a directory must fail")
		}
		return errOut.String()
	}
	if raw := run(false); !strings.Contains(raw, "Error:") || !strings.Contains(raw, acct) {
		t.Fatalf("positive control: cobra's error line should carry the path: %q", raw)
	}
	if got := run(true); !strings.Contains(got, "Error:") || strings.Contains(got, acct) {
		t.Fatalf("cobra's error line leaked the account through the wrapped stderr: %q", got)
	}
}

// TestA131ScrubCarriesTheReplacedTail 는 교체 직전에 경로를 연 쓰기(감사 줄 · 재열기 전 기록 줄)가 교체 뒤 옛 inode 에 쓴 줄이
// 새 파일로 같은 규칙을 거쳐 옮겨짐을 잼(리뷰 R-4c).
func TestA131ScrubCarriesTheReplacedTail(t *testing.T) {
	f := a131LegacyFiles(t)
	early, err := os.OpenFile(f.auditPath, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer early.Close()
	late, _ := json.Marshal(audit.Entry{At: time.Now().UTC(), Subject: "op", Action: audit.ActionRemoteLogin,
		Setting: "console.remote", New: "ok", Detail: "account=" + a131ScrubAccount})
	previous := scrubAfterReplace
	scrubAfterReplace = func(path string) {
		if path == f.auditPath {
			_, _ = early.Write(append(late, '\n'))
		}
	}
	t.Cleanup(func() { scrubAfterReplace = previous })
	if _, _, err := runCLI(t, "--config-dir", f.configDir, "scrub-account-records", "--apply"); err != nil {
		t.Fatalf("apply: %v", err)
	}
	log, _ := audit.Open(audit.Options{Path: f.auditPath})
	got, err := log.Entries()
	if err != nil || len(got) != 3 || got[2].Action != audit.ActionRemoteLogin {
		t.Fatalf("the line written into the replaced inode was lost: %+v, %v", got, err)
	}
	raw, _ := os.ReadFile(f.auditPath)
	if strings.Contains(string(raw), a131ScrubAccount) {
		t.Fatalf("the carried tail was not scrubbed: %s", raw)
	}
}
