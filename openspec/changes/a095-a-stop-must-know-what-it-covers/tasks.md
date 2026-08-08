# a095 · tasks

- **Change**: `a095-a-stop-must-know-what-it-covers`
- **위험 등급**: **High-risk** — 손절 관련 알림 등급과 총위험 보고. §0.3 적용.
- **base-commit**: `ec29dc72c0fd589daa2069ccf26bad26baeb2a04`

## 0. 게이트 선행

- [x] 0.1 `base-commit.txt` 고정
- [x] 0.2 `openspec validate a095-a-stop-must-know-what-it-covers --strict` — **통과**
- [x] 0.3 **AST 산출물이 문서보다 먼저** — 함수 9개, 분기 98개
- [x] 0.4 `check_analysis.py --change a095-…` — **통과**(`evidence complete`).
      두 Markdown은 **측정으로** 채웠다: 조건은 소스 원문, 창의 호출·return은
      `ast.json` 좌표, 진입 여부는 `go test -covermode=set` 프로파일이다
- [ ] 0.5 **proposal-freeze 리뷰**(적대적 Eng 필수) → `review.md`.
      **교차 모델을 여기서 지킨다** — a092가 여섯 라운드, a094가 한 라운드 미충족이다

## 1. 산출물 (완료 — 문서보다 먼저)

- [x] 1.1 `obs.SeverityOf` (분기 1 · return 2)
- [x] 1.2 `obs.Notifier.Notify` (분기 1 · return 2)
- [x] 1.3 `obs.Notifier.publishBestEffort` (분기 2 · return 1)
- [x] 1.4 `engine.ReconcileDriver.checkExternalIncrease` (분기 3 · return 3)
- [x] 1.5 `engine.ReconcileDriver.alertUnmanaged` (분기 6 · return 1)
- [x] 1.6 `journal.Journal.OpenExitState` (분기 20 · return 15)
- [x] 1.7 `journal.resetExitStateForReadoptTx` (분기 6 · return 7)
- [x] 1.8 `journal.Journal.ApplyPositionAdjustment` (분기 27 · return 17)
- [x] 1.9 `exitpolicy.EvaluateLadder` (분기 32 · return 23)
- [x] 1.10 **Branch Test Map** — 위 9개. **분기 98개 중 진입 62 · 미진입 27 · 자체 블록 없음 9**
      (`go test ./internal/{obs,app/engine,journal,exitpolicy}/... -count=1 -covermode=set`)

      | 함수 | 분기 | 미진입 | 블록없음 |
      | --- | --- | --- | --- |
      | `obs.SeverityOf` | 1 | 0 | 0 |
      | `obs.Notifier.Notify` | 1 | 0 | 0 |
      | `obs.Notifier.publishBestEffort` | 2 | 0 | 0 |
      | `engine.ReconcileDriver.checkExternalIncrease` | 3 | **2** | 0 |
      | `engine.ReconcileDriver.alertUnmanaged` | 6 | 0 | 1 |
      | `journal.Journal.OpenExitState` | 20 | 4 | 1 |
      | `journal.resetExitStateForReadoptTx` | 6 | **4** | 1 |
      | `journal.Journal.ApplyPositionAdjustment` | 27 | 6 | 5 |
      | `exitpolicy.EvaluateLadder` | 32 | 11 | 1 |

      **가장 무거운 실측**: `checkExternalIncrease`는 분기 3개 중 **2개가 미진입**이다 —
      `B1 :442`(프로세스당 1회 억제)와 `B2 :446`(편입 기록 없음의 조용한 반환).
      **a095의 R2가 정확히 그 둘 위에 얹힌다.** 2.3·2.4가 먼저 덮는다.

      **둘째로 무거운 것**: `resetExitStateForReadoptTx`는 분기 6개 중 **4개 미진입**이고
      넷 다 오류 처리다. 손절가를 덮어쓰는 유일한 자리의 오류 경로가 거의 시험되지
      않았다 — **D4가 자동 갱신 경로를 지금 붙이지 않는 이유의 하나다**
- [ ] 1.11 **발신 자리 넷의 문구 대조** — `adoption.go:418`·`:456`,
      `exitloop.go:1501`, `exitwiring.go:104`가 실제로 **같은 사실**을 말하는지.
      다른 사실이면 등급을 함께 올리는 것이 틀렸다(design D7)
