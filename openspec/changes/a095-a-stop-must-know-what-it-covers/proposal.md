# a095 · 손절은 자기가 무엇을 덮는지 알아야 한다 — 4판

- **Feature**: `FEAT-TOS-009` — Exit line truth and position policy lifecycle
- **Story**: `STORY-TOS-a095`
- **Spec**: `exit-policy` · `engine-safety`
- **위험 등급**: **High-risk** — 무보호 보고의 등급과 진입 차단 도달. §0.3·§0.6 적용.
- **base-commit**: `027163575e78482cdd4d8129f027e928b769ae87` (3판 재고정, `0b17784e`)
- **판**: 3판(2026-09-27). 1판·2판 본문은 git 이력에 있다(2판 = `2fbdcd78` 시점).

> **작성 순서.** 이 문서의 분기 주장은 전부 `analysis/function-logic/`의 AST 산출물에서 왔다.
> 3판은 base `02716357`에서 번들 **21개**를 이 문서보다 먼저 갖췄다 — 새로 뽑음 12(호출자 `judgeHoldings` ·
> `ExitObserver.workingSet` · `notifierAlerter.ExternalPositionFound` 포함) · 줄 이동으로 다시 뽑음 4 ·
> 소스 해시가 그대로여서 산문만 고침 5. 분기 표는 `analysis/harness/render_bundles.py`가
> `ast.json`과 커버리지 프로파일(`analysis/harness/coverage/`)에서 기계로 만든다.
> 손으로 적은 분기 주장은 없다. 결정이 덮지 않는 지점은 **`[비움 — Qn]`**으로 비워 두고
> 끝의 「열린 질문」에 올린다.

## 0. 3판의 입력 — 사용자 결정 (2026-09-25) 원문

**출처와 한계.** 사용자가 대화에서 쓴 날문장은 저장소에 없다. 아래는 그 결정을 기록한 두 자리를
**글자 그대로** 옮긴 것이다 — `review.md` §2.16 표와, 그것을 착지시킨 커밋 `2fbdcd78`의 메시지 ③.

**물음** (`review.md` §2.11, 2라운드가 올린 것):

> 1. **a095의 착지를 a092에 묶을지.** a092가 「exit goroutine에서 동기 deliver 없음」을 약속·착지할 때까지 a095 R1을 동결할지, 아니면 a095가 자체 완화를 설계할지(3판 입력 4)
> 2. **알림 비활성·미설정 엔진과 `adoption.enabled=false`(기본값) 엔진에서 무관리 보유가 ENTRY_BLOCKED를 부르는 교환을 받아들일지.** 받아들이면 정본 exit-policy `:81`의 「false = 기존 동작」 SHALL을 바꾸는 결정이다(안전 불변식 3)
> 3. **범위를 옮길지.** 동기 포지션이 전부 닫힌 지금, 실측으로 드러난 진짜 구멍 — **엔진이 직접 연 포지션의 수량 증가 미검사**와 R2의 키 설계 — 으로 초점을 옮길지

**결정** (`review.md` §2.16, 원문 표):

> | §2.11 | 결정 | 3판이 받는 것 |
> |---|---|---|
> | 1 | **묶지 않는다** — a095 는 a092 와 독립 | 2.10 항목 4 의 ②: exit goroutine 에 critical Notify 를 **새로 두지 않는다**. 발신은 reconcile 쪽(`adoption.go` 경로)에서만 critical, exit 루프 자리(`exitloop.go:518` `alertUnmanaged`)는 normal 로 둔다. 옛 약속 사본(tasks 2.8 · §4 행 · D7) 제거 |
> | 2 | **거부** | 알림 off · `adoption.enabled=false` · `exclude_symbols` 에서 ENTRY_BLOCKED 를 부르지 않는다. 등급은 이벤트 종류가 아니라 **사실**로 — 운영자가 고른 상태는 critical 에서 뺀다. 정본 exit-policy `:81` SHALL 은 그대로(MODIFIED 아님, 안전 불변식 3) |
> | 3 | **수용 — 범위를 옮긴다** | (i) 엔진이 직접 연 포지션의 수량 증가 검사 — `adoption.go:108-111` 이 `Adopted()` 일 때만 `checkExternalIncrease` 를 부른다 (ii) R2 키 설계를 재알림 창 SHALL NOT 과 대조(B-P1-4) (iii) 두 발신 자리의 키 분리(B-P1-5). R2-B2 와 「편입 기록 없음 = 보호 없음」 SHALL·시나리오 삭제. 제목은 유지 |
>
> 3판 순서: FLM 재추출(호출자 `judgeHoldings` · `ExitObserver.workingSet` · `notifierAlerter.ExternalPositionFound` 추가, stale 4 재추출, base 재고정) → 문서 → 3라운드(두 보이스 결과 실제 수합 + 교차 모델). 저자 로트는 Opus 리셋(2026-09-26 19:00) 뒤.

