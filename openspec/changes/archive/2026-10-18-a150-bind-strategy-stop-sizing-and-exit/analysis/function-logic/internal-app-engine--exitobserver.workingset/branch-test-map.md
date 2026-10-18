# Branch Test Map: `ExitObserver.workingSet`

Current AST branch inventory. Scenario rows name conditions; they do not claim isolated branch coverage. Focused receipts are in the implementation evidence; full regression remains pending.

| Branch | Source line and condition | Verification status |
|---|---|---|
| B1 | 552: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B2 | 556: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B3 | 560: `for _, result := range stateResults {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B4 | 565: `for _, p := range positions {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B5 | 566: `if p.State == journal.PositionClosed \|\| isZeroQuantity(p.Quantity) {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B6 | 569: `if !p.ExitEligible() {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B7 | 583: `if !p.Adopted() {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B8 | 586: `if err == nil && !protection.Legacy && ok && result.Corruption == nil && !strategyProtectionMatchesExit(protection, p, state) {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B9 | 589: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B10 | 592: `if campaign != "" {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B11 | 593: `if latchErr := o.opts.Journal.ReconcileStrategyProtection(ctx, campaign, p); latchErr != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B12 | 598: `if cycle.Err == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B13 | 602: `if !ok && protection.Policy == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B14 | 609: `if !ok {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B15 | 611: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B16 | 612: `if cycle.Err == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B17 | 617: `if opened.PositionID == "" {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B18 | 628: `if result.Corruption != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B19 | 631: `if qerr != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B20 | 632: `if cycle.Err == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B21 | 642: `if q, active, qerr := o.opts.Journal.ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq); qerr != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B22 | 647: `} else if active && !q.NeedsReJudgement() {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B23 | 643: `if cycle.Err == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B24 | 647: `} else if active && !q.NeedsReJudgement() {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B25 | 651: `} else if active {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B26 | 651: `} else if active {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B27 | 675: `if identityErr != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B28 | 678: `if qerr != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B29 | 679: `if cycle.Err == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B30 | 693: `if protection.Policy != nil && protection.PositionGeneration == p.InstanceSeq {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B31 | 694: `if cmp, compareErr := riskcalc.CompareDecimal(protection.Risk.StopPrice, state.Baseline); compareErr == nil && cmp > 0 {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
