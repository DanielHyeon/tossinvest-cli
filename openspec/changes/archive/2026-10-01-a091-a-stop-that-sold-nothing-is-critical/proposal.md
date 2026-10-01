# a091 · 한 주도 못 판 손절은 critical이다

- **Feature**: `FEAT-TOS-009` — Exit line truth and position policy lifecycle
- **Story**: `STORY-TOS-a091`
- **Spec**: `engine-safety`
- **위험 등급**: **High-risk** (손절 경로의 알림 등급. §0.3 적용.)

> **작성 순서**: 이 문서의 분기 주장은 전부 `analysis/function-logic/`의 AST 산출물에서
> 나왔다. 산출물이 문서보다 **먼저** 만들어졌다 (`.claude/CLAUDE.md`「단계 건너뛰기 금지」).
>
> **판 이력**: 2026-08-06 proposal-freeze **거부**(review.md — 차단 C1·C2). C2(동기의
> 실측 과장)는 현행 Why에 정정 반영됨. 2026-09-30 재작성(Manager): H1·H2·M1·M2·M4·M5
> 반영 + 리뷰 이후의 세계 변화 반영 — a089 불구현 아카이브(2026-09-28 사용자 결정),
> a092 신설·freeze 종료·기록 입구 착지(design D5). **발효 조건은 C1 그대로 — a092 완주
> 뒤 base 재고정 + freeze 재리뷰.** 본문의 코드 좌표는 base `ec29dc72` 시점이며 재고정
> 시 재검증한다.
>
> **3판(2026-10-01, 재freeze 입력)**: C1 발효(a092 아카이브 `75d138b5`) → base 재고정 `b30318d6`(`16cb1a1a`) →
> 좌표 · 수치를 base 에서 다시 쟀다(FLM 번들 14 — `analysis/function-logic/`). 바뀐 판단: M1 해소(`isZeroQuantity` 는
> base 에서 수치 비교 — design D6), §0.4 문장 정정(`applyFloor` 는 RECONCILE 에서 브로커를 읽는다 — a091 이 더하는 요청 0),
> 새 종류 이름 `exit.stop_sold_nothing`(design D1), 루프에 남는 몫의 이름(design D5). 아래 `ec29dc72` 좌표는 3판에서 base 좌표로 바꿨다.

## Why

2026-08-02, 042660(`pos-522745e0`). 원장 `exit_events`와 `engine.log` 대조.

```text
23:23:25 ~ 23:26:21  STOP_LOSS_LADDER × 13 → 전부 PROPOSAL_REFUSED
                     exit.proposal_capped × 13, severity=normal
                     "the RECONCILE confirmed floor authorises 0
                      (broker sellable quantity), and 5 stays unsold"
23:27:42             ADJUSTMENT_CLOSED — 손절은 끝내 나가지 않았다
```

**4판 — 원장 재독(2026-10-01, 읽기 전용, design 「8/2 원장 재독」)**: 13회 동안 계좌 보유는 **5**였고 한정 항은 **매도가능 0**이었다.
같은 4분 사이 보유가 **엔진 밖 매도로** 10 → 5(23:17) → 2(23:26:39) → 0(23:27:41) 으로 줄었다 — 운영자가 손으로 내던 매도 주문이
남은 주식을 잡고 있었다. 포지션은 엔진 밖에서 종결됐고(`exit.position_closed_externally`), 엔진이 편입해 보호하던 그 포지션의 손절은
그동안 한 번도 나가지 못했다. 이 사실은 a091 뒤에도 보고 대상이다(보유 0 이 아니다 — design D3).

**손절이 3분 동안 13번 완전히 막혔고, `alert_outbox`에 남은 행은 0건이다.**

