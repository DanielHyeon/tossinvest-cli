package officialbars

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

// ---- 시각 기준점 ----
//
// 2026-08-14은 금요일이고 미국 시장은 서머타임(EDT, -04:00)이다.
// 정규장 09:30~16:00 ET는 13:30~20:00 UTC다.
var (
	usOpen  = time.Date(2026, 8, 14, 13, 30, 0, 0, time.UTC)
	usClose = time.Date(2026, 8, 14, 20, 0, 0, 0, time.UTC)
	// 한국 정규장 09:00~15:30 KST는 00:00~06:30 UTC다.
	krOpen = time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
)

const (
	usSymbol = "AAPL"
	krSymbol = "005930"
)

func mustLocation(t *testing.T, market marketclock.Market) *time.Location {
	t.Helper()
	loc, err := market.Location()
	if err != nil {
		t.Fatalf("location for %s: %v", market, err)
	}
	return loc
}

func usCalendarAt(t *testing.T, fetchedAt time.Time) scheduler.CalendarSnapshot {
	t.Helper()
	eastern := mustLocation(t, marketclock.MarketUS)
	day := func(date string) official.MarketCalendarDay {
		parsed, err := time.ParseInLocation("2006-01-02", date, eastern)
		if err != nil {
			t.Fatalf("parsing %s: %v", date, err)
		}
		year, month, dayOfMonth := parsed.Date()
		return official.MarketCalendarDay{
			Date: date,
			RegularMarket: &official.MarketCalendarSession{
				StartTime: time.Date(year, month, dayOfMonth, 9, 30, 0, 0, eastern),
				EndTime:   time.Date(year, month, dayOfMonth, 16, 0, 0, 0, eastern),
			},
		}
	}
	snapshot, err := scheduler.AdaptOfficialCalendar(marketclock.MarketUS, official.MarketCalendarResponse{
		PreviousBusinessDay: day("2026-08-13"),
		Today:               day("2026-08-14"),
		NextBusinessDay:     day("2026-08-17"),
	}, fetchedAt)
	if err != nil {
		t.Fatalf("AdaptOfficialCalendar(US): %v", err)
	}
	return snapshot
}

func krCalendarAt(t *testing.T, fetchedAt time.Time) scheduler.CalendarSnapshot {
	t.Helper()
	seoul := mustLocation(t, marketclock.MarketKR)
	day := func(date string) official.MarketCalendarDay {
		parsed, err := time.ParseInLocation("2006-01-02", date, seoul)
		if err != nil {
			t.Fatalf("parsing %s: %v", date, err)
		}
		year, month, dayOfMonth := parsed.Date()
		return official.MarketCalendarDay{
			Date: date,
			Integrated: &official.MarketCalendarSessions{
				RegularMarket: &official.MarketCalendarSession{
					StartTime: time.Date(year, month, dayOfMonth, 9, 0, 0, 0, seoul),
					EndTime:   time.Date(year, month, dayOfMonth, 15, 30, 0, 0, seoul),
				},
			},
		}
	}
	snapshot, err := scheduler.AdaptOfficialCalendar(marketclock.MarketKR, official.MarketCalendarResponse{
		PreviousBusinessDay: day("2026-08-13"),
		Today:               day("2026-08-14"),
		NextBusinessDay:     day("2026-08-17"),
	}, fetchedAt)
	if err != nil {
		t.Fatalf("AdaptOfficialCalendar(KR): %v", err)
	}
	return snapshot
}

// brokerTimestamp는 브로커가 쓰는 그대로의 시각 표기다(KST 고정 오프셋, 밀리초 셋).
func brokerTimestamp(t *testing.T, instant time.Time) string {
	t.Helper()
	return instant.In(mustLocation(t, marketclock.MarketKR)).Format("2006-01-02T15:04:05.000-07:00")
}

func usBar(t *testing.T, open time.Time, closePrice string) official.RawMinuteCandle {
	t.Helper()
	return official.RawMinuteCandle{
		Timestamp: brokerTimestamp(t, open), Open: "231.4350", High: "231.5000",
		Low: "231.4000", Close: closePrice, Volume: "1234", Currency: "USD",
	}
}

func krBar(t *testing.T, open time.Time, closePrice string) official.RawMinuteCandle {
	t.Helper()
	return official.RawMinuteCandle{
		Timestamp: brokerTimestamp(t, open), Open: "71000", High: "71100",
		Low: "70900", Close: closePrice, Volume: "12", Currency: "KRW",
	}
}

