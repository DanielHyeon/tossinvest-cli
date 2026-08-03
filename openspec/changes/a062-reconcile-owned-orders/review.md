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

- Final re-review confirmed canonical direct PLACE/AMEND ownership takes precedence and lineage is only a same-scope fallback, so a parent PLACE plus parent→child AMEND is not counted as two owners.
- CANCEL attempts never create fill ownership; owner evidence must be exact-scope, unique-intent, and strictly earlier than the observation.
- The external cumulative 10 → parent PLACE → AMEND → late parent cumulative 12 regression projects and appends the full delta 12.
- Focused race and full journal/engine tests passed; no projection, early reservation release, live order, toggle mutation, or authorization expansion is introduced.

### Maintainability — 2026-08-03

Decision: ACCEPT, P0=0, P1=0.

- Direct-owner precedence is canonical-scope local, while duplicate direct owners and cross-scope ambiguity remain fail closed.
- The late-parent regression pins baseline reset, full delta projection, and one exact scoped event.
- Focused race and vet checks passed. Generated Function Logic Maps were regenerated afterward; logic-map and full diff checks pass.
- Non-blocking follow-up: extract the repeated canonical temporal ownership SQL predicates into a repository query abstraction to reduce drift risk.

### Test — 2026-08-03

Initial decision: REJECT, P0=0, P1=1, P2=3.

- P1: snapshot persistence, detector deduplication/lineage, and reconciliation comparison still collapsed canonical identities back to `order_id`; a different-scope broker order could make recovery falsely clean.
- Required remediation: composite durable snapshot identity, canonical detector keys, canonical local/broker comparison, and an end-to-end reused-identifier regression.
- P2 follow-ups requested: direct legacy non-lineage reuse, decision-bound reservation negative coverage, and production recovery integration coverage.

The canonical snapshot/detector/comparison remediation, legacy temporal binding, and decision-bound reservation negatives are implemented.

Final decision: ACCEPT, P0=0, P1=0.

- The direct-owner precedence, same-second future-owner, duplicate-owner, CANCEL non-owner, external baseline reset, late replacement-parent fill, provenance/outcome, reconcile lineage, tracer, and exit E2E regressions pass.
- The uncached related-package suite and focused race suite pass with no failing command and no missing P0/P1 test.

## Delivery and live verification — 2026-08-03

Decision: ACCEPT.

- The feature was merged to local `main`, pushed to `origin/main`, built into the local container image, and deployed with both HTTPS services healthy.
- No live order command, broker mutation route, or operating-toggle change was invoked during deployment or recovery.
- With the engine stopped and its lock held exclusively, `engine reconcile-resolve` released the obsolete quantity-mismatch blocks only after three stable official snapshots and a clean blocking diff; the engine was then restarted.
- The first post-restart reconcile cycle reported `blocked=0` and adopted three external holdings: one KR and two US. Subsequent cycles remained clean with no new adoption block.
- The live `/positions` response shows every holding as `MANAGED` with adoption evidence. Fresh canonical evaluations render take-profit, initial stop, recovery, active baseline, and high-water values; evaluations outside the display freshness bound retain the existing fail-closed `—` presentation and stale warning.
- Read-only browser rendering at the desktop viewport confirmed the compact table columns remain aligned, `상세 보기` stays in the final column, and no row displays the former reconcile-blocked or adoption-pending verdict.
