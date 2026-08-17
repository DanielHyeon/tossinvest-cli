// Package officialbars는 공식 API에서 읽은 1분봉 페이지를 breakout 증거로 바꿔 적재한다.
//
// 이 꾸러미가 지키는 것은 딱 세 가지다.
//
//  1. 폴에서 가장 새 봉은 절대 적재하지 않는다. 봉이 닫혔다는 증거는 "그 다음 봉을
//     실제로 보았다"는 것뿐이고, 가장 새 봉에는 아직 다음 봉이 없다(a112 결정 6).
//  2. 페이지를 전부 받아 검사한 다음에야 한 줄이라도 적재한다. 중간에 계약이 깨지면
//     아무것도 적재하지 않는다.
//  3. 저장은 정규장 창 안의 봉만 한다. 창 밖의 봉은 세어서 보고하고 successor로만 쓴다.
//
// 시계는 밖에서 준다(PollAt). 이 꾸러미 안에는 time.Now도, 고루틴도, 로그도 없다.
package officialbars

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

// CandleReader는 이 생산자가 필요로 하는 읽기 능력 전부다.
// 실제로는 official.Client.StrictMinuteCandles가 들어오고, 시험에서는 가짜가 들어온다.
type CandleReader interface {
	StrictMinuteCandles(ctx context.Context, market, symbol string, count int, before string) (official.StrictMinutePage, error)
}

// BarStore는 이 생산자가 필요로 하는 저장 능력 전부다.
type BarStore interface {
	Append(context.Context, strategyevidence.Envelope) (strategyevidence.AppendResult, error)
	SealBarSeries(context.Context, strategyevidence.BarSeriesQuery) (strategyevidence.BarSeries, error)
}

// BarPollInput은 폴 한 번에 필요한 값 전부다.
type BarPollInput struct {
	Market   marketclock.Market
	Symbol   string
	Calendar scheduler.CalendarSnapshot
	// PollAt은 이 폴의 시각이다. 첫 페이지의 상한이자 두 컷오프의 기준이다.
	PollAt time.Time
	// LowerBound보다 오래된 봉을 만나면 기어가기를 멈춘다. 비어 있으면 정규장 시작.
	LowerBound time.Time
	// MaxPages는 한 폴이 읽을 페이지 수 상한이다. 0이면 4, 허용 범위는 1..8.
	MaxPages int
}

// BarRefusal은 봉 하나가 증거가 되지 못한 이유다. 나머지 봉은 그대로 적재된다.
type BarRefusal struct {
	OpenAt time.Time
	Reason string
}

// MinuteGap은 "이어진 두 관측 봉 사이"에서 비어 있는 분들 중 정규장 창 안에 드는
// 부분이다. 창 밖으로 나가는 부분은 잘라 낸다. 관측된 봉의 바깥(가장 새 봉보다 뒤,
// 가장 오래된 봉보다 앞)은 세지 않는다 — 거기는 아직 보지 않은 구간이지 빠진 구간이 아니다.
// 보고용이고, 이것 때문에 관측된 봉을 감추지는 않는다(빠진 분으로 세션을 거절하는 것은
// 읽는 쪽 규칙이다).
type MinuteGap struct {
	From    time.Time
	To      time.Time
	Minutes int
}

// BarPollResult는 폴 한 번의 결과 전부다.
type BarPollResult struct {
	SessionID       string
	CalendarVersion string

	Pages       int
	Observed    int
	Admitted    int
	Unchanged   int
	Corrections int
	Conflicts   int

	Refused        []BarRefusal
	NewestObserved time.Time

	Terminal          bool
	Truncated         bool
	RateGated         bool
	ReachedLowerBound bool

	Gaps []MinuteGap
}

// PollRefusal은 폴을 멈춘 이유다. 계약 위반과 아래 계층의 실패를 구분해서 들고 있다.
type PollRefusal struct {
	Reason string
	Detail string
	Err    error
}

func (r *PollRefusal) Error() string {
	message := "official bars poll: " + r.Reason + ": " + r.Detail
	if r.Err != nil {
		return message + ": " + r.Err.Error()
	}
	return message
}

func (r *PollRefusal) Unwrap() error { return r.Err }

