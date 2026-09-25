# a124 design — 판정은 원장이 커밋한 시도 수에 서고, 승인이 이긴다

> **5판 (2026-09-26, Teammate)** — freeze 1~4판 REJECT(`review.md` §0 · §0.2 · §0.3 · §0.4) 반영. Manager 결정 M1 = **B′**(2026-09-26 재결정,
> C 에서 바꿈), M2 = **(ii)**. 4판 → 5판: D7 을 `EntryGate` 의 사유별 **해제 세대**로 다시 썼다 — 실행자는 `n.mu` 를 잡지 않고 `Acknowledge` 는
> 무편집(4판의 한 줄 제거). D8 계수의 수명을 해제 세대에 묶었다. Q4 · Q5 · Q6 반영(D7 · D8 · D6).
> 5판 2쇄(freeze 5회차 §0.5 V1~V6): 세대를 임차 **전후로 두 번** 읽고(V1), 울타리가 실패하면 버리기 전에 행이 아직 미전달인지 원장에서
> 다시 확인한다(V2 — 해제의 「미전달 0」은 해제 순간이 아니라 그 앞의 셈이다). spec 의 잠금 문장을 「경계 있는 게이트 구간 허용」으로(V4).
> 5판 3쇄(freeze 6회차 §0.6 W1~W5): 판정을 버리는 근거는 원장의 **ACKNOWLEDGED** 뿐(전달은 사람 확인이 아니다, W1), 경합이 이어지면
> 무조건 잠그지 않고 **다음 사이클로 미룬다**(W3), 나열 계수 울타리(W2), 「잠금만」 판정 운반(W4).
>
> **⚠ freeze 7회차 REJECT (`review.md` §0.7 X1~X5)** — 3쇄의 「원장 상태로 버릴지 정한다」는 W1 과 X2 가 같은 DELIVERED 에서 반대 답을 요구해
> 성립하지 않는다. 판정 원칙 E(늦은 적용 = 제때 적용, 해제와 근거의 **시간 순서**로 가른다)를 Manager 에게 올렸다 — 확인 전까지 D7 을 고치지 않는다.
>
> **Manager 지시와 다른 점 하나 (확인 요청)**: 해제 세대를 「`Clear` 가 **실제로 지웠을 때만**」이 아니라 「그 사유의 **해제 요청마다**(래치 유무 무관)」
> 올린다. 지시대로면 판별 케이스 ㉣(래치가 서기 전에 운영자가 backlog 를 비움)에서 승인 뒤 재잠금 + durable 승격이 남는다(D7). `revision`
> (전략 ABA 봉인)은 지시·기존대로 실제 삭제 때만 오른다.
>
> 분기 주장의 근거는 `analysis/function-logic/` AST 번들뿐이다. 2판 이후 새로 기대는 함수 셋(`Journal.settleUnderClaim` ·
> `Notifier.Acknowledge` · `Journal.TransitionOperatingMode`)의 번들은 이 문서보다 먼저 base `4798d399` 에서 추출했다.

## Context

a098 이 critical outbox 를 비우는 보조 실행자를 세웠고(감독 밖, 결정 9-2), a099 가 배제를 원장 임차로 옮겼다.
둘 다 「지속 실패 → 진입 차단·모드 승격」을 실행자에 두지 않았다 — 오늘 그 일은 동기 경로 `Notifier.deliver` 가
한다(proposal 표). a092 가 그 동기 시도를 엔진 루프에서 빼면 그 일의 주인이 사라진다. 이 change 가 그 주인을
실행자로 정한다.

## D1 — 주인은 실행자, 판정 입력은 **원장이 커밋한 증가 후 `attempts`**

| 후보 | 재시작 | 두 발송자 | 채택 |
|---|---|---|---|
| 실행자 메모리의 연속 실패 사이클 수 | 지워진다 | 각자 센다 | ✗ |
| 나열 행의 `alert.Attempts` (+1) | 남는다 | **낡는다** — 배치 뒤쪽 행은 나열 뒤 최대 ~10×10 s 에 임차되고, 그 사이 다른 발송자가 올리거나 재무장이 0 으로 되돌린다(`outbox.go:337`) | ✗ (F1) |
| **시도 기록 트랜잭션이 커밋한 값** | 남는다 | CAS(id · PENDING · 토큰) 아래의 값이라 이 에피소드·이 임차의 것 | **✓** |

HEAD 는 이 값을 돌려주지 않는다: `settleUnderClaim` 의 적용 경로 B5 는 `SettleResult{Outcome: SettleApplied}` 만 돌려주고
(FLM `internal-journal--journal.settleunderclaim` B5, `alert_claim.go:330-334`), `SettleResult` 에 attempts 필드가 없다(:140-145).

**journal API (additive).** `SettleResult` 에 `Attempts int` 를 더한다. `settleUnderClaim` 의 B5 경로에서 커밋 **전에** 같은
트랜잭션으로 `SELECT attempts FROM alert_outbox WHERE id = ?` 를 읽어 채운다. 적용되지 않은 결과(`LeaseLost` ·
`AlreadySettled` · `NotFound`)와 오류에서는 0 이며 **판정에 쓰지 않는다.** 읽기 실패는 롤백 후 오류다(정산하지 않은 것 —
B3 과 같은 결과). 읽기는 반드시 **그 트랜잭션**으로 한다(`tx.QueryRowContext`) — 원장 연결이 하나라 `j.db` 로 읽으면 자기 트랜잭션이 쥔 연결을 기다린다.
스키마 무변경, 세 호출자(`MarkAlertDelivered` · `MarkAlertAttemptFailed` · `ReleaseAlertClaim`)의 SQL·CAS 불변,
기존 호출자는 새 필드를 무시한다. `RETURNING` 대신 트랜잭션 안 SELECT 를 쓰는 이유: 같은 연결·같은 트랜잭션이라 값이 같고,
드라이버 판본 의존을 만들지 않는다.

**판정.** 정산 결과의 `err` 를 **먼저** 본다 — 오류의 `SettleResult{}` 는 `Outcome` 영값이 `SettleApplied` 다(`alert_claim.go:108`, F10).
두 정산 모두 판정한다. 동기 경로와의 대응은 FLM `internal-obs--notifier.deliver`(B7~B18)와 `notifyCritical` B4 에서 가져왔다
(`claimAndDeliver` 는 `lost` 면 `owed=false` 로 돌려 승격을 막는다 — `notifier.go:309-315`).

**실패 기록** (`MarkAlertAttemptFailed`, `deliverOne` B8 · B9 끝):

| 결과 | 판정 | 동기 경로 대응 |
|---|---|---|
| `err != nil` | D8 계수 | B13·B15 — 로그 후 계속, 소진 시 잠금 |
| `Applied`, `Attempts < 한도` | 없음 | 다음 시도 |
| `Applied`, `Attempts >= 한도` | **잠금 + 승격** | B26·B27 잠금 + `notifyCritical` B4 승격 |
| `AlreadySettled` · `LeaseLost` | **없음** — 승인(또는 남의 발송)이 먼저 | B14·B16 → `lost` → 잠금·승격 없음 |
| `NotFound` | **잠금만** (승격 없음) | B18 `NotFound && Gate` → `Block` :520, `lost=true` → 승격 없음. 2판의 「승격 parity」는 **틀렸다**(N6) |
| 모르는 값 | **잠금만** | 동기 경로는 이 자리에서 잠그지 않는다 — 보수 방향으로 한 칸 더(fail-safe) |

**전달 정산** (`MarkAlertDelivered`, `deliverOne` B11 · B12 — 2판이 빠뜨린 자리, N3):

