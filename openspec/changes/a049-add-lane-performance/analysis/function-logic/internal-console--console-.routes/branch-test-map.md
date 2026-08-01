# Branch Test Map: `Console.routes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at line 697: `fmt.Fprintf(c.out, " %s %s (%s) — %s\n", a.Kind, a.ID, a.Symbol, why)`; invariant: missing/corrupt/alternate path is explicit | `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric`, `TestPerformanceHistoryIsMethodReadOnlyMobileAccessibleAndCSPCompatible` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B2 | `if` at line 772: `// The operator overview (change console-operator-overview). It registers`; invariant: missing/corrupt/alternate path is explicit | `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric`, `TestPerformanceHistoryIsMethodReadOnlyMobileAccessibleAndCSPCompatible` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
