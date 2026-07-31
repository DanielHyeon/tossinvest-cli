# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| root config | nil/default or explicit config directory | Cobra root options | path-resolution failure is non-fatal; dashboard journal is unwired |
| console options | validated local/remote options | `remoteAccessOptions` | invalid remote configuration returns before serving |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | context is nil | installs background context | none | existing console command tests |
| B2 | remote options invalid | none | return error | existing remote option tests |
| B3 | prerequisite path/seam resolution fails | none | return error | existing command seam tests |
| B4 | journal path resolution fails | prints diagnostic; injects empty path | continue serving | existing static/read-only tests |
| B5 | explicit config directory exists | inject `<config-dir>/journal.db` | continue serving | `TestConsoleJournalPathFollowsTheEngineProfile` |
| B6 | no explicit config directory | inject `journal.DefaultPath()` | continue serving | `TestConsoleJournalPathFollowsTheEngineProfile` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleJournalPath` (new leaf) | mirror engine profile journal identity | returns path/error; no I/O | CodeGraph + AST + engine contract |
| `engineJournalDir` | marker and lock identity | non-fatal diagnostic on error | CodeGraph + AST |
| `console.ListenAndServe` | inject read-only journal path | server lifecycle contract unchanged | CodeGraph + AST |

## State mutations and fallbacks

- Only the selected path string changes. No journal is opened or written here.
- Resolution failure retains the existing empty-path fallback and warning.

## Safety conclusion

- Safe edit boundary: replace unconditional default resolution with a pure
  profile-aware resolver; preserve every later seam and lifecycle branch.
- High-risk impact: yes — journal identity, but read-only and no schema/write
  authority.
