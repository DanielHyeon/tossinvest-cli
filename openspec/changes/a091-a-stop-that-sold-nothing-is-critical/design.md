# a091 설계 — 0주는 캡이 아니라 실패다

> **3판 (2026-10-01, 재freeze 입력)** — base `b30318d6`(a092 아카이브 `75d138b5` 뒤)에서 좌표 · 수치 · 전제를 다시 쟀다.
> 2판(2026-09-30 Manager 재작성)과 달라진 것: D1 에 새 종류의 **이름**과 「부분 캡 전용」 문장의 정정(익절 0주가 남는다) ·
> a095 선례, D3 에 익절 B2 의 처분, D5 에 착지한 기록 경로의 수치와 루프에 남는 몫의 이름, M1 의 해소(D6 신설),
> 범위 표 재집계(호출자 단위), §0.4 문장 정정(`applyFloor` 는 RECONCILE 에서 브로커를 읽는다 — a091 이 더하는 요청은 0).
> 분기 · 호출 주장은 전부 `analysis/function-logic/` 번들 14 개(base AST · 커버리지)에서 나왔다.

## D1 — 새 이벤트 종류를 만든다

`SeverityOf`는 `criticalEvents` 맵만 보는 순수 함수다(`event.go:371-376`, AST branches 1). **등급은 이벤트 종류에 붙어
있다.** 같은 종류 안에서 등급을 나눌 수 없다. base 의 등급표는 **19** 종이다(`event.go:337-361` — 주문 · 브로커 · 알림 ·
모드 · 루프 13, exit 관측 5, a095 편입 실패 1).

선택지는 셋이다(리뷰 H1 — 초판의 A/B 이분법이 C를 빠뜨렸다).

| 안 | 방법 | 문제 |
| --- | --- | --- |
| A | `EventExitProposalCapped`를 `criticalEvents`에 추가 | **부분 캡 · 익절 0주까지 critical**이 된다. 8/2가 보여준 결함은 보호 청산의 0주다 |
| B | **보호 청산 0주 전용 종류 신설** + 그것만 `criticalEvents`에 등록 | 소비자 확인 필요(`issues.md` 소비자 조사 — base 에서 깨지는 소비자 0) |
| C | 이벤트별 override — 호출 자리에서 등급 지정 | `SeverityOf` 순수성이 깨지고, **종류 열거를 훑는 class rule 이 실제 등급을 세지 못한다** |

**B를 택한다.** 근거는 C가 부수는 것이다(리뷰 H1): `obs.CriticalEvents()`를 순회하는 class rule 이 base 에 셋 있다 —
`measurement_test.go:48`(subject `measurement` 금지) · `a074_quarantine_event_test.go:34`(격리 사건 등재) ·
`cmd/tossctl/a109_…_test.go:293`(강등 사건 비등재). 셋 다 등급이 종류에 붙어 있다는 전제 위에 있다 — per-call override 로
critical 이 된 사건은 어떤 class rule 도 세지 못한다. **a095 가 같은 모양으로 착지했다**: 같은 무관리 보고 가운데 사실
하나(편입 시도 실패)만 새 종류 `EventExitPositionAdoptionFailed` 로 떼어 critical 로 올렸다(`event.go:255-258` · `:358-360`).

**이름(3판 — freeze 가 고정한다)**: `obs.EventExitStopSoldNothing` = `"exit.stop_sold_nothing"`. subject 는 `exit`
(`Subject()` — `log.go:196` 의 `FieldSubject`). 로그 필터 · outbox `event_type` 에 남는 계약 문자열이므로 이름은 freeze 에서
정한다(바꾸면 기존 배포의 필터가 조용히 무효 — a109 가 같은 이유로 값을 핀했다).

**`EventExitProposalCapped` 의 남는 범위(3판 정정)**: 2판은 「부분 캡 전용으로 좁아진다」고 썼다 — 거짓이다. 이 change 뒤
그 종류는 **보호 청산의 0주를 뺀 모든 캡**을 뜻한다: 부분 캡(보호 · 익절) + **익절의 0주**(D3 · 델타 「익절이 0주로
깎인다」 — 종전 등급). 등급 normal · 부분 캡 문구 무변화.

## D2 — 보호/익절 구분은 호출자가 넘긴다

