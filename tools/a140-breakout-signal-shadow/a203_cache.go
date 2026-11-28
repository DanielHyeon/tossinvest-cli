package main

// a203_cache.go 는 a203 재생의 저장소 밖 캐시와 그 앞의 전송 관문(design §3-1 · §8).
//
// 모든 브로커 GET 은 a203Get 한 곳을 지남: 캐시에 있으면 GET 0, 없으면 간격(직전 전송 시작부터 1 초) → 시간대 거절(결정 M3) →
// 실행당 상한 → 전송 → 캐시 기록. 전송 실패(429 · 예산 거절 · 그 밖 오류)는 재시도하지 않고 그 실행 전체를 멈춤(sticky).
// 다시 돌리면 캐시에 있는 것은 다시 묻지 않으므로 그 자리에서 이어짐.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

const a203CacheSchema = "a203-cache/v1"

// a203Live 는 a203 이 부르는 공식 읽기 GET 넷(design §8) — *official.Client 가 그대로 만족함. 주문 · 계좌 · 쓰기 표면 없음.
type a203Live interface {
	StrictMinuteCandles(ctx context.Context, market, symbol string, count int, before string) (official.StrictMinutePage, error)
	Candles(ctx context.Context, symbol, interval string, count int, before string, adjusted bool) (domain.Chart, error)
	Rankings(ctx context.Context, typ, marketCountry, duration string, excludeCaution bool, count int) (domain.Ranking, error)
	TypedMarketCalendar(ctx context.Context, country, date string) (official.MarketCalendarResponse, error)
}

var (
	errA203NotCached = errors.New("not in the a203 cache (replay and dry-run never send)")
	// 캐시 열쇠가 파일 이름 한 조각이 되므로 경로 이탈 글자를 요청 전에 막음.
	a203CacheSymbol = regexp.MustCompile(`^[A-Z0-9]{6}$`)
	a203CacheDate   = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
	a203CacheBefore = regexp.MustCompile(`^[0-9T:.+Z-]{1,40}$`)
)

// a203Request 는 캐시 한 칸의 요청 인자(종류마다 쓰는 칸이 다름).
type a203Request struct {
	Market   string `json:"market,omitempty"`
	Symbol   string `json:"symbol,omitempty"`
	Interval string `json:"interval,omitempty"`
	Count    int    `json:"count,omitempty"`
	Before   string `json:"before,omitempty"`
	Adjusted bool   `json:"adjusted,omitempty"`
	Country  string `json:"country,omitempty"`
	Date     string `json:"date,omitempty"`
	Type     string `json:"type,omitempty"`
	Duration string `json:"duration,omitempty"`
}

// a203Record 는 캐시 파일 하나. Response 는 응답 값 그대로(1분봉은 StrictMinutePage 에서 rate 예산 칸만 RateBudget 으로 옮김 —
// 아래 a203Fetcher.minute 주석).
type a203Record struct {
	Schema     string          `json:"schema"`
	Kind       string          `json:"kind"`
	Request    a203Request     `json:"request"`
	FetchedAt  time.Time       `json:"fetched_at"`
	Response   json.RawMessage `json:"response"`
	RateBudget json.RawMessage `json:"rate_budget,omitempty"`
}

// a203StopError 는 실행 전체를 멈춘 사유 — 체크포인트에 그대로 적힘.
type a203StopError struct {
	Reason string
	Key    string
	Err    error
}

func (e *a203StopError) Error() string {
	msg := "a203 collection stopped: " + e.Reason
	if e.Key != "" {
		msg += " (at " + e.Key + ")"
	}
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	return msg
}

func (e *a203StopError) Unwrap() error { return e.Err }

// a203Fetcher 는 캐시 + 전송 관문. live 가 nil 이면 캐시만 읽음(재생 · dry-run — 전송 0).
type a203Fetcher struct {
	live    a203Live
	root    string // 캐시 디렉터리
	now     func() time.Time
	sleep   func(context.Context, time.Duration) error
	gap     time.Duration
	maxGets int
	gets    int
	byKind  map[string]int
	hits    int
	last    time.Time
	stop    error
}

func newA203Fetcher(live a203Live, root string, now func() time.Time, sleep func(context.Context, time.Duration) error, gap time.Duration, maxGets int) *a203Fetcher {
	return &a203Fetcher{live: live, root: root, now: now, sleep: sleep, gap: gap, maxGets: maxGets, byKind: map[string]int{}}
}

