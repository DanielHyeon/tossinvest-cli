package weeklyvaluelane

import (
	"reflect"
	"testing"
)

// 이 파일은 **weekly 경로의 위임 한 칸**을 증명한다.
//
// 1차 진입 backstop(`internal/app/engine`)은 봉인된 execution terms identity 두
// 개를 비교해서 위조를 막는다. 그 identity 는 정책에서 `policy.identity` 하나만
// 해시한다 — 그러니 `ExecutionPolicy` 의 나머지 여덟 필드는 그 identity 가
// 자기들의 함수일 때만 backstop 이 지킨다.
//
// continuation·reversal 은 그 여덟을 비운 채 정책을 만들므로 잃을 값이 없다.
// **weekly 는 아홉을 다 채운다**(`strategyflow/adapters.go` 의 `weeklyPolicy`).
// 그러니 위임이 성립하는지는 이 패키지가 답해야 하고, 11차 적대 리뷰가 추적해
// 보니 성립한다: `Identity` 는 `{seal, DecisionDigest, CalendarDigest,
// CapSnapshotID}` 에서 나오고, `seal` 은 나머지 다섯 값 필드를 담는다.
//
// 성립한다는 것과 **못 박혀 있다**는 것은 다르다. 여기 표가 없으면
// `weeklyExecutionPreimageSeal` 의 preimage 에서 필드 하나를 빼는 편집이
// backstop 의 사정거리에서 그 값을 조용히 빼면서 모든 시험을 초록으로 통과한다.

// TestWeeklyExecutionPolicyIdentityChangesWithEveryFieldItCovers 는
// `RRExecutionPolicy` 의 여덟 필드를 **하나씩** 바꿔 `Identity` 가 달라지는지 본다.
//
// 값 다섯은 preimage 를 바꾸고 **다시 봉인해서** 넣는다. 봉인 없이 바꾼 preimage
// 는 `valid()` 가 거절하므로(`v.seal == weeklyExecutionPreimageSeal(v)`) 상류로
// 갈 수 없다 — 여기서 재봉인하는 것이 곧 그 경로를 흉내 내는 것이다.
func TestWeeklyExecutionPolicyIdentityChangesWithEveryFieldItCovers(t *testing.T) {
	basePreimage := func() ExecutionTermsPreimage {
		v := ExecutionTermsPreimage{planDigest: "sha256:plan", evidenceDigest: "sha256:evidence",
			entry: "100", staged: "120", fair: "115", entryCosts: "3", exitCosts: "2", minimumRR: 2_000_000}
		v.seal = weeklyExecutionPreimageSeal(v)
		return v
	}
	baseLineage := ResultLineage{DecisionDigest: "sha256:decision", CalendarDigest: "sha256:calendar",
		CapSnapshotID: "cap-1"}
	want := executionPolicy(basePreimage(), baseLineage).Identity
	if want == "" {
		t.Fatal("기준 정책의 Identity 가 비었다 — 아래 케이스가 전부 공허해진다")
	}
	reseal := func(mutate func(*ExecutionTermsPreimage)) func(*ExecutionTermsPreimage, *ResultLineage) {
		return func(v *ExecutionTermsPreimage, _ *ResultLineage) {
			mutate(v)
			v.seal = weeklyExecutionPreimageSeal(*v)
		}
	}
	cases := map[string]func(*ExecutionTermsPreimage, *ResultLineage){
		"StagedTargetMinor": reseal(func(v *ExecutionTermsPreimage) { v.staged += "9" }),
		"FairValueMinor":    reseal(func(v *ExecutionTermsPreimage) { v.fair += "9" }),
		"EntryCostsMinor":   reseal(func(v *ExecutionTermsPreimage) { v.entryCosts += "9" }),
		"ExitCostsMinor":    reseal(func(v *ExecutionTermsPreimage) { v.exitCosts += "9" }),
		"MinimumRRPPM":      reseal(func(v *ExecutionTermsPreimage) { v.minimumRR++ }),
		"DecisionDigest":    func(_ *ExecutionTermsPreimage, l *ResultLineage) { l.DecisionDigest += "0" },
		"CalendarDigest":    func(_ *ExecutionTermsPreimage, l *ResultLineage) { l.CalendarDigest += "0" },
		"CapSnapshotID":     func(_ *ExecutionTermsPreimage, l *ResultLineage) { l.CapSnapshotID += "-other" },
	}
	policy := reflect.TypeOf(RRExecutionPolicy{})
	if len(cases) != policy.NumField()-1 {
		t.Fatalf("표가 %d 개를 세는데 RRExecutionPolicy 는 %d 필드다(Identity 제외 %d)",
			len(cases), policy.NumField(), policy.NumField()-1)
	}
	if _, ok := policy.FieldByName("Identity"); !ok {
		t.Fatal("Identity 필드가 없다 — 위 개수에서 뺀 것이 그것이라는 근거가 사라진다")
	}
	for name, mutate := range cases {
		if _, ok := policy.FieldByName(name); !ok {
			t.Fatalf("케이스 이름 %q 가 RRExecutionPolicy 의 필드가 아니다", name)
		}
		t.Run(name, func(t *testing.T) {
			preimage, lineage := basePreimage(), baseLineage
			mutate(&preimage, &lineage)
			moved := executionPolicy(preimage, lineage)
			if moved.Identity == want {
				t.Fatalf("%s 를 바꿨는데 Identity 가 같다 — 그 값은 1차 진입 backstop 의 사정거리 밖이다", name)
			}
		})
	}
}
