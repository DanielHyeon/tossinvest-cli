# Function Logic Map: `EvaluateLadder`

- Source: `internal/exitpolicy/ladder.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Policy + state identity | validated immutable policy; state identity matches | registry/journal state | refusal, no state/order side effect |
| Decimal price/state inputs | positive entry/observed/high-water/baseline; legal rung | journal + quote | refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | policy invalid or state identity mismatch | none | refusal | policy identity tests |
| B2 | invalid prices/ratio/rung | none | refusal | existing validation tests |
| B3 | observation reaches new highest rung | local watermark/rung promotion | continue | rung tests |
| B4 | lock/runner raises baseline | local monotone baseline | continue | common-policy runner tests |
| B5 | completed state | undo decision-time promotion | early state-only return | completed test |
| B6 | observed below newly composed baseline | full stop or duplicate suppression | early return | breach/pending tests |
| B7 | final/partial/state-only rung | proposal or promotion-only | continue/return | rung outcome tests |
| B8 | pending proposal | keep promotion; suppress order | return | pending test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `LadderPolicy.Validate` | identity/table/monotonicity guard | pure/refusal | CodeGraph + AST |
| `ComputeProtectedStop` | compose stored lock and optional runner | fail closed | CodeGraph + AST |
| `lockPrice`/decimal parsers | exact decimal boundaries | pure/refusal | CodeGraph + AST |

## State mutations and fallbacks

- Pure decision-time computation; execution-time fields are copied unchanged.
- Promotion precedes breach, and breach displaces a pending partial.

## Safety conclusion

- Safe edit boundary: add immutable identity and snapshot projection without changing preset values or breach ordering.
- High-risk impact: yes — stop/take-profit judgement.
