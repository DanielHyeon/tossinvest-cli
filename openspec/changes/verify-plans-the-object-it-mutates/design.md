# Design: verify-plans-the-object-it-mutates

## D1. 종목 해석은 한 곳에 남고, 카탈로그가 자기 대상을 선언한다

`mutationSymbol`은 이미 "이 단계가 무엇에 대해 주문을 내는가"를 답하는 유일한 자리다.
계획(`plan.go:564`)과 preflight의 시장 검사(`runner.go:550`)가 둘 다 여기를 읽는다.
고칠 자리는 여기다. 호출부에 종목 결정을 흩으면 두 번째 권위가 생기고, 이 change가
고치는 결함이 바로 그 종류다.

카탈로그는 이미 자기 mutation을 데이터로 들고 있다(`Step.Mutations`). 대상도 같은 방식으로
선언한다:

```go
// ActsOnConditional says the step operates on the conditional order this
// verification already registered, not on the run's probe symbol.
ActsOnConditional bool `json:"acts_on_conditional,omitempty"`
```

해석 순서:

| 조건 | 종목 |
|---|---|
| `ActsOnConditional`이고 살아 있는 조건주문이 있다 | 그 조건주문의 종목 |
| `ActsOnConditional`인데 아직 없다 | `holdingSymbol` — 등록 단계가 쓸 종목이 곧 그것이다 |
| `NeedsHolding` | `holdingSymbol` (무변경) |
| 그 밖 | `r.symbol` (무변경) |

두 번째 줄이 신선한 전체 실행을 덮는다: 계획 시점에는 조건주문이 아직 없고,
`conditional-register`가 `holdingSymbol`에 등록한 뒤 실행 시점의 `liveConditional()`이
같은 종목을 돌려주므로 계획과 실행이 일치한다.

### 남는 좁은 구간과 그 실패 방향

이전 실행이 종목 A에 남긴 조건주문이 정리 대상이면서, 같은 실행이 다른 보유 B에 대해
`conditional-register`를 다시 도는 경우 — 계획은 A를 싣고 실행 시점에는 A가 정리되고 B가
살아 있다. 결과는 `ErrOutsidePlan`: **아무것도 전송되지 않고 실행이 멈춘다.** 다음 실행에서는
A가 기록의 outstanding 집합에서 빠져 있으므로 계획이 B를 싣고 정상 진행한다. 스스로 복구되며,
틀린 방향으로 보내는 것이 아니라 거절하는 방향으로 틀린다. 이 구간을 계획 단계에서 미리
풀려면 `Plan`의 `willRun`을 `mutationSymbol`까지 끌고 들어가야 하는데, 그것은 preflight의
run-time 호출부와 의미가 갈라진다 — 지금 고치는 결함과 같은 종류의 분기다. 만들지 않는다.

## D2. `NeedsHolding: true`를 붙이지 않는 이유

가장 짧은 수정처럼 보이지만 틀린다. `NeedsHolding`은 `preflightStatic`(runner.go:545)에서
"계좌에 쓸 수 있는 보유가 없으면 이 단계를 건너뛴다"까지 뜻한다. 그러면 **보유가 사라진
계좌에 조건주문만 남은 상태** — 잔여물을 반드시 치워야 하는 바로 그 상태 — 에서
`conditional-cancel`이 skip된다. 지금 계좌에는 보유가 있어 당장 드러나지 않지만, 이 도구가
남긴 것을 이 도구가 치우지 못하게 되는 구성을 만들 이유가 없다.

또 하나: 잔여 조건주문이 현재 `holdingSymbol`과 **다른** 종목일 수 있다.
`firstUsableHoldingIn`은 "이 시장의 첫 쓸 만한 보유"를 고르고, 그것이 조건주문이 걸린 종목과
같다는 보장은 없다. `NeedsHolding`은 그 경우에도 여전히 틀린 종목을 싣는다.

## D3. 이름 붙일 수 없는 대상은 목록에 오르지 않는다

`plan.go`의 자기 서술이 이미 규칙을 말하고 있다 — "an unreadable quantity is not something to
ask a person to approve"(`planUnknown`). 종목도 같다. mutating 단계의 종목이 빈 문자열이면
그 단계는 계획에서 빠지고 사유가 표시된다.

오늘 이것이 닿을 수 있는 곳은 US 실행이다. `MarketOf("")`는 `"US"`를 돌려주므로
(`pricing.go:258` — 6자리 숫자가 아니면 US), 보유가 없는 US 계좌에서 `r.symbol`이 비면
preflight의 시장 검사가 그것을 잡지 못하고 **종목 없는 계획 줄**이 생긴다. KR 실행에서는
같은 값이 시장 불일치로 먼저 걸린다. 이 비대칭은 우연이며, 승인 목록의 완전성이 걸린 자리에
남겨 둘 성질이 아니다.

## D4. 인가에서 와일드카드를 없앤다

```go
// 전
if m.Symbol != "" && !strings.EqualFold(m.Symbol, strings.TrimSpace(symbol)) { continue }
// 후
if !strings.EqualFold(strings.TrimSpace(m.Symbol), strings.TrimSpace(symbol)) { continue }
```

- 종목이 있는 계획 줄: 동작 동일.
- 종목이 없는 계획 줄: 종목이 **있는** 요청을 더 이상 인가하지 않는다.
- 정리 줄(`cleanup.go:188`)은 artifact의 종목을 싣고 `cancelOrder`/`cancelConditional`에
  같은 값을 넘긴다. 둘 다 비어 있는 옛 기록에서도 `"" == ""`로 계속 인가된다.

D3이 계획 쪽에서 막고 D4가 인가 쪽에서 막는다. 둘 중 하나만으로도 오늘의 결함은 고쳐지지
않으므로 둘 다 이 change의 것이 아니라고 볼 수도 있으나, 이 change가 세우는 성질은
"**계획의 각 줄은 그 요청이 무엇에 대한 것인지 말한다**"이고 와일드카드는 그 성질의 부정이다.
같은 문장 안에 있다.

## D5. 무엇을 테스트가 고정하는가

결함의 종류는 "계획이 말한 것과 실행이 보내는 것이 다르다"이다. 종목 상수 하나를 비교하는
테스트는 같은 결함이 다른 필드에서 재발하는 것을 못 잡는다. 그래서 고정하는 것은 **경로
전체**다: probe 종목과 다른 종목에 조건주문이 살아 있는 기록으로 runner를 만들고,
`conditional-modify`·`conditional-cancel`이 실제로 인가를 통과해 브로커에 도달하는지를
fake broker로 확인한다. 이 테스트는 오늘의 기록 모양을 그대로 쓴다 —
probe `005930`, 조건주문은 `333430`.

보조로 두 단위 성질을 고정한다: ① `ActsOnConditional` 단계의 계획 줄 종목이 살아 있는
조건주문의 종목이라는 것, ② 종목 없는 계획 줄이 종목 있는 요청을 인가하지 않는다는 것.

## D6. 범위 밖

- KR 정규장 밖에서 정정·취소가 통과하는지 — 실측 대상이고 이 change의 주장이 아니다.
  22:43 KST에 `conditional-register`가 통과한 것은 관측된 사실이지만 정정·취소가 같다는
  근거가 아니다.
- 콘솔에 종목을 넘기는 입구 신설 — 만들지 않는다. 계획이 스스로 옳은 종목을 고르는 것이
  이 change의 답이고, 사람이 종목을 손으로 맞춰 넣게 하는 것은 같은 결함을 운영 절차로
  옮기는 것이다.
