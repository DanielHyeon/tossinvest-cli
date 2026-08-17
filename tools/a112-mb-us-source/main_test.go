package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

type recordingReader struct {
	calls []string
}

type fixedIdentity struct{ value identitySnapshot }

func (f fixedIdentity) snapshot(config) (identitySnapshot, error) { return f.value, nil }

type selectedGoIdentity struct{}

func (selectedGoIdentity) snapshot(cfg config) (identitySnapshot, error) {
	goBinary, err := verifiedGoToolchain(cfg.goBinary)
	if err != nil {
		return identitySnapshot{}, err
	}
	snapshot := testSnapshot()
	snapshot.GoBinary = goBinary.Path
	snapshot.GoToolchainSHA256 = goBinary.SHA256
	snapshot.goExecution = goBinary
	return snapshot, nil
}

type tamperedGoExecutionSnapshotIdentity struct{}

func (tamperedGoExecutionSnapshotIdentity) snapshot(cfg config) (identitySnapshot, error) {
	goBinary, err := verifiedGoToolchain(cfg.goBinary)
	if err != nil {
		return identitySnapshot{}, err
	}
	snapshot := testSnapshot()
	snapshot.GoBinary = goBinary.Path
	snapshot.GoToolchainSHA256 = goBinary.SHA256
	snapshot.goExecution = goBinary
	if err := os.Chmod(goBinary.execution.executionPath, 0o700); err != nil {
		_ = goBinary.close()
		return identitySnapshot{}, err
	}
	if err := os.WriteFile(goBinary.execution.executionPath, []byte("#!/bin/sh\nexit 9\n"), 0o700); err != nil {
		_ = goBinary.close()
		return identitySnapshot{}, err
	}
	return snapshot, nil
}

func (r *recordingReader) Candle(_ context.Context, before []byte) (sourceResult, error) {
	r.calls = append(r.calls, "candle:"+string(before))
	return sourceResult{body: []byte(`{"result":{"nextBefore":null}}`), cursorJSON: []byte("null"), rateHeaders: validRateHeaders()}, nil
}

func (r *recordingReader) Orderbook(context.Context) (sourceResult, error) {
	r.calls = append(r.calls, "orderbook")
	return sourceResult{body: []byte(`{"result":{}}`), rateHeaders: validRateHeaders()}, nil
}

func (r *recordingReader) Calendar(_ context.Context, date string) (sourceResult, error) {
	r.calls = append(r.calls, "calendar:"+date)
	return sourceResult{body: []byte(`{"result":{}}`), rateHeaders: validRateHeaders()}, nil
}

func validRateHeaders() map[string][]string {
	return map[string][]string{
		"X-Ratelimit-Limit": {"10"}, "X-Ratelimit-Remaining": {"9"}, "X-Ratelimit-Reset": {"1"},
	}
}

func TestRunCollectsTerminalCandleThenSingleOrderbookAndCalendar(t *testing.T) {
	root := t.TempDir()
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}
	cache := filepath.Join(root, "token-cache.json")
	if err := os.WriteFile(cache, []byte("not read by fake"), 0o600); err != nil {
		t.Fatal(err)
	}
	reader := &recordingReader{}
	err := run(context.Background(), config{tokenCache: cache, goBinary: "/tmp/test-go", receiptRoot: root, sessionDate: "2026-08-14", before: []byte("first")}, dependencies{
		reader: reader, now: func() time.Time { return time.Date(2026, 8, 15, 1, 0, 0, 0, time.UTC) }, identity: testIdentity(), builder: testBuilder(),
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	want := []string{"candle:first", "orderbook", "calendar:2026-08-14"}
	if len(reader.calls) != len(want) {
		t.Fatalf("calls=%v want=%v", reader.calls, want)
	}
	for i := range want {
		if reader.calls[i] != want[i] {
			t.Fatalf("calls=%v want=%v", reader.calls, want)
		}
	}
	assertSealedPrivateReceipt(t, root)
}

func TestRunRejectsInsecureReceiptBeforeReader(t *testing.T) {
	root := t.TempDir()
	if err := os.Chmod(root, 0o755); err != nil {
		t.Fatal(err)
	}
	reader := &recordingReader{}
	err := run(context.Background(), config{tokenCache: "/tmp/a112-token", goBinary: "/tmp/test-go", receiptRoot: root, sessionDate: "2026-08-14"}, dependencies{reader: reader, now: time.Now, identity: testIdentity(), builder: testBuilder()})
	if err == nil {
		t.Fatal("insecure receipt root accepted")
	}
	if len(reader.calls) != 0 {
		t.Fatalf("insecure receipt made network calls: %v", reader.calls)
	}
}

func TestRunStopsOnReaderErrorWithoutFallback(t *testing.T) {
	root := t.TempDir()
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}
	reader := &errorReader{}
	err := run(context.Background(), config{tokenCache: "/tmp/a112-token", goBinary: "/tmp/test-go", receiptRoot: root, sessionDate: "2026-08-14"}, dependencies{reader: reader, now: time.Now, identity: testIdentity(), builder: testBuilder()})
	if err == nil {
		t.Fatal("reader error accepted")
	}
	if reader.calls != 1 {
		t.Fatalf("calls=%d want 1", reader.calls)
	}
}

func TestRunUsesDecodedCursorBytesWithoutNormalization(t *testing.T) {
	root := privateRoot(t)
	reader := &pagedReader{pages: []sourceResult{
		{body: []byte(`{"result":{"nextBefore":"A%20\u00c5"}}`), cursorJSON: []byte(`"A%20\u00c5"`), cursorValue: []byte("A%20Å"), rateHeaders: validRateHeaders()},
		{body: []byte(`{"result":{"nextBefore":null}}`), cursorJSON: []byte("null"), rateHeaders: validRateHeaders()},
	}}
	if err := run(context.Background(), testConfig(root), testDependencies(reader)); err != nil {
		t.Fatalf("run: %v", err)
	}
	want := []byte("A%20Å")
	if len(reader.befores) != 2 || !bytes.Equal(reader.befores[1], want) {
		t.Fatalf("decoded cursor was changed: %q", reader.befores)
	}
}

func TestRunHoldsBeforeReaderWhenLessThanOneRequestBudgetRemains(t *testing.T) {
	root := privateRoot(t)
	reader := &recordingReader{}
	started := time.Date(2026, 8, 15, 1, 0, 0, 0, time.UTC)
	clocks := 0
	err := run(context.Background(), testConfig(root), dependencies{reader: reader, identity: testIdentity(), builder: testBuilder(), now: func() time.Time {
		clocks++
		if clocks == 1 {
			return started
		}
		return started.Add(overallBudget - requestBudget + time.Nanosecond)
	}})
	if err == nil || len(reader.calls) != 0 {
		t.Fatalf("deadline HOLD=%v calls=%v", err, reader.calls)
	}
}

func TestRunHoldsOnCursorLoopBeforeOrderbook(t *testing.T) {
	root := privateRoot(t)
	reader := &pagedReader{pages: []sourceResult{
		{body: []byte(`{"result":{"nextBefore":"same"}}`), cursorJSON: []byte(`"same"`), cursorValue: []byte("same"), rateHeaders: validRateHeaders()},
		{body: []byte(`{"result":{"nextBefore":"same"}}`), cursorJSON: []byte(`"same"`), cursorValue: []byte("same"), rateHeaders: validRateHeaders()},
	}}
	err := run(context.Background(), testConfig(root), testDependencies(reader))
	if err == nil || reader.orderbookCalls != 0 || reader.calendarCalls != 0 {
		t.Fatalf("cursor loop accepted or fell through: err=%v orderbook=%d calendar=%d", err, reader.orderbookCalls, reader.calendarCalls)
	}
}

// 캔들 4페이지가 모두 정상 cursor를 주면 그건 HOLD가 아니다. 5번째 캔들 요청 없이
// 멈추고 receipt에 "cap을 다 썼다"고 적은 다음 orderbook/calendar/seal까지 그대로 간다.
func TestRunRecordsCapExhaustionAndStillCollectsOrderbookCalendarAndSeals(t *testing.T) {
	root := privateRoot(t)
	reader := &orderedReader{pages: []sourceResult{
		candlePageWithCursor("c1"), candlePageWithCursor("c2"), candlePageWithCursor("c3"), candlePageWithCursor("c4"),
	}}
	if err := run(context.Background(), testConfig(root), testDependencies(reader)); err != nil {
		t.Fatalf("cap exhaustion must not HOLD: %v", err)
	}
	want := []string{"candle:first", "candle:c1", "candle:c2", "candle:c3", "orderbook", "calendar:2026-08-14"}
	if !reflect.DeepEqual(reader.calls, want) {
		t.Fatalf("calls=%v want=%v", reader.calls, want)
	}
	assertSealedPrivateReceipt(t, root)
	runDirectory := receiptRunDirectory(t, root)
	if _, err := os.Stat(filepath.Join(runDirectory, "TAINTED")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("successful cap exhaustion tainted the receipt: %v", err)
	}
	record := readCandleCrawlRecord(t, runDirectory)
	if record.Schema != "a112-mb-us-candle-crawl:v1" || record.Pages != 4 || record.Terminal != "cap_exhausted" {
		t.Fatalf("candle crawl record=%+v", record)
	}
	if record.LastCursorSHA256 != digestBytes([]byte("c4")) {
		t.Fatalf("last cursor digest=%q want=%q", record.LastCursorSHA256, digestBytes([]byte("c4")))
	}
	// 마지막 cursor 해시는 4번째 페이지 meta가 적어 둔 cursor 해시와 같은 값이어야 한다.
	meta := readCandleMetaRecord(t, runDirectory, "candle-04.meta.json")
	if meta.CursorValueSHA256 == "" || record.LastCursorSHA256 != meta.CursorValueSHA256 {
		t.Fatalf("crawl last_cursor_sha256=%q candle-04 cursor_value_sha256=%q", record.LastCursorSHA256, meta.CursorValueSHA256)
	}
	if !manifestListsFile(t, runDirectory, "candle-crawl.json") {
		t.Fatalf("manifest does not list candle-crawl.json: %v", manifestFileNames(t, runDirectory))
	}
}

