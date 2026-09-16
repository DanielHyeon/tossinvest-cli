## 1. Repair specification completeness

- [x] 1.1 Complete both proposal-declared capability deltas and independently review their scope and canonical compatibility.
- [x] 1.2 Run strict a119 validation and repository-wide validation; record actual results without claiming implementation completion.

## 2. Establish implementation evidence

- [ ] 2.1 Capture the implementation baseline before code changes and record current hard evidence and applicable function analysis.
- [ ] 2.2 Obtain sanitized supported-host event fixtures and effective configuration-loading evidence; identify one Codex-owned launch path.
- [ ] 2.3 Complete proposal-freeze gstack review on the evidence-backed implementation plan.

## 3. Implement and verify

- [ ] 3.1 Add matcher and duplicate-registration regressions and implement the smallest reviewed correction through a separate teammate.
- [ ] 3.2 Verify isolation, redaction, atomic persistence, lock-owner preservation and unchanged tool results in isolated tests.
- [ ] 3.3 Observe supported-host event delivery and single startup; keep unavailable runtime proof explicitly pending.
- [ ] 3.4 Complete separate adversarial review, gstack review, required tests and SDD checks, final gate and Manager acceptance before PM synchronization and archive.