| 결과 | 판정 | 동기 경로 대응 |
|---|---|---|
| `Applied` | 없음 | B7 |
| `AlreadySettled` · `LeaseLost` | **없음**, 임차 유지 | B8 — 「게이트 사건 아님」(:440-451) |
| `NotFound` · 모르는 값 | **즉시 잠금 + 승격**, 임차 유지 | B9·B10 → 즉시 `Block` :484, `false,false` → `notifyCritical` B4 승격 |
| `err != nil` | **즉시 잠금 + 승격**(D7 해제 세대 울타리를 통과할 때), 임차 유지 | B11·B12 → 즉시 `Block`. 동기 경로는 `n.mu` 를 발행부터 쥐어 그 사이 승인이 끼지 못하지만, 실행자는 `n.mu` 를 잡지 않으므로 승인이 발행과 정산 사이에 낄 수 있다 — 오류는 그것을 가리므로 울타리가 필요하다 |

전달 정산 실패를 D8 계수가 아니라 **즉시** 잠그는 것은 동기 경로 parity 다(「발행은 됐는데 기록이 안 됨은 정산되지 않은 것」, :465-476).
임차를 쥔 채 두는 것도 같다 — 반납하면 다음 사이클이 이미 나간 알림을 곧장 다시 보낸다(2026-08-08 폭풍, :486-490). 보장의 크기는
**임차가 살아 있고 교체되지 않은 동안**의 재발행 억제뿐이다. 임차(81 s)가 만료되면 다음 사이클이 다시 잡아 재발행할 수 있다 — 동기 경로도
같다(「Expiry lets go eventually」, :486-490). 배치 서비스가 81 s 를 넘으면 바로 다음 사이클일 수도 있다(R6).

게이트는 멱등이고(`Block` 은 **없을 때만 삽입**, detail 은 처음 것이 남는다 — `execgw/retry.go:526-533`, F9), 승격도 멱등
(FLM `internal-journal--journal.transitionoperatingmode` B15: 같은 모드면 행을 안 쓰고 `changed=false`, B16: 더 엄한 모드면 그대로).
그래서 한도 이상의 행이 매 사이클 다시 시도돼도 폭풍이 아니라 같은 사실의 재확인이다(비용: 실패 행마다 BEGIN IMMEDIATE 읽기 한 번 — F12).

## D2 — 한도 값: **ㄱ `attempts >= 3` (확정)**

상수 하나 `alertAttemptLimit = obs.DefaultCriticalAttempts` 를 두고 시험은 그 상수를 인용한다(숫자를 복사하지 않는다).

계산값(측정 아님; 원장 쓰기·반납 대기 제외). 기본값: `DefaultCriticalAttempts = 3` (`notifier.go:45`) · `DefaultRetryDelay = 2 s` (`:48`)
· `DefaultPublishTimeout = 10 s` (`alert_lease.go:17`) · `alertDeliveryInterval = 2 s` · 배치 10 (`alertdelivery.go:63·74`).

| 상황 (한도 3) | 동기 경로 | 실행자 |
|---|---:|---:|
| 행 하나, 즉시 실패 | 4 s | 4 s |
| 행 하나, 매번 10 s timeout | 34 s | 34 s |

같은 실패 양상에서 한도 3 은 동기 경로와 **같은 시간**이다. 1판 proposal 의 「6 s 대 34 s」는 서로 다른 양상을 비교했다(F11 — proposal
교정). ㄴ(등가 시간)은 한 값으로 정의되지 않는다: 17 회면 즉시 실패 32 s 지만 timeout 이면 `12×17−2 = 202 s` 로 동기 경로보다 덜
보수적이다. a092 델타도 「시도 횟수를 줄이지 않는다(SHALL NOT)」(`a092…/spec.md:74`). 한도를 **늘리는** 안은 진입 차단 완화이므로
그때만 사람 결정이다.

## D3 — publisher 부재는 실패 시도로 센다 (확정)

B8(`Publisher == nil`)도 `MarkAlertAttemptFailed(…, "no publisher is configured")` 를 거쳐 D1 판정을 탄다(고정 원인 문구 — D9).
근거: `newNotifier` 문서 *"nil publisher … the entry gate latches, and sustained failure escalates to ENTRY_BLOCKED. That is the
specified direction"* (`exitwiring.go:60-70`), 동기 경로 `deliver` B3 → B26·B27, `notifyCritical` B4, a092 델타 :76 · 시나리오 :108.
세지 않는 안은 무설정 엔진이 영구히 진입을 열어 두는 구멍이다.

지연은 같지 않다(F13): 동기 경로는 publisher 가 없으면 첫 반복에서 `break` 후 곧장 잠근다(≈0 s), 실행자는 한 사이클 한 시도라
≈4~6 s. 규칙을 하나(「사이클당 한 행 한 시도」, a092 델타 :74)로 두기 위해 이 차이를 받아들인다.

## D4 — 굶주림 없는 선택

`PendingAlerts` 를 `ORDER BY (attempts >= ?), id` 로 바꾼다(한도 아래 먼저, 그 안에서 오래된 것 먼저). 한도는 D2 의 상수를 인자로
받는다 — SQL 에 값을 박지 않는다. 버리는 경로는 없다. 호출자 4 중 셋(`alertOps.Pending` · `Flush` · `Acknowledge`)은 `limit 0` 이라
**집합은 불변이고 순서만** 바뀐다(`evidence-reconciliation.md` R2). 운영자 목록(`alertOps.Pending` → `engine alerts`)의 표시 순서가
바뀌므로 그 출력 시험을 1.4 에서 확인한다.

R2 가 막는 것과 못 막는 것(F5, 기록 — 수용):

| 경우 | 결과 | 유계? |
|---|---|---|
| 한도 행이 배치를 채우고 새 행이 온다 | 새 행이 다음 사이클 먼저 — **R2 의 대상** | 예 |
| 한도 아래 `HeldElsewhere` 행 10 개가 선택을 채운다 | 시도 없는 사이클 | 쥔 발송자가 죽었으면 예 — 임차 만료(`DefaultAlertLease = 81 s`, `alert_claim.go:34`) 뒤 탈취. 살아서 정산하면 그 발송자가 판정한다. 임차를 반복 갱신하는 발송자는 무계(D6) |
| 한도 층의 오래된 10 행이 영구 실패(독 행) | 그 뒤 한도 행은 재시도 안 됨 | 아니오 — 단 게이트는 이미 잠겼고 행은 미전달로 세어진다. 전달만 늦다 |
| 한도 아래 행이 끝없이 온다 | 한도 행은 잔여 자리를 못 얻는다 | 아니오 — 설계 의도(새 알림 우선), 게이트는 이미 잠김 |

임차 가능 여부를 정렬에 넣는 안(`claim_token = ''` 먼저)은 YAGNI 로 기록만 한다.

## D5 — 배선

`alertDeliverer` 에 `Gate *execgw.EntryGate` · `AccountRef string` 을 더하고 `Context.AlertDeliverer`(`auxiliary.go:156-190`)에서 `c.Entry` ·
`c.AccountRef` 를 넘긴다(`c.Entry` 는 이미 필수 — :157). `Journal` 은 이미 있다. **`Notifier` 는 넘기지 않는다** — 실행자는 `n.mu` 와 무관하다(D7).
승격은 `Journal.EscalateOperatingMode(ctx, AccountRef, ModeTriggerCriticalAlertUndelivered, nil)` — `Notifier.escalate` 와 같은 호출(:382)이고
announcer 는 nil(실행자는 `Notify` 를 재진입할 이유가 없다; 전달 수단이 막 실패한 참이다 — `notifier.go` escalate 문서).

## D6 — 래치 시간: 조건부 식 (R3, 2판 재유도 — F4 · N5)

