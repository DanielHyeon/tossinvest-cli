package candidate

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

var syntheticActivationEvidence = []byte("synthetic a046 activation evidence; not a numeric approval")

func thresholdJSON(market, session, evidenceDigest, seenLate string) string {
	return fmt.Sprintf(`{
  "version":"candidate-veto-2026-07-31.1",
  "market":%q,
  "session":%q,
  "metrics":[
    {"key":"seen_late","definition":"first-sighting rank percentile","value":%q},
    {"key":"extended","definition":"gain from stored first price","value":"50"},
    {"key":"near_high","definition":"distance below intraday high","value":"2.0"}
  ],
  "sample_window":{"from":"2026-07-01T00:00:00Z","to":"2026-07-31T00:00:00Z"},
  "sample_count":100,
  "missing_rate":"0.1",
  "evidence_digest":%q
}`, market, session, seenLate, evidenceDigest)
}

func syntheticActivationJSON(version, market, session, setDigest, evidenceDigest string, approvedAt time.Time) string {
	return fmt.Sprintf(`{
  "version":%q,
  "market":%q,
  "session":%q,
  "set_digest":%q,
  "evidence_digest":%q,
  "approved_at":%q,
  "approved_by":"synthetic-human-review-record-not-an-approval"
}`, version, market, session, setDigest, evidenceDigest, approvedAt.UTC().Format(time.RFC3339))
}

func syntheticApprovedInputs(t *testing.T, market, session, seenLate string) (string, ActivationRecord) {
	t.Helper()
	evidenceDigest := DigestEvidence(syntheticActivationEvidence)
	document := thresholdJSON(market, session, evidenceDigest, seenLate)
	scope := ThresholdScope{Market: market, Session: session}
	setDigest, err := DigestThresholdSetDocument(strings.NewReader(document), scope)
	if err != nil {
		t.Fatalf("DigestThresholdSetDocument synthetic fixture: %v", err)
	}
	record, err := LoadActivationRecord(strings.NewReader(syntheticActivationJSON(
		"candidate-veto-2026-07-31.1", market, session, setDigest, evidenceDigest,
		time.Date(2026, 7, 31, 1, 0, 0, 0, time.UTC))))
	if err != nil {
		t.Fatalf("LoadActivationRecord synthetic fixture: %v", err)
	}
	return document, record
}

func loadSyntheticThresholdSet(t *testing.T) ThresholdSet {
	t.Helper()
	document, activation := syntheticApprovedInputs(t, MarketKR, SessionRegular, "80")
	set, err := LoadThresholdSet(strings.NewReader(document), syntheticActivationEvidence, activation,
		ThresholdScope{Market: MarketKR, Session: SessionRegular},
		time.Date(2026, 7, 31, 1, 0, 0, 0, time.UTC), time.Minute)
	if err != nil {
		t.Fatalf("LoadThresholdSet synthetic fixture: %v", err)
	}
	return set
}

func TestThresholdSetLoaderBindsOpaqueEvidenceCanonicalSetAndActivation(t *testing.T) {
	set := loadSyntheticThresholdSet(t)
	if set.Version() != "candidate-veto-2026-07-31.1" || set.EvidenceDigest() != DigestEvidence(syntheticActivationEvidence) {
		t.Fatalf("loaded metadata = version:%q evidence:%q", set.Version(), set.EvidenceDigest())
	}
	if set.SetDigest() == "" || set.ApprovedBy() != "synthetic-human-review-record-not-an-approval" {
		t.Fatalf("activation metadata not bound: set_digest=%q approved_by=%q", set.SetDigest(), set.ApprovedBy())
	}
	got := set.VetoThresholds()
	if got.SeenLatePercentilePct != "80" || got.ExtendedGainPct != "50" || got.NearHighDistancePct != "2.0" {
		t.Fatalf("thresholds = %+v", got)
	}
}

