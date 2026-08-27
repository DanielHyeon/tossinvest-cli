package officialbars

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"go/ast"
	"path/filepath"
	"strings"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

// ---- 시각 기준점 ----
//
// 2026-08-14 금요일 미국 정규장 안의 한순간이다. 호가는 봉과 달리 분 경계에 놓이지 않는다.
var (
	quoteSource   = time.Date(2026, 8, 14, 13, 45, 12, 500*int(time.Millisecond), time.UTC)
	quoteReceived = time.Date(2026, 8, 14, 13, 45, 12, 720*int(time.Millisecond), time.UTC)
)

// ---- 가짜 리더 ----

type quoteReaderCall struct {
	half           string
	market, symbol string
}

type fakeQuoteReader struct {
	top     official.StrictTopOfBook
	topErr  error
	last    official.StrictLastPrice
	lastErr error
	calls   []quoteReaderCall
}

func (r *fakeQuoteReader) StrictOrderbookTop(_ context.Context, market, symbol string) (official.StrictTopOfBook, error) {
	r.calls = append(r.calls, quoteReaderCall{half: "orderbook", market: market, symbol: symbol})
	if r.topErr != nil {
		return official.StrictTopOfBook{}, r.topErr
	}
	return r.top, nil
}

func (r *fakeQuoteReader) StrictLastPrice(_ context.Context, market, symbol string) (official.StrictLastPrice, error) {
	r.calls = append(r.calls, quoteReaderCall{half: "prices", market: market, symbol: symbol})
	if r.lastErr != nil {
		return official.StrictLastPrice{}, r.lastErr
	}
	return r.last, nil
}

func (r *fakeQuoteReader) halves() []string {
	out := make([]string, 0, len(r.calls))
	for _, call := range r.calls {
		out = append(out, call.half)
	}
	return out
}

// ---- 고정 값들 ----

const (
	quoteAskPrice = "231.7000"
	quoteBidPrice = "231.6500"
	quoteLastText = "231.6800"
	orderbookSeal = "orderbook-body"
	pricesSeal    = "prices-body"
)

func usTop(source, readAt time.Time, ask, bid, label string) official.StrictTopOfBook {
	return official.StrictTopOfBook{
		Market: "US", Symbol: usSymbol, Currency: "USD",
		Ask:           official.StrictQuoteLevel{Price: ask, Volume: "160"},
		Bid:           official.StrictQuoteLevel{Price: bid, Volume: "40"},
		SourceInstant: source, ReadAt: readAt, StatusCode: 200, BodyDigest: digestFor(label),
	}
}

func usLast(source, readAt time.Time, last, label string) official.StrictLastPrice {
	return official.StrictLastPrice{
		Market: "US", Symbol: usSymbol, Currency: "USD", Last: last,
		SourceInstant: source, ReadAt: readAt, StatusCode: 200, BodyDigest: digestFor(label),
	}
}

func usQuoteReader() *fakeQuoteReader {
	return &fakeQuoteReader{
		top:  usTop(quoteSource, quoteReceived, quoteAskPrice, quoteBidPrice, orderbookSeal),
		last: usLast(quoteSource, quoteReceived, quoteLastText, pricesSeal),
	}
}

