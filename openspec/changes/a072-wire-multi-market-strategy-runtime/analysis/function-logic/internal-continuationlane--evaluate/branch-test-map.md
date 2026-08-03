# Branch Test Map: `evaluate`

| Branch | Scenario | Required test |
|---|---|---|
| B1 | invalid plan | existing plan refusal tests |
| B2 | invalidation requested | existing invalidation tests |
| B3 | invalidation code absent | existing invalidation tests |
| B4 | lane disabled | existing OFF tests |
| B5 | market scope mismatch | existing scope tests |
| B6 | signal refusal | existing evidence tests |
| B7 | public saved scalar has no valid private authority | `TestCallerForgedSavedStopProvenanceFailsClosed` |
| B8 | effective stop invalid | existing stop tests |
| B9 | invalid leg | allocation tests |
| B10 | terminal leg | allocation tests |
| B11 | risk latch | risk tests |
| B12 | risk-plan mismatch | risk tests |
| B13 | zero planned quantity | allocation tests |
| B14 | invalid A066 cap | cap tests |
| B15 | filled parse failure | arithmetic tests |
| B16 | held parse failure | arithmetic tests |
| B17 | reservation parse failure | arithmetic tests |
| B18 | checked addition failure | arithmetic tests |
| B19 | budget parse failure | arithmetic tests |
| B20 | budget exceeded | risk tests |
| B21 | missing, mutated, or unordered execution terms | `TestContinuationExecutionTermsMissingOrMutatedFailClosed` |

`TestPublicSavedStopScalarCannotRetreatSealedAuthority` covers empty, candidate-equal and below-candidate scalar mutations while preserving sealed stop 100.
