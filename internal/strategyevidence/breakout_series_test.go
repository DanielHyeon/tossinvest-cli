package strategyevidence

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestSealBarSeriesReturnsOrderedBarsAndDeterministicDigest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	clk := marketclock.NewFake(usTestBarOpenAt.Add(30 * time.Minute))
	store := openTestStoreWithClock(t, clk)
	ingest := usTestBarOpenAt.Add(30 * time.Minute)
	// 일부러 시간 순서를 섞어서 넣는다. 읽을 때 open_at 오름차순으로 정렬되어야 한다.
	for _, offset := range []time.Duration{2 * time.Minute, 0, time.Minute} {
		appendTestBar(t, store, clk, usTestBarOpenAt.Add(offset), 1, ingest, nil, nil)
	}

	query := usTestBarSeriesQuery(usTestBarOpenAt.Add(10*time.Minute), ingest)
	series, err := store.SealBarSeries(ctx, query)
	if err != nil {
		t.Fatalf("SealBarSeries: %v", err)
	}
	if len(series.Bars) != 3 {
		t.Fatalf("bars = %d, want 3: %+v", len(series.Bars), series.Bars)
	}
	for index, bar := range series.Bars {
		want := uint64(usTestBarOpenAt.Add(time.Duration(index) * time.Minute).UnixMilli())
		if bar.Payload.OpenAtMS != want {
			t.Fatalf("bar %d open_at = %d, want %d", index, bar.Payload.OpenAtMS, want)
		}
		if bar.EvidenceID == "" || bar.RevisionIdentity != "r1" || len(bar.PayloadDigest) != 64 {
			t.Fatalf("bar %d identity = %+v", index, bar)
		}
		if bar.Payload.SessionID != usTestSessionID || bar.Payload.Symbol != "AAPL" {
			t.Fatalf("bar %d scope = %+v", index, bar.Payload)
		}
	}
	if series.Query.MaxBars != 512 {
		t.Fatalf("query max bars = %d, want the default 512", series.Query.MaxBars)
	}
	if len(series.Digest) != 64 {
		t.Fatalf("series digest = %q", series.Digest)
	}
	replayed, err := store.SealBarSeries(ctx, query)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if replayed.Digest != series.Digest {
		t.Fatalf("series digest is not deterministic: %s != %s", replayed.Digest, series.Digest)
	}
	// 읽기 전용이어야 한다: 스냅샷 표에 아무 줄도 남기지 않는다.
	var snapshotRows int
	if err := store.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM snapshots`).Scan(&snapshotRows); err != nil {
		t.Fatal(err)
	}
	if snapshotRows != 0 {
		t.Fatalf("SealBarSeries wrote %d snapshot rows; it must be SELECT-only", snapshotRows)
	}
}

func TestSealBarSeriesDigestIsDomainSeparatedFromSnapshotDigest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := openTestStore(t)
	evaluationAt := usTestBarOpenAt.Add(10 * time.Minute)
	ingestionCutoff := usTestBarOpenAt.Add(30 * time.Minute)
	series, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(evaluationAt, ingestionCutoff))
	if err != nil {
		t.Fatalf("SealBarSeries: %v", err)
	}
	collision := snapshotDigest(SnapshotQuery{
		Market:               marketclock.MarketUS,
		Symbol:               "AAPL",
		IssuerIdentity:       usTestSessionID,
		IssuerMappingVersion: "60000",
		EvaluationAt:         evaluationAt.UTC(),
		IngestionCutoff:      ingestionCutoff.UTC(),
	}, nil)
	if series.Digest == collision {
		t.Fatal("bar series digest collided with a symbol snapshot digest over the same field values")
	}
}

func TestSealBarSeriesCorrectionIsAppendOnlyAndReplayStable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	firstIngest := usTestBarOpenAt.Add(10 * time.Minute)
	laterIngest := usTestBarOpenAt.Add(20 * time.Minute)
	clk := marketclock.NewFake(firstIngest)
	store := openTestStoreWithClock(t, clk)
	appendTestBar(t, store, clk, usTestBarOpenAt, 1, firstIngest, nil, nil)

	evaluationAt := usTestBarOpenAt.Add(5 * time.Minute)
	before, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(evaluationAt, firstIngest))
	if err != nil {
		t.Fatalf("SealBarSeries before correction: %v", err)
	}
	if len(before.Bars) != 1 || before.Bars[0].Payload.CloseMinor != 2316500 {
		t.Fatalf("series before correction = %+v", before.Bars)
	}

	appendTestBar(t, store, clk, usTestBarOpenAt, 2, laterIngest, nil, func(fields map[string]any) {
		fields["close_minor"] = uint64(2317000)
		fields["raw"].(map[string]any)["close"] = "231.7"
	})

	unchanged, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(evaluationAt, firstIngest))
	if err != nil {
		t.Fatalf("SealBarSeries replay at the earlier cutoff: %v", err)
	}
	if unchanged.Digest != before.Digest {
		t.Fatalf("earlier ingestion cutoff changed after a correction: %s != %s", unchanged.Digest, before.Digest)
	}
	if len(unchanged.Bars) != 1 || unchanged.Bars[0].RevisionIdentity != "r1" || unchanged.Bars[0].Payload.CloseMinor != 2316500 {
		t.Fatalf("correction leaked into the earlier cutoff: %+v", unchanged.Bars)
	}

	after, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(evaluationAt, laterIngest))
	if err != nil {
		t.Fatalf("SealBarSeries at the later cutoff: %v", err)
	}
	if len(after.Bars) != 1 || after.Bars[0].RevisionIdentity != "r2" || after.Bars[0].Payload.CloseMinor != 2317000 {
		t.Fatalf("correction is not visible at the later cutoff: %+v", after.Bars)
	}
	if after.Digest == before.Digest {
		t.Fatal("correction did not change the later series digest")
	}
	if count, err := store.EvidenceCount(ctx); err != nil || count != 2 {
		t.Fatalf("evidence count = %d, %v — correction must be append-only", count, err)
	}
}

// 같은 봉에 두 판(revision)이 동시에 보일 때는 언제나 더 큰 판이 이긴다.
func TestSealBarSeriesPrefersTheLatestVisibleRevision(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ingest := usTestBarOpenAt.Add(10 * time.Minute)
	clk := marketclock.NewFake(ingest)
	store := openTestStoreWithClock(t, clk)
	appendTestBar(t, store, clk, usTestBarOpenAt, 1, ingest, nil, nil)
	// 앞 판을 대체한다고 말하지 않는 더 높은 판. 두 줄 모두 컷오프 아래에서 보인다.
	appendTestBar(t, store, clk, usTestBarOpenAt, 2, ingest, func(header *Header) {
		header.SupersedesRevisionIdentity = ""
	}, func(fields map[string]any) {
		fields["close_minor"] = uint64(2317500)
		fields["raw"].(map[string]any)["close"] = "231.75"
	})

	series, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(usTestBarOpenAt.Add(5*time.Minute), ingest))
	if err != nil {
		t.Fatalf("SealBarSeries: %v", err)
	}
	if len(series.Bars) != 1 {
		t.Fatalf("bars = %d, want 1 winner per bar identity: %+v", len(series.Bars), series.Bars)
	}
	if series.Bars[0].RevisionIdentity != "r2" || series.Bars[0].Payload.CloseMinor != 2317500 {
		t.Fatalf("older revision won: %+v", series.Bars[0])
	}
}

func TestSealBarSeriesDualCutoffIndependence(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	early := usTestBarOpenAt
	late := usTestBarOpenAt.Add(5 * time.Minute)
	clk := marketclock.NewFake(usTestBarOpenAt.Add(20 * time.Minute))
	store := openTestStoreWithClock(t, clk)
	appendTestBar(t, store, clk, late, 1, usTestBarOpenAt.Add(20*time.Minute), nil, nil)
	appendTestBar(t, store, clk, early, 1, usTestBarOpenAt.Add(40*time.Minute), nil, nil)

	availableOnly, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(early.Add(time.Minute), usTestBarOpenAt.Add(40*time.Minute)))
	if err != nil {
		t.Fatalf("SealBarSeries with the source cutoff: %v", err)
	}
	if len(availableOnly.Bars) != 1 || availableOnly.Bars[0].Payload.OpenAtMS != uint64(early.UnixMilli()) {
		t.Fatalf("source cutoff did not bound the series: %+v", availableOnly.Bars)
	}
	ingestedOnly, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(late.Add(5*time.Minute), usTestBarOpenAt.Add(30*time.Minute)))
	if err != nil {
		t.Fatalf("SealBarSeries with the ingestion cutoff: %v", err)
	}
	if len(ingestedOnly.Bars) != 1 || ingestedOnly.Bars[0].Payload.OpenAtMS != uint64(late.UnixMilli()) {
		t.Fatalf("ingestion cutoff did not bound the series: %+v", ingestedOnly.Bars)
	}
}

func TestSealBarSeriesRefusesMoreThanFiveHundredTwelveBars(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ingest := usTestBarOpenAt.Add(600 * time.Minute)
	clk := marketclock.NewFake(ingest)
	store := openTestStoreWithClock(t, clk)
	for index := 0; index < 513; index++ {
		appendTestBar(t, store, clk, usTestBarOpenAt.Add(time.Duration(index)*time.Minute), 1, ingest, nil, nil)
	}

	refused, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(ingest, ingest))
	var refusal *ValidationError
	if !errors.As(err, &refusal) {
		t.Fatalf("513 bars were not refused: err=%v bars=%d", err, len(refused.Bars))
	}
	if len(refused.Bars) != 0 || refused.Digest != "" {
		t.Fatalf("refusal returned a truncated series: %d bars, digest %q", len(refused.Bars), refused.Digest)
	}
	// 512개까지는 그대로 통과해야 한다. 경계가 512라는 뜻이다.
	bounded, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(usTestBarOpenAt.Add(512*time.Minute), ingest))
	if err != nil {
		t.Fatalf("512 bars were refused: %v", err)
	}
	if len(bounded.Bars) != 512 {
		t.Fatalf("bars = %d, want exactly 512", len(bounded.Bars))
	}
}

func TestSealBarSeriesRegularSessionFilter(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ingest := usTestBarOpenAt.Add(30 * time.Minute)
	clk := marketclock.NewFake(ingest)
	store := openTestStoreWithClock(t, clk)
	appendTestBar(t, store, clk, usTestBarOpenAt, 1, ingest, nil, nil)
	appendTestBar(t, store, clk, usTestBarOpenAt.Add(time.Minute), 1, ingest, nil, func(fields map[string]any) {
		fields["regular_session"] = false
	})
	appendTestBar(t, store, clk, usTestBarOpenAt.Add(2*time.Minute), 1, ingest, nil, nil)

	query := usTestBarSeriesQuery(usTestBarOpenAt.Add(10*time.Minute), ingest)
	all, err := store.SealBarSeries(ctx, query)
	if err != nil {
		t.Fatalf("SealBarSeries: %v", err)
	}
	if len(all.Bars) != 3 {
		t.Fatalf("unfiltered bars = %d, want 3", len(all.Bars))
	}
	query.RegularSessionOnly = true
	regular, err := store.SealBarSeries(ctx, query)
	if err != nil {
		t.Fatalf("SealBarSeries regular only: %v", err)
	}
	if len(regular.Bars) != 2 {
		t.Fatalf("regular-session bars = %d, want 2: %+v", len(regular.Bars), regular.Bars)
	}
	for _, bar := range regular.Bars {
		if !bar.Payload.RegularSession {
			t.Fatalf("extended-hours bar crossed the regular-session filter: %+v", bar.Payload)
		}
	}
	if regular.Digest == all.Digest {
		t.Fatal("the regular-session filter is not part of the series digest")
	}
}

func TestSealBarSeriesRefusesInvalidQuery(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := openTestStore(t)
	evaluationAt := usTestBarOpenAt.Add(10 * time.Minute)
	ingestionCutoff := usTestBarOpenAt.Add(30 * time.Minute)
	for _, tt := range []struct {
		name   string
		mutate func(*BarSeriesQuery)
	}{
		{"unknown market", func(q *BarSeriesQuery) { q.Market = "JP" }},
		{"empty market", func(q *BarSeriesQuery) { q.Market = "" }},
		{"empty symbol", func(q *BarSeriesQuery) { q.Symbol = " " }},
		{"empty session", func(q *BarSeriesQuery) { q.SessionID = "" }},
		{"session calendar does not match market", func(q *BarSeriesQuery) { q.SessionID = krTestSessionID }},
		{"interval is not one minute", func(q *BarSeriesQuery) { q.IntervalMS = 30000 }},
		{"max bars above the hard cap", func(q *BarSeriesQuery) { q.MaxBars = 513 }},
		{"negative max bars", func(q *BarSeriesQuery) { q.MaxBars = -1 }},
		{"zero evaluation clock", func(q *BarSeriesQuery) { q.EvaluationAt = time.Time{} }},
		{"zero ingestion cutoff", func(q *BarSeriesQuery) { q.IngestionCutoff = time.Time{} }},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			query := usTestBarSeriesQuery(evaluationAt, ingestionCutoff)
			tt.mutate(&query)
			series, err := store.SealBarSeries(ctx, query)
			var refusal *ValidationError
			if !errors.As(err, &refusal) {
				t.Fatalf("%s: want a typed refusal, got %v", tt.name, err)
			}
			if len(series.Bars) != 0 || series.Digest != "" {
				t.Fatalf("%s: refusal returned a series: %+v", tt.name, series)
			}
		})
	}
}

func TestSealBarSeriesIgnoresOtherKindsSymbolsAndSessions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ingest := usTestBarOpenAt.Add(30 * time.Minute)
	clk := marketclock.NewFake(ingest)
	store := openTestStoreWithClock(t, clk)
	appendTestBar(t, store, clk, usTestBarOpenAt, 1, ingest, nil, nil)

	otherSymbol, err := newClosedBar1mHeader(marketclock.MarketUS, "MSFT", usTestSessionID, usTestCalendarVersion,
		usTestBarOpenAt, 1, usTestBarOpenAt.Add(2*time.Minute), "USD")
	if err != nil {
		t.Fatal(err)
	}
	otherSymbolFields := usTestBarFields(uint64(usTestBarOpenAt.UnixMilli()))
	otherSymbolFields["symbol"] = "MSFT"
	if _, err := store.Append(ctx, mustEnvelopeBytes(t, otherSymbol, barPayloadBytes(t, otherSymbolFields, nil))); err != nil {
		t.Fatal(err)
	}

	previousDay := usTestBarOpenAt.AddDate(0, 0, -1)
	otherSession, err := newClosedBar1mHeader(marketclock.MarketUS, "AAPL", "US:2026-08-13", usTestCalendarVersion,
		previousDay, 1, previousDay.Add(2*time.Minute), "USD")
	if err != nil {
		t.Fatal(err)
	}
	otherSessionFields := usTestBarFields(uint64(previousDay.UnixMilli()))
	otherSessionFields["session_id"] = "US:2026-08-13"
	if _, err := store.Append(ctx, mustEnvelopeBytes(t, otherSession, barPayloadBytes(t, otherSessionFields, nil))); err != nil {
		t.Fatal(err)
	}

	quoteObserved := usTestBarOpenAt.Add(3 * time.Minute)
	quoteHeader, err := newQuoteL1Header(marketclock.MarketUS, "AAPL", quoteObserved, 1, quoteObserved.Add(time.Second), "USD")
	if err != nil {
		t.Fatal(err)
	}
	quoteFields := usTestQuoteFields()
	quoteFields["source_observed_at_ms"] = uint64(quoteObserved.UnixMilli())
	quoteFields["received_at_ms"] = uint64(quoteObserved.UnixMilli()) + 1000
	if _, err := store.Append(ctx, mustEnvelopeBytes(t, quoteHeader, barPayloadBytes(t, quoteFields, nil))); err != nil {
		t.Fatal(err)
	}

	legacy := validHeader(marketclock.MarketUS, KindUSParticipation, "legacy-in-the-same-store", "rev-1")
	if _, err := store.Append(ctx, mustEnvelope(t, legacy, `{"value":1}`)); err != nil {
		t.Fatal(err)
	}

	series, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(usTestBarOpenAt.Add(10*time.Minute), ingest))
	if err != nil {
		t.Fatalf("SealBarSeries: %v", err)
	}
	if len(series.Bars) != 1 || series.Bars[0].Payload.Symbol != "AAPL" || series.Bars[0].Payload.SessionID != usTestSessionID {
		t.Fatalf("series is not scoped to one kind/symbol/session: %+v", series.Bars)
	}
}

func TestClosedBarAppendQuarantinesSameRevisionWithADifferentDigest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ingest := usTestBarOpenAt.Add(30 * time.Minute)
	clk := marketclock.NewFake(ingest)
	store := openTestStoreWithClock(t, clk)
	appendTestBar(t, store, clk, usTestBarOpenAt, 1, ingest, nil, nil)

	conflicting := usTestBarFields(uint64(usTestBarOpenAt.UnixMilli()))
	conflicting["close_minor"] = uint64(2317000)
	conflicting["raw"].(map[string]any)["close"] = "231.7"
	_, err := store.Append(ctx, mustEnvelopeBytes(t, usTestBarHeader(t, 1), barPayloadBytes(t, conflicting, nil)))
	if !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("want ErrRevisionConflict, got %v", err)
	}
	if count, err := store.QuarantineCount(ctx); err != nil || count != 1 {
		t.Fatalf("quarantine count = %d, %v", count, err)
	}
	if count, err := store.EvidenceCount(ctx); err != nil || count != 1 {
		t.Fatalf("evidence count = %d, %v", count, err)
	}
	series, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(usTestBarOpenAt.Add(10*time.Minute), ingest))
	if err != nil {
		t.Fatalf("SealBarSeries: %v", err)
	}
	if len(series.Bars) != 1 || series.Bars[0].Payload.CloseMinor != 2316500 {
		t.Fatalf("quarantined payload reached the series: %+v", series.Bars)
	}
}

// 결합 생성자로 만든 봉이 저장·재생을 거쳐 같은 값으로 돌아오는지 본다.
func TestSealBarSeriesRoundTripsEnvelopesBuiltByTheConstructor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ingest := usTestBarOpenAt.Add(30 * time.Minute)
	clk := marketclock.NewFake(ingest)
	store := openTestStoreWithClock(t, clk)
	input := usTestBarInput()
	built, err := NewClosedBar1mEnvelope(input)
	if err != nil {
		t.Fatalf("NewClosedBar1mEnvelope: %v", err)
	}
	if _, err := store.Append(ctx, built); err != nil {
		t.Fatalf("Append: %v", err)
	}

	series, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(usTestBarOpenAt.Add(10*time.Minute), ingest))
	if err != nil {
		t.Fatalf("SealBarSeries: %v", err)
	}
	if len(series.Bars) != 1 {
		t.Fatalf("bars = %d, want 1", len(series.Bars))
	}
	bar := series.Bars[0]
	if bar.EvidenceID != built.Header().EvidenceID || bar.PayloadDigest != built.PayloadDigest() || bar.RevisionIdentity != "r1" {
		t.Fatalf("stored identity drifted: %+v", bar)
	}
	if bar.Payload.PriceScale != 4 || bar.Payload.CloseMinor != 2316500 || bar.Payload.LowMinor != 2311000 {
		t.Fatalf("stored minors drifted: %+v", bar.Payload)
	}
	if bar.Payload.Raw != input.Raw {
		t.Fatalf("raw strings drifted: %+v want %+v", bar.Payload.Raw, input.Raw)
	}
	if bar.Payload.OpenAtMS != uint64(usTestBarOpenAt.UnixMilli()) || !bar.Payload.RegularSession {
		t.Fatalf("stored bar identity drifted: %+v", bar.Payload)
	}
}

// 머리말은 US, 본문은 KR. NewEnvelope은 둘을 맞대어 볼 수 없으므로 저장은 되고, 읽을 때 거절된다.
// 두 값은 같은 순간(2026-08-18 00:00Z)이라 미국 현지로는 08-17, 한국 현지로는 08-18 세션이다.
func TestSealBarSeriesRefusesAPayloadThatDisagreesWithItsHeader(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sharedOpen := krTestBarOpenAt
	ingest := sharedOpen.Add(30 * time.Minute)
	clk := marketclock.NewFake(ingest)
	store := openTestStoreWithClock(t, clk)
	usHeader, err := newClosedBar1mHeader(marketclock.MarketUS, "AAPL", "US:2026-08-17", usTestCalendarVersion,
		sharedOpen, 1, sharedOpen.Add(2*time.Minute), "USD")
	if err != nil {
		t.Fatalf("US header: %v", err)
	}
	krFields := krTestBarFields(uint64(sharedOpen.UnixMilli()))
	if _, err := store.Append(ctx, mustEnvelopeBytes(t, usHeader, barPayloadBytes(t, krFields, nil))); err != nil {
		t.Fatalf("Append: %v", err)
	}

	query := usTestBarSeriesQuery(sharedOpen.Add(10*time.Minute), ingest)
	query.SessionID = "US:2026-08-17"
	series, err := store.SealBarSeries(ctx, query)
	var refusal *ValidationError
	if !errors.As(err, &refusal) || refusal.Code != RefusalIdentityMismatch {
		t.Fatalf("header/payload mismatch was not refused on read: %v", err)
	}
	if !strings.Contains(refusal.Detail, usHeader.EvidenceID) {
		t.Fatalf("refusal does not name the offending evidence: %q", refusal.Detail)
	}
	if len(series.Bars) != 0 {
		t.Fatalf("refusal returned bars: %+v", series.Bars)
	}
}

// ---- 읽기 쪽 머리말↔본문 결속 (RED 먼저) ----

// 어긋난 줄은 하나하나 "어느 단계에서" 막혀야 하는지까지 시험한다.
// 단계가 밀리면(예: 읽기에서 막히던 것이 조용히 통과하면) 그 자체로 실패다.
type driftStage string

const (
	driftAtEnvelope driftStage = "envelope"
	driftAtAppend   driftStage = "append"
	driftAtRead     driftStage = "read"
)

func TestSealBarSeriesRefusesEveryHeaderPayloadDrift(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		stage driftStage
		field string
		build func(t *testing.T) (Header, []byte)
	}{
		{"payload minute drifts earlier while every other clock agrees", driftAtRead, "open_at_ms",
			driftBuilder(nil, func(fields map[string]any) {
				earlier := uint64(usTestBarOpenAt.Add(-6 * time.Hour).UnixMilli())
				fields["open_at_ms"] = earlier
				fields["successor_open_at_ms"] = earlier + 60000
			})},
		{"payload observation instant differs from the header", driftAtRead, "source_observed_at_ms",
			driftBuilder(nil, func(fields map[string]any) {
				fields["source_observed_at_ms"] = uint64(usTestBarOpenAt.Add(3 * time.Minute).UnixMilli())
			})},
		{"payload currency differs from the header", driftAtRead, "currency",
			driftBuilder(func(header *Header) { header.Currency = "KRW" }, nil)},
		{"header availability is not the bar close", driftAtRead, "source_available_at",
			driftBuilder(func(header *Header) { header.SourceAvailableAt = header.SourceEventAt }, nil)},
		{"header authority is not the official api", driftAtRead, "authority",
			driftBuilder(func(header *Header) { header.Authority = AuthoritySEC }, nil)},
		{"header schema version is foreign", driftAtRead, "schema_version",
			driftBuilder(func(header *Header) { header.SchemaVersion = "official_closed_bar_1m:v2" }, nil)},
		{"header issuer identity is foreign", driftAtRead, "issuer_identity",
			driftBuilder(func(header *Header) { header.IssuerIdentity = "US:MSFT" }, nil)},
		{"header issuer mapping version is foreign", driftAtRead, "issuer_mapping_version",
			driftBuilder(func(header *Header) { header.IssuerMappingVersion = "a112-bar-issuer-v2" }, nil)},
		{"header market effective date is not the session date", driftAtRead, "market_effective_date",
			driftBuilder(func(header *Header) { header.MarketEffectiveDate = "2026-08-13" }, nil)},
		{"header unit is not minor", driftAtRead, "unit",
			driftBuilder(func(header *Header) { header.Unit = "major" }, nil)},
		{"revision identity disagrees with the payload revision", driftAtRead, "revision",
			driftBuilder(func(header *Header) {
				header.RevisionIdentity = "r2"
				header.EvidenceID = header.SourceRecordID + ":r2"
			}, nil)},
		{"payload symbol differs from the header symbol", driftAtRead, "source_record_id",
			driftBuilder(nil, func(fields map[string]any) { fields["symbol"] = "MSFT" })},
		{"a foreign session forged under our record prefix", driftAtRead, "source_record_id", forgedForeignSessionRow},
		{"payload session and bar day disagree", driftAtEnvelope, "session_id",
			driftBuilder(nil, func(fields map[string]any) { fields["session_id"] = "US:2026-08-13" })},
		{"correction references a revision that was never stored", driftAtAppend, "supersedes",
			driftBuilder(func(header *Header) {
				header.RevisionIdentity = "r2"
				header.SupersedesRevisionIdentity = "r1"
				header.EvidenceID = header.SourceRecordID + ":r2"
			}, func(fields map[string]any) { fields["revision"] = uint64(2) })},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ingest := usTestBarOpenAt.Add(12 * time.Hour)
			clk := marketclock.NewFake(ingest)
			store := openTestStoreWithClock(t, clk)
			header, payload := tt.build(t)

			envelope, err := NewEnvelope(header, payload)
			if err != nil {
				assertDriftStage(t, tt.name, tt.stage, driftAtEnvelope, tt.field, err)
				return
			}
			if _, err := store.Append(ctx, envelope); err != nil {
				assertDriftStage(t, tt.name, tt.stage, driftAtAppend, tt.field, err)
				return
			}
			series, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(usTestBarOpenAt.Add(13*time.Hour), ingest))
			if err == nil {
				t.Fatalf("%s: drifted row was not refused at any stage (bars=%d)", tt.name, len(series.Bars))
			}
			assertDriftStage(t, tt.name, tt.stage, driftAtRead, tt.field, err)
			if len(series.Bars) != 0 {
				t.Fatalf("%s: refusal returned bars", tt.name)
			}
			var refusal *ValidationError
			if !errors.As(err, &refusal) || refusal.Code != RefusalIdentityMismatch {
				t.Fatalf("%s: read refusal = %v, want %s", tt.name, err, RefusalIdentityMismatch)
			}
			if !strings.Contains(refusal.Detail, header.EvidenceID) {
				t.Fatalf("%s: refusal does not name the offending evidence: %q", tt.name, refusal.Detail)
			}
		})
	}
}

func assertDriftStage(t *testing.T, name string, want, got driftStage, field string, err error) {
	t.Helper()
	if want != got {
		t.Fatalf("%s: refused at the %s stage, want the %s stage (err=%v)", name, got, want, err)
	}
	var refusal *ValidationError
	if !errors.As(err, &refusal) {
		t.Fatalf("%s: want a typed refusal at the %s stage, got %v", name, got, err)
	}
	if field != "" && !strings.Contains(refusal.Field+" "+refusal.Detail, field) {
		t.Fatalf("%s: refusal names %q/%q, want it to name %q", name, refusal.Field, refusal.Detail, field)
	}
}

func driftBuilder(headerMutate func(*Header), payloadMutate func(map[string]any)) func(*testing.T) (Header, []byte) {
	return func(t *testing.T) (Header, []byte) {
		t.Helper()
		header := usTestBarHeader(t, 1)
		if headerMutate != nil {
			headerMutate(&header)
		}
		return header, barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), payloadMutate)
	}
}

// 08-13 세션의 봉을 안쪽은 완전히 앞뒤가 맞게 만든 뒤, 기록 번호만 08-14 접두사로 위조한다.
// SQL 접두사 범위는 통과하지만 기록 번호 검사가 잡는다.
func forgedForeignSessionRow(t *testing.T) (Header, []byte) {
	t.Helper()
	foreignOpen := usTestBarOpenAt.AddDate(0, 0, -1)
	header, err := newClosedBar1mHeader(marketclock.MarketUS, "AAPL", "US:2026-08-13", usTestCalendarVersion,
		foreignOpen, 1, foreignOpen.Add(2*time.Minute), "USD")
	if err != nil {
		t.Fatalf("foreign session header: %v", err)
	}
	header.SourceRecordID = "US:AAPL:US:2026-08-14:60000:" + formatUint(uint64(foreignOpen.UnixMilli()))
	header.EvidenceID = header.SourceRecordID + ":forged:r1"
	fields := usTestBarFields(uint64(foreignOpen.UnixMilli()))
	fields["session_id"] = "US:2026-08-13"
	return header, barPayloadBytes(t, fields, nil)
}

// 헤더 분(13:30)과 본문 분(19:30)이 다르면, 평가 시점이 13:32라도 그 줄은 결과에 절대 들어가지 않는다.
func TestSealBarSeriesRefusesRatherThanHidingADriftedMinute(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ingest := usTestBarOpenAt.Add(12 * time.Hour)
	clk := marketclock.NewFake(ingest)
	store := openTestStoreWithClock(t, clk)
	drifted := uint64(usTestBarOpenAt.Add(6 * time.Hour).UnixMilli())
	envelope, err := NewEnvelope(usTestBarHeader(t, 1), barPayloadBytes(t, usTestBarFields(uint64(usTestBarOpenAt.UnixMilli())), func(fields map[string]any) {
		fields["open_at_ms"] = drifted
		fields["source_observed_at_ms"] = drifted + 120000
	}))
	if err != nil {
		t.Skipf("envelope already refused: %v", err)
	}
	if _, err := store.Append(ctx, envelope); err != nil {
		t.Fatalf("Append: %v", err)
	}
	series, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(usTestBarOpenAt.Add(2*time.Minute), ingest))
	if err == nil {
		t.Fatalf("drifted minute was silently hidden instead of refused: %+v", series.Bars)
	}
	if len(series.Bars) != 0 {
		t.Fatalf("refusal returned bars: %+v", series.Bars)
	}
}

// 같은 기록 번호(13:30)를 쓰면서 머리말 시계와 본문을 나란히 13:31로 옮긴 r2.
// 안쪽은 완벽하게 앞뒤가 맞으므로 다른 검사는 전부 통과한다. 기록 번호 검사만이 이것을 잡는다.
func TestSealBarSeriesRefusesACorrectionThatMovesTheBar(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ingest := usTestBarOpenAt.Add(12 * time.Hour)
	clk := marketclock.NewFake(ingest)
	store := openTestStoreWithClock(t, clk)
	appendTestBar(t, store, clk, usTestBarOpenAt, 1, ingest, nil, nil)

	movedOpen := usTestBarOpenAt.Add(time.Minute)
	moved := usTestBarHeader(t, 2) // 기록 번호는 13:30 그대로다.
	moved.SourceEventAt = movedOpen
	moved.SourceAvailableAt = movedOpen.Add(time.Minute)
	moved.ObservedAt = movedOpen.Add(2 * time.Minute)
	moved.IngestedAt = moved.ObservedAt
	fields := usTestBarFields(uint64(movedOpen.UnixMilli()))
	fields["revision"] = uint64(2)
	if _, err := store.Append(ctx, mustEnvelopeBytes(t, moved, barPayloadBytes(t, fields, nil))); err != nil {
		t.Fatalf("Append: %v", err)
	}

	series, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(usTestBarOpenAt.Add(13*time.Hour), ingest))
	var refusal *ValidationError
	if !errors.As(err, &refusal) {
		t.Fatalf("an identity change disguised as a correction was accepted: %+v", series.Bars)
	}
	if refusal.Field != "source_record_id" {
		t.Fatalf("refused by %q, want the record-id binding to catch it", refusal.Field)
	}
	if len(series.Bars) != 0 {
		t.Fatalf("refusal returned bars: %+v", series.Bars)
	}
}

// ---- 지문 골든 벡터와 성질 ----

func TestSealBarSeriesDigestChangesWithEveryInput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ingest := usTestBarOpenAt.Add(30 * time.Minute)
	clk := marketclock.NewFake(ingest)
	store := openTestStoreWithClock(t, clk)
	appendTestBar(t, store, clk, usTestBarOpenAt, 1, ingest, nil, nil)
	appendTestBar(t, store, clk, usTestBarOpenAt.Add(time.Minute), 1, ingest, nil, func(fields map[string]any) {
		fields["regular_session"] = false
	})
	base := usTestBarSeriesQuery(usTestBarOpenAt.Add(10*time.Minute), ingest)
	baseline, err := store.SealBarSeries(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name   string
		mutate func(*BarSeriesQuery)
	}{
		{"regular session only", func(q *BarSeriesQuery) { q.RegularSessionOnly = true }},
		{"max bars 256", func(q *BarSeriesQuery) { q.MaxBars = 256 }},
		{"later evaluation", func(q *BarSeriesQuery) { q.EvaluationAt = q.EvaluationAt.Add(time.Nanosecond) }},
		{"later ingestion cutoff", func(q *BarSeriesQuery) { q.IngestionCutoff = q.IngestionCutoff.Add(time.Nanosecond) }},
	} {
		tt := tt
		query := base
		tt.mutate(&query)
		other, err := store.SealBarSeries(ctx, query)
		if err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		if other.Digest == baseline.Digest {
			t.Fatalf("%s: digest did not change", tt.name)
		}
	}

	// 같은 신원, 다른 본문 바이트 → 다른 지문이어야 한다.
	other := openTestStoreWithClock(t, marketclock.NewFake(ingest))
	appendTestBar(t, other, marketclock.NewFake(ingest), usTestBarOpenAt, 1, ingest, nil, func(fields map[string]any) {
		fields["close_minor"] = uint64(2317000)
		fields["raw"].(map[string]any)["close"] = "231.7"
	})
	appendTestBar(t, other, marketclock.NewFake(ingest), usTestBarOpenAt.Add(time.Minute), 1, ingest, nil, func(fields map[string]any) {
		fields["regular_session"] = false
	})
	changed, err := other.SealBarSeries(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	if changed.Digest == baseline.Digest {
		t.Fatal("different payload bytes under the same identities produced the same digest")
	}
}

// ---- SQL 범위 ----

func TestSealBarSeriesScopeExcludesLaterSessionsOtherKindsAndOtherSymbols(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ingest := usTestBarOpenAt.Add(48 * time.Hour)
	clk := marketclock.NewFake(ingest)
	store := openTestStoreWithClock(t, clk)
	appendTestBar(t, store, clk, usTestBarOpenAt, 1, ingest, nil, nil)

	// 다음 세션(US:2026-08-15)의 봉은 위쪽 경계 밖이다.
	nextDay := usTestBarOpenAt.AddDate(0, 0, 1)
	nextHeader, err := newClosedBar1mHeader(marketclock.MarketUS, "AAPL", "US:2026-08-15", usTestCalendarVersion,
		nextDay, 1, nextDay.Add(2*time.Minute), "USD")
	if err != nil {
		t.Fatal(err)
	}
	nextFields := usTestBarFields(uint64(nextDay.UnixMilli()))
	nextFields["session_id"] = "US:2026-08-15"
	if _, err := store.Append(ctx, mustEnvelopeBytes(t, nextHeader, barPayloadBytes(t, nextFields, nil))); err != nil {
		t.Fatal(err)
	}

	// 우리 세션 접두사를 그대로 흉내 낸 옛 종류의 줄은 kind 조건이 막는다.
	legacy := validHeader(marketclock.MarketUS, KindUSParticipation, "legacy-inside-the-prefix", "rev-1")
	legacy.SourceRecordID = "US:AAPL:US:2026-08-14:60000:" + formatUint(uint64(usTestBarOpenAt.UnixMilli()))
	legacy.Symbol = "AAPL"
	if _, err := store.Append(ctx, mustEnvelope(t, legacy, `{"value":1}`)); err != nil {
		t.Fatal(err)
	}

	// 같은 접두사를 쓰지만 종목 열이 다른 줄은 symbol 조건이 막는다.
	foreign := validHeader(marketclock.MarketUS, KindUSParticipation, "legacy-other-symbol", "rev-1")
	foreign.SourceRecordID = "US:AAPL:US:2026-08-14:60000:" + formatUint(uint64(usTestBarOpenAt.Add(time.Minute).UnixMilli()))
	foreign.Symbol = "MSFT"
	if _, err := store.Append(ctx, mustEnvelope(t, foreign, `{"value":1}`)); err != nil {
		t.Fatal(err)
	}

	// 다른 종목의 닫힌 봉인데 기록 번호만 우리 접두사 안으로 위조한 줄.
	// 오직 SQL의 symbol 조건만이 이 줄을 막는다.
	otherSymbol, err := newClosedBar1mHeader(marketclock.MarketUS, "MSFT", usTestSessionID, usTestCalendarVersion,
		usTestBarOpenAt.Add(2*time.Minute), 1, usTestBarOpenAt.Add(4*time.Minute), "USD")
	if err != nil {
		t.Fatal(err)
	}
	otherSymbol.SourceRecordID = "US:AAPL:US:2026-08-14:60000:" + formatUint(uint64(usTestBarOpenAt.Add(2*time.Minute).UnixMilli()))
	otherSymbol.EvidenceID = otherSymbol.SourceRecordID + ":msft:r1"
	otherFields := usTestBarFields(uint64(usTestBarOpenAt.Add(2 * time.Minute).UnixMilli()))
	otherFields["symbol"] = "MSFT"
	if _, err := store.Append(ctx, mustEnvelopeBytes(t, otherSymbol, barPayloadBytes(t, otherFields, nil))); err != nil {
		t.Fatal(err)
	}

	series, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(usTestBarOpenAt.Add(47*time.Hour), ingest))
	if err != nil {
		t.Fatalf("SealBarSeries: %v", err)
	}
	if len(series.Bars) != 1 || series.Bars[0].Payload.SessionID != usTestSessionID || series.Bars[0].Payload.Symbol != "AAPL" {
		t.Fatalf("scope leaked: %+v", series.Bars)
	}
}

// SealBarSeries는 구조적으로 읽기 전용이어야 한다 (consumer_static_test.go와 같은 방식).
func TestSealBarSeriesIsStructurallySelectOnly(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "breakout_series.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var method *ast.FuncDecl
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Recv != nil && function.Name.Name == "SealBarSeries" {
			method = function
		}
	}
	if method == nil {
		t.Fatal("SealBarSeries not found")
	}
	ast.Inspect(method.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.SelectorExpr:
			for _, forbidden := range []string{"Exec", "ExecContext", "Begin", "BeginTx"} {
				if value.Sel.Name == forbidden {
					t.Errorf("SealBarSeries contains mutating database call %s", forbidden)
				}
			}
		case *ast.BasicLit:
			if value.Kind != token.STRING {
				break
			}
			literal, unquoteErr := strconv.Unquote(value.Value)
			if unquoteErr != nil {
				break
			}
			upper := strings.ToUpper(literal)
			for _, forbidden := range []string{"INSERT ", "UPDATE ", "DELETE ", "ALTER ", "DROP ", "PRAGMA "} {
				if strings.Contains(upper, forbidden) {
					t.Errorf("SealBarSeries contains mutating SQL %q", forbidden)
				}
			}
		}
		return true
	})
	source, err := os.ReadFile("breakout_series.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"INSERT ", "UPDATE ", "DELETE ", "ALTER TABLE", "DROP ", "ExecContext", "BeginTx"} {
		if strings.Contains(strings.ToUpper(string(source)), strings.ToUpper(forbidden)) {
			t.Errorf("breakout_series.go contains the mutating construct %q", forbidden)
		}
	}
}

// ---- 시험 보조 ----

func usTestBarSeriesQuery(evaluationAt, ingestionCutoff time.Time) BarSeriesQuery {
	return BarSeriesQuery{
		Market:          "US",
		Symbol:          "AAPL",
		SessionID:       usTestSessionID,
		IntervalMS:      60000,
		EvaluationAt:    evaluationAt,
		IngestionCutoff: ingestionCutoff,
	}
}

func appendTestBar(t *testing.T, store *Store, clk *marketclock.Fake, openAt time.Time, revision uint64,
	ingestedAt time.Time, mutateHeader func(*Header), mutate func(map[string]any)) Envelope {
	t.Helper()
	header, err := newClosedBar1mHeader(marketclock.MarketUS, "AAPL", usTestSessionID, usTestCalendarVersion,
		openAt, revision, openAt.Add(2*time.Minute), "USD")
	if err != nil {
		t.Fatalf("newClosedBar1mHeader: %v", err)
	}
	if mutateHeader != nil {
		mutateHeader(&header)
	}
	fields := usTestBarFields(uint64(openAt.UnixMilli()))
	fields["revision"] = revision
	envelope := mustEnvelopeBytes(t, header, barPayloadBytes(t, fields, mutate))
	clk.Set(ingestedAt)
	stored, err := store.Append(context.Background(), envelope)
	if err != nil {
		t.Fatalf("Append bar at %s revision %d: %v", openAt, revision, err)
	}
	return stored.Evidence
}

func mustEnvelopeBytes(t *testing.T, header Header, payload []byte) Envelope {
	t.Helper()
	envelope, err := NewEnvelope(header, payload)
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	return envelope
}

// 지문 골든 벡터. 값 하나를 한 번 재서 박아 두고, 그 뒤로는 바뀌면 즉시 드러나게 한다.
// 기대값은 두 갈래로 확인한다: (1) 손으로 적은 preimage를 여기서 직접 해시하고,
// (2) 박아 둔 16진수와 비교한다. 구현이 한 칸이라도 달라지면 둘 다 깨진다.
func TestSealBarSeriesDigestGoldenVector(t *testing.T) {
	t.Parallel()
	const (
		goldenFirstEvidenceID     = "US:AAPL:US:2026-08-14:60000:1786714200000:r1"
		goldenSecondEvidenceID    = "US:AAPL:US:2026-08-14:60000:1786714260000:r1"
		goldenFirstPayloadDigest  = "a0e71356998f0dda4a29b96e4f5264bb91624aad918531b41ebfff16fdcd62ee"
		goldenSecondPayloadDigest = "fbb18a99d354b563bb3e836b76a0fc118732c8bb6dda2b079edba9481ce4a99b"
		goldenSeriesDigest        = "51d803804d83a41a55adc81355bace7898faa66859d1f43923ff4b1c1ae79d87"
	)
	ctx := context.Background()
	ingest := usTestBarOpenAt.Add(30 * time.Minute)
	clk := marketclock.NewFake(ingest)
	store := openTestStoreWithClock(t, clk)
	for _, offset := range []time.Duration{0, time.Minute} {
		input := usTestBarInput()
		input.OpenAt = usTestBarOpenAt.Add(offset)
		input.ObservedAt = input.OpenAt.Add(2 * time.Minute)
		input.SuccessorOpenAt = input.OpenAt.Add(time.Minute)
		envelope, err := NewClosedBar1mEnvelope(input)
		if err != nil {
			t.Fatalf("constructor: %v", err)
		}
		if _, err := store.Append(ctx, envelope); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
	series, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(usTestBarOpenAt.Add(10*time.Minute), ingest))
	if err != nil {
		t.Fatalf("SealBarSeries: %v", err)
	}
	if len(series.Bars) != 2 {
		t.Fatalf("bars = %d", len(series.Bars))
	}
	preimage := []string{
		barSeriesDigestDomain,
		"US", "AAPL", "US:2026-08-14", "60000",
		"2026-08-14T13:40:00.000000000Z", "2026-08-14T14:00:00.000000000Z",
		"false", "512",
		goldenFirstEvidenceID, "r1", goldenFirstPayloadDigest, "1786714200000",
		goldenSecondEvidenceID, "r1", goldenSecondPayloadDigest, "1786714260000",
	}
	hash := sha256.New()
	for _, field := range preimage {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len(field)))
		_, _ = hash.Write(length[:])
		_, _ = hash.Write([]byte(field))
	}
	recomputed := hex.EncodeToString(hash.Sum(nil))
	if series.Bars[0].EvidenceID != goldenFirstEvidenceID || series.Bars[1].EvidenceID != goldenSecondEvidenceID {
		t.Fatalf("evidence ids drifted: %s / %s", series.Bars[0].EvidenceID, series.Bars[1].EvidenceID)
	}
	if series.Bars[0].PayloadDigest != goldenFirstPayloadDigest || series.Bars[1].PayloadDigest != goldenSecondPayloadDigest {
		t.Fatalf("payload digests drifted: %s / %s", series.Bars[0].PayloadDigest, series.Bars[1].PayloadDigest)
	}
	if recomputed != goldenSeriesDigest {
		t.Fatalf("hand-written preimage hashes to %s, want the pinned %s", recomputed, goldenSeriesDigest)
	}
	if series.Digest != goldenSeriesDigest {
		t.Fatalf("series digest = %s, want the pinned %s", series.Digest, goldenSeriesDigest)
	}
}

// 판 번호는 글자 순서가 아니라 숫자로 비교해야 한다.
// SQLite가 revision_identity를 글자순으로 주면 "r1","r10","r11","r2",... 순서가 되어,
// 더 높은 판(r11)을 먼저 본 뒤에 더 낮은 판(r2~r9)이 뒤늦게 들어온다.
// 그때 낮은 판을 건너뛰는 가지(continue)가 실제로 실행되는 경우다.
func TestSealBarSeriesPicksTheHighestRevisionNumericallyNotTextually(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	clk := marketclock.NewFake(usTestBarOpenAt.Add(10 * time.Minute))
	store := openTestStoreWithClock(t, clk)
	ingestOf := func(revision uint64) time.Time {
		return usTestBarOpenAt.Add(time.Duration(10+revision) * time.Minute)
	}
	for revision := uint64(1); revision <= 11; revision++ {
		revision := revision
		appendTestBar(t, store, clk, usTestBarOpenAt, revision, ingestOf(revision), nil, func(fields map[string]any) {
			fields["volume"] = 10000 + revision
			fields["raw"].(map[string]any)["volume"] = formatUint(10000 + revision)
		})
	}
	// 저장된 줄이 글자순으로 나오는지 확인한다. 그래야 이 시험이 노리는 순서가 실제로 만들어진다.
	rows, err := store.db.QueryContext(ctx, `SELECT revision_identity FROM evidence_records
        WHERE evidence_kind=? ORDER BY source_record_id, revision_identity, evidence_id`, KindOfficialClosedBar1m)
	if err != nil {
		t.Fatal(err)
	}
	var scanned []string
	for rows.Next() {
		var identity string
		if err := rows.Scan(&identity); err != nil {
			t.Fatal(err)
		}
		scanned = append(scanned, identity)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	if len(scanned) != 11 || scanned[1] != "r10" || scanned[3] != "r2" {
		t.Fatalf("scan order = %v, want the textual order that puts r10/r11 before r2", scanned)
	}

	evaluationAt := usTestBarOpenAt.Add(5 * time.Minute)
	all, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(evaluationAt, ingestOf(11)))
	if err != nil {
		t.Fatalf("SealBarSeries: %v", err)
	}
	if len(all.Bars) != 1 {
		t.Fatalf("bars = %d, want exactly one winner per bar identity: %+v", len(all.Bars), all.Bars)
	}
	if all.Bars[0].RevisionIdentity != "r11" || all.Bars[0].Payload.Revision != 11 || all.Bars[0].Payload.Volume != 10011 {
		t.Fatalf("winner = %+v, want revision 11", all.Bars[0])
	}

	// r11이 아직 안 들어온 시점에서는 r10이 이겨야 한다. r2~r9를 건너뛴 결과가 남아 있는지 본다.
	upToTen, err := store.SealBarSeries(ctx, usTestBarSeriesQuery(evaluationAt, ingestOf(10)))
	if err != nil {
		t.Fatalf("SealBarSeries at the earlier cutoff: %v", err)
	}
	if len(upToTen.Bars) != 1 {
		t.Fatalf("bars = %d, want exactly one: %+v", len(upToTen.Bars), upToTen.Bars)
	}
	if upToTen.Bars[0].RevisionIdentity != "r10" || upToTen.Bars[0].Payload.Revision != 10 || upToTen.Bars[0].Payload.Volume != 10010 {
		t.Fatalf("winner = %+v, want revision 10", upToTen.Bars[0])
	}
}
