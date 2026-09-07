# Branch Test Map: `Journal.CreatePositionCampaign`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | exact retry | `TestCreatePositionCampaignProspectiveCASAndRetry` | yes | yes |
| B2 | concurrent claim | `TestConcurrentProspectiveCampaignCreationHasOneWinner` | yes | yes |
| B3 | missing decision / OPEN Position | `TestCreatePositionCampaignRejectsMissingDecisionAndOpenPosition` | yes | yes |
| B4 | authoritative CLOSED version | `TestCreatePositionCampaignUsesAuthoritativeClosedPositionVersion` | yes | yes |
| B5 | legacy Position remains unversioned | `TestMigrationV20AddsCampaignSchemaWithoutBackfillingPositions` | yes | yes |
| B6 | validation branch group | create hardening/CAS suite | yes | yes |
| B7 | validation branch group | create hardening/CAS suite | yes | yes |
| B8 | command lookup branch group | create retry suite | yes | yes |
| B9 | command lookup branch group | create retry suite | yes | yes |
| B10 | transaction branch group | create CAS suite | yes | yes |
| B11 | decision authority branch group | missing-decision test | yes | yes |
| B12 | decision authority branch group | missing-decision test | yes | yes |
| B13 | Position query branch group | OPEN/CLOSED/legacy tests | yes | yes |
| B14 | Position query branch group | OPEN/CLOSED/legacy tests | yes | yes |
| B15 | Position query branch group | OPEN/CLOSED/legacy tests | yes | yes |
| B16 | Position query branch group | OPEN/CLOSED/legacy tests | yes | yes |
| B17 | claim CAS branch group | concurrent create test | yes | yes |
| B18 | claim CAS branch group | concurrent create test | yes | yes |
| B19 | campaign insert branch group | CAS tests | yes | yes |
| B20 | campaign insert branch group | CAS tests | yes | yes |
| B21 | claim insert branch group | CAS tests | yes | yes |
| B22 | command/event append branch group | replay tests | yes | yes |
| B23 | command/event append branch group | replay tests | yes | yes |
| B24 | commit branch | crash/retry tests | yes | yes |
| B25 | final read branch | create tests | yes | yes |
| B26 | strategy lineage absent | `TestCreatePositionCampaignRequiresExactStrategyDecisionLineage/decision_without_lineage` | yes | yes |
| B27 | strategy market/symbol/lane/version/evidence mismatch | `TestCreatePositionCampaignRequiresExactStrategyDecisionLineage/cross_market_lineage` + exact create fixtures | yes | yes |
| B28 | strategy lineage query storage error | journal query error propagation contract | yes | yes |
