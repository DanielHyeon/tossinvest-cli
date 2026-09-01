package engine

import (
	"go/ast"
	"go/token"
	"go/types"
	"sort"
	"strings"
	"testing"
)

// 이 파일은 1차 진입 backstop 이 **봉인된 identity 자체**를 비교한다는 것을
// 구조로 못 박는다. 단언은 하나다.
//
// 왜 하나여야 하는가. 5~9차 적대 리뷰가 매 라운드 같은 결론에 왔다: 엔진 쪽에서
// "이 축을 바꿔도 거절하는가"를 시험으로 쪼개면 **다음 축이 항상 남는다.**
// `executionTermsIdentity` 가 담는 스칼라가 32 개이고(정확히 32 개라는 것은
// `strategyflow/execution_terms_identity_fields_test.go` 가 타입에서 읽어 못
// 박는다), 그중 대부분은 시험 seam 이 하나만 움직이지도 못한다. 라운드마다
// 하나씩 사면 끝나지 않는 값을 사는 것이다.
//
// 의무를 둘로 나누면 각각 끝난다.
//
//  1. **무엇이 identity 에 담기는가** — `strategyflow` 가 자기 필드로 32 개를 센다.
//  2. **이 가드가 identity 를 비교하는가** — 아래 단언 하나.
//
// 그리고 이것은 이 change 가 다섯 라운드에 걸쳐 지운 **철자 census 가 아니다.**
// 그것들은 "토큰 X 가 범위 어딘가에 있는가"를 물었고, 그래서 X 가 다른 곳에
// 나타나면 뚫렸다. 여기서는 **그 비교의 피연산자가 이 식인가**를 묻는다 —
// 가드 그 자체인 AST 노드에 대한 구조 등식이라, 다른 곳의 등장으로는 만족시킬
// 수 없다. 행동 시험 넷이 "이 가드가 돌고 거절한다"를 이미 증명하므로,
// 행동 + 구조 + 해시 완전성 셋이 함께여야 닫힌다.

// firstLegIdentityRefusal 은 위조된 제안을 거절하는 문구다. 행동 시험 넷이
// 요구하는 문구와 같아야 한다 — 여기서 문구가 갈리면 두 시험이 서로 다른 줄을
// 보게 된다.
const firstLegIdentityRefusal = "production proposal identity changed"

// TestTheFirstLegBackstopComparesTheSealedIdentitiesThemselves 는 그 거절이
// 걸린 `if` 의 조건이 **정확히 두 identity 비교의 논리합**임을 요구한다.
//
// 6~9차의 모든 뮤턴트가 이 단언에서 죽는다: `Lineage.Market` 만, `Lineage.Identity`
// 만, `CampaignID` 만, `ExecutionTerms.Identity()` 만, 세 가격의 `priceMinor` 만,
// 거기에 `EffectiveStop().Source()` 를 더한 것 — 전부 이 피연산자들을 바꾼다.
func TestTheFirstLegBackstopComparesTheSealedIdentitiesThemselves(t *testing.T) {
	file := parseEngineFile(t, "strategy_account_first_leg_authority.go")
	guards := make([]*ast.IfStmt, 0, 1)
	ast.Inspect(file, func(node ast.Node) bool {
		if stmt, ok := node.(*ast.IfStmt); ok && refusesWith(stmt.Body, firstLegIdentityRefusal) {
			guards = append(guards, stmt)
		}
		return true
	})
	if len(guards) != 1 {
		t.Fatalf("%q 로 거절하는 if 를 %d 개 찾았다 — 정확히 하나여야 이 단언이 그 가드를 가리킨다",
			firstLegIdentityRefusal, len(guards))
	}
	requireOnlyTheRefusal(t, guards[0].Body)
	requireGuardFollowsTheRederivation(t, file, guards[0])
	got := disjunctComparisons(t, guards[0].Cond)
	want := []string{
		"accepted.result.ExecutionTerms.Identity() != result.ExecutionTerms.Identity()",
		"accepted.result.Lineage.Identity != result.Lineage.Identity",
	}
	if strings.Join(got, " | ") != strings.Join(want, " | ") {
		t.Fatalf("backstop 의 비교가 바뀌었다.\n got: %v\nwant: %v\n\n"+
			"이 가드는 **봉인된 identity 두 개**를 비교해야 한다. 비교를 필드별로 펼치면"+
			" (가격·수량·출처를 따로 나열하면) 하나를 빠뜨리는 순간 그 값이 위조 가능해지고,"+
			" 빠뜨린 필드마다 엔진 시험을 새로 만드는 일은 끝나지 않는다. identity 가 모든"+
			" 필드를 담는다는 것은 strategyflow/execution_terms_identity_fields_test.go 가"+
			" 타입에서 세어 증명한다 — 그러니 여기서는 identity 를 그대로 비교하면 된다.\n\n"+
			"왼쪽 항(Lineage.Identity)은 오른쪽에 포섭되지만 남겨 둔다: 지워도 안전하나,"+
			" 지우는 편집과 오른쪽을 지우는 편집은 한 글자 차이다.", got, want)
	}
}

