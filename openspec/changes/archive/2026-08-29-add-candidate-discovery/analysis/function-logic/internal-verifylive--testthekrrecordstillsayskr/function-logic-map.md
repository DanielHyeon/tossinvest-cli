# Function Logic Map: `TestTheKRRecordStillSaysKR`

- Source: `internal/verifylive/us_market_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트**다(HEAD L287-304). 앞선 커밋 `f62457c`(`apply-us-measurement-fixes`)의 것이며,
`add-candidate-discovery`의 base가 그보다 앞서 이 change의 diff에 잡힌다.

`TestTheRecordDoesNotCallAUSRequestAKROne`의 **나머지 반쪽**이다. 시장을 이름으로 부르게
만드는 수정이 "아무것도 부르지 않기"로 변질되면 안 된다 — detail을 비워 버리는 것도
US 테스트를 통과시키기 때문이다. 그래서 KR 실행에서는 `KR`이 여전히 있어야 하고,
정정 detail에는 브로커가 요구하는 `quantity`가 여전히 있어야 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 하네스 | `newHarness(t, newFakeBroker().withHolding("005930", 3), alwaysConfirm())` | `verifylive_test.go` | 실계좌 무접촉 |
| 실행 | `Options{HoldingSymbol: "005930"}` — 기본 시장(KR) | 이 테스트 | 실패 시 `t.Fatalf` |
| 관측 | `StepOrderCancel:order.place.ok`, `StepOrderAmend:order.amend.ok` | `observationDetail` | 부재면 `t.Fatalf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L289) | 실행 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 (if, L294) | KR 배치 detail이 `KR`을 말하지 않음 | 없음 | `t.Errorf` | 자체 실행 |
| B3 (if, L298) | KR 정정 detail이 `KR`을 말하지 않음 | 없음 | `t.Errorf` | 자체 실행 |
| B4 (if, L301) | KR 정정 detail이 `quantity`를 말하지 않음 | 없음 | `t.Errorf` — KR 정정은 브로커가 수량을 요구한다 | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newHarness` / `newFakeBroker().withHolding` / `alwaysConfirm` | fakeBroker 하네스 | 실주문 없음 | ast.json calls |
| `h.run` | KR 실행 | 오류 그대로 단언 | ast.json calls |
| `observationDetail` | 기록에서 detail 추출 | 부재는 `t.Fatalf` | ast.json calls |
| `strings.Contains` | 문자열 단언 | 순수 | ast.json calls |

## State mutations and fallbacks

- fakeBroker와 `t.TempDir()`만 쓴다. **LIVE 주문 side effect 0**.

## Safety conclusion

- Safe edit boundary: 신규 테스트 가산.
- High-risk impact: **no** — 테스트 전용, fakeBroker.
  이 테스트의 가치는 **hold-the-line**이다: US 문장을 고치는 수정이 KR 기록을 훼손하거나
  detail을 비워 버리는 방향으로 가지 못하게 막는다.
