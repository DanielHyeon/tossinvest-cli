package strategyevidence

// a064 가 "주문 의도 0건, 브로커 요청 0건, 토글 변경 0건"이라고 주장한 성질을 실제로
// 재는 자리다.
//
// 원래 근거는 갓 만든 저널에서 intents/mutation_attempts/risk_reservations 의 행을 세는
// 것이었는데, 그 세 표는 실행 경로가 애초에 쓰지 않는다 — 코드를 어떻게 바꾸든 0 이라
// **실패할 수 없는** 단언이었다. 여기서는 두 가지를 대신 잰다.
//
//  1. 이 패키지의 import 폐포(transitive)에 브로커·주문·토글·HTTP 가 하나도 없다.
//     새 import 가 하나 생기면 빨개진다.
//  2. dormant 읽기 경로에서 **도달 가능한 함수 전부**가 SELECT 전용이다. 파일 하나·
//     메서드 하나만 보는 검사는 같은 패키지 헬퍼로 우회되므로, 호출 사슬을 따라간다.
//
// 두 시험 모두 양성 대조군을 갖는다. 계측기가 눈멀었는지 아닌지를 가르는 것은
// "아무것도 못 찾았다"가 아니라 "찾아야 할 것을 찾는다"이다.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const evidenceModulePath = "github.com/JungHoonGhae/tossinvest-cli"

// mutationPathFragments 는 주문·자금·운영 상태를 바꿀 수 있는 경로의 이름 조각이다.
var mutationPathFragments = []string{
	"net/http", "/broker", "/dispatch", "/execgw", "/guardian",
	"/operating", "/runtime", "/toggle", "/journal", "/order",
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// importClosure 는 저장소 안 패키지를 따라 import 를 전부 펼친다. 저장소 밖(표준
// 라이브러리·의존성)은 이름만 보고 멈춘다 — 우리가 막고 싶은 것은 "이 패키지가 무엇에
// 닿을 수 있는가"이고, 그 답은 우리 코드의 가장자리에서 결정된다.
func importClosure(t *testing.T, root, start string) map[string]bool {
	t.Helper()
	seen := map[string]bool{}
	var visit func(string)
	visit = func(pkg string) {
		if seen[pkg] {
			return
		}
		seen[pkg] = true
		if !strings.HasPrefix(pkg, evidenceModulePath+"/") {
			return
		}
		dir := filepath.Join(root, strings.TrimPrefix(pkg, evidenceModulePath+"/"))
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("reading package %s: %v", pkg, err)
		}
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			// 빌드 태그로 다른 OS 에만 들어가는 파일도 읽는다. 이 판정에서 넓게 보는
			// 것은 안전한 방향이다 — 놓치는 쪽이 아니라 더 잡는 쪽으로 틀린다.
			file, parseErr := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, name), nil, parser.ImportsOnly)
			if parseErr != nil {
				t.Fatalf("parsing %s: %v", filepath.Join(dir, name), parseErr)
			}
			for _, imported := range file.Imports {
				path, unquoteErr := strconv.Unquote(imported.Path.Value)
				if unquoteErr != nil {
					t.Fatal(unquoteErr)
				}
				visit(path)
			}
		}
	}
	visit(start)
	return seen
}

func mutationPathsIn(closure map[string]bool) []string {
	var found []string
	for pkg := range closure {
		for _, fragment := range mutationPathFragments {
			if strings.Contains(pkg, fragment) {
				found = append(found, pkg)
				break
			}
		}
	}
	return found
}

func TestStrategyEvidenceImportClosureReachesNoMutationPath(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	closure := importClosure(t, root, evidenceModulePath+"/internal/strategyevidence")
	if found := mutationPathsIn(closure); len(found) != 0 {
		t.Fatalf("strategy evidence can reach mutation/runtime packages: %v", found)
	}

	// 양성 대조군: 같은 계측기가 실제로 net/http 에 닿는 패키지를 잡는다. 이것이 없으면
	// "조각 목록이 아무것도 안 맞는다"와 "위반이 없다"를 구별할 수 없다.
	control := importClosure(t, root, evidenceModulePath+"/internal/obs")
	if found := mutationPathsIn(control); len(found) == 0 {
		t.Fatalf("import-closure detector found nothing in internal/obs; it cannot fail")
	}
}

