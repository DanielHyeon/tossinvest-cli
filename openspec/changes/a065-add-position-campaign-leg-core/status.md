# Status — a065-add-position-campaign-leg-core

- Date: 2026-08-04
- State: Wave 1A implementation complete; final repository-wide gate pending
- Schema ownership: journal v20 only

## Delivered

- Strategy-neutral PositionCampaign/CampaignLeg aggregate identity and complete transition tables.
- Prospective generation/version CAS with unique, non-reused token and set-once successor binding.
- Ordered legs, immutable plan/intent/attempt/order/replacement lineage and per-order cumulative watermarks.
- Same-transaction campaign fill apply between Position projection and Exit apply.
- Exact preservation of terminal predecessor late fills, aggregate-over-cap RECONCILE and retry delta zero.
- Delta-zero terminal cancellation, positive long-stop monotonic composition with full candidate provenance.
- Frozen command kinds, canonical command-key syntax and length-prefixed typed identity.
- Pure/read-only deterministic reconstruction with stable mismatch reasons and no repair/broker path.
- Authoritative Position generation/state/version CAS, decision/intent/attempt/order scope binding and immutable
  one-owner/one-successor order lineage.
- CLOSED/ambiguous late-fill preservation with deterministic reconciliation evidence and complete projection digest.
- DB-enforced append-only campaign commands/events and version-conditional v20 read-only schema preflight.
- State-machine event replay with exact campaign/leg transitions, request/version evidence, delta arithmetic,
  aggregate leg quantities and per-order cap-based remaining validation.
- Immutable strategy-decision identity checks plus projection checkpoints bound to expected generation/version
  and the durable claim row; pre-v20 Positions report explicit `LEGACY_UNKNOWN` lineage.
- Durable caller-ambiguity and authoritative quantity refusal for order links, including successor-cap remaining.

## Safety boundary

- No `internal/app/engine`, broker client, live order, runtime toggle or lane activation was changed.
- `CampaignApplierBound()` remains false in production until a later integration change explicitly wires it.
- Position remains the sole quantity and average-price projection authority.

## Verification

- PositionCampaign unit and race suites: PASS.
- Journal focused migration/CAS/hook/restart/late-fill/terminal suites and focused race: PASS.
- Journal full suite: PASS (286.182s) after round-5 replay/claim/lineage hardening.
- Focused journal race: PASS (57.877s); PositionCampaign race: PASS (1.012s).
- Focused vet: PASS.
- OpenSpec strict validation and Function Logic Map completeness: PASS.
- Production broker/toggle wiring assertion: PASS; campaign planning/replay remains disconnected.
- CodeGraph hard sync completed; CodeGraphContext advisory refresh stalled and was interrupted after 60+ seconds.
- `make sdd-check`: PASS once against a fresh hard-evidence fingerprint; advisory stale warnings retained.
  Concurrent KR/US lane edits moved the combined-worktree fingerprint afterward, so the final integration refresh
  remains pending rather than repeatedly syncing during active parallel writes.
- Final combined-worktree `make sdd-check`/`make test`/`make gate` and independent manager review remain pending task 6.3/6.4.