func digestFor(label string) string {
	sum := sha256.Sum256([]byte(label))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// pageOf는 읽기 결과 한 장을 만든다. cursor가 빈 문자열이면 마지막 페이지다.
func pageOf(market, symbol string, readAt time.Time, label, cursor string, candles ...official.RawMinuteCandle) official.StrictMinutePage {
	page := official.StrictMinutePage{
		Market: market, Symbol: symbol, Candles: candles, ReadAt: readAt,
		StatusCode: 200, BodyDigest: digestFor(label),
	}
	if cursor == "" {
		page.Terminal = true
	} else {
		page.NextBefore = cursor
	}
	return page
}

func usPage(t *testing.T, readAt time.Time, label, cursor string, candles ...official.RawMinuteCandle) official.StrictMinutePage {
	t.Helper()
	return pageOf("US", usSymbol, readAt, label, cursor, candles...)
}

// ---- 가짜 리더 ----

type readerCall struct {
	market, symbol, before string
	count                  int
}

type readerReply struct {
	page official.StrictMinutePage
	err  error
}

type fakeReader struct {
	replies []readerReply
	calls   []readerCall
}

func (r *fakeReader) StrictMinuteCandles(_ context.Context, market, symbol string, count int, before string) (official.StrictMinutePage, error) {
	r.calls = append(r.calls, readerCall{market: market, symbol: symbol, before: before, count: count})
	index := len(r.calls) - 1
	if index >= len(r.replies) {
		return official.StrictMinutePage{}, fmt.Errorf("fake reader: unscripted page %d", index+1)
	}
	return r.replies[index].page, r.replies[index].err
}

func readerOf(pages ...official.StrictMinutePage) *fakeReader {
	replies := make([]readerReply, 0, len(pages))
	for _, page := range pages {
		replies = append(replies, readerReply{page: page})
	}
	return &fakeReader{replies: replies}
}

// ---- 저장소 도우미 ----

func openTestStore(t *testing.T, clock marketclock.Clock) *strategyevidence.Store {
	t.Helper()
	store, err := strategyevidence.Open(context.Background(), strategyevidence.Options{
		Path: filepath.Join(t.TempDir(), "breakout-bars.db"), Clock: clock,
	})
	if err != nil {
		t.Fatalf("strategyevidence.Open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

// cursorBefore는 실측된 커서다: 그 페이지의 가장 오래된 봉보다 정확히 1분 이르다.
func cursorBefore(t *testing.T, oldest time.Time) string {
	t.Helper()
	return brokerTimestamp(t, oldest.Add(-time.Minute))
}

// recordingStore는 적재된 봉투를 그대로 붙잡아 둔다. 머리말과 본문을 시험이 직접 볼 수 있다.
type recordingStore struct {
	inner    BarStore
	appended []strategyevidence.Envelope
}

func (s *recordingStore) Append(ctx context.Context, envelope strategyevidence.Envelope) (strategyevidence.AppendResult, error) {
	result, err := s.inner.Append(ctx, envelope)
	if err == nil {
		s.appended = append(s.appended, envelope)
	}
	return result, err
}

func (s *recordingStore) SealBarSeries(ctx context.Context, query strategyevidence.BarSeriesQuery) (strategyevidence.BarSeries, error) {
	return s.inner.SealBarSeries(ctx, query)
}

func payloadOf(t *testing.T, envelope strategyevidence.Envelope) strategyevidence.ClosedBar1mPayload {
	t.Helper()
	payload, err := strategyevidence.DecodeClosedBar1mPayload(envelope.CanonicalPayload())
	if err != nil {
		t.Fatalf("DecodeClosedBar1mPayload: %v", err)
	}
	return payload
}

// flakyStore는 정해진 순번의 Append만 실패시킨다.
type flakyStore struct {
	inner  BarStore
	failAt int
	calls  int
	err    error
}

func (s *flakyStore) Append(ctx context.Context, envelope strategyevidence.Envelope) (strategyevidence.AppendResult, error) {
	s.calls++
	if s.calls == s.failAt {
		return strategyevidence.AppendResult{}, s.err
	}
	return s.inner.Append(ctx, envelope)
}

func (s *flakyStore) SealBarSeries(ctx context.Context, query strategyevidence.BarSeriesQuery) (strategyevidence.BarSeries, error) {
	return s.inner.SealBarSeries(ctx, query)
}

// injectingStore는 중복 제거 읽기가 끝난 "직후"에 다른 글쓴이가 나타나는 상황을 만든다.
type injectingStore struct {
	inner  BarStore
	inject func()
}

func (s *injectingStore) Append(ctx context.Context, envelope strategyevidence.Envelope) (strategyevidence.AppendResult, error) {
	return s.inner.Append(ctx, envelope)
}

func (s *injectingStore) SealBarSeries(ctx context.Context, query strategyevidence.BarSeriesQuery) (strategyevidence.BarSeries, error) {
	series, err := s.inner.SealBarSeries(ctx, query)
	if s.inject != nil {
		s.inject()
	}
	return series, err
}

func usSeries(t *testing.T, store *strategyevidence.Store, evaluationAt, ingestionCutoff time.Time) strategyevidence.BarSeries {
	t.Helper()
	series, err := store.SealBarSeries(context.Background(), strategyevidence.BarSeriesQuery{
		Market: "US", Symbol: usSymbol, SessionID: "US:2026-08-14",
		IntervalMS:   strategyevidence.ClosedBar1mIntervalMS,
		EvaluationAt: evaluationAt, IngestionCutoff: ingestionCutoff,
	})
	if err != nil {
		t.Fatalf("SealBarSeries: %v", err)
	}
	return series
}

func refusalOf(t *testing.T, err error) *PollRefusal {
	t.Helper()
	var refusal *PollRefusal
	if !errors.As(err, &refusal) {
		t.Fatalf("error %v is not a PollRefusal", err)
	}
	return refusal
}

// ---- 결정 6: 폴의 가장 새 봉은 절대 적재하지 않는다 ----

func TestPollNeverAdmitsTheNewestObservedBar(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	clock := marketclock.NewFake(pollAt)
	store := openTestStore(t, clock)
	reader := readerOf(usPage(t, pollAt, "page-1", "",
		usBar(t, usOpen.Add(3*time.Minute), "231.4400"),
		usBar(t, usOpen.Add(2*time.Minute), "231.4500"),
		usBar(t, usOpen.Add(time.Minute), "231.4600")))

	result, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if err != nil {
		t.Fatalf("PollClosedBars: %v", err)
	}
	if result.Observed != 3 || result.Admitted != 2 || result.Unchanged != 0 || result.Corrections != 0 {
		t.Fatalf("counts = %+v", result)
	}
	if result.SessionID != "US:2026-08-14" || result.CalendarVersion == "" || !result.Terminal {
		t.Fatalf("scope = %+v", result)
	}
	if !result.NewestObserved.Equal(usOpen.Add(3 * time.Minute)) {
		t.Fatalf("newest observed = %v", result.NewestObserved)
	}
	series := usSeries(t, store, pollAt, pollAt)
	if len(series.Bars) != 2 {
		t.Fatalf("stored bars = %d, want 2", len(series.Bars))
	}
	wantSuccessor := map[uint64]uint64{
		uint64(usOpen.Add(time.Minute).UnixMilli()):     uint64(usOpen.Add(2 * time.Minute).UnixMilli()),
		uint64(usOpen.Add(2 * time.Minute).UnixMilli()): uint64(usOpen.Add(3 * time.Minute).UnixMilli()),
	}
	for _, bar := range series.Bars {
		if bar.Payload.SuccessorOpenAtMS != wantSuccessor[bar.Payload.OpenAtMS] {
			t.Fatalf("bar %d successor = %d, want %d", bar.Payload.OpenAtMS,
				bar.Payload.SuccessorOpenAtMS, wantSuccessor[bar.Payload.OpenAtMS])
		}
		if bar.Payload.Revision != 1 || !bar.Payload.RegularSession || bar.Payload.Finality != strategyevidence.BarFinalitySuccessorObserved {
			t.Fatalf("bar payload = %+v", bar.Payload)
		}
	}
	if _, exists := wantSuccessor[uint64(usOpen.Add(3*time.Minute).UnixMilli())]; exists {
		t.Fatal("test fixture is wrong")
	}
}

func TestPollAdmitsThePreviousNewestBarOnceASuccessorExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	calendar := usCalendarAt(t, usOpen.Add(-time.Hour))
	firstPollAt := usOpen.Add(10 * time.Minute)
	clock := marketclock.NewFake(firstPollAt)
	store := openTestStore(t, clock)

	first := readerOf(usPage(t, firstPollAt, "page-1", "",
		usBar(t, usOpen.Add(2*time.Minute), "231.4500"),
		usBar(t, usOpen.Add(time.Minute), "231.4600")))
	if _, err := PollClosedBars(ctx, first, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: firstPollAt,
	}); err != nil {
		t.Fatalf("first poll: %v", err)
	}

	secondPollAt := firstPollAt.Add(time.Minute)
	clock.Advance(time.Minute)
	second := readerOf(usPage(t, secondPollAt, "page-2", "",
		usBar(t, usOpen.Add(3*time.Minute), "231.4300"),
		usBar(t, usOpen.Add(2*time.Minute), "231.4500"),
		usBar(t, usOpen.Add(time.Minute), "231.4600")))
	result, err := PollClosedBars(ctx, second, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: secondPollAt,
	})
	if err != nil {
		t.Fatalf("second poll: %v", err)
	}
	if result.Admitted != 1 || result.Unchanged != 1 || result.Corrections != 0 || result.Conflicts != 0 {
		t.Fatalf("second poll counts = %+v", result)
	}
	series := usSeries(t, store, secondPollAt, secondPollAt)
	if len(series.Bars) != 2 {
		t.Fatalf("stored bars = %d, want 2", len(series.Bars))
	}
	if series.Bars[1].Payload.OpenAtMS != uint64(usOpen.Add(2*time.Minute).UnixMilli()) {
		t.Fatalf("newest stored bar = %d", series.Bars[1].Payload.OpenAtMS)
	}
}

// ---- 정규장 창 밖의 봉은 관측하되 저장하지 않는다 ----

func TestPollStoresOnlyRegularSessionBarsButUsesTheOthersAsSuccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usClose.Add(5 * time.Minute)
	clock := marketclock.NewFake(pollAt)
	store := openTestStore(t, clock)
	reader := readerOf(usPage(t, pollAt, "page-1", "",
		usBar(t, usClose.Add(time.Minute), "231.4400"),  // 16:01 ET, after hours
		usBar(t, usClose, "231.4500"),                   // 16:00 ET, close is exclusive
		usBar(t, usClose.Add(-time.Minute), "231.4600"), // 15:59 ET, last regular bar
		usBar(t, usOpen, "231.4700"),                    // 09:30 ET, first regular bar
		usBar(t, usOpen.Add(-time.Minute), "231.4800"))) // 09:29 ET, pre-session

	result, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usClose.Add(-time.Hour)), PollAt: pollAt,
	})
	if err != nil {
		t.Fatalf("PollClosedBars: %v", err)
	}
	if result.Observed != 5 || result.Admitted != 2 {
		t.Fatalf("counts = %+v", result)
	}
	series := usSeries(t, store, pollAt, pollAt)
	if len(series.Bars) != 2 {
		t.Fatalf("stored bars = %d", len(series.Bars))
	}
	if series.Bars[0].Payload.OpenAtMS != uint64(usOpen.UnixMilli()) ||
		series.Bars[1].Payload.OpenAtMS != uint64(usClose.Add(-time.Minute).UnixMilli()) {
		t.Fatalf("stored minutes = %+v", series.Bars)
	}
	// 15:59 봉의 successor는 정규장 밖의 16:00 봉이다.
	if series.Bars[1].Payload.SuccessorOpenAtMS != uint64(usClose.UnixMilli()) {
		t.Fatalf("15:59 successor = %d, want the 16:00 after-hours bar", series.Bars[1].Payload.SuccessorOpenAtMS)
	}
	if len(result.Gaps) != 1 || result.Gaps[0].Minutes != 388 ||
		!result.Gaps[0].From.Equal(usOpen.Add(time.Minute)) || !result.Gaps[0].To.Equal(usClose.Add(-2*time.Minute)) {
		t.Fatalf("gaps = %+v", result.Gaps)
	}
}