`applyFloor`는 제안이 손절인지 모른다(번들 Inputs 표 — 맥락이 인자에 없다). `submit`은 `proposal exitpolicy.Proposal`을
갖고 있고(`exitloop.go:1395`) `isProtective`가 이미 있다(`:1375-1377` — BaselineBreach · LadderStop). 인자 하나를 더
넘긴다(`:1397`). 첫 리뷰가 확인한 이분(`Action.Orderable()` 5종의 정확한 이분 — 누락 · 오분류 없음)은 base 에서 무변경이다.

**판정기를 건드리지 않는다.**

## D3 — 0주가 되는 두 경로를 같게 다룬다

| 경로 | 조건 | base 현재 |
| --- | --- | --- |
| B2 `:1622` → `:1628` | 확정 하한을 계산할 수 없다(`ConfirmedFloor` 오류) | `logErr(EventExitProposalCapped, …)` 한 줄(`:1626`), **알림 없음**, 반환 `"0", true` |
| 끝 `:1644` → `:1660` | 확정 하한이 원안보다 작고 그 값이 `"0"` | `alert(EventExitProposalCapped)`(normal → 이관 버퍼, outbox 행 0) |

**보호 청산**이면 두 경로 모두 새 종류 · critical 로 보고하고 원인은 detail 에 담는다(델타 문단 1). B2 는 알림을
**추가**한다. `logErr`는 유지한다(오류 객체를 담는 유일한 자리). **B2 의 `logErr` 종류도 새 종류로 바꾼다(리뷰 H2)** —
로그와 알림은 한 종류다. 알림 경로의 로그는 이미 한 종류다: `RecordOnly.Notify` 가 `logEvent`(`record_only.go:51`)를 사건의
종류로 쓴다. 그래서 보호 B2 는 로그 두 줄(오류 객체의 `logErr` · 기록의 `logEvent`)이 **같은 새 종류**로 남는다.

**익절**의 두 경로는 종전 그대로다: B2 는 `logErr(EventExitProposalCapped)` 한 줄 · 알림 없음, 끝은 `EventExitProposalCapped`
(normal). 단 끝 경로의 **문구**는 0주면 결과에 맞춘다(D4 — 델타의 문구 문장은 등급과 무관하게 적용된다).

## D4 — 문구

0주일 때 「일부만 나갔다」(`:1647`)는 거짓이다. 제목과 본문을 결과에 맞춘다 — 보호 0주(새 종류)와 익절 0주(옛 종류) 둘 다.
부분 캡의 문구는 **건드리지 않는다**(참이다).

이 문구 결함은 8/2 사건의 증거가 아니다(리뷰 M2) — 8/2 로그 본문은 영어였고 한국어 제목은 8/2 **이후**에 들어왔다.
현행 코드의 정확성 결함으로서 고친다.

## D5 — 새 critical 은 a092 의 기록 계약 위에 선다

a092 는 아카이브됐다(`75d138b5`). exit 관측 goroutine 의 알림 부품은 `obs.RecordOnly`(`exitwiring.go:348-349`)이고,
새 종류는 `o.alert`(`exitloop.go:1814-1821`)의 기존 경로를 그대로 탄다 — **a091 은 발송 경로를 만들지 않는다.**

| 등급 | 경로 (`RecordOnly.Notify`, `record_only.go:45-61`) | 루프가 기다리는 것 |
| --- | --- | --- |
| normal(현행 0주) | `logEvent` 한 줄 → `NormalRelay.Offer`(비차단 `select`, 차면 버림을 기록 — `normal_relay.go:53-57`) | 로그 한 줄 |
| **critical(a091 뒤 보호 0주)** | `logEvent` 한 줄 → `n.mu` 아래 `Journal.RecordAlert`(임차 없음 · `BEGIN IMMEDIATE` · `busy_timeout` 5s · `synchronous=FULL`) | 로그 한 줄 + **로컬 원장 트랜잭션 하나** + `n.mu` 대기 |

**루프에 남는 몫의 이름(정본 a092 「루프에 남는 몫은 이름을 갖고 편성」)**: a091 이 exit 관측 루프에 더하는 것은 **보호 0주
사건 하나당 outbox 기록 트랜잭션 하나**다(이름: 「0주 기록」). 그 사이클에서 그 포지션은 제출할 것이 없으므로(0주) 그
포지션의 손절을 늦추지 않는다. 같은 사이클의 **뒤쪽 포지션**은 그 트랜잭션만큼 늦는다 — 원격 왕복이 아니라 로컬 쓰기이고
기한은 없다(`n.mu` · 원장 쓰기 — a092 정본이 이미 인정한 성질). 첫 리뷰 C1 의 34s(발송 3회 + 대기 2회)는 base 에 없다.

