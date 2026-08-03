# Branch Test Map: `evaluate`

| Branch | Scenario | Required test |
|---|---|---|
| B1 | invalid plan | existing plan refusal tests |
| B2 | invalidation requested | existing invalidation tests |
| B3 | invalidation code absent | existing invalidation tests |
| B4 | lane disabled | existing OFF tests |
| B5 | market scope mismatch | existing scope tests |
| B6 | signal refusal | existing evidence tests |
| B7 | effective stop invalid | existing stop tests |
| B8 | invalid leg | allocation tests |
| B9 | terminal leg | allocation tests |
| B10 | risk latch | risk tests |
| B11 | risk-plan mismatch | risk tests |
| B12 | zero planned quantity | allocation tests |
| B13 | invalid A066 cap | cap tests |
| B14 | filled parse failure | arithmetic tests |
| B15 | held parse failure | arithmetic tests |
| B16 | reservation parse failure | arithmetic tests |
| B17 | checked addition failure | arithmetic tests |
| B18 | budget parse failure | arithmetic tests |
| B19 | budget exceeded | risk tests |
| B20 | missing, mutated, or unordered execution terms | `TestContinuationExecutionTermsMissingOrMutatedFailClosed` |

Accepted KR/US tests assert exact caller-supplied entry, effective stop and target.
