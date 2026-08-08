# a096 — 한 조건은 한 알림이다

## 왜

2026-08-08(토) 10:10:43 KST, 거래일이 아닌 날 운영자의 기기로 같은 critical 알림이
**약 5.6초마다** 도착하기 시작했다. 10:16:16 기준 60건이고, 그대로 두면 하루 약 15,000건이다.

원장에는 그 알림이 **한 건**으로 기록돼 있다.

```
sqlite> select id, event_type, state, attempts, created_at from alert_outbox order by id desc limit 1;
14|exit.proposal_refused|DELIVERED|1|2026-08-08T01:10:43Z
```

행은 하나, 전송은 60번. 이 change는 그 간극을 닫는다.

## 무엇을 관측했나

`~/.config/tossctl/engine.log` (2026-08-08T01:10:43Z 이후, 실측):

```
WARN  exit.proposal_refused  symbol=272210 position_id=pos-8a171e1ca7188c731a93f5b6
      action=LADDER_PARTIAL level=2
      detail=broker refused the mutation with HTTP 422: official: API error 422:
             {"error":{"code":"order-hours-closed","message":"주문가능일이 아닙니다."}}
...5.6초마다 반복...
ERROR engine.alert_undelivered  error="journal: no such alert: 14 (or it is no longer pending)"
```

