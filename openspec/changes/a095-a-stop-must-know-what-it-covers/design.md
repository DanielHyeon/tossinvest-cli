# a095 · 설계 — 10판 (구현 로트 — 열린 질문의 답, Manager 판정 2026-09-30)

> 분기 인용은 전부 `analysis/function-logic/`의 AST 산출물에서 온다(번들 31개 — 각 번들의 `ast.json`이 묶인 커밋은
> `analysis/harness/coverage/README.md`).
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
| B5 enabled 시도 실패 — **알림 켜짐 · 편입 시도(`adopt` B8 `d.adoptOne`)가 거짓**(아래 범주 ①②) | **critical** | 결정 (1)「발신은 reconcile 쪽에서만 critical」 + (2)「운영자가 고른 상태는 critical 에서 뺀다」 |
| 알림 off 엔진에서 생긴 모든 a095 사실 | **critical로 매기지 않는다**(→ ENTRY_BLOCKED에 닿지 않음) | 결정 (2). 4판 r3 N1 — 아래 「알림 off에서 무엇이 막혀야 하는가」 |
| `adopt` B2 · B6 · B7 연기분(오늘은 B5 사유로 모임) | **normal** — 조건 칸 「편입 켜짐 — 연기」 · 「include 지정 — 연기」로 분리(10판, Q2(c) 답) | 4판 r3 N2 · Manager 판정 2026-09-30 |
| B3 설정 거부 · B6 include 시도 실패(편입 꺼짐) | **normal** — 오늘 등급 유지(10판, Q2(a)(b) 답) | 거부 블록에 의도적으로 끈 엔진이 섞인다(9판 r8 N3) · `adoption.enabled` false 동등성 — Manager 판정 2026-09-30 |
| 알림 on이지만 transport가 죽음 | **이관 — 정본(a124)** | 사용자 결정 2026-09-28(Q2(d)) |

**a091과 같은 경계다.** a091 design(`design.md` 「승격하지 않는다」 절)은 *"경계는 '엔진이 그 포지션을
보호하기로 했는가'이고, 그 술어는 이미 코드에 있다"*고 적는다. B5는 `adoption.enabled`가 참인 — 엔진이
보호하기로 한 — 보유의 편입 실패다.

### 후보별 결과와 억제 키 (5판, r4 R4-1 · R4-3 — Manager 처분으로 편집 경계를 연다)

**문제**: `adopt`(`reconciledriver.adopt`)는 편입된 id 집합만 돌려주므로, 호출자 `judgeHoldings` B14에서 연기
(`adopt` B2 · B6 · B7)와 시도 실패(B8 — `d.adoptOne` 거짓)가 구별되지 않는다. 그리고 `alertUnmanaged`
(`reconciledriver.alertunmanaged`) B1 `:393` `if d.unmanaged[p.ID] {`는 포지션 id만으로 억제한다 — 해제는
`adoptOne` B3 창의 `delete(d.unmanaged, …)`(편입 성공)뿐이다(`reconciledriver.adoptone`). 그래서 Q2(c)에 「연기 =
normal」로 답하면, 앞 사이클의 연기 보고가 래치를 걸고 뒤 사이클의 시도 실패(critical)는 등급 판정 전에 B1에서
반환된다. 필수 critical이 사라지고, 재시작하면 결과가 바뀐다.

**편집 경계(Manager 처분 2026-09-29)**: `alertUnmanaged` B1 래치와 `adopt`의 결과 형태를 편집 경계에 넣는다(둘 다
번들이 이미 있다). 형태(후보별 결과 열거, 래치 키 구성)는 구현 로트가 정하되, Q2(c)의 어느 답(연기 = normal ·
critical)도 막지 않는 형태여야 한다.

**설계 원칙 — 6판(r5 R5-1, Manager 처분 2026-09-29가 5판 원칙을 대체).** 5판은 「억제 키는 (사실, 등급)이거나 등급이
오르면 풀린다」였고, 5라운드가 그 원칙이 **계속되는 critical 조건을 영구히 삼킨다**고 보였다: 만료 없는 메모리 래치 때문에
배달된 뒤 실패가 계속돼도 정본 재알림 창에 닿지 않고, 래치가 `d.alert` **앞**에서 걸리므로 기록이 실패해도
(`reconciledriver.alert` B2가 `Notify` 오류를 로그로만 남긴다) 저장소가 회복된 뒤 다시 기록되지 않는다. 6판 원칙:

