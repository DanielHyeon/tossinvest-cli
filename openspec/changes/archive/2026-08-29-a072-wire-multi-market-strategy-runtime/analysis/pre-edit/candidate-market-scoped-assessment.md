# Candidate market-scoped assessment — pre-edit map

## Function Logic Map

1. `Summaries` retains its released all-market operator contract.
2. `summariesForMarket` uses the same decoder but adds `WHERE market=?`.
3. `Assess` asks only for its validated market and still excludes expired lives.
4. Observation history remains market-scoped as before.

## Branch Test Map

| Branch | Expected |
|---|---|
| KR assessment with many US rows | only KR summaries materialized |
| US assessment | no KR row returned |
| Existing all-market Summaries caller | unchanged ordered result |
| Invalid/empty market | empty scoped result, no cross-market fallback |
