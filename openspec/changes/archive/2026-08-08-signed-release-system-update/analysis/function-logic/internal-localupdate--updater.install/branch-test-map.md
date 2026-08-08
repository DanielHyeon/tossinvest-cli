# Branch Test Map: `Updater.Install`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Unsupported replacement platform refuses | existing unsupported-platform test | baseline | pass |
| B2 | Empty reviewed SHA refuses before candidate preparation | existing empty-review test | baseline | pass |
| B3 | Candidate preparation or validation fails | existing invalid-candidate tests | baseline | pass |
| B4 | Prepared candidate digest differs from the reviewed digest | existing candidate-changed test | baseline | pass |
| B5 | Current executable cannot be opened descriptor-safely | existing current-changed test | baseline | pass |
| B6 | Current executable metadata/build validation fails | existing current-invalid test | baseline | pass |
| B7 | Current bytes differ from the process-start fingerprint | existing current-replaced test | baseline | pass |
| B8 | Optional commit guard is present | commit-guard tests | baseline | pass |
| B9 | Commit guard refuses before rollback/current mutation | commit-guard refusal test | baseline | pass |
| B10 | Rollback temporary copy fails | existing rollback-copy failure test | baseline | pass |
| B11 | Prepared rollback validation fails | existing rollback-validation test | baseline | pass |
| B12 | Rollback digest differs from the open current descriptor | existing current-race test | baseline | pass |
| B13 | Current path no longer names the opened file before rollback publish | existing descriptor-race test | baseline | pass |
| B14 | Publishing the rollback sibling fails | existing rename failure test | baseline | pass |
| B15 | Syncing the rollback directory fails | existing sync failure test | baseline | pass |
| B16 | Current path changes after durable rollback publication | existing descriptor-race test | baseline | pass |
| B17 | Atomic current-executable replacement fails | existing replacement failure test | baseline | pass |
| B18 | Replacement directory sync fails and restoration is attempted | existing restore-after-sync-failure test | baseline | pass |
| B19 | Rollback restoration also fails and both failures are reported | existing double-failure test | baseline | pass |
