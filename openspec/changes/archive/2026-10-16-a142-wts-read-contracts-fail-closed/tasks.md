## 1. Freeze and evidence

- [x] 1.1 Resolve watchlist optional-field/ordering contract from upstream fixtures; complete four-perspective proposal review and PM registration.
- [x] 1.2 After a141 landing capture and commit immutable baseline; collect fresh CodeGraph, AST, branch/logic maps.

## 2. Implement and test

- [x] 2.1 Add RED malformed-order collection tests, minimal fail-closed guards and partial-market failure coverage.
- [x] 2.2 Add RED dedicated-watchlist fixtures and minimally adapt the read route/decoder while preserving mutation contracts.
- [x] 2.3 Refresh final maps; pass focused and existing affected consumer/mutation regression tests using offline fixtures.

- [x] 2.4 Align only the existing watchlist probe definition with the dedicated read route/collection shape; capture fresh pre-edit evidence and pass offline probe contract tests.

## 3. Review and verification

- [x] 3.1 Complete separate adversarial review and resolve actionable findings.
- [x] 3.2 Complete additional gstack review; record actual reviewers and disposition in review.md.
- [x] 3.3 Pass focused offline checks; Manager inspect final diff, test receipts and user constraints before the completion gate.

Completion additionally requires a successful `make gate CHANGE=a142-wts-read-contracts-fail-closed` receipt, verified by Manager after the tasks complete. No deployment or operating settings change is authorized.
