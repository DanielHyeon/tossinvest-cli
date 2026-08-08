# Branch Test Map: `ExitObserver.submit`

AST 기준 분기 11 / 이탈 9. 이탈 9개 중 주문이 살아나는 것은 B7 하나, 존재 미상이 B8 하나.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:1240` `applyFloor` 오류 → `return err` `:1241` | **없음** | no | no |
| B2 | `:1243` 확정 하한이 0 → 조용히 `release` `:1248` | `TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable` `:953` · `TestAFloorThatCannotBeComputedSellsNothing` `:982` | no | yes |
| B3 | `:1263` `IssueReduction` 오류 → 알림 + `release` `:1265` | 간접 | no | yes |
| B4 | `:1272` `AttachExitIntent` 오류 → `return err` `:1273` | **없음** | no | no |
| B5 | `:1277` `sellIntent` 오류 → `alertRefused` + `release` `:1279` | **없음** | no | no |
| B6 | `:1287` `switch out.State` | 다수 | no | yes |
| B7 | `:1288` `StateConfirmed` → 로그, `return nil` `:1295` | 다수 | no | yes |
| B8 | `:1296` `InDoubt`/`UnresolvedInDoubt` → **제안 유지**, `return nil` `:1300` | `TestAnInDoubtSubmissionKeepsTheProposalArmed` `:1036` | no | yes |
| B9 | `:1301` `ReasonSymbolInFlight` → **`noteDelay`** + `release(Cancelled)` `:1303` | 간접 | no | yes |
| B10 | `:1304` default(브로커 거부) → 알림 + `release(Refused)` `:1310` | `TestARefusedProposalReleasesTheLevelAndAlerts` `:1013` | no | yes |
| B11 | `:1306` `detail` 비었고 `err` 있으면 detail 보정 | 간접 | no | yes |

## B2가 2026-08-02 사건의 경로다 (실측)

`applyFloor`가 `floor.Quantity`(0)를 돌려주고(`:1446`) B2가 조용히 `release`한다.
그 직전 `applyFloor`가 `EventExitProposalCapped`를 올리지만 **`criticalEvents`에 없다** →
`severity: normal` → **outbox 행 없음 · 게이트 무반응 · 재전달 없음**.

```text
2026-08-02 23:23:25 ~ 23:26:21  pos-522745e0 (042660)
  STOP_LOSS_LADDER × 13 → 전부 PROPOSAL_REFUSED
  exit.proposal_capped × 13, severity=normal
  detail: "the RECONCILE confirmed floor authorises 0 (broker sellable quantity)"
  alert_outbox 행: 0
  종료: ADJUSTMENT_CLOSED — 손절은 끝내 나가지 않았다
```

알림 제목은 "청산이 확정 하한에 걸려 **일부만 나갔다**"인데 나간 수량은 **0**이다.
기존 테스트 `:953`·`:982`는 "아무것도 제출되지 않고 레벨은 재발의 가능"까지만 단언하고
**등급·durability·반복 계수는 단언하지 않는다.**

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R1 | 하한 0으로 보호 청산이 13주기 연속 막힘 | 반복이 원장에 남고 등급이 critical로 올라간다 |
| R2 | B8(InDoubt) | 미제출로 **계수하지 않는다** — 주문이 존재할 수 있다 |
| R3 | B1·B4·B5 각각 | 최소 1건씩 — 현재 이탈 3개가 무테스트 |
| R4 | B7(접수) | 계수·시계 초기화 |

**B5는 a087이 바꾸려는 함수(`sellIntent`)의 오류 경로다 — a087 착수 전 RED가 선행해야 한다.**
