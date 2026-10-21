# official-read-compatibility — a163 delta

## ADDED Requirements

### Requirement: General price reads preserve the broker data timestamp

The general official price adapter SHALL preserve the `timestamp` of each `/api/v1/prices` row as a separate broker data time on the domain quote, decoded from its RFC 3339 form with its numeric offset and normalised to UTC. The adapter MUST NOT replace, shift or derive the existing received time (`FetchedAt`) from the broker data time. A null, absent, empty or unparseable `timestamp` SHALL leave the broker data time at its zero value without returning an error and MUST NOT be filled from a local clock. No additional request, retry or rate-budget cost SHALL be introduced.

#### Scenario: Both markets carry a KST offset
- **WHEN** a KR row carries `2026-08-27T19:59:59.000+09:00` and a US row carries `2026-08-28T05:28:30.000+09:00`
- **THEN** each quote's broker data time is the same absolute instant in UTC and its received time is unchanged

#### Scenario: Missing timestamp
- **WHEN** a row has a null, absent, empty or malformed `timestamp`
- **THEN** adaptation succeeds, the broker data time is zero and the received time is still set

#### Scenario: No new broker call
- **WHEN** a batch of symbols is read
- **THEN** exactly one `/api/v1/prices` request is made, as before
