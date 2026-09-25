# Function Logic Map: `strategyRuntimeAttachment.observe`

- Source: `cmd/tossctl/httpapi_strategy_attach.go` (:274–312) · AST: `ast.json`
> 인용 전용 번들(a115 freeze 리뷰 P1-4) — a115 는 이 함수를 편집하지 않는다.

## Branches and early returns
| Branch | Condition (:line) | 효과 |
|---|---|---|
| B1 | :275 `err != nil && requestCancelled(ctx, err)` | 판정 없음(false) — 요청자 취소는 endpoint 소식이 아니다(F1) |
| B2 | :280 `seat != a.seat` | 옛 자리 소식 무시(F2) — 회복 직후 탈착 방지 |
| B3 | :285 `err == nil` | failed=false·attached 복원(G3)·필요 시 부착 전이 로그 |
| B4 | :296 `announce` (B3 안) | reportAttached 1회 |
| B5 | :306 `announce` | 탈착 전이 로그 — 문구가 「데몬은 그대로 돈다」(issues R1) |

**모든 답한 오류(코드 붙은 rpcError 포함)가 탈착이다** — 엔진 Validate 실패의 503, 콘솔 쪽 decode 거절도 탈착으로 읽힌다. a115 로 콘솔이 물려받는 깜빡임 병(리뷰 P2-1, issues R4)의 근거.

## Safety conclusion
- 무편집. High-risk impact: no.
