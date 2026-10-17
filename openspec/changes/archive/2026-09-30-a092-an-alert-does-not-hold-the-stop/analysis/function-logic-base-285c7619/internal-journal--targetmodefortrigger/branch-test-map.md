# Branch Test Map: `TargetModeForTrigger`

a092는 이 함수를 편집하지 않는다. 표는 **전칭 주장의 근거가 실재함**을 AST로 고정한다.

| Branch | 위치 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:538` | 트리거 문자열을 트림해 분기한다 | 기존 커버리지 | n/a | n/a |
| B2 | `:539` | **자동 트리거 6개가 한 case에 모여 `ENTRY_BLOCKED`를 돌려준다** — 값을 돌려주는 case가 이것 하나뿐이라는 것이 전칭 주장의 근거다 | 기존 커버리지 | n/a | n/a |
| B3 | `:546` | 열거 밖 트리거는 `("", false)` — 승격이 일어나지 않는다 | 기존 커버리지 | n/a | n/a |

## 이 표가 떠받치는 문서 주장

| 주장 | 어디 | 근거 |
|---|---|---|
| "위 네 트리거의 목표는 전부 `ModeEntryBlocked`다" | `notify-reach.md` | B2가 유일한 값 반환 case |
| "`ModeTriggerExitObservationOutage`의 목표는 `ENTRY_BLOCKED`" | `proposal.md`, `checkoutage` FLM | B2의 case 목록 |
| "`HALT_ALL` 자동 진입은 없다" | 이 함수 주석 `:531-535` | `HALT_ALL`을 돌려주는 case가 **존재하지 않는다**(AST branches 3) |