// ---- 달력 결합 ----

func TestPollLabelsOvernightUSBarsWithTheirEasternDate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// 01:00 KST 2026-08-15 == 16:00 UTC 2026-08-14 == 12:00 ET 2026-08-14.
	pollAt := time.Date(2026, 8, 14, 16, 0, 0, 0, time.UTC)
	clock := marketclock.NewFake(pollAt)
	store := openTestStore(t, clock)
	calendar := usCalendarAt(t, usOpen.Add(-30*time.Minute))
	reader := readerOf(usPage(t, pollAt, "page-1", "",
		usBar(t, pollAt.Add(-2*time.Minute), "231.4400"),
		usBar(t, pollAt.Add(-3*time.Minute), "231.4500")))

	result, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: pollAt,
	})
	if err != nil {
		t.Fatalf("PollClosedBars: %v", err)
	}
	if result.SessionID != "US:2026-08-14" || result.CalendarVersion != calendar.Version {
		t.Fatalf("session = %q version = %q", result.SessionID, result.CalendarVersion)
	}
	if len(reader.calls) != 1 || reader.calls[0].before != "2026-08-15T01:00:00.000+09:00" {
		t.Fatalf("page-1 before = %+v", reader.calls)
	}
	if reader.calls[0].count != 200 || reader.calls[0].market != "US" || reader.calls[0].symbol != usSymbol {
		t.Fatalf("page-1 request = %+v", reader.calls[0])
	}
}

func TestPollUsesTheKoreanCalendarAndSessionPrefix(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := krOpen.Add(10 * time.Minute)
	clock := marketclock.NewFake(pollAt)
	store := openTestStore(t, clock)
	reader := readerOf(pageOf("KR", krSymbol, pollAt, "page-1", "",
		krBar(t, krOpen.Add(2*time.Minute), "71050"),
		krBar(t, krOpen.Add(time.Minute), "71000")))

	result, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketKR, Symbol: krSymbol,
		Calendar: krCalendarAt(t, krOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if err != nil {
		t.Fatalf("PollClosedBars: %v", err)
	}
	if result.SessionID != "KRX:2026-08-14" || result.Admitted != 1 {
		t.Fatalf("result = %+v", result)
	}
	if reader.calls[0].before != "2026-08-14T09:10:00.000+09:00" {
		t.Fatalf("page-1 before = %q", reader.calls[0].before)
	}
}

// ---- 입력 거절 ----

func TestPollRefusesInvalidInputBeforeAnyRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	valid := usCalendarAt(t, usOpen.Add(-time.Hour))
	noRegular := usCalendarAt(t, usOpen.Add(-time.Hour))
	noRegular.Today.Regular = nil
	skewed := usCalendarAt(t, pollAt.Add(time.Hour))

	for _, testCase := range []struct {
		name  string
		input BarPollInput
		want  string
	}{
		{"unknown market", BarPollInput{Market: "jp", Symbol: usSymbol, Calendar: valid, PollAt: pollAt}, RefusalMarketInvalid},
		{"empty symbol", BarPollInput{Market: marketclock.MarketUS, Symbol: "", Calendar: valid, PollAt: pollAt}, RefusalSymbolInvalid},
		{"symbol with a space", BarPollInput{Market: marketclock.MarketUS, Symbol: "AA PL", Calendar: valid, PollAt: pollAt}, RefusalSymbolInvalid},
		{"calendar market mismatch", BarPollInput{Market: marketclock.MarketKR, Symbol: krSymbol, Calendar: valid, PollAt: pollAt}, RefusalCalendarMarket},
		{"calendar fetched after the poll", BarPollInput{Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: skewed, PollAt: pollAt}, RefusalCalendarInvalid},
		{"poll instant on another trading day", BarPollInput{Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: valid, PollAt: pollAt.Add(48 * time.Hour)}, RefusalCalendarDay},
		{"no regular session", BarPollInput{Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: noRegular, PollAt: pollAt}, RefusalNoRegularSession},
		{"poll instant missing", BarPollInput{Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: valid}, RefusalPollAtInvalid},
		{"max pages above the bound", BarPollInput{Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: valid, PollAt: pollAt, MaxPages: 9}, RefusalMaxPagesInvalid},
		{"max pages negative", BarPollInput{Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: valid, PollAt: pollAt, MaxPages: -1}, RefusalMaxPagesInvalid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			store := openTestStore(t, marketclock.NewFake(pollAt))
			reader := readerOf(usPage(t, pollAt, "unused", ""))
			_, err := PollClosedBars(ctx, reader, store, testCase.input)
			if err == nil {
				t.Fatalf("accepted %+v", testCase.input)
			}
			if got := refusalOf(t, err); got.Reason != testCase.want {
				t.Fatalf("reason = %q, want %q", got.Reason, testCase.want)
			}
			if len(reader.calls) != 0 {
				t.Fatalf("a refused poll still called the reader: %+v", reader.calls)
			}
		})
	}
}

// ---- 여러 페이지 기어가기 ----

// 이 시험은 실측된 페이지 모양을 쓴다: 커서는 그 페이지의 마지막 봉보다 1분 이르고,
// 다음 페이지는 그 커서에서(또는 그보다 아래에서) 시작한다. 겹치는 봉은 없다.
func TestPollCrawlsPagesInTheMeasuredShape(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	clock := marketclock.NewFake(pollAt)
	store := openTestStore(t, clock)
	cursor := cursorBefore(t, usOpen.Add(3*time.Minute))
	reader := readerOf(
		usPage(t, pollAt, "page-1", cursor,
			usBar(t, usOpen.Add(5*time.Minute), "231.4200"),
			usBar(t, usOpen.Add(4*time.Minute), "231.4300"),
			usBar(t, usOpen.Add(3*time.Minute), "231.4400")),
		usPage(t, pollAt, "page-2", "",
			usBar(t, usOpen.Add(2*time.Minute), "231.4500"),
			usBar(t, usOpen.Add(time.Minute), "231.4600")))

	result, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if err != nil {
		t.Fatalf("PollClosedBars: %v", err)
	}
	if result.Pages != 2 || result.Observed != 5 || result.Admitted != 4 {
		t.Fatalf("result = %+v", result)
	}
	if len(reader.calls) != 2 || reader.calls[1].before != cursor {
		t.Fatalf("page-2 request = %+v", reader.calls)
	}
	series := usSeries(t, store, pollAt, pollAt)
	if len(series.Bars) != 4 {
		t.Fatalf("stored bars = %d, want 4", len(series.Bars))
	}
}

// 인터페이스 가드 시험. 진짜 official.Client를 통해서는 도달할 수 없는 모양이다
// (리더가 커서 < 가장 오래된 봉, 모든 봉 <= before를 강제한다). 다른 CandleReader
// 구현이 들어왔을 때를 위해 생산자 쪽에 남겨 둔 방어다.
func TestPollInterfaceGuardRefusesAnOverlapThatIsNotByteIdentical(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	clock := marketclock.NewFake(pollAt)
	store := openTestStore(t, clock)
	cursor := brokerTimestamp(t, usOpen.Add(3*time.Minute))
	reader := readerOf(
		usPage(t, pollAt, "page-1", cursor,
			usBar(t, usOpen.Add(5*time.Minute), "231.4200"),
			usBar(t, usOpen.Add(4*time.Minute), "231.4300"),
			usBar(t, usOpen.Add(3*time.Minute), "231.4400")),
		usPage(t, pollAt, "page-2", "",
			usBar(t, usOpen.Add(3*time.Minute), "231.4900"), // 같은 분, 다른 바이트
			usBar(t, usOpen.Add(2*time.Minute), "231.4500")))

	_, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if err == nil {
		t.Fatal("accepted a page whose overlapping bar changed under it")
	}
	if got := refusalOf(t, err); got.Reason != RefusalOverlapMismatch {
		t.Fatalf("reason = %q", got.Reason)
	}
	if series := usSeries(t, store, pollAt, pollAt); len(series.Bars) != 0 {
		t.Fatalf("a refused poll appended %d bars", len(series.Bars))
	}
}

