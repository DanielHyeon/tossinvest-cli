# Function Logic Map: `Console.openJournal`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Options.JournalPath` | empty (unwired) or exact engine-selected path, including whitespace | `cmd/tossctl.consoleJournalPath` / engine profile rule | empty reports unwired; non-empty is opened exactly as supplied |
| selected journal | missing, too new, too old, current, or other read failure | `journal.OpenReadOnly` | typed display state; never fallback or write |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | exact supplied path is empty | set view to unwired | nil handle and unwired view | existing unwired portfolio tests |
| B2/B3 | read-only open succeeds | record schema version only | readable handle and OK view | `TestPositionsReadsOnlyTheSelectedJournal` |
| B4 | selected file is missing | set missing state | nil handle | existing missing-journal tests |
| B5 | selected schema is too new | set typed state/detail | nil handle | existing schema-direction tests |
| B6 | selected schema is too old | set typed state/detail | nil handle | `TestOpenReadOnlyRejectsV8BeforePolicyQuery` plus portfolio tests |
| B7 | other read failure | set failed state/detail | nil handle | existing invalid-journal tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `journal.OpenReadOnly` | open the exact selected journal without creation or migration | context-bound, no retry and no fallback | CodeGraph + AST |
| `ReadOnly.SchemaVersion` | render the selected journal's compatible version | only after successful open | CodeGraph + AST |

## State mutations and fallbacks

- Mutates only the request-local `journalView`; successful callers close the
  read-only handle.
- The path is identity data, not user text. Trimming it changes the selected
  file and can redirect a relative whitespace profile to a different absolute
  path.
- No migration, creation, writable handle, or alternate-path fallback exists.

## Safety conclusion

- Safe edit boundary: remove only the text normalization from the injected path;
  retain empty detection and all typed read-only states.
- High-risk impact: yes — journal identity controls whether protection state is
  reported truthfully, although the access remains read-only.
