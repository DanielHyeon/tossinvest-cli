# Review

## Proposal freeze — 2026-08-03

Voices: Manager, root-cause explorer, test inspector, security auditor.

Decision: ACCEPT with the following required implementation constraints.

- Root-cause evidence confirms desired/effective adoption are ON and an account-wide permanent block wins before adoption.
- Live journal evidence confirms the initiating missing order has no local mutation attempt or lineage, only an external non-terminal snapshot.
- Test review found no engine→journal→console transition test and identified the adopted-without-exit recovery window as follow-up coverage.
- Security review found that adoption/runtime ignored non-quantity journal causes, release visibility preceded durable persistence, and include-only constructor validation could panic.
- The design accepts all relevant P0/P1 findings: positive ownership evidence, full journal authority, durable-first release, fail-closed cycle abort, and local audited recovery only.
- No web/HTTP mutation route, live order command, toggle flip, or synthetic protection price is allowed.

## Independent implementation review

### Security — 2026-08-03

Decision: ACCEPT, P0=0, P1=0.

- Confirmed AMEND lineage requires exact intent/attempt ownership, target-parent, broker-child, and canonical scope.
- Malformed legacy edges and cross-account/market/day identifier reuse cannot promote external snapshots.
- Ambiguity durably enters `IDENTIFIER_CONFLICT`; no projection, hook, early reservation release, live order, or toggle mutation is introduced.

### Maintainability — 2026-08-03

Decision: ACCEPT, P0=0, P1=0.

- Startup reservation recovery precedes nonce pruning and engine decisions; held reservations retain spent-nonce evidence.
- Migration-shaped regressions exercise actual blank-scope rows for both tracked lineage and terminal live-order queries.
- Non-blocking follow-up: extract the repeated canonical ownership SQL predicates into a repository query abstraction to reduce drift risk.

### Test — 2026-08-03

Pending independent final test verdict and repository-wide delivery gates.
