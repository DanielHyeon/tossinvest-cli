# Function Logic Map: `strategyflowEvaluation`

- Source: `internal/weeklyvaluelane/strategyflow_testseam.go`
- AST: `ast.json`, source SHA recorded in the AST.
- Scope: `tossos_testseams` fixture only. Default production build excludes this file.

## Inputs and invariants

| Input/state | Source of truth | Failure behavior |
|---|---|---|
| `ordinal` | tagged fixture caller | outside 1..7 returns `RefusalPlanExhausted` |
| native plan | `BuildCampaignPlan` | return its error |
| reservation progression | `ReservationState` + native `ApplyReservation` | exact next ordinal and unique active reservation required |
| filled/held risk | `RiskState` + `ApplyPositiveFillAtomic` | native refusal, no hand-written state repair |
| market week | `MarketWeekEvidence` | native `ValidateMarketWeek` during reservation |

## Branches and tests

| Branch | Condition | Outcome | Required focused test |
|---|---|---|---|
| B1 | ordinal outside 1..7 | `RefusalPlanExhausted` | existing fixture-boundary test |
| B2 | ordinal 1 | no prefix transition | existing direct fixture consumer |
| B3 | ordinal 2..7 | native prefix reservation/fill transitions | new KR/US k=3..7 integration coverage |
| B4 | prefix reservation/fill refuses | fixture error preserving native code | missing/intermediate reservation negative |
| B5 | final reservation refuses | fixture error preserving native code | current reservation-missing assessment test |

## Calls and state

- `BuildCampaignPlan` creates the original tagged plan only.
- `ApplyReservation` and `ApplyPositiveFillAtomic` are the only prefix-state transitions.
- `mintRiskState` initializes once before any prefix; no loop reset is allowed.
- `strategyflowNextWeek` advances each synthetic week. The updated fixture must also advance evidence timestamps consistently.

## Safety conclusion

Edit only tagged test fixture. No production constructor, FX authority, broker/journal path, or runtime access is introduced.
