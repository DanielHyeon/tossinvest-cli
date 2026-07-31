# Function Logic Map: `LoadActivationRecord`

- Source: `internal/candidate/thresholdset.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| activation JSON reader | exactly one strict object | a046 design decision 9 | zero record + error |
| version/scope/digests/time/approver | complete; KR/US regular; sha256 lower hex | separate human activation schema | zero record + field-specific error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | strict decode fails | none | zero + wrapped decode error | strict unknown/trailing tests |
| B2 | validate required record fields | none | first invalid field errors | activation tests |
| B3 | version absent | none | zero + version error | strict activation matrix |
| B4 | market unsupported | none | zero + market error | strict activation matrix |
| B5 | session unsupported | none | zero + session error | strict activation matrix |
| B6 | set digest malformed | none | zero + set_digest error | strict activation matrix |
| B7 | evidence digest malformed | none | zero + evidence_digest error | strict activation matrix |
| B8 | approval instant absent | none | zero + approved_at error | strict activation matrix |
| B9 | approver absent | none | zero + approved_by error | missing approver test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `decodeOneStrictJSON`, string/time normalization, digest regex | strict immutable parse | synchronous; no retry/I/O beyond supplied reader | CodeGraph + AST |

## State mutations and fallbacks

- Local document normalization only. No registry mutation, clock read, numeric fallback, or activation side effect.

## Safety conclusion

- Safe edit boundary: parse an inert record; activation occurs only in the binding loader.
- High-risk impact: yes for approval integrity, covered by zero-on-every-error tests.
