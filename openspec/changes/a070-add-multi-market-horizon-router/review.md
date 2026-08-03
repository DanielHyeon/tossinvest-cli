# Review — a070-add-multi-market-horizon-router

- Date: 2026-08-03
- Stage: core implementation complete; a067–a069/a066 runtime integration and root gate pending
- Voices: Manager architecture/safety review, independent architecture/test review, round-2 authority-boundary review

## Findings and disposition

- **Accepted blocker:** ownership key is `(account, market, canonical symbol, position_generation)`;
  horizon is admission/attribution only. Routing checks every active horizon owner before scoring.
- Market/horizon rate capabilities are anti-replay/admission subscopes over one physical endpoint and
  reset-generation quota authority; they cannot multiply provider capacity or safety reserve.
- KR and US desired state uses per-market record/revision/lock/activation CAS. Legacy disabled migrates
  both OFF; a verified single-market state may migrate only that market while its peer remains OFF.
- Official exchange calendars and IANA zones define independent sessions; market failure and retry state
  do not cross-contaminate.
- **Resolved blocker:** routing now requires an exact sealed market record/revision in READY state. A caller
  cannot turn an OFF durable market ON by submitting an ON candidate.
- **Resolved blocker:** quota freshness is checked with the authority's trusted clock, not caller-provided
  observation time; backdating cannot revive a stale physical quota snapshot.
- **Resolved blocker:** owner snapshots, ON market records and physical quota snapshots have package-private
  attestation constructors. External callers can inspect values but cannot mint a valid seal. The only public
  market constructor creates a sealed OFF/UNOBSERVED record.
- Rollback accepts OFF targets only. It cannot replay a historical ON activation as fresh authority.
- Inactive same-key owner rows do not compete. Any row from another account/market/symbol/generation corrupts
  the bounded snapshot even when inactive; more than one active same-key row is reconstruction mismatch.

## Verification

- Strict OpenSpec validation: PASS.
- `go test ./internal/strategyrouter -count=1`: PASS.
- `go test -race ./internal/strategyrouter -count=1`: PASS.
- `go vet ./internal/strategyrouter`: PASS.
- `FuzzOwnerKeyNeverIncludesHorizon` and `FuzzLegacyMigrationRetryConverges` (3s each): PASS.
- Cross-horizon ownership, durable OFF binding, shared-quota exhaustion/backdating, concurrent CAS,
  OFF-only rollback and crash migration tests: PASS.
- External-package forgery and exported authority-constructor checks: PASS.
- No combined KR+US approval or automatic activation is introduced.

## Verdict

Core pure/sealed-port implementation approved for integration review. KR and US ship in the same release,
remain independently OFF/UNOBSERVED by default, and share one physical quota authority. Runtime wiring,
broader safety-loop regressions and the repository SDD/gate remain intentionally pending at root integration.
