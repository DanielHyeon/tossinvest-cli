# Independent a152 design review

Verdict: DESIGN FREEZE PASS. No blocking ambiguity or actionable contradiction identified in the reviewed bounded v1 contract. This accepts design only; implementation, tests, aggregate gate and actual source availability remain separate evidence.

## Acceptance reasoning

The envelope, dataset metadata and kind-specific records form an executable closed schema. Global limits and envelope failures are separate from per-dataset invalid results. Decimal arithmetic, explicit IDs, duplicate JSON-key refusal and independent dataset evaluation provide clear trust-boundary requirements. No proxy record conversion, arbitrary executable input, new dependency or profile framework is required.

Classification describes supplied form; eligible_unverified is only structural qualification. Source/coverage/license claims remain unverified. Synthetic inputs cannot become empirical VP inputs, production_authorized remains false, and unsupported VP leaves non-profile research unblocked. The existing strategy LVN/HVN boundary is not wired or granted authority.

Trade qualification checks requested event interval and receipt cutoff separately, permits delayed out-of-order execution timestamps while requiring nondecreasing receipt order, and rejects duplicate identities. Corrections, cancellations and unresolved/unknown policies are unsupported rather than silently reconstructed. none_declared is explicitly not proof that no correction occurred. This is an appropriately small path without a correction replay engine.

Coverage is an exact requested-interval assertion with complete status/sequence and no declared gaps; neither endpoint observations nor volume sums establish genuine completeness. Bins require executed shares, explicit zero intervals and contiguous fixed-width grid coverage, while individual trades cannot be reconstructed from bins. Empty evidence and zero-volume bins remain not_ready.

KR/US session dates and offsets use explicit IANA checks without claiming official holidays or immediate prior trading day. Prior-session and prefix information cutoff semantics are defined. Exclusive/no-follow output creation, strict bounds and immutable input handling are specified with no crash-durability overclaim.

## Feedback history and revised timing contract

The first reviewed design used one as_of for both data end and knowledge time, which forced current-prefix bin build/receipt timestamps to equal that end. This reviewer initially accepted it as a conservative limitation. Following gstack objection, Manager superseded that decision with one explicit cutoff_at field; the earlier equality limitation no longer applies.

Rereview PASS: current-prefix requires open<cutoff_at<=as_of<=close; prior-session requires cutoff_at=close with later-date as_of. Requested coverage and execution intervals end at cutoff_at, while receipt must be no later than as_of. Bins require cutoff_at<=built_at<=received_at<=as_of. Delayed available prefix snapshots can therefore qualify without importing executions outside the frozen interval or future receipts. No correction replay or new framework is introduced. Task 2.3 verifies emitted exclusion reasons/non_profile_research_blocked=false and input immutability; it does not run downstream experiments.

## Review limits

Read full current design and delta spec. Earlier read-only investigation covered proposal and existing LVN/HVN input boundary; search-first skill was read, while external registry/GitHub/MCP research was explicitly not performed by this reviewer. No source edits, tests, gate, API calls or data collection. Only this report written. Actual tape/bin availability is not established by synthetic success.

## Reviewed SHA-256

- `design.md`: `cedd8535d7d97c2bb1390a0e65af63922ce3fa633b2f4132dcbaa380933a9b63`
- `specs/candidate-discovery/spec.md`: `3b164f0fa793fe3aa670148a742cf18a7ba960f844b74550fcae3ae140f4e896`
- `proposal.md`: `4f227cebf2eafd86f42bf90a7936733ff78413808d080f6cfcfea08050599bf8`
- `tasks.md`: `cf58f2f5a2d155ee695fcf3751845e220e26431f0214612efffc7e7c5a5f13d4`
