# a124 · 계속 실패하는 배달 실행자는 진입을 막는다

- **Feature**: `FEAT-TOS-009` — Exit line truth and position policy lifecycle
- **Story**: `STORY-TOS-a124`
- **Spec**: `engine-safety` (ADDED 2)
- **위험 등급**: **High-risk** — 진입 게이트·운영 모드(Guardian·kill switch·운영 모드 경로). 적대적 Eng 리뷰 + 교차 모델 + Pre-Edit 선언.
- **출처**: a092 20라운드 사용자 결정 20-1 **ⓒ**(2026-09-25) — a098 설계 D8.4 가 「미배정 후속」으로 남긴 것을 별도 change 로 먼저 낸다.
- 작성 2026-09-25, Manager(Fable, tossos-d6). HEAD `463cc895`.

> **작성 순서**: 이 문서의 분기 주장은 전부 `analysis/function-logic/` 의 AST 산출물에서 나왔고, 산출물이
> 문서보다 **먼저** 만들어졌다(`.claude/CLAUDE.md` 「단계 건너뛰기 금지」). 함수 5개 — `alertDeliverer.cycle` ·
> `alertDeliverer.deliverOne` · `Notifier.deliver` · `Notifier.notifyCritical` · `restoreAlertEntryLatch`, 분기 48.
> 편집 대상은 앞의 둘이고 뒤의 셋은 대조 근거다.

## Why — 오늘 잰 것

정본 「등급화된 알림」은 *"전달 실패가 지속되면 신규 진입을 차단한다(SHALL)"* 이고, 시나리오 「critical 알림 전달
실패 지속」은 *"전송이 재시도 한도까지 실패하면 신규 진입이 차단되고 … 전달 복구 후 수동 확인으로 해제한다"* 다.
운영 모드 트리거 `CRITICAL_ALERT_UNDELIVERED` 의 정의도 *"critical 알림 outbox 전달 실패 지속"*
(`internal/journal/operating_mode.go:80-81`)이다.

**HEAD 에서 그 SHALL 을 지키는 자리는 전부 동기 알림 경로에 있다.**

| 자리 | 무엇 | 근거 |
|---|---|---|
| `Notifier.deliver` :484 · :520 · :571 | 전송 실패 3 종에서 `Gate.Block(ReasonAlertUndelivered)` | FLM `internal-obs--notifier.deliver` B12 · B18 · B27 |
| `Notifier.claimAndDeliver` :280 | 기록 실패에서 같은 잠금 | grep — `notifier.go` 의 비테스트 `Gate.Block(ReasonAlertUndelivered)` 는 이 넷이 전부 |
| `Notifier.notifyCritical` :219 · :228 | `escalate` → `EscalateOperatingMode(…, CRITICAL_ALERT_UNDELIVERED)` :383 | FLM `internal-obs--notifier.notifycritical` B3 · B4; `escalate` 의 호출자는 이 둘뿐 |
| `restoreAlertEntryLatch` :164 | **기동 시 1회** 미전달 수로 잠금 복원 | FLM `internal-app-engine--restorealertentrylatch` 종단; 호출자 `gateway.go:269` 하나 |

**a098 의 배달 실행자에는 하나도 없다.** `internal/app/engine/alertdelivery.go` 에 `Gate` · `Escalate` 는 0 회(grep).
`deliverOne` 의 전송 실패(B9)와 publisher 부재(B8)는 `attempts+1`(`outbox.go:471`) · 임차 반납 · 로그 한 줄로
끝나고(:205-212 · :221-226), `cycle` 은 `deliverOne` 의 결과를 받지 않아 배치 전부가 실패해도 `nil` 을 돌려준다
(:163). `deliverOne` 은 행의 `attempts` 를 **읽지 않는다**(`Attempts` 0 회, grep). a098 설계 D8.4 는 이것을 스스로
적었다 — *"배달 실행자가 매 주기 publish 에 실패하면서 계속 도는 상태는 a098 이 자동으로 안 잡는다 … 그때 진입을
막는 것은 오늘의 다섯 자리 중 `deliver` 의 실패 경로이고, 그것은 동기 알림이 한 번이라도 시도될 때만 돈다.
후속으로 남긴다"*, tasks 「안 하는 것」 :1802 *"미배정 후속"*.

