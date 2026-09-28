# a095 · 설계 — 3판

> 분기 인용은 전부 `analysis/function-logic/`의 AST 산출물에서 온다(base `02716357`, 번들 21개).
> 번들 이름은 디렉터리 이름의 뒷부분으로 적는다(예: `reconciledriver.judgeholdings`).
> 결정 (1)~(3)의 원문은 `proposal.md` §0에 있다. 결정이 덮지 않는 절은 `[비움 — Qn]`으로 두고
> 질문은 `proposal.md` 「열린 질문」에 있다.

## D0. 하나의 원칙 — 유지

**보호의 기준은 얼려 둔다. 보호가 무엇을 덮는지는 매번 다시 잰다.**

기준(`entry_price` · `initial_stop` · `initial_risk`)은 이미 보고된 R의 분모다. `checkExternalIncrease`의
주석(adoption.go `:437-440`)이 그 동결을 설계로 선언하고 그 논거는 옳다. 3판도 기준을 쓰지 않는다.

**3판이 더하는 원칙: 등급은 이벤트 종류가 아니라 사실이 정한다**(결정 (2)). 같은 「관리하지 않는
보유」라도 운영자가 고른 상태와 엔진이 보호하기로 했는데 실패한 상태는 다른 사실이다.

## D1. R1′ — 등급을 사실로

### 사실의 지도 (분기로 고정)

`judgeHoldings`(`reconciledriver.judgeholdings`, 분기 15)가 reconcile 쪽의 유일한 입구다.

| 분기 | 조건 (원문, 번들 표에서) | 창의 결과 | 3판 사실 |
| --- | --- | --- | --- |
| B7 | `:108` `if p.ExitEligible() {` | B8을 거쳐 `continue` | 보호 중 — 무관리 아님 |
| B8 | `:109` `if p.Adopted() {` | `d.checkExternalIncrease` | 편입 후 수량 증가만 검사(D2) |
| B9 | `:116` `if d.blocked(market, symbol) {` | `continue` | 전이 상태 — 무알림 |
| B10 | `:119` `if !fresh {` | `continue` | 전이 상태 — 무알림 |
| B11 | `:127` `if d.opts.Adoption.Excludes(symbol) {` | `unmanaged`에 넣음 | **운영자가 고름** |
| B12 | `:135` `if !d.opts.Adoption.Enabled && !d.opts.Adoption.Included(symbol) {` | `unmanaged`에 넣음 | **운영자가 고름** |
| B14 | `:144` `if !adopted[c.position.ID] {` | `unmanaged`에 넣음 | 편입 대상인데 편입되지 않음 |
| B15 | `:152` `for _, p := range unmanaged {` | `d.alertUnmanaged` | 모인 것을 알림 |

B14의 입력은 `adopt`(`reconciledriver.adopt`, 분기 8)가 돌려준 집합이다. `adopt`는 B2(시세 읽기 오류) ·
B6(종목 관측 없음) · B7(관측 묵음)에서 후보를 편입하지 않고, B8(`d.adoptOne`)이 거짓이어도 편입하지
않는다. 그 후보 전부가 B14로 무관리가 된다.

`alertUnmanaged`(`reconciledriver.alertunmanaged`, 분기 6)의 why-matrix B2 switch가 사유 문구를 고른다:
B3 설정 거부 · B4 exclude · B5 enabled 시도 실패 · B6 include 지정 시도 실패 · 기본(off ∧ 미지정).
**사유가 이미 사실별로 갈려 있다** — 3판은 이 분기를 등급의 근거로 쓴다.

### 결정이 정한 등급

