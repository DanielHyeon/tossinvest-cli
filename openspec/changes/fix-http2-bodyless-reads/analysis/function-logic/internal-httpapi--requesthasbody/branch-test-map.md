# Branch Test Map: `requestHasBody`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil request is safe | middleware/router construction tests | yes | yes |
| B2 | known and unknown-length bodies are present | `TestHTTP2BodylessReadsAndStreamRespectWireBody` | yes | yes |
| B3 | absent or present preserved header iterates safely | `TestRequestHasBodyUsesPreservedHTTP2ContentLength` | yes | yes |
| B4 | comma-capable header value splitting preserves classification | `TestRequestHasBodyUsesPreservedHTTP2ContentLength` | yes | yes |
| B5 | every rune in a present header part is inspected | `TestRequestHasBodyUsesPreservedHTTP2ContentLength` | yes | yes |
| B6 | malformed/signed-zero header rejects while digit-only zero continues | `TestRequestHasBodyUsesPreservedHTTP2ContentLength` | yes | yes |
| B7 | positive header rejects while parsed zero remains bodyless | `TestRequestHasBodyUsesPreservedHTTP2ContentLength` | yes | yes |
