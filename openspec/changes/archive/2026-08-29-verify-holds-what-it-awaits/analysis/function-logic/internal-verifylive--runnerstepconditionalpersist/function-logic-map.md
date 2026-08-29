# Function Logic Map: `Runner.stepConditionalPersist`

- Source: `internal/verifylive/steps.go`
- Function: `internal/verifylive/steps.go:Runner.stepConditionalPersist`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-holds-what-it-awaits`

등록한 프로세스가 죽은 뒤에도 조건주문이 남아 있는지 본다(M18·M39 — 2c의 가장 중요한 전제). 이 change의 편집은 표시 한 줄이고, **여기서는 그 호출이 오늘 no-op이다** — 이 단계는 객체를 되읽기만 하고 artifact를 기록하지 않는다(실기록 `run-OJRFYBGI4UOBM4MD`의 artifacts가 없다). 그래서 붙잡음은 여전히 등록 줄이 선언한 것이 유효하고, 이 change가 그 사실을 바꾸지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| id string | 되읽을 조건주문 | liveConditional | 없으면 위에서 skip |
| registrar | 등록한 프로세스의 instance id | 기록 | 같으면 awaiting-restart |
| r.chainOf(...) | 기존 사슬 | 기록 | 표시가 no-op이라 결과에 안 실린다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` (line 686) — `if !ok {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `if` (line 691) — `if found && registrar == r.process.InstanceID {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B3 | `if` (line 704) — `if err != nil {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.liveConditional` | ast.json calls (line 685) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `sr.skip` | ast.json calls (line 687) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.registeringProcess` | ast.json calls (line 690) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `sr.awaitRestart` | ast.json calls (line 692) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `fmt.Sprintf` | ast.json calls (line 692) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `fmt.Fprintf` | ast.json calls (line 697) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.readConditional` | ast.json calls (line 703) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `sr.observe` | ast.json calls (line 705) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `truncateError` | ast.json calls (line 705) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `sr.fail` | ast.json calls (line 706) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `shortID` | ast.json calls (line 711) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `orDash` | ast.json calls (line 711) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `sr.markHeld` | ast.json calls (line 713) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.chainOf` | ast.json calls (line 713) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 없음 — 읽고 관측만 한다. 표시 호출은 기록된 artifact가 없어 아무것도 바꾸지 않는다.

## Safety conclusion

- Safe edit boundary: 표시 한 줄. 존속 판정·프로세스 경계 비교는 무변경.
- High-risk impact: no — 조회 전용 단계다.
