# Branch Test Map: `Console.mutating`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | non-POST is still rejected before all mutation logic | existing `mutating` method tests | inherited | inherited |
| B2 | explicit invalid evidence is rejected before form/CSRF/handler | `TestRemoteSameOriginEvidencePrecedence` | expected failure | yes |
| B3 | malformed form remains a 400 after origin acceptance | existing malformed form tests | inherited | inherited |
| B4 | canonical headerless TLS request with wrong CSRF reaches CSRF rejection | `TestHeaderlessCanonicalTLSPostReachesCSRFAndHandlerGates` | expected failure | yes |
| fallthrough | canonical headerless TLS request with valid CSRF reaches handler | `TestHeaderlessCanonicalTLSPostReachesCSRFAndHandlerGates` | expected failure | yes |