- [ ] 1.12 **a091과의 텍스트 충돌 확인** — 둘 다 `criticalEvents` map을 늘린다.
      논리 의존은 없으나 병합 순서를 정한다

## 2. R1·R2 — 등급과 억제 (D1·D2)

- [ ] 2.0 **Pre-Edit 선언** — `internal/obs/event.go`(map only),
      `internal/app/engine/adoption.go` `checkExternalIncrease`
- [ ] 2.1 **RED** — `SeverityOf(EventExitPositionUnmanaged)` == `SeverityCritical`
- [ ] 2.2 **RED** — 그 이벤트가 `Notify`에서 **B1 `:111`을 타지 않고** `notifyCritical`로
      가서 `alert_outbox` 행을 만든다. 전달 실패 시 재시도가 붙는다
- [ ] 2.2a **RED** — `CriticalEvents()`의 고정 목록이 19종이 된다(기존 18 + 1).
      **다른 18종의 등급은 무변화**
- [ ] 2.3 **RED (미진입 분기 B2 `:446`)** — 편입 기록이 **없는** 포지션에서
      `checkExternalIncrease`가 조용히 반환하지 않는다. 그 포지션은 **보호 상태 없음**으로
      보고된다. **오늘 이 분기를 밟는 시험이 하나도 없다**
- [ ] 2.3a **RED** — 진짜 조회 오류(DB 오류)는 종전대로 조용히 반환한다.
      **「기록 없음」과 「조회 실패」를 가른다** — 후자를 알리면 DB 장애가 알림 폭주가 된다
- [ ] 2.4 **RED (미진입 분기 B1 `:442`)** — 수량 증가를 알린 뒤 **같은 프로세스에서**
      수량이 또 늘면 **다시 알린다**. 수량이 그대로면 다시 알리지 않는다.
      **오늘 이 분기를 밟는 시험이 하나도 없다**
- [ ] 2.5 **RED** — `alertUnmanaged`의 `d.unmanaged` 억제는 **무변화**(프로세스당 1회).
      무보호는 수량과 무관한 상태이고 critical의 outbox 재시도가 반복을 대신 책임진다
      (design D6의 결정을 시험으로 고정)
- [ ] 2.6 **RED** — `alertUnmanaged`의 사유 switch(B3~B6) **무변화**
- [ ] 2.7 **GREEN** — `criticalEvents`에 `EventExitPositionUnmanaged` 한 줄 +
      `checkExternalIncrease`의 억제 키를 수량 기준으로 + B2 분기 분리
- [ ] 2.8 **§0.3 확인** — 이 change는 판정·제출 경로에 새 호출을 넣지 않는다.
      diff로 보인다

## 3. R3 — 총위험 (D3)

- [ ] 3.1 **RED** — 알림이 `(평단 − **유효손절**) × 현재수량`을 담는다.
      **유효손절은 `baseline_price`이지 `initial_stop`이 아니다.**
      475150 형태(평단 58,000 · 유효손절 57,900 · 32주)에서 **3,200**이 나온다.
      **`initial_stop`(56,163)을 쓰면 58,784가 나오고 그것은 18.4배 과대다** — 1판이
      그렇게 적었고 1라운드가 잡았다
- [ ] 3.2 **RED** — `exit_states`의 `entry_price`·`initial_stop`·`initial_risk`·
      `policy_*`에 **어떤 쓰기도 일어나지 않는다.** 구조로 고정
- [ ] 3.3 **RED** — 평단이 `firstNonEmpty`(`position_adjustments.go:312`)로 이어받은
      값이면 그 불확실성이 보고에 담긴다
- [ ] 3.4 **RED** — 평단이나 손절가를 읽지 못하면 총위험을 **추측하지 않는다.**
      "미측정"으로 보고하고 알림 자체는 계속 나간다
      (수를 못 만든 것이 알림을 막으면 안 된다)
- [ ] 3.4a **RED** — 유효 손절이 평단 **위**면 그 수는 음수다. **「위험」이 아니라
      「고정된 이익」으로 보고한다.** 272210(−7,220)·TSLA가 그 경우다.
      음수를 위험으로 적으면 운영자가 정반대로 읽는다
