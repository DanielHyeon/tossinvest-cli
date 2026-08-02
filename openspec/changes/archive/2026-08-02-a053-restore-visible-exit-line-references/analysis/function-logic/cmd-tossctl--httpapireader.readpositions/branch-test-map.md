# Branch Test Map: `httpAPIReader.readPositions`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | required seams absent | production reader suite | pre-existing | yes |
| B2 | account read failure | production reader suite | pre-existing | yes |
| B3 | holdings failure | production reader suite | pre-existing | yes |
| B4 | journal readable/unavailable | reader management tests | pre-existing | yes |
| B5 | live-position read failure | reader failure tests | pre-existing | yes |
| B6 | lifecycle read failure | reader lifecycle tests | pre-existing | yes |
| B7 | lifecycle map rows | released/generation tests | yes | yes |
| B8 | journal map rows | stored evidence tests | pre-existing | yes |
| B9 | broker rows | cache/join tests | pre-existing | yes |
| B10 | matched stored row | legacy reference and generation tests | yes | yes |
| B11 | unmatched broker row | US adoption plan test | yes | yes |
| B12 | matched released row | released projection test | pre-existing | yes |
| B13 | journal unavailable reason | runtime failure tests | pre-existing | yes |
| B14 | remaining journal rows | stored evidence tests | pre-existing | yes |
| B15 | duplicate broker+journal key | join tests | pre-existing | yes |
| B16 | journal-only released row | released projection test | pre-existing | yes |
