# Branch Test Map: `LimitRequestBody`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil handler stays fail-closed | existing `LimitRequestBody(nil)` coverage/static contract | yes | yes |
| B2 | declared/unknown body remains bounded | `TestLimitRequestBodyAcceptsBoundaryAndRejectsOneByteOver`; HTTP/2 body tests | yes | yes |
| B3 | HTTP/2 known-empty body is not wrapped as an actual body | new TLS HTTP/2 GET/HEAD test | yes | yes |
