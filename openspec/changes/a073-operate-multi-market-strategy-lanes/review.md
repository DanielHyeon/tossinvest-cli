# Review — a073-operate-multi-market-strategy-lanes

- Date: 2026-08-04
- Stage: shared operational projection wave implemented; performance, Compose and deployment waves pending
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
  at a time and reverse-rolls only the replaced subset. Incompatible rollback keeps entry OFF and forbids
  destructive downgrade.
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
- Full performance/Compose/repository gates and final independent implementation review remain pending.

## Verdict

The read-only operational projection wave is ready for independent implementation review. This is not a
release or deployment verdict: performance, Compose, repository gate and dormant deployment tasks remain
open. Market activation remains an explicit later human decision.
