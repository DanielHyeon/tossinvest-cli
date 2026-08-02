# Function Logic Map: `Console.handleSettings`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json` (revision: base)
- Change: a055-console-settings-cadence · category: tab
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

The settings screen became four sub-screens classified by whether a change can be undone and how often it is made. What is read, and in what order, did not change. Was the screen; is now the redirect to the default tab. Every read it used to do moved to settingsView byte for byte — the seam checks, the three Loads, the gate and autostart reads, the update inspection. It also answers ?section=<anchor>, which is the half of an old /settings#adoption link a server can see: the fragment itself never reaches one.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 147 | no domain mutation; a GET render or a redirect | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B2 | if at line 150 | no domain mutation; a GET render or a redirect | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B3 | if at line 155 | no domain mutation; a GET render or a redirect | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B4 | if at line 158 | no domain mutation; a GET render or a redirect | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B5 | if at line 163 | no domain mutation; a GET render or a redirect | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B6 | if at line 166 | no domain mutation; a GET render or a redirect | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B7 | if at line 175 | no domain mutation; a GET render or a redirect | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B8 | if at line 178 | no domain mutation; a GET render or a redirect | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |
| B9 | if at line 184 | no domain mutation; a GET render or a redirect | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 15 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- No domain mutation; a get render or a redirect.
- No new broker call, no new config key, no new audit record, no new POST route.

## Safety conclusion

- Safe edit boundary: display, classification and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate's judgement, the ledger, reconciliation, authentication or fill handling. The Guardian limits change in DISPLAY only: the submit path, the field names, GuardianLimits.Validate, CeilingViolations and the writer are untouched, and the screen states the direction of a change without refusing one. The console's set of state-changing acts does not grow — the four routes this change adds are GET reads with no form.
