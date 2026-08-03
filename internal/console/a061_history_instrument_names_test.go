package console

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

type a061InstrumentNames struct {
	mu    sync.Mutex
	rows  []InstrumentName
	err   error
	calls int
	refs  []InstrumentRef
}

type a061BlockingInstrumentNames struct {
	entered chan struct{}
	release chan struct{}
}

func (f *a061BlockingInstrumentNames) Names(ctx context.Context, _ []InstrumentRef) ([]InstrumentName, error) {
	select {
	case <-f.entered:
	default:
		close(f.entered)
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-f.release:
		return []InstrumentName{{Market: "kr", Symbol: "005930", Name: "삼성전자"}}, nil
	}
}

func (f *a061InstrumentNames) Names(_ context.Context, refs []InstrumentRef) ([]InstrumentName, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.refs = append([]InstrumentRef(nil), refs...)
	return append([]InstrumentName(nil), f.rows...), f.err
}

func (f *a061InstrumentNames) snapshot() (int, []InstrumentRef) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls, append([]InstrumentRef(nil), f.refs...)
}

func seedA060USHistory(t *testing.T, path string) {
	t.Helper()
	execRaw(t, path, `
INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
                       quantity, avg_price, opened_at, closed_at)
VALUES ('pos-us-closed','123-45-678901','us','IONQ',1,'d-1','CLOSED','0','0',
        '2026-07-25T00:30:00Z','2026-07-25T06:30:00Z');
INSERT INTO trade_outcomes (position_id, realized_pnl_after_costs, realized_r, initial_risk,
                            initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at)
VALUES ('pos-us-closed','12.5','0.5','2','4',3600,'HALF_RISK',1,'2026-07-25T06:30:00Z');
INSERT INTO exit_events (position_id, observed_price, high_water, baseline_after, level_after,
                         action, created_at)
VALUES ('pos-us-closed','36.44','37','35','HALF_RISK','OBSERVED','2026-07-25T05:30:00Z');
`)
}

func TestA061HistoryShowsCodeAndNameForTripsAndEvents(t *testing.T) {
	names := &a061InstrumentNames{rows: []InstrumentName{
		{Market: "kr", Symbol: "005380", Name: "현대차"},
		{Market: "kr", Symbol: "005930", Name: "삼성전자"},
		{Market: "us", Symbol: "IONQ", Name: "아이온큐"},
	}}
	h := newDashboardHarness(t, func(o *Options) { o.InstrumentNames = names })
	seedJournal(t, h.journal)
	seedA060USHistory(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/history")
	for _, want := range []string{
		`<code>005380</code> · 현대차`,
		`<code>005930</code> · 삼성전자`,
		`<code>IONQ</code> · 아이온큐`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("history does not keep code and name together: missing %q", want)
		}
	}
	if got := strings.Count(page, `<th scope="col">종목</th>`); got != 2 {
		t.Errorf("history has %d semantic instrument headers, want 2", got)
	}
	got, refs := names.snapshot()
	if got != 1 || len(refs) != 4 {
		t.Fatalf("name lookup = %d call(s), %d unique ref(s), want one batch with four refs: %+v", got, len(refs), refs)
	}
	seen := map[string]bool{}
	for _, ref := range refs {
		key := symbolKey(ref.Market, ref.Symbol)
		if seen[key] {
			t.Errorf("duplicate instrument ref in batch: %s", key)
		}
		seen[key] = true
	}
}

func TestA061HistoryMissingNameKeepsTheSymbol(t *testing.T) {
	names := &a061InstrumentNames{rows: []InstrumentName{
		{Market: "kr", Symbol: "005380", Name: "현대차"},
	}}
	h := newDashboardHarness(t, func(o *Options) { o.InstrumentNames = names })
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/history")
	row, ok := historyRow(page, "051910")
	if !ok || !strings.Contains(row, `<code>051910</code>`) {
		t.Fatalf("missing metadata removed or replaced the authoritative symbol: %s", row)
	}
}

func TestA061HistoryNameFailureKeepsFrozenRowsAndSymbols(t *testing.T) {
	names := &a061InstrumentNames{err: errors.New("metadata unavailable")}
	h := newDashboardHarness(t, func(o *Options) { o.InstrumentNames = names })
	seedJournal(t, h.journal)
	h.authenticate(t)

	response := h.get(t, "/history")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET /history after name failure = %d, want 200", response.StatusCode)
	}
	page := body(t, response)
	for _, want := range []string{"005380", "48000", "2.4", "종목명 조회 실패", "metadata unavailable"} {
		if !strings.Contains(page, want) {
			t.Errorf("name failure hid frozen history or its honest warning: missing %q", want)
		}
	}
}

