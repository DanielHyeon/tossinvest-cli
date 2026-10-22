# a165 CodeGraph baseline (0.3)

- 도구: codegraph **1.6.0**, 색인은 이 worktree(`/tmp/tossos-a165-20261023`)에서 `make sdd-sync` 가 새로 만든 것(2,981 files · 51,899 nodes · 192,152 edges, `raw/codegraph-status.txt`).
- 질의 시각 HEAD: `5975af2b9` (base `34e12967e` 뒤 문서 커밋 4개 — `.go` 변경 0, `git diff --name-only 34e12967e 5975af2b9 | grep .go$` 0줄).
  HEAD 가 base 와 다른 이유는 issues.md 「sdd-sync 가 worktree 를 fast-forward 함」.
- 원문: `raw/<symbol>.txt` (callers · callees · impact). 아래는 결론만.

| 심볼 | callers (생산) | 결론 |
|---|---|---|
| `execgw.Gateway.submit` | `Gateway.place`·`Cancel`·`Amend` | 주문 mutation 의 유일한 브로커 문. 재시도는 여기 한 곳에 둔다 |
| `execgw.Gateway.place` | `Gateway.Place`(:358), `PlaceClaimedStrategy`(strategy_gateway.go:64) | Place 생산 호출자: exitloop.go:1539 · tracer.go:442 · strategydispatch/adapters.go:112 · flatten/liquidate.go:506 (proposal 표, CodeGraph `Place` callers 로 재확인 — `raw/place.txt`) |
| `execgw.Gateway.Cancel` | exitloop.go:1686 · flatten/flatten.go:382 (grep), 시험 | 둘 다 **노출 감소 경로**(청산·flatten) |
| `execgw.Gateway.Amend` | CodeGraph 는 `ops.modifyHandler`(write_operations.go:217)를 적지만 그 함수는 `d.Trading.Amend`(:241, trading.Service)를 부른다 — 이름 충돌(1.3 S11 정정). 엔진 생산 호출자 0 | 409 재시도는 Amend 에 적용해도 엔진 생산 경로에선 닿지 않음 |
| `execgw.classifyMutation` | `Gateway.submit` 1곳 | 분류는 한 자리 |
| `official.Client.send` | `deleteAcct`·`getWithHeaders`·`postWithHeaders` | 401 재전송은 모든 verb 공유 |
| `execgw.Gateway.withStrategyEntryGateAuthority` | `Gateway.submit` 1곳 | strategy first-leg 의 브로커 호출은 EntryGate 잠금 안 |
| `journal.ClassifyHTTPMutation` | `classifyMutation`, 시험 | — |
| `official.Client.PlaceOrder` | engine `officialBroker.PlacePendingOrder`(broker.go:24, CodeGraph 미검출 — grep), hybrid·verifylive·ops | 엔진 경로는 official 단일 |

`codegraph affected` 의 Go 시험 판별 결함(하네스 주석)에 따라 영향 시험은 3.3 에서 `--filter '*_test.go'` 로 잰다(편집 전이라 0.3 에선 미실행).
