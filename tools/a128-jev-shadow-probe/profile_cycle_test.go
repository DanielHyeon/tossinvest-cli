package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

type localRoundTrip func(*http.Request) (*http.Response, error)

func (f localRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func handlerTransport(handler http.HandlerFunc) *http.Client {
	return &http.Client{Transport: localRoundTrip(func(r *http.Request) (*http.Response, error) {
		w := httptest.NewRecorder()
		handler(w, r)
		return w.Result(), nil
	})}
}
func TestCompareCycleSharesCollectionAndLabels(t *testing.T) {
	m := &officialMock{t: t, prices: map[string]string{"005930": "70000"}, rateLimited: map[string]int{}, failStatus: map[string]int{}, calls: map[string]int{}, methods: map[string]bool{}, queries: map[string]string{}, fxBase: "USD", fxQuote: "KRW", fxRate: "1250", now: at(10, 0)}
	m.setRanking(rankingRow{symbol: "005930", price: "70000", currency: "KRW"})
	src := official.New(official.Credentials{APIKey: "mock", SecretKey: "mock"}, filepath.Join(t.TempDir(), "token"), official.WithBaseURL("http://fixture"), official.WithHTTPClient(handlerTransport(m.serve)))
	var requests []systemOneRequest
	judge, _ := newJudgeClient("fake", "http://fixture", "")
	judge.hc = handlerTransport(func(w http.ResponseWriter, r *http.Request) {
		var req systemOneRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		requests = append(requests, req)
		answers := map[string]any{}
		v := 0.2
		if len(requests)%2 == 0 {
			v = 0.8
		}
		for k := range req.Questions {
			answers[k] = map[string]any{"type": "noul", "noul": v}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"model": defaultJudgeModel, "answers": answers})
	})
	judge.sleep = func(context.Context, time.Duration) error { return nil }
	dir := t.TempDir()
	l, err := openLedger(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	p := newProbe(src, judge, l, config{InputProfile: "compare", Cohort: "fixture", TopK: 1, Pool: 1, Horizon: time.Hour, Interval: 10 * time.Minute, CapSource: fixedCap(500000, "KRW")})
	clock := &fakeClock{now: at(10, 0)}
	p.now, p.sleep = clock.Now, clock.Sleep
	p.seed = func() uint64 { return 42 }
	p.logf = func(string, ...any) {}
	if err := p.cycle(context.Background(), "KR", session{Open: at(9, 0), Close: at(15, 30), Source: sessionSourceCalendar}, at(10, 0)); err != nil {
		t.Fatal(err)
	}
	c, err := readLedger(filepath.Join(dir, ledgerFileName))
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != 2 || len(c.Judgments) != len(p.questions)*2 {
		t.Fatalf("requests %d judgments %d gaps %+v", len(requests), len(c.Judgments), c.Gaps)
	}
	if requests[1].State.Intraday != nil || requests[1].State.News != newsAbsent {
		t.Fatal("initial state leaked extended data")
	}
	if !reflect.DeepEqual(requests[0].Questions, requests[1].Questions) || requests[0].Model != requests[1].Model {
		t.Fatal("question or model differs")
	}
	if requests[0].State.Intraday == nil || requests[0].State.Daily == nil {
		t.Fatal("fixture lacks current extended data")
	}
	raw, err := json.Marshal(requests[1].State)
	if err != nil {
		t.Fatal(err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"intraday", "daily", "market_context", "warning_types", "price_limits", "investor_flows_per_stock", "news_headlines"} {
		if _, ok := keys[key]; ok {
			t.Fatalf("initial request retained %s", key)
		}
	}

	if !c.Judgments[0].Screen.Finalist || c.Judgments[len(p.questions)].Screen.Finalist {
		t.Fatal("screening was not profile-specific")
	}
	if p.pendingCount() != 1 || len(p.pending[0].JudgmentIDs) != len(p.questions)*2 {
		t.Fatal("labels not shared")
	}
	if c.Judgments[0].SourceStateDigest == "" || c.Judgments[0].SourceStateDigest != c.Judgments[len(p.questions)].SourceStateDigest || c.Judgments[0].ID == c.Judgments[len(p.questions)].ID {
		t.Fatal("provenance mismatch")
	}
	p.cfg.Cohort = "second"
	m.setRanking(rankingRow{symbol: "005930", price: "80000", currency: "KRW"})
	if err := p.cycle(context.Background(), "KR", session{Open: at(9, 0), Close: at(15, 30), Source: sessionSourceCalendar}, at(10, 0)); err != nil {
		t.Fatal(err)
	}
	reopened, err := readLedger(filepath.Join(dir, ledgerFileName))
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, row := range reopened.Judgments {
		if seen[row.ID] {
			t.Fatalf("duplicate ID %s", row.ID)
		}
		seen[row.ID] = true
	}
	restored := newProbe(src, judge, l, p.cfg)
	restored.restorePending(reopened)
	if restored.pendingCount() != 2 {
		t.Fatalf("restored %d pending groups", restored.pendingCount())
	}
	for _, group := range restored.pending {
		if len(group.JudgmentIDs) != len(p.questions)*2 {
			t.Fatalf("restored %d IDs", len(group.JudgmentIDs))
		}
	}
	price := 90000.0
	for _, group := range restored.pending {
		if err := restored.writeLabels(group, at(11, 0), &price, ""); err != nil {
			t.Fatal(err)
		}
	}
	labeled, err := readLedger(filepath.Join(dir, ledgerFileName))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range reopened.Judgments {
		got := labeled.Labels[row.ID].RealizedReturn
		if got == nil {
			t.Fatal("missing label")
		}
		want := round(price/row.BasePrice-1, 8)
		if *got != want {
			t.Fatalf("cohort %s return %v want %v", row.Cohort, *got, want)
		}
	}

	var logs []string
	p.cfg.InputProfile = "initial"
	p.cfg.Cohort = "third"
	p.logf = func(f string, args ...any) { logs = append(logs, fmt.Sprintf(f, args...)) }
	if err := p.cycle(context.Background(), "KR", session{Open: at(9, 0), Close: at(15, 30), Source: sessionSourceCalendar}, at(10, 0)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(logs, "\n"), "judged 1 · finalists 1") {
		t.Fatalf("initial log incorrect: %v", logs)
	}
}

func TestCompareReportRejectsMismatchesAndDuplicates(t *testing.T) {
	base := judgmentRow{Cohort: "c", SourceStateDigest: "sha256:" + strings.Repeat("a", 64), StateDigest: "sha256:" + strings.Repeat("b", 64), Market: "KR", Symbol: "X", At: "2026-10-01T00:00:00Z", QuestionID: screenRedFlagQuestion, QuestionRev: "v", QuestionDigest: "d", ModelRequested: "m", ModelAnswered: "m", HorizonMinutes: 60, BasePrice: 1}
	current := base
	current.InputProfile = "current"
	current.P = .59
	initial := base
	initial.InputProfile = "initial"
	initial.P = .61
	duplicate := base
	duplicate.Symbol = "Y"
	duplicate.InputProfile = "current"
	mismatch := base
	mismatch.Symbol = "Z"
	mismatch.InputProfile = "initial"
	mismatch.ModelAnswered = "different"
	out := renderProfileComparison(ledgerContents{Judgments: []judgmentRow{current, initial, duplicate, duplicate, mismatch}, Labels: map[string]labelRow{}})
	for _, part := range []string{"Paired question rows: 1", "unpaired: 1", "duplicate groups: 1", "0.020000", "crossings: 1 / 1"} {
		if !strings.Contains(out, part) {
			t.Fatalf("missing %q in %s", part, out)
		}
	}

}

func TestBothProfileFailuresKeepSourceAndAreCounted(t *testing.T) {
	m := &officialMock{t: t, prices: map[string]string{"005930": "70000"}, rateLimited: map[string]int{}, failStatus: map[string]int{}, calls: map[string]int{}, methods: map[string]bool{}, queries: map[string]string{}, fxBase: "USD", fxQuote: "KRW", fxRate: "1250", now: at(10, 0)}
	m.setRanking(rankingRow{symbol: "005930", price: "70000", currency: "KRW"})
	src := official.New(official.Credentials{APIKey: "mock", SecretKey: "mock"}, filepath.Join(t.TempDir(), "token"), official.WithBaseURL("http://fixture"), official.WithHTTPClient(handlerTransport(m.serve)))
	judge, _ := newJudgeClient("fake", "http://fixture", "")
	judge.hc = handlerTransport(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnprocessableEntity) })
	dir := t.TempDir()
	l, err := openLedger(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	p := newProbe(src, judge, l, config{InputProfile: "compare", Cohort: "failure", TopK: 1, Pool: 1, Horizon: time.Hour, Interval: 10 * time.Minute, CapSource: fixedCap(500000, "KRW")})
	clock := &fakeClock{now: at(10, 0)}
	p.now, p.sleep = clock.Now, clock.Sleep
	p.seed = func() uint64 { return 42 }
	p.logf = func(string, ...any) {}
	if err := p.cycle(context.Background(), "KR", session{Open: at(9, 0), Close: at(15, 30), Source: sessionSourceCalendar}, at(10, 0)); err != nil {
		t.Fatal(err)
	}
	c, err := readLedger(filepath.Join(dir, ledgerFileName))
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Judgments) != 0 || len(c.Gaps) != 2 || c.Gaps[0].SourceStateDigest == "" || c.Gaps[0].SourceStateDigest != c.Gaps[1].SourceStateDigest {
		t.Fatalf("bad failure evidence: %+v", c)
	}
	if !strings.Contains(renderProfileComparison(c), "both requests failed: 1") {
		t.Fatal(renderProfileComparison(c))
	}
}