두 번째 줄이 결정적 증거다. 이 오류는 `MarkAlertDelivered`가 **이미 DELIVERED인 행**을
갱신하려다 0행을 갱신했을 때만 나온다([outbox.go:161](../../../internal/journal/outbox.go#L161),
`requireOneRow`는 [outbox.go:287](../../../internal/journal/outbox.go#L287)에서 0행을 오류로
바꾼다). 그 시점은 `Publish`가 **이미 성공한 뒤**다([notifier.go:256-261](../../../internal/obs/notifier.go#L256-L261)).
즉 이 오류가 한 번 찍힐 때마다 push는 이미 한 번 나갔다. 53건 찍혀 있다.

알림 채널은 ntfy push다(`config.json`의 `engine.notifications.enabled = true`).

## 무엇이 원인인가 — AST 증거

`tools/logic-map` 산출물을 먼저 만들었다. 아래 줄 번호는 전부 base `ec29dc72` 기준이다
(현재 산출물은 9함수 46분기이며 2판 코드를 가리킨다).

**① `Journal.EnqueueAlert` B5 — 중복을 알리지 않고 삼킨다**

[outbox.go:111-151](../../../internal/journal/outbox.go#L111-L151), 9분기.
**B5@130** `case err == nil:` → `return existing, tx.Commit()`.

이 함수의 주석은 스스로 이렇게 말한다:

> Enqueuing the same event key twice is not an error and does not create a second
> row: the caller observing the same condition again is the normal case, and
> **duplicating it would turn one problem into a pager storm.**

의도는 정확하다. 그러나 반환값 `(int64, error)`에는 **이 id가 방금 삽입된 것인지 이미
있던 것인지, 있었다면 이미 전달됐는지**를 담을 자리가 없다. 호출자는 구별할 수 없다.

**② `Notifier.notifyCritical` B4 — 구별하지 않으므로 무조건 보낸다**

[notifier.go:153-190](../../../internal/obs/notifier.go#L153-L190), 4분기.
**B4@182** `if !n.deliver(ctx, id, e)` — `EnqueueAlert`가 무엇을 돌려줬든 `deliver`를 부른다.
`deliver`는 [notifier.go:256](../../../internal/obs/notifier.go#L256)에서 `Publisher.Publish`를
조건 없이 실행한다.

**③ 그래서 중복 제거는 행에만 적용된다**

`event_key`의 UNIQUE 제약이 행을 하나로 묶는다. 전송에는 아무것도 적용되지 않는다.
주석이 막겠다고 쓴 pager storm을, 그 주석이 설명하는 바로 그 중복 제거가 **보장한다** —
행이 하나로 접히는 덕분에 조건이 사라지지 않고 매 사이클 되살아나기 때문이다.

## 왜 지금까지 안 잡혔나

이 계약을 지키기로 한 테스트가 이미 있다. 그리고 틀린 것을 센다.

[obs_test.go:500-523](../../../internal/obs/obs_test.go#L500-L523) `TestTheSameConditionEnqueuesOnce`:

```go
// three consecutive polls observing one contradiction is one alert, not a pager storm.
...
if len(pending) != 1 {
    t.Fatalf("outbox rows = %d, want 1 — the same condition is one alert", len(pending))
}
```

이름도 주석도 pager storm을 말하는데 **세는 것은 outbox 행 수**다. 행 수는 `event_key`
UNIQUE가 이미 보장하므로 이 검사는 스키마를 확인할 뿐 전송 경로를 통과하지 않는다.
게다가 이 테스트는 `failingPublisher{fail: true}`를 쓴다 — 행이 PENDING으로 남는 유일한
경우, 즉 **재전송이 정당한 경우**만 검증하고 있었다.

`pub.callCount()`는 이 파일 안에 이미 있고([obs_test.go:326](../../../internal/obs/obs_test.go#L326)),
다른 테스트들이 쓴다. 이 테스트만 부르지 않았다.

그리고 폭주 자체는 **이미 매 CI에서 실행되고 있었다.**
[obs_test.go:620-641](../../../internal/obs/obs_test.go#L620-L641) `TestNotifierIsConcurrencySafe`는
전송이 **성공하는** publisher로 6개 key × 20회 = critical `Notify` 120회를 돌린다. key당 20회 중
19회가 재전송이다. 그리고 이 테스트에는 **단언이 하나도 없다** — race detector용 smoke test다.

커버리지 실측이 이것을 확인한다. `deliver`의 `markErr != nil` 블록
([notifier.go:258.88-260.5](../../../internal/obs/notifier.go#L258-L260))은 `count 1`로 **진입한다**.
그 블록에 들어간다는 것은 `MarkAlertDelivered`가 이미 DELIVERED인 행을 갱신하려다 실패했다는
뜻이고, 그것은 운영에서 관측된 `no such alert` 오류와 **같은 지점**이다. 이 테스트만 격리해
측정해도 진입한다(다른 어떤 테스트도 진입하지 않는다).

즉 이 결함은 테스트되지 않은 것이 아니라 **단언 없이 실행되고 있었다.**

## 무엇을 바꾸나

**R1. 원장이 "이 전송이 아직 필요한가"에 답한다.** 새 함수
`Journal.ClaimAlertForDelivery(ctx, Alert, remindAfter) (int64, bool, error)`. 본체는 base의
`EnqueueAlert`를 그대로 옮긴 것이고, 더한 것은 `owed` 판정과 **재무장**이다.
id를 결정하는 것과 같은 트랜잭션 안의 읽기·쓰기이므로 TOCTOU 창이 없다.

판정은 순수 함수 `claimOwed`로 다시 분리한다: PENDING이면 owed, 종결(DELIVERED·
ACKNOWLEDGED)이면 재알림 창이 지났을 때만 owed, **모르는 상태면 owed이면서 PENDING으로
재무장**한다. 재무장하지 않으면 전달 완료 표시가 실패해 다음 관측에서 다시 전송된다.

`EnqueueAlert`는 **서명도 계약도 그대로** 두고 `remindAfter = 0`으로 위임한다. 기록만 하는
호출자(`execgw.Gateway.parkAlert` — replay 경로는 전송 루프가 없다)는 전달 여부를 답할 수
있는 위치가 아니고, 남의 리마인더 정책을 대신 실행해서도 안 된다.

**R2. 억제는 영구가 아니라 창이다.** 창을 넘긴 종결 행은 **PENDING으로 재무장**되어,
리마인더가 최초 전달과 **같은 경로**를 걷는다 — 재시도 예산, gate 잠금, 운영 모드 승격이
전부 그대로 적용된다. 기본 창은 1시간이다.

영구 억제는 1판의 설계였고 독립 리뷰가 두 가지로 깼다. 반복 전송이 **알림 경로가 살아
있다는 유일한 주기적 증거**였으므로 영구 억제는 transport 사망을 미탐지로 만들고, event key가
조건만 담고 원인을 담지 않으므로 한 원인의 알림이 같은 key의 다른 원인을 영구히 가린다.

**R3. 판정과 전송을 한 배타 구간에 둔다.** `claimAndDeliver`가 `n.mu`를 claim부터 send까지
잡는다. `Flush`도 같은 잠금을 잡는다 — 1판까지 그 함수는 잠금을 전혀 잡지 않았다.

**R4. 계약을 전송 횟수로 검증한다.** 기존 테스트가 세던 것을 바꾸지 않고, 전송이 성공하는
상태에서 전송 횟수를 세는 검사를 새로 만든다. 커버리지는 근거가 되지 못한다 —
`covermode=set`은 실행 여부만 기록하고 횟수를 세지 않는다.

### 설계를 두 번 고쳤다

**한 번은 게이트가, 한 번은 독립 리뷰가 고치게 했다.**

처음에는 `EnqueueAlert`의 서명을 `(int64, string, error)`로 넓혔다. 컴파일러가 모든
호출자에게 이 구별을 강제한다는 것이 근거였다. 그러자 `check_analysis.py`가 증거 누락
**8건**을 보고했다 — `Gateway.parkAlert`와 `internal/journal/outbox_test.go`의 기존 테스트
7개. arity 하나 때문에 원장 패키지의 기존 테스트 전부가 "수정된 기존 함수"가 됐고,
그 각각이 Function Logic Map을 요구했다. 생산됐을 28개 파일은 각 테스트의 안전성이 아니라
**인자 개수가 바뀌었다는 사실을 28번 적은 것**이었다.

좁힌 설계는 그 28개를 없앨 뿐 아니라 더 낫다. 강제하려던 유일한 기존 호출자
(`parkAlert`)는 전송을 하지 않으므로 강제할 것이 애초에 없었다.

그리고 1판의 억제는 **영구**였다. Codex 리뷰가 blocker 4건으로 깼고, 그중 둘이 설계
결함이었다 — 영구 억제가 알림 경로의 생존 감지기를 없애고(D5), 같은 key의 다른 원인을
영구히 가린다(D4). 2판의 창과 재무장이 그 정정이다. 나머지 둘은 동시성(D5ter)과
증거 정합(covermode 오독, FLM 헤더 stale)이었다.

## 무엇을 바꾸지 않나

- **등급.** critical은 그대로 critical이다.
- **재시도 예산과 진입 차단.** `deliver`의 3회·2초와 소진 뒤 `ModeEntryBlocked`는 a074 계약
  그대로다. 리마인더는 재무장을 거쳐 **같은 경로**를 걸으므로 그 계약이 자동으로 적용된다 —
  1판처럼 새 경로를 만들어 계약을 다시 검증해야 하는 상황이 아니다.
- **구조화 로그.** `logEvent`는 `Notify` 안에서 등급 분기보다 **먼저** 실행되므로
  ([notifier.go:126](../../../internal/obs/notifier.go#L126)) 관측마다 한 줄이 계속 남는다.
  억제되는 것은 push뿐이고 관측 사실의 기록이 아니다.
- **원장 스키마.** 열도 제약도 마이그레이션도 없다. 재무장은 기존 `state` 열의 UPDATE다.
  `SchemaVersion`은 30 그대로다.
- **손절·익절·사이징.** 이 change는 주문 경로를 건드리지 않는다. Go diff는
  `internal/journal/outbox.go`와 `internal/obs/notifier.go` 두 파일뿐이다.

## 범위 밖 — 그러나 같은 사고에서 나왔다

이 사고에는 결함이 **둘** 있고 서로 독립이다. a096은 ②만 다룬다.

**① 엔진에 장 운영 시간 게이트가 없다.** 비거래일에도 가격 피드가 답하고
(`exit_states.last_observed_at = 2026-08-08T01:15:19Z`, `ObservedPrice = 70600`),
사다리는 그 값을 정상 판단해 익절 발의를 만들고, 브로커가 `order-hours-closed`로 거절한다.
`GetTradingHours`는 [marketdata.go:568](../../../internal/client/marketdata.go#L568)에 있지만
CLI([market.go:44](../../../cmd/tossctl/market.go#L44))만 쓴다. 청산 경로는 조회하지 않는다.

**②만 고쳐도 알림 폭주는 멎지만 거절된 주문은 계속 나간다.** ①은 별도 change가 필요하다.
`order-hours-closed`는 a094가 다루는 `opposite-pending-order-exists`와 같은 모양(확정 거절)이라
그 R1 분류에 붙일 수 있다. 이 change에서 처리한다고 주장하지 않는다.

또한 ②는 ①과 무관하게 **반복되는 모든 critical 조건**에서 같은 폭주를 낸다. ①이 고쳐져도
②는 남는다. 그래서 먼저 고친다.

## 선행·후행

- 선행 없음. a094·a095와 파일이 겹치지 않는다(`internal/obs/notifier.go`는 a095가 `event.go`만 만진다).
- a091·a095가 `criticalEvents`를 확장하면 critical 이벤트가 늘어난다. 늘어난 각각이 반복
  조건이면 폭주 후보가 하나씩 늘어난다는 뜻이므로, a096은 그 둘보다 **먼저** 착지하는 것이 낫다.
- **a092와의 상호작용.** a092는 "알림이 관측 사이클을 붙잡지 않게" 만든다. a096은 잠금
  구간을 claim까지 넓혔으므로 그 구간이 조금 길어졌다(claim 트랜잭션 하나). 두 change가
  만나면 배타 구간을 비동기 경로로 옮기는 판단이 필요하다 — a096이 만든 문제는 아니지만
  a096이 관련되므로 tasks 7.5에 적어 둔다.