| 사실 | 등급 | 근거 |
| --- | --- | --- |
| exit 관측 자리 | normal — 무변화 | 결정 (1) |
| B4 exclude | normal — 무변화 | 결정 (2) |
| 기본 사유(off ∧ 미지정) | normal — 무변화 | 결정 (2) · 정본 exit-policy `adoption.enabled` false 동등 |
| B5 enabled 시도 실패 | **critical** | 결정 (1)「발신은 reconcile 쪽에서만 critical」 + (2)「운영자가 고른 상태는 critical 에서 뺀다」 |
| 알림 off 엔진 | 어떤 a095 사실도 ENTRY_BLOCKED에 닿지 않음 | 결정 (2) |
| B3 설정 거부 · B6 include 시도 실패 · `adopt` 연기분(B5 사유로 모임) · 알림 on이지만 transport 죽음 | `[비움 — Q2]` | 결정이 이름 대지 않음 |

**a091과 같은 경계다.** a091 design(`design.md` 「승격하지 않는다」 절)은 *"경계는 '엔진이 그 포지션을
보호하기로 했는가'이고, 그 술어는 이미 코드에 있다"*고 적는다. B5는 `adoption.enabled`가 참인 — 엔진이
보호하기로 한 — 보유의 편입 실패다.

### 싣는 방식 — `[비움 — Q1]`

`SeverityOf`(`severityof`, 분기 1)의 B1 `:348` `if criticalEvents[t] {`는 종류만 본다. 따라서 같은 종류
`EventExitPositionUnmanaged`를 쓰는 exit 관측 자리(normal)와 reconcile B5(critical)를 표 한 줄로 가를 수
없다. 방식(새 종류 · 등급 필드 · 기타)과 「알림 off」를 발신 자리가 아는 방법이 정해지기 전에는 이 절을
쓰지 않는다.

### 알림 off에서 무엇이 막혀야 하는가 (결정 (2)의 근거 경로)

critical 사실이 알림 off 엔진에 닿으면 다음이 일어난다 — 전부 번들의 분기다.

```text
Notify B1 :134 (critical이면 통과) → notifyCritical
  └ claimAndDeliver  n.mu.Lock … n.deliver (뮤텍스 안)
       └ deliver B3 :429  if n.Publisher == nil { lastErr = …; break }   ← 진입 실측: 아니오
            └ B27 :570  if n.Gate != nil → n.Gate.Block(...)
  └ notifyCritical B4 :223  if owed && !sent → n.escalate(ctx, e)
```

`deliver` B3은 **어떤 시험도 밟지 않는다**(`notifier.deliver`). 결정 (2)는 a095의 사실이 이 사슬에
**들어오지 않을** 것을 요구한다. 3판은 사슬 안에 거르는 분기를 더하지 않는다(그 함수들은 a092 · a124의
영역이고 다른 critical 사건에도 쓰인다). 거름은 발신 쪽에 있어야 하며 그 방식이 Q1이다.

### exit 관측 자리가 normal이어야 하는 이유 (결정 (1)의 근거 경로)

`ObserveOnce`(`exitobserver.observeonce`)의 호출 좌표: `o.workingSet` `:426` → `o.observe` `:441` →
`o.judge` `:465`. `workingSet`(`exitobserver.workingset`) B6 `:512` `if !p.ExitEligible() {` 창이
`o.alertUnmanaged`를 부른다. 이 자리의 알림은 **그 사이클의 모든 손절 판정 앞**에 선다.

normal이면 `Notify` B1 창의 `n.publishBestEffort`로 간다 — 그 함수(`notifier.publishbesteffort`)는
뮤텍스를 잡지 않고 outbox를 쓰지 않는다. 결정 (1)로 이 자리는 오늘과 같다. 기존 시험
`TestAPositionWithNoEntryDecisionIsSkippedAndAlertedOnce`가 그 등급을 이미 고정한다
(`exitloop_test.go:508`).

### 남는 경합 — **a092 영역으로 이관** (사용자 결정 2026-09-28)

`claimAndDeliver`(`notifier.claimanddeliver`)의 호출 목록은 `n.mu.Lock` · `n.mu.Unlock`(defer)로 열고
`n.deliver`로 끝난다 — 배달 전체가 뮤텍스 안이다. reconcile B5 사실이 critical이 되면 그 배달 동안
exit goroutine의 **기존** critical 발신(`exitloop.go:831` · `:1633` · `:1657` · `:1687`)이 같은 뮤텍스를
기다린다. 결정 (1)은 exit goroutine에 **새** critical Notify를 두지 않는 것이고, 이 경합은 새 발신이
아니라 **기존 발신의 대기 증가**다.

