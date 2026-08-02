# Function Logic Map: `Console.redirectSettings`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json` (revision: current)
- Change: a055-console-settings-cadence · category: card
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unchanged by this change | as before | as before | as before |

## Branches and early returns

The card standard: every settings form states its current value, what applying it changes and when, a NAMED reason wherever the save surface is missing or risky, and its result beside itself. No new judgement is introduced — every reason already existed on the old screen as a paragraph. Derives the tab and the card from the route that is answering, so a save's result renders next to the form that caused it instead of at the top of a screen with eight forms on it. The route knows which form it is; the notice text does not, and a text-matching rule would go quietly wrong the first time a sentence was reworded. The notice string itself is passed through unchanged.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 419 | no domain mutation; display only | unchanged behaviour | `TestTheNavigationSaysWhatEachScreenAnswers`, `TestNoScreenIsReachableOnlyFromInsideAnother`, `TestTheFourSettingsTabsAreRegisteredGetRoutes`, `TestEachSettingControlAppearsOnExactlyOneTab`, `TestEveryCardEitherSavesOrSaysWhyNot`, `TestASaveResultComesBackToTheFormThatCausedIt`, `TestNoWarningIsHiddenInsideADisclosure`, `TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 4 call sites, unchanged in contract by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- No domain mutation; display only.
- No new broker call, no new config key, no new audit record, no new POST route.

## Safety conclusion

- Safe edit boundary: display, classification and routing.
- Neither this change nor this function touches order submission, stop-loss, take-profit, sizing, the Guardian gate's judgement, the ledger, reconciliation, authentication or fill handling. The Guardian limits change in DISPLAY only: the submit path, the field names, GuardianLimits.Validate, CeilingViolations and the writer are untouched, and the screen states the direction of a change without refusing one. The console's set of state-changing acts does not grow — the four routes this change adds are GET reads with no form.
