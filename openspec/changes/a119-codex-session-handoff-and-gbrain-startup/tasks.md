## 1. Repair specification completeness

- [x] 1.1 Complete both proposal-declared capability deltas and independently review their scope and canonical compatibility.
- [x] 1.2 Run strict a119 validation and repository-wide validation; record actual results without claiming implementation completion.

## 2. Establish implementation evidence

- [x] 2.1 Capture the implementation baseline before code changes and record current hard evidence and applicable function analysis.
- [x] 2.2 Obtain sanitized supported-host event fixtures and effective configuration-loading evidence; identify one Codex-owned launch path.
- [x] 2.3 Complete proposal-freeze gstack review on the evidence-backed plan.
      > 2026-09-25 first pass: **REJECT** (review.md) — plan no longer matched the 2026-08-29 Why; scope decision handed to the user.
      > 2026-09-25 user decision **(a)**: proposal rewritten to evidence + regression pins; symptoms to named follow-ups
      > (proposal "Follow-ups", issues I-1). Second pass 2026-09-25: **PASS** — independent adversarial voice (Claude Sonnet,
      > read-only; run early on purpose during the Opus limit), blocking 0 · should-fix 0 · note 3, Manager spot-checked (review.md
      > "Re-freeze review"). Frozen under scope (a). 3.x may start (separate teammate).

## 3. Implement and verify

- [ ] 3.1 Land the regression pins from `wip/a119-3.1` through a separate teammate: fixture × matcher test, saver stdout test,
      one-effective-registration test, mutation harness with a green no-mutation control.
- [ ] 3.2 Verify isolation, redaction, atomic persistence, lock-owner preservation and unchanged tool results in isolated tests
      (design evidence map), and that the configuration files are byte-identical to `54004f44`.
- [ ] 3.3 Record in `analysis/host-evidence.md` §4 what stays unobserved (interactive-host delivery, per-thread startup) and
      that it belongs to the named follow-ups; make no runtime claim here.
- [ ] 3.4 Complete separate adversarial review, gstack review, required tests and SDD checks, final gate and Manager acceptance
      before PM synchronization and archive.
