# Risk Pattern Report: `internal/strategyevidence/breakout_bar.go` — `NewClosedBar1mEnvelope`

| Rule | Location | Message |
|---|---|---|
| — | — | No configured risk pattern matched |

> Findings are review candidates, not automatic defect verdicts.


Manual note (`NewClosedBar1mEnvelope`, breakout_bar.go:299–360): builds bytes with `encoding/json` and delegates to `NewEnvelope`; no file/HTTP I/O, goroutine, defer, panic or float64 cast (AST). Nothing to classify.