// a203Get 은 캐시 한 칸을 읽거나, 없으면 관문을 지나 한 번 보내고 기록함. encode 는 캐시에 쓸 응답 바이트(기본 json.Marshal).
func a203Get[T any](ctx context.Context, f *a203Fetcher, kind, key string, req a203Request,
	call func(context.Context, a203Live) (T, error), encode func(T) (json.RawMessage, json.RawMessage, error),
) (T, error) {
	var zero T
	if f.stop != nil {
		return zero, f.stop
	}
	path := filepath.Join(f.root, kind, key)
	raw, err := os.ReadFile(path)
	if err == nil {
		var rec a203Record
		if err := json.Unmarshal(raw, &rec); err != nil || rec.Schema != a203CacheSchema || rec.Kind != kind {
			return zero, fmt.Errorf("the a203 cache file %s is unreadable (schema %q): %v", path, rec.Schema, err)
		}
		var value T
		if err := json.Unmarshal(rec.Response, &value); err != nil {
			return zero, fmt.Errorf("the a203 cache file %s: %w", path, err)
		}
		f.hits++
		return value, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return zero, err
	}
	if f.live == nil {
		return zero, fmt.Errorf("%s/%s: %w", kind, key, errA203NotCached)
	}
	at := kind + "/" + key
	if f.gets >= f.maxGets {
		f.stop = &a203StopError{Reason: fmt.Sprintf("the per-run GET cap %d is reached", f.maxGets), Key: at}
		return zero, f.stop
	}
	// 간격은 직전 전송 시작부터(응답 시간과 무관) — 기다린 **뒤에** 시간대를 봐야 기다리는 사이 열린 거절 창에 보내지 않음.
	if !f.last.IsZero() {
		if wait := f.last.Add(f.gap).Sub(f.now()); wait > 0 {
			if err := f.sleep(ctx, wait); err != nil {
				f.stop = &a203StopError{Reason: "interrupted", Key: at, Err: err}
				return zero, f.stop
			}
		}
	}
	sendAt := f.now()
	if refused, why := a203Refused(sendAt); refused {
		f.stop = &a203StopError{Reason: "time window refused at " + sendAt.UTC().Format(time.RFC3339) + " — " + why, Key: at}
		return zero, f.stop
	}
	f.gets++
	f.byKind[kind]++
	f.last = sendAt
	value, err := call(ctx, f.live)
	if err != nil {
		f.stop = &a203StopError{Reason: "the GET failed (no retry)", Key: at, Err: err}
		return zero, f.stop
	}
	if encode == nil {
		encode = func(v T) (json.RawMessage, json.RawMessage, error) { b, err := json.Marshal(v); return b, nil, err }
	}
	response, rate, err := encode(value)
	if err == nil {
		err = a203WriteFile(path, a203Record{Schema: a203CacheSchema, Kind: kind, Request: req, FetchedAt: sendAt.UTC(), Response: response, RateBudget: rate})
	}
	if err != nil {
		// 받은 응답을 못 적으면 다음 실행이 같은 GET 을 또 보냄 — 쿼터를 지키려 여기서 멈춤.
		f.stop = &a203StopError{Reason: "writing the cache failed", Key: at, Err: err}
		return zero, f.stop
	}
	return value, nil
}

// minute 은 엄격 1분봉 한 페이지(officialbars.CandleReader 를 만족하도록 a203MinuteReader 가 감쌈).
//
// rate 예산 칸은 캐시 응답에서 빼 RateBudget 에 따로 둠: 남은 몫 0 을 실은 페이지가 그대로 캐시되면 재생 · 이어하기마다
// officialbars 가 그 페이지에서 페이지 넘김을 멈춰(RateGated) 그 종목-세션이 영구 결손이 됨. 수집 중의 RateGated 는
// 호출자(collect)가 실행 정지로 다룸.
func (f *a203Fetcher) minute(ctx context.Context, market, symbol string, count int, before string) (official.StrictMinutePage, error) {
	if !a203CacheSymbol.MatchString(symbol) || !a203CacheBefore.MatchString(before) || market != "KR" {
		return official.StrictMinutePage{}, fmt.Errorf("a203 cache: refusing the minute request %s/%s before %q", market, symbol, before)
	}
	key := symbol + "/" + strings.ReplaceAll(before, ":", "") + "-c" + fmt.Sprint(count) + ".json"
	req := a203Request{Market: market, Symbol: symbol, Count: count, Before: before}
	return a203Get(ctx, f, "minute", key, req, func(ctx context.Context, live a203Live) (official.StrictMinutePage, error) {
		return live.StrictMinuteCandles(ctx, market, symbol, count, before)
	}, func(page official.StrictMinutePage) (json.RawMessage, json.RawMessage, error) {
		raw, err := json.Marshal(page)
		if err != nil {
			return nil, nil, err
		}
		fields := map[string]json.RawMessage{}
		if err := json.Unmarshal(raw, &fields); err != nil {
			return nil, nil, err
		}
		rate := fields["Budget"]
		delete(fields, "Budget")
		raw, err = json.Marshal(fields)
		return raw, rate, err
	})
}

