# Function Logic Map: `isFullExit`

- Source: `internal/app/engine/exitloop.go` (lines 1139–1142, revision `base` (base `ac2bbfcc`))
- AST evidence: `ast.json` (`source_sha256: 3cf97f20c5eafa4b8f4d57bdbc1bc9d9f639c1f425590e4f212b238e7d0d5c8c`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk**

## What it does

포지션을 통째로 닫으려는 제안인지 보고한다. 본문은 바뀌지 않았다 — 개정 2가 바로
아래에 `isProtective`를 새로 넣으면서 base 쪽 hunk가 이 함수의 줄 범위에 닿았다.
`isProtective`는 이 셋 중 손절 둘만 골라낸다: 익절(`ActionLadderTakeProfit`)은 보류해도
상단만 잃지만, 손절을 보류하면 §0.3을 깬다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `p.Action` | `exitpolicy.Action` 상수 | 평가기가 만든 제안 | 그 외 action은 false — 부분 청산은 심볼을 비울 필요가 없다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Required test |
|---|---|---|---|
| B1 | 분기 없음 | 단일 반환식 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | 호출 없음. 순수 비교 셋 | — | AST `calls` 없음 |

## State mutations and fallbacks

- 없음. 순수 함수.

## Safety conclusion

- Safe edit boundary: 세 action 집합. 여기서 action을 빼면 통짜 청산이 미결 주문을 남긴 채 제출된다.
- High-risk impact: yes — 심볼 청소 여부를 결정한다.
