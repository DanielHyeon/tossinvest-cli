# Branch Test Map: `Gateway.checkReservation`

Measured at HEAD `0004536c` with `go test -count=1 -coverprofile ./internal/execgw (untagged, full package)` (statement coverage; `covered` = the branch body ran at least once in the package suite, it does not say which test). Named tests are attributed per test with a single-test `-coverprofile` run. Harness: `analysis/harness/branch_coverage_rows.py`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 890:2 — risk-reducing decision returns before any reservation read | `TestAnExitNeedsNoReservation`, `TestACancelNeedsNoReservation` | pre-a066 existing | covered at `0004536c` |
| B2 | if at 894:2 — `if err != nil {`; then `return reject(ReasonGuardianReservationMissing,` (line last changed by `a6a396ab`) | package suite `go test -count=1 -coverprofile ./internal/execgw (untagged, full package)` | n/a — branch line last changed by `a6a396ab`, not an a066 commit | NOT covered at `0004536c` |
| B3 | range at 899:2 — `for _, reservation := range reservations {`; then `if reservation.Held() {` (line last changed by `a6a396ab`) | package suite `go test -count=1 -coverprofile ./internal/execgw (untagged, full package)` | n/a — branch line last changed by `a6a396ab`, not an a066 commit | covered at `0004536c` |
| B4 | if at 900:3 — HELD aggregate reservation triggers q_final revalidation | `TestGatewayRefusesQFinalMarkedDecisionWithoutExactAdmissionBeforeBroker` | q_final checkpoint 2026-08-04 | covered at `0004536c` |
| B5 | if at 902:4 — q_final revalidation error refuses | `TestGatewayRefusesQFinalMarkedDecisionWithoutExactAdmissionBeforeBroker`, `TestRevokedDecisionIsRefusedAtTheLastMoment` | q_final checkpoint 2026-08-04 | covered at `0004536c` |
| B6 | if at 903:5 — decision disappeared during revalidation keeps the Guardian-missing reason | `TestRevokedDecisionIsRefusedAtTheLastMoment` | reason regression, q_final checkpoint 2026-08-04 | covered at `0004536c` |
| B7 | if at 908:5 — entry loss lock refusal is reported as its own reason (entry_loss_lock_active), chosen by errors.Is on the journal sentinel type, not by text | `TestA066DecisionIssuedBeforeLockIsRefusedAtSubmit` | mutants M30 (mapping removed) and M31 (mapped to mismatch) CAUGHT (analysis/mutation-5.5/ledger-run3.tsv) | covered at `0004536c` |