// raw null로 끝난 기존 경로도 같은 기록 파일을 남긴다. 다만 terminal은 "null"이고
// 마지막 cursor 해시는 없다.
func TestRunRecordsNullTerminalInCandleCrawlRecord(t *testing.T) {
	root := privateRoot(t)
	reader := &recordingReader{}
	if err := run(context.Background(), testConfig(root), testDependencies(reader)); err != nil {
		t.Fatalf("run: %v", err)
	}
	assertSealedPrivateReceipt(t, root)
	runDirectory := receiptRunDirectory(t, root)
	record := readCandleCrawlRecord(t, runDirectory)
	if record.Schema != "a112-mb-us-candle-crawl:v1" || record.Pages != 1 || record.Terminal != "null" {
		t.Fatalf("candle crawl record=%+v", record)
	}
	if record.LastCursorSHA256 != "" {
		t.Fatalf("terminal null recorded a cursor digest: %q", record.LastCursorSHA256)
	}
	if !manifestListsFile(t, runDirectory, "candle-crawl.json") {
		t.Fatalf("manifest does not list candle-crawl.json: %v", manifestFileNames(t, runDirectory))
	}
	if len(reader.calls) != 3 {
		t.Fatalf("calls=%v want one candle, one orderbook and one calendar", reader.calls)
	}
}

