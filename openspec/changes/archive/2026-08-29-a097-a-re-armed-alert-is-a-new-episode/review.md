# a097 리뷰 기록

## §0. 판정 요약

| 라운드 | 시점 | 대상 | 판정 |
|---|---|---|---|
| 1 | 2026-08-09 | proposal-freeze (proposal·design·spec delta·tasks 초판) | **BLOCK — P1 2건, P2 7건** |
| 2 | 2026-08-09 | 구현 후 독립 검증 (working tree diff + 새 테스트 4파일) | **BLOCK — P2 4건, P3 1건** |

초판은 구현 착수 전에 막혔다. 그것이 이 게이트의 목적이며, 여기서 막힌 두 P1은 **구현
뒤에 발견됐다면 안전 계약을 잘못 구현한 채 통과했을 것**이다.

특히 R2의 결론이 **뒤집혔다**: 초판은 claim 실패 시 운영 모드 승격을 **금지**했고,
2판은 **요구**한다.

2라운드는 구현을 봤고, 가장 아픈 지적은 **Function Logic Map이 구현과 반대를 말하고
있다**는 것이었다 — a096 3판에서 내가 남의 문서에 대해 지적했던 바로 그 결함을 내가
만들었다.

이 문서에는 **되돌린 판정이 하나 있다.** 1라운드 지적을 받아 "P2 6건 중 5건만 닫는다"로
낮췄던 것을, 뮤테이션 실측이 6건 전부 검증된다고 말해서 되돌렸다. 근거는 리뷰에 대한
반론이 아니라 숫자이며 §6에 있다.

## §1. 라운드 1 — proposal-freeze, 적대적 교차모델

- 날짜: 2026-08-09
- 보이스: 적대적 staff engineer 관점 (Codex, `model_reasoning_effort=high`, read-only 샌드박스)
- 대상: 초판 4문서 + `internal/journal/outbox.go`, `internal/obs/notifier.go`,
  `internal/execgw/retry.go`, `internal/flatten/flatten.go`
- 지시: 각 주장을 CONFIRMED-SAFE / DEFECT로 판정하고 구체적 실패 시나리오와 file:line을
  요구. 칭찬 금지.
- 판정: **BLOCK**

위험 등급이 High-risk(원장·알림·진입 게이트)이므로 `docs/WORKFLOW.md` §리뷰 게이트가
요구하는 적대적 Eng 관점을 이 라운드가 충족한다.

### 수용한 지적

| # | 등급 | 지적 | 반영 |
|---|---|---|---|
| 2 | P2 | 본문을 덮으면서 `delivered_at`을 남기면 그 행은 "지금 담고 있는 내용이 T1에 전달됐다"고 말한다 — 거짓이고 복원할 곳도 없다 | D1에서 `delivered_at`을 **지우는 쪽으로 뒤집음.** "세 번째 범주"라는 초판의 구분을 폐기 |
| 3 | P2 | lock order는 안전하지만 D2가 말한 운영자 해제 경로가 **배선되어 있지 않다** — `Notifier.Acknowledge`에 비테스트 호출자 0건 | 직접 확인함(0건). D2에서 그 주장을 삭제하고, 실질적 탈출구가 재시작임을 명시. §8에 미배선 사실을 남김 |
| 4 | **P1** | 승격을 금지하면 래치가 메모리뿐이라 재시작이 유일한 대응을 지운다. claim 실패 시에는 알림 행조차 없다 | **R2를 뒤집음.** `notifyCritical` 오류 분기에서 `escalate` 호출. spec 요구사항도 금지→요구로 재작성 |
| 6 | P2 | DB 효과 관측도 결정론이 아니다 — 스케줄러가 Acknowledge goroutine을 안 돌리면 통과한다 | 인정. D4를 "닫는다"에서 **"개선만 하고 닫지 않는다"**로 바꾸고 §8로 이관 |
| 7 | P2 | "P2 6건을 닫는다"고 쓰면서 그중 하나를 범위 밖으로 미룬다 | proposal·tasks 첫 문단을 **"5건을 닫고 1건은 개선만"**으로 정정 |
| 8 | **P1** | tasks 2.10이 2.1~2.9 전부의 RED를 요구하지만 2.4·2.5·2.7~2.9는 초록으로 도착한다. 구현자는 RED 증거를 지어내거나 게이트를 미완으로 둘 수밖에 없다 | §2를 **2a(지금 실패해야 하는 것)**와 **2b(초록으로 도착하는 것)**로 분리. 2b의 가치는 뮤테이션과 반복 실행으로 증명 |
| 9 | P2 | `last_attempt_at`이 이전 에피소드에서 남는다 — `attempts=0`인데 T1에 시도한 행 | D1 표와 UPDATE에 `last_attempt_at = NULL` 추가 |
| 10 | P2 | `payload`를 갱신해도 `Flush`는 `Title`·`Body`만 보내므로 backlog 경로로 문맥이 안 간다 | spec 시나리오를 **title/body로 좁힘.** 비대칭을 §8에 기록 |
| 11 | P2 | Pre-Edit 선언이 GREEN **뒤**에 있고 FLM 완성이 구현 뒤로 미뤄져 있다 | Pre-Edit을 §3으로 올려 GREEN(§4) 앞에 둠. FLM·BTM 6쌍은 **이 리뷰 전에** 이미 완성 |

