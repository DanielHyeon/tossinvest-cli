# Function Logic Map: `Console.handleSettings`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| authenticated request | GET `/settings` | route/session wrapper | unauthenticated request never reaches handler |
| each settings seam | nil or a closed read interface | `console.Options` wiring | nil renders explicit unwired state |
| load result | value or error | config service | error is rendered in its own section |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | adoption seam wired / load error | read only | render block or section error | existing settings tests |
| B3-B4 | limits seam wired / load error | read only | render gate or section error | existing limits tests |
| B5-B6 | trading seam wired / load error | read only | render policy or section error | existing operating tests |
| B7 | updater seam wired | inspect local candidate | render update state | existing update tests |

The change adds autostart wired/load-error branches isolated from all existing
sections.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Settings.Load` | adoption display | error stays local to section | AST |
| `Limits.Load` | Guardian/gate display | error stays local to section | AST |
| `TradingPolicy.Load` | LIVE policy display | error stays local to section | AST |
| `Autostart.Load` | new process-lifecycle display | error must not trigger start | design + new test |
| `render` | emit authenticated settings page | template escaping applies | AST |

## State mutations and fallbacks

- This GET handler performs no save and no process start.
- Each seam is independently optional and errors do not erase other sections.

## Safety conclusion

- Safe edit boundary: add one independent load block and view fields before render.
- High-risk impact: yes — the displayed value is the human approval state.