func TestCompareReportDistinguishesLabelsAndInvalidRows(t *testing.T) {
	base := judgmentRow{Cohort: "c", SourceStateDigest: "sha256:" + strings.Repeat("a", 64), StateDigest: "sha256:" + strings.Repeat("b", 64), Market: "KR", Symbol: "X", At: "2026-10-01T00:00:00Z", QuestionID: screenRedFlagQuestion, QuestionRev: "v", QuestionDigest: "d", ModelRequested: "m", ModelAnswered: "m", HorizonMinutes: 60, BasePrice: 1}
	a := base
	a.ID = "a"
	a.InputProfile = "current"
	b := base
	b.ID = "b"
	b.InputProfile = "initial"
	c := base
	c.ID = "c"
	c.Symbol = "Y"
	c.InputProfile = "current"
	d := base
	d.ID = "d"
	d.Symbol = "Z"
	d.InputProfile = "initial"
	d.QuestionRev = ""
	e := base
	e.ID = "e"
	e.InputProfile = "initial"
	e.Cohort = ""
	ret := .1
	out := renderProfileComparison(ledgerContents{Judgments: []judgmentRow{a, b, c, d, e}, Labels: map[string]labelRow{"a": {RealizedReturn: &ret}, "b": {}}})
	for _, part := range []string{"Invalid judgment rows: 2;", "current: judgments 2; realized 1; label gaps 0; pending 1", "initial: judgments 3; realized 0; label gaps 1; pending 2"} {
		if !strings.Contains(out, part) {
			t.Fatalf("missing %q: %s", part, out)
		}
	}
}

