# Function Logic Map: `brokerNumber`

- Source: `internal/console/portfolio.go`
- Change: `add-candidate-discovery`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**본문 무변경** — 이 change의 base 대비 함수 본문이 byte 동일하다(base·HEAD 두 판본의 선언 범위 텍스트를 직접 비교해 확인했다). 인접 hunk 교차로 evidence가 요구됐고, `ast.json`은 base revision에서 뜬 것이다.

브로커 숫자의 렌더. **본문 무변경** — base 대비 byte 동일이며, 바로 아래에 `livePositions`가 삽입되면서 diff hunk가 교차해 evidence가 요구됐다. AST는 base revision이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `present` | bool — 브로커 응답에 이 심볼이 있었는가 | `positionRow.InBroker` | false면 `—` |
| `v` | float64 | `domain.Position`의 수량·단가·평가액 | 해당 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `!present` | 없음 | `"—"` | `TestThePositionsScreenRendersWithEitherSourceMissing`(원장에만 있는 행이 브로커 칸을 `—`로 렌더) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `decimalText` | 지수 표기·후행 0 없이 3자리 구분 | 순수 문자열 변환 | portfolio.go:721 |
| (금지 바인딩) | 계좌·원장 무접촉 | 순수 함수 | ast.json calls |

## State mutations and fallbacks

- 없음(순수 함수·본문 무변경).

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입(`livePositions`)만 존재한다.
- High-risk impact: no (렌더 전용, 산술 없음). 다만 `—`와 `0`을 가르는 지점이므로 판정이 뒤집히면 '브로커에 없는 보유'가 0으로 보인다 — 이 저장소가 반복해서 막아 온 실패 형태다.
