# Branch Test Map: `TestTheStatusColumnHeaderSaysAdoption`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | new accessible adoption header missing | TestTheStatusColumnHeaderSaysAdoption | yes — compact table initially had `관리` | yes |
| B2 | old ambiguous header remains | TestTheStatusColumnHeaderSaysAdoption | yes — compact table initially had legacy wording | yes |
