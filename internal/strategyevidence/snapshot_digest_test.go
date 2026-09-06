package strategyevidence

// snapshotDigest 가 **무엇을 결속하는가**를 재는 유일한 자리다.
//
// 왜 replay 시험으로는 못 재는가: consumer.go 의 snapshotItemMatchesQuery 가 market·
// symbol·issuer·mapping·source 시각·effective date 여섯 가지를 digest 재계산 **앞에서**
// 따로 다시 검사하고 거기서 ErrSnapshotUnavailable 을 낸다. 그래서 digest 의 preimage 를
// EvidenceID+payload digest 로 줄여도 replay 시험은 전부 초록이 된다(한 규칙을 두 곳에서
// 판정하면 서로가 서로의 시험을 통과시킨다). 여기서는 replay 를 거치지 않고
// snapshotDigest 를 직접 부른다.

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"reflect"
	"sort"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// goldenSnapshotDigest 는 생산 코드를 **부르지 않는** 독립 프로그램이 낳은 값이다.
// 설계가 규정한 preimage(질의 6개 필드 → EvidenceID 오름차순 항목 → 항목마다 Header
// 20개 필드 + payload digest, 각 값 앞에 8바이트 big-endian 길이)를 손으로 다시 만들어
// SHA-256 한 결과다. 실행 중인 시스템에서 읽어 온 값이 아니므로, 생산 preimage 가
// 한 필드라도 줄면 이 상수와 갈라진다.
const goldenSnapshotDigest = "4d4ec1f561389e1feafda66b1b9c1871e04d799fc6f839032546de6489537fab"

// goldenSnapshotDigestDescending 은 같은 두 항목을 EvidenceID **내림차순**으로 넣었을 때
// 나오는 값이다. 정렬 비교자가 뒤집히면 생산 결과가 이 값이 된다.
const goldenSnapshotDigestDescending = "1e3b9ba0b3c10968a7a2c67a46577d1cd889ec198d2716d9d9ccaa7f493bc535"

const (
	goldenPayloadA = `{"value":1}`
	goldenPayloadB = `{"value":2}`
	goldenDigestA  = "48208f9428d64634bd8e28ff345bf0eab60d53c18fa2fbdb0b9bc1e84df2b5f6"
	goldenDigestB  = "49c987621f206f09e5fbe23b516b55a36f838cb14867961f1d84a554d3a35b6b"
)

func goldenSnapshotQuery() SnapshotQuery {
	return SnapshotQuery{
		Market:               marketclock.MarketUS,
		Symbol:               "AAPL",
		IssuerIdentity:       "issuer-aapl",
		IssuerMappingVersion: "issuer-map-v1",
		EvaluationAt:         time.Date(2026, 8, 3, 14, 0, 0, 0, time.UTC),
		IngestionCutoff:      time.Date(2026, 8, 3, 15, 0, 0, 0, time.UTC),
	}
}

func goldenHeader(evidenceID, recordID, revision, supersedes string, confidence Confidence) Header {
	return Header{
		EvidenceID:                 evidenceID,
		Market:                     marketclock.MarketUS,
		Symbol:                     "AAPL",
		IssuerIdentity:             "issuer-aapl",
		IssuerMappingVersion:       "issuer-map-v1",
		Kind:                       KindUSParticipation,
		SchemaVersion:              "v1",
		Authority:                  AuthoritySEC,
		SourceRecordID:             recordID,
		RevisionIdentity:           revision,
		SupersedesRevisionIdentity: supersedes,
		MarketEffectiveDate:        "2026-08-03",
		SourceEventAt:              time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
		SourceAvailableAt:          time.Date(2026, 8, 3, 13, 0, 0, 0, time.UTC),
		ObservedAt:                 time.Date(2026, 8, 3, 13, 1, 0, 0, time.UTC),
		IngestedAt:                 time.Date(2026, 8, 3, 13, 2, 0, 0, time.UTC),
		Currency:                   "USD",
		Unit:                       "minor",
		Availability:               AvailabilityAvailable,
		Confidence:                 confidence,
	}
}

