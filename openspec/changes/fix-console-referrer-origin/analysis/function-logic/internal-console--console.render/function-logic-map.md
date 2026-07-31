# Function Logic Map: `Console.render`

- Source: `internal/console/pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `w` | active response writer | route handler | receives a deterministic error or rendered document |
| `name` | registered template name | `pages` template set | B1 returns HTTP 500 |
| `data` | shape expected by `name` | calling page handler | B1 returns HTTP 500 |
| response headers | HTML content type, no-store, referrer, MIME policy | `Console.render` | wrong referrer policy can override the remote wrapper and break browser origin evidence |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `pages.ExecuteTemplate` returns an error | renders into a buffer only; writes error response | HTTP 500, early return | existing template compilation/render tests plus full console suite |
| final | template execution succeeds | sets headers, writes HTTP 200 and buffered HTML | rendered page | `TestConsoleDocumentsUseSameOriginReferrerPolicy` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `pages.ExecuteTemplate` | render before committing response headers/status | error is converted to HTTP 500; no retry | AST B1 |
| `http.Error` | return deterministic rendering failure | HTTP 500 and early return | AST B1 |
| `w.Header().Set` | apply HTML, cache, referrer, and MIME headers | deterministic in-memory mutation; no retry | AST lines 382-385 |
| `w.WriteHeader` / `w.Write` | commit HTTP 200 and buffered page | write errors intentionally ignored by existing behavior | AST lines 386-387 |

## State mutations and fallbacks

- Mutates only the current response header map and response bytes/status.
- Does not mutate console state, settings, journal, engine, orders, risk
  controls, or operating toggles.
- Template execution is buffered, so an execution failure cannot leak a
  partially rendered page.

## Safety conclusion

- Safe edit boundary: replace only the rendered response `Referrer-Policy`
  literal; keep template execution, remaining headers, status, and writes
  unchanged.
- High-risk impact: yes, because the later rendered header determines the
  effective browser document policy on normal console pages.
