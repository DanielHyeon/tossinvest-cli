# Branch Test Map: `EntryGate.Block`

Source: `internal/execgw/retry.go` (498-505). AST 기준 분기 1 / 이탈 0 /
defers 1 / go_statements 0.

`return` 문이 0개인 void 함수다. 관측은 전부 게이트의 후속 판정(`CheckEntry`)으로 한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:501` 처음 래치 → `latches`에 들어가고 진입이 막힌다 | `internal/execgw/entrylatch_test.go TestEveryEntryBlockingConditionReachesTheChain:35`, `TestTheLatchSurfaceNeverDisagreesWithTheGate:198` | no | yes |
| — | 이미 래치된 이유로 다시 호출 → 무동작 | **없음** — `detail`이 덮어써지지 않는지, `revision`이 한 번만 느는지 보는 테스트가 없다 | no | no |

두 번째 행은 B1의 **거짓 갈래**이고 AST에 별도 ID가 없다. 분기 ID를 지어내지 않는다.

## a092가 이 함수에 대해 지는 것

편집하지 않으므로 이 함수에 대한 새 RED는 없다.

**미테스트인 B1 거짓 갈래가 17판에서 처음으로 뜨겁게 돌기 시작한다.** HEAD에서는
`claimAndDeliver`/`deliver`가 관측당 최대 한 번 `Block`을 부르지만, 17판의 배달
루프는 **주기마다** 실패한 행에 대해 부른다. 멱등성이 깨져 있으면 `revision`이
2초마다 늘고, `revision`을 캐시 무효화에 쓰는 소비자가 매 주기 캐시를 버린다.

코드를 읽으면 멱등하다. **그러나 그것을 관측하는 테스트는 없다.**
§6.0에 이 관측을 더한다 — **R17-12**로 새로 세운다(이 산출물이 만든 항목이다).

`ReasonAlertUndelivered`의 해제 경로 부재는 §6.0 R17-11이 지고 간다.
