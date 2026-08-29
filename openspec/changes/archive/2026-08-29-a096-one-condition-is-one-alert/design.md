# a096 설계 — 전송을 중복 제거한다

증거: `analysis/function-logic/` 9함수 46분기(AST), base commit `ec29dc72`.
커버리지 실측: `go test -covermode=set` — obs 84.8% → 84.7%, journal 74.9% → 75.0%.

**2판이다.** 1판은 영구 억제였고 독립 리뷰 1라운드(Codex)가 blocker 4건으로 깼다.
D3·D4·D5가 그 정정이다.

## D1. 누가 "보내도 되나"에 답하는가

`notifyCritical`이 알아야 하는 것은 **"방금 넣은 행인가"**가 아니라 **"이 전달이 아직
필요한가"**다. 전자는 삽입 여부이고 후자는 상태이며, 둘은 다르다 — PENDING인 기존 행은
삽입이 아니지만 전송은 해야 한다.

네 가지를 검토했다.

**(a) `EnqueueAlert`의 서명을 넓힌다.** `(int64, error)` → `(int64, string, error)`.
컴파일러가 모든 호출자에게 구별을 강제한다.

**(b) 별도 조회 `AlertState(ctx, id)`를 추가한다.** 서명을 안 바꾸지만 왕복이 하나 늘고,
`EnqueueAlert`의 커밋과 그 조회 사이에 TOCTOU 창이 생긴다. 5.6초 주기에서 성능은
문제가 아니지만 창은 근거 없이 만든 것이다.

**(c) `notifyCritical`이 `PendingAlerts`를 훑어 자기 key를 찾는다.** 전체 스캔이고,
outbox가 커질수록 사이클마다 비싸진다.

**(d) 새 함수 `ClaimAlertForDelivery(ctx, Alert, remindAfter) (int64, bool, error)`를 만들고
`EnqueueAlert`는 그것에 위임한다.** 본체·SQL·트랜잭션 경계는 base 그대로 옮기고,
`SELECT id`를 `SELECT id, state, delivered_at, acknowledged_at`으로 넓혀 `owed`를 함께
돌려준다. 창을 넘긴 종결 행은 **같은 트랜잭션 안에서 PENDING으로 재무장**한다(D3).

**(d)를 택했다. 처음에는 (a)를 택했고, 게이트가 그것을 되돌렸다.**

(a)로 구현하자 `check_analysis.py`가 증거 누락 8건을 보고했다:
`internal/execgw/replay.go:Gateway.parkAlert`와 `internal/journal/outbox_test.go`의 기존
테스트 7개. arity 하나 때문에 원장 패키지의 기존 테스트 전부가 "수정된 기존 함수"가 됐고,
각각이 Function Logic Map을 요구했다. 그렇게 생산됐을 28개 파일은 각 테스트의 안전성이
아니라 **인자 개수가 바뀌었다는 사실을 28번 적은 것**이다.

