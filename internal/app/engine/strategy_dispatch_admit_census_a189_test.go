package engine

// a189 D13 — admitSites 좁힘의 fixture 시험.
//
// 앞 판본의 admitSites 는 이름이 문(Admit · AdmitEachOwnerScope)과 같은 식별자를 전부 셌음. 그래서 a087 이 넣은 엔진 사설
// 인터페이스의 메서드 선언 lowerLimitAdmission.Admit 이 경계 자리 「type 블록:Admit」 로 잡혀 생산 census 가 깨졌음.
// 좁힌 판본이 (가) 경계 밖의 같은 이름은 세지 않고 (나) 경계 패키지의 이름은 별칭 · 함수 값으로도 여전히 세며
// (다) dot import 파일에서는 보수적으로 맨 이름을 다 센다는 것을 여기서 셈. 기존 TestAdmitCensusSeesEverySpellingItClaims 는
// 무편집으로 같은 철자 다섯을 계속 잼. gstack 리뷰로 더한 것: (라) 원시 문자열·이스케이프 import 경로 해독, (마) 엔진 생산
// 파일의 unsafe import 금지, (바) 모듈 전체의 경계 문 //go:linkname 금지 — 셋 다 철자를 세는 census 밖에서 문에 닿는 길임.

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

// 원시 문자열(`…`)·이스케이프(\x68) 철자의 경계 import 도 경계로 알아봄 — 따옴표만 벗기던 판본은 둘 다 놓쳤음(a189 리뷰).
// gofmt 가 두 철자를 큰따옴표로 정규화하지만, census 는 lint 를 건너뛴 소스에서도 같은 답을 내야 함.
func TestA189AdmitCensusDecodesTheImportPathLiteral(t *testing.T) {
	for name, imp := range map[string]string{
		"raw-string": "sh `example.com/internal/strategyhandoff`",
		"escaped":    `sh "example.com/internal/strategy\x68andoff"`,
	} {
		fixture := "package engine\n\nimport " + imp + "\n\nfunc second() { _ = sh.Admit(true, nil) }\n"
		if got, want := a189AdmitSites(t, fixture), "second:Admit"; got != want {
			t.Errorf("%s: admitSites=[%s], want [%s] — 경계 import 철자를 못 알아봄", name, got, want)
		}
	}
}

// 엔진 생산 파일은 unsafe 를 import 하지 않음 — //go:linkname 으로 경계 문을 다른 이름의 몸체 없는 함수에 묶으면 선택자도
// 맨 이름도 남지 않아 admitSites 가 못 셈(a189 리뷰, 이름 기반 앞 판본도 이름이 다르면 못 셈). linkname 지시어는 unsafe
// import 가 있는 파일에서만 컴파일되므로 이 금지는 엔진 패키지 안의 linkname 을 닫음. 다른 패키지에 둔 linkname 은
// 아래 TestA189NoPackageLinknamesTheSeamDoors 가 닫음(이 금지만으로는 안 닫힘 — gstack 리뷰 2회차 실측).
func TestA189EngineProductionNeverImportsUnsafe(t *testing.T) {
	files := engineProductionFiles(t)
	for _, path := range files {
		for _, spec := range parseEngineFile(t, path).Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil || importPath == "unsafe" {
				t.Errorf("%s imports %s — unsafe 는 //go:linkname 으로 경계 문을 census 밖에서 부를 길을 엶", filepath.Base(path), spec.Path.Value)
			}
		}
	}
}

// 모듈의 어느 비시험 Go 파일도 //go:linkname 으로 경계(strategyhandoff)의 이름에 묶지 않음. 엔진 밖 패키지가 unsafe 를 들여와
// 문을 다른 이름으로 다시 내보내면 strategyhandoff 를 import 하지 않으므로 TestOnlyTheEngineImportsThisSeam 도, 엔진이
// 그 이름을 부르면 admitSites 도 못 봄(gstack 리뷰 2회차 실측). unsafe 는 모듈의 다른 자리(strategyprojectionrpc)가 정당하게
// 쓰므로 모듈 전체 import 금지 대신 지시어의 대상 이름을 봄.
func TestA189NoPackageLinknamesTheSeamDoors(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	scanned := 0
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if path != root && (strings.HasPrefix(name, ".") || name == "testdata" || name == "node_modules" || name == "vendor") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, perr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments) // 지시어는 import 뒤에 오므로 ImportsOnly 로 자르면 안 보임
		if perr != nil {
			// 조용히 건너뛰면 그 파일이 문이 됨 — 못 읽는 Go 파일은 실패시킴
			t.Errorf("parse %s: %v", path, perr)
			return nil
		}
		scanned++
		for _, group := range file.Comments {
			for _, c := range group.List {
				fields := strings.Fields(c.Text)
				if len(fields) >= 3 && fields[0] == "//go:linkname" && strings.Contains(fields[2], "internal/strategyhandoff.") {
					t.Errorf("%s: %s — 몸체 없는 별칭이 admit census 밖에서 경계 문에 닿음", path, c.Text)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk module: %v", err)
	}
	if scanned < 500 {
		t.Fatalf("only %d non-test Go files scanned — the linkname census would be partial", scanned)
	}
}
