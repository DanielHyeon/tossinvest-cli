package strategyflow

// 결정 (5) 의 봉인 로트 몫 — strategyflow 의 **봉인된 Result 를 만드는 문** census(a112 6.2 봉인 로트, Manager 판정 2026-10-01).
//
// 봉인(`proposalSeal`)은 비밀이 아니라 결과 값의 해시다(`proposalResultSeal`). 그러니 「봉인된 Result 를 만든다」는 곧 「proposalSeal
// 에 그 해시를 쓴다」이고, 그 쓰기는 이 패키지 안에서만 가능하다(필드가 비공개). 이 파일은 그 쓰기와 그 문을 세 겹으로 센다:
//   ① 공개 표면 동결 — 기본 빌드(태그 없음)의 공개 이름 · 서명 전부를 testdata/exported_surface.golden 과 대조(조용한 새 문 금지).
//   ② 봉인 쓰기 census — 기본 빌드에서 proposalSeal 에 0 이 아닌 값을 쓰는 자리는 `sealProposalResult` 하나, 그것을 부르는 자리는
//      `Propose` 하나. 태그(`tossos_testseams`) 파일의 시험 전용 주조기 둘은 태그 파일 안에서만 부른다. 구조체 필드 `proposalSeal`
//      을 선언하는 타입은 Result 하나뿐 — 같은 필드 이름을 가진 쌍둥이 구조체(변환으로 봉인을 들여오는 모양)나 그 필드를 품은
//      제약(제네릭 주조)을 막는다.
//   ②' 언급 census — 봉인 관련 세 이름의 **모든 등장**을 함수별로 세어 표와 대조(함수 값 별칭 · 주소 경유 쓰기 차단, 6.2 리뷰 codex #2),
//      봉인 필드의 주소 취득 · 슬라이싱 금지, reflect · unsafe import 금지.
//   ③ 함수 수준 AST 정본 digest 동결 — sealProposalResult · proposalResultSeal · ValidProposal 의 본문(주석 제외 · 좌표 독립)을
//      상수로 고정하고, 그 상수가 **어느 리뷰 기록(openspec review.md)에 적혀 있어야** 통과한다(재고정-리뷰 결속 — 5.2.2.1 리뷰
//      이월 #3). 패키지 전체 동결은 하지 않는다 — ~2,900줄이고 형제 로트가 편집한다.
//
// **③ 의 한계(6.2 리뷰 보이스 B #8):** review 결속은 openspec review 기록에 digest 문자열이 **적혀 있는지**만 본다 — 저자가 같은 커밋에서
// 스스로 만족시킬 수 있고, 증명하는 것은 「적혔다」이지 「독립 리뷰됐다」가 아니다. 동결 셋 밖의 해시 입력 함수(writeLineageString ·
// writeLineageUint64 · executionTermsIdentity)는 동결하지 않는다. 같은 이름의 함수가 여러 파일에 있으면 마지막 정의만 본다(플랫폼별
// 파일은 위에서 거절하므로 오늘은 생기지 않는다).
//
// **못 보는 모양(이름 붙여 둔다).** ② 는 모양을 센다: 필드 이름을 쓰지 않는 위치 기반 합성 리터럴로 Result 를 통째로 만드는 편집
// (`Result{a, b, …, seal}`)은 ① 의 표면 · ③ 의 동결에 걸리지 않을 수 있다 — 위치 기반 Result 리터럴 자체를 아래에서 금지하는 것으로
// 대신한다. 이 census 는 완전성을 주장하지 않는다. 주문 경로에서의 종결은 1차 레그 권한의 소유자 범위 재유도(엔진)가 진다 —
// 봉인이 위조돼도 조립이 중재하지 않은 제안은 거기서 거절된다.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type flowFile struct {
	name   string
	file   *ast.File
	tagged bool
}