// cap이 HOLD가 아니게 되어도 같은 cursor가 다시 오면 여전히 HOLD다.
func TestRunStillHoldsOnCursorLoopAfterCapChange(t *testing.T) {
	root := privateRoot(t)
	reader := &orderedReader{pages: []sourceResult{
		candlePageWithCursor("c1"), candlePageWithCursor("c2"), candlePageWithCursor("c2"),
	}}
	err := run(context.Background(), testConfig(root), testDependencies(reader))
	if err == nil {
		t.Fatal("cursor loop was accepted")
	}
	want := []string{"candle:first", "candle:c1", "candle:c2"}
	if !reflect.DeepEqual(reader.calls, want) {
		t.Fatalf("calls=%v want=%v", reader.calls, want)
	}
	runDirectory := receiptRunDirectory(t, root)
	if _, statErr := os.Stat(filepath.Join(runDirectory, "TAINTED")); statErr != nil {
		t.Fatalf("cursor loop HOLD did not taint the receipt: %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(runDirectory, "candle-crawl.json")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("cursor loop recorded a crawl outcome: %v", statErr)
	}
}

// crawl 기록 파일을 못 쓰면 그건 조용히 넘어갈 일이 아니다. orderbook 요청 없이 HOLD하고
// receipt에 TAINTED를 남기며 manifest는 생기지 않아야 한다.
func TestRunHoldsAndTaintsWhenCandleCrawlRecordWriteFails(t *testing.T) {
	root := privateRoot(t)
	reader := &orderedReader{pages: []sourceResult{
		candlePageWithCursor("c1"), candlePageWithCursor("c2"), candlePageWithCursor("c3"), candlePageWithCursor("c4"),
	}}
	deps := testDependencies(reader)
	deps.openReceipt = func(path string) (receiptStore, error) {
		store, err := openReceipt(path)
		if err != nil {
			return nil, err
		}
		return &crawlRecordFailingStore{receiptStore: store}, nil
	}
	err := run(context.Background(), testConfig(root), deps)
	if err == nil {
		t.Fatal("failed candle crawl receipt write was accepted")
	}
	if !strings.Contains(err.Error(), "candle crawl receipt") {
		t.Fatalf("error=%v want it to name the candle crawl receipt", err)
	}
	want := []string{"candle:first", "candle:c1", "candle:c2", "candle:c3"}
	if !reflect.DeepEqual(reader.calls, want) {
		t.Fatalf("calls=%v want=%v", reader.calls, want)
	}
	runDirectory := receiptRunDirectory(t, root)
	if _, statErr := os.Stat(filepath.Join(runDirectory, "TAINTED")); statErr != nil {
		t.Fatalf("failed crawl record write did not taint the receipt: %v", statErr)
	}
	if receiptHasManifest(root) {
		t.Fatal("failed crawl record write wrote a success manifest")
	}
}

// 4번째 페이지의 rate 예산이 바닥나면, 그 뒤에 캔들 요청이 없더라도 여전히 HOLD다.
func TestRunHoldsWhenPageFourRateBudgetIsExhausted(t *testing.T) {
	root := privateRoot(t)
	reader := &orderedReader{pages: []sourceResult{
		candlePageWithCursor("c1"), candlePageWithCursor("c2"), candlePageWithCursor("c3"),
		candlePageWithCursorAndRemaining("c4", "0"),
	}}
	err := run(context.Background(), testConfig(root), testDependencies(reader))
	if err == nil {
		t.Fatal("exhausted page 4 rate budget was accepted")
	}
	if !strings.Contains(err.Error(), "candle page 4 rate budget") {
		t.Fatalf("error=%v want the page 4 rate budget HOLD", err)
	}
	want := []string{"candle:first", "candle:c1", "candle:c2", "candle:c3"}
	if !reflect.DeepEqual(reader.calls, want) {
		t.Fatalf("calls=%v want=%v", reader.calls, want)
	}
	runDirectory := receiptRunDirectory(t, root)
	if _, statErr := os.Stat(filepath.Join(runDirectory, "TAINTED")); statErr != nil {
		t.Fatalf("page 4 rate budget HOLD did not taint the receipt: %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(runDirectory, "candle-crawl.json")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("page 4 rate budget HOLD recorded a crawl outcome: %v", statErr)
	}
}

// 앞쪽 페이지의 rate 예산이 바닥나면 다음 캔들 요청 자체를 하지 않는다.
func TestRunHoldsWhenEarlierPageRateBudgetIsExhausted(t *testing.T) {
	root := privateRoot(t)
	reader := &orderedReader{pages: []sourceResult{
		candlePageWithCursor("c1"), candlePageWithCursorAndRemaining("c2", "0"),
		candlePageWithCursor("c3"), candlePageWithCursor("c4"),
	}}
	err := run(context.Background(), testConfig(root), testDependencies(reader))
	if err == nil {
		t.Fatal("exhausted page 2 rate budget was accepted")
	}
	if !strings.Contains(err.Error(), "candle page 2 rate budget") {
		t.Fatalf("error=%v want the page 2 rate budget HOLD", err)
	}
	want := []string{"candle:first", "candle:c1"}
	if !reflect.DeepEqual(reader.calls, want) {
		t.Fatalf("calls=%v want=%v", reader.calls, want)
	}
	runDirectory := receiptRunDirectory(t, root)
	if _, statErr := os.Stat(filepath.Join(runDirectory, "candle-crawl.json")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("page 2 rate budget HOLD recorded a crawl outcome: %v", statErr)
	}
}

// 두 페이지를 받고 raw null로 끝나면 pages는 2다. 그리고 마지막 cursor 열쇠말 자체가
// 파일에 없어야 한다(omitempty).
func TestRunRecordsPagesBeforeNullTerminal(t *testing.T) {
	root := privateRoot(t)
	reader := &orderedReader{pages: []sourceResult{candlePageWithCursor("c1"), candlePageWithNullCursor()}}
	if err := run(context.Background(), testConfig(root), testDependencies(reader)); err != nil {
		t.Fatalf("run: %v", err)
	}
	want := []string{"candle:first", "candle:c1", "orderbook", "calendar:2026-08-14"}
	if !reflect.DeepEqual(reader.calls, want) {
		t.Fatalf("calls=%v want=%v", reader.calls, want)
	}
	runDirectory := receiptRunDirectory(t, root)
	record := readCandleCrawlRecord(t, runDirectory)
	if record.Pages != 2 || record.Terminal != "null" {
		t.Fatalf("candle crawl record=%+v", record)
	}
	data, err := os.ReadFile(filepath.Join(runDirectory, "candle-crawl.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "last_cursor_sha256") {
		t.Fatalf("terminal null kept the last cursor key: %s", data)
	}
}

// receipt에 적히는 해시는 SHA-256이다. 알고리즘이 조용히 바뀌면 이 값이 달라진다.
func TestDigestBytesPinsSHA256Vector(t *testing.T) {
	const want = "sha256:0012a3fa000c5dc26ee658c3c58e12cecd58d6455cec3d5621f0c787675b38aa"
	if got := digestBytes([]byte("c4")); got != want {
		t.Fatalf("digestBytes(\"c4\")=%q want=%q", got, want)
	}
}

func TestRunHoldsWhenFirstCursorRepeatsExplicitInitialBefore(t *testing.T) {
	root := privateRoot(t)
	reader := &pagedReader{pages: []sourceResult{{body: []byte(`{"result":{"nextBefore":"first"}}`), cursorJSON: []byte(`"first"`), cursorValue: []byte("first"), rateHeaders: validRateHeaders()}}}
	err := run(context.Background(), testConfig(root), testDependencies(reader))
	if err == nil || reader.orderbookCalls != 0 || reader.calendarCalls != 0 {
		t.Fatalf("initial cursor loop accepted or fell through: err=%v", err)
	}
}

func TestRunHoldsOnNonStringRawCursorBeforeOrderbook(t *testing.T) {
	root := privateRoot(t)
	reader := &pagedReader{pages: []sourceResult{{body: []byte(`{"result":{"nextBefore":7}}`), cursorJSON: []byte("7"), cursorValue: []byte("7"), rateHeaders: validRateHeaders()}}}
	err := run(context.Background(), testConfig(root), testDependencies(reader))
	if err == nil || reader.orderbookCalls != 0 || reader.calendarCalls != 0 {
		t.Fatalf("non-string cursor accepted or fell through: err=%v", err)
	}
}

func TestRunHoldsOnSecretLikeRawBodyBeforeNextRequest(t *testing.T) {
	root := privateRoot(t)
	reader := &pagedReader{pages: []sourceResult{{body: []byte(`{"result":{"accessToken":"must-not-persist","nextBefore":null}}`), cursorJSON: []byte("null"), rateHeaders: validRateHeaders()}}}
	err := run(context.Background(), testConfig(root), testDependencies(reader))
	if err == nil || reader.orderbookCalls != 0 || reader.calendarCalls != 0 {
		t.Fatalf("secret raw body accepted or fell through: err=%v", err)
	}
}

func TestRunHoldsWhenIdentityDriftsAfterNetwork(t *testing.T) {
	root := privateRoot(t)
	reader := &recordingReader{}
	identity := &changingIdentity{}
	err := run(context.Background(), testConfig(root), dependencies{reader: reader, now: fixedNow, identity: identity, builder: testBuilder()})
	if err == nil || len(reader.calls) != 3 {
		t.Fatalf("identity drift was not HOLD after bounded reads: err=%v calls=%v", err, reader.calls)
	}
}

func TestRunHoldsWhenCancelledDuringPostReadFinalization(t *testing.T) {
	root := privateRoot(t)
	ctx, cancel := context.WithCancel(context.Background())
	identity := &cancellingIdentity{cancel: cancel}
	err := run(ctx, testConfig(root), dependencies{reader: &recordingReader{}, now: fixedNow, identity: identity, builder: testBuilder()})
	if err == nil {
		t.Fatal("cancel during finalization returned success")
	}
	if receiptHasManifest(root) {
		t.Fatal("cancelled measurement wrote a success manifest")
	}
}

func TestRunHoldsWhenClockExceedsOverallBudgetAfterFinalResponse(t *testing.T) {
	root := privateRoot(t)
	started := fixedNow()
	clockCalls := 0
	err := run(context.Background(), testConfig(root), dependencies{reader: &recordingReader{}, identity: testIdentity(), builder: testBuilder(), now: func() time.Time {
		clockCalls++
		if clockCalls <= 4 { // start plus all three request admissions
			return started
		}
		return started.Add(overallBudget + time.Nanosecond)
	}})
	if err == nil {
		t.Fatal("overall deadline overrun returned success")
	}
	if receiptHasManifest(root) {
		t.Fatal("deadline-overrun measurement wrote a success manifest")
	}
}

func TestRunHoldsOnNestedNeutralKeyBearerValue(t *testing.T) {
	root := privateRoot(t)
	reader := &pagedReader{pages: []sourceResult{{body: []byte(`{"result":{"payload":{"note":"Bearer abc.def.ghi"},"nextBefore":null}}`), cursorJSON: []byte("null"), rateHeaders: validRateHeaders()}}}
	err := run(context.Background(), testConfig(root), testDependencies(reader))
	if err == nil {
		t.Fatal("neutral key bearer value was persisted")
	}
	if receiptHasRawPayload(root) {
		t.Fatal("secret-like raw payload was written")
	}
}

func TestRunHoldsWhenUnexpectedReceiptEntryAppearsBeforeSeal(t *testing.T) {
	root := privateRoot(t)
	identity := &poisoningIdentity{root: root}
	err := run(context.Background(), testConfig(root), dependencies{reader: &recordingReader{}, now: fixedNow, identity: identity, builder: testBuilder()})
	if err == nil {
		t.Fatal("unexpected receipt entry was accepted")
	}
	if receiptHasManifest(root) {
		t.Fatal("poisoned receipt wrote a success manifest")
	}
}

func TestSealedReceiptDetectsPostSealModeDowngrade(t *testing.T) {
	root := privateRoot(t)
	store, err := openReceipt(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.close() })
	if err := store.writeJSON("preflight.json", receiptPreflight{}); err != nil {
		t.Fatal(err)
	}
	if err := store.seal(receiptSeal{}); err != nil {
		t.Fatal(err)
	}
	run := receiptRunDirectory(t, root)
	if err := os.Chmod(filepath.Join(run, "manifest.json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := store.verifySealed(); err == nil {
		t.Fatal("post-seal 0644 mode downgrade was accepted")
	}
}

func TestValidateConfigRejectsCallerClaimedBuildCommand(t *testing.T) {
	if _, exists := reflect.TypeOf(config{}).FieldByName("buildCommand"); exists {
		t.Fatal("caller-claimed build command remains configurable")
	}
}

func TestValidateConfigRequiresExplicitAbsoluteGoBinary(t *testing.T) {
	root := privateRoot(t)
	for _, goBinary := range []string{"", "go", "relative/go"} {
		cfg := testConfig(root)
		cfg.goBinary = goBinary
		if err := validateConfig(cfg); err == nil {
			t.Fatalf("go binary %q was accepted", goBinary)
		}
	}
}

func TestRunRejectsUnsafeSelectedGoBeforeReader(t *testing.T) {
	root := privateRoot(t)
	safe := writeSafeGoBinary(t, root, "safe-go")
	worldWritable := writeSafeGoBinary(t, root, "world-writable-go")
	if err := os.Chmod(worldWritable, 0o777); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(root, "not-a-go")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(root, "go-link")
	if err := os.Symlink(safe, symlink); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	for name, goBinary := range map[string]string{
		"missing":        filepath.Join(root, "missing-go"),
		"relative":       "relative/go",
		"symlink":        symlink,
		"non_regular":    directory,
		"world_writable": worldWritable,
	} {
		t.Run(name, func(t *testing.T) {
			reader := &recordingReader{}
			cfg := testConfig(root)
			cfg.goBinary = goBinary
			err := run(context.Background(), cfg, dependencies{reader: reader, now: fixedNow, identity: selectedGoIdentity{}, builder: testBuilder()})
			if err == nil {
				t.Fatal("unsafe Go binary reached collection")
			}
			if len(reader.calls) != 0 {
				t.Fatalf("unsafe Go binary made reader calls: %v", reader.calls)
			}
		})
	}
}

func TestRunRejectsTamperedGoExecutionSnapshotBeforeReader(t *testing.T) {
	root := privateRoot(t)
	reader := &recordingReader{}
	goBinary := writeSafeGoBinary(t, root, "drifting-go")
	cfg := testConfig(root)
	cfg.goBinary = goBinary
	err := run(context.Background(), cfg, dependencies{reader: reader, now: fixedNow, identity: tamperedGoExecutionSnapshotIdentity{}, builder: testBuilder()})
	if err == nil {
		t.Fatal("tampered Go execution snapshot reached collection")
	}
	if len(reader.calls) != 0 {
		t.Fatalf("tampered Go execution snapshot made reader calls: %v", reader.calls)
	}
}

// This is the concrete swap-after-hash/restore-before-post-hash attack.  The
// selected pathname is first measured as harmless; the replacement restores
// the harmless bytes before the post-command check.  Pathname re-hashing alone
// therefore cannot prove what actually executed.
func TestGoExecutionResistsSwapAfterHashAndRestoreBeforePostHash(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "goroot", "bin", "go")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Dir(filepath.Dir(path)), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	original := "#!/bin/sh\n[ x$1 = xenv ] && printf go1.selected\n"
	if err := os.WriteFile(path, []byte(original), 0o700); err != nil {
		t.Fatal(err)
	}
	selected, err := verifiedGoToolchain(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = selected.close() }()
	if err := selected.revalidate(); err != nil {
		t.Fatalf("pre-command revalidation: %v", err)
	}
	attacker := "#!/bin/sh\nprintf %s '" + original + "' > \"$0\"\nprintf go1.attacker\n"
	if err := os.WriteFile(path, []byte(attacker), 0o700); err != nil {
		t.Fatal(err)
	}
	executable, err := selected.commandPath()
	if err != nil {
		t.Fatalf("execution snapshot: %v", err)
	}
	gotBytes, err := exec.Command(executable, "env", "GOVERSION").Output()
	if err != nil {
		t.Fatalf("selected Go invocation: %v", err)
	}
	if err := selected.revalidate(); err != nil {
		t.Fatalf("post-command revalidation: %v", err)
	}
	got := strings.TrimSpace(string(gotBytes))
	if got != "go1.selected" {
		t.Fatalf("swap/restore executed %q; expected verified Go bytes", got)
	}
}

// Git has the same provenance hazard as Go.  This test names the private
// execution-snapshot capability that Git must use; prior pathname-hash code
// has no such capability at all.
func TestGitExecutionSnapshotResistsSwapAfterHashAndRestoreBeforePostHash(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "selected-git")
	original := "#!/bin/sh\nprintf git.selected\n"
	if err := os.WriteFile(path, []byte(original), 0o700); err != nil {
		t.Fatal(err)
	}
	snapshot, err := snapshotExecutionBinary(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = snapshot.close() }()
	attacker := "#!/bin/sh\nprintf %s '" + original + "' > \"$0\"\nprintf git.attacker\n"
	if err := os.WriteFile(path, []byte(attacker), 0o700); err != nil {
		t.Fatal(err)
	}
	executable, err := snapshot.commandPath()
	if err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(executable).Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(output)); got != "git.selected" {
		t.Fatalf("swap/restore executed %q; expected verified Git bytes", got)
	}
}

