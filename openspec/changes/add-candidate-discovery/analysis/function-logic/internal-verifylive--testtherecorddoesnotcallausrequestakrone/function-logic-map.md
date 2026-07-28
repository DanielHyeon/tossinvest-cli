# Function Logic Map: `TestTheRecordDoesNotCallAUSRequestAKROne`

- Source: `internal/verifylive/us_market_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트**다(HEAD L256-283). 앞선 커밋 `f62457c`(`apply-us-measurement-fixes`)의 것이며,
`add-candidate-discovery`의 base가 그보다 앞서 이 change의 diff에 잡힌다.

요청은 이미 시장을 따르고 있었지만(`TestTheUSAmendSendsNoQuantity`) 그 **옆에 적히는
문장**은 아니었다 — 두 detail이 KR을 이름으로 박은 고정 문자열이었다. 그래서 US 실행이
"브로커가 KR price+quantity 정정을 받았다"고 기록했고, 그 요청은 수량을 아예 싣지 않았다.

이것은 미관 문제가 아니다. **기록은 이후 change가 브로커의 행동을 판단하기 위해 읽는
증거**이고, 거기 있는 거짓 문장은 없는 문장보다 나쁘다 — change 2c가 귀속 규칙을 쓰기 위해
바로 이 줄들을 읽는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 하네스 | `newHarness(t, usBroker(), alwaysConfirm())` — **fakeBroker** | `verifylive_test.go` | 실계좌·실브로커 무접촉 |
| 실행 | `Options{Market: MarketUS, Symbol: "MWG", HoldingSymbol: "MWG"}` | 이 테스트 | 실패 시 `t.Fatalf` |
| 관측 | `StepOrderCancel:order.place.ok`, `StepOrderAmend:order.amend.ok` | `observationDetail` | 부재면 `t.Fatalf` |

불변식: 양방향 단언이다 — KR이 **없어야** 하고 US가 **있어야** 한다. 한쪽만 재면
detail을 비워 버리는 "수정"이 통과한다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L258) | 실행 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 (if, L264) | 배치 detail에 `KR`이 있음 | 없음 | `t.Errorf` — US 배치를 KR로 기록했다 | 자체 실행 |
| B3 (if, L267) | 배치 detail에 `US`가 없음 | 없음 | `t.Errorf` — 어느 시장을 쟀는지 말하지 않는다 | 자체 실행 |
| B4 (if, L272) | 정정 detail에 `KR`이 있음 | 없음 | `t.Errorf` | 자체 실행 |
| B5 (if, L275) | 정정 detail에 `US`가 없음 | 없음 | `t.Errorf` | 자체 실행 |
| B6 (if, L280) | 정정 detail에 `quantity`가 있음 | 없음 | `t.Errorf` — 요청이 싣지 않은 수량을 주장하는 문장은 브로커가 `us-modify-quantity-not-supported`로 거절했을 요청을 서술한다 | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newHarness` / `usBroker` / `alwaysConfirm` | fakeBroker 하네스 | 실주문 없음 | ast.json calls |
| `h.run` | US 실행 | 오류 그대로 단언 | ast.json calls |
| `observationDetail` | 기록에서 detail 추출 | 부재는 `t.Fatalf` | ast.json calls |
| `strings.Contains` | 문자열 단언 | 순수 | ast.json calls |

## State mutations and fallbacks

- fakeBroker와 `t.TempDir()`만 쓴다. **LIVE 주문 side effect 0** — 실계좌·실브로커 무접촉.

## Safety conclusion

- Safe edit boundary: 신규 테스트 가산.
- High-risk impact: **no** — 테스트 전용이고 fakeBroker를 쓴다.
  재는 대상은 High-risk(라이브 주문 단계의 기록)이며, 이 테스트가 없으면 US 실행의
  기록이 다시 조용히 거짓이 될 수 있다.
