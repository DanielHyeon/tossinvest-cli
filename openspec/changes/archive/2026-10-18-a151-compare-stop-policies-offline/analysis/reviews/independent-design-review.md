# Independent a151 design review

Verdict: DESIGN FREEZE PASS. The initial P2 finding below is resolved in the reviewed revision. No remaining actionable design contradiction identified. This is design acceptance, not implementation or profitability acceptance.

## Initial finding history — P2 entry-bar open precedence (resolved)

Design Fill and cost model requires evaluation from the entry bar, but introduces open-based stop/target precedence only for subsequent bars (후속 봉). Entry uses adverse slippage, so valid geometry does not imply that the raw entry-bar open exceeds every stop.

Counterexample: tick=1, next open=100, entry slippage=2 ticks, simulated entry=102, confirmed signal low=102, structural stop=101. These satisfy stop<entry and can satisfy all caps, ATR and bounded-grid rules. An entry bar with open=100/high=103/low=100/close=102 touches the structural stop. Applying only touch pricing returns101; applying open precedence immediately after entry returns100 before exit slippage. Both readings fit the current distinction between entry and subsequent bars, producing different PnL and paired differences.

Required clarification: explicitly apply open precedence to the entry bar immediately after the simulated entry, including immediate gap-through stop exit; alternatively explicitly exclude this entry geometry across all arms. Recommend the former because it preserves the existing entry/slippage model and conservative stop execution. Add this counterexample to verification requirements. Do not infer tick-level sequence from OHLC.

## Other reviewed boundaries

KR/KRW and US/USD independent units, no FX/account return aggregation, same-session ATR15 anchor/14-TR seed, strict post-confirmation scheduled entry, common target/caps, exact fees/tax, complete three-arm cohort, separate arm sensitivity endpoints and inconclusive profitability are mutually consistent. EMA is explicitly unsupported. Caller-provided calendar and bounded tick assertions do not claim provider authenticity. No additional user scope decision is needed for the narrow entry-bar clarification.

Existing source facts checked in the preceding read-only investigation: a149 is KR-only and is not to be extended; production quantity helpers are not portfolio simulators; a148 actual realized evidence must not be fabricated for simulated R. No live dataset, implementation correctness or actual profitability is attested here.

Reviewer read five design artifacts; ran no tests, commands regenerating evidence, network calls or runtime operations. Only this report was written; other writers' files preserved.

## Resolution rereview

Manager explicitly applies open-gap precedence to every bar including entry. The added spec scenario requires immediate entry-bar exit at open less exit slippage, resolving the numeric counterexample without inventing intrabar order.

The revised timing contract distinguishes hypothetical pre-scheduled open execution from complete-bar receipt. Signal inputs and all model/budget/cap snapshots are frozen by confirmation; structural low and ATR are frozen then. Only the next scheduled open subsequently determines entry, absolute fixed/ATR stops, common target and quantity. Entry-bar HLC and later observations cannot influence those values. Outcome-bar receipt after close is expressly allowed as observation evidence, not retrospective live execution. This resolves the former apparent conflict between open execution and full-bar availability. These definitions are consistent with a pure offline model and leave no new user decision open within the approved scope.

No source or tests changed or executed by this reviewer. Initial FAIL history retained above; final verdict applies to the hashes below.

## Reviewed artifacts SHA-256

- `proposal.md`: `380f9ca931c674484a064752e1dc6da72bea38cc442a2567c44975c796618468`
- `design.md`: `80bdb4c959e04451548bad2e718eb65bfa5a895ea4d6fdebab22a1f9f06bd21d`
- `specs/strategy-optimization/spec.md`: `2bd0ee278f988450e6a319315917942a9551d4716600b7bc5291b549e3fd4662`
- `tasks.md`: `4ccf85422f211c9ef6edbfc9ef518f3a1525c6ad3b2dc2d4a6a83afffdd77d82`
- `review.md`: `ee5e1deb93c8240e0a3313962f079273520e896fa7934c27f02f2d89a48f9ece`