**일반 상한은 없다.** 아래는 전제를 이름으로 단 조건부 식이고, 전제가 깨지는 경우는 표 뒤에 따로 적는다. 2판이 「≤ … (상한)」과
781 s 를 운영 상한처럼 쓴 것은 철회한다(N5).

기호: `L` 한도(3) · `I` 사이클 대기(2 s) · `B` 한 사이클이 시도하는 행 수(≤ 10) · `T` 한 행의 발행 실패 시간(즉시 ≈ 0, timeout 10 s —
transport 가 소유, `ntfy.go:95-100`) · `S` 한 행의 원장 비용(임차 · 정산 · 반납) · `M` D7 해제 세대 읽기·울타리의 `g.mu` 대기 · `E` 판정 뒤 승격·로그 비용.

실행자는 **사이클 먼저, 대기 뒤**(`alertdelivery.go:122-136`)이고 한 사이클은 선택한 행을 차례로 시도한다(`:154-162`). 한 행의 재시도
사이에는 **같은 사이클의 다른 행 전부**가 낀다 — 1판 식이 이것을 빠뜨렸다.

**전제 H (동질 두절)** — 전부 성립해야 식이 선다(Q6): ① transport 가 끊겨 모든 발행이 같은 `T` 로 실패한다 ② 두절 시점에 모든 PENDING 행의
`attempts` 가 한도 아래이고 사이클마다 선택되는 행 수 `B` 가 고정이다(새 행 유입 없음) ③ 임차 경합이 없다 ④ 그동안 해제 요청이 없다(D7 울타리가
판정을 버리지 않는다) ⑤ 원장 정산이 오류 없이 `Applied` 로 끝난다 ⑥ `S`·`M`·`E` 가 유한하다. 이때 두절 뒤 첫 사이클의
첫 행은 한도 아래 행 중 가장 오래된 것이고(R2 정렬), 한도에 닿을 때까지 그 자리를 지킨다(더 새 행은 id 로 뒤). 게이트는 **어느 행이든**
처음 한도에 닿을 때 잠긴다. 그래서

- 사이클 길이 `C = B·(T + S + M + E) + I_list + I` (`I_list` = 나열 비용)
- 두절 뒤 첫 사이클 시작까지 `Q ≤ C` (진행 중 사이클의 남은 부분 + 대기)
- **래치 ≈ `Q + (L−1)·C + I_list + (T + S + M)`** (사이클 시작 기준; 마지막 사이클의 나열 뒤 첫 행의 L 번째 시도 끝에서 울타리가 성공하는 순간.
  전제 H 아래의 **조건부 상한**이고, 등호는 두절 순간 모든 행의 `attempts` 가 0 일 때만)

| 예시 (전제 H, `S = M = E = I_list = 0`) | 첫 사이클 시작부터 | `Q` 최대 포함 |
|---|---:|---:|
| 행 1, 즉시 실패 (`C = 2`) | 4 s | 6 s |
| 행 1, 10 s timeout (`C = 12`) | 34 s | 46 s |
| **배치 10 행 전부 timeout** (`C = 102`) | **214 s** (끝 행 304 s) | **316 s** |

1판의 `100 + 3×2 + 3×10 = 136 s` 는 첫 행 기준 214 s 보다 작았다.

**전제 H 가 없을 때 (이질 실패 — 특정 행만 실패)**: 실패하는 행보다 오래된 한도 아래 행 `K` 개가 **성공하며** 빠지는 동안 기다린다 —
`Q ≤ C + ⌈K/B⌉ · C_ok` (`C_ok` = 성공 행들의 사이클 길이). `K` 는 작업량이 정하므로 이 change 가 값을 줄 수 없다.

**유한한 값이 없는 경우** (이름만 적는다 — a092 델타 「보장이 아니다」와 같은 결): 원장 정지(연결 하나 — `journal.go:174`; `S` 무계),
`g.mu` 를 쥔 다른 보유자의 시간(`M` — 전략 진입 dispatch 는 브로커 전송 동안 쥔다: `execgw/strategy_entry_gate_authority.go:60-72`; 전략 진입은 이 빌드에서
휴면. 이것은 **실행자의 판정만** 늦춘다), 임차를 갱신하며 쥐는 발송자, `T` 에 상한이 없는 `Publisher` 구현.

임차 만료는 그 자체로 정산을 무효로 만들지 않는다 — 정산은 토큰·상태 CAS 만 본다(`outbox.go:455·472`). 만료 뒤 **다른 발송자가 잡아 토큰이
바뀌었을 때만** `LeaseLost` 가 된다(FLM settleUnderClaim 종단, `alert_claim.go:349-365`). 3판의 「`M ≤ 54 s` 라 81 s 안」은 철회한다(R7). 5판의
`M` 은 `g.mu` 대기다. 해제 세대 읽기는 임차를 쥔 채 `g.mu` 를 기다릴 수 있다 — 그 대기가 임차(81 s)를 넘기면 만료 뒤 탈취가 가능하고, 그러면 정산이
`LeaseLost` 라 판정이 없다(다음 사이클에 다시). 울타리는 반납 **뒤**라 임차에 걸리지 않는다.

tasks 4.4 는 예시 세 줄을 실측해 review 에 적는다. a092 델타의 식 `다음 배달 사이클까지의 시간 + 시도 횟수 × 1회 상한 + (시도 횟수 − 1) ×
사이클 주기` 는 「사이클 주기」를 `C` 로 **재정의하면 발행 시간을 이중으로 센다**(`Q + L·T + (L−1)·C`, Q6). a092 21판에는 **식을 교체**하라고
전한다: `Q + (L−1)·C + I_list + (T + S + M)` 와 전제 H(tasks 5.2).

## D7 — 승인이 이긴다: `EntryGate` 의 사유별 해제 세대 (M1 = B′, F2 · N1 · R1 · R4 · Q1 · Q2 · Q3 · Q4)

**문제.** 실행자의 `Block` 이 운영자의 해제 **뒤에** 떨어지면 빈 backlog 위에 래치가 서고 운영자가 방금 본 사건이 차단으로 되살아난다. 이미 정해진
규칙이 이것을 금한다: a092 델타 「운영자의 승인은 … 전송의 성공 여부는 그 사유를 되살리지 않는다(SHALL NOT)」(`a092…/spec.md:66-68`, 사용자 결정
2026-08-10 결정 2 = 안 1).

**C 가 떨어진 이유 (§0.3 R1 · §0.4 Q1).** `Notifier` 의 `n.mu` 를 빌리는 모든 형태는 손절 경로(exit `Notify` — `exitloop.go:1710`, 비상 청산 —
`flatten.go:694`)가 취소 없이 기다리는 뮤텍스 안에서 무언가를 기다린다. 3판은 원장을, 4판은 `g.mu` 를 — 그리고 전략 진입 dispatch 는 `g.mu` 를
브로커 전송 동안 쥔다(`execgw/strategy_entry_gate_authority.go:60-72` → `gateway.go:692-708`). 그래서 5판의 실행자는 **`n.mu` 를 잡지 않는다.**

### 수단 (execgw — High-risk 표면)

`EntryGate` 에 필드 하나, 메서드 둘, `Clear` 한 줄:

