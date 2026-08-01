# Branch Test Map: `appendExitEventFromHook`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | insert failure aborts the surrounding apply transaction | existing apply-hook rollback suite | yes | yes |
| B2 | row-count lookup failure is returned | ApplyTx error contract | yes | yes |
| B3 | missing exit-state row cannot create an unbound event | existing missing-exit-state hook suite | yes | yes |

The positive generation-2 path is covered by `TestReadoptedGenerationFillAndCompletionEventsStayGenerationTwo`.
