# a064 · Issues

## I1 — `ExitCycle.Err`는 쓰이기만 하고 읽히지 않았다

`ObserveOnce`는 사이클 실패를 `cycle.Err`에 담고, `Run`은 `_ = o.ObserveOnce(ctx)`로
사이클을 통째로 버렸다. exit 루프는 runtime에 `Health` 없이 등록되어 있어
(`cmd/tossctl/engine.go`) 감독자의 열화 사다리도 이 값을 셀 수 없다.

즉 **격리 생성만의 문제가 아니었다.** 원장 write 실패, 심볼 하나의 break-even 계산
실패, 판정 기록 실패 — 이 루프의 모든 사이클 실패가 관측 불가였다. 가격 관측 두절만
별도 경로(`checkOutage`)로 알림된다.

필드 선언 주석은 처음부터 이렇게 쓰여 있었다.

> `Err is the cycle's first failure, if any. It is reported and not returned by Run`

**보고하는 코드가 없었다.** a064가 그 문장을 참으로 만든다.

## I2 — `attempts=0`은 전송 실패가 아니라 시도 없음이다

운영 원장의 critical 알림 6건이 전부 `state=PENDING, attempts=0`이었다. 처음에는
"전송이 계속 실패하는구나"로 읽었으나 `Notifier.deliver`를 보면 반대다.

```go
for attempt := 1; attempt <= attempts; attempt++ {
    if n.Publisher == nil {
        lastErr = errors.New("no notification publisher is configured")
        break            // ← 시도를 기록하기 전에 빠져나간다
    }
```

`attempts=0`은 **한 번도 보내려 하지 않았다**는 뜻이다. 가장 오래된 행은
2026-07-31T09:55:49Z의 `engine.operating_mode`이고, 그때부터 지금까지 어떤 알림도
기계 밖으로 나간 적이 없다.

## I3 — 조작된 격리는 원장으로 재현할 수 없지만, **어긋난 두 기록**은 재현된다

a069 issues I4는 "조작한 판정을 `RecordExitJudgement`에 넣어 재격리를 보는 것은
불가능하다"고 기록했다. 여전히 맞다 — 저장된 snapshot을 손대면
`ValidateRecoveryDerivation`이나 output digest가 먼저 corrupt로 걸러낸다.

a064는 다른 각도를 찾았다. **저장된 snapshot은 그대로 두고 `exit_states.entry_price`
칼럼만 바꾼다.** snapshot은 자기 자신에 대해 여전히 유효하고, 다음 평가가 그 칼럼에서
다시 계산한 line만 달라진다 → `saved.EntryPrice != recomputed.EntryPrice` →
`ErrRecoveryIdentity` → `ambiguous_recovery` 격리. 운영에서 실제로 벌어진 것과 같은
모양이며, 이제 `TestAnAmbiguousRecoveryQuarantineIsAnnouncedInTheSameCycle`이
그 경로를 실행한다.

**a063의 7.2가 주장을 좁혔던 이유가 이것으로 해소된다.** 후속으로 a062의
`TestAGenuinelyAmbiguousJudgementIsStillQuarantined`가 실제로 격리 경로를 보고 있는지
이 기법으로 재확인할 수 있다.

## I4 — 변이 검증이 테스트의 빈 곳을 하나 잡았다

M3(in-process latch 제거)이 처음에는 **아무 테스트도 깨뜨리지 않았다.**

이유는 latch가 걸리는 자리를 잘못 짚었기 때문이다. `TestAnAlreadyActiveQuarantine…`은
`workingSet` B15 경로인데 그 경로는 애초에 발행하지 않는다. latch가 실제로 하중을 받는
곳은 **corrupt snapshot 경로(B11)** 다 — corruption 검사가 활성 격리 검사보다 **앞**에
있어서, corrupt 행은 매 사이클 `QuarantineExitSnapshot`을 호출하고 같은 행을 돌려받는다.

테스트를 1사이클에서 3사이클로 고치자 M3이 "announcements = 3, want exactly one"으로
정확히 RED가 됐다. latch가 없으면 5초마다 gate block과 error 줄이 영원히 반복된다.

## I5 — 알림 설정에는 토큰을 담을 필드가 없다

설계 결정이며 구조로 강제한다. `config.Notifications`에는 `token` 필드가 아예 없고,
`rawNotifications`에도 없다. 설정 파일에 토큰을 써도 **착지할 자리가 없다.**

`TOSSCTL_NTFY_TOPIC`이 파일의 topic을 덮을 수 있는 것도 같은 이유다 — ntfy.sh에서는
topic 이름이 유일한 접근 제어이므로 config를 읽을 수 있는 모든 것이 알림 채널에 쓸 수
있게 된다. 자체 호스팅 + 토큰 구성에서는 topic이 비밀이 아니므로 파일에 두는 편이
diff 가능하고 낫다. 두 운영 형태가 실제로 다르므로 둘 다 허용한다.

## I6 — 알림 설정은 아직 꺼져 있다

a064는 **켤 수 있게** 만들 뿐 켜지 않는다. 기본값은 off이고 off는 오늘과 구별
불가능하다. 실제로 켜려면 운영자가 `config.json`의 `engine.notifications`를 쓰고
`TOSSCTL_NTFY_TOKEN`을 컨테이너 환경에 넣어야 하며, 그것은 §0.7 사람 판정이다.

**켜지기 전까지 critical 알림은 여전히 outbox에서 멈춘다.** a064가 바꾸는 것은
"멈추는 이유가 배선 부재"에서 "멈추는 이유가 설정 미선택"으로 옮기는 것이다.

## I7 — 활성 격리 3건은 여전히 그대로다

a064는 격리를 **보이게** 만들고, a063은 **풀 수 있게** 만들었다. 실제로 푸는 것은
사람의 일이다 (§0.7). 466100·IONQ·TSLA는 배포 후 콘솔에서 해제할 수 있으며, 해제하지
않으면 계속 판정되지 않는다.

## I8 — exit 루프는 여전히 `Health`가 없다

runtime의 열화 사다리(연속 5주기 실패 → critical + ENTRY_BLOCKED)는 exit 루프에
적용되지 않는다. 주석이 이유를 적고 있다 — "the exit observer's degradation ladder is
its own landed 60-second observation-outage contract, and a second threshold on top
would be two definitions of 'exit observation is down' that could disagree."

a064는 그 결정을 바꾸지 않았다. 다만 이제 사이클 실패가 **로그에는** 남으므로, 두절
계약이 잡지 못하는 종류의 반복 실패(판정 기록 실패 등)를 사람이 셀 수 있다. 그것을
자동 임계로 만들지 여부는 별개의 판단이며 후속 change 후보다.
