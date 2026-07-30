# Function Logic Map: `engineArgs`

- Source: `cmd/tossctl/engineproc.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root` | nil or parsed global options | Cobra root command | nil/empty values use child defaults |
| config dir | empty or absolute profile path | `--config-dir` | omitted leaves default profile |
| session file | empty or absolute credential path | `--session-file` | omitted leaves child default resolution |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | root and config dir non-empty | prepend global config flag | child uses same config profile | existing engineArgs test |

The change adds an independent branch for the explicit session file so a
container child reads the copied 0600 secret rather than a nonexistent default.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | pure argument construction | no I/O or retry | AST |

## State mutations and fallbacks

- Returns a new slice; does not mutate root.
- Global flags precede `engine run`, matching Cobra parsing.

## Safety conclusion

- Safe edit boundary: append one explicit global flag pair.
- High-risk impact: yes — the child must use exactly the console's authenticated session.
