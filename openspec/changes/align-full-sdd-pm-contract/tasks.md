## 1. Contract and baseline

- [x] 1.1 Record the StockOS-to-TossOS SDD comparison, preserved TossOS constraints, and evidence reconciliation
- [x] 1.2 Validate the OpenSpec proposal, design, delta spec, and implementation tasks in strict mode
- [x] 1.3 Capture the change base commit and complete the proposal-freeze review

## 2. PM enforcement tests

- [x] 2.1 Add RED tests that reject bootstrap allowlists, orphan active changes, duplicate Story mappings, invalid paths, and manually stored Story status
- [x] 2.2 Add RED tests for evidence-derived Story lifecycle states

## 3. PM portfolio migration

- [x] 3.1 Add TossOS Initiative, Epic, and Feature records that classify every active OpenSpec change without changing product-specific intent
- [x] 3.2 Backfill exactly one Story for each active OpenSpec change and register every Story in its Feature and the PM registry
- [x] 3.3 Convert the existing archived SDD Story to the canonical OpenSpec mapping shape
- [x] 3.4 Remove the bootstrap allowlist and make PM validation enforce bidirectional Story-to-active-change one-to-one coverage
- [x] 3.5 Derive Story status from OpenSpec proposal, task, and archive evidence and regenerate the master trackers

## 4. Full SDD workflow alignment

- [x] 4.1 Align docs/WORKFLOW.md with the StockOS Full SDD sequence, READY gate, evidence reconciliation, Pre-Edit Gate, TDD, completion, and archive/PM order
- [x] 4.2 Preserve TossOS-specific Go AST and test gates, official Toss Open API safety, upstream compatibility, journal/Guardian rules, namespaces, and deployment controls
- [x] 4.3 Remove contradictory bootstrap-exception wording from the SDD workflow contract through the OpenSpec delta

## 5. Verification and completion

- [x] 5.1 Run PM generator unit tests and confirm RED-to-GREEN evidence
- [x] 5.2 Run strict OpenSpec validation for all changes and regenerate trackers without drift
- [ ] 5.3 Run make sdd-sync, make sdd-check, and make gate CHANGE=align-full-sdd-pm-contract
- [x] 5.4 Complete an independent review, update PM-derived evidence, and record the final completion report
