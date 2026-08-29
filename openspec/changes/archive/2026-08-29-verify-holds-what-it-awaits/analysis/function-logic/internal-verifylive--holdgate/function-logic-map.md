# Function Logic Map: `holdGate`

- Source: `internal/verifylive/cleanup.go`
- Function: `internal/verifylive/cleanup.go:holdGate`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-holds-what-it-awaits`

객체 하나가 **누구의 판정을 기다리는지** 답한다. 신규 함수이지만 새 정책이 아니라 이 파일이 이미 갖고 있던 정책을 **글로 적은 것**이다. 두 기본값이 그 정책이다 — 조건주문은 `conditional-cancel`이 gate였고(그래서 구 기록의 `Deliberate`는 정확히 그 뜻이다), 주문은 gate가 없었다(이 도구가 남긴 주문은 전부 치워야 할 실수였으니까). 발동 측정이 뒤집는 것은 두 번째다: 조건주문이 발동해 만든 child 매도는 **체결되어야** 하고, '이번 실행이 취소하지 않았다'가 누수가 아니라 측정 내용이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| a.HeldUntil | 이 객체를 놓아줄 단계 | artifact를 쓴 줄 | 비면 kind 기본값으로 떨어진다 |
| a.Kind | order 또는 conditional-order | record.go 상수 | 그 외 kind는 gate 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` (line 156) — `if a.HeldUntil != "" {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `if` (line 159) — `if a.Kind == KindConditional {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 없음 | 호출 없음 | — | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 없음 — 순수 함수. 입력을 읽고 StepID 하나를 돌려준다.

## Safety conclusion

- Safe edit boundary: 신규 leaf. 기존 판정을 기본값으로 재현하는 것이 유일한 계약이며 legacy 표 5종이 그것을 고정한다.
- High-risk impact: yes — 여기서 빈 값을 돌려주면 그 객체는 무조건 취소 대상이 된다.
