# Function Logic Map: `proposalQuantity`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| remaining | canonical non-negative whole-share decimal in engine path | journal position projection | error for invalid; non-positive clamps/reads zero |
| ratio | empty defaults to 1; canonical decimal expected in (0,1] | exitpolicy proposal | parse error; >1 clamps to 1 for legacy safety |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | remaining cannot parse | none | error | quantity unit test |
| B2 | ratio empty/nonempty invalid | default 1 or error | error/continue | quantity unit test |
| B3 | ratio > 1 | clamp to 1 | continue | quantity unit test |
| B4 | product fractional | floor whole units | value | one-share partial test |
| B5 | negative result | clamp to zero | value | quantity unit test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `math/big.Rat.SetString` | exact ratio multiplication | pure parse | CodeGraph + AST |
| `big.Int.Quo` | truncate toward zero/whole-share projection | pure | CodeGraph + AST |

## State mutations and fallbacks

- Pure helper; current caller filters zero only after projection. Snapshot evaluator will own the same arithmetic.

## Safety conclusion

- Safe edit boundary: delegate to a shared exitpolicy projector without changing floor/clamp behavior.
- High-risk impact: yes — sizing, but conservative whole-share behavior is preserved and explicitly tested.