`EventExitProposalCapped`가 `criticalEvents`(`obs/event.go`)에 없다. `SeverityOf`는
그 map만 보는 순수 함수이고(base `event.go:371-376`, AST branches 1) 미등록은 `SeverityNormal`로
**조용히 강등된다**. 8/2 당시 normal 등급은 `publishBestEffort`로 가서(base 에서는 a092 의 이관 버퍼 `NormalRelay` 로 간다 —
outbox 행이 생기지 않는 것은 같다)

- outbox 행이 생기지 않는다 → 원장에 흔적 없음
- 전달 실패해도 재시도가 없다
- 게이트가 반응하지 않는다
- (8/2 당시) publisher 가 nil 이면 normal 은 로그도 없이 반환했다. base 에서는 이관 버퍼가 버림을 로그로 기록한다(`normal_relay.go:43-58`)

**8/2에 운영자가 한 번도 호출받지 못한 이유는 등급이 아니라 transport 부재다** —
알림 배선 커밋 `e540668f`는 2026-08-04이고, `alert_outbox` id 1~9(7/31~08-04)는 전부
`attempts=0`(= `notifier.go:252` `Publisher == nil` 분기)이며, `engine.log`에
`"no notification publisher is configured"`가 8/1과 8/3 **양쪽에** 있다.
등급을 올렸어도 그날의 호출 횟수는 0이었다.

**그래서 이 change의 이익은 운영자 호출이 아니라 원장에 남는 흔적 하나다.**
normal은 outbox 행을 만들지 않으므로 8/2의 13회는 **사후에 재구성할 방법이 없었다**.
그 이익만으로 change는 성립한다.

### 문구도 사실과 다르다

```go
Title: "… 청산이 확정 하한에 걸려 일부만 나갔다"
```

`floor.Quantity == 0`이면 나간 수량은 **0**이다. "일부만 나갔다"는 반대를 말한다.

단, 이것은 8/2의 증거가 아니다(리뷰 M2) — 8/2 로그 본문은 영어였고 이 한국어 Title은
8/2 **이후**에 들어왔다. 운영자에게 전달된 적 없는, 현행 코드의 정확성 결함이다.

### 0주가 되는 경로는 둘이고 하나는 알림조차 없다

`applyFloor`의 AST는 분기 6·반환 7이다(`analysis/…/applyfloor/ast.json`, base `exitloop.go:1617-1661`).

| 경로 | 조건 | 남는 것 | 제출 수량 |
| --- | --- | --- | --- |
| B2 `:1622→:1628` | 확정 하한을 **계산할 수 없다** | `logErr` 한 줄. **알림 없음** | 0 |
| 끝 `:1644→:1660` | 확정 하한이 **0을 허용한다** | `EventExitProposalCapped` (normal) | 0 |

둘 다 `submit`의 `isZeroQuantity` 분기(`:1401`)로 가서 조용히 `release`된다.
B2의 fail-closed 방향은 옳다 — 문제는 **그 사실이 보고되지 않는 것**이다.

`isZeroQuantity`의 모양(리뷰 M1): 첫 리뷰 시점에는 정확히 `"0"` 문자열 비교였으나 **base 에서는 수치 비교**다
(`exitloop.go:1871-1878` — `CompareDecimal(q, "0") <= 0`, 빈 문자열 · 파싱 실패도 0). "0주 경로는 둘"의 입력 쪽 불변식
(`quantity` 는 양의 정수, `floor.Quantity` 의 0 은 `"0"` 한 철자)은 생산 출처까지 인용했다 — design D6.

### 기존 테스트는 이것을 잡을 수 없다

`TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable`(base `exitloop_test.go:971`)과
`TestAFloorThatCannotBeComputedSellsNothing`(`:1000`)은 "아무것도 제출되지 않고 레벨은
재발의 가능"까지만 단언한다. **등급·durability·문구는 단언하지 않는다.**
그래서 13회가 반복되는 동안 아무 테스트도 깨지지 않았다.

## What Changes

### 보호 청산이 0주로 깎이면 critical로 보고한다

