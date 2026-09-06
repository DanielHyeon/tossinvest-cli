# Branch Test Map: `TestAdapterEnforcesResponseBytesWindowAndRetryAfter`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | an oversized body is refused | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` | existing coverage | PASS |
| B2 | the first call inside the window succeeds | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` | existing coverage | PASS |
| B3 | the second call in the same window is rate limited | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` | existing coverage | PASS |
| B4 | a 429 is retried once and then succeeds | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` | existing coverage | PASS |
| B5 | exactly the advertised Retry-After is waited, once | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` | existing coverage | PASS |
