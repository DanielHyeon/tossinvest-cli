# Branch Test Map: `remoteRuntime.sameOriginForMutation`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | explicit Origin/Referer evidence is final including empty/repeated/contradictory values | `TestRemoteSameOriginEvidencePrecedence` | retained | yes |
| final | headerless canonical direct TLS passes; plaintext/wrong host/forwarded evidence fails | `TestHeaderlessCanonicalTLSPostReachesCSRFAndHandlerGates`, `TestRemoteMutationOriginFallbackRejectsIndirectEvidence` | retained | yes |