// 인터페이스 가드 시험(위와 같은 이유로 진짜 리더로는 도달 불가).
func TestPollInterfaceGuardRefusesAPageThatStartsNewerThanThePreviousPageEnded(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	store := openTestStore(t, marketclock.NewFake(pollAt))
	cursor := brokerTimestamp(t, usOpen.Add(3*time.Minute))
	reader := readerOf(
		usPage(t, pollAt, "page-1", cursor,
			usBar(t, usOpen.Add(5*time.Minute), "231.4200"),
			usBar(t, usOpen.Add(4*time.Minute), "231.4300")),
		usPage(t, pollAt, "page-2", "",
			usBar(t, usOpen.Add(6*time.Minute), "231.4100"),
			usBar(t, usOpen.Add(2*time.Minute), "231.4500")))

	_, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if err == nil {
		t.Fatal("accepted a page that walked forward in time")
	}
	if got := refusalOf(t, err); got.Reason != RefusalOverlapMismatch {
		t.Fatalf("reason = %q", got.Reason)
	}
}

func TestPollRefusesACursorThatDoesNotMoveBackwards(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	store := openTestStore(t, marketclock.NewFake(pollAt))
	cursor := cursorBefore(t, usOpen.Add(4*time.Minute))
	reader := readerOf(
		usPage(t, pollAt, "page-1", cursor,
			usBar(t, usOpen.Add(5*time.Minute), "231.4200"),
			usBar(t, usOpen.Add(4*time.Minute), "231.4300")),
		usPage(t, pollAt, "page-2", cursor, // 같은 커서를 또 준다
			usBar(t, usOpen.Add(3*time.Minute), "231.4400"),
			usBar(t, usOpen.Add(2*time.Minute), "231.4500")))

	_, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if err == nil {
		t.Fatal("accepted a cursor that did not move backwards")
	}
	if got := refusalOf(t, err); got.Reason != RefusalCursorLoop {
		t.Fatalf("reason = %q", got.Reason)
	}
}

func TestPollStopConditions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	calendar := usCalendarAt(t, usOpen.Add(-time.Hour))
	cursorAt := func(offset time.Duration) string { return cursorBefore(t, usOpen.Add(offset)) }

	t.Run("empty page", func(t *testing.T) {
		t.Parallel()
		store := openTestStore(t, marketclock.NewFake(pollAt))
		reader := readerOf(
			usPage(t, pollAt, "page-1", cursorAt(4*time.Minute),
				usBar(t, usOpen.Add(5*time.Minute), "231.4200"),
				usBar(t, usOpen.Add(4*time.Minute), "231.4300")),
			usPage(t, pollAt, "page-2", cursorAt(2*time.Minute)))
		result, err := PollClosedBars(ctx, reader, store, BarPollInput{
			Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: pollAt,
		})
		if err != nil {
			t.Fatalf("PollClosedBars: %v", err)
		}
		if result.Pages != 2 || result.Terminal || result.Truncated || result.RateGated || result.ReachedLowerBound {
			t.Fatalf("result = %+v", result)
		}
	})

	t.Run("max pages", func(t *testing.T) {
		t.Parallel()
		store := openTestStore(t, marketclock.NewFake(pollAt))
		reader := readerOf(
			usPage(t, pollAt, "page-1", cursorAt(5*time.Minute),
				usBar(t, usOpen.Add(6*time.Minute), "231.4200"),
				usBar(t, usOpen.Add(5*time.Minute), "231.4300")),
			usPage(t, pollAt, "page-2", cursorAt(3*time.Minute),
				usBar(t, usOpen.Add(4*time.Minute), "231.4400"),
				usBar(t, usOpen.Add(3*time.Minute), "231.4500")))
		result, err := PollClosedBars(ctx, reader, store, BarPollInput{
			Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: pollAt, MaxPages: 2,
		})
		if err != nil {
			t.Fatalf("PollClosedBars: %v", err)
		}
		if result.Pages != 2 || !result.Truncated || len(reader.calls) != 2 {
			t.Fatalf("result = %+v calls = %d", result, len(reader.calls))
		}
	})

	t.Run("lower bound", func(t *testing.T) {
		t.Parallel()
		store := openTestStore(t, marketclock.NewFake(pollAt))
		reader := readerOf(usPage(t, pollAt, "page-1", cursorAt(-time.Minute),
			usBar(t, usOpen.Add(time.Minute), "231.4200"),
			usBar(t, usOpen.Add(-time.Minute), "231.4300")))
		result, err := PollClosedBars(ctx, reader, store, BarPollInput{
			Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: pollAt,
		})
		if err != nil {
			t.Fatalf("PollClosedBars: %v", err)
		}
		if !result.ReachedLowerBound || result.Pages != 1 || len(reader.calls) != 1 {
			t.Fatalf("result = %+v", result)
		}
	})

	t.Run("rate gated", func(t *testing.T) {
		t.Parallel()
		store := openTestStore(t, marketclock.NewFake(pollAt))
		page := usPage(t, pollAt, "page-1", cursorAt(4*time.Minute),
			usBar(t, usOpen.Add(5*time.Minute), "231.4200"),
			usBar(t, usOpen.Add(4*time.Minute), "231.4300"))
		page.Budget = official.RateBudget{Path: "/api/v1/candles", Limit: 20, Remaining: 0, Reported: true}
		reader := readerOf(page, usPage(t, pollAt, "page-2", "", usBar(t, usOpen.Add(3*time.Minute), "231.4400")))
		result, err := PollClosedBars(ctx, reader, store, BarPollInput{
			Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: pollAt,
		})
		if err != nil {
			t.Fatalf("PollClosedBars: %v", err)
		}
		if !result.RateGated || result.Pages != 1 || len(reader.calls) != 1 {
			t.Fatalf("result = %+v calls = %d", result, len(reader.calls))
		}
	})
}

// ---- 되풀이 폴, 정정, 충돌 ----

func TestPollRepolledWithIdenticalContentAppendsNothing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	calendar := usCalendarAt(t, usOpen.Add(-time.Hour))
	pollAt := usOpen.Add(10 * time.Minute)
	// 저장소 시계는 폴 시각보다 뒤에 있다. 운영에서 늘 그렇다(적재는 폴 뒤에 일어난다).
	clock := marketclock.NewFake(pollAt.Add(time.Second))
	store := openTestStore(t, clock)
	bars := []official.RawMinuteCandle{
		usBar(t, usOpen.Add(3*time.Minute), "231.4400"),
		usBar(t, usOpen.Add(2*time.Minute), "231.4500"),
		usBar(t, usOpen.Add(time.Minute), "231.4600"),
	}
	if _, err := PollClosedBars(ctx, readerOf(usPage(t, pollAt, "page-1", "", bars...)), store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: pollAt,
	}); err != nil {
		t.Fatalf("first poll: %v", err)
	}
	first := usSeries(t, store, pollAt, pollAt.Add(2*time.Second))

	secondPollAt := pollAt.Add(time.Minute)
	clock.Advance(time.Minute)
	result, err := PollClosedBars(ctx, readerOf(usPage(t, secondPollAt, "page-2", "", bars...)), store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: secondPollAt,
	})
	if err != nil {
		t.Fatalf("second poll: %v", err)
	}
	if result.Unchanged != 2 || result.Admitted != 0 || result.Conflicts != 0 || result.Corrections != 0 {
		t.Fatalf("second poll counts = %+v", result)
	}
	if again := usSeries(t, store, pollAt, pollAt.Add(2*time.Second)); again.Digest != first.Digest || len(again.Bars) != 2 {
		t.Fatalf("re-poll changed the sealed series: %q -> %q", first.Digest, again.Digest)
	}
}

