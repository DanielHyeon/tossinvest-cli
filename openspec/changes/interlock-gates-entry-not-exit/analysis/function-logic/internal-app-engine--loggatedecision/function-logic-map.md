# Function Logic Map: `logGateDecision`

- Source: `internal/app/engine/interlock.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

`entry_permitted` 필드와, 검증됐지만 보호가 미배선일 때의 설명 문장이 추가됐다(design D6).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| status.Verified | true/false | runInterlock | false면 문장을 붙이지 않는다 |
| status.Protection | WIRED | UNWIRED | 동상 | WIRED면 문장을 붙이지 않는다 |

## Branches and early returns

추가된 분기는 `Verified && Protection != WIRED` 하나다. 기존 refusal 분기는 무수정.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` @ internal/app/engine/interlock.go:456 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ internal/app/engine/interlock.go:461 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| logger.Warn / logger.Event | 운영 모드 1행 | 없음 | 기존 |

## State mutations and fallbacks

- 없음 — 로그만 낸다.

## Safety conclusion

- Safe edit boundary: 필드 하나와 조건부 detail 하나의 추가. 기존 필드·레벨·이벤트명은 무수정.
- High-risk impact: no