1. **프로세스 메모리 억제 래치는 normal 보고 전용이다.** critical 보고는 래치를 무조건 지난다 — 매 관측이 durable 기록을
   시도한다.
2. **critical의 중복 판정은 outbox 키와 정본 재알림 창이 맡는다**(정본 engine-safety 「같은 조건의 critical 알림은 재알림 창
   안에서 한 번만 전송한다」 — 창 안의 같은 key는 `claimAndDeliver` B5 `ClaimSettled`로 흡수되고, 창이 지나면 `claimOwed`
   B7 창 `:410`이 재무장한다).
3. **근거: 판정이 둘이면 반증이 죽는다.** 정본 재알림 창이 이미 critical의 중복 판정 기계인데 그 앞에 메모리 래치를 겹치면
   같은 판정이 두 자리에 있고, 앞 자리가 뒤 자리를 가려 5라운드의 영구 삼킴이 생긴다. 판정은 상태가 사는 한 자리(원장)에 둔다.
4. **기록 실패는 삼키지 않는다** — 래치가 더는 막지 않으므로 다음 관측이 다시 기록을 시도한다(자연 성립). 사람 소유의
   진입 게이트 래치는 이 원칙과 무관하게 **불변**이다.
5. **normal 래치의 키는 사실 식별자다**(아래 「사실 식별자」). 같은 종목의 다른 사실은 normal 래치에 삼켜지지 않는다.

**귀결(기록만)**: critical이 매 관측마다 `Notify`를 부르면, 행이 PENDING인 동안 `claimOwed` B2 `:381`이 owed를 주고,
대사 goroutine은 **claim을 얻으면** 동기 배달을 시도한다(`claimAndDeliver` → `n.deliver`). 다른 발송자가 그 행의 임차를
쥐고 있으면 `claimAndDeliver` B6 `:294`(`case journal.ClaimHeldElsewhere:`)의 창 return `:305`로 배달 없이 돌아간다(7판
r6 R6-1 정정 — 6판은 「관측마다 동기 배달을 시도한다」로 과대 서술했다). 이것은 이관된 Q7의 **둘째 면**(대사
goroutine 자신의 원격 대기 — a092 영역의 이름 붙은 잔여)과 같은 실체이며 a095가 새로 정하지 않는다.

시험(전부 **생산 배선** — 실 `Notifier` · outbox · 배달 실행자): 「연기(normal) → 같은 프로세스에서 시도 실패(critical)」가
재시작 없이 기록된다(tasks 2.12) · **배달됨 → 재알림 창 경과 → 여전히 실패**면 다시 전송된다(2.15) · **outbox 기록 실패 →
저장소 회복 → 같은 실패**가 다음 관측에서 기록된다(2.16).

### 사실 식별자 (6판, r5 R5-2)

**계약**: 무관리 보고의 사실은 **(포지션, 조건)**으로 식별한다. 조건은 다음 일곱 칸이다(8판 r7b F8 — 모호함 제거):
운영자 제외 · 편입 꺼짐∧미지정 · 설정 거부 · include 지정 — 시도 실패 · include 지정 — 연기 · 편입 켜짐 — 시도 실패 ·
편입 켜짐 — 연기. 편입이 켜진 엔진의 include 지정 종목은 `alertUnmanaged`가 B5 `:409`(`Enabled`)를 B6 `:412`(`Included`)보다 먼저 보므로 **「편입 켜짐」 칸**에 든다(9판 r8 N8). **연기는 경로마다 조건 하나다** — 시세 읽기 오류(`adopt` B2) · 관측 없음(B6) · 관측 묵음(B7)의 셋은 **진단
원인**이다. 오늘 include 지정 후보의 연기는 `alertUnmanaged` B6(「시도했는데 이번 사이클 실패」) 사유로 알려지므로, 이 계약은
그 사유 문구와 다른 칸을 연다(문구 정정은 구현 로트). **진단 원인**(오류 문구, 연기의 세 원인, `adoptOne`이 로그에 남기는 거절
사유 등)은 식별자가 아니다 — 본문 · payload에 싣는다. 이 계약은 두 `alertUnmanaged` 발신 자리의 보고에만 적용되고, 편입 후 수량
증가 보고(같은 이벤트 종류)는 Q4에 남는다(8판 r7a F2). durable event key는 이 식별자와
정렬한다(exit 관측 자리와 reconcile 자리의 분리 — 결정 (3)(iii) — 위에 조건을 더한다). 철자(10판): 조건 칸은 `rejected` · `excluded` ·
`enabled_failed` · `enabled_deferred` · `include_failed` · `include_deferred` · `off_undesignated`이고, reconcile 자리의 key는
`<종류>|reconcile|<조건>|<posID>`다(critical은 종류가 `exit.position_adoption_failed`). exit 관측 자리의 key `exit.position_unmanaged|<posID>`는
그대로 두어도 이미 다르다 — 그 자리(`exitloop.go`)는 a095가 편집하지 않는다.

