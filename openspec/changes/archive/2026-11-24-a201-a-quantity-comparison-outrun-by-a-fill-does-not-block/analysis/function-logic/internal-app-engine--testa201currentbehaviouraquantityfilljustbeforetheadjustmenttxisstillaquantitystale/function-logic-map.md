# Function Logic Map: `TestA201CurrentBehaviourAQuantityFillJustBeforeTheAdjustmentTxIsStillAQuantityStale` (지운 base 판)

- Source: `internal/app/engine/a201_stale_quantity_comparison_repro_test.go` (361-376), base `35d28a535`, source sha256 `8f5f0dc8e157…`.
- AST evidence: `ast.json`(revision: base, 분기 1 — base 바이트 사본에서 `go run ./tools/logic-map` 로 뽑고 `file` 을 저장소 경로로 적음).
- Risk scan: `risk-pattern-report.md`.
- 처분: 조정 tx 직전 체결의 편집 전 동작(수량 재검증 stale · 고착) 고정 — v3 계약으로 뒤집혀 지움(tasks 2.2). 뒤를 잇는 시험 `TestA201AFillJustBeforeTheAdjustmentTxIsCreditedLikeAnAdjustment`(`internal/app/engine/a201_fill_credit_cycles_test.go`).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 임시 장부 · 가짜 브로커 · 합성 픽스처 | `t.TempDir()` | `newA201World` | 단언 실패면 시험 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 if@370 | `c.Err == nil \|\| !strings.Contains(c.Err.Error(), "expected quantity 31 but found 0")` | 없음 | t.Fatalf | 시험 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 1단계 하네스(newA201World · start · next · logCycle · assertStuck · settleClean) | 주기 돌리기 · 장부 읽기 | 오류면 t.Fatal | AST |

## State mutations and fallbacks

- 시험 함수 — 생산 상태 변경 없음(임시 장부만).

## Safety conclusion

- Safe edit boundary: 시험 전용. 생산 코드 영향 없음.
- High-risk impact: no(시험 함수 — 지운 기대는 새 계약 시험이 이어받음).
