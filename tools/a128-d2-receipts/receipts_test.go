package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// 사전등록 3절의 목표 시각과 요청 묶음 — KR 은 KST, US 는 서머타임 전후 오프셋이 바뀌어야 함.
func TestPlanFor(t *testing.T) {
	cases := []struct {
		market, date string
		want         []string // "before/count"
	}{
		{"KR", "2026-10-23", []string{"2026-10-23T10:00:00+09:00/6", "2026-10-23T14:10:00+09:00/76"}},
		{"US", "2026-10-23", []string{"2026-10-23T11:00:00-04:00/66", "2026-10-23T14:00:00-04:00/66"}},
		{"US", "2026-11-02", []string{"2026-11-02T11:00:00-05:00/66", "2026-11-02T14:00:00-05:00/66"}},
	}
	for _, c := range cases {
		day, _ := time.Parse("2006-01-02", c.date)
		plans, err := planFor(c.market, day)
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, p := range plans {
			got = append(got, fmt.Sprintf("%s/%d", p.Before.Format(time.RFC3339), p.Count))
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s %s: got %v want %v", c.market, c.date, got, c.want)
		}
	}
	if _, err := planFor("JP", time.Now()); err == nil {
		t.Error("unknown market must be refused")
	}
}

func j2Line(market, symbol, at, question string) string {
	return fmt.Sprintf(`{"kind":"judgment","question_id":%q,"market":%q,"symbol":%q,"at":%q,"p":0.5}`, question, market, symbol, at)
}

// 목록 = 그 세션 날짜의 현지 10:00:00 이상 10:10:00 미만 J2 판단 종목(사전등록 2절).
func TestListSymbols(t *testing.T) {
	lines := []string{
		j2Line("KR", "000002", "2026-10-23T01:09:59Z", j2Question), // 10:09:59 KST — 포함
		j2Line("KR", "000001", "2026-10-23T01:00:00Z", j2Question), // 10:00 — 포함
		j2Line("KR", "000001", "2026-10-23T01:05:00Z", j2Question), // 중복 — 한 번만
		j2Line("KR", "000003", "2026-10-23T01:10:00Z", j2Question), // 10:10 — 밖
		j2Line("KR", "000004", "2026-10-23T00:59:00Z", j2Question), // 09:59 — 밖
		j2Line("KR", "000005", "2026-10-23T01:00:00Z", "j1_other"), // 다른 질문
		j2Line("KR", "000006", "2026-10-22T01:00:00Z", j2Question), // 다른 날짜
		`{"kind":"label","judgment_id":"x"}`,
		j2Line("US", "AAA", "2026-10-23T14:00:00Z", j2Question), // 10:00 EDT
		j2Line("US", "BBB", "2026-10-23T13:59:00Z", j2Question), // 09:59 EDT — 밖
		`{"kind":"judgment","question_id":"j2_hold`,             // 쓰는 중 잘린 마지막 줄 — 건너뜀
	}
	ledger := strings.Join(lines, "\n")
	got, digest, err := listSymbols(strings.NewReader(ledger), "KR", "2026-10-23")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"000001", "000002"}) || !strings.HasPrefix(digest, "sha256:") {
		t.Fatalf("KR list %v %s", got, digest)
	}
	got, _, err = listSymbols(strings.NewReader(ledger), "US", "2026-10-23")
	if err != nil || !reflect.DeepEqual(got, []string{"AAA"}) {
		t.Fatalf("US list %v %v", got, err)
	}
	// US 서머타임 뒤(EST): 10:00 ET = 15:00Z, 14:00Z 는 09:00 ET 라 밖
	est := j2Line("US", "CCC", "2026-11-02T15:00:00Z", j2Question) + "\n" + j2Line("US", "DDD", "2026-11-02T14:00:00Z", j2Question)
	if got, _, _ := listSymbols(strings.NewReader(est), "US", "2026-11-02"); !reflect.DeepEqual(got, []string{"CCC"}) {
		t.Fatalf("US EST list %v", got)
	}
	// 중간 줄 손상은 오류
	broken := "{broken\n" + j2Line("KR", "000001", "2026-10-23T01:00:00Z", j2Question)
	if _, _, err := listSymbols(strings.NewReader(broken), "KR", "2026-10-23"); err == nil {
		t.Fatal("a corrupt middle line must be an error")
	}
}

