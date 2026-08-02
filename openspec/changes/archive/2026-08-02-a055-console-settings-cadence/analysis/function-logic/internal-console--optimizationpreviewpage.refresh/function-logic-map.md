# Function Logic Map: `optimizationPreviewPage.Refresh`

- Source: `internal/console/optimization_view.go`
- AST evidence: `ast.json` (revision: base)
- Change: a055-console-settings-cadence · category: shell
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

This function belongs to a054-console-status-shell, the shared status shell. It is in a055's diff because the two changes were committed together in 01a4caa1 — they share about fifteen console files and neither half compiles alone — not because a055 edited it. The Function Logic Map of record is a054's, in `openspec/changes/archive/2026-08-02-a054-console-status-shell/analysis/function-logic/`. That classification is read off the archive: a054 has an artifact directory for this exact source and function. a054 removed it: chrome carries Refresh as a field, so the screens that used to declare their own method no longer need one. The evidence below is the base revision, which is the last one in which this function existed.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the one path through this function | none by a055 | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 0 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- None by a055.
- No new broker call, no new config key, no new audit record, no new POST route.

## Safety conclusion

- Safe edit boundary: display, classification and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate's judgement, the ledger, reconciliation, authentication or fill handling. The Guardian limits change in DISPLAY only: the submit path, the field names, GuardianLimits.Validate, CeilingViolations and the writer are untouched, and the screen states the direction of a change without refusing one. The console's set of state-changing acts does not grow — the four routes this change adds are GET reads with no form.
