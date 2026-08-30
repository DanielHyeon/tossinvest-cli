# Function Logic Map: `Runner.sweepStep`

- Source: `internal/verifylive/cleanup.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `verify-survives-already-processing`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 이 함수의 입력 | 시그니처가 정의한 범위 | 호출자 | 범위 밖 값은 정규화되거나 거부된다 |
| 증거 기록의 Outstanding | 이 도구가 만들고 취소되지 않은 객체 | capability-verify*.jsonl | 기록이 없으면 대상도 없다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if sr.abort != nil \|\| ctx.Err() != nil {` (internal/verifylive/cleanup.go:209, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B2 | `for _, a := range Outstanding(mine) {` (internal/verifylive/cleanup.go:213, range) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B3 | `if a.Kind != KindOrder {` (internal/verifylive/cleanup.go:214, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B4 | `if err := r.cancelOrder(ctx, sr, a.ID, a.Symbol, "이 단계가 실패해 남긴 주문 — 다음 단계의 노출 상한을 비운다"); err != nil {` (internal/verifylive/cleanup.go:219, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 | 실패한 단계가 자기가 만들고 취소하지 않은 주문을 정리한다. 범위는 그 단계의 산출물뿐이고, 기록의 이전 실행 잔여물은 prologue의 몫이다. | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- 실패한 단계가 자기가 만들고 취소하지 않은 주문을 정리한다. 범위는 그 단계의 산출물뿐이고, 기록의 이전 실행 잔여물은 prologue의 몫이다.
- 승인 게이트·계획 인가(Plan.Authorises)·노출 상한·1주 규칙은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: sr.abort가 설정됐거나 컨텍스트가 죽었으면 아무것도 보내지 않는다 — 계획 밖 요청으로 멈춘 실행이 '정리니까 한 건만 더'를 보내면 그 오류의 의미를 부정한다. 조건주문은 대상이 아니다(존속 측정이 읽어야 한다). 단계의 verdict는 바꾸지 않는다.
- High-risk impact: yes — 실계좌에 나가는 요청의 목록·순서를 결정한다.
