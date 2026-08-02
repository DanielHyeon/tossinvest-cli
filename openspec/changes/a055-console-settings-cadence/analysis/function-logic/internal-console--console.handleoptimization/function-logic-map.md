# Function Logic Map: `Console.handleOptimization`

- Source: `internal/console/optimization.go`
- AST evidence: `ast.json` (revision: current)
- Change: a055-console-settings-cadence · category: shell
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

This function belongs to a054-console-status-shell, the shared status shell. It is in a055's diff because the two changes were committed together in 01a4caa1 — they share about fifteen console files and neither half compiles alone — not because a055 edited it. The Function Logic Map of record is a054's, in `openspec/changes/archive/2026-08-02-a054-console-status-shell/analysis/function-logic/`. That classification is read off the archive: a054 has an artifact directory for this exact source and function.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 66 | none by a055 | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B2 | if at line 73 | none by a055 | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B3 | if at line 80 | none by a055 | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B4 | if at line 85 | none by a055 | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B5 | if at line 87 | none by a055 | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B6 | else at line 89 | none by a055 | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B7 | if at line 93 | none by a055 | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B8 | if at line 95 | none by a055 | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B9 | else at line 97 | none by a055 | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B10 | range at line 106 | none by a055 | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B11 | if at line 108 | none by a055 | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B12 | if at line 115 | none by a055 | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B13 | if at line 118 | none by a055 | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B14 | else at line 120 | none by a055 | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 28 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- None by a055.
- No new broker call, no new config key, no new audit record, no new POST route.

## Safety conclusion

- Safe edit boundary: display, classification and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate's judgement, the ledger, reconciliation, authentication or fill handling. The Guardian limits change in DISPLAY only: the submit path, the field names, GuardianLimits.Validate, CeilingViolations and the writer are untouched, and the screen states the direction of a change without refusing one. The console's set of state-changing acts does not grow — the four routes this change adds are GET reads with no form.