func TestA061HistoryUnwiredNameReaderIsExplicit(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/history")
	if !strings.Contains(page, "종목명 조회 기능이 연결되지 않아 심볼만 표시한다") {
		t.Fatal("nil metadata seam was disguised as genuine missing metadata")
	}
}

func TestA061HistoryRejectsUnsafeDisplayNames(t *testing.T) {
	names := &a061InstrumentNames{rows: []InstrumentName{
		{Market: "kr", Symbol: "005380", Name: `</code><img src=x onerror="alert(1)">`},
		{Market: "kr", Symbol: "051910", Name: "safe\u202Espoof"},
	}}
	h := newDashboardHarness(t, func(o *Options) { o.InstrumentNames = names })
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/history")
	if strings.Contains(page, `<img src=x`) || strings.Contains(page, "safe\u202Espoof") {
		t.Fatal("untrusted metadata produced markup or a bidi-spoofed label")
	}
	for _, symbol := range []string{"005380", "051910"} {
		if !strings.Contains(page, `<code>`+symbol+`</code>`) {
			t.Errorf("unsafe name rejection removed authoritative symbol %s", symbol)
		}
	}
}

func TestA061HistoryPOSTAndVerificationHoldSpendNoMetadataBudget(t *testing.T) {
	names := &a061InstrumentNames{rows: []InstrumentName{{Market: "kr", Symbol: "005380", Name: "현대차"}}}
	h := newDashboardHarness(t, func(o *Options) { o.InstrumentNames = names })
	seedJournal(t, h.journal)
	h.authenticate(t)

	response := h.post(t, "/history", url.Values{})
	if response.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("POST /history = %d, want 405", response.StatusCode)
	}
	if calls, _ := names.snapshot(); calls != 0 {
		t.Fatalf("POST /history made %d metadata call(s), want 0", calls)
	}

	holdRunLock(t, h.runlock, h.clock.Now())
	page := h.page(t, "/history")
	if calls, _ := names.snapshot(); calls != 0 {
		t.Fatalf("verification-held /history made %d metadata call(s), want 0", calls)
	}
	if !strings.Contains(page, "실계좌 검증") || !strings.Contains(page, "종목명 조회 보류") {
		t.Error("verification-held history does not explain why names are code-only")
	}
}

func TestA061InstrumentNameCacheBoundsLookupAndReusesTheResult(t *testing.T) {
	refs := make([]InstrumentRef, 0, historyInstrumentRefLimit+1)
	rows := make([]InstrumentName, 0, historyInstrumentRefLimit)
	for i := 0; i <= historyInstrumentRefLimit; i++ {
		symbol := fmt.Sprintf("K%06d", i)
		refs = append(refs, InstrumentRef{Market: "kr", Symbol: symbol})
		if i < historyInstrumentRefLimit {
			rows = append(rows, InstrumentName{Market: "kr", Symbol: symbol, Name: "name"})
		}
	}
	names := &a061InstrumentNames{rows: rows}
	var cache instrumentNameCache
	now := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	resolved, notice := cache.get(context.Background(), names, refs, now, false, "")
	if calls, gotRefs := names.snapshot(); calls != 1 || len(gotRefs) != historyInstrumentRefLimit {
		t.Fatalf("bounded lookup = %d call(s), %d refs, want one call with %d refs", calls, len(gotRefs), historyInstrumentRefLimit)
	}
	if len(resolved) != historyInstrumentRefLimit || !strings.Contains(notice, "조회 한도") {
		t.Fatalf("bounded result = %d names, notice %q", len(resolved), notice)
	}
	cache.get(context.Background(), names, refs[:historyInstrumentRefLimit], now.Add(time.Hour), false, "")
	if calls, _ := names.snapshot(); calls != 1 {
		t.Fatalf("24-hour cache made %d reader calls, want 1", calls)
	}
}

