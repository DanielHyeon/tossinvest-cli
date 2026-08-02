# Branch Test Map: `attachPositionExitLines`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty and mixed rows | console focused/full suite | pre-existing | yes |
| B2 | desired-designated US row with commander unavailable | `TestPositionsShowRuntimeUnknownWhenCommanderUnavailableButDesiredIncludesUS` | yes | yes |
| B3 | no exit + pending/blocked/unmanaged/runtime unknown | candidate/runtime matrix | yes | yes |
| B4 | cross-lifecycle raw/canonical evidence | `TestPositionsSuppressCrossLifecycleExitEvidence` | yes | yes |
| B5 | released lifecycle with raw/canonical evidence | released trading-view regressions | pre-existing | yes |
| B6 | released valid versus corrupt reference | released canonical/corrupt matrices | yes | yes |
| B7 | fresh and stale canonical | `TestPositionsRenderCanonicalExitLineFixtures` | pre-existing | yes |
| B8 | lifecycle-unverified canonical/raw | unsafe-evidence canonical matrix | yes | yes |
| B9 | raw allowlisted versus corrupt | legacy and unsafe-evidence matrices | yes | yes |
