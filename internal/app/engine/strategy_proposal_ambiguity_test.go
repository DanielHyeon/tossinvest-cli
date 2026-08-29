package engine

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// 리뷰 지적 C2: 한 종목의 fail-closed 가 시장 수준의 fail-open 이 되면 안 된다.
//
// 무슨 일이 벌어질 수 있었나. 제안 단계는 종목마다 `batch.For(symbol)` 을 부른다.
// 한 종목이 두 가족 이상을 제안하면 `For` 는 "고르지 않겠다"며 닫아서 거절한다.
// 그런데 부르는 쪽이 그것을 그냥 `refused++` 로 세고 넘어가면 그 종목은 목록에서
// 빠진다. 목록이 둘에서 하나로 줄면, 아래 세 관문이 모두 쓰는 `len(entries) != 1`
// 이 오히려 *만족*되어 남은 *다른* 종목의 발주가 풀린다. 한 종목을 막으려던
// 결정이 다른 종목을 열어 준다.
//
// 관문 셋(`ResultAuthority`, `strategyAccountAuthorityLoader.collectMarket`,
// `strategyProjectionFromAssembly`)은 모두 같은 `strategyProposalMarketAuthority.entries`
// 를 읽는다. 그래서 고칠 자리는 관문 셋이 아니라 그 목록을 만드는 한 곳,
// 즉 `strategyProposalAuthorityLoader.collectMarket` 이다.
//
// 이 검사는 그 한 곳에서 (1) 모호함을 묻고 (2) 목록을 만들기 *전에* 시장을 닫는지를
// 구문으로 확인한다. 순서가 뒤집히면 고친 의미가 없으므로 순서까지 본다.
func TestProposalCollectMarketClosesTheMarketBeforeBuildingEntries(t *testing.T) {
	const source = "strategy_proposal_authority.go"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, source, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", source, err)
	}
	var target *ast.FuncDecl
	ast.Inspect(file, func(node ast.Node) bool {
		decl, ok := node.(*ast.FuncDecl)
		if ok && decl.Name.Name == "collectMarket" && decl.Recv != nil {
			target = decl
		}
		return true
	})
	if target == nil {
		t.Fatalf("%s no longer declares collectMarket", source)
	}

	ambiguousGuardEnd := token.NoPos
	ast.Inspect(target, func(node ast.Node) bool {
		statement, ok := node.(*ast.IfStmt)
		if !ok {
			return true
		}
		calls, returns := false, false
		ast.Inspect(statement.Cond, func(inner ast.Node) bool {
			call, isCall := inner.(*ast.CallExpr)
			if !isCall {
				return true
			}
			if selector, isSelector := call.Fun.(*ast.SelectorExpr); isSelector && selector.Sel.Name == "Ambiguous" {
				calls = true
			}
			return true
		})
		for _, inner := range statement.Body.List {
			ast.Inspect(inner, func(node ast.Node) bool {
				if _, isReturn := node.(*ast.ReturnStmt); isReturn {
					returns = true
				}
				return true
			})
		}
		if calls && returns && (ambiguousGuardEnd == token.NoPos || statement.End() < ambiguousGuardEnd) {
			ambiguousGuardEnd = statement.End()
		}
		return true
	})
	if ambiguousGuardEnd == token.NoPos {
		t.Fatal("collectMarket does not refuse the whole market when a symbol proposes more than one family; " +
			"dropping that symbol instead makes len(entries)==1 satisfiable and releases a different symbol")
	}

	firstEntryAppend := token.NoPos
	ast.Inspect(target, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			return true
		}
		ident, isIdent := assign.Lhs[0].(*ast.Ident)
		if !isIdent || ident.Name != "entries" {
			return true
		}
		call, isCall := assign.Rhs[0].(*ast.CallExpr)
		if !isCall {
			return true
		}
		if fun, isIdent := call.Fun.(*ast.Ident); isIdent && fun.Name == "append" {
			if firstEntryAppend == token.NoPos || assign.Pos() < firstEntryAppend {
				firstEntryAppend = assign.Pos()
			}
		}
		return true
	})
	if firstEntryAppend == token.NoPos {
		t.Fatal("collectMarket no longer appends to entries; this guard's ordering claim has no anchor")
	}
	if ambiguousGuardEnd >= firstEntryAppend {
		t.Fatalf("the ambiguity refusal ends at %s but entries are first appended at %s; "+
			"a guard that runs after the list is built cannot stop the list from shrinking",
			fset.Position(ambiguousGuardEnd), fset.Position(firstEntryAppend))
	}
}
