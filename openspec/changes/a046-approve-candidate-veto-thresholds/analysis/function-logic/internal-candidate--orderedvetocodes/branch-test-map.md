# Branch Test Map: `OrderedVetoCodes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | external caller mutates returned copy | `TestOrderedVetoCodesReturnsAnExternallyMutableCopy` | missing accessor / exported mutable variable | later call returns original D3 order |
