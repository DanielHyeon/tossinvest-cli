# Branch Test Map: `remoteRuntime.sameOrigin`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid configured origin refuses | startup validation + `networkboundary.TestCanonicalOriginUsesSchemeHostAndEffectivePort` | retained defensive path | yes |
| delegated | Origin presence is final; empty/repeated/wrong/opaque Origin is not rescued by Referer | `TestRemoteSameOriginEvidencePrecedence`, `networkboundary.TestOriginPrecedenceRejectsMalformedExplicitEvidence` | retained + new RED observed | yes |
| delegated | missing/empty/repeated/malformed Referer refuses; login has no headerless fallback | `TestRemoteSameOriginEvidencePrecedence`, `TestHeaderlessRemoteLoginRemainsStrict` | retained | yes |
| final | Referer path/query/fragment ignored; different port refused | `networkboundary.TestOriginIgnoresRefererPathAndRejectsDifferentPort` + console regression | new RED observed | yes |