// fakeSource 는 종목별로 정해 둔 응답을 돌려주고 받은 요청을 적음.
type fakeSource struct {
	calls []string
	reply func(symbol, before string, count int) (official.StrictMinutePage, error)
}

func (f *fakeSource) StrictMinuteCandles(_ context.Context, market, symbol string, count int, before string) (official.StrictMinutePage, error) {
	f.calls = append(f.calls, fmt.Sprintf("%s %s %s %d", market, symbol, before, count))
	return f.reply(symbol, before, count)
}

// barsUpTo 는 before 부터 거꾸로 1분 간격 봉 n 개(최신이 앞)를 만듦. 종가 = 분(minute-of-day).
func barsUpTo(before string, n int) []official.RawMinuteCandle {
	end, _ := time.Parse(time.RFC3339, before)
	out := make([]official.RawMinuteCandle, 0, n)
	for i := 0; i < n; i++ {
		at := end.Add(-time.Duration(i) * time.Minute)
		out = append(out, official.RawMinuteCandle{Timestamp: at.Format(time.RFC3339),
			Close: fmt.Sprint(at.Hour()*60 + at.Minute())})
	}
	return out
}

func krPlans(t *testing.T) []fetchPlan {
	day, _ := time.Parse("2006-01-02", "2026-10-23")
	plans, err := planFor("KR", day)
	if err != nil {
		t.Fatal(err)
	}
	return plans
}

func TestCollectKeepsOnlyTargetWindows(t *testing.T) {
	src := &fakeSource{reply: func(symbol, before string, count int) (official.StrictMinutePage, error) {
		if symbol == "000002" && strings.HasPrefix(before, "2026-10-23T10:00") {
			return official.StrictMinutePage{}, &official.StrictMinuteCandlesError{Reason: official.StrictReasonBody, Detail: "x"}
		}
		return official.StrictMinutePage{Candles: barsUpTo(before, count), BodyDigest: "sha256:ab"}, nil
	}}
	slept := 0
	sleep := func(context.Context, time.Duration) error { slept++; return nil }
	clock := time.Date(2026, 10, 23, 7, 0, 0, 0, time.UTC)
	now := func() time.Time { clock = clock.Add(time.Second); return clock }
	rc, err := collect(context.Background(), src, "KR", "2026-10-23", []string{"000001", "000002"}, krPlans(t),
		time.Millisecond, sleep, now)
	if err != nil {
		t.Fatal(err)
	}
	if rc.Calls != 4 || len(src.calls) != 4 || slept != 3 {
		t.Fatalf("calls %d/%d slept %d", rc.Calls, len(src.calls), slept)
	}
	if src.calls[0] != "KR 000001 2026-10-23T10:00:00+09:00 6" || src.calls[1] != "KR 000001 2026-10-23T14:10:00+09:00 76" {
		t.Fatalf("requests %v", src.calls)
	}
	// 000001: 09:55~10:00(6) + 12:55~13:00(6) + 14:05~14:10(6) = 18 봉, 13:01~14:04 은 버림
	var times []string
	for _, b := range rc.Bars["000001"] {
		times = append(times, b.Time[11:16])
	}
	want := []string{"09:55", "09:56", "09:57", "09:58", "09:59", "10:00", "12:55", "12:56", "12:57", "12:58", "12:59", "13:00",
		"14:05", "14:06", "14:07", "14:08", "14:09", "14:10"}
	if !reflect.DeepEqual(times, want) {
		t.Fatalf("kept %v", times)
	}
	if rc.Bars["000001"][5].Close != "600" { // 10:00 = 600 분
		t.Fatalf("close %v", rc.Bars["000001"][5])
	}
	// 000002: 10:00 요청은 계약 거절 → 기록하고 계속, 14:10 요청 봉은 남음
	if rc.Fetches[2].Status != "refused: "+official.StrictReasonBody || rc.Fetches[3].Status != "ok" || len(rc.Bars["000002"]) != 12 {
		t.Fatalf("fetches %+v bars %d", rc.Fetches, len(rc.Bars["000002"]))
	}
	if rc.Fetches[1].Candles != 76 || rc.Fetches[1].Kept != 12 || rc.ElapsedSeconds != 1 {
		t.Fatalf("record %+v elapsed %v", rc.Fetches[1], rc.ElapsedSeconds)
	}
}

