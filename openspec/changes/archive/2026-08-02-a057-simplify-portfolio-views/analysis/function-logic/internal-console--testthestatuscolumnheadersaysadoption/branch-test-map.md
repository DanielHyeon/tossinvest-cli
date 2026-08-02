# Branch Test Map: `TestTheStatusColumnHeaderSaysAdoption`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Iterate compact required headers | `TestTheStatusColumnHeaderSaysAdoption` | yes — old table lacked `라인` | yes — focused and full console suites |
| B2 | Required header is absent | `TestTheStatusColumnHeaderSaysAdoption` | yes — old table lacked `라인` | yes — focused and full console suites |
| B3 | Obsolete dedicated adoption header remains | `TestTheStatusColumnHeaderSaysAdoption` | yes — old table rendered it | yes — focused and full console suites |
| B4 | Instrument row loses adoption status | `TestTheStatusColumnHeaderSaysAdoption` | yes — status was previously isolated | yes — focused and full console suites |
