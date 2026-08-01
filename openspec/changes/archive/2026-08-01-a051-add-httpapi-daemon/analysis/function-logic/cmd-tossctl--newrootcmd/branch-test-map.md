# Branch Test Map: `newRootCmd`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid output format still fails before any daemon handler | existing root/output tests | existing behavior | GREEN in full suite |
| B2 | optional update cache resolves or remains advisory-only | existing update-notice tests | existing behavior | GREEN in full suite |
| B3 | optional config status is read without widening authority | existing config warning tests | existing behavior | GREEN in full suite |
| B4 | onboarding hint is considered only for table output | existing root/onboarding tests | existing behavior | GREEN in full suite |
| B5 | credential path resolves before loading official credentials | existing onboarding tests | existing behavior | GREEN in full suite |
| B6 | credentials are considered present only after a successful non-nil load | existing onboarding tests | existing behavior | GREEN in full suite |
| B7 | interactive first-run hint is emitted only when policy says so | existing onboarding tests | existing behavior | GREEN in full suite |
| Registration | `httpapi` is present once with `source: both`, never `mutating:true` | `TestHTTPAPICommandIsRegisteredReadOnlyAndNonMutating` | RED before registration | GREEN focused/full suite |
| Construction | constructing/helping `httpapi` opens no listener, engine, writer or broker mutation seam | `TestHTTPAPICommandConstructionHasNoRuntimeSideEffects` | RED before handler | GREEN focused/full suite |
