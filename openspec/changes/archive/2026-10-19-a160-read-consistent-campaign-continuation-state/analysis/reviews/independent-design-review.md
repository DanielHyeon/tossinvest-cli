# Independent a160 design review

Verdict: PASS — final revised design. Initial HOLD and its resolved finding are preserved below. 2026-10-19, baseline 5f3506599fd9cbd40b888aa954ae880244bedf9b. No implementation/testing performed.

## P2: Permit the mandatory read-only method census update

`internal/journal/readonly_test.go:117`, TestTheReadOnlyHandleHasNoWriteMethods, enumerates the exact public method set. Every unlisted method fails. A new public ReadOnly reader therefore requires its exact name to be registered in this existing test while retaining all rejection assertions. Current proposal impact/design exclude existing-function changes and promise only a new reader/test. Clarify the narrow test-only allowlist addition as authorized; no production helper change or guard weakening is necessary. Otherwise the prescribed change necessarily fails the full gate or forces an undeclared scope expansion.

## Remaining design assessment

The ReadOnly-only, single-transaction design preserves existing authority boundaries. Exact requested account/campaign identity, normalized KR/US scope, current OPEN position and matching actual generation/projection version prevent old and current position mixing. RECONCILE/entryblocked remains inspectable. FullPlanUnavailable on every success truthfully excludes qFinal, budget reconstruction, monetary summation, sellable, decision, reservations and lease creation. Existing query/scanner reuse is feasible: positionCampaignQuery accepts sql.Tx through its current query interface; scanPosition/scanCampaignLeg/scanCampaignOrder accept row scanners. CurrentPositionCampaignCAS's two standalone queries need not be changed or reused unsafely.

Concurrency verification is feasible with a private transaction reader and deterministic test arrangement, provided the actual public wrapper's use of that same transaction path is demonstrated. A helper-only test that cannot detect a wrapper reverting to independent DB queries is insufficient. The specified behavioral mutation/fault evidence must detect transaction or scope loss, rather than merely inspect SQL text. No production hook is required by the design.

Implementation acceptance must preserve: no partial result on scan/rows/commit/context failure; exact campaign and watermark-leg linkage; deterministic sequence/order-ID ordering; valid planned leg with no watermark distinct from measured zero; missing authoritative projection version distinct from zero; current position state/gen and blocked campaign facts from one snapshot. Stored facts are not a complete replay audit. Row-content/count preservation, close/reopen and peer isolation tests are adequate bounded evidence; no whole SQLite-file hash requirement is introduced.

No additional policy decision, framework, runtime connection or widened ledger reconstruction is needed. a156/live/a157 remain excluded.

## Reviewed hashes

- `openspec/changes/a160-read-consistent-campaign-continuation-state/proposal.md`: `9e85c7ae5289829a4c20fd1a75f186074e36c3715b429481f74d14ca86ee4e78`
- `openspec/changes/a160-read-consistent-campaign-continuation-state/design.md`: `883a2096b42d285e2fe6717e054ac724c2df029b9c486110b22d68a78a7a93c8`
- `openspec/changes/a160-read-consistent-campaign-continuation-state/tasks.md`: `36fe73200de595e80a433881c6c96c39dea140d7e4616a02a10445d767eaea89`
- `openspec/changes/a160-read-consistent-campaign-continuation-state/specs/position-campaigns/spec.md`: `881ffbd12ca8cbb35982fea2934e68dccaa64372dfc99d6decabde9a1008cb32`
- `internal/journal/readonly_test.go`: `c9389b9d8b4eb40672d2f09cfe4e1cfaf60f100a1554190c5b1876d668b5f6b7`
- `internal/journal/position_campaign_cas_read.go`: `2410ff1745956a24e7ac4d86ab62b2ccde40743ad0a2f92ae6b4a27423d5af18`

## Final disposition

Manager amended Migration Plan to permit exactly one SELECT-only method-name registration and rationale in TestTheReadOnlyHandleHasNoWriteMethods, retaining assertions and requiring AST/FLM/BTM evidence. Initial P2 is resolved. The revised exact nonempty position EntryDecisionID == campaign.DecisionID guard also closes same-scope/generation but foreign-decision ambiguity. No remaining actionable design blocker; implementation and gate results remain unverified.

Updated artifact hashes (earlier hashes above identify the initial reviewed version):

- `proposal.md`: `f185c414dce9e040bc62856c83c9791303bc55800b37be6a8fb0712dc232b071`
- `design.md`: `2f65f47516f91cabff435fb1ac9d3fb72269912a534fc9027239b8c2b85ce860`
- `tasks.md`: `36fe73200de595e80a433881c6c96c39dea140d7e4616a02a10445d767eaea89`
- `specs/position-campaigns/spec.md`: `e1b3335b97fde646249b9882502d46ed489aba0fbb19a66b7b0f9d881a3e333b`
- `review.md`: `a2dd02032b231ad8c815818a8692af2e31e6181b5b1ec923480ec443fbc9ad65`
