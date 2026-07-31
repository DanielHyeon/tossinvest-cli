# Branch Test Map: `DesiredStore.Save`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1-B2 | canceled/invalid desired state is not written | desired validation and context tests | existing | yes |
| B3-B6 | directory/lock/current-read failure cannot reach rename | focused filesystem contracts + full suite | existing structure | yes |
| B7 | stale ON after committed OFF returns revision conflict and leaves OFF revision 2 | `TestDesiredStateRevisionCASPreservesCommittedOperatorOff` | yes, stale save overwrote OFF | yes |
| B8 | max revision cannot wrap to zero | `TestDesiredRevisionCannotWrap` | revision field absent | yes |
| B9-B17 | temp install errors propagate and pre-commit cancellation preserves old file | atomic write structure + existing permission/round-trip tests | existing | yes |
| cross-process | child Save blocks while parent holds state flock, then commits OFF | `TestDesiredStoreSaveSerializesAcrossProcesses` | no lock existed | yes |
| bounded wait | held flock honors caller deadline and default 2s ceiling without mutation | `TestDesiredStoreSaveCancelsWhileWaitingForProcessLock`, `TestDesiredLockWaitIsBoundedWithoutCallerDeadline` | blocking flock ignored context forever | yes |
| final | successful save persists mode 0600 and increments revision | round-trip + CAS tests | revision absent | yes |
| B1 | initial context cancellation | context test | existing | yes |
| B2 | desired validation/future approval | validation tests | existing | yes |
| B3 | desired directory creation | round-trip test | existing | yes |
| B4 | flock acquisition | subprocess serialization test | lock absent | yes |
| B5 | cancellation after lock | context gate/code path | existing guard | yes |
| B6 | current strict load | malformed/CAS tests | existing | yes |
| B9 | JSON marshal | round-trip test | existing | yes |
| B10 | temp creation | round-trip/permission test | existing | yes |
| B11 | chmod 0600 | round-trip mode assertion | existing | yes |
| B12 | temp write | round-trip test | existing | yes |
| B13 | temp fsync | round-trip test | existing | yes |
| B14 | temp close | round-trip test | existing | yes |
| B15 | pre-commit context gate | context contract | existing guard | yes |
| B16 | atomic rename | round-trip/CAS tests | existing | yes |
| B17 | directory open before sync | round-trip durability test | existing | yes |
