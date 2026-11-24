# Function Logic Map: `TestA201CurrentBehaviourAStaleStopsTheTrailingSymbolWhoseFillLandsAfterObserve` (지운 base 판)

- Source: `internal/app/engine/a201_stale_quantity_comparison_repro_test.go` (435-462), base `35d28a535`, source sha256 `8f5f0dc8e157…`.
- AST evidence: `ast.json`(revision: base, 분기 2 — base 바이트 사본에서 `go run ./tools/logic-map` 로 뽑고 `file` 을 저장소 경로로 적음).
- Risk scan: `risk-pattern-report.md`.
- 처분: 판 중단 뒤 종목 B 의 잔여 재현 — 앞 종목 A 도 고착이라 단언했으나 v3 에서 A 는 W3-a 체결 credit 으로 풀리므로 지움(tasks 2.2). B 의 고착(design §4 남는 것)은 `TestA201TheTrailingSymbolWhoseFillLandsAfterObserveStaysStuck` 가, A 의 해제는 `TestA201W3aFillAfterConvergencePositionsIsCreditedLikeAnAdjustment` 가 이어받음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 임시 장부 · 가짜 브로커 · 합성 픽스처(두 종목) | `t.TempDir()` | `newA201World` | 단언 실패면 시험 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 if@444 | 시작 주기 오류 또는 두 종목 중 미해제 행 있음 | 없음 | t.Fatalf | 시험 자신 |
| B2 if@454 | `c.Err == nil \|\| !strings.Contains(c.Err.Error(), a) \|\| c.Converged != 0` | 없음 | t.Fatalf | 시험 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 1단계 하네스(newA201World · hold · holds · stopOpenAtBroker · accountSold · next · logCycle · stopFills · settleClean · assertStuck) | 주기 돌리기 · 장부 읽기 | 오류면 t.Fatal | AST |

## State mutations and fallbacks

- 시험 함수 — 생산 상태 변경 없음(임시 장부만).

## Safety conclusion

- Safe edit boundary: 시험 전용. 생산 코드 영향 없음.
- High-risk impact: no(시험 함수 — 지운 기대는 새 계약 시험 둘이 이어받음).
