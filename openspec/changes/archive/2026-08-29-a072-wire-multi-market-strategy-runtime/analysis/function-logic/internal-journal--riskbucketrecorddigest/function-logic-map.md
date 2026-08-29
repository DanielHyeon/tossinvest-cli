# Function Logic Map: `riskBucketRecordDigest`

- Source: `internal/journal/risk_bucket.go`
- Qualified function: `riskBucketRecordDigest`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

`value` is a versioned risk-bucket record whose JSON representation is the immutable authority preimage.
Callers must supply canonical types; maps with unstable semantic ordering are not accepted as an alternate format.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | JSON encoding fails | none | wrapped record-encoding error | admission/fill rollback tests |
| success | exact bytes encode | none | lowercase SHA-256 hex | record tamper/replay suites |

## Calls and live bindings

Uses only standard canonical JSON encoding and SHA-256. It reads no mutable runtime authority.

## State mutations and fallbacks

Pure calculation; no zero/default digest is emitted and enclosing journal work must roll back on failure.

## Safety conclusion

- Safe edit boundary: digest exactly the versioned record bytes without repair.
- High-risk impact: yes — record mismatch must fail closed rather than revive stale capacity.