- [ ] 3.5 **RED** — `resetExitStateForReadoptTx`의 호출자는 **여전히 하나**다
      (`positionpolicy.ActionReadopt`). a095는 새 호출자를 만들지 않는다.
      **구조로 고정** — D4가 자동 갱신을 미루는 근거가 이 불변식이다
- [ ] 3.6 **GREEN** — 산출과 필드 추가. 읽기와 산술뿐

## 4. 하지 않는 것을 고정한다 (D4·D6)

- [ ] 4.1 **RED (§6)** — 평단이 **내려간** 포지션에서 유효 손절가가 **내려가지 않는다.**
      실측 3종목(066570·080220·272210)이 그 경우다
- [ ] 4.2 **RED** — `EvaluateLadder`의 산출이 **무변화**. 모든 선은 계속 `entry_price`에서 나온다
- [ ] 4.3 `issues.md`에 **래칫의 선행 조건 셋** 기록 — 쓰기 경로 단일성 · 하향 거부의
      기록 · 진입 기준 이동의 사다리 파급 계측. **StockOS의 SHALL과 그것이 dormant라는
      사실을 함께 적는다**

## 5. 실측 재생

- [ ] 5.1 2026-08-07 원장 상태를 fixture로 재생 — 010170(무보호 30주)이 critical로
      보고되는지, 475150(3→32주)이 총위험 **3,200**(유효손절 57,900 기준)과 함께 보고되는지
- [ ] 5.2 `alert_outbox`에 `exit.position_unmanaged` 행이 **생기는지**
      (오늘 0행인 것이 이 change의 출발점이다)
- [ ] 5.3 결과를 `issues.md`에 기록. **a091·a094와의 상호작용**을 명시한다

## 6. 게이트

- [ ] 6.1 `go test ./... -count=1 -race` 회귀 0
- [ ] 6.2 **§0.3 확인 — 호출 수가 아니라 체류 시간으로 잰다**(1라운드 §1.4).
      1판은 *"제출 경로에 새 호출이 0개임을 diff로 보인다"*고 약속했고 **호출 수 diff는
      이것을 구조적으로 볼 수 없다** — 호출이 늘지 않고 **기존 호출의 등급이 바뀔** 뿐이다.
      `alertUnmanaged`가 `observe` **전에** 있다는 사실(`exitloop.go:515` vs `:443`)과
      `deliver`가 `n.mu`를 잡는 구간을 **실측**한다
- [ ] 6.2a **RED** — `notifierAlerter.ExternalPositionFound`(`exitwiring.go:98-103`)의
      오류 전파를 다룬다. `Notify`는 critical outbox 쓰기 실패에서만 오류를 반환하므로
      (`notifier.go:104-107`) **`EnqueueAlert` 실패가 이제 외부 포지션 대사를 실패시킨다**
      (`internal/reconcile/external.go:278-295`). 종전에는 불가능했다.
      **그리고 그 발신 자리만 latch가 없다** — 적용되는 fold마다 발화한다(`external.go:277-281`)
- [ ] 6.3 **§0.4 확인** — 새 브로커 조회 **0건**. 총위험은 이미 읽은 값의 산술이다
- [ ] 6.4 **토글 OFF 동등성** — 이 change는 토글을 도입하지 않는다.
      도입하지 않았음을 명시한다(`not-applicable` 아님 — 무도입)
- [ ] 6.5 **`openspec validate --strict`의 한계** — ADDED만 쓰므로 이 change에는
      정본 치환 위험이 없다. **그 사실을 확인하고 적는다**(a094 1라운드 차단 1의 교훈)
- [ ] 6.6 FLM·AST **재생성** (구현 후) + `check_analysis.py` 통과
- [ ] 6.7 `make sdd-sync` → `make sdd-check`
- [ ] 6.8 **격리 worktree에서** `make gate CHANGE=a095-a-stop-must-know-what-it-covers`
- [ ] 6.9 **독립 리뷰**(구현과 분리된 컨텍스트). **교차 모델을 지킨다**
- [ ] 6.10 PM 동기화 → `openspec archive`

## 7. 배포와 운영 — 사람이 승인한다

