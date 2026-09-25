# Function Logic Map: `strategyRuntimeAttachment.wake`

- Source: `cmd/tossctl/httpapi_strategy_attach.go` (:166–177) · AST: `ast.json`
> 인용 전용 번들(a115 freeze 리뷰 P1-4) — a115 는 이 함수를 편집하지 않는다.

## Branches and early returns
| Branch | Condition (:line) | 효과 |
|---|---|---|
| B1 | :170 `a.trying || tooSoon || a.ctx.Err() != nil` | early return — single-flight·rate limit(interval)·수명 종료가 게이트. 아니면 trying=true, lastTry=now, go attempt() |

무조건 wake 의 비용이 「간격당 시도 1회」로 고정되는 근거(:110–112 주석과 일치) — design 펌프 결정(P1-2 수용)의 근거.

## Safety conclusion
- 무편집. High-risk impact: no. 주의: interval≤0 이면 tooSoon 이 항상 false — 콘솔 펌프 쪽 가드는 별도(P2-5).
