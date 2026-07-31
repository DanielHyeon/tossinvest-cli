# Branch Test Map: `remoteRuntime.sameOrigin`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Origin-key presence is final; canonical Origin ignores malformed/cross-origin Referer and invalid Origin is not rescued | `TestRemoteSameOriginEvidencePrecedence` | expected failure | yes |
| B2 | empty, whitespace, or multiple Origin is rejected | `TestRemoteSameOriginEvidencePrecedence` | expected failure | yes |
| B3 | absent/multiple Referer and headerless `/login` reject without mutation | `TestRemoteSameOriginEvidencePrecedence` / `TestHeaderlessRemoteLoginRemainsStrict` | expected failure | yes |
| B4 | empty or whitespace Referer is rejected | `TestRemoteSameOriginEvidencePrecedence` | retained | yes |
| B5 | malformed/relative/cross-origin Referer rejects while canonical paths pass | `TestRemoteSameOriginEvidencePrecedence` | retained | yes |
| SM-final | headerless direct TLS canonical Host reaches the following CSRF/handler gates | `TestHeaderlessCanonicalTLSPostReachesCSRFAndHandlerGates` | expected failure | yes |
| SM-final | non-TLS, wrong Host/port, and `X-Forwarded-Host`/`X-Forwarded-Proto` evidence are rejected | `TestRemoteMutationOriginFallbackRejectsIndirectEvidence` | retained | yes |
