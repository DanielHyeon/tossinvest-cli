# Branch Test Map: `remoteRuntime.sameOriginForMutation`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | explicit header evidence delegates and is final, including empty/multiple and contradictory values | `TestRemoteSameOriginEvidencePrecedence` | expected failure | yes |
| final | canonical direct TLS+Host passes; non-TLS, wrong Host/port, and forwarded headers fail | `TestHeaderlessCanonicalTLSPostReachesCSRFAndHandlerGates` / `TestRemoteMutationOriginFallbackRejectsIndirectEvidence` | expected failure | yes |
