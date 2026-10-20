# a074 · 보호가 멈춘 순간이 보이고, critical 알림이 실제로 도착한다

## 왜

운영 원장에서 세 건을 나란히 읽으면 결함이 그대로 드러난다.

| 포지션 | 격리 생성 (`quarantined_at`) | 첫 알림 (`alert_outbox.created_at`) | 간격 |
|---|---|---|---|
| kr 466100 | `2026-08-03T09:03:40Z` | `09:03:45Z` | 5초 |
| us IONQ | `2026-08-03T13:45:25Z` | `13:45:30Z` | 5초 |
| us TSLA | `2026-08-03T14:53:11Z` | `14:53:16Z` | 5초 |

5초는 exit 관측 주기다. **보호가 멈춘 사이클에는 아무 기록도 없고, 그 다음 사이클이
결과를 보고한다.** 그리고 그 결과 알림조차 지금까지 한 번도 발송되지 않았다.

### 결함 1 — 격리를 만든 사이클은 아무것도 남기지 않는다

세 건 모두 `ambiguous_recovery`이고, 그 격리는 원장 트랜잭션 안
(`exit_state.go:499`)에서 만들어진 뒤 호출자에게 `ErrExitSnapshotQuarantined`로
돌아온다. 그 error가 어디로 가는지 따라가면 이렇다.

```
exitloop.go:1053  recorded, err := RecordExitJudgementResult(...)
exitloop.go:1055  errors.Is(err, ErrProposalPending) → 아님
exitloop.go:1061  return fmt.Errorf("engine: recording the exit judgement of %s: %w", …)
exitloop.go:429   if err := o.judge(…); err != nil && cycle.Err == nil { cycle.Err = err }
exitloop.go:349   _ = o.ObserveOnce(ctx)      ← 사이클이 통째로 버려진다
```

`ExitCycle.Err`의 주석은 이렇게 쓰여 있다.

> `Err is the cycle's first failure, if any. It is reported and not returned by Run`

**보고하는 코드가 없다.** `Run`은 반환값을 `_`에 버리고, exit 루프는 runtime에
`Health` 없이 등록되어 있어(`cmd/tossctl/engine.go:373-380`) 감독자도 사이클 실패를
셀 수 없다. 격리 생성뿐 아니라 이 루프의 **모든** 사이클 실패가 지금 관측 불가다.

### 결함 2 — 격리 생성에는 자기 이벤트가 없다

지금 남는 유일한 신호는 다음 사이클의 `exit.judgement_refused`다. 그것은
**결과**이지 사건이 아니고, 포지션당 in-process latch(`o.refused[positionID]`)가
걸려 있어 이미 다른 이유로 거부 중이던 포지션이 격리되면 **한 줄도 남지 않는다.**
격리 version·사유·증거·격리 시각 중 어느 것도 알림에 실리지 않는다.

### 결함 3 — critical 알림은 outbox에서 멈춘다

```
id  event_type                created_at            state    attempts
1   engine.operating_mode     2026-07-31T09:55:49Z  PENDING  0
2   engine.loop_degraded      2026-08-01T14:55:13Z  PENDING  0
3   exit.observation_outage   2026-08-01T19:31:19Z  PENDING  0
4-6 exit.judgement_refused    2026-08-03T…          PENDING  0
```

`attempts=0`이 결정적이다. `Notifier.deliver`는 `Publisher == nil`이면 시도를
기록하기 전에 `break`한다. 즉 **전송이 실패한 것이 아니라 시도된 적이 없다.**
`obs.Ntfy`는 완성돼 있지만 `cmd/tossctl/engine_assembly.go`는 Publisher를 의도적으로
비워 두고, 그 주석은 이유를 이렇게 적어 두었다 — "configuring a transport is an
operational setting with an audit trail, which is a change of its own."

**a074가 그 change다.**

## 무엇을

세 결함을 각각 그 원인 지점에서 고친다.

1. **사이클 실패를 보고한다.** `ExitObserver.Run`이 각 사이클의 결과를 읽고, 실패한
   사이클을 구조화 로그에 error 등급으로 남긴다. 주석이 이미 약속한 동작이다.
2. **격리 생성에 자기 이벤트를 준다.** 새 critical 이벤트
   `exit.snapshot_quarantined`를 격리가 생긴 그 사이클에, 격리를 만드는 세 경로
   전부에 대해 발행한다. version·사유·증거·격리 시각·세대를 싣는다.
3. **전송 경로를 설정 가능하게 만든다.** `engine.notifications` 설정 블록을 추가하고
   `NewContext`가 그것으로 `obs.Ntfy`를 조립한다. 토큰은 설정 파일이 아니라
   환경변수에서 읽고, 설정은 audit trail에 기록한다.

## 무엇을 하지 않는가

- 격리를 만드는 **조건**을 넓히거나 좁히지 않는다. a062가 고친 판정 로직과
  `SelectRecoverySnapshot`을 건드리지 않는다.
- 사이클 실패를 critical 알림으로 만들지 않는다. 일시적 원장 오류가
  ENTRY_BLOCKED를 유발하게 되기 때문이다 — 로그 등급까지만 올린다.
- 격리를 자동으로 해제하지 않는다. 해제는 a063이 만든 사람 경로뿐이다 (§0.7).
- 알림 설정을 켜지 않는다. 기본값은 off이고 off는 오늘 동작과 구별 불가다 (§0.2).
- 토큰·토픽을 로그·audit·기억에 남기지 않는다 (§0.8).

## 영향

| 영역 | 내용 |
|---|---|
| spec | `exit-policy` ADDED 1 · `engine-safety` ADDED 1 |
| 편집되는 기존 함수 | 6 (`Run`·`workingSet`·`record`·`mergeEngine`·`recordGateSettings`·`NewContext`) |
| 새 파일 | `internal/config` 알림 블록 · `cmd/tossctl` publisher 조립 · 테스트 |
| High-risk | exit 관측 루프와 엔진 기동 경로를 편집한다. 판정·주문 로직은 바꾸지 않는다 |
| §0.2 | `notifications.enabled` false = 현재 동작 (Publisher nil) |
