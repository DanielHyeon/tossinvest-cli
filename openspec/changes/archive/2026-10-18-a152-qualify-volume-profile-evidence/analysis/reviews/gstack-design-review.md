# a152 gstack design review

Verdict: DESIGN FREEZE PASS after corrective rereview. No unresolved actionable design finding remains. Codex-host gstack design review using critical/informational and security/QA criteria. No Claude/cross-model run, implementation, test execution, lifecycle operation, telemetry or external disclosure performed. Only this report edited. Base reviewed: `0f95adce191ca4c82a4abd2f29aa1da02a5c2c42`.

## Initial findings — resolved (history preserved)

### P2 — Prefix bins require zero-latency timestamps

Design defines current_prefix requested interval as [open, as_of), then requires requested interval end <= built_at <= received_at <= as_of. These constraints force end=build=receipt=as_of. A documented prefix built or delivered even slightly after its data cutoff cannot qualify at any later information time, because moving as_of also moves its required coverage end.

Recommended minimal fix: distinguish the requested data cutoff (`profile_end` or `window_end`) from information availability `as_of`; require open < end <= close, end <= as_of, executed_at < end and received_at <= as_of, with bins end <= built_at <= received_at <= as_of. Report both and require exact coverage for the requested interval. Alternatively explicitly make delayed prefix bins unsupported in v1, but that narrows the intended useful capability. Do not silently reinterpret caller coverage.end as a different cutoff.

### P2 — Task 2.3 implies out-of-scope downstream execution

Task 2.3 asks to verify that downstream experiments exclude only VP arm, while design excludes a153 execution and decisions about other arm execution. Replace with a validator-output test: unsupported emits vp_input_eligible=false and reasons, non_profile_research_blocked=false; no downstream run or wiring required.

## Resolution verified

Manager added explicit `cutoff_at` separate from `as_of`. Prefix now requires open < cutoff_at <= as_of <= close and uses [open, cutoff_at); prior requires cutoff_at=close and prior local date. Bin build/receipt bounds apply after requested cutoff and before as_of, allowing delayed delivery without future information. Trade execution bounds remain interval-based and receipts availability-based. Task 2.3 now tests emitted VP exclusion, non-profile-unblocked flag and immutable input, expressly excluding downstream execution. Rechecked design/spec/tasks for contradictory old interval requirements; none found.

## Satisfactory scope and boundaries

- Classification and eligibility are distinct. exact_trades means declared data structure, not externally authenticated execution tape. Source/coverage/license claims stay unverified; synthetic fixtures cannot become empirical VP authority.
- Corrections are explicitly unsupported rather than guessed, silently deduplicated or replayed. Event/receipt time ordering and unique event IDs are required; delayed execution order is allowed. Empty records and zero-bin totals remain not_ready, not proof of no trading.
- Coverage requires explicit interval, sequence/gap and venue/trade-type declarations. Bins require contiguous zero-inclusive grid and executed-share units. Orderbook size and OHLCV cannot be relabeled execution volume. Prior session is not asserted to be the prior trading day.
- Bounded strict JSON, exact decimals, no raw-event error disclosure, local inputs, no runtime/network/subprocess, exclusive no-follow output and no crash-durability overclaim are appropriate. Independent dataset errors must not suppress valid siblings; envelope/limit failure is a distinct CLI error.
- Minimal stdlib leaf is justified. It consumes declarations and validates evidence; it does not calculate LVN/HVN, optimize bins, collect data or integrate trading. Unsupported actual availability is normal completion, distinct from synthetic correctness.
- Research-first evidence is appropriately limited: local repository reuse analysis and explicitly skipped external package/GitHub/MCP channels. Reviewed search-first's relevant-channel guidance; no external search coverage is claimed. Read local official apiTrade fields (price/volume/timestamp/currency) and Trades comment documenting latest maximum 50; neither establishes correction or session-coverage guarantees. Existing a147/a151 paths are research consumers, not drop-in profile qualifiers. No provider/network call made. Absence conclusion must stay limited to inspected sources, not whole-disk proof.
- grill-me/grilling decision tree is documented; the cutoff frontier above is now explicitly resolved. impeccable is N/A for no UI change. Historical batch freeze in review.md is not evidence of approval for the current expanded contract.

Implementation review, independent test receipts, actual availability report and full gate remain later requirements. This verdict approves the bounded design only, not implementation correctness, empirical coverage, source authenticity or runtime activation.

## Reviewed document hashes

```text
4f227cebf2eafd86f42bf90a7936733ff78413808d080f6cfcfea08050599bf8  proposal.md
cedd8535d7d97c2bb1390a0e65af63922ce3fa633b2f4132dcbaa380933a9b63  design.md
3b164f0fa793fe3aa670148a742cf18a7ba960f844b74550fcae3ae140f4e896  specs/candidate-discovery/spec.md
cf58f2f5a2d155ee695fcf3751845e220e26431f0214612efffc7e7c5a5f13d4  tasks.md
9cb97d65082918e1ab70f5c849bc6019b808bbc9f5c40f69aad70447124d7cfc  review.md
```
