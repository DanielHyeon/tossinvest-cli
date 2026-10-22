## ADDED Requirements

### Requirement: verification reconciliation reads refuse ambiguous broker responses

Verification cleanup reconciliation (absence and expiry modes) SHALL refuse, and write nothing, when a list page, a list row
or its nested condition objects carry a duplicated JSON key or trailing data after the JSON value, because the standard
decoder keeps only the last duplicate and a hidden earlier value could conceal a triggered order of another id. The
reconciliation reads SHALL accept exactly one success status contract, and the design SHALL state that contract in the same
words as the code.

#### Scenario: a duplicated trigger key in another id's row is refused

- **WHEN** a CLOSED conditional-order row naming a different id carries `first.triggeredOrderId` twice, once with an order id
  and once as `null`
- **THEN** reconciliation refuses with a schema refusal in both absence and expiry modes and appends no line to the record

#### Scenario: trailing data after a list page is refused

- **WHEN** a reconciliation list page result is followed by further JSON data
- **THEN** reconciliation refuses and appends no line to the record