// 거절 사유 이름표.
const (
	RefusalDependencyMissing = "DEPENDENCY_MISSING"
	RefusalMarketInvalid     = "MARKET_INVALID"
	RefusalSymbolInvalid     = "SYMBOL_INVALID"
	RefusalCalendarMarket    = "CALENDAR_MARKET_MISMATCH"
	RefusalCalendarInvalid   = "CALENDAR_INVALID"
	RefusalNoRegularSession  = "NO_REGULAR_SESSION"
	RefusalPollAtInvalid     = "POLL_AT_INVALID"
	RefusalMaxPagesInvalid   = "MAX_PAGES_INVALID"
	RefusalReaderError       = "READER_ERROR"
	RefusalPageInvalid       = "PAGE_INVALID"
	RefusalOverlapMismatch   = "OVERLAP_MISMATCH"
	RefusalCursorLoop        = "CURSOR_LOOP"
	RefusalStoreError        = "STORE_ERROR"
	RefusalCalendarDay       = "CALENDAR_DAY_MISMATCH"
	RefusalPageIdentity      = "PAGE_IDENTITY_MISMATCH"
	RefusalPageOrder         = "PAGE_ORDER_INVALID"
)

const (
	// 한 페이지에 200봉을 요청한다(문서 상한이자 실측값).
	pageCount = 200
	// 페이지 상한의 기본값과 허용 범위.
	defaultMaxPages = 4
	maxAllowedPages = 8
	// 첫 페이지 상한의 표기. 실측된 요청 문자열 그대로다(양쪽 시장 모두 KST).
	beforeLayout = "2006-01-02T15:04:05.000-07:00"
	// 봉 하나의 길이. 브로커 시각표를 여는 시각으로 옮길 때 쓴다.
	barInterval = time.Duration(strategyevidence.ClosedBar1mIntervalMS) * time.Millisecond
)

// knownReadHorizon은 "이미 저장된 것"을 읽을 때 컷오프를 얼마나 멀리 두는지다.
//
// 중복 제거는 이 세션에 이미 들어 있는 모든 줄을 봐야 한다. 적재 시각으로 잘라 버리면
// 방금 넣은 줄이 다음 폴에 보이지 않고, 같은 판 번호로 다시 적재되어 충돌 격리가 쌓인다.
// 충돌은 "다른 글쓴이"를 가리키는 신호이므로 자기 자신이 만들어서는 안 된다.
// 되감기(replay) 컷오프는 읽는 쪽(L3)의 일이지 생산자의 일이 아니다.
//
// 값을 크게 잡아도 읽는 양은 늘지 않는다. 질의가 시장·종목·세션·간격으로 이미 좁혀져
// 있고 MaxBars 512는 어떤 정규장 세션보다도 크기 때문이다(정규장은 최대 390분).
// 실측 비용은 390봉 한 세션에 봉마다 판 하나면 약 47밀리초, 판이 넷이면 약 159밀리초다.
// 폴마다 한 번만 부르므로 감당할 수 있는 값이고, 1분마다 도는 뜨거운 경로에서는
// 절대 부르지 않는다.
const knownReadHorizon = 366 * 24 * time.Hour

func refuse(reason, detail string, cause error) error {
	return &PollRefusal{Reason: reason, Detail: detail, Err: cause}
}

// observedBar는 한 봉과, 그 봉을 어느 응답에서 보았는지를 함께 들고 다닌다.
// openAt은 이미 변환을 마친 **여는 시각**이다(브로커가 보낸 닫는 시각이 아니다).
type observedBar struct {
	candle official.RawMinuteCandle
	openAt time.Time
	readAt time.Time
	digest string
}

