# Branch Test Map: `dispatchValidated`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | positive exact confirmed path | `TestValidatedDispatchPlansOnceAndPersistsExactOfficialOutcome` | stale tests could not mint opaque decision | pass via private seam |
| B2 | plan-time gate mutation | `TestValidatedDispatchPlanTimeGateChangeRefusesBeforeOfficialCall` | TOCTOU window | pass |
| B3 | exact expiry reached during plan | expiry boundary test | only initial expiry check | pass |
| B4 | empty attempt/definitive/ambiguous outcomes | `TestOfficialOutcomeDispositionPreservesAttemptEvidence` | string-only result | pass |
| B5 | manifest lease blocks revocation | manifest lease test | validation/call race | pass |
| Security | every ManifestBinding field mismatch | `TestManifestVerificationRejectsMismatchInEveryField` | deleted activation coverage | pass, 32/32 |
| Security | every DecisionRecord field mismatch | `TestDecisionBindingRejectsMismatchInEveryDecisionRecordField` | partial gate binding | pass, 60/60 |
| B6 | atomic issuer failure | issuer/journal tests | missing | pass |
| B7 | gate lease sees revision/binding drift | TOCTOU test | race window | pass |
| B8 | gate predicates change under lease | TOCTOU test | race window | pass |
| B9 | manifest digest changes after plan | manifest revocation test | race window | pass |
| B10 | decision expires exactly before call | expiry boundary test | missing recheck | pass |
| B11 | gateway invoked only after all leases | positive/lease tests | missing | pass |
| B12 | manifest error before gateway maps TOCTOU | revocation test | missing | pass |
| B13 | successful lease enters outcome switch | positive/outcome tests | missing | pass |
| B14 | DISPATCHED outcome | positive test | missing | pass |
| B15 | dispatched persistence failure surfaces plan error | issuer spy error path | missing | pass |
| B16 | REFUSED outcome | outcome disposition table | missing | pass |
| B17 | refusal persistence failure joins error | issuer failure path | missing | pass |
| B18 | IN_DOUBT outcome | outcome disposition table | missing | pass |
| B19 | in-doubt persistence failure joins error | issuer failure path | missing | pass |
| B20 | lease not entered or typed TOCTOU | gate mutation test | missing | pass |
| B21 | TOCTOU refusal persistence | gate mutation test | missing | pass |
| B22 | typed TOCTOU returned | gate mutation test | missing | pass |
| B23 | post-call definitive refusal | outcome table | string-only draft | pass |
| B24 | post-call refusal persistence | outcome table | missing | pass |
| B25 | post-call exact dispatch proof | positive table | missing | pass |
| B26 | post-call dispatch persistence | positive test | missing | pass |
| B27 | post-call ambiguous outcome persists IN_DOUBT | outcome table | missing | pass |
