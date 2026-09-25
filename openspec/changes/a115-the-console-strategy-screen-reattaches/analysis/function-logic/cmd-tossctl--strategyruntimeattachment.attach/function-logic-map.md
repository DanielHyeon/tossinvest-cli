# Function Logic Map: `strategyRuntimeAttachment.attach`

- Source: `cmd/tossctl/httpapi_strategy_attach.go`
- AST evidence: `ast.json` (base `8688f74f` = HEAD — 이 파일은 a115 가 편집하지 않는다, :158–163, 분기 0)
- Risk scan: `risk-pattern-report.md`

> **인용 전용 번들**(a115 freeze 리뷰 P1-4) — a115 는 이 함수를 편집하지 않는다. design D1 의 재사용·펌프 논거가 이 분기들에 기댄다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reader`·`live` | 부팅 해석의 세 모양(nil·sentinel·client)과 live | resolve 결과 | 없음 — 무엇이든 출발점 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (분기 없음) | 본문 전체 (:158–163) | 부팅 해석 결과를 자리에 앉힌다(attached=live, failed=!live, seat++) | — | `TestTheConsoleBootKeepsAnUnreachableEndpointUnreachable` · `TestTheConsoleStrategyScreenRecoversWhenTheEngineStartsLater` · `TestTheDaemonAttachesWhenTheEngineComesUpLater` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a.mu.Lock`/`Unlock` | 자리 잠금 | 비차단(짧은 임계구역) | AST |

## State mutations and fallbacks

- 잠금 아래 reader·attached(=live)·failed(=!live) 대입 + seat++. a115 함의: 부팅 1회 해석의 세 값 전부가 유효한 출발점 — design D1 「부팅 해석을 그대로 받는다」의 근거. 부팅은 `lastTry` 를 찍지 않는다(첫 wake 가 막히지 않음).

## Safety conclusion

- Safe edit boundary: 편집 없음(인용 전용).
- High-risk impact: no — 조회 전용 projection 재부착, 주문 경로 아님.
