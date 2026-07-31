# Branch Test Map: `ExitObserver.submit`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | floor lookup errors | existing floor tests | existing | yes |
| B2 | confirmed floor is zero | existing reconcile floor tests | existing | yes |
| B3 | Guardian refuses | existing reduction tests | existing | yes |
| B4 | intent attach fails | crash ordering tests | existing | yes |
| B5 | sell intent invalid | existing market/decimal tests | existing | yes |
| B6 | confirmed mutation | existing submission tests | existing | yes |
| B7 | in-doubt mutation remains armed | existing in-doubt tests | existing | yes |
| B8 | symbol in flight releases cancelled | existing delay tests | existing | yes |
| B9 | other refusal releases refused | existing gateway refusal tests | existing | yes |
| B10 | winning concurrent proposal submits matching provenance | concurrent engine/journal test | yes | yes |
| B11 | losing proposal never reaches mutation gateway | concurrent engine/journal test | yes | yes |