// wantCompositeDigest는 결정 37이 선언한 preimage를 시험이 직접 다시 만든 값이다.
// 생산자의 계산을 베끼지 않고 글로 적힌 규칙에서 유도해야 규칙이 바뀐 것을 잡을 수 있다.
func wantCompositeDigest(t *testing.T, orderbookLabel, pricesLabel string) string {
	t.Helper()
	orderbook := strings.TrimPrefix(digestFor(orderbookLabel), "sha256:")
	prices := strings.TrimPrefix(digestFor(pricesLabel), "sha256:")
	sum := sha256.Sum256([]byte("a112-quote-l1-v1\n" + orderbook + "\n" + prices))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// countingQuoteStore는 적재 시도 횟수를 센다. "한 줄도 적재하지 않았다"를 증명하는 데 쓴다.
type countingQuoteStore struct {
	inner    QuoteStore
	attempts int
}

func (s *countingQuoteStore) Append(ctx context.Context, envelope strategyevidence.Envelope) (strategyevidence.AppendResult, error) {
	s.attempts++
	return s.inner.Append(ctx, envelope)
}

func quotePayloadOf(t *testing.T, store *strategyevidence.Store, at time.Time) []strategyevidence.QuoteL1Payload {
	t.Helper()
	snapshot, err := store.SealSnapshot(context.Background(), strategyevidence.SnapshotQuery{
		Market: marketclock.MarketUS, Symbol: usSymbol,
		IssuerIdentity: "US:" + usSymbol, IssuerMappingVersion: strategyevidence.BarIssuerMappingVersion,
		EvaluationAt: at, IngestionCutoff: at,
	})
	if err != nil {
		t.Fatalf("SealSnapshot: %v", err)
	}
	out := make([]strategyevidence.QuoteL1Payload, 0, len(snapshot.Items))
	for _, item := range snapshot.Items {
		payload, err := strategyevidence.DecodeQuoteL1Payload(item.CanonicalPayload())
		if err != nil {
			t.Fatalf("DecodeQuoteL1Payload: %v", err)
		}
		out = append(out, payload)
	}
	return out
}

// ---- 결정 34·37·40: 두 읽기가 한 봉인이 된다 ----

func TestPollQuoteL1SealsTwoReadsIntoOneRow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	clock := marketclock.NewFake(quoteReceived.Add(time.Second))
	store := openTestStore(t, clock)
	reader := usQuoteReader()

	result, err := PollQuoteL1(ctx, reader, store, QuotePollInput{Market: marketclock.MarketUS, Symbol: usSymbol})
	if err != nil {
		t.Fatalf("PollQuoteL1: %v", err)
	}
	if got := reader.halves(); len(got) != 2 || got[0] != "orderbook" || got[1] != "prices" {
		t.Fatalf("reads = %v, want the orderbook first and exactly one of each", got)
	}
	if result.Outcome != QuoteOutcomeAdmitted || result.Revision != 1 {
		t.Fatalf("result = %+v", result)
	}
	if !result.SourceObservedAt.Equal(quoteSource) || !result.ReceivedAt.Equal(quoteReceived) {
		t.Fatalf("instants = %v / %v", result.SourceObservedAt, result.ReceivedAt)
	}
	if result.SourceResponseDigest != wantCompositeDigest(t, orderbookSeal, pricesSeal) {
		t.Fatalf("digest = %q", result.SourceResponseDigest)
	}

	payloads := quotePayloadOf(t, store, quoteReceived.Add(time.Hour))
	if len(payloads) != 1 {
		t.Fatalf("stored rows = %d, want 1", len(payloads))
	}
	quote := payloads[0]
	if quote.Market != "US" || quote.Symbol != usSymbol || quote.Currency != "USD" || quote.PriceScale != 4 {
		t.Fatalf("identity = %+v", quote)
	}
	if quote.BidMinor != 2316500 || quote.AskMinor != 2317000 || quote.LastMinor != 2316800 {
		t.Fatalf("minors = %d / %d / %d", quote.BidMinor, quote.AskMinor, quote.LastMinor)
	}
	if quote.Raw.Bid != quoteBidPrice || quote.Raw.Ask != quoteAskPrice || quote.Raw.Last != quoteLastText {
		t.Fatalf("raw = %+v", quote.Raw)
	}
	if quote.SourceObservedAtMS != uint64(quoteSource.UnixMilli()) || quote.ReceivedAtMS != uint64(quoteReceived.UnixMilli()) {
		t.Fatalf("payload instants = %d / %d", quote.SourceObservedAtMS, quote.ReceivedAtMS)
	}
	if quote.SourceResponseDigest != result.SourceResponseDigest || quote.Revision != 1 {
		t.Fatalf("payload provenance = %+v", quote)
	}
}

// ---- 결정 36: 오래된 시각과 늦은 영수증, 둘 다 거절하는 방향 ----

