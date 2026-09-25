# a124 design — 판정은 원장이 커밋한 시도 수에 서고, 승인이 이긴다

> **4판 (2026-09-26, Teammate)** — freeze 1판(§0) · 2판(§0.2) · 3판(§0.3) REJECT 반영. Manager 결정 M1 = **C**, M2 = **(ii)**.
> 3판 → 4판: D7 의 배제를 **메모리 전용**으로 좁혔다 — 원장 연산은 전부 배제 밖, 배제 안에는 「승인 세대 비교 + 게이트 map 삽입」만(codex R1 P0).
> 그 대가로 `Acknowledge` 에 승인 세대 증가 **한 줄**이 들어간다(해제 조건 불변). D8 계수를 행별로 되돌리고 수명을 승인 세대에 묶었다(R3 · R4 · R8).
> D9 새 로그는 계좌·원문 오류 금지(R2). 임차 유지의 보장을 실제 크기로 줄였다(R6). D6 의 54 s 주장 철회(R7).
>
> **⚠ freeze 4판 REJECT (`review.md` §0.4)** — D7 의 수단(M1 = C, `n.mu` 배제)은 두 형태 모두 손절 경로 대기를 만든다. 대안 B′(게이트의
> 사유별 해제 세대)는 Manager 결정 M1 을 바꾸는 것이라 **Manager 재결정 전까지 이 판을 고치지 않는다.** D7 · D8 의 승인 세대 문장은 그 결정에 따라 다시 쓴다.
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
B3 과 같은 결과). 스키마 무변경, 세 호출자(`MarkAlertDelivered` · `MarkAlertAttemptFailed` · `ReleaseAlertClaim`)의 SQL·CAS 불변,
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
| `err != nil` | **즉시 잠금 + 승격**(D7 승인 세대 울타리를 통과할 때), 임차 유지 | B11·B12 → 즉시 `Block`. 동기 경로는 `n.mu` 를 발행부터 쥐어 그 사이 승인이 끼지 못하지만, 실행자는 발행을 배제 밖에서 하므로 승인이 발행과 정산 사이에 낄 수 있다 — 오류는 그것을 가리므로 울타리가 필요하다 |

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

`alertDeliverer` 에 `Gate *execgw.EntryGate` · `AccountRef string` · `Exclusion deliveryExclusion` 을 더하고 `Context.AlertDeliverer`
(`auxiliary.go:156-190`)에서 `c.Entry` · `c.AccountRef` · `c.Notifier` 를 넘긴다. `Journal` 은 이미 있다. 승격은
`Journal.EscalateOperatingMode(ctx, AccountRef, ModeTriggerCriticalAlertUndelivered, nil)` — `Notifier.escalate` 와 같은 호출(:382)이고
announcer 는 nil(실행자는 `Notify` 를 재진입할 이유가 없다; 전달 수단이 막 실패한 참이다 — `notifier.go` escalate 문서).
`Exclusion` 은 `AckGeneration` · `LatchUnlessAcknowledgedSince` 두 메서드의 인터페이스다(D7). `c.Notifier` 가 nil 이면 울타리 없이 바로 잠근다 —
그 배선에는 승인 경로도 없다(`alertops.go:98` 이 `c.Notifier == nil` 을 거부).

## D6 — 래치 시간: 조건부 식 (R3, 2판 재유도 — F4 · N5)

**일반 상한은 없다.** 아래는 전제를 이름으로 단 조건부 식이고, 전제가 깨지는 경우는 표 뒤에 따로 적는다. 2판이 「≤ … (상한)」과
781 s 를 운영 상한처럼 쓴 것은 철회한다(N5).

기호: `L` 한도(3) · `I` 사이클 대기(2 s) · `B` 한 사이클이 시도하는 행 수(≤ 10) · `T` 한 행의 발행 실패 시간(즉시 ≈ 0, timeout 10 s —
transport 가 소유, `ntfy.go:95-100`) · `S` 한 행의 원장 비용(임차 · 정산 · 반납) · `M` D7 울타리의 `n.mu` 대기.

