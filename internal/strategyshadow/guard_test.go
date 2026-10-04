package strategyshadow

// a112 7.3.1 SHADOW 경계 가드(브리프 v3.3 §1 · §2).
//
//   - 저작: shadow 매니페스트 바이트를 만드는 표면(Encode · Document)은 정의와 도구 `tools/a112-family-shadow` 만 쓴다(생산 작성자 0).
//   - 주조: strategyrouter.FamilyActivation 의 비영 composite literal 은 저장소 생산 코드 전체에서 정확히 1 자리(활성화 적재기 반환) —
//     shadow 경로가 활성화를 만들 수 없다는 것을 타입 가시성(필드 비공개) 위에 셈으로 한 번 더 못 박는다. `*_testseam.go` 는 이름으로 뺀다.
//   - 폐포: strategyshadow 를 뿌리로 한 **생산** 그래프 둘(deps · deps-tagged)의 모듈 안 패키지가 `unsafe` · `reflect` 를 직접 import 하지
//     않는다(비공개 필드 우회 차단). 표준 라이브러리 내부 사용(encoding/json 등)은 범위 밖 — codex 재검 3 N2. 시험 그래프는 범위 밖(재검 4).
//   - 동결: 이 패키지 생산 소스의 gofmt 정본 digest(strategyhandoff source_freeze_test.go 선례) — 바꾸려면 같은 커밋에서 상수를 다시 적는다.

import (
	"crypto/sha256"
	"encoding/hex"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

const (
	shadowImportPath = "github.com/JungHoonGhae/tossinvest-cli/internal/strategyshadow"
	routerImportPath = "github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	modulePrefix     = "github.com/JungHoonGhae/tossinvest-cli/"
)

// frozenShadowSourceDigest 는 이 패키지 생산 소스의 gofmt 정본 digest 다. 재고정 절차는 strategyhandoff/source_freeze_test.go 와 같다.
const frozenShadowSourceDigest = "sha256:7faa518830d790574ee727e43faa5805511fe33f576337ce501aa7e98db49b41"

func repositoryGoFiles(t *testing.T, visit func(relative string, file *ast.File)) {
	t.Helper()
	root := filepath.Join("..", "..")
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "vendor", "node_modules", ".sdd":
				return fs.SkipDir
			}
			if strings.HasPrefix(entry.Name(), "_") || entry.Name() == "testdata" {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		parsed, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		visit(relative, parsed)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// importNames 는 그 파일에서 importPath 를 가리키는 이름들이다(별칭 포함, 점 import 면 dot=true).
func importNames(file *ast.File, importPath string) (map[string]bool, bool) {
	names, dot := map[string]bool{}, false
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil || path != importPath {
			continue
		}
		switch {
		case spec.Name == nil:
			names[path[strings.LastIndex(path, "/")+1:]] = true
		case spec.Name.Name == ".":
			dot = true
		case spec.Name.Name != "_":
			names[spec.Name.Name] = true
		}
	}
	return names, dot
}

func TestOnlyTheAuthoringToolCanBuildShadowBytes(t *testing.T) {
	symbols := []string{"EncodeProductionFamilyShadow", "Document"}
	allowed := map[string]bool{
		filepath.Join("internal", "strategyshadow", "shadow.go"): true,
		filepath.Join("tools", "a112-family-shadow", "main.go"):  true,
	}
	seen := map[string]int{}
	repositoryGoFiles(t, func(relative string, file *ast.File) {
		bare := filepath.Dir(relative) == filepath.Join("internal", "strategyshadow")
		aliases, dot := importNames(file, shadowImportPath)
		bare = bare || dot
		ast.Inspect(file, func(node ast.Node) bool {
			name := ""
			switch value := node.(type) {
			case *ast.SelectorExpr:
				if ident, ok := value.X.(*ast.Ident); ok && aliases[ident.Name] {
					name = value.Sel.Name
				}
			case *ast.Ident:
				if bare {
					name = value.Name
				}
			}
			for _, symbol := range symbols {
				if name == symbol {
					seen[relative+"\x00"+symbol]++
					if !allowed[relative] {
						t.Errorf("%s builds shadow manifest bytes (%s) — only tools/a112-family-shadow may", relative, symbol)
					}
				}
			}
			return true
		})
	})
	for path := range allowed {
		for _, symbol := range symbols {
			if seen[path+"\x00"+symbol] == 0 {
				t.Errorf("%s has no %s reference — the guard counts a name that no longer exists", path, symbol)
			}
		}
	}
}

func TestExactlyOneProductionSiteMintsANonZeroFamilyActivation(t *testing.T) {
	var sites []string
	repositoryGoFiles(t, func(relative string, file *ast.File) {
		if strings.HasSuffix(relative, "_testseam.go") {
			return
		}
		inRouter := filepath.Dir(relative) == filepath.Join("internal", "strategyrouter")
		aliases, dot := importNames(file, routerImportPath)
		ast.Inspect(file, func(node ast.Node) bool {
			literal, ok := node.(*ast.CompositeLit)
			if !ok || len(literal.Elts) == 0 {
				return true
			}
			switch typ := literal.Type.(type) {
			case *ast.Ident:
				if typ.Name == "FamilyActivation" && (inRouter || dot) {
					sites = append(sites, relative)
				}
			case *ast.SelectorExpr:
				if ident, ok := typ.X.(*ast.Ident); ok && aliases[ident.Name] && typ.Sel.Name == "FamilyActivation" {
					sites = append(sites, relative)
				}
			}
			return true
		})
	})
	want := []string{filepath.Join("internal", "strategyrouter", "production_family_activation.go")}
	if strings.Join(sites, ",") != strings.Join(want, ",") {
		t.Fatalf("non-zero FamilyActivation literals at %v, want exactly %v (the activation loader's return)", sites, want)
	}
}

func TestNoModulePackageInTheShadowProductionClosureImportsUnsafeOrReflect(t *testing.T) {
	for _, mode := range testenv.WalkModes() {
		if mode.Tests {
			continue
		}
		graph := testenv.ListDeps(t, ".", mode)
		reached := graph.Reachable(graph.Roots(shadowImportPath), nil)
		modules := 0
		for name := range reached {
			if !strings.HasPrefix(name, modulePrefix) {
				continue
			}
			modules++
			for _, imported := range graph[name] {
				if imported == "unsafe" || imported == "reflect" {
					t.Errorf("%s: module package %s imports %s directly", mode.Name, name, imported)
				}
			}
		}
		// 하한(양성 대조): 걸음이 strategyshadow · strategyrouter 를 실제로 지났다.
		if modules < 2 || !reached[shadowImportPath] || !reached[routerImportPath] {
			t.Fatalf("%s: the walk reached %d module packages (shadow=%v router=%v) — the census read nothing",
				mode.Name, modules, reached[shadowImportPath], reached[routerImportPath])
		}
	}
}

func TestTheShadowPackageSourceIsFrozen(t *testing.T) {
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	hash := sha256.New()
	files := 0
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		canonical, err := format.Source(raw)
		if err != nil {
			t.Fatalf("gofmt %s: %v", path, err)
		}
		hash.Write([]byte(path + "\x00"))
		hash.Write(canonical)
		files++
	}
	if files == 0 {
		t.Fatal("no production source was hashed")
	}
	if got := "sha256:" + hex.EncodeToString(hash.Sum(nil)); got != frozenShadowSourceDigest {
		t.Fatalf("strategyshadow production source digest %s, frozen %s — re-pin in the same commit with the reason (source freeze procedure)",
			got, frozenShadowSourceDigest)
	}
}
