# Branch Test Map: `claimOwed`

**GREEN 칸은 실측해서 채운다.**
`ast.json`의 열거가 정본이다: 분기 8 · 이탈 7.

**a099는 이 함수를 편집하지 않는다.** 아래 GREEN은 전부 기존 테스트이며,
a099가 요구하는 새 테스트는 **B2 하나에 대한 회귀 핀**뿐이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:275` 상태로 갈린다 | 기존 `outbox_test.go` | no | **yes (기존)** |
| B2 | `:276` **PENDING → `true, false`** — 재무장하지 않는다 | **a099 R1** (회귀 핀) + 기존 | **no — 오늘도 통과한다. born-GREEN이다** | **yes (기존)** |
| B3 | `:279` DELIVERED / ACKNOWLEDGED | 기존 (a096) | no | **yes (기존)** |
| B4 | `:280` `remindAfter <= 0` | 기존 — `EnqueueAlert` 경로 | no | **yes (기존)** |
| B5 | `:284` 날짜를 못 읽는 정착 행 | 기존 (a096 round 2) | no | **yes (기존)** |
| B6 | `:290` 시계가 뒤로 갔다 | 기존 (a096 round 2) | no | **yes (기존)** |
| B7 | `:304` 창이 아직 안 지났다 | 기존 | no | **yes (기존)** |
| B8 | `:308` 미지의 상태 | 기존 | no | **yes (기존)** |

## ⚠ B2의 테스트는 RED가 아니다 — 회귀 핀이다

a099 R1이 이 분기를 **관측**하지만 오늘도 통과한다. `claimOwed`가 PENDING에 대해
`true, false`를 주는 것은 **옳고, a099 이후에도 옳다.**

**틀린 것은 이 함수가 아니라 그 반환을 받는 쪽이다** — `ClaimAlertForDelivery`가
`rearm=false`를 *"아무것도 하지 마라"*로 읽는 것.

이것을 RED라고 적으면 born-GREEN을 RED로 위장하는 것이고, a092 19라운드 A-P6이
정확히 그 오류를 잡았다. **그래서 여기에 회귀 핀이라고 적는다.**

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

B1·B3~B8 전부. **`not-applicable`: a099가 이 함수를 편집하지 않고, 이 일곱 분기를
근거로 아무것도 주장하지 않는다.** 근거로 쓰는 것은 B2 하나다.

## 덮이지 않은 것을 이름으로 적는다

- **`latestStamp`의 파싱 실패 경로**(`outbox.go:325`의 `continue`)는 이 함수의
  분기가 아니라 그 함수의 것이다. B5가 그 결과를 받을 뿐이다. a099는 안 건드린다.