**결정** (커밋 `2fbdcd78` 메시지 ③, 원문):

> ③ a095 §2.11: 1 독립(exit goroutine 에 critical Notify 안 둠) · 2 거부(기본 설정 엔진 ENTRY_BLOCKED 불가, 불변식 3) ·
>    3 수용(엔진 개설 포지션 수량 증가 미검사 + R2 키 + 발신 자리 키 분리로 범위 이동). review §2.16, tasks 0.5.

## 1. 2라운드 P0 넷을 어떻게 해소하는가

`review.md` §2.5의 차단 넷. 각 행의 「근거 번들」은 `analysis/function-logic/`의 디렉터리 이름이다.

| P0 | 2라운드가 잡은 것 | 3판의 해소 | 근거 번들 · 결정 |
| --- | --- | --- | --- |
| **P0-1** FLM이 함수 경계에서 멈췄다 | R2-B2가 「미편입·엔진 개설 포지션을 삼킨다」고 했으나 그 둘은 `checkExternalIncrease`에 오지 않는다. 엔진 개설 포지션의 수량 증가는 어디서도 검사되지 않는다. 델타가 설계보다 넓다 | **R2-B2와 그 SHALL·시나리오를 삭제**한다. 호출자 `judgeHoldings`의 번들을 새로 뽑아 사실을 고정했다: B7(`p.ExitEligible()`) 창이 `continue`하고 B8(`p.Adopted()`)만 `checkExternalIncrease`를 부른다. 엔진 개설 포지션의 수량 증가 검사를 **범위에 넣는다** — 비교 기준은 `[비움 — Q3]`. 델타는 결정된 것만 싣는다 | `reconciledriver.judgeholdings` · `reconciledriver.checkexternalincrease` · 결정 (3)(i) |
| **P0-2** R2 재알림을 outbox dedupe가 삼킨다 | 키 `…\|grown\|<posID>`에 수량이 없고, 같은 키의 전달된 행은 재알림 창 안에서 `ClaimSettled` | 이 P0는 **수량 증가 사실이 critical일 때만** 성립한다(`claimAndDeliver` B5는 critical 경로에만 있다). 그 알림의 본문은 스스로 *"늘어난 수량은 원래 수량 기준으로 산정된 손절의 보호를 받는다"*고 쓴다 — 무보호가 아니다. 그 사실의 종류·등급과 키를 `[비움 — Q4]`로 올린다. 결정 (3)(ii) 「재알림 창 SHALL NOT 과 대조」는 Q4의 답이 critical일 때 적용한다 | `notifier.claimanddeliver` · `reconciledriver.checkexternalincrease` · 결정 (3)(ii) |
| **P0-3** exit-policy 델타의 「쓰기 경로는 하나」가 거짓 | `baseline_price`의 쓰기 자리는 하나가 아니고 하향 거부는 이미 있다 | 거짓 전제를 **지운다.** 쓰기 자리 넷을 AST로 열거해 issues I1에 적는다: 최초 INSERT(`OpenExitState`) · 판정 UPDATE(`recordExitJudgementTx` — 옛 경로는 B25 `notBelow("baseline", …)`, 스냅샷 경로는 B29 창 `SelectRecoverySnapshot`) · 관측 갱신(`RefreshExitObservation` — B23이 보호가가 다르면 거절하므로 값은 그대로) · 재편입 reset(`resetExitStateForReadoptTx` — 분기 여섯 중 이전 기준선과 비교하는 것이 없다). 선행 조건을 참인 문장으로 다시 SHALL로 적을지는 `[비움 — Q6]` | `journal.recordexitjudgementtx` · `journal.refreshexitobservation` · `resetexitstateforreadopttx` · `journal.openexitstate` |
| **P0-4** critical 승격이 exit 루프의 손절 판정 앞에 최악 54s 체류를 넣는다 | `workingSet`이 `observe`·`judge` 앞에서 `alertUnmanaged`를 부르고, critical이면 `n.mu`를 쥔 채 재시도 예산을 돈다. a092는 그것을 약속하지 않는다 | **exit 관측 자리는 normal로 남는다**(결정 (1)). `ObserveOnce`의 호출 순서 `o.workingSet :426 → o.observe :441 → o.judge :465`와 `workingSet` B6 창의 `o.alertUnmanaged`는 그대로이고, 그 사실은 `Notify` B1 창의 `publishBestEffort`로 간다 — `Notify`의 두 경로 중 `n.mu`를 잡는 것은 critical 경로의 `claimAndDeliver`뿐이다(그 밖의 잠금 자리 `Flush` `notifier.go:734` · `Acknowledge` `:851`은 발신 경로가 아니다). a092에 묶지 않는다. 옛 약속 사본(tasks 2.8 · 안전표 §4 행 · design D7 행)을 지웠다. **남는 것**: reconcile 자리의 critical 배달이 쥔 `n.mu`를 exit goroutine의 **기존** critical 발신이 기다리는 경합 — `[비움 — Q7]` | `exitobserver.observeonce` · `exitobserver.workingset` · `notifier.notify` · `notifier.claimanddeliver` · 결정 (1) |

