# Branch Test Map: `ValidateRecoveryDerivation`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | malformed identity/definition | exact derivation and forgery tests | yes | yes |
| B2 | valid ratchet rederivation | existing ratchet persistence tests | yes | yes |
| B3 | LADDER before first rung (`-1`) | `TestRecoveryAllowsLadderBeforeFirstRung` | yes | yes |
| B4 | rung below `-1` or beyond table | `TestRecoveryRefusesInvalidLadderRungBounds` | yes | yes |
| B5 | next-line or projection mismatch | digest/semantic output forgery tests | yes | yes |
| B6 | remaining quantity validation | semantic output tests | yes | yes |
| B7 | select ratchet vs ladder arm | ratchet and ladder recovery suites | yes | yes |
| B8 | ratchet identity resolution | exact policy derivation tests | yes | yes |
| B9 | ratchet output validation | persistence/forgery tests | yes | yes |
| B10 | ratchet risk/high/breakeven parsing | decimal validation tests | yes | yes |
| B11 | ratchet next-line derivation error | forged derivation tests | yes | yes |
| B12 | ladder definition validation | ladder policy validation tests | yes | yes |
| B13 | ladder identity resolution | exact ladder identity tests | yes | yes |
| B14 | ladder output/rung validation | rung-bound and semantic output tables | yes | yes |
| B15 | ladder next-line derivation error | pre-first/bounds tests | yes | yes |
| B16 | derived target/protection equality | forged next-line test | yes | yes |
