package strategyevidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// 한 번에 읽을 수 있는 봉의 최대 개수다. 더 많으면 자르지 않고 거절한다.
// 자르면 "무엇이 빠졌는지" 모르는 채로 전략이 판단하게 되기 때문이다.
const MaxBarSeriesBars = 512

// 봉 묶음 지문에 맨 앞에 붙는 이름표다. 종목 스냅샷 지문과 절대 섞이지 않게 한다.
const barSeriesDigestDomain = "tossos.strategyevidence.bar_series.v1"

// BarSeriesQuery는 "어느 시장·종목·세션의 1분봉을, 어느 두 시점 기준으로 볼지"를 적는다.
type BarSeriesQuery struct {
	Market             string
	Symbol             string
	SessionID          string
	IntervalMS         uint64
	EvaluationAt       time.Time
	IngestionCutoff    time.Time
	RegularSessionOnly bool
	MaxBars            int
}

// ClosedBarRecord는 저장된 봉 한 줄이다. 신원과 본문을 함께 들고 다닌다.
type ClosedBarRecord struct {
	EvidenceID       string
	RevisionIdentity string
	PayloadDigest    string
	Payload          ClosedBar1mPayload
}

// BarSeries는 한 세션의 봉 묶음과, 그 묶음을 그대로 다시 만들 수 있는 지문이다.
type BarSeries struct {
	Query  BarSeriesQuery
	Bars   []ClosedBarRecord
	Digest string
}

// SealBarSeries는 저장된 닫힌 1분봉을 시간 순서대로 읽는다. 쓰기는 전혀 하지 않는다.
//
// 비용(이 저장소에서 직접 재 본 값): 정규장 390봉 한 세션을 읽는 데 봉마다 판이 하나면
// 약 47밀리초, 판이 넷이면(1,560줄) 약 159밀리초다. 리뷰에서 잰 값(42밀리초/168밀리초)과
// 같은 크기다. 한 줄을 두 번 해독하기 때문이다
// (scanEnvelope 안의 NewEnvelope에서 한 번, DecodeClosedBar1mPayload에서 또 한 번).
// 그래서 L3/L5는 1분마다 도는 뜨거운 경로에서 이것을 그대로 부르면 안 된다.
// 한 번 읽어서 해독된 ClosedBarRecord를 재사용하는 것이 정석이다.
// 두 컷오프(자료가 생긴 시점 / 우리가 받아 적은 시점)를 따로 적용하므로,
// 나중에 들어온 정정본은 "그 뒤 시점"으로 볼 때만 보인다.
func (s *Store) SealBarSeries(ctx context.Context, query BarSeriesQuery) (BarSeries, error) {
	normalized, market, err := normalizeBarSeriesQuery(query)
	if err != nil {
		return BarSeries{}, err
	}
	prefix := strings.Join([]string{normalized.Market, normalized.Symbol, normalized.SessionID, formatUint(normalized.IntervalMS), ""}, ":")
	upperBound, ok := textPrefixUpperBound(prefix)
	if !ok {
		return BarSeries{}, invalid(RefusalIdentityMismatch, "bar_series_query", "session scope is not a usable record prefix")
	}
	evaluation := stamp(normalized.EvaluationAt)
	ingestion := stamp(normalized.IngestionCutoff)
	rows, err := s.db.QueryContext(ctx, `SELECT `+envelopeColumns("e")+`
      FROM evidence_records e
     WHERE e.evidence_kind=? AND e.market=? AND e.symbol=?
       AND e.source_record_id>=? AND e.source_record_id<?
       AND e.source_event_at<=? AND e.source_available_at<=? AND e.ingested_at<=?
     ORDER BY e.source_record_id, e.revision_identity, e.evidence_id`,
		KindOfficialClosedBar1m, market, normalized.Symbol, prefix, upperBound,
		evaluation, evaluation, ingestion)
	if err != nil {
		return BarSeries{}, err
	}
	winners := make(map[string]ClosedBarRecord)
	for rows.Next() {
		record, recordID, err := scanClosedBarRecord(rows, normalized)
		if err != nil {
			_ = rows.Close()
			return BarSeries{}, err
		}
		// 같은 봉에 여러 판이 동시에 보이면 언제나 더 큰 판이 이긴다.
		if current, exists := winners[recordID]; exists && record.Payload.Revision <= current.Payload.Revision {
			continue
		}
		winners[recordID] = record
	}
	if err := rows.Close(); err != nil {
		return BarSeries{}, err
	}
	if err := rows.Err(); err != nil {
		return BarSeries{}, err
	}
	bars := make([]ClosedBarRecord, 0, len(winners))
	for _, record := range winners {
		if normalized.RegularSessionOnly && !record.Payload.RegularSession {
			continue
		}
		bars = append(bars, record)
	}
	sort.Slice(bars, func(left, right int) bool { return bars[left].Payload.OpenAtMS < bars[right].Payload.OpenAtMS })
	if len(bars) > normalized.MaxBars {
		return BarSeries{}, invalid(RefusalPayloadInvalid, "bar_series",
			"session carries "+strconv.Itoa(len(bars))+" visible bars, above the bound of "+strconv.Itoa(normalized.MaxBars))
	}
	return BarSeries{Query: normalized, Bars: bars, Digest: barSeriesDigest(normalized, bars)}, nil
}