func TestExecutionSnapshotRejectsPrivateContentModeAndEntryTampering(t *testing.T) {
	path := writeSafeGoBinary(t, t.TempDir(), "tool")
	for name, mutate := range map[string]func(*executionSnapshot) error{
		"content_and_mode": func(snapshot *executionSnapshot) error {
			if err := os.Chmod(snapshot.executionPath, 0o700); err != nil {
				return err
			}
			return os.WriteFile(snapshot.executionPath, []byte("#!/bin/sh\nexit 9\n"), 0o700)
		},
		"unexpected_entry": func(snapshot *executionSnapshot) error {
			entry := filepath.Join(snapshot.directory, "unexpected")
			return os.WriteFile(entry, []byte("unexpected"), 0o600)
		},
		"unexpected_symlink": func(snapshot *executionSnapshot) error {
			outside := filepath.Join(t.TempDir(), "outside")
			if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
				return err
			}
			return os.Symlink(outside, filepath.Join(snapshot.directory, "unexpected-link"))
		},
	} {
		t.Run(name, func(t *testing.T) {
			snapshot, err := snapshotExecutionBinary(path)
			if err != nil {
				t.Fatal(err)
			}
			privateDirectory := snapshot.directory
			if err := mutate(snapshot); err != nil {
				t.Fatal(err)
			}
			if _, err := snapshot.commandPath(); err == nil {
				t.Fatal("tampered private execution snapshot was accepted")
			}
			if err := snapshot.close(); err != nil {
				t.Fatalf("tampered private execution cleanup: %v", err)
			}
			if _, err := os.Stat(privateDirectory); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("tampered private execution snapshot remained at %q: %v", privateDirectory, err)
			}
		})
	}
}

func TestExecutionSnapshotCloseRemovesPrivateCapability(t *testing.T) {
	snapshot, err := snapshotExecutionBinary(writeSafeGoBinary(t, t.TempDir(), "tool"))
	if err != nil {
		t.Fatal(err)
	}
	directory := snapshot.directory
	if err := snapshot.close(); err != nil {
		t.Fatalf("close private execution snapshot: %v", err)
	}
	if _, err := os.Stat(directory); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("closed private capability remained at %q: %v", directory, err)
	}
}

func TestRealGoExecutionSnapshotRunsUnderSanitizedEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("real selected Go execution snapshot")
	}
	selected, err := verifiedGoToolchain(realGoBinary(t))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = selected.close() }()
	executable, err := selected.commandPath()
	if err != nil {
		t.Fatal(err)
	}
	environment, err := selected.environment("", "", "/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "env", "GOVERSION")
	command.Env = environment
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("private selected Go execution: %v\n%s", err, output)
	}
	if strings.TrimSpace(string(output)) == "" {
		t.Fatal("private selected Go returned an empty GOVERSION")
	}
}

func TestPrivateGoDistributionResistsOriginalToolAndSourceSwapRestore(t *testing.T) {
	for name, fixture := range map[string]func(*testing.T) (string, string){
		"source": func(t *testing.T) (string, string) {
			root := t.TempDir()
			goPath := writeMiniGoDistribution(t, root, "IFS= read -r line < \"$GOROOT/src/marker\"\nprintf source.selected > \"$GOROOT/src/marker\"\nprintf %s \"$line\"\n")
			marker := filepath.Join(root, "src", "marker")
			if err := os.WriteFile(marker, []byte("source.selected"), 0o600); err != nil {
				t.Fatal(err)
			}
			return goPath, marker
		},
		"tool": func(t *testing.T) (string, string) {
			root := t.TempDir()
			tool := filepath.Join(root, "pkg", "tool", "linux_amd64", "compile")
			goPath := writeMiniGoDistribution(t, root, "exec \"$GOROOT/pkg/tool/linux_amd64/compile\"\n")
			if err := os.WriteFile(tool, []byte("#!/bin/sh\nprintf tool.selected\n"), 0o700); err != nil {
				t.Fatal(err)
			}
			return goPath, tool
		},
	} {
		t.Run(name, func(t *testing.T) {
			goPath, target := fixture(t)
			selected, err := verifiedGoToolchain(goPath)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = selected.close() }()
			original, err := os.ReadFile(target)
			if err != nil {
				t.Fatal(err)
			}
			attacker := "#!/bin/sh\nprintf %s '" + string(original) + "' > \"$0\"\nprintf " + name + ".attacker\n"
			if name == "source" {
				attacker = name + ".attacker"
			}
			if err := os.WriteFile(target, []byte(attacker), 0o700); err != nil {
				t.Fatal(err)
			}
			got, err := selectedGoEnv(context.Background(), selected, "GOVERSION")
			if err != nil {
				t.Fatal(err)
			}
			if got != name+".selected" {
				t.Errorf("unbound original %s executed %q; expected private distribution bytes", name, got)
			}
		})
	}
}

func TestPrivateGoDistributionRejectsPrivateDriftAndCleansUp(t *testing.T) {
	root := t.TempDir()
	goPath := writeMiniGoDistribution(t, root, "printf go1.private\n")
	marker := filepath.Join(root, "src", "marker")
	if err := os.WriteFile(marker, []byte("source.private"), 0o600); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(goBinaryIdentity) error{
		"content": func(selected goBinaryIdentity) error {
			path := filepath.Join(selected.distribution.privateRoot, "src", "marker")
			if err := os.Chmod(path, 0o600); err != nil {
				return err
			}
			return os.WriteFile(path, []byte("tampered"), 0o600)
		},
		"unexpected_entry": func(selected goBinaryIdentity) error {
			return os.WriteFile(filepath.Join(selected.distribution.privateRoot, "unexpected"), []byte("unexpected"), 0o400)
		},
	} {
		t.Run(name, func(t *testing.T) {
			selected, err := verifiedGoToolchain(goPath)
			if err != nil {
				t.Fatal(err)
			}
			privateRoot := selected.distribution.privateRoot
			if err := mutate(selected); err != nil {
				t.Fatal(err)
			}
			if err := selected.revalidate(); err == nil {
				t.Fatal("private Go distribution drift was accepted")
			}
			if err := selected.close(); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(privateRoot); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("private Go distribution remained after HOLD cleanup: %q: %v", privateRoot, err)
			}
		})
	}
}

func TestPrivateGoDistributionSnapshotHonorsCancellation(t *testing.T) {
	root := t.TempDir()
	goPath := writeMiniGoDistribution(t, root, "printf go1.cancel\n")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := snapshotGoDistributionContext(ctx, goPath, digestBytes([]byte("not-read"))); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled distribution snapshot err=%v want context.Canceled", err)
	}
}

