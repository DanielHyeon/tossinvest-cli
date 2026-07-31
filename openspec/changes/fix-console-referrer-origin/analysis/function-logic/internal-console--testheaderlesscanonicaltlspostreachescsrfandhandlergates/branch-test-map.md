# Branch Test Map: `TestHeaderlessCanonicalTLSPostReachesCSRFAndHandlerGates`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | valid headerless canonical request returns 204 | this function | unchanged | passed |
| B2 | valid request invokes handler exactly once | this function | unchanged | passed |
| B3 | wrong CSRF is refused with CSRF-specific text | this function | unchanged | passed |
| B4 | wrong CSRF does not invoke handler | this function | unchanged | passed |