func TestPollWritesACorrectionAsARevisionAndLeavesTheEarlierReplayAlone(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	calendar := usCalendarAt(t, usOpen.Add(-time.Hour))
	pollAt := usOpen.Add(10 * time.Minute)
	clock := marketclock.NewFake(pollAt.Add(time.Second))
	store := openTestStore(t, clock)
	if _, err := PollClosedBars(ctx, readerOf(usPage(t, pollAt, "page-1", "",
		usBar(t, usOpen.Add(3*time.Minute), "231.4400"),
		usBar(t, usOpen.Add(2*time.Minute), "231.4500"),
		usBar(t, usOpen.Add(time.Minute), "231.4600"))), store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: pollAt,
	}); err != nil {
		t.Fatalf("first poll: %v", err)
	}
	before := usSeries(t, store, pollAt, pollAt.Add(2*time.Second))

	secondPollAt := pollAt.Add(time.Minute)
	clock.Advance(time.Minute)
	recorder := &recordingStore{inner: store}
	result, err := PollClosedBars(ctx, readerOf(usPage(t, secondPollAt, "page-2", "",
		usBar(t, usOpen.Add(3*time.Minute), "231.4400"),
		usBar(t, usOpen.Add(2*time.Minute), "231.4900"), // 정정된 종가
		usBar(t, usOpen.Add(time.Minute), "231.4600"))), recorder, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: secondPollAt,
	})
	if err != nil {
		t.Fatalf("second poll: %v", err)
	}
	if result.Corrections != 1 || result.Admitted != 1 || result.Unchanged != 1 {
		t.Fatalf("correction counts = %+v", result)
	}
	after := usSeries(t, store, secondPollAt, secondPollAt.Add(2*time.Second))
	corrected := after.Bars[1]
	if corrected.Payload.Revision != 2 || corrected.RevisionIdentity != "r2" || corrected.Payload.Raw.Close != "231.4900" {
		t.Fatalf("corrected bar = %+v", corrected)
	}
	// 정정본은 자기가 무엇을 대신하는지 머리말에 적어야 한다. 그래야 사슬이 끊기지 않는다.
	if len(recorder.appended) != 1 || recorder.appended[0].Header().SupersedesRevisionIdentity != "r1" {
		t.Fatalf("supersedes = %+v", recorder.appended)
	}
	if replay := usSeries(t, store, pollAt, pollAt.Add(2*time.Second)); replay.Digest != before.Digest {
		t.Fatalf("the earlier ingestion cutoff no longer replays: %q -> %q", before.Digest, replay.Digest)
	}
}

func TestPollCountsAConflictWhenAForeignWriterHoldsTheSameRevision(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	calendar := usCalendarAt(t, usOpen.Add(-time.Hour))
	pollAt := usOpen.Add(10 * time.Minute)
	// 이물질 줄은 우리가 이미 읽은 *뒤에* 나타난다. 그것이 충돌이 실제로 생기는 유일한
	// 방식이다(같은 판 번호, 다른 지문). 중복 제거 읽기는 모든 줄을 보므로, 미리 넣어 두면
	// 충돌이 아니라 정정이 된다.
	clock := marketclock.NewFake(pollAt.Add(10 * time.Minute))
	store := openTestStore(t, clock)
	foreign, err := strategyevidence.NewClosedBar1mEnvelope(strategyevidence.ClosedBar1mInput{
		Market: marketclock.MarketUS, Symbol: usSymbol, SessionID: "US:2026-08-14",
		CalendarVersion: calendar.Version, OpenAt: usOpen.Add(2 * time.Minute), Revision: 1,
		ObservedAt: pollAt, Currency: "USD",
		Raw: strategyevidence.RawClosedBar1m{Open: "231.4350", High: "231.5000", Low: "231.4000",
			Close: "231.4999", Volume: "1234"},
		SuccessorOpenAt: usOpen.Add(3 * time.Minute), RegularSession: true,
		SourceResponseDigest: digestFor("foreign"),
	})
	if err != nil {
		t.Fatalf("building the foreign row: %v", err)
	}
	injected := &injectingStore{inner: store, inject: func() {
		if _, err := store.Append(ctx, foreign); err != nil {
			t.Errorf("appending the foreign row: %v", err)
		}
	}}

	result, err := PollClosedBars(ctx, readerOf(usPage(t, pollAt, "page-1", "",
		usBar(t, usOpen.Add(3*time.Minute), "231.4400"),
		usBar(t, usOpen.Add(2*time.Minute), "231.4500"),
		usBar(t, usOpen.Add(time.Minute), "231.4600"))), injected, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: pollAt,
	})
	if err != nil {
		t.Fatalf("PollClosedBars: %v", err)
	}
	if result.Conflicts != 1 || result.Admitted != 1 {
		t.Fatalf("result = %+v", result)
	}
}

func TestPollRecordsAConstructorRefusalAndKeepsTheOtherBars(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	store := openTestStore(t, marketclock.NewFake(pollAt))
	bad := usBar(t, usOpen.Add(2*time.Minute), "231.4500")
	bad.Open = "00231.4350" // 앞자리 0은 evidence 계약이 거절한다
	reader := readerOf(usPage(t, pollAt, "page-1", "",
		usBar(t, usOpen.Add(3*time.Minute), "231.4400"),
		bad,
		usBar(t, usOpen.Add(time.Minute), "231.4600")))

	result, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if err != nil {
		t.Fatalf("PollClosedBars: %v", err)
	}
	if len(result.Refused) != 1 || !result.Refused[0].OpenAt.Equal(usOpen.Add(2*time.Minute)) {
		t.Fatalf("refused = %+v", result.Refused)
	}
	if result.Admitted != 1 {
		t.Fatalf("the other bar was not admitted: %+v", result)
	}
}

func TestPollReturnsStoreErrorWithTheCountsSoFar(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	inner := openTestStore(t, marketclock.NewFake(pollAt))
	boom := errors.New("disk is on fire")
	store := &flakyStore{inner: inner, failAt: 2, err: boom}
	reader := readerOf(usPage(t, pollAt, "page-1", "",
		usBar(t, usOpen.Add(3*time.Minute), "231.4400"),
		usBar(t, usOpen.Add(2*time.Minute), "231.4500"),
		usBar(t, usOpen.Add(time.Minute), "231.4600")))

	result, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if err == nil {
		t.Fatal("a store failure was swallowed")
	}
	if got := refusalOf(t, err); got.Reason != RefusalStoreError {
		t.Fatalf("reason = %q", got.Reason)
	}
	if !errors.Is(err, boom) {
		t.Fatalf("the underlying store error was lost: %v", err)
	}
	if result.Admitted != 1 || result.Pages != 1 {
		t.Fatalf("counts so far were not reported: %+v", result)
	}
}

func TestPollReturnsAReaderErrorWithoutAppending(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	store := openTestStore(t, marketclock.NewFake(pollAt))
	reader := &fakeReader{replies: []readerReply{{err: official.ErrRateLimited}}}

	result, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if got := refusalOf(t, err); got.Reason != RefusalReaderError {
		t.Fatalf("reason = %q", got.Reason)
	}
	if !errors.Is(err, official.ErrRateLimited) {
		t.Fatalf("the transport classification was lost: %v", err)
	}
	if result.Admitted != 0 {
		t.Fatalf("a failed read still appended: %+v", result)
	}
}

func TestPollIsDeterministic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	calendar := usCalendarAt(t, usOpen.Add(-time.Hour))
	bars := []official.RawMinuteCandle{
		usBar(t, usOpen.Add(4*time.Minute), "231.4400"),
		usBar(t, usOpen.Add(3*time.Minute), "231.4500"),
		usBar(t, usOpen.Add(time.Minute), "231.4600"),
	}
	run := func() (BarPollResult, string) {
		store := openTestStore(t, marketclock.NewFake(pollAt))
		result, err := PollClosedBars(ctx, readerOf(usPage(t, pollAt, "page-1", "", bars...)), store, BarPollInput{
			Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: pollAt,
		})
		if err != nil {
			t.Fatalf("PollClosedBars: %v", err)
		}
		return result, usSeries(t, store, pollAt, pollAt).Digest
	}
	firstResult, firstDigest := run()
	secondResult, secondDigest := run()
	if !reflect.DeepEqual(firstResult, secondResult) {
		t.Fatalf("results differ:\n%+v\n%+v", firstResult, secondResult)
	}
	if firstDigest != secondDigest {
		t.Fatalf("series digests differ: %q vs %q", firstDigest, secondDigest)
	}
	if len(firstResult.Gaps) != 1 || firstResult.Gaps[0].Minutes != 1 {
		t.Fatalf("gaps = %+v", firstResult.Gaps)
	}
}

