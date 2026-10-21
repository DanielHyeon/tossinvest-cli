# official-read-compatibility — a163 delta

## ADDED Requirements

### Requirement: General price reads preserve the broker data timestamp

The general official price adapter SHALL preserve the `timestamp` of each `/api/v1/prices` row verbatim, as a separate raw broker data time string on the domain quote, without decoding it and without adding any branch that can drop a row or return an error because of that field. The adapter MUST NOT replace, shift or derive the existing received time (`FetchedAt`) from the broker data time. A null or absent `timestamp` SHALL leave the broker data time empty and MUST NOT be filled from a local clock; a malformed `timestamp` SHALL be preserved as received. The domain quote's JSON form SHALL omit the broker data time when it is empty, so that quotes from producers that do not set it serialise byte-for-byte as before. No additional request, retry or rate-budget cost SHALL be introduced.

#### Scenario: Both markets carry a KST offset
- **WHEN** a KR row carries `2026-08-27T19:59:59.000+09:00` and a US row carries `2026-08-28T05:28:30.000+09:00`
- **THEN** each quote's broker data time is that string unchanged and its received time is unchanged

#### Scenario: Missing or malformed timestamp
- **WHEN** a row has a null, absent, empty or malformed `timestamp`
- **THEN** adaptation succeeds with the row kept, the broker data time is empty (null, absent, empty) or the received text (malformed), and the received time is still set

#### Scenario: Quotes without a broker data time serialise as before
- **WHEN** a quote produced by the WTS quote reader or the official stock-metadata adapter is written as JSON
- **THEN** the output contains no broker data time key and is byte-identical to the output before this change

#### Scenario: No new broker call
- **WHEN** a batch of symbols is read
- **THEN** exactly one `/api/v1/prices` request is made, as before
