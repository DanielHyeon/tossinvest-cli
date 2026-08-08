# Branch Test Map: `Notifier.notifyCritical`

측정: `go test -covermode=set -coverprofile ./internal/obs/`.
**RED 85.4% → GREEN 86.7%.** `-covermode=set`은 횟수를 세지 않는다.
줄번호는 GREEN 칸이 구현 후 재생성한 `ast.json` 기준이다.

**a097 2판이 `B3`의 본문을 바꾼다** — 초판은 이 함수를 편집 대상에서 뺐고,
proposal-freeze 리뷰의 P1이 그것을 뒤집었다. B3은 이제 오류를 반환하기 전에
`n.escalate(ctx, e)`를 부른다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:171` journal 미배선 → best-effort 강등 | `TestOrdinaryAlertsAreBestEffort` 계열 | 진입 (`171-175 count=1`) | 진입 (`171-175 count=1`) |
| B2 | `:175` 강등 경고를 남길 logger가 있다 | 위와 같음 | 진입 (`175-179 count=1`) | 진입 (`175-179 count=1`) |
| B3 | `:195` `claimAndDeliver`가 오류를 냈다 → **승격 시도** | `TestAClaimThatFailsAttemptsTheDurableBlock` (a097 2.5) | **미진입** (`195-197 count=0`) | **진입** (`195-215 count=1`) |
| B4 | `:217` owed였는데 정착하지 못했다 → 승격 | `TestPersistentDeliveryFailureBlocksEntries` 계열 | 진입 (`199-205 count=1`) | 진입 (`217-223 count=1`) |

## B3과 B4는 다른 실패를 다룬다

`B4`는 **전송**이 owed였는데 정착하지 못한 경우다. `B3`은 **기록**이 실패해 owed 판단
자체가 없는 경우이며, `claimAndDeliver`가 `owed=false`를 돌려주므로 B4에 도달하지 않는다.

초판은 그 도달 불가를 근거로 "claim 실패에는 승격이 없고 그것이 의도"라고 적었다.
**리뷰의 P1이 그 의도를 기각했다**: `EntryGate`의 래치는 메모리에만 있고, claim이 실패하면
원장에 알림 행조차 없다. 재시작 한 번이면 차단도 증거도 사라지고 진입이 다시 열린다.

그래서 a097은 B4를 바꾸는 대신 **B3 안에 승격을 직접 넣었다**. 두 분기는 여전히 서로
독립이며, RED에서 미진입이던 B3이 GREEN에서 진입으로 바뀐 것이 그 변경의 직접 증거다.

`B4`의 값은 RED와 GREEN에서 모두 진입이다 — a097이 전송 실패 경로를 건드리지 않았다는
확인이다.

## 승격이 실패해도 부른다

`TestAClaimThatFailsAttemptsTheDurableBlock`은 **닫힌 journal**로 돌린다. 그래서 승격도
실패하고, `escalate`는 그 실패를 `EventOperatingMode` error 로그로 남긴다. 테스트가
단언하는 것이 그 로그다.

원장이 죽어 있으면 내구적 차단을 만들 수 없다는 것은 사실이고 바꿀 수 없다. 바꿀 수 있는
것은 **그 사실이 기록되는가**이며, 지금은 기록된다.
