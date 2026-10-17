# Function Logic Map: `classifyMutation`

- Source: `internal/execgw/classify.go` (`21`–`99`)
- Qualified: `classifyMutation`
- AST evidence: `ast.json` (`source_sha256` 020dba811b666e94…) — **구현 로트(2026-09-30) 편집 뒤 재생성**. 편집 전 판본은 base `1ffe2295` 의 같은 파일(`808e462c…`)이며 B3~B5 가 없었다
- Risk scan: `risk-pattern-report.md`
- 분기 10 · return 7

**역할.** 브로커 호출 하나의 결과를 원장의 dispatch 분류로 바꾼다. **순서가 설계다** — 로컬 거부 → **본문 code(a094 R1)** → 브로커의 서술적 거절 → status.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `err` | 브로커/로컬 오류 | 호출자(`Gateway.submit` `gateway.go` 발송 결과) | B1 이 nil 이면 Acked |
| `result` | `domain.MutationResult` | 브로커 응답 | B1 에서 broker order id 추출 |
| `send` | 전송 진행도 | dispatch 추적기 | B8 · B10 에서 `ClassifyHTTPMutation` 에 넘어간다 |

불변식: 확정 거절(`DispatchRejected`)은 증거가 있을 때만 — B4(목록 안 code, 모순 없음) · B6(서술적 거절) · status 확정 목록(B8 안). 모르면 모호.

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/execgw/ -count=1 -covermode=set` 프로파일(2026-09-30)에서 그 줄로 시작하는 블록의 count.

| Branch | 종류 | 조건 (원문) | return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:22` `if err == nil {` | :23 Acked | 예 |
| B2 | if | `:32` `if reason, refused := policyRefusal(err); refused {` | :33 NotSent | 예 |
| B3 | switch | `:44` `switch reason, verdict := classifyRefusalCode(err); verdict {` | — | 예 |
| B4 | case | `:45` `case refusalCodeDefinitive:` | :46 Rejected(code 의 reason) | 예 |
| B5 | case | `:52` `case refusalCodeContradictory:` | :53 Ambiguous — 뒤의 분류를 타지 않음 | 예 |
| B6 | if | `:66` `if reason, refused := ClassifyBrokerRefusal(err); refused {` | :74 | 예 |
| B7 | if | `:69` `if errors.As(err, &branch) && branch.Source == trading.BranchSourcePostPrepareConfirmation {` | — (class 만 바꿈) | 아니오 |
| B8 | if | `:84` `if status, known := statusOf(err); known {` | :93 | 예 |
| B9 | if | `:87` `if outcome.Detail == "" {` | — | 아니오 |
| B10 | else | `:89` `} else {` | :98(함수 끝 갈래) | 예 |

## Calls and live bindings

`policyRefusal`(B2) · `classifyRefusalCode`(B3 — 새 파일 `refusal_code.go`, 공식 `APIError.Body` 만 읽음) · `ClassifyBrokerRefusal`(B6) · `errors.As`(B7) · `statusOf`(B8) · `journal.ClassifyHTTPMutation`(B8 안 · 끝) · `reasonForClass`. 전부 순수 함수 — 브로커 · 원장 호출 0, 오류 · 타임아웃 계약 없음.

## State mutations and fallbacks

없다(분류만). 판정 없음(`refusalCodeNone`)은 B3 의 switch 를 빠져나와 종전 경로(B6 → B8 → 끝)로 간다 — 이것이 "모르는 것을 확정으로 바꾸지 않는다" 의 fallback 이다.

## Safety conclusion

- **Safe edit boundary**: a094 R1 은 B3~B5 를 **B6(ClassifyBrokerRefusal) 앞**에 끼웠다(D−3.5). B5 가 뒤의 분류를 건너뛰는 것이 핵심 — 두 자리 code 모순은 422 여도 확정 거절이 아니다(시험 2.5f). 상태 코드 표(`isDefinitiveRejection`)는 건드리지 않았다(시험 2.6). 재생 분류(`classifyReplay`)는 이 함수도 B3 의 분류기도 부르지 않는다(구조 시험 2.11).
- **High-risk impact**: yes — 주문 분류의 최상위 진입점. 확정 거절은 attempt 를 종결시켜 종목 차단(`checkSymbolFree`)을 푼다.
