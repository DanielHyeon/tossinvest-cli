# Function Logic Map: `cleanupFrom`

- Source: `internal/verifylive/cleanup.go`
- Function: `internal/verifylive/cleanup.go:cleanupFrom`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-holds-what-it-awaits`

이 change의 중심. **어떤 살아 있는 객체를 취소 대상으로 올릴지** 정하는 유일한 자리다. 여기서 나온 목록이 승인 화면의 첫 줄들이 되고, 승인되면 라이브 취소 요청이 나간다. 이전에는 kind로 갈라 두 규칙을 썼다 — 주문은 무조건 대상, 조건주문은 `conditional-cancel` 판정 기준. 이제 한 규칙이다: **줄이 지목한 gate가 그 줄보다 뒤에 판정했는가.** kind별 동작은 `holdGate`의 기본값으로 남아 기존 기록의 판정을 바이트 단위로 보존한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| entries []Entry | append-only 기록 전체 | capability-verify*.jsonl | 빈 기록이면 빈 목록 |
| settled func(StepID) bool | gate가 terminal 판정을 받았는가 | Runner.settled 또는 Settled() | false면 대상 아님(fail-closed) |
| outstandingLines(entries) | 취소되지 않은 객체와 그 줄의 index | record.go | 취소된 것은 애초에 안 나온다 |
| holdGate(a) | 이 객체를 놓아줄 단계 | artifact.HeldUntil 또는 kind 기본값 | 빈 값이면 gate 없음 = 대상 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 125) — `for _, l := range outstandingLines(entries) {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `if` (line 127) — `if gate == "" {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B3 | `if` (line 131) — `if settled(gate) && heldAfter(entries, gate, l.at) {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `outstandingLines` | ast.json calls (line 125) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `holdGate` | ast.json calls (line 126) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `append` | ast.json calls (line 128) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `settled` | ast.json calls (line 131) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `heldAfter` | ast.json calls (line 131) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 없음 — 판정만 하고 slice를 새로 만들어 돌려준다. 기록도 브로커도 건드리지 않는다.

## Safety conclusion

- Safe edit boundary: 판정 경로가 바뀌었지만 **결과는 기존 기록에서 동일하다**(hold_test.go의 legacy 표 5종). 새 동작은 `HeldUntil`이 실린 줄에서만 나타나고, 그 필드를 주문에 찍는 코드는 이 change에 없다.
- High-risk impact: yes — 이 함수의 반환이 라이브 취소 요청의 대상 목록이다. 다만 이 change의 방향은 대상을 **줄이는** 쪽이다.