func packageDeclarations(t *testing.T, dir string) map[string][]*ast.FuncDecl {
	t.Helper()
	declarations := map[string][]*ast.FuncDecl{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, parseErr := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, name), nil, 0)
		if parseErr != nil {
			t.Fatalf("parsing %s: %v", name, parseErr)
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			declarations[function.Name.Name] = append(declarations[function.Name.Name], function)
		}
	}
	if len(declarations) == 0 {
		t.Fatalf("no function declarations parsed from %s", dir)
	}
	return declarations
}

// reachableFrom 은 이름만으로 호출 사슬을 펼친다. 수신자 타입을 풀지 않으므로 같은
// 이름의 메서드를 전부 끌어온다 — 넓게 보는 쪽으로 틀리고, 그것이 안전한 방향이다.
func reachableFrom(declarations map[string][]*ast.FuncDecl, roots ...string) map[string]bool {
	reached := map[string]bool{}
	queue := append([]string(nil), roots...)
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if reached[name] {
			continue
		}
		reached[name] = true
		for _, function := range declarations[name] {
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				switch callee := call.Fun.(type) {
				case *ast.Ident:
					if !reached[callee.Name] {
						queue = append(queue, callee.Name)
					}
				case *ast.SelectorExpr:
					if !reached[callee.Sel.Name] {
						queue = append(queue, callee.Sel.Name)
					}
				}
				return true
			})
		}
	}
	return reached
}

var mutatingSQLKeywords = []string{
	"INSERT ", "UPDATE ", "DELETE ", "ALTER ", "DROP ", "CREATE ",
	"REPLACE ", "TRUNCATE ", "ATTACH ", "VACUUM ", "PRAGMA ",
}

var mutatingDatabaseCalls = []string{"Exec", "ExecContext", "Begin", "BeginTx", "Prepare", "PrepareContext"}

// normalizeSQL 은 공백을 하나로 접고 대문자로 만든다. 그래야 "DELETE\nFROM" 이나
// "CREATE  TABLE" 처럼 띄어쓰기만 다른 문장이 검사를 비켜 가지 못한다.
func normalizeSQL(literal string) string {
	return strings.Join(strings.Fields(strings.ToUpper(literal)), " ") + " "
}

func mutationsIn(t *testing.T, declarations map[string][]*ast.FuncDecl, reached map[string]bool) []string {
	t.Helper()
	var findings []string
	for name := range reached {
		for _, function := range declarations[name] {
			ast.Inspect(function.Body, func(node ast.Node) bool {
				switch value := node.(type) {
				case *ast.SelectorExpr:
					for _, forbidden := range mutatingDatabaseCalls {
						if value.Sel.Name == forbidden {
							findings = append(findings, name+" calls "+forbidden)
						}
					}
				case *ast.BasicLit:
					if value.Kind != token.STRING {
						break
					}
					literal, err := strconv.Unquote(value.Value)
					if err != nil {
						break
					}
					normalized := normalizeSQL(literal)
					for _, forbidden := range mutatingSQLKeywords {
						if strings.Contains(normalized, forbidden) {
							findings = append(findings, name+" contains SQL "+strings.TrimSpace(forbidden))
						}
					}
				}
				return true
			})
		}
	}
	return findings
}

func TestDormantSnapshotReadIsSelectOnlyAcrossTheWholeCallChain(t *testing.T) {
	t.Parallel()
	declarations := packageDeclarations(t, ".")
	for _, root := range []string{"Replay", "NewDormantSnapshotReadPort"} {
		if len(declarations[root]) == 0 {
			t.Fatalf("dormant read entry point %s is not declared in this package", root)
		}
	}
	reached := reachableFrom(declarations, "Replay", "NewDormantSnapshotReadPort")
	if !reached["snapshotDigest"] || !reached["scanEnvelope"] || !reached["snapshotItemMatchesQuery"] {
		t.Fatalf("call-chain walker did not reach the dormant read helpers: %d names", len(reached))
	}
	if findings := mutationsIn(t, declarations, reached); len(findings) != 0 {
		t.Fatalf("dormant snapshot read chain can mutate the database: %v", findings)
	}

	// 양성 대조군: 같은 계측기를 봉인 경로에 대면 반드시 쓰기를 찾아야 한다.
	sealing := reachableFrom(declarations, "SealSnapshot")
	if findings := mutationsIn(t, declarations, sealing); len(findings) == 0 {
		t.Fatal("SELECT-only detector found no write in SealSnapshot; it cannot fail")
	}
}
