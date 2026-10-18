# Branch Test Map: `ExitObserver.ObserveOnce`

Current AST branch inventory. Scenario rows name conditions; they do not claim isolated branch coverage. Focused receipts are in the implementation evidence; full regression remains pending.

| Branch | Source line and condition | Verification status |
|---|---|---|
| B1 | 465: `if o.opts.SLO != nil && o.opts.SLO.FillDetectionBehind() {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B2 | 475: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B3 | 479: `if len(states) == 0 {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B4 | 492: `if err != nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B5 | 501: `for _, state := range states {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B6 | 503: `if !ok {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B7 | 511: `if !o.quoteUsable(quote) {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
| B8 | 519: `if err := o.judge(ctx, state, quote, observation, &cycle); err != nil && cycle.Err == nil {` | Inspect branch behavior in source; focused tests exercise relevant admission, recovery and refusal paths. Individual branch coverage is not asserted here. |
