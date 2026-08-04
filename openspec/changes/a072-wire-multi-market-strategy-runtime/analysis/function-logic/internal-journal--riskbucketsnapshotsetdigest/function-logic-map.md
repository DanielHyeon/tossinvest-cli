# Function Logic Map: `riskBucketSnapshotSetDigest`

- Source: `internal/journal/risk_bucket.go`
- Qualified function: `riskBucketSnapshotSetDigest`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

The ordered snapshot-reference slice is already canonical and immutable for one admission transaction.
Order is part of the authority; this function must not sort, normalize or omit fields.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | canonical JSON encoding fails | none | wrapped encoding error | admission rollback/unsupported-value tests |
| success | encoding succeeds | none | lowercase SHA-256 hex | snapshot substitution/replay tests |

## Calls and live bindings

`json.Marshal`, `sha256.Sum256` and `hex.EncodeToString` bind the exact ordered reference bytes. No
external state, clock, broker or config is read.

## State mutations and fallbacks

Pure calculation; there is no fallback digest and callers must abort the admission transaction on error.

## Safety conclusion

- Safe edit boundary: preserve canonical field order and exact SHA-256 bytes.
- High-risk impact: yes — this digest binds the snapshots consumed by paired KR/US risk authority.