## Why

사용자 보고(2026-08-06): *"익절 손절이 잘못 계산되고 있는것 같은데 추가 구매하면 평균
단가가 낮아 지는데 이를 반영 하지 않고 있는 것 같아."*

2판까지의 진단은 셋이었다 — 손절은 `entry_price`에서 파생되고 평단을 따르지 않는다, 총위험을
세는 곳이 없다, 그리고 **보호 범위에 대한 보고가 원장에 남지 않는다.** 3판은 셋째를 중심에 두고,
2라운드가 드러낸 **실제 구멍**으로 범위를 옮긴다(결정 (3)).

### 무관리 보유의 보고는 한 번도 원장에 남지 않았다

- `EventExitPositionUnmanaged`는 `criticalEvents`(18종)에 없고, `SeverityOf` B1은 그 map만 본다
  (`severityof`). 미등재는 normal이고, `Notify` B1 창이 `publishBestEffort`로 보낸다 — outbox 행도
  재시도도 없다(`notifier.notify` · `notifier.publishbesteffort`).
- `alert_outbox` 실측: 2026-08-07 13행, 2026-09-25 16행 — **둘 다 전부 critical이고
  `exit.position_unmanaged`는 0행**이다(2라운드 §2.6 P1-6, 읽기 전용 조회). 결함 계열은 남아 있다.

### 그러나 「가진 것에 손절이 없다」는 사실은 한 종류가 아니다

2라운드의 세 검증이 서로 보지 않고 같은 결론에 닿았다(§2.9 · §2.14 · §2.15). 3판의 번들이 그 차이를
분기로 고정한다:

| 발신 자리 | 들어오는 길(분기) | 사실 | 누가 고른 상태인가 |
| --- | --- | --- | --- |
| exit 관측 `ExitObserver.alertUnmanaged` | `workingSet` B6(`!p.ExitEligible()`) | 적격하지 않은 보유 | 판정 앞, 전이 상태 판정 없음 |
| reconcile `ReconcileDriver.alertUnmanaged` B4 | `judgeHoldings` B11(exclude) | 의도적으로 뺀 종목 | **운영자** |
| 같은 함수, 기본 사유 | `judgeHoldings` B12(off ∧ 미지정) | 편입을 켜지 않았다 | **운영자** |
| 같은 함수 B5 | `judgeHoldings` B14(편입 안 됨) ← `adopt` B2 · B6 · B7 · `adoptOne` 실패 | 편입하라 했는데 못 했다 | 운영자가 고르지 않았다 |
| 같은 함수 B3 · B6 | 설정 거부 · include 지정 시도 실패 | — | `[비움 — Q2]` |
| `checkExternalIncrease` | `judgeHoldings` B8(편입됨) | 편입 후 수량 증가 — **증가분도 원래 손절의 보호를 받는다** | 무보호가 아니다 — `[비움 — Q4]` |
| `notifierAlerter.ExternalPositionFound` | `IngestExternalPositions` B13 | fold 알림 | **생산에서 도달하지 않는다** — B12(`in.Alert == nil`), `ReconcileDriver`가 `d.ingest.Alert = nil`로 둔다(`reconcileloop.go:338`) |