**처분(2026-09-28)**: a092 21판이 소유한다 — 그 engine-safety 델타가 exit goroutine이 줄 서는 잠금의 **모든
보유자**에게 원격 전송 동안 잠금을 놓으라고 요구한다(SHALL). 대사 goroutine 자신의 대기(둘째 면)는 a092 21판
밖의 **이름 붙은 잔여**다. a095는 이 경합에 대해 설계하지 않는다. 교차 인용은 `proposal.md` 「이관 기록」.

### 키 분리 (결정 (3)(iii))

두 `alertUnmanaged`(`exitobserver.alertunmanaged` · `reconciledriver.alertunmanaged`)가 같은 철자
`string(obs.EventExitPositionUnmanaged) + "|" + p.ID`를 쓴다. 3판은 두 자리가 **다른 키**를 쓰게 한다.
exit 관측 자리가 normal인 동안 키는 outbox에 닿지 않지만(normal 경로는 `eventKey`를 쓰지 않는다),
등급이 사실로 갈리는 순간 같은 키는 한 outbox 행으로 합쳐져 먼저 온 문구만 남는다(보이스 B B-P1-5).
철자는 Q1의 방식이 정해지면 따라 정해진다.

### 발신 자리 넷의 3판 처분

| 자리 | 3판 |
| --- | --- |
| `ExitObserver.alertUnmanaged` (exit 관측) | normal 유지 · 키 분리 |
| `ReconcileDriver.alertUnmanaged` (reconcile) | 사유별 등급(위 표) · 키 분리 |
| `checkExternalIncrease` (수량 증가) | D2 · Q4 |
| `notifierAlerter.ExternalPositionFound` | **무변화.** 생산 호출 자리 `IngestExternalPositions` B13은 B12 `:278` `if !folded.Applied \|\| in.Alert == nil {`에 막힌다(`reconcile--ingestor.ingestexternalpositions`) — `ReconcileDriver`가 `d.ingest.Alert = nil`로 복사한다(`reconcileloop.go:338`). 등급이 normal로 남으므로 2판 tasks 6.2a의 오류 전파 변화도 없다 |

## D2. R2′ — 수량 증가

### R2-B2 삭제 (결정 (3))

`checkExternalIncrease`(`reconciledriver.checkexternalincrease`, 분기 3)의 B2 `:446` `if err != nil {`는
`AdoptionOf` 실패에서 반환한다. 호출자 `judgeHoldings` B8이 `p.Adopted()`인 포지션만 부르고,
`positions.adoption_id`는 `REFERENCES position_adoptions(id)`(journal/adoption.go `:92`)이며 연결이
`foreign_keys(on)`(journal.go `:223`)이다. **B2가 받는 입력은 조회 오류다.** 2판의 「편입 기록이 없는
보유가 여기로 온다」는 거짓이었고, 그 위에 선 R2-B2 · SHALL · 시나리오를 지운다. B2는 무변화.

### (i) 엔진 개설 포지션 — `[비움 — Q3]`

`judgeHoldings` B7 창: 적격이면 B8만 검사하고 `continue`한다. 편입 기록이 없는 적격 포지션 — 엔진이
진입 결정으로 연 것 — 은 수량이 늘어도 아무 검사를 받지 않는다. 결정 (3)(i)이 이것을 범위에 넣는다.

`exit_states`에는 수량 열이 없다(`core_domain.go` 스키마). 비교 기준이 정해지기 전에는 설계를 쓰지 않는다.

### (ii) 수량 증가 사실의 종류·등급과 키 — `[비움 — Q4]`

