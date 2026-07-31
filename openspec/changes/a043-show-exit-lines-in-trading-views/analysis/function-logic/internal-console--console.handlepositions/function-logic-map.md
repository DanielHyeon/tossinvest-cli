# Function Logic Map: `Console.handlePositions`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| authenticated `*http.Request` | session wrapper has already accepted it; a043 additionally requires GET/HEAD only | `Console.routes` + `Console.readOnly` | unsupported method returns 405 before this handler |
| positions snapshot | broker and read-only journal halves may independently be absent | `Console.positions` | preserve partial view and typed unknown reasons |
| adoption settings snapshot | optional; display-only labels may be stamped but no mutation controls may render | `Options.Settings.Load` | load failure leaves labels unstamped; positions remain readable |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | settings seam is nil | none | render positions without designation labels | `TestTradingViewsAreInputFreeAndRejectPOST` |
| B2 | settings load succeeds | stamp include/exclude display state only | render positions with the same exit read model | existing settings label tests + a043 DOM test |
| B3 | settings load fails | none | render without claiming designation/exclusion state | existing unwired/error tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Console.positions` | assemble broker/journal snapshot | failures are values, not handler errors | CodeGraph: only caller is this handler |
| `AdoptionSettings.Load` | preserve existing display-only designation context | no retry; error degrades locally | current HEAD + AST |
| `operatorview.FromSnapshot` (planned) | build one canonical exit-line DTO from persisted snapshot | unknown/stale stays typed and renders as `—` | a041/a042 API + planned adapter tests |
| `Console.render` | server-render semantic HTML under existing CSP | render failure follows common console error path | current HEAD + AST |

## State mutations and fallbacks

- Current code puts CSRF-backed include/exclude forms directly on `/positions`; a043 removes those action surfaces while retaining the settings endpoints on `/settings`.
- Planned mutation is view-local only: attach value DTOs and render. No journal, broker, config, or account write is reachable.

## Safety conclusion

- Safe edit boundary: remove action capability from the page, preserve read seams and typed failures, and map persisted values without recomputation.
- High-risk impact: no. Exit-policy evaluation, live order submission, journal writes, and operating toggles are untouched.