### 그리고 엔진이 직접 연 포지션의 수량 증가는 아무도 보지 않는다

`judgeHoldings` B7 창은 적격 포지션에서 `continue`하고, B8(`p.Adopted()`)만 `checkExternalIncrease`를
부른다. **엔진이 진입 결정으로 연 포지션에 앱에서 더 사도 검사가 없다.** 결정 (3)(i)이 이것을 범위에
넣는다.

### 측정 시점

2판의 원장 표(066570 · 080220 · 272210 · 475150 · TSLA · 010170)는 **2026-08-07** 측정이다. 2026-09-25
재조회에서 그 동기 포지션은 **TSLA 먼지 1건을 빼고 전부 CLOSED**였다(2라운드 P1-6). 3판은 그 표를
근거로 쓰지 않고, 위 분기 사실과 outbox 0행을 근거로 쓴다. 475150의 편입 수량은 2판의 3이 아니라
**2**였다(P2-1) — 이력으로만 남긴다.

## What Changes (3판)

### R1′ — 등급은 이벤트 종류가 아니라 사실로 (결정 (1)·(2))

| 사실 | 3판 등급 | 결정 |
| --- | --- | --- |
| exit 관측 자리(`workingSet` B6 → `ExitObserver.alertUnmanaged`) | **normal — 무변화** | (1) |
| exclude 종목(`judgeHoldings` B11 → `alertUnmanaged` B4) | **normal — 무변화** | (2) |
| 편입 off ∧ 미지정(`judgeHoldings` B12 → 기본 사유) | **normal — 무변화** | (2) — 정본 exit-policy「`adoption.enabled` … false에서의 동작은 … 기존 동작과 동일」 유지 |
| 알림 off 엔진에서 생긴 모든 a095 사실 | **critical로 매기지 않는다**(→ ENTRY_BLOCKED에 닿지 않음) | (2). 4판 r3 N1 — critical 요구의 전제가 「알림 켜짐」이다 |
| 알림 켜짐 · 편입하라 했는데 **편입 시도가 실패**했다(`adopt` B8 거짓 → `judgeHoldings` B14 → `alertUnmanaged` B5) | **critical** | (1) 「발신은 reconcile 쪽에서만 critical」 + (2) 「운영자가 고른 상태는 critical 에서 뺀다」의 귀결 |
| 설정 거부(B3) · include 지정 시도 실패(B6) · 시세 연기로 편입 안 된 후보(`adopt` B2 · B6 · B7) | `[비움 — Q2(a)(b)(c)]` — 연기분은 4판 델타의 critical 요구에서 **뺐다**(r3 N2) | 결정이 이름 대지 않음 |

**싣는 방식**은 `[비움 — Q1]`. `SeverityOf`는 종류만 보므로(B1) 같은 종류의 두 자리에 다른 등급을
줄 수 없다. 방식이 정해지기 전에는 이 change가 `criticalEvents`에 무엇을 더하는지 적지 않는다.

**a091과 같은 경계다.** a091 design은 `EventExitPositionUnmanaged`를 승격하지 않는 이유를
*"운영자가 선택한 정상 상태"*로 적고 경계를 *"엔진이 그 포지션을 보호하기로 했는가"*로 둔다. 3판이
critical로 올리는 것은 `adoption.enabled`가 참이어서 **엔진이 보호하기로 한** 보유의 편입 실패뿐이다.

### R2′ — 수량 증가 (결정 (3))

- **(i) 엔진 개설 포지션도 검사한다.** 자리는 `judgeHoldings` B7 창. 비교 기준(`exit_states`에는
  수량 열이 없다) `[비움 — Q3]`.
