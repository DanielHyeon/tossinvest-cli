# Risk Pattern Report: `internal/strategyevidence/breakout_series.go` — `scanClosedBarRecord`

| Rule | Location | Message |
|---|---|---|
| — | — | No configured risk pattern matched |

> Findings are review candidates, not automatic defect verdicts.


Manual note (`scanClosedBarRecord`, breakout_series.go:118–173): reads one already-fetched row and compares fields; no SQL text, file write, HTTP, goroutine, defer, panic or float64 cast (AST/ast-grep). Nothing to classify.