**기존 비용과의 비교(§0.3 판단의 척도)**: 같은 `applyFloor` 호출이 RECONCILE 에서 이미 하는 브로커 읽기(`ConfirmedFloor` —
Query 2회, 한 Query 최악 ≈ 98s, 번들 calls 표)가 이 로컬 트랜잭션보다 수 자릿수 크다. a091 은 그 읽기를 바꾸지 않는다.

**접힘 · 재알림(tasks 5.1)**: 사건 키는 `exit.stop_sold_nothing|<position id>` 다(끝 경로의 현행 키 모양 `type|position` 을
따른다 — `:1646`). outbox 는 `event_key` 당 행 하나이고 `recordAlertTx`(`outbox.go:276-`)가 판정한다: 행이 PENDING 이면 새
행 없음 · 발송 빚 유지, 정착(전달 · 승인) 행은 재알림 창(`DefaultRemindAfter` **1h**, `notifier.go:59`)이 지나야 재무장.
그래서 8/2 의 13회(3분)는 **행 1 · PENDING 1** 로 접힌다. B2 와 끝 경로는 **같은 키**를 쓴다(같은 사건 — 델타 문단 1) — PENDING 행의 제목 · 본문은 첫 기록의 것으로 남고 재무장 때만 새 기록으로 바뀐다(`outbox.go:295-` — a097). 사이클마다 원인이 바뀌면 그 변화는 구조화 로그가 담는다. 첫 리뷰 H3 의 「2회차부터 `MarkAlertDelivered` 가 PENDING 에 걸려
`alert_undelivered` ERROR 12줄」은 옛 동기 발송(`notifyCritical`)의 모양이다 — base 의 exit 경로는 발송하지 않으므로 그 줄의
발생원이 없다. 5.1 재생이 그것을 잰다.

## D6 — M1 의 해소: 0주 판정은 철자에 의존하지 않는다 (3판 신설)

첫 리뷰 M1 은 「`isZeroQuantity` 는 정확히 `"0"` 문자열 비교」라고 적었다. **base 에서 그 문장은 거짓이다**: `isZeroQuantity`
(`exitloop.go:1871-1878`)는 공백 제거 뒤 빈 문자열이면 0, 아니면 `CompareDecimal(q, "0") <= 0`, 파싱 실패도 0 이다.
`"0.0"` · `" 0"` 도 0 이다. 그래도 「0주 경로는 둘」의 성립에는 입력 쪽 불변식이 필요하므로 생산 출처를 전부 인용한다(번들 근거):

| 값 | 생산 | 0 의 철자 | 번들 |
| --- | --- | --- | --- |
| 포지션 수량 | `Converger.ConvergeQuantities` 가 `NewQuantity: mismatch.Authority()`(= 계좌 값, `converge.go:218` · `compare.go:266`) | 무관 — exit 루프가 `isZeroQuantity(p.Quantity)`(수치) 포지션을 건너뛰고(`exitloop.go:541`), 나머지는 `canonicalSnapshotContext` 가 **양수 강제** + `RatString`(`snapshot.go:263` · `:267`) | convergequantities · canonicalsnapshotcontext |
| 투영 수량 `quantity` | `ProjectWholeShares` → `units.String()`(`snapshot.go:93`) | `"0"` 하나 | projectwholeshares |
| 주문 가능성 | `orderable = projected != "0"`(`snapshot.go:143` 사다리 · `:208` 래칫), `ExecutableProposal` 가 0 이면 빈 제안(`:55`), `record` 의 `orderable`(`exitloop.go:1233`) 이 거짓이면 `submit` 없음(B11 `:1314`) | — | evaluateladdersnapshot · evaluateratchetsnapshot · executableproposal · record |
| 하한 `floor.Quantity` | `ConfirmedFloorQuantity` — `MaxDecimal("0", …)` → `CanonicalDecimal`(`decimal.go:92-101`) 또는 `zeroFloor` 리터럴 `"0"`(`confirmed_floor.go:236-243`) | `"0"` 하나 | confirmedfloorquantity |