- **(ii) 키와 재알림 창.** 수량 증가 사실의 종류·등급이 먼저다 — `[비움 — Q4]`. critical이면 키를
  정본 engine-safety 「같은 조건의 critical 알림은 재알림 창 안에서 한 번만 전송한다」(SHALL NOT)와
  대조해 정하고, 475150 원장 순서(3·4·5·8·26·32)를 재생하는 시험으로 못 박는다.
- **(iii) 발신 자리의 키를 가른다.** exit 관측 자리와 reconcile 자리가 오늘 같은 철자
  `exit.position_unmanaged|<posID>`를 쓴다(두 `alertUnmanaged` 번들의 변이 절). 두 자리는 서로 다른
  키를 쓴다.
- **R2-B2 삭제.** `checkExternalIncrease` B2는 바꾸지 않는다. 호출자 가드(`judgeHoldings` B8)와
  `positions.adoption_id REFERENCES position_adoptions(id)` · `foreign_keys(on)` 때문에 B2가 받는 입력은
  조회 오류다. 「편입 기록 없음 = 보호 없음」 SHALL·시나리오도 삭제한다.

### R3 — 총위험 보고 — **보류, 델타에서 SHALL 해제 (Q5 결정)**

> **사용자 결정 (2026-09-28, 일괄 승인 — Manager 전달 원문):** *"Q5 = R3(총위험) 보류 + delta 의 SHALL 지위
> 해제(archive 차단 해소, 후속 change 후보로 기록) / Q2(d)·Q7 = a092·a124 영역 이관(a092 21판이 그 소유를
> 인수했고 a124 는 아카이브됨 — 교차 인용 갱신)."*

2판의 R3(`(평단 − 유효 손절) × 현재 수량`)는 a095가 구현하지 않는다. exit-policy 델타의 R3 요구사항
(「보호는 자기가 덮는 수량과 총위험을 말할 수 있어야 한다」)을 **지웠다** — archive가 논쟁 중인 SHALL을
정본에 넣지 않게 하기 위해서다. **후속 change 후보**로 `issues.md` I2에 남긴다: 그 후보가 먼저 풀어야 할
것은 평단의 출처다(보이스 A A-5 — 수량 수렴 경로에서 평단이 설계상 낡는다는 주장, 3판 미재검증).

### 하지 않는 것

- **손절가의 평단 기준 재계산** — 물타기 방향에서 손절이 내려간다(§6). 3판은 손절가를 올리지도
  내리지도 않는다.
- **`EventExitPositionUnmanaged`의 일괄 승격** — 결정 (1)·(2)가 금지하는 형태다.
- **a092에 묶기** — 결정 (1).
- **정본 exit-policy의 `adoption.enabled` false 동등성 변경** — 결정 (2), MODIFIED 없음.
- **`positions.avg_price` 보정 · 수량 상한(사이징)** — 2판과 같다.
- **평단이 높아졌을 때 유효 손절을 올리는 자동 경로(불타기 래칫)** — 도입하지 않는다. 별도 change의 주제이고
  선행 사실은 `issues.md` I1. (3판까지 exit-policy 델타에 SHALL NOT으로 있던 문장 — 정본에 병합되면 영구 금지로
  굳으므로 4판이 여기 범위 문장으로 옮겼다, 3라운드 V-N4)
- **운영자 재편입 reset의 하향에 승인 · audit를 붙이는 것** — 하지 않는다. `issues.md` I6(해제 원칙 가족의 잔여 질문)

## 열린 질문 — Manager 에게 (결정 (1)~(3)이 덮지 않는 지점)

1. **Q1 — 등급을 싣는 방식.** `SeverityOf` B1은 종류만 본다. reconcile 자리의 비선택 사실만 critical로
   만드는 방식은 (a) 그 사실에 새 이벤트 종류를 주고 `criticalEvents`에 등재, (b) `Event`에 등급을 싣고
   `SeverityOf`의 계약을 바꿈, (c) 다른 방식 — 어느 것인가? 그리고 발신 자리가 「알림 off」를 어떻게
   아는가(`Notifier.Publisher`가 nil인지를 읽을지, 설정값을 `ReconcileDriver` 옵션으로 넘길지)?
   **제약(3판 문서 리뷰가 더함)**: 정본 engine-safety 「등급화된 알림」은 critical의 *"전달 실패가
   지속되면 신규 진입을 차단한다(SHALL)"*고 적는다. 따라서 알림 off에서 진입 차단을 막으려면 그 상태에서
   a095의 사실이 **critical로 매겨지지 않아야** 한다 — critical로 매긴 뒤 차단만 거르면 정본과 어긋나고
   MODIFIED가 필요하다(결정 (2)는 MODIFIED를 금한다).
