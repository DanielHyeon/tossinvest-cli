# a160 gstack design review

2026-10-19. **PASS — scoped design review after linkage clarification.** Codex-host installed gstack design/review checklist applied directly. This is not implementation/test/gate approval or a156/live completion. No code, operational state or other reviewer's files changed. Baseline 5f3506599fd9cbd40b888aa954ae880244bedf9b.

## Resolved finding

P2 contract ambiguity: matching scope/generation alone did not explicitly bind the position's entry decision to the requested campaign. Existing ReconcileStrategyProtection verifies that identity as well. Manager added nonempty position.EntryDecisionID == campaign.DecisionID to design decision 4 and the mismatch scenario. Empty/mismatched decision must produce no usable partial result. Clarification resolved; no outstanding design blocker.

## Review assessment

A new method only on the existing ReadOnly handle, reused scanners and one transaction are sufficient; no general snapshot framework, schema, event reconstruction or engine consumer needed. Existing separate read methods do not themselves establish this multi-query snapshot guarantee. The transaction's first read defines the snapshot; returning a version does not guarantee freshness after return or create a future CAS/dispatch permit.

Current OPEN position must be latest for exact scoped identity and agree with requested and campaign actual generation plus entry decision. Missing position/version, prospective/closed binding, orphan watermark and missing required leg fail without fallback zeros or partial data. RECONCILE/entry-blocked campaigns remain readable: inspection must not clear protection or treat inability to enter as inability to observe. Stored facts are not a full event/protection audit. Planned legs without watermarks do not prove zero fills.

Stable FullPlanUnavailable on every success is appropriate under the inspected storage limitation. Do not reconstruct immutable plan/budget from a digest, sum five overlapping monetary bucket dimensions, expose Ready or emit quantity/decision/reservation/lease fields. This prevents the reader from silently becoming a later-leg admission gate.

Testing contract is substantive: public wrapper tied to actual transaction reader; deterministic concurrent writer after first read; no mixed version/position/watermark; KR/US and peer scopes; error/closed/cancelled handles; reopen identity; logical table row preservation rather than SQLite-file hash; at least one meaningful transaction/scope fault or mutation. Internal transaction-reader-only tests must not pass while public wrapper omits BeginTx. No production/public test hook is authorized.

Exact registration of one SELECT-only method in TestTheReadOnlyHandleHasNoWriteMethods is a bounded test compatibility exception, with pre/post AST and required FLM/BTM. Preserve existing forbidden-method assertions; do not turn the allowlist into a permissive pattern. No other existing-function edits are preapproved.

## Corrected strategy context and limits

Manager direct recheck supersedes the prior broad claim that strategy allocation/count/budget policy was unspecified. Canonical strategy-engine already specifies continuation 8:4:2, reversal 2:4:8 and weekly up to seven legs with immutable allocation/budget constraints. My earlier next-leg mapping review accepted that overbroad policy-gap statement; it must not be carried forward. Existing policies should be reused, not re-decided. The first-leg dispatch boundary and missing coherent persisted full-plan/risk restoration remain separate real gaps. This reader intentionally does not solve those later authorization steps.

Read relevant ReadOnly/campaign/CAS/reconcile source and Manager correction. Five prep AST source hashes match current source. No fresh test, index mutation, external service or memory write performed. Implementation review and actual gate remain future steps; a156/a157 and historical evidence remain outside scope.

## Reviewed document hashes

- `openspec/changes/a160-read-consistent-campaign-continuation-state/proposal.md`: `f185c414dce9e040bc62856c83c9791303bc55800b37be6a8fb0712dc232b071`
- `openspec/changes/a160-read-consistent-campaign-continuation-state/design.md`: `2f65f47516f91cabff435fb1ac9d3fb72269912a534fc9027239b8c2b85ce860`
- `openspec/changes/a160-read-consistent-campaign-continuation-state/tasks.md`: `36fe73200de595e80a433881c6c96c39dea140d7e4616a02a10445d767eaea89`
- `openspec/changes/a160-read-consistent-campaign-continuation-state/specs/position-campaigns/spec.md`: `e1b3335b97fde646249b9882502d46ed489aba0fbb19a66b7b0f9d881a3e333b`