func TestPrivateGoDistributionSnapshotCancelsCopyAndCleansPartialTree(t *testing.T) {
	if testing.Short() {
		t.Skip("large cancellation-copy fixture")
	}
	root := t.TempDir()
	goPath := writeMiniGoDistribution(t, root, "printf go1.cancel-copy\n")
	large := filepath.Join(root, "src", "cancel-copy-fixture")
	file, err := os.OpenFile(large, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(64 << 20); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	before, err := filepath.Glob(filepath.Join(os.TempDir(), "a112-mb-us-goroot-*"))
	if err != nil {
		t.Fatal(err)
	}
	known := make(map[string]struct{}, len(before))
	for _, path := range before {
		known[path] = struct{}{}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	privateRootCreated := make(chan struct{})
	stopWatching := make(chan struct{})
	go func() {
		ticker := time.NewTicker(100 * time.Microsecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopWatching:
				return
			case <-ticker.C:
				paths, globErr := filepath.Glob(filepath.Join(os.TempDir(), "a112-mb-us-goroot-*"))
				if globErr != nil {
					return
				}
				for _, path := range paths {
					if _, existed := known[path]; !existed {
						close(privateRootCreated)
						return
					}
				}
			}
		}
	}()
	go func() {
		select {
		case <-privateRootCreated:
			cancel()
		case <-stopWatching:
		}
	}()
	if _, err := snapshotGoDistributionContext(ctx, goPath, digestBytes([]byte("not-read"))); !errors.Is(err, context.Canceled) {
		close(stopWatching)
		t.Fatalf("mid-copy cancelled distribution snapshot err=%v want context.Canceled", err)
	}
	close(stopWatching)
	after, err := filepath.Glob(filepath.Join(os.TempDir(), "a112-mb-us-goroot-*"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range after {
		if _, existed := known[path]; !existed {
			t.Fatalf("cancelled distribution snapshot leaked private root %q", path)
		}
	}
}

func TestPrivateGoDistributionManifestIsDeterministic(t *testing.T) {
	root := t.TempDir()
	goPath := writeMiniGoDistribution(t, root, "printf go1.deterministic\n")
	first, err := verifiedGoToolchain(goPath)
	if err != nil {
		t.Fatal(err)
	}
	second, err := verifiedGoToolchain(goPath)
	if err != nil {
		_ = first.close()
		t.Fatal(err)
	}
	if first.distribution.digest != second.distribution.digest {
		t.Fatalf("private distribution manifest drifted across identical snapshots: %q != %q", first.distribution.digest, second.distribution.digest)
	}
	firstRoot, secondRoot := first.distribution.privateRoot, second.distribution.privateRoot
	if err := errors.Join(first.close(), second.close()); err != nil {
		t.Fatal(err)
	}
	for _, privateRoot := range []string{firstRoot, secondRoot} {
		if _, err := os.Stat(privateRoot); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("deterministic snapshot cleanup left %q: %v", privateRoot, err)
		}
	}
}

func TestSelectedGoBinaryAloneServesIdentityAndBuildClosure(t *testing.T) {
	root, err := canonicalRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	moduleCache := filepath.Join(dir, "module-cache")
	if err := os.Mkdir(moduleCache, 0o700); err != nil {
		t.Fatal(err)
	}
	goPath, logPath := writeRecordingGoBinary(t, dir, root, moduleCache)
	selected, err := verifiedGoToolchain(goPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = selected.close() }()
	t.Setenv("GOROOT", "/hostile/goroot")
	t.Setenv("PATH", "/hostile/path")
	t.Setenv("GOFLAGS", "-toolexec=/hostile/toolexec")
	t.Setenv("GOENV", "/hostile/goenv")
	t.Setenv("GOWORK", "/hostile/go.work")
	t.Setenv("GOPROXY", "https://hostile.invalid")
	t.Setenv("GOSUMDB", "hostile.invalid")
	t.Setenv("CGO_ENABLED", "1")
	t.Setenv("CC", "/hostile/cc")
	t.Setenv("CXX", "/hostile/cxx")
	if version, err := selectedGoEnv(context.Background(), selected, "GOVERSION"); err != nil || version != "go1.selected" {
		t.Fatalf("selected go env GOVERSION=%q err=%v", version, err)
	}
	if target, err := selectedGoEnv(context.Background(), selected, "GOOS"); err != nil || target != "linux" {
		t.Fatalf("selected go env GOOS=%q err=%v", target, err)
	}
	cache, err := verifiedModuleCache(context.Background(), selected)
	if err != nil || cache != moduleCache {
		t.Fatalf("selected go env GOMODCACHE=%q err=%v", cache, err)
	}
	buildCache := filepath.Join(dir, "build-cache")
	if err := os.Mkdir(buildCache, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := verifyOfflineModuleClosure(context.Background(), selected, root, buildCache, moduleCache); err != nil {
		t.Fatalf("selected go module/list closure: %v", err)
	}
	if _, err := compiledSourceClosure(context.Background(), root, selected, "linux"); err != nil {
		t.Fatalf("selected go compiled closure: %v", err)
	}
	if _, command, err := (systemBuildRunner{}).rebuild(context.Background(), root, selected); err != nil || command != prescribedBuildCommand {
		t.Fatalf("selected go prescribed build: command=%q err=%v", command, err)
	}
	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"cmd:env GOVERSION", "cmd:env GOOS", "cmd:env GOMODCACHE", "cmd:mod verify", "cmd:list -mod=readonly -deps", "cmd:list -mod=readonly -json -deps", "cmd:build -mod=readonly -trimpath -buildvcs=false"} {
		if !bytes.Contains(log, []byte(command)) {
			t.Fatalf("selected Go did not receive %q:\n%s", command, log)
		}
	}
	for _, hostile := range []string{"/hostile/goroot", "/hostile/path", "-toolexec=/hostile/toolexec", "/hostile/goenv", "/hostile/go.work", "https://hostile.invalid", "hostile.invalid", "/hostile/cc", "/hostile/cxx"} {
		if bytes.Contains(log, []byte(hostile)) {
			t.Fatalf("hostile inherited environment controlled selected Go: %q in\n%s", hostile, log)
		}
	}
}

func TestTrimpathBuiltCollectorCompletesOfflineIdentityPreflightBeforeReader(t *testing.T) {
	if testing.Short() {
		t.Skip("real prescribed trimpath preflight")
	}
	repoRoot, err := canonicalRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })
	selected, err := verifiedGoToolchain(realGoBinary(t))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = selected.close() }()
	collector := buildTrimpathCollector(t, repoRoot, selected)
	receiptRoot := privateRoot(t)
	reader := &recordingReader{}
	cfg := config{tokenCache: filepath.Join(receiptRoot, "missing-token-cache"), goBinary: selected.Path, receiptRoot: receiptRoot, sessionDate: "2026-08-14"}
	started := fixedNow()
	clockCalls := 0
	err = run(context.Background(), cfg, dependencies{
		reader: reader,
		now: func() time.Time {
			clockCalls++
			if clockCalls >= 8 {
				return started.Add(overallBudget - requestBudget + time.Nanosecond)
			}
			return started
		},
		identity: systemIdentity{executable: func() (string, error) { return collector, nil }},
		builder:  systemBuildRunner{},
	})
	if err == nil || !strings.Contains(err.Error(), "less than 15 seconds remain") {
		t.Fatalf("trimpath collector did not complete offline identity preflight: %v", err)
	}
	if len(reader.calls) != 0 {
		t.Fatalf("trimpath identity preflight reached recording reader: %v", reader.calls)
	}
	preflight := filepath.Join(receiptRunDirectory(t, receiptRoot), "preflight.json")
	data, err := os.ReadFile(preflight)
	if err != nil {
		t.Fatalf("preflight sentinel missing after trimpath identity: %v", err)
	}
	if !bytes.Contains(data, []byte(`"go_binary":"`+selected.Path+`"`)) || !bytes.Contains(data, []byte(`"go_toolchain_sha256":"`+selected.SHA256+`"`)) {
		t.Fatalf("preflight did not seal selected Go identity: %s", data)
	}
}

func TestTrimpathBuiltCollectorProcessHoldsBeforeExternalReadOnMissingCache(t *testing.T) {
	if testing.Short() {
		t.Skip("real trimpath collector process")
	}
	repoRoot, err := canonicalRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	selected, err := verifiedGoToolchain(realGoBinary(t))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = selected.close() }()
	collector := buildTrimpathCollector(t, repoRoot, selected)
	receiptRoot := privateRoot(t)
	command := exec.Command(collector, "--token-cache", filepath.Join(receiptRoot, "missing-token-cache"), "--go-binary", selected.Path, "--receipt-root", receiptRoot, "--session-date", "2026-08-14", "--before=")
	command.Dir = repoRoot
	command.Env = []string{"HOME=/nonexistent", "GOROOT=/hostile/goroot", "PATH=/hostile/path", "GOFLAGS=-toolexec=/hostile/toolexec", "GOENV=/hostile/goenv", "GOWORK=/hostile/go.work", "GOPROXY=https://hostile.invalid", "GOSUMDB=hostile.invalid", "CGO_ENABLED=1", "CC=/hostile/cc", "CXX=/hostile/cxx"}
	output, err := command.CombinedOutput()
	if err == nil || !bytes.Contains(output, []byte("candle page 1")) {
		t.Fatalf("trimpath collector unexpectedly passed or missed cache HOLD: err=%v output=%s", err, output)
	}
	run := receiptRunDirectory(t, receiptRoot)
	if _, err := os.Stat(filepath.Join(run, "preflight.json")); err != nil {
		t.Fatalf("collector process did not finish identity preflight: %v", err)
	}
	if raw, err := filepath.Glob(filepath.Join(run, "*.raw.json")); err != nil || len(raw) != 0 {
		t.Fatalf("missing cache path produced external raw receipt: matches=%v err=%v", raw, err)
	}
}

func TestIdentityComparisonDetectsUntrackedContentChangeWithSameName(t *testing.T) {
	first := identitySnapshot{Untracked: []string{"tools/a112-mb-us-source/new.go"}, UntrackedContent: map[string]string{"tools/a112-mb-us-source/new.go": "sha256:first"}}
	second := identitySnapshot{Untracked: []string{"tools/a112-mb-us-source/new.go"}, UntrackedContent: map[string]string{"tools/a112-mb-us-source/new.go": "sha256:second"}}
	if first.equal(second) {
		t.Fatal("same-name untracked content drift was accepted")
	}
}

func TestVerifyPrescribedBinaryRejectsExecutableHashMismatch(t *testing.T) {
	err := verifyPrescribedBinary(context.Background(), identitySnapshot{RepoRoot: t.TempDir(), ExecutableSHA256: digestBytes([]byte("running"))}, fakeBuildRunner{binary: []byte("rebuilt")})
	if err == nil {
		t.Fatal("prescribed rebuild mismatch was accepted")
	}
}

