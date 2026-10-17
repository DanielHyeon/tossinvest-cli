package journal

// a066 6.1 (B)(2) — loadRiskBucketFillTransition B12: 한 owner 의 여러 결정이 같은 bucket 에 다른 snapshot 한도를 들고
// 있으면 체결 계상의 한도는 **가장 작은 값**(보수 방향)임. 이 분기는 어떤 시험도 실행하지 않았음 — 기존 scale-in 시험은
// 모두 뒤 결정의 한도가 크거나 같았음. 순서(owner_sequence)대로 100 → 80 을 주면 80 이, 80 → 100 을 주면 여전히 80 이
// 되어야 함(앞 결정이 작으면 뒤의 큰 값이 한도를 넓히지 못함).

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

func TestA066FillTransitionUsesTheSmallestDecisionLimitPerBucket(t *testing.T) {
	for _, tc := range []struct {
		name, firstLimit, scaleInLimit string
	}{
		{"later decision tighter", "100", "80"},
		{"earlier decision tighter", "80", "100"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			suffix := "limit-min-" + tc.firstLimit + "-" + tc.scaleInLimit
			j := openTestJournal(t)
			key, firstDecision, firstReserved := seedRiskBucketFillFixtureWithLimit(t, j, suffix, "risk-"+suffix, tc.firstLimit)
			if err := j.RegisterRiskBucketOrder(ctx, RiskBucketOrderPlan{OrderID: "risk-" + suffix, DecisionID: firstDecision, OrderQuantity: 10, ReservedMinor: firstReserved, CreatedAt: riskFillNow}); err != nil {
				t.Fatal(err)
			}
			// a066 6.5: 앞 결정이 더 좁으면(80) scale-in 은 기록된 최소 한도 안에서만 admit 됨(원장 50 + 30 ≤ 80) — 수량 6.
			commitRiskBucketScaleInQuantity(t, j, key, suffix+"-second", tc.scaleInLimit, "50", 6)
			tx, err := j.db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			order, found, err := riskBucketOrderByID(ctx, tx, key, firstDecision, "risk-"+suffix)
			if err != nil || !found {
				t.Fatalf("order record found=%v err=%v", found, err)
			}
			state, _, err := loadRiskBucketFillTransition(ctx, tx, order, "1", nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(state.Buckets) != 5 {
				t.Fatalf("buckets=%d", len(state.Buckets))
			}
			for bucket, usage := range state.Buckets {
				if usage.LimitMinor != "80" {
					t.Fatalf("%s limit=%s, want the smaller decision limit 80", bucket.Dimension, usage.LimitMinor)
				}
			}
		})
	}
}

// commitRiskBucketScaleInQuantity 는 commitRiskBucketScaleIn 과 같되 후보 수량을 정함(기록된 최소 한도 안에 들게 하려고).
func commitRiskBucketScaleInQuantity(t *testing.T, j *Journal, key riskbucket.OwnerKey, suffix, limit, held string, quantity uint64) {
	t.Helper()
	existingID := "existing-fill-" + suffix
	seedExistingRiskReservation(t, j, existingID, key.AccountID)
	plan := riskBucketAdmissionFixture(t, "fill-"+suffix, key.AccountID, "lane-short", "campaign-1", key.ProspectiveGeneration, limit, held)
	plan.ExistingReservationID = existingID
	plan.Owner.Key = key
	plan.Admission.QCandidate, plan.Admission.QExistingGuardian = quantity, quantity
	plan.Admission.Policy.QuoteCurrency = "USD"
	plan.Admission.Policy.AccountCurrency = "KRW"
	rebindRiskBucket(t, &plan, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: string(key.Market), PolicyVersion: "policy-v1"})
	rebindRiskBucket(t, &plan, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: key.Symbol, PolicyVersion: "policy-v1"})
	if _, err := j.CommitRiskBucketAdmission(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
}