func TestThresholdSetLoaderFailsClosedForMissingOrMismatchedBinding(t *testing.T) {
	scope := ThresholdScope{Market: MarketKR, Session: SessionRegular}
	document, activation := syntheticApprovedInputs(t, MarketKR, SessionRegular, "80")
	asOf := time.Date(2026, 7, 31, 1, 0, 0, 0, time.UTC)
	wrongDocument, wrongSetActivation := syntheticApprovedInputs(t, MarketKR, SessionRegular, "81")
	_ = wrongDocument
	loadRecord := func(raw string) ActivationRecord {
		t.Helper()
		record, err := LoadActivationRecord(strings.NewReader(raw))
		if err != nil {
			t.Fatalf("synthetic activation: %v", err)
		}
		return record
	}
	wrongVersionActivation := loadRecord(syntheticActivationJSON("candidate-veto-2026-07-31.2",
		MarketKR, SessionRegular, activation.SetDigest(), activation.EvidenceDigest(), activation.ApprovedAt()))
	wrongScopeActivation := loadRecord(syntheticActivationJSON(activation.Version(),
		MarketUS, SessionRegular, activation.SetDigest(), activation.EvidenceDigest(), activation.ApprovedAt()))
	wrongEvidenceActivation := loadRecord(syntheticActivationJSON(activation.Version(),
		MarketKR, SessionRegular, activation.SetDigest(), DigestEvidence([]byte("other evidence")), activation.ApprovedAt()))

	for _, tc := range []struct {
		name       string
		document   string
		evidence   []byte
		activation ActivationRecord
		expected   ThresholdScope
		asOf       time.Time
		skew       time.Duration
		want       string
	}{
		{name: "missing evidence", document: document, activation: activation, expected: scope, asOf: asOf, skew: time.Minute, want: "evidence bytes"},
		{name: "evidence digest mismatch", document: document, evidence: []byte("different opaque evidence"), activation: activation, expected: scope, asOf: asOf, skew: time.Minute, want: "evidence digest"},
		{name: "missing activation", document: document, evidence: syntheticActivationEvidence, expected: scope, asOf: asOf, skew: time.Minute, want: "activation"},
		{name: "set digest mismatch", document: document, evidence: syntheticActivationEvidence, activation: wrongSetActivation, expected: scope, asOf: asOf, skew: time.Minute, want: "set digest"},
		{name: "activation version mismatch", document: document, evidence: syntheticActivationEvidence, activation: wrongVersionActivation, expected: scope, asOf: asOf, skew: time.Minute, want: "version"},
		{name: "activation scope mismatch", document: document, evidence: syntheticActivationEvidence, activation: wrongScopeActivation, expected: scope, asOf: asOf, skew: time.Minute, want: "activation scope"},
		{name: "activation evidence mismatch", document: document, evidence: syntheticActivationEvidence, activation: wrongEvidenceActivation, expected: scope, asOf: asOf, skew: time.Minute, want: "activation evidence"},
		{name: "wrong expected market", document: document, evidence: syntheticActivationEvidence, activation: activation, expected: ThresholdScope{Market: MarketUS, Session: SessionRegular}, asOf: asOf, skew: time.Minute, want: "scope"},
		{name: "negative skew", document: document, evidence: syntheticActivationEvidence, activation: activation, expected: scope, asOf: asOf, skew: -time.Second, want: "skew"},
		{name: "missing asOf", document: document, evidence: syntheticActivationEvidence, activation: activation, expected: scope, skew: time.Minute, want: "asof"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var zero ThresholdSet
			got, err := LoadThresholdSet(strings.NewReader(tc.document), tc.evidence, tc.activation,
				tc.expected, tc.asOf, tc.skew)
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.want)) {
				t.Fatalf("LoadThresholdSet error = %v, want containing %q", err, tc.want)
			}
			if got != zero {
				t.Fatalf("rejected binding returned usable state: %+v", got)
			}
		})
	}
}

