package main

// receipts.go 는 a128 사전등록 D2 의 공식 1분봉 영수증을 만듦.
//
// 무엇을 하나
//
// D2(openspec/changes/a128-jev-shadow-judgment-probe/analysis/prereg-2026-10-23-d2-twice-daily-index-correlation.md)
// 평가기는 진입·청산 봉을 state 합집합에서 먼저 찾고, 없을 때만 이 영수증을 읽음(3절). 이 도구는 그 영수증을
// 만듦 — 원장에서 그날 10:00 회차 목록을 읽고, 목록 종목마다 진입·청산 시각 T 의 [T−5분, T] 봉만 남겨 파일로 고정함.
//
// 무엇을 하지 않나
//
// 주문·계좌 조회 없음(공식 1분봉 GET 만). 원장·state 를 쓰지 않음. 429 를 받으면 재시도하지 않고 멈추며 파일을
// 남기지 않음. 이미 있는 영수증을 덮어쓰지 않음(증거 파일).

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

const j2Question = "j2_hold_off_red_flag"

// candleSource 는 공식 클라이언트의 엄격 1분봉 리더 한 개만 씀 — 시험에서 가짜로 바꿔 끼우려는 경계.
type candleSource interface {
	StrictMinuteCandles(ctx context.Context, market, symbol string, count int, before string) (official.StrictMinutePage, error)
}

// fetchPlan 은 한 번의 요청이다. Before(봉 닫는 시각의 포함 상한)부터 Count 개를 받아 Targets 의 [T−5분, T] 봉만 남김.
type fetchPlan struct {
	Before  time.Time
	Count   int
	Targets []time.Time
}

// marketLocation 은 시장 현지 시계(KR = KST 고정, US = America/New_York 서머타임 반영)임.
func marketLocation(market string) (*time.Location, error) {
	switch market {
	case "KR":
		return time.FixedZone("KST", 9*3600), nil
	case "US":
		return time.LoadLocation("America/New_York")
	}
	return nil, fmt.Errorf("market must be KR or US, got %q", market)
}

// planFor 는 사전등록 3절의 진입·청산 시각을 요청 두 개로 묶음(종목당 호출 2 회로 줄이려는 것).
// KR: 10:00 진입 / 13:00 진입 + 14:10 청산. US: 10:00 진입 + 11:00 청산 / 13:00 진입 + 14:00 청산(ET).
// Count = 가장 이른 목표의 T−5분 봉까지 닿는 봉 수(분당 봉 하나 가정; 거래 없는 분이 있으면 더 앞까지 받을 뿐 결과는 거름).
// 최대 76(KR 12:55~14:10) — 엄격 리더 상한 200 안.
func planFor(market string, day time.Time) ([]fetchPlan, error) {
	loc, err := marketLocation(market)
	if err != nil {
		return nil, err
	}
	at := func(h, m int) time.Time { return time.Date(day.Year(), day.Month(), day.Day(), h, m, 0, 0, loc) }
	groups := map[string][][]time.Time{
		"KR": {{at(10, 0)}, {at(13, 0), at(14, 10)}},
		"US": {{at(10, 0), at(11, 0)}, {at(13, 0), at(14, 0)}},
	}[market]
	plans := make([]fetchPlan, 0, len(groups))
	for _, targets := range groups {
		before := targets[len(targets)-1]
		count := int(before.Sub(targets[0])/time.Minute) + 6
		plans = append(plans, fetchPlan{Before: before, Count: count, Targets: targets})
	}
	return plans, nil
}

