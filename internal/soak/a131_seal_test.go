package soak_test

// a131_seal_test.go — soak 기록은 계좌를 봉인해 쓰고, 읽는 쪽은 열어 attestation 의 계좌 결속이 종전과 같음(design D4).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
)

const a131SoakAccount = "98765432188"

func writeSealedSoak(t *testing.T, dir string, cycles []soak.Cycle) string {
	t.Helper()
	path := filepath.Join(dir, soak.FileName)
	rec, err := soak.OpenRecorder(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cycles {
		c.AccountRef = a131SoakAccount
		if err := rec.Append(c); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
	if err := rec.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestA131SoakRecordSealsTheAccount(t *testing.T) {
	dir := t.TempDir()
	c := sampleCycle(soakStart)
	c.Endpoints[0].Error = "broker said account " + a131SoakAccount + " is busy"
	redact.Register(a131SoakAccount)
	path := writeSealedSoak(t, dir, []soak.Cycle{c})
	raw, _ := os.ReadFile(path)
	if strings.Contains(string(raw), a131SoakAccount) || strings.Contains(string(raw), `"account_ref"`) {
		t.Fatalf("soak record carries the account in the clear: %s", raw)
	}
	if !strings.Contains(string(raw), `"`+redact.SealedField+`"`) {
		t.Fatalf("soak record has no sealed account: %s", raw)
	}
}

func TestA131SealedSoakRecordStillAttestsTheSameAccount(t *testing.T) {
	dir := t.TempDir()
	path := writeSealedSoak(t, dir, threeCleanDays())
	cycles, err := soak.LoadCycles(path)
	if err != nil {
		t.Fatal(err)
	}
	for i, c := range cycles {
		if c.AccountRef != a131SoakAccount {
			t.Fatalf("cycle %d AccountRef = %q, want the opened account", i, c.AccountRef)
		}
	}
	s := soak.Summarize(cycles)
	now := soakStart.AddDate(0, 0, 3)
	a, err := soak.BuildAttestation(s, criteria(), now, "tester", "", nil)
	if err != nil {
		t.Fatalf("BuildAttestation: %v", err)
	}
	if a.AccountRef != a131SoakAccount {
		t.Fatalf("attestation account = %q, want %q", a.AccountRef, a131SoakAccount)
	}
	if err := a.Verify(now, a131SoakAccount, soak.RequiredEndpoints()); err != nil {
		t.Fatalf("engine interlock check fails: %v", err)
	}
}

func TestA131LegacyAndSealedLinesReadTheSame(t *testing.T) {
	dir := t.TempDir()
	days := threeCleanDays()
	path := writeSealedSoak(t, dir, days[3:])
	// 앞 절반은 a131 이전 판본이 쓴 평문 줄(구조체 직렬화 그대로).
	var plain []byte
	for _, c := range days[:3] {
		c.AccountRef = a131SoakAccount
		line, err := json.Marshal(c)
		if err != nil {
			t.Fatal(err)
		}
		plain = append(append(plain, line...), '\n')
	}
	sealed, _ := os.ReadFile(path)
	if err := os.WriteFile(path, append(plain, sealed...), 0o600); err != nil {
		t.Fatal(err)
	}
	cycles, err := soak.LoadCycles(path)
	if err != nil || len(cycles) != len(days) {
		t.Fatalf("LoadCycles = %d, %v", len(cycles), err)
	}
	s := soak.Summarize(cycles)
	if s.AccountRef != a131SoakAccount || len(s.AccountRefs) != 1 {
		t.Fatalf("mixed record reads as %q / %v", s.AccountRef, s.AccountRefs)
	}
}

func TestA131MissingSealKeyFailsClosed(t *testing.T) {
	dir := t.TempDir()
	path := writeSealedSoak(t, dir, threeCleanDays())
	if err := os.Remove(filepath.Join(dir, redact.KeyFileName)); err != nil {
		t.Fatal(err)
	}
	// 열 수 없는 봉인 줄은 해독 오류 — 빈 계좌로 읽어 연속일수에 세지 않음(리뷰 R-2).
	if cycles, err := soak.LoadCycles(path); err == nil {
		t.Fatalf("LoadCycles without the key = %d cycle(s), nil; want an error", len(cycles))
	}
	// 키를 새로 만들어 이어 쓰는 길도 닫힘.
	if _, err := soak.OpenRecorder(path); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("OpenRecorder on a sealed record without its key = %v; want a refusal", err)
	}
}

func TestA131UnopenableSealedLineIsAnError(t *testing.T) {
	dir := t.TempDir()
	days := threeCleanDays()
	path := writeSealedSoak(t, dir, days)
	// 다른 키로 봉인한 줄을 **마지막** 줄로 덧붙임 — 찢김 관용을 받으면 안 됨.
	otherDir := t.TempDir()
	other := writeSealedSoak(t, otherDir, days[:1])
	foreign, _ := os.ReadFile(other)
	f, _ := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
	_, _ = f.Write(foreign)
	_ = f.Close()
	if _, err := soak.LoadCycles(path); err == nil {
		t.Fatal("a line sealed with another key was read (as the account-less torn tail?)")
	}
	// 평문 줄 + 다른 키 줄의 섞인 기록으로 attestation 이 나가지 않음.
	plain, _ := json.Marshal(days[0])
	mixed := filepath.Join(t.TempDir(), soak.FileName)
	_ = os.WriteFile(mixed, append(append(plain, '\n'), foreign...), 0o600)
	if _, err := soak.LoadCycles(mixed); err == nil {
		t.Fatal("a mixed record with an unopenable line was read")
	}
}

func TestA131SoakAppendReopensAReplacedRecord(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, soak.FileName)
	rec, err := soak.OpenRecorder(path)
	if err != nil {
		t.Fatal(err)
	}
	defer rec.Close()
	if err := rec.Append(sampleCycle(soakStart)); err != nil {
		t.Fatal(err)
	}
	// 정리 명령이 하는 원자 교체.
	data, _ := os.ReadFile(path)
	tmp := path + ".tmp"
	_ = os.WriteFile(tmp, data, 0o600)
	if err := os.Rename(tmp, path); err != nil {
		t.Fatal(err)
	}
	if err := rec.Append(sampleCycle(soakStart.Add(time.Hour))); err != nil {
		t.Fatal(err)
	}
	cycles, err := soak.LoadCycles(path)
	if err != nil || len(cycles) != 2 {
		t.Fatalf("after the replace the record holds %d cycle(s), %v; want 2 (the second append went to the old inode)", len(cycles), err)
	}
}

func TestA131SoakRecordFormatIsTheSealedFormat(t *testing.T) {
	if soak.RecordFormatVersion != redact.SealedFormatVersion {
		t.Fatalf("soak.RecordFormatVersion = %d, redact.SealedFormatVersion = %d", soak.RecordFormatVersion, redact.SealedFormatVersion)
	}
}