```go
clearEpochs map[ReasonCode]uint64          // 지연 생성, g.mu 아래에서만

func (g *EntryGate) Clear(reason ReasonCode) {
	g.mu.Lock(); defer g.mu.Unlock()
	g.bumpClearEpochLocked(reason)             // ← 추가: 해제 요청마다 +1 (래치 유무 무관)
	if _, exists := g.latches[reason]; exists { // B1 — 기존 그대로
		delete(g.latches, reason); g.revision++ //   revision 은 실제 삭제 때만 (기존 의미 불변)
	}
}
func (g *EntryGate) ClearEpoch(reason ReasonCode) uint64                                   // g.mu, 읽기만
func (g *EntryGate) BlockUnlessClearedSince(reason ReasonCode, epoch uint64, detail string) bool // g.mu
	// clearEpochs[reason] != epoch 이면 아무것도 안 하고 false.
	// 같으면 Block 과 같은 규칙(없을 때만 삽입 + revision++, FLM entrygate.block B1)으로 true.
```

두 메서드는 `g.mu` 만 잡고 밖을 부르지 않는다(FLM `internal-execgw--entrygate.block` · `.clear` — 둘 다 map 연산뿐).

**`ReasonAlertUndelivered` 를 푸는 길은 `Clear` 하나다**: 이 사유를 `g.latches` 에서 지우는 다른 코드는 없고(직접 삭제는 `modegate.go:37` 의
`ReasonOperatingModeBlocked`, `symbolgate.go:189` 의 대사 계열뿐 — grep), 이 사유로 `Clear` 를 부르는 곳은 `Notifier.Acknowledge` 의 두 자리 —
B3(무원장, :846)과 B10(**승인 뒤 미전달 수 = 0**, :875) — 뿐이다(grep; FLM `internal-obs--notifier.acknowledge`).

**`Acknowledge` 는 편집하지 않는다.** 해제 조건(승인 + 미전달 0)도 그대로다.

### 실행자의 순서

```text
deliverOne(row)
  e1 := Gate.ClearEpoch(R)          ← 임차 전 (g.mu, map 읽기)        R = ReasonAlertUndelivered
  claim                             ← 잠금 없음 (원장)
  e2 := Gate.ClearEpoch(R)          ← 임차 직후
  e1 != e2 → 임차 반납, 이 행은 이번 사이클에 보내지 않는다         (V1 — 해제가 임차와 겹쳤다; 다음 사이클이 새로 읽는다)
  e := e1
  publish · 정산 트랜잭션            ← 잠금 없음 (원격 · 원장)
  release (실패 기록 경로만)
  D1 판정 (메모리)
  잠금이면 latchConfirmed(row, e, escalate)            ← escalate = D1 표의 「잠금 + 승격」/「잠금만」 (W4)
     for 시도 := 1..3 {
        if Gate.BlockUnlessClearedSince(R, e, 고정 detail) { escalate 면 승격 · 로그 (잠금 없음); return }
        // 해제 요청이 그 사이에 있었다. 버리기 전에 원장에 묻는다 (V2)
        e = Gate.ClearEpoch(R)                               ← 먼저 새 세대를 읽고
        a, err := Journal.LookupAlert(row.ID)                 ← 그 뒤 행을 읽는다 (잠금 없음)
        if err == nil && a.State == ACKNOWLEDGED { 판정 버림; return }   // 사람이 그 행을 봤다 — 승인이 이긴다
        // PENDING(같은 id 의 새 에피소드 포함) · DELIVERED(전달 복구 ≠ 사람 확인, W1) · 읽기 오류 → 새 세대로 다시
     }
     deferred[row.ID] = {escalate}                            ← 세 번 연속 해제와 경합: 버리지도 무조건 잠그지도 않고 다음 사이클로 미룬다 (W3)
```

**미룬 판정.** `deferred` 는 실행자 메모리의 작은 맵이다. 매 사이클 시작(나열 전)에 각 항목을 새 세대로 `latchConfirmed` 에 다시 넣는다 — 원장이
ACKNOWLEDGED 라고 답하면 버리고, 아니면 잠근다. 해제는 사람의 행위라 사이클마다 세 번 경합이 이어지는 일은 사람이 그렇게 하는 동안뿐이고, 그 동안
판정은 사라지지 않고 기다린다. 재시작하면 사라진다 — 그 행이 PENDING 이면 기동 복원이 잠그고, 이미 DELIVERED 면 잠기지 않는다. 후자는 오늘의 메모리 래치가
재시작에 지워지는 것과 같은 성질이다(기동 복원은 PENDING 만 센다, `gateway.go:153-167`) — 새 구멍이 아니라 기존 한계이고, 기록한다.

임차 오류(`deliverOne` B1)는 `e1` 만 읽고 D8 계수에 쓴다. 나열 오류는 사이클 시작에 읽은 세대를 쓴다.

### 판별 케이스별 논증

핵심 사실 **(P)** — 무엇이 보장되고 무엇이 안 되는가.

- (P1) `e1 == e2` 이면 임차를 얻는 동안 이 사유의 해제 요청이 없었다. 그러므로 판정 근거인 에피소드는 임차 순간 PENDING 이었고, 세대 `e` 는 그 에피소드의
  근거보다 **앞선** 값이다. 임차 후 세대만 읽으면(5판 1쇄) 「임차 → 승인·해제 → 새 세대 읽기 → 전달 정산 오류 → 잠금」에서 승인된 에피소드가 새 세대를
  물려받아 다시 잠근다(V1). 임차 전 세대만 읽으면 「읽기 → 승인·해제 → 같은 id 재무장 → 임차」에서 새 에피소드의 판정이 옛 세대로 버려진다. 두 번 읽고
  다르면 보내지 않는 것이 둘을 함께 막는다.
- (P2) 울타리가 실패하면 **해제 요청이 있었다**는 것만 확실하다. `Acknowledge` 는 미전달 수를 센 **뒤에** `Clear` 한다(FLM acknowledge B9 → B10,
  :870-875). 그 사이에 `n.mu` 를 거치지 않는 기록자 — `execgw/replay.go:551` 의 `EnqueueAlert` — 가 새 행을 넣을 수 있다(V2). 그래서 「해제 순간 미전달
  0 이었다」는 추론은 **틀렸다**(5판 1쇄의 (P)). 판정을 버리는 근거는 세대가 아니라 **원장**이다: 새 세대를 읽은 뒤 행을 다시 읽어 **ACKNOWLEDGED**
  이면 버리고, 그 밖(PENDING · DELIVERED · 읽기 오류)이면 새 세대로 다시 잠근다. DELIVERED 로 버리면 안 된다(W1): 셈~해제 사이에 들어온 행 B 가 한도에
  닿고, 옛 해제가 세대를 올리고, 다른 발송자가 B 를 전달하면 B 는 **사람이 본 적 없이** 판정을 잃는다. 정본은 「전달 복구 후 **수동 확인**으로 해제」다
  (`openspec/specs/engine-safety/spec.md:183`) — 전달은 사람 확인이 아니다.
- (P3) 원장 확인 뒤 · 재시도 울타리 전에 또 해제가 끼면 한 번 더 돈다. 세 번 연속이면 **다음 사이클로 미룬다**(무조건 잠그지 않는다 — 그 사이 마지막 해제가
  그 행을 실제로 승인했다면 무조건 잠금이 승인을 뒤집는다, W3).

**`Acknowledge` 의 셈~해제 구간 자체의 경합(V2 의 뿌리)은 이 change 가 고치지 않는다** — 그 구간에서 독립 기록자가 넣은 행이 있어도 게이트가 열리는 것은
HEAD 의 `Acknowledge` 성질이고, a092 델타가 「무엇이 그 구간을 지키는지 적어야 한다」(`a092…/spec.md:72`)로 소유한다. a124 는 그 행의 판정을 잃지 않는
것까지만 진다(위 P2) — 그 행이 한도에 닿으면 다시 잠긴다. proposal Follow-ups 에 적었다.

