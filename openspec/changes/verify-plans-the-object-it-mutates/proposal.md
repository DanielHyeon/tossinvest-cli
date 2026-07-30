# Change: verify-plans-the-object-it-mutates

## Why

2026-07-29 22:28 KST, 사용자가 콘솔에서 `[재측정]`을 두 번 눌렀고 두 번 다 같은 자리에서
멈췄다. 계좌에는 아무것도 전송되지 않았다.

```
13:28:16Z  approval             pass   승인 2건: conditional-modify, conditional-cancel
13:28:20Z  conditional-modify   fail   ErrOutsidePlan
13:28:54Z  approval             pass   (같은 digest sha256:c2a06fd2…)
13:28:57Z  conditional-modify   fail   ErrOutsidePlan
```

사유 원문:

```
conditional-modify is about to modify-conditional for SELL 1 333430,
which is not on the list approved at the start of this run
```

승인된 줄과 실제 요청을 필드별로 대조하면 어긋난 것은 **종목 하나**다.

| 필드 | 계획 | 요청 | |
|---|---|---|---|
| kind | modify-conditional | modify-conditional | 일치 |
| side | sell | SELL | 일치 |
| quantity | max 1 | 1 | 일치 |
| **symbol** | **005930** | **333430** | **불일치** |

### 결함 — 계획과 실행이 같은 객체를 가리키지 않는다

- 계획: `plan.go:564` → `runner.go:557` `mutationSymbol`. `step.NeedsHolding`이 아니면
  `r.symbol`을 쓴다. 콘솔의 `r.symbol`은 `cmd/tossctl/console.go:73`
  `consoleProbeSymbolKR = "005930"` — **매수측 probe 종목**이다.
- 실행: `steps.go:723`·`steps.go:775`. 두 단계 모두 `liveConditional()`이 돌려주는
  **살아 있는 조건주문의 종목**(`333430`)으로 요청을 만든다.
- `plan.go:398` `Authorises`는 종목을 정확히 비교한다 → 불일치 → `ErrOutsidePlan`.

`NeedsHolding: true`는 `sell-boundary`(verifylive.go:323)와
`conditional-register`(verifylive.go:369)에만 있다. **`conditional-modify`와
`conditional-cancel`에는 없다.** 그런데 `--holding-symbol` 플래그 설명(`verify.go:184`)은
"Held symbol the sell-side **and conditional steps** use"라고 적혀 있다 — 의도는 문서에
있었고 카탈로그가 그것을 따르지 않았다.

정리 prologue는 같은 자리를 제대로 짚는다: `cleanup.go:188`은 `Symbol: a.Symbol`,
즉 **정리 대상 객체 자신의 종목**을 싣는다.

### 왜 이제 드러났나

이 결함은 `verify-reopens-conditional-chain`이 만든 것이 아니라 **그 change가 처음으로
도달 가능하게 만든 잠복 결함**이다. `conditional-modify`·`conditional-cancel`은 지금까지
모든 실행에서 "이 검증이 만든 살아 있는 조건주문이 없다"로 skip돼 인가 검사에 닿은 적이
없었다. 교착이 풀리자마자 첫 도달에서 드러났다.

### 지금의 귀결

- `conditional-cancel`도 같은 불일치다. 즉 **정상 경로로는 그 조건주문을 제거할 수 없다.**
- 정리 prologue도 이것을 건드리지 않는다 — `conditional-cancel`의 판정(07-26)이
  이 객체(07-29)보다 먼저 기록됐고, `decidedAfter`가 그것을 올바르게 거절한다.
- 조건주문 `p7hQz7HAXc…` (333430)이 계좌에 살아 있다. MARKET 손절, 발동가 최종체결가
  20% 아래, 1주, 만료 1주.
- `verify-execution-capability` task 2.5는 `conditional-modify`·`conditional-cancel`이
  남아 미완이고, 그래서 2c `add-protection-orders`는 계속 금지 상태다(tasks.md:3).