// ---- 결정 24: 달력 신선도는 기록을 막지 않는다 ----

func TestPollAcceptsAStaleOrLateCalendarThroughTheWholeSession(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// 정규장은 6시간 30분인데 달력 신선도 창은 6시간이다. 옛 규칙에서는 open+6h 뒤의
	// 어떤 봉도 증거가 될 수 없었다. 마감 직전의 봉이야말로 돌파 판단에 필요한 봉이다.
	for _, offset := range []time.Duration{6 * time.Hour, 6*time.Hour + 29*time.Minute} {
		t.Run(offset.String(), func(t *testing.T) {
			t.Parallel()
			pollAt := usOpen.Add(offset)
			store := openTestStore(t, marketclock.NewFake(pollAt.Add(time.Second)))
			reader := readerOf(usPage(t, pollAt, "page-1", "",
				usBar(t, pollAt.Add(-time.Minute).Truncate(time.Minute), "231.4400"),
				usBar(t, pollAt.Add(-2*time.Minute).Truncate(time.Minute), "231.4500"),
				usBar(t, pollAt.Add(-3*time.Minute).Truncate(time.Minute), "231.4600")))
			result, err := PollClosedBars(ctx, reader, store, BarPollInput{
				Market: marketclock.MarketUS, Symbol: usSymbol,
				Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
			})
			if err != nil {
				t.Fatalf("PollClosedBars at open+%s: %v", offset, err)
			}
			if result.Admitted != 2 {
				t.Fatalf("admitted = %d, want 2 (%+v)", result.Admitted, result)
			}
		})
	}
}

// ---- 결정 25: 중복 제거는 저장된 모든 줄을 본다 ----

func TestPollDeduplicatesEveryStoredRowRegardlessOfIngestionInstant(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	calendar := usCalendarAt(t, usOpen.Add(-time.Hour))
	pollAt := usOpen.Add(10 * time.Minute)
	// 저장소 시계는 폴 시각보다 뒤다. 옛 규칙(적재 컷오프 = PollAt)에서는 방금 넣은 줄이
	// 다음 폴에 보이지 않아 같은 판 번호로 다시 적재되고 충돌 격리가 쌓였다.
	store := openTestStore(t, marketclock.NewFake(pollAt.Add(time.Second)))
	bars := []official.RawMinuteCandle{
		usBar(t, usOpen.Add(3*time.Minute), "231.4400"),
		usBar(t, usOpen.Add(2*time.Minute), "231.4500"),
		usBar(t, usOpen.Add(time.Minute), "231.4600"),
	}
	if _, err := PollClosedBars(ctx, readerOf(usPage(t, pollAt, "first-read", "", bars...)), store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: pollAt,
	}); err != nil {
		t.Fatalf("first poll: %v", err)
	}
	// 같은 순간에 다시 폴한다. 응답 지문만 다르다(관측 metadata는 폴마다 달라진다).
	result, err := PollClosedBars(ctx, readerOf(usPage(t, pollAt, "second-read", "", bars...)), store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: pollAt,
	})
	if err != nil {
		t.Fatalf("duplicate poll: %v", err)
	}
	if result.Unchanged != 2 || result.Admitted != 0 || result.Conflicts != 0 {
		t.Fatalf("duplicate poll = %+v", result)
	}
}

func TestPollRetryAfterAStoreErrorDoesNotConflict(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	calendar := usCalendarAt(t, usOpen.Add(-time.Hour))
	pollAt := usOpen.Add(10 * time.Minute)
	inner := openTestStore(t, marketclock.NewFake(pollAt.Add(time.Second)))
	bars := []official.RawMinuteCandle{
		usBar(t, usOpen.Add(3*time.Minute), "231.4400"),
		usBar(t, usOpen.Add(2*time.Minute), "231.4500"),
		usBar(t, usOpen.Add(time.Minute), "231.4600"),
	}
	failing := &flakyStore{inner: inner, failAt: 2, err: errors.New("disk is on fire")}
	first, err := PollClosedBars(ctx, readerOf(usPage(t, pollAt, "attempt-1", "", bars...)), failing, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: pollAt,
	})
	if err == nil || first.Admitted != 1 {
		t.Fatalf("first attempt = %+v err = %v", first, err)
	}
	retry, err := PollClosedBars(ctx, readerOf(usPage(t, pollAt, "attempt-2", "", bars...)), inner, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: pollAt,
	})
	if err != nil {
		t.Fatalf("retry: %v", err)
	}
	if retry.Unchanged != 1 || retry.Admitted != 1 || retry.Conflicts != 0 {
		t.Fatalf("retry = %+v", retry)
	}
}

// ---- successor와 관측 시각의 출처 ----

func TestPollBindsEachSuccessorToTheNextNewerObservedBar(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(15 * time.Minute)
	// 실측 크롤에는 1분이 아닌 간격이 27번 있었다. successor는 "다음 분"이 아니라
	// "실제로 본 다음 봉"이어야 한다.
	newest, middle, oldest := usOpen.Add(10*time.Minute), usOpen.Add(5*time.Minute), usOpen.Add(time.Minute)
	recorder := &recordingStore{inner: openTestStore(t, marketclock.NewFake(pollAt.Add(time.Second)))}
	result, err := PollClosedBars(ctx, readerOf(usPage(t, pollAt, "page-1", "",
		usBar(t, newest, "231.4400"), usBar(t, middle, "231.4500"), usBar(t, oldest, "231.4600"))),
		recorder, BarPollInput{
			Market: marketclock.MarketUS, Symbol: usSymbol,
			Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
		})
	if err != nil || result.Admitted != 2 || len(recorder.appended) != 2 {
		t.Fatalf("result = %+v err = %v appended = %d", result, err, len(recorder.appended))
	}
	want := map[uint64]uint64{
		uint64(middle.UnixMilli()): uint64(newest.UnixMilli()),
		uint64(oldest.UnixMilli()): uint64(middle.UnixMilli()),
	}
	for _, envelope := range recorder.appended {
		payload := payloadOf(t, envelope)
		if payload.SuccessorOpenAtMS != want[payload.OpenAtMS] {
			t.Fatalf("bar %d successor = %d, want %d", payload.OpenAtMS, payload.SuccessorOpenAtMS, want[payload.OpenAtMS])
		}
	}
}

func TestPollTakesObservedAtFromTheResponseNotThePollInstant(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(20 * time.Minute)
	readAt := pollAt.Add(-10 * time.Minute)
	recorder := &recordingStore{inner: openTestStore(t, marketclock.NewFake(pollAt))}
	result, err := PollClosedBars(ctx, readerOf(usPage(t, readAt, "page-1", "",
		usBar(t, readAt.Add(-time.Minute), "231.4400"),
		usBar(t, readAt.Add(-2*time.Minute), "231.4500"),
		usBar(t, readAt.Add(-3*time.Minute), "231.4600"))),
		recorder, BarPollInput{
			Market: marketclock.MarketUS, Symbol: usSymbol,
			Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
		})
	if err != nil || result.Admitted != 2 {
		t.Fatalf("result = %+v err = %v", result, err)
	}
	for _, envelope := range recorder.appended {
		payload := payloadOf(t, envelope)
		if payload.SourceObservedAtMS != uint64(readAt.UnixMilli()) {
			t.Fatalf("observed at = %d, want the response instant %d (poll was %d)",
				payload.SourceObservedAtMS, readAt.UnixMilli(), pollAt.UnixMilli())
		}
	}
}

// ---- 멈춤 조건의 나머지 ----

func TestPollStopsAtAnExplicitLowerBound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	store := openTestStore(t, marketclock.NewFake(pollAt.Add(time.Second)))
	reader := readerOf(
		usPage(t, pollAt, "page-1", cursorBefore(t, usOpen.Add(3*time.Minute)),
			usBar(t, usOpen.Add(5*time.Minute), "231.4200"),
			usBar(t, usOpen.Add(4*time.Minute), "231.4300"),
			usBar(t, usOpen.Add(3*time.Minute), "231.4400")),
		usPage(t, pollAt, "page-2", "", usBar(t, usOpen.Add(2*time.Minute), "231.4500")))
	result, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
		LowerBound: usOpen.Add(4 * time.Minute),
	})
	if err != nil {
		t.Fatalf("PollClosedBars: %v", err)
	}
	if !result.ReachedLowerBound || len(reader.calls) != 1 || result.Pages != 1 {
		t.Fatalf("result = %+v calls = %d", result, len(reader.calls))
	}
}

