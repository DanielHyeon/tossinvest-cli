# Function Logic Map: `positionsPage.BrokerState`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json` (revision: current)
- Change: a055-console-settings-cadence · category: fold
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

A screen that reloads itself cannot keep a native <details> open, so its explanatory disclosures are driven by a URL parameter that the reload carries with it. The parameter is display-only and reaches no judgement, save or audit record. New, for the same reason ordersPage.Screen exists: the broker-state banner is invoked with the holdings snapshot as its dot and its fold needs the page's state.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the one path through this function | no domain mutation; a GET render | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 0 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- No domain mutation; a get render.
- No new broker call, no new config key, no new audit record, no new POST route.

## Safety conclusion

- Safe edit boundary: display, classification and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate's judgement, the ledger, reconciliation, authentication or fill handling. The Guardian limits change in DISPLAY only: the submit path, the field names, GuardianLimits.Validate, CeilingViolations and the writer are untouched, and the screen states the direction of a change without refusing one. The console's set of state-changing acts does not grow — the four routes this change adds are GET reads with no form.
