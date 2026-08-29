# Function Logic Map: `verifyProjectionRiskIntent`

- Source: `internal/journal/strategyflow_projection.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| projection | verified accepted lineage and terms | strategyflow projection | reject invalid scope/terms |
| risk | exact q_final and executable prices | Guardian RiskIntent | reject mismatch; never repair authority |
| schema | v2 literal-minor or v3 canonical-major semantics | payload dispatcher | unknown schema fails closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | quantity invalid, zero, or above candidate | none | exact-binding error | q_final drift table |
| B2 | account/market/symbol mismatch | none | exact-binding error | scope drift table |
| B3 | v2 literal price differs from sealed minor | none | exact-binding error | v2 compatibility test |
| B4 | v3 cannot derive canonical major decimal | none | exact-binding error | malformed provenance test |
| B5 | v3 major price differs | none | exact-binding error | paired price drift table |
| B6 | exact scope and prices | none | nil | paired KR/US success |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `NormalizeDecimal` / `ParseUint` | authenticate q_final integer | deterministic; no retry | AST |
| `PriceProvenance.MajorDecimal` | derive decimal without floats | deterministic `(string,bool)` | pre-edit contract |

## State mutations and fallbacks

- Pure verifier with no database, network, lease, broker, activation, or fallback.
- v2 is never silently reinterpreted as v3.

## Safety conclusion

- Safe edit boundary: compare-only function before journal mutation.
- High-risk impact: yes; a US unit error changes notional by 100x.
