package candidate

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

const testEvidenceDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func approvedThresholdJSON(market, session string) string {
	return fmt.Sprintf(`{
  "version":"candidate-veto-2026-07-31.1",
  "market":%q,
  "session":%q,
  "metrics":[
    {"key":"seen_late","definition":"first-sighting rank percentile","value":"80"},
    {"key":"extended","definition":"gain from stored first price","value":"50"},
    {"key":"near_high","definition":"distance below intraday high","value":"2.0"}
  ],
  "sample_window":{"from":"2026-07-01T00:00:00Z","to":"2026-07-31T00:00:00Z"},
  "sample_count":100,
  "missing_rate":"0.1",
  "evidence_digest":%q,
  "approved_at":"2026-07-31T01:00:00Z",
  "approved_by":"human-review-record-1"
}`, market, session, testEvidenceDigest)
}

func TestThresholdSetLoaderAcceptsOnlyCompleteApprovedExpectedScope(t *testing.T) {
	scope := ThresholdScope{Market: MarketKR, Session: SessionRegular}
	set, err := LoadThresholdSet(strings.NewReader(approvedThresholdJSON(MarketKR, SessionRegular)), scope)
	if err != nil {
		t.Fatalf("LoadThresholdSet: %v", err)
	}
	if set.Version() != "candidate-veto-2026-07-31.1" || set.EvidenceDigest() != testEvidenceDigest {
		t.Fatalf("loaded metadata = version:%q digest:%q", set.Version(), set.EvidenceDigest())
	}
	got := set.VetoThresholds()
	if got.SeenLatePercentilePct != "80" || got.ExtendedGainPct != "50" || got.NearHighDistancePct != "2.0" {
		t.Fatalf("thresholds = %+v", got)
	}
}

func TestThresholdSetLoaderFailsClosedForIncompleteWrongMarketAndUnapproved(t *testing.T) {
	scope := ThresholdScope{Market: MarketKR, Session: SessionRegular}
	complete := approvedThresholdJSON(MarketKR, SessionRegular)
	for _, tc := range []struct {
		name string
		doc  string
		want string
	}{
		{name: "wrong market", doc: approvedThresholdJSON(MarketUS, SessionRegular), want: "scope"},
		{name: "wrong session", doc: approvedThresholdJSON(MarketKR, "extended-hours"), want: "scope"},
		{name: "missing sample count", doc: strings.Replace(complete, `"sample_count":100`, `"sample_count":0`, 1), want: "sample_count"},
		{name: "missing digest", doc: strings.Replace(complete, testEvidenceDigest, "", 1), want: "evidence_digest"},
		{name: "unapproved", doc: strings.Replace(complete, `"approved_by":"human-review-record-1"`, `"approved_by":""`, 1), want: "approved_by"},
		{name: "partial metrics", doc: strings.Replace(complete, `,
    {"key":"extended","definition":"gain from stored first price","value":"50"}`, "", 1), want: "extended"},
		{name: "unknown field", doc: strings.Replace(complete, `"version":`, `"invented":true,"version":`, 1), want: "unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var zero ThresholdSet
			got, err := LoadThresholdSet(strings.NewReader(tc.doc), scope)
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.want)) {
				t.Fatalf("LoadThresholdSet error = %v, want containing %q", err, tc.want)
			}
			if got != zero {
				t.Fatalf("rejected document returned usable state: %+v", got)
			}
		})
	}
}

func TestThresholdSetDoesNotAliasTheSourceDocument(t *testing.T) {
	document := []byte(approvedThresholdJSON(MarketKR, SessionRegular))
	set, err := LoadThresholdSet(strings.NewReader(string(document)),
		ThresholdScope{Market: MarketKR, Session: SessionRegular})
	if err != nil {
		t.Fatal(err)
	}
	for index := range document {
		document[index] = 'x'
	}
	if set.Version() != "candidate-veto-2026-07-31.1" || set.VetoThresholds().NearHighDistancePct != "2.0" {
		t.Fatalf("loaded immutable set changed after input mutation: version=%q thresholds=%+v",
			set.Version(), set.VetoThresholds())
	}
}

func TestApprovedCandidateCarriesThresholdVersionWithoutOrderAuthority(t *testing.T) {
	set, err := LoadThresholdSet(strings.NewReader(approvedThresholdJSON(MarketKR, SessionRegular)),
		ThresholdScope{Market: MarketKR, Session: SessionRegular})
	if err != nil {
		t.Fatal(err)
	}
	at := t0.Add(time.Minute)
	got, err := AssessApprovedCandidate(VetoInputs{
		Candidate: aCandidate(t0),
		Sighting:  Sighting{Measured: true, Rank: 90, RankTotal: 100},
		Expansion: Expansion{Measured: true, FirstPrice: "100", LastPrice: "110",
			FirstAt: t0, LastAt: at},
		Range: RangePosition{Measured: true, High: "120", Price: "100", At: at},
		At:    at,
	}, set)
	if err != nil {
		t.Fatalf("AssessApprovedCandidate: %v", err)
	}
	if !got.Chase.Passed() || got.ThresholdVersion != set.Version() {
		t.Fatalf("approved verdict = %+v, want pass attributed to %q", got, set.Version())
	}
	if got.Key != (Key{Market: MarketKR, Symbol: "005930"}) {
		t.Fatalf("approved key = %+v", got.Key)
	}
}

func TestThresholdDescriptorsAreDormantAndLegacyNearHighIsNotEffective(t *testing.T) {
	descriptors := CandidateFilterDescriptors()
	if len(descriptors) != len([]string{MarketKR, MarketUS})*len(VetoCodes) {
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
