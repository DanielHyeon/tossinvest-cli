# a091 설계 — 0주는 캡이 아니라 실패다

> **4판 (2026-10-01, freeze 재리뷰 2라운드 반영)** — `review.md` 「freeze 재리뷰 2라운드」 R2-1~R2-12 와 Manager 판정 Q1~Q3
> (같은 절 「Manager 판정」). 3판 대비 바뀐 것: D1 에 알림 켜짐 게이트(Q1 — a095 정본 문자 그대로), D3 에 원인 분류와 보유 0
> 배제(Q2), D5 를 루프 몫 항목별 편성으로 다시 씀(R2-4 — 기록 실패 승격 · 종료 취소 · `n.mu` · 시세 신선도), D7 원인 계약
> (Q3), D8 계좌 가림(R2-3), 8/2 원장 재독 결과(측정 — 추론을 대체), 범위 표에 모드 통지 경로(R2-9), 수치 표현 정정(R2-11).
> **5판 (2026-10-01, 3라운드 표적 재확인 반영)**: R3-1 ③ 동치를 한 방향 + 조건으로 정정 · 새는 칸의 처분, R3-2 취소 판정을 출처(오류가
> `context.Canceled`)로 · 기록은 `context.WithoutCancel`, R3-3 계좌 — 기록 실패 로그와 `escalate` 두 줄까지 가림, R3-4 몫 배정(a092
> `alertLoopShare` 750ms 대입)과 수락 기준, R3-5 알림 꺼짐 B2 는 알림 0, R3-6 원장 재독의 측정/추론 구분 · 시간 정정. `review.md` 「5판」.
> **6판 (2026-10-01, 4라운드 — A APPROVE · codex FAIL 좁음)**: 취소 억제를 「취소뿐인 오류」로(합쳐진 오류 제외), 몫 배정의 출처 정정(a092
> 17판 산정 · 21판 철회 — a091 이 소유) · 수락을 보고 호출 전체 경과로 · `Acknowledge` 밀린 행 수 고정, `WithoutCancel` 의 종료 지연(기한 없음) 명명 ·
> 시험, 진입점 명시(`o.opts.Alerts.Notify`), 로컬 매도 0 의 근거를 전 기간 intent 조회로.
> **7판 (2026-10-01, 구현 로트 정지 · 보고 → Manager 판정)**: 수락 기준 (i) 정정 — 750ms 판정은 a091 소유 셀만, 승인 경합 셀은 측정 · 기록
> (근거 셋), 5.4 실측 주입 변형, `Acknowledge` 잠금 해제는 후속 후보. `review.md` 「구현 로트 — 수락 기준 정정」.
> base 는 그대로 `b30318d6`. 분기 · 호출 주장은 `analysis/function-logic/` 번들 14 에서 나왔다.

## 8/2 원장 재독 (2026-10-01, 읽기 전용 · Manager Q2 승인)

운영 원장(`~/.config/tossctl/journal.db`, `mode=ro` + `query_only`)과 `engine.log` 를 읽었다. 계좌 필드는 읽는 자리에서 가렸다.

| 시각 (UTC) | 원장 · 로그 | 뜻 |
| --- | --- | --- |
| 07-31 03:53:52 | `position_adjustments` EXTERNAL 0 → **10**, `exit_events` OPENED | 엔진 밖에서 산 보유를 편입 |
| 08-02 23:17:16 | adjustment UNKNOWN 10 → **5**(「the account reports 5 … converged」) | **엔진 밖 매도 1** |
| 23:23:25 ~ 23:26:21 | `exit.proposal_capped` × 13, 전부 `quantity:"0"` · `proposed:"5"` · `floor_bound:"broker sellable quantity"` · `severity:normal` | 확정 하한 0 — **한정 항은 매도가능 수량**(보유는 5) |
| 23:26:39 | adjustment 5 → **2** | **엔진 밖 매도 2** |
| 23:27:41 | adjustment 2 → **0**, `exit.position_closed_externally`(normal), `exit_events` ADJUSTMENT_CLOSED | 엔진 밖 종결 |

**측정(5판 — 측정과 추론을 가름).**