func TestPollQuoteL1BindsTheHalvesConservatively(t *testing.T) {
	t.Parallel()
	early := quoteSource
	late := quoteSource.Add(300 * time.Millisecond)
	earlyRead := quoteReceived
	lateRead := quoteReceived.Add(400 * time.Millisecond)
	for _, testCase := range []struct {
		name                       string
		topSource, lastSource      time.Time
		topRead, lastRead          time.Time
		wantObserved, wantReceived time.Time
	}{
		{"orderbook is older and read first", early, late, earlyRead, lateRead, early, lateRead},
		{"prices is older and read first", late, early, lateRead, earlyRead, early, lateRead},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			clock := marketclock.NewFake(lateRead.Add(time.Second))
			store := openTestStore(t, clock)
			reader := &fakeQuoteReader{
				top:  usTop(testCase.topSource, testCase.topRead, quoteAskPrice, quoteBidPrice, orderbookSeal),
				last: usLast(testCase.lastSource, testCase.lastRead, quoteLastText, pricesSeal),
			}
			result, err := PollQuoteL1(context.Background(), reader, store,
				QuotePollInput{Market: marketclock.MarketUS, Symbol: usSymbol})
			if err != nil {
				t.Fatalf("PollQuoteL1: %v", err)
			}
			if !result.SourceObservedAt.Equal(testCase.wantObserved) {
				t.Fatalf("source observed = %v, want the older of the two source instants (%v)",
					result.SourceObservedAt, testCase.wantObserved)
			}
			if !result.ReceivedAt.Equal(testCase.wantReceived) {
				t.Fatalf("received = %v, want the later of the two read instants (%v)",
					result.ReceivedAt, testCase.wantReceived)
			}
		})
	}
}

// ---- 결정 34: 한쪽이 실패하면 아무것도 적재하지 않는다 ----

func TestPollQuoteL1AppendsNothingWhenEitherHalfFails(t *testing.T) {
	t.Parallel()
	broken := errors.New("broker said no")
	for _, testCase := range []struct {
		name      string
		reader    *fakeQuoteReader
		wantCalls int
	}{
		{"orderbook fails", &fakeQuoteReader{topErr: broken, last: usLast(quoteSource, quoteReceived, quoteLastText, pricesSeal)}, 1},
		{"prices fails", &fakeQuoteReader{top: usTop(quoteSource, quoteReceived, quoteAskPrice, quoteBidPrice, orderbookSeal), lastErr: broken}, 2},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			clock := marketclock.NewFake(quoteReceived.Add(time.Second))
			store := &countingQuoteStore{inner: openTestStore(t, clock)}
			_, err := PollQuoteL1(context.Background(), testCase.reader, store,
				QuotePollInput{Market: marketclock.MarketUS, Symbol: usSymbol})
			if err == nil {
				t.Fatal("a failed half still produced a quote")
			}
			if refusal := refusalOf(t, err); refusal.Reason != RefusalQuoteReaderError {
				t.Fatalf("reason = %q", refusal.Reason)
			}
			if !errors.Is(err, broken) {
				t.Fatalf("the reader's own error was dropped: %v", err)
			}
			if store.attempts != 0 {
				t.Fatalf("appends = %d, want 0", store.attempts)
			}
			// 실패한 절반은 다시 부르지 않는다. 재시도는 두 절반 사이의 틈을
			// 보이지 않게 벌린다(결정 34).
			if len(testCase.reader.calls) != testCase.wantCalls {
				t.Fatalf("reader calls = %v", testCase.reader.halves())
			}
		})
	}
}

// ---- 교차한 호가와 계약 밖 소수는 적재 전에 거절된다 ----