func goldenItems(t *testing.T) []Envelope {
	t.Helper()
	first := mustEnvelope(t, goldenHeader("golden-a", "record-a", "rev-a2", "rev-a1", ConfidenceVerified), goldenPayloadA)
	second := mustEnvelope(t, goldenHeader("golden-b", "record-b", "rev-b1", "", ConfidenceUnverified), goldenPayloadB)
	if first.PayloadDigest() != goldenDigestA || second.PayloadDigest() != goldenDigestB {
		t.Fatalf("golden payload digests moved: %s %s", first.PayloadDigest(), second.PayloadDigest())
	}
	return []Envelope{first, second}
}

// withHeaderField 는 Envelope 를 한 필드만 바꿔 다시 만든다. NewEnvelope 를 거치지 않는
// 이유는 정규화가 거부할 값(예: 시장만 바꾼 header)까지 digest 에 넣어 봐야 하기 때문이다.
// digest 함수는 정규화된 header 만 받는다는 가정 위에 서 있으므로, 그 가정을 시험이
// 대신 지킨다.
func withHeaderField(item Envelope, mutate func(*Header)) Envelope {
	header := item.Header()
	mutate(&header)
	return Envelope{header: header, payload: item.CanonicalPayload(), digest: item.PayloadDigest()}
}

// referenceSnapshotDigest 는 golden 을 낳은 것과 같은 규정을 시험 쪽에서 다시 쓴 것이다.
// 생산 함수를 부르지 않으므로 "생산이 무엇을 하든 같은 답"이 되지 않는다. golden 상수와
// 짝을 이뤄, 둘 다 조용히 함께 바뀌는 일을 막는다.
func referenceSnapshotDigest(query SnapshotQuery, items []Envelope) string {
	utc := func(value time.Time) string { return value.UTC().Format("2006-01-02T15:04:05.000000000Z") }
	write := func(destination hash.Hash, value string) {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len(value)))
		_, _ = destination.Write(length[:])
		_, _ = destination.Write([]byte(value))
	}
	sum := sha256.New()
	for _, value := range []string{
		string(query.Market), query.Symbol, query.IssuerIdentity, query.IssuerMappingVersion,
		utc(query.EvaluationAt), utc(query.IngestionCutoff),
	} {
		write(sum, value)
	}
	ordered := append([]Envelope(nil), items...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Header().EvidenceID < ordered[j].Header().EvidenceID
	})
	for _, item := range ordered {
		header := item.Header()
		for _, value := range []string{
			header.EvidenceID, string(header.Market), header.Symbol, header.IssuerIdentity,
			header.IssuerMappingVersion, string(header.Kind), header.SchemaVersion, string(header.Authority),
			header.SourceRecordID, header.RevisionIdentity, header.SupersedesRevisionIdentity,
			header.MarketEffectiveDate, utc(header.SourceEventAt), utc(header.SourceAvailableAt),
			utc(header.ObservedAt), utc(header.IngestedAt), header.Currency, header.Unit,
			string(header.Availability), string(header.Confidence), item.PayloadDigest(),
		} {
			write(sum, value)
		}
	}
	return hex.EncodeToString(sum.Sum(nil))
}

func TestSnapshotDigestMatchesFrozenGoldenVector(t *testing.T) {
	t.Parallel()
	query := goldenSnapshotQuery()
	items := goldenItems(t)
	if got := snapshotDigest(query, items); got != goldenSnapshotDigest {
		t.Fatalf("snapshotDigest = %s, want frozen golden %s", got, goldenSnapshotDigest)
	}
	if got := referenceSnapshotDigest(query, items); got != goldenSnapshotDigest {
		t.Fatalf("test-side reference = %s, want frozen golden %s", got, goldenSnapshotDigest)
	}
}

