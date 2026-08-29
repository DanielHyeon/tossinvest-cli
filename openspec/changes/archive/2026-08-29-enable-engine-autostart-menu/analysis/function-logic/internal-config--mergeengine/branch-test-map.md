# Branch Test Map: `mergeEngine`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | absent engine block remains all-off | `TestAutostartDefaultsOffWhenEngineIsMissing` | no | no |
| B2 | present/absent autostart is copied/defaulted | `TestAutostartLoadExplicitAndMissing` | no | no |
| B3 | existing exit-policy merge does not change | existing exit-policy config tests | yes | yes |
| B4 | absent automation gate remains off | existing gate default tests | yes | yes |
| B5 | explicit gate enabled remains copied | existing gate parsing tests | yes | yes |
