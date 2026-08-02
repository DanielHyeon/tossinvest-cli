# Branch Test Map: `Console.handlePositions`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | policy seam present/absent | console focused/full suite | pre-existing + a053 | yes |
| B2 | lifecycle list success/failure | runtime/lifecycle tests | pre-existing + a053 | yes |
| B3 | zero/multiple lifecycle rows | console focused suite | pre-existing | yes |
| B4 | settings seam present/absent | settings/positions tests | pre-existing | yes |
| B5 | desired load success/failure | desired/effective tests | pre-existing | yes |
| B6 | KR/US designation stamping | a053 market matrix | yes | yes |
| B7 | runtime attempted/unavailable | runtime-unknown projector matrix | yes | yes |
| B8 | mixed rows projected | a053 KR/US matrix | yes | yes |
| B9 | journal row and broker-only row | a053 legacy/candidate | yes | yes |
| B10 | lifecycle missing | legacy generation 1 fixture | pre-existing | yes |
| B11 | managed/released lifecycle and generation | released + generation tests | yes | yes |
| B12 | block present/absent | US blocked/pending tests | yes | yes |
