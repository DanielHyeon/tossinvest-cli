# Branch Test Map: `Console.handleSettings`

| Branch | Scenario | Test | RED observed | GREEN observed |
| --- | --- | --- | --- | --- |
| B1 | settings seam present loads the block/verdict | settings seam tests | baseline | baseline |
| B2 | settings load error renders independently | settings error-seam tests | baseline | baseline |
| B3 | limits seam present loads the one canonical gate snapshot | limit settings tests | baseline | baseline |
| B4 | limits load error renders independently | limit error-seam tests | baseline | baseline |
| B5 | trading policy seam present loads effective policy | operating settings tests | baseline | baseline |
| B6 | trading policy error renders independently | operating error-seam tests | baseline | baseline |
| B7 | engine boot seam present loads autostart | autostart settings tests | baseline | baseline |
| B8 | engine boot load error renders independently | autostart error-seam tests | baseline | baseline |
| B9 | updater present adds inspect/receipt state without a write | system update view tests | baseline | baseline |
| Render | escaped page renders successfully | settings handler/template tests | baseline | baseline |
| A047 | fixed lane/constants/provenance and blockers render `not_configured/OFF` with no input/textarea/contenteditable/free form | strategy-runtime DOM/static tests | missing | pass |
