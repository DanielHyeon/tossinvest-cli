package strategyevidence

import (
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestProjectKeepsFatalAndLaneEvidenceSeparate(t *testing.T) {
	t.Parallel()
	tradability := validHeader(marketclock.MarketUS, KindTradability, "halt", "rev-1")
	participation := validHeader(marketclock.MarketUS, KindUSParticipation, "score", "rev-1")
	participation.SourceRecordID = "score-1"
	snapshot := Snapshot{
		Market: marketclock.MarketUS, Symbol: "AAPL", IssuerIdentity: participation.IssuerIdentity, IssuerMappingVersion: participation.IssuerMappingVersion, EvaluationAt: participation.ObservedAt.Add(time.Minute),
		Items: []Envelope{
			mustEnvelope(t, tradability, `{"blocked":true,"code":"TRADING_HALT"}`),
			mustEnvelope(t, participation, `{"score_ppm":999999}`),
		},
	}
	result, err := Project(snapshot, ProjectionPolicy{
		Version: "policy-v1", Market: marketclock.MarketUS,
		Fatal:    []Requirement{{Kind: KindTradability, Required: true, MaxAge: time.Hour, Authorities: []SourceAuthority{AuthoritySEC}}},
		Required: []Requirement{{Kind: KindUSParticipation, Required: true, MaxAge: time.Hour, Authorities: []SourceAuthority{AuthoritySEC}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Fatal.Blocked || result.Fatal.Reasons[0].Code != RefusalFatalEvidence {
		t.Fatalf("fatal = %+v", result.Fatal)
	}
	if !result.Lane.Eligible || len(result.Lane.Evidence) != 1 {
		t.Fatalf("lane evidence should remain independently projected: %+v", result.Lane)
	}
}

func TestProjectRejectsFatalPayloadWithoutExplicitBlockedState(t *testing.T) {
	t.Parallel()
	h := validHeader(marketclock.MarketUS, KindTradability, "halt", "rev-1")
	snapshot := Snapshot{
		Market: marketclock.MarketUS, Symbol: "AAPL", IssuerIdentity: h.IssuerIdentity, IssuerMappingVersion: h.IssuerMappingVersion, EvaluationAt: h.ObservedAt.Add(time.Minute),
		Items: []Envelope{mustEnvelope(t, h, `{"code":"TRADING_HALT"}`)},
	}
	result, err := Project(snapshot, ProjectionPolicy{
		Version: "policy-v1", Market: marketclock.MarketUS,
		Fatal: []Requirement{{Kind: KindTradability, Required: true, MaxAge: time.Hour, Authorities: []SourceAuthority{AuthoritySEC}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Fatal.Blocked || !hasRefusal(result.Fatal.Reasons, RefusalPayloadInvalid) {
		t.Fatalf("fatal payload without blocked field did not fail closed: %+v", result.Fatal)
	}
}

func TestProjectRequiredMissingAndStaleFailClosedWhileOptionalIsTypedUnavailable(t *testing.T) {
	t.Parallel()
	h := validHeader(marketclock.MarketKR, KindKRNetFlow, "flow", "rev")
	h.Authority = AuthorityKRX
	h.Currency = "KRW"
	snapshot := Snapshot{Market: marketclock.MarketKR, Symbol: "AAPL", IssuerIdentity: h.IssuerIdentity, IssuerMappingVersion: h.IssuerMappingVersion, EvaluationAt: h.ObservedAt.Add(2 * time.Hour), Items: []Envelope{mustEnvelope(t, h, `{"flow":1}`)}}
	result, err := Project(snapshot, ProjectionPolicy{
		Version: "policy-v1", Market: marketclock.MarketKR,
		Required: []Requirement{
			{Kind: KindKRNetFlow, Required: true, MaxAge: time.Minute, Authorities: []SourceAuthority{AuthorityKRX}},
			{Kind: KindDisclosure, Required: true, MaxAge: time.Hour, Authorities: []SourceAuthority{AuthorityOpenDART}},
		},
		Optional: []Requirement{{Kind: KindDisclosureRisk, MaxAge: time.Hour, Authorities: []SourceAuthority{AuthorityOpenDART}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Lane.Eligible {
		t.Fatal("required stale/missing evidence was accepted")
	}
	if !hasRefusal(result.Lane.Refusals, RefusalEvidenceStale) || !hasRefusal(result.Lane.Refusals, RefusalEvidenceMissing) {
		t.Fatalf("refusals = %+v", result.Lane.Refusals)
	}
	if result.Lane.Unavailable[KindDisclosureRisk].Code != RefusalEvidenceMissing {
		t.Fatalf("optional unavailable = %+v", result.Lane.Unavailable)
	}
}

func TestProjectRequiredEvidenceConflictsAndScopeMismatchesFailClosed(t *testing.T) {
	t.Parallel()
	base := validHeader(marketclock.MarketUS, KindUSParticipation, "one", "rev-1")
	base.SourceRecordID = "record-one"
	other := base
	other.EvidenceID = "two"
	other.SourceRecordID = "record-two"

	tests := []struct {
		name        string
		items       func(*testing.T) []Envelope
		requirement Requirement
		want        RefusalCode
	}{
		{
			name: "conflicting authoritative records",
			items: func(t *testing.T) []Envelope {
				return []Envelope{mustEnvelope(t, base, `{"value":1}`), mustEnvelope(t, other, `{"value":2}`)}
			},
			requirement: Requirement{Kind: KindUSParticipation, Required: true, MaxAge: time.Hour, Authorities: []SourceAuthority{AuthoritySEC}},
			want:        RefusalEvidenceConflict,
		},
		{
			name:        "currency mismatch",
			items:       func(t *testing.T) []Envelope { return []Envelope{mustEnvelope(t, base, `{"value":1}`)} },
			requirement: Requirement{Kind: KindUSParticipation, Required: true, MaxAge: time.Hour, Authorities: []SourceAuthority{AuthoritySEC}, Currency: "KRW"},
			want:        RefusalCurrencyUnresolved,
		},
		{
			name:        "unit mismatch",
			items:       func(t *testing.T) []Envelope { return []Envelope{mustEnvelope(t, base, `{"value":1}`)} },
			requirement: Requirement{Kind: KindUSParticipation, Required: true, MaxAge: time.Hour, Authorities: []SourceAuthority{AuthoritySEC}, Unit: "shares"},
			want:        RefusalUnitUnresolved,
		},
		{
			name: "identity mismatch",
			items: func(t *testing.T) []Envelope {
				mismatched := base
				mismatched.Symbol = "MSFT"
				return []Envelope{mustEnvelope(t, mismatched, `{"value":1}`)}
			},
			requirement: Requirement{Kind: KindUSParticipation, Required: true, MaxAge: time.Hour, Authorities: []SourceAuthority{AuthoritySEC}},
			want:        RefusalIdentityMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			snapshot := Snapshot{Market: marketclock.MarketUS, Symbol: "AAPL", IssuerIdentity: base.IssuerIdentity, IssuerMappingVersion: base.IssuerMappingVersion, EvaluationAt: base.ObservedAt.Add(time.Minute), Items: tt.items(t)}
			result, err := Project(snapshot, ProjectionPolicy{Version: "policy-v1", Market: marketclock.MarketUS, Required: []Requirement{tt.requirement}})
			if err != nil {
				t.Fatal(err)
			}
			if result.Lane.Eligible || !hasRefusal(result.Lane.Refusals, tt.want) {
				t.Fatalf("want fail-closed %s, got %+v", tt.want, result.Lane)
			}
		})
	}
}

func hasRefusal(refusals []Refusal, code RefusalCode) bool {
	for _, refusal := range refusals {
		if refusal.Code == code {
			return true
		}
	}
	return false
}