func TestA061InstrumentNameCacheExpiresAtTwentyFourHours(t *testing.T) {
	names := &a061InstrumentNames{rows: []InstrumentName{{Market: "kr", Symbol: "005930", Name: "삼성전자"}}}
	refs := []InstrumentRef{{Market: "kr", Symbol: "005930"}}
	var cache instrumentNameCache
	now := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	cache.get(context.Background(), names, refs, now, false, "")
	cache.get(context.Background(), names, refs, now.Add(instrumentNameTTL-time.Nanosecond), false, "")
	if calls, _ := names.snapshot(); calls != 1 {
		t.Fatalf("cache before TTL made %d calls, want 1", calls)
	}
	cache.get(context.Background(), names, refs, now.Add(instrumentNameTTL), false, "")
	if calls, _ := names.snapshot(); calls != 2 {
		t.Fatalf("cache at TTL boundary made %d calls, want 2", calls)
	}
}

func TestA061InstrumentNameCacheIsBoundedToTwoThousandFortyEightEntries(t *testing.T) {
	names := &a061InstrumentNames{}
	var cache instrumentNameCache
	now := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	for batch := 0; batch < 6; batch++ {
		refs := make([]InstrumentRef, historyInstrumentRefLimit)
		for i := range refs {
			refs[i] = InstrumentRef{Market: "kr", Symbol: fmt.Sprintf("%02d%06d", batch, i)}
			names.rows = append(names.rows, InstrumentName{Market: "kr", Symbol: refs[i].Symbol, Name: "name"})
		}
		cache.get(context.Background(), names, refs, now, false, "")
	}
	if got := len(cache.entries); got != instrumentNameCacheLimit {
		t.Fatalf("cache entries = %d, want bounded at %d", got, instrumentNameCacheLimit)
	}
}

func TestA061PartialMetadataGetsAWarningAndShortRetryInsteadOfANegativeDayCache(t *testing.T) {
	names := &a061InstrumentNames{rows: []InstrumentName{{Market: "kr", Symbol: "005930", Name: "삼성전자"}}}
	refs := []InstrumentRef{{Market: "kr", Symbol: "005930"}, {Market: "kr", Symbol: "000660"}}
	var cache instrumentNameCache
	now := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	resolved, notice := cache.get(context.Background(), names, refs, now, false, "")
	if resolved["kr|005930"] != "삼성전자" || !strings.Contains(notice, "일부") {
		t.Fatalf("partial response result/notice = %+v/%q", resolved, notice)
	}
	cache.get(context.Background(), names, refs, now.Add(instrumentNameRetryAfter-time.Nanosecond), false, "")
	if calls, _ := names.snapshot(); calls != 1 {
		t.Fatalf("partial response retried inside short backoff: %d calls", calls)
	}
	retried, _ := cache.get(context.Background(), names, refs, now.Add(instrumentNameRetryAfter), false, "")
	if calls, retriedRefs := names.snapshot(); calls != 2 {
		t.Fatalf("partial response at retry boundary made %d calls, want 2", calls)
	} else if len(retriedRefs) != 1 || symbolKey(retriedRefs[0].Market, retriedRefs[0].Symbol) != "kr|000660" {
		t.Fatalf("partial response retried refs = %+v, want only omitted kr|000660", retriedRefs)
	}
	if retried["kr|005930"] != "삼성전자" {
		t.Fatalf("accepted cache entry did not survive omitted-ref retry: %+v", retried)
	}
}

func TestA061ConcurrentMetadataWaitHonorsTheCallerDeadline(t *testing.T) {
	names := &a061BlockingInstrumentNames{entered: make(chan struct{}), release: make(chan struct{})}
	var cache instrumentNameCache
	refs := []InstrumentRef{{Market: "kr", Symbol: "005930"}}
	firstDone := make(chan struct{})
	go func() {
		defer close(firstDone)
		cache.get(context.Background(), names, refs, time.Now(), false, "")
	}()
	<-names.entered
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, notice := cache.get(ctx, names, refs, time.Now(), false, "")
	if elapsed := time.Since(started); elapsed > time.Second || !strings.Contains(notice, "시간 초과") {
		t.Fatalf("contended lookup elapsed/notice = %s/%q", elapsed, notice)
	}
	close(names.release)
	<-firstDone
}