func TestPollQuoteL1RefusesQuotesTheContractCannotHold(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name           string
		ask, bid, last string
		market         marketclock.Market
		symbol         string
		currency       string
	}{
		{name: "bid above ask", ask: "231.7000", bid: "231.7500", last: quoteLastText},
		{name: "us raw with five decimals", ask: "231.70001", bid: quoteBidPrice, last: quoteLastText},
		{name: "us last with five decimals", ask: quoteAskPrice, bid: quoteBidPrice, last: "231.68001"},
		{name: "raw is not a decimal", ask: "231,70", bid: quoteBidPrice, last: quoteLastText},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			clock := marketclock.NewFake(quoteReceived.Add(time.Second))
			store := &countingQuoteStore{inner: openTestStore(t, clock)}
			reader := &fakeQuoteReader{
				top:  usTop(quoteSource, quoteReceived, testCase.ask, testCase.bid, orderbookSeal),
				last: usLast(quoteSource, quoteReceived, testCase.last, pricesSeal),
			}
			_, err := PollQuoteL1(context.Background(), reader, store,
				QuotePollInput{Market: marketclock.MarketUS, Symbol: usSymbol})
			if err == nil {
				t.Fatal("a quote outside the contract was admitted")
			}
			if refusal := refusalOf(t, err); refusal.Reason != RefusalQuoteEnvelope {
				t.Fatalf("reason = %q", refusal.Reason)
			}
			if store.attempts != 0 {
				t.Fatalf("appends = %d, want 0", store.attempts)
			}
			// 같은 입력을 증거 생성자에게 직접 줘도 거절이어야 한다. 두 문이 같은
			// 규칙을 지키는지 확인하는 것이지, 한 문만 믿는 것이 아니다.
			if _, err := strategyevidence.NewQuoteL1Envelope(strategyevidence.QuoteL1Input{
				Market: marketclock.MarketUS, Symbol: usSymbol,
				SourceObservedAt: quoteSource, ReceivedAt: quoteReceived, Revision: 1, Currency: "USD",
				Raw:                  strategyevidence.RawQuoteL1{Bid: testCase.bid, Ask: testCase.ask, Last: testCase.last},
				SourceResponseDigest: wantCompositeDigest(t, orderbookSeal, pricesSeal),
			}); err == nil {
				t.Fatal("NewQuoteL1Envelope accepted what the producer refused")
			}
		})
	}
}

func TestPollQuoteL1RefusesFractionalKRW(t *testing.T) {
	t.Parallel()
	clock := marketclock.NewFake(quoteReceived.Add(time.Second))
	store := &countingQuoteStore{inner: openTestStore(t, clock)}
	reader := &fakeQuoteReader{
		top: official.StrictTopOfBook{
			Market: "KR", Symbol: krSymbol, Currency: "KRW",
			Ask:           official.StrictQuoteLevel{Price: "284000.5", Volume: "10"},
			Bid:           official.StrictQuoteLevel{Price: "283500", Volume: "11"},
			SourceInstant: quoteSource, ReadAt: quoteReceived, StatusCode: 200, BodyDigest: digestFor(orderbookSeal),
		},
		last: official.StrictLastPrice{
			Market: "KR", Symbol: krSymbol, Currency: "KRW", Last: "283900",
			SourceInstant: quoteSource, ReadAt: quoteReceived, StatusCode: 200, BodyDigest: digestFor(pricesSeal),
		},
	}
	_, err := PollQuoteL1(context.Background(), reader, store,
		QuotePollInput{Market: marketclock.MarketKR, Symbol: krSymbol})
	if err == nil {
		t.Fatal("a fractional KRW price was admitted")
	}
	if refusal := refusalOf(t, err); refusal.Reason != RefusalQuoteEnvelope {
		t.Fatalf("reason = %q", refusal.Reason)
	}
	if store.attempts != 0 {
		t.Fatalf("appends = %d, want 0", store.attempts)
	}
}

// ---- 결정 40: 같은 순간을 다시 읽으면 ----

func TestPollQuoteL1IsIdempotentOnAnIdenticalReRead(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	clock := marketclock.NewFake(quoteReceived.Add(time.Second))
	store := openTestStore(t, clock)
	first, err := PollQuoteL1(ctx, usQuoteReader(), store, QuotePollInput{Market: marketclock.MarketUS, Symbol: usSymbol})
	if err != nil || first.Outcome != QuoteOutcomeAdmitted {
		t.Fatalf("first poll = %+v, %v", first, err)
	}
	second, err := PollQuoteL1(ctx, usQuoteReader(), store, QuotePollInput{Market: marketclock.MarketUS, Symbol: usSymbol})
	if err != nil {
		t.Fatalf("second poll: %v", err)
	}
	if second.Outcome != QuoteOutcomeUnchanged {
		t.Fatalf("second outcome = %q, want %q", second.Outcome, QuoteOutcomeUnchanged)
	}
	count, err := store.EvidenceCount(ctx)
	if err != nil {
		t.Fatalf("EvidenceCount: %v", err)
	}
	if count != 1 {
		t.Fatalf("stored rows = %d, want 1", count)
	}
}

