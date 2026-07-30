# Function Logic Map: `Runner.stepConditionalModify`

- Source: `internal/verifylive/steps.go`
- Function: `internal/verifylive/steps.go:Runner.stepConditionalModify`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-holds-what-it-awaits`

조건주문 정정이 원자적인지 실측한다 — 새 식별자를 발급하고 옛 것을 무효화하는가(M19·M40, 양 시장 동일). 이 change의 편집은 끝의 표시 한 줄이고, 여기서 **후속 객체가 선행 객체의 사슬을 물려받는다.** 그것이 필요한 이유가 이 단계의 측정 결과 자체다: 옛 식별자가 즉시 404가 되므로 그 뒤로는 기록만이 둘이 한 보호였다고 말할 수 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| id string | 정정 전 조건주문 | liveConditional | 없으면 위에서 skip |
| newID string | 정정이 돌려준 새 식별자 | POST modify 응답 | 같으면 not-applicable로 관측 |
| r.chainOf(sr, ..., id) | 선행 객체의 사슬 | 기록 | 비면 후속도 사슬 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` (line 731) — `if !ok {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `if` (line 736) — `if err != nil {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B3 | `if` (line 740) — `if newTrigger <= 0 {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B4 | `if` (line 753) — `if err != nil {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B5 | `if` (line 759) — `if newID != id {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B6 | `else` (line 769) — `} else {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B7 | `if` (line 760) — `if _, err := r.readConditional(ctx, sr, id); err != nil {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B8 | `else` (line 763) — `} else {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B9 | `if` (line 773) — `if co, err := r.readConditional(ctx, sr, newID); err == nil {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.liveConditional` | ast.json calls (line 730) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `sr.skip` | ast.json calls (line 732) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.readConditional` | ast.json calls (line 735) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `OneTickFurther` | ast.json calls (line 739) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `MarketOf` | ast.json calls (line 739) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `trim` | ast.json calls (line 747) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.expireDate` | ast.json calls (line 749) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.modifyConditional` | ast.json calls (line 752) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `sr.observe` | ast.json calls (line 756) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `strconv.FormatBool` | ast.json calls (line 757) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `truncateError` | ast.json calls (line 762) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `sr.fail` | ast.json calls (line 766) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `orDash` | ast.json calls (line 774) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `sr.markHeld` | ast.json calls (line 780) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.chainOf` | ast.json calls (line 780) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 브로커의 조건주문을 정정하고(승인 목록 경유) 옛 것을 취소로, 새 것을 생성으로 기록한 뒤 표시한다.

## Safety conclusion

- Safe edit boundary: 표시 한 줄. 정정 요청·발동가 계산·옛 식별자 무효화 관측·fail 경로는 전부 무변경.
- High-risk impact: yes — 라이브 정정 요청을 보낸다. 이 change는 그 요청을 건드리지 않는다.
