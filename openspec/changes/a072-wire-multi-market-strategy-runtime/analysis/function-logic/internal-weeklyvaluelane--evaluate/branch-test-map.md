# Branch Test Map: `evaluate`

| Branch | Scenario | Required test |
|---|---|---|
| B1 | planned ordinal in range | plan tests |
| B2 | FX lineage present | FX tests |
| B3 | authorization invalid | OFF test |
| B4 | candidate identity invalid | schema tests |
| B5 | plan invalid | plan tests |
| B6 | market mismatch | scope tests |
| B7 | source mismatch | source tests |
| B8 | invalidation requested | invalidation tests |
| B9 | invalidation code missing | invalidation tests |
| B10 | evidence rejected | evidence tests |
| B11 | evidence scope/config mismatch | config tests |
| B12 | week market mismatch | calendar tests |
| B13 | calendar refusal | calendar tests |
| B14 | invalid leg | allocation tests |
| B15 | reservation state invalid | reservation tests |
| B16 | reservation missing | reservation tests |
| B17 | reservation identity/ordinal mismatch | reservation tests |
| B18 | reservation terminal | reservation tests |
| B19 | seven-leg plan exhausted | allocation tests |
| B20 | zero quantity | allocation tests |
| B21 | invalid A066 cap | cap tests |
| B22 | public saved scalar has no valid private authority | forged-authority subcase in `TestWeeklySavedStopNeverInheritsCandidateProvenance` |
| B23 | invalid effective stop | stop tests |
| B24 | selected saved stop lacks valid private authority | `TestWeeklySavedStopNeverInheritsCandidateProvenance` |
| B25 | missing/noncanonical entry | exact terms tests |
| B26 | missing/noncanonical staged target | `TestWeeklyMissingExplicitTargetFailsClosed` |
| B27 | invalid entry/stop arithmetic | stop tests |
| B28 | structural stop distance exceeded | stop cap tests |
| B29 | RR refused | RR tests |
| B30 | computed exact terms invalid | exact terms tests |
| B31 | risk admission refusal | risk tests |
| B32 | unsupported currency scale | execution authority tests |
| B33 | saved stop replaces candidate provenance | `TestWeeklySavedStopNeverInheritsCandidateProvenance` |

`TestWeeklyPublicSavedStopScalarCannotRetreatSealedAuthority` covers empty, candidate-equal and below-candidate scalar mutations while preserving sealed stop 100.
