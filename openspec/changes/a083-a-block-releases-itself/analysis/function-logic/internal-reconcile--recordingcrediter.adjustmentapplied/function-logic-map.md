# Function Logic Map: `recordingCrediter.AdjustmentApplied`

- Source: `internal/reconcile/converge_test.go` (lines 26–29)
- AST evidence: `ast.json` (`source_sha256: ca763ad418c5a277dda72275cc8f35f5f545243d04ed1163c323f1a940591264`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 테스트 코드. 주문·손절·원장 판정 경로가 아니다.

## What it does

`AdjustmentCrediter`의 테스트 대역. 전달된 심볼과 비교 as-of를 순서대로 기록한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 fixture | 해당 테스트가 구성한 값 | 테스트 본문 | 단언 실패로 드러난다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — 단일 경로 | 아래 mutations 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 27, 'column': 15}, 'text': 'append'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 27, 'column': 34}, 'text': 'append'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 27, 'column': 41}}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 28, 'column': 18}, 'text': 'append'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 테스트 로컬 상태만. 프로덕션 상태를 변경하지 않는다.

## Safety conclusion

- Safe edit boundary: a083이 바꾼 것은 credit·diff에 비교 as-of를 붙이고 관측 사이에 시계를 진행시킨 것뿐이다. 어떤 단언도 완화하지 않았다.
- High-risk impact: no — 테스트 함수다. 다만 이 테스트들이 검증하는 대상은 High-risk 경로이므로, 단언을 약화하지 않았다는 것이 이 map의 요점이다.