실행자는 **사이클 먼저, 대기 뒤**(`alertdelivery.go:122-136`)이고 한 사이클은 선택한 행을 차례로 시도한다(`:154-162`). 한 행의 재시도
사이에는 **같은 사이클의 다른 행 전부**가 낀다 — 1판 식이 이것을 빠뜨렸다.

**전제 H (동질 두절)**: transport 가 끊겨 모든 발행이 같은 `T` 로 실패하고, 임차 경합이 없고, `S`·`M` 이 유한하다. 이때 두절 뒤 첫 사이클의
첫 행은 한도 아래 행 중 가장 오래된 것이고(R2 정렬), 한도에 닿을 때까지 그 자리를 지킨다(더 새 행은 id 로 뒤). 게이트는 **어느 행이든**
처음 한도에 닿을 때 잠긴다. 그래서

- 사이클 길이 `C = B·(T + S + M) + I`
- 두절 뒤 첫 사이클 시작까지 `Q ≤ C` (진행 중 사이클의 남은 부분 + 대기)
- **래치 ≈ `Q + (L−1)·C + (T + S + M)`**

| 예시 (전제 H, `S = M = 0`) | 첫 사이클 시작부터 | `Q` 최대 포함 |
|---|---:|---:|
| 행 1, 즉시 실패 (`C = 2`) | 4 s | 6 s |
| 행 1, 10 s timeout (`C = 12`) | 34 s | 46 s |
| **배치 10 행 전부 timeout** (`C = 102`) | **214 s** (끝 행 304 s) | **316 s** |

1판의 `100 + 3×2 + 3×10 = 136 s` 는 첫 행 기준 214 s 보다 작았다.

**전제 H 가 없을 때 (이질 실패 — 특정 행만 실패)**: 실패하는 행보다 오래된 한도 아래 행 `K` 개가 **성공하며** 빠지는 동안 기다린다 —
`Q ≤ C + ⌈K/B⌉ · C_ok` (`C_ok` = 성공 행들의 사이클 길이). `K` 는 작업량이 정하므로 이 change 가 값을 줄 수 없다.

**유한한 값이 없는 경우** (이름만 적는다 — a092 델타 「보장이 아니다」와 같은 결): 원장 정지(연결 하나 — `journal.go:174`; `S` 무계),
`n.mu` 를 쥔 다른 보유자의 시간(`M` — 동기 `deliver` 는 a092 착지 전 원격 전송 동안, `Flush` 는 backlog 전체 동안 쥔다: `notifier.go:254-255·734-741`;
이것은 판정만 늦춘다), 임차를 갱신하며 쥐는 발송자, `T` 에 상한이 없는 `Publisher` 구현.

임차 만료는 그 자체로 정산을 무효로 만들지 않는다 — 정산은 토큰·상태 CAS 만 본다(`outbox.go:455·472`). 만료 뒤 **다른 발송자가 잡아 토큰이
바뀌었을 때만** `LeaseLost` 가 된다(FLM settleUnderClaim 종단, `alert_claim.go:349-365`). 3판의 「`M ≤ 54 s` 라 81 s 안」은 철회한다(R7) — 4판의
`M` 은 D7 의 메모리 울타리 대기이고, 그것은 판정을 늦출 뿐 정산과 임차에 걸리지 않는다.

tasks 4.4 는 예시 세 줄을 실측해 review 에 적는다. a092 델타의 식 `다음 배달 사이클까지의 시간 + 시도 횟수 × 1회 상한 + (시도 횟수 − 1) ×
사이클 주기` 의 「사이클 주기」는 **배치 서비스 시간을 포함한 `C`** 라는 것을 a092 21판에 전한다(tasks 5.2).

## D7 — 승인이 이긴다: 승인 세대 울타리, 배제는 메모리만 (M1 = C, F2 · N1 · R1 · R4 · R5)

