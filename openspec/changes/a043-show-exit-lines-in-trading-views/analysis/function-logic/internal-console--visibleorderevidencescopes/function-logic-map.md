# Function Logic Map: `visibleOrderEvidenceScopes`

- Source: `internal/console/orders.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| one broker reading | visible open/closed/conditional rows | orders cache | omit invalid identity and deduplicate exact composite |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid/already seen key | local set only | skip | visible-scope tests |
| B2 | open rows | append exact scopes | continue | orders tests |
| B3 | closed rows | append exact scopes | continue | orders tests |
| B4 | conditional rows | append own-id scope only | continue | conditional tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `evidenceKey` | derive exact identity | invalid rows omitted | AST + orders tests |

## State mutations and fallbacks

- Builds a bounded request only from rows already returned for this screen.

## Safety conclusion

- Safe edit boundary: read-only evidence request construction.
- High-risk impact: no live call or mutation.