// PollClosedBars는 한 종목의 닫힌 1분봉을 한 번 긁어서 증거로 적재한다.
func PollClosedBars(ctx context.Context, reader CandleReader, store BarStore, in BarPollInput) (BarPollResult, error) {
	var result BarPollResult
	if reader == nil || store == nil {
		return result, refuse(RefusalDependencyMissing, "a reader and a store are required", nil)
	}
	market, err := marketclock.ParseMarket(string(in.Market))
	if err != nil {
		return result, refuse(RefusalMarketInvalid, "market "+strconv.Quote(string(in.Market))+" is not kr or us", nil)
	}
	code := marketCode(market)
	if err := checkSymbol(in.Symbol); err != nil {
		return result, err
	}
	if in.Calendar.Market != market {
		return result, refuse(RefusalCalendarMarket,
			"calendar is for market "+strconv.Quote(string(in.Calendar.Market))+", not "+string(market), nil)
	}
	if in.PollAt.IsZero() {
		return result, refuse(RefusalPollAtInvalid, "the poll instant is required", nil)
	}
	// ValidityAt은 "지금 이 달력을 믿고 매매해도 되는가"를 묻는 관문이고, 그 신선도 창은
	// 6시간인데 정규장은 6시간 30분이다. 기록에까지 그 관문을 쓰면 장 마지막 30분의 봉은
	// 어떤 달력으로도 증거가 될 수 없다 — 돌파 판단에 가장 필요한 봉들이다.
	// 그래서 여기서는 "시계가 거꾸로 간 달력"만 거절하고, 대신 폴 시각이 정말로 이 달력이
	// 말하는 하루에 속하는지를 직접 확인한다. 어느 달력을 썼는지는 모든 payload의
	// calendar_version에 남으므로 읽는 쪽이 언제든 되짚을 수 있다.
	if validity := in.Calendar.ValidityAt(in.PollAt); validity == scheduler.CalendarClockSkew {
		return result, refuse(RefusalCalendarInvalid, "calendar validity is "+string(validity), nil)
	}
	if in.Calendar.Today.Regular == nil {
		return result, refuse(RefusalNoRegularSession, "today carries no regular session", nil)
	}
	// 정규장 창의 양 끝은 그 달력이 말하는 하루 위에 있어야 한다.
	// AdaptOfficialCalendar를 거쳐 온 snapshot은 이미 이 조건을 지키므로, 이 갈래는
	// 달력을 손으로 조립한 호출자만 밟는다 — 그래도 값싸게 닫아 둔다.
	for _, edge := range []struct {
		name    string
		instant time.Time
	}{{"open", in.Calendar.Today.Regular.Open}, {"close", in.Calendar.Today.Regular.Close}} {
		edgeDay, err := market.TradingDay(edge.instant)
		if err != nil || edgeDay != in.Calendar.Today.Date {
			return result, refuse(RefusalCalendarInvalid,
				"the regular session "+edge.name+" falls on market-local day "+strconv.Quote(edgeDay)+
					", not on the calendar's today "+strconv.Quote(in.Calendar.Today.Date), err)
		}
	}
	pollDay, err := market.TradingDay(in.PollAt)
	if err != nil {
		return result, refuse(RefusalCalendarDay, "reading the market-local day of the poll instant", err)
	}
	if pollDay != in.Calendar.Today.Date {
		return result, refuse(RefusalCalendarDay,
			"the poll instant falls on market-local day "+pollDay+
				", not on the calendar's today "+in.Calendar.Today.Date, nil)
	}
	if in.MaxPages < 0 || in.MaxPages > maxAllowedPages {
		return result, refuse(RefusalMaxPagesInvalid,
			"max pages must be 0 (the default of "+strconv.Itoa(defaultMaxPages)+
				") or between 1 and "+strconv.Itoa(maxAllowedPages), nil)
	}
	maxPages := in.MaxPages
	if maxPages == 0 {
		maxPages = defaultMaxPages
	}
	seoul, err := marketclock.MarketKR.Location()
	if err != nil {
		// 도달 불가 방어: time/tzdata가 바이너리에 박혀 있으므로 이 조회는 실패하지 않는다.
		return result, refuse(RefusalDependencyMissing, "the Asia/Seoul zone is unavailable", err)
	}

	sessionID := sessionCalendar(code) + ":" + in.Calendar.Today.Date
	regularOpen := in.Calendar.Today.Regular.Open
	regularClose := in.Calendar.Today.Regular.Close
	lowerBound := in.LowerBound
	if lowerBound.IsZero() {
		lowerBound = regularOpen
	}
	result.SessionID = sessionID
	result.CalendarVersion = in.Calendar.Version

	// ---- (c) 페이지를 전부 받아 검사한다. 여기서는 아직 아무것도 적재하지 않는다. ----
	before := in.PollAt.Truncate(time.Second).In(seoul).Format(beforeLayout)
	beforeInstant := in.PollAt.Truncate(time.Second)
	var bars []observedBar
	var previous []observedBar
	for {
		// 페이지 수는 요청을 보내기 전에 센다. 실패한 시도도 한 페이지 값의 쿼터를 썼고,
		// 오류를 받은 호출자도 "어디까지 갔는지"를 알아야 한다.
		result.Pages++
		page, err := reader.StrictMinuteCandles(ctx, code, in.Symbol, pageCount, before)
		if err != nil {
			return result, refuse(RefusalReaderError, "reading page "+strconv.Itoa(result.Pages), err)
		}
		if page.Market != code || page.Symbol != in.Symbol {
			return result, refuse(RefusalPageIdentity,
				"page carries "+strconv.Quote(page.Market+"/"+page.Symbol)+
					", not "+strconv.Quote(code+"/"+in.Symbol), nil)
		}
		current, err := adoptPage(page)
		if err != nil {
			return result, err
		}
		admitted := current
		if len(previous) > 0 && len(current) > 0 {
			previousLast := previous[len(previous)-1]
			if err := checkOverlap(previousLast, current[0]); err != nil {
				return result, err
			}
			// 겹치는 봉은 정확히 한 개, 그것도 같은 분일 때만 버린다. 조용히 중복을
			// 걸러 내면 "왜 사라졌는지" 모르는 채 지나가므로, 나머지 어긋남은 아래에서
			// 통째로 거절한다.
			if current[0].openAt.Equal(previousLast.openAt) {
				admitted = current[1:]
			}
		}
		bars = append(bars, admitted...)
		if page.Terminal {
			result.Terminal = true
			break
		}
		if len(current) == 0 {
			break
		}
		if current[len(current)-1].openAt.Before(lowerBound) {
			result.ReachedLowerBound = true
			break
		}
		cursor, err := time.Parse(time.RFC3339, page.NextBefore)
		if err != nil {
			// 도달 불가 방어: 리더는 커서를 이미 같은 문법으로 통과시킨 뒤에만 넘긴다.
			return result, refuse(RefusalPageInvalid, "cursor "+strconv.Quote(page.NextBefore)+" is not a timestamp", err)
		}
		// 커서와 before는 둘 다 브로커 이름표(=닫는 시각) 공간의 값이다. 여는 시각으로
		// 옮기지 않고 그대로 견준다 — 상한과 상한을 비교하는 것이 맞다.
		if !cursor.Before(beforeInstant) {
			return result, refuse(RefusalCursorLoop,
				"cursor "+strconv.Quote(page.NextBefore)+" is not strictly older than the bound it answered", nil)
		}
		if result.Pages >= maxPages {
			result.Truncated = true
			break
		}
		if page.Budget.Exhausted() {
			result.RateGated = true
			break
		}
		before, beforeInstant, previous = page.NextBefore, cursor, current
	}
	// 이어 붙인 목록은 새 봉이 앞이고 엄격한 내림차순이어야 한다. 여기서 어긋나면
	// 어느 봉이 어느 봉의 successor인지가 무너지므로 폴 전체를 거절한다.
	//
	// not-applicable: checkOverlap과 같은 이유로 진짜 official.Client를 통해서는 도달할 수
	// 없다. 리더가 페이지 안에서 엄격한 내림차순을 강제하고, 커서가 그 페이지의 가장
	// 오래된 봉보다 이르며, 다음 페이지의 모든 봉이 그 커서 이하이기 때문이다.
	// 다른 CandleReader 구현을 위한 방어이며, 겹침 규칙이 첫 벌만 버리므로
	// "같은 분이 잇달아 두 번" 오는 경우는 여기서만 잡힌다.
	for index := 1; index < len(bars); index++ {
		if !bars[index].openAt.Before(bars[index-1].openAt) {
			return result, refuse(RefusalPageOrder,
				"the merged bars are not strictly newest-first at "+bars[index].openAt.Format(time.RFC3339), nil)
		}
	}

	result.Observed = len(bars)
	if len(bars) > 0 {
		result.NewestObserved = bars[0].openAt
	}
	result.Gaps = minuteGaps(bars, regularOpen, regularClose)

	// ---- (e) 이미 저장된 것을 한 번만 읽는다 ----
	series, err := store.SealBarSeries(ctx, strategyevidence.BarSeriesQuery{
		Market: code, Symbol: in.Symbol, SessionID: sessionID,
		IntervalMS:   strategyevidence.ClosedBar1mIntervalMS,
		EvaluationAt: in.PollAt.Add(knownReadHorizon), IngestionCutoff: in.PollAt.Add(knownReadHorizon),
		RegularSessionOnly: false, MaxBars: strategyevidence.MaxBarSeriesBars,
	})
	if err != nil {
		return result, refuse(RefusalStoreError, "reading the stored series", err)
	}
	// 질의가 이미 시장·종목·세션·간격을 고정했으므로, 이 묶음 안에서 봉의 신원은 여는 분이다.
	known := make(map[uint64]strategyevidence.ClosedBarRecord, len(series.Bars))
	for _, record := range series.Bars {
		known[record.Payload.OpenAtMS] = record
	}

	// ---- (f) 적재 ----
	// index는 1부터 돈다. bars[0]은 이 폴의 가장 새 봉이고, 그 봉이 닫혔다는 증거인
	// "다음 봉"은 아직 관측되지 않았다. 다음 폴에서 후계 봉이 보이면 그때 적재된다.
	for index := 1; index < len(bars); index++ {
		bar := bars[index]
		if bar.openAt.Before(regularOpen) || !bar.openAt.Before(regularClose) {
			continue
		}
		raw := strategyevidence.RawClosedBar1m{
			Open: bar.candle.Open, High: bar.candle.High, Low: bar.candle.Low,
			Close: bar.candle.Close, Volume: bar.candle.Volume,
		}
		revision := uint64(1)
		correction := false
		if record, found := known[uint64(bar.openAt.UnixMilli())]; found {
			if record.Payload.Raw == raw {
				result.Unchanged++
				continue
			}
			revision = record.Payload.Revision + 1
			correction = true
		}
		envelope, err := strategyevidence.NewClosedBar1mEnvelope(strategyevidence.ClosedBar1mInput{
			Market: market, Symbol: in.Symbol, SessionID: sessionID,
			CalendarVersion: in.Calendar.Version, OpenAt: bar.openAt, Revision: revision,
			ObservedAt: bar.readAt, Currency: bar.candle.Currency, Raw: raw,
			SuccessorOpenAt: bars[index-1].openAt, RegularSession: true,
			SourceResponseDigest: bar.digest,
		})
		if err != nil {
			result.Refused = append(result.Refused, BarRefusal{OpenAt: bar.openAt, Reason: err.Error()})
			continue
		}
		if _, err := store.Append(ctx, envelope); err != nil {
			if errors.Is(err, strategyevidence.ErrRevisionConflict) {
				result.Conflicts++
				continue
			}
			return result, refuse(RefusalStoreError, "appending the bar that opens at "+bar.openAt.UTC().Format(time.RFC3339), err)
		}
		result.Admitted++
		if correction {
			result.Corrections++
		}
	}
	return result, nil
}

