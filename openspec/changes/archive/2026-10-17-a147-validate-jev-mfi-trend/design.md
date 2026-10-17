## Context

The discovery report at `outputs/mfi-jev-20261018/report.md` describes an observational association, not the effect of feeding MFI to Jev. US calendar bar offsets were not preserved in that snapshot, so those observations were excluded. The present request authorizes this design and review only; future implementation and any paid/model/API collection require separate authorization.

## Goals / Non-Goals

**Goals:** preregister independent additional regular sessions; preserve time and provenance; relate a fixed MFI trend to existing r1 selection and 60-minute outcomes by market and session.

**Non-Goals:** alter Jev r1 prompts, model, inputs, ranking, screening, decisions, trading or runtime settings; reinterpret historical lots; infer account returns, a same-K strategy improvement, causality, or statistical sufficiency from five sessions.

## Decisions

1. **Freeze a prospective plan before observation.** Before the first observation, register one immutable manifest for the whole evaluation: all market-local session dates or an objective consecutive-session selection rule, five new regular sessions per market as an exploratory target, candidate pool/symbol selection, existing cycle schedule, price/label horizon and timing, source and timezone, missingness thresholds, maximum latest-bar age and maximum 60-minute label delay, anticipated session interruptions, interim reporting and halt rules, exclusions, and analysis/version/digests. Any later amendment must be separately versioned with its reason and distinguish pre- from post-amendment findings; never reinterpret prior observations under changed rules. Safety or operational halts remain allowed, while outcome-driven early stops cannot establish confirmatory success. Register replacement and partial-session handling in advance; do not drop bad days or silently replace them based on outcomes. Interrupted, partial, missing or insufficient sessions remain visible, and partial sessions do not count as full sessions. Five sessions are an exploration target, not an evidential power calculation or a guarantee of significance.
2. **Preserve source evidence and r1.** Record a new append-only, separately identified observation lot and bar sidecar keyed by market, session, cycle, symbol and evidence digest. Preserve raw bar close timestamp with UTC offset or UTC plus its source and original timezone metadata, OHLCV, retrieval/observation time and provenance. Link bars to the original judgment input or to identical evidence actually received before the judgment. A later API retrieval of past bars is ineligible unless as-of availability and identity are independently proven; close time and offset alone are insufficient. Keep the exact r1 JSON bytes unchanged; preserve the original timestamps with the smallest feasible separate sidecar path. Do not mutate the r1 input, questions or selection; derive MFI only after that input is fixed, and never feed labels or future bars into the earlier judgment. The existing closed-bar input alone cannot reconstruct missing calendar offsets, so do not backfill by inference or reclassify past data.
3. **Fix calculation, exclusions and comparison.** Compute 1-minute MFI14 at the last completed bar and at three bars earlier using at least 18 valid consecutive one-minute bars. Typical price is `(high+low+close)/3`; signed flow is typical price times volume according to its change from the previous bar. Equal typical prices contribute neither sign. With positive and negative flow sums P and N, MFI is `100 - 100/(1 + P/N)` when N > 0; N=0 and P>0 gives 100, P=0 and N>0 gives 0, and both zero is undefined. Delta within 1e-9 MFI points of zero is unchanged. Exclude and count missing, ambiguous timezone, future, duplicate, nonconsecutive, invalid or inadequate bars and missing/invalid 60-minute labels. The preregistered maximum latest-bar age and label delay are mandatory; missing limits deny the run, and observations exceeding them are excluded and counted. Keep markets and local sessions separate; join J1/J2 to a single candidate by state/symbol/cycle and labels via judgment ID. Do not double count questions.
4. **Report observational results.** Preserve denominators from all observed candidates through MFI-valid/excluded reasons and label-valid/missing counts by market, session, selected, nonselected and selection-unknown. Never classify a missing or ambiguous selection/join as nonselected. For selected and nonselected, report MFI rising/falling/unchanged candidate counts, missing labels, cost-free 60-minute mean and fraction of strictly positive returns. Also report within-cycle rising-minus-falling means, equal weighting across cycles that contain both groups, and the number of matched cycles. This is not a same-K strategy comparison. Disclose repeated symbols, overlapping horizons, missingness, session incompleteness and small samples. Insufficient or ambiguous evidence is inconclusive, not PASS.

Reuse the a145 offline reader and existing snapshot/manifest conventions and the discovery analysis formula where applicable. A new service, scheduler or database is unnecessary.

## Risks / Trade-offs

- [Timezone or future-data leakage] → retain offset-bearing original close timestamps and reject unprovable as-of ordering.
- [Selective sessions and dependence] → preregister before collection, retain failures and report partial sessions; separate KR/US and local dates.
- [Apparent improvement is only association] → preserve r1 and label results as observational; any input-change experiment needs a separate design and authorization.

## Migration Plan

No migration or execution is authorized now. Once separately approved, implement and test only the offline/sidecar path, independently review it, and register the prospective plan before any collection. New collection is separately authorized; append new evidence without editing prior lots. Stop and report uncertainty rather than modify existing trading behavior.

## Open Questions

Before future collection, the operator must fix specific session dates, candidate rules, acceptable missingness, interruption/stop conditions, and any costs or API budget in the preregistration. None are assumed approved here.