// requireGuardFollowsTheRederivation 은 가드가 비교하는 `result` 가 **loader 자기
// 권한 쌍에서 그대로 꺼낸 값**임을 요구한다. 재유도는 두 고리짜리 사슬이다.
//
//	proposal, ... := loader.proposals.forMarket(market)     // 사슬의 뿌리
//	proposalAuthority := proposal.entries[0].authority      // 고리 1
//	result := proposalAuthority.Proposal()                  // 고리 2
//	<가드>
//
// 세 이름이 각각 한 번만 대입되고, 세 문장이 이 순서로 붙어 있어야 한다.
//
// 12차 적대 리뷰가 고리 1 만 안 잡혀 있을 때를 보였다. 아래는 6.2 가 실제로 쓸
// 편집(네 가족 중 조정자가 고른 것을 찾기)이고, 고리 2 와 가드는 한 바이트도
// 안 바뀐다:
//
//	proposalAuthority := proposal.entries[0].authority
//	for _, entry := range proposal.entries {
//	    if entry.authority.Proposal().Lineage.Identity == accepted.result.Lineage.Identity {
//	        proposalAuthority = entry.authority
//	    }
//	}
//
// 그러면 가드가 **자기 참조**가 된다 — accepted 와 맞는 항목을 골라 놓고 그것을
// accepted 와 비교한다. 오늘은 `:217` 의 `len(entries) != 1` 이 그 loop 를 한
// 항목으로 묶어 두어 해가 없지만, **6.2 가 바로 그 둘을 바꾼다.**
//
// 11차 적대 리뷰가 왜 필요한지 보였다. 10차의 탈출구를 가드 본문 밖, **한 줄 위**로
// 옮기면 가드 노드는 한 바이트도 안 바뀐다:
//
//	if result.Lineage.PlannedCeiling != accepted.result.Lineage.PlannedCeiling {
//	    result = accepted.result
//	}
//	<가드 그대로>
//
// 대입이 비교보다 먼저 일어나므로 가드가 볼 때는 identity 가 이미 같고, 가드는
// 옳게 아무것도 안 한다. 구조 단언·행동 시험 여섯·두 스위트가 전부 초록이었고
// 위조된 수량 9 · 손절 80 짜리 1차 진입이 나갔다.
//
// **그때 B3 에 "가드 앞은 다른 가드(B2·B4)가 맡는다"고 적었는데 그것은 틀렸다.**
// 리뷰어가 재 보니 B2(`:217` 개수 관문)도 B4(`:227` 위험 범위)도 그 편집에서
// 초록이다. 아무도 안 맡고 있었다 — 이름만 적고 주인을 잘못 적은 것이다.
//
// 1 번은 감싸기도 함께 막는다. 가드를 바깥 `if` 안에 넣으면 최상위 목록에서
// 사라지므로 여기서 터진다. (`refusesWith` 가 재귀하기 때문에 감싼 `if` 도
// 후보로 잡혀 "가드 2 개"로도 터지지만, 그것은 **우연히** 막는 것이다.
// 나중에 `refusesWith` 를 직속 문장만 보도록 "정리"하면 그 방어는 조용히
// 사라진다. 여기 1 번이 그것을 의도적으로 막는 자리다.)
func requireGuardFollowsTheRederivation(t *testing.T, file *ast.File, guard *ast.IfStmt) {
	t.Helper()
	const rederivation = "result := proposalAuthority.Proposal()"
	var body *ast.BlockStmt
	ast.Inspect(file, func(node ast.Node) bool {
		decl, ok := node.(*ast.FuncDecl)
		if ok && decl.Name != nil && decl.Name.Name == "collectStrategyFirstLegAuthority" {
			body = decl.Body
		}
		return body == nil
	})
	if body == nil {
		t.Fatal("collectStrategyFirstLegAuthority 를 못 찾았다 — 이 단언이 공허해진다")
	}
	position := -1
	for index, statement := range body.List {
		if statement == ast.Stmt(guard) {
			position = index
			break
		}
	}
	if position < 0 {
		t.Fatal("가드가 함수 최상위 문장이 아니다 — 다른 블록 안에 있으면 그 블록의" +
			" 조건이 참일 때만 도는 가드가 된다")
	}
	chain := []string{"proposalAuthority := proposal.entries[0].authority", rederivation}
	if position < len(chain) {
		t.Fatalf("가드 앞에 문장이 %d 개뿐이다 — 재유도 사슬 %d 고리가 들어갈 자리가 없다",
			position, len(chain))
	}
	for offset, want := range chain {
		got := statementText(body.List[position-len(chain)+offset])
		if got != want {
			t.Fatalf("재유도 사슬의 고리 %d 이 %q 가 아니라 %q 다 — 가드가 보는 값이"+
				" loader 자기 권한 쌍에서 그대로 나왔다는 보장이 사라진다.\n\n"+
				"L6 6.2 가 여러 항목 중 하나를 고르도록 바꾼다면, **그 선택이 accepted 에"+
				" 기대면 안 된다.** accepted 와 맞는 항목을 골라 놓고 그것을 accepted 와"+
				" 비교하면 이 가드는 자기 참조가 되고, 이 시험은 그대로 초록이다."+
				" 기대 문자열만 고쳐서 통과시키는 것이 가장 쉬운 길이므로 여기 적어 둔다 —"+
				" 고칠 때는 새 선택이 무엇에 기대는지를 먼저 정하고, 그 성질을 여기에"+
				" 다시 못 박아야 한다.", offset+1, want, got)
		}
	}
	// 사슬의 세 이름이 각각 한 번만 대입돼야 한다. 하나라도 두 번 대입되면 가드가
	// 비교하는 값이 재유도된 값이 아닐 수 있다.
	for _, name := range []string{"proposal", "proposalAuthority", "result"} {
		if count := assignmentsTo(body, name); count != 1 {
			t.Fatalf("%s 가 함수 안에서 %d 번 대입된다 — 한 번이어야 한다."+
				" 두 번째 대입은 가드가 무엇을 비교하는지를 바꾼다", name, count)
		}
	}
}

