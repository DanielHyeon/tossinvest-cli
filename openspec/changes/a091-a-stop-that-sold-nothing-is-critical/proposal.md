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

## Why

2026-08-02, 042660(`pos-522745e0`). 원장 `exit_events`와 `engine.log` 대조.

```text
23:23:25 ~ 23:26:21  STOP_LOSS_LADDER × 13 → 전부 PROPOSAL_REFUSED
                     exit.proposal_capped × 13, severity=normal
                     "the RECONCILE confirmed floor authorises 0
                      (broker sellable quantity), and 5 stays unsold"
23:27:42             ADJUSTMENT_CLOSED — 손절은 끝내 나가지 않았다
```

**손절이 3분 동안 13번 완전히 막혔고, `alert_outbox`에 남은 행은 0건이다.**

`EventExitProposalCapped`가 `criticalEvents`(`obs/event.go`)에 없다. `SeverityOf`는
그 map만 보는 순수 함수이고(`event.go:309-314`, AST branches 1) 미등록은 `SeverityNormal`로
**조용히 강등된다**. normal 등급은 `publishBestEffort`로 가서

- outbox 행이 생기지 않는다 → 원장에 흔적 없음
- 전달 실패해도 재시도가 없다
- 게이트가 반응하지 않는다
- publisher가 nil이면 `notifier.go:139-141`이 **로그도 없이 반환한다**

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

`applyFloor`의 AST는 분기 6·이탈 7이다(`analysis/…/applyfloor/ast.json`).

| 경로 | 조건 | 남는 것 | 제출 수량 |
| --- | --- | --- | --- |
| B2 `:1408→:1414` | 확정 하한을 **계산할 수 없다** | `logErr` 한 줄. **알림 없음** | 0 |
| `:1446` | 확정 하한이 **0을 허용한다** | `EventExitProposalCapped` (normal) | 0 |

둘 다 `submit`의 `isZeroQuantity` 분기(`:1243`)로 가서 조용히 `release`된다.
B2의 fail-closed 방향은 옳다 — 문제는 **그 사실이 보고되지 않는 것**이다.

`isZeroQuantity`(`:1657-1664`)의 모양에 주의(리뷰 M1): **정확히 `"0"` 문자열 비교**이고
`""`·파싱 실패도 0으로 보되 `"0.0"`·`" 0"`은 통과한다. "0주 경로는 둘"의 성립은
`quantity`가 정규형(`"0"`)이라는 전제 위에 있고, 그 불변식의 출처 인용은 freeze
입력이다(`issues.md`).

### 기존 테스트는 이것을 잡을 수 없다

`TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable`(`:953`)과
`TestAFloorThatCannotBeComputedSellsNothing`(`:982`)은 "아무것도 제출되지 않고 레벨은
재발의 가능"까지만 단언한다. **등급·durability·문구는 단언하지 않는다.**
그래서 13회가 반복되는 동안 아무 테스트도 깨지지 않았다.

## What Changes

### 보호 청산이 0주로 깎이면 critical로 보고한다

`applyFloor`가 보호 제안에 대해 **0주**를 돌려주는 두 경로(B2 · `:1446`)에서 critical
등급의 이벤트를 올린다. critical은 durable outbox에 기록되고 전달 실패가 게이트로 이어진다.

**부분 캡은 종전 등급을 유지한다.** 일부라도 나갔으면 그것은 "보호되지 않은 노출"이
아니라 축소된 노출이다.

**익절의 0주 캡도 종전 등급을 유지한다.** 익절이 안 나가도 노출은 그대로다.

### 문구를 결과에 맞춘다

0주일 때는 "일부만 나갔다"가 아니라 **한 주도 나가지 않았다**고 말한다.

### 보호/익절 구분은 호출자가 넘긴다

`applyFloor`는 제안이 손절인지 익절인지 모른다(FLM 입력 표). `submit`은 `proposal`을
갖고 있으므로(`:1237` 시그니처) 전달할 수 있다.

## Impact

- **Specs**: `engine-safety` (**MODIFIED 1** — 리뷰 M4. 「등급화된 알림」의 critical
  열거에 사건 하나를 더한다. a092가 같은 요구를 MODIFIED로 싣는 선행이므로 재고정 시
  a092-뒤 정본 위로 재기저화한다 — delta 머리와 tasks 0.5)
- **Code**: `internal/app/engine/exitloop.go` (`applyFloor` 알림 경로 · `submit`의 인자
  전달 · B2 `logErr`의 종류 — H2), `internal/obs/event.go` (새 종류 등록 + 종류 목록
  주석 갱신)
- **Tests** (리뷰 M5): `internal/obs/a091_*_test.go` **신규** — 새 종류의 `criticalEvents`
  등록 누락을 잡는 유일한 장치, `internal/app/engine/exitloop_test.go` (기존 0주 시험 둘
  옆에 등급·종류·문구 단언 추가)
- **Docs** (리뷰 M5): `issues.md`(소비자 조사·낡음 대장), `docs/pm/generated/` 3종,
  `cmd/tossctl/engine_assembly.go:31-35`의 Publisher 주석 — nil transport 서술이
  현행 배선과 맞는지 확인·갱신
- **Schema**: **없음**
- **§0.3**: 손절을 지연시키지 않는다. **제출 수량 계산(B1~B6·`:1446`의 반환값)을
  건드리지 않는다** — 바꾸는 것은 보고뿐이다
- **§0.4**: 브로커 요청 무변경 (`applyFloor`는 브로커에 닿지 않는다 — FLM calls 표)
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

## 미해결 → 처분 (2026-09-30 재작성)

- **새 이벤트 종류 vs 등급 분기** — **해소: design D1이 B(신설)로 결정했다.** 진짜
  근거는 리뷰 H1이 세운 class rule(D1 표의 C안 행 참조)
- **§0.3 — 승격이 만드는 동기 지연 (C1).** 리뷰 시점(base `ec29dc72`) HEAD에서 normal은
  `publishBestEffort` publish 1회(상한 10s), critical은 `deliver` 최대 3회 + 대기 2회
  (**34s**, `n.mu` 보유)였고 `applyFloor`는 `ObserveOnce`(순차 순회) 안에서 불린다.
  **a092가 이 성질 자체를 제거한다**(critical은 기록까지만 동기 — design D5). a092는
  freeze 종료(24라운드)·기록 입구 착지(`Journal.RecordAlert`) 상태로 구현 중이다.
  **발효 조건은 a092 완주(아카이브)이며, 그 뒤 base 재고정 + freeze 재리뷰가 이
  change의 다음 단계다**
