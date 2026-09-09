# a079 · Issues

## I1 — 운영 원장의 활성 격리는 3건이고, 셋 다 a078 결함이다

2026-08-04 읽기 전용 확인. 보유 5건 중 3건이다.

```
kr 466100  gen 1  v1  ambiguous_recovery  2026-08-03T09:03:40Z
us IONQ    gen 3  v1  ambiguous_recovery  2026-08-03T13:45:25Z
us TSLA    gen 1  v1  ambiguous_recovery  2026-08-03T14:53:11Z
```

셋 다 evidence가 맨몸 `exitpolicy: recovery candidate identity mismatch`다 —
a062가 고친 그 경로다. 8/3 조사 시점에는 466100 한 건이었고, 그 뒤 IONQ와 TSLA가
차례로 첫 rung을 밟으며 같은 결함에 걸렸다. a078 조사에서 예측한 blast radius가
그대로 실현된 것이다.

**a063은 이 3건을 자동으로 풀지 않는다.** 도구만 만든다. 실행은 §0.7 사람 판정이다.

## I2 — 도달 불가능한 안전 가드를 변이 검증이 잡아냈다

`ReleaseQuarantine`의 첫 초안에는 version 비교가 **두 번** 있었다.

```go
current, err := s.findQuarantine(ctx, repo, grant.request)   // 여기서 이미 비교
if current.Version != grant.row.Version { … }                 // 도달 불가
```

`Preview`는 `row.Version == req.Version`인 후보에만 grant를 발급하므로
`grant.row.Version == grant.request.Version`이 항상 참이고, 따라서 두 번째 비교는
어떤 입력으로도 참이 되지 않는다. 변이 검증에서 그것을 지웠는데 **아무 테스트도
실패하지 않아** 발견했다.

안전 경로의 죽은 가드는 검증되지 않은 보호처럼 읽힌다. 제거하고 왜 하나로 충분한지를
주석에 남겼다. 이 change에서 변이 검증이 실제로 값을 한 건이다.

## I3 — `State.AdoptionGeneration`은 `positions.instance_seq`가 아니다

콘솔 행 조립의 첫 초안은 격리를 `quarantine.Generation == state.AdoptionGeneration`로
매칭했다. 두 값은 다르다.

- `AdoptionGeneration`은 `position_policy_lifecycles.adoption_generation`이고,
  lifecycle 명령이 한 번도 없던 포지션에서는 **기본값 1**이다.
- 격리의 generation은 `positions.instance_seq`다.

운영 원장에서 IONQ는 `instance_seq=3`, `adoption_generation=NULL(→1)`이다. 즉 초안
그대로였으면 **IONQ와 042660의 격리 badge가 조용히 숨겨졌다.** 세대 규칙을 원장
읽기 한 곳(`q.position_generation = p.instance_seq`)에만 두고 콘솔은 position id로만
매칭하도록 고쳤다.

## I4 — 조작된 격리 판정은 원장 경로로 재현할 수 없다 (설계대로)

7.2를 "조작한 crossed-axes 판정을 `RecordExitJudgement`에 넣어 재격리를 본다"로
쓰려 했으나 불가능하다. `ValidateRecoveryDerivation`이 정확한 evaluator 입력을 다시
돌려 저장 line과 field 단위로 일치할 것을 요구하므로, 손으로 고친 후보는 복구 선택에
도달하기 전에 invalid로 거부된다.

또한 `SnapshotID`는 `InputDigest`와 policy digest만 봉인하고 rung·protection은 보지
않으므로, 이전 snapshot을 복사해 값만 바꾸면 **decision id가 같아져** 중복 판정으로
걸러진다. a062의 `TestAGenuinelyAmbiguousJudgementIsStillQuarantined`가 두 오류를
모두 허용하도록 쓰여 있는 것은 이 때문이며, 그 테스트는 실제로는 격리 경로가 아니라
중복/무효 경로를 확인하고 있었을 가능성이 있다. **후속 확인 대상.**

a063은 재현 가능한 것만 주장하도록 7.2를 다시 썼다 — 해제가 원장의 재격리 능력을
소모하지 않는다는 것. 판정 쪽은 a062의 exitpolicy 테스트가 이미 고정하고 있고
a063은 그것을 건드리지 않는다.

## I5 — 격리를 만드는 조건은 여전히 셋이고 a062는 하나만 없앴다

`stored_snapshot_corrupt`와 `legacy_policy_identity_unknown`은 그대로 살아 있다.
둘 다 정당한 fail-closed이며 없애면 안 된다. a063이 만드는 것은 그것들이 발동했을 때
**출구가 있다**는 사실이다.

## I6 — 해제 이력을 보여주는 화면이 없다

`exit_snapshot_quarantines`는 `released_at`·`release_kind`·`release_evidence`를
보존하지만 어떤 화면도 과거 격리 이력을 보여주지 않는다. 활성 격리만 보인다.
"이 포지션이 지난달에 몇 번 격리됐다 풀렸나"는 지금 원장을 직접 읽어야 알 수 있다.
a079 범위 밖이며 후속 change 후보다.

## I7 — 격리 생성 시점의 관측과 알림은 a074가 맡는다

a063은 이미 만들어진 격리를 다루는 change다. 격리가 **생기는** 순간이 로그에 없다는
사실과 critical 알림 publisher가 미배선이라는 사실은 a074의 범위다.

## 배포 후 실측 (2026-08-04T04:28Z, task 11.1 부분)

병합 빌드를 올린 뒤 `/position-management`를 GET으로만 읽었다. POST는 하나도 보내지
않았다.