func TestSnapshotDigestOrdersItemsAscendingWithoutMutatingCaller(t *testing.T) {
	t.Parallel()
	query := goldenSnapshotQuery()
	items := goldenItems(t)
	reversed := []Envelope{items[1], items[0]}
	got := snapshotDigest(query, reversed)
	if got != goldenSnapshotDigest {
		t.Fatalf("reversed input digest = %s, want %s (ordering is not applied)", got, goldenSnapshotDigest)
	}
	if got == goldenSnapshotDigestDescending {
		t.Fatalf("snapshotDigest ordered items descending: %s", got)
	}
	if reversed[0].Header().EvidenceID != "golden-b" || reversed[1].Header().EvidenceID != "golden-a" {
		t.Fatalf("snapshotDigest mutated the caller slice: %s, %s",
			reversed[0].Header().EvidenceID, reversed[1].Header().EvidenceID)
	}
}

func TestSnapshotDigestBindsEveryHeaderFieldAndPayloadDigest(t *testing.T) {
	t.Parallel()
	query := goldenSnapshotQuery()
	base := goldenItems(t)
	baseline := snapshotDigest(query, base)

	mutations := map[string]func(*Header){
		"EvidenceID":                 func(h *Header) { h.EvidenceID = "golden-a2" },
		"Market":                     func(h *Header) { h.Market = marketclock.MarketKR },
		"Symbol":                     func(h *Header) { h.Symbol = "MSFT" },
		"IssuerIdentity":             func(h *Header) { h.IssuerIdentity = "issuer-msft" },
		"IssuerMappingVersion":       func(h *Header) { h.IssuerMappingVersion = "issuer-map-v2" },
		"Kind":                       func(h *Header) { h.Kind = KindDisclosure },
		"SchemaVersion":              func(h *Header) { h.SchemaVersion = "v2" },
		"Authority":                  func(h *Header) { h.Authority = AuthorityTossOpenAPI },
		"SourceRecordID":             func(h *Header) { h.SourceRecordID = "record-z" },
		"RevisionIdentity":           func(h *Header) { h.RevisionIdentity = "rev-a3" },
		"SupersedesRevisionIdentity": func(h *Header) { h.SupersedesRevisionIdentity = "rev-a0" },
		"MarketEffectiveDate":        func(h *Header) { h.MarketEffectiveDate = "2026-08-02" },
		"SourceEventAt":              func(h *Header) { h.SourceEventAt = h.SourceEventAt.Add(-time.Nanosecond) },
		"SourceAvailableAt":          func(h *Header) { h.SourceAvailableAt = h.SourceAvailableAt.Add(-time.Nanosecond) },
		"ObservedAt":                 func(h *Header) { h.ObservedAt = h.ObservedAt.Add(time.Nanosecond) },
		"IngestedAt":                 func(h *Header) { h.IngestedAt = h.IngestedAt.Add(time.Nanosecond) },
		"Currency":                   func(h *Header) { h.Currency = "KRW" },
		"Unit":                       func(h *Header) { h.Unit = "major" },
		"Availability":               func(h *Header) { h.Availability = AvailabilityUnavailable },
		"Confidence":                 func(h *Header) { h.Confidence = ConfidenceUnverified },
	}
	assertCoversEveryField(t, reflect.TypeOf(Header{}), mutations, "Header")

	for name, mutate := range mutations {
		t.Run("header/"+name, func(t *testing.T) {
			changed := withHeaderField(base[0], mutate)
			assertMutatesOnly(t, name, base[0].Header(), changed.Header())
			mutated := []Envelope{changed, base[1]}
			if got := snapshotDigest(query, mutated); got == baseline {
				t.Fatalf("changing Header.%s left the snapshot digest at %s", name, got)
			}
		})
	}

	t.Run("payload-digest", func(t *testing.T) {
		// payload digest 는 Header 밖에 있는 21번째 preimage 값이다. 항목 하나의
		// canonical payload 가 바뀌면 snapshot identity 도 바뀌어야 한다.
		swapped := []Envelope{
			{header: base[0].Header(), payload: base[0].CanonicalPayload(), digest: goldenDigestB},
			base[1],
		}
		if got := snapshotDigest(query, swapped); got == baseline {
			t.Fatalf("changing the item payload digest left the snapshot digest at %s", got)
		}
	})
}

