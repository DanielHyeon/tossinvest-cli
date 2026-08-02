# Branch Test Map: `Console.accountPanelFrom`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Iterate enriched rows for page-global facts | a057 shared projection test | yes — dashboard notice missing | yes — focused test |
| B2 | Broker holding is absent from readable journal | a057 shared projection test | yes — notice missing | yes — focused test |
| B3 | Iterate configured markets | existing overview split tests | pre-existing | yes — full console suite |
| B4 | Market cache is present | existing cold/failed broker overview tests | pre-existing | yes — full console suite |
| B5 | Iterate joined holding rows | a057 shared projection test | yes — dashboard discarded rows | yes — focused test |
| B6 | Row belongs to current market | existing two-market guard | pre-existing | yes — full console suite |
| B7 | Classify management state | existing overview split tests | pre-existing | yes — full console suite |
| B8 | Unknown management case | existing journal-failure tests | pre-existing | yes — full console suite |
| B9 | Managed case | existing managed aggregate tests | pre-existing | yes — full console suite |
| B10 | Unmanaged case | existing unmanaged aggregate tests | pre-existing | yes — full console suite |
| B11 | Market has holdings | existing aggregate tests | pre-existing | yes — full console suite |
| B12 | Market has broker total | existing known-total tests | pre-existing | yes — full console suite |
| B13 | Market total fallback | existing empty/failed tests | pre-existing | yes — full console suite |
| B14 | Other-market panel is present | unknown-market test | pre-existing | yes — full console suite |
| B15 | Iterate rows for other market | unknown-market test | pre-existing | yes — full console suite |
| B16 | Skip known KR/US row | existing two-market guard | pre-existing | yes — full console suite |
| B17 | Append named unknown-market panel | unknown-market test | pre-existing | yes — full console suite |