| 케이스 | 경과 | 결과 |
|---|---|---|
| ㉠ **진행 중인 승인과 겹침**(Q2) | 승인이 `n.mu` 를 잡고 행을 승인하는 중에 실행자가 세대를 읽고 정산·판정 | 세대는 승인 **시작**이 아니라 **해제 순간**(`Clear`)에 오른다. 울타리가 해제 전이면 `Block` 이 먼저 서고 해제가 지운다 — 승인이 이긴다. 해제 뒤면 울타리 실패 → 원장 확인: 행 ACKNOWLEDGED → 버림 |
| ㉡ 해제 뒤 늦은 실패 · 기록 오류(R4 포함) | 정산이 `Applied`·오류로 끝났지만 그 사이 해제 요청 | 울타리 실패 → 원장 확인. 승인됐으면 버림, 아직 PENDING 이면 새 세대로 잠금 — 오류가 가린 정보를 원장이 답한다 |
| ㉢ **무관한 승인 · 실패한 승인**(Q3) | 다른 행만 승인(미전달 남음 → B10 거짓) 또는 승인이 원장 오류로 중단(B5 · B8 · B9) | `Clear` 가 안 불린다 → 세대 불변 → 판정은 **그대로 선다**. 4판처럼 무관한 승인이 판정을 버리는 일이 없다 |
| ㉣ **래치가 서기 전의 전체 승인** | 행이 방금 한도에 닿았고(정산 `Applied`) 게이트는 아직 안 잠김. 울타리 전에 운영자가 전부 승인 → 미전달 0 → `Clear` — 지울 래치가 없다 | 세대를 **해제 요청마다** 올리므로 울타리 실패 → 원장 확인: ACKNOWLEDGED → 버림, 승격도 없음. **지시대로 「실제로 지웠을 때만」 올리면** 여기서 세대가 그대로라 빈 backlog 위에 래치가 서고 durable 승격까지 남는다 — 5판이 지시와 다르게 적은 이유 |
| ㉤ **전송 성공 뒤 잠금**(Q3 · Q4) | 판정 뒤 · 울타리 전에 다른 발송자가 그 행을 전달(그리고 같은 id 재무장까지) | 해제 요청이 없으므로 세대 불변 → **잠근다**. 이것은 틀린 잠금이 아니다: 한도에 닿은 순간 잠겼어야 했고, 정본은 「전달 복구 후 **수동 확인으로** 해제」다 — 제때 잠갔다면 전달 뒤에도 잠겨 있었을 상태와 같다. 재무장된 새 에피소드는 자기 시도로 따로 판정된다 |
| ㉥ 승인이 정산 **전**에 끝남 | 행은 ACKNOWLEDGED(`outbox.go:494-499`) | 정산 `AlreadySettled` → 판정 없음(D1) |
| ㉦ 동기 경로의 같은 id 재무장이 정산 **전** | 재무장은 임차 열을 지우고 새 토큰(`outbox.go:334-342`) | 옛 토큰 정산 `LeaseLost` → 판정 없음 |
| ㉨ **임차와 해제가 겹침**(V1) | 임차 → 승인·해제 → (5판 1쇄라면 새 세대 읽기) → 발행 성공 → 전달 정산 오류 | `e1 != e2` → 보내지 않고 반납. 다음 사이클 나열에 그 행은 없다(승인됨) |
| ㉩ **셈~해제 사이의 독립 기록**(V2) | 승인이 0 을 셈 → `EnqueueAlert` 로 B 생성 → 실행자가 B 임차(`e1 == e2`, 옛 세대) → 발행 성공, 정산 오류 → 승인의 `Clear` → 울타리 실패 | 원장 확인: B 는 PENDING → 새 세대로 **잠금 + 승격**. 5판 1쇄는 여기서 B 의 판정을 버렸다 |
| ㉪ **㉩ + 그 뒤 다른 발송자가 B 를 전달**(W1) | ㉩ 의 울타리 실패 뒤 · 원장 확인 전에 동기 경로가 B 를 전달 | 원장 확인: B 는 DELIVERED — 사람 확인이 아니므로 **새 세대로 잠금 + 승격**. 5판 2쇄 초고는 「PENDING 아님」으로 버렸다 |
| ㉫ **해제 경합 세 번 연속**(W3) | 울타리 → 해제 → 원장(ACK 아님) → 울타리 → 해제 … 세 번 | 다음 사이클로 미룸. 그 사이 마지막 해제가 B 를 승인했으면 다음 사이클 원장 확인이 ACKNOWLEDGED → 버림 |
| ㉧ 한 행만 승인, 다른 행 남음, 이 행의 전달 정산이 오류 | 해제 없음 → 세대 불변 → 잠금 | 이 행은 승인됐지만 다른 미전달이 있다 — 게이트가 잠긴 것은 정본 해제 조건(미전달 0)과 어긋나지 않는다. 보수 방향 잔여(기록) |

**이 기계는 잠그는 쪽으로만 틀린다 (재시작 제외).** 판정을 **버리는** 길은 하나뿐이다 — 울타리 실패 **그리고** 새 세대를 읽은 뒤 원장이 그 행을
**ACKNOWLEDGED** 라고 답할 때. 그때 사람이 그 행을 봤다. 나머지 갈래(세대 같음 · 원장 PENDING/DELIVERED · 원장 읽기 오류)는 **잠그고**, 세 번 연속 경합은
**미뤄서 다시 묻는다**. 미룬 판정이 재시작에 사라지는 것은 기존 메모리 래치와 같은 한계다(위 「미룬 판정」).
- 낡은 근거로 잠금이 **더** 서는 경우(㉤ · ㉧, 같은 id 새 에피소드) → 해제는 운영자의 빈 목록 승인(FLM acknowledge B4 → B10)이 한다.
- 세대 구현이 망가져 **안 오르면** `BlockUnlessClearedSince` 는 `Block` 과 같아진다 — 4판 이전의 「승인 뒤 늦은 래치」로 돌아갈 뿐 잠금을 잃지 않는다.
- 세대가 **엉뚱하게 오르면**(다른 사유·다른 연산이 올림) 원장 확인이 막는다 — 잘못 오른 세대는 판정을 버리게 하지 못하고 한 번 더 도는 비용만 든다.
  그래도 계약(해제 요청 없이 불변)은 spec 과 시험 2.13 이 못 박는다.

**손절 경로가 새로 기다리는 것: 경계 있는 게이트 구간 하나(Q1 · V4).** 실행자는 `n.mu` 를 잡지 않는다. 실행자가 쥐는 `g.mu` 구간은 map 읽기·삽입뿐이고
그 안에서 밖을 부르지 않는다. 손절 경로 중 `g.mu` 를 잡는 것(동기 `deliver` 의 `Gate.Block`, 비상 청산의 `blockEntries` — `flatten.go:635-643`)이
실행자 때문에 기다리는 시간은 그 map 연산 하나로 **경계가 있다**. 「아무것도 기다리지 않는다」가 아니다 — 원격·원장·다른 잠금에 대한 대기가 전파되지 않는다. 반대로
실행자가 전략 dispatch 의 `g.mu` 보유(브로커 전송)를 기다리는 것은 **실행자 판정의 지연**(D6 `M`)이지 손절 경로의 지연이 아니다. 원장 연결 하나를
exit 루프와 나눠 쓰는 것은 오늘 실행자와 같다(F12) — tasks 2.6 이 잰다.

**승격은 `BlockUnlessClearedSince` 가 true 일 때만, 잠금 밖에서.** 승인은 모드를 풀지 않는다(완화는 사람 승인 — FLM `transitionoperatingmode` B17~B19).
사람이 모드를 울타리~승격 사이에 완화하면 모드가 다시 막힌다 — 보수 방향, 다시 완화하면 된다. 운영 순서는 **승인 먼저, 완화 뒤**(Risks).

