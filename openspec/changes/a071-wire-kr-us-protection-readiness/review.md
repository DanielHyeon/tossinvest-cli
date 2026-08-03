# Review — a071-wire-kr-us-protection-readiness

- Date: 2026-08-03
- Stage: proposal freeze; high-risk implementation not started
- Voices: Manager safety review, independent operations/security review, final semantic re-review

## Findings and disposition

- Readiness is market-scoped `WIRED|UNWIRED` plus typed refusal; no combined KR+US state exists.
- Signed attestation binds pinned trust roots, key ID/algorithm allowlist, rotation/revocation, monotonic
  serial, maximum lifetime and durable trusted-time floor.
- Scope includes exact broker client-key echo, lookup/uniqueness, pending/terminal/cancel query,
  dedup/idempotency and replace semantics. Missing capability is `UNWIRED`, not guessed support.
- Submit/cancel unknown and orphan orders use exact identity reconciliation only; no symbol/time inference,
  blind resubmit or inferred cancellation is allowed.

## Verification

- Strict OpenSpec validation: PASS.
- Final independent security/operations re-review: PASS, no open blocker.
- Fixture/isolated official-broker tests structurally exclude live transport and automatic activation.

## Verdict

Proposal freeze approved for high-risk RED implementation. Actual signed KR/US capability evidence still
requires a separate human-approved workflow; absent evidence keeps that market `UNWIRED`.