**왜**: 오늘 연기와 시도 실패는 같은 B5 사유 · 같은 key(`exit.position_unmanaged|<posID>`)를 공유한다. Q2(c)=critical이면
둘의 등급도 같아 「등급 상승」으로 가를 수 없고, 메모리 래치만 풀어도 같은 key의 행이 창 안에서 `ClaimSettled`로 흡수된다.
반대로 바뀌는 오류 문구를 새 사실로 치면 매 관측이 새 key가 되어 폭주한다.

시험(tasks 2.17): 같은 등급의 연기 → 시도 실패가 서로 다른 key로 기록된다(Q2(c)=critical 가정) · 창 안에서 A → B → A는
**A가 정착(DELIVERED)했으면** A의 재전송을 창 규칙대로 흡수하고, **A가 아직 PENDING이면** A는 재시도 대상이다(정본 — 전송
실패는 중복이 아니라 미완) · 다른 발송자가 A의 임차를 쥐고 있으면 이 관측은 배달하지 않는다 · 오류 문구만 바뀐 반복은
같은 key다. Q2의 등급 선택은 열어 둔다(7판 r6 R6-1).

**「시도 실패」의 경계 — `adoptOne`의 실패 세 범주**(`reconciledriver.adoptone`, 분기 3):

| 범주 | 분기 | 반환 | 5판 등급 |
| --- | --- | --- | --- |
| ① 편입 전 거절 | B1 `:320` — `exitpolicy.SyntheticStop` 실패, `AdoptPosition` 호출 전 | false | 시도 실패 — critical(알림 켜짐) |
| ② 영속 실패 | B2 `:339` — `AdoptPosition` 거절 · 원장 검증 · 트랜잭션 실패 | false | 시도 실패 — critical(알림 켜짐) |
| ③ 커밋 뒤 보호 미개설 | B3 `:344` 창 — `OpenAdoptedExitState` 실패에도 로그 뒤 `:379` **true** | true | **critical 요구 밖 — 이름 붙은 경계.** 이 함수가 성공으로 답하고, 주석은 다음 사이클이 개설을 마친다고 적는다. 그 자체가 후속 후보다(`issues.md` I7) |

B1 · B2는 **미진입**이다(커버리지 `analysis/harness/coverage/r5-app-engine.out`). 범주 ①②를 critical로 기록하는
시험(tasks 2.2)이 두 분기의 첫 진입이 된다.

### 싣는 방식 — Q1 답 (10판, Manager 판정 2026-09-30)

`SeverityOf`(`severityof`, 분기 1)의 B1 `:348`(현재 `event.go:359`) `if criticalEvents[t] {`는 종류만 본다. 그래서 **사실에 종류를
준다**: 새 종류 `obs.EventExitPositionAdoptionFailed = "exit.position_adoption_failed"`를 `criticalEvents`에 등재한다.
`SeverityOf` · `Notify` 본문은 바뀌지 않는다(Impact의 (a) 경우 — 경계 재선언 불요).

발신 자리(`ReconcileDriver.alertUnmanaged`)가 사실을 정한다:

```text
조건 = 사실 식별자 칸(아래) — 설정 거부 · 운영자 제외 · 편입 켜짐(시도 실패 | 연기) · include 지정(시도 실패 | 연기) · 편입 꺼짐∧미지정
critical ⇔ 알림 켜짐 ∧ 조건 = 「편입 켜짐 — 시도 실패」
  critical → 종류 exit.position_adoption_failed, 메모리 래치를 거치지 않는다(6판 원칙 1)
  그 밖   → 종류 exit.position_unmanaged(normal), 래치 키 = (포지션, 조건)
```

「알림 켜짐」은 `ReconcileDriverOptions.NotificationsEnabled`이고, 생산 조립 `Context.ReconcileDriver`가
`c.Config.Engine.Notifications.Enabled`로 **항상** 채운다(`Adoption`을 채우는 것과 같은 자리). 로드된 값이므로 거부된 블록은 거짓이다
(`mergeNotifications` B3). 전송기 유무(`Publisher`)는 읽지 않는다(5판 r4 R4-2).