func TestOneProfileFailureKeepsOtherResult(t *testing.T) {
	m := &officialMock{t: t, prices: map[string]string{"005930": "70000"}, rateLimited: map[string]int{}, failStatus: map[string]int{}, calls: map[string]int{}, methods: map[string]bool{}, queries: map[string]string{}, fxBase: "USD", fxQuote: "KRW", fxRate: "1250", now: at(10, 0)}
	m.setRanking(rankingRow{symbol: "005930", price: "70000", currency: "KRW"})
	src := official.New(official.Credentials{APIKey: "mock", SecretKey: "mock"}, filepath.Join(t.TempDir(), "token"), official.WithBaseURL("http://fixture"), official.WithHTTPClient(handlerTransport(m.serve)))
	judge, _ := newJudgeClient("fake", "http://fixture", "")
	calls := 0
	judge.hc = handlerTransport(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		var req systemOneRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		answers := map[string]any{}
		for k := range req.Questions {
			answers[k] = map[string]any{"type": "noul", "noul": 0.2}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"model": defaultJudgeModel, "answers": answers})
	})
	dir := t.TempDir()
	l, err := openLedger(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	p := newProbe(src, judge, l, config{InputProfile: "compare", Cohort: "failure", TopK: 1, Pool: 1, Horizon: time.Hour, Interval: 10 * time.Minute, CapSource: fixedCap(500000, "KRW")})
	clock := &fakeClock{now: at(10, 0)}
	p.now, p.sleep = clock.Now, clock.Sleep
	p.seed = func() uint64 { return 42 }
	p.logf = func(string, ...any) {}
	if err := p.cycle(context.Background(), "KR", session{Open: at(9, 0), Close: at(15, 30), Source: sessionSourceCalendar}, at(10, 0)); err != nil {
		t.Fatal(err)
	}
	c, err := readLedger(filepath.Join(dir, ledgerFileName))
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Judgments) != len(p.questions) || len(c.Gaps) != 1 || c.Gaps[0].InputProfile != "current" || c.Judgments[0].InputProfile != "initial" || c.Gaps[0].SourceStateDigest != c.Judgments[0].SourceStateDigest {
		t.Fatalf("failure lost: %+v", c)
	}
	if !strings.Contains(renderProfileComparison(c), "one failed: 1") {
		t.Fatal(renderProfileComparison(c))
	}
}