2. **Q2 — 결정 (2)가 이름 대지 않은 사유.** (a) 설정 거부(`alertUnmanaged` B3) (b) include 지정 시도 실패
   (B6) — 특히 `adoption.enabled=false`에 include 목록이 있을 때 정본 「false = 기존 동작」과의 관계
   (c) `adopt` B2(시세 읽기 오류) · B6(관측 없음) · B7(관측 묵음)로 **연기된** 후보가 B5 사유로 알려진다 —
   일시적 시세 실패가 critical이 되는 것을 받아들이는가 — **(a)(b)(c) 열림: 구현 로트행**(Manager 분류
   2026-09-27, 코드 영수증으로 확정).
   ~~(d) 알림이 켜져 있지만 transport가 죽은 엔진에서 B5 사실이 ENTRY_BLOCKED를 부르는 교환~~ →
   **이관(사용자 결정 2026-09-28): a124 영역.** 그 교환은 a095가 정하지 않는다 — critical 전달 실패의 진입
   차단 · 운영 모드 승격의 주체와 판정은 a124가 착지시킨 정본 engine-safety 「배달 실행자는 지속 실패를 진입
   차단과 운영 모드 승격으로 잇는다」(a124 archive `c1e34dc4`)가 소유한다. a095가 B5를 critical로 매기면 그
   사실은 그 정본을 따른다.
3. **Q3 — 엔진 개설 포지션 수량 증가의 비교 기준.** `exit_states`에 수량 열이 없다. 진입 fill 합계,
   `position_adjustments` 이력, 마지막으로 보고한 수량(메모리) 중 무엇을 기준으로 하는가?
4. **Q4 — 수량 증가 사실의 종류·등급.** 증가분은 원래 손절의 보호를 받는다(무보호 아님). 별도 종류로
   가르는가, 그리고 normal인가 critical인가? critical이면 키에 수량을 넣는가(행마다 ack · 진입 차단
   비용), 재알림 창을 따르는가?
5. ~~**Q5 — R3의 지위.**~~ → **결정(2026-09-28): 보류 + 델타 SHALL 해제**, 후속 change 후보(위 R3 절).
6. **Q6 — 래칫 선행 조건.** exit-policy 델타의 남은 요구(평단 하락 비하향)에서 거짓 전제를 지웠다. 쓰기 자리 넷의 사실로
   선행 조건을 다시 SHALL로 적는가, issues I1에만 두고 후속 change에 넘기는가?
7. ~~**Q7 — `n.mu` 경합.**~~ → **이관(사용자 결정 2026-09-28): a092 영역**(아래 「이관 기록」). 원문: reconcile 자리의 critical 배달은 `claimAndDeliver`가 `n.mu`를 쥔 채
   `n.deliver`를 부른다. 그동안 exit goroutine의 **기존** critical 발신(`exitloop.go:831`
   `EventExitObservationOutage` · `:1633` `EventExitJudgementRefused` · `:1657` `EventExitProposalRefused` ·
   `:1687` `EventExitLiquidationDelayed`)이 같은 뮤텍스를 기다린다. a095가 이 경합의 모집단을 늘리는 것을
   받아들이고 a092 · a124의 소유로 두는가, a095 안에서 상한을 두는가?
   **같은 뿌리의 둘째 면(3판 문서 리뷰가 더함)**: 그 배달은 reconcile goroutine 자신도 붙잡는다.
   `judgeHoldings` B15가 무관리 보유마다 `alertUnmanaged`를 **차례로** 부르므로, transport가 죽은 채
   B5 사실이 N건이면(예: 시세 경로 장애로 `adopt` B2가 후보 전원을 연기) 대사 사이클이 N × 배달 예산만큼
   멈춘다. 손절 경로는 아니지만 대사 주기가 늦어진다.

**이관 기록 (Q7 · Q2(d), 2026-09-28).**

