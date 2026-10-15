## 1. Freeze and evidence

- [x] 1.1 Register Story and generated PM tracker; freeze proposal through recorded four-perspective review.
- [x] 1.2 Capture and commit immutable baseline; collect current CodeGraph/caller evidence and AST/logic/branch maps before code changes.

## 2. Implementation and regression tests

- [x] 2.1 Demonstrate routing regressions with offline tests and implement effective-backend dependency enforcement.
- [x] 2.2 Demonstrate exact-number truncation regression and preserve retained JSON numbers with minimal decoding changes.
- [x] 2.3 Refresh final AST/branch mappings and test negative paths, enabled behavior and unchanged safety boundaries.

## 3. Independent verification

- [x] 3.1 Complete separate adversarial review and resolve every actionable finding.
- [x] 3.2 Complete additional gstack review and record actual reviewer/model and unresolved findings in review.md.
- [x] 3.3 Pass required offline checks; Manager inspect final diff, test receipts and user constraints before the completion gate.

Completion requires an additional successful `make gate CHANGE=a141-mcp-respects-backend-and-exact-numbers` receipt after all tasks are complete. Manager verifies that receipt before declaring completion. No deploy tasks are authorized.