- [ ] 7.1 배포 전 `main`과 **SchemaVersion 대조** (낮으면 엔진이 조용히 죽는다)
- [ ] 7.1a **공시** — alert transport가 죽고 무보호 보유가 있으면 **엔진이 신규 진입을
      멈춘다**(`ModeEntryBlocked`, `notifier.go:216-218`, `Acknowledge`로만 해제).
      **손절·익절·취소는 막지 않는다.** 그것이 a095가 선택하는 교환임을 운영자가
      미리 알아야 한다(design D1의 근거 셋)
- [ ] 7.2 **배포 직후 알림 폭주 여부를 실측한다.** 열린 포지션 6건 중 최소 2건이
      즉시 critical로 울린다(010170 무보호 · 475150 수량 증가). 그것이 **의도한 동작**임을
      미리 적어 두고, 그 이상이 울면 억제 키를 재검토한다
- [ ] 7.3 **이 change는 현재 열린 포지션을 소급 보호하지 않는다.**
      **010170의 30주는 배포 후에도 무보호이며, 배포가 하는 일은 그것을 보이게 하는 것뿐이다.**
      배포 전까지 사람이 처리한다
- [ ] 7.4 배포 후 `alert_outbox`에 `exit.position_unmanaged` 행이 실제로 생기는지 확인

## 선후 관계

```text
a095 (이 change) ── 보호 범위의 보고
   │
   ├─ a091 한 주도 못 판 손절      **같은 map을 늘린다.** 논리 의존 없음. 병합 순서만
   ├─ a094 손절이 길을 치운다       독립. 같은 종목이 등장하나 원인이 다르다(409 오분류)
   ├─ a089 나가지 못한 손절을 센다  독립. 계측
   └─ a092 알림이 손절을 잡지 않는다 **선행 조건.** 아래 참조
```

**a095는 a092 뒤에 간다(1라운드 §1.7 정정).** a092가 존재하는 이유가 손절 경로의
**동기 알림 체류를 유계로 만드는 것**이고, a095 R1은 **그 경로를 타는 알림의 모집단을
늘린다.** 1판은 *"a092와 겹치지 않는다 — 알림 체류"*라고 적었고 **거꾸로였다.**

`alertUnmanaged`는 `workingSet` 안 `exitloop.go:515`에서 불린다 — **`observe`(`:443`)와
포지션 판정(`:453-465`) 전이다.** critical이 되면 `deliver`가 `n.mu`를 잡은 채
재시도 예산(`DefaultCriticalAttempts=3` × `DefaultRetryDelay=2s`)을 돈다.

**a091과는 합산된다.** 둘 다 같은 `criticalEvents` map에 행을 더하고, 그것은 텍스트
충돌이 아니라 **직렬화된 `deliver` 예산에 두 종류가 더 들어오는 것**이다.
어느 계획도 그 합을 소유하지 않으므로 **먼저 가는 쪽이 합을 측정한다.**

**후속(별도 change)**: 불타기 방향의 손절 상향 래칫. 선행 조건 셋은 `issues.md`와
`specs/exit-policy/spec.md`에 SHALL로 있다.

## 안전 불변식 확인

| 불변식 | 이 change에서 |
| --- | --- |
| §1 사람 승인 없는 LIVE 주문 side effect 금지 | 이 change는 **주문을 내지 않는다.** 알림과 산술뿐. 배포는 7절에서 사람이 승인 |
| §2 `mutating: true` 자동 실행 금지 | 준수 |
| §3 토글 OFF는 upstream과 동일 | **토글을 도입하지 않는다** |
| §4 손절 즉시성을 약화·지연하지 않는다 | 6.2가 제출 경로에 새 호출 0개임을 보인다. 알림 경로만 건드린다 |
| §5 High-risk 경로 | 손절 알림·보호 범위. Pre-Edit 선언은 2.0 |
| §6 보수 방향만 | **손절가를 내리는 변경을 명시적으로 거부한다**(4.1). 등급은 올리는 방향, 알림은 늘리는 방향, 총위험은 **과소보고를 없애는** 방향 |
| §7 운영 토글 flip과 live 검증은 사람이 | 7절 |
| §8 시크릿·계좌 개인정보 저장 금지 | 원장 인용은 종목코드·수량·가격·시각까지. 계좌번호·잔고 절대액 없음 |
