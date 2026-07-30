# Function Logic Map: `Plan.Authorises`

- Source: `internal/verifylive/plan.go`
- Function: `internal/verifylive/plan.go:Plan.Authorises`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-plans-the-object-it-mutates`

전송 직전의 인가 표면 전체. mutate.go는 이 함수가 false를 준 것을 보내지 않는다. 이 change는 종목 비교에서 '계획 줄에 종목이 없으면 무엇이든 통과' 와일드카드를 없앤다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| step / kind | 단계와 요청 종류 | 호출부(mutate.go) | 불일치는 다음 줄로 continue |
| symbol | 실제 요청의 종목 | 단계 본문 | **계획 줄과 정확히 같아야 한다 — 빈 값 포함** |
| side | 매수/매도 | 단계 본문 | 계획 줄이 비어 있으면 검사하지 않는다(무변경) |
| quantity | 주 수 | 단계 본문 | MaxQuantity 상한 + 1e-9 허용오차(무변경) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 394) — `for _, m := range p.Mutations {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `if` (line 395) — `if m.Step != step \|\| m.Kind != kind {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B3 | `if` (line 403) — `if !strings.EqualFold(strings.TrimSpace(m.Symbol), strings.TrimSpace(symbol)) {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B4 | `if` (line 406) — `if m.Side != "" && !strings.EqualFold(m.Side, strings.TrimSpace(side)) {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B5 | `if` (line 409) — `if m.MaxQuantity <= 0 {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B6 | `if` (line 412) — `if quantity > 0 {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B7 | `if` (line 417) — `if quantity <= m.MaxQuantity+quantityTolerance {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.EqualFold` | ast.json calls (line 403) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `strings.TrimSpace` | ast.json calls (line 403) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 없음 — 순수 판정 함수.

## Safety conclusion

- Safe edit boundary: 종목 비교만 좁아진다. kind·side·quantity 규칙은 무변경이고, 넓어지는 분기는 없다.
- High-risk impact: yes — 라이브 요청이 전송되는지를 최종 결정한다. 이 change의 변경 방향은 순수한 축소다.
