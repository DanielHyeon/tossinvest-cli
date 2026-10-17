package main

// main_test.go 는 도구의 입력 경계 · 스키마 가드 · 분봉 읽기를 잼 — 임시 원장 · 가짜 분봉만 씀(운영 원장 · 브로커 호출 없음).

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func tempJournal(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "journal.db")
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     path,
		Clock:    clock.NewFake(time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

// noNetwork 는 분봉 읽기가 불리면 시험을 실패시킴 — 묶음이 없으면 브로커를 부르지 않아야 함.
func noNetwork(t *testing.T) func(string) (candleReader, string, error) {
	return func(string) (candleReader, string, error) {
		t.Error("the tool opened the broker client although there was no lot to price")
		return nil, "", errors.New("no network in tests")
	}
}

func testDeps(t *testing.T, out *bytes.Buffer) deps {
	return deps{now: func() time.Time { return time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC) }, stdout: out, candles: noNetwork(t)}
}

func TestTheToolRefusesWithoutAJournalPathOrAConfirmation(t *testing.T) {
	for _, args := range [][]string{
		{},
		{"--journal", tempJournal(t), "--execute"},
		{"--journal", tempJournal(t), "--confirm", "abc"},
		{"--journal", filepath.Join(t.TempDir(), "missing.db")},
	} {
		var out bytes.Buffer
		if code, err := run(context.Background(), args, testDeps(t, &out)); code != 1 || err == nil {
			t.Errorf("args %v: code=%d err=%v, want a refusal", args, code, err)
		}
	}
}

func TestThePreviewPrintsATokenAndAWrongTokenWritesNothing(t *testing.T) {
	path := tempJournal(t)
	var out bytes.Buffer
	if code, err := run(context.Background(), []string{"--journal", path}, testDeps(t, &out)); code != 0 || err != nil {
		t.Fatalf("preview: code=%d err=%v", code, err)
	}
	if !strings.Contains(out.String(), "confirmation token: ") || !strings.Contains(out.String(), "nothing was written") {
		t.Errorf("preview output = %q", out.String())
	}
	out.Reset()
	code, err := run(context.Background(), []string{"--journal", path, "--execute", "--confirm", "ffffffffffffffff"}, testDeps(t, &out))
	if code != 1 || !errors.Is(err, journal.ErrBackfillTokenMismatch) || !strings.Contains(out.String(), "written: 0") {
		t.Errorf("wrong token: code=%d err=%v out=%q, want a refusal with nothing written", code, err, out.String())
	}
}

// TestAnOlderSchemaJournalIsLeftByteIdentical(리뷰 R3-4): 스키마가 빌드보다 낮은 원장에 실행하면 쓰기 연결을 열기 전에 거절하고 파일은
// 바이트 단위로 그대로임(journal.Open 은 이관하므로 열리면 안 됨).
func TestAnOlderSchemaJournalIsLeftByteIdentical(t *testing.T) {
	path := tempJournal(t)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("PRAGMA user_version = 32"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	before := digestOf(t, path)
	var out bytes.Buffer
	code, err := run(context.Background(), []string{"--journal", path, "--execute", "--confirm", "ffffffffffffffff"}, testDeps(t, &out))
	if code != 1 || err == nil || !strings.Contains(err.Error(), "never migrates") {
		t.Fatalf("code=%d err=%v, want the schema refusal", code, err)
	}
	if after := digestOf(t, path); after != before {
		t.Error("a refused execution changed the journal file")
	}
}

func digestOf(t *testing.T, path string) [32]byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return sha256.Sum256(b)
}

func TestTheSchemaGuardRefusesAnyOtherVersion(t *testing.T) {
	if err := schemaGuard(32, 35); err == nil || !strings.Contains(err.Error(), "never migrates") {
		t.Errorf("schemaGuard(32, 35) = %v, want a refusal", err)
	}
	if err := schemaGuard(36, 35); err == nil {
		t.Error("a newer journal must be refused too")
	}
	if err := schemaGuard(35, 35); err != nil {
		t.Errorf("schemaGuard(35, 35) = %v, want nil", err)
	}
}

// TestFetchPricesKeepsOnlyTheExactMinuteAndPaces: 그 분의 봉이 있는 묶음만 가격을 얻고(돌려받은 이름을 그대로 적음), 읽기 사이에 간격을 둠,
// 쓰는 토큰 캐시 경로를 출력함.
func TestFetchPricesKeepsOnlyTheExactMinuteAndPaces(t *testing.T) {
	lots := []journal.IntakeLot{
		{Market: "kr", Symbol: "010170", CreatedAt: time.Date(2026, 10, 7, 0, 18, 54, 0, time.UTC), Quantity: "3"},
		{Market: "kr", Symbol: "333430", CreatedAt: time.Date(2026, 7, 31, 6, 40, 41, 0, time.UTC), Quantity: "25"},
	}
	var calls []time.Time
	reader := func(_ context.Context, market, symbol string, label time.Time) (string, time.Time, bool, error) {
		calls = append(calls, time.Now())
		if market == "kr" && symbol == "010170" && label.Equal(time.Date(2026, 10, 7, 0, 19, 0, 0, time.UTC)) {
			return "18430", label, true, nil
		}
		return "", label.Add(-10 * time.Minute), false, nil
	}
	var out bytes.Buffer
	d := deps{stdout: &out, pace: 20 * time.Millisecond,
		candles: func(string) (candleReader, string, error) { return reader, "/engine/config/openapi-token.json", nil }}
	prices, err := fetchPrices(context.Background(), d, "/engine/config", lots)
	if err != nil {
		t.Fatal(err)
	}
	if len(prices) != 1 || prices[lots[0].Key()].Close != "18430" ||
		!prices[lots[0].Key()].Label.Equal(time.Date(2026, 10, 7, 0, 19, 0, 0, time.UTC)) ||
		prices[lots[0].Key()].Source != journal.IntakeSourceBrokerCandle1m {
		t.Errorf("prices = %+v, want only the 010170 09:19 KST bar", prices)
	}
	if len(calls) != 2 || calls[1].Sub(calls[0]) < 20*time.Millisecond {
		t.Errorf("calls = %v, want two reads paced apart", calls)
	}
	if !strings.Contains(out.String(), "token cache: /engine/config/openapi-token.json") {
		t.Errorf("output %q does not name the token cache it used", out.String())
	}
}

// TestBarsAreNotReadWithoutTheEngineConfigDir(리뷰 4차 P1-1): 묶음이 있는데 --config-dir 가 없으면 브로커 client 를 열기 전에 거절함 —
// 기본값은 다른 토큰 캐시라 엔진 토큰을 무효화하는 발급을 일으킴(a082).
func TestBarsAreNotReadWithoutTheEngineConfigDir(t *testing.T) {
	lots := []journal.IntakeLot{{Market: "kr", Symbol: "010170", CreatedAt: time.Date(2026, 10, 7, 0, 18, 54, 0, time.UTC), Quantity: "3"}}
	var out bytes.Buffer
	d := deps{stdout: &out, candles: noNetwork(t)}
	if _, err := fetchPrices(context.Background(), d, "", lots); err == nil || !strings.Contains(err.Error(), "--config-dir") {
		t.Errorf("err = %v, want the --config-dir refusal", err)
	}
}