**잠금 순서.** 실행자는 `g.mu` 하나만, 원장 연결은 잠금 밖에서만 잡는다 — 순환이 생길 자리가 없다. (3판의 「journal 에는 콜백 메서드가 없다」는
틀렸다 — `TransitionOperatingMode` 는 트랜잭션 안에서 `Auditor.RecordAction` 을 부른다(FLM B23). 5판의 논증은 그것에 기대지 않는다.)

## D8 — 원장에 시도를 남기지 못하면 (M2 = (ii), F3 · N2 · N4 · R3 · R4 · R8)

`MarkAlertAttemptFailed` 가 오류를 돌려주면(FLM `deliverOne` B10; 원천은 `settleUnderClaim` B2·B3·B4·B6·B7·B8) attempts 가 오르지 않아
D1 판정이 영원히 서지 않는다. a092 뒤에는 이것을 가려 줄 동기 경로가 없다.

**계수는 행별이고 수명은 해제 세대에 묶는다.** 3판의 실행자 단위 계수는 다른 행의 성공이 지웠다(R3). 4판은 승인 세대로 묶어 무관한 승인이 계수를
지웠다(Q3). 5판은 D7 의 해제 세대를 쓴다 — 세대가 바뀌면 운영자가 backlog 를 비웠다는 **신호**이므로 계수를 0 으로 한다. 이 신호는 (P2) 에 따라 틀릴 수 있다 —
그래도 결과는 **다시 세는 지연**뿐이다(판정을 버리는 것은 D7 의 원장 확인만 한다).

| 사건 | 그 행의 계수 |
|---|---|
| 그 행의 실패 기록 오류(`deliverOne` B10) · 임차 요청 오류(B1) | +1. 항목에 **그 `deliverOne` 이 읽은 해제 세대 `e`**(임차 직후; 임차 오류면 그 자리)를 적는다 — 항목의 세대가 `e` 와 다르면 먼저 0 으로(끝난 에피소드) |
| 그 행의 정산이 **`Applied`** 로 끝남(실패 기록이든 전달이든) | 지움 — 원장이 그 행에 썼다. `LeaseLost`·`AlreadySettled` 는 0 행을 썼으므로 지우지 **않는다**(R3) |
| 그 행이 임차에서 「이미 정산됨」(B6) | 지움 — PENDING 을 떠났다 |
| 사이클 시작에 본 해제 세대가 항목의 세대와 다름 | 그 항목의 수를 0 으로 — 운영자가 backlog 를 비웠다는 신호다. (P2) 로 이것이 틀릴 수 있으므로(셈~해제 사이의 독립 기록) 결과는 **지연**뿐이다: 기록 오류가 계속되면 다시 센다. **무관한·실패한 승인은 세대를 안 바꾸므로 계수를 지우지 않는다**(Q3) |
| 나열이 **완전할 때**(`len(pending) < batch` — 배치로 잘리지 않음) 나열에 없는 id | 지움 — PENDING 전체를 봤고 그 행은 없다. 잘린 나열에서는 거르지 않는다(밀려난 PENDING 행의 계수를 지우면 영영 한도에 못 닿는다) |
| 오류가 났을 때 **실행자 수명 ctx**(Run 의 ctx)가 끝나 있음 | 세지 않음 — 엔진 종료. transport timeout 은 ntfy 의 자식 ctx 라 발행 실패(`perr`)로 온다 |
| 재시작 | 사라짐 — 대신 기동 복원이 PENDING 수로 다시 잠근다(`gateway.go:153-167`) |

**맵의 크기 (Q5).** 남이 전달해 PENDING 을 떠난 행의 항목은 다음 완전 나열이나 다음 해제 요청에서 지워진다. 나열이 계속 잘리고 해제도 없으면 그 사이
기록 오류를 낸 서로 다른 행 수만큼 남는다 — 항목은 id · 수 · 세대 셋이고 재시작이 비운다. 막지 않고 기록한다(잘린 나열에서 지우는 것이 더 나쁘다 —
위 표).

미전달 **나열** 오류(`cycle` B1)는 행이 없으므로 실행자 단위 계수 하나로 센다 — 증가 **전에** 사이클 시작의 해제 세대를 계수의 세대와 비교해 다르면 먼저 0 으로
(V5), 그 세대를 적고 +1. 나열 성공이 지운다.

**같은 id 의 에피소드가 바뀌어도(해제 없이 전달 → 재무장) 계수는 id 에 남는다(V5).** 새 에피소드가 옛 기록 오류를 물려받아 **더 일찍** 잠길 수 있다 —
보수 방향이라 받아들이고 시험 2.10 에 기대 결과(3 회째 기록 오류에서 잠금)로 적는다.

**한도에 닿으면.** 행 계수 또는 나열 계수가 `alertAttemptLimit` 에 닿으면 `BlockUnlessClearedSince(ReasonAlertUndelivered, 항목의 세대, 고정 detail)` 로
잠그고(D7 의 `latchConfirmed(row, e, 승격)` — 울타리 실패면 원장 확인 후 재시도, 세 번이면 미룸), 잠갔으면 잠금 밖에서 승격을 시도한다(원장이 쓰기를
거부하는 중이면 실패할 것이다 — D9). 원장 확인이 ACKNOWLEDGED 면 계수를 버린다.

**나열 계수의 울타리 (W2).** 나열 오류에는 행이 없어 원장 확인을 할 행이 없다. 나열 계수가 한도에 닿으면 `BlockUnlessClearedSince(R, 계수의 세대, …)` 로
잠그고 승격한다. 울타리가 실패하면 계수를 0 으로 되돌리고 **다시 센다**(판정을 버리는 것이 아니라 증거를 버린다): 해제 요청은 `Acknowledge` B10 에서만 오고,
B10 에 닿으려면 그 호출이 `PendingAlerts`(B4, 비었을 때)·`AcknowledgeAlert`·`UndeliveredCount`(B9)를 원장에서 **성공적으로** 했어야 한다 — 「원장을
읽지 못한다」는 연속의 전제가 그 순간 반증됐다. 나열이 계속 실패하면 3 사이클 뒤 다시 잠근다. 결정적 시험: 해제 경합 · 나열 회복 · 새 행 · 계속 실패(2.10).

행별 계수는 행마다 사이클당 한 번 오르므로 **최소 ≈ 2 사이클(≈ 4 s)** 이다(R8). 나열 계수도 사이클당 한 번이다.

**fail-closed 가 거부할 정상 입력 (먼저 열거 — Manager 지시).**

| 정상 입력 | 결과 | 이 change 의 목적 안인가 |
|---|---|---|
| **일시적 원장 쓰기 오류가 같은 행에서 3 연속** (busy timeout 초과 · 디스크 가득 참 등; 사이클마다 한 번이라 최소 ≈ 4 s), 그 사이 backlog 를 비우는 승인 없음 | **운영자 승인 전까지 신규 진입 차단**, 승격 시도 | **예.** 미전달을 원장에 적지 못하는 것은 「미전달 감시가 실패했다」이고, 감시가 실패한 동안 진입하지 않는 것이 이 change 의 목적(정본 「전달 실패가 지속되면 신규 진입을 차단」)이다 |
| 미전달 나열이 3 사이클 연속 실패 | 같음 | 예 — 같은 이유 |
| 발행 성공 뒤 전달 정산 쓰기 오류 **한 번** | **즉시** 차단 + 승격 | 예 — 동기 경로 parity(D1 전달 정산 표) |
| 짧은 transport 흔들림(ntfy 재시작 ≈ 4 s 동안 즉시 실패 3 회) | 차단 + durable 승격 → 사람이 완화 | 예 — 동기 경로가 오늘 같은 입력에 같은 결과(D2 표) |
| publisher 미설정 엔진에 critical 행 하나 | ≈ 4~6 s 에 차단 + 승격 | 예 — `newNotifier` 문서의 「specified direction」(D3) |
| 발송 중 운영자 승인 | **차단 안 함** | — 승인이 이긴다(D7) |
| 엔진 종료 중 기록 실패 | **안 셈** | — 취소는 원장 결함이 아니다 |