되돌아보니 (a)의 근거도 약했다. 컴파일러가 강제하려던 유일한 기존 호출자 `parkAlert`는
알림을 **기록만** 하고 전송 루프를 갖지 않는다(주석: "the notifier's Flush picks the row
up and delivers it"). 전달이 아직 필요한지는 그 호출자가 답할 수 있는 질문이 아니며,
강제할 것이 애초에 없었다.

(d)는 (b)의 TOCTOU도 (c)의 스캔도 없다 — 읽기가 id를 결정하는 바로 그 트랜잭션 안에 있다.
그리고 이름이 의도를 말한다: 호출자는 id를 얻는 것이 아니라 **보낼 권한을 청구**한다.

새 삽입 경로는 `true`를 돌려준다. INSERT가 `AlertPending`을 썼기 때문이고, 상수를 다시
쓰는 것이 아니라 INSERT가 쓴 값을 그대로 반영하는 것이다.

### `owed`의 정의를 왜 원장에 두는가

`owed`는 상태와 그 상태에 도달한 시각을 함께 읽어야 나오는 값이고, 재무장이라는 쓰기를
동반한다. 그 셋은 한 트랜잭션 안에 있어야 하며, 원장 밖에서는 그 경계를 만들 수 없다.

호출자에 두면 상태 집합의 해석이 두 군데가 되고 새 상태가 생기는 날 갈라진다.
1판이 정확히 그 형태였다 — 인식되지 않는 상태를 `!= AlertPending`으로 뭉뚱그려
조용히 종결로 취급했다(D3의 마지막 줄).

판정 자체는 `claimOwed`라는 **순수 함수**로 다시 한 번 분리한다. 판정과 쓰기를 나누면
판정을 표로 놓고 볼 수 있고, 1판이 놓친 두 칸은 표로 놓고 보지 않아서 놓친 것이다.

## D2. 분기를 어디에 두는가

`notifyCritical`은 RED에서도 GREEN에서도 4분기다. 분기 **수**는 같고 B4의 조건이 바뀌었다.
억제 판정과 전송은 새 함수 `claimAndDeliver`로 옮겼다 — 둘이 한 잠금 안에 있어야 하기
때문이다(D5ter).

| 분기 | RED | GREEN | 조건 | 하는 일 |
|---|---|---|---|---|
| B1 | :154 | :171 | `n.Journal == nil` | durable 저장소 없음 → best-effort로 강등하고 반환 |
| B2 | :158 | :175 | `n.Log != nil` | 그 강등을 로그로 알린다 |
| B3 | :178 | :195 | claim/전송이 오류 | 유일하게 호출자에게 올라가는 오류 |
| B4 | :182 `!n.deliver(...)` | :199 **`owed && !sent`** | 전송 실패면 escalate |

**`owed &&`가 B4의 안전 조건이다.** 보낼 필요가 없어서 안 보낸 것은 전달 실패가 아니므로
gate를 잠그지 않는다. `!sent` 하나만으로 판정하면 억제될 때마다 진입이 차단된다.

억제 분기(`!owed`)는 `claimAndDeliver` 안에 있고 `deliver` **호출 자체**를 막는다.
`deliver` 안에서 거르지 않는 이유: 그 함수는 재시도 예산 전체를 잠금 안에서 돈다
([notifier.go:303-347](../../internal/obs/notifier.go#L303-L347)). 이미 전달된 알림 때문에
그 예산 구간에 들어가는 것은, 진짜 전달이 필요한 다른 알림을 그동안 줄 세우는 것이다.

## D3. 억제는 창이지 무덤이 아니다

이 change의 유일한 실질 위험은 **보내야 할 알림을 안 보내는 것**이다. 1판은 그 위험을
"PENDING만 owed"로 막았다고 주장했고, 독립 리뷰가 두 반례를 냈다.

| 행 상태 | RED | 1판 | 2판 | 근거 |
|---|---|---|---|---|
| 새 삽입(PENDING) | 전송 | 전송 | **전송** | 바뀌지 않는다 |
| PENDING (전송 실패 뒤) | 전송 | 전송 | **전송** | 실패는 중복이 아니라 미완. a074 재시도 계약 |
| DELIVERED, 창 안 | 전송 | 억제 | **억제** | 운영자는 방금 받았다 |
| DELIVERED, 창 밖 | 전송 | **억제** | **전송(재무장)** | D4·D5 |
| ACKNOWLEDGED, 창 안 | 전송 | 억제 | **억제** | 운영자가 봤다고 직접 말했다 |
| ACKNOWLEDGED, 창 밖 | 전송 | **억제** | **전송(재무장)** | 같은 이유 |
| 모르는 상태 | 전송 | **억제** | **전송(재무장)** | CHECK 제약이 없다. 재무장하지 않으면 전달 완료 표시가 실패해 다시 폭주한다 |

마지막 줄도 리뷰가 지적했다. `alert_outbox.state`에는 CHECK 제약이 없으므로 PENDING이
아닌 **어떤** 문자열도 1판에서는 `owed=false`가 됐다. 모르는 상태는 이 빌드가 이해하지
못하는 행이고, "운영자가 받았는지 모른다"의 안전한 해석은 보내는 것이다. 이때 행도
PENDING으로 재무장해야 한다. `owed=true`만 돌려주면 발행은 성공해도
`MarkAlertDelivered`의 `state = PENDING` 조건을 통과하지 못해, 다음 관측이 다시 발행한다.

`claimOwed`가 이 표 전체다(`outbox.go:237-268`, 7분기). 순수 함수로 뽑은 이유는 판정과
쓰기를 따로 볼 수 있게 하기 위해서다 — 1판에서 이 판정은 SQL 사이에 끼어 있었고,
그래서 표로 놓고 보지 않았고, 그래서 두 칸이 틀린 것을 놓쳤다.

### 재무장이 왜 옳은 형태인가

창을 넘긴 행은 `PENDING`으로 **되돌린다**. 새 경로를 만들지 않는 것이 요점이다:
리마인더는 최초 전달과 **완전히 같은** 코드를 걷고, 따라서 재시도 예산·gate 잠금·운영 모드
승격이 자동으로 그대로 적용된다. a074 계약을 다시 검증할 필요가 없다.

`delivered_at`·`acknowledged_at`은 지우지 않는다. 이전 episode의 기록이고 감사 흔적이다.
창의 기준은 둘 중 나중 것이며 `latestStamp`가 고른다.

## D4. 영구 억제는 나중의 *다른* 발생을 삼킨다

1판의 D4는 "되살아나는 조건이 다시 알려지지 않는 것은 `event_key` UNIQUE의 기존 한계"라고
주장했다. **틀렸다.**

행 수명의 한계는 기존이었다. 그러나 base `ec29dc72`에서 UNIQUE key는 옛 행을 돌려줬고
`notifyCritical`은 **여전히 발행했다.** 운영자에게 보이는 억제는 1판이 처음 만든 것이다.

그리고 그 억제가 특히 위험한 이유는 **event key가 조건을 담고 원인을 담지 않기**
때문이다. `exit.proposal_refused`의 key는 `type|position_id|action|level`이며
(`exitloop.go:1550-1551`) 거절 사유가 없다. 주말의 `order-hours-closed` 거절이 한 번
전달되고 나면, 같은 포지션·같은 단계의 **조치가 필요한 다른 거절**이 영구히 조용해진다.

이 논증이 특히 뼈아픈 이유는, 1판이 **버그로 지목한 바로 그 혼동의 거울상**이기 때문이다.
"행에는 적용되고 전송에는 안 된다"를 결함으로 쓴 뒤, 같은 구별을 정당화에 썼다.

창은 이것을 닫는다. 다른 원인의 재발은 최대 한 창 뒤에 전달된다. key에 사유를 넣는 것은
별개 문제이며 a096 범위 밖이다(tasks 7.2).

## D5. 영구 억제는 진입 차단을 **덜** 일어나게 만든다

1판의 D5는 "escalate 경로에 닿지 않으며 진입 차단이 늘지도 줄지도 않는다"고 주장했다.
줄지 않는다는 쪽이 **거짓이었다.**

내가 쓴 반박은 *"억제 대상은 DELIVERED이고, DELIVERED가 되려면 그 행은 이미 한 번 전송에
성공했다. 성공한 전송이 escalate를 만든 적은 없다"* 였다. 이것은 **과거의 전송이
성공했다**를 **미래의 전송도 성공한다**로 바꿔치기한 것이다.

구체적 실패:

1. key K가 전달돼 행이 DELIVERED가 된다
2. 이후 알림 transport가 죽는다
3. K가 재발한다
4. 1판: `owed=false` → transport를 시험조차 하지 않는다 → 예산 소진 없음 → gate 잠김 없음
   → `ENTRY_BLOCKED` 없음. **엔진은 알림 경로가 죽은 채로 계속 거래한다.**

리뷰는 ACKNOWLEDGED 경로의 더 나쁜 판도 냈다: 운영자가 차단을 풀려고 acknowledge하면
그 행은 영구히 `owed=false`가 되어, 같은 조건이 다음 장애 중에 재발해도 차단이 다시
걸리지 않는다.

즉 **폭주가 알림 경로의 생존 감지기 노릇을 겸하고 있었다.** 폭주를 없애면서 감지기까지
없앤 것이 1판이다. 창이 감지기를 되돌려준다 — 최대 한 창 늦게, 그러나 확실히.

반대 방향(억제 때문에 escalate가 **더** 일어나는가)은 `notifyCritical` B4의 `owed &&`가
막는다. 보낼 필요가 없어서 안 보낸 것은 전달 실패가 아니므로 gate를 잠그지 않는다.

`TestADeadTransportIsStillFoundAfterASuccessfulDelivery`가 이 절 전체를 검증한다.

## D5bis. 창의 길이

1시간(`DefaultRemindAfter`)으로 정했다.

- 5.6초 주기 관측에서 한 조건이 주말 60시간 지속되면 **60건.** 폭주는 15,000건이었다.
- transport가 전달 직후 죽어도 **1시간 안에** 발견되고 gate가 잠긴다.
- 관측당 한 줄인 구조화 로그는 그대로이므로 상세는 잃지 않는다.

지수 backoff(1분→5분→1시간→4시간)도 검토했다. 재알림 회차를 세는 durable 카운터가 필요하고,
`attempts`는 전송 시도 수라 의미가 다르다. 고정 창이 더 단순하고 예측 가능하며, 무엇보다
**설명 가능한 상한**을 준다. 필요해지면 그때 바꾼다.

## D5ter. 배타 구간

판정과 전송은 `claimAndDeliver`가 `n.mu` 하나로 감싼다.

1판은 claim이 잠금 **밖**이었다. 트랜잭션은 커밋되고 연결이 반환된 뒤에야 `deliver`가
잠금을 잡았으므로, 두 관측이 아직 전달되지 않은 같은 행을 읽고 둘 다 owed로 판정한 뒤
차례로 발행할 수 있었다. 두 번째 발행은 이미 DELIVERED인 행을 표시하려다 실패하고,
그 실패가 `no such alert` 줄이다 — 폭주가 남긴 것과 같은 줄.

durable CAS(예: `SENDING` 상태)도 가능하지만 스키마 변경이 필요하고, 이 저장소에서
`SchemaVersion`을 올리는 것은 배포된 콘솔·엔진과의 정합 문제를 만든다. 엔진은
`engine.lock` flock으로 프로세스가 하나뿐이므로 in-process 뮤텍스면 충분하다.

`Flush`도 같은 잠금을 잡는다. 1판까지 그 함수는 잠금을 **전혀** 잡지 않았고, 리뷰가
지적한 것보다 넓은 구멍이었다.

## D6. 테스트가 무엇을 세야 하는가

기존 [`TestTheSameConditionEnqueuesOnce`](../../internal/obs/obs_test.go#L500)는 이름이
말하는 것을 세지 않는다. 두 가지가 겹쳐 있다.

1. **행 수를 센다.** 행 수는 `event_key` UNIQUE가 보장하므로 전송 경로를 통과하지 않는다.
2. **`failingPublisher{fail: true}`를 쓴다.** 행이 PENDING으로 남는 유일한 경우, 즉
   재전송이 **정당한** 경우만 검증한다.

이 테스트를 고치지 않고 새 검사를 추가한다 — 기존 테스트의 주장(행은 하나)은 여전히 참이고
지킬 가치가 있다. 새 검사는 전송이 **성공하는** 상태에서 `pub.callCount()`를 센다.
그 헬퍼는 [obs_test.go:326](../../internal/obs/obs_test.go#L326)에 이미 있다.

커버리지를 근거로 삼아서도 안 된다. `-covermode=set`은 블록의 실행 **여부**를 0/1로
기록하고 횟수를 세지 않으므로, 폭주하는 코드와 한 번만 도는 코드가 같은 값을 낸다.
1판 문서가 `deliver` B5의 `1`을 횟수로 읽었고 독립 리뷰가 그것을 지적했다. 이 change에서
커버리지로 말할 수 있는 것은 "그 경로가 발생했다/않았다"까지다.

새 코드는 새 파일에 둔다(`internal/obs/a096_one_send_per_condition_test.go`,
`internal/journal/a096_claim_for_delivery_test.go`) — 기존 파일에 새 함수를 넣으면 다음
change의 logic-map 범위가 그 파일 전체로 번진다. 같은 이유로 **기존 테스트를 한 줄도
고치지 않는 것**이 D1의 설계 제약이었다.

## D7. 무엇을 측정해 완료로 보는가

RED에서 측정한 값(1판 기준, 조건은 2판에서도 같다):

```text
--- FAIL: TestOneConditionIsOneSend        sends = 3, want 1
--- FAIL: TestSuppressingTheSendKeepsTheRecord
      sends = 5, want 1
      log lines = 9, want 5
```

`log lines = 9`는 관측 5줄 + `alert_undelivered` 오류 4줄이다. 그 4줄이 운영 로그의
`no such alert`와 같은 것이고, 각각이 이미 나간 push 하나를 뜻한다.

GREEN에서 세 테스트 모두 통과하고, 커버리지가 독립적으로 같은 것을 말한다:

```text
RED   notifier.go:258.88,260.5 1 1   ← MarkAlertDelivered가 오류를 돌려준 블록: 진입
GREEN notifier.go:276.88,278.5 1 0   ← 같은 블록(+18): 미진입
```

이 블록은 이미 전달된 행에 다시 전달 표시를 시도했을 때만 들어간다. 미진입은 그 경로가
스위트 전체에서 **발생하지 않았다**는 뜻이다 — 몇 번 발생했는지는 `set` 모드가 말하지
않으므로 그 이상 읽지 않는다.

배포 후 실측 조건은 같은 신호의 운영 측 대응물이다: `engine.log`에서
`engine.alert_undelivered`의 `no such alert` 오류가 0건.
