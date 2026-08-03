# a061 · Issues

## I1 — `httpapi_reader.go`에 같은 결함이 있다 (범위 밖, 기록)

```go
// cmd/tossctl/httpapi_reader.go:25, :256
const httpAPIPositionFreshness = 30 * time.Second
view := stored.Exit.Snapshot.WithFreshness(asOf, httpAPIPositionFreshness)
```

콘솔과 완전히 같은 원인이다. a051의 private operator daemon이 반환하는 모든
`ExitLine`도 30초 뒤 닫힌다.

**a061에서 고치지 않은 이유**: httpapi는 별도 프로세스이고 엔진 마커를 배선받지
않는다. 콘솔의 D2 판정을 그대로 옮기려면 그 daemon의 wiring과 spec을 함께 건드려야
하고, 그것은 a061이 고치라고 요청받은 화면이 아니다.

**옳은 해소 경로**: 엔진이 `last_observed_at`을 정직한 하트비트로 만드는 후속
change. 그것이 오면 httpapi는 상수 하나로 정상 동작한다 — 나이 판정이 다시
의미를 갖기 때문이다. 그때까지 httpapi 소비자는 `ExitLine`이 아니라
`StoredExitEvidence`를 봐야 한다.

## I2 — 466100 클로봇의 quarantine은 운영자 결정이 필요한 실제 보호 공백

```
pos-3b14217c40e2a96c3f16c35e  466100  ambiguous_recovery
  evidence: exitpolicy: recovery candidate identity mismatch
  quarantined_at 2026-08-03T09:03:40Z   released_at NULL
```

엔진은 09:03:45Z에 `exit.judgement_refused`를 남겼다 — "the stored protection state
or the observed price is not usable, so this position is not being judged at all".
**2026-08-03 11:34 기준 2시간 반째 손절 판정 대상이 아니다.**

a061은 이 사실을 화면에 **표시**만 한다. 해제는 `QuarantineReleaseHumanRepair` 또는
`QuarantineReleaseAuthoritativeReconcile`이고 §0.7에 따라 사람이 직접 승인한다.
콘솔에 해제 경로를 만들지 않는다.

## I3 — critical 알림이 아무 데도 전달되지 않았다

같은 시각 로그: `engine.alert_undelivered … no notification publisher is configured`.

I2가 2시간 반 동안 아무에게도 도달하지 않은 이유다. 판정 격리는 심각도 critical로
분류돼 있는데 publisher가 없으면 로그 파일에만 남는다. a061이 화면에 올리는 것은
그 공백을 부분적으로 메우지만(운영자가 화면을 볼 때에 한해), publisher 배선은 별개
문제다.

## I4 — `WithFreshness`는 남는다

a061은 `journal.ExitSnapshotView.WithFreshness`를 삭제하거나 바꾸지 않는다.
순수 함수이고 journal 테스트와 httpapi가 쓰며, 미배선 콘솔(D2 세 번째 갈래)도
계속 쓴다. 후속 change가 `last_observed_at`을 정직하게 만들면 이 함수의 의미가
비로소 이름과 일치하게 된다 — 지금은 함수가 옳고 그것이 읽는 컬럼이 틀렸다.

## I5 — quarantine 표시는 exit state가 있는 행에만 닿는다

`attachPositionExitLines`는 `!row.HasExit`이면 reference만 쓰고 `continue`하므로,
exit state가 없는 포지션에 quarantine이 걸려 있으면 그 사실이 화면에 나오지 않는다.

현재 코드에서 이 조합은 발생하지 않는다. quarantine은 `exitloop`가 exit state를
읽는 도중에만 만들어지고(`quarantineExitSnapshotTx`), 그 state 행은 지워지지 않으며
`accountExitStates`가 completed 여부와 무관하게 전부 되돌려주므로 `HasExit`는 참이다.

그러나 이것은 **강제되는 불변식이 아니라 현재 HEAD에 대한 관찰**이다. exit state를
지우거나 quarantine을 다른 경로에서 만드는 변경이 생기면 조용히 깨진다. 후속
change에서 `HasExit`와 무관하게 quarantine을 표시하도록 옮기는 것이 옳다.

## I6 — CodeGraphContext advisory 인덱스가 갱신되지 않았다

`make sdd-sync`에서 `codegraphcontext update . --quiet`가 300초 timeout으로 실패했다.
CodeGraph hard-evidence 인덱스는 정상 갱신됐고 `make sdd-check`가 worktree와 일치를
확인했다. CodeGraphContext와 GBrain은 advisory이므로 gate를 막지 않는다.
GBrain은 다른 세션이 소유 중이라 busy로 건너뛰었다.
