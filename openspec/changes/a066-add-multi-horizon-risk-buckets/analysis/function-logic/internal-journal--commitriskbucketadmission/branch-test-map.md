# Branch Test Map: `(*Journal).CommitRiskBucketAdmission`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | calculated risk refusal | existing admission validation suite | existing | GREEN |
| B2 | invalid plan | existing admission validation suite | existing | GREEN |
| B3 | preimage error | existing admission validation suite | existing | GREEN |
| B4 | begin error | existing atomic suite | existing | GREEN |
| B5 | prior transaction exists | existing replay suite | existing | GREEN |
| B6 | digest/q mismatch | existing replay suite | existing | GREEN |
| B7 | replay commit error | existing crash suite | existing | GREEN |
| B8 | transaction lookup hard error | existing replay suite | existing | GREEN |
| B9 | reservation lookup error | existing binding suite | existing | GREEN |
| B10 | reservation missing | existing binding suite | existing | GREEN |
| B11 | reservation binding mismatch | existing binding suite | existing | GREEN |
| B12 | exact-scope latch/reconcile gate before any owner lookup or insert | `TestReleasedOwnerLateFillBlocksFirstFreshAdmissionOnlyInExactMarket`, `TestRiskBucketLateFillCannotBindReopenedOwner` | first fresh owner INSERT escaped | GREEN |
| B13 | owner lookup switch | existing owner suite | existing | GREEN |
| B14 | active owner found | existing owner reuse suite | existing | GREEN |
| B15 | active owner identity conflict | existing owner conflict suite | existing | GREEN |
| B16 | owner absent | existing first admission suite | existing | GREEN |
| B17 | new owner insert fails | existing owner race suite | existing | GREEN |
| B18 | unique race becomes conflict | existing owner race suite | existing | GREEN |
| B19 | owner lookup hard error | existing atomic suite | existing | GREEN |
| B20 | owner reused | existing scale-in suite | existing | GREEN |
| B21 | state digest mismatch | existing replay suite | existing | GREEN |
| B22 | existing bucket query error | existing scale-in suite | existing | GREEN |
| B23 | scan existing buckets | existing scale-in suite | existing | GREEN |
| B24 | bucket scan error | existing scale-in suite | existing | GREEN |
| B25 | close existing bucket rows | existing scale-in suite | existing | GREEN |
| B26 | bucket count mismatch | existing scale-in suite | existing | GREEN |
| B27 | required cap iteration | existing five-dimension suite | existing | GREEN |
| B28 | bucket identity mismatch | existing scale-in suite | existing | GREEN |
| B29 | owner sequence error | existing atomic suite | existing | GREEN |
| B30 | snapshot set digest error | existing validation suite | existing | GREEN |
| B31 | decision insert error | existing atomic suite | existing | GREEN |
| B32 | per-cap write loop | existing five-dimension suite | existing | GREEN |
| B33 | policy record digest error | existing validation suite | existing | GREEN |
| B34 | policy insert error | existing collision suite | existing | GREEN |
| B35 | immutable policy collision | existing collision suite | existing | GREEN |
| B36 | snapshot record digest error | existing validation suite | existing | GREEN |
| B37 | snapshot insert error | existing atomic suite | existing | GREEN |
| B38 | immutable snapshot collision | existing collision suite | existing | GREEN |
| B39 | reservation insert error | existing atomic suite | existing | GREEN |
| B40 | state replay error | existing crash/replay suite | existing | GREEN |
| B41 | state sequence read error | existing atomic suite | existing | GREEN |
| B42 | state snapshot insert error | existing atomic suite | existing | GREEN |
| B43 | admission event insert error | existing atomic suite | existing | GREEN |
| B44 | final commit error | existing crash suite | existing | GREEN |
