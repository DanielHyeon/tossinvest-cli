# a153 gstack design review

Current verdict: DESIGN FREEZE PASS after Manager clarification. Codex-host gstack plan-engineering review; no Claude or cross-model execution claimed. This is design approval only, not implementation, tests, gate or profitability approval. Only this assigned report edited. No telemetry, configuration, memory writes, network collection or test execution.

## Scope and evidence

Reviewed current proposal/design/spec/tasks, research-first evidence and AST interface inventory references; directly read a151 evaluate/compare, bounded numeric/bar parsing and a147 mfi/delta_direction with its a145 dependency. Existing TODO live-order work is unrelated and remains excluded. Applied gstack plan-eng-review architecture, quality, tests and performance sections within the assigned read-only/report-only scope. Operational preambles, additional reviewer launches and cross-model calls are outside authorization. Separate independent review is coordinated by Manager. Impeccable: N/A, no UI.

## Findings and resolution

1. Primary ambiguity policy needed to be explicit (P2, confidence 8/10). Earlier design retained both stop-first and target-first diagnostics but said only “closed valid baseline” for cohort and bootstrap. a151 evaluate assigns stop_first to fixed.base and target_first to fixed.sensitivity; choosing either changes paired means and eligibility. Manager now states primary scoring, complete cohort, means and bootstrap use only fixed.base; sensitivity remains diagnostic and never chooses a favorable outcome. Resolved by rereading design Execution and costs.
2. Opportunity identity needed an exact key (P2, confidence 8/10). Earlier “signal close/confirmation” allowed interpretation that changing receipt/confirmation creates an independent candidate. Manager now keys (market, symbol, local session date, normalized signal_close_at instant), rejects changed IDs/confirmation and normalizes equivalent offsets. Resolved by rereading Input and feature contracts.

No unresolved blocker remains. Historical research text saying scope pending describes the earlier research stage; current Manager design/proposal explicitly record the user-approved narrower scope.

## Architecture and quality assessment

Minimal leaf reuse is appropriate: strict bounded caller manifest -> causal feature extraction -> unchanged a151 fixed-arm results at original/doubled slippage -> common complete-session comparison -> exclusive report. a147 MFI14 uses 18 consecutive observations and three-minute delta with 1e-9 direction tolerance. Wrapper must establish receipt causality before calling it; its old timestamp parser is not a sufficient trust boundary. Current design explicitly requires strict seconds/microsecond and bounded-offset validation on every timestamp before either reused tool runs.

The inherited a151 ATR warmup/all-three geometry restriction is explicitly part of the universe, despite scoring only fixed. No duplicated simulator or alteration of historical tools is needed. Existing decimals are bounded to 64 characters, bars to 390/78; exact execution arithmetic is preserved, while finite MFI conversion and undefined-flow behavior remain explicit. Module paths must remain repository-owned local dependencies, not caller-provided import targets.

Absent/late prerequisites remain unknown, false filters mean known no-trade contribution zero, and unresolved baseline is never zero-filled. Features remain causally independent of future execution availability. KR/KRW and US/USD are separated, session declarations consistent, input/source hashes identify bytes without claiming authenticity. Synthetic provenance cannot be upgraded through self-declaration or hashes.

## Statistical and scope assessment

The design states the estimand: candidate-mean simulated-risk-normalized contribution within a session, then equal session weighting. A false filter's paired difference is minus its baseline R; executed-trade mean R has separate denominators. Both scenarios and all three filters use the same complete-session cohort. Excluded session/candidate IDs and outcome-conditioned selection are disclosed, so complete-case success cannot masquerade as population evidence.

2000 replacement resamples, fresh seed 153, sorted dates and linear percentile interpolation make exploratory bounds deterministic. Fewer than 20 sessions is insufficient; 20 is not proof of adequacy. Dependence, missing equity and lack of authenticated holdout remain explicit. All outputs stay inconclusive/unauthorized without rankings or promotion. Cost stress recomputes the entire path; fee monotonicity is tested only with fills held fixed. Portfolio MDD, partial exits, EMA, VP and empirical holdout enforcement remain deferred rather than fabricated.