func TestA061InstrumentNameCacheSingleFlightsSuccessAndFailure(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "success"},
		{name: "failure", err: errors.New("official outage")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			names := &a061InstrumentNames{rows: []InstrumentName{{Market: "kr", Symbol: "005930", Name: "삼성전자"}}, err: tc.err}
			var cache instrumentNameCache
			now := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
			var wg sync.WaitGroup
			start := make(chan struct{})
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					cache.get(context.Background(), names, []InstrumentRef{{Market: "kr", Symbol: "005930"}}, now, false, "")
				}()
			}
			close(start)
			wg.Wait()
			if calls, _ := names.snapshot(); calls != 1 {
				t.Fatalf("concurrent %s made %d reader calls, want 1", tc.name, calls)
			}
			cache.get(context.Background(), names, []InstrumentRef{{Market: "kr", Symbol: "005930"}}, now.Add(instrumentNameRetryAfter-time.Nanosecond), false, "")
			if calls, _ := names.snapshot(); calls != 1 {
				t.Fatalf("%s retried inside cache/backoff window: %d calls", tc.name, calls)
			}
			if tc.err != nil {
				cache.get(context.Background(), names, []InstrumentRef{{Market: "kr", Symbol: "005930"}}, now.Add(instrumentNameRetryAfter), false, "")
				if calls, _ := names.snapshot(); calls != 2 {
					t.Fatalf("failure backoff at exact boundary made %d calls, want 2", calls)
				}
			}
		})
	}
}

type a061DeadlineNames struct {
	calls     int
	remaining time.Duration
}

func (f *a061DeadlineNames) Names(ctx context.Context, _ []InstrumentRef) ([]InstrumentName, error) {
	f.calls++
	deadline, ok := ctx.Deadline()
	if !ok {
		return nil, errors.New("missing total metadata deadline")
	}
	f.remaining = time.Until(deadline)
	return nil, errors.New("captured deadline")
}

func TestA061InstrumentNameLookupHonorsTotalTimeout(t *testing.T) {
	reader := &a061DeadlineNames{}
	var cache instrumentNameCache
	_, notice := cache.get(context.Background(), reader, []InstrumentRef{{Market: "kr", Symbol: "005930"}}, time.Now(), false, "")
	if reader.calls != 1 || !strings.Contains(notice, "captured deadline") {
		t.Fatalf("timeout calls/notice = %d/%q", reader.calls, notice)
	}
	if reader.remaining < instrumentNameTimeout-time.Second || reader.remaining > instrumentNameTimeout {
		t.Fatalf("metadata deadline remaining = %s, want approximately %s", reader.remaining, instrumentNameTimeout)
	}
}

func TestA061InstrumentNameConflictAndUnsafeTextFailClosed(t *testing.T) {
	names := &a061InstrumentNames{rows: []InstrumentName{
		{Market: "kr", Symbol: "005930", Name: "삼성전자"},
		{Market: "kr", Symbol: "005930", Name: "다른 이름"},
	}}
	var cache instrumentNameCache
	resolved, _ := cache.get(context.Background(), names, []InstrumentRef{{Market: "kr", Symbol: "005930"}}, time.Now(), false, "")
	if len(resolved) != 0 {
		t.Fatalf("conflicting names attached: %+v", resolved)
	}
	unsafe := []string{"control\x00name", strings.Repeat("가", maxInstrumentNameRunes+1)}
	for _, r := range []rune{'\u061c', '\u200e', '\u200f', '\u202a', '\u202b', '\u202c', '\u202d', '\u202e', '\u2066', '\u2067', '\u2068', '\u2069'} {
		unsafe = append(unsafe, "name"+string(r)+"spoof")
	}
	for _, name := range unsafe {
		if got, ok := safeInstrumentName(name); ok || got != "" {
			t.Errorf("unsafe name accepted: %q", name)
		}
	}
}

func TestA061VerificationHoldUsesCachedNameWithoutNewRead(t *testing.T) {
	names := &a061InstrumentNames{rows: []InstrumentName{{Market: "kr", Symbol: "005930", Name: "삼성전자"}}}
	refs := []InstrumentRef{{Market: "kr", Symbol: "005930"}}
	var cache instrumentNameCache
	now := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	cache.get(context.Background(), names, refs, now, false, "")
	resolved, _ := cache.get(context.Background(), names, refs, now.Add(time.Hour), true, "verification")
	if calls, _ := names.snapshot(); calls != 1 || resolved["kr|005930"] != "삼성전자" {
		t.Fatalf("held cached lookup calls/result = %d/%+v", calls, resolved)
	}
}