// assignmentsTo 는 그 이름이 몇 번 대입되는지 센다.
//
// `&x` 도 대입으로 센다. 포인터를 넘기면 부르는 쪽이 값을 바꿀 수 있고, 그것은
// `*ast.AssignStmt` 가 아니라 `*ast.UnaryExpr` 라서 대입만 세면 안 보인다
// (12차 적대 리뷰). 오늘 그 구멍으로는 해를 못 만든다 — 가드가 뒤따르는 모든
// 문장을 지배하므로 가드 뒤의 채택은 identity 가 이미 같을 때만 돌고, identity
// 동일은 봉인된 값 전체의 동일이라 채택이 무의미하다. 그래도 세는 편이 싸다.
func assignmentsTo(body *ast.BlockStmt, name string) int {
	count := 0
	ast.Inspect(body, func(node ast.Node) bool {
		switch stmt := node.(type) {
		case *ast.AssignStmt:
			for _, target := range stmt.Lhs {
				if identifierNamed(target, name) {
					count++
				}
			}
		case *ast.RangeStmt:
			if identifierNamed(stmt.Key, name) || identifierNamed(stmt.Value, name) {
				count++
			}
		case *ast.UnaryExpr:
			if stmt.Op == token.AND && identifierNamed(stmt.X, name) {
				count++
			}
		}
		return true
	})
	return count
}

// statementText 는 문장을 한 줄 문자열로 만든다. 대입문만 필요하므로 그 모양만 쓴다.
func statementText(statement ast.Stmt) string {
	assign, ok := statement.(*ast.AssignStmt)
	if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return ""
	}
	return types.ExprString(assign.Lhs[0]) + " " + assign.Tok.String() + " " + types.ExprString(assign.Rhs[0])
}

func identifierNamed(node ast.Expr, name string) bool {
	identifier, ok := node.(*ast.Ident)
	return ok && identifier.Name == name
}

