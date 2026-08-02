# Branch Test Map: `router.serveStream`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | wrong method rejected | existing stream method tests | yes | yes |
| B2 | bodyless HTTP/2 accepted; declared/unknown body rejected | new TLS HTTP/2 regression test | yes | yes |
| B3 | query rejected | existing stream query tests | yes | yes |
| B4 | nil stream unavailable | existing stream unavailable test | yes | yes |
| B5 | fixed bodyless stream delegated | existing SSE tests plus TLS HTTP/2 stub | yes | yes |