따라서 `applyFloor` 에 들어오는 `quantity` 는 **양의 정수**이고, 반환이 0 이 되는 자리는 B2(리터럴) · 끝(`floor.Quantity`)
둘뿐이다(통과 반환 B1 · B3 · B5 는 양의 `quantity` 를 그대로 돌려준다). 첫 리뷰가 올리라던 「수치 비교로 세우는 안」은 base 가
이미 그렇게 서 있다.

## 범위 완결성 — exit 관측 goroutine 의 알림 전수 (3판 재집계, 호출자 단위)

base 의 exit 관측 goroutine 이 올리는 사건을 **호출자 단위**로 세었다(`ExitObserver` 메서드를 가진 파일 전부의 `obs.Event{` 줄,
2판 정정의 교훈 — 파일 단위 열거는 빠뜨린다).

| 자리 | 이벤트 | 등급 | 함수 |
| --- | --- | --- | --- |
| `exitloop.go:866` | `EventExitObservationOutage` | CRITICAL | `checkOutage` |
| `exit_unobserved.go:248` | `EventExitObservationOutage` | CRITICAL | a090 포지션 단위 미관측 |
| **`exitloop.go:1644`** | **`EventExitProposalCapped`** | **normal** | `applyFloor` ← **이 change** |
| `exitloop.go:1714` | `EventExitPositionUnmanaged` | normal | `alertUnmanaged` |
| `exitloop.go:1740` | `EventExitJudgementRefused` | CRITICAL | `alertRefused` |
| `exitloop.go:1764` | `EventExitProposalRefused` | CRITICAL | `alertProposalRefused` |
| `exitloop.go:1794` | `EventExitLiquidationDelayed` | CRITICAL | `noteDelay` |
| `exit_held_proposal.go:203` · `:246` · `:284` | `EventExitLiquidationDelayed` | CRITICAL | a094 판정 진입 알림(`RecordCritical`, 창 0) |
| `exit_held_proposal.go:222` | `EventOrderUnresolved` | CRITICAL | a094 park 원인 |
| `exit_quarantine_announce.go:71` | `EventExitSnapshotQuarantined` | CRITICAL | `announceQuarantine` |

normal 은 둘이고 `EventExitPositionUnmanaged` 는 **승격하지 않는다** — 운영자가 선택한 정상 상태(편입하지 않은 보유)다.
a095 가 그 경계를 정본으로 세웠다: 같은 무관리 보고 중 **편입 시도 실패만** 새 종류로 critical 이고, exit 관측 자리
(`:1714`)는 normal 로 남는다(`event.go:358-359` 주석 · a095 결정 (1)). a091 의 경계도 같다 — 「엔진이 그 포지션을
보호하기로 했는가」. `applyFloor` 의 보호 0주는 엔진이 이미 보호하기로 한 포지션의 보호가 실패한 것이다.

## 건드리지 않는 것

- **제출 수량 계산** — `applyFloor` 의 반환값(B1~B6 · 끝)은 한 글자도 바꾸지 않는다. §0.3 · §0.9 가 걸리는 자리
- **브로커 요청** — `ConfirmedFloor` 의 RECONCILE 읽기(Query 2회)를 포함해 무변경. a091 이 더하는 브로커 요청은 0(§0.4)
- **0주가 되는 원인** — RECONCILE 확정 하한은 대사 영역
- **부분 캡의 등급 · 문구**, **익절 0주의 등급**
- **`exitpolicy` 판정기**, **스키마**(`alert_outbox.event_type` 은 CHECK 없는 `TEXT NOT NULL` — `outbox.go:51`)

## 검증

- 보호 + 0주(두 경로 각각) → 새 종류 · critical, outbox 행 생성
- 보호 + 0주(두 경로 각각) → **로그와 알림이 같은 종류**(H2 — B2 는 `logErr` · `logEvent` 두 줄 다 새 종류)
- 보호 + 부분 캡 → **종전 종류 · 등급 · 문구 무변화**
- 익절 + 0주(두 경로) → **종전 종류 · 등급 무변화**, 끝 경로 문구만 0주에 맞춤
- 미등록 종류는 여전히 normal · 기존 19 종 등급 무변화 · 새 종류는 subject `exit`(class rule 셋 통과)
- **반환값 회귀**: 위 전부에서 `applyFloor` 의 (수량, capped, err) 무변화
- 8/2 재생: 13회(3분)를 fixture 로 돌려 행 1 · PENDING · 재알림 창 판정 확인(D5)
- `go test ./... -count=1 -race` 회귀 0
