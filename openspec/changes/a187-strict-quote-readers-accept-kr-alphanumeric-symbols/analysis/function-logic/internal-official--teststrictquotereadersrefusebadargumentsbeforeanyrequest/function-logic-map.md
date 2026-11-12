# Function Logic Map: `TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest`

- Source: `internal/official/strict_quote_reads_test.go`(73~108행, base 6dbb752f6)
- AST evidence: `ast.json`(분기 8)
- Risk scan: `risk-pattern-report.md`
- 작성 시점: 편집 전(tasks 1.2). 시험 함수지만 본문(사례 표)을 고치므로 gate ⑤ 대상임.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 사례 표 | 두 호가 리더가 요청 전에 거절해야 하는 (market, symbol) | 시험 표 | 리더가 받아들이거나 요청을 보내면 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (75행) | `range` 사례 표 | 하위 시험 생성 | — | 이 함수 자신 |
| B2 (93행) | `StrictOrderbookTop` 오류 nil | — | `t.Fatal` | 자신 |
| B3 (95행) | else | — | — | 자신 |
| B4 (95행) | 기대 사유가 있고 다름 | — | `t.Fatalf` | 자신 |
| B5 (98행) | `StrictLastPrice` 오류 nil | — | `t.Fatal` | 자신 |
| B6 (100행) | else | — | — | 자신 |
| B7 (100행) | 기대 사유가 있고 다름 | — | `t.Fatalf` | 자신 |
| B8 (103행) | 요청 수 ≠ 0 | — | `t.Fatalf` | 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `harness.client.StrictOrderbookTop`·`StrictLastPrice` | 거절 확인 | 오류여야 함 | AST |
| `strictQuoteRefusal` | `StrictQuoteError` 형 확인 | 아니면 실패 | AST |
| `harness.hitCount` | 요청 0회 확인 | — | AST |

## State mutations and fallbacks

- 없음(시험).

## 편집 계획

- 사례 표에서 `{"kr symbol with letters", "KR", "00593A", ""}`·`{"kr alphanumeric minute-only symbol", "KR", "0101N0", StrictQuoteReasonSymbol}` 두 줄을 빼고, 같은 두 입력을 수락 시험(`a187_kr_alphanumeric_test.go`)으로 옮겨 뜻을 뒤집어 고정함(요청 1회·정확한 query·같은 종목 반환). 분기 구조는 바뀌지 않음.
- 대신 이 표에 남는 KR 거절 사례(`00593`)는 그대로 두고, 영숫자 문법 밖의 KR 거절 모양(소문자·공백·길이 5·7·구두점·비 ASCII)은 새 시험이 두 호가 리더에 대해 사유까지 고정함.

## Safety conclusion

- Safe edit boundary: 사례 표 두 줄. 분기·단언 코드는 그대로.
- High-risk impact: 없음(시험).
