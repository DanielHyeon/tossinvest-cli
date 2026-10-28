# Function Logic Map: `runVerifyReconcile`

- Source: `cmd/tossctl/verify_reconcile.go` (82-131)
- Qualified function: `runVerifyReconcile`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 4008a00b6a32f2da…)
- AST evidence: `ast.json` — AST branches 7, 반환 6(줄 91, 95, 100, 106, 122, 130), 호출 이름 15
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: 명령 흐름. a164 편집: `Expired: opts.expired` 전달, 결과 문구 라벨(B7).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts *verifyReconcileOptions` | market · record · (a164) expired | cobra | — |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 줄 84: `if ctx == nil {` | 없음 | `ctx = context.Background()` | 미덮임(시험별 coverage 측정 0) — cobra 가 늘 ctx 를 줌 |
| B2 | if 줄 90: `if err != nil {` | 없음 | `return err` | `TestReconcileRefusesPreflightBeforeAnyLockOrRead` |
| B3 | if 줄 94: `if err != nil {` | 없음 | `return reconcileRefusal(verifylive.RefuseExecutionLock, err.Error())` | `TestReconcileRefusesWhenTheExecutionLockIsHeld` |
| B4 | if 줄 99: `if err != nil {` | 없음 | `return reconcileRefusal(verifylive.RefuseRateBudget, err.Error())` | `TestReconcileRefusesWhenTheRateBudgetLeaseIsHeld` |
| B5 | if 줄 105: `if err != nil {` | 없음 | `return err` | 미덮임(시험별 coverage 측정 0) — 리더 생성 거절 — 생성자 시험(`TestReconcileNarrowReaderRefusesAccountShapes`)이 함수 층에서 잼 |
| B6 | if 줄 121: `if err != nil {` | 결과 문구 출력 없음(거절) | `return err` | `TestReconcileDefaultApprovalStopsTheCommandBeforeAnyListRead` |
| B7 | if 줄 125: `if res.Artifact.ReconciledExpired {` | 만료 라벨 | `ending = "reconciled expired"` | `TestExpiryReconcileCommandAppendsWithOnlyGetReadsAndLeavesProofsUnchanged` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `reconcilePreflight`·`acquireVerifyExecutionLock`·`acquireVerifyRateBudget` | 사전 검사·배제·lease | 거절 | AST |
| `verifyReconcileReaderFactory` | 좁은 리더 | 거절 | AST |
| `verifylive.Reconcile` | 판정 | 거절 | AST |

AST 호출 이름 전체: `acquireVerifyExecutionLock`, `acquireVerifyRateBudget`, `cmd.Context`, `cmd.OutOrStdout`, `context.Background`, `err.Error`, `fmt.Fprintf`, `lease.Release`, `lock.Release`, `reconcilePreflight`, `reconcileRefusal`, `verifyReconcileReaderFactory`, `verifyReconcileTerminal`, `verifylive.NormalizeMarket`, `verifylive.Reconcile`.

## State mutations and fallbacks

- 잠금·lease 를 defer 로 해제.

## Safety conclusion

- Safe edit boundary: 순서 불변, Params 필드 하나 추가.
- High-risk impact: 인접.

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 6 에서 편집 뒤 7 로. `source_sha256` 82787649cc41… → 4008a00b6a32….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: equal 옛 B1..B6 = 새 B1..B6 · insert 옛 B7..B6(0) → 새 B7..B7(1).
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
