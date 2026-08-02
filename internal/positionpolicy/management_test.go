package positionpolicy

import (
	"testing"
	"time"
)

func TestProjectManagementUsesOneStablePriorityTable(t *testing.T) {
	now := time.Date(2026, 8, 2, 1, 2, 3, 0, time.UTC)
	accountBlock := ReconcileBlock{
		Scope: ScopeAccount, Reason: "RECONCILE_PERMANENT",
		Detail: "quantity mismatch", StartedAt: now, Permanent: true,
	}
	effective := NewAdoptionSettings(true, 0.03,
		[]string{"AAPL", "BOTH", "MANAGED"}, []string{"EXCLUDED", "BOTH", "MANAGED"}, "")
	runtime := ManagementRuntime{
		Effective: effective, EffectiveKnown: true,
		Blocks: []ReconcileBlock{accountBlock}, BlockSource: AdoptionBlockingTrackerProjection,
	}
	tests := []struct {
		name                                                      string
		input                                                     ManagementInput
		want                                                      ManagementStatus
		known, designated, included, excluded, candidate, blocked bool
	}{
		{name: "journal unknown wins", input: ManagementInput{Symbol: "AAPL", Runtime: runtime},
			want: ManagementStatusUnknown},
		{name: "managed wins over exclude and block", input: ManagementInput{JournalKnown: true, Managed: true, Symbol: "MANAGED", Runtime: runtime},
			want: ManagementStatusManaged, known: true, designated: true, included: true, excluded: true},
		{name: "operator release wins over candidate and block", input: ManagementInput{JournalKnown: true, Released: true, Market: "us", Symbol: "AAPL", Runtime: runtime},
			want: ManagementStatusUnmanaged, known: true, designated: true, included: true, candidate: true},
		{name: "exclude wins over include and block", input: ManagementInput{JournalKnown: true, Symbol: "BOTH", Runtime: runtime},
			want: ManagementStatusExcluded, known: true, designated: true, included: true, excluded: true},
		{name: "candidate covered by block", input: ManagementInput{JournalKnown: true, Market: "us", Symbol: "AAPL", Runtime: runtime},
			want: ManagementStatusReconcileBlocked, known: true, designated: true, included: true, candidate: true, blocked: true},
		{name: "candidate pending", input: ManagementInput{JournalKnown: true, Market: "us", Symbol: "AAPL", Runtime: ManagementRuntime{Effective: effective, EffectiveKnown: true}},
			want: ManagementStatusAdoptionPending, known: true, designated: true, included: true, candidate: true},
		{name: "unselected", input: ManagementInput{JournalKnown: true, Symbol: "NONE", Runtime: ManagementRuntime{Effective: NewAdoptionSettings(false, .05, nil, nil, ""), EffectiveKnown: true}},
			want: ManagementStatusUnmanaged, known: true, designated: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProjectManagement(tt.input)
			if got.Status != tt.want || got.StatusKnown != tt.known || got.DesignationKnown != tt.designated ||
				got.Included != tt.included || got.Excluded != tt.excluded || got.Candidate != tt.candidate ||
				(got.Block != nil) != tt.blocked {
				t.Fatalf("projection=%+v, want status=%s known=%v designation=%v included=%v excluded=%v candidate=%v blocked=%v",
					got, tt.want, tt.known, tt.designated, tt.included, tt.excluded, tt.candidate, tt.blocked)
			}
		})
	}
}

func TestProjectManagementNamesOperatorReleasedLifecycle(t *testing.T) {
	got := ProjectManagement(ManagementInput{JournalKnown: true, Released: true, Symbol: "AAPL",
		Runtime: ManagementRuntime{EffectiveKnown: true,
			Effective: NewAdoptionSettings(false, .05, []string{"AAPL"}, nil, "")}})
	if got.Status != ManagementStatusUnmanaged || !got.StatusKnown ||
		got.Reason != ManagementReasonOperatorReleased || got.Label != "관리 외(운영자 해제)" ||
		!got.DesignationKnown || !got.Included || !got.Candidate || got.Block != nil {
		t.Fatalf("projection=%+v", got)
	}
}

func TestProjectManagementKeepsRuntimeUnavailableExplicit(t *testing.T) {
	managed := ProjectManagement(ManagementInput{JournalKnown: true, Managed: true, Symbol: "AAPL"})
	if managed.Status != ManagementStatusManaged || !managed.StatusKnown || managed.DesignationKnown {
		t.Fatalf("managed runtime-unavailable projection=%+v", managed)
	}
	unknown := ProjectManagement(ManagementInput{JournalKnown: true, Symbol: "AAPL"})
	if unknown.Status != ManagementStatusUnknown || unknown.StatusKnown || unknown.DesignationKnown ||
		unknown.Reason != ManagementReasonRuntimeUnavailable {
		t.Fatalf("unmanaged runtime-unavailable projection=%+v", unknown)
	}
}

func TestReconcileBlockCoverageIsScopeAwareAndFailClosed(t *testing.T) {
	tests := []struct {
		block          ReconcileBlock
		market, symbol string
		want           bool
	}{
		{block: ReconcileBlock{Scope: ScopeAccount}, market: "us", symbol: "AAPL", want: true},
		{block: ReconcileBlock{Scope: ScopeMarket, Market: "US"}, market: "us", symbol: "AAPL", want: true},
		{block: ReconcileBlock{Scope: ScopeSymbol, Market: "us", Symbol: "aapl"}, market: "US", symbol: "AAPL", want: true},
		{block: ReconcileBlock{Scope: ReconcileScope("future")}, market: "us", symbol: "AAPL", want: true},
	}
	for _, tt := range tests {
		if got := tt.block.Covers(tt.market, tt.symbol); got != tt.want {
			t.Fatalf("block=%+v covers=%v want=%v", tt.block, got, tt.want)
		}
	}
}

func TestNewAdoptionSettingsCopiesAndNormalizesLists(t *testing.T) {
	include := []string{" aapl ", "AAPL", "", "msft"}
	exclude := []string{" tsla "}
	got := NewAdoptionSettings(false, .05, include, exclude, " rejected\nreason ")
	include[0], exclude[0] = "MUTATED", "MUTATED"
	if len(got.IncludeSymbols) != 2 || got.IncludeSymbols[0] != "AAPL" || got.IncludeSymbols[1] != "MSFT" ||
		len(got.ExcludeSymbols) != 1 || got.ExcludeSymbols[0] != "TSLA" || got.Rejected != "rejected reason" {
		t.Fatalf("settings=%+v", got)
	}
}
