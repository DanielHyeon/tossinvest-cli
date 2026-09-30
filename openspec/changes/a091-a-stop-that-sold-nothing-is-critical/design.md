# a091 설계 — 0주는 캡이 아니라 실패다

> **4판 (2026-10-01, freeze 재리뷰 2라운드 반영)** — `review.md` 「freeze 재리뷰 2라운드」 R2-1~R2-12 와 Manager 판정 Q1~Q3
> (같은 절 「Manager 판정」). 3판 대비 바뀐 것: D1 에 알림 켜짐 게이트(Q1 — a095 정본 문자 그대로), D3 에 원인 분류와 보유 0
> 배제(Q2), D5 를 루프 몫 항목별 편성으로 다시 씀(R2-4 — 기록 실패 승격 · 종료 취소 · `n.mu` · 시세 신선도), D7 원인 계약
> (Q3), D8 계좌 가림(R2-3), 8/2 원장 재독 결과(측정 — 추론을 대체), 범위 표에 모드 통지 경로(R2-9), 수치 표현 정정(R2-11).
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

**측정 결론.** 13회 동안 계좌 보유는 5 였고(한정 항 `FloorBoundSellable` 은 `sellable < holdings` 일 때만 붙는다 —
`confirmed_floor.go:169-172`), 매도가능 수량이 0 이었다. 같은 4분 사이 보유가 엔진 밖에서 10 → 5 → 2 → 0 으로 줄었다 —
**운영자가 손으로 매도 주문을 내고 있었고, 그 미체결 매도가 남은 주식을 전부 잡아 매도가능이 0 이었다**(엔진의 로컬 미체결
매도는 하한 계산에서 따로 빼므로 — `localOpenSells` — 그것이 아니다). 2라운드 보이스 B 의 추론을 측정이 확인한다.

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
  종류 · outbox · 게이트 동작은 오늘과 같다. 달라지는 것은 0주일 때 문구(D4)뿐이다.
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
| ④ 종료 취소 | B2 · `ctx.Err() != nil` | 알림 없음(로그 한 줄) | R2-4 — 끝나는 관측의 조회 실패는 사실이 아니다 |

**③ 의 판정이 정확함(번들 근거 — `ConfirmedFloorQuantity` AST B8 · B11).** `Bound == FloorBoundHoldings ∧ Quantity == "0"` ⟺
신선한 보유 스냅숏이 0. (⇐) 보유 0 이면 `base = min(0, sellable) = 0`, `sellable` 이 0 이어도 `sellable == holdings` 라 한정 항은
Holdings 로 남고(`:169`), `floor = max(0, 0 − local) = 0 = base` 라 LocalSells 로 바뀌지 않는다(`:184`). (⇒) 한정 항이 Holdings 이고
0 이면 `floor == base`(아니면 LocalSells) 이고 `base == holdings` 다. `reconcileFloor` 는 계좌가 그 종목을 보고하지 않으면 보유 `"0"`
스냅숏을 정직하게 만든다(`exitwiring.go:216-220`). 이 동치는 시험으로 핀한다(tasks 3.3a) — riskcalc 는 편집하지 않는다.

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

새 종류는 `o.alert`(`exitloop.go:1814-1821`) → `obs.RecordOnly.Notify`(`record_only.go:45-61`, 생산 배선 `exitwiring.go:348-349`)의
기존 경로를 탄다 — **발송 경로를 만들지 않는다.** 정본 exit-policy 가 이미 편성한 항목(구조화 로그 줄 · critical 이면 임차 없는 기록
트랜잭션 · `n.mu` 대기)에 a091 이 **새 이름 둘**을 붙인다.

| 이름 | 무엇 | 크기 · 기한 | 근거 좌표 |
| --- | --- | --- | --- |
| **「0주 기록」** | 보호 0주(①②, 알림 켜짐) 사건당: `logEvent` 한 줄 + `n.mu` 대기 + `Journal.RecordAlert` 트랜잭션 하나(`BEGIN IMMEDIATE`, `synchronous=FULL`) | 원격 0. **기한 없음**: `n.mu` 보유자(다른 기록 · 운영자 승인 `Acknowledge` — 밀린 행 수에 비례, `notifier.go:957-981`)와 원장 연결 풀(`SetMaxOpenConns(1)`, `journal.go:174` — 풀 대기는 `busy_timeout` 5s 밖) | `record_only.go:136-137` |
| **「0주 기록 실패 승격」** | 기록 실패일 때만: 로그 한 줄 + `Gate.Block`(메모리) + **`escalate` — `EscalateOperatingMode` 원장 트랜잭션 하나, exit goroutine 에서 동기**(잠금 밖) | 원격 0, 기한 없음(원장) | `record_only.go:139-157` · `notifier.go:425-449` |

- **그 포지션의 손절**: 그 사이클에 제출할 것이 없다(0주) — 늦출 것이 없다.
- **같은 사이클 뒤쪽 포지션**: 두 몫만큼 늦게 판정된다(`ObserveOnce` 는 포지션을 순차로 돈다). 늦음이 시세 증거 수명
  (`QueryPriceEvidenceDuration` 15s — `retry.go:188-192`, `exitloop.go:477-495` 의 `quoteUsable`)을 넘기면 그 포지션은 그 사이클
  판정을 건너뛴다(§0.3 에 닿는 경로). 이 성질은 a092 가 이미 정본으로 인정한 critical 기록 경로의 것이고, a091 이 바꾸는 것은
  **빈도**다 — RECONCILE 동안 보호 0주 포지션마다 매 사이클 몫 하나(8/2: 3분에 13).
