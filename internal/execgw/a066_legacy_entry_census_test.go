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
// censusRoots 는 걷는 최상위 디렉터리이고 censusRootFiles 는 마지막 걷기에서 각 root 가 준 비시험 Go 파일 수임.
// 걷기 범위가 줄면(root 가 빠지거나 비면) census 는 공허하게 통과하므로 그 수를 따로 단언함.
var (
	censusRoots     = []string{"internal", "cmd", "tools"}
	censusRootFiles = map[string]int{}
)

func enclosingSites(t *testing.T, match func(ast.Node) bool) []string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	// tools/ 도 같은 모듈의 main 패키지들이라 셈(적대 리뷰 2026-09-27).
	for _, top := range censusRoots {
		censusRootFiles[top] = 0
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
			censusRootFiles[top]++
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

// references 는 이름 하나를 가리키는 모든 식별자(호출·메서드 값·타입 참조·new(T)·var x T·&T{})를 셈.
// 적대 리뷰(2026-09-27)가 calls·literalOf 만으로는 메서드 값(f := g.IssueEntry)·new(GuardianAdapter)·
// var a GuardianAdapter·&Tracer{} 가 빠진다고 지적함 — 이 판정은 그 모양 전부를 잡음.
func references(name string) func(ast.Node) bool {
	return func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.Ident:
			return n.Name == name
		case *ast.SelectorExpr:
			return n.Sel.Name == name
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

// exposureRaisingProducer 는 결정의 SafetyClass 를 **정하는** 자리를 셈 — 복합 리터럴 필드(SafetyClass: v)와
// 대입(x.SafetyClass = v) 둘 다, 값이 RISK_REDUCING 상수가 아니면 전부. 값이 상수가 아닌 변수·문자열
// 리터럴("EXPOSURE_RAISING")이어도 셈(적대 리뷰 2026-09-27: 대입·문자열 모양이 빠졌었음).
func exposureRaisingProducer(node ast.Node) bool {
	var key, value ast.Expr
	switch n := node.(type) {
	case *ast.KeyValueExpr:
		key, value = n.Key, n.Value
	case *ast.AssignStmt:
		if len(n.Lhs) != 1 || len(n.Rhs) != 1 {
			return false
		}
		selector, ok := n.Lhs[0].(*ast.SelectorExpr)
		if !ok {
			return false
		}
		key, value = selector.Sel, n.Rhs[0]
	default:
		return false
	}
	if ident, ok := key.(*ast.Ident); !ok || ident.Name != "SafetyClass" {
		return false
	}
	switch v := value.(type) {
	case *ast.Ident:
		return v.Name != "SafetyClassRiskReducing"
	case *ast.SelectorExpr:
		return v.Sel.Name != "SafetyClassRiskReducing"
	}
	return true
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
		// 모양을 가리지 않는 참조 전수 — 자기 선언 자리만 남아야 함.
		// tracer.go 패키지 수준 = EntryIssuer 인터페이스의 메서드 선언.
		{"any reference to IssueEntry", references("IssueEntry"), []string{
			"internal/app/engine/tracer.go:<package-level>",
			"internal/app/engine/tracer.go:Tracer.submitEntry",
			"internal/execgw/riskguardian.go:RiskGuardian.IssueStrategyEntry",
		}},
		{"any reference to IssueStrategyEntry", references("IssueStrategyEntry"), []string{
			"internal/strategydispatch/adapters.go:GuardianAdapter.IssueAndPlan",
		}},
		{"any reference to NewTracer", references("NewTracer"), nil},
		// 타입 선언 자신과 생성자 NewTracer 안의 &Tracer{} 뿐 — NewTracer 호출이 0 이므로 생성 0.
		{"any reference to Tracer", references("Tracer"), []string{
			"internal/app/engine/tracer.go:<package-level>",
			"internal/app/engine/tracer.go:NewTracer",
		}},
		// 타입 선언 자신뿐(수신자 선언은 FuncDecl.Recv 라 세지 않음).
		{"any reference to GuardianAdapter", references("GuardianAdapter"), []string{
			"internal/strategydispatch/adapters.go:<package-level>",
		}},
		// EXPOSURE_RAISING 결정을 짓는 자리 전부. q_final 둘(IssuePrechecked*)과 legacy 둘.
		// 넷에 더해 **이미 정해진 class 를 옮기는** 자리 넷이 잡힘(2026-09-27 실측, 각 자리 원문):
		// gateway.go prepareRequest `req.SafetyClass = decision.SafetyClass`(기존 결정 복사),
		// decision.go DecisionRequest.build `SafetyClass: class`(요청에 이미 정해진 값의 검증 후 저장),
		// durability.go scanAttempt `rec.SafetyClass = safetyClass.String`(DB 읽기),
		// risk_bucket_issuance.go qFinalIssueDigest `SafetyClass: decision.SafetyClass`(digest 입력).
		{"exposure-raising decision producers", exposureRaisingProducer, []string{
			"internal/execgw/gateway.go:Gateway.prepareRequest",
			"internal/execgw/issue.go:Issuer.IssueEntry",
			"internal/execgw/riskguardian.go:RiskGuardian.IssueEntry",
			"internal/execgw/riskguardian_first_leg.go:RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg",
			"internal/execgw/riskguardian_qfinal.go:RiskGuardian.IssuePrecheckedQFinalEntry",
			"internal/journal/decision.go:DecisionRequest.build",
			"internal/journal/durability.go:scanAttempt",
			"internal/journal/risk_bucket_issuance.go:qFinalIssueDigest",
		}},
	}
	// 걷기 범위 자체의 단언 — 세 root 가 모두 걸리고 각각 비시험 Go 파일을 줌.
	enclosingSites(t, func(ast.Node) bool { return false })
	if strings.Join(censusRoots, ",") != "internal,cmd,tools" {
		t.Fatalf("census roots changed: %v", censusRoots)
	}
	for _, top := range censusRoots {
		if censusRootFiles[top] == 0 {
			t.Fatalf("census walked no non-test Go file under %s/ (%v)", top, censusRootFiles)
		}
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

// TestA066LedgerUsageHasOneComputation 는 a066 5.6.1 F1(과 5.7 체결 overage)의 "생산 reader 와 같은 합" 을 산문이 아니라 호출 그래프로 고정함:
// 원장 사용량을 읽고 더하는 두 내부 함수는 ReadJournalBucketUsage 만 부르고, 그 함수를 부르는 자리는 생산 snapshot
// reader 와 journal admission 의 stale 대조 둘뿐이며, 그 대조는 두 admission 트랜잭션에서만 불림.
func TestA066LedgerUsageHasOneComputation(t *testing.T) {
	cases := []struct {
		name string
		want []string
	}{
		{"readProductionRiskUsage", []string{"internal/riskbucket/production_snapshot_authority.go:ReadJournalBucketUsage"}},
		{"aggregateProductionRiskUsage", []string{"internal/riskbucket/production_snapshot_authority.go:ReadJournalBucketUsage"}},
		// a066 5.7: 체결 계상의 공유 bucket overage 도 같은 함수로 다른 진입의 사용량을 읽음.
		// a066 6.5: 제출 재검증도 같은 함수로 결정의 bucket latch 를 읽음(latchedUsageRefusal 규칙과 짝).
		{"ReadJournalBucketUsage", []string{
			"internal/journal/risk_bucket_fill.go:riskBucketSharedUsage",
			"internal/journal/risk_bucket_issuance.go:Journal.RevalidateQFinalAdmission",
			"internal/journal/risk_bucket_usage.go:refuseStaleBucketUsage",
			"internal/riskbucket/production_snapshot_authority.go:loadProductionRiskEntries",
		}},
		{"riskBucketSharedUsage", []string{
			"internal/journal/risk_bucket_fill.go:loadRiskBucketFillTransition",
		}},
		{"refuseStaleBucketUsage", []string{
			"internal/journal/risk_bucket.go:Journal.CommitRiskBucketAdmission",
			"internal/journal/risk_bucket_issuance.go:commitFreshRiskBucketAdmissionTx",
		}},
		// a066 6.5: latch 거절 규칙은 admission 대조와 제출 재검증 두 자리에서만 불림(같은 규칙 함수).
		{"latchedUsageRefusal", []string{
			"internal/journal/risk_bucket_issuance.go:Journal.RevalidateQFinalAdmission",
			"internal/journal/risk_bucket_usage.go:refuseStaleBucketUsage",
		}},
		{"ensureRiskBucketEntryScopeClean", []string{
			"internal/journal/risk_bucket.go:Journal.CommitRiskBucketAdmission",
			"internal/journal/risk_bucket_issuance.go:Journal.RevalidateQFinalAdmission",
			"internal/journal/risk_bucket_issuance.go:commitFreshRiskBucketAdmissionTx",
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := enclosingSites(t, calls(tc.name)); strings.Join(got, "\n") != strings.Join(tc.want, "\n") {
				t.Fatalf("callers of %s changed:\n got: %v\nwant: %v", tc.name, got, tc.want)
			}
		})
	}
}
