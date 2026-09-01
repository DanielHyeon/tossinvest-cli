package strategyflow

import (
	"reflect"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// 이 파일은 **"terms identity 는 모든 필드의 함수다"** 하나만 증명한다.
//
// 왜 여기인가. 1차 진입 backstop(`internal/app/engine`)이 이 identity 두 개를
// 비교해서 위조를 막는다. 그 비교가 무엇을 지키는지는 **이 해시가 무엇을 담느냐**
// 로 정해진다. 5~9차 적대 리뷰는 엔진 쪽에서 축을 하나씩 시험으로 쪼개 봤고,
// 매 라운드 다음 축이 남았다 — `executionTermsIdentity` 가 해시하는 스칼라가
// 32 개이기 때문이다(아래 표가 그 32 개다). 엔진에 32 개 fixture 를 두는 것은
// 끝나지 않는 일이고, 그 대부분은 시험 seam 이 축 하나만 움직이지도 못한다.
//
// 필드를 가진 타입 안에서는 그 일이 싸고 **끝난다.** 여기서는 비공개 필드를
// 직접 만질 수 있으므로 fixture 도 seam 도 필요 없다.

func baseExecutionTerms() ExecutionTerms {
	price := func(minor, source string) PriceProvenance {
		return PriceProvenance{priceMinor: minor, source: source, version: "v1", digest: "sha256:aaa",
			asOf: "2026-09-01T00:00:00Z", currency: "KRW", minorScale: 0, unitVersion: "minor-v1"}
	}
	return ExecutionTerms{accountRef: "acct-1", market: strategyrouter.MarketKR, symbol: "005930",
		campaignID: "campaign-1", legOrdinal: 1, quantity: 8,
		entry: price("100", "entry-contract"), stop: price("95", "stop-contract"), target: price("120", "target-contract"),
		policy:          ExecutionPolicy{identity: "policy-identity-1"},
		lineageIdentity: "lineage-identity-1"}
}

// TestExecutionTermsIdentityChangesWithEveryFieldItCovers 는 identity 가 담는
// 스칼라 32 개를 **하나씩** 바꿔 해시가 달라지는지 본다.
//
// 하나라도 안 담기면 그 필드는 1차 진입 backstop 이 지키지 않는 값이 된다 —
// 9차 리뷰가 `stop.digest` 로 그 모양을 보였다(그때는 엔진 쪽 시험이 세 필드를
// 함께 옮겨서 못 잡았다).
func TestExecutionTermsIdentityChangesWithEveryFieldItCovers(t *testing.T) {
	base := baseExecutionTerms()
	want := executionTermsIdentity(base)
	priceOf := map[string]func(*ExecutionTerms) *PriceProvenance{
		"entry":  func(terms *ExecutionTerms) *PriceProvenance { return &terms.entry },
		"stop":   func(terms *ExecutionTerms) *PriceProvenance { return &terms.stop },
		"target": func(terms *ExecutionTerms) *PriceProvenance { return &terms.target },
	}
	priceField := map[string]func(*PriceProvenance){
		"priceMinor":  func(p *PriceProvenance) { p.priceMinor += "9" },
		"source":      func(p *PriceProvenance) { p.source += "-other" },
		"version":     func(p *PriceProvenance) { p.version += "-other" },
		"digest":      func(p *PriceProvenance) { p.digest += "0" },
		"asOf":        func(p *PriceProvenance) { p.asOf = "2026-09-02T00:00:00Z" },
		"currency":    func(p *PriceProvenance) { p.currency = "USD" },
		"minorScale":  func(p *PriceProvenance) { p.minorScale++ },
		"unitVersion": func(p *PriceProvenance) { p.unitVersion += "-other" },
	}
	cases := map[string]func(*ExecutionTerms){
		"accountRef":      func(terms *ExecutionTerms) { terms.accountRef += "-other" },
		"market":          func(terms *ExecutionTerms) { terms.market = strategyrouter.MarketUS },
		"symbol":          func(terms *ExecutionTerms) { terms.symbol = "000660" },
		"campaignID":      func(terms *ExecutionTerms) { terms.campaignID += "-other" },
		"legOrdinal":      func(terms *ExecutionTerms) { terms.legOrdinal++ },
		"quantity":        func(terms *ExecutionTerms) { terms.quantity++ },
		"policy.identity": func(terms *ExecutionTerms) { terms.policy.identity += "-other" },
		"lineageIdentity": func(terms *ExecutionTerms) { terms.lineageIdentity += "-other" },
	}
	for priceName, pick := range priceOf {
		for fieldName, mutate := range priceField {
			mutate, pick := mutate, pick
			cases[priceName+"."+fieldName] = func(terms *ExecutionTerms) { mutate(pick(terms)) }
		}
	}
	if len(cases) != 32 {
		t.Fatalf("이 표가 세는 스칼라 수 = %d, want 32", len(cases))
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			moved := baseExecutionTerms()
			mutate(&moved)
			if moved == base {
				t.Fatal("뮤테이션이 아무것도 안 바꿨다 — 이 케이스는 아무것도 증명하지 못한다")
			}
			if got := executionTermsIdentity(moved); got == want {
				t.Fatalf("%s 를 바꿨는데 identity 가 같다 — 이 필드는 identity 가 담지 않는다", name)
			}
		})
	}
}