func TestPollDefaultsToFourPages(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(20 * time.Minute)
	store := openTestStore(t, marketclock.NewFake(pollAt.Add(time.Second)))
	page := func(label string, newest time.Duration, terminal bool) official.StrictMinutePage {
		cursor := ""
		if !terminal {
			cursor = cursorBefore(t, usOpen.Add(newest-time.Minute))
		}
		return usPage(t, pollAt, label, cursor,
			usBar(t, usOpen.Add(newest), "231.4400"),
			usBar(t, usOpen.Add(newest-time.Minute), "231.4500"))
	}
	reader := readerOf(
		page("page-1", 11*time.Minute, false), page("page-2", 9*time.Minute, false),
		page("page-3", 7*time.Minute, false), page("page-4", 5*time.Minute, false),
		page("page-5", 3*time.Minute, true))
	result, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if err != nil {
		t.Fatalf("PollClosedBars: %v", err)
	}
	if result.Pages != 4 || !result.Truncated || len(reader.calls) != 4 {
		t.Fatalf("the default page bound is not four: %+v calls = %d", result, len(reader.calls))
	}
}

func TestPollGapsAreClampedToTheRegularWindow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usClose.Add(5 * time.Minute)
	store := openTestStore(t, marketclock.NewFake(pollAt.Add(time.Second)))
	result, err := PollClosedBars(ctx, readerOf(usPage(t, pollAt, "page-1", "",
		usBar(t, usClose.Add(2*time.Minute), "231.4300"),
		usBar(t, usOpen.Add(3*time.Minute), "231.4400"),
		usBar(t, usOpen.Add(time.Minute), "231.4500"),
		usBar(t, usOpen.Add(-5*time.Minute), "231.4600"))), store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usClose.Add(-time.Hour)), PollAt: pollAt,
	})
	if err != nil {
		t.Fatalf("PollClosedBars: %v", err)
	}
	if len(result.Gaps) != 3 {
		t.Fatalf("gaps = %+v", result.Gaps)
	}
	if result.Gaps[0].Minutes != 386 || !result.Gaps[0].From.Equal(usOpen.Add(4*time.Minute)) ||
		!result.Gaps[0].To.Equal(usClose.Add(-time.Minute)) {
		t.Fatalf("after-hours side is not clamped to the close: %+v", result.Gaps[0])
	}
	if result.Gaps[1].Minutes != 1 || !result.Gaps[1].From.Equal(usOpen.Add(2*time.Minute)) {
		t.Fatalf("inner gap = %+v", result.Gaps[1])
	}
	if result.Gaps[2].Minutes != 1 || !result.Gaps[2].From.Equal(usOpen) {
		t.Fatalf("pre-session side is not clamped to the open: %+v", result.Gaps[2])
	}
}

func TestPollTruncatesTheSubSecondPollInstant(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10*time.Minute + 500*time.Millisecond)
	store := openTestStore(t, marketclock.NewFake(pollAt.Add(time.Second)))
	// 커서를 잘라 낸 상한과 정확히 같게 준다. 자르지 않으면 이 커서가 상한보다 이르게 보여
	// 되돌이 감지를 빠져나간다.
	reader := readerOf(usPage(t, pollAt, "page-1", brokerTimestamp(t, usOpen.Add(10*time.Minute)),
		usBar(t, usOpen.Add(5*time.Minute), "231.4400"),
		usBar(t, usOpen.Add(4*time.Minute), "231.4500")))
	_, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if len(reader.calls) != 1 || reader.calls[0].before != "2026-08-14T22:40:00.000+09:00" {
		t.Fatalf("page-1 before = %+v", reader.calls)
	}
	if err == nil {
		t.Fatal("a cursor equal to the truncated bound was accepted")
	}
	if got := refusalOf(t, err); got.Reason != RefusalCursorLoop {
		t.Fatalf("reason = %q", got.Reason)
	}
}

// ---- 페이지 신원과 오류 경로 ----

func TestPollRefusesAPageForAnotherMarketOrSymbol(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	store := openTestStore(t, marketclock.NewFake(pollAt.Add(time.Second)))
	reader := readerOf(pageOf("KR", krSymbol, pollAt, "page-1", "",
		usBar(t, usOpen.Add(3*time.Minute), "231.4400"),
		usBar(t, usOpen.Add(2*time.Minute), "231.4500")))
	_, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if err == nil {
		t.Fatal("a page for another market and symbol was adopted")
	}
	if got := refusalOf(t, err); got.Reason != RefusalPageIdentity {
		t.Fatalf("reason = %q", got.Reason)
	}
}

func TestPollKeepsThePageCountWhenTheReaderFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	store := openTestStore(t, marketclock.NewFake(pollAt.Add(time.Second)))
	reader := &fakeReader{replies: []readerReply{
		{page: usPage(t, pollAt, "page-1", cursorBefore(t, usOpen.Add(4*time.Minute)),
			usBar(t, usOpen.Add(5*time.Minute), "231.4400"),
			usBar(t, usOpen.Add(4*time.Minute), "231.4500"))},
		{err: official.ErrServer},
	}}
	result, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if got := refusalOf(t, err); got.Reason != RefusalReaderError {
		t.Fatalf("reason = %q", got.Reason)
	}
	if result.Pages != 2 || result.SessionID != "US:2026-08-14" {
		t.Fatalf("the failed poll lost its counts: %+v", result)
	}
}

func TestPollRefusesAMergedListThatIsNotStrictlyDescending(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(15 * time.Minute)
	store := openTestStore(t, marketclock.NewFake(pollAt.Add(time.Second)))
	reader := readerOf(
		usPage(t, pollAt, "page-1", cursorBefore(t, usOpen.Add(4*time.Minute)),
			usBar(t, usOpen.Add(5*time.Minute), "231.4400"),
			usBar(t, usOpen.Add(4*time.Minute), "231.4500")),
		usPage(t, pollAt, "page-2", "",
			usBar(t, usOpen.Add(3*time.Minute), "231.4600"),
			usBar(t, usOpen.Add(6*time.Minute), "231.4700"))) // 뒤로 가지 않고 앞으로 갔다
	_, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if err == nil {
		t.Fatal("an ascending merged list was accepted")
	}
	if got := refusalOf(t, err); got.Reason != RefusalPageOrder {
		t.Fatalf("reason = %q", got.Reason)
	}
}

// 인터페이스 가드 시험: 커서가 마지막 봉과 같은 분을 가리키는 리더가 들어오면
// 겹치는 봉 한 개만 버리고 나머지는 그대로 이어 붙인다.
func TestPollInterfaceGuardDropsTheEqualInstantOverlapBar(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	store := openTestStore(t, marketclock.NewFake(pollAt.Add(time.Second)))
	overlap := usBar(t, usOpen.Add(3*time.Minute), "231.4400")
	reader := readerOf(
		usPage(t, pollAt, "page-1", brokerTimestamp(t, usOpen.Add(3*time.Minute)),
			usBar(t, usOpen.Add(5*time.Minute), "231.4200"),
			usBar(t, usOpen.Add(4*time.Minute), "231.4300"), overlap),
		usPage(t, pollAt, "page-2", "", overlap,
			usBar(t, usOpen.Add(2*time.Minute), "231.4500"),
			usBar(t, usOpen.Add(time.Minute), "231.4600")))
	result, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if err != nil {
		t.Fatalf("PollClosedBars: %v", err)
	}
	if result.Observed != 5 || result.Admitted != 4 {
		t.Fatalf("result = %+v", result)
	}
}

// ---- 진짜 official.Client와 끝까지 붙여 보는 시험 ----

func candleJSON(candle official.RawMinuteCandle) string {
	return `{"timestamp":"` + candle.Timestamp + `","openPrice":"` + candle.Open +
		`","highPrice":"` + candle.High + `","lowPrice":"` + candle.Low +
		`","closePrice":"` + candle.Close + `","volume":"` + candle.Volume +
		`","currency":"` + candle.Currency + `"}`
}

