# Branch Test Map: `adaptExchangeRate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Exact base/quote, rate/mid-rate and validity strings survive the unit adapter unchanged | `official.TestAdaptExchangeRatePreservesAuthorityFields` | compile failed: raw fields absent | pass |
| HTTP boundary | Exact fields survive the official client envelope and query path unchanged | `official.TestExchangeRatePreservesAuthorityFieldsAcrossHTTPBoundary` | compile failed: raw fields absent | pass |
| display compatibility | Existing `Base`/`Close` floats retain their current values | `official.TestAdaptExchangeRatePreservesAuthorityFields` | compile failed before assertions | pass |
| downstream fail-closed | Invalid decimals, missing/unparseable/reversed validity cannot mint sealed evidence | `officialfx.TestReadOfficialRefusesInvalidAuthority` | compile failed: adapter absent | pass |
