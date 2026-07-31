# Branch Test Map: `Journal.RecordExitJudgement`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing position | existing validation tests | existing | yes |
| B2 | invalid judgement provenance | journal provenance validation test | yes | yes |
| B3 | invalid proposal | existing proposal validation tests | existing | yes |
| B4 | proposal provenance mismatch | journal provenance validation test | yes | yes |
| B5 | transaction begin/read fails | journal failure tests | existing | yes |
| B6 | completed state | existing completed tests | existing | yes |
| B7 | high-water decreases | existing monotonicity tests | existing | yes |
| B8 | baseline decreases | existing monotonicity tests | existing | yes |
| B9 | empty ratchet level preserves current | existing ladder tests | existing | yes |
| B10 | state update fails | journal failure tests | existing | yes |
| B11 | no proposal appends state-only event | one-share state-only test | yes | yes |
| B12 | proposal arm succeeds | existing proposal tests | existing | yes |
| B13 | proposal arm loses concurrent race | concurrent engine/journal test | yes | yes |
| B14 | event append fails | journal failure tests | existing | yes |
| B15 | commit fails | journal failure tests | existing | yes |
| B16 | complete provenance reaches transaction unchanged | concurrent engine test | yes | yes |
| B17 | legacy zero provenance remains compatible | full journal suite | existing | yes |
