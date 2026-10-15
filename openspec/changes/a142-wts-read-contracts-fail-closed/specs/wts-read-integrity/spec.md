## ADDED Requirements

### Requirement: Order collections distinguish unknown from empty

WTS pending and completed reads SHALL reject missing, null or incorrectly typed required order collections. Explicit empty arrays SHALL remain successful. A failure in any required market response MUST NOT return a partial successful aggregate or prove order absence.

#### Scenario: Pending result has no array
- **WHEN** result is absent, null, an object or a scalar
- **THEN** pending-order read returns an error and no authoritative empty list

#### Scenario: Completed result body has no array
- **WHEN** result or result.body is absent/null or body has a wrong type in any required market response
- **THEN** completed-order range returns an error with no successful partial aggregate

#### Scenario: Empty arrays and valid orders
- **WHEN** required collections contain [] or valid existing order fixtures
- **THEN** empty reads succeed and valid order parsing remains compatible

### Requirement: Watchlist reads preserve the flat contract

Flat watchlist reads SHALL use the dedicated new-watchlists GET endpoint and retain group and item identity, currency, base/current prices and stable ascending group/item ordering with wire-order ties. Symbol SHALL remain item.code; currency SHALL be copied verbatim with missing currency empty; nullable base/close prices SHALL keep existing zero defaults. Required collection schema failures SHALL be errors. Existing mutation request contracts MUST remain unchanged.

#### Scenario: Multiple groups and currencies
- **WHEN** a valid dedicated-endpoint response contains multiple groups and KRW/USD items
- **THEN** flat output preserves each group association, identity, currency and price without calling the legacy dashboard endpoint

#### Scenario: Missing collections and empty watchlist
- **WHEN** required collections are missing/null/wrong-type, or explicitly empty
- **THEN** malformed responses fail and explicit empty arrays succeed

#### Scenario: Existing mutation consumers
- **WHEN** the shared read decoder is used alongside existing watchlist mutation operations
- **THEN** mutation methods, payloads and operation IDs retain their existing contract

#### Scenario: Optional values and sorting ties
- **WHEN** groups/items are out of ordering sequence with equal ordering values, missing currency or nullable prices
- **THEN** stable ascending sorting preserves wire-order ties, Symbol remains the product code, missing currency is empty and nullable prices remain zero
