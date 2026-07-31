# Branch Test Map: `ExitObserver.workingSet`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | position/state query fails | existing journal error tests | existing | yes |
| B2 | closed/zero position is skipped | existing projection tests | existing | yes |
| B3 | ineligible position alerts | existing unmanaged tests | existing | yes |
| B4 | missing state opens | existing opening tests | existing | yes |
| B5 | open failure retained on cycle | existing opening failure tests | existing | yes |
| B6 | completed duplicate is skipped | existing completed-state tests | existing | yes |
| B7 | opened state increments count | existing cycle tests | existing | yes |
| B8 | legacy default/common ID resolves fixed identity | policy identity integration tests | yes | yes |
| B9 | unknown meaning is carried as refusal | policy digest conflict tests | yes | yes |
| B10 | managed position preserves identity error per position | reinterpretation refusal test | yes | yes |
