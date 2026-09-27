package execgw_test

// a066 5.5 잔여 격 판정 — q_final 표식 없는(legacy) 진입 결정은 horizon 이 없어 진입 손실 잠금 밖에 있음.
// 그 잔여가 생산에서 실제로 도달 가능한지 소스 전수로 셈(Manager 지시 2026-09-27: "전칭 판정은 표본이 0이면
// 통과한다 — 표본이 언제 채워지는지까지").
//
// 측정(2026-09-27): legacy 진입 발급자 RiskGuardian.IssueEntry 로 가는 길은 Tracer.submitEntry 와
// RiskGuardian.IssueStrategyEntry 둘뿐이고, 그 위의 NewTracer 호출과 strategydispatch.GuardianAdapter
// 생성은 비시험 코드에 0 개임 → 생산 조립에서 q_final 없이 도달 가능한 진입 경로 0.
//
// 이 시험은 그 0 을 **표본이 채워지는 순간** 깨지게 고정함: 시험은 매 실행 시 저장소의 비시험 소스 전부를
// 다시 파싱하므로, 누가 생산 코드에 Tracer·GuardianAdapter 를 조립하거나 새 EXPOSURE_RAISING 결정 생산자를
// 더하면 그 커밋의 스위트가 빨개짐. 그때는 그 경로가 진입 손실 잠금을 거치는지 먼저 판정해야 함.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// enclosingSites 는 비시험 Go 소스에서 match 가 참인 노드를 "파일:수신자.함수" 좌표로 모음.
// 좌표는 줄 번호가 아니라 감싼 선언이라 남의 편집에 흔들리지 않음.
func enclosingSites(t *testing.T, match func(ast.Node) bool) []string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, top := range []string{"internal", "cmd"} {
		err := filepath.WalkDir(filepath.Join(root, top), func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(root, path)
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				name := fn.Name.Name
				if fn.Recv != nil && len(fn.Recv.List) == 1 {
					recv := fn.Recv.List[0].Type
					if star, ok := recv.(*ast.StarExpr); ok {
						recv = star.X
					}
					if ident, ok := recv.(*ast.Ident); ok {
						name = ident.Name + "." + name
					}
				}
				ast.Inspect(fn.Body, func(node ast.Node) bool {
					if node != nil && match(node) {
						seen[filepath.ToSlash(rel)+":"+name] = true
					}
					return true
				})
			}
			// 함수 밖(패키지 수준 var 초기화 등)의 조립도 세어야 함.
			for _, decl := range file.Decls {
				if gen, ok := decl.(*ast.GenDecl); ok {
					ast.Inspect(gen, func(node ast.Node) bool {
						if node != nil && match(node) {
							seen[filepath.ToSlash(rel)+":<package-level>"] = true
						}
						return true
					})
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	sites := make([]string, 0, len(seen))
	for site := range seen {
		sites = append(sites, site)
	}
	sort.Strings(sites)
	return sites
}

func calls(name string) func(ast.Node) bool {
	return func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return false
		}
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			return fun.Name == name
		case *ast.SelectorExpr:
			return fun.Sel.Name == name
		}
		return false
	}
}

func literalOf(typeName string) func(ast.Node) bool {
	return func(node ast.Node) bool {
		lit, ok := node.(*ast.CompositeLit)
		if !ok {
			return false
		}
		switch typ := lit.Type.(type) {
		case *ast.Ident:
			return typ.Name == typeName
		case *ast.SelectorExpr:
			return typ.Sel.Name == typeName
		}
		return false
	}
}

func exposureRaisingProducer(node ast.Node) bool {
	kv, ok := node.(*ast.KeyValueExpr)
	if !ok {
		return false
	}
	key, ok := kv.Key.(*ast.Ident)
	if !ok || key.Name != "SafetyClass" {
		return false
	}
	switch value := kv.Value.(type) {
	case *ast.Ident:
		return value.Name == "SafetyClassExposureRaising"
	case *ast.SelectorExpr:
		return value.Sel.Name == "SafetyClassExposureRaising"
	}
	return false
}

func TestA066LegacyEntryPathsAreUnreachableFromProductionAssembly(t *testing.T) {
	cases := []struct {
		name  string
		match func(ast.Node) bool
		want  []string
	}{
		// legacy 발급자 RiskGuardian.IssueEntry 를 부르는 자리(같은 이름의 Issuer.IssueEntry 호출자 포함).
		{"IssueEntry callers", calls("IssueEntry"), []string{
			"internal/app/engine/tracer.go:Tracer.submitEntry",
			"internal/execgw/riskguardian.go:RiskGuardian.IssueStrategyEntry",
		}},
		{"IssueStrategyEntry callers", calls("IssueStrategyEntry"), []string{
			"internal/strategydispatch/adapters.go:GuardianAdapter.IssueAndPlan",
		}},
		// 두 legacy 경로의 조립 지점 — 0 이 잔여의 격을 "생산 도달 0"으로 정함.
		{"NewTracer callers", calls("NewTracer"), nil},
		{"GuardianAdapter construction", literalOf("GuardianAdapter"), nil},
		// EXPOSURE_RAISING 결정을 짓는 자리 전부. q_final 둘(IssuePrechecked*)과 legacy 둘.
		{"exposure-raising decision producers", exposureRaisingProducer, []string{
			"internal/execgw/issue.go:Issuer.IssueEntry",
			"internal/execgw/riskguardian.go:RiskGuardian.IssueEntry",
			"internal/execgw/riskguardian_first_leg.go:RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg",
			"internal/execgw/riskguardian_qfinal.go:RiskGuardian.IssuePrecheckedQFinalEntry",
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := enclosingSites(t, tc.match)
			if strings.Join(got, "\n") != strings.Join(tc.want, "\n") {
				t.Fatalf("census changed — decide whether the new path passes the a066 entry loss lock before updating it.\n got: %v\nwant: %v", got, tc.want)
			}
		})
	}
}
