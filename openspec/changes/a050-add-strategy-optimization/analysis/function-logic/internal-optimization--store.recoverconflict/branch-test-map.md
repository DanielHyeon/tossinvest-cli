# Branch Test Map: `Store.recoverConflict`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | blank token/actor rejected | actor-binding tests | actor was mutable at baseline | PASS |
| B2 | unknown capability rejected | lifecycle invalid-token test | baseline lifecycle | PASS |
| B3 | storage error propagated | DB error coverage | baseline lifecycle | PASS |
| B4 | actor/time/boolean/raw payload/MAC tamper rejected | `TestCandidateCapabilityIsBoundToPreviewActorAcrossApplyRecoveryAndReplay`, MAC tamper tests | metadata trusted before hardening | PASS |
| B5 | invalid category/source/reason rejected | candidate metadata tamper matrix | metadata trusted before hardening | PASS |
| B6 | malformed or empty changes rejected | raw payload tamper test | metadata trusted before hardening | PASS |
| B7 | every attempted change is revalidated | candidate tamper matrix | registry drift unchecked before hardening | PASS |
| B8 | unknown/read-only/wrong-category/timing/safety/option rejected | candidate tamper matrix | registry drift unchecked before hardening | PASS |
| B9 | corrupt latest snapshot fails closed | snapshot digest corruption test | snapshot digest incomplete before hardening | PASS |