| 면 | 소유 | 교차 인용 |
| --- | --- | --- |
| Q7 첫째 면 — exit goroutine의 기존 critical 발신이 대사 쪽 배달이 쥔 `n.mu`를 기다림 | **a092 21판** | a092 engine-safety 델타 「exit 관측 goroutine이 기다리는 잠금은 원격 전송을 덮어서는 안 된다(SHALL NOT)」 — 그 잠금의 **모든 보유자**가 원격 전송 동안 잠금을 놓아야 한다(design D0.3e 5번). a092 review(21판 Manager 판정 Q1)가 *"a095 Q7의 첫째 면 … 21판의 잠금 범위가 닫는다"*고 적었다 |
| Q7 둘째 면 — 대사 goroutine 자신이 무관리 보유마다 원격 왕복을 기다림 | **a092 영역 · 21판 밖의 이름 붙은 잔여** | 같은 판정: *"둘째 면(대사 goroutine 자신이 무관리 보유마다 원격 왕복을 기다림)은 Q1 문자 해석상 21판 밖에 남는다."* a095는 이 잔여를 **이름으로 남기고** 고치지 않는다 |
| Q2(d) — transport 사망 시 ENTRY_BLOCKED 교환 | **a124**(archive `c1e34dc4`) → 정본 | 정본 engine-safety 「배달 실행자는 지속 실패를 진입 차단과 운영 모드 승격으로 잇는다」 |

**아직 열림 — 구현 로트행**(Manager 분류 2026-09-27, a095가 스케줄될 때 코드 영수증으로 확정): Q1 · Q2(a)(b)(c) ·
Q3(정지 조건 — `exit_states`에 수량 열이 없다는 스키마 질문) · Q4 · Q6 · **Q8(정지 조건, 4판)**.

8. **Q8 — 사실이 해소된 뒤의 critical 행 (4판, r3 N3 · Manager 설계 지시의 조건 검사).** 알림 켜짐 · 편입 실패로
   만든 critical 행이 PENDING인 채 다음 사이클에 편입이 성공하면, 그 행은 나중에 「지금 무보호」 문구로 배달되고
   운영자 승인 전에는 PENDING을 떠나지 않는다. Manager 지시는 「기존 재개방/본문 갱신 기계로 갱신 또는
   해소-정산」이었고, **그 기계는 PENDING 행에 쓸 수 없다** — 재무장은 `claimOwed`가 정착 행에만 준다
   (`claimowed` B2 창 return `:381` owed=참 · rearm=거짓), 해소 정산 연산은 outbox에 없다(design D1 「사실이
   해소된 뒤의 행」 표). 무엇으로 처리하는가 — (a) 새 outbox 연산(해소 정산 또는 PENDING 본문 갱신), (b) 행 문구를
   「그 시각의 실패 사건」으로 바꿔 해소 뒤 배달도 참이 되게, (c) 기타? **구현 로트의 정지 조건이다.** 사람 소유
   래치를 자동으로 푸는 답은 제외한다.

## Impact

| | 자리 | 성격 |
| --- | --- | --- |
| R1′ | reconcile 쪽 무관리 보고(`adoption.go` `alertUnmanaged` · 그 호출자) · 등급 표 | 방식 Q1. exit 관측 자리 · `publishBestEffort` · `notifyCritical` · `deliver` 본문은 무변화. **`SeverityOf` · `Notify`의 편집 경계는 Q1의 답에 달렸다**(4판 r3 N4) — (a) 새 종류 등재면 본문 불변, (b) 등급을 `Event`에 싣고 `SeverityOf` 계약을 바꾸면 경계 재선언 · 번들 재생성 · 재리뷰 |
| R2′ | `adoption.go` `judgeHoldings` B7 창 · 발신 키 | Q3 · Q4 |
| R3 | **보류 — 델타에서 지움, 후속 change 후보** | 결정 Q5 |

spec: `engine-safety`(사실별 등급 · 진입 차단 비도달 · 키 분리), `exit-policy`(R3 요구 삭제 · 래칫 요구의
거짓 전제 삭제).

**기존 함수 내부를 고치므로 Function Logic Map 면제는 없다.** 번들 26개(3판 21 + 4판 5 — `recordAlertTx` · `claimOwed` ·
`alertDeliverer.cycle` · `alertDeliverer.deliverOne` · `SelectRecoverySnapshot`)가 있고, 구현 후 다시 뽑는다.
