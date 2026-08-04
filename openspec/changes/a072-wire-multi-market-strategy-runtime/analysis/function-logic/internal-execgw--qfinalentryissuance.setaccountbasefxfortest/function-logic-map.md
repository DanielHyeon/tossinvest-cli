# Function Logic Map: `QFinalEntryIssuance.SetAccountBaseFXForTest`

- Source: `internal/execgw/export_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sealed account-base FX authority | package-sealed `risk.AccountBaseFX`; may be zero only for an explicit refusal test | tagged risk test seam or production-equivalent package constructor used by the test | the downstream q_final precheck rejects an invalid or absent seal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | called by an execgw package test | copies the opaque value and marks the private test slot present | none | paired KR/US q_final account-base tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | direct assignment keeps scalar FX data inaccessible to execgw tests | no fallback or conversion | AST |

## State mutations and fallbacks

- This helper exists only in `_test.go`; production callers cannot inject or reconstruct FX authority.
- It does not validate the seal. The same production precheck and issue-time revalidation remain authoritative.

## Safety conclusion

- Safe edit boundary: test-only transport of the opaque risk package value.
- High-risk impact: **no** — unavailable in production builds.
