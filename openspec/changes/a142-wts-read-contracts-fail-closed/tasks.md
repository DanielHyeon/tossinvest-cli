## 1. Freeze and evidence

- [ ] 1.1 Resolve watchlist optional-field/ordering contract from upstream fixtures; complete four-perspective proposal review and PM registration.
- [ ] 1.2 After a141 landing capture and commit immutable baseline; collect fresh CodeGraph, AST, branch/logic maps.

## 2. Implement and test

- [ ] 2.1 Add RED malformed-order collection tests, minimal fail-closed guards and partial-market failure coverage.
- [ ] 2.2 Add RED dedicated-watchlist fixtures and minimally adapt the read route/decoder while preserving mutation contracts.
- [ ] 2.3 Refresh final maps; pass focused and existing affected consumer/mutation regression tests using offline fixtures.

## 3. Review and verification

- [ ] 3.1 Complete separate adversarial review and resolve actionable findings.
- [ ] 3.2 Complete additional gstack review; record actual reviewers and disposition in review.md.
- [ ] 3.3 Pass required offline checks; Manager inspect final diff, test receipts and user constraints before the completion gate.

Completion additionally requires a successful `make gate CHANGE=a142-wts-read-contracts-fail-closed` receipt, verified by Manager after the tasks complete. No deployment or operating settings change is authorized.