func TestThresholdDocumentRemainsStrictCompleteAndScopeSpecific(t *testing.T) {
	scope := ThresholdScope{Market: MarketKR, Session: SessionRegular}
	document, activation := syntheticApprovedInputs(t, MarketKR, SessionRegular, "80")
	asOf := time.Date(2026, 7, 31, 1, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name string
		doc  string
		want string
	}{
		{name: "wrong market", doc: strings.Replace(document, `"market":"KR"`, `"market":"US"`, 1), want: "scope"},
		{name: "wrong session", doc: strings.Replace(document, `"session":"regular"`, `"session":"extended-hours"`, 1), want: "scope"},
		{name: "missing sample count", doc: strings.Replace(document, `"sample_count":100`, `"sample_count":0`, 1), want: "sample_count"},
		{name: "missing digest", doc: strings.Replace(document, DigestEvidence(syntheticActivationEvidence), "", 1), want: "evidence_digest"},
		{name: "partial metrics", doc: strings.Replace(document, `,
    {"key":"extended","definition":"gain from stored first price","value":"50"}`, "", 1), want: "extended"},
		{name: "unknown field", doc: strings.Replace(document, `"version":`, `"invented":true,"version":`, 1), want: "unknown"},
		{name: "duplicate field", doc: strings.Replace(document, `"version":`, `"version":"shadowed","version":`, 1), want: "duplicate"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var zero ThresholdSet
			got, err := LoadThresholdSet(strings.NewReader(tc.doc), syntheticActivationEvidence, activation,
				scope, asOf, time.Minute)
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.want)) || got != zero {
				t.Fatalf("LoadThresholdSet = (%+v, %v), want zero and %q error", got, err, tc.want)
			}
		})
	}
}

func TestCanonicalSetDigestIgnoresJSONWhitespaceAndMetricInputOrder(t *testing.T) {
	evidenceDigest := DigestEvidence(syntheticActivationEvidence)
	document := thresholdJSON(MarketKR, SessionRegular, evidenceDigest, "80")
	reordered := strings.Replace(document, `    {"key":"seen_late","definition":"first-sighting rank percentile","value":"80"},
    {"key":"extended","definition":"gain from stored first price","value":"50"},
    {"key":"near_high","definition":"distance below intraday high","value":"2.0"}`,
		`    {"key":"near_high","definition":"distance below intraday high","value":"2.0"},
    {"key":"seen_late","definition":"first-sighting rank percentile","value":"80"},
    {"key":"extended","definition":"gain from stored first price","value":"50"}`, 1)
	reordered = strings.ReplaceAll(reordered, "  ", "    ")
	scope := ThresholdScope{Market: MarketKR, Session: SessionRegular}
	first, err := DigestThresholdSetDocument(strings.NewReader(document), scope)
	if err != nil {
		t.Fatal(err)
	}
	second, err := DigestThresholdSetDocument(strings.NewReader(reordered), scope)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("canonical digest depends on JSON layout/order: %q != %q", first, second)
	}
}

