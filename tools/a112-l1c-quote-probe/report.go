package main

// report.go는 관측을 사람이 읽는 보고서로 바꾼다. 값은 절대 싣지 않는다.
//
// 결정 42는 "모양만" 보고하라고 했다. 그래서 이 파일은 소수의 **자릿수**만 세고,
// 가격과 잔량 자체는 한 번도 문자열에 넣지 않는다. 값이 새면 그 보고서는 저장소에
// 남길 수 없는 것이 되고, 남길 수 없는 보고서는 증거가 되지 못한다.

import (
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// observation은 한 번의 프로브가 본 것 전부다.
type observation struct {
	Market string
	Symbol string
	Top    official.StrictTopOfBook
	Last   official.StrictLastPrice
}

// decimalShape는 소수 문자열의 정수부·소수부 자릿수를 센다.
// 값을 해석하지 않는다 — 문법이 계약과 다르면 그냥 아니라고 답한다.
func decimalShape(raw string) (int, int, bool) {
	if raw == "" {
		return 0, 0, false
	}
	integerPart, fractionPart := raw, ""
	if index := strings.IndexByte(raw, '.'); index >= 0 {
		integerPart, fractionPart = raw[:index], raw[index+1:]
		if fractionPart == "" || strings.IndexByte(fractionPart, '.') >= 0 {
			return 0, 0, false
		}
	}
	if integerPart == "" || !onlyDigits(integerPart) || !onlyDigits(fractionPart) {
		return 0, 0, false
	}
	return len(integerPart), len(fractionPart), true
}

func onlyDigits(text string) bool {
	for index := 0; index < len(text); index++ {
		if text[index] < '0' || text[index] > '9' {
			return false
		}
	}
	return true
}

// marketScale은 시장의 소수 배율이다. 증거 꾸러미가 쓰는 값과 같은 숫자지만 이 도구는
// 그 꾸러미를 부르지 않는다 — 적재 경로에 닿지 않는 것이 이 프로브의 성질이고,
// 같은 파일의 정적 가드가 그 이름이 나타나기만 해도 실패한다.
func marketScale(market string) (int, bool) {
	switch market {
	case "KR":
		return 0, true
	case "US":
		return 4, true
	default:
		return 0, false
	}
}

type decimalRow struct {
	name string
	raw  string
	// price는 이 소수가 배율 판정 대상인지다. 잔량은 증거 본문에 들어가지 않는다.
	price bool
}

func renderReport(observed observation) string {
	var out strings.Builder
	line := func(format string, args ...any) {
		fmt.Fprintf(&out, format+"\n", args...)
	}
	currency := observed.Top.Currency
	line("a112 L1c quote probe — shape only, no prices and no volumes")
	line("market: %s  symbol: %s  currency: %s", observed.Market, observed.Symbol, currency)
	line("orderbook: HTTP %d  digest %s", observed.Top.StatusCode, observed.Top.BodyDigest)
	line("prices:    HTTP %d  digest %s", observed.Last.StatusCode, observed.Last.BodyDigest)
	line("broker instants: orderbook %s | prices %s", observed.Top.SourceTimestamp, observed.Last.SourceTimestamp)
	line("broker instant difference: %s", gap(observed.Top.SourceInstant, observed.Last.SourceInstant))
	line("read instant difference: %s", gap(observed.Top.ReadAt, observed.Last.ReadAt))
	line("seal binding (decision 36): source_observed_at = the %s half; received_at = the %s half",
		half(observed.Last.SourceInstant.Before(observed.Top.SourceInstant)),
		half(observed.Last.ReadAt.After(observed.Top.ReadAt)))

	line("decimal shape (digit counts only, values withheld):")
	rows := []decimalRow{
		{"ask.price", observed.Top.Ask.Price, true},
		{"ask.volume", observed.Top.Ask.Volume, false},
		{"bid.price", observed.Top.Bid.Price, true},
		{"bid.volume", observed.Top.Bid.Volume, false},
		{"last", observed.Last.Last, true},
	}
	scale, known := marketScale(observed.Market)
	overPrecise := make([]string, 0, len(rows))
	malformed := make([]string, 0, len(rows))
	for _, row := range rows {
		integer, fraction, ok := decimalShape(row.raw)
		if !ok {
			line("  %-11s not a plain decimal string", row.name)
			malformed = append(malformed, row.name)
			continue
		}
		line("  %-11s %d integer digits, %d fraction digits", row.name, integer, fraction)
		if row.price && known && fraction > scale {
			overPrecise = append(overPrecise, row.name)
		}
	}

	switch {
	case !known:
		line("scale verdict: market %q has no known scale", observed.Market)
	case len(malformed) > 0:
		line("scale verdict: %s scale %d — %s is not a plain decimal, so the quote would be refused",
			currency, scale, strings.Join(malformed, ", "))
	case len(overPrecise) > 0:
		line("scale verdict: %s scale %d — %s carries more fraction digits than the scale, so the quote would be refused",
			currency, scale, strings.Join(overPrecise, ", "))
	default:
		line("scale verdict: %s scale %d — every decimal fits the market scale, so the producer would admit this quote",
			currency, scale)
	}
	line("level count: not measured here — the reader reads index 0 only (decision 33); " +
		"KR ten levels / US one stands from the 2026-08-18 console probe")
	return out.String()
}

// gap은 두 시각 사이의 절대 간격이다. 어느 쪽이 앞인지는 위의 seal binding 줄이 말한다.
func gap(left, right time.Time) time.Duration {
	difference := right.Sub(left)
	if difference < 0 {
		return -difference
	}
	return difference
}

func half(lastPriceWins bool) string {
	if lastPriceWins {
		return "last-price"
	}
	return "top-of-book"
}
