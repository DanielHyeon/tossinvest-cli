# Branch Test Map: `Tracker.syncGate`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | nil gate | direct tracker tests | preserve | yes |
| B2 | active blocks are enumerated for account projection | A110 permanent/blank-symbol tests | preserve | yes |
| B3 | permanent or blank-symbol projects account-wide | A110 permanent/blank-symbol tests | yes | yes |
| B4 | active blocks are enumerated for symbol projection | changing-symbol incident | preserve | yes |
| B5 | real symbol stays narrow | changing-symbol incident | preserve | yes |
| B6 | existing symbol gates are enumerated | existing gate ownership suite | preserve | yes |
| B7 | foreign reason remains untouched | existing gate ownership suite | preserve | yes |
| B8 | absent reconcile symbol reason clears | existing gate ownership suite | preserve | yes |
| B9 | account reconcile reasons are enumerated | commit ambiguity and operator tests | preserve | yes |
| B10 | present permanent/ordinary reason blocks | commit ambiguity and operator tests | yes | yes |
| B11 | absent account reason clears independently | operator tests | preserve | yes |
