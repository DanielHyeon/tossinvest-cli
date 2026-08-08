# Branch Test Map: `ExitObserver.applyFloor`

AST 기준 분기 6 / 이탈 7. 기존 테스트는 `internal/app/engine/exitloop_test.go`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:1404` Floor 미주입 → 무캡 | `TestNoFloorSourceCapsNothing` `:997` | no | yes |
| B2 | `:1408` 하한 계산 실패 → 0주, **logErr만** | `TestAFloorThatCannotBeComputedSellsNothing` `:982` | no | 부분 |
| B3 | `:1416` 하한 미적용 → 무캡 | 간접 | no | yes |
| B4 | `:1420` `CompareDecimal` 오류 | **없음** | no | no |
| B5 | `:1423` 하한이 충분 → 무캡 | 간접 | no | yes |
| B6 | `:1427` `SubDecimal` 오류 | **없음** | no | no |

`:1446`(캡 성립)은 분기 id가 없지만 `TestTheConfirmedFloorCapsTheLiquidation` `:927`과
`TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable` `:953`이 도달한다.

## 기존 테스트가 단언하지 않는 것

`:953`·`:982`는 "아무것도 제출되지 않고 레벨은 재발의 가능"까지만 본다.
**등급·durability·반복 계수·문구는 단언하지 않는다.** 그래서 8/2에 13회 반복되는 동안
어떤 테스트도 깨지지 않았다.

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R1 | 보호 제안 + `floor.Quantity == 0` | critical 등급, outbox 행 생성 |
| R2 | 보호 제안 + 하한 계산 실패(B2) | R1과 동일 — 원인이 달라도 결과는 "손절이 0주" |
| R3 | 보호 제안 + **부분** 캡(`floor.Quantity > 0`) | **종전 등급 유지** (일부는 나갔다) |
| R4 | 익절 제안 + `floor.Quantity == 0` | **종전 등급 유지** — 무보호 노출이 아니다 |
| R5 | 알림 문구 | 0주일 때 "일부만 나갔다"라고 말하지 않는다 |
| R6 | 제출 수량 반환값 | R1~R5 전부에서 **무변화** (§0.9) |
