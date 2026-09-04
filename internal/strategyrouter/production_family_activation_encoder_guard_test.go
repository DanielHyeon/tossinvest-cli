package strategyrouter

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// 태스크 8.8.3 [a112 결정 61]: **매니페스트 바이트를 만드는 능력은 도구 하나만 갖는다.**
//
// `EncodeProductionFamilyActivation` 은 exported 여야 한다 — 저장소 밖 도구가
// 아니라 `tools/` 의 도구가 부르는데, 그 도구는 별도 패키지이기 때문이다. 그래서
// "이 패키지 밖에서는 영값만 만들 수 있다" 는 성질을 타입 가시성으로 지킬 수 없고,
// 대신 **참조를 센다**.
//
// 왜 이것이 안전에 관계되는가. 인코더 자체는 아무것도 승격하지 못한다(바이트를 낼
// 뿐이고 env 핀이 그 바이트를 가리켜야 무언가 켜진다). 그러나 생산 코드가 그것을
// 부르기 시작하면 "매니페스트는 프로세스 밖에서 태어난다" 가 더 이상 참이 아니게 되고,
// 그때부터 이 문장은 주석일 뿐 계약이 아니다.
//
// 문자열 검색이 아니라 **import 별칭을 해석한 선택자**로 센다 — 이름이 있는지만 보는
// 가드는 별칭 하나로 뚫린다(4.3.2 가 같은 이유로 같은 모양이다).
const familyActivationEncoderImportPath = "github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"

// familyActivationAuthoringSymbols 는 바이트를 만드는 표면 전부다.
// 인코더만 세면 문서 타입을 채워 자기 직렬화기로 내보내는 우회가 남는다.
var familyActivationAuthoringSymbols = []string{
	"EncodeProductionFamilyActivation",
	"FamilyActivationDocument",
}

// 이 둘만이 바이트를 만들 수 있다. 다른 곳은 전부 위반이다.
var familyActivationAuthoringAllowed = map[string]bool{
	// 정의 자체.
	filepath.Join("internal", "strategyrouter", "production_family_activation.go"): true,
	// 사람이 매니페스트를 만드는 유일한 도구.
	filepath.Join("tools", "a112-family-activation", "main.go"): true,
}

func TestOnlyTheAuthoringToolCanBuildActivationBytes(t *testing.T) {
	root := filepath.Join("..", "..")
	fset := token.NewFileSet()
	seen := map[string]int{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "vendor", "node_modules", ".sdd":
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		parsed, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		counts := countFamilyActivationAuthoringReferences(parsed, relative)
		for symbol, count := range counts {
			if count == 0 {
				continue
			}
			seen[relative+"\x00"+symbol] = count
			if !familyActivationAuthoringAllowed[relative] {
				t.Errorf("%s 가 활성화 매니페스트 바이트를 만든다 (%s 참조 %d개).\n"+
					"매니페스트는 프로세스 밖에서 태어나야 한다 — 만드는 자리는 "+
					"tools/a112-family-activation 하나다 [a112 결정 61]", relative, symbol, count)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// 대조군은 **심볼마다** 센다. 합집합으로 세면 심볼 하나가 이름이 바뀌어
	// 사라져도 나머지가 대신 세어져 이 가드가 조용히 통과한다 — 반증 M7 이
	// 정확히 그렇게 살아남았다. 셈 단위를 (파일 × 심볼) 로 올려서 막는다.
	for allowed := range familyActivationAuthoringAllowed {
		for _, symbol := range familyActivationAuthoringSymbols {
			if seen[allowed+"\x00"+symbol] == 0 {
				t.Errorf("%s 에 %s 참조가 하나도 없다 — 이름이 바뀌었다면 이 가드가 "+
					"세는 대상(familyActivationAuthoringSymbols)도 함께 바뀌어야 한다",
					allowed, symbol)
			}
		}
	}
}

// countFamilyActivationAuthoringReferences 는 그 파일이 저작 표면을 몇 번 부르는지 센다.
//
// 같은 패키지(정의 파일이 사는 곳)에서는 맨 이름으로, 다른 패키지에서는 해석된
// 별칭의 선택자로 나타난다. 둘 다 센다 — 한쪽만 세면 다른 쪽이 통째로 안 보인다.
func countFamilyActivationAuthoringReferences(file *ast.File, relative string) map[string]int {
	// 맨 이름으로 부를 수 있는 파일은 둘이다: 정의가 사는 패키지 안, 그리고 점
	// import 를 한 파일. 뒤엣것을 빼먹은 것이 아래 P2 였다.
	bare := filepath.Dir(relative) == filepath.Join("internal", "strategyrouter")
	aliases, dotImported := familyActivationRouterNames(file)
	if dotImported {
		bare = true
	}
	counts := map[string]int{}
	ast.Inspect(file, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.SelectorExpr:
			ident, ok := value.X.(*ast.Ident)
			if !ok || !aliases[ident.Name] {
				return true
			}
			if familyActivationAuthoringSymbol(value.Sel.Name) {
				counts[value.Sel.Name]++
			}
		case *ast.Ident:
			// 점 import 인 파일에서는 선택자의 Sel 도 여기 한 번 더 걸려 두 번
			// 세어질 수 있다. 판정은 "0 인가" 뿐이라 과다 계수는 결론을 바꾸지
			// 않으므로 그대로 둔다 — 갈래를 늘리는 쪽이 더 틀리기 쉽다.
			if bare && familyActivationAuthoringSymbol(value.Name) {
				counts[value.Name]++
			}
		}
		return true
	})
	return counts
}

func familyActivationAuthoringSymbol(name string) bool {
	for _, symbol := range familyActivationAuthoringSymbols {
		if name == symbol {
			return true
		}
	}
	return false
}

// familyActivationRouterNames 는 이 파일이 강령 패키지를 부를 수 있는 **모든**
// 이름과, 이름 없이(점 import) 부르는지를 낸다.
//
// 앞 판본은 첫 import spec 하나만 보고 `.` 이나 `_` 를 만나면 "이 파일은 안 부른다"
// 로 답했다. 그래서 두 우회가 가드를 그냥 지나갔다(2026-09-04 독립 적대 리뷰 P2-1,
// 둘 다 실제로 컴파일해 확인했다).
//
//   - 점 import 는 맨 이름으로 부르는데, 바로 그 갈래를 껐다.
//   - `_ "…/strategyrouter"` 를 진짜 별칭보다 **먼저** 적으면 뒤의 별칭을 못 봤다.
//
// 기전은 5.5·M7 과 같다 — 세는 범위가 입력의 함수라서, 입력을 고르면 셈이 0 이 된다.
// 그래서 spec 하나가 아니라 전부를 보고, 이름 집합과 점 여부를 함께 낸다.
func familyActivationRouterNames(file *ast.File) (map[string]bool, bool) {
	aliases := map[string]bool{}
	dotImported := false
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil || path != familyActivationEncoderImportPath {
			continue
		}
		switch {
		case spec.Name == nil:
			aliases["strategyrouter"] = true
		case spec.Name.Name == ".":
			dotImported = true
		case spec.Name.Name == "_":
			// 이 spec 으로는 못 부른다. 여기서 멈추지 않고 다음 spec 을 본다.
		default:
			aliases[spec.Name.Name] = true
		}
	}
	return aliases, dotImported
}