| 항목 | 결과 |
|---|---|
| 활성 격리 badge | **확인.** 응답 본문에 `판정 격리` 10회. 격리 다섯 건 모두 사유·경과 시간·유지 보호선과 함께 렌더된다 |
| 해제 action | **확인.** `판정 격리 해제`가 CSRF 토큰을 실은 form으로 행에 붙어 있다 |
| 해제 **화면** | **미측정.** 해제 화면은 `/position-management/quarantine/preview` POST가 만든다. preview는 `mutating` 게이트 뒤에 있고, 에이전트가 자동으로 보내지 않는다(§0.7, task 11.2와 같은 이유) |

badge와 action이 뜨는 것까지 확인했고 화면 자체는 사람이 눌러야 보이므로 task 11.1은
열어 둔다.

## 해제 화면은 이미 세 번 쓰였다 (2026-09-09 원장 실측, task 11.1 나머지 하나)

8월 기록은 badge와 해제 action까지 확인하고 해제 **화면**을 "미측정"으로 남겼다.
그 기록을 쓴 04:28Z로부터 21분 뒤, 사람이 콘솔에서 격리 셋을 실제로 해제했다.
원장이 그것을 보존하고 있다.

| 종목 | `released_at` | `release_kind` | `release_evidence` |
|---|---|---|---|
| kr 042660 | 2026-08-04T04:49:48Z | `HUMAN_REPAIR` | `LOCAL_OPERATOR released quarantine v1 (reason=ambiguous_recovery) from the console at 2026-08-04T04:49:48Z` |
| us IONQ | 2026-08-04T04:50:09Z | `HUMAN_REPAIR` | 같은 형식, 같은 시각 |
| us TSLA | 2026-08-04T04:50:23Z | `HUMAN_REPAIR` | 같은 형식, 같은 시각 |

`HUMAN_REPAIR` + 서버가 조립한 `LOCAL_OPERATOR … from the console` 문장은 task 4.7이
고정한 성공 경로 그 자체다.

**해제가 preview 화면을 반드시 거쳤다는 구조 근거.** `ReleaseQuarantine`은 32바이트
난수 capability를 검증하고(`exit_quarantine_command.go:231`), 그 난수를 만드는
`issueQuarantineCapability`의 호출자는 **하나뿐**이다 — `PreviewQuarantineRelease`
(`:119`, 저장소 전체에서 유일). 그 값이 브라우저에 닿는 길도 하나뿐이다 —
`quarantine-release-preview` 템플릿 렌더(`internal/console/exit_quarantine.go:217`).
게다가 `exitQuarantineCapabilityDelay = 3 * time.Second`라 발급 3초 안에는 거부되고,
확인 없는 apply도 거부된다(4.3). 즉 해제가 원장에 남았다는 사실은 preview 화면이
운영자 앞에 그려지고 3초 넘게 읽힌 뒤 확인을 받았다는 뜻이다. 세 번.

세 해제의 간격(21초, 14초)도 사람이 화면 셋을 차례로 넘긴 모양이다.

**task 11.1을 닫는다** — badge와 action은 2026-08-04T04:28Z GET으로, 해제 화면은
같은 날 04:49~04:50Z의 해제 세 건으로 측정됐다. task 11.2가 못 박은 대로 누른 것은
사람이고 에이전트가 아니다.

## 남은 창이 닫혔다 — 지금은 재측정할 수 없다 (2026-09-09)

같은 실측에서 확인한 것이다. 오늘은 해제 화면에 도달할 방법이 없다.

| 자리 | 술어 | 오늘 |
|---|---|---|
| `internal/journal/exit_quarantine_list.go:46` | `q.released_at IS NULL AND q.position_generation = p.instance_seq` | 5건 |
| `internal/journal/position_policy.go:85` | `WHERE p.state <> 'CLOSED'` | 3건 |
| `internal/console/position_policy.go:321` | 두 번째 집합의 행에만 badge를 붙인다 | **0건** |

활성 격리 5건은 전부 CLOSED 포지션이고, 열린 포지션 3건은 하나도 격리되지 않았다.
badge가 없으면 서명된 `quarantine_token`이 없고, token이 없으면 preview POST가
성립하지 않는다.

배선 자체는 살아 있다는 것을 따로 확인했다.

- 배포 바이너리에 a079 문자열이 있다(`격리 해제 표면 없음` 1건, `판정 격리 v` 1건).
- 콘솔이 받는 `*consolePositionPolicyCommander`가 세 메서드를 모두 갖고
  (`cmd/tossctl/exit_quarantine_commander.go:34,42,51`), 그 `lifecycle`인
  `*positionpolicyrpc.Client`도 같은 셋을 만족한다(컴파일 시점 확인). 따라서
  `exitQuarantines()`는 언제나 ok를 돌려주고 `Quarantines(ctx)`는 실제 RPC다.
- `/position-management` 응답에 `판정 격리 상태를 읽지 못했다` notice가 **없다**.
  즉 그 RPC가 성공했다 — 엔진이 5건을 돌려줬고 화면이 조용히 버렸다.

## I8 — 활성 격리 read와 화면 범위가 어긋난다 (후속 change 후보)

`ActiveExitSnapshotQuarantines`의 주석은 "관측할 수 없는 것에 대한 해제 버튼을 주지
않는다"를 세대 술어의 근거로 적지만 `p.state`는 거르지 않는다. 그래서 read가
"활성"이라 부르는 5건을 화면은 구조적으로 못 그리고, 어떤 문장도 그 사실을 말하지
않는다. CLOSED 포지션의 격리가 무해하다면 read가 걸러야 하고, 무해하지 않다면 화면이
말해야 한다. 지금은 둘 다 아니다. a079 범위 밖이다.
