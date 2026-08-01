# Function Logic Map: `TestTheSeamsWithVerbExemptionsAreEnumeratedAndArgued`

- Source: `internal/console/orders_static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| console capability exemptions | exact field/type/method names with non-empty rationale | static capability registry | assertion failure on any unreviewed mutation verb exemption |

## Branches and early returns

| Branch group | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B17 | enumerate capability exemptions and shared verb lists | none | assertion failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| static capability registry | audits injected seam names | deterministic source contract | AST |

## State mutations and fallbacks

- The a050 `OptimizationCommander.Apply` exemption is explicitly enumerated; shared mutation verbs remain intact.

## Safety conclusion

- Safe edit boundary: add one argued, exact least-capability seam entry.
- High-risk impact: yes; this prevents capability-name exemptions from silently widening.
