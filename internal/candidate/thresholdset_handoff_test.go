package candidate

import (
	"errors"
	"reflect"
	"testing"
)

func TestAssessApprovedCandidateFailsClosedForAllAndMixedVetoStates(t *testing.T) {
	set := loadSyntheticThresholdSet(t)
	pass := passingApprovedInputs(aCandidate(t0))
	allRaised := pass
	allRaised.Sighting = Sighting{Measured: true, Rank: 1, RankTotal: 100}
	allRaised.Expansion.LastPrice = "200"
	allRaised.Range = RangePosition{Measured: true, High: "120", Price: "119", At: pass.At}
	allUnmeasured := pass
	allUnmeasured.Sighting = Sighting{}
	allUnmeasured.Expansion = Expansion{}
	allUnmeasured.Range = RangePosition{}
	mixed := pass
	mixed.Sighting = allRaised.Sighting
	mixed.Range = RangePosition{}

	for _, tc := range []struct {
		name     string
		input    VetoInputs
		wantKind ApprovalErrorKind
		want     []VetoCode
	}{
		{name: "all raised", input: allRaised, wantKind: ApprovalVetoRaised,
			want: []VetoCode{VetoSeenLate, VetoExtended, VetoNearHigh}},
		{name: "all unmeasured", input: allUnmeasured, wantKind: ApprovalVetoUnmeasured,
			want: []VetoCode{VetoSeenLate, VetoExtended, VetoNearHigh}},
		{name: "raised takes precedence over unmeasured", input: mixed, wantKind: ApprovalVetoRaised,
			want: []VetoCode{VetoSeenLate}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := AssessApprovedCandidate(tc.input, set)
			if got != (ApprovedCandidate{}) {
				t.Fatalf("refused approval = %+v, want exact zero", got)
			}
			var approvalErr *ApprovalError
			if !errors.As(err, &approvalErr) {
				t.Fatalf("error = %T %v, want *ApprovalError", err, err)
			}
			if approvalErr.Kind() != tc.wantKind || !reflect.DeepEqual(approvalErr.Vetoes(), tc.want) {
				t.Fatalf("refusal = kind:%q vetoes:%v, want kind:%q vetoes:%v",
					approvalErr.Kind(), approvalErr.Vetoes(), tc.wantKind, tc.want)
			}
		})
	}
}
