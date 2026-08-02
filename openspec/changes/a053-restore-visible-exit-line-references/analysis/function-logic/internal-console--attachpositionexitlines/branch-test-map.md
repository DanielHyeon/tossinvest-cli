# Branch Test Map: `attachPositionExitLines`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty and mixed rows | console focused/full suite | pre-existing | yes |
| B2 | no exit + pending/blocked/unmanaged/runtime unknown | candidate/runtime matrix | yes | yes |
| B3 | cross-lifecycle raw/canonical evidence | `TestPositionsSuppressCrossLifecycleExitEvidence` | yes | yes |
| B4 | released lifecycle with raw/canonical evidence | released trading-view regressions | pre-existing | yes |
| B5 | released valid versus corrupt reference | released canonical/corrupt matrices | yes | yes |
| B6 | fresh and stale canonical | `TestPositionsRenderCanonicalExitLineFixtures` | pre-existing | yes |
| B7 | lifecycle-unverified canonical/raw | unsafe-evidence canonical matrix | yes | yes |
| B8 | raw allowlisted versus corrupt | legacy and unsafe-evidence matrices | yes | yes |
