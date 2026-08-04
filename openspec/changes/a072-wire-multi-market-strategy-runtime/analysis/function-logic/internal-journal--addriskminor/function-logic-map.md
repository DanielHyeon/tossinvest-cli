# Function Logic Map: `addRiskMinor`

- Source: `internal/journal/risk_bucket.go`
- Qualified function: `addRiskMinor`
- AST evidence: `ast.json` (base revision)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

`left` and `right` are unsigned base-10 minor-unit integers. Empty left is legacy zero; right has no
empty fallback. Result precision is bounded to 256 bits.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | left is empty | local left becomes canonical zero | continue | replay compatibility tests |
| B2 | left is non-decimal or negative | none | replay-mismatch error | malformed replay tests |
| B3 | right is non-decimal or negative | none | replay-mismatch error | malformed delta tests |
| B4 | sum exceeds 256 bits | none | replay-mismatch overflow | boundary tests |

## Calls and live bindings

`big.Int.SetString`, `Sign`, `Add` and `BitLen` provide exact integer arithmetic; there is no float,
currency conversion, saturation or truncation fallback.

## State mutations and fallbacks

Only local big integers are mutated. Durable callers commit the returned canonical decimal or fail their
encompassing transaction. The current implementation moved this responsibility into the versioned riskbucket core.

## Safety conclusion

- Safe edit boundary: preserve exact unsigned decimal and 256-bit overflow refusal.
- High-risk impact: yes — silent wrap/saturation would understate account-wide held/filled risk.
