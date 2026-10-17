# Risk Pattern Report: `internal/strategyevidence/breakout_series.go` — `Store.SealBarSeries`

| Rule | Location | Message |
|---|---|---|
| — | — | No configured risk pattern matched |

> Findings are review candidates, not automatic defect verdicts.


Manual note (`Store.SealBarSeries`, breakout_series.go:59–114): the SQL text is built by concatenating the static column list (`envelopeColumns("e")`) with a literal SELECT; every query value (kind, market, symbol, record-id range, cutoffs) is a bound `?` parameter — no user text enters the SQL string. SELECT-only, enforced by `TestSealBarSeriesIsStructurallySelectOnly`. No file write, HTTP, panic or float64 cast (AST/ast-grep).
