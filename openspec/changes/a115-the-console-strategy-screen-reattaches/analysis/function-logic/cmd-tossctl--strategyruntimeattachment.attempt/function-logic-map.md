# Function Logic Map: `strategyRuntimeAttachment.attempt`

- Source: `cmd/tossctl/httpapi_strategy_attach.go` (:180–215) · AST: `ast.json`
> 인용 전용 번들(a115 freeze 리뷰 P1-4) — a115 는 이 함수를 편집하지 않는다.

## Branches and early returns
| Branch | Condition (:line) | 효과 |
|---|---|---|
| B1 | :184 `!live` | 실패 침묵. **붙어 있는 자리 불변**(live→sentinel 격하 금지) |
| B2 | :197 `a.reader == nil && reader != nil` (B1 안) | 부재→sentinel **승격**(a109 G1), seat++ — 영구 NOT_CONFIGURED 탈출로 |
| B3 | :212 `announce` | 부착 전이 1회 로그. live 성공 시 evicted 를 잠금 밖에서 Close(G5), seat++ |

콘솔 재사용 시 함의: G2(:102–112 주석)가 「시도 대상만 깨우기」게이트를 기각했다 — live 로 보이는 죽은 자리는 failed 게이트로 안 드러난다. design 의 펌프 무조건 wake 결정(P1-2)의 근거.

## Safety conclusion
- 무편집. High-risk impact: no(주문 경로 아님).
