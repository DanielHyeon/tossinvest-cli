# a091 · 한 주도 못 판 손절은 critical이다

- **Feature**: `FEAT-TOS-009` — Exit line truth and position policy lifecycle
- **Story**: `STORY-TOS-a091`
- **Spec**: `engine-safety`
- **위험 등급**: **High-risk** (손절 경로의 알림 등급. §0.3 적용.)

> **작성 순서**: 이 문서의 분기 주장은 전부 `analysis/function-logic/`의 AST 산출물에서
> 나왔다. 산출물이 문서보다 **먼저** 만들어졌다 (`.claude/CLAUDE.md`「단계 건너뛰기 금지」).

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

### 0주가 되는 경로는 둘이고 하나는 알림조차 없다

`applyFloor`의 AST는 분기 6·이탈 7이다(`analysis/…/applyfloor/ast.json`).

| 경로 | 조건 | 남는 것 | 제출 수량 |
| --- | --- | --- | --- |
| B2 `:1408→:1414` | 확정 하한을 **계산할 수 없다** | `logErr` 한 줄. **알림 없음** | 0 |
| `:1446` | 확정 하한이 **0을 허용한다** | `EventExitProposalCapped` (normal) | 0 |

둘 다 `submit`의 `isZeroQuantity` 분기(`:1243`)로 가서 조용히 `release`된다.
B2의 fail-closed 방향은 옳다 — 문제는 **그 사실이 보고되지 않는 것**이다.

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

- **Specs**: `engine-safety` (ADDED 1)
- **Code**: `internal/app/engine/exitloop.go` (`applyFloor` 알림 경로 · `submit`의 인자 전달),
  `internal/obs/event.go` (새 이벤트 종류를 `criticalEvents`에 등록하는 경우)
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
- **outbox 재발 장부** → a089
- **보호 청산의 가격** → a087

## 미해결

- **새 이벤트 종류를 만들 것인가, 기존 종류에 등급 분기를 둘 것인가.** `SeverityOf`는
  종류만 보므로 등급을 나누려면 종류를 나눠야 한다. 이벤트 종류 추가는
  `AllReasonCodes` 계열의 rename 규칙과 무관하지만 소비자(콘솔·로그 필터)를 확인해야 한다.
  design D1에서 결정한다
- **§0.3 — 승격이 만드는 동기 지연.** normal은 `publishBestEffort`로 publish 1회(상한
  10s), critical은 `deliver`로 최대 3회 + 대기 2회(**34s**, `n.mu` 보유)다.
  `applyFloor`는 `ObserveOnce`(순차 순회) 안에서 불리므로 승격은 그 경로의 동기 체류를
  10s에서 34s로 늘린다. **이것은 a092가 해소한 뒤에 발효해야 한다** —
  근거는 `openspec/changes/a092-an-alert-does-not-hold-the-stop/`의 AST 산출물
