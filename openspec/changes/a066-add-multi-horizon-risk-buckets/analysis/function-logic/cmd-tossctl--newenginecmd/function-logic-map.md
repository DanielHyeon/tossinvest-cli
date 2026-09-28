# Function Logic Map: `newEngineCmd`

- Source: `cmd/tossctl/engine.go`
- AST evidence: `ast.json` (post-edit, 114–127, **0 branches**)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| root | the CLI root options | `newRootCmd` | — (no branch) |

## Branches and early returns

None. The function builds the `engine` command and registers subcommands. The a066 5.5 edit adds one statement,
`cmd.AddCommand(newEngineRiskRelaxationCmds(root)...)`, which registers `entry-lock-release`, `risk-latch-release`
(both `mutating: true`) and `risk-latch-show` (`mutating: false`).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newEngineRiskRelaxationCmds(root)` | the three a066 5.5 commands (`engine_risk_relaxation.go`) | none at construction | AST + `TestEngineRiskRelaxationCommandsDeclareWhatTheyAre` |

## State mutations and fallbacks

- None at construction. The release commands reach the running engine's control endpoint only when executed.

## Safety conclusion

- Registration only. Both release commands are declared mutating, so interactive agents never auto-run them (pinned by
  `TestMutatingAnnotationOnTradeCommands` and mutation C06).
- High-risk impact: indirect (it exposes the relaxation commands); the release semantics are in the journal API.