// daily 는 일봉 200 개(표준 해독기 Client.Candles — design §5, 엄격 일봉 리더 없음). 종목당 1 회.
func (f *a203Fetcher) daily(ctx context.Context, symbol string) (domain.Chart, error) {
	if !a203CacheSymbol.MatchString(symbol) {
		return domain.Chart{}, fmt.Errorf("a203 cache: refusing the daily request for %q", symbol)
	}
	req := a203Request{Symbol: symbol, Interval: "1d", Count: 200}
	return a203Get(ctx, f, "daily", symbol+".json", req, func(ctx context.Context, live a203Live) (domain.Chart, error) {
		return live.Candles(ctx, symbol, "1d", 200, "", false)
	}, nil)
}

// calendar 는 KR 달력 한 날짜(today · 전 · 다음 영업일).
func (f *a203Fetcher) calendar(ctx context.Context, date string) (official.MarketCalendarResponse, error) {
	if !a203CacheDate.MatchString(date) {
		return official.MarketCalendarResponse{}, fmt.Errorf("a203 cache: refusing the calendar date %q", date)
	}
	req := a203Request{Country: "KR", Date: date}
	return a203Get(ctx, f, "calendar", date+".json", req, func(ctx context.Context, live a203Live) (official.MarketCalendarResponse, error) {
		return live.TypedMarketCalendar(ctx, "KR", date)
	}, nil)
}

// a203PBRanking 은 P(b) 랭킹 인자(design §5) — 수집 시작 때 1 회, 그 뒤로는 캐시.
var a203PBRanking = a203Request{Type: "MARKET_TRADING_AMOUNT", Country: "KR", Duration: "6mo", Count: 100}

func (f *a203Fetcher) ranking(ctx context.Context) (domain.Ranking, error) {
	r := a203PBRanking
	return a203Get(ctx, f, "ranking", "pb-6mo-100.json", r, func(ctx context.Context, live a203Live) (domain.Ranking, error) {
		return live.Rankings(ctx, r.Type, r.Country, r.Duration, false, r.Count)
	}, nil)
}

// a203MinuteReader 는 fetcher 를 officialbars.CandleReader 로 보이게 함.
type a203MinuteReader struct{ f *a203Fetcher }

func (r a203MinuteReader) StrictMinuteCandles(ctx context.Context, market, symbol string, count int, before string) (official.StrictMinutePage, error) {
	return r.f.minute(ctx, market, symbol, count, before)
}

// a203NullStore 는 수집 · dry-run 용 officialbars.BarStore — 적재하지 않음. 페이지 넘김은 저장소 내용과 무관하므로
// 수집은 재생과 같은 요청 열(같은 before · 커서)을 만들고, 봉 적재는 재생(report)이 실제 evidence 저장소에 함.
type a203NullStore struct{}

func (a203NullStore) Append(context.Context, strategyevidence.Envelope) (strategyevidence.AppendResult, error) {
	return strategyevidence.AppendResult{}, nil
}

func (a203NullStore) SealBarSeries(context.Context, strategyevidence.BarSeriesQuery) (strategyevidence.BarSeries, error) {
	return strategyevidence.BarSeries{}, nil
}

// a203WriteFile 은 JSON 한 벌을 0600 파일로 원자적으로 씀(임시 파일 → fsync → rename, 디렉터리 0700).
func a203WriteFile(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", " ")
	if err != nil {
		return err
	}
	return a203WriteBytes(path, append(raw, '\n'))
}

func a203WriteBytes(path string, raw []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // rename 뒤에는 없는 파일 — 실패 경로의 찌꺼기만 지움
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// a203SHA 는 바이트의 sha256 16진수.
func a203SHA(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// a203CacheFingerprint 는 캐시 디렉터리 지문 = 「상대 경로 + 공백 + 파일 sha256」 줄(경로순)의 sha256, 그리고 파일 수(design §8).
func a203CacheFingerprint(root string) (string, int, error) {
	var lines []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) && path == root {
				return filepath.SkipAll
			}
			return err
		}
		if d.IsDir() || strings.HasPrefix(d.Name(), ".tmp-") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(raw)
		lines = append(lines, filepath.ToSlash(rel)+" "+hex.EncodeToString(sum[:]))
		return nil
	})
	if err != nil {
		return "", 0, err
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:]), len(lines), nil
}
