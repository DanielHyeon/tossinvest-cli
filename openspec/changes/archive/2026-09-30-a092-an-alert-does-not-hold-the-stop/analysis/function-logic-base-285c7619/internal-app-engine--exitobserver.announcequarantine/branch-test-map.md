# Branch Test Map: `ExitObserver.announceQuarantine`

a092는 이 함수를 편집하지 않는다. 표는 **인용한 분기가 실재함**을 AST로 고정한다.

| Branch | 위치 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:61` | 맵이 없으면 지연 생성한다 — `NewExitObserver`를 건드리지 않으려는 의도 | 기존 커버리지 | n/a | n/a |
| B2 | `:67` | 같은 세대·버전의 격리는 두 번 알리지 않는다 — **래치** | 기존 커버리지 | n/a | n/a |

## 이 표가 정정하는 3판의 누락

`notify-reach.md`는 이 자리를 이미 열거했다. 못 쓴 쪽은 `delivery-latency.md`다.

| 3판 `delivery-latency.md` | 4판 |
|---|---|
| publish 유발 이벤트 14종 | **16종** — `exit.snapshot_quarantined`와 `flatten.complete` 포함. 그중 프로덕션 publish는 **12종**(flatten 4종은 `Notifier` nil) |
| 판별자 3개 | **4개** — snapshot_quarantined는 레벨(WARN=Notify / INFO=로그 전용)로 가른다 |