// TestBreakoutPolicyIdentityChangesWithEveryFieldItCovers 는 정책의 나머지
// 여덟 필드가 **위임으로** 덮이는지 본다.
//
// `executionTermsIdentity` 는 정책에서 `policy.identity` 하나만 해시한다.
// `ExecutionPolicy` 는 아홉 필드이므로 나머지 여덟은 그 identity 가 자기들의
// 함수일 때만 덮인다 — `lineageIdentity` 가 lineage 를 대신하는 것과 같은 위임이다.
// 10차 적대 리뷰가 위 표의 산술이 이 타입을 아예 안 읽는다는 것을 보였다.
//
// **위임이 확인되는 범위를 정확히 적는다.** TossOS 안에서 정책 identity 를
// 유도하는 곳은 `breakoutPolicyIdentity` 하나이고 아래가 그것을 잰다.
// continuation·reversal 경로는 `ExecutionPolicy{identity: …}` 로 나머지 여덟을
// 비운 채 만들므로(`adapters.go` 의 `fromContinuation`·`fromReversal`) 잃을 값이
// 없다. **weekly 경로는 다르다** — `weeklyPolicy` 가 아홉 필드를 다 채우면서
// identity 는 lane 이 계산한 것을 그대로 받는다. 그 유도는 이 패키지가 확인하지
// 않는다. 확인 안 된 위임으로 남긴다.
func TestBreakoutPolicyIdentityChangesWithEveryFieldItCovers(t *testing.T) {
	base := ExecutionPolicy{stagedTargetMinor: "120", fairValueMinor: "115", entryCostsMinor: "3",
		exitCostsMinor: "2", minimumRRPPM: 2_000_000, decisionDigest: "sha256:d", calendarDigest: "sha256:c",
		capSnapshotID: "cap-1"}
	want := breakoutPolicyIdentity(base)
	if want == "" {
		t.Fatal("기준 정책의 identity 가 비었다 — 아래 케이스가 전부 공허해진다")
	}
	cases := map[string]func(*ExecutionPolicy){
		"stagedTargetMinor": func(p *ExecutionPolicy) { p.stagedTargetMinor += "9" },
		"fairValueMinor":    func(p *ExecutionPolicy) { p.fairValueMinor += "9" },
		"entryCostsMinor":   func(p *ExecutionPolicy) { p.entryCostsMinor += "9" },
		"exitCostsMinor":    func(p *ExecutionPolicy) { p.exitCostsMinor += "9" },
		"minimumRRPPM":      func(p *ExecutionPolicy) { p.minimumRRPPM++ },
		"decisionDigest":    func(p *ExecutionPolicy) { p.decisionDigest += "0" },
		"calendarDigest":    func(p *ExecutionPolicy) { p.calendarDigest += "0" },
		"capSnapshotID":     func(p *ExecutionPolicy) { p.capSnapshotID += "-other" },
	}
	policy := reflect.TypeOf(ExecutionPolicy{})
	if _, ok := policy.FieldByName("identity"); !ok {
		t.Fatal("identity 필드가 없다 — 아래 개수에서 뺀 것이 그것이라는 근거가 사라진다")
	}
	if len(cases) != policy.NumField()-1 {
		t.Fatalf("이 표가 세는 정책 필드 = %d, ExecutionPolicy 는 %d 필드(identity 제외 %d)",
			len(cases), policy.NumField(), policy.NumField()-1)
	}
	// **이름으로** 맞춘다. 개수만 맞추면 필드 하나를 더하고 하나를 빼는 편집이
	// 그대로 통과한다 — 11차 적대 리뷰가 그 잔여를 지적했다.
	for index := 0; index < policy.NumField(); index++ {
		name := policy.Field(index).Name
		if name == "identity" {
			continue
		}
		if _, covered := cases[name]; !covered {
			t.Fatalf("%s 가 이 표에 없다 — 위임을 세는 범위 밖이다", name)
		}
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			moved := base
			mutate(&moved)
			if moved == base {
				t.Fatal("뮤테이션이 아무것도 안 바꿨다 — 이 케이스는 아무것도 증명하지 못한다")
			}
			if got := breakoutPolicyIdentity(moved); got == want {
				t.Fatalf("%s 를 바꿨는데 정책 identity 가 같다 — 그 필드는 terms identity 로 위임되지 않는다", name)
			}
		})
	}
}

// TestExecutionTermsIdentityTableCoversEveryDeclaredField 는 위 두 표가 **선언된
// 필드 전부**를 세는지 본다.
//
// 표가 32 개를 세는 것과 타입이 32 개를 가진 것은 다른 말이다. 나중에 필드가
// 하나 늘면 표는 그대로 32 개라 초록으로 남는다 — 그 순간 이 파일의 완전성
// 주장은 거짓이 된다. 개수를 **세 타입 모두**에서 읽어 못 박는다.
//
// 앞 판본은 `ExecutionPolicy` 를 스칼라 하나로 하드코딩해서, 그 타입에 필드를
// 더해도 초록이었다(10차 적대 리뷰). 세는 범위가 곧 완전성의 범위다.
func TestExecutionTermsIdentityTableCoversEveryDeclaredField(t *testing.T) {
	terms := reflect.TypeOf(ExecutionTerms{})
	price := reflect.TypeOf(PriceProvenance{})
	policy := reflect.TypeOf(ExecutionPolicy{})
	// terms.identity 는 해시의 입력이 아니라 출력이므로 뺀다. 가격 셋은 필드가
	// 아니라 여덟 개짜리 구조체로 센다.
	direct := terms.NumField() - 1 - 3 + 3*price.NumField()
	// policy.identity 는 직접 해시되고(위 direct 에 이미 포함), 나머지는 그
	// identity 를 통해 위임된다.
	delegated := policy.NumField() - 1
	if direct != 32 || delegated != 8 {
		t.Fatalf("직접 %d(want 32) · 위임 %d(want 8) — ExecutionTerms %d · PriceProvenance %d · ExecutionPolicy %d 필드",
			direct, delegated, terms.NumField(), price.NumField(), policy.NumField())
	}
}
