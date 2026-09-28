# Function Logic Map: `Journal.applyRiskBucketOwnerBindingInTx`

- Source: `internal/journal/risk_bucket_owner.go`
- AST evidence: `ast.json` (post-edit, 491–563, 17 branches, 16 returns). Pre-edit AST and map: `analysis/pre-edit/6.x-owner-bind/` (HEAD `47b48ae4`, 491–556, 16 branches, 15 returns)
- Risk scan: `risk-pattern-report.md` (no configured pattern matched)

## Why this bundle exists (6.x finding, 2026-09-28)

The 6.1 (B)(3) storage-exit census counted B2 (`campaignQuantity(fill.Delta)` failure) as a non-storage exit that
returns `nil`. The function header says "semantic gaps latch entry for this owner", and B9/B13–B16 do latch. B2 and
the error half of B3 return `nil` with **no owner bind and no latch**. A registered risk order's BUY fill then
leaves no trace in risk accounting. Manager ruling 2026-09-28: repair it in this lot, with the FLM first, the probe
promoted to RED, a minimal repair in the direction of the sibling latches, then mutation. If the repair spreads into
release semantics, stop and report.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `fill.Side` | `BUY` binds; anything else is not an exposure-raising fill | `AppliedFill` built by `RecordFill` | non-BUY returns nil (B1): nothing to bind |
| `fill.Delta` | canonical non-negative decimal | `RecordFill` delta | **pre-edit**: unreadable returns nil, no latch (B2); zero returns nil (B3, legitimate re-observation) |
| registered order → active owner | exactly one active owner scope | `risk_bucket_orders` ⋈ `risk_bucket_final_decisions` ⋈ active `risk_bucket_owners` | zero: released-owner late-fill latch (B8); many: REPLAY_MISMATCH on each + OWNER_BIND_REFUSED (B9–B12) |
| transaction | caller's fill transaction (`runApplyHooks` / strategy runtime) | caller | a returned error aborts the whole fill transaction; a latch keeps the fill |

## Post-edit branch alignment

The pre-edit rows below keep their pre-edit IDs. Post-edit: B1 unchanged. B2 (delta unreadable/negative) now
latches through `latchRiskBucketFillFailureForScope`. New **B3** is the `CompareDecimal` error, which also latches;
it is unreachable, and `FuzzA066CampaignQuantityIsComparable` pins the premise. New **B4** is the zero-delta no-op
(the second half of the old B3). Post-edit B5–B17 are pre-edit B4–B16, body-identical. Tests per row:
`branch-test-map.md`.

## Branches and early returns (pre-edit)

| Branch | Condition (AST source line) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `strings.ToUpper(strings.TrimSpace(fill.Side)) != "BUY"` (492) | none | nil | a SELL fill through `runApplyHooks` (existing hook tests with SELL; not a066-specific) |
| B2 | `err != nil` after `campaignQuantity(fill.Delta)` (496) | **none — the gap** | nil | `TestA066OwnerBindUnreadableFillDeltaLatchesEntryWithoutBlockingTheFill` (RED at `47b48ae4`) |
| B3 | `err != nil \|\| positive <= 0` after `CompareDecimal(delta,"0")` (500) | none | nil | zero delta: legitimate no-op. The error half is unreachable after B2 passes: `campaignQuantity` returned a canonical decimal |
| B4 | owner query error (509) | none | err (aborts fill tx) | storage exit — structural test P1/P2 |
| B5 | `rows.Next()` loop (513) | collect keys | — | `TestRunApplyHooksBindsRiskBucketOwnerAfterCampaignInSameTransaction` |
| B6 | scan error (515) | rows.Close | err | storage exit — structural test |
| B7 | close error (521) | none | err | storage exit — structural test |
| B8 | `len(keys) == 0` (524) | `latchReleasedOwnerLateFillInTx` | its result | `TestReleasedOwnerLateFillBlocksFirstFreshAdmissionOnlyInExactMarket`, `TestRiskBucketLateFillCannotBindReopenedOwner` |
| B9 | `len(keys) != 1` (527) | latch each key REPLAY_MISMATCH + OWNER_BIND_REFUSED event | nil | multiple-active-scopes fixture (per-test coverage in the BTM) |
| B10 | range keys (528) | as B9 | — | as B9 |
| B11 | latch write error (529) | none | err | storage exit — structural test |
| B12 | state record error (532) | none | err | storage exit — structural test |
| B13 | bind error (538) | classified below | — | `TestRiskBucketFillHookLatchesBindGapWithoutReturningError` |
| B14 | not lifecycle, not semantic, not ErrNoRows (540) | none | err (aborts fill tx) | storage/driver error |
| B15 | `lifecycle == nil` — semantic/replay drift (543) | REPLAY_MISMATCH latch, no state snapshot | latch result | semantic drift fixture |
| B16 | lifecycle refusal latch error (550) | none | err | storage exit — structural test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `campaignQuantity` (position_campaign.go:2018) | canonicalize delta, refuse negative | error on non-decimal or negative | AST + source |
| `riskcalc.CompareDecimal` | positive check | error only on non-decimal input | source |
| `latchReleasedOwnerLateFillInTx` | zero active owner | storage errors propagate | source |
| `latchRiskBucketFillFailure` / `latchRiskBucketFillFailureForScope` (risk_bucket_fill.go:1135/1153) | REPLAY_MISMATCH scope latch + unknown-actual owner/reservation latch + FILL_UNACCOUNTED event | storage errors propagate; ForScope latches every **active** registered owner of this order, and none when there is none | source |
| `bindRiskBucketOwnerActualInTx` | authoritative bind | lifecycle/semantic errors latch; others abort | source |

Callers (grep of internal/journal, non-test): `apply_hook.go:285` (`runApplyHooks`, after the Campaign hook) and
`strategy_dispatch_runtime.go:1384`.

## State mutations and fallbacks

- Every semantic gap except B2 (and the unreachable error half of B3) ends in a durable latch on the active owner
  scope. The latch blocks new entry only. The fill transaction commits, so position, campaign and the exit hook
  are unaffected.
- Planned minimal repair: B2 and the error half of B3 call
  `latchRiskBucketFillFailureForScope(ctx, tx, fill, "owner bind: fill delta unreadable: …")`. That is the same latch
  RecordFill uses for ambiguous ownership. It latches the active owners registered for this order, and nothing when
  the order is not a registered risk order (identical to today for unregistered orders). No release, owner row or
  reservation amount changes. The zero-delta no-op stays.

## Safety conclusion

- Safe edit boundary: B2 and the error half of B3 only. The latch is entry-blocking only; the fill still commits
  (2.7 contract: fill detection is never delayed or rejected).
- High-risk impact: yes (fill/owner accounting). The direction is conservative: it adds a latch where there was
  none and removes no refusal.
