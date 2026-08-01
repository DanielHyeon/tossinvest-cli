# Branch Test Map: `exactEventResult`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | overlapping ACTIVE result from a different event kind or payload | `TestRepositoryStaleRetryCannotMasqueradeAsDifferentEventKind` | registration active falsely matched replace active | false |
| B2 | same kind/fields but different timestamp | stale retry table | subset predicate only | false |
| B3 | dispatch known/unknown kind | event-specific repository suite | generic caller event | exact case only |
| B4 | begin registration exact/mismatch | round-trip and conflict tests | arbitrary saga update | exact only |
| B5 | active result exact attempt/broker | forged lineage test | mismatched lineage accepted | exact only |
| B6 | mutation unknown exact attempt/reason | event-specific suite | no event API | exact only |
| B7 | replace begin exact pending state | event-specific suite | caller whole-state update | exact only |
| B8 | zero quantity defaults to active quantity | replace default test | n/a | exact derived result |
| B9 | replace active exact attempt/broker | event-specific suite | weak lineage | exact only |
| B10 | trigger exact broker | trigger lineage test | broker mismatch ignored | exact only |
| B11 | close exact broker | close lineage test | empty broker wildcard | exact only |
| B12 | discrepancy exact reason | event-specific suite | partial result match | exact only |
| B13 | unsupported event kind | private invariant/default test | generic public event | false |
