package verifylive

// a131_seal_test.go — 검증 기록(단계 줄 · 대사 줄)은 계좌(마스크)를 봉인해 쓰고, 읽는 쪽(LoadEntries · 대사 엄격 해독)은 열어
// 감독 증거 결속(SameAccountMasked)과 대사 계좌 결속(bindReconcileAccount)이 종전 값을 봄(design D4).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
)

const a131VerifyFull = "98765432177"

var a131VerifyMask = attest.Mask(a131VerifyFull)

func a131Entry(at time.Time) Entry {
	return Entry{RunID: "run-1", StepID: "s1", Title: "t", StartedAt: at, FinishedAt: at,
		AccountRef: a131VerifyMask, Mutating: true, Verdict: VerdictPass,
		Reason:    "broker echoed account " + a131VerifyFull,
		Calls:     []Call{{Endpoint: "POST /api/v1/orders", At: at}},
		Artifacts: []Artifact{{Kind: KindConditional, ID: "c-1", Symbol: "005930", CreatedAt: at}}}
}

func TestA131VerifyRecordSealsTheAccount(t *testing.T) {
	redact.Register(a131VerifyFull)
	path := filepath.Join(t.TempDir(), RecordFileName(MarketKR))
	rec, err := OpenRecorder(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := rec.Append(a131Entry(time.Now().UTC())); err != nil {
		t.Fatalf("Append: %v", err)
	}
	_ = rec.Close()
	raw, _ := os.ReadFile(path)
	for _, leak := range []string{a131VerifyFull, a131VerifyMask, `"account_ref"`} {
		if strings.Contains(string(raw), leak) {
			t.Fatalf("verify record carries %q: %s", leak, raw)
		}
	}
	entries, err := LoadEntries(path)
	if err != nil || len(entries) != 1 || entries[0].AccountRef != a131VerifyMask {
		t.Fatalf("LoadEntries = %+v, %v; want the opened mask", entries, err)
	}
}

func TestA131SealedVerifyProofStillBinds(t *testing.T) {
	path := filepath.Join(t.TempDir(), RecordFileName(MarketKR))
	rec, err := OpenRecorder(path)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := rec.Append(a131Entry(now.Add(-time.Hour))); err != nil {
		t.Fatal(err)
	}
	_ = rec.Close()
	entries, err := LoadEntries(path)
	if err != nil {
		t.Fatal(err)
	}
	ev := SucceededEndpoints(entries, now, 24*time.Hour)
	if len(ev.AccountRefs) != 1 || !attest.SameAccountMasked(a131VerifyFull, ev.AccountRefs[0]) {
		t.Fatalf("supervised evidence no longer binds: %v", ev.AccountRefs)
	}
	if _, ok := ev.Endpoints["POST /api/v1/orders"]; !ok {
		t.Fatalf("endpoint evidence lost: %v", ev.Endpoints)
	}
}

func TestA131ReconcileLineIsSealedAndStillBinds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, RecordFileName(MarketKR))
	rec, err := OpenRecorder(path)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := rec.Append(a131Entry(now)); err != nil {
		t.Fatal(err)
	}
	_ = rec.Close()

	raw, entries, err := readRecordStrictNoTail(path)
	if err != nil || len(entries) != 1 {
		t.Fatalf("readRecordStrictNoTail = %d, %v", len(entries), err)
	}
	target := entries[0].Artifacts[0]
	if _, _, err := bindReconcileAccount(entries, target, a131VerifyFull); err != nil {
		t.Fatalf("reconcile account binding broke on a sealed record: %v", err)
	}

	line, err := encodeReconcileLine(a131VerifyMask, target, "basis", now, now)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := lockRecordForAppend(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.appendLine(line, int64(len(raw))); err != nil {
		t.Fatalf("appendLine: %v", err)
	}
	lock.release()
	after, _ := os.ReadFile(path)
	if strings.Contains(string(after), a131VerifyMask) || strings.Contains(string(after), `"account_ref"`) {
		t.Fatalf("reconcile line carries the account: %s", after)
	}
	_, entries, err = readRecordStrictNoTail(path)
	if err != nil || len(entries) != 2 {
		t.Fatalf("re-read = %d, %v", len(entries), err)
	}
	for i, e := range entries {
		if e.AccountRef != a131VerifyMask {
			t.Fatalf("entry %d AccountRef = %q, want the opened mask", i, e.AccountRef)
		}
	}
}

func TestA131UnopenableVerifyLineIsAnError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, RecordFileName(MarketKR))
	rec, err := OpenRecorder(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := rec.Append(a131Entry(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	_ = rec.Close()
	if err := os.Remove(filepath.Join(dir, redact.KeyFileName)); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadEntries(path); err == nil {
		t.Fatal("LoadEntries read a sealed line without its key")
	}
	if _, _, err := readRecordStrictNoTail(path); err == nil {
		t.Fatal("the strict reconcile reader read a sealed line without its key")
	}
	if _, err := OpenRecorder(path); err == nil {
		t.Fatal("OpenRecorder started a new key for a record with sealed lines")
	}
}

func TestA131VerifyAppendReopensAReplacedRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), RecordFileName(MarketKR))
	rec, err := OpenRecorder(path)
	if err != nil {
		t.Fatal(err)
	}
	defer rec.Close()
	now := time.Now().UTC()
	if err := rec.Append(a131Entry(now)); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	_ = os.WriteFile(path+".tmp", data, 0o600)
	if err := os.Rename(path+".tmp", path); err != nil {
		t.Fatal(err)
	}
	if err := rec.Append(a131Entry(now)); err != nil {
		t.Fatal(err)
	}
	entries, err := LoadEntries(path)
	if err != nil || len(entries) != 2 {
		t.Fatalf("after the replace: %d entries, %v; want 2", len(entries), err)
	}
}

func TestA131VerifyRecordFormatIsTheSealedFormat(t *testing.T) {
	if RecordFormatVersion != redact.SealedFormatVersion {
		t.Fatalf("verifylive.RecordFormatVersion = %d, redact.SealedFormatVersion = %d", RecordFormatVersion, redact.SealedFormatVersion)
	}
}

func TestA131OpenRecorderRefusesAMissingKeyForASealedRecord(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, RecordFileName(MarketKR))
	rec, err := OpenRecorder(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = rec.Append(a131Entry(time.Now().UTC()))
	_ = rec.Close()
	_ = os.Remove(filepath.Join(dir, redact.KeyFileName))
	if _, err := OpenRecorder(path); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("OpenRecorder = %v; want the missing-key refusal before any broker call", err)
	}
}
