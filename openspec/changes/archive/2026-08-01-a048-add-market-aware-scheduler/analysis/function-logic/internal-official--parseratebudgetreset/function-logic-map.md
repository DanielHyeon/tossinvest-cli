# Function Logic Map: `ParseRateBudgetReset`

- Source: `internal/official/ratebudget.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| raw | optional decimal reset header, surrounding whitespace allowed | official HTTP header | canonicalized with `TrimSpace`; malformed/negative is unparsed |
| observedAt | non-zero response-completion instant | official client clock | zero makes nonempty reset unparsed |
| threshold/bounds | epoch at exactly 1,000,000,000; plausible delta from -1m through +24h | official constants | anything outside exact parser semantics is unparsed with raw preserved |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | canonical raw is empty | none | absent/zero reset | absent reset tests |
| B2 | decimal parse fails, value is negative, or observed-at is zero | none | unparsed/zero reset | malformed/zero-time tests |
| B3 | seconds are at or above exact threshold | derives Unix UTC candidate | epoch if plausible, otherwise unparsed | threshold/epoch boundary tests |
| B4 | seconds are below threshold | calls overflow-safe delta add | delta if plausible, otherwise unparsed | delta/wrapping boundary tests |
| B5 | derived instant is outside inclusive plausibility window | discards derived instant but preserves canonical raw | unparsed | -1m/+24h boundary tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` / `strconv.ParseInt` | reproduce official header normalization and signed decimal parse | parse errors become `ResetUnparsed` | AST + parser tests |
| `addResetDelta` | convert seconds and add without duration wrap | false becomes `ResetUnparsed` | AST + wrapping tests |
| `plausibleReset` | enforce inclusive [-1m,+24h] window | false becomes `ResetUnparsed` | AST + boundary tests |

## State mutations and fallbacks

- Pure, read-only derivation. No client state, headers, or budget objects are mutated.
- Invalid evidence retains canonical raw provenance while clearing the derived instant.

## Safety conclusion

- Safe edit boundary: official reset-header semantics.
- High-risk impact: yes, because downstream scheduling uses this as its trust root.
