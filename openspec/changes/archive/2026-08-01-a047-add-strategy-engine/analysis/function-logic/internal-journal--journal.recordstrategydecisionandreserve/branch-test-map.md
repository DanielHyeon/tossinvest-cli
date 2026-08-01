# Branch Test Map: `Journal.RecordStrategyDecisionAndReserve`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil Journal/database refuses | no direct test found | missing | unverified |
| B2 | canonical decision/reservation request build fails | journal invalid-request tests (callee coverage); no exact direct row identified | baseline | partial |
| B3 | built decision is not exposure-raising canonical RiskIntent | strategy risk-binding refusal tests | draft accepted raw JSON | pass |
| B4 | zero plan creation time is replaced with journal UTC now | `TestStrategyProductionIssuanceCommitsAuthorityReservationAndLineageTogether` | missing | pass |
| B5 | supplied plan creation time is normalized to UTC | no direct production-issuance test identified | missing | unverified |
| B6 | lineage/attempt/manifest/revision completeness check fails | divergent projection and noncanonical payload tests | weak draft | pass |
| B7 | canonical RiskIntent/full payload/denormalized binding verification fails | exhaustive projection, identity and strict payload tests | partial decode accepted divergence | pass |
| B8 | `BeginTx` fails | no injected database-begin failure test found | missing | unverified |
| B9 | reservation precheck fails | reservation/recollection refusal suite | baseline | pass |
| B10 | decision row insert fails | no exact direct production-issuance failure isolated at this statement | missing | unverified |
| B11 | reservation row insertion fails | production issuance rollback suite/callee reservation tests | baseline | partial |
| B12 | exact strategy-decision insert/collision fails | divergent projection/collision helper tests; direct full-function collision not isolated | missing | partial |
| B13 | exact strategy-attempt insert/collision fails | `TestStrategyProductionIssuanceFailureRollsBackReservationAndAllAuthority` | split transaction | pass |
| B14 | exact `DISPATCH_START` insert/collision fails | low-level exact execution replay tests; direct full-function failure not isolated | missing | partial |
| B15 | transaction commit fails | no injected commit failure test found | missing | unverified |
| Invariant | every denormalized DecisionRecord lineage field matches canonical payload | exhaustive lineage mutation table | payload subset accepted divergence | pass |
| Invariant | unknown/trailing/noncanonical payload and full-record identity mutation fail | strict payload table | partial decode accepted divergence | pass |
| Scenario | every statement succeeds and one result plus canonical receipt is returned | `TestStrategyProductionIssuanceCommitsAuthorityReservationAndLineageTogether` | missing | pass |
