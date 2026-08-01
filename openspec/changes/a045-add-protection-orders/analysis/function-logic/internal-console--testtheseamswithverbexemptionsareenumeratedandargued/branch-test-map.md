# Branch Test Map: `TestTheSeamsWithVerbExemptionsAreEnumeratedAndArgued`
| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | enumerate option seams | self/static suite | Protections absent | pass |
| B2 | reject missing classification | self/static suite | Protections absent | pass |
| B3 | reject weak rationale | self/static suite | rationale absent | pass |
| B4 | enumerate route declarations | self/static suite | routes absent | pass |
| B5 | reject missing route metadata | self/static suite | routes absent | pass |
| B6 | reject duplicate route metadata | self/static suite | existing coverage | pass |
| B7 | enumerate verb exemptions | self/static suite | existing coverage | pass |
| B8 | reject unexplained exemption | self/static suite | existing coverage | pass |
| B9 | enumerate capability references | self/static suite | Protections absent | pass |
| B10 | reject unknown capability | self/static suite | existing coverage | pass |
| B11 | enumerate mutating routes | self/static suite | protection routes absent | pass |
| B12 | compare classifications | self/static suite | protection routes absent | pass |
| B13 | reject read/mutation mismatch | self/static suite | existing coverage | pass |
| B14 | reject missing CSRF coverage | self/static suite | preview/apply absent | pass |
| B15 | enumerate known seams | self/static suite | Protections absent | pass |
| B16 | reject undeclared seam | self/static suite | existing coverage | pass |
| B17 | reject unused declaration | self/static suite | existing coverage | pass |
