# Branch Test Map: `evaluate`

| Branch | Scenario | Required test |
|---|---|---|
| B1 | disabled lane | existing OFF tests |
| B2 | structural invalidation | existing invalidation tests |
| B3 | plan or scope mismatch | plan tests |
| B4 | stale FX | FX tests |
| B5 | metric refusal without code | schema tests |
| B6 | typed metric refusal | schema tests |
| B7 | risk latch | risk tests |
| B8 | invalid or retreating stop | stop tests |
| B9 | invalid leg | allocation tests |
| B10 | invalid cap | cap tests |
| B11 | final-leg structure refusal | structure tests |
| B12 | zero quantity | allocation tests |
| B13 | reservation mismatch | cap tests |
| B14 | risk admission refusal | risk tests |
| B15 | missing, mutated, or unordered execution terms | `TestReversalExecutionTermsMissingOrMutatedFailClosed` |
| B16 | first leg action | accepted exact-terms test |
| B17 | later leg action | allocation tests |

Accepted KR/US tests assert exact explicit entry, stop and target.
