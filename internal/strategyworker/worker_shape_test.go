package strategyworker

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"
)

// 이 파일은 `owns` 가 **무엇을 비교하는지**를 구조로 못 박는다.
//
// 왜 행동 시험만으로는 안 되는가. 다섯 비교 중 넷은 서로를 가려 준다 —
// 뮤테이션으로 재어 보니 시장·레인 ID·레인 버전·지평 비교를 **하나씩** 지워도
// 열두 시험이 전부 초록이었다. 이유는 우연이 아니다: 가족 유도
// (`strategyarbiter.familyScore`)가 이미 (horizon, lane_id, lane_version) 세 축으로
// 권한의 점수 행을 찾고, 레인 ID 는 시장마다 다르다. 그래서 한 축을 지워도 남은
// 축이 같은 입력을 막는다.
//
// 그것을 "동등 변이"라고 넘기면 안 된다. 넷을 **함께** 지우면 US worker 가 KR
// 제안을 받는다 — 즉 각 비교는 정답에 필요하고, 다만 행동으로 개별 관측이 안 될
// 뿐이다. 축마다 시험을 하나씩 사는 일은 끝나지 않는다는 것이 이 change 가
// 13 라운드에 걸쳐 얻은 결론이고(`strategy_first_leg_backstop_shape_test.go`),
// 답도 같다: 행동은 "가드가 돌고 거절한다"를 증명하고, 구조는 "가드가 이 다섯을
// 비교한다"를 증명한다.
//
// 이것은 철자 census 가 아니다. "토큰이 파일 어딘가에 있는가"가 아니라
// "이 반환식의 피연산자 목록이 정확히 이것인가"를 묻는다 — 다른 곳의 등장으로는
// 만족시킬 수 없다.
func TestOwnsComparesExactlyTheFiveThingsThatBindAProposalToThisWorker(t *testing.T) {
	body := ownsBody(t)

	// 1. 봉인부터 본다. 이 문장이 없으면 아래 계보 비교는 위조된 값을 읽는다.
	guard, ok := body.List[0].(*ast.IfStmt)
	if !ok || render(t, guard.Cond) != "!proposal.Result.ValidProposal()" {
		t.Fatalf("owns 의 첫 문장은 봉인 검사여야 한다: %s", render(t, body.List[0]))
	}

	final, ok := body.List[len(body.List)-1].(*ast.ReturnStmt)
	if !ok || len(final.Results) != 1 {
		t.Fatal("owns 의 마지막 문장은 값 하나를 돌려주는 return 이어야 한다")
	}

	got := conjuncts(t, final.Results[0])
	want := []string{
		"lineage.Market == worker.key.Market",
		"lineage.LaneID == worker.key.LaneID",
		"lineage.LaneVersion == worker.key.LaneVersion",
		"lineage.Horizon == worker.horizon",
		"strategyarbiter.ProposalFamily(proposal) == worker.key.Family",
	}
	if strings.Join(got, " & ") != strings.Join(want, " & ") {
		t.Fatalf("owns 의 비교가 바뀌었다.\n got: %v\nwant: %v\n\n"+
			"다섯 중 넷은 행동 시험이 개별로 못 잡는다(서로를 가려 준다). 하나를 지우면"+
			" 그 순간은 초록이지만, 둘째를 지우는 순간 남의 시장·남의 레인 제안이 통과한다."+
			" 지워도 된다고 판단했다면 그 판단을 여기 want 에 적고 이유를 남길 것 —"+
			" 조용히 사라지면 아무도 그 판단을 안 본다.", got, want)
	}
}

// 가족은 계보가 신고한 값이 아니라 봉인된 권한에서 유도한 값이어야 한다.
//
// 위 단언은 문자열로 그것을 요구하지만, 문자열이 맞아도 그 함수가 계보를 그냥
// 되읽는 것이면 검사는 언제나 참이다. 그래서 유도 함수의 인자가 **제안 전체**인지
// (즉 권한을 들고 가는지) 따로 확인한다 — `ProposalFamily(lineage)` 였다면
// 유도가 아니라 신고를 읽는 것이다.
func TestTheFamilyIsDerivedFromTheWholeSealedProposal(t *testing.T) {
	found := 0
	ast.Inspect(ownsBody(t), func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || render(t, call.Fun) != "strategyarbiter.ProposalFamily" {
			return true
		}
		found++
		if len(call.Args) != 1 || render(t, call.Args[0]) != "proposal" {
			t.Errorf("가족 유도가 제안 전체를 받지 않는다: %s", render(t, call))
		}
		return true
	})
	if found != 1 {
		t.Fatalf("가족 유도 호출을 %d 개 찾았다 — 정확히 하나여야 한다", found)
	}
}

func ownsBody(t *testing.T) *ast.BlockStmt {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "worker.go", nil, 0)
	if err != nil {
		t.Fatalf("parse worker.go: %v", err)
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "owns" || fn.Recv == nil || fn.Body == nil {
			continue
		}
		if len(fn.Body.List) < 2 {
			t.Fatal("owns 의 본문이 너무 짧다")
		}
		return fn.Body
	}
	t.Fatal("worker.go 에서 owns 메서드를 찾지 못했다")
	return nil
}

// conjuncts 는 `a && b && c` 를 [a b c] 로 편다. 다른 연산자가 섞이면 실패한다 —
// `||` 하나가 끼면 다섯 중 넷이 있으나 마나가 되기 때문이다.
func conjuncts(t *testing.T, expr ast.Expr) []string {
	t.Helper()
	binary, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return []string{render(t, expr)}
	}
	if binary.Op == token.LAND {
		return append(conjuncts(t, binary.X), conjuncts(t, binary.Y)...)
	}
	if binary.Op != token.EQL {
		t.Fatalf("owns 의 반환식에 %s 연산자가 있다 — && 로 이은 == 비교만 있어야 한다", binary.Op)
	}
	return []string{render(t, expr)}
}

func render(t *testing.T, node ast.Node) string {
	t.Helper()
	expr, ok := node.(ast.Expr)
	if !ok {
		return "<not an expression>"
	}
	return types.ExprString(expr)
}