**문제.** 실행자는 뮤텍스를 잡지 않는다(`alertdelivery.go:19-20`). `Acknowledge` 는 「승인 → 미전달 수 → 0 이면 해제」 전체를 `n.mu`
아래에서 한다(FLM `internal-obs--notifier.acknowledge` B4~B10, :851-876). 실행자의 `Block` 이 그 해제 **뒤에** 떨어지면 빈 backlog 위에
래치가 서고 운영자가 방금 본 사건이 차단으로 되살아난다. 이미 정해진 규칙이 이것을 금한다: a092 델타 「운영자의 승인은 … 전송의 성공 여부는
그 사유를 되살리지 않는다(SHALL NOT)」(`a092…/spec.md:66-68`, 사용자 결정 2026-08-10 결정 2 = 안 1).

**3판이 틀린 곳 (R1 P0).** 3판은 정산 트랜잭션을 `n.mu` 안에 넣었다. `n.mu` 는 exit 관측 루프의 동기 `Notify`(`exitloop.go:1710` →
`claimAndDeliver` :254)와 비상 청산(`flatten.go:694`, `context.Background()`)이 잡는 뮤텍스이고 `sync.Mutex` 는 ctx 를 보지 않는다. 실행자가
그 안에서 원장 연결을 기다리면 손절 경로가 **취소할 수 없는** 대기를 얻는다. 「트랜잭션 하나」는 연산 수를 묶을 뿐 시간을 묶지 않는다.

**4판 수단 — 승인 세대.** `Notifier` 에 `ackGen atomic.Uint64` 를 두고 `Acknowledge` 가 `n.mu` 를 잡은 **직후** 한 번 올린다(해제 조건은 불변 —
FLM acknowledge B4~B10 무편집, 줄 하나 추가). `obs` 에 진입점 **둘**을 더한다:

- `AckGeneration() uint64` — 원자 읽기, 잠금 없음
- `LatchUnlessAcknowledgedSince(gen uint64, latch func()) bool` — `n.mu` 를 잡고 `ackGen == gen` 이면 `latch()` 를 부르고 true

`latch` 는 `Gate.Block(고정 detail)` 하나다 — **메모리 map 삽입뿐**(`retry.go:526-533`). 원장·로그·`Notifier` 메서드는 부르지 않는다.

```text
deliverOne(row)
  gen := Exclusion.AckGeneration()     ← 원자 읽기
  claim · publish · 정산 트랜잭션         ← 전부 배제 밖 (원장 대기는 여기서만)
  release (실패 기록 경로만)              ← 울타리 **전에** — n.mu 를 기다리는 동안 임차를 쥐고 있지 않게
  D1 판정 (메모리)
  잠금이면  ok := Exclusion.LatchUnlessAcknowledgedSince(gen, Gate.Block)   ← n.mu, 메모리만
           ok 이면 배제 밖에서 승격 · 로그
           아니면 판정을 버린다 (다음 실패가 다시 판정 — attempts 는 원장에 남았다)
```

**exit 루프가 새로 기다리는 것.** 실행자가 쥐는 `n.mu` 구간은 원자 비교 + map 삽입(`g.mu`)이다 — 원장·I/O 없음. exit `Notify` 의 추가 대기는
그 구간 하나다. 반대로 실행자는 `Acknowledge`·동기 `deliver`·`Flush` 가 쥔 `n.mu` 를 기다릴 수 있지만 그것은 **판정이 늦어지는** 것이지 손절 경로가
기다리는 것이 아니다(D6 `M`). tasks 2.6 은 (a) 실행자 판정 경로 전체(정산 포함)를 인위 지연해도 exit 체류가 늘지 않음, (b) 원장 연결을 막아도
exit `Notify` 가 실행자 때문에 `n.mu` 에서 기다리지 않음을 결정적으로 보인다. 늘면 멈추고 보고한다.

**인터리빙.** 판정의 효력(`Block`)과 `Acknowledge` 의 세고-푸는 구간이 같은 `n.mu` 로 직렬화되고, 세대가 「그 사이 승인이 있었는가」를 답한다.

