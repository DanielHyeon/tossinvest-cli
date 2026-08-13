# Branch Test Map: `Tracker.Observe`

| Branch | Scenario | Test/evidence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | first mismatch initializes the block set | `TestQuantityMismatchBlocksEntries` | existing | existing |
| B2 | non-blocking comparison clears transient promotion streaks without clearing permanent state | existing mismatch permanent/reset regressions plus a110 reset case | pending | pending |
| B3 | a blocking comparison derives unique canonical dispute streaks instead of incrementing a shared scalar | a110 changing-symbol, tuple and missing-order regressions | pending | pending |
| B4 | a clean quantity read does not release a foreign-cause block | existing foreign-cause preservation regression | preserve | preserve |
| B5 | a clean quantity read does not release an operator-only/differently released block | existing permanent/operator release regressions | preserve | preserve |
| B6 | no strictly earlier adjustment credit leaves the block awaiting adjustment | existing a083/a083b credit tests | preserve | preserve |
| B7 | a later comparison that still disputes the symbol cannot spend its credit | existing a083/a083b refuted-credit tests | preserve | preserve |
| B8 | every current mismatch is projected to an ordinary block | `TestQuantityMismatchBlocksEntries` | existing | existing |
| B9 | a new ordinary block is marked pending and latched before journal I/O | `TestNewMismatchLatchesTheGateBeforeJournalIO` | existing | existing |
| B10 | any one exact current key at threshold proposes permanent promotion; changing keys cannot pool | a110 incident/exact-identity regressions | pending | pending |
| B11 | the absent account permanent key is added pending without replacing durable foreign authority | existing conflict test plus a110 pending-promotion cases | pending | pending |
| B12 | outcome slices are deterministic | existing block-ordering assertions | preserve | preserve |
| B13 | current additions are indexed before generic pending retry | existing failed-enter retry plus a110 pending split | pending | pending |
| B14 | ordinary older pending entries retry; stale account promotion retries only while its earning key continues | a110 failed permanent enter clean/different/same-key cases | pending | pending |
| B15 | journal-confirmed additions replace pending rows as durable | existing write-through tests plus a110 same-key retry | pending | pending |
| B16 | a different durable authoritative cause replaces the local proposal | existing cause-conflict regression | preserve | preserve |
| B17 | only journal-confirmed exact releases leave the in-memory block set | existing partial/exact release regressions | preserve | preserve |
| B18 | committed releases are indexed by normalized symbol for credit accounting | existing a083/a083b credit tests | preserve | preserve |
| B19 | credits are examined independently per symbol | existing a083/a083b multi-symbol tests | preserve | preserve |
| B20 | same/earlier/unparseable comparison time cannot spend a credit | existing a083/a083b time-order tests | preserve | preserve |
| B21 | refuted, committed or orphaned credits are removed at their bounded lifetime | existing a083/a083b credit-lifetime tests | preserve | preserve |
| Restore/Refresh | restart preserves already-durable permanent blocks without recreating transient streak evidence | existing restore tests plus new a110 restart regression | pending | pending |

Additional B3/B10 acceptance vectors: equivalent `1`/`1.0` continues one streak; a float64 2^53
collision pair remains distinct; blank/malformed/non-finite values stay ordinarily blocked but earn no
promotion count; a valid sibling in the same diff still counts; duplicate identities count once; complete
missing-order identity scopes opaque-ID reuse; each blank required missing-order component remains an
ordinary block but cannot earn a promotion streak, while a valid sibling still counts.

`pending` means the implementation Teammate must first demonstrate RED against the frozen base, then record GREEN after the minimal fix.