**a092 가 그 동기 시도를 뺀다.** a092 델타는 *"원격 전송·시도 간 대기·시도별 실패 기록·전달 실패에 따른 진입 게이트
래치·운영 모드 승격은 루프 밖의 배달 경로가 수행해야 한다(SHALL)"* 고 적고, `claimAndDeliver` 에서 `deliver` 를
뺀다. 그러면 위 표의 첫 세 줄은 엔진 루프 경로에서 **도달 불가**가 되고, 남는 것은 넷째 줄 — 재시작해야 성립하는
차단 — 뿐이다. a092 20라운드 두 보이스가 독립으로 이것을 P0 로 냈다(A-1 = B-1): *"적힌 대로 GREEN 이면 전송이
영구 실패해도 진입이 안 막힌다."* **옮기는 task 가 a092 에 없고, 그 주인은 a098 의 실행자다.**

**같은 실행자가 새 행을 굶긴다.** `PendingAlerts` 는 `WHERE state = ? ORDER BY id` + `LIMIT ?`
(`outbox.go:518-521`), 배치 상한 `alertDeliveryBatch = 10`(:74). 실패한 행은 `PENDING` 으로 남으므로 시도를 다 쓴
오래된 행 10개가 매 사이클을 채우면 새 critical 행은 **첫 시도조차** 못 한다(a092 20라운드 B-2). 오늘은 동기 경로가
이 막힘을 우회하고, a092 뒤에는 우회가 없다. a092 델타가 이미 「굶주림 금지」 문단을 적었는데 그 성질이 HEAD 실행자에
없다 — 그 문단은 실행자의 성질이므로 이 change 로 옮긴다(사용자 결정 20-1).

## What Changes

**R1 — 배달 실행자가 지속 실패의 주인이다.** 어떤 행의 시도 수가 재시도 한도에 이르렀는데 전달되지 않았으면
배달 실행자가 `Gate.Block(ReasonAlertUndelivered)` 와 `EscalateOperatingMode(…, CRITICAL_ALERT_UNDELIVERED)` 를
부른다. 판정은 원장 행의 `attempts`(재시작을 넘긴다)에 선다 — 프로세스 메모리의 사이클 계수가 아니다.
사유 코드도 트리거도 **정본의 것을 그대로** 쓴다: 그래서 해제는 정본 「운영자가 밀린 알림을 해제한다」의 두 조건
(승인 + 미전달 0) 그대로이고, 모드 완화는 사람 승인 그대로이며, 재시작은 `restoreAlertEntryLatch` 가 오늘처럼
다시 잠근다. 동기 경로가 한 번도 시도하지 않는 구성에서 이 성질이 성립해야 한다 — 그것이 a092 뒤의 세계다.

**R2 — 행 선택은 굶주림을 만들지 않는다.** 한 사이클은 아직 한도에 이르지 않은 행을 **먼저** 고르고, 한도에 이른
행은 잔여 자리에만 들어간다. 한도에 이른 행을 버리지 않는다 — 미전달로 남고 미전달 수에 세어진다. 빠지는 것은
우선순위뿐이다(a092 델타 문단의 실체를 그대로 가져온다).

**R3 — 최악 래치 시간을 적는다.** 첫 시도 전의 큐 대기(선행 배치의 서비스 시간)를 포함한 값을 실측으로 적는다
(a092 델타 「최악 시간은 change 에 적혀야 한다」를 이 change 가 진다).

**a092 와의 관계.** a092 21판은 「exit goroutine 에서 동기 deliver 제거」로 좁히고 이 change 를 착수 조건으로
인용한다. 이 change 의 착지 전에 a092 가 `deliver` 를 빼면 정본 SHALL 이 빈다 — 순서는 a124 → a092 다.