알려진 한계(기록): 원장이 간헐적으로 실패해 같은 행의 `Applied` 가 3 번 안에 한 번씩 끼면 계수는 한도에 닿지 않는다 — 그 경우 attempts 는
그 `Applied` 들로 오르므로 D1 판정이 결국 선다.

## D9 — 게이트 detail · 로그 · 승격 실패 (불변식 8, F8)

- 게이트 detail 은 **고정 문구 + 수**만: 예 `"a critical alert reached the delivery attempt limit (3) without being delivered; an operator
  has to acknowledge the backlog"`. 행의 제목·본문·payload·`last_error`·계좌·토큰·원문 오류를 넣지 않는다 — 선례 `auxiliary.go:206-211`
  (고정 문구) · `gateway.go:161-166`(수만). 동기 경로 detail 형식(`notifier.go:563-571`, 원문 오류 포함)은 베끼지 않는다.
- `MarkAlertAttemptFailed` 의 원인 문자열: 전송 실패는 오늘처럼 `perr.Error()`(원장 안에 남고 게이트 detail 로 나가지 않는다), publisher
  부재는 고정 문구.
- 로그: **이 change 가 새로 쓰는 줄**(래치 · 승격 · 기록 실패 한도)은 허용 목록 필드만 쓴다 — 사건 종류 · 고정 detail · `alert_id` · 사유 코드 ·
  수. **계좌 참조와 원문 오류 문자열은 넣지 않는다**(불변식 8, R2): `Logger.Error` 는 `err.Error()` 를 그대로 붙이고(`obs/log.go:162-165`) 모드
  전이 오류는 계좌를 담는다(FLM `transitionoperatingmode` B11 · B22) — 그래서 승격 오류는 `err` 를 넘기지 않고 고정 분류(`context` 취소 / 원장 오류)만
  적는다. 3판의 「기존 관례라 계좌 허용」은 철회한다 — 기존 줄의 관례는 새 누출의 근거가 아니다. 기존 줄(`deliverOne` 의 `logf` 호출들)은 이 change 가
  바꾸지 않는다. 행 제목·본문·payload·`last_error`·임차 토큰은 어느 줄에도 없다(`logf` 계약 :312-313). 시험은 계좌·토큰 sentinel 로 새 줄을 잰다(2.11).
  로그는 전부 배제 **밖**(D7).
- 승격 실패: error 로그, 메모리 래치는 남고 재시작은 기동 복원이 덮는다 — `Notifier.escalate` 의 「오류 반환 없음」 계약과 같다.
- 래치·승격 로그는 **바뀐 때만** 한 줄(`changed == true`, 또는 이 실행자가 처음 잠근 때) — 매 사이클 재확인은 조용하다(R21 선례).

## 안전 불변식 대조

- **손절 즉시성**: exit 관측 루프를 만지지 않는다. 진입 게이트는 노출을 늘리는 변이에만 묻는다(`execgw/gateway.go:855-859`), 트리거는
  `ENTRY_BLOCKED` 로만 간다(`operating_mode.go:537-545`). 실행자는 `n.mu` 를 잡지 않고 `g.mu` 는 map 연산 동안만 쥔다(D7) — 손절 경로가 실행자 때문에 새로 기다리는 것은 그 게이트 구간 하나로 경계가 있고,
원격·원장·다른 잠금 대기는 전파되지 않는다. 원장 연결 공유는 tasks 2.6 에서 잰다 —
  `a098_the_backlog_does_not_delay_protection_test.go` 가 그대로 초록이어야 하고 publisher 없음 · `g.mu` 경합 변형을 더한다.
- **토글 OFF = upstream**: 새 토글 없음. 1판의 *"알림 게이트 OFF 면 Notifier 자체가 없다"* 는 **틀렸다**(F7) — Notifier 는 엔진에서
  무조건 생성된다(`gateway.go:323`). 실제 OFF 경계는 automation gate OFF 에서 `engine run` 이 기동을 거부하는 것이다
  (`cmd/tossctl/engine.go:220-221` `errEngineGateOff`, `TestAGateOffEngineRefusesWithoutEnumeratingClauses` rc 0, 2026-09-26). 엔진이 없으면 실행자도 없다.
- **보수 방향**: 이 change 가 여는 문은 없다 — 오늘 안 잠기던 자리가 잠긴다. 한도는 동기 경로와 같고(D2), `Acknowledge` 는 편집하지 않으며 해제
  조건(승인 + 미전달 0)은 그대로다. 새 execgw 기계는 잠그는 쪽으로만 틀린다(D7). `revision`(전략 ABA 봉인)의 의미는 바꾸지 않는다.
- **audit**: 모드 승격은 `operating_modes` 행으로 남는다(자동·보수 — 정본 「보수 방향은 자동·즉시·durable」).

## Risks

- **D2 ㄱ 은 durable 승격을 더 자주 만든다** — `Notify` 없이 기록되는 행(`execgw/replay.go:551` `parkAlert`)과 a092 뒤의 모든 행. 완화는
  사람 승인. 운영 순서는 **승인 먼저, 모드 완화 뒤**(D7) — 거꾸로 하면 다음 실패가 다시 승격한다.
- **배포 직후(F6 · N8)**: 기존 PENDING 행은 기동 복원이 게이트를 이미 잠근다. `attempts ≥ 3` 인 행은 한도 층으로 내려가고, **그 행이 다음에
  선택되어 실패하면** 승격을 시도한다 — 모드가 실제로 바뀔 때만 `operating_modes` 행이 생긴다(같거나 더 엄하면 없음, FLM `transitionoperatingmode`
  B15·B16). 새 행에 밀리거나 다음 발행이 성공하면 승격은 없다. **운영 원장의 해당 행 수는 이 change 가 조회하지 않았다 — 사람 몫**(배포 전 확인).
- **원장 쓰기 증가(F12)**: D3 로 publisher 없는 엔진은 사이클마다 행당 기록 1 회(10 행이면 2 s 마다 10 회, 정산 1 회 5.584 ms 실측 기준
  ≈ 56 ms)를 더 쓴다. 같은 연결을 exit 루프가 쓴다 → tasks 2.6.
- **D8 의 오탐 비용**: 원장 쓰기 3 연속 실패가 승인 요구로 이어진다(D8 표 첫 줄) — 목적 안이라고 판단했다.
- **execgw 표면 추가(M1 = B′)**: High-risk 표면에 필드 1 · 메서드 2 · `Clear` 한 줄. 해제 세대의 두 계약(해제 요청 없이 불변 · 이 사유의 해제는
  미전달 0 에서만)이 (P) 의 전제다 — 시험 2.13 이 둘을 못 박는다. 1.2 Pre-Edit 대상에 `EntryGate` 를 넣는다.
- **기존 결함 (범위 밖, proposal Follow-ups)**: 동기 경로가 `n.mu` 를 쥔 채 `Gate.Block` 을 부른다(`notifier.go:280·484·520·571`) — 전략 진입이
  켜져 `g.mu` 를 전송 동안 쥐면 손절 경로가 브로커 전송을 기다린다. 전략 진입 활성화 전 필수 선행.
