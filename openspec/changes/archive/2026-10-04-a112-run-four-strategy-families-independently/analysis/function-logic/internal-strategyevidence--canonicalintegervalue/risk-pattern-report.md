# Risk Pattern Report: `internal/strategyevidence/breakout_bar.go` — `canonicalIntegerValue`

| Rule | Location | Message |
|---|---|---|
| — | — | No configured risk pattern matched |

> Findings are review candidates, not automatic defect verdicts.


Manual note (`canonicalIntegerValue`, breakout_bar.go:788–818): `strconv` parsing and `uint64` arithmetic only; no float64 cast, I/O, goroutine, defer or panic. Nothing to classify.
