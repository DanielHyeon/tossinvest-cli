# Branch Test Map: `Journal.ClaimAlertForDelivery`

측정: `go test -covermode=set ./internal/journal/...` — RED 74.9%, GREEN 75.0%.
RED 열은 base `ec29dc72`의 `EnqueueAlert` 본체(같은 SQL·같은 트랜잭션)의 대응 분기다
(325.2s, 74.9%). 이 함수 자체는 a096이 만들었으므로 RED에 존재하지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 빈 event key `:173` | 기존 `TestEnqueueAlertRequiresAKeyAndAType` (위임 경유) | 진입 | 진입 |
| B2 | 빈 event type `:176` | 위와 같음 | 진입 | 진입 |
| B3 | `BeginTx` 실패 `:183` | 없음 — DB 핸들 실패 주입 하니스가 없다 | 미진입 | 미진입 |
| B4 | `switch` 진입 `:194` | — (자기 블록 없음) | — | — |
| B5 | **기존 행 → `claimOwed`로 판정** `:195` | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing` 외 5건 | 진입(판정 없음) | 진입 |
| B6 | **owed 행을 재무장한다** `:197` | `TestClaimingADeliveredAlertPastTheWindowIsReArmed`, `TestClaimingAnAcknowledgedAlertFollowsTheSameWindow`, `TestAnUnrecognisedAlertStateOwesDelivery` | **분기 없음** | 진입 |
| B7 | 재무장 UPDATE 실패 `:202` | 없음 — DB 실패 주입 하니스 없음 | **분기 없음** | — |
| B8 | 조회가 `ErrNoRows` 아닌 오류 `:209` | 없음 | 미진입 | 미진입 |
| B9 | `INSERT` 실패 `:217` | 없음 | 미진입 | 미진입 |
| B10 | `LastInsertId` 실패 `:221` | 없음 | 미진입 | 미진입 |
| B11 | `Commit` 실패 `:224` | 기존 테스트가 정상 커밋으로 진입 | 진입 | 진입 |

## B5·B6이 이 change다

RED에서 이 자리는 `return existing, tx.Commit()` 한 줄이었다 — id만 돌려주고 상태를 버렸다.
지금은 `claimOwed`가 판정하고, 창을 넘긴 종결 행은 **같은 트랜잭션 안에서 PENDING으로
재무장**된다. 재무장 덕분에 리마인더는 최초 전달과 완전히 같은 경로를 걷는다: 재시도 예산,
gate 잠금, 운영 모드 승격이 전부 그대로 적용된다.

`claimOwed`의 세 상태 분기는 그 함수의 BTM이 따로 덮는다. 여기서 덮는 것은
**재무장이라는 부작용**이다:

| 상태 | 창 | 기대 | 테스트 |
|---|---|---|---|
| DELIVERED | 안 | 억제, 행 그대로 | `TestClaimingADeliveredAlertInsideTheWindowOwesNothing` |
| DELIVERED | 밖 | owed, 행이 PENDING | `TestClaimingADeliveredAlertPastTheWindowIsReArmed` |
| ACKNOWLEDGED | 안/밖 | 억제 → owed | `TestClaimingAnAcknowledgedAlertFollowsTheSameWindow` |
| PENDING | — | owed, 재무장 없음 | `TestClaimingAnUndeliveredAlertStillOwesDelivery` |
| 모르는 상태 | — | owed, 행이 PENDING, 완료 표시 성공 | `TestAnUnrecognisedAlertStateOwesDelivery` |
| — | 창 0 | 영구 억제, 재무장 없음 | `TestAZeroReminderWindowNeverReArms` |

마지막 줄이 `EnqueueAlert`의 위임을 지킨다: 기록만 하는 호출자는 남의 리마인더 정책을
대신 실행하지 않는다.

## 미커버 분기에 대한 판단

B3·B7·B8·B9·B10은 모두 **DB 계층 실패 주입**이 필요하고 이 저장소에는 그 하니스가 없다.
B7은 a096이 새로 만든 분기지만 같은 종류다 — UPDATE 한 문장의 실패 경로이고, 실패하면
트랜잭션이 롤백되어 행은 손대기 전 상태로 남는다.
`not-applicable`: 실패 주입 하니스 부재.