// scanClosedBarRecord는 한 줄을 다시 봉투로 만들고(=엄격 해독을 다시 통과시키고)
// 본문이 자기 신원과 같은 이야기를 하는지 확인한다.
func scanClosedBarRecord(rows scanner, query BarSeriesQuery) (ClosedBarRecord, string, error) {
	envelope, err := scanEnvelope(rows)
	if err != nil {
		return ClosedBarRecord{}, "", err
	}
	header := envelope.Header()
	payload, err := DecodeClosedBar1mPayload(envelope.CanonicalPayload())
	if err != nil {
		return ClosedBarRecord{}, "", err
	}
	// 아래 검사들은 "쓸 때는 통과했는데 읽을 때 말이 달라지는" 줄을 전부 막는다.
	// 하나라도 어긋나면 자르지 않고 통째로 거절하고, 어느 줄이 범인인지 이름을 적는다.
	mismatch := func(field, detail string) error {
		return invalid(RefusalIdentityMismatch, field, "evidence "+header.EvidenceID+" "+detail)
	}
	sessionDate, err := sessionDateFor(payload.Market, payload.SessionID)
	if err != nil {
		return ClosedBarRecord{}, "", mismatch("session_id", err.Error())
	}
	// 기록 번호는 본문이 주장하는 신원 그대로여야 한다. SQL이 이미 질의 접두사로 걸렀으므로,
	// 이 한 줄이 시장·종목·세션·간격이 질의와 같다는 것까지 함께 증명한다.
	expectedRecordID := closedBar1mRecordID(payload.Market, payload.Symbol, payload.SessionID,
		payload.IntervalMS, uint64(header.SourceEventAt.UnixMilli()))
	switch {
	case header.SourceRecordID != expectedRecordID:
		return ClosedBarRecord{}, "", mismatch("source_record_id", "does not carry the bar identity its own payload and header clocks describe")
	case header.Authority != AuthorityTossOpenAPI:
		return ClosedBarRecord{}, "", mismatch("authority", "was not issued by the official Toss Open API")
	case header.SchemaVersion != SchemaOfficialClosedBar1m:
		return ClosedBarRecord{}, "", mismatch("schema_version", "carries a foreign schema version")
	case header.IssuerIdentity != query.Market+":"+query.Symbol:
		return ClosedBarRecord{}, "", mismatch("issuer_identity", "carries a foreign issuer identity")
	case header.IssuerMappingVersion != BarIssuerMappingVersion:
		return ClosedBarRecord{}, "", mismatch("issuer_mapping_version", "carries a foreign issuer mapping version")
	case header.MarketEffectiveDate != sessionDate:
		return ClosedBarRecord{}, "", mismatch("market_effective_date", "is not the session date its payload carries")
	case header.Unit != "minor":
		return ClosedBarRecord{}, "", mismatch("unit", "is not recorded in integer minor units")
	case !header.SourceAvailableAt.Equal(header.SourceEventAt.Add(time.Duration(query.IntervalMS) * time.Millisecond)):
		return ClosedBarRecord{}, "", mismatch("source_available_at", "does not become available exactly one interval after it opened")
	case payload.OpenAtMS != uint64(header.SourceEventAt.UnixMilli()):
		return ClosedBarRecord{}, "", mismatch("open_at_ms", "claims a minute its header does not")
	case payload.SourceObservedAtMS != uint64(header.ObservedAt.UnixMilli()):
		return ClosedBarRecord{}, "", mismatch("source_observed_at_ms", "claims an observation instant its header does not")
	case payload.Currency != header.Currency:
		return ClosedBarRecord{}, "", mismatch("currency", "carries a payload currency that differs from its header currency")
	case header.RevisionIdentity != revisionIdentityFor(payload.Revision):
		return ClosedBarRecord{}, "", mismatch("revision", "carries a payload revision that differs from its revision identity")
	}
	return ClosedBarRecord{
		EvidenceID:       header.EvidenceID,
		RevisionIdentity: header.RevisionIdentity,
		PayloadDigest:    envelope.PayloadDigest(),
		Payload:          payload,
	}, header.SourceRecordID, nil
}