func TestPollQuoteL1QuarantinesADifferentBodyAtTheSameInstant(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	clock := marketclock.NewFake(quoteReceived.Add(time.Second))
	store := openTestStore(t, clock)
	if _, err := PollQuoteL1(ctx, usQuoteReader(), store, QuotePollInput{Market: marketclock.MarketUS, Symbol: usSymbol}); err != nil {
		t.Fatalf("first poll: %v", err)
	}
	rival := &fakeQuoteReader{
		top:  usTop(quoteSource, quoteReceived, quoteAskPrice, quoteBidPrice, "other-orderbook-body"),
		last: usLast(quoteSource, quoteReceived, "231.6900", "other-prices-body"),
	}
	second, err := PollQuoteL1(ctx, rival, store, QuotePollInput{Market: marketclock.MarketUS, Symbol: usSymbol})
	if err != nil {
		t.Fatalf("a conflict must be reported, not returned as an error: %v", err)
	}
	if second.Outcome != QuoteOutcomeConflict {
		t.Fatalf("outcome = %q, want %q", second.Outcome, QuoteOutcomeConflict)
	}
	quarantined, err := store.QuarantineCount(ctx)
	if err != nil {
		t.Fatalf("QuarantineCount: %v", err)
	}
	if quarantined != 1 {
		t.Fatalf("quarantined = %d, want 1", quarantined)
	}
	payloads := quotePayloadOf(t, store, quoteReceived.Add(time.Hour))
	if len(payloads) != 1 || payloads[0].Raw.Last != quoteLastText {
		t.Fatalf("the first revision was not kept: %+v", payloads)
	}
}

// ---- 두 절반이 서로 다른 것을 말하면 거절한다 ----

func TestPollQuoteL1RefusesHalvesThatDisagree(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name   string
		mutate func(reader *fakeQuoteReader)
		reason string
	}{
		{"orderbook answers another symbol", func(reader *fakeQuoteReader) { reader.top.Symbol = "MSFT" }, RefusalQuoteIdentity},
		{"prices answers another symbol", func(reader *fakeQuoteReader) { reader.last.Symbol = "MSFT" }, RefusalQuoteIdentity},
		{"orderbook answers another market", func(reader *fakeQuoteReader) { reader.top.Market = "KR" }, RefusalQuoteIdentity},
		{"prices answers another currency", func(reader *fakeQuoteReader) { reader.last.Currency = "KRW" }, RefusalQuoteCurrency},
		{"orderbook carries no source instant", func(reader *fakeQuoteReader) { reader.top.SourceInstant = time.Time{} }, RefusalQuoteInstant},
		{"prices carries no source instant", func(reader *fakeQuoteReader) { reader.last.SourceInstant = time.Time{} }, RefusalQuoteInstant},
		{"orderbook carries no read instant", func(reader *fakeQuoteReader) { reader.top.ReadAt = time.Time{} }, RefusalQuoteInstant},
		{"a body digest is malformed", func(reader *fakeQuoteReader) { reader.last.BodyDigest = "nothex" }, RefusalQuoteDigest},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			clock := marketclock.NewFake(quoteReceived.Add(time.Second))
			store := &countingQuoteStore{inner: openTestStore(t, clock)}
			reader := usQuoteReader()
			testCase.mutate(reader)
			_, err := PollQuoteL1(context.Background(), reader, store,
				QuotePollInput{Market: marketclock.MarketUS, Symbol: usSymbol})
			if err == nil {
				t.Fatal("two halves that disagree were sealed together")
			}
			if refusal := refusalOf(t, err); refusal.Reason != testCase.reason {
				t.Fatalf("reason = %q, want %q", refusal.Reason, testCase.reason)
			}
			if store.attempts != 0 {
				t.Fatalf("appends = %d, want 0", store.attempts)
			}
		})
	}
}