오늘의 키는 `…|grown|<posID>`이고 수량이 없다. 그 알림의 본문은 스스로 *"늘어난 수량은 원래 수량
기준으로 산정된 손절의 보호를 받는다"*고 쓴다 — **무보호가 아니다.** 그런데 종류는
`EventExitPositionUnmanaged`이고 제목은 *"고정된 t0가 증가분을 덮지 않는다"*다.

`claimAndDeliver` B5 `:285` `case journal.ClaimSettled:`가 같은 키의 재전송을 재알림 창 안에서 삼키는 것은
**critical 경로에서만** 일어난다. 따라서 2라운드 P0-2는 이 사실이 critical일 때의 문제다. 종류·등급이
정해지면: critical이면 키를 정본 engine-safety 「같은 조건의 critical 알림은 재알림 창 안에서 한 번만
전송한다」와 대조해 정하고, normal이면 재알림 창과 무관하게 `d.grown`(B1) 래치의 기준만 정한다.

## D3. R3 — 총위험 — **보류, 델타 SHALL 해제** (사용자 결정 2026-09-28)

a095는 총위험을 구현하지 않고, exit-policy 델타에서 그 요구를 지웠다. 후속 change 후보는 `issues.md` I2.
2판 D3는 이력으로 남긴다. 보이스 A A-5의 주장 — 수량 수렴 경로(`ConvergeQuantities`)가
`NewAvgPrice: ""`를 보내 `firstNonEmpty`로 평단이 이어지므로 R3가 겨냥한 바로 그 경우에 평단이 낡는다 —
은 **3판이 재검증하지 않았다**(그 함수의 번들이 없다). 후속 change가 그 번들을 먼저 뽑는다.

## D4. 손절가를 평단 기준으로 다시 계산하지 않는다 — 사실 정정

### 방향이 대칭이 아니다 — 유지

물타기(평단↓)에서 평단 기준 손절은 지금보다 낮다 — 내리면 §6 위반이다. 3판은 손절가를 올리지도
내리지도 않는다.

### `baseline_price`의 쓰기 자리 — 2판의 「하나뿐」 정정

| 자리 | 번들 | 쓰는 값 | 하향에 대한 분기 |
| --- | --- | --- | --- |
| 최초 INSERT | `journal.openexitstate` | 진입 손절 | — (최초) |
| 판정 UPDATE | `journal.recordexitjudgementtx` | 판정의 기준선 | 옛 경로: B23 `:487` `if recomputed == nil {` → B25 `:491` `notBelow("baseline", …)`. 스냅샷 경로: B28 `:506` → B29 창 `exitpolicy.SelectRecoverySnapshot` |
| 관측 갱신 UPDATE | `journal.refreshexitobservation` | 저장된 effective 스냅샷의 보호가와 같은 값 | B23 `:127`이 `sameExitOperationalLine` 거짓이면 거절 — 값이 움직이지 않는다 |
| 재편입 reset UPDATE | `resetexitstateforreadopttx` | 재편입 관측의 합성 손절 | 분기 여섯(B1~B6) 모두 오류·행 수 검사 — **이전 기준선과 비교하는 분기가 없다.** 운영자 행동(`positionpolicy.ActionReadopt`)에서만 불린다 |

주석의 *"the only reset writer for the four guarded execution-time columns"*에서 네 열은
`taken_ratio_total` · `pending_action` · `pending_level` · `pending_intent_id`이고 손절 열이 아니다(보이스 A
A-4). 2판이 이것을 「유효 손절의 유일한 쓰기 경로」로 읽은 것이 P0-3이다.

### 래칫 상향의 선행 조건 — `[비움 — Q6]`

2판의 선행 조건 셋 중 첫째(「쓰기 경로 하나」)는 거짓이었고, 둘째(「하향 거부」)는 판정 경로에 이미
있다(위 표). 참인 문장으로 다시 SHALL을 세울지, 위 표를 issues I1에만 두고 후속 change에 넘길지가 Q6이다.
그때까지 exit-policy 델타는 「평단이 내려도 유효 손절가를 낮추지 않는다」만 싣는다.

