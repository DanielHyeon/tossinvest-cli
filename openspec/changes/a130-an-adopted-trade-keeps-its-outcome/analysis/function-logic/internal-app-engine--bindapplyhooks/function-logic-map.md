# Function Logic Map: `bindApplyHooks`

- Source: `internal/app/engine/gateway.go` (편집 전 `187`–`202`, base `233d5d9c`)
- AST evidence: `ast.json` (편집 전 `source_sha256` 3d6dabe90a54dd3f…) — branches 1 · returns 2 · defers 0 · go 0
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(tasks 0.5).

**역할.** 엔진 기동 때 저널에 Project · Campaign · Exit 훅과 비용 모델을 한 번 묶는다.
호출자(CodeGraph 1.6.0, `analysis/code-context/raw/`): 생산 `NewContext`(engine.go:503) 1, 시험 14.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `j` | 열린 저널 | NewContext | SetApplyHooks 오류 → B1 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:188` SetApplyHooks 오류 | 없음 | 감싼 err `:199` | TestTheProjectionGuardAsksTheBindingAndNotTheData(두 번째 바인딩 거절, interlock_internal_test.go:202) |
| (fallthrough) | 성공 | — | nil `:201` | TestBindingWiresBothApplyHooks |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.SetApplyHooks` | 바인딩 | 두 번째 → 오류 | AST |

Live config binding: 없음.

## State mutations and fallbacks

- 저널의 훅 바인딩(한 번).

## a130 편집 계획 (2판)

시그니처에 선택 인자 `refused ...func(journal.TradeOutcomeRefusal)` 를 더하고(시험 14곳 그대로 컴파일), 리터럴에 `OutcomeRefused: outcomeRefused` 한 필드. 리터럴은 gateway.go 에 남음. B1 · 반환 무변경.

## Safety conclusion

- Safe edit boundary: 위 「a130 편집 계획」 만. 주문 · 손절 · 체결 반영의 동작 무변경.
- High-risk impact: yes(원장 / 엔진 기동) — 성과 산식 무변경, 거절 셋 추가(보수 방향), 거절 기록 추가.