`applyFloor`가 보호 제안에 대해 **0주**를 돌려주는 두 경로(B2 `:1622` · 끝 `:1644`)에서, **알림이 켜진 엔진이면** 새 종류
`exit.stop_sold_nothing`(critical)으로 보고한다. critical은 durable outbox에 기록되고 전달 실패가 게이트로 이어진다.

**알림이 꺼진 엔진(기본값)은 종전 등급 그대로다** — 정본 a095 「무관리 보유 보고의 등급은 사실이 정한다」의 규칙을 문자 그대로 따른다
(보낼 수 없는 critical 이 진입 차단 · ENTRY_BLOCKED 로 가지 않게 — 불변식 3). 알림이 켜져 있는데 전송이 실패하는 엔진의 래치는
의도된 동작이다(design D1).

**계좌 보유 0**(엔진 밖 종결 진행 중)이 원인인 0주와 **관측을 끝내는 취소**로 하한 조회가 중단된 경우도 종전 등급이다(design D3).

**부분 캡은 종전 등급을 유지한다.** 일부라도 나갔으면 그것은 "보호되지 않은 노출"이
아니라 축소된 노출이다.

**익절의 0주 캡도 종전 등급을 유지한다.** 익절이 안 나가도 노출은 그대로다.

### 문구를 결과에 맞춘다

0주일 때는 "일부만 나갔다"가 아니라 **한 주도 나가지 않았다**고 말한다. 알림 행 본문은 에피소드의 **첫 원인과 시각**을 담고,
관측마다의 원인은 구조화 로그 줄이 담는다(design D7). 하한 계산 실패의 원문 오류는 계좌를 가린 로그에만 남긴다(design D8).

### 보호/익절 구분은 호출자가 넘긴다

`applyFloor`는 제안이 손절인지 익절인지 모른다(FLM 입력 표). `submit`은 `proposal`을
갖고 있으므로(`:1395` 시그니처) 전달할 수 있다.

## Impact

- **Specs**: `engine-safety` (**MODIFIED 1** — 리뷰 M4. 「등급화된 알림」의 critical
  열거에 사건 하나를 더한다. a092-뒤 정본 위로 재기저화 — 3판), `exit-policy` (**MODIFIED 1** — 4판, R2-2.
  「관측 경로와 fail-safe」의 「캡 발생은 알림된다(… 일반 등급 …)」 뒤에 보호 0주 예외 한 문장)
- **Code**: `internal/app/engine/exitloop.go` (`applyFloor` 원인 분류 · 게이트 · 알림 경로 · B2 오류 줄(계좌 가림) · `submit`의 인자
  전달 · `ExitObserverOptions.NotificationsEnabled`), `internal/app/engine/exitwiring.go` (설정 값으로 덮기),
  `internal/obs/event.go` (새 종류 등록 + 종류 목록 주석)
- **Tests** (리뷰 M5 · 4판): `internal/obs/a091_*_test.go` **신규**(등록 · 값 핀 · class rule), `internal/app/engine/a091_*_test.go` **신규**
  (실제 `RecordOnly` + 원장 + 로그 캡처 하네스 · 5 액션 표 · 게이트 두 팔 · 보유 0 · 종료 취소 · 계좌 카나리 · 원인 두 순서 ·
  8/2 재생 네 팔 + 배달 실행자 내보내기 훅), `internal/riskcalc` 보유 0 동치 표 시험, `exitloop_test.go` 기존 0주 시험 둘 확장(레벨 해제 ·
  재발의). 깨지는 기존 시험 0(측정 — 2라운드 보이스 B)
- **Docs** (리뷰 M5 · 4판): `issues.md`(소비자 조사 · 낡음 대장), `docs/pm/generated/` 3종, **`docs/operations.md` 새 절**(`exit.stop_sold_nothing` —
  확인 · 조치 · 승인 `tossctl engine alerts ack` · 모드 해제 `tossctl engine mode-release`). `cmd/tossctl/engine_assembly.go:31-35`의
  Publisher 주석은 이미 참이다(2라운드 보이스 B — 갱신 불요)
