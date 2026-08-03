# Branch Test Map: `evaluateWith`

| Branch | Scenario | Required test |
|---|---|---|
| B1 | invalid approved candidate | existing candidate refusal |
| B2 | candidate/router scope mismatch | existing scope refusal |
| B3 | router typed refusal | existing router refusal |
| B4 | malformed route decision | lineage mismatch |
| B5 | unsupported descriptor | unsupported binding |
| B6 | wrong tagged lane input | unsupported binding |
| B7 | registry binding absent | unsupported binding |
| B8 | native lane refusal | preserve native code |
| B9 | lane lineage mismatch | incomplete sealed lineage |
| B10 | existing owner campaign mismatch | lineage mismatch |
| B11 | campaign/leg/risk incomplete | lineage incomplete |
| B12 | accepted lane has missing/malformed terms | execution terms invalid |

Post-edit execution-term validation adds tests for missing, malformed, tampered and complete sealed terms across all six bindings.