## Non-goals

- `Notifier` 동기 경로(`claimAndDeliver` · `deliver` · `notifyCritical`)를 편집하지 않는다 — a092 의 표면이다.
  `obs` 에는 승인 세대 진입점 둘과 필드 하나를 더하고, `Acknowledge` 에는 세대 증가 **한 줄**만 넣는다(design D7 — 해제 조건 불변,
  동기 경로 함수 무편집).
- 새 사유 코드 · 새 트리거 · 새 토글 · 새 실행자를 만들지 않는다. `ReasonAlertSenderDown`(실행자 정지)의 의미도
  그대로다 — 이것은 「시도가 실패했다」이고 그것은 「시도할 주체가 없다」다(정본 결정 8-1).
- `Acknowledge` 의 해제 조건 · 운영 모드 완화 경로 · 기동 복원을 바꾸지 않는다.
- 죽은 실행자를 되살리지 않는다(정본 결정 11-1). exit 관측 루프를 만지지 않는다.
- 행을 버리는 어떤 형태의 "정리"도 하지 않는다.

## 결정 (freeze 1판 리뷰 뒤 확정 — design 2판 D2·D3·D7·D8)

- **한도 = 3** (`obs.DefaultCriticalAttempts` 인용, design D2). 1판의 *"같은 3 이면 약 6 s … 동기 경로는 약 34 s"* 는
  서로 다른 실패 양상을 비교한 **잘못된 문장**이었다(review F11). 같은 양상에서 두 경로는 같은 시간이다 — 즉시 실패 4 s,
  매번 10 s timeout 34 s.
- **publisher 부재 = 실패 시도** (design D3). 결과는 동기 경로와 같고 지연만 ≈0 s 대 ≈4~6 s 로 다르다(F13).
- **승인이 이긴다** (Manager 결정 M1 = C, design D7): 차단 **적용**만 `Acknowledge` 와 같은 `n.mu` 배제 아래(메모리 전용 — 승인 세대 비교 +
  게이트 map 삽입), 원격 전송·원장 트랜잭션·승격·로그는 밖. 판정 근거를 얻은 뒤 승인이 있었으면 그 판정은 버린다.
- **기록 자체의 실패도 지속 실패다** (Manager 결정 M2 = (ii), design D8): 한 행에서 원장에 시도를 남기지 못한 연속 횟수가 같은 한도에
  이르면 잠근다(승인이 연속을 끊는다). 발행 뒤 전달 기록 실패는 동기 경로처럼 즉시 잠근다(D1).

## Impact

- `internal/app/engine/alertdelivery.go` — `alertDeliverer` 에 게이트·계정 참조·배제 배선, `deliverOne` 한도 판정(D1)과
  연속 기록 실패 계수(D8), `cycle`/`PendingAlerts` 선택 순서. FLM 은 편집 뒤 `revision: current` 재추출.
- `internal/journal/outbox.go` — `PendingAlerts` 의 정렬(한도 인자).
- `internal/journal/alert_claim.go` — `SettleResult.Attempts` (additive 필드), `settleUnderClaim` 적용 경로에서 같은 트랜잭션
  읽기(D1). 스키마 무변경. FLM `internal-journal--journal.settleunderclaim`.
- `internal/obs/notifier.go` — `ackGen` 필드, `AckGeneration` · `LatchUnlessAcknowledgedSince`(새 leaf), `Acknowledge` 에 세대 증가 한 줄(D7).
  FLM `internal-obs--notifier.acknowledge`(편집 대상, 1.3 재추출).
- `internal/app/engine/auxiliary.go` — 실행자 생성 시 배선만.
- 시험: `a098_*` 는 그대로 초록이어야 한다(특히 `a098_the_backlog_does_not_delay_protection_test.go`). 새 RED 는 tasks 2.x.
- 스펙: `engine-safety` ADDED 2. a092 델타의 「굶주림」 문단은 a092 21판이 지운다.