- **Schema**: **없음**
- **§0.3**: 제출 수량 계산(B1~B6 · 끝의 반환값)을 **건드리지 않는다**. 루프에 남는 몫은 이름 둘로 편성한다 — 「0주 기록」(로그 줄 · `n.mu`
  대기 · 기록 트랜잭션) · 「0주 기록 실패 승격」(기록 실패일 때만 동기 승격 트랜잭션). 둘 다 원격 0 · 기한 없음, 같은 사이클 뒤쪽 포지션이
  그만큼 늦게 판정된다. 소요는 미실측 — 구현 로트가 잰다(design D5, tasks 5.3)
- **§0.4**: 브로커 요청 무변경. **3판 정정**: 2판의 「`applyFloor`는 브로커에 닿지 않는다」는 거짓이었다 — RECONCILE 에서
  `ConfirmedFloor` 가 `Retrier.Query` 2회(Holdings · SellableQuantity)로 브로커를 읽는다(번들 calls 표 — 최악 수치 포함).
  a091 은 그 읽기를 바꾸지 않고 새 요청을 더하지 않는다(더하는 것은 로컬 원장 쓰기 — design D5)
- **§0.9**: 임계·가격·수량 무변경

## Non-goals

- **0주가 되는 원인 자체** — RECONCILE 확정 하한이 0을 주는 것은 대사 영역이다.
  이 change는 그것이 **보이게** 만들 뿐 고치지 않는다
- **부분 캡의 등급** — 현행 유지
- **`EventExitProposalCapped`를 통째로 `criticalEvents`에 넣기** — 부분 캡까지 critical이
  된다. `SeverityOf` FLM의 Safety conclusion 참조
- **관측 누락 계측** → a090
- **재발·접힘·재알림 의미론** → a092 재알림 창(`remindAfter`)·episode(`event_key`)가
  소유한다. 종전 판이 여기 두었던 "outbox 재발 장부 → a089"는 낡았다 — a089는 불구현
  아카이브됐다(2026-09-28 사용자 결정)
- **보호 청산의 가격** → a087

## 미해결 → 처분 (2026-09-30 재작성 · 2026-10-01 4판)

- **새 이벤트 종류 vs 등급 분기** — **해소: design D1이 B(신설)로 결정했다.** 진짜
  근거는 리뷰 H1이 세운 class rule(D1 표의 C안 행 참조)
- **§0.3 — 승격이 만드는 동기 지연 (C1).** 리뷰 시점(base `ec29dc72`) HEAD에서 normal은
  `publishBestEffort` publish 1회(상한 10s), critical은 `deliver` 최대 3회 + 대기 2회
  (**34s**, `n.mu` 보유)였고 `applyFloor`는 `ObserveOnce`(순차 순회) 안에서 불린다.
  **a092가 이 성질 자체를 제거했다**(critical은 기록까지만 동기 — design D5). **C1 발효**: a092 아카이브 `75d138b5`
  (2026-09-30) → base 재고정 `b30318d6` → freeze 재리뷰(3판 REJECT → 4판, tasks 0.5). 루프에 남는 몫은 design D5 의 이름 둘
- **알림 꺼진 엔진의 등급 (2라운드 R2-1 → Manager Q1, 2026-10-01)** — 해소: a095 정본 게이트(enabled 플래그만). 사용자 거부권 항목으로 Manager 가 보고
- **보유 0 원인의 0주 (R2-5 → Q2)** — 해소: 배제(종전 등급 · 본문에 원인). 8/2 는 보유 0 이 아니었다(원장 재독 — 매도가능 0)
- **원인 계약 (R2-6 → Q3)** — 해소: 에피소드 단일 키 유지, 행 본문 = 첫 원인 + 시각, 관측마다의 원인 = 로그(design D7)