func TestComparisonRejectsTamperedStoredState(t *testing.T) {
	dir := t.TempDir()
	l, err := openLedger(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	digest, err := l.storeState(PublicState{Market: "KR", Symbol: "X", News: newsAbsent, MissingInputs: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	row := judgmentRow{Cohort: "c", SourceStateDigest: digest, StateDigest: digest, Market: "KR", Symbol: "X", At: "2026-10-01T00:00:00Z", QuestionID: screenRedFlagQuestion, QuestionRev: "v", QuestionDigest: "d", ModelRequested: "m", ModelAnswered: "m", HorizonMinutes: 60, BasePrice: 1}
	a := row
	a.InputProfile = "current"
	b := row
	b.InputProfile = "initial"
	contents := ledgerContents{Judgments: []judgmentRow{a, b}, Labels: map[string]labelRow{}}
	if !strings.Contains(renderProfileComparison(verifyComparisonStates(contents, filepath.Join(dir, statesDirName))), "Paired question rows: 1") {
		t.Fatal("valid pair rejected")
	}
	if err := os.WriteFile(filepath.Join(dir, statesDirName, digest[7:]+".json"), []byte("{}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(renderProfileComparison(verifyComparisonStates(contents, filepath.Join(dir, statesDirName))), "Invalid judgment rows: 2;") {
		t.Fatal("tampering accepted")
	}
}

func TestDefaultAndExplicitCurrentProduceSameRows(t *testing.T) {
	m := &officialMock{t: t, prices: map[string]string{"005930": "70000"}, rateLimited: map[string]int{}, failStatus: map[string]int{}, calls: map[string]int{}, methods: map[string]bool{}, queries: map[string]string{}, fxBase: "USD", fxQuote: "KRW", fxRate: "1250", now: at(10, 0)}
	m.setRanking(rankingRow{symbol: "005930", price: "70000", currency: "KRW"})
	src := official.New(official.Credentials{APIKey: "mock", SecretKey: "mock"}, filepath.Join(t.TempDir(), "token"), official.WithBaseURL("http://fixture"), official.WithHTTPClient(handlerTransport(m.serve)))
	var requests [][]byte
	judge, _ := newJudgeClient("fake", "http://fixture", "")
	judge.hc = handlerTransport(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		requests = append(requests, raw)
		var req systemOneRequest
		_ = json.Unmarshal(raw, &req)
		answers := map[string]any{}
		for k := range req.Questions {
			answers[k] = map[string]any{"type": "noul", "noul": 0.2}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"model": defaultJudgeModel, "answers": answers})
	})
	var rows [][]judgmentRow
	for _, profile := range []string{"", "current"} {
		dir := t.TempDir()
		l, err := openLedger(dir)
		if err != nil {
			t.Fatal(err)
		}
		p := newProbe(src, judge, l, config{InputProfile: profile, TopK: 1, Pool: 1, Horizon: time.Hour, Interval: 10 * time.Minute, CapSource: fixedCap(500000, "KRW")})
		clock := &fakeClock{now: at(10, 0)}
		p.now, p.sleep = clock.Now, clock.Sleep
		p.seed = func() uint64 { return 42 }
		p.logf = func(string, ...any) {}
		if err := p.cycle(context.Background(), "KR", session{Open: at(9, 0), Close: at(15, 30), Source: sessionSourceCalendar}, at(10, 0)); err != nil {
			t.Fatal(err)
		}
		contents, err := readLedger(filepath.Join(dir, ledgerFileName))
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, contents.Judgments)
		l.Close()
	}
	if len(requests) != 2 || string(requests[0]) != string(requests[1]) || !reflect.DeepEqual(rows[0], rows[1]) {
		t.Fatal("default and explicit current differ")
	}
}
