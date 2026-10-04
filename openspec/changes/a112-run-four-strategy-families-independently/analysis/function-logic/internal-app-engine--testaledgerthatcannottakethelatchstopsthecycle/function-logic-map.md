# Function Logic Map: `TestALedgerThatCannotTakeTheLatchStopsTheCycle (시험)`

- Source: `internal/app/engine/a112_lane_latch_durability_test.go`
- Source SHA-256: `bf27865d0f81e8a13f1443fa18f697676ed951c8054bffb5f2355da7aa1295de`
- Signature: `TestALedgerThatCannotTakeTheLatchStopsTheCycle(params=1, results=0)`
- Source range: `282:1`–`294:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 288:2 | 시험 자신의 갈래 |
| B2 | if | 291:2 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `laneLatchFixture` | 283:13 |
| `context.Background` | 284:9 |
| `mustProductionStrategyLanes` | 285:13 |
| `runtime.lanesFor` | 286:10 |
| `latchOneLane` | 287:2 |
| `c.Journal.Close` | 288:12 |
| `t.Fatalf` | 289:3 |
| `runtime.evaluate` | 291:12 |
| `t.Fatal` | 292:3 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
