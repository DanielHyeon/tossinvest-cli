# Branch Test Map: `TestSchemaTablesAndColumns`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | table query error | this test | existing | yes |
| B2 | table row loop | this test | existing | yes |
| B3 | table scan error | this test | existing | yes |
| B4 | table rows error | this test | existing | yes |
| B5 | joined table set differs from the golden table list (`if` at 180:2) | this test | yes | yes |
| B6 | iterate every golden table's expected columns (`range` at 390:2) | this test | yes | yes |
| B7 | column set mismatch (`if` at 393:3) | this test | yes | yes |

Rows B5/B6 were re-described in Wave 2A (2026-09-25) from the AST source lines; the earlier text
("expected table loop" / "pragma query/scan") did not match the branch at those positions.
GREEN re-measured: `go test -count=1 ./internal/journal/...` PASS at HEAD d72bc401 (495.5s).
