# Branch Test Map: `SealVersionedMarketInput`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | caller provenance/version laundering | `TestVersionedMarketInputRejectsCallerProvenanceLaundering` | existing | existing |
| B2 | trading-day calendar missing/too short, shifted from 09:00 KST, or crosses KST day; valid regular/early close derive close-minus-45m | cutoff derivation, short-session and pinned-KRX-day tests | caller-authored window/cutoff | pass |
| B3 | iterate both required decimals | malformed VWAP and EMA9 rows | missing | pass |
| B4 | required decimal is nonpositive or malformed | malformed required-decimal table | missing malformed rows | pass |
| B5 | slope decimal malformed | malformed slope row | missing | pass |
| B6 | LVN decimal malformed | malformed LVN row | missing | pass |
| B7 | tangled score negative/malformed | negative tangled row | missing | pass |
| B8 | current price optional presence/absence | present baseline plus optional-absent success | indirect only | pass |
| B9 | present current price malformed | malformed current-price row | nonpositive lane refusal did not cover parse failure | pass |
| B10 | iterate both optional expansion/HVN decimals | first/second optional malformed rows | missing | pass |
| B11 | optional expansion/HVN presence/absence | present baseline plus optional-absent success | indirect only | pass |
| B12 | present optional decimal malformed | malformed expansion and HVN rows | missing | pass |
