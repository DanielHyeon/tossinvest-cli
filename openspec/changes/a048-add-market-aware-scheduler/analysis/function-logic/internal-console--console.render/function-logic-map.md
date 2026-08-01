# Function Logic Map: `Console.render`

- Source: `internal/console/pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| template name/data | registered server template and typed page view | `pages` template set | render failure returns generic HTTP 500 |
| response headers | fixed local security policy | `Console.render` | no cache, same-origin referrer, nosniff and restrictive CSP |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | template execution fails | writes no partially rendered page | `http.Error` then return | existing template execution tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `pages.ExecuteTemplate` | render into an in-memory buffer before committing headers | error is handled once; no retry | CodeGraph + AST |
| `http.ResponseWriter.Header/WriteHeader/Write` | publish successful HTML and fixed security headers | write errors are not actionable after response commit | AST + console page tests |

## State mutations and fallbacks

- Only the response buffer and headers are mutated. The added CSP is static and
  grants no network, script, frame, or navigation capability.

## Safety conclusion

- Safe edit boundary: one fixed `Content-Security-Policy` header on successful HTML responses.
- High-risk impact: low; full console tests verify the exact header and every existing page still renders.