### 확인만 된 지적 (CONFIRMED-SAFE)

| # | 주장 | 확인 방법 |
|---|---|---|
| 1 | `attempts` 0 복귀는 재시도 자격·예산에 영향이 없다 | 저장소 전역 검색으로 `attempts` 컬럼을 필터·정렬·판단에 쓰는 비테스트 코드가 없음. 예산은 메모리의 `n.Attempts`(`notifier.go:306`) |
| 5 | R3의 소스 읽기가 옳고 문서만 고치는 것이 맞다 | `state=QUARANTINED, remindAfter=0`이 `default`로 직행함을 확인. 재무장을 막으면 `MarkAlertDelivered`가 행을 거부해 반복 publish로 되돌아감 |

### 반영하지 않은 것

없음. 9건 전부 문서에 반영했고 2건은 확인으로 종결했다.

## §2. 이 라운드가 바꾼 안전 결론

초판 D2의 논증은 **"잠금은 싸고 승격은 비싸다"**였다. 그 전제가 사실을 잘못 읽었다.

`EntryGate.Block`(`execgw/retry.go:498`)은 `g.latches` 맵에 쓰고 그것은 **메모리에만**
있다. claim이 실패한 경우 원장에는 알림 행조차 없다. 따라서 초판의 대응은 재시작 한
번으로 완전히 사라지고, 운영자가 아무것도 받지 못한 채 진입이 다시 열린다.

결정적인 것은 코드가 이미 그 말을 하고 있었다는 점이다. `escalate`의 주석
(`notifier.go:264-271`)이 승격 실패 시 잃는 것을 **"재시작을 넘겨 살아남는 부분"**이라고
이름 붙이고 있다. 초판은 그 부분을 자발적으로 포기하면서 그것을 보수적 선택이라고
불렀다.

같은 주석이 반대 방향의 걱정도 해소한다: `escalate`는 실패해도 오류를 버블링하지 않고
error 로그를 남긴다. 원장이 죽어 있으면 승격도 실패하지만 **그 실패가 기록된다.**
지금 그 자리에는 아무것도 없다.

## §3. 방법에 대한 기록

이 라운드에서 값을 준 것은 세 가지다.

1. **AST를 문서보다 먼저 만든 것.** R3의 판정(`B4@256`이 `B3@255`의 자식, `default`는
   형제)은 손으로 읽었으면 "default가 `remindAfter`를 무시한다"에서 멈췄을 것이다.
   교차모델 리뷰도 같은 구조 읽기로 CONFIRMED-SAFE를 냈다.
2. **RED 커버리지를 측정하고 인용한 것.** `claimAndDeliver` `B1@227`이 `count=0`이라는
   사실 — 오늘 어떤 테스트도 claim 실패를 지나지 않는다 — 이 R2의 필요성을 주장이 아니라
   측정으로 만들었다. `flatten.Saga.event`의 `Notify` 줄도 `count=0`이다.
3. **리뷰에게 "칭찬 금지"와 "구체적 실패 시나리오"를 요구한 것.** 9건 중 2건이 P1이었고
   둘 다 문서를 읽는 것만으로는 나오지 않는 종류였다 — 하나는 `EntryGate`의 저장 위치,
   하나는 task 목록의 내부 모순이다.

## §4. 남은 위험 (완료 보고에 다시 적는다)

> 이 절의 첫 항목은 1라운드 직후 "`Acknowledge` 상호배제는 여전히 미검증"이었다.
> **§6의 뮤테이션 실측이 그것을 뒤집었다** — 구판·신판 모두 유휴 50/50, 부하 30/30으로
> 뮤턴트를 죽인다. 아래는 정정된 목록이다.

