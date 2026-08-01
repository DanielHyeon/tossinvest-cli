# Function Logic Map: `visibleOrderEvidenceScopes`

- Source: `internal/console/orders.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| filtered rendered rows | only rows in `v.Rows` | local filter result | omit invalid identity and deduplicate id+account+market+day |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate rendered rows | local set only | continue | filtered visible evidence test |
| B2 | invalid/already seen key | none | skip | invalid identity/duplicate tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | keys were derived while preserving raw broker identity | invalid rows omitted | AST + orders tests |

## State mutations and fallbacks

- Builds a bounded request only after local links have selected the rendered rows.

## Safety conclusion

- Safe edit boundary: read-only evidence request construction.
- High-risk impact: no live call or mutation.
