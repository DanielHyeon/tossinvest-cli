# a068 · Issues

## I1 — 466100의 기존 격리는 이 수정으로 풀리지 않는다 (운영자 결정 필요)

```
pos-3b14217c40e2a96c3f16c35e  466100  gen 1  version 1
  reason:   ambiguous_recovery
  evidence: exitpolicy: recovery candidate identity mismatch
  quarantined_at 2026-08-03T09:03:40Z   released_at NULL
```

a062는 **앞으로의** 격리를 막는다. 이미 원장에 있는 행은 그대로이고 `exitloop`는
계속 이 포지션을 건너뛴다. 해제 경로는 세 가지이며 전부 §0.7 사람 판정이다.

| 경로 | 상태 | 결과 |
|---|---|---|
| `Journal.ReleaseExitSnapshotQuarantine(HUMAN_REPAIR)` | **호출자 없음** — CLI에도 콘솔에도 배선돼 있지 않다 | 배선하려면 별도 change |
| 콘솔 `자동관리 해제` → `새 generation 재편입` | 배선돼 있음 | 새 세대는 옛 격리와 세대가 달라 막히지 않는다. 다만 **재편입은 현재가 기준으로 기준선을 다시 만든다** — 진입가 25,700/손절 24,929가 사라지고 현재가 기준 새 손절폭이 잡힌다 |
| 그대로 두고 수동 관리 | — | 엔진이 이 포지션을 보호하지 않는다는 사실을 운영자가 계속 안고 간다 |

**즉 현재 코드에는 "격리를 풀고 원래 기준선을 유지하는" 경로가 없다.** 이것 자체가
설계 공백이고, 후속 change 후보다.

## I2 — 격리 생성 순간이 로그에 없다

`quarantineExitSnapshotTx`는 조용히 행을 만든다. 09:03:40에 보호가 멈춘 사건은
어떤 로그에도 없고, 5초 뒤 그 **결과**인 `exit.judgement_refused`만 남았다.

```
09:03:40Z  (격리 생성 — 로그 없음)
09:03:45Z  WARN exit.judgement_refused severity=critical … "not being judged at all"
```

사후 조사에서 "언제부터 보호가 없었나"를 원장 타임스탬프로만 알 수 있었다. 격리
생성 자체가 critical 이벤트여야 한다.

## I3 — critical 알림이 전달되지 않는다

```
09:03:45Z ERROR engine.alert_undelivered event=exit.judgement_refused alert_id=4
          error: no notification publisher is configured
```

"이 포지션은 판정되지 않는다"가 critical로 분류돼 있는데 publisher가 없어 로그
파일에만 남았다. a061이 이 사실을 화면에 올리므로 운영자가 콘솔을 볼 때는 보이지만,
그것은 push가 아니다. publisher 배선은 별개 문제다.

## I4 — 결함이 무테스트로 출하됐다

`internal/exitpolicy/recovery_test.go`의 두 테스트는 전부 `ActiveRung: NoRung`
fixture만 쓴다. rung을 활성화한 후보가 등장하는 테스트가 **하나도 없다**. 그래서
ladder의 첫 rung 전이라는, 모든 ladder 포지션이 반드시 통과하는 경로가 한 번도
검증되지 않았다.

a062는 그 전이를 양방향으로 고정한다(2.1, 2.2).

## I5 — 042660이 살아남은 이유는 운이다

8/2 23:23:25의 첫 ladder 판정이 곧바로 rung 0을 활성화했고, 그 시점에는 비교할
`saved_snapshot_json`이 NULL이었다. `SelectRecoverySnapshot(nil, …)`은 비교 없이
recomputed를 반환한다. 익절선 **아래**에서 한동안 머물며 canonical snapshot을 쌓은
뒤 넘는 종목만 걸린다 — 즉 정상적인 보유 대부분이다.