func strategyflowSourceFiles(t *testing.T) []flowFile {
	t.Helper()
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	var out []flowFile
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		// 「tagged」는 시험 호스트의 빌드 맥락(build.Default — cgo · GOOS)으로 정하지 않는다(6.2 리뷰 보이스 B #1): 생산 이미지는
		// CGO_ENABLED=0 이고 release 는 여러 GOOS 로 빌드하므로, `!cgo` · `_windows.go` 같은 파일은 호스트에서는 빠져도 생산에는 들어간다.
		// 태그 파일은 정확히 `//go:build tossos_testseams` 인 파일뿐이고, 그 밖의 빌드 제약 · GOOS/GOARCH 접미사 파일은 거절한다(모르면 실패).
		constraint := ""
		for _, group := range file.Comments {
			for _, comment := range group.List {
				if strings.HasPrefix(comment.Text, "//go:build ") && comment.Pos() < file.Package {
					constraint = strings.TrimSpace(strings.TrimPrefix(comment.Text, "//go:build "))
				}
			}
		}
		if constraint != "" && constraint != "tossos_testseams" {
			t.Fatalf("%s carries the build constraint %q — only tossos_testseams is recognised; a constraint the census does not model "+
				"could ship a file the census skipped", path, constraint)
		}
		if goosArchSuffix(path) {
			t.Fatalf("%s has a GOOS/GOARCH file-name suffix — the census does not model per-platform files", path)
		}
		out = append(out, flowFile{name: path, file: file, tagged: constraint == "tossos_testseams"})
	}
	if len(out) == 0 {
		t.Fatal("no strategyflow source file")
	}
	return out
}

// goosArchSuffix 는 파일 이름이 go 의 암묵 빌드 제약(`_linux.go` · `_amd64.go` · `_windows_arm64.go` …)을 지는지 답한다.
func goosArchSuffix(path string) bool {
	name := strings.TrimSuffix(filepath.Base(path), ".go")
	parts := strings.Split(name, "_")
	known := map[string]bool{}
	for _, value := range strings.Fields("aix android darwin dragonfly freebsd hurd illumos ios js linux nacl netbsd openbsd plan9 solaris " +
		"wasip1 windows zos 386 amd64 amd64p32 arm arm64 arm64be armbe loong64 mips mipsle mips64 mips64le mips64p32 mips64p32le ppc ppc64 " +
		"ppc64le riscv riscv64 s390 s390x sparc sparc64 wasm") {
		known[value] = true
	}
	return len(parts) > 1 && known[parts[len(parts)-1]]
}

func funcName(function *ast.FuncDecl) string {
	if function.Recv != nil && len(function.Recv.List) == 1 {
		receiver := function.Recv.List[0].Type
		if star, ok := receiver.(*ast.StarExpr); ok {
			receiver = star.X
		}
		return types.ExprString(receiver) + "." + function.Name.Name
	}
	return function.Name.Name
}

