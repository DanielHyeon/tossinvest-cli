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
| B17 | reservation terminal | reservation tests |
| B18 | seven-leg plan exhausted | allocation tests |
| B19 | zero quantity | allocation tests |
| B20 | invalid A066 cap | cap tests |
| B21 | invalid effective stop | stop tests |
| B22 | missing/noncanonical entry | exact terms tests |
| B23 | missing/noncanonical staged target | `TestWeeklyMissingExplicitTargetFailsClosed` |
| B24 | invalid entry/stop arithmetic | stop tests |
| B25 | structural stop distance exceeded | stop cap tests |
| B26 | RR refused | RR tests |
| B27 | computed exact terms invalid | exact terms tests |
| B28 | risk admission refusal | risk tests |
| B29 | unsupported currency scale | execution authority tests |
| B30 | final exact provenance composition | six-lane integration test |

Accepted KR/US tests assert the exact fair-value-capped target from `CalculateRR`.