// seedRiskBucketFillFixtureWithLimit 는 seedRiskBucketFillFixture 와 같되 첫 결정의 bucket 한도를 정함.
func seedRiskBucketFillFixtureWithLimit(t *testing.T, j *Journal, suffix, orderID, limit string) (riskbucket.OwnerKey, string, map[riskbucket.BucketKey]string) {
	t.Helper()
	seedExistingRiskReservation(t, j, "existing-fill-"+suffix, "acct-1")
	plan := riskBucketAdmissionFixture(t, "fill-"+suffix, "acct-1", "lane-short", "campaign-1", "prospective-fill-"+suffix, limit, "0")
	plan.ExistingReservationID = "existing-fill-" + suffix
	plan.Owner.Key.Market = riskbucket.MarketUS
	plan.Owner.Key.Symbol = "AAPL"
	plan.Admission.Policy.QuoteCurrency = "USD"
	plan.Admission.Policy.AccountCurrency = "KRW"
	rebindRiskBucket(t, &plan, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: string(riskbucket.MarketUS), PolicyVersion: "policy-v1"})
	rebindRiskBucket(t, &plan, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "AAPL", PolicyVersion: "policy-v1"})
	receipt, err := j.CommitRiskBucketAdmission(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrder(t, j, "risk-intent-"+suffix, "risk-attempt-"+suffix, orderID)
	bindRiskOrderAttemptDecision(t, j, orderID, receipt.DecisionID)
	return plan.Owner.Key, receipt.DecisionID, riskReservedMap("50")
}

// TestA066RevalidateReportsNotRequiredForNonQFinalDecisions 는 RevalidateQFinalAdmission B3·B4 를 고정함: q_final 표식이
// 없는 결정은 "q_final 권위가 필요 없음"(false, nil)으로 답하고, 그 뒤의 bucket·잠금 판정을 하지 않음. Gateway 의
// checkReservation 은 이 답을 받으면 기존 집계 예약 검사만으로 진행함 — a066 이전 결정과 위험 감소 결정이 a066 판정에
// 끌려 들어가지 않는다는 경계(토글 OFF = upstream 동일).
func TestA066RevalidateReportsNotRequiredForNonQFinalDecisions(t *testing.T) {
	ctx := context.Background()
	// B3 를 끄면 RiskIntent 가 아닌 결정은 영값 RiskIntent 로 B4 에 닿고, 빈 정책 버전은 q_final 표식이 아니므로 같은
	// 답(false, nil)이 남(변이 "B3 skipped" 생존 — mutation-6.1/revalidate-ledger.tsv). 그 백스톱의 전제를 못 박음.
	if _, _, required := splitQFinalPolicyVersion(""); required {
		t.Fatal("an empty policy version must not carry q_final authority (B3's backstop)")
	}
	for _, tc := range []struct {
		name    string
		request func(*testing.T, *Journal) DecisionRequest
	}{
		// B3: RiskIntent 가 아닌 preimage(위험 감소 결정).
		{"reduction preimage", func(t *testing.T, _ *Journal) DecisionRequest { return reductionRequest(t) }},
		// B4: RiskIntent 이지만 q_final 정책 버전 표식이 없음(a066 이전 진입 결정).
		{"legacy risk intent", func(t *testing.T, _ *Journal) DecisionRequest { return riskRequest(t) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			decision, err := j.RecordDecision(ctx, tc.request(t, j))
			if err != nil {
				t.Fatal(err)
			}
			// 잠금을 켜 두어도 필요 없음 판정이 먼저라 잠금 판정까지 가지 않음.
			for _, market := range []riskbucket.Market{riskbucket.MarketKR, riskbucket.MarketUS} {
				for _, horizon := range []riskbucket.Horizon{riskbucket.HorizonShort, riskbucket.HorizonMedium} {
					if _, _, err := j.ActivateEntryLossLock(ctx, EntryLossLock{AccountRef: "acct-1", Market: market, Horizon: horizon, Cause: "a066 not-required probe", ActivatedAt: riskFillNow}); err != nil {
						t.Fatal(err)
					}
				}
			}
			required, err := j.RevalidateQFinalAdmission(ctx, decision.ID)
			if err != nil || required {
				t.Fatalf("non-q_final decision: required=%v err=%v, want false/nil", required, err)
			}
		})
	}
}

// TestA066RevalidateRefusesAMissingDimensionReservation 는 RevalidateQFinalAdmission 이 다섯 차원 예약 중 하나가 빠진
// q_final 결정을 제출 전 재검증에서 거절함을 고정함(BTM 행 변이: 차원 누락 루프 B15 를 꺼도 스위트가 초록이었음 —
// 누락을 만드는 시험이 없었음). 어느 가드가 먼저 서는지는 원장 층위의 문제라 여기서는 결과(거절 · 필요함)만 단언함.
func TestA066RevalidateRefusesAMissingDimensionReservation(t *testing.T) {
	j := openTestJournal(t)
	request := qFinalIssueFixture(t, j, "missing-dimension")
	if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`DELETE FROM risk_bucket_reservations WHERE decision_id=? AND bucket_dimension='sector'`, request.Issue.Decision.ID); err != nil {
		t.Fatal(err)
	}
	required, err := j.RevalidateQFinalAdmission(context.Background(), request.Issue.Decision.ID)
	if err == nil || !required {
		t.Fatalf("a q_final decision missing its sector reservation must be refused: required=%v err=%v", required, err)
	}
	t.Logf("refused by: %v", err)
}