// adoptPage는 한 페이지의 봉을 풀어 놓으면서, 브로커의 시각표를 **여는 시각**으로 옮긴다.
//
// 왜 빼는가 (2026-08-18 03:29 KST 사람이 직접 잰 값, 결정 30):
// 벽시계 03:29:14에 가장 새 봉의 이름표가 `03:30:00`이고 거래량 251이었는데,
// 26초 뒤 03:29:40에 **같은 이름표** `03:30:00`의 거래량이 1,089로 늘어 있었다.
// 그동안 이름표 `03:29:00` 봉은 2,997에서 꼼짝하지 않았다. 즉 벽시계가 [03:29, 03:30)
// 안에 있는 동안 자라던 봉의 이름표는 `03:30`이었다 — 이름표는 봉이 **닫는** 순간이다.
// 문서(openapi "봉 시작 시각")도, candle_reads.go의 주석도, a047 KR 선례도 이 점에서
// 모두 틀렸다. 26초짜리 측정 하나가 그 셋을 뒤집었다.
//
// 저장되는 증거의 시각 규약은 그대로 여는 시각(bar_label = "open_at")이다. 바뀐 것은
// 전선에서 들어오는 값의 뜻뿐이므로, 변환은 여기 한 곳에서만 한다. 이 뺄셈을 "고쳐서"
// 되돌리지 말 것 — 되돌리면 모든 봉이 1분씩 밀리고, 정규장 경계와 successor가 함께 어긋난다.
func adoptPage(page official.StrictMinutePage) ([]observedBar, error) {
	out := make([]observedBar, 0, len(page.Candles))
	for index, candle := range page.Candles {
		closeAt, err := time.Parse(time.RFC3339, candle.Timestamp)
		if err != nil {
			return nil, refuse(RefusalPageInvalid,
				"candle "+strconv.Itoa(index)+" carries the unreadable timestamp "+strconv.Quote(candle.Timestamp), err)
		}
		openAt := closeAt.UTC().Add(-barInterval)
		// 리더가 이름표의 분 정렬을 이미 확인했으므로 옮긴 값도 분 경계에 있어야 한다.
		// 그래도 값싸게 못을 박아 둔다. 이 불변식이 깨지면 봉 신원 자체가 무너진다.
		if openAt.Second() != 0 || openAt.Nanosecond() != 0 {
			return nil, refuse(RefusalPageInvalid,
				"candle "+strconv.Itoa(index)+" opens at "+openAt.Format(time.RFC3339Nano)+
					", which is not a whole minute", nil)
		}
		out = append(out, observedBar{candle: candle, openAt: openAt, readAt: page.ReadAt, digest: page.BodyDigest})
	}
	return out, nil
}

