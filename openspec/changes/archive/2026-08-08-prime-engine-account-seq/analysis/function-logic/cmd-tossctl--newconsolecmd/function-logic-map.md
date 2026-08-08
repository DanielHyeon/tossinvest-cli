# Function Logic Map: `newConsoleCmd`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root` | CLI root options | root command | forwarded to `runConsole` |
| help/annotation | loopback-only, human-approved mutating console | route implementation | stale operator contract |
| port | integer, zero means OS-selected | Cobra | parse refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | happy path (AST is branchless) | constructs command/flag | returns command | registration/help tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | normalize help | cannot fail | CodeGraph + AST |
| `runConsole` | execute server assembly | propagates error | CodeGraph + AST |
| `IntVar` | bind loopback port | Cobra validation | AST |

## State mutations and fallbacks

- No account or executable mutation occurs during command construction.
- `mutating=true` remains because authenticated routes can verify live orders and install the fixed candidate.

## Safety conclusion

- Safe edit boundary: help text only; annotations, flag set and `RunE` stay fixed.
- High-risk impact: yes — operator contract for live verification and self-replacement.
