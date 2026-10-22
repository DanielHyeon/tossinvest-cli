## ADDED Requirements

### Requirement: Stock basic info exposes the broker security type

The official stock basic-info read (`GET /api/v1/stocks`) SHALL carry the broker's `securityType` value verbatim into its
output as `security_type` for every returned symbol, because ETF, ETN and stock rows cannot otherwise be told apart by any
consumer. Producers that do not receive the value (WTS reads, `/prices`) SHALL leave the field empty and omit it from JSON, so
their output bytes are unchanged.

#### Scenario: an ETF and a stock are told apart

- **WHEN** `tossctl ops call stocks` reads two symbols whose broker responses carry `securityType` `ETF` and `STOCK`
- **THEN** the output rows carry `security_type` `ETF` and `STOCK` respectively, unchanged from the broker value

#### Scenario: producers without the value keep their bytes

- **WHEN** a quote is produced by a read that has no `securityType` (for example `/prices`)
- **THEN** its JSON has no `security_type` key and is byte-identical to the output before this change
