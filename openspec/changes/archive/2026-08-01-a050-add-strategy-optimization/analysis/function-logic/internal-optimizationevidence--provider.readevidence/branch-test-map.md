# Branch Test Map: `Provider.ReadEvidence`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil provider/reader/clock cannot produce evidence | `TestProviderRejectsNilDependenciesAndZeroClock` | provider absent at `948e721` | PASS |
| B2 | zero observation clock fails closed | `TestProviderRejectsNilDependenciesAndZeroClock` | provider absent at `948e721` | PASS |
| B3 | dashboard error is wrapped and unavailable | `TestProviderDashboardErrorIsUnavailableAndNeverLaunderedAsEvidence` | provider absent at `948e721` | PASS |
| B4 | dashboard query echo differs in any fixed dimension | `TestProviderRejectsDashboardThatDoesNotEchoTheExactFixedQuery` | provider could trust a different query | PASS |
| B5 | canonical digest error stays unavailable; supported fixed structs bind normal path | deterministic digest test plus AST review | provider absent at `948e721` | normal path PASS; defensive branch reviewed |
| B6 | persisted source time enters explicit temporal classification | persisted source-time tests | as-of could launder freshness | PASS |
| B7 | missing or future persisted source time is stale with explicit source-time reason | `TestProviderRefusesMissingOrFuturePersistedSourceTime` | as-of could launder freshness | PASS |
| B8 | source older than 72 hours is stale | `TestProviderUsesPersistedSourceTimeAndMarksOldEvidenceStale` | stale evidence could appear complete | PASS |
| B9 | temporal reason is deduplicated/sorted and stale takes precedence | persisted source-time tests | temporal defect absent | PASS |
| B10 | without temporal defect, ordinary missing evidence selects insufficient | fail-closed evidence matrices | provider absent at `948e721` | PASS |
| B11 | non-empty ordinary missing list selects insufficient; empty list remains complete | fixed query complete case and missing-class tests | provider absent at `948e721` | PASS |
