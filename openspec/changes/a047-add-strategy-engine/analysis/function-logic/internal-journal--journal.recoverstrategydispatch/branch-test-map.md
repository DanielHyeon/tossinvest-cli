# Branch Test Map: `Journal.RecoverStrategyDispatch`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | zero core attempts | `TestPendingStrategyPlansAreAccountScopedAndZeroAttemptRecoversToRefused` | no durable state | pass |
| B2 | one state across exact terminal map | `TestStrategyRecoveryClassifiesOneCoreAttemptExactly` | all non-confirmed in doubt | pass |
| B3 | multiple core attempts | `TestStrategyRecoveryMultipleCoreAttemptsIsDurableInDoubt` | cardinality unchecked | pass |
| B4 | current IN_DOUBT becomes DISPATCHED on proof | promotion test | PLANNED-only CAS | pass |
| B5 | account/state receipt mismatch | account-scoped pending tests | missing | pass |
| B6 | core query error | journal suite | missing | pass |
| B7 | row scan/iteration error | journal suite | missing | pass |
| B8 | zero attempt terminal write failure joins error | terminal CAS tests | missing | pass |
| B9 | >1 while already IN_DOUBT stays blocked | cardinality test | missing | pass |
| B10 | >1 planned persists reason | cardinality test | missing | pass |
| B11 | confirmed requires non-empty broker id | state table | missing | pass |
| B12 | definitive terminal refusal persists reason | state table | wrong draft class | pass |
| B13 | nonterminal planned persists IN_DOUBT | state table | missing | pass |
| B14 | nonterminal current IN_DOUBT stays blocked | promotion/current tests | missing | pass |
| B15 | fallback returns typed recovery error | state table | missing | pass |
