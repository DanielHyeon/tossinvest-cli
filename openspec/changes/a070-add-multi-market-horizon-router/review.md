# Review — a070-add-multi-market-horizon-router

- Date: 2026-08-03
- Stage: proposal freeze; implementation not started
- Voices: Manager architecture/safety review, independent architecture/test review, round-2 adversarial review

## Findings and disposition

- **Accepted blocker:** ownership key is `(account, market, canonical symbol, position_generation)`;
  horizon is admission/attribution only. Routing checks every active horizon owner before scoring.
- Market/horizon rate capabilities are anti-replay/admission subscopes over one physical endpoint and
  reset-generation quota authority; they cannot multiply provider capacity or safety reserve.
- KR and US desired state uses per-market record/revision/lock/activation CAS. Legacy disabled migrates
  both OFF; a verified single-market state may migrate only that market while its peer remains OFF.
- Official exchange calendars and IANA zones define independent sessions; market failure and retry state
  do not cross-contaminate.

## Verification

- Strict OpenSpec validation: PASS.
- Cross-horizon ownership, shared-quota exhaustion, concurrent CAS, rollback and crash migration are RED tasks.
- No combined KR+US approval or automatic activation is introduced.

## Verdict

Proposal freeze approved. KR and US routing implementation proceeds in the same wave with independent
activation and shared account-wide mutation authority.