| 순서 | 결과 |
|---|---|
| 승인이 정산 **전**에 끝남 | 행은 ACKNOWLEDGED(`outbox.go:494-499`) → 정산 `AlreadySettled` → 판정 없음. (세대도 바뀌었다) |
| 정산 뒤 · 울타리 전에 승인 | 세대가 바뀜 → `Block` 안 함 → 판정 버림. 행이 승인됐으면 다시 나열되지 않는다 |
| 울타리 뒤에 승인 | `Block` 이 먼저 → 승인이 세고 0 이면 해제 — 승인이 이긴다 |
| 정산이 원장 **오류**, 그 사이 승인 + 새 행 B 생성(R4) | 세대가 바뀜 → 오류 쪽 판정 버림. B 는 자기 시도로 판정된다 — 3판의 전역 `UndeliveredCount` 울타리가 틀렸던 자리 |
| 동기 경로가 같은 id 재무장 | 재무장은 임차 열을 지우고 새 토큰(`outbox.go:334-342`) → 옛 토큰 정산 `LeaseLost` → 판정 없음 |
| 다른 행만 승인(미전달 남음) | 세대가 바뀜 → 이번 판정 버림 — 게이트는 승인이 해제하지 않았으므로 이미 잠겨 있으면 그대로, 처음이면 **한 사이클 늦게** 잠긴다 |

마지막 줄이 이 수단의 대가다: 승인이 있었던 판정 창은 버린다. 승인은 사람의 행위라 매 사이클 반복되지 않으므로 지연은 유계(다음 실패)이고,
버린 판정의 근거(attempts)는 원장에 남는다.

**승격을 배제 밖에 두는 근거.** 승인은 모드를 풀지 않는다(완화는 사람 승인 — 정본; FLM `transitionoperatingmode` B17~B19). 승격은 `Block` 이
성공한 판정에서만 부르고, 그 판정은 울타리 시점에 승인보다 앞섰다. 사람이 **모드 완화**를 울타리~승격 사이에 넣으면 모드가 다시 막힌다 — 보수
방향이고 다시 완화하면 된다. 이 창은 spec 에 적었다(「승격은 배제 밖」).

**잠금 순서.** 실행자의 `n.mu` 구간 안에는 `g.mu`(map 삽입)만 있다. `g.mu` 는 잡은 채 밖을 부르지 않는다(`retry.go:526-543`; 권위 갱신은 `g.mu`
전 — :481-482). 원장 연결은 `n.mu` 밖에서만 잡으므로 「연결을 쥔 채 `n.mu`」와의 순환이 생길 수 없다. (3판의 「journal 에는 콜백 메서드가 없다」는
틀렸다 — `TransitionOperatingMode` 는 트랜잭션 안에서 `Auditor.RecordAction` 을 부른다(FLM B23 :460-465). 실행자의 승격은 `Auditor` 를 넘기지 않고,
4판에서는 그것이 교착 논증에 필요하지도 않다.)

## D8 — 원장에 시도를 남기지 못하면 (M2 = (ii), F3 · N2 · N4 · R3 · R4 · R8)

`MarkAlertAttemptFailed` 가 오류를 돌려주면(FLM `deliverOne` B10; 원천은 `settleUnderClaim` B2·B3·B4·B6·B7·B8) attempts 가 오르지 않아
D1 판정이 영원히 서지 않는다. a092 뒤에는 이것을 가려 줄 동기 경로가 없다.

**계수는 행별이고 수명은 승인 세대에 묶는다.** 3판의 실행자 단위 계수는 다른 행의 성공이 지웠다(R3). 행별 맵의 수명 문제(N4)는 세 규칙으로 닫는다.

