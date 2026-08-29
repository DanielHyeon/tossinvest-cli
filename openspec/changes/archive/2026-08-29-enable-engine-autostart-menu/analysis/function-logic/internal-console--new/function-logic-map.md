# Function Logic Map: `New`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Options.StartVerify` | non-nil | cmd wiring/test harness | nil returns `ErrNoVerifyWiring` |
| `Options.Now` | nil or clock | caller | nil uses UTC system clock |
| `Options.Remote` | empty local or complete remote contract | CLI validation | invalid remote config returns error |
| optional seams | nil or least-capability interfaces | cmd wiring | nil renders explicit unavailable state |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | StartVerify nil | none | constructor refusal | existing constructor test |
| B2 | Now nil | install system clock | continue | existing clock tests |
| B3 | remote runtime error | none | return error | remote validation tests |
| B4 | Out nil | use `io.Discard` | continue | existing constructor tests |
| B5 | Binary nil | use `binstamp.Self` | continue | existing update tests |

The change copies an optional initial engine note and stamps it only when non-empty.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newRemoteRuntime` | validate remote security contract | fail constructor on error | AST |
| `Binary` | fingerprint running executable | error becomes unknown stamp | AST |
| cache constructors | bound broker reads | no eager broker call | AST |
| `routes` | freeze HTTP surface | constructor completes only after handler exists | AST |

## State mutations and fallbacks

- Mints session and CSRF from crypto randomness.
- Does not start an engine or touch config/account.
- Initial note is display-only state supplied by `runConsole`.

## Safety conclusion

- Safe edit boundary: initialize two existing engine-note fields from one new option.
- High-risk impact: yes — must never treat a note as authority or trigger.
