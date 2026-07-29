# Function Logic Map: `Runner.Plan`

- Source: `internal/verifylive/plan.go`
- Function: `internal/verifylive/plan.go:Runner.Plan`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-plans-the-object-it-mutates`

실행이 보낼 수 있는 모든 것을 계산한다. 이 change는 여기에 분기 하나를 더한다 — 대상 종목을 이름할 수 없는 mutating 단계는 목록에 올리지 않고 사유와 함께 제외한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Steps() | 카탈로그 전체 | verifylive.go | settled·preflight·unknown 각각 제외 사유가 다르다 |
| r.mutationSymbol(step) | 이 단계의 대상 종목 | runner.go | **빈 값이면 이 change가 추가한 분기가 단계를 제외한다** |
| sellable / sellableKnown | 매도가능 수량 | 계좌 1회 읽기 | 모르면 해당 단계 제외(기존) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 542) — `for _, line := range r.planCleanup() {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `range` (line 547) — `for _, step := range Steps() {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B3 | `if` (line 548) — `if settled, verdict := r.settled(step.ID); settled {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B4 | `if` (line 549) — `if step.Mutates {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B5 | `if` (line 558) — `if reason, skip := r.preflightStatic(step, passed); skip {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B6 | `if` (line 559) — `if step.Mutates {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B7 | `if` (line 565) — `if !step.Mutates {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B8 | `if` (line 570) — `if strings.TrimSpace(symbol) == "" {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B9 | `if` (line 585) — `if !ok {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B10 | `range` (line 594) — `for _, line := range lines {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `maskedAccount` | ast.json calls (line 533) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.planSellable` | ast.json calls (line 534) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `Passed` | ast.json calls (line 537) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.planCleanup` | ast.json calls (line 542) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `append` | ast.json calls (line 545) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `Steps` | ast.json calls (line 547) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.settled` | ast.json calls (line 548) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `fmt.Sprintf` | ast.json calls (line 552) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.preflightStatic` | ast.json calls (line 558) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.mutationSymbol` | ast.json calls (line 569) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `strings.TrimSpace` | ast.json calls (line 570) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.planStep` | ast.json calls (line 584) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 지역 `plan` 값과 `willRun` 맵만 쓴다. 계좌 읽기 1건(planSellable) 외에 호출 없음.

## Safety conclusion

- Safe edit boundary: 추가된 분기는 기존 `planUnknown` 제외와 같은 자리·같은 방향이다. 기존 제외 경로들은 무변경.
- High-risk impact: yes — 사람이 승인하는 목록 자체를 만든다. 목록이 불완전하면 승인이 승인이 아니게 된다.