`adopt`의 결과 형태는 id 집합에서 **후보별 결과**(편입됨 · 시도 실패 · 연기)로 바뀐다 — B8(`d.adoptOne` 거짓)이 「시도 실패」,
B2 · B6 · B7이 「연기」다. `adoptOne` · `ReconcileDriver.alert` 본문은 바뀌지 않는다(`adoptOne`의 `delete(d.unmanaged, id)`는 래치의 모양이
포지션 → 조건 집합으로 바뀌어도 그대로 포지션의 모든 조건을 지운다).

**critical 행 문구 — 시점 사건(Q8 답, 아래).** 제목 「<종목> 편입 시도 실패 — 보호가 열리지 않았다」, 본문은 「<시각>에 편입 시도가
실패했다. 그 시각 기준 이 보유에는 손절 · 익절이 걸려 있지 않았다 …」. 시각은 그 관측의 `d.clk.Now()`(RFC3339).

### 「알림 켜짐」의 판정 근거 (5판, r4 R4-2)

설정의 `notifications.enabled`로만 판정한다. `resolveNotificationPublisher`(`resolvenotificationpublisher`)가
전송기 nil을 내는 경우는 셋이다 — B2 `:77`(설정 거부) · B3 `:83` `if !cfg.Enabled {`(꺼짐) · B5 `:94`(켜짐 +
topic 없음). 그러므로 `Publisher == nil`은 「꺼짐」의 대용이 아니다. 켜짐 + topic 없음 엔진의 B5 사실은 critical이고,
그 행은 배달 실행자 `deliverOne` B8(전송기 nil → 실패 시도)과 a124 정본을 탄다. 시험: 켜짐 + topic 없음(critical) ·
꺼짐 + topic 유지(critical 아님)를 생산 배선으로(tasks 2.5 · 2.5a).

**거부된 알림 블록 = 꺼짐(8판, r7a F4 · r7b F3, Manager 처분).** `mergeNotifications`(`config.mergenotifications`) B3 `:105`
창이 검증에 실패한 블록을 `Notifications{Rejected: why}`로 만든다 — 로드된 `Enabled`는 거짓이다. 그러므로 파일에 `enabled:
true`를 적었어도 블록이 거부되면 a095에게는 「꺼짐」이고 a095 사실은 critical이 아니다(시험 2.5b). **R4-2 논거와의 방향 차이를
인정한다**: R4-2는 「전송 수단이 없다」를 꺼짐으로 읽지 말라고 했고(그러면 필수 critical이 사라진다), 여기서는 거부를 꺼짐으로
읽는다 — 거부는 전송 수단의 부재가 아니라 설정 자체가 무효라는 사실이고, 결과가 critical을 만들지 않는 안전 방향이므로 명시가
우선이다.

**거부된 편입 블록은 「운영자가 고른 상태」가 아니다(8판, r7b F1).** `mergeAdoption`(`config.mergeadoption`) B3 `:279` 창도
거부된 블록을 `Adoption{Rejected: why}`로 만들어 꺼짐 · 미지정과 모양이 같다. 정본 exit-policy는 범위 검증이 「enabled가
참이거나 include 목록이 비어 있지 않을 때」 요구된다고 적는다 — 델타는 정의에 「거부되지 않은」을 넣고 거부의 등급은 Q2(a)로
남긴다(정의로 닫지 않는다). **다만 거부된 엔진이 모두 보호를 요청한 엔진은 아니다(9판 r8 N3)**: `Adoption.validate`
(`config.adoption.validate`) B1 `:161`은 편입 꺼짐 · include 없음 · `DefaultStopPct == 0`일 때만 검증을 건너뛰므로, 의도적으로
끈 블록에 범위 밖 `default_stop_pct`가 남아 있어도 B2 `:164`에서 거부된다. 이 모양은 Q2(a)의 입력으로 기록한다(`proposal.md` Q2).

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
**들어오지 않을** 것을 요구한다.