// TestA066RevalidateDimensionGuardsAreLayered 는 RevalidateQFinalAdmission 의 차원 누락 가드 둘을 선언된 층위로 고정함
// (Manager 규칙 2026-09-28: 변이가 상호 엄폐를 실증한 쌍만 핀). 누락 차원은 먼저 verifyRiskBucketStateDigest(상태 재구성의
// "bucket count")가 거절하고, 뒤의 RequiredDimensionOrder 루프(B15/B16)는 백스톱임 — BTM 행 변이에서 루프를 꺼도
// 앞 가드가 막아 살아남았음(mutation-6.1/revalidate-ledger.tsv). 행동 시험은 먼저 서는 가드만 보므로 위치를 AST 로 봄:
// ① 상태 digest 대조가 함수 본문 최상위 직선 경로에서 ② 차원 루프보다 앞이고 ③ 루프는 누락 차원에서 return 하며
// ④ 루프 뒤에 차원 수 대조가 있음.
func TestA066RevalidateDimensionGuardsAreLayered(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "risk_bucket_issuance.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "RevalidateQFinalAdmission" {
			body = fn.Body
		}
	}
	if body == nil {
		t.Fatal("RevalidateQFinalAdmission not found")
	}
	digestAt, loopAt := -1, -1
	for i, stmt := range body.List {
		if digestAt < 0 && callsOnStraightPath(stmt, "verifyRiskBucketStateDigest") {
			digestAt = i
		}
		rng, ok := stmt.(*ast.RangeStmt)
		if !ok || !callsAnywhere(rng.X, "RequiredDimensionOrder") && !selectorCall(rng.X, "RequiredDimensionOrder") {
			continue
		}
		// 루프는 모든 차원을 걸어야 함: 첫 문장이 누락-거절 if 이고 본문에 분기문(continue 등)이 없어야 함.
		if hasBranchStmtJ(rng.Body) || len(rng.Body.List) == 0 {
			continue
		}
		for _, inner := range rng.Body.List[:1] {
			if ifs, ok := inner.(*ast.IfStmt); ok {
				if un, ok := ifs.Cond.(*ast.UnaryExpr); ok && un.Op == token.NOT {
					if idx, ok := un.X.(*ast.IndexExpr); ok {
						if id, ok := idx.X.(*ast.Ident); ok && id.Name == "seen" && len(ifs.Body.List) > 0 {
							if _, ok := ifs.Body.List[0].(*ast.ReturnStmt); ok {
								loopAt = i
							}
						}
					}
				}
			}
		}
	}
	if digestAt < 0 {
		t.Fatal("① verifyRiskBucketStateDigest is not on the straight path of RevalidateQFinalAdmission")
	}
	if loopAt < 0 {
		t.Fatal("③ the RequiredDimensionOrder backstop loop no longer refuses a missing dimension")
	}
	if digestAt >= loopAt {
		t.Fatalf("② the state digest check (stmt %d) must precede the dimension backstop loop (stmt %d)", digestAt, loopAt)
	}
	// ④ 셋째 층: 루프 뒤의 `len(seen) != len(…)` 차원 수 대조(B17). 셋을 모두 꺼야 행동 시험이 빨개짐(R4) — 둘까지는 남은
	// 하나가 막음(R3b 생존). 그래서 셋째도 자리로 못 박음.
	countAt := -1
	for i, stmt := range body.List {
		if ifs, ok := stmt.(*ast.IfStmt); ok && i > loopAt {
			if bin, ok := ifs.Cond.(*ast.BinaryExpr); ok && bin.Op == token.NEQ {
				if call, ok := bin.X.(*ast.CallExpr); ok && len(call.Args) == 1 {
					if id, ok := call.Args[0].(*ast.Ident); ok && id.Name == "seen" {
						countAt = i
					}
				}
			}
		}
	}
	if countAt < 0 {
		t.Fatal("④ the dimension-count check after the backstop loop is gone")
	}
}

func selectorCall(e ast.Expr, name string) bool {
	call, ok := e.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == name
}

// hasBranchStmtJ 는 블록 안에 continue·break·goto 가 있는지 봄(중첩 함수 리터럴 제외).
func hasBranchStmtJ(block *ast.BlockStmt) bool {
	found := false
	ast.Inspect(block, func(n ast.Node) bool {
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		if _, ok := n.(*ast.BranchStmt); ok {
			found = true
		}
		return !found
	})
	return found
}