// checkOverlap은 다음 페이지가 앞 페이지와 어긋나지 않는지 본다.
//
// not-applicable: 진짜 official.Client를 통해서는 두 갈래 모두 도달할 수 없다. 리더가
// "커서 < 그 페이지의 가장 오래된 봉"과 "모든 봉 <= before"를 강제하므로, 다음 페이지는
// 언제나 앞 페이지보다 오래된 곳에서 시작한다. 이 검사는 CandleReader 인터페이스에
// 다른 구현이 들어왔을 때를 위한 방어다(L1a의 record-id 포섭 선례와 같은 성격).
// 실측 계약은 "첫 봉 == 커서"가 아니라 "첫 봉 <= 커서"다 — 장외 시간에는 커서가 가리키는
// 분에 체결이 없을 수 있어서, 같기를 요구하면 정상 크롤을 거절하게 된다.
func checkOverlap(previousLast, currentFirst observedBar) error {
	switch {
	case currentFirst.openAt.After(previousLast.openAt):
		return refuse(RefusalOverlapMismatch,
			"the next page starts at "+currentFirst.openAt.Format(time.RFC3339)+
				", newer than the previous page's oldest bar at "+previousLast.openAt.Format(time.RFC3339), nil)
	case currentFirst.openAt.Equal(previousLast.openAt) && currentFirst.candle != previousLast.candle:
		// 커서는 포함 상한이라 같은 봉이 두 페이지에 걸쳐 나올 수 있다. 그때 두 벌의
		// 7개 필드가 다르면 어느 쪽이 사실인지 알 수 없으므로 폴 전체를 거절한다.
		return refuse(RefusalOverlapMismatch,
			"the bar at "+currentFirst.openAt.Format(time.RFC3339)+" differs between the two pages that carry it", nil)
	}
	return nil
}

