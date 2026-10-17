# Risk Pattern Report: `internal/strategyevidence/breakout_bar.go` — `minorFromRawDecimal`

| Rule | Location | Message |
|---|---|---|
| — | — | No configured risk pattern matched |

> Findings are review candidates, not automatic defect verdicts.


Manual note (`minorFromRawDecimal`, breakout_bar.go:822–861): integer arithmetic only (`uint64`, no float64 cast — the `go-float64-cast` rule did not fire); no I/O, goroutine, defer or panic. Nothing to classify.