- **세 동시성 테스트의 통과 경로는 기회 의존적이다.** 검증은 뮤테이션 실측이지 구성이
  아니다. 경합 상대가 스케줄되지 않으면 아무 일도 일어나지 않고, 그때 값은 배제와
  구분되지 않는다. 결정론을 구성으로 얻으려면 프로덕션 seam이 필요하다(design D4).
- **운영자 해제 경로(`Notifier.Acknowledge`)가 미배선이다.** a097 이후 claim 실패로 잠긴
  게이트는 승격 기록은 남기지만, 사람이 푸는 경로는 여전히 없다. 실질적 탈출구는 재시작
  이고, 승격이 durable한 지금은 재시작이 그것을 지우지 못한다 — 즉 **운영자가 개입할
  경로 없이 진입이 막힌 상태가 지속될 수 있다.** 별도 change가 필요하다.
- **일시적 `SQLITE_BUSY` 한 번이 durable한 운영 모드 전환을 일으킬 수 있다.** 방향은
  보수적이지만 운영 부담이며, 위 항목과 겹치면 부담이 커진다. tasks §7의 배포 후 실측에서
  거짓 승격이 없는지 확인해야 한다.
- **선택 배선이 비면 이 change의 보호가 미치지 않는다.** `Gate`·`AccountRef`가 없는 조립은
  차단할 게이트도 기록할 계정도 없다. 엔진 조립은 셋 다 배선하지만
  (`internal/app/engine/exitwiring.go:73-78`), 다른 조립 지점이 생기면 같은 확인이 필요하다.
- **`n.mu` 아래의 로깅은 a074부터의 성질이다.** `Notify`를 재호출하는 `io.Writer`를 꽂으면
  교착한다. 그런 writer는 이 저장소에 없지만 금지 장치도 없다(§5 지적 2).

## §5. 라운드 2 — 구현 후 독립 검증, 판정 BLOCK

- 날짜: 2026-08-09
- 보이스: 독립 리뷰어 관점 (Codex, `model_reasoning_effort=high`, read-only 샌드박스).
  구현을 만든 컨텍스트와 분리된 패스 — `docs/WORKFLOW.md` §9 요구.
- 대상: working tree diff(`outbox.go`, `notifier.go`, `a096_one_send_per_condition_test.go`)와
  새 테스트 3파일, 그리고 이 change의 계약 문서 전부
- 판정: **BLOCK — P2 4건, P3 1건**

**절차상 내 실수 하나를 먼저 적는다.** 이 리뷰를 뮤테이션 실험과 **동시에** 돌렸다.
리뷰어가 `notifier.go`가 실행 중에 바뀌는 것을 감지하고 "의도된 뮤턴트를 구현으로
보고하지 않도록" 스스로 대기했다. 결과는 유효했지만, 흔들리는 트리에 리뷰를 거는 것은
리뷰를 무효로 만들 수 있었다. **독립 검증은 트리를 고정한 뒤에 건다.**

### 수용한 지적

| # | 등급 | 지적 | 반영 |
|---|---|---|---|
| 1 | P2 | `Gate`·`Log`·`AccountRef`가 모두 비면 claim 실패의 결과가 전부 사라진다. spec의 무조건 SHALL과 어긋난다 | spec에 **"이 요구사항이 미치는 범위는 배선된 것까지"** 문단과 시나리오 추가. 오류 반환만은 배선과 무관하다는 SHALL NOT 유지. `TestAFailedClaimWithNothingWiredStillReports` 신규 |
| 3 | P2 | 새 배타 테스트들이 여전히 경과 시간에 의존하고, 무한 채널 대기로 전체 타임아웃까지 멈출 수 있다 | 시간 의존은 **측정으로 답한다**(§6). 무한 대기는 실제 결함이라 세 곳에 10초 상한과 문장을 붙였다 |
| 4 | **P2** | **FLM이 구현과 반대를 말한다** — `delivered_at`을 "의도적으로 남긴다"고 적혀 있고, `notifyCritical` FLM은 claim 실패 승격이 "없는 것이 의도"라고 적혀 있다 | 둘 다 고쳤다. 각각 왜 뒤집혔는지 함께 남겼다 |
| 5 | P3 | 몇몇 테스트는 a097 변경을 고정하지 않는다. 특히 `TestAFailedClaimStillReturnsItsError`는 되돌려도 통과한다 | 그 테스트의 주석을 **"pin이 아니라 guard"**로 고쳐 적었다. 되돌려도 통과하는 것이 의도이며, 왜 그런 테스트가 필요한지 적었다 |

