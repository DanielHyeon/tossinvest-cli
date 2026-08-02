# Branch Test Map: `Console.decoratePositionRows`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Policy adapter exists | policy runtime tests | pre-existing | yes — full console suite |
| B2 | Lifecycle list succeeds | lifecycle indexing tests | pre-existing | yes — full console suite |
| B3 | Iterate lifecycle states | multi-position lifecycle tests | pre-existing | yes — full console suite |
| B4 | Settings adapter exists | settings/adoption tests | pre-existing | yes — full console suite |
| B5 | Settings load succeeds | designation tests | pre-existing | yes — full console suite |
| B6 | Iterate KR/US rows for desired flags | a053 US + a057 parity tests | yes — dashboard path was absent | yes — focused and full suites |
| B7 | Runtime read was attempted | runtime-known/unknown tests | pre-existing | yes — full console suite |
| B8 | Iterate rows for effective projection | a057 parity test | yes — shared projection absent | yes — focused and full suites |
| B9 | Row exists in journal | managed lifecycle tests | pre-existing | yes — full console suite |
| B10 | Lifecycle state exists | lifecycle indexed tests | pre-existing | yes — full console suite |
| B11 | Lifecycle state is missing | lifecycle fail-closed tests | pre-existing | yes — full console suite |
| B12 | Reconciliation block exists | a052 reconcile block tests | pre-existing | yes — full console suite |
