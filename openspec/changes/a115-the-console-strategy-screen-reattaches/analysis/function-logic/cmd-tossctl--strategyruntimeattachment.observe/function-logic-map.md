# Function Logic Map: `strategyRuntimeAttachment.observe`

- Source: `cmd/tossctl/httpapi_strategy_attach.go`
- AST evidence: `ast.json` (base `8688f74f` = HEAD — 이 파일은 a115 가 편집하지 않는다, :274–312, 분기 5)
- Risk scan: `risk-pattern-report.md`

> **인용 전용 번들**(a115 freeze 리뷰 P1-4) — a115 는 이 함수를 편집하지 않는다. design D1 의 재사용·펌프 논거가 이 분기들에 기댄다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx`·`seat`·`err` | 읽기의 ctx·그 읽기가 쓴 자리 세대·결과 | `Read` | 판정 반환(재부착 대상인가) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil && requestCancelled(ctx, err) {` (:275) | 요청자 취소 — 판정 없음(F1) | — | `TestACancelledRequestDoesNotDetachAHealthyClient` |
| B2 | `if seat != a.seat {` (:280) | 옛 자리의 소식 무시(F2) — 회복 직후 탈착 방지 | — | `TestALateReadFailureDoesNotUnseatTheNewAttachment` |
| B3 | `if err == nil {` (:285) | 성공 — failed=false·attached 복원(G3) | — | `TestARecoveredReadIsAnAttachmentAgain` · `TestTheAttachmentReportsOnlyTransitions` |
| B4 | `if announce {` (:296) | (B3 안) 복원이 전이면 부착 로그 1회 | — | `TestTheAttachmentReportsOnlyTransitions` |
| B5 | `if announce {` (:306) | 실패 전이면 탈착 로그 1회 — 문구 「데몬은 그대로 돈다」(issues R1) | — | `TestTheAttachmentReportsOnlyTransitions` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `requestCancelled` | 취소 판정 | ctx 가 실제로 끝났을 때만 | AST |
| `a.reportAttached`/`a.report` | 전이 로그 | 전이 시 1회 | AST |

## State mutations and fallbacks

- failed·attached 갱신. **모든 답한 오류(코드 붙은 rpcError 포함)가 탈착이다** — 엔진 Validate 실패의 503, decode 거절도 탈착으로 읽힌다. a115 로 콘솔이 물려받는 깜빡임 병(freeze 리뷰 P2-1, issues R4)의 근거.

## Safety conclusion

- Safe edit boundary: 편집 없음(인용 전용).
- High-risk impact: no — 조회 전용 projection 재부착, 주문 경로 아님.