// wideUSCalendar는 하루를 거의 통째로 정규장으로 삼는 달력이다. 이 시험은 벽시계로
// 돌기 때문에, 하루 중 언제 돌려도 봉이 창 안에 들어오게 하려는 것이다.
func wideUSCalendar(t *testing.T, day string, fetchedAt time.Time) scheduler.CalendarSnapshot {
	t.Helper()
	eastern := mustLocation(t, marketclock.MarketUS)
	build := func(date string) official.MarketCalendarDay {
		parsed, err := time.ParseInLocation("2006-01-02", date, eastern)
		if err != nil {
			t.Fatalf("parsing %s: %v", date, err)
		}
		year, month, dayOfMonth := parsed.Date()
		return official.MarketCalendarDay{Date: date, RegularMarket: &official.MarketCalendarSession{
			StartTime: time.Date(year, month, dayOfMonth, 0, 10, 0, 0, eastern),
			EndTime:   time.Date(year, month, dayOfMonth, 23, 50, 0, 0, eastern),
		}}
	}
	base, err := time.ParseInLocation("2006-01-02", day, eastern)
	if err != nil {
		t.Fatalf("parsing %s: %v", day, err)
	}
	snapshot, err := scheduler.AdaptOfficialCalendar(marketclock.MarketUS, official.MarketCalendarResponse{
		PreviousBusinessDay: build(base.AddDate(0, 0, -1).Format("2006-01-02")),
		Today:               build(day),
		NextBusinessDay:     build(base.AddDate(0, 0, 1).Format("2006-01-02")),
	}, fetchedAt)
	if err != nil {
		t.Fatalf("AdaptOfficialCalendar: %v", err)
	}
	return snapshot
}

func TestPollEndToEndThroughTheRealOfficialClient(t *testing.T) {
	ctx := context.Background()
	pollAt := time.Now()
	day, err := marketclock.MarketUS.TradingDay(pollAt)
	if err != nil {
		t.Fatalf("TradingDay: %v", err)
	}
	calendar := wideUSCalendar(t, day, pollAt.Add(-time.Hour))
	minute := pollAt.UTC().Truncate(time.Minute)
	instants := []time.Time{minute.Add(-2 * time.Minute), minute.Add(-3 * time.Minute), minute.Add(-4 * time.Minute)}
	if instants[2].Before(calendar.Today.Regular.Open) || !instants[0].Before(calendar.Today.Regular.Close) {
		// 창은 00:10~23:50 ET이고 봉은 지금보다 2~4분 이르다. 그래서 하루에
		// 00:00~00:14와 23:52~24:00, 합쳐서 약 22분만 건너뛴다.
		t.Skipf("wall clock %s sits inside the ~22 minutes a day next to the eastern day boundary", pollAt)
	}
	bars := []official.RawMinuteCandle{
		usBar(t, instants[0], "231.4400"), usBar(t, instants[1], "231.4500"), usBar(t, instants[2], "231.4600"),
	}
	firstBefore := pollAt.Truncate(time.Second).In(mustLocation(t, marketclock.MarketKR)).Format(beforeLayout)
	cursor := brokerTimestamp(t, instants[1].Add(-time.Minute))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			fmt.Fprint(w, `{"access_token":"token","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		switch r.URL.Query().Get("before") {
		case firstBefore:
			fmt.Fprint(w, `{"result":{"candles":[`+candleJSON(bars[0])+`,`+candleJSON(bars[1])+
				`],"nextBefore":"`+cursor+`"}}`)
		case cursor:
			fmt.Fprint(w, `{"result":{"candles":[`+candleJSON(bars[2])+`],"nextBefore":null}}`)
		default:
			t.Errorf("unexpected before %q", r.URL.Query().Get("before"))
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	client := official.New(official.Credentials{APIKey: "k", SecretKey: "s"}, t.TempDir()+"/token",
		official.WithBaseURL(server.URL), official.WithHTTPClient(server.Client()))
	// 진짜 시계를 쓰는 저장소다(Options.Clock 기본값). 적재 시각은 응답 시각 뒤에 온다.
	store, err := strategyevidence.Open(ctx, strategyevidence.Options{
		Path: filepath.Join(t.TempDir(), "breakout-bars.db"),
	})
	if err != nil {
		t.Fatalf("strategyevidence.Open: %v", err)
	}
	defer func() { _ = store.Close() }()
	recorder := &recordingStore{inner: store}

	result, err := PollClosedBars(ctx, client, recorder, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: calendar, PollAt: pollAt,
	})
	if err != nil {
		t.Fatalf("PollClosedBars: %v", err)
	}
	if result.Pages != 2 || result.Observed != 3 || result.Admitted != 2 || !result.Terminal {
		t.Fatalf("result = %+v", result)
	}
	if result.SessionID != "US:"+day {
		t.Fatalf("session = %q, want US:%s", result.SessionID, day)
	}
	for _, envelope := range recorder.appended {
		payload := payloadOf(t, envelope)
		// 관측 시각은 전송 계층이 본문을 다 읽은 순간이다. PollAt은 그보다 먼저 찍혔으므로
		// 엄격히 뒤여야 한다. 같기만 해도 "관측 시각 = 폴 시각"을 넣는 실수를 놓친다.
		if payload.SourceObservedAtMS <= uint64(pollAt.UnixMilli()) {
			t.Fatalf("observed at %d is not after the poll instant %d", payload.SourceObservedAtMS, pollAt.UnixMilli())
		}
	}
}

// 인터페이스 가드 시험: 다음 페이지가 같은 봉을 두 번 담아 온다. 겹침 규칙이 첫 벌만
// 버리므로 두 번째 벌이 남고, 이어 붙인 목록에 같은 분이 두 번 들어온다.
// 같은 분은 "뒤로 갔다"가 아니라 "제자리"이므로, 내림차순 검사가 엄격해야만 잡힌다.
func TestPollInterfaceGuardRefusesAnAdjacentDuplicateInstant(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(15 * time.Minute)
	store := openTestStore(t, marketclock.NewFake(pollAt.Add(time.Second)))
	duplicate := usBar(t, usOpen.Add(4*time.Minute), "231.4500")
	reader := readerOf(
		usPage(t, pollAt, "page-1", brokerTimestamp(t, usOpen.Add(4*time.Minute)),
			usBar(t, usOpen.Add(5*time.Minute), "231.4400"), duplicate),
		usPage(t, pollAt, "page-2", "", duplicate, duplicate))
	_, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		Calendar: usCalendarAt(t, usOpen.Add(-time.Hour)), PollAt: pollAt,
	})
	if err == nil {
		t.Fatal("a merged list carrying the same minute twice was accepted")
	}
	if got := refusalOf(t, err); got.Reason != RefusalPageOrder {
		t.Fatalf("reason = %q", got.Reason)
	}
	if series := usSeries(t, store, pollAt, pollAt.Add(2*time.Second)); len(series.Bars) != 0 {
		t.Fatalf("a refused poll appended %d bars", len(series.Bars))
	}
}

func TestPollRefusesACalendarWindowThatIsNotOnItsOwnDay(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pollAt := usOpen.Add(10 * time.Minute)
	store := openTestStore(t, marketclock.NewFake(pollAt.Add(time.Second)))
	// 손으로 만든 snapshot이다. AdaptOfficialCalendar를 거쳐 온 것은 이미 이 조건을
	// 지키므로, 이 갈래는 달력을 직접 조립한 호출자만 밟을 수 있다.
	foreign := usCalendarAt(t, usOpen.Add(-time.Hour))
	foreign.Today.Regular = &scheduler.SessionWindow{
		Open: usOpen.AddDate(0, 0, -1), Close: usClose.AddDate(0, 0, -1),
	}
	reader := readerOf(usPage(t, pollAt, "unused", ""))
	_, err := PollClosedBars(ctx, reader, store, BarPollInput{
		Market: marketclock.MarketUS, Symbol: usSymbol, Calendar: foreign, PollAt: pollAt,
	})
	if err == nil {
		t.Fatal("a regular window belonging to another day was accepted")
	}
	if got := refusalOf(t, err); got.Reason != RefusalCalendarInvalid {
		t.Fatalf("reason = %q", got.Reason)
	}
	if len(reader.calls) != 0 {
		t.Fatalf("a refused poll still called the reader: %+v", reader.calls)
	}
}
