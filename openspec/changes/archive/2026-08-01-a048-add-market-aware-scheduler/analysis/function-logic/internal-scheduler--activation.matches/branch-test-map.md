# Branch Test Map: `Activation.matches`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil activation refuses | `TestDecisionStatesAreTypedAndOrderedFailClosed/not activated` | existing | yes |
| predicate false | revision/config/calendar/approval mismatch refuses | `TestActivationIsOpaqueAndBoundToExactDesiredState`, changed-calendar test | yes for current calendar digest | yes |
| predicate true | exact restored capability proceeds to later gates | decision allowed/holiday tests | existing | yes |
