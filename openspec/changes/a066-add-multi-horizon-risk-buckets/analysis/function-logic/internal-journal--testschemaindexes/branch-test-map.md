# Branch Test Map: `TestSchemaIndexes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | query error | this test | existing | yes |
| B2 | row iteration | this test | existing | yes |
| B3 | scan error | this test | existing | yes |
| B4 | rows error | this test | existing | yes |
| B5 | expected iteration | this test | yes | yes |
| B6 | missing index | this test | yes | yes |

Wave 2A (2026-09-25): AST re-extracted at HEAD (644–715); old/new branch alignment identical B1–B6.
GREEN re-measured: `go test -count=1 ./internal/journal/...` PASS at HEAD d72bc401 (495.5s).