func normalizeBarSeriesQuery(query BarSeriesQuery) (BarSeriesQuery, marketclock.Market, error) {
	market, err := marketclock.ParseMarket(query.Market)
	if err != nil {
		return BarSeriesQuery{}, "", invalid(RefusalIdentityMismatch, "bar_series_query.market", "only kr and us are supported")
	}
	query.Market = strings.ToUpper(string(market))
	query.Symbol = strings.ToUpper(strings.TrimSpace(query.Symbol))
	query.SessionID = strings.TrimSpace(query.SessionID)
	if !canonicalSymbolText(query.Symbol) {
		return BarSeriesQuery{}, "", invalid(RefusalIdentityMismatch, "bar_series_query.symbol",
			"a canonical symbol without the ':' record separator is required")
	}
	if _, err := sessionDateFor(query.Market, query.SessionID); err != nil {
		return BarSeriesQuery{}, "", invalid(RefusalIdentityMismatch, "bar_series_query.session_id", err.Error())
	}
	if query.IntervalMS != ClosedBar1mIntervalMS {
		return BarSeriesQuery{}, "", invalid(RefusalIdentityMismatch, "bar_series_query.interval_ms",
			"only "+formatUint(ClosedBar1mIntervalMS)+" millisecond bars are stored")
	}
	if query.MaxBars < 0 || query.MaxBars > MaxBarSeriesBars {
		return BarSeriesQuery{}, "", invalid(RefusalPayloadInvalid, "bar_series_query.max_bars",
			"must be between 1 and "+strconv.Itoa(MaxBarSeriesBars))
	}
	if query.MaxBars == 0 {
		query.MaxBars = MaxBarSeriesBars
	}
	if query.EvaluationAt.IsZero() || query.IngestionCutoff.IsZero() {
		return BarSeriesQuery{}, "", invalid(RefusalTimestampInvalid, "bar_series_query", "both cutoffs are required")
	}
	query.EvaluationAt = query.EvaluationAt.UTC()
	query.IngestionCutoff = query.IngestionCutoff.UTC()
	return query, market, nil
}

// barSeriesDigest는 질문과 결과를 길이-접두 방식으로 이어 붙여 한 개의 지문으로 만든다.
// 맨 앞의 이름표 덕분에 종목 스냅샷 지문과 값이 같아질 수 없다.
func barSeriesDigest(query BarSeriesQuery, bars []ClosedBarRecord) string {
	hash := sha256.New()
	writeSnapshotDigestField(hash, barSeriesDigestDomain)
	for _, value := range []string{
		query.Market, query.Symbol, query.SessionID, formatUint(query.IntervalMS),
		stamp(query.EvaluationAt), stamp(query.IngestionCutoff),
		strconv.FormatBool(query.RegularSessionOnly), strconv.Itoa(query.MaxBars),
	} {
		writeSnapshotDigestField(hash, value)
	}
	for _, bar := range bars {
		for _, value := range []string{
			bar.EvidenceID, bar.RevisionIdentity, bar.PayloadDigest, formatUint(bar.Payload.OpenAtMS),
		} {
			writeSnapshotDigestField(hash, value)
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// textPrefixUpperBound는 "이 글자로 시작하는 모든 값"의 바로 다음 값을 만든다.
// SQLite의 기본 TEXT 비교는 바이트 순서이므로 [prefix, upperBound) 범위가 정확히 접두 일치다.
func textPrefixUpperBound(prefix string) (string, bool) {
	raw := []byte(prefix)
	for index := len(raw) - 1; index >= 0; index-- {
		if raw[index] < 0xFF {
			raw[index]++
			return string(raw[:index+1]), true
		}
	}
	return "", false
}
