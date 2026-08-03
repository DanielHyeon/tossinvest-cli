# Review — a073-operate-multi-market-strategy-lanes

- Date: 2026-08-04
- Stage: shared operational projection and pure deployment-guard waves implemented; performance and actual deployment pending
- Voices: Manager scope review, independent operations/security/UI review, projection contract review

## Findings and disposition

- Console, JSON and SSE consume one server-owned market projection. KR and US have separate status/error
  envelopes, exact `WIRED|UNWIRED` readiness and honest unavailable/not-configured states.
- The existing authenticated GET-only strategy-runtime pattern is reused; no order, gate, activation,
  autostart, protection-weakening route or free input is added.
- Performance attribution requires the complete persisted market-to-close identity chain and conserves
  partial/staged-close quantity, basis, fees, taxes, FX and PnL. Missing facts are `link_missing` or
  `not_measured`, never zero/current-FX inference.
- Compose replacement pins image/schema/config/activation/protection/volume preimages, replaces one service
  at a time and reverse-rolls only the replaced subset. When complete compatibility evidence proves entry
  remains OFF, incompatible rollback retains that observed state and forbids destructive downgrade.
- Deployment actions seal UTC issue/deadline times within five minutes. Their canonical evidence digest binds
  action/service/image/schema/health/state/environment/mount facts, so replay, late observations and mutation
  after evidence generation are rejected without advancing the plan.
- Independent HIGH review found three fail-closed gaps: an applied-but-unhealthy current attempt could be
  omitted from rollback, arbitrary/unbounded health digests could be replayed, and an unhealthy compatibility
  read could authorize rollback. The implementation now distinguishes `APPLIED|NOT_APPLIED`, rolls the applied
  current attempt first, seals the action window/canonical observation and refuses destructive rollback on
  unhealthy, timed-out or invalid compatibility evidence.
- Follow-up HIGH review found recovery records could falsely report entry `OFF` during state drift and health
  failure could hide simultaneous preservation drift. Recovery now reports exact common observed `ON|OFF`
  only from complete KR/US evidence, otherwise `UNKNOWN`; preservation drift is classified before schema or
  health failure and missing compatibility evidence cannot authorize rollback.
- The implemented shared model owns the exact paired KR/US shape and validation. Console, REST, SSE and the
  authenticated Unix reader consume that model without recomputing readiness, effective state or refusal.
- Unix transport review confirms a bounded strict decoder, exact private descriptor/socket permissions,
  no-follow plus same-file descriptor checks, constant-time bearer comparison and a `Read`-only client type.
- Integrated partial-failure and reconnect tests preserve the unaffected market byte-for-value and replace a
  prior partial state with one complete fresh snapshot; no zero/current/cross-market fallback exists.

## Design and DX disposition

The plan extends the existing `/strategy-runtime` information architecture rather than introducing a new
interaction model. Primary scan order is market identity → desired/effective/refusal → evidence/scheduler/
protection/reconciliation → campaign/risk → provenance/performance. Loading/unavailable, dormant, partial,
current and lineage-missing states are specified; mobile/accessibility and console/API parity reuse existing
golden-contract patterns. Operator time-to-diagnosis improves because every blocked market exposes its first
typed refusal without requiring journal joins.

## Projection-wave verification

- Strict OpenSpec validation: PASS.
- `go test -count=1 ./internal/strategyprojection ./internal/strategyprojectionrpc ./internal/console ./internal/httpapi`: PASS.
- The same four-package command with `-race`: PASS.
- `go vet` for the same four packages: PASS.
- Targeted OpenAPI/strategy-runtime contract tests: PASS.
- `git diff --check`: PASS.
- Legacy single-market console coverage was replaced by paired authority-projection, invalid/read-error
  fail-closed, authenticated GET/HEAD, no-input/no-mutation, responsive/CSP, partial-market and real Unix
  console/API/SSE convergence tests.
- Full performance/repository gates, real Compose preimage verification and final implementation review remain pending.

## Deployment-guard verification

- Pure `internal/deployguard` package and repository boundary tests cover immutable preimage refusal, exact
  rendered target images, frozen service order, bounded sequential actions, canonical evidence, typed
  timeouts, applied/not-applied subset accounting, reverse rollback and incompatible/read-failed recovery.
- `go test -count=1`, `go test -race -count=1`, `go test -count=25` and `go vet` for
  `./internal/deployguard`: PASS. Existing Compose/API separation static test, strict a073 OpenSpec validation
  and `git diff --check`: PASS.
- No Docker/Compose build, pull, replace, stop, down, rollback or other runtime mutation was executed. Tasks
  5.4 and 6 remain open until a human-authorized dormant deployment window has real immutable evidence.

## Verdict

The read-only operational projection and pure deployment-guard waves are ready for the remaining repository
gates. This is not a release or deployment verdict: performance, actual immutable preimage verification and
dormant deployment tasks remain open. Market activation remains an explicit later human decision.