func TestActivationRecordIsStrictAndApprovalTimeIsBounded(t *testing.T) {
	document, activation := syntheticApprovedInputs(t, MarketKR, SessionRegular, "80")
	scope := ThresholdScope{Market: MarketKR, Session: SessionRegular}
	asOf := time.Date(2026, 7, 31, 1, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		name string
		json string
		want string
	}{
		{name: "unknown field", json: strings.Replace(syntheticActivationJSON(activation.Version(), MarketKR, SessionRegular,
			activation.SetDigest(), activation.EvidenceDigest(), activation.ApprovedAt()), `"version":`, `"invented":true,"version":`, 1), want: "unknown"},
		{name: "trailing value", json: syntheticActivationJSON(activation.Version(), MarketKR, SessionRegular,
			activation.SetDigest(), activation.EvidenceDigest(), activation.ApprovedAt()) + `{}`, want: "trailing"},
		{name: "duplicate field", json: strings.Replace(syntheticActivationJSON(activation.Version(), MarketKR, SessionRegular,
			activation.SetDigest(), activation.EvidenceDigest(), activation.ApprovedAt()),
			`"version":`, `"version":"shadowed","version":`, 1), want: "duplicate"},
		{name: "missing approver", json: strings.Replace(syntheticActivationJSON(activation.Version(), MarketKR, SessionRegular,
			activation.SetDigest(), activation.EvidenceDigest(), activation.ApprovedAt()),
			`"approved_by":"synthetic-human-review-record-not-an-approval"`, `"approved_by":""`, 1), want: "approved_by"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var zero ActivationRecord
			got, err := LoadActivationRecord(strings.NewReader(tc.json))
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), tc.want) || got != zero {
				t.Fatalf("LoadActivationRecord = (%+v, %v), want zero and %q error", got, err, tc.want)
			}
		})
	}

	setDigest := activation.SetDigest()
	evidenceDigest := activation.EvidenceDigest()
	for _, tc := range []struct {
		name       string
		approvedAt time.Time
		asOf       time.Time
		want       string
	}{
		{name: "before sample window end", approvedAt: time.Date(2026, 7, 30, 23, 59, 59, 0, time.UTC), asOf: asOf, want: "sample_window.to"},
		{name: "beyond explicit future skew", approvedAt: asOf.Add(time.Minute + time.Second), asOf: asOf, want: "asof"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			record, err := LoadActivationRecord(strings.NewReader(syntheticActivationJSON(
				activation.Version(), MarketKR, SessionRegular, setDigest, evidenceDigest, tc.approvedAt)))
			if err != nil {
				t.Fatal(err)
			}
			var zero ThresholdSet
			got, err := LoadThresholdSet(strings.NewReader(document), syntheticActivationEvidence, record,
				scope, tc.asOf, time.Minute)
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.want)) || got != zero {
				t.Fatalf("LoadThresholdSet = (%+v, %v), want zero and %q error", got, err, tc.want)
			}
		})
	}
}

func TestThresholdRegistryRejectsSameVersionWithDifferentCanonicalDigest(t *testing.T) {
	registry := NewThresholdRegistry()
	asOf := time.Date(2026, 7, 31, 1, 0, 0, 0, time.UTC)
	scope := ThresholdScope{Market: MarketKR, Session: SessionRegular}
	firstDoc, firstActivation := syntheticApprovedInputs(t, MarketKR, SessionRegular, "80")
	if _, err := registry.LoadThresholdSet(strings.NewReader(firstDoc), syntheticActivationEvidence,
		firstActivation, scope, asOf, time.Minute); err != nil {
		t.Fatalf("first registry load: %v", err)
	}
	secondDoc, secondActivation := syntheticApprovedInputs(t, MarketKR, SessionRegular, "81")
	var zero ThresholdSet
	got, err := registry.LoadThresholdSet(strings.NewReader(secondDoc), syntheticActivationEvidence,
		secondActivation, scope, asOf, time.Minute)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "same version") || got != zero {
		t.Fatalf("conflicting registry load = (%+v, %v), want zero same-version error", got, err)
	}
}

func TestThresholdRegistryConcurrentSameVersionSameDigestIsIdempotent(t *testing.T) {
	registry := NewThresholdRegistry()
	document, activation := syntheticApprovedInputs(t, MarketKR, SessionRegular, "80")
	scope := ThresholdScope{Market: MarketKR, Session: SessionRegular}
	asOf := time.Date(2026, 7, 31, 1, 0, 0, 0, time.UTC)
	const workers = 16
	errors := make(chan error, workers)
	var wait sync.WaitGroup
	for range workers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := registry.LoadThresholdSet(strings.NewReader(document), syntheticActivationEvidence,
				activation, scope, asOf, time.Minute)
			errors <- err
		}()
	}
	wait.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Errorf("idempotent concurrent registry load: %v", err)
		}
	}
}