// ① 공개 표면 동결.
func TestTheStrategyflowSurfaceIsFrozen(t *testing.T) {
	var lines []string
	for _, source := range strategyflowSourceFiles(t) {
		if source.tagged {
			continue
		}
		for _, decl := range source.file.Decls {
			switch value := decl.(type) {
			case *ast.FuncDecl:
				if !value.Name.IsExported() {
					continue
				}
				if value.Recv != nil {
					receiver := value.Recv.List[0].Type
					if star, ok := receiver.(*ast.StarExpr); ok {
						receiver = star.X
					}
					if ident, ok := receiver.(*ast.Ident); !ok || !ident.IsExported() {
						continue
					}
				}
				lines = append(lines, "func "+funcName(value)+" "+types.ExprString(value.Type))
			case *ast.GenDecl:
				for _, spec := range value.Specs {
					switch inner := spec.(type) {
					case *ast.TypeSpec:
						if inner.Name.IsExported() {
							lines = append(lines, "type "+inner.Name.Name+" "+types.ExprString(inner.Type))
						}
					case *ast.ValueSpec:
						for _, name := range inner.Names {
							if name.IsExported() {
								lines = append(lines, value.Tok.String()+" "+name.Name)
							}
						}
					}
				}
			}
		}
	}
	sort.Strings(lines)
	got := strings.Join(lines, "\n") + "\n"
	// 재생성은 의도적 재고정이다 — 파일 diff 가 리뷰에 보인다.
	if os.Getenv("STRATEGYFLOW_REGENERATE_SURFACE") == "1" {
		if err := os.WriteFile(filepath.Join("testdata", "exported_surface.golden"), []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		// 같은 실행에서 통과시키지 않는다(6.2 리뷰 보이스 B #7) — 재생성은 멈추고, 다시 돌려 diff 를 보게 한다.
		t.Fatal("regenerated testdata/exported_surface.golden — rerun without STRATEGYFLOW_REGENERATE_SURFACE and review the diff")
	}
	want, err := os.ReadFile(filepath.Join("testdata", "exported_surface.golden"))
	if err != nil {
		t.Fatalf("read the frozen surface: %v", err)
	}
	if got != string(want) {
		t.Fatalf("strategyflow's exported surface (default build) changed. A new public name or signature is a new door for sealed "+
			"Results; declare it by regenerating testdata/exported_surface.golden in the same commit, and name it in review.\n--- got ---\n%s", got)
	}
}

// ② 봉인 쓰기 census.
func TestOnlyProposeSealsAProposalInTheProductionBuild(t *testing.T) {
	sealWriters := map[string]bool{}   // proposalSeal 에 0 이 아닌 값을 쓰는 함수
	sealCallers := map[string]string{} // sealProposalResult 를 부르는 함수 → tagged/untagged
	fieldDecls := 0
	var problems []string
	for _, source := range strategyflowSourceFiles(t) {
		kind := "untagged"
		if source.tagged {
			kind = "tagged"
		}
		for _, decl := range source.file.Decls {
			owner := "(package level)"
			if function, ok := decl.(*ast.FuncDecl); ok {
				owner = funcName(function)
			}
			ast.Inspect(decl, func(node ast.Node) bool {
				switch value := node.(type) {
				case *ast.StructType:
					for _, field := range value.Fields.List {
						for _, name := range field.Names {
							if name.Name == "proposalSeal" {
								fieldDecls++
							}
						}
					}
				case *ast.AssignStmt:
					for index, left := range value.Lhs {
						selector, ok := left.(*ast.SelectorExpr)
						if !ok || selector.Sel.Name != "proposalSeal" {
							continue
						}
						if index < len(value.Rhs) && types.ExprString(value.Rhs[index]) == "[32]byte{}" {
							continue // 봉인을 지우는 쓰기(안전 방향)는 어디서나 허용
						}
						sealWriters[source.name+":"+owner] = true
					}
				case *ast.KeyValueExpr:
					if key, ok := value.Key.(*ast.Ident); ok && key.Name == "proposalSeal" {
						problems = append(problems, source.name+":"+owner+" writes proposalSeal in a composite literal")
					}
				case *ast.CompositeLit:
					if types.ExprString(value.Type) == "Result" && len(value.Elts) > 0 {
						if _, keyed := value.Elts[0].(*ast.KeyValueExpr); !keyed {
							problems = append(problems, source.name+":"+owner+" builds a Result positionally")
						}
					}
				case *ast.CallExpr:
					if ident, ok := value.Fun.(*ast.Ident); ok && ident.Name == "sealProposalResult" {
						sealCallers[source.name+":"+owner] = kind
					}
				}
				return true
			})
		}
	}
	if fieldDecls != 1 {
		problems = append(problems, "struct fields named proposalSeal="+strconv.Itoa(fieldDecls)+", want exactly 1 (Result) — a twin struct or constraint can smuggle a seal in by conversion")
	}
	writers := sortedKeys(sealWriters)
	if strings.Join(writers, ",") != "types.go:sealProposalResult" {
		problems = append(problems, "non-zero proposalSeal writers="+strings.Join(writers, ",")+", want types.go:sealProposalResult")
	}
	wantCallers := map[string]string{
		"flow.go:Propose": "untagged",
		"authority_testseam.go:AcceptedResultForAuthorityTest":                          "tagged",
		"authority_stop_provenance_testseam.go:ResultWithRestatedStopProvenanceForTest": "tagged",
	}
	for site, kind := range sealCallers {
		if wantCallers[site] != kind {
			problems = append(problems, "sealProposalResult is called from "+site+" ("+kind+") which is not a declared minting door")
		}
	}
	for site := range wantCallers {
		if _, ok := sealCallers[site]; !ok {
			problems = append(problems, "declared minting door "+site+" no longer calls sealProposalResult — update the census deliberately")
		}
	}
	if len(problems) != 0 {
		sort.Strings(problems)
		t.Fatalf("%v", problems)
	}
}

// ②' 언급 census(6.2 리뷰 codex #2): 봉인 쓰기 census 는 **호출식과 대입 좌변**만 봤으므로 함수 값 별칭(`mint := sealProposalResult;
// mint(r)`)과 주소 경유 쓰기(`p := &r.proposalSeal; *p = …`)를 못 봤다. 여기서는 모양을 가리지 않고 **언급을 센다**: 봉인과 관련된
// 세 이름(`sealProposalResult` · `proposalResultSeal` 식별자, `.proposalSeal` 선택자)의 모든 등장을 감싼 함수별로 세어 아래 표와 대조한다 —
// 새 등장은 그 모양이 무엇이든 표와 어긋난다. 그리고 봉인 필드의 주소 취득 · 슬라이싱, reflect · unsafe import(비공개 필드 우회)를 금지한다.
var sealMentionCensus = map[string]int{
	"flow.go:Propose:sealProposalResult":                                                               1,
	"types.go:Result.ValidProposal:.proposalSeal":                                                      2,
	"types.go:Result.ValidProposal:proposalResultSeal":                                                 1,
	"types.go:sealProposalResult:.proposalSeal":                                                        2,
	"types.go:sealProposalResult:proposalResultSeal":                                                   1,
	"types.go:FinalizeProposalQuantity:.proposalSeal":                                                  1,
	"authority_testseam.go:AcceptedResultForAuthorityTest:sealProposalResult":                          1,
	"authority_stop_provenance_testseam.go:ResultWithRestatedStopProvenanceForTest:sealProposalResult": 1,
}

func TestEveryMentionOfTheSealIsWhereTheCensusSaysItIs(t *testing.T) {
	found := map[string]int{}
	var problems []string
	for _, source := range strategyflowSourceFiles(t) {
		for _, spec := range source.file.Imports {
			if path := strings.Trim(spec.Path.Value, `"`); path == "reflect" || path == "unsafe" {
				problems = append(problems, source.name+" imports "+path+" — it can reach the unexported seal field")
			}
		}
		for _, decl := range source.file.Decls {
			owner := "(package level)"
			var declName *ast.Ident
			if function, ok := decl.(*ast.FuncDecl); ok {
				owner, declName = funcName(function), function.Name
			}
			ast.Inspect(decl, func(node ast.Node) bool {
				switch value := node.(type) {
				case *ast.Ident:
					if value != declName && (value.Name == "sealProposalResult" || value.Name == "proposalResultSeal") {
						found[source.name+":"+owner+":"+value.Name]++
					}
				case *ast.SelectorExpr:
					if value.Sel.Name == "proposalSeal" {
						found[source.name+":"+owner+":.proposalSeal"]++
					}
				case *ast.UnaryExpr:
					if selector, ok := value.X.(*ast.SelectorExpr); ok && value.Op == token.AND && selector.Sel.Name == "proposalSeal" {
						problems = append(problems, source.name+":"+owner+" takes the address of proposalSeal")
					}
				case *ast.SliceExpr:
					if selector, ok := value.X.(*ast.SelectorExpr); ok && selector.Sel.Name == "proposalSeal" {
						problems = append(problems, source.name+":"+owner+" slices proposalSeal")
					}
				}
				return true
			})
		}
	}
	// 양성 대조: 알려진 쓰기 자리를 실제로 셌는가.
	if found["types.go:sealProposalResult:.proposalSeal"] == 0 {
		t.Fatal("the mention census did not see sealProposalResult's own seal write — it is blind")
	}
	for key, count := range found {
		if sealMentionCensus[key] != count {
			problems = append(problems, key+" appears "+strconv.Itoa(count)+" time(s), census says "+strconv.Itoa(sealMentionCensus[key]))
		}
	}
	for key, count := range sealMentionCensus {
		if found[key] != count {
			problems = append(problems, key+" is in the census "+strconv.Itoa(count)+" time(s) but the source has "+strconv.Itoa(found[key]))
		}
	}
	if len(problems) != 0 {
		sort.Strings(problems)
		t.Fatalf("a mention of the seal moved or appeared: %v", problems)
	}
}

// ③ 함수 수준 AST 정본 digest 동결 + 재고정-리뷰 결속.
var frozenSealFunctions = map[string]string{
	"sealProposalResult":   "sha256:5e45db69120f540a463f43b2c6d8ba8cd0e7f94d3b070fcb9a80b42629d6bd6a",
	"proposalResultSeal":   "sha256:311b1ccc0808e9ac01c35fa5c4b169ed255b12eac3445842b698deaddfcfb69e",
	"Result.ValidProposal": "sha256:60114abda6d2266c6fb27c269c74f6089cdb637f2524cd739c032d7a3b74a4ee",
}

func canonicalFuncDigest(t *testing.T, function *ast.FuncDecl) string {
	t.Helper()
	clone := *function
	clone.Doc = nil
	var buffer bytes.Buffer
	// 주석 없이(파일 주석 맵을 넘기지 않음) 인쇄 — 좌표 · 주석 · 들여쓰기와 무관한 정본.
	if err := printer.Fprint(&buffer, token.NewFileSet(), &clone); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(buffer.Bytes())
	return "sha256:" + hex.EncodeToString(sum[:])
}

func TestTheSealFunctionsAreFrozenAndTheirPinsAreReviewed(t *testing.T) {
	found := map[string]string{}
	for _, source := range strategyflowSourceFiles(t) {
		for _, decl := range source.file.Decls {
			if function, ok := decl.(*ast.FuncDecl); ok {
				if _, frozen := frozenSealFunctions[funcName(function)]; frozen {
					found[funcName(function)] = canonicalFuncDigest(t, function)
				}
			}
		}
	}
	reviews := reviewRecords(t)
	for name, want := range frozenSealFunctions {
		got, ok := found[name]
		if !ok {
			t.Errorf("frozen seal function %s is gone", name)
			continue
		}
		if got != want {
			t.Errorf("seal function %s changed: digest %s, frozen %s. Re-pin in the SAME commit, write the new digest into the "+
				"change's review.md with the reason, and land only after an independent review has read the diff.", name, got, want)
			continue
		}
		if !strings.Contains(reviews, want) {
			t.Errorf("the frozen digest of %s (%s) is not written in any openspec review record — a re-pin must be reviewed", name, want)
		}
	}
}

// reviewRecords 는 openspec 의 모든 review.md(진행 중 · archive)를 이어 붙인다 — archive 로 옮겨도 결속이 끊기지 않게.
func reviewRecords(t *testing.T) string {
	t.Helper()
	var all strings.Builder
	for _, pattern := range []string{"../../openspec/changes/*/review.md", "../../openspec/changes/archive/*/review.md"} {
		paths, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			all.Write(raw)
		}
	}
	if all.Len() == 0 {
		t.Fatal("no openspec review record found — the re-pin binding cannot be checked")
	}
	return all.String()
}

func sortedKeys(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for key := range values {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}
