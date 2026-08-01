# Branch Test Map: `TestConsolePerformanceCapabilitiesOpenOneProfileDatabaseForBothReadSeams`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | open succeeds | this test | no wiring | yes |
| B2 | both read seams present | this test | no wiring | yes |
| B3 | DB exists in profile | this test | no wiring | yes |
| B4 | evidence read succeeds | this test | no wiring | yes |
| B5 | empty DB is insufficient with digest | this test | no wiring | yes |
| B6 | lifecycle close succeeds | this test | no close | yes |
| B7 | post-close read fails | this test | no close | yes |
| B8 | evidence read before close succeeds | this test | no read-only evidence seam | yes |
| B9 | lifecycle close makes the dashboard seam unusable | this test | no shared close lifecycle | yes |
