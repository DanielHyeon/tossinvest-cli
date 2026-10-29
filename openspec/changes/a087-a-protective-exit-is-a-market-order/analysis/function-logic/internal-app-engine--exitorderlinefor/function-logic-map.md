# Function Logic Map: `exitOrderLineFor`

- Source: `internal/app/engine/exit_order_log.go:102-114`
- AST evidence: `ast.json` (Pre-Edit HEAD `a284a2b0a`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| selected limit | lower-limit, observed, baseline, empty | `submit` single decision | sent price must equal line price |
| observed | original observed string | quote | never overwrite log field |
| baseline | persisted state string | exit state | fallback only if observed empty |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | observed blank | price uses baseline | return line | baseline source golden |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | current fallback canonicalization | no error | AST B1 |

## State mutations and fallbacks

- signature gains selected price/source while `observedPrice` stays original.
- lower-limit source is legal only when selected price came from the immutable board.

## Safety conclusion

- Safe edit boundary: a129 diagnostic line only; broker write path is untouched.
- High-risk impact: yes — log must not assert a different price than the intent.
