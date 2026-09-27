package journal

// a066 6.1 (B)(2) — CommitRiskBucketAdmission · commitFreshRiskBucketAdmissionTx 의 미도달 분기 중 입력으로 닿는 것:
//   - CRA B1: 계산이 거절하면(한도 소진) 그 거절을 그대로 돌려주고 아무것도 쓰지 않음.
//   - CRA B19 · commitFresh B13: 해제된 owner 의 prospective generation 을 다시 쓰면 owner INSERT 가 기본 키에 걸림 →
//     ErrRiskBucketOwnerConflict(해제된 owner 는 활성 조회에 안 잡히므로 SELECT 가 아니라 INSERT 가 막음).
//   - commitFresh B2(같은 transaction id·decision id 의 기존 발급): 두 호출자(RecordQFinalDecisionAndReserve,
//     strategy_first_leg_atomic)가 같은 트랜잭션에서 recoverQFinalIssueReplayTx(같은 술어)를 먼저 부르므로 도달하지
//     않는 백스톱임 — 속성은 행동으로, 호출 순서는 AST 로 고정함(우연한 중복을 선언된 층위로).

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

func TestA066AdmissionCalculatorRefusalWritesNothing(t *testing.T) {
	j := openTestJournal(t)
	seedExistingRiskReservation(t, j, "existing-refused", "acct-1")
	plan := riskBucketAdmissionFixture(t, "refused", "acct-1", "lane-short", "campaign-1", "prospective-refused", "100", "100")
	_, err := j.CommitRiskBucketAdmission(context.Background(), plan)
	if !riskbucket.IsRefusal(err, riskbucket.RefusalBucketCapExhausted) {
		t.Fatalf("exhausted cap: err=%v, want BUCKET_CAP_EXHAUSTED", err)
	}
	for _, table := range []string{"risk_bucket_final_decisions", "risk_bucket_owners", "risk_bucket_reservations"} {
		if got := countRiskBucketRows(t, j, table); got != 0 {
			t.Fatalf("refused admission wrote %d rows to %s", got, table)
		}
	}
}

// releaseOwnerRowForTest 는 owner 행을 해제된 모양으로 만듦(해제 절차의 증거 사슬은 이 시험의 주제가 아님 — 주제는
// 해제된 generation 의 재사용이 INSERT 에서 막히는지).
func releaseOwnerRowForTest(t *testing.T, j *Journal, prospective string) {
	t.Helper()
	if _, err := j.db.Exec(`UPDATE risk_bucket_owners SET released_at='2026-03-30T00:40:00Z' WHERE prospective_generation=?`, prospective); err != nil {
		t.Fatalf("release owner row: %v", err)
	}
}

func TestA066ReleasedProspectiveGenerationCannotBeReacquired(t *testing.T) {
	t.Run("CommitRiskBucketAdmission", func(t *testing.T) {
		j := openTestJournal(t)
		seedExistingRiskReservation(t, j, "existing-reuse-a", "acct-1")
		first := riskBucketAdmissionFixture(t, "reuse-a", "acct-1", "lane-short", "campaign-1", "prospective-reuse", "100", "0")
		if _, err := j.CommitRiskBucketAdmission(context.Background(), first); err != nil {
			t.Fatal(err)
		}
		releaseOwnerRowForTest(t, j, "prospective-reuse")
		seedExistingRiskReservation(t, j, "existing-reuse-b", "acct-1")
		second := riskBucketAdmissionFixture(t, "reuse-b", "acct-1", "lane-short", "campaign-1", "prospective-reuse", "100", "50")
		if _, err := j.CommitRiskBucketAdmission(context.Background(), second); !errors.Is(err, ErrRiskBucketOwnerConflict) {
			t.Fatalf("re-acquiring a released prospective generation: err=%v, want owner conflict", err)
		}
		if got := countRiskBucketRows(t, j, "risk_bucket_final_decisions"); got != 1 {
			t.Fatalf("decisions=%d", got)
		}
	})
	t.Run("RecordQFinalDecisionAndReserve", func(t *testing.T) {
		j := openTestJournal(t)
		first := qFinalIssueFixture(t, j, "reuse-fresh-a")
		if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), first); err != nil {
			t.Fatal(err)
		}
		releaseOwnerRowForTest(t, j, first.Admission.Owner.Key.ProspectiveGeneration)
		second := qFinalIssueFixture(t, j, "reuse-fresh-b")
		second.Admission.Owner = first.Admission.Owner
		refreshSnapshotUsageFromLedger(t, j, &second.Admission)
		if _, err := issueSecondWithFreshVersion(t, j, second); !errors.Is(err, ErrRiskBucketOwnerConflict) {
			t.Fatalf("re-acquiring a released prospective generation on the issuance path: err=%v, want owner conflict", err)
		}
		assertSingleQFinalDecision(t, j)
	})
}

