# Branch Test Map: `Store.queryTrades`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | SQL query error is wrapped | DB error coverage | a049 baseline | PASS |
| B2 | every joined metric row is scanned | dashboard aggregate tests | a049 baseline | PASS |
| B3 | scan mismatch/corruption returns error | schema/scan coverage | a049 baseline | PASS |
| B4 | malformed persisted trade-close or metric-observation timestamp fails closed | corrupt source timestamp test | freshness validation absent at `948e721` | PASS |
| B5 | first row initializes one trade and stable order entry | dashboard aggregation tests | a049 baseline | PASS |
| B6 | 10,001st distinct trade returns row-limit error | row-limit test | a049 baseline | PASS |
| B7 | newest persisted timestamp across joined rows is retained | source freshness maximum test | freshness absent at `948e721` | PASS |
| B8 | valid metric row populates value/status | metric dashboard tests | a049 baseline | PASS |
| B9 | markout metric uses cost-adjusted value | markout dashboard test | a049 baseline | PASS |
| B10 | non-empty source stores provenance with version | provenance test | a049 baseline | PASS |
| B11 | row iterator error returns no partial result | DB error coverage | a049 baseline | PASS |
| B12 | distinct trades return in stable filtered order | deterministic dashboard test | a049 baseline | PASS |
