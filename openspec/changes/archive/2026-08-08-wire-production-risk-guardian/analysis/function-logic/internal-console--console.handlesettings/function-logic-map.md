# Function Logic Map: `Console.handleSettings`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| authenticated request | GET settings after session gate | routes | no mutation |
| optional setting seams | nil or narrow reader/editor | `Console.Options` | render each section unwired/error |
| updater | nil or fixed-sibling inspector | CLI injection | render unavailable or reviewed metadata |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B6 | each legacy seam wired/load success or error | populate only that section | render error locally | settings suites |
| B7 | `SystemUpdater != nil` | call read-only `Inspect` and populate update panel | invalid candidate becomes reason/no button | update render test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| settings/limits/policy `Load` | read current settings | error stays section-local | CodeGraph + AST |
| `SystemUpdater.Inspect` | inspect fixed candidate metadata without execution | returns explicit refusal reason | updater contract |
| `c.render` | render one coherent settings page | template error handled centrally | CodeGraph |

## State mutations and fallbacks

- No settings or executable bytes are changed by GET.
- Candidate path is supplied by the updater, never the request.

## Safety conclusion

- Safe edit boundary: append a read-only update view after existing settings reads.
- High-risk impact: yes — metadata is the human review boundary for executable installation.