// minuteGaps는 정규장 창 안에서 관측되지 않은 분들을 센다. 보고만 한다.
func minuteGaps(bars []observedBar, regularOpen, regularClose time.Time) []MinuteGap {
	var gaps []MinuteGap
	for index := 1; index < len(bars); index++ {
		newer, older := bars[index-1].openAt, bars[index].openAt
		start, end := older.Add(time.Minute), newer
		if start.Before(regularOpen) {
			start = regularOpen
		}
		if end.After(regularClose) {
			end = regularClose
		}
		if !start.Before(end) {
			continue
		}
		minutes := int(end.Sub(start) / time.Minute)
		if minutes <= 0 {
			continue
		}
		gaps = append(gaps, MinuteGap{From: start.UTC(), To: end.Add(-time.Minute).UTC(), Minutes: minutes})
	}
	return gaps
}

// checkSymbol은 "요청을 만들 수조차 없는" 종목만 여기서 막는다. 시장별 문법은
// 리더가 요청을 보내기 전에, 증거 생성자와 SealBarSeries가 다시 한 번 거절한다.
// 같은 문법을 여기에 한 벌 더 두면 세 곳이 서로 어긋날 수 있다.
func checkSymbol(symbol string) error {
	if strings.TrimSpace(symbol) == "" || strings.ContainsAny(symbol, " \t\r\n") {
		return refuse(RefusalSymbolInvalid,
			"symbol "+strconv.Quote(symbol)+" is empty or carries whitespace", nil)
	}
	return nil
}

func marketCode(market marketclock.Market) string {
	if market == marketclock.MarketKR {
		return "KR"
	}
	return "US"
}

// sessionCalendar는 세션 이름표의 앞자리다. 한국은 KRX, 미국은 US다.
func sessionCalendar(code string) string {
	if code == "KR" {
		return "KRX"
	}
	return "US"
}