- a092 가 먼저 착지하면 정본 SHALL 이 빈다. 순서를 a092 tasks 의 착수 조건으로 못 박는다.
- a092 델타에도 「운영자의 승인은 … 되살리지 않는다」 문단(`a092…/spec.md:66`)이 있다. 두 change 가 같은 요구를 각자 싣고 아카이브되면
  정본에 사본 둘이 생긴다 → a092 21판이 그 문단을 이 change 의 요구로 가리키게 한다(tasks 5.2 에 더함).

## freeze 발견 처분 (`review.md` §0 · §0.2 · §0.3 · §0.4)

| id | 처분 | 자리 |
|---|---|---|
| F1 P0 | 수용 — 커밋된 증가 후 attempts, `SettleResult.Attempts` additive | D1 · tasks 2.7 · 3.4 |
| F2 P0 | 수용 — M1 = C(2~4판) → **B′**(5판): 해제 세대 울타리, 실행자는 `n.mu` 무관, 승격은 잠금 밖 | D7 · spec 「승인이 이긴다」 · tasks 2.8 · 2.9 · 3.5 |
| F3 P0/P1 | 수용 — M2 = (ii), 정상 입력 거부 열거 | D8 · spec 「기록 실패도 지속 실패」 · tasks 2.10 · 3.6 |
| F4 P1 | 수용 — 식 재유도, 214 s / 316 s | D6 · tasks 4.4 |
| F5 P1/P2 | 부분 수용 — 네 경우와 유계·비유계 기록, 정렬 확장은 YAGNI | D4 |
| F6 P1/P2 | 수용(기록) — 배포 직후 동작, 운영 원장 조회는 사람 몫 | Risks |
| F7 P1/P2 | 수용 — 불변식 대조 문장 교정 | 안전 불변식 대조 |
| F8 P1 | 수용 — 고정 detail · 로그 필드 · 승격 실패 처리 | D9 · tasks 2.11 |
| F9 P3 | 수용 — `Block` 은 없을 때만 삽입 | D1 |
| F10 P2 | 수용 — `err` 먼저 | D1 · tasks 2.7 |
| F11 P2 | 수용 — proposal 문장 교정 | proposal 「결정」 · D2 |
| F12 P2 | 수용 — publisher 없음 변형의 exit 체류 측정 | Risks · tasks 2.6 |
| F13 P2 | 기록 — 지연 차이 수용 | D3 |
| N1 P1 | 수용 — 3판: 정산 트랜잭션 하나로 좁힘 → 4판(R1): 메모리 전용 | D7 · tasks 2.6 |
| N2 P1 | 수용 — 3판: 원장 읽기 울타리 → 4판: 승인 세대 → 5판: 해제 세대 | D7 판별 케이스 · D8 · tasks 2.10 |
| N3 P1 | 수용 — 전달 정산 실패는 동기 경로처럼 즉시 잠금 + 승격, 임차 유지 | D1 전달 정산 표 · spec · tasks 2.12 |
| N4 P1 | 수용 — 3판: 실행자 단위 → 5판: 행별 + 해제 세대 수명 | D8 · tasks 2.10 |
| N5 P1 | 수용 — 일반 상한 주장 철회, 전제 H 조건부 식 + 무계 경우 명명 | D6 |
| N6 P2 | 수용 — 실패 기록의 `NotFound` 는 잠금만(동기 parity) | D1 |
| N7 P1 | 수용 — spec 에 나열 오류·전달 정산 실패·로그 정제, tasks 에 인터리빙·수명·재진입 금지 | spec · tasks 2.9~2.12 |
| N8 P2 | 수용 — 배포 문장을 조건부로 | Risks |
| R1 P0 | 수용 — 4판: 메모리 전용 배제 → 5판: 배제 없음(`n.mu` 무관, B′) | D7 · spec · tasks 2.6 · 3.5 |
| R2 P1 | 수용 — 새 로그 줄은 허용 목록 필드, 계좌·원문 오류 금지, sentinel 시험 | D9 · spec · tasks 2.11 |
| R3 P1 | 수용 — 행별 계수, `Applied` 만 지움 | D8 · tasks 2.10 |
| R4 P1 | 수용 — 전역 수 울타리 폐기, 해제 세대 울타리(㉡) | D7 · D8 · tasks 2.9 · 2.10 |
| R5 P1 | 수용 — 배제 계약 하나(메모리 전용)로 spec·tasks·처분표 정렬 | spec · tasks 3.5 · 4.1 |
| R6 P1 | 수용 — 보장을 「임차가 살아 있고 교체되지 않은 동안」으로 | D1 · spec · tasks 2.12 |
| R7 P2 | 수용 — 54 s 주장 철회, 만료 ≠ `LeaseLost` | D6 |
| R8 P2 | 수용 — 행별 계수로 최소 ≈ 4 s 복원 | D8 |
| Q1 P0 | 수용 — M1 = B′: 실행자가 `n.mu` 를 잡지 않음 | D7 · tasks 2.6 · 3.5 |
| Q2 P1 | 수용 — 세대는 해제 **순간**(`Clear`)에 오른다(케이스 ㉠) | D7 · tasks 2.9 |
| Q3 P1 | 수용 — 무관·실패한 승인은 세대 불변 → 판정·계수 보존(㉢), 판정이 버려지는 것은 (P) 로 불필요한 것뿐 | D7 · D8 · tasks 2.9 · 2.10 |
| Q4 P2 | 수용 — 전달·재무장 뒤의 늦은 잠금은 「제때 잠갔을 상태」와 같음을 논증(㉤) | D7 · tasks 2.8 |
| Q5 P2 | 수용 — 해제 세대 · 완전 나열로 거름, 남는 크기 기록 | D8 · tasks 2.10 |
| Q6 P2 | 수용 — 전제 H 여섯 조건, `E`·`I_list` 항, a092 식은 재정의가 아니라 교체 | D6 · tasks 5.2 |
| Q7 P3 | 수용 — F2 행 교정 | 이 표 |
| V1 P1 | 수용 — 임차 전후 세대 두 번, 다르면 보내지 않음(㉨) | D7 · spec · tasks 2.9 |
| V2 P1 | 수용 — 울타리 실패 시 원장 확인 후 재시도, 세 번이면 무조건 잠금(㉩). `Acknowledge` 셈~해제 경합 자체는 a092 소유(Follow-ups) | D7 · D8 · spec · tasks 2.9 |
| V3 P1 | 수용 — proposal 세대 읽기 시점 문장 정렬 | proposal |
| V4 P1 | 수용 — 「경계 있는 게이트 구간 허용, 그 안 외부 작업 금지」로 spec 교정 | D7 · spec |
| V5 P2 | 수용 — 에피소드 carryover 는 보수 방향으로 수용·시험 기대값, 나열 계수 초기화 순서 명시 | D8 · tasks 2.10 |
| V6 P2 | 수용 — `I_list` 항, 「조건부 상한」 표현, tasks 5.2 문장 교체 | D6 · tasks 5.2 |
| W1 P0 | 수용 — 버리는 근거를 ACKNOWLEDGED 로만, DELIVERED 는 잠금(㉪) | D7 · spec · tasks 2.9 |
| W2 P1 | 수용 — 나열 계수의 울타리 실패는 증거 초기화(재계수), 근거는 `Acknowledge` B10 의 원장 읽기 성공 | D8 · tasks 2.10 |
| W3 P1 | 수용 — 무조건 잠금 폐기, 다음 사이클로 미룸(㉫), 재시작 소실은 기존 메모리 래치 한계와 같음을 기록 | D7 · spec · tasks 2.9 |
| W4 P2 | 수용 — `latchConfirmed` 가 「잠금만」 판정을 운반 | D7 · tasks 2.8 |
| W5 P3 | 수용 — D8 · spec 해제 세대 요구 · tasks 3.5 · 불변식 문장 정렬 | 각 자리 |
