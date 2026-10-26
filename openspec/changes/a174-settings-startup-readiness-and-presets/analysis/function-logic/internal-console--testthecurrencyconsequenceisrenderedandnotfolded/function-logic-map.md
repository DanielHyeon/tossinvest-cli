# Function Logic Map: `TestTheCurrencyConsequenceIsRenderedAndNotFolded`

- Source: `internal/console/settings_card_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 현재 통화 결과 | limits 카드의 notice 1개 | `LimitCurrencyConsequence()` | 부재·접힘이면 시험 실패 |
| paired FX 설명 | 계좌 기준 통화·KR/US·FX·같은 구현 웨이브 | 현재 통화 결과 문장 | 누락 시 시험 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | consequence 속성 부재 | 즉시 중단 | `t.Fatal` | 현재 notice 존재 |
| B2 | notice class 부재 | 없음 | `t.Error` | never-folded notice |
| B3 | 필수 paired-FX 표현 순회 | 없음 | 각 표현 검사 | 현재 통화 설명 |
| B4 | 표현 부재 | 없음 | `t.Errorf` | paired readiness 유지 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `limitsHarness`, `h.page` | daily GET 렌더 생성 | 시험 fixture, 재시도 없음 | AST calls |
| `strings.Index`, `strings.Contains` | notice 위치·문구 검사 | 부재면 시험 실패 | AST calls |

## State mutations and fallbacks

- 시험은 화면 HTML만 읽고 form submit을 수행하지 않음.
- 현재 통화가 없을 때도 실제 설정값을 꾸며 표시하지 않음.

## Safety conclusion

- Safe edit boundary: current consequence의 펼쳐진 notice 계약만 검증함.
- High-risk impact: no. 한도·통화 저장값을 변경하지 않음.