func TestValidateFrozenBaseRejectsInvalidOrUnknownCommit(t *testing.T) {
	for _, base := range []string{"not-a-commit", strings.Repeat("0", 40)} {
		if err := validateFrozenBase(base); err == nil {
			t.Fatalf("invalid frozen base %q was accepted", base)
		}
	}
}

func TestSystemIdentityRejectsExecutionOutsideCanonicalRepoRoot(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if _, err := (systemIdentity{}).snapshot(config{}); err == nil {
		t.Fatal("lookalike/non-root execution was accepted")
	}
}

func TestSystemBuildRunnerUsesPrescribedOfflineBuild(t *testing.T) {
	if testing.Short() || os.Getenv("A112_REAL_BUILD_TEST") != "1" {
		t.Skip("real prescribed build")
	}
	root, err := canonicalRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	goBinary, err := verifiedGoToolchain(realGoBinary(t))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = goBinary.close() }()
	binary, command, err := (systemBuildRunner{}).rebuild(context.Background(), root, goBinary)
	if err != nil {
		t.Fatalf("prescribed build: %v", err)
	}
	if command != prescribedBuildCommand || len(binary) == 0 {
		t.Fatalf("unexpected prescribed build evidence: command=%q bytes=%d", command, len(binary))
	}
}

func TestPrescribedBuildDoesNotInheritAttackerGOFLAGSToolExec(t *testing.T) {
	if testing.Short() || os.Getenv("A112_REAL_BUILD_TEST") != "1" {
		t.Skip("real prescribed build")
	}
	root, err := canonicalRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	marker := filepath.Join(dir, "toolexec-used")
	wrapper := filepath.Join(dir, "toolexec.sh")
	script := "#!/bin/sh\nprintf used > '" + marker + "'\nexec \"$@\"\n"
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	old, hadOld := os.LookupEnv("GOFLAGS")
	if err := os.Setenv("GOFLAGS", "-toolexec="+wrapper); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if hadOld {
			_ = os.Setenv("GOFLAGS", old)
		} else {
			_ = os.Unsetenv("GOFLAGS")
		}
	})
	goBinary, err := verifiedGoToolchain(realGoBinary(t))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = goBinary.close() }()
	if _, _, err := (systemBuildRunner{}).rebuild(context.Background(), root, goBinary); err != nil {
		t.Fatalf("rebuild under inherited GOFLAGS: %v", err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("inherited GOFLAGS -toolexec controlled prescribed rebuild")
	}
}

func TestOfflineClosureRunsModuleVerificationBeforeList(t *testing.T) {
	source, err := os.ReadFile("identity.go")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(source, []byte(`"mod", "verify"`)) {
		t.Fatal("offline closure accepts a tampered extracted module because go mod verify is absent")
	}
}

func TestIdentityIncludesDerivedCompiledSourceClosure(t *testing.T) {
	source, err := os.ReadFile("identity.go")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(source, []byte("CompiledSourceSHA256")) || !bytes.Contains(source, []byte(`"list", "-mod=readonly", "-json", "-deps"`)) {
		t.Fatal("same-package compiled evil.go would be omitted from source identity")
	}
}

func TestCompiledClosureRejectsExtraUntrackedA112OfficialFile(t *testing.T) {
	if err := validateCompiledInput("internal/official/a112_mbus_evil.go"); err == nil {
		t.Fatal("untracked nonignored a112_mbus_evil.go was accepted by a broad prefix allow")
	}
}

