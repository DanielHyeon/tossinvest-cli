# Status — a065-add-position-campaign-leg-core

- Date: 2026-09-07 (updated; implementation 2026-08-04)
- State: Wave 1A implementation complete; task 6.4 adversarial review findings resolved
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

Corrected on 2026-09-07. The two claims struck below were true when a065 landed and false
afterwards; a072 wired the core. They are recorded rather than deleted because `review.md`
cited them as verification. See `issues.md` §0 for the measurement.

- ~~No `internal/app/engine`, broker client, live order, runtime toggle or lane activation was changed.~~
  a065 itself changed none of those. But `internal/app/engine/gateway.go` binds
  `journal.ApplyPositionCampaignFill` as the Campaign apply hook, unconditionally at start-up,
  and a072's first-leg dispatch path calls `campaignExposureBlockedInTx`. The core is live.
- ~~`CampaignApplierBound()` remains false in production until a later integration change explicitly wires it.~~
  That later change is a072 (`8022f578`), already on this branch. The guard a065 added to hold
  the claim was renamed and inverted there: it now asserts the wiring exists.
- Position remains the sole quantity and average-price projection authority. (still true —
  the `positions` writers are three and the campaign core is not among them)

Measured live surface (non-test callers, 2026-09-07):

| entry point | non-test callers |
|---|---|
| `ApplyPositionCampaignFill` | 1 (engine gateway, unconditional) |
| `campaignExposureBlockedInTx` | 3 (a072 first-leg path is live) |
| `(*Journal).LinkCampaignOrder` | 0 |
| `(*Journal).PlanCampaignLeg` | 0 |
| `(*Journal).UpdateCampaignStop` | 0 |

## Verification

- PositionCampaign unit and race suites: PASS.
- Journal focused migration/CAS/hook/restart/late-fill/terminal suites and focused race: PASS.
- Journal full suite: PASS (286.182s) after round-5 replay/claim/lineage hardening.
- Focused journal race: PASS (57.877s); PositionCampaign race: PASS (1.012s).
- Focused vet: PASS.
- OpenSpec strict validation and Function Logic Map completeness: PASS.
- ~~Production broker/toggle wiring assertion: PASS; campaign planning/replay remains disconnected.~~
  Corrected: the test named in `review.md` no longer exists under that name, and the assertion
  is now the opposite one. A verification result without a live instrument is not evidence.
- CodeGraph hard sync completed; CodeGraphContext advisory refresh stalled and was interrupted after 60+ seconds.
- `make sdd-check`: PASS once against a fresh hard-evidence fingerprint; advisory stale warnings retained.
  Concurrent KR/US lane edits moved the combined-worktree fingerprint afterward, so the final integration refresh
  remains pending rather than repeatedly syncing during active parallel writes.
- Task 6.4 independent adversarial review (four axes): initially BLOCK on all four.
  Three P0 and six P1 findings resolved; every fix carries a behavioural test, an AST test
  pinning that the rule has one implementation, and a mutation proving the test goes RED
  without the fix. Remaining debt is named in `issues.md` §4.
- Task 6.3/6.4 gate commands and their results are recorded in `review.md`.