func TestPollQuoteL1RefusesBadArgumentsBeforeAnyRead(t *testing.T) {
	t.Parallel()
	clock := marketclock.NewFake(quoteReceived.Add(time.Second))
	store := openTestStore(t, clock)
	if _, err := PollQuoteL1(context.Background(), nil, store, QuotePollInput{Market: marketclock.MarketUS, Symbol: usSymbol}); err == nil {
		t.Fatal("a missing reader was accepted")
	} else if refusalOf(t, err).Reason != RefusalDependencyMissing {
		t.Fatalf("reason = %q", refusalOf(t, err).Reason)
	}
	if _, err := PollQuoteL1(context.Background(), usQuoteReader(), nil, QuotePollInput{Market: marketclock.MarketUS, Symbol: usSymbol}); err == nil {
		t.Fatal("a missing store was accepted")
	}
	for _, testCase := range []struct {
		name   string
		input  QuotePollInput
		reason string
	}{
		{"unknown market", QuotePollInput{Market: marketclock.Market("JP"), Symbol: usSymbol}, RefusalMarketInvalid},
		{"empty symbol", QuotePollInput{Market: marketclock.MarketUS, Symbol: "  "}, RefusalSymbolInvalid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			reader := usQuoteReader()
			_, err := PollQuoteL1(context.Background(), reader, store, testCase.input)
			if err == nil {
				t.Fatal("a bad argument was accepted")
			}
			if refusal := refusalOf(t, err); refusal.Reason != testCase.reason {
				t.Fatalf("reason = %q, want %q", refusal.Reason, testCase.reason)
			}
			if len(reader.calls) != 0 {
				t.Fatalf("the broker was read %d times for an argument the producer must refuse", len(reader.calls))
			}
		})
	}
}

// ---- 정적 가드 ----

// TestProductionNeverTouchesFloatingPoint는 이 꾸러미가 소수를 부동소수로 바꾸지 못하게 한다.
// 가격은 브로커가 보낸 문자열로만 다니고, 정수 minor는 strategyevidence가 만든다.
func TestProductionNeverTouchesFloatingPoint(t *testing.T) {
	t.Parallel()
	forbidden := map[string]bool{"float64": true, "float32": true, "ParseFloat": true, "FormatFloat": true}
	checked := 0
	for _, file := range productionFiles(t) {
		ast.Inspect(parseFile(t, file), func(node ast.Node) bool {
			identifier, ok := node.(*ast.Ident)
			if !ok {
				return true
			}
			checked++
			if forbidden[identifier.Name] {
				t.Fatalf("%s names %s; prices stay decimal strings in this package", filepath.Base(file), identifier.Name)
			}
			return true
		})
	}
	if checked == 0 {
		t.Fatal("the guard inspected no identifier at all")
	}
}

// TestPollQuoteL1RefusesAQuoteReceivedBeforeItWasObserved는 브로커 시각이 우리 읽은
// 시각보다 뒤에 있는 경우(시계 어긋남)를 붙잡는다. 생산자는 값을 끌어당겨 맞추지 않는다 —
// 끌어당기면 "받기 전에 관측된" 호가가 조용히 정상으로 보이고, 신선도 거부가 그만큼 헐거워진다.
func TestPollQuoteL1RefusesAQuoteReceivedBeforeItWasObserved(t *testing.T) {
	t.Parallel()
	ahead := quoteReceived.Add(2 * time.Second)
	clock := marketclock.NewFake(ahead.Add(time.Second))
	store := &countingQuoteStore{inner: openTestStore(t, clock)}
	reader := &fakeQuoteReader{
		top:  usTop(ahead, quoteReceived, quoteAskPrice, quoteBidPrice, orderbookSeal),
		last: usLast(ahead, quoteReceived, quoteLastText, pricesSeal),
	}
	_, err := PollQuoteL1(context.Background(), reader, store,
		QuotePollInput{Market: marketclock.MarketUS, Symbol: usSymbol})
	if err == nil {
		t.Fatal("a quote whose broker instant is after the read instant was admitted")
	}
	if refusal := refusalOf(t, err); refusal.Reason != RefusalQuoteEnvelope {
		t.Fatalf("reason = %q", refusal.Reason)
	}
	if store.attempts != 0 {
		t.Fatalf("appends = %d, want 0", store.attempts)
	}
}