| 사건 | 그 행의 계수 |
|---|---|
| 그 행의 실패 기록 오류(`deliverOne` B10) · 임차 요청 오류(B1) | +1. 첫 증가 때 **그때의 승인 세대**를 함께 적는다 |
| 그 행의 정산이 **`Applied`** 로 끝남(실패 기록이든 전달이든) | 지움 — 원장이 그 행에 썼다. `LeaseLost`·`AlreadySettled` 는 0 행을 썼으므로 지우지 **않는다**(R3) |
| 그 행이 임차에서 「이미 정산됨」(B6) | 지움 — PENDING 을 떠났다 |
| 승인 세대가 바뀜(아무 `Acknowledge`) — 사이클 시작과 울타리에서 `AckGeneration()` 을 마지막으로 본 값과 비교 | **맵 전체를 비움** — 승인이 이긴다, 그리고 승인은 행을 PENDING 에서 빼므로 남은 계수는 옛 에피소드의 것이다 |
| 오류가 났을 때 **실행자 수명 ctx**(Run 의 ctx)가 끝나 있음 | 세지 않음 — 엔진 종료. transport timeout 은 ntfy 의 자식 ctx 라 발행 실패(`perr`)로 온다 |
| 재시작 | 사라짐 — 대신 기동 복원이 PENDING 수로 다시 잠근다(`gateway.go:153-167`) |

같은 id 재무장이 비워지지 않은 계수를 물려받는 경우(승인 없이 동기 경로가 전달 → 재무장, 실행자는 그 사이 그 행을 못 봄)는 **더 일찍 잠그는** 쪽이다 —
보수 방향이라 막지 않고 기록한다. 맵의 크기는 「마지막 승인 이후 기록 오류를 낸 서로 다른 행 수」로 묶인다.

미전달 **나열** 오류(`cycle` B1)는 행이 없으므로 실행자 단위 계수 하나로 센다 — 나열 성공이 지운다(그 계수는 「나열조차 못 한다」만 잰다).

**한도에 닿으면.** 행 계수 또는 나열 계수가 `alertAttemptLimit` 에 닿으면 D7 울타리(`LatchUnlessAcknowledgedSince(적어 둔 세대, …)`)로 잠그고,
통과하면 배제 밖에서 승격을 시도한다(원장이 쓰기를 거부하는 중이면 실패할 것이다 — D9). 울타리를 못 통과하면(그 사이 승인) 계수를 버린다.
원장을 읽는 울타리(3판)는 없앴다 — 전역 수는 오류 난 에피소드가 아직 남았는지 답하지 못했다(R4).

행별 계수는 행마다 사이클당 한 번 오르므로 **최소 ≈ 2 사이클(≈ 4 s)** 이다(R8 — 3판의 실행자 단위 계수는 한 배치 안에서 3 을 채울 수 있었다).
나열 계수도 사이클당 한 번이다.

**fail-closed 가 거부할 정상 입력 (먼저 열거 — Manager 지시).**

| 정상 입력 | 결과 | 이 change 의 목적 안인가 |
|---|---|---|
| **일시적 원장 쓰기 오류가 같은 행에서 3 연속** (busy timeout 초과 · 디스크 가득 참 등; 사이클마다 한 번이라 최소 ≈ 4 s), 그 사이 승인 없음 | **운영자 승인 전까지 신규 진입 차단**, 승격 시도 | **예.** 미전달을 원장에 적지 못하는 것은 「미전달 감시가 실패했다」이고, 감시가 실패한 동안 진입하지 않는 것이 이 change 의 목적(정본 「전달 실패가 지속되면 신규 진입을 차단」)이다 |
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
  `ENTRY_BLOCKED` 로만 간다(`operating_mode.go:537-545`). D7 의 배제는 원격 전송도 원장 연산도 덮지 않는다(메모리 비교 + map 삽입). exit `Notify` 가 새로 기다릴 수 있는 그 몫은 tasks 2.6 에서 잰다 —
  `a098_the_backlog_does_not_delay_protection_test.go` 가 그대로 초록이어야 하고 publisher 없음 · 배제 경합 변형을 더한다.
- **토글 OFF = upstream**: 새 토글 없음. 1판의 *"알림 게이트 OFF 면 Notifier 자체가 없다"* 는 **틀렸다**(F7) — Notifier 는 엔진에서
  무조건 생성된다(`gateway.go:323`). 실제 OFF 경계는 automation gate OFF 에서 `engine run` 이 기동을 거부하는 것이다
  (`cmd/tossctl/engine.go:220-221` `errEngineGateOff`, `TestAGateOffEngineRefusesWithoutEnumeratingClauses` rc 0, 2026-09-26). 엔진이 없으면 실행자도 없다.