func TestThresholdSetDoesNotAliasTheSourceEvidenceOrActivation(t *testing.T) {
	document, activation := syntheticApprovedInputs(t, MarketKR, SessionRegular, "80")
	documentBytes := []byte(document)
	evidence := append([]byte(nil), syntheticActivationEvidence...)
	set, err := LoadThresholdSet(strings.NewReader(string(documentBytes)), evidence, activation,
		ThresholdScope{Market: MarketKR, Session: SessionRegular},
		time.Date(2026, 7, 31, 1, 0, 0, 0, time.UTC), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	for index := range documentBytes {
		documentBytes[index] = 'x'
	}
	for index := range evidence {
		evidence[index] = 'x'
	}
	if set.Version() != "candidate-veto-2026-07-31.1" || set.VetoThresholds().NearHighDistancePct != "2.0" {
		t.Fatalf("loaded immutable set changed after input mutation: version=%q thresholds=%+v",
			set.Version(), set.VetoThresholds())
	}
}

func passingApprovedInputs(candidate Candidate) VetoInputs {
	at := candidate.FirstSeenAt.Add(time.Minute)
	return VetoInputs{
		Candidate: candidate,
		Sighting:  Sighting{Measured: true, Rank: 90, RankTotal: 100},
		Expansion: Expansion{Measured: true, FirstPrice: "100", LastPrice: "110",
			FirstAt: candidate.FirstSeenAt, LastAt: at},
		Range: RangePosition{Measured: true, High: "120", Price: "100", At: at},
		At:    at,
	}
}

func TestAssessApprovedCandidateFailsClosed(t *testing.T) {
	set := loadSyntheticThresholdSet(t)
	pass := passingApprovedInputs(aCandidate(t0))
	dangerous := pass
	dangerous.Range = RangePosition{Measured: true, High: "120", Price: "119", At: pass.At}
	unmeasured := pass
	unmeasured.Range = RangePosition{}
	wrongMarket := pass
	wrongMarket.Candidate.Market = MarketUS
	invalidLife := pass
	invalidLife.Candidate.Symbol = ""
	zeroFirstSeen := pass
	zeroFirstSeen.Candidate.FirstSeenAt = time.Time{}

	for _, tc := range []struct {
		name       string
		input      VetoInputs
		set        ThresholdSet
		wantKind   ApprovalErrorKind
		wantVetoes []VetoCode
	}{
		{name: "invalid set", input: pass, wantKind: ApprovalInvalidSet},
		{name: "wrong market", input: wrongMarket, set: set, wantKind: ApprovalScopeMismatch},
		{name: "invalid candidate life", input: invalidLife, set: set, wantKind: ApprovalInvalidCandidateLife},
		{name: "zero first seen", input: zeroFirstSeen, set: set, wantKind: ApprovalInvalidCandidateLife},
		{name: "dangerous", input: dangerous, set: set, wantKind: ApprovalVetoRaised,
			wantVetoes: []VetoCode{VetoNearHigh}},
		{name: "unmeasured", input: unmeasured, set: set, wantKind: ApprovalVetoUnmeasured,
			wantVetoes: []VetoCode{VetoNearHigh}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := AssessApprovedCandidate(tc.input, tc.set)
			if got != (ApprovedCandidate{}) {
				t.Fatalf("refused approval = %+v, want exact zero value", got)
			}
			var approvalErr *ApprovalError
			if !errors.As(err, &approvalErr) {
				t.Fatalf("error = %T %v, want *ApprovalError", err, err)
			}
			if approvalErr.Kind() != tc.wantKind {
				t.Fatalf("error kind = %q, want %q", approvalErr.Kind(), tc.wantKind)
			}
			if !reflect.DeepEqual(approvalErr.Vetoes(), tc.wantVetoes) {
				t.Fatalf("error vetoes = %v, want %v", approvalErr.Vetoes(), tc.wantVetoes)
			}
		})
	}
}