// 429 는 재시도 없이 즉시 멈추고 영수증이 없음. 다른 전송 오류도 멈춤.
func TestCollectStopsOnRateLimitAndTransport(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want error
	}{{official.ErrRateLimited, errRateLimited}, {official.ErrTransport, official.ErrTransport}} {
		src := &fakeSource{reply: func(symbol, before string, count int) (official.StrictMinutePage, error) {
			if symbol == "000002" {
				return official.StrictMinutePage{}, fmt.Errorf("wrapped: %w", tc.err)
			}
			return official.StrictMinutePage{Candles: barsUpTo(before, count)}, nil
		}}
		_, err := collect(context.Background(), src, "KR", "2026-10-23", []string{"000001", "000002", "000003"}, krPlans(t),
			0, func(context.Context, time.Duration) error { return nil }, time.Now)
		if !errors.Is(err, tc.want) {
			t.Fatalf("err %v want %v", err, tc.want)
		}
		if len(src.calls) != 3 { // 000001 두 번 + 000002 첫 요청에서 멈춤
			t.Fatalf("calls after stop %v", src.calls)
		}
	}
}

// 영수증 형식은 평가기(prereg_eval.py d2_load_receipts)가 읽는 {"bars": {sym: [{"time","close"}]}} 이고, 덮어쓰기는 거절.
func TestWriteReceiptShapeAndNoOverwrite(t *testing.T) {
	dir := t.TempDir()
	rc := receipt{Market: "KR", Date: "2026-10-23", Bars: map[string][]receiptBar{
		"000001": {{Time: "2026-10-23T10:00:00+09:00", Close: "71200"}}}}
	path, err := writeReceipt(dir, rc)
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(dir, "minute-bars", "KR-2026-10-23.json") {
		t.Fatalf("path %s", path)
	}
	raw, _ := os.ReadFile(path)
	var back struct {
		Bars map[string][]struct {
			Time  string  `json:"time"`
			Close float64 `json:"close"` // 숫자여야 함(문자열이면 해독 실패)
		} `json:"bars"`
	}
	if err := json.Unmarshal(raw, &back); err != nil || back.Bars["000001"][0].Close != 71200 {
		t.Fatalf("shape %v %s", err, raw)
	}
	if _, err := writeReceipt(dir, rc); !errors.Is(err, os.ErrExist) {
		t.Fatalf("overwrite must be refused, got %v", err)
	}
}

// --config-dir 없이는 네트워크 전에 멈춤(기본 토큰 캐시 사용 금지). 목록·계획은 원장만으로 만들어짐.
func TestRunRequiresConfigDirBeforeNetwork(t *testing.T) {
	dir := t.TempDir()
	ledger := filepath.Join(dir, "ledger.jsonl")
	if err := os.WriteFile(ledger, []byte(j2Line("KR", "000001", "2026-10-23T01:00:00Z", j2Question)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := run("KR", "2026-10-23", "", ledger, dir, 0, 0, false, false)
	if err == nil || !strings.Contains(err.Error(), "--config-dir is required") {
		t.Fatalf("got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "minute-bars")); !os.IsNotExist(statErr) {
		t.Fatal("no receipt directory may be created")
	}
	if err := run("KR", "2026-10-24", "", ledger, dir, 0, 0, true, false); err == nil {
		t.Fatal("an empty list must be an error, not an empty receipt")
	}
}
