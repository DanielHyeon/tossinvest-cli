# Function Logic Map: `newUpdateCmd`

- Source: `cmd/tossctl/update.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| root output format | table/json/csv | persistent CLI option | parse error stops before network |
| installed executable | resolvable non-dev install | `os.Executable`, symlink resolution, `selfupdate.DetectInstallMethod` | dev builds fail closed |
| latest version | semantic release string or empty | `updatecheck.Checker` | no update is reported |
| install approval | interactive confirmation or explicit `--yes` | Cobra flag + TTY | cancel returns without mutation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid format/path/dev build | none | error | existing update command tests |
| B2 | `--check` | release metadata read only | formatted result | check-mode tests |
| B3 | no newer release | console output only | nil | existing up-to-date test |
| B4 | operator cancels | none | nil | confirmation test |
| B5 | legacy update proceeds | invokes checksum-only self updater | updater result | existing self-update tests |
| B6 | any non-check invocation | prints deprecation/signed-settings warning before possible mutation | continues through existing flow | `TestLegacyUpdateWarnsBeforeMutation` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `updatecheck.LatestStable` | fixed release discovery | cache-backed; empty means unavailable | CodeGraph + AST |
| `selfupdate.Run` | legacy installer | existing platform-specific contract | CodeGraph + AST |
| `writeUpdateCheckResult` | read-only output | propagates encoder error | CodeGraph + AST |

## State mutations and fallbacks

- Only the additional warning writes output. No confirmation, version selection,
  installer call, or account state changes.

## Safety conclusion

- Safe edit boundary: emit an honest deprecation warning only on non-check runs.
- High-risk impact: yes; this command can replace the installed binary, so its
  control flow remains otherwise unchanged.
