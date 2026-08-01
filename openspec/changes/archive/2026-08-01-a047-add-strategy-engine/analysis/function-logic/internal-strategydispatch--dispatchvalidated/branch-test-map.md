# Branch Test Map: `dispatchValidated`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | each nil dependency refuses before authority | `TestPostValidationDispatchRejectsEveryMissingDependencyBeforeAuthority` | review found no direct mapping | pass |
| B2 | zero clock and exact expiry refuse stale | `TestPostValidationDispatchPrePlanFailuresAreExactAndCallOfficialGatewayZeroTimes` | review found branch mislabeled | pass |
| B3 | gate read error retains cause | same pre-plan failure table | review found no injected read failure | pass |
| B4 | initial blocker returns stable exact reason | `TestPostValidationDispatchCoreRefusesEveryInitialGateBeforeIssuerWithStablePrecedence` | partial precedence only | pass |
| B5 | manifest verification error stops before issuer/gateway | pre-plan failure table, manifest row | review found no direct mapping | pass |
| B6 | atomic issuer error retains cause and gateway stays zero | pre-plan failure table, issue-and-plan row | failure injection absent | pass |
| B7 | leased binding differs after planning | `TestValidatedDispatchPostPlanAuthorityChangesPersistTOCTOUAndCallOfficialGatewayZeroTimes` | direct drift fixture absent | pass |
| B8 | leased operational blocker changes after planning | same post-plan authority table | direct leased blocker absent | pass |
| B9 | plan manifest digest differs from leased activation | same post-plan authority table | direct digest fixture absent | pass |
| B10 | exact expiry is reached under leases | `TestValidatedDispatchExpiryReachedDuringPlanningRefusesAtExactBoundary` | only initial expiry existed | pass |
| B11 | pre-gateway manifest error maps to TOCTOU | post-plan authority table, revoke/digest rows | direct revoke row absent | pass |
| B12 | successful lease enters terminal classification | `TestValidatedDispatchSuccessfulLeaseOutcomePersistenceBranches` | generic outcome helper was not dispatch evidence | pass |
| B13 | switch handles all three dispositions | same successful-lease terminal table | dispatch arms incomplete | pass |
| B14 | exact confirmed writes dispatched outcome | positive spy test plus successful-lease table | existing positive seam | pass |
| B15 | dispatched persistence failure surfaces plan error | successful-lease table, dispatched failure row | failure injection absent | pass |
| B16 | not-dispatched successful call writes refusal | successful-lease table | classifier-only coverage | pass |
| B17 | successful-lease refusal persistence error is joined | successful-lease table, refusal failure row | failure injection absent | pass |
| B18 | malformed success writes in-doubt | successful-lease table | classifier-only coverage | pass |
| B19 | successful-lease in-doubt persistence error is joined | successful-lease table, in-doubt failure row | failure injection absent | pass |
| B20 | pre-callback error and typed leased TOCTOU write refusal | lease/post-call error table plus authority-drift tests | only one TOCTOU shape existed | pass |
| B21 | TOCTOU refusal persistence error is joined | lease/post-call error table, pre-callback row | failure injection absent | pass |
| B22 | existing typed TOCTOU is returned unchanged | plan-time mutation and leased drift tests | indirect | pass |
| B23 | post-call definitive failure writes refusal | `TestValidatedDispatchPostCallDefinitiveRefusalPersistsAndReturnsGatewayCause` | classifier-only coverage | pass |
| B24 | post-call refusal persistence error is joined | lease/post-call error table, definitive refusal row | failure injection absent | pass |
| B25 | non-nil post-call error cannot classify dispatched | `TestPostCallErrorCannotReachStructurallyUnreachableDispatchedPersistenceBranch` | map claimed reachable | structural-unreachable invariant pass |
| B26 | dispatched persistence under B25 | same invariant | map claimed a nonexistent issuer path | structural-unreachable with B25; no method call possible |
| B27 | post-call ambiguity persists in-doubt; persistence failure joins | lease/post-call error table plus structural invariant test | failure injection absent | pass |
| Invariant | no test uses a real gateway, order command, manifest installer, or toggle | all dispatch tests use package-local spies | required safety boundary | pass |
| Invariant | manifest writer is queued while verified callback owns read lease | `TestManifestLeaseBlocksRevocationAcrossCallback` | writer scheduling was sampled without a start barrier | pass with writer-start barrier |
| Order | activation → lane → kill → protection → reconcile → scheduler → autostart → gate → LIVE | initial gate table covers every adjacent simultaneous pair | partial | pass |
| Success | exact confirmed spy outcome writes one dispatched link | `TestPostValidationDispatchCorePlansOnceAndPersistsExactOfficialOutcomeWithSpies` | existing | pass; package-private seam, not production-positive evidence |
