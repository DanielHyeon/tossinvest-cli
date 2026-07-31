# Function Logic Map: `CommonLadderForPosition`

- Source: `internal/exitpolicy/common_policy.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| common policy id | one registered stable option id | server registry | unknown id error |
| adopted | bool; only RUNNER changes to floor-only semantics | stored position provenance | derived immutable adopted version/digest |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | registry miss | none | error | unknown common policy test |
| B2 | adopted RUNNER | clone partial ratios zero; derive distinct version/digest | ladder value | adopted runner tests |
| B3 | other policy/provenance | no semantic change | ladder value | common policy tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `CommonPolicyByID` | deep-copy registered policy | pure lookup | CodeGraph + AST |
| `LadderPolicy.Identity` | bind adopted variant to its actual semantics | pure/fail closed | CodeGraph + AST |

## State mutations and fallbacks

- Only returned copies are mutated. Process registry remains immutable.

## Safety conclusion

- Safe edit boundary: preserve StockOS adopted RUNNER floor-only behavior while giving it a distinct immutable version/digest.
- High-risk impact: yes — adopted exit policy resolution.
