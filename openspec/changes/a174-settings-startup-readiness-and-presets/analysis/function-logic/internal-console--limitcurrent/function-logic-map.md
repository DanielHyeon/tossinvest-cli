# Function Logic Map: `limitCurrent`

- Source: `internal/console/settings_limits_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Guardian current table | `id="guardian-current"`부터 닫는 table까지 | daily 렌더 HTML | 부재·미닫힘이면 시험 중단 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | current table 시작점 부재 | 즉시 중단 | `t.Fatal` | 미설정값 시험 |
| B2 | table 닫힘 부재 | 즉시 중단 | `t.Fatal` | current table 경계 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.Index` | 현재값 table 경계 탐색 | 부재면 시험 중단 | AST calls |
| `t.Helper`, `t.Fatal` | 호출 지점과 계약 위반 표시 | 제품 상태 변경 없음 | AST calls |

## State mutations and fallbacks

- 입력 문자열의 부분 범위만 반환하며 저장·주문·토글을 호출하지 않음.
- card에 프리셋 숫자가 있어도 current table 검사에 섞이지 않도록 범위를 제한함.

## Safety conclusion

- Safe edit boundary: 시험용 HTML 범위 추출만 유지함.
- High-risk impact: no. 제품 경로에 도달하지 않음.
