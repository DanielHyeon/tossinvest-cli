# Branch Test Map: `Journal.RecordExitJudgement`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing position | judgement validation | yes | pending |
| B2 | invalid provenance | provenance tests | yes | pending |
| B3 | proposal present | arm-before-submit tests | yes | pending |
| B4 | invalid proposal | existing state-only refusal | yes | pending |
| B5 | invalid proposal provenance | provenance mismatch test | yes | pending |
| B6 | zero/nonzero mismatch | provenance mismatch test | yes | pending |
| B7 | unequal tuple | provenance mismatch test | yes | pending |
| B8 | transaction begin error | fault matrix | yes | pending |
| B9 | state read error | fault matrix | yes | pending |
| B10 | completed state | existing completed test | yes | pending |
| B11 | descending high-water | existing monotone test | yes | pending |
| B12 | descending protection | existing monotone test | yes | pending |
| B13 | omitted ratchet level | existing judgement test | yes | pending |
| B14 | state update/fault | atomic rollback test | yes | pending |
| B15 | proposal arm/fault | atomic rollback test | yes | pending |
| B16 | event append/fault | atomic rollback test | yes | pending |
| B17 | commit or duplicate decision | crash/reopen and concurrent exactly-one tests | yes | pending |
| B18 | snapshot candidate supplied | persistence test | yes | yes |
| B19 | snapshot generation mismatch | generation refusal test | yes | yes |
| B20 | duplicate decision lookup hit | duplicate decision test | yes | yes |
| B21 | duplicate lookup SQL failure | atomic fault coverage | yes | yes |
| B22 | legacy high-water guard | existing monotonicity test | yes | yes |
| B23 | legacy baseline guard | existing monotonicity test | yes | yes |
| B24 | saved snapshot available | stale recovery test | yes | yes |
| B25 | recovery selection ambiguity | quarantine test | yes | yes |
| B26 | quarantine insert failure | transaction rollback test | yes | yes |
| B27 | quarantine commit failure | migration/storage fault suite | yes | yes |
| B28 | saved candidate selected | whole saved tuple test | yes | yes |
| B29 | recomputed candidate selected | first evaluation test | yes | yes |
| B30 | effective scalar projection | read-model round trip test | yes | yes |
| B31 | effective JSON encode failure | output digest validation test | yes | yes |
| B32 | state update failure | fault matrix | yes | yes |
| B33 | after-state injected failure | fault matrix | yes | yes |
| B34 | proposal arm branch | fault matrix | yes | yes |
| B35 | after-arm injected failure | fault matrix | yes | yes |
| B36 | evaluation event branch | persistence test | yes | yes |
| B37 | event append failure | fault matrix | yes | yes |
| B38 | after-event injected failure | fault matrix | yes | yes |
| B39 | transaction commit failure | durability suite | yes | yes |
