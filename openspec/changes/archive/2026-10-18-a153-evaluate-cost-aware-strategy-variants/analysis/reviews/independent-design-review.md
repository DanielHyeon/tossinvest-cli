# Independent a153 design review

Verdict: DESIGN FREEZE PASS. No actionable design contradiction identified. Acceptance applies to the user-approved minimum, not the original broader empirical deliverable.

## Source-grounded acceptance

The design explicitly retains a151 evaluate's 15-five-minute-bar warmup and all-three-valid geometry/quantity prerequisite despite scoring only the fixed arm. This closes the earlier hidden-universe restriction concern without changing a151. Post-entry sibling outcomes are not required for fixed scoring; base/stress fixed outcomes remain independently observable.

The MFI definition matches a147 source: last 18 consecutive one-minute bars, 14-flow values ending at indices17 and14, three-minute delta, rising strictly above1e-9. Wrapper must validate receipt/close cutoff and strict timestamp syntax before finite float conversion; a147 does not itself validate receipt times. Unknown/undefined/stale MFI cannot become false. Signal candle and immediately previous volume bar are fixed causal inputs, independent of future execution success.

Known false selection contributes zero without fabricating a zero-risk fill. Paired difference still requires closed baseline with positive own initial risk and equals minus that baseline R. Known true difference is zero. Unknown filters/unresolved baseline exclude the whole market/date session from the common cohort across all three comparisons and both base/stress scenarios; original IDs and reasons remain visible. This is deliberately outcome-conditioned and not unbiased population evidence.

Within-session candidate means followed by equal-weight session means avoid weighting busy sessions more heavily. Repeated symbols remain clustered. Market/date calendar consistency and duplicate opportunity rejection prevent fabricated independence. Paired bootstrap parameters and percentile interpolation are explicit; empty cohorts remain undefined, and twenty sessions never imply adequacy. KR/US remain separate.

Full evaluator rerun under doubled entry/exit slippage recomputes prices, sizes and paths; stress uses its own initial risk. Fee monotonicity is asserted only with held fills, avoiding a false monotonicity claim when paths change. Explicit deferral of partial exits, equity MDD, MAE/MFE, FX, EMA, authenticated holdout and recommendation reconciles the reduced scope. All results remain inconclusive/nonpromotable.

## Implementation acceptance watchpoints (not additional scope)

- Canonicalize validated timestamp instants when comparing duplicate opportunities; equivalent offset spellings must not manufacture unique samples.
- Preserve each excluded candidate's session association; an invalid candidate must not silently disappear and permit a falsely complete session.
- Validate every reused timestamp field before a151/a147 parsing. Keep prior source hashes unchanged, including a145 imported by a147.
- Report conditional selected-trade mean/count separately from full-cohort normalized contribution mean. A false filter with unresolved baseline cannot supply a known paired difference.
- Test unequal session sizes and base/stress common-cohort failures with independently derived arithmetic; do not claim hash as freeze/holdout authority.

Reviewed proposal, design, delta spec and tasks against previously inspected a151/a147 source interfaces and current research notes. The research note's pending-decision language is historical pre-approval context; current Manager design/proposal record approval. Reviewer made no source edits, ran no tests or gate, and wrote only this report. Implementation remains unverified.

## Reviewed SHA-256

- `proposal.md`: `37cd5a8773e8608a7f4161aac8623c62b593e35b4ed180a0b556c277d9bb1a31`
- `design.md`: `2e97611c8fec2703ff3d45b8a524ae1f58d6d0b8486915f25ecbca8c37f89b27`
- `specs/strategy-optimization/spec.md`: `9a67a790e16d0bf51c88be5e515052906d142ec98af3cb7f9d4f36e5ba9ab3cb`
- `tasks.md`: `8bfebe79acd9fd91d04db7e3015a9a2fae920b6b4c58ef8f6ffb8c5d04fa4b4b`

## Final clarification rereview

DESIGN FREEZE PASS retained. Primary scoring, common-cohort membership, means and bootstrap use fixed.base stop_first only in each base/stress run. Target-first sensitivity stays diagnostic and cannot select the scored path. Duplicate opportunity identity is exactly market, symbol, session date and normalized signal_close_at instant; changing confirmation does not create another opportunity. This resolves the potential alternate-time spelling and delayed-confirmation duplication without expanding scope. Prior reviewed hashes above are retained as chronology; final hashes follow.

## Final design artifacts SHA-256

- `proposal.md`: `37cd5a8773e8608a7f4161aac8623c62b593e35b4ed180a0b556c277d9bb1a31`
- `design.md`: `fe14fd4132041aaeb0f50f243816b40ca7ef25889459632179af34e54dd7ef9f`
- `specs/strategy-optimization/spec.md`: `9a67a790e16d0bf51c88be5e515052906d142ec98af3cb7f9d4f36e5ba9ab3cb`
- `tasks.md`: `8bfebe79acd9fd91d04db7e3015a9a2fae920b6b4c58ef8f6ffb8c5d04fa4b4b`
