package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

const (
	fixtureAskPrice  = "231.7000"
	fixtureAskVolume = "160"
	fixtureBidPrice  = "231.6500"
	fixtureBidVolume = "40"
	fixtureLast      = "231.6800"
)

func fixtureObservation(t *testing.T) observation {
	t.Helper()
	bookInstant, err := time.Parse(time.RFC3339, "2026-08-28T03:29:14.120+09:00")
	if err != nil {
		t.Fatal(err)
	}
	priceInstant := bookInstant.Add(-20 * time.Millisecond)
	bookRead := time.Date(2026, 8, 27, 18, 29, 14, 300*int(time.Millisecond), time.UTC)
	priceRead := bookRead.Add(143 * time.Millisecond)
	return observation{
		Market: "US", Symbol: "AAPL",
		Top: official.StrictTopOfBook{
			Market: "US", Symbol: "AAPL", Currency: "USD",
			Ask:             official.StrictQuoteLevel{Price: fixtureAskPrice, Volume: fixtureAskVolume},
			Bid:             official.StrictQuoteLevel{Price: fixtureBidPrice, Volume: fixtureBidVolume},
			SourceTimestamp: "2026-08-28T03:29:14.120+09:00", SourceInstant: bookInstant,
			ReadAt: bookRead, StatusCode: 200, BodyDigest: "sha256:aa",
		},
		Last: official.StrictLastPrice{
			Market: "US", Symbol: "AAPL", Currency: "USD", Last: fixtureLast,
			SourceTimestamp: "2026-08-28T03:29:14.100+09:00", SourceInstant: priceInstant,
			ReadAt: priceRead, StatusCode: 200, BodyDigest: "sha256:bb",
		},
	}
}

// TestReportNeverCarriesAValue는 이 도구의 존재 이유를 지킨다. 결정 42는 모양만
// 보고하라고 했다 — 값이 한 번이라도 새면 그 보고서는 저장소에 남기면 안 되는 것이 된다.
func TestReportNeverCarriesAValue(t *testing.T) {
	t.Parallel()
	report := renderReport(fixtureObservation(t))
	for _, secret := range []string{fixtureAskPrice, fixtureBidPrice, fixtureLast, fixtureAskVolume, fixtureBidVolume} {
		if strings.Contains(report, secret) {
			t.Fatalf("the report leaks the value %q:\n%s", secret, report)
		}
	}
}

// TestReportCarriesTheShapeDecisionFortyTwoAsksFor는 보고서가 실제로 무엇을 답하는지 고정한다.
func TestReportCarriesTheShapeDecisionFortyTwoAsksFor(t *testing.T) {
	t.Parallel()
	report := renderReport(fixtureObservation(t))
	for _, want := range []string{
		"market: US", "symbol: AAPL", "currency: USD",
		"ask.price", "ask.volume", "bid.price", "bid.volume", "last",
		"3 integer digits, 4 fraction digits",
		"3 integer digits, 0 fraction digits",
		"broker instant difference: 20ms",
		"read instant difference: 143ms",
		"source_observed_at = the last-price half",
		"received_at = the last-price half",
		"sha256:aa", "sha256:bb", "HTTP 200",
		"scale 4",
		"every decimal fits the market scale",
		"level count: not measured here",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("the report does not carry %q:\n%s", want, report)
		}
	}
}

func TestDecimalShapeCountsDigitsWithoutInterpretingTheValue(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		raw               string
		integer, fraction int
		ok                bool
	}{
		{"231.7000", 3, 4, true},
		{"284000", 6, 0, true},
		{"0.5", 1, 1, true},
		{"", 0, 0, false},
		{"231.", 0, 0, false},
		{".5", 0, 0, false},
		{"1.2.3", 0, 0, false},
		{"-1", 0, 0, false},
		{"1e5", 0, 0, false},
		{"1 000", 0, 0, false},
	} {
		t.Run(testCase.raw, func(t *testing.T) {
			t.Parallel()
			integer, fraction, ok := decimalShape(testCase.raw)
			if ok != testCase.ok || integer != testCase.integer || fraction != testCase.fraction {
				t.Fatalf("decimalShape(%q) = %d, %d, %v", testCase.raw, integer, fraction, ok)
			}
		})
	}
}

// TestScaleVerdictFollowsTheMarketScale는 소수 자릿수가 시장 배율을 넘으면 생산자가
// 거절하리라는 것을 보고서가 미리 말하는지 본다(적재는 하지 않는다).
func TestScaleVerdictFollowsTheMarketScale(t *testing.T) {
	t.Parallel()
	observed := fixtureObservation(t)
	if verdict := renderReport(observed); !strings.Contains(verdict, "every decimal fits the market scale") {
		t.Fatalf("a four-decimal USD quote should fit:\n%s", verdict)
	}
	overPrecise := fixtureObservation(t)
	overPrecise.Top.Ask.Price = "231.70001"
	report := renderReport(overPrecise)
	if !strings.Contains(report, "would be refused") {
		t.Fatalf("a five-decimal USD price must be reported as refusable:\n%s", report)
	}
	kr := fixtureObservation(t)
	kr.Market, kr.Symbol = "KR", "005930"
	kr.Top.Market, kr.Top.Symbol, kr.Top.Currency = "KR", "005930", "KRW"
	kr.Last.Market, kr.Last.Symbol, kr.Last.Currency = "KR", "005930", "KRW"
	kr.Top.Ask.Price, kr.Top.Bid.Price, kr.Last.Last = "284000", "283500", "283900"
	if report := renderReport(kr); !strings.Contains(report, "scale 0") {
		t.Fatalf("KRW carries scale 0:\n%s", report)
	}
}

// TestProbeReachesOnlyTheTwoStrictReaders는 이 도구가 읽기 전용임을 코드로 확인한다.
// 약속이 아니라 도달 가능한 것으로 막는다(결정 45).
func TestProbeReachesOnlyTheTwoStrictReaders(t *testing.T) {
	t.Parallel()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	files, err := filepath.Glob(filepath.Join(filepath.Dir(thisFile), "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{"StrictOrderbookTop": true, "StrictLastPrice": true}
	checked := 0
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", file, err)
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			receiver, ok := selector.X.(*ast.Ident)
			if !ok || receiver.Name != "client" {
				return true
			}
			checked++
			if !allowed[selector.Sel.Name] {
				t.Fatalf("%s calls client.%s; this probe may only read the two strict quote readers",
					filepath.Base(file), selector.Sel.Name)
			}
			return true
		})
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"os.WriteFile", "os.Create", "strategyevidence", "PollQuoteL1", "Append("} {
			if strings.Contains(string(source), forbidden) {
				t.Fatalf("%s spells %q; the probe writes nothing and appends no evidence", filepath.Base(file), forbidden)
			}
		}
	}
	if checked == 0 {
		t.Fatal("the guard found no client call at all")
	}
}