func TestSnapshotDigestBindsEveryQueryField(t *testing.T) {
	t.Parallel()
	base := goldenItems(t)
	baseline := snapshotDigest(goldenSnapshotQuery(), base)

	mutations := map[string]func(*SnapshotQuery){
		"Market":               func(q *SnapshotQuery) { q.Market = marketclock.MarketKR },
		"Symbol":               func(q *SnapshotQuery) { q.Symbol = "MSFT" },
		"IssuerIdentity":       func(q *SnapshotQuery) { q.IssuerIdentity = "issuer-msft" },
		"IssuerMappingVersion": func(q *SnapshotQuery) { q.IssuerMappingVersion = "issuer-map-v2" },
		"EvaluationAt":         func(q *SnapshotQuery) { q.EvaluationAt = q.EvaluationAt.Add(time.Nanosecond) },
		"IngestionCutoff":      func(q *SnapshotQuery) { q.IngestionCutoff = q.IngestionCutoff.Add(time.Nanosecond) },
	}
	assertCoversEveryField(t, reflect.TypeOf(SnapshotQuery{}), mutations, "SnapshotQuery")

	for name, mutate := range mutations {
		t.Run("query/"+name, func(t *testing.T) {
			query := goldenSnapshotQuery()
			mutate(&query)
			assertMutatesOnly(t, name, goldenSnapshotQuery(), query)
			if got := snapshotDigest(query, base); got == baseline {
				t.Fatalf("changing SnapshotQuery.%s left the snapshot digest at %s", name, got)
			}
		})
	}
}

// assertCoversEveryField 는 "이 이름이 표에 있나"가 아니라 "구조체의 필드 **전부**가
// 표에 있나"를 센다. 새 필드가 생기고 결속에 안 들어가면, 그 필드의 변형 시험이
// 없다는 사실 자체로 여기서 먼저 걸린다.
func assertCoversEveryField[T any](t *testing.T, structType reflect.Type, covered map[string]T, label string) {
	t.Helper()
	for index := 0; index < structType.NumField(); index++ {
		name := structType.Field(index).Name
		if _, ok := covered[name]; !ok {
			t.Fatalf("%s.%s has no binding case in this table", label, name)
		}
	}
	if len(covered) != structType.NumField() {
		t.Fatalf("%s binding table has %d entries for %d struct fields", label, len(covered), structType.NumField())
	}
}

// assertMutatesOnly 는 표의 키가 **이름표가 아니라 역할**이게 만든다. "Currency" 로
// 적힌 칸이 실제로는 Unit 을 바꾸고 있어도 digest 는 달라지므로, 키만 보는 표는
// 두 필드 중 하나를 결속에서 빼도 통과한다(2026-09-07 독립 리뷰 지적).
func assertMutatesOnly[T any](t *testing.T, field string, before, after T) {
	t.Helper()
	value := reflect.ValueOf(before)
	changed := reflect.ValueOf(after)
	moved := []string{}
	for index := 0; index < value.NumField(); index++ {
		name := value.Type().Field(index).Name
		if !reflect.DeepEqual(value.Field(index).Interface(), changed.Field(index).Interface()) {
			moved = append(moved, name)
		}
	}
	if len(moved) != 1 || moved[0] != field {
		t.Fatalf("the %q case moved %v; a table keyed by a name it does not honour proves nothing", field, moved)
	}
}