func TestSealRejectsPayloadByteTamperBeforeManifest(t *testing.T) {
	root := privateRoot(t)
	store, err := openReceipt(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.close()
	if err := store.writeJSON("preflight.json", receiptPreflight{}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(receiptRunDirectory(t, root), "preflight.json"), []byte("tampered\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := store.seal(receiptSeal{}); err == nil {
		t.Fatal("tampered payload was sealed")
	}
}

func TestVerifySealedRejectsManifestContentTamper(t *testing.T) {
	root := privateRoot(t)
	store, err := openReceipt(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.close()
	if err := store.writeJSON("preflight.json", receiptPreflight{}); err != nil {
		t.Fatal(err)
	}
	if err := store.seal(receiptSeal{}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(receiptRunDirectory(t, root), "manifest.json"), []byte(`{"schema":"forged"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := store.verifySealed(); err == nil {
		t.Fatal("forged manifest content was accepted")
	}
}

func TestReceiptMetadataBindsCanonicalQuery(t *testing.T) {
	meta, err := marshalReceiptMeta(sourceResult{query: "adjusted=false&before=A%2FB&count=200"})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(meta, []byte(`"canonical_query":"adjusted=false&before=A%2FB&count=200"`)) {
		t.Fatal("canonical query missing")
	}
}

func TestSecretScanRejectsNestedDuplicateKeyHiddenBearer(t *testing.T) {
	if safeRawBody([]byte(`{"result":{"note":"Bearer abcdefghijklmnop","note":"safe"}}`)) {
		t.Fatal("duplicate-key hidden bearer accepted")
	}
}

func TestIdentityUsesVettedSanitizedGitAndNoFollowSources(t *testing.T) {
	source, err := os.ReadFile("identity.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, needle := range []string{"verifiedGitBinary", "snapshotExecutionBinary(candidate)", "git.commandPath()", "sanitizedGitEnvironment", "readRegularNoFollow"} {
		if !bytes.Contains(source, []byte(needle)) {
			t.Fatalf("identity hardening %s missing", needle)
		}
	}
	if bytes.Contains(source, []byte("exec.Command(git,")) {
		t.Fatal("verified Git command reopened caller pathname")
	}
}

func TestIdentityDoesNotDiscoverGoAuthorityFromRuntimeOrAmbientEnvironment(t *testing.T) {
	source, err := os.ReadFile("identity.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"runtime.GOROOT", "exec.LookPath(\"go\")", "os.Getenv(\"GOROOT\")"} {
		if bytes.Contains(source, []byte(forbidden)) {
			t.Fatalf("forbidden Go authority locator remains: %s", forbidden)
		}
	}
	for _, required := range []string{"verifiedGoToolchainContext(ctx, cfg.goBinary)", "snapshotExecutionBinary(path)", "snapshotGoDistributionContext(ctx, path, snapshot.digest)", "goBinary.commandPath()", "goBinary.revalidate()"} {
		if !bytes.Contains(source, []byte(required)) {
			t.Fatalf("selected Go identity binding is missing: %s", required)
		}
	}
	for _, forbidden := range []string{"exec.CommandContext(ctx, goBinary.Path", "exec.Command(goBinary.Path"} {
		if bytes.Contains(source, []byte(forbidden)) {
			t.Fatalf("verified Go command reopened caller pathname: %s", forbidden)
		}
	}
}

func TestIdentityIgnoresHostileGitEnvironment(t *testing.T) {
	want, err := filepath.EvalSymlinks("../..")
	if err != nil {
		t.Fatal(err)
	}
	want, err = filepath.Abs(want)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())
	t.Setenv("GIT_DIR", filepath.Join(t.TempDir(), "attacker.git"))
	t.Setenv("GIT_INDEX_FILE", filepath.Join(t.TempDir(), "attacker.index"))
	got, err := canonicalRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("root=%q want=%q", got, want)
	}
}

func TestIdentityReadRejectsSourceSymlink(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real.go")
	link := filepath.Join(dir, "source.go")
	if err := os.WriteFile(real, []byte("package source\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if _, err := readRegularNoFollow(link); err == nil {
		t.Fatal("source symlink accepted")
	}
}

func TestVerifySealedRejectsPayloadTamperAfterManifest(t *testing.T) {
	root := privateRoot(t)
	store, err := openReceipt(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.close()
	if err := store.writeJSON("preflight.json", receiptPreflight{}); err != nil {
		t.Fatal(err)
	}
	if err := store.seal(receiptSeal{}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(receiptRunDirectory(t, root), "preflight.json"), []byte("same mode, new bytes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := store.verifySealed(); err == nil {
		t.Fatal("post-manifest payload tamper accepted")
	}
}

func TestSecretScanRejectsDuplicateKeysAtEveryDepth(t *testing.T) {
	for _, body := range []string{
		`{"authorization":"Bearer abcdefghijklmnop","authorization":"safe"}`,
		`{"outer":{"api_key":"token_abcdefghijklmnop","api_key":"safe"}}`,
		`{"outer":[{"credential":"abcdefghijklm","credential":"safe"}]}`,
	} {
		if safeRawBody([]byte(body)) {
			t.Fatalf("duplicate-key body accepted: %s", body)
		}
	}
}

func TestCollectorLiveRejectsExactOverallDeadline(t *testing.T) {
	start := fixedNow()
	for _, tc := range []struct {
		delta time.Duration
		hold  bool
	}{{overallBudget - time.Nanosecond, false}, {overallBudget, true}, {overallBudget + time.Nanosecond, true}} {
		err := collectorLive(context.Background(), func() time.Time { return start.Add(tc.delta) }, start)
		if (err != nil) != tc.hold {
			t.Fatalf("delta=%v err=%v", tc.delta, err)
		}
	}
}

func TestSealAndVerifyRejectRunDirectoryModeDrift(t *testing.T) {
	for _, mode := range []os.FileMode{0o755, 0o777} {
		t.Run(mode.String(), func(t *testing.T) {
			root := privateRoot(t)
			store, err := openReceipt(root)
			if err != nil {
				t.Fatal(err)
			}
			defer store.close()
			if err := store.writeJSON("preflight.json", receiptPreflight{}); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(receiptRunDirectory(t, root), mode); err != nil {
				t.Fatal(err)
			}
			sealErr := store.seal(receiptSeal{})
			verifyErr := store.verifySealed()
			if sealErr == nil || verifyErr == nil {
				t.Fatalf("run directory mode %o accepted: seal=%v verify=%v", mode, sealErr, verifyErr)
			}
		})
	}
}

func TestIdentityReadRejectsBeyond256MiBWithoutTailCollision(t *testing.T) {
	if maxIdentityInputBytes != 256<<20 {
		t.Fatalf("production identity limit=%d", maxIdentityInputBytes)
	}
	const limit = int64(16)
	dir := t.TempDir()
	exact := filepath.Join(dir, "exact")
	over := filepath.Join(dir, "over")
	prefix := []byte("0123456789abcdef")
	if err := os.WriteFile(exact, prefix, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(over, append(append([]byte(nil), prefix...), 'X'), 0o600); err != nil {
		t.Fatal(err)
	}
	exactBytes, exactErr := readRegularNoFollowWithLimit(exact, limit)
	exactDigest := digestBytes(exactBytes)
	overBytes, overErr := readRegularNoFollowWithLimit(over, limit)
	if exactErr != nil || len(exactBytes) != int(limit) || exactDigest != digestBytes(prefix) || len(overBytes) != 0 || overErr == nil {
		t.Fatalf("limit handling: exactErr=%v exactLen=%d overErr=%v overLen=%d", exactErr, len(exactBytes), overErr, len(overBytes))
	}
}

func TestReceiptPayloadReadRejectsAppendedTailBeyond64MiB(t *testing.T) {
	if maxReceiptPayloadBytes != 64<<20 {
		t.Fatalf("production receipt payload limit=%d", maxReceiptPayloadBytes)
	}
	const limit = 4 << 10
	root := privateRoot(t)
	store, err := openReceipt(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.close()
	receipt := store.(*unixReceipt)
	receipt.payloadLimit = limit
	if err := receipt.write("payload.bin", make([]byte, limit)); err != nil {
		t.Fatal(err)
	}
	if err := receipt.seal(receiptSeal{}); err != nil {
		t.Fatalf("exact injected limit did not seal: %v", err)
	}
	if err := receipt.verifySealed(); err != nil {
		t.Fatalf("exact injected limit did not verify: %v", err)
	}
	payload := filepath.Join(receiptRunDirectory(t, root), "payload.bin")
	file, err := os.OpenFile(payload, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte{'X'}); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	tailAccepted := receipt.verifySealed() == nil

	overRoot := privateRoot(t)
	overStore, err := openReceipt(overRoot)
	if err != nil {
		t.Fatal(err)
	}
	defer overStore.close()
	overReceipt := overStore.(*unixReceipt)
	overReceipt.payloadLimit = limit
	if err := overReceipt.write("payload.bin", make([]byte, limit+1)); err != nil {
		t.Fatal(err)
	}
	if err := overReceipt.seal(receiptSeal{}); err == nil {
		t.Fatal("injected limit+1 payload was sealed")
	}
	if tailAccepted {
		t.Fatal("post-seal tail beyond injected limit was accepted")
	}
}

func TestSplitNULPathsPreservesWhitespaceInUntrackedNames(t *testing.T) {
	got := splitNULPaths([]byte("normal.go\x00file with spaces.go\x00"))
	want := []string{"normal.go", "file with spaces.go"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("paths=%q want=%q", got, want)
	}
}

func privateRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}
	return root
}

func testConfig(root string) config {
	return config{tokenCache: filepath.Join(root, "token-cache.json"), goBinary: "/tmp/test-go", receiptRoot: root, sessionDate: "2026-08-14", before: []byte("first")}
}

func testDependencies(reader sourceReader) dependencies {
	return dependencies{reader: reader, now: fixedNow, identity: testIdentity(), builder: testBuilder()}
}
func fixedNow() time.Time { return time.Date(2026, 8, 15, 1, 0, 0, 0, time.UTC) }

func realGoBinary(t *testing.T) string {
	t.Helper()
	if explicit := os.Getenv("A112_TEST_GO_BINARY"); explicit != "" {
		data, err := readRegularNoFollow(explicit)
		if err != nil {
			t.Fatalf("A112_TEST_GO_BINARY is unsafe: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("A112_TEST_GO_BINARY is empty")
		}
		return explicit
	}
	path, err := exec.LookPath("go")
	if err != nil {
		t.Skipf("test Go binary unavailable: %v", err)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		t.Skipf("test Go binary is not resolvable: %v", err)
	}
	data, err := readRegularNoFollow(path)
	if err != nil {
		t.Skipf("test Go binary is not a safe explicit binary: %v", err)
	}
	if len(data) == 0 {
		t.Skip("test Go binary is empty")
	}
	return path
}

func writeSafeGoBinary(t *testing.T, directory, name string) string {
	t.Helper()
	path := filepath.Join(directory, name, "bin", "go")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Dir(filepath.Dir(path)), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeMiniGoDistribution(t *testing.T, root, goEnvBody string) string {
	t.Helper()
	for _, directory := range []string{root, filepath.Join(root, "bin"), filepath.Join(root, "src"), filepath.Join(root, "pkg"), filepath.Join(root, "pkg", "tool"), filepath.Join(root, "pkg", "tool", "linux_amd64")} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(root, "bin", "go")
	script := "#!/bin/sh\ncase \"$1:$2\" in\nenv:GOVERSION)\n" + goEnvBody + ";;\nenv:GOOS) printf linux ;;\nenv:GOMODCACHE) printf /tmp/module-cache ;;\n*) exit 9 ;;\nesac\n"
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeRecordingGoBinary(t *testing.T, directory, repoRoot, moduleCache string) (string, string) {
	t.Helper()
	path := filepath.Join(directory, "selected-goroot", "bin", "go")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Dir(filepath.Dir(path)), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(directory, "selected-go.log")
	script := fmt.Sprintf(`#!/bin/sh
printf 'cmd:%%s goroot:%%s path:%%s goflags:%%s goenv:%%s gowork:%%s goproxy:%%s gosumdb:%%s cgo:%%s cc:%%s cxx:%%s\n' "$*" "${GOROOT-}" "${PATH-}" "${GOFLAGS-}" "${GOENV-}" "${GOWORK-}" "${GOPROXY-}" "${GOSUMDB-}" "${CGO_ENABLED-}" "${CC-}" "${CXX-}" >> %q
case "$1" in
env)
  case "$2" in
  GOVERSION) printf 'go1.selected\n' ;;
  GOOS) printf 'linux\n' ;;
  GOMODCACHE) printf '%%s\n' %q ;;
  *) exit 11 ;;
  esac
  ;;
mod) exit 0 ;;
list)
  case "$*" in
  *-json*) printf '%%s\n' %q ;;
  esac
  ;;
build)
  output=''
  previous=''
  for argument in "$@"; do
    if [ "$previous" = '-o' ]; then output="$argument"; fi
    previous="$argument"
  done
  [ -n "$output" ] || exit 12
  printf 'rebuilt' > "$output"
  ;;