func TestAssessApprovedCandidateReturnsPassWithImmutableProvenance(t *testing.T) {
	set := loadSyntheticThresholdSet(t)
	input := passingApprovedInputs(aCandidate(t0))
	got, err := AssessApprovedCandidate(input, set)
	if err != nil {
		t.Fatalf("AssessApprovedCandidate: %v", err)
	}
	if !got.Valid() || !got.Chase().Passed() {
		t.Fatalf("approved verdict = %+v, want valid measured-and-clear pass", got)
	}
	if got.Key() != input.Candidate.Key || !got.FirstSeenAt().Equal(input.Candidate.FirstSeenAt) {
		t.Fatalf("approved life = (%+v, %v), want (%+v, %v)",
			got.Key(), got.FirstSeenAt(), input.Candidate.Key, input.Candidate.FirstSeenAt)
	}
	if got.ThresholdVersion() != set.Version() || got.SetDigest() != set.SetDigest() ||
		got.EvidenceDigest() != set.EvidenceDigest() || !got.ApprovedAt().Equal(set.ApprovedAt()) {
		t.Fatalf("approved provenance = version:%q set:%q evidence:%q at:%v",
			got.ThresholdVersion(), got.SetDigest(), got.EvidenceDigest(), got.ApprovedAt())
	}

	typ := reflect.TypeOf(ApprovedCandidate{})
	for index := range typ.NumField() {
		field := typ.Field(index)
		if field.IsExported() {
			t.Errorf("ApprovedCandidate field %s is exported; provenance/pass invariant is mutable", field.Name)
		}
	}
}

func TestApprovedCandidateLifeIdentityUsesKeyAndFirstSeenAt(t *testing.T) {
	set := loadSyntheticThresholdSet(t)
	wantPayload := "tossos:candidate-life:v1\x00" + MarketKR + "\x00" + "005930" + "\x00" +
		t0.UTC().Format(time.RFC3339Nano)
	wantSum := sha256.Sum256([]byte(wantPayload))
	wantID := CandidateLifeID("candidate-life:v1:sha256:" + hex.EncodeToString(wantSum[:]))

	approve := func(t *testing.T, firstSeen time.Time) ApprovedCandidate {
		t.Helper()
		candidate := aCandidate(firstSeen)
		got, err := AssessApprovedCandidate(passingApprovedInputs(candidate), set)
		if err != nil {
			t.Fatalf("AssessApprovedCandidate: %v", err)
		}
		return got
	}
	one := approve(t, t0)
	sameInstant := approve(t, t0.In(time.FixedZone("same-instant", 9*60*60)))
	differentLife := approve(t, t0.Add(time.Nanosecond))
	if one.CandidateLifeID() != wantID || sameInstant.CandidateLifeID() != wantID {
		t.Fatalf("same life IDs = (%q, %q), want %q", one.CandidateLifeID(), sameInstant.CandidateLifeID(), wantID)
	}
	if differentLife.CandidateLifeID() == wantID {
		t.Fatalf("different FirstSeenAt reused candidate-life ID %q", wantID)
	}
	if strings.Contains(string(one.CandidateLifeID()), inputSymbolForTest) {
		t.Fatalf("candidate-life ID %q exposes raw symbol", one.CandidateLifeID())
	}
}

const inputSymbolForTest = "005930"

func TestThresholdDescriptorsAreDormantAndLegacyNearHighIsNotEffective(t *testing.T) {
	descriptors := CandidateFilterDescriptors()
	if len(descriptors) != len([]string{MarketKR, MarketUS})*len(OrderedVetoCodes()) {
		t.Fatalf("descriptors = %d, want market x veto matrix", len(descriptors))
	}
	for _, d := range descriptors {
		if d.Category != "candidate-filters" || !d.ReadOnly || d.DefaultState != ThresholdStateUnapproved ||
			d.DesiredValue != "" || d.EffectiveValue != "" || d.SampleState != "not_measured" ||
			d.EvidenceState != "not_measured" || !d.CASRequired || d.ApplyTiming == "" {
			t.Errorf("descriptor is not dormant/fail-closed: %+v", d)
		}
		if d.Key == string(VetoNearHigh) {
			if d.LegacyValue != LegacyNearHighThresholdPct || d.Provenance != "legacy-unapproved" {
				t.Errorf("near_high provenance = %+v", d)
			}
		} else if d.LegacyValue != "" {
			t.Errorf("%s invented legacy value %q", d.Key, d.LegacyValue)
		}
	}
}