**4판(r3 N1): 거르는 자리는 등급이다.** 배달 실행자도 같은 규칙을 따른다 — `alertDeliverer.deliverOne` B8
(`if d.Publisher == nil {`)은 전송 수단 부재를 실패 시도로 세고(`alertdeliverer.deliverone`), 정본 「배달 실행자는
지속 실패를 진입 차단과 운영 모드 승격으로 잇는다」가 그 셈을 진입 차단으로 잇는다. 따라서 알림 off에서 critical
행이 하나라도 생기면, 동기 경로를 우회해도 차단에 닿는다. 알림 off에서 생긴 사실은 **critical로 매겨지지 않아야**
한다 — `review.md` §3.7 R1과 `proposal.md` Q1의 제약과 같은 뿌리다. critical로 매긴 뒤 차단만 거르면 정본
「등급화된 알림」과 어긋나 MODIFIED가 필요하고, 결정 (2)는 MODIFIED를 금한다. 3판은 사슬 안에 거르는 분기를 더하지 않는다(그 함수들은 a092 · a124의
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

**6판이 첫째 면을 키운다 — 기록 + 착수 전 승인(8판 기록, 9판 재문언 — r7a F1 · r7b F5 · r8 N1, Manager 처분 2026-09-29).** 6판 원칙(critical은 메모리
래치를 지난다) 전에는 대사 쪽 critical 배달이 포지션당 프로세스에 약 한 번이었다. 6판 뒤에는 행이 PENDING이고 claim을 얻는
관측마다 대사 goroutine이 배달 예산 동안 `n.mu`를 쥘 수 있고(`claimAndDeliver` — `n.mu.Lock` `:254` · `n.deliver` `:309`),
그동안 exit goroutine의 기존 critical 발신(`exitloop.go:831` · `:1633` · `:1657` · `:1687`)이 같은 뮤텍스를 기다린다.
이것은 **범위가 아니라 구현 순서**의 문제다 — 결정 (1)(a095 설계 · 범위를 a092에 묶지 않음)은 그대로이고 a095는 a092와
독립이다. 구현 착수 순서는 Manager 스케줄링 사항이다. **기본 스케줄은 a092 「모든 보유자」 착지 이후다.** 그 **전에** 착수하려면 Manager 승인만으로는 부족하고 **사용자 확인**이 필요하다 — 추가되는 대기가 손절 루프의 `o.alert` 네 자리(`exitloop.go:831` 관측 두절 · `:1633` 판정 거부 · `:1657` 제안 거부 · `:1687` 청산 지연)를 지나 안전 불변식 4(손절 즉시성)에 닿기 때문이다. **깨질 때의 행동**: 착지 전인데 사용자 확인 기록이 없으면 착수 금지. 사용자 확인과 Manager 승인은 이 change의 `review.md` 「착수 승인 기록」 절에 적고, tasks 7.3 공시에 증폭 수용을 명기한다. 「a092 『모든 보유자』 착지」의 판정 기준은 a092 tasks **21.4**(GREEN — `claimAndDeliver` 잠금 범위)가 체크된 커밋이 이 브랜치 역사에 있고 그 커밋에서 a092 tasks **21.3 (f)**(다른 호출자의 동기 발송이 원격 전송 중일 때 exit 기록이 그 전송을 기다리지 않는다) 시험이 통과하는 것이다 — 그 커밋 해시를 기록에 적는다. (`proposal.md` 「이관 기록」 · tasks 0.7 · 7.3. 8판의 「a092 착지 이후 또는 같은 창을 전제로 한다 · 깨지면
P1」은 스펙 전제로 읽혀 선후 관계 절과 모순됐다 — 9판이 스케줄링 사항으로 다시 적었다.)

### 사실이 해소된 뒤의 행 — Q8 답: (b) 시점 사건 문구 (r3 N3 · 10판 Manager 판정 2026-09-30)

**지시**: 사실이 해소되면(다음 사이클에 편입 성공) PENDING 행은 기존 재개방/본문 갱신 기계로 갱신되거나
해소-정산된다 — **outbox의 기존 기계로 표현할 수 있을 때만** 설계로 세우고, 없으면 번호 질문으로 비운다.

**검사 결과: 맞는 기계가 없다.** 전부 번들의 분기다.

| 기계 | 번들 · 분기 | PENDING 행에 쓸 수 있나 |
| --- | --- | --- |
| 재무장 시 본문 교체(`outbox.go:294-345`) | `journal.recordalerttx` B4 `:295` `if rearm {` → B5 UPDATE(제목 · 본문 · payload 교체) | **아니오** — `rearm`은 `claimOwed`가 준다. `claimOwed` B2 `:379` `case AlertPending:`의 창 return `:381`은 owed=참 · rearm=거짓이다(`claimowed`). 재무장은 정착(DELIVERED · ACKNOWLEDGED) 행이 재알림 창을 지났을 때(B7 창 `:410`)와 날짜 없는 · 미래 · 모르는 상태(B5 · B6 · B8)뿐 |
| a089 R1 「최신 발생 반영」 계보 | a089 archive `design.md:94` (*"제목·본문·payload를 최신 발생으로 갱신한다"*) | **아니오** — 같은 archive `review.md:471`이 R1을 *"대체됨 — a096~a099, 다른(정본) 형태로"*로 적었다. 착지한 형태는 위의 「창 뒤 재무장」이고 PENDING 행은 대상이 아니다 |
| 사실 해소 정산 | `outbox.go` 함수 목록 — `MarkAlertDelivered` · `MarkAlertAttemptFailed` · `AcknowledgeAlert` · 임차 연산 | **없다.** 행 상태는 PENDING · DELIVERED · ACKNOWLEDGED이고, PENDING을 떠나는 길은 전달 성공과 운영자 승인뿐이다. 승인은 사람의 행위이며 자동으로 대신할 수 없다(`review.md` §3.8 codex N3 제안의 「사람 소유 래치를 조용히 풀지 말 것」) |
| 배달 실행자 | `alertdeliverer.cycle` B3 `:254`(`for _, alert := range pending {`)의 루프 몸체 — 생성 표의 좌표 창으로는 B4 `:258` 창의 `d.deliverOne`(`:261`) · `alertdeliverer.deliverone` B9 창 `Publish(Title: alert.Title, Body: alert.Body)` | 실행자는 PENDING 행을 원장에서 골라 **저장된 문구**를 보낸다 — 사실이 해소됐는지 묻지 않는다 |

따라서 오늘의 기계로는 편입 실패 critical 행이 PENDING인 채 편입이 성공하면, 그 행은 나중에 「지금 무보호」
문구로 배달된다(전달 실패면 정본대로 차단으로 이어진다). **편입 회복 자체는 그 행을 갱신하지도 정산하지도 않는다**
— 행을 정산하는 것은 배달 성공(DELIVERED, `MarkAlertDelivered` `outbox.go:449-456`)이나 운영자 승인뿐이고, 행의 상태와
사람 소유의 진입 게이트 래치는 **서로 다른 수명주기**다(5판 r4 R4-4 정정 — 4판은 「운영자 승인 전에는 PENDING을 떠나지
않는다」고 잘못 적었다). 이 절은
3판~9판에서 **Q8**로 비웠다(구현 로트의 정지 조건).

**10판 답 (b)**: 행을 갱신 · 정산하는 기계를 새로 만들지 않고, 행이 말하는 문장을 **그 시각의 실패 사건**으로 쓴다 — 「<시각>에 편입 시도가
실패했다. 그 시각 기준 보호가 없었다」. 편입이 나중에 성공해도 그 문장은 참이므로 늦은 배달이 거짓을 전하지 않는다. PENDING 행의 본문은
`claimOwed` B2 창(owed=참 · rearm=거짓)이 교체하지 않으므로 첫 실패의 시각이 남고, 재알림 창이 지난 정착 행의 재무장(B7 창)은
`recordAlertTx` B5가 제목 · 본문을 **그 관측의** 문장으로 바꾼다 — 에피소드마다 자기 시각을 말한다(정본 「재무장된 outbox 행은 통째로
이번 에피소드를 말한다」). 사람 소유 래치는 자동으로 풀지 않는다. 해소 뒤 PENDING 행의 전달 실패가 진입을 막는 것은 정본대로이며, 그
교환은 a124 정본(Q2(d) 이관)이 소유한다.

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

### (i) 엔진 개설 포지션 — Q3 답: 원장 조정의 순증 (10판, Manager 판정 2026-09-30)

`judgeHoldings` B7 창: 적격이면 B8만 검사하고 `continue`한다. 편입 기록이 없는 적격 포지션 — 엔진이
진입 결정으로 연 것 — 은 수량이 늘어도 아무 검사를 받지 않는다. 결정 (3)(i)이 이것을 범위에 넣는다.

`exit_states`에는 수량 열이 없다(`core_domain.go:203-230`). **비교 기준은 원장 `position_adjustments`다.**

- **「외부 증가 사실」의 의미 — 쓰기 코드에서.** 투영은 체결로 만들어지고, 계좌가 체결로 설명되지 않는 수량을 보이면 대사 수렴이
  조정 행 하나를 **추가**하고 투영을 계좌 값으로 옮긴다: `converge.go:209-227`(`ApplyPositionAdjustment` — `ExpectedPrevQuantity:
  mismatch.Local` · `NewQuantity: mismatch.Authority()`), `position_adjustments.go:304-316`(행의 `prev_quantity`는 투영이 들던 값
  `held`, `new_quantity`는 계좌 값) · `:332-343`(INSERT) · `:346-350`(같은 트랜잭션의 투영 UPDATE). 살아 있는 인스턴스는 **제자리에서**
  조정된다(`:281` `adjustInPlace`). 그러므로 한 포지션 인스턴스의 조정 행들의 `Σ(new_quantity − prev_quantity)`는 **현재 투영 수량 −
  체결이 설명하는 수량**이다.
- **술어**: 엔진 개설 포지션(`p.ExitEligible() ∧ ¬p.Adopted()`)은 그 순증이 **0보다 클 때** 수량이 체결로 설명되지 않게 늘었다.
- **읽기**: journal 읽기 전용 새 leaf `Journal.NetAdjustedQuantity(ctx, positionID)` — 그 인스턴스의 조정 행을 읽어 순증을 십진 문자열로
  돌려준다. 스키마 무변경, 쓰기 없음, 브로커 호출 없음.
- **이름 붙은 잔여**: 체결 기록이 계좌보다 늦게 들어오면 수렴이 먼저 +조정을 쓰고 뒤의 체결이 −조정을 부른다 — 그 사이의 순증 > 0은
  일시적 오보다. 등급이 normal(Q4)이므로 진입 · 청산에 닿지 않고 보고 한 줄이다.

### (ii) 수량 증가 사실의 종류·등급과 키 — Q4 답: normal (10판, Manager 판정 2026-09-30)

오늘의 키는 `…|grown|<posID>`이고 수량이 없다. 그 알림의 본문은 스스로 *"늘어난 수량은 원래 수량
기준으로 산정된 손절의 보호를 받는다"*고 쓴다 — **무보호가 아니다.** 그런데 종류는
`EventExitPositionUnmanaged`이고 제목은 *"고정된 t0가 증가분을 덮지 않는다"*다.

`claimAndDeliver` B5 `:285` `case journal.ClaimSettled:`가 같은 키의 재전송을 재알림 창 안에서 삼키는 것은
**critical 경로에서만** 일어난다. 따라서 2라운드 P0-2는 이 사실이 critical일 때의 문제다. 종류·등급이
정해지면: critical이면 키를 정본 engine-safety 「같은 조건의 critical 알림은 재알림 창 안에서 한 번만
전송한다」와 대조해 정하고, normal이면 재알림 창과 무관하게 `d.grown`(B1) 래치의 기준만 정한다.

**10판 답**: 수량 증가 사실은 **normal**이다. 종류(`EventExitPositionUnmanaged`)와 key 철자(`…|grown|<posID>`)는 편입 포지션 자리에서
그대로이고 엔진 개설 포지션도 같은 모양을 쓴다(normal 경로는 outbox에 닿지 않으므로 P0-2의 창 흡수가 없다). **래치는 (포지션, 보고한
최대 수량)이다** — 보고한 최대 수량보다 현재 수량이 클 때만 다시 보고한다. 그래서 475150의 원장 순서(편입 2 → 3 · 4 · 5 · 8 · 26 · 32)는
3 · 4 · 5 · 8 · 26 · 32를 각각 한 번씩 보고하고 32가 운영자에게 전해진다(tasks 3.4). 같은 수량의 반복과 감소는 보고하지 않는다.

**normal의 전제(Manager 조건, 영수증은 `review.md` §4.4)**: 손절이 발동하면 제안 수량은 그 시점의 투영 수량 전량이다 —
`RemainingQuantity: m.position.Quantity`(`exitloop.go:1092`) × 손절 비율 `"1"`(`ratchet.go:430` · `ladder.go:446`) →
`ProjectWholeShares`(`snapshot.go:137` · `:206`), 투영 수량은 조정이 계좌 값으로 수렴시킨다(`position_adjustments.go:346-350`). 이
사슬이 바뀌면 이 등급은 다시 판정되어야 한다.

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
| 최초 INSERT | `journal.openexitstate` | 진입 손절 | — (행을 **만든다** — 기존 값을 낮추는 쓰기가 아니다) |
| 판정 UPDATE | `journal.recordexitjudgementtx` | 판정의 기준선 | 옛 경로: B23 `:487` `if recomputed == nil {` → B25 `:491` `notBelow("baseline", …)` — **스칼라**와 비교. 이 경로는 B37 `:557` `if effective != nil {`이 거짓이라 스칼라 열만 쓰고 effective JSON은 다시 쓰지 않는다. 스냅샷 경로: B28 `:506` → B29 창 `exitpolicy.SelectRecoverySnapshot` — **저장된 effective 스냅샷**과 비교하고, 저장 스냅샷이 없으면(`selectrecoverysnapshot` B2 `:138`) 비교 없이 재계산값을 받는다 |
| 관측 갱신 UPDATE | `journal.refreshexitobservation` | 저장된 effective 스냅샷의 보호가 | B23 `:127`이 `sameExitOperationalLine` 거짓이면 거절 — 비교 대상은 effective JSON이다. **스칼라와 effective 스냅샷이 일치할 때만** 값이 움직이지 않고, 갈라져 있으면 스칼라를 스냅샷 보호가로 되돌린다(낮출 수 있다). 그 갈라짐에 생산이 도달하는지는 재지 않았다 |
| 재편입 reset UPDATE | `resetexitstateforreadopttx` | 재편입 관측의 합성 손절 | 분기 여섯(B1~B6) 모두 오류·행 수 검사 — **이전 기준선과 비교하는 분기가 없다 — 낮출 수 있다.** 운영자 행동(`positionpolicy.ActionReadopt`)에서만 불린다 |

주석의 *"the only reset writer for the four guarded execution-time columns"*에서 네 열은
`taken_ratio_total` · `pending_action` · `pending_level` · `pending_intent_id`이고 손절 열이 아니다(보이스 A
A-4). 2판이 이것을 「유효 손절의 유일한 쓰기 경로」로 읽은 것이 P0-3이다.

### 래칫 상향의 선행 조건 — Q6 답: 후속 change로 (10판)

2판의 선행 조건 셋 중 첫째(「쓰기 경로 하나」)는 거짓이었고, 둘째(「하향 거부」)는 판정 경로에 이미
있다(위 표). 참인 문장으로 다시 SHALL을 세울지, 위 표를 issues I1에만 두고 후속 change에 넘길지가 Q6이다.
그때까지 exit-policy 델타는 「평단이 내려도 유효 손절가를 낮추지 않는다」만 싣는다.

**10판 답 (Manager 판정 2026-09-30)**: 선행 조건을 SHALL로 다시 세우지 않는다. 쓰기 자리 넷의 사실은 `issues.md` I1에 두고 불타기 래칫을
도입할 후속 change가 받는다(tasks 5.4).

## D5. 셋의 상호작용 (3판)

```text
reconcile 사이클 — judgeHoldings
  ├─ B7 적격 ──┬─ B8 편입됨 → checkExternalIncrease ── D2 (Q4)
  │           └─ 편입 안 됨(엔진 개설) → continue ── D2 (i) (Q3)  ← 오늘의 구멍
  ├─ B9·B10 전이 상태 → continue (무알림, 유지)
  ├─ B11 exclude ─────────┐
  ├─ B12 off ∧ 미지정 ────┤→ unmanaged → B15 → alertUnmanaged
  └─ B14 편입 안 됨 ───────┘              ├─ B4 · 기본 사유 → normal (결정 2)
        ↑ adopt B2·B6·B7·B8                ├─ B5(알림 켜짐 · adoptOne 실패) → critical (결정 1·2) — 방식 Q1 · 행 해소 Q8
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
| 일시적 시세 실패가 critical이 되는가? | `adopt` B2 · B6 · B7의 연기분은 오늘 B5 사유로 모이지만, 4판 델타는 critical 요구에서 **연기분을 뺐다**(r3 N2) — 등급은 Q2(c) |
| 편입 실패 critical 행이 편입 성공 뒤에도 남는가? | **남는다 — 오늘의 기계로는**(D1 「사실이 해소된 뒤의 행」). Q8, 구현 로트의 정지 조건 |
| a091과 충돌하는가? | 경계가 같다(D1). 다만 `criticalEvents` 크기를 고정하는 시험(2판 2.2a의 「19종」)은 병합 순서에 따라 틀리므로 3판은 수를 적지 않는다 |
