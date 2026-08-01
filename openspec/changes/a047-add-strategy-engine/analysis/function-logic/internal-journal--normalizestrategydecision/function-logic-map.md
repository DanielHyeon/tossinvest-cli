# Function Logic Map: `normalizeStrategyDecision`

- Source: `internal/journal/strategy_lineage.go`
- CodeGraph callers/callees: journal plan and production issuance preflight
- AST: generated after implementation

## Inputs and invariants

| Input/state | Range | Source of truth | Failure behavior |
|---|---|---|---|
| scalar identifiers | surrounding whitespace removed | typed journal request | later completeness check rejects empty |
| decision payload | exact canonical bytes, never rewritten | opaque decision serializer | strict verifier rejects whitespace/noncanonical bytes |
| creation time | supplied UTC or injected journal time | journal clock | normalize to UTC |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Test |
|---|---|---|---|---|
| B1 | `value.CreatedAt.IsZero()` after all named scalar fields are trimmed | local copy only | injected caller-supplied time | production issuance tests |
| B2 | lineage creation time is nonzero | local copy only | UTC-normalized time | exact plan/idempotency tests |
| Invariant | payload bytes may be whitespace/noncanonical | payload remains unchanged | later verifier rejects | strict payload table |

## Calls and live bindings

| Callee | Contract | Failure path | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | identifiers only, never signed payload | completeness validation | tests |
| `time.UTC` | deterministic journal timestamp | none | replay tests |

## State mutations and fallbacks

- Mutates only a local value copy. Signed payload bytes are never trimmed or rewritten; no fallback exists.

## Safety conclusion

- Pure normalization; signed/canonical payload bytes remain immutable.