*) exit 13 ;;
esac
`, logPath, moduleCache, fmt.Sprintf(`{"Dir":%q,"GoFiles":[]}`, filepath.Join(repoRoot, "tools", "a112-mb-us-source")))
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path, logPath
}

func buildTrimpathCollector(t *testing.T, repoRoot string, selected goBinaryIdentity) string {
	t.Helper()
	moduleCache, err := verifiedModuleCache(context.Background(), selected)
	if err != nil {
		t.Fatalf("verified module cache: %v", err)
	}
	temporary, err := os.MkdirTemp("/tmp", "a112-mb-us-trimpath-")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(temporary, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(temporary) })
	cache := filepath.Join(temporary, "gocache")
	if err := os.Mkdir(cache, 0o700); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(temporary, "a112-mb-us-source")
	if err := selected.revalidate(); err != nil {
		t.Fatal(err)
	}
	executable, err := selected.commandPath()
	if err != nil {
		t.Fatal(err)
	}
	environment, err := prescribedBuildEnvironment(selected, cache, moduleCache)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "build", "-mod=readonly", "-trimpath", "-buildvcs=false", "-o", output, "./tools/a112-mb-us-source")
	command.Dir = repoRoot
	command.Env = environment
	if data, err := command.CombinedOutput(); err != nil {
		t.Fatalf("trimpath build: %v\n%s", err, data)
	}
	if err := selected.revalidate(); err != nil {
		t.Fatal(err)
	}
	if _, err := readRegularNoFollow(output); err != nil {
		t.Fatalf("trimpath output identity: %v", err)
	}
	return output
}

func assertSealedPrivateReceipt(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read receipt root: %v", err)
	}
	var runEntry os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "a112-mb-us-") {
			if runEntry != nil {
				t.Fatalf("multiple receipt directories: %v", entries)
			}
			runEntry = entry
		}
	}
	if runEntry == nil {
		t.Fatalf("receipt root entries=%v err=%v", entries, err)
	}
	run := filepath.Join(root, runEntry.Name())
	info, err := os.Stat(run)
	if err != nil || info.Mode().Perm() != 0o700 {
		t.Fatalf("receipt directory mode=%v err=%v", info.Mode(), err)
	}
	manifestBytes, err := os.ReadFile(filepath.Join(run, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Files []struct {
			Name string `json:"name"`
		}
	}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, file := range manifest.Files {
		if file.Name == "manifest.json" {
			t.Fatal("manifest must be self-excluding")
		}
		fileInfo, err := os.Stat(filepath.Join(run, file.Name))
		if err != nil || fileInfo.Mode().Perm() != 0o600 {
			t.Fatalf("receipt payload %q mode=%v err=%v", file.Name, fileInfo.Mode(), err)
		}
	}
}

type pagedReader struct {
	pages                         []sourceResult
	befores                       [][]byte
	orderbookCalls, calendarCalls int
}

func (r *pagedReader) Candle(_ context.Context, before []byte) (sourceResult, error) {
	r.befores = append(r.befores, bytes.Clone(before))
	if len(r.pages) == 0 {
		return sourceResult{}, errors.New("unexpected candle page")
	}
	next := r.pages[0]
	r.pages = r.pages[1:]
	return next, nil
}
func (r *pagedReader) Orderbook(context.Context) (sourceResult, error) {
	r.orderbookCalls++
	return sourceResult{body: []byte(`{"result":{}}`), rateHeaders: validRateHeaders()}, nil
}
func (r *pagedReader) Calendar(context.Context, string) (sourceResult, error) {
	r.calendarCalls++
	return sourceResult{body: []byte(`{"result":{}}`), rateHeaders: validRateHeaders()}, nil
}

// orderedReader는 준비된 캔들 페이지를 순서대로 돌려주면서 모든 호출을 순서 그대로 적는다.
// 캔들이 몇 번 불렸는지와 orderbook/calendar가 그 뒤에 왔는지를 한 줄로 확인하려는 것이다.
type orderedReader struct {
	pages []sourceResult
	calls []string
}

func (r *orderedReader) Candle(_ context.Context, before []byte) (sourceResult, error) {
	r.calls = append(r.calls, "candle:"+string(before))
	if len(r.pages) == 0 {
		return sourceResult{}, errors.New("unexpected candle page")
	}
	next := r.pages[0]
	r.pages = r.pages[1:]
	return next, nil
}

func (r *orderedReader) Orderbook(context.Context) (sourceResult, error) {
	r.calls = append(r.calls, "orderbook")
	return sourceResult{body: []byte(`{"result":{}}`), rateHeaders: validRateHeaders()}, nil
}

func (r *orderedReader) Calendar(_ context.Context, date string) (sourceResult, error) {
	r.calls = append(r.calls, "calendar:"+date)
	return sourceResult{body: []byte(`{"result":{}}`), rateHeaders: validRateHeaders()}, nil
}

// candlePageWithCursor는 null이 아닌 문자열 cursor 하나를 가진 캔들 페이지를 만든다.
func candlePageWithCursor(value string) sourceResult {
	return sourceResult{
		body:        []byte(`{"result":{"candles":[],"nextBefore":"` + value + `"}}`),
		cursorJSON:  []byte(`"` + value + `"`),
		cursorValue: []byte(value),
		rateHeaders: validRateHeaders(),
	}
}

// candlePageWithCursorAndRemaining은 남은 rate 예산만 바꾼 캔들 페이지를 만든다.
func candlePageWithCursorAndRemaining(value, remaining string) sourceResult {
	page := candlePageWithCursor(value)
	page.rateHeaders = map[string][]string{
		"X-Ratelimit-Limit": {"10"}, "X-Ratelimit-Remaining": {remaining}, "X-Ratelimit-Reset": {"1"},
	}
	return page
}

// candlePageWithNullCursor는 raw null로 끝나는 마지막 캔들 페이지를 만든다.
func candlePageWithNullCursor() sourceResult {
	return sourceResult{
		body:        []byte(`{"result":{"candles":[],"nextBefore":null}}`),
		cursorJSON:  []byte("null"),
		rateHeaders: validRateHeaders(),
	}
}

// crawlRecordFailingStore는 candle-crawl.json 쓰기만 실패시키고 나머지는 진짜 receipt에
// 그대로 넘긴다. 그 한 번의 실패를 run이 삼키지 않는지 보려는 것이다.
type crawlRecordFailingStore struct {
	receiptStore
}

func (s *crawlRecordFailingStore) writeJSON(name string, value any) error {
	if name == "candle-crawl.json" {
		return errors.New("injected candle crawl record write failure")
	}
	return s.receiptStore.writeJSON(name, value)
}

func (s *crawlRecordFailingStore) setLive(live func() error) {
	if guarded, ok := s.receiptStore.(interface{ setLive(func() error) }); ok {
		guarded.setLive(live)
	}
}

// candleMetaRecord는 candle-NN.meta.json에서 cursor 해시만 꺼내 보기 위한 모양이다.
type candleMetaRecord struct {
	CursorValueSHA256 string `json:"cursor_value_sha256"`
}

func readCandleMetaRecord(t *testing.T, runDirectory, name string) candleMetaRecord {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(runDirectory, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	var record candleMetaRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	return record
}

// candleCrawlRecord는 receipt에 적힌 candle-crawl.json을 테스트에서 읽기 위한 모양이다.
type candleCrawlRecord struct {
	Schema           string `json:"schema"`
	Pages            int    `json:"pages"`
	Terminal         string `json:"terminal"`
	LastCursorSHA256 string `json:"last_cursor_sha256"`
}

func readCandleCrawlRecord(t *testing.T, runDirectory string) candleCrawlRecord {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(runDirectory, "candle-crawl.json"))
	if err != nil {
		t.Fatalf("read candle-crawl.json: %v", err)
	}
	var record candleCrawlRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("decode candle-crawl.json: %v", err)
	}
	return record
}

func manifestFileNames(t *testing.T, runDirectory string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(runDirectory, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest.json: %v", err)
	}
	var manifest struct {
		Files []struct {
			Name string `json:"name"`
		} `json:"files"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode manifest.json: %v", err)
	}
	names := make([]string, 0, len(manifest.Files))
	for _, file := range manifest.Files {
		names = append(names, file.Name)
	}
	return names
}

func manifestListsFile(t *testing.T, runDirectory, name string) bool {
	t.Helper()
	for _, listed := range manifestFileNames(t, runDirectory) {
		if listed == name {
			return true
		}
	}
	return false
}

type changingIdentity struct{ calls int }

func (i *changingIdentity) snapshot(config) (identitySnapshot, error) {
	i.calls++
	value := testSnapshot()
	value.WorktreeHEAD = fmt.Sprintf("head-%d", i.calls)
	return value, nil
}

type cancellingIdentity struct {
	calls  int
	cancel context.CancelFunc
}

func (i *cancellingIdentity) snapshot(config) (identitySnapshot, error) {
	i.calls++
	if i.calls == 2 {
		i.cancel()
	}
	return testSnapshot(), nil
}

type poisoningIdentity struct {
	calls int
	root  string
}

func (i *poisoningIdentity) snapshot(config) (identitySnapshot, error) {
	i.calls++
	if i.calls == 2 {
		entries, err := os.ReadDir(i.root)
		if err != nil {
			return identitySnapshot{}, err
		}
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "a112-mb-us-") {
				if err := os.WriteFile(filepath.Join(i.root, entry.Name(), "unexpected.txt"), []byte("poison"), 0o600); err != nil {
					return identitySnapshot{}, err
				}
				break
			}
		}
	}
	return testSnapshot(), nil
}

func receiptHasManifest(root string) bool {
	entries, _ := os.ReadDir(root)
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "a112-mb-us-") {
			_, err := os.Stat(filepath.Join(root, entry.Name(), "manifest.json"))
			return err == nil
		}
	}
	return false
}

func receiptHasRawPayload(root string) bool {
	entries, _ := os.ReadDir(root)
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "a112-mb-us-") {
			matches, _ := filepath.Glob(filepath.Join(root, entry.Name(), "*.raw.json"))
			return len(matches) != 0
		}
	}
	return false
}

func receiptRunDirectory(t *testing.T, root string) string {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "a112-mb-us-") {
			return filepath.Join(root, entry.Name())
		}
	}
	t.Fatal("receipt run directory missing")
	return ""
}

type errorReader struct{ calls int }

func (r *errorReader) Candle(context.Context, []byte) (sourceResult, error) {
	r.calls++
	return sourceResult{}, errors.New("fixture read failure")
}
func (*errorReader) Orderbook(context.Context) (sourceResult, error) {
	panic("must not call orderbook")
}
func (*errorReader) Calendar(context.Context, string) (sourceResult, error) {
	panic("must not call calendar")
}

type fakeBuildRunner struct{ binary []byte }

func (f fakeBuildRunner) rebuild(context.Context, string, goBinaryIdentity) ([]byte, string, error) {
	return f.binary, prescribedBuildCommand, nil
}

func testIdentity() fixedIdentity {
	return fixedIdentity{value: testSnapshot()}
}

func testBuilder() buildRunner { return fakeBuildRunner{binary: []byte("running")} }

func testSnapshot() identitySnapshot {
	return identitySnapshot{RepoRoot: "test-root", ExecutableSHA256: digestBytes([]byte("running")), BuildCommand: prescribedBuildCommand}
}