// listSymbols 는 원장에서 D2 목록을 읽음 — 세션 날짜(at 의 UTC 날짜) = date 이고 at 의 현지 시각이 10:00:00 이상
// 10:10:00 미만인 J2 판단의 종목(사전등록 2절). 마지막 줄이 쓰는 중이라 잘렸으면 건너뜀, 중간 줄 손상은 오류.
// 반환: 정렬된 종목, 읽은 원장 바이트의 sha256.
func listSymbols(r io.Reader, market, date string) ([]string, string, error) {
	loc, err := marketLocation(market)
	if err != nil {
		return nil, "", err
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(data)
	lines := bytes.Split(data, []byte("\n"))
	last := -1
	for i, ln := range lines {
		if len(bytes.TrimSpace(ln)) > 0 {
			last = i
		}
	}
	seen := map[string]bool{}
	for i, ln := range lines {
		if len(bytes.TrimSpace(ln)) == 0 {
			continue
		}
		var row struct {
			Kind       string `json:"kind"`
			QuestionID string `json:"question_id"`
			Market     string `json:"market"`
			Symbol     string `json:"symbol"`
			At         string `json:"at"`
		}
		if err := json.Unmarshal(ln, &row); err != nil {
			if i == last {
				continue // 쓰는 중인 마지막 줄
			}
			return nil, "", fmt.Errorf("ledger line %d: %w", i+1, err)
		}
		if row.Kind != "judgment" || row.QuestionID != j2Question || row.Market != market {
			continue
		}
		at, err := time.Parse(time.RFC3339, row.At)
		if err != nil {
			return nil, "", fmt.Errorf("ledger line %d: at %q: %w", i+1, row.At, err)
		}
		local := at.In(loc)
		if at.UTC().Format("2006-01-02") == date && local.Hour() == 10 && local.Minute() < 10 {
			seen[row.Symbol] = true
		}
	}
	symbols := make([]string, 0, len(seen))
	for s := range seen {
		symbols = append(symbols, s)
	}
	sort.Strings(symbols)
	return symbols, "sha256:" + hex.EncodeToString(sum[:]), nil
}

type receiptBar struct {
	Time  string      `json:"time"`
	Close json.Number `json:"close"`
}

type fetchRecord struct {
	Symbol     string `json:"symbol"`
	Before     string `json:"before"`
	Count      int    `json:"count"`
	Status     string `json:"status"` // "ok" 또는 "refused: <사유>"(엄격 리더 거절 — 그 요청만 빠지고 계속)
	Candles    int    `json:"candles"`
	Kept       int    `json:"kept"`
	BodyDigest string `json:"body_digest,omitempty"`
}

// receipt 는 영수증 파일 내용이다. 평가기는 bars 만 읽고 나머지는 감사용 기록임.
type receipt struct {
	Market         string                  `json:"market"`
	Date           string                  `json:"date"`
	Prereg         string                  `json:"prereg"`
	ListRule       string                  `json:"list_rule"`
	LedgerDigest   string                  `json:"ledger_digest"`
	Symbols        []string                `json:"symbols"`
	FetchedAt      string                  `json:"fetched_at"`
	ElapsedSeconds float64                 `json:"elapsed_seconds"`
	Calls          int                     `json:"calls"`
	Fetches        []fetchRecord           `json:"fetches"`
	Bars           map[string][]receiptBar `json:"bars"`
}

// errRateLimited 는 429 로 멈췄다는 뜻 — 파일을 남기지 않음.
var errRateLimited = errors.New("rate limited (429): stopped without retry; no receipt written")

// collect 는 목록 종목마다 계획한 요청을 차례로 보내고 [T−5분, T] 봉만 남김.
// 429 는 즉시 멈춤(재시도 없음). 엄격 리더의 계약 거절은 그 요청만 기록하고 계속. 그 밖의 오류(전송·인증·5xx)는 멈춤.
func collect(ctx context.Context, src candleSource, market, date string, symbols []string, plans []fetchPlan,
	gap time.Duration, sleep func(context.Context, time.Duration) error, now func() time.Time) (receipt, error) {
	started := now()
	rc := receipt{Market: market, Date: date, Symbols: symbols, Bars: map[string][]receiptBar{}, Fetches: []fetchRecord{}}
	for _, sym := range symbols {
		for _, p := range plans {
			if rc.Calls > 0 && gap > 0 {
				if err := sleep(ctx, gap); err != nil {
					return receipt{}, err
				}
			}
			before := p.Before.Format(time.RFC3339)
			rc.Calls++
			page, err := src.StrictMinuteCandles(ctx, market, sym, p.Count, before)
			rec := fetchRecord{Symbol: sym, Before: before, Count: p.Count}
			var refused *official.StrictMinuteCandlesError
			switch {
			case errors.Is(err, official.ErrRateLimited):
				return receipt{}, fmt.Errorf("%w (call %d, %s before %s)", errRateLimited, rc.Calls, sym, before)
			case errors.As(err, &refused):
				rec.Status = "refused: " + refused.Reason
				rc.Fetches = append(rc.Fetches, rec)
				continue
			case err != nil:
				return receipt{}, fmt.Errorf("call %d, %s before %s: %w", rc.Calls, sym, before, err)
			}
			rec.Status, rec.Candles, rec.BodyDigest = "ok", len(page.Candles), page.BodyDigest
			for _, c := range page.Candles {
				closeAt, err := time.Parse(time.RFC3339, c.Timestamp)
				if err != nil {
					return receipt{}, fmt.Errorf("%s: candle timestamp %q: %w", sym, c.Timestamp, err)
				}
				if inAnyWindow(closeAt, p.Targets) {
					rc.Bars[sym] = append(rc.Bars[sym], receiptBar{Time: c.Timestamp, Close: json.Number(c.Close)})
					rec.Kept++
				}
			}
			rc.Fetches = append(rc.Fetches, rec)
		}
		sort.Slice(rc.Bars[sym], func(i, j int) bool { return rc.Bars[sym][i].Time < rc.Bars[sym][j].Time })
		if len(rc.Bars[sym]) == 0 {
			delete(rc.Bars, sym)
		}
	}
	end := now()
	rc.FetchedAt = end.UTC().Format(time.RFC3339)
	rc.ElapsedSeconds = end.Sub(started).Seconds()
	return rc, nil
}

// inAnyWindow 는 봉 닫는 시각이 어느 목표 T 의 [T−5분, T] 안인지(양끝 포함, 사전등록 3절).
func inAnyWindow(closeAt time.Time, targets []time.Time) bool {
	for _, t := range targets {
		if !closeAt.Before(t.Add(-5*time.Minute)) && !closeAt.After(t) {
			return true
		}
	}
	return false
}

// writeReceipt 는 <out>/minute-bars/<MARKET>-<date>.json 을 새로 만듦. 이미 있으면 거절(증거 덮어쓰기 금지).
func writeReceipt(outDir string, rc receipt) (string, error) {
	dir := filepath.Join(outDir, "minute-bars")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, rc.Market+"-"+rc.Date+".json")
	body, err := json.MarshalIndent(rc, "", "  ")
	if err != nil {
		return "", err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	// 쓰다 실패하면 부분 파일을 지움 — 남으면 O_EXCL 이 재수집을 막음
	w := bufio.NewWriter(f)
	_, err = w.Write(append(body, '\n'))
	if err == nil {
		err = w.Flush()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(path)
		return "", err
	}
	return path, nil
}
