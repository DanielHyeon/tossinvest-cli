# Branch Test Map: `ReadOnly.checkSchema`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | newer schema refuses before query | `TestOpenReadOnlyRejectsNewerSchema` | existing | pending |
| B2 | released baseline table missing | existing readonly schema tests | existing | pending |
| B3 | released baseline column missing retains specific contract | existing v8/v9/v14 readonly tests | existing | pending |
| B4 | v20 campaign table missing | `TestOpenReadOnlyRejectsDamagedV20CampaignSchema` | pending | pending |
| B5 | v20 campaign column missing | `TestOpenReadOnlyRejectsDamagedV20CampaignSchema` | pending | pending |
| B6 | complete v20 opens read-only | campaign readonly/replay tests | existing | pending |
| B7 | baseline column scan | readonly schema tests | existing | yes |
| B8 | baseline column missing | v8/v9/v14 tests | existing | yes |
| B9 | baseline inspection error | readonly tests | existing | yes |
| B10 | baseline missing summary | readonly tests | existing | yes |
| B11 | v20 version gate | damaged-v20 tests | yes | yes |
| B12 | campaign table scan | damaged-v20 tests | yes | yes |
| B13 | campaign table missing | damaged-v20 tests | yes | yes |
| B14 | campaign table inspection error | damaged-v20 tests | yes | yes |
| B15 | campaign column scan | damaged-v20 tests | yes | yes |
| B16 | campaign column missing | damaged-v20 tests | yes | yes |
| B17 | campaign column inspection error | damaged-v20 tests | yes | yes |
| B18 | v20 missing summary | damaged-v20 tests | yes | yes |
| B19 | schema complete | readonly current tests | existing | yes |
| B20 | older specific error ordering | v14 test | existing | yes |
| B21 | no migration fallback | readonly no-migrate test | existing | yes |
| B22 | return nil | readonly current tests | existing | yes |