func TestA066QFinalTransactionIDCannotBeReusedByAnotherDecision(t *testing.T) {
	j := openTestJournal(t)
	first := qFinalIssueFixture(t, j, "txid-a")
	if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	second := qFinalIssueFixture(t, j, "txid-b")
	second.Admission.TransactionID = first.Admission.TransactionID
	intent := second.Issue.Decision.Preimage.(RiskIntent)
	intent.PolicyVersion = first.Issue.Decision.Preimage.(RiskIntent).PolicyVersion
	second.Issue.Decision.Preimage = intent
	refreshSnapshotUsageFromLedger(t, j, &second.Admission)
	_, err := issueSecondWithFreshVersion(t, j, second)
	// 먼저 서는 것은 recoverQFinalIssueReplayTx 의 "divergent q_final issuance replay" 임(B2 는 그 뒤의 백스톱).
	if !errors.Is(err, ErrRiskBucketReplayMismatch) || !strings.Contains(err.Error(), "divergent q_final issuance replay") {
		t.Fatalf("transaction id reused by another decision: err=%v", err)
	}
	assertSingleQFinalDecision(t, j)
}

// TestA066FreshAdmissionIdentityBackstopRunsAfterReplayRecovery 는 commitFreshRiskBucketAdmissionTx 를 부르는 모든
// 비시험 함수가 **같은 직선 경로에서** 그보다 먼저 recoverQFinalIssueReplayTx 를 부름을 AST 로 단언함: 두 호출 모두 함수
// 본문 최상위 문장에 있어야 하고(갈래 본문 안의 recover 는 그 경로가 commitFresh 로 가지 않을 수 있으므로 치지 않음 —
// if 문은 init·조건만 봄), recover 문장이 앞서야 함. 호출자 수도 셈 — 새 호출자가 recover 없이 commitFresh 를 부르면
// B2 가 유일한 가드가 되므로 빨개져야 함.
func TestA066FreshAdmissionIdentityBackstopRunsAfterReplayRecovery(t *testing.T) {
	names, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	callers := 0
	for _, name := range names {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || !callsAnywhere(fn.Body, "commitFreshRiskBucketAdmissionTx") {
				continue
			}
			callers++
			recoverAt, commitAt := -1, -1
			for i, stmt := range fn.Body.List {
				if recoverAt < 0 && callsOnStraightPath(stmt, "recoverQFinalIssueReplayTx") {
					recoverAt = i
				}
				if commitAt < 0 && callsOnStraightPath(stmt, "commitFreshRiskBucketAdmissionTx") {
					commitAt = i
				}
			}
			if commitAt < 0 || recoverAt < 0 || recoverAt >= commitAt {
				t.Errorf("%s: %s — commitFreshRiskBucketAdmissionTx (top-level stmt %d) is not preceded on the straight path by recoverQFinalIssueReplayTx (stmt %d)",
					fset.Position(fn.Pos()), fn.Name.Name, commitAt, recoverAt)
			}
		}
	}
	if callers != 2 {
		t.Errorf("commitFreshRiskBucketAdmissionTx callers=%d, want 2 (RecordQFinalDecisionAndReserve, strategy first leg) — re-check the backstop layering", callers)
	}
}

func callsAnywhere(node ast.Node, name string) bool {
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if id, ok := call.Fun.(*ast.Ident); ok && id.Name == name {
				found = true
			}
		}
		return !found
	})
	return found
}

// callsOnStraightPath 는 최상위 문장이 갈래 본문 밖에서 name 을 부르는지 봄(if 는 init·조건만).
func callsOnStraightPath(stmt ast.Stmt, name string) bool {
	switch s := stmt.(type) {
	case *ast.IfStmt:
		return (s.Init != nil && callsAnywhere(s.Init, name)) || callsAnywhere(s.Cond, name)
	case *ast.AssignStmt, *ast.ExprStmt, *ast.DeclStmt:
		return callsAnywhere(s, name)
	}
	return false
}