## Planned branch and user-flow coverage

```text
local JSON CLI
  -> size/schema/duplicate/time/identity/session guards -> reject without output
  -> causal candle / volume / MFI -> true / false / unknown with reasons
  -> unchanged a151, original and doubled slippage
       -> upstream excluded / unfilled / fixed invalid / censored / closed
       -> sensitivity retained separately
  -> market/date common cohort -> excluded session or complete session
  -> exact session contributions -> empty / insufficient / exploratory bootstrap
  -> report with immutable hashes -> exclusive new output or safe refusal
```

Design requires deterministic unit tests and actual CLI checks separately from aggregate gate. Implementation review should verify negative cases for malformed timestamps (including nanosecond truncation and +08:60), receipt-late/gapped MFI, unchanged/undefined flow, doji/equal-volume false, future-path mutation preserving features, changed-confirmation duplicate identity, inherited geometry exclusions, one unresolved candidate excluding its whole session, no pooling of currencies, unequal session sizes, empty cohort and bootstrap oracle. Primary/sensitivity separation must be tested with a same-bar ambiguous fixture. Stress must demonstrably change entry/size/path, not merely fees. CLI tests must preserve input bytes and reject existing targets and ancestor/leaf symlinks. These are implementation acceptance checks, not claims that tests already exist or passed.

## Performance and delivery

1000 candidates, bounded input and inherited per-session bar limits bound two evaluator passes. Bootstrap over at most 1000 session values with 2000 draws needs no framework, service, cache or database. Compute features once per candidate; avoid expanding paths or repeated parsing for every resample. Local stdlib CLI plus README/example and dedicated tests suffices; package publication, scheduling and deployment are outside scope. No performance blocker identified; validate actual bounded implementation before approval.

## Reviewed manifest

Manifest SHA-256: `30da55b3e026401f0958471791880388e0edc515f17d38fb6588f79c7d0459ca`

```text
37cd5a8773e8608a7f4161aac8623c62b593e35b4ed180a0b556c277d9bb1a31  openspec/changes/a153-evaluate-cost-aware-strategy-variants/proposal.md
fe14fd4132041aaeb0f50f243816b40ca7ef25889459632179af34e54dd7ef9f  openspec/changes/a153-evaluate-cost-aware-strategy-variants/design.md
8bfebe79acd9fd91d04db7e3015a9a2fae920b6b4c58ef8f6ffb8c5d04fa4b4b  openspec/changes/a153-evaluate-cost-aware-strategy-variants/tasks.md
9a67a790e16d0bf51c88be5e515052906d142ec98af3cb7f9d4f36e5ba9ab3cb  openspec/changes/a153-evaluate-cost-aware-strategy-variants/specs/strategy-optimization/spec.md
fb1f46ea164b4e2138e77f1eb84e10be5e640aceae23bfe6d50ff6ea6ad5bfea  openspec/changes/a153-evaluate-cost-aware-strategy-variants/analysis/research-first.md
3d7849652be408103211883888b9b6ef7a5828693f9f51d51b4408c07bbecc54  tools/a151-compare-stop-policies-offline/analysis.py
c8d802e3f6423e100353310412923c845e50464bb637e391c81c33eaeb395c75  tools/a147-jev-mfi-offline/analysis.py
fa28c8ee364ff534fdad0d3109cf5dc7252a6266942b414b5a74afc5a4b45368  tools/a145-jev-offline-analysis/analysis.py
```

## GSTACK REVIEW REPORT

| Review | Runs | Status | Findings |
| --- | --- | --- | --- |
| Codex-host gstack plan engineering | 1 plus corrective reread | PASS | 2 contract ambiguities resolved |
| Architecture / code quality | 1 | PASS | Local reuse and strict wrapper boundary appropriate |
| Test plan / performance | 1 | PASS at design scope | Acceptance branches specified; no execution claimed |
| Cross-model execution | 0 | Not run | No claim |

VERDICT: DESIGN FREEZE PASS. Implementation and full gate remain pending.
NO UNRESOLVED DECISIONS
