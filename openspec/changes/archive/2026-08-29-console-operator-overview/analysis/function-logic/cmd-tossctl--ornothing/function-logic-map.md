# Function Logic Map: `orNothing`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json` (revision=base, L400–405, 분기 1개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `47672c6f` — 본문 **byte 동일**. 인접 삽입의 diff hunk 교차로 evidence가 요구되었다 (revision=base)

빈 문자열을 진단 메시지에서 `(nothing)`으로 바꾸는 테스트 헬퍼다. 이 change의 수정 대상이 아니며 인접 삽입(Guardian 한도·주문 seam 테스트 블록)의 diff hunk 교차로 evidence가 요구되었다 — 본문은 base와 byte 동일하다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s string` | 임의 | confirmer 진단 문자열 | 빈 문자열이면 `(nothing)` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `s == ""` | 없음 | `"(nothing)"` | `TestVerifyRunStillConfirmsAtTheTerminalOnly`, `TestConsoleWiresTheWebConfirmerAndRefusesPerMutationPrompts`의 실패 메시지 경로 |
| (else) | 비어 있지 않음 | 없음 | `s` | 동일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls=null |

## State mutations and fallbacks

- 순수 함수. 상태 없음.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입만 존재.
- High-risk impact: no (테스트 진단 문자열 헬퍼).
