package engine

// a189 D13 — admitSites 좁힘의 fixture 시험.
//
// 앞 판본의 admitSites 는 이름이 문(Admit · AdmitEachOwnerScope)과 같은 식별자를 전부 셌음. 그래서 a087 이 넣은 엔진 사설
// 인터페이스의 메서드 선언 lowerLimitAdmission.Admit 이 경계 자리 「type 블록:Admit」 로 잡혀 생산 census 가 깨졌음.
// 좁힌 판본이 (가) 경계 밖의 같은 이름은 세지 않고 (나) 경계 패키지의 이름은 별칭 · 함수 값으로도 여전히 세며
// (다) dot import 파일에서는 보수적으로 맨 이름을 다 센다는 것을 여기서 셈. 기존 TestAdmitCensusSeesEverySpellingItClaims 는
// 무편집으로 같은 철자 다섯을 계속 잼.

import (
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"testing"
)

// a189AdmitSites 는 fixture 하나를 파싱해 admitSites 결과를 정렬해 돌려줌.
func a189AdmitSites(t *testing.T, fixture string) string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "fixture.go", fixture, 0)
	if err != nil {
		t.Fatalf("fixture 파싱 실패: %v", err)
	}
	got := admitSites(file, map[string]bool{"Admit": true, "AdmitEachOwnerScope": true})
	sort.Strings(got)
	return strings.Join(got, ", ")
}

// 경계 밖의 같은 이름 다섯 모양은 세지 않고, 별칭 import 로 쓴 경계 이름 둘(호출 · 함수 값)만 셈.
func TestA189AdmitCensusIgnoresSameNamedMembersOutsideTheSeam(t *testing.T) {
	const fixture = `package engine

import (
	"context"

	sh "example.com/internal/strategyhandoff"
)

// 1) 엔진 사설 인터페이스의 메서드 선언 — a087 lowerLimitAdmission 과 같은 모양(세지 않음)
type lowerLimitAdmission interface{ Admit(context.Context) error }

// 2) 같은 이름 필드(세지 않음)
type minter struct{ Admit func() }

// 3) 같은 이름 메서드 선언(세지 않음)
func (m minter) Admit() {}

type board struct{ admission lowerLimitAdmission }

// 4) 다른 값의 같은 이름 메서드 호출 — 선택자 X 가 경계 패키지 이름이 아님(세지 않음)
func benign(ctx context.Context, b board) error { return b.admission.Admit(ctx) }

// 5) 지역 변수의 같은 이름 메서드 값(세지 않음)
func local(admission lowerLimitAdmission) { _ = admission.Admit }

// 6) 별칭으로 부른 경계 문(셈)
func viaAlias() { _ = sh.Admit(true, nil) }

// 7) 별칭으로 빼 둔 경계 문의 함수 값(셈 — 호출이 아니라 참조를 셈)
func viaValue() { f := sh.AdmitEachOwnerScope; _ = f }
`
	want := "viaAlias:Admit, viaValue:AdmitEachOwnerScope"
	if got := a189AdmitSites(t, fixture); got != want {
		t.Fatalf("admitSites=[%s], want [%s] — 경계 밖의 같은 이름을 세거나 경계 이름을 놓침", got, want)
	}
}

// 별칭 없는 import 의 기본 지역 이름 strategyhandoff 도 셈(생산 엔진의 철자).
func TestA189AdmitCensusCountsTheUnaliasedSeamName(t *testing.T) {
	const fixture = `package engine

import "github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"

func dispatchHandoff() { _, _ = strategyhandoff.Admit(true, nil) }

var spare = strategyhandoff.Admit
`
	want := "dispatchHandoff:Admit, spare:Admit"
	if got := a189AdmitSites(t, fixture); got != want {
		t.Fatalf("admitSites=[%s], want [%s]", got, want)
	}
}

// dot import 가 있는 파일은 맨 이름을 보수적으로 다 셈 — 인터페이스 메서드 선언도 셈(실패 쪽으로 틀림).
// 선택자의 Sel 자리(b.Admit)는 dot import 파일에서도 세지 않음.
func TestA189AdmitCensusIsConservativeInADotImportedFile(t *testing.T) {
	const fixture = `package engine

import . "example.com/other"

type gate interface{ Admit() }

func bare() { _ = Admit(true, nil) }

func viaValue(b gate) { b.Admit() }
`
	want := "bare:Admit, type 블록:Admit"
	if got := a189AdmitSites(t, fixture); got != want {
		t.Fatalf("admitSites=[%s], want [%s]", got, want)
	}
}