// snapshotItemMatchesQuery 는 digest 와 **같은 규칙을 두 번째로 판정하는 자리**다.
// digest 를 직접 시험해도 이쪽은 여전히 아무도 안 묶고 있었다 — `if true { return true }`
// 로 바꿔도 패키지 스위트가 초록이었다(2026-09-07 독립 리뷰 S1). 한쪽만 묶으면 나머지가
// 상대의 시험을 대신 통과시키는 것이 이 저장소가 아는 병이고, 그것이 여기 그대로 남아
// 있었다.
//
// 여기서는 digest 를 **일부러 맞춰 놓고** 잰다. Snapshot.Valid 은 항목 범위 검사를
// digest 비교보다 먼저 하므로, digest 가 맞는데도 거짓이 나오면 그 판정은 오직 범위
// 검사에서 나온 것이다.
func TestSnapshotItemScopeIsCheckedIndependentlyOfTheDigest(t *testing.T) {
	t.Parallel()
	evaluationAt := time.Date(2026, 8, 3, 14, 0, 0, 0, time.UTC)
	ingestionCutoff := time.Date(2026, 8, 3, 15, 0, 0, 0, time.UTC)

	inScope := func() Header {
		return goldenHeader("scope-item", "record-scope", "rev-scope", "", ConfidenceVerified)
	}
	for _, test := range []struct {
		name   string
		mutate func(*Header)
	}{
		{"symbol", func(h *Header) { h.Symbol = "MSFT" }},
		{"issuer", func(h *Header) { h.IssuerIdentity = "issuer-other" }},
		{"mapping-version", func(h *Header) { h.IssuerMappingVersion = "issuer-map-v2" }},
		{"market", func(h *Header) {
			h.Market = marketclock.MarketKR
			h.Kind = KindKRNetFlow
			h.Authority = AuthorityKRX
			h.Currency = "KRW"
		}},
		{"source-event-after-evaluation", func(h *Header) {
			h.SourceEventAt = evaluationAt.Add(time.Nanosecond)
			h.SourceAvailableAt = h.SourceEventAt
			h.ObservedAt = h.SourceEventAt
			h.IngestedAt = h.SourceEventAt
		}},
		{"source-available-after-evaluation", func(h *Header) {
			h.SourceAvailableAt = evaluationAt.Add(time.Nanosecond)
			h.ObservedAt = h.SourceAvailableAt
			h.IngestedAt = h.SourceAvailableAt
		}},
		{"ingested-after-cutoff", func(h *Header) {
			h.IngestedAt = ingestionCutoff.Add(time.Nanosecond)
		}},
		{"effective-date-after-trading-day", func(h *Header) { h.MarketEffectiveDate = "2026-12-31" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			header := inScope()
			test.mutate(&header)
			items := []Envelope{mustEnvelope(t, header, goldenPayloadA)}
			query := SnapshotQuery{
				Market: marketclock.MarketUS, Symbol: "AAPL", IssuerIdentity: "issuer-aapl",
				IssuerMappingVersion: "issuer-map-v1", EvaluationAt: evaluationAt, IngestionCutoff: ingestionCutoff,
			}
			// digest 는 바로 이 항목 집합 위에서 계산한다. 그러므로 digest 비교는
			// 반드시 통과하고, 거짓은 범위 검사에서만 나올 수 있다.
			digest := snapshotDigest(query, items)
			snapshot := Snapshot{
				ID: "snapshot-" + digest, Digest: digest,
				Market: query.Market, Symbol: query.Symbol, IssuerIdentity: query.IssuerIdentity,
				IssuerMappingVersion: query.IssuerMappingVersion,
				EvaluationAt:         query.EvaluationAt, IngestionCutoff: query.IngestionCutoff,
				Items: items,
			}
			if snapshot.Valid() {
				t.Fatalf("an out-of-scope item (%s) passed with a matching digest: scope checking is not enforced", test.name)
			}
		})
	}

	// 양성 대조군: 범위 안 항목은 통과해야 한다. 이것이 없으면 "무엇이든 거짓"이
	// 위의 여덟을 전부 통과시킨다.
	t.Run("in-scope", func(t *testing.T) {
		t.Parallel()
		items := []Envelope{mustEnvelope(t, inScope(), goldenPayloadA)}
		query := SnapshotQuery{
			Market: marketclock.MarketUS, Symbol: "AAPL", IssuerIdentity: "issuer-aapl",
			IssuerMappingVersion: "issuer-map-v1", EvaluationAt: evaluationAt, IngestionCutoff: ingestionCutoff,
		}
		digest := snapshotDigest(query, items)
		snapshot := Snapshot{
			ID: "snapshot-" + digest, Digest: digest,
			Market: query.Market, Symbol: query.Symbol, IssuerIdentity: query.IssuerIdentity,
			IssuerMappingVersion: query.IssuerMappingVersion,
			EvaluationAt:         query.EvaluationAt, IngestionCutoff: query.IngestionCutoff,
			Items: items,
		}
		if !snapshot.Valid() {
			t.Fatal("an in-scope sealed snapshot was rejected; the scope check refuses normal input")
		}
	})
}
