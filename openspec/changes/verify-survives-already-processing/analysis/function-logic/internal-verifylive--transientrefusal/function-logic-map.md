# Function Logic Map: `transientRefusal`

- Source: `internal/verifylive/mutate.go` (revision: current)
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
| B1 | `if !errors.As(err, &apiErr) \|\| apiErr.Code != http.StatusConflict {` (internal/verifylive/mutate.go:373, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B2 | `if jsonErr := json.Unmarshal([]byte(apiErr.Body), &body); jsonErr != nil {` (internal/verifylive/mutate.go:384, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B3 | `if body.Error.Code != "already-processing" {` (internal/verifylive/mutate.go:387, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B4 | `if wait <= 0 {` (internal/verifylive/mutate.go:391, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B5 | `if wait > TransientRetryMaxWait {` (internal/verifylive/mutate.go:394, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 | 일시적 거절(409 already-processing) 재시도를 정정까지 넓혔다. 판정 함수는 하나이며 조건은 종전과 동일하게 좁다 — HTTP 상태와 본문 오류 코드가 모두 일치할 때만 참이다. | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- 일시적 거절(409 already-processing) 재시도를 정정까지 넓혔다. 판정 함수는 하나이며 조건은 종전과 동일하게 좁다 — HTTP 상태와 본문 오류 코드가 모두 일치할 때만 참이다.
- 승인 게이트·계획 인가(Plan.Authorises)·노출 상한·1주 규칙은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: 접수(placeOrder)와 조건주문 생성(createConditional)에는 절대 붙지 않는다 — 그 둘은 반복하면 두 번째 라이브 객체가 생긴다. 취소·정정은 생성 연산이 아니다. 재시도 루프는 gate와 같은 함수 안에 있어야 한다(정적 가드).
- High-risk impact: yes — 실계좌에 나가는 요청의 목록·순서를 결정한다.