// 가드가 세는 범위 자체를 시험한다 [P2-1 수정, 2026-09-04 독립 적대 리뷰].
//
// 위 워크 시험은 저장소에 위반이 **없다**는 것만 잰다. 그것은 가드가 실제로 볼 줄
// 아는지에 대해 아무 말도 하지 않는다 — 아무것도 못 보는 가드도 똑같이 초록이다.
// 리뷰가 컴파일해 보인 두 우회를 여기에 **입력으로** 심어, 셈이 그 둘을 본다는 것을
// 못 박는다. 저장소에 위반 파일을 둘 수는 없으므로(그 파일이 워크 시험을 빨갛게
// 만든다) 원천을 문자열로 파싱한다.
func TestTheEncoderGuardCountsEveryWayToNameThePackage(t *testing.T) {
	call := "var _ = %sEncodeProductionFamilyActivation(%sFamilyActivationDocument{})\n"
	cases := map[string]struct {
		imports string
		prefix  string
	}{
		"평범한 import": {
			imports: "import \"" + familyActivationEncoderImportPath + "\"\n",
			prefix:  "strategyrouter.",
		},
		"별칭 import": {
			imports: "import sr \"" + familyActivationEncoderImportPath + "\"\n",
			prefix:  "sr.",
		},
		"점 import — 맨 이름으로 부른다": {
			imports: "import . \"" + familyActivationEncoderImportPath + "\"\n",
			prefix:  "",
		},
		"빈 import 를 진짜 별칭보다 먼저 적는다": {
			imports: "import (\n\t_ \"" + familyActivationEncoderImportPath + "\"\n\tsr \"" +
				familyActivationEncoderImportPath + "\"\n)\n",
			prefix: "sr.",
		},
	}
	for name, sample := range cases {
		t.Run(name, func(t *testing.T) {
			source := "package probe\n\n" + sample.imports + "\n" +
				strings.Replace(call, "%s", sample.prefix, -1)
			parsed, err := parser.ParseFile(token.NewFileSet(), "probe.go", source, 0)
			if err != nil {
				t.Fatalf("표본을 파싱하지 못했다: %v\n%s", err, source)
			}
			// 정의 패키지 **밖**의 경로로 센다 — 안이면 맨 이름 갈래가 경로만으로
			// 켜져서, 점 import 를 실제로 보는지 알 수 없게 된다.
			counts := countFamilyActivationAuthoringReferences(parsed,
				filepath.Join("internal", "app", "engine", "probe.go"))
			for _, symbol := range familyActivationAuthoringSymbols {
				if counts[symbol] == 0 {
					t.Errorf("가드가 %s 를 못 봤다 — 이 철자로 저작 표면을 부르는 "+
						"생산 파일이 조용히 통과한다\n%s", symbol, source)
				}
			}
		})
	}
}

// 부르지 않는 파일은 세지 않는다.
//
// 위 시험만 있으면 "무엇이든 센다" 는 판본이 통과한다. 그러면 가드가 저장소 전체를
// 빨갛게 만들어 쓸모가 없어지는데, 그 상태는 위 시험이 못 본다.
func TestTheEncoderGuardCountsNothingWhenThePackageIsNotCalled(t *testing.T) {
	source := "package probe\n\nimport \"strings\"\n\n" +
		"var _ = strings.TrimSpace(\"EncodeProductionFamilyActivation\")\n"
	parsed, err := parser.ParseFile(token.NewFileSet(), "probe.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	counts := countFamilyActivationAuthoringReferences(parsed,
		filepath.Join("internal", "app", "engine", "probe.go"))
	for _, symbol := range familyActivationAuthoringSymbols {
		if counts[symbol] != 0 {
			t.Errorf("가드가 부르지 않는 파일에서 %s 를 %d 번 셌다", symbol, counts[symbol])
		}
	}
}
