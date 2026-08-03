# Review — a072-wire-multi-market-strategy-runtime

- Date: 2026-08-03
- Stage: proposal freeze; high-risk implementation not started
- Voices: Manager safety review, independent operations/security review, round-2 lease adversarial review

## Findings and disposition

- KR and US evaluation workers start in the same runtime/release with independent calendar, activation,
  evidence cursor, budget and failure envelope; one market never waits for the other's stability.
- Dispatch uses durable owner epoch/fencing and irreversible
  `ISSUED→CLAIMED→SUBMITTING→SUBMITTED|AMBIGUOUS|REFUSED`; authority A→B→A cannot revive a lease.
- Round 2 froze reservation disposition atomically with exact outcome: pre-transport or definitive broker
  rejection/no-accept/no-fill is `REFUSED+RELEASED`; acceptance is `SUBMITTED+TRANSFERRED`; durable
  transport uncertainty alone is `AMBIGUOUS+HELD`. Reconciliation changes disposition, never lease state.
- Market-worker faults latch/restart only that market's entry worker. Central integrity faults block all
  entry and require fenced safety-only fallback within 60 seconds while broker protection is preserved.

## Verification

- Strict OpenSpec validation: PASS.
- Final independent semantic re-review: PASS, no open blocker.
- No lane/autostart/automation/LIVE approval mutation or second broker path is introduced.

## Verdict

Proposal freeze approved for high-risk RED implementation. Dormant deployment and operational activation
remain separate states.
