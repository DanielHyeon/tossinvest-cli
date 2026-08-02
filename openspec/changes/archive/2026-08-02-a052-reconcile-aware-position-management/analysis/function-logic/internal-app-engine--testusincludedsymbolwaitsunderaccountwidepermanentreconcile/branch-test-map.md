# Branch Test Map: `TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `for` line 113 | `for i := 0; i < reconcile.DefaultMaxFailures; i++ {` true/entered and complementary path | TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `if` line 118 | `if err != nil {` true/entered and complementary path | TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `if` line 121 | `if i+1 == reconcile.DefaultMaxFailures && !out.Permanent {` true/entered and complementary path | TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B4 | `if` line 126 | `if rejected := h.tracker.EntryAllowed("us", "AAPL"); rejected == nil \|\|` true/entered and complementary path | TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B5 | `if` line 133 | `if cycle.Folded != 1 \|\| cycle.Adopted != 0 \|\| cycle.Unmanaged != 0 \|\| h.prices.calls != 0 {` true/entered and complementary path | TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B6 | `if` line 137 | `if p.Adopted() {` true/entered and complementary path | TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