### 사실 확인 후 반영하지 않은 것

| # | 지적 | 확인 결과 |
|---|---|---|
| 2 | `n.Log.Error`를 `n.mu` 아래에서 부르므로, `Notify`를 재호출하는 동기 writer를 꽂으면 교착한다 | **a097이 만든 노출이 아니다.** `deliver`는 `n.mu`를 잡은 채 이미 `n.Log.Error`를 세 번 부른다(`notifier.go:374`, `:385`, `:399`). 그 함수의 PRECONDITION 주석(`:336`)이 호출자의 잠금 보유를 명시한다. 즉 "잠금 아래 로깅"은 a074/a096b부터의 성질이고 a097은 같은 패턴을 하나 더한다. 시나리오도 `Notify`를 재호출하는 `io.Writer`를 요구하는데 이 저장소에 그런 writer는 없다. **기록만 하고 여기서 고치지 않는다** |

### 리뷰어가 확인해 준 것

- `delivered_at`·`last_attempt_at`을 지워도 현재 독자를 깨뜨리지 않는다. PENDING이
  `latestStamp` 앞에서 단락되고, `PendingAlerts`·`UndeliveredCount`는 state로 거르며,
  `LookupAlert`는 nullable로 스캔한다. `DELIVERED`인데 `delivered_at`이 NULL인 상태는
  공개 경로로 도달 불가다(`MarkAlertDelivered`가 둘을 한 번에 쓴다).
- **owed였던 알림이 unowed가 되지 않는다.** owed 판정이 재무장 UPDATE보다 먼저 일어난다.
- **`EntryGate.Block`은 청산을 막지 않는다.** 노출을 늘리는 계획만 게이트를 본다 —
  취소는 `raisesExposure=false`이고 매도는 노출 증가로 분류되지 않는다
  (`internal/execgw/gateway.go:377`, `:415`, `:855`). 안전 불변식 §0-3 무영향 확인.

## §6. 측정이 리뷰를 두 번 정정했다

a096 2라운드 리뷰가 낸 P2 ⑤·⑥은 테스트에 대한 것이었고, 둘 다 **처방을 그대로 넣지
않았다.** 넣기 전에 쟀기 때문이다. 방법은 하나다 — 지키려는 잠금을 제거한 뮤턴트를
만들고 각 변형이 몇 번 죽이는지 센다.

### P2 ⑤ — 진단은 맞고 처방은 틀렸다

뮤턴트: `claimAndDeliver`의 `n.mu.Lock`/`Unlock` 제거. `GOMAXPROCS=1`, 각 100회.

| 변형 | 탐지 |
|---|---|
| 원본 (barrier 없음) | 96 / 100 |
| **barrier만 추가 (리뷰의 처방)** | **91 / 100 — 개선 없음** (차이 1.4σ) |
| 차단 publisher + barrier | **100 / 100** |

정상 코드 오검출 0/30. 리뷰가 본 오통과는 실재했다(4%). 그러나 그것을 고치는 것은
goroutine을 함께 출발시키는 것이 아니라 **첫 전송을 `Publish` 안에 붙잡아 경합 창을
구성으로 여는 것**이었다. barrier는 비용이 없어 남겼고, 탐지에 기여하지 않는다고 적었다.

### P2 ⑥ — 우려가 재현되지 않았다

뮤턴트: `Acknowledge`의 `n.mu.Lock`/`Unlock` 제거.

| 변형 | 유휴 (n=50) | 20코어 부하 (n=30) |
|---|---|---|
| 구판 (50ms `time.After`) | 50 / 50 | 30 / 30 |
| 신판 (원장 효과 + 순서) | 50 / 50 | 30 / 30 |

**구판도 부하 아래에서 뮤턴트를 한 번도 놓치지 않았다.** 따라서 이 잠금은 검증되며,
신판의 이점은 측정된 탐지율이 아니라 판정의 종류다 — 지속 시간이 아니라 원장에 남은
효과를 본다. 그 차이가 값으로 나타나지 않았다는 사실을 그대로 적는다.

### Flush의 잠금 (P2 ④)