- **보수 방향**: 이 change 가 여는 문은 없다 — 오늘 안 잠기던 자리가 잠긴다. 한도는 동기 경로와 같고(D2), `Acknowledge` 의 해제 조건
  (승인 + 미전달 0)은 바꾸지 않는다. 바꾸는 안이 필요해지면 멈추고 사람 결정으로 올린다.
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
- a092 가 먼저 착지하면 정본 SHALL 이 빈다. 순서를 a092 tasks 의 착수 조건으로 못 박는다.
- a092 델타에도 「운영자의 승인은 … 되살리지 않는다」 문단(`a092…/spec.md:66`)이 있다. 두 change 가 같은 요구를 각자 싣고 아카이브되면
  정본에 사본 둘이 생긴다 → a092 21판이 그 문단을 이 change 의 요구로 가리키게 한다(tasks 5.2 에 더함).

## freeze 발견 처분 (`review.md` §0 · §0.2 · §0.3)

| id | 처분 | 자리 |
|---|---|---|
| F1 P0 | 수용 — 커밋된 증가 후 attempts, `SettleResult.Attempts` additive | D1 · tasks 2.7 · 3.4 |
| F2 P0 | 수용 — M1 = C. 4판: 차단 **적용**만 `n.mu` 아래(메모리), 승격은 밖 — **freeze 4판 REJECT(§0.4 Q1·Q2·Q3), M1 재결정 대기** | D7 · spec 「승인이 이긴다」 · tasks 2.8 · 2.9 · 3.5 |
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
| N2 P1 | 수용 — 3판: 원장 읽기 울타리 → 4판(R4): 승인 세대 울타리 | D7 인터리빙 표 · D8 · tasks 2.10 |
| N3 P1 | 수용 — 전달 정산 실패는 동기 경로처럼 즉시 잠금 + 승격, 임차 유지 | D1 전달 정산 표 · spec · tasks 2.12 |
| N4 P1 | 수용 — 3판: 실행자 단위 → 4판(R3): 행별 + 승인 세대 수명 | D8 · tasks 2.10 |
| N5 P1 | 수용 — 일반 상한 주장 철회, 전제 H 조건부 식 + 무계 경우 명명 | D6 |
| N6 P2 | 수용 — 실패 기록의 `NotFound` 는 잠금만(동기 parity) | D1 |
| N7 P1 | 수용 — spec 에 나열 오류·전달 정산 실패·로그 정제, tasks 에 인터리빙·수명·재진입 금지 | spec · tasks 2.9~2.12 |
| N8 P2 | 수용 — 배포 문장을 조건부로 | Risks |
| R1 P0 | 수용 — 배제를 메모리 전용(승인 세대 비교 + map 삽입)으로, 원장 연산은 전부 밖 | D7 · spec · tasks 2.6 · 3.5 |
| R2 P1 | 수용 — 새 로그 줄은 허용 목록 필드, 계좌·원문 오류 금지, sentinel 시험 | D9 · spec · tasks 2.11 |
| R3 P1 | 수용 — 행별 계수, `Applied` 만 지움 | D8 · tasks 2.10 |
| R4 P1 | 수용 — 전역 수 울타리 폐기, 승인 세대 울타리 | D7 · D8 · tasks 2.9 · 2.10 |
| R5 P1 | 수용 — 배제 계약 하나(메모리 전용)로 spec·tasks·처분표 정렬 | spec · tasks 3.5 · 4.1 |
| R6 P1 | 수용 — 보장을 「임차가 살아 있고 교체되지 않은 동안」으로 | D1 · spec · tasks 2.12 |
| R7 P2 | 수용 — 54 s 주장 철회, 만료 ≠ `LeaseLost` | D6 |
| R8 P2 | 수용 — 행별 계수로 최소 ≈ 4 s 복원 | D8 |
