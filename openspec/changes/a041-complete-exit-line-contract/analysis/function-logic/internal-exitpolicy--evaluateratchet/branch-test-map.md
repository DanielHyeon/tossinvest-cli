# Branch Test Map: `EvaluateRatchet`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | supplied ratchet configuration overrides defaults | existing custom config tests | existing | yes |
| B2 | invalid configuration refuses evaluation | existing config validation tests | existing | yes |
| B3 | invalid entry/initial-stop risk refuses evaluation | `TestUnusableInputsAreRefused` | existing | yes |
| B4 | invalid observed price refuses evaluation | `TestUnusableInputsAreRefused` | existing | yes |
| B5 | invalid high-water refuses evaluation | `TestUnusableInputsAreRefused` | existing | yes |
| B6 | invalid previous baseline refuses evaluation | `TestUnusableInputsAreRefused` | existing | yes |
| B7 | invalid real break-even refuses evaluation | `TestUnusableInputsAreRefused` | existing | yes |
| B8 | invalid taken ratio refuses evaluation | existing ratio validation tests | existing | yes |
| B9 | a higher observation advances the watermark | `TestTripleMonotoneProperty` | existing | yes |
| B10 | ratchet candidate computation failure refuses evaluation | existing custom config tests | existing | yes |
| B11 | reached level contributes a protection candidate | existing trigger-table tests | existing | yes |
| B12 | break-even participates only at or above its level | existing real-break-even tests | existing | yes |
| B13 | protected-stop composition failure refuses evaluation | existing composition tests | existing | yes |
| B14 | malformed composed baseline refuses evaluation | existing composition tests | existing | yes |
| B15 | malformed decision baseline refuses evaluation | existing baseline tests | existing | yes |
| B16 | observed price below newly composed protection proposes full breach | breach tests + a041 promoted-breach snapshot test | existing | yes |
| B17 | an already-pending breach is suppressed | existing duplicate breach test | existing | yes |
| B18 | partial trigger enters the once-only proposal path | existing partial tests | existing | yes |
| B19 | partial outcome switches on taken/pending/free state | existing pending/taken table tests | existing | yes |
| B20 | already-taken partial is suppressed permanently | existing once-per-position test | existing | yes |
| B21 | unresolved proposal suppresses another partial | existing pending proposal test | existing | yes |
| B22 | free partial trigger emits the configured ratio | existing partial proposal test | existing | yes |
