# Function Logic Map: `Console.renderHTML`

- Source: `internal/console/pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| status/template/data/CSP | explicit caller-owned values and parsed templates | render/refuse/optimization preview | template failure returns 500 before partial page output |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | template execution fails | error response only | 500 and early return | existing template renderer tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `pages.ExecuteTemplate` | render fully before headers/body | single attempt | CodeGraph + AST |
| `http.ResponseWriter` | writes no-store, referrer, nosniff and caller CSP | exact caller status | security response tests |

## State mutations and fallbacks

- Buffers the entire template before writing; no domain state is touched.

## Safety conclusion

- Safe edit boundary: one shared HTML header/status writer.
- High-risk impact: yes; callers cannot accidentally omit the standard response headers.
