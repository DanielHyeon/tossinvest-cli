# Function Logic Map: `Console.handleOptimization`

- Source: `internal/console/optimization.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| HTTP method | GET or HEAD only | handler contract | 405, no read or write |
| `category` query | six registry IDs or empty | a050 UI IA | unknown falls back to overview with warning |
| optimization commander | optional narrow lifecycle seam | `Console.Options` | nil/unavailable renders read-only state |
| exit-policy compatibility seam | optional read-only fallback | `ExitPolicySettings` | read failure is visible; no fabricated value |
| protection commander | optional engine-owned opaque capability seam | `Console.Options` | nil renders explicit OFF/UNWIRED; read failure is visible |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | method is not GET/HEAD | refusal response only | 405 | method refusal test |
| B2 | unknown category | none | overview + warning | category fallback test |
| B3 | commander nil/unavailable | none | read-only/unavailable categories | unwired state test |
| B4 | lifecycle snapshot read fails | none | stale/error status, controls absent | reader failure test |
| B5 | valid snapshot | template render only | 200 | six-category/UI contract test |
| B6 | protection seam is wired | one bounded status-list read | status rows or visible error | protection UI tests |
| B7 | protection status read fails | none | fail-closed read error | protection UI error test |
| B8 | protection status read succeeds | none | opaque row actions only | protection row test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OptimizationCommander.Read` | loads immutable lifecycle projection | one bounded call; error is rendered, never retried | CodeGraph + AST |
| `ExitPolicySettings.Load` | compatibility read until adapter lands | one call; error rendered | CodeGraph + AST |
| `ProtectionCommander.List` | loads broker-resident protection display rows | one bounded call; error rendered, never retried | CodeGraph + AST |
| `Console.render` | emits CSP-compatible server HTML | template error is centralized | CodeGraph + AST |

## State mutations and fallbacks

- No account, journal, broker, lane, gate, kill-switch, or LIVE mutation. Protection rows expose only opaque, server-issued action tokens.
- Unknown/missing owner descriptors remain read-only and never receive guessed defaults.

## Safety conclusion

- Safe edit boundary: query parsing, immutable view composition, template rendering.
- High-risk impact: yes (operator settings surface), mitigated by capability-free GET and unavailable fail-closed state.
