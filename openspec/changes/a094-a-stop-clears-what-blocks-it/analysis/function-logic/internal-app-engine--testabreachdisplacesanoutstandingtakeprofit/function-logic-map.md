# Function Logic Map: `TestABreachDisplacesAnOutstandingTakeProfit`

- Source: `internal/app/engine/exitloop_test.go` (`842`–`874`)
- Qualified: `TestABreachDisplacesAnOutstandingTakeProfit`
- AST evidence: `ast.json` (`source_sha256` f09993a5df74088e…)
- Risk scan: `risk-pattern-report.md`

**역할.** 기존 시험 — a094 D−4.3 로 손절이 취소한 주기 다음 주기에 나감(+1 관측). 이름 붙은 대가(design.md:403 · :311).

## Inputs and invariants

시험 픽스처(임시 원장). 생산 코드 아님 — 비례 원칙상 이 번들은 게이트의 요구 집합(수정된 기존 함수)을 채우는 기록이다.

## Branches and early returns

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:850` `if cycle.Err != nil {` | — |
| B2 | if | `:853` `if len(h.submit.cancels) == 0 {` | — |
| B3 | if | `:858` `if len(h.submit.places) != 1 {` | — |
| B4 | if | `:861` `if cycle := h.observe(); cycle.Err != nil {` | — |
| B5 | if | `:864` `if len(h.submit.places) != 2 {` | — |
| B6 | if | `:868` `if state.PendingAction != string(exitpolicy.ActionBaselineBreach) {` | — |
| B7 | if | `:871` `if state.PendingIntentID != h.submit.places[1].IntentID {` | — |

## Calls and live bindings

시험 대상 API(원장 · 관측 루프)만 부른다. 브로커 호출 0.

## State mutations and fallbacks

임시 원장만.

## Safety conclusion

- High-risk impact: no(시험). 단언은 약화되지 않았다 — 바뀐 것은 새 호출 형태(기대 intent) 또는 이름 붙은 대가의 +1 관측뿐.
