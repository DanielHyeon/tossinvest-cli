# CodeGraph baseline

Base: `cf9260f3f5e89a1d8382f2d0437a70a088d36704`. `codegraph status .` reported index up to date before pre-edit sync. Raw query, callers, callees and impact output is in `codegraph-queries.txt`; the isolated worktree source is the authority for exact behavior.

`RiskBasedQuantity` is called by `StrategyEntryQuantity`; graph impact reaches the Guardian and tests. `OpeningRangeATR` feeds breakout calculation and has dedicated tests, but its initial-TR and final rounding rules differ from a151. `CalculateSessionFeatures` is an offline a149 KR-only leaf. `RegularSession` is timezone-aware but its fixed normal hours and weekend handling are not a holiday or early-close calendar. Python `safe_output` and `write_once` are used by the existing a147 offline tool; validate suitability against the new single-manifest input/output contract before reuse. No changes to these existing functions are planned; the a151 implementation is an independent new leaf.

Pre-edit `sdd_sync.py --no-gbrain` completed exit 0 with CodeGraph and CGC; its raw output is `/tmp/tossos-a151-discovery-sync.log`. GBrain was deliberately omitted as advisory. Final sync and full gate after implementation remain separate requirements.