## D5. 셋의 상호작용 (3판)

```text
reconcile 사이클 — judgeHoldings
  ├─ B7 적격 ──┬─ B8 편입됨 → checkExternalIncrease ── D2 (Q4)
  │           └─ 편입 안 됨(엔진 개설) → continue ── D2 (i) (Q3)  ← 오늘의 구멍
  ├─ B9·B10 전이 상태 → continue (무알림, 유지)
  ├─ B11 exclude ─────────┐
  ├─ B12 off ∧ 미지정 ────┤→ unmanaged → B15 → alertUnmanaged
  └─ B14 편입 안 됨 ───────┘              ├─ B4 · 기본 사유 → normal (결정 2)
        ↑ adopt B2·B6·B7·B8                ├─ B5 → critical (결정 1·2)  — 방식 Q1
                                           └─ B3 · B6 → Q2

exit 관측 사이클 — ObserveOnce
  workingSet :426 ── B6 !ExitEligible → ExitObserver.alertUnmanaged → normal (결정 1, 무변화)
  observe    :441
  judge      :465   ← 이 앞에 새 critical 체류 없음. 기존 critical 발신의 n.mu 대기는 a092 영역
```

## D6. 무엇을 하지 않는가 (3판)

| | 결정 | 근거 |
| --- | --- | --- |
| `EventExitPositionUnmanaged`를 종류째 `criticalEvents`에 등재 | **안 한다** | 결정 (1)·(2) — exit 관측 자리와 운영자 선택 상태가 함께 올라간다 |
| exit goroutine에 critical Notify를 새로 둠 | **안 한다** | 결정 (1) |
| a092 착지를 선행 조건으로 둠 | **안 한다** | 결정 (1) |
| 정본 exit-policy `adoption.enabled` false 동등성 수정 | **안 한다** | 결정 (2) — MODIFIED 없음 |
| `checkExternalIncrease` B2를 `alertUnmanaged`로 보냄(R2-B2) | **삭제** | 결정 (3) · D2 |
| `notifyCritical` · `claimAndDeliver` · `deliver` 본문 수정 | **안 한다** | 거름은 발신 쪽(D1) |
| `ExternalPositionFound`의 등급 · 오류 전파 변경 | **안 한다** | 생산에서 도달하지 않는다(D1 표) |
| 손절가의 평단 기준 재계산 · 기준 컬럼 쓰기 | **안 한다** | D0 · D4 |

## D7. 실패 모드 재검토 (3판)

| 우려 | 답 |
| --- | --- |
| 손절 판정 앞에 critical 체류가 들어가는가? | **들어가지 않는다.** exit 관측 자리는 normal(결정 (1)) — `publishBestEffort`는 뮤텍스를 잡지 않는다. 남는 것은 reconcile 쪽 배달이 쥔 `n.mu`를 기존 exit 발신이 기다리는 경합이고 Q7이다 |
| 기본 설정 엔진(알림 off · 편입 off)이 무관리 보유 하나로 진입을 막는가? | **막지 않는다**(결정 (2)). off ∧ 미지정 · exclude는 normal이고, 알림 off에서는 어떤 a095 사실도 사슬(D1)에 들어오지 않는다 — 방식 Q1 |
| 편입된 · 보호 중 포지션에 「무보호」 critical이 나가는가? | 3판이 critical로 올리는 것은 B5(편입 대상인데 편입 안 됨)뿐이다. 수량 증가(보호 중)는 Q4, `ExternalPositionFound`는 생산 도달 불가 |
| 일시적 시세 실패가 critical이 되는가? | `adopt` B2 · B6 · B7의 연기분이 B5 사유로 모인다 — Q2(c) |
| a091과 충돌하는가? | 경계가 같다(D1). 다만 `criticalEvents` 크기를 고정하는 시험(2판 2.2a의 「19종」)은 병합 순서에 따라 틀리므로 3판은 수를 적지 않는다 |
