# Review — a065-add-position-campaign-leg-core

- Date: 2026-08-04
- Stage: Wave 1A core implemented; production entry wiring intentionally disconnected
- Voices: Manager scope/safety review, independent adversarial ledger review, final semantic re-review

## Findings and disposition

- **Accepted:** campaign creation before first fill requires a prospective position-generation CAS and a
  set-once first-fill binding; manual/external position drift enters reconciliation rather than renumbering.
- **Accepted:** Campaign and Leg transition tables include rejection, cancel pending/cancelled, expiry,
  submit ambiguity, replacement, late fill and recovery terminality.
- **Accepted:** cumulative fill watermarks are broker-order scoped and replacement lineage aggregates
  predecessor/child deltas without assuming the new order continues the predecessor counter.
- **Accepted after round 2:** a late positive fill on a replaced/cancelled predecessor advances the immutable
  order watermark and Position exactly once in the same transaction. Cap or lineage ambiguity preserves
  the fill, recalculates remaining quantity and latches campaign `RECONCILE`/new-entry block.
- **Accepted after round 3:** CREATE now compares the latest authoritative Position generation/state and a
  post-v20 projection version companion. Legacy rows remain unversioned, and existing OPEN/manual positions
  cannot acquire a prospective campaign.
- **Accepted after round 3:** decision/intent/attempt/order/replacement scope is journal-authoritative. One
  scoped broker order has one campaign leg owner and one predecessor has one successor; caller ambiguity flags
  cannot bypass the scoped lineage.
- **Accepted after round 3:** duplicate legacy campaign matches and CLOSED late fills never reject the
  authoritative fill. They append deterministic evidence, preserve Position progress, retain CLOSED terminality
  where applicable, and latch durable reconciliation/new-entry blocking.
- **Accepted after round 4:** every durable refusal/latch advances campaign version only with the matching
  deterministic command, append-only event and complete projection digest. Exact retry is stable, while SQLite
  triggers reject UPDATE/DELETE of command/event evidence.
- **Accepted after round 5:** offline reconstruction now derives Campaign/Leg transitions from immutable event
  kind, campaign version, request digest, leg quantities and per-order delta/watermark facts. Stored state cannot
  invent a transition, and per-order remaining is recomputed from that order's cap.
- **Accepted after round 5:** projection checkpoints bind campaign account/market/symbol/lane/version/decision/
  evidence identity, expected generation/version and the full durable claim row. Claim deletion/mutation and
  identity drift are reported as snapshot drift.
- **Accepted after round 5:** CREATE requires exact immutable strategy-decision lineage. LINK includes caller
  ambiguity in its command digest, durably refuses `true`, validates initial intent or replacement-edge quantity,
  and never persists a caller-defaulted false without authoritative evidence.
- **Accepted after round 5:** pre-v20 Positions expose explicit `LEGACY_UNKNOWN` campaign lineage with no
  synthetic campaign identifier.

## Verification

- Strict OpenSpec validation: PASS.
- RED→GREEN coverage includes complete Campaign/Leg transition tables, prospective CAS races,
  expected-version/command retries, restart/rollback, replacement and late predecessor fills,
  delta-zero terminal cancellation, aggregate cap excess, monotone stops and offline replay.
- Independent adversarial review findings were resolved: immutable aggregate lineage, CLOSED/rebind
  replay refusal, aggregate cap reconciliation, calculation-before-watermark commit, positive stop
  provenance, frozen collision-safe command identity, terminal delta-zero application and successor
  query error propagation.
- `go test ./internal/journal -count=1`: PASS (286.182s, including round-5 hardening and legacy journal regressions).
- `go test -race ./internal/positioncampaign`: PASS (1.017s).
- focused `go test -race ./internal/journal`: PASS (57.877s).
- `go vet ./internal/positioncampaign ./internal/journal`: PASS.
- `openspec validate a065-add-position-campaign-leg-core --strict --no-interactive`: PASS.
- Function Logic Map completeness check: PASS.
- Broker/config assertion: `TestCampaignCoreHasNoProductionBrokerOrToggleWiring` PASS; production Campaign hook remains unbound.
- CodeGraph hard sync completed; CodeGraphContext update stalled for more than 60 seconds and was interrupted as
  advisory-only. `make sdd-check` passed once with a fresh hard-evidence fingerprint and advisory
  CodeGraphContext/GBrain stale warnings retained. Because KR/US lane changes are landing concurrently, the final
  combined-worktree fingerprint refresh/check remains an integration-stage action.
- Exit-first and non-retreating stop authority remain outside the campaign policy and are not weakened.

## Verdict

Wave 1A hardening is ready for the final combined-worktree fingerprint refresh, repository-wide gate and independent manager review. The core is strategy neutral, v20-only, and grants
no broker or activation capability. Production `internal/app/engine` wiring remains unchanged, so this
change cannot activate a KR or US lane by itself.