- 측정: 13회 동안 계좌 보유는 5 였고 한정 항은 매도가능이었다(`FloorBoundSellable` 은 `sellable < holdings` 일 때만 —
  `confirmed_floor.go:169-172`) — 매도가능 0. 보유는 엔진 밖에서 10 → 5(23:17:16) → 2(23:26:39) → 0(23:27:41) 으로 **10분 25초**에 걸쳐
  줄었다(13 관측은 그 안의 23:23:25 ~ 23:26:21 — 3분). 그 종목의 엔진 intent 는 **전 기간 1 건이고 그것은 2026-08-18** 이다(`intents` 읽기 전용 전수 조회 — 6판, 4라운드 codex #4: 로컬 미체결 매도는
  거래일을 넘겨 셀 수 있으므로 당일 조회로는 모자랐다) — 8/2 에 엔진 주문은 없었고 로컬 미체결 매도는 0 이었다(한정 항 Sellable 은 로컬 매도가 있어도 바뀌지 않으므로 — 3라운드 codex #6 — 한정 항이 아니라 이 조회가 근거다).
- 추론(주문 수준 증거 없음): 매도가능을 잡은 것은 운영자가 손으로 낸 미체결 매도다. 엔진 밖 체결 셋이 같은 10분에 있었다는 것이 그
  추론의 근거이고, 누가 · 어떤 주문이 잡았는지는 원장에 없다.

**이 측정이 a091 에 주는 것.** 8/2 는 「보유 0」(엔진 밖 종결이 이미 끝남)이 **아니다** — Q2 의 배제(D3)에 들지 않고, 알림 켜진
엔진이었다면 a091 뒤 critical 이다. 그 알림의 참인 내용은 「손절이 매도가능 0 때문에 나가지 못했다」이고, 그것은 사람이 낸
매도가 주식을 잡고 있는 동안에도 참이다 — 그 주문이 체결될지 · 손절가보다 위에 있는지 엔진은 모른다. 등급은 **엔진이
보호하기로 한 포지션의 손절이 나가지 못했다**는 사실에 붙는다(a095 원칙 — 사실별 등급).

## D1 — 새 이벤트 종류를 만든다 · 알림 켜짐일 때만 critical

`SeverityOf`는 `criticalEvents` 맵만 보는 순수 함수다(`event.go:371-376`). **등급은 종류에 붙는다.** base 의 등급표는
**19** 종(`event.go:337-361`).

| 안 | 방법 | 문제 |
| --- | --- | --- |
| A | `EventExitProposalCapped` 를 `criticalEvents` 에 추가 | 부분 캡 · 익절 0주까지 critical |
| B | **보호 청산 0주 전용 종류 신설** + 그것만 등록 | 소비자 확인(`issues.md`) |
| C | 호출 자리 등급 지정 | class rule 셋이 실제 등급을 못 센다(`measurement_test.go:48` · `a074_quarantine_event_test.go:34` · `cmd/tossctl/a109_…_test.go:293`) |

**B 를 택한다.** 이름 `obs.EventExitStopSoldNothing = "exit.stop_sold_nothing"`(subject `exit` — 계약 문자열이므로 freeze 가 고정).

**알림 켜짐 게이트(4판, Manager Q1 — a095 정본 문자 그대로).** 정본 engine-safety 「무관리 보유 보고의 등급은 사실이 정한다」
(`spec.md:1846-1852`)는 **알림이 꺼진 엔진의 사실을 critical 로 매기지 않는다** — 매기면 보낼 수단이 없는 행이 배달 실행자에게
실패 시도로 세어져(`alertdelivery.go:316-326`) 세 사이클 뒤 진입 게이트 래치와 ENTRY_BLOCKED 승격(`:446-495`)에 닿고,
「알림을 끈 것은 운영자가 고른 상태」다. a095 는 그것을 발신 쪽 등급 선택으로 걸렀다(`adoption.go:451` —
`d.opts.NotificationsEnabled && …`). a091 은 같은 자리에서 같게 거른다:

```text
보호 청산 · 0주 · 보고 대상 원인(D3) ∧ 로드된 설정 notifications.enabled = true  →  exit.stop_sold_nothing (critical)
그 밖(알림 꺼짐 포함)                                                           →  exit.proposal_capped (normal, 오늘 등급)
```

- **게이트는 enabled 플래그 하나다(codex 세부 구분 채택).** 「알림 켜짐 + 전송 실패/부재」는 게이트 대상이 **아니다** — 그때
  래치 · 승격은 정본 a092 · a124 의 **의도된** 의미론(보낼 수단을 운영자가 켜 두었는데 닿지 않음)이다.
- **불변식 3(토글 OFF = upstream)**: 알림 꺼짐(기본값 — `internal/config/a074_notifications_test.go:28-30`) 엔진에서 a091 뒤의 등급 ·
  종류 · outbox · 게이트 동작은 오늘과 같다. **B2 는 오늘처럼 알림 0 · 로그 한 줄**(옛 종류, 계좌 없음 — D8)이고, 끝 경로는 오늘처럼
  옛 종류 normal 알림 하나다. 달라지는 것은 0주일 때 문구(D4)와 로그 줄의 계좌 필드(D8 — B2 줄, 그리고 `escalate` 두 줄은 다른 critical 종류의 줄도 바뀐다)다(5판 R3-5 · 6판).
- **꺼짐에서 남는 흔적**: 옛 종류 normal 알림(이관 버퍼)과 구조화 로그 줄 — outbox 행은 없다. 이 change 의 원장 흔적(outbox 행)
  이익은 **알림 켜진 엔진**에 한정된다(Manager Q1 판정의 「원장 흔적 유지」는 로그 흔적으로 읽는다 — 정정 기록 `review.md`).
- 배선: `ExitObserverOptions.NotificationsEnabled` 를 생산 배선이 로드된 설정에서 **덮어쓴다**(`exitwiring.go` 의 기존 덮기 규칙 —
  호출자 값과 무관, a092 25라운드 보이스 B #4 · a095 `reconcileloop.go:369` 와 같은 원천 `c.Config.Engine.Notifications.Enabled`).

**`EventExitProposalCapped` 의 남는 범위**: 부분 캡(보호 · 익절) + 익절 0주 + 보호 0주 중 게이트 밖(알림 꺼짐 · 보유 0 · 종료 취소 —
D3). 등급 normal · 부분 캡 문구 무변화.

## D2 — 보호/익절 구분은 호출자가 넘긴다

`submit`(`exitloop.go:1395`)이 `isProtective(proposal)`(`:1375-1377` — BaselineBreach · LadderStop)을 `applyFloor`(`:1397`)에 넘긴다.
5 주문 액션의 정확한 이분이다(`ratchet.go:94-118` — 2라운드 세 보이스 확인, recovery 는 여섯째 액션을 만들지 않는다 — codex:
`recovery.go:223-236` · `exit_snapshot.go:170-175`). **판정기를 건드리지 않는다.**

## D3 — 0주의 원인을 가른다

`applyFloor` 가 0 을 돌려주는 자리는 둘이다(D6): B2(하한 계산 실패, `:1622` → `:1628`) · 끝(`floor.Quantity` = `"0"`, `:1644` → `:1660`).
보호 청산에서 원인을 넷으로 가른다.

| 원인 | 판정 | 보고 (알림 켜짐) | 근거 |
| --- | --- | --- | --- |
| ① 하한 계산 실패 | B2 · 호출자 ctx 살아 있음 | **새 종류 critical** | 하한을 모르는 채 손절이 안 나감 — fail-closed 는 옳고 보고가 빠져 있었다 |
| ② 매도가능 · 로컬 매도 · 스냅숏 | 끝 · `Bound` ∈ {Sellable, LocalSells, NoSnapshot, StaleSnapshot} | **새 종류 critical** | 8/2 가 이것(Sellable) |
| ③ **보유 0**(엔진 밖 종결 진행 중) | 끝 · `Bound == FloorBoundHoldings` | 옛 종류 normal(Q2 배제) | 정본 `EventExitPositionClosedExternally` = normal 과 같은 사실 — 보호할 주식이 계좌에 없다 |
| ④ 종료 취소 | B2 · 하한 오류가 **취소뿐**(오류 나무의 잎이 전부 `context.Canceled`) **이고** 호출자 `ctx.Err() != nil` | 알림 없음(로그 한 줄) | R2-4 · R3-2 — 오류 자체가 취소일 때만. 끝나는 관측의 조회 실패는 사실이 아니다 |

**③ 의 판정(5판 정정 — 3라운드 세 보이스 수렴 R3-1).** 4판의 「⟺」는 거짓이었다. 참인 것은:

- (⇒, 정확) `Bound == FloorBoundHoldings ∧ Quantity == "0"` 이면 신선한 보유 스냅숏이 0 이다: 한정 항이 Holdings 이고 0 이면
  `floor == base`(아니면 LocalSells, `:184`) 이고 `base == holdings`(아니면 Sellable, `:169`) 다(`ConfirmedFloorQuantity` AST B8 · B11).
- (⇐, 조건부) 신선한 보유 0 이 Holdings 한정 0 을 내는 것은 **매도가능 스냅숏도 신선하고 로컬 수량이 유효할 때만**이다. 매도가능이
  없으면 `zeroFloor(Sellable)` — NoSnapshot, 낡았으면 StaleSnapshot(`:150-158`), 로컬 수량이 비정상이면 오류(`:141-144`).

**새는 칸의 처분**: 보유 0 을 읽었는데 ③ 로 떨어지지 않는 칸은 셋이다 — (a) 매도가능 조회 실패 → B2 ①(`exitwiring.go:231-240`),
(b) 매도가능 조회가 길어 보유 스냅숏이 평가 시각(`Now` 는 두 조회 뒤 — `exitwiring.go:248`)에 낡음(한계 `AccountSnapshotStaleness` 10s — `riskcalc.go:91`, 보이스 A) → ② StaleSnapshot,
(c) 로컬 수량 오류 → B2 ①. 셋 다 **critical 로 남긴다**(과보고 방향 — 보고가 틀려도 손절 판정은 같다). ③ 를 넓히려면 `reconcileFloor`
가 보유 0 을 입력으로 넘겨야 하고 그것은 riskcalc · 하한 공급자 편집이다 — 범위 밖, 이름 붙인 잔여. 이 칸들은 시험 표에 넣는다(tasks 3.3a).
**또 하나의 잔여**: ③ 는 보유 조회 한 번의 0 을 믿는다 — 브로커 응답이 일시적으로 종목을 빠뜨리면 실제 실패가 normal 로 내려간다(base 와 같은 성질).

익절의 두 경로는 종전 그대로(B2: `logErr` 옛 종류 · 알림 없음, 끝: 옛 종류 normal). 보호 B4 · B6(`:1634` · `:1641` — 십진 비교 ·
빼기 실패)은 **범위 밖**이다: 0 이 아니라 오류를 돌려주고 `submit` B1(`:1398`)이 그 오류를 올린다 — 발의는 해제되지 않고
무장된 채 남으며 관측 루프의 오류 경로가 보고한다. 0 투영 보호 액션(1주 미만 포지션 — `snapshot.go:136-144`)도 범위 밖이다:
`applyFloor` 에 오지 않는다(`record` B11). 번들의 「유일한 자리」 문장은 「**확정 하한이** 손절을 0 으로 깎는 유일한 자리」로 좁혔다.

**H2(로그와 알림은 한 종류)**: 한 사건의 로그 줄과 알림은 그 엔진 설정에서 고른 **같은 종류**다. 보호 · 알림 켜짐 · ①이면 B2 의
오류 로그(`logErr` 대체 — D8)와 기록 로그(`RecordOnly` 의 `logEvent`, `record_only.go:51`)가 둘 다 새 종류다. 게이트 밖이면 둘 다 옛 종류다.

## D4 — 문구

0주일 때 「일부만 나갔다」(`:1647`)는 거짓이다. 보호 0주(두 종류 모두)와 익절 0주의 제목 · 본문을 **「한 주도 나가지 않았다」**로
맞추고, 정본 a085 의 문구 규칙을 따른다(`engine-safety/spec.md:801-813` — 한국어 · `이름(코드)` · 계좌 없음). 본문에 원인을 한국어
범주로 적는다(D7). 부분 캡 문구는 **건드리지 않는다**(참이다). 이 문구 결함은 8/2 의 증거가 아니다(리뷰 M2 — 8/2 로그는 영어였다).

## D5 — 루프에 남는 몫 (정본 exit-policy 「관측 경로와 fail-safe」의 편성 규칙, 4판 다시 씀)

새 보고의 진입점은 **`o.opts.Alerts.Notify(context.WithoutCancel(ctx), e)`** 다 — 생산에서 `obs.RecordOnly.Notify`(`record_only.go:45-61`,
배선 `exitwiring.go:348-349`)이고, `o.alert`(`exitloop.go:1814-1821`)와 같은 부품을 부르되 `o.alert` 의 실패 로그(`logErr`, 계좌 원문)를
거치지 않는다(D8). 재알림 창은 그 부품의 것(1h)이 그대로 쓰인다 — `opts.Critical.RecordCritical(…, 0)` 로 바꾸면 창이 0 이 되어 5.1a 가
깨진다. **발송 경로를 만들지 않는다.** 정본 exit-policy 가 이미 편성한 항목(구조화 로그 줄 · critical 이면 임차 없는 기록
트랜잭션 · `n.mu` 대기)에 a091 이 **새 이름 둘**을 붙인다.

| 이름 | 무엇 | 크기 · 기한 | 근거 좌표 |
| --- | --- | --- | --- |
| **「0주 기록」** | 보호 0주(①②, 알림 켜짐) 사건당: `logEvent` 한 줄 + `n.mu` 대기 + `Journal.RecordAlert` 트랜잭션 하나(`BEGIN IMMEDIATE`, `synchronous=FULL`) | 원격 0. **기한 없음**: `n.mu` 보유자(다른 기록 · 운영자 승인 `Acknowledge` — 밀린 행 수에 비례, `notifier.go:957-981`)와 원장 연결 풀(`SetMaxOpenConns(1)`, `journal.go:174` — 풀 대기는 `busy_timeout` 5s 밖) | `record_only.go:136-137` |
| **「0주 기록 실패 승격」** | 기록 실패일 때만: 로그 **두** 줄(`recordCritical` 의 가린 줄 + a091 보고의 실패 줄 — D8) + `Gate.Block`(메모리) + **`escalate` — `EscalateOperatingMode` 원장 트랜잭션 하나와 로그 한 줄, exit goroutine 에서 동기**(잠금 밖) | 원격 0, 기한 없음(원장) | `record_only.go:139-157` · `notifier.go:425-447` |

- **그 포지션의 손절**: 그 사이클에 제출할 것이 없다(0주) — 늦출 것이 없다.
- **같은 사이클 뒤쪽 포지션**: 두 몫만큼 늦게 판정된다(`ObserveOnce` 는 포지션을 순차로 돈다). 늦음이 시세 증거 수명
  (`QueryPriceEvidenceDuration` 15s — `retry.go:188-192`, `exitloop.go:477-495` 의 `quoteUsable`)을 넘기면 그 포지션은 그 사이클
  판정을 건너뛴다(§0.3 에 닿는 경로). 이 성질은 a092 가 이미 정본으로 인정한 critical 기록 경로의 것이고, a091 이 바꾸는 것은
  **빈도**다 — RECONCILE 동안 보호 0주 포지션마다 매 사이클 몫 하나(8/2: 3분에 13).
- **래치를 겹치지 않는다**(정본 a095 6판 원칙 — critical 은 매 관측이 기록을 시도하고 중복은 outbox 키와 재알림 창이 맡는다; 메모리
  래치는 기록 실패 · 창 뒤 재알림을 영구히 삼킨다).
- **배정(6판 정정)**: 두 몫의 합에 사건당 **750ms 를 a091 이 배정한다**(이름 「0주 보고 몫」). 출처: a092 17판이 같은 규칙(실측 최악의 2배
  이상)으로 산정한 값이다 — 그 근거 356.1ms 는 claim + 실패 기록 + 게이트 래치 + 승격 트랜잭션 + 로그 전부를 잰 최악이다
  (`archive/2026-09-30-a092-…/design.md:1094-1098`). **그 상수 `alertLoopShare` 는 a092 21판에서 철회되어 착지하지 않았다**(a092 `tasks.md:1709` 「21판 이후 철회 — 알림 예산 상수 파일 …」,
  저장소 Go 에 `alertLoopShare` 0 건 — 4라운드 보이스 A) — a091 은 그 값을 인용하는 것이 아니라 **같은 측정 위에 자기 배정을 둔다**. 그 측정은 쓰기 경합 · `EnqueueAlert` 경합
  아래였고 `Acknowledge` 의 `n.mu` 보유는 재지 않았다(a092 `design.md:1685-1692`) — 그 칸의 근거는 5.3 의 실측과 멈춤 규칙뿐이다.
  750ms 는 관측 주기 5s 보다 작다(정본의 판정). **이 경로를 직접 잰 값은 아직 없다.**
- **수락 기준(7판 — 구현 로트 정지 · 보고 뒤 Manager 판정 2026-10-01, tasks 5.3 · 5.4)**: (i) **a091 이 소유한 셀**(경합 없는 원장 ·
  기록 실패 승격 · 연결 풀 경합)은 보고 호출 전체 경과의 관측 최악 ≤ 750ms — 실측 11~26ms(실패 승격 0.1ms 미만). **승인 경합 셀(`Acknowledge`
  가 밀린 행 100 개를 승인하는 중)은 측정 · 기록**이다(실측 최악 1.06~1.28s, 2026-10-01 · 20 보고 × 5 판). 근거 셋: ① 이 대기는 a092 정본이 이미
  이름 붙인 항(「승인은 그 잠금 아래에서 밀린 행을 하나씩 승인한다 — 밀린 행 수에 비례」)이고 기존 exit critical 전부가 같은 대기를 진다 —
  a091 은 빈도만 더한다. ② 관측 주기 유계: 실측 최악 1.28s < 관측 주기 5s(정본 「주기보다 작아야 한다」 충족 — 시험이 이 부등식을 판정한다).
  ③ 750ms 배정의 출처 측정(356.1ms)은 이 셀을 재지 않았다 — 배정을 실측에 맞춰 올리는 것은 남의 항을 a091 예산으로 세탁하는 것이라
  기각(Manager). (ii) 같은 사이클에서 보호 0주 포지션 뒤에 선 **보호 포지션**이, 앞 포지션의 보고에 **배정값(750ms)과 실측 최악(1.3s)**
  지연을 각각 주입해도 그 사이클에 판정 · 제출된다(증명은 측정값으로 — Manager 보강 1).
- **실측(i1 뒤, 2026-10-01, 커밋 `2567df05` · 개발 머신 · 플래그 없음 · 20 사이클 · 관측 사이클 전체 경과 — 보고 몫의 상한)**: 소유 칸 — 경합 없음
  14~16ms, 기록 실패 + 실제 승격 트랜잭션 16~21ms, 연결 풀 경합 54~69ms. 승인 칸(밀린 행 100 을 측정 전에 채움) 0.30~0.34s. 앞선 측정(Notify 만,
  행 채움이 측정 중에 섞임, 일부는 `-race` · 커버리지 실행)은 소유 칸 11~26ms · 승인 칸 1.06~1.28s 였다 — 재는 양과 조건이 달라 같은 줄에 놓지
  않는다. 5.4 의 실측 주입값 1.3s 는 그 앞선 최악을 올림한 값으로 남긴다(보수 방향).
- **후속 후보(정식 기재)**: 「`Notifier.Acknowledge` 의 행별 · 배치 잠금 해제 — 셈-해제 배제 불변식(a092 · a124) 아래의 별도 change」.
  실측 좌표: `internal/obs/notifier.go:957-981`(승인이 `n.mu` 를 쥔 채 `PendingAlerts` · 행마다 `AcknowledgeAlert` — 행당 fsync 트랜잭션 약 12ms),
  a091 `TestA091TheReportFitsItsShare` 승인 경합 셀 1.06~1.28s(밀린 행 100).
- **종료 취소(④, 5판 · 6판)**: 억제는 **출처**로 판정한다 — 하한 오류가 **취소뿐**이고(오류 나무를 `Unwrap() error` · `Unwrap() []error` 로
  내려가 잎이 전부 `context.Canceled`) 호출자 ctx 가 끝났을 때만. `errors.Is(err, context.Canceled)` 하나로는 모자라다 — 생산 `Retrier.Query`
  는 인증 거절과 승격 실패를 **합쳐** 돌려주므로(`retry.go:357-365` · `:409-422`) 종료가 승격의 원장 연산을 실패시키면 「인증 거절 + 취소」가
  되고, 그것을 억제하면 진짜 브로커 실패를 숨긴다(4라운드 codex #1). 재시도 대기 중 취소되면 `Query` 는 **직전 일시 오류**를 돌려준다
  (`retry.go:383-385`) — 그것은 취소뿐이 아니므로 ① 로 기록된다(재시작 뒤 배달 — 델타의 「취소와 무관한 실패는 취소가 뒤따라도 대상」과 같다,
  이름 붙인 성질). HTTP 시한 · 호출자 기한의 `DeadlineExceeded` 는 ①이다(exit 루프 ctx 에는 사이클 기한이 없다 — `runtime.go:299` · `:345`).
  **기록은 `context.WithoutCancel(ctx)` 로 한다**(a094 `recordThawReleaseFailure` 선례) — 판정 뒤 · 기록 전에 종료가 끼어들어도 가짜 래치 ·
  승격이 없다. 종료 순서는 안전하다: 런타임은 `loopCtx` 를 취소한 뒤 모든 루프를 기다리고(`runtime.go:345-347`) 원장은 `Run` 이 돌아온 뒤
  닫힌다(`cmd/tossctl/engine.go:224`). **대가(이름: 「종료 중 보고 대기」)**: 그 기록은 `n.mu` · 연결 풀 대기에 기한이 없으므로 종료가 그만큼,
  **기한 없이** 늦을 수 있다(원격 0). 시험으로 잰다(tasks 3.3c).
- **이름 붙인 잔여 — 취소 원인이 지워지는 오류(5라운드 codex P2)**: 공식 클라이언트는 요청 · 본문 읽기 오류를 문자열로 감싼다
  (`fmt.Errorf("%w: %s", ErrTransport, err)` — `internal/official/client.go:196` · `:203`). 그래서 **HTTP 요청 중의 종료 취소**는 잎이 `ErrTransport` 인
  오류로 오고 `Retrier` 는 그것을 일시 오류로 분류한다(`retry.go:81-82`) — 취소뿐이 아니므로 ① 로 기록된다(과보고 방향, 가짜 래치는
  `WithoutCancel` 로 없다). 실제로 억제되는 취소는 오류 나무가 `context.Canceled` 를 보존하는 경우(원장 읽기 `localOpenSells` 의 `%w`,
  `Retrier` 의 `ClassCanceled` 반환)뿐이다. 3라운드 보이스 A 의 「HTTP 중 취소는 `*url.Error` 가 감싼다」는 공식 클라이언트에서 거짓이다.
  공식 클라이언트를 고치는 것은 범위 밖 — 시험으로 과보고를 핀한다(tasks 3.3b (vii)).
- **브로커 비용과의 비교 문장은 뺐다**(R2-11 — 3판의 「수 자릿수 크다」는 실측 없는 로컬 트랜잭션을 가정한 HTTP 추정과 견준 것이었다).

## D6 — M1 의 해소: 0주 판정은 철자에 의존하지 않는다

3판 그대로(2라운드 세 보이스 좌표 확인). `isZeroQuantity`(`exitloop.go:1871-1878`)는 수치 비교다. 입력 불변식의 생산 출처:

| 값 | 생산 | 0 의 철자 | 번들 |
| --- | --- | --- | --- |
| 포지션 수량 | `Converger.ConvergeQuantities` — `NewQuantity: mismatch.Authority()`(`converge.go:218` · `compare.go:266`) | 무관 — exit 루프가 수치 0 포지션을 건너뛰고(`exitloop.go:541`), `canonicalSnapshotContext` 가 양수 강제 + `RatString`(`snapshot.go:263` · `:267`) | convergequantities · canonicalsnapshotcontext |
| 투영 수량 | `ProjectWholeShares` → `units.String()`(`snapshot.go:93`) | `"0"` 하나 | projectwholeshares |
| 주문 가능성 | `orderable = projected != "0"`(`snapshot.go:143` · `:208`), `ExecutableProposal`(`:55`), `record` B11(`exitloop.go:1314`) | — | evaluate*snapshot · executableproposal · record |
| 하한 | `ConfirmedFloorQuantity` — `MaxDecimal` → `CanonicalDecimal`(`decimal.go:92-101`) 또는 `zeroFloor` `"0"`(`confirmed_floor.go:236-243`) | `"0"` 하나 | confirmedfloorquantity |

## D7 — 원인 계약: 에피소드 첫 원인은 행에, 관측마다의 원인은 로그에 (4판, Manager Q3)

사건 키는 `exit.stop_sold_nothing|<position id>` 하나다(끝 경로 현행 모양 `type|position`, `:1646`). **원인별 키를 쓰지 않는다** —
정본 a092 의 에피소드 의미론(행 = 에피소드)과 충돌하고, 8/2 처럼 원인이 흔들리면 13회가 원인마다 새 행이 된다.

outbox 는 키당 행 하나이고 PENDING 행은 첫 기록의 제목 · 본문을 유지한다(`outbox.go:285-345` — 재무장 때만 새 내용, a097). 배달은
제목 · 본문만 보낸다(`alertdelivery.go:324-332`). 그래서 계약은:

- **행 본문 = 에피소드 첫 원인**(한국어 범주 + 관측 시각 UTC). 운영자가 늦게 읽어도 거짓이 되지 않게 시각을 싣는다(정본 a095 의
  「같은 key는 … 먼저 온 자리의 사유만 남는다」 함정을 문장으로 인정).
- **관측마다의 원인 = 구조화 로그 줄**. `RecordOnly.Notify` 는 필드를 지우고 로그하지만(`withoutFields`, `record_only.go:51`)
  **본문은 detail 로 남긴다**(`logEvent`, `notifier.go:168-170`) — 매 관측의 `Notify` 가 그 관측의 원인 · 시각을 담은 본문으로
  로그 한 줄을 쓴다. 행이 PENDING 이어도 로그는 매번 쓰인다.
- B2 의 원문 오류(영어 · 원장 · 계좌를 품을 수 있음)는 **본문에 넣지 않는다** — 가린 오류 로그 한 줄에만(D8).

3판 델타의 「원인은 세부 정보로 구분해 담아야 한다」는 이 계약으로 **정정**했다(정정 경위 `review.md` 「4판」).

## D8 — 계좌 가림 (4판 R2-3 · 5판 R3-3)

a091 이 편집 · 추가하거나 **새로 닿게 하는** 줄은 계좌 원문을 싣지 않는다(안전 불변식 8 · Manager 상임 지시 「계좌 원문 로그 금지」).

- **B2 오류 줄(모든 경우 — 보호 · 익절 · 게이트 밖)**: 현행 `logErr` 는 `obs.FieldAccount, o.opts.AccountRef`(`exitloop.go:1838`)를 싣는다 —
  생산 `AccountRef` 는 계좌번호다(`interlock.go:680-684`, 로거는 가리지 않는다; 운영 `engine.log` 에 원문 줄 66,009 — 2026-10-01 재독).
  B2 자리의 줄은 계좌 필드 없이 쓰고 오류는 `obs.MaskAccount(err, AccountRef)`. 종류는 D3 이 고른 것.
- **a091 보고의 기록 실패 줄**: 새 보고는 `o.alert` 를 거치지 않고 알림기를 직접 부른다 — `o.alert` 의 실패 로그는 `logErr`(계좌 원문,
  `exitloop.go:1818-1819`)이기 때문이다(3라운드 codex #3 · 보이스 A #8). 실패 줄은 계좌 필드 없이 가린 오류로.
- **`Notifier.escalate` 의 두 줄(`notifier.go:433` · `:440`)**: `FieldAccount` 원문을 **뺀다**(5판 — 3라운드 codex P0: 이름 붙인 잔여로는 불변식이
  만족되지 않는다). 그 함수는 critical 기록 실패 · 동기 발송 실패의 공유 경로(`notifier.go:238` · `:261` · `record_only.go:157`)이고,
  편집은 로그 필드 하나 — 판정 · 반환 · 원장 호출은 같다(번들 `notifier.escalate`). 알림기는 계좌 하나에 묶이므로 뜻은 잃지 않는다.
- 새 알림: 필드에 계좌를 싣지 않는다(키는 포지션 id — 계좌 무염 해시).
- 카나리(tasks 3.8): 계좌 sentinel 로 — B2 줄 · 기록 줄 · 기록 실패 줄 · `escalate` 성공/실패 줄 · 행(제목 · 본문 · payload) 어디에도 없음.
- **남는 잔여(이름)**: exit 루프의 **다른** `logErr` 호출자 · base 의 나머지 계좌 필드 줄은 건드리지 않는다 — 사람 결정 큐 「계좌 가림 설계」
  (a090 D12 · D13). a091 이 새로 닿게 하는 줄은 위에서 전부 가렸다.

## 범위 완결성 — exit 관측 goroutine 에서 알림 경로에 닿는 경로 (4판, R2-9)

정본 a092 의 방법(호출 자리가 아니라 **알림 경로에 도달하는 경로**)으로 다시 셌다.

| 경로 | 이벤트 | 등급 | 자리 |
| --- | --- | --- | --- |
| `checkOutage` | `EventExitObservationOutage` | CRITICAL | `exitloop.go:866` |
| `checkOutage` → `EscalateOperatingMode(…, Announcer)` | `EventOperatingMode` | CRITICAL | `exitloop.go:882-883` |
| a090 포지션 단위 미관측 | `EventExitObservationOutage` · 그 강화의 `EventOperatingMode` | CRITICAL | `exit_unobserved.go:248` · `:229-230` |
| **`applyFloor`** | **`EventExitProposalCapped`** | **normal** | **`exitloop.go:1644`** ← 이 change |
| `applyFloor` → `ConfirmedFloor` → exit Retrier 401/403 강화 | `EventOperatingMode` | CRITICAL | `exit_record_only.go:15-21` · `retry.go:360-364` |
| 가격 조회 → exit Retrier 401/403 강화 | `EventOperatingMode` | CRITICAL | 같은 Retrier |
| `alertUnmanaged` | `EventExitPositionUnmanaged` | normal | `exitloop.go:1714` |
| `alertRefused` · `alertProposalRefused` · `noteDelay` | Judgement/Proposal Refused · LiquidationDelayed | CRITICAL | `:1740` · `:1764` · `:1794` |
| a094 판정 진입 알림 | `EventExitLiquidationDelayed` · `EventOrderUnresolved` | CRITICAL | `exit_held_proposal.go:203` · `:222` · `:246` · `:284` |
| `announceQuarantine` | `EventExitSnapshotQuarantined` | CRITICAL | `exit_quarantine_announce.go:71` |

normal 은 둘이고 `EventExitPositionUnmanaged` 는 승격하지 않는다(a095 정본 — 운영자가 고른 상태). **겹침**: `applyFloor` 한 호출에서
하한 조회의 401 이 모드 통지 기록(critical)을 쓰고 B2 로 돌아와 a091 의 새 critical 기록을 또 쓴다 — 한 호출에 기록 둘(「0주 기록」 ×
1 + 정본이 이미 편성한 모드 통지 기록 1). 시험으로 핀한다(tasks 3.9).

## 건드리지 않는 것

- **제출 수량 계산** — `applyFloor` 의 반환값(B1~B6 · 끝)은 한 글자도 바꾸지 않는다(§0.3 · §0.9)
- **브로커 요청** — `ConfirmedFloor` 의 RECONCILE 읽기(`exitwiring.go:207` Holdings · `:231` SellableQuantity, 각각 `Retrier.Query`)를
  포함해 무변경. a091 이 더하는 브로커 요청은 0(§0.4)
- 0주의 원인 자체(대사 영역) · 부분 캡 등급 · 문구 · 익절 0주 등급 · `exitpolicy` 판정기 · `riskcalc` · 스키마 ·
  `Notifier.escalate` 의 판정(로그 필드만 편집 — D8) · 배달 실행자 · 알림 꺼짐 엔진의 동작(D1 — 문구 제외)

## 검증 (tasks 가 RED 로 옮긴다)

- 알림 켜짐 · 보호 · ①② → 새 종류 critical · outbox 행 1 · 키 `exit.stop_sold_nothing|<pos>` · 두 경로 같은 키
- 알림 **꺼짐** · 보호 · ① → 알림 0 · 로그 한 줄 / ② → 옛 종류 normal 알림 하나 · outbox 행 0 · 게이트 · 모드 무변화(불변식 3)
- ③ 보유 0 → 옛 종류 normal(켜짐에서도) · 본문이 「계좌에 보유가 없다」
- ④ 종료 취소(취소뿐인 오류) → 알림 0 · 래치 0 · 합쳐진 오류(인증 거절 + 취소)는 ① · 진짜 오류 뒤 종료는 ① · 판정 뒤 종료가 끼어도 가짜 래치 0 ·
  막힌 기록 동안 종료는 기록이 끝날 때까지 기다린다
- 보호 5 액션 · 익절 3 액션 표 — 익절은 어떤 원인이든 옛 종류 normal
- 부분 캡 → 종류 · 등급 · 문구 무변화
- 반환값 · 관측 결과 무변화: B2 뒤 레벨 해제(재발의 가능) · 하한이 풀리면 같은 레벨 재발의 · 제출 수량
- 로그와 알림 같은 종류(로그 캡처 하네스) · 본문 한국어 · `이름(코드)` · 원문 오류 없음 · 계좌 sentinel 없음
- 원인 두 순서(B2 → 끝, 끝 → B2)에서 행은 첫 원인, 로그는 관측마다
- 8/2 재생: 실제 `RecordOnly` + 원장 + 배달 실행자, 팔 셋(정상 전송 · 실패 전송 · publisher 없음) + 정착 행 1h 경계 · 알림 꺼짐 대조
- 보고 호출 전체 경과 실측(세 칸, 밀린 행 100) 관측 최악 ≤ 750ms · 뒤쪽 보호 포지션이 배정 지연 아래 판정 · 제출
- 보유 0 새는 칸 셋이 critical 로 남음 · 알림 꺼짐 B2 알림 0 · 계좌 카나리(실패 경로 · `escalate` 포함)
- `go test ./... -count=1 -race` 회귀 0
