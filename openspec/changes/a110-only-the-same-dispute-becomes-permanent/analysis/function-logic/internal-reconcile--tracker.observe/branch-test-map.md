# Branch Test Map: `Tracker.Observe`

| Branch | Scenario | Test/evidence | RED | GREEN |
|---|---|---|---|---|
| B1 | first observation initializes blocks | ordinary mismatch suite | preserve | yes |
| B2 | clean path selected | A110 clean/reset | yes | yes |
| B3 | blocking path selected | A110 identity tables | yes | yes |
| B4 | clean resolves pending permanent authority | commit-timeout/read-error cases | yes | yes |
| B5 | clean inspects active blocks | a083 credit suites | preserve | yes |
| B6 | foreign cause/release owner is skipped | cause ownership suites | preserve | yes |
| B7 | absent/stale adjustment credit awaits | a083 time suites | preserve | yes |
| B8 | still-disputed symbol cannot release | a083b reclassification suite | preserve | yes |
| B9 | still-disputed symbol cannot release | a083b reclassification suite | preserve | yes |
| B10 | pending earning key disappeared | clean/different/same-key cases | yes | yes |
| B11 | current ordinary blocks are enumerated before authority read | `TestA110AuthorityOutageStillProjectsCurrentDifferentOrdinaryMismatch` | yes (M25) | yes |
| B12 | unseen current block is inserted pending | same F9 test | yes | yes |
| B13 | authority-read error returns only after current gate projection | same F9 test | yes | yes |
| B14 | normal path advances streaks | identity tables | preserve | yes |
| B15 | non-prelatched current blocks require insertion | ordinary latch tests | preserve | yes |
| B16 | current ordinary blocks enumerated | changing-symbol/invalid tables | yes | yes |
| B17 | unseen ordinary block added pending | pre-I/O latch tests | preserve | yes |
| B18 | deterministic exact key reaches threshold | exact identity/handoff tests | yes | yes |
| B19 | account permanent key added once | durable threshold test | preserve | yes |
| B20 | confirmed durable additions enumerated | retry tests | preserve | yes |
| B21 | durable permanent clears pending identity | same-key retry tests | yes | yes |
| B22 | authoritative conflicts enumerated | cause-conflict regression | preserve | yes |
| B23 | account authority clears pending identity only | conflict regression | preserve | yes |
| B24 | confirmed releases delete exact block | release failure/partial suites | preserve | yes |
| B25 | committed releases indexed | a083 credit suites | preserve | yes |
| B26 | credits enumerated independently | a083 multi-symbol suite | preserve | yes |
| B27 | non-later comparison preserves credit | a083 time-order suite | preserve | yes |
| B28 | refuted/spent/orphan credit removed | a083b lifetime suite | preserve | yes |

Restore/Refresh acceptance: transient streaks never reconstruct; already durable permanent rows do.