- **래치를 겹치지 않는다**(정본 a095 6판 원칙 — critical 은 매 관측이 기록을 시도하고 중복은 outbox 키와 재알림 창이 맡는다; 메모리
  래치는 기록 실패 · 창 뒤 재알림을 영구히 삼킨다).
- **측정 상태**: 「0주 기록」의 소요는 **이 경로에서 잰 값이 없다**. 정본의 대입 규칙(상위집합 실측을 보수적으로 대입 허용,
  대입 사실과 실측 의무를 남김 — engine-safety 「등급화된 알림」)에 따라 구현 로트가 실측한다(tasks 5.3 — 경합 없는 원장 ·
  승인 경합 · 연결 풀 경합 세 칸). 실측 전에 「주기보다 작다」를 주장하지 않는다.
- **종료 취소(④)**: 종료 중 ctx 가 끝나면 하한 조회가 ClassCanceled 로 돌아와(`retry.go:365-366`) B2 로 간다. 이때 알림을 올리면
  끝난 ctx 로 `RecordAlert` 가 실패해 **가짜 래치 · 승격 시도**가 난다(보이스 A R2-4). 그래서 B2 는 `ctx.Err() != nil` 이면 보고하지
  않는다 — 판정은 **호출자 ctx** 이지 오류 문구가 아니다(HTTP 시한의 `DeadlineExceeded` 는 ①이다).
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

## D8 — 계좌 가림 (4판, R2-3)

a091 이 편집 · 추가하는 줄은 계좌 원문을 싣지 않는다.

- B2 의 오류 로그: 현행 `logErr` 는 `obs.FieldAccount, o.opts.AccountRef`(`exitloop.go:1838`)를 싣는다 — 생산 `AccountRef` 는 계좌번호다
  (`interlock.go:680-684`, 로거는 가리지 않는다 `log.go:190-199`; 운영 `engine.log` 에 원문 줄이 실재 — 2026-10-01 재독). a091 의 B2 줄은
  계좌 필드 없이, 오류는 `obs.MaskAccount(err, AccountRef)` 로 쓴다(a092 선례). 다른 `logErr` 호출자는 건드리지 않는다.
- 새 알림: 필드에 계좌를 싣지 않는다(키는 포지션 id — 계좌 무염 해시).
- 카나리 시험: 계좌 sentinel 로 B2 로그 · 기록 로그 · 행(제목 · 본문 · payload)에 그 문자열이 없음을 단언(tasks 3.8).
- **이름 붙인 잔여**: 기록 실패 → `Notifier.escalate` 의 두 로그 줄(`notifier.go:433-444`)은 `FieldAccount` 원문을 싣는다(오류는 이미
  `MaskAccount`). 그 함수는 모든 critical 기록 경로가 공유하는 a092 · a124 영역이고, base 의 계좌 로그 관행은 사람 결정 큐 「계좌 가림
  설계」(a090 D12 · D13)에 있다. a091 은 그 경로에 닿는 **빈도**를 늘린다(알림 켜짐 · 기록 실패 때만). 넓히는 수리는 그 큐의 결정 뒤.

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
  `Notifier.escalate`(D8 잔여) · 배달 실행자 · 알림 꺼짐 엔진의 동작(D1 — 문구 제외)

## 검증 (tasks 가 RED 로 옮긴다)

- 알림 켜짐 · 보호 · ①② → 새 종류 critical · outbox 행 1 · 키 `exit.stop_sold_nothing|<pos>` · 두 경로 같은 키
- 알림 **꺼짐** · 보호 · ①② → 옛 종류 normal · outbox 행 0 · 게이트 · 모드 무변화(불변식 3)
- ③ 보유 0 → 옛 종류 normal(켜짐에서도) · 본문이 「계좌에 보유가 없다」
- ④ 종료 취소 → 알림 0 · 래치 0
- 보호 5 액션 · 익절 3 액션 표 — 익절은 어떤 원인이든 옛 종류 normal
- 부분 캡 → 종류 · 등급 · 문구 무변화
- 반환값 · 관측 결과 무변화: B2 뒤 레벨 해제(재발의 가능) · 하한이 풀리면 같은 레벨 재발의 · 제출 수량
- 로그와 알림 같은 종류(로그 캡처 하네스) · 본문 한국어 · `이름(코드)` · 원문 오류 없음 · 계좌 sentinel 없음
- 원인 두 순서(B2 → 끝, 끝 → B2)에서 행은 첫 원인, 로그는 관측마다
- 8/2 재생: 실제 `RecordOnly` + 원장 + 배달 실행자, 팔 셋(정상 전송 · 실패 전송 · publisher 없음) + 정착 행 1h 경계 · 알림 꺼짐 대조
- 「0주 기록」 실측(세 칸)
- `go test ./... -count=1 -race` 회귀 0