게이트는 **의도대로 동작했다** — 아무것도 전송되지 않았고 실행이 멈췄다. 고칠 것은
게이트가 아니라 게이트에 주어지는 목록이다.

## What Changes

- **승인 계획은 단계가 실제로 지목할 객체를 이름한다.** 등록된 조건주문 위에서 동작하는
  단계(`conditional-modify`, `conditional-cancel`)는 계획에서도 그 조건주문의 종목을
  싣는다. 카탈로그가 그 사실을 데이터로 선언하고(`Step.ActsOnConditional`), 종목 해석이
  한 곳(`mutationSymbol`)에 남는다.
- **이름 붙일 수 없는 대상은 목록에 오르지 않는다.** mutating 단계의 종목이 비면 그 단계는
  계획에서 제외되고 사유가 표시된다. 모르는 채로 승인을 받지 않는다는 기존 방향
  (`planUnknown`)과 같다.
- **인가는 빈 종목을 와일드카드로 쓰지 않는다.** `Authorises`의 종목 비교에서 "계획 줄에
  종목이 없으면 무엇이든 통과" 분기를 없앤다. 종목이 있는 줄의 동작은 그대로이고,
  종목이 없는 줄이 종목 있는 요청을 인가하던 경로만 닫힌다 — 좁히는 방향이다(§0.9).

## Non-Goals

- 인가 규칙 완화 — 없다. 이 change는 **목록을 정확하게** 만들 뿐 게이트를 느슨하게 하지
  않는다. 목록 밖 요청이 실행을 세우는 레일은 무변경이다.
- `NeedsHolding: true`를 조건주문 정정·취소에 붙이기 — 거절한다. 그 플래그는 preflight
  까지 막으므로, 보유가 사라진 계좌에서 **잔여 조건주문을 취소해야 할 바로 그 순간에**
  cancel이 skip된다. design.md D2.
- 승인 창(5분)·노출 상한·1주 규칙·수량 상한 변경 — 무변경.
- 조건주문 발동(`conditional-trigger`) — 그대로 deferred다.
- 2c `add-protection-orders` 스펙 작성 — 금지 유지.
- KR 정규장 밖 정정·취소 동작에 대한 결론 — 이 change는 그것을 **측정 가능하게** 만들 뿐
  결과를 미리 주장하지 않는다.

## Capabilities

### Modified Capabilities

- `order-execution`: 승인 계획의 각 줄이 그 단계가 실제로 지목할 객체를 이름한다

## Impact

- Affected code: `internal/verifylive/verifylive.go`(`Step.ActsOnConditional` +
  두 단계 선언), `internal/verifylive/runner.go`(`mutationSymbol`),
  `internal/verifylive/plan.go`(`Plan` 제외 경로, `Authorises` 종목 비교)
- High-risk 여부: **yes** — 라이브 주문 정정·취소의 승인 인가 표면이다. 적대적 Eng 리뷰 +
  Pre-Edit 선언 + Function Logic Map 필수.
- 안전 검토(§0):
  - §0.1: 새 mutation 종류 0. 이 change가 보낼 수 있게 하는 요청은 **이미 승인 화면에
    표시되고 있던 두 줄**이다(`approval.requests_listed = 2`). 사람의 승인 행위 하나가
    여전히 모든 요청 앞에 있다.
  - §0.3: 손절·비상 청산 경로 무변경. 대상 조건주문은 검증이 만든 것이고 발동가가 시장
    20% 아래다.
  - §0.4: 새 호출 종류 0, 호출 수 증가 0. 오히려 실패로 끝나던 재시도가 줄어든다.
  - §0.6: `Step.ActsOnConditional`은 가산·`omitempty`. 기록 스키마(`Entry`) 무변경.
  - §0.9: 세 변경 모두 **좁히거나 정확하게 하는** 방향이다. 인가가 넓어지는 변경은 없다 —
    `Authorises`의 와일드카드 제거는 순수한 축소이고, 계획 줄이 정확해지는 것은 인가
    범위를 옮기는 것이지 늘리는 것이 아니다.
