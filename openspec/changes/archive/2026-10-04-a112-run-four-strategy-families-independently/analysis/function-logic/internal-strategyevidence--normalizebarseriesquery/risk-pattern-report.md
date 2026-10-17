# Risk Pattern Report: `internal/strategyevidence/breakout_series.go` — `normalizeBarSeriesQuery`

| Rule | Location | Message |
|---|---|---|
| — | — | No configured risk pattern matched |

> Findings are review candidates, not automatic defect verdicts.


Manual note (`normalizeBarSeriesQuery`, breakout_series.go:175–207): pure query normalisation; no SQL, I/O, goroutine, defer, panic or float64 cast (AST/ast-grep). Nothing to classify.
