# Function Logic Map: `strategyRuntimeAttachment.Read`

- Source: `cmd/tossctl/httpapi_strategy_attach.go` (:119–137) · AST: `ast.json`
> 인용 전용 번들(a115 freeze 리뷰 P1-4) — a115 는 이 함수를 편집하지 않는다.

## Branches and early returns
| Branch | Condition (:line) | 효과 |
|---|---|---|
| B1 | :121 `wanted`(=failed) | wake — 시도 대상일 때 요청이 재부착을 깨움 |
| B2 | :124 `reader == nil` | error 반환 — 부재를 스냅샷으로 지어내지 않음(design 대안 2 기각의 근거) |
| B3 | :131 `a.observe(ctx, seat, err)` | 실패 판정 시 즉시 wake — 다음 요청을 기다리지 않음 |

## Safety conclusion
- 무편집. High-risk impact: no. 요청 경로는 dial 하지 않는다(dial 은 attempt 에만) — spec 의 SHALL NOT 근거.
