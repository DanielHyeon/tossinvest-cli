## ADDED Requirements

### Requirement: verification reconciliation reads accept only HTTP 200

Each official GET read of the reconciliation operation (absence and expiry modes) — the conditional-order list page, the
plain-order list page, the conditional-order read by identifier and the instrument read — SHALL succeed only when the
response that decided it, the last HTTP attempt after any token-refresh retry, has status exactly 200. Any other 2xx
status SHALL fail that read. The shared transport of order, modify and cancel requests SHALL keep its status handling.

#### Scenario: a success status other than 200 is refused

- **WHEN** any of the four reconciliation reads is answered with HTTP 201, 202, 203, 206 or another 2xx status other than
  200, including when a token-refresh retry ends on such a status after an earlier 401
- **THEN** reconciliation refuses with the refusal code that read uses for a failed read and appends no line to the record

#### Scenario: the silent not-found classification is unchanged

- **WHEN** the read by identifier is answered with HTTP 404 whose error code is `conditional-order-not-found`
- **THEN** the read is classified as silent exactly as before

#### Scenario: order requests keep their status handling

- **WHEN** an order, modify or cancel request receives any 2xx status
- **THEN** it is handled exactly as before this requirement

### Requirement: verification reconciliation reads refuse ambiguous response bodies

A reconciliation read answered with HTTP 200 SHALL fail when any JSON object anywhere in its body, the envelope included,
carries two keys the standard decoder would merge into one field — identical keys, or keys equal under case folding — or
when data follows the JSON value. Such a failure SHALL carry the refusal code that read uses for a failed read
(`read-error`, `instrument-control` or `expiry-read-error`) at that read's place in the guard order; nothing is appended.

#### Scenario: an emptied trigger key in another identifier's EXPIRED CLOSED row is refused in expiry mode

- **WHEN** in expiry mode a CLOSED conditional-order row naming a different identifier with status `EXPIRED` carries
  `first.triggeredOrderId` twice, first with an order identifier and then as an empty string, so that the standard decoder
  would read the row as never triggered
- **THEN** reconciliation refuses with `read-error` and appends no line to the record

#### Scenario: an open row hidden by a repeated collection key is refused in both modes

- **WHEN** an OPEN conditional-order page carries `conditionalOrders` twice, first with a live row and then empty, or the
  envelope carries `result` twice with the later one an empty page
- **THEN** reconciliation refuses with `read-error` in both absence and expiry modes and appends no line to the record

#### Scenario: a second leg or a row hidden by a repeated key is refused

- **WHEN** a list row carries `second` as an object and again as `null`, or a list page carries its row collection twice
  with the later one empty, or the envelope carries `result` twice
- **THEN** reconciliation refuses with the refusal code of that read and appends no line

#### Scenario: a case-variant key that would overwrite a contract key is refused

- **WHEN** a plain-order row carries `symbol` and also `Symbol` with a different value, or the envelope carries `result`
  and also `Result`
- **THEN** reconciliation refuses with the refusal code of that read and appends no line

#### Scenario: trailing data after a response is refused

- **WHEN** a reconciliation read's response body is followed by further data after its JSON value
- **THEN** reconciliation refuses and appends no line to the record

#### Scenario: a well-formed response decodes as before

- **WHEN** every reconciliation read is answered with HTTP 200 and a body whose objects carry no repeated or case-variant
  keys
- **THEN** each read returns the same decoded page, statement or instrument echo as before this requirement

### Requirement: verification reconciliation reads refuse case variants of the keys they read

A reconciliation read answered with HTTP 200 SHALL fail when any JSON object anywhere in its body carries a key that is
equal under case folding to a key some reconciliation reader reads, but is not byte-identical to it, so that every
reconciliation decoder reads the same bytes. The set of keys the readers read SHALL be kept in one place. Such a failure
SHALL carry the refusal code that read uses for a failed read, and nothing SHALL be appended.

#### Scenario: a lone case-variant trigger or second leg in a by-identifier statement is refused

- **WHEN** in expiry mode both reads by identifier answer HTTP 200 with a statement whose first leg carries
  `TriggeredOrderId` instead of `triggeredOrderId`, or which carries `Second` instead of `second`, and the CLOSED list does
  not name the artifact
- **THEN** reconciliation refuses with `expiry-read-error` and appends no line to the record

#### Scenario: a lone case-variant key in a list response is refused

- **WHEN** a list response carries `Result` instead of `result`, or a list row carries `TriggeredOrderId` instead of
  `triggeredOrderId`
- **THEN** reconciliation refuses with `read-error` and appends no line to the record
