# Function Logic Map: `Gateway.checkDecision`

- Source: `internal/execgw/guardian.go`
- Qualified function: `Gateway.checkDecision`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| durable decision and opaque reference | matching generation/account and intact risk hash/preimage | journal row re-read by Gateway | typed refusal; broker zero |
| actual mutation plan | class and preimage must match actual order shape | private `mutationPlan`, never caller class alone | class/preimage refusal |
| Gateway clock | strictly before nonzero decision expiry | injected clock | missing/expired refusal |
| limits and optional account-base FX envelope | complete supported snapshot; paired KR/US strategy path carries exact private FX capability | persisted decision plus private strategy capability | limit/FX refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | stored risk hash differs from stored preimage | none | tampered refusal | Guardian tamper suite |
| B2 | persisted preimage cannot be parsed | none | tampered refusal | malformed preimage suite |
| B3-B5 | generation/account identity switch finds mismatch | none | tampered or intent-mismatch refusal | generation/account substitution tests |
| B6-B11 | verified safety class does not match actual exposure shape or is unknown | none | class-mismatch refusal | class/shape matrix |
| B12 | exact persisted preimage differs from order | none | field-specific mismatch refusal | preimage matrix |
| B13 | decision has no expiry | none | Guardian missing refusal | expiry tests |
| B14 | current time reaches/passes expiry | none | expired refusal | boundary tests |
| B15 | derived idempotency key is absent/mismatched/invalid | none | key-mismatch refusal | key tests |
| success | all prior checks pass | none | delegate to `checkLimits` | legacy and paired account-base suites |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `journal.HashPreimage` / `journal.ParsePreimage` | verify durable decision integrity | no repair/fallback | AST + tamper tests |
| `checkPreimage` | compare exact durable risk intent to actual order | first mismatch refuses | preimage tests |
| `checkIdempotencyKey` | bind attempt operation key to decision | no caller key fallback | replay/key tests |
| `checkLimits(dec, plan, now)` | apply quantity and legacy/account-base money limits at this exact clock | read-only; any unsupported envelope fails closed | checkLimits map + paired KR/US Gateway tests |

## State mutations and fallbacks

- Read-only. It is called before dispatch recording and again on a fresh journal row immediately before transport.
- It never fixes malformed authority, substitutes a peer market, or derives FX from scalar caller data.

## Safety conclusion

- Safe edit boundary: preserve integrity → identity → class/shape → preimage → time → key → limits order.
- High-risk impact: yes — this is the repeatable last decision fence before real broker transport.
