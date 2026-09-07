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
- ~~Broker/config assertion: `TestCampaignCoreHasNoProductionBrokerOrToggleWiring` PASS; production Campaign hook remains unbound.~~ **Corrected 2026-09-07 (task 6.4).** No test of that name exists in the tree: a072 (`8022f578`) renamed it and inverted its assertion, which now requires the wiring to exist. A verification result reported from an instrument that is not there is not evidence. The measured state is that `internal/app/engine/gateway.go` binds `journal.ApplyPositionCampaignFill` unconditionally at start-up. See `issues.md` §0.
- CodeGraph hard sync completed; CodeGraphContext update stalled for more than 60 seconds and was interrupted as
  advisory-only. `make sdd-check` passed once with a fresh hard-evidence fingerprint and advisory
  CodeGraphContext/GBrain stale warnings retained. Because KR/US lane changes are landing concurrently, the final
  combined-worktree fingerprint refresh/check remains an integration-stage action.
- Exit-first and non-retreating stop authority remain outside the campaign policy and are not weakened.

## Task 6.4 — independent adversarial review (2026-09-07)

Four independent axes were reviewed: D4 transition tables / D8 reconstruction, journal atomicity,
EXIT FIRST and non-retreating stops, and the dormancy claims. **All four returned BLOCK.**

Nearly every finding had one root cause: the same rule implemented in two places by two different
methods. Tests pin whichever side their author was reading, so both sides pass and both survive
mutation. Four such rules were found and each is now a single function that both the ledger writer
and the offline reconstruction call — `LegStateAfterFill`, `LatchEntryBlocked`,
`StoredOrderRemaining` and `RemainingFromCap`, all in `internal/positioncampaign/fill.go`.

Three P0 and six P1 findings were resolved. The full finding list, the resolution of each, the
measured live surface that corrects this document's dormancy claims, and the named remaining debt
are in `issues.md`.

Evidence discipline applied to every fix: a behavioural test, an AST test pinning that the rule
has exactly one implementation, and a mutation reverting the fix to prove the test goes RED.
Eleven mutations were run; all eleven were RED and every revert was verified byte-identical.

Two of the fixes change live admission behaviour and are recorded here explicitly:

- **More blocking (conservative).** A campaign bound to a position generation whose row is CLOSED
  no longer admits an exposure-raising leg. Scoping the refusal to the *bound* generation is what
  keeps the spec's own "re-entry after close" scenario working; refusing on "latest row is CLOSED"
  would have made re-entry permanently impossible.
- **Less blocking (a loosening, stated as such).** The unresolved risk-reducing predicate is now
  bounded to intents created after the current position generation opened. The unbounded form had
  no resolution path for two shapes — an intent with no attempt, and a CONFIRMED order whose only
  fill observation was refused (`terminal=0, fail_closed=1` is permanent) — so one dead 2020 intent
  blocked every future entry on that symbol. A gate that refuses all normal input is not a guard.
  The residual hazard (a SELL resting at the broker from a superseded generation) is named as debt
  D-4; the unbounded form did not meaningfully cover it either.

### Gate results (tasks 6.3 / 6.4), 2026-09-07

Repository-wide commands, run at HEAD `f90c25c8` with the fixes in the working tree:

| command | result |
|---|---|
| `openspec validate … --strict --no-interactive` | PASS (`is valid`) |
| `make sdd-check` | PASS (RC=0, after `make sdd-sync` and PM tracker regeneration) |
| `make test` | PASS (RC=0; `internal/journal` 521.759s, `internal/app/engine` 173.571s, `internal/execgw` 95.552s, `internal/positioncampaign` 0.015s) |
| `make test-seams` | PASS (RC=0) |
| `make test-race` | PASS (RC=0; `internal/journal` 533.182s) |
| `make lint` | PASS (RC=0) |
| `make vet` | PASS (RC=0) |
| `make validate` | PASS (RC=0; 61 items, 0 failed) |

`make gate CHANGE=a065-add-position-campaign-leg-core` at HEAD: **steps 1–4 OK, step 5 fails.**
This is the documented stacked-change artifact, not a defect in this work, and the split is
recorded here rather than worked around silently:

- a065's `base-commit.txt` is `c57915dd` (2026-08-03). Step 5 diffs that base against the working
  tree, so a month of unrelated changes on this branch appear as a065's own diff.
- Measured attribution: a pristine worktree at `f90c25c8` with **none** of this work applied
  produces **344** step-5 findings. The same command with all of this work applied produces
  **344**. The delta is zero, and no finding names any file this work touched.
- Change-scoped steps 1–5 were therefore run on a065's own commit slice (`75cb371a`) with every
  fix applied, where the comparison base is meaningful: tasks.md present, 0 unchecked tasks,
  no deploy pair, review.md present, and `check_analysis` reports
  `evidence complete or diff-proven exempt`.
- The seven remaining stale-hash findings at HEAD are in `apply_hook.go`, `readonly.go`,
  `schema_test.go` and `readonly_test.go` — files later changes moved and this work did not
  touch. One is a stale citation (`TestOpenReadOnlyRejectsNewerSchema` was renamed elsewhere).
  They are a065's inherited debt against a base that has drifted, not a regression from 6.4.

### Function Logic Map refresh (task 6.3)

Four bundles pin `internal/journal/position_campaign.go` and were regenerated:

| bundle | branches | line range |
|---|---|---|
| `ApplyPositionCampaignFill` | 54 → 49 | 944–1161 → 974–1195 |
| `Journal.LinkCampaignOrder` | 45 → 46 | 603–785 → 603–812 |
| `Journal.CreatePositionCampaign` | 28 → 28 | unchanged (hash only) |
| `Journal.PlanCampaignLeg` | 23 → 23 | unchanged (hash only) |

Branch IDs are positional, so renumbering by hand would silently re-point every row past the
edit. The two changed maps were renumbered by aligning the old and new `ast.json` branch
sequences with difflib, and each map records that alignment. Four branches in the fill applier
collapsed into one because the judgement they performed moved into
`positioncampaign.LegStateAfterFill`; the rows for the collapsed groups were rewritten against
their measured source lines rather than carried over.

## Verdict

Wave 1A plus the task 6.4 adversarial-review fixes. The core remains strategy neutral and v20-only
and grants no broker or activation capability by itself — but it is **not dormant**: the engine
gateway binds its fill applier unconditionally and a072's first-leg path uses its admission port.
The earlier claim to the contrary is corrected above and in `status.md`.