// requireOnlyTheRefusal 은 가드의 본문이 **그 거절 하나뿐**임을 요구한다.
//
// 10차 적대 리뷰가 이것이 없을 때를 보였다. 조건을 한 글자도 안 바꾸고 본문
// 안에 분기를 하나 넣으면 — `PlannedCeiling` 이 다르면 `result = accepted.result`
// 로 건너온 값을 그대로 믿고, 아니면 거절 — 구조 단언과 행동 시험 여섯이 전부
// 초록이었고 위조된 수량 9 · 손절 80 짜리 1차 진입이 나갔다. fixture 가 전부
// ceiling 8 이라 그 탈출구를 밟지 않았기 때문이다.
//
// 앞 판본의 `refusesWith` 는 "이 블록 어딘가에 그 문구가 있는가"를 물었다.
// 그것은 이 change 가 다섯 라운드에 걸쳐 지운 **존재 검사**이고, 조건은 구조로
// 못 박으면서 본문은 존재로 본 셈이다. 이제 본문도 구조로 본다: 문장 하나,
// 반환 둘, 두 번째가 그 문구의 `errors.New` 여야 한다.
func requireOnlyTheRefusal(t *testing.T, body *ast.BlockStmt) {
	t.Helper()
	if body == nil || len(body.List) != 1 {
		t.Fatalf("가드 본문이 문장 %d 개다 — 거절 하나여야 한다. 본문에 무엇이든 더하면"+
			" 조건이 그대로여도 건너온 값을 믿는 길이 생긴다", len(body.List))
	}
	ret, ok := body.List[0].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 2 {
		t.Fatalf("가드 본문이 두 값을 돌려주는 return 이 아니다: %T", body.List[0])
	}
	if got := types.ExprString(ret.Results[0]); got != "execgw.QFinalCampaignFirstLegIssuance{}" {
		t.Fatalf("가드가 빈 발주가 아닌 것을 돌려준다: %s", got)
	}
	call, ok := ret.Results[1].(*ast.CallExpr)
	if !ok || types.ExprString(call.Fun) != "errors.New" || len(call.Args) != 1 {
		t.Fatalf("가드가 errors.New 로 거절하지 않는다: %s", types.ExprString(ret.Results[1]))
	}
	literal, ok := call.Args[0].(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING || literal.Value != `"`+firstLegIdentityRefusal+`"` {
		t.Fatalf("거절 문구가 %q 가 아니다: %s", firstLegIdentityRefusal, types.ExprString(call.Args[0]))
	}
}

// refusesWith 는 **가드를 찾는** 데만 쓴다. 이것은 존재 검사이고, 그래도 되는
// 이유는 여기서 나온 후보를 requireOnlyTheRefusal 이 구조로 다시 보기 때문이다.
// 찾는 데 존재 검사를 쓰는 것과 증명하는 데 쓰는 것은 다르다.
func refusesWith(body *ast.BlockStmt, message string) bool {
	if body == nil {
		return false
	}
	found := false
	ast.Inspect(body, func(node ast.Node) bool {
		literal, ok := node.(*ast.BasicLit)
		if ok && literal.Kind == token.STRING && literal.Value == `"`+message+`"` {
			found = true
		}
		return !found
	})
	return found
}

// disjunctComparisons 는 조건을 `||` 로 펼쳐 각 항이 `!=` 비교인지 보고,
// 비교마다 두 피연산자를 정렬해 문자열로 돌려준다. 정렬은 좌우를 바꿔 쓴 것이
// 다른 가드로 읽히지 않게 하려는 것이고, 항 자체도 정렬해 순서에 기대지 않는다.
func disjunctComparisons(t *testing.T, cond ast.Expr) []string {
	t.Helper()
	var walk func(ast.Expr) []ast.Expr
	walk = func(node ast.Expr) []ast.Expr {
		if binary, ok := node.(*ast.BinaryExpr); ok && binary.Op == token.LOR {
			return append(walk(binary.X), walk(binary.Y)...)
		}
		return []ast.Expr{node}
	}
	terms := walk(cond)
	out := make([]string, 0, len(terms))
	for _, term := range terms {
		binary, ok := term.(*ast.BinaryExpr)
		if !ok || binary.Op != token.NEQ {
			t.Fatalf("조건의 항이 `!=` 비교가 아니다: %s", types.ExprString(term))
		}
		sides := []string{types.ExprString(binary.X), types.ExprString(binary.Y)}
		sort.Strings(sides)
		out = append(out, sides[0]+" != "+sides[1])
	}
	sort.Strings(out)
	return out
}
