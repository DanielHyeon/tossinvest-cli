# a079 · 운영자가 판정 격리를 해제할 수 있다

## 왜

exit snapshot 격리는 **단방향 문**이다. 현재 HEAD에서 손잡이를 세면 이렇다.

| 손잡이 | 개수 | 위치 |
|---|---|---|
| 격리를 **만드는** 경로 | 3 | `exitloop.go:494` `stored_snapshot_corrupt` · `exitloop.go:518` `legacy_policy_identity_unknown` · `exit_state.go:499` `ambiguous_recovery` |
| 격리를 **읽고 판정을 거부하는** 경로 | 1 | `exitloop.go:506` |
| 격리를 **푸는** 경로 | **0** | `Journal.ReleaseExitSnapshotQuarantine`는 프로덕션 호출자가 없다 |

`ReleaseExitSnapshotQuarantine`은 v10부터 존재하고 CAS·kind·evidence 검증까지 갖췄지만
CLI에도 콘솔에도 배선돼 있지 않다. `HUMAN_REPAIR`와 `AUTHORITATIVE_RECONCILE` 둘 다
호출자가 없고, 유일한 호출은 `migration_v10_test.go`의 테스트다.

설계 의도는 "판정할 수 없으니 사람이 보고 풀어라"인데 **사람이 잡을 손잡이가 없다.**

### 지금 걸려 있는 것

2026-08-04 기준 운영 원장의 활성 격리는 3건이고, 보유 5건 중 3건이다.

```
kr 466100  gen 1  v1  ambiguous_recovery  2026-08-03T09:03:40Z
us IONQ    gen 3  v1  ambiguous_recovery  2026-08-03T13:45:25Z
us TSLA    gen 1  v1  ambiguous_recovery  2026-08-03T14:53:11Z
```

셋 다 a062가 고친 결함이 원인이다. a062는 **앞으로의** 격리를 막지만 이미 만들어진
행은 그대로 두고, `exitloop`는 계속 이 세 포지션을 건너뛴다 — 손절 평가 포함.

a062의 수정만으로는 이 세 건이 돌아오지 않는다. 현재 코드에서 가능한 유일한 우회는
콘솔의 `자동관리 해제 → 새 generation 재편입`인데, 재편입은 **현재가 기준으로 기준선을
다시 만든다**. 466100이라면 진입가 25,700과 초기 손절 24,929가 사라지고 오늘 가격에서
새 손절폭이 잡힌다. 그것은 해제가 아니라 다른 포지션으로 바꾸는 일이다.

**"격리를 풀고 원래 기준선을 유지하는" 경로가 코드에 없다.** 이 change가 그것을 만든다.

## 무엇을

`/position-management`에 격리된 포지션 행의 「판정 격리 해제」 동작을 추가한다.

- 이미 존재하는 `Journal.ReleaseExitSnapshotQuarantine(HUMAN_REPAIR)`를 호출한다.
  그 함수를 바꾸지 않는다.
- 정책 CAS pipeline에 끼워 넣지 않고 **자기 seam**을 갖는다 (D1). 격리 해제는 정책
  상태를 전혀 바꾸지 않으므로 그 pipeline의 generation/version 의미와 맞지 않는다.
- 승인 모양은 정책 pipeline과 같다 — 서버 발급 1회용 capability, 3초 위험 지연,
  체크박스 확인, 운영자가 본 그 quarantine version에 대한 CAS.
- evidence는 **서버가 만든다**. 사용자가 문구를 타이핑하지 않는다 (D3).
- `exit_states`·`positions`·`position_policy_*`에 쓰지 않는다. 기준선은 그대로다 (D5).

## 무엇을 하지 않는가

- 격리 **생성** 조건을 넓히거나 좁히지 않는다.
- `AUTHORITATIVE_RECONCILE` 자동 해제를 배선하지 않는다. 자동 해제 경로는 만들지 않는다 (§0.7).
- 손절·익절 계산식과 평가기를 건드리지 않는다.
- 격리 생성 시점의 관측·알림은 a074가 맡는다. 이 change의 범위가 아니다.

## 영향

- spec: `position-exit-policy-management` ADDED 1건
- 신규 패키지 `internal/exitquarantine` (capability-neutral 계약)
- 신규 `internal/app/engine/exit_quarantine_command.go`, `internal/console/exit_quarantine.go`
- 기존 편집: RPC transport 라우트 3개, 콘솔 Options·라우트·행 조립, CLI 배선
- 원장 스키마 변경 없음. 신규 공식 API 호출 없음 (§0.4 해당 없음).
