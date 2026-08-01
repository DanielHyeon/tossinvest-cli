# Branch Test Map: `router.ServeHTTP`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | exact mutation allowlist and method | existing mutation route contract tests | yes | yes |
| B2 | stream delegation | existing SSE tests plus TLS HTTP/2 stream test | yes | yes |
| B3 | unknown resource | existing resource-not-found tests | yes | yes |
| B4 | wrong method | existing method-not-allowed tests | yes | yes |
| B5 | bodyless HTTP/2 accepted; declared/unknown body rejected | new TLS HTTP/2 regression test | yes | yes |
| B6 | query rejected | existing query contract tests | yes | yes |
| B7 | reader unavailable | existing unavailable tests | yes | yes |
| B8 | fixed read succeeds | existing router tests plus new TLS HTTP/2 GET/HEAD | yes | yes |