뮤턴트: `Flush`의 `n.mu.Lock`/`Unlock` 제거 → `TestFlushCannotPublishBesideASend`
**20/20 실패**. 같은 뮤턴트에서 obs 스위트의 다른 **기존** 테스트는 하나도 죽지 않았다.
이 테스트가 그 잠금을 단독으로 고정한다.

### 그래서 P2 6건은 전부 닫힌다 — 다만 두 건은 리뷰가 쓴 방식이 아니다

세 잠금 모두 **통과 경로는 기회 의존적**이다. 검증은 구성이 아니라 뮤테이션이며, 세 곳에
같은 기준을 적용한다. 한쪽에 20/20을 근거로 완결을 주고 다른 쪽에 50/50을 근거로 미완을
주면 그것이 두 기준이다.

## §7. 실측

| | |
|---|---|
| `go vet ./...` | 통과 |
| `go test -race ./internal/obs/` | 통과 82.2s |
| `go test ./internal/execgw/` | 통과 38.3s |
| `go test ./internal/flatten/` | 통과 5.1s |
| 커버리지 journal | RED **75.0%** → GREEN **75.0%** (새 분기 없음) |
| 커버리지 obs | RED **85.4%** → GREEN **86.7%** |

분기 주장은 커버리지 프로파일로 직접 대조했다:

- `claimAndDeliver` claim 실패 분기: **미진입(count=0) → 진입.** RED에서 이 저장소의 어떤
  테스트도 지나지 않던 경로다.
- `notifyCritical` B3(claim 오류 → 승격 시도): **미진입 → 진입.**
- `notifyCritical` B4(전송 실패 → 승격): RED·GREEN 모두 진입 — 전송 실패 경로를 건드리지
  않았다는 확인.
- `claimOwed` 8분기: RED와 GREEN이 완전히 동일(B5만 미진입) — 이 함수를 편집하지 않았다는
  확인.
- `flatten.Saga.event`의 `Notify` 호출 줄: RED·GREEN 모두 **미진입**. 오류를 버리는 그 줄은
  flatten 스위트에서 한 번도 실행된 적이 없다.

**예측 하나가 틀렸고 그대로 적었다.** BTM 초판은 "obs의 claim 실패 테스트가 닫힌 DB로
`ClaimAlertForDelivery` B3을 지나므로 journal 커버리지가 0→1로 바뀔 것"이라고 적었다.
GREEN에서도 0이다 — Go 커버리지는 **테스트 대상 패키지 기준**으로 집계되므로
`./internal/obs/` 실행 중의 journal 코드는 journal 프로파일에 기록되지 않는다.
"진입"으로 적고 넘어갔다면 그것이 a096 3판이 지적한 미측정 수치가 됐을 것이다.

## §8. 게이트와 sdd-sync advisory (task 6.5·6.6)

`make gate CHANGE=a097-a-re-armed-alert-is-a-new-episode` → **GATE PASS 8/8** (`GATE_EXIT=0`).
`make sdd-sync` → `SYNC_EXIT=0`.

**1차 실행은 5/8에서 실패했다.** `make sdd-check`의 PM 검사가
`generated tracker stale: 00-master-tracker.md / 01-active-change-map.md`를 냈다 —
tracker를 생성한 **뒤에** 46개 task 상자를 전부 체크해서 생성물이 입력보다 낡았다.
2차는 PM 재생성 → `sdd-sync` → `gate`를 **한 셸 호출**로 묶어 통과했다
(중간에 파일이 바뀌면 CodeGraph fingerprint가 다시 깨지기 때문이다).

**advisory 실패를 그대로 적는다 — 침묵한 생략 금지.** `make sdd-sync`는 `SYNC_EXIT=0`이지만
아래 넷은 실패했고, 전부 advisory 계층이라 hard evidence를 대체하지 않는다.

| advisory | 출력 | 뜻 |
|---|---|---|
| GBrain source probe | `schema read failed` | 소스 등록 확인 실패 |
| GBrain sync | `busy; keeping its previous freshness` (owner pid=456 `gbrain serve`) | 의미 색인이 **갱신되지 않았다** |
| sdd-hook / commit-event | `warning: offline` | 이력 훅이 원격에 못 붙었다 |
| context-graph | `CCG adapter failed; its checkpoint was not advanced` | CCG 체크포인트 미전진 |

넷 중 어느 것도 `make sdd-check`의 index-freshness(=CodeGraph hard evidence)를 건드리지
않는다. 그 검사는 통과했다: `[index-freshness] CodeGraph hard-evidence index matches the worktree`.
