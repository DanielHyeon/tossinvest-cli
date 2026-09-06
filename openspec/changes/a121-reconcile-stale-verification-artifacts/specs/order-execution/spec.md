## ADDED Requirements

### Requirement: verification cleanup reconciliation requires authoritative absence

A verification artifact that remains outstanding after a failed cleanup SHALL
remain outstanding unless a dedicated reconciliation operation obtains a fresh,
complete, account-scoped official read proving that the exact record-owned
artifact is absent. A cleanup DELETE error, including HTTP 404, and a generic
operator observation SHALL NOT by themselves be treated as cancellation, fill,
or terminal absence.

The reconciliation read SHALL cover the supported conditional-order status
groups with complete bounded pagination and exact opaque identifier comparison.
Missing/repeated cursors, stale data, read failure, wrong profile or account,
wrong object type, identity mismatch, ambiguous result, or unsupported record
schema SHALL fail closed without appending a terminal event.

#### Scenario: DELETE 404 remains nonterminal

- **WHEN** cleanup receives `conditional-order-not-found` from the official API
- **THEN** the artifact remains outstanding and a later resume cannot treat the
  404 itself as cleanup success

#### Scenario: complete official read proves absence

- **WHEN** the selected record-owned conditional is absent from a fresh,
  complete official OPEN and CLOSED read for the same profile/account and symbol
- **THEN** the tool appends one distinct `reconciled-absent` event and does not
  schedule a new cleanup mutation for that artifact

#### Scenario: absence evidence is incomplete

- **WHEN** any reconciliation page, cursor, identity, profile/account binding,
  freshness check, or object type cannot be verified
- **THEN** no reconciliation event is appended and the artifact remains
  outstanding

### Requirement: reconciled absence preserves audit and attestation boundaries

`reconciled-absent` SHALL be append-only, idempotent, and attributable to a
bounded official-read basis without retaining raw account or broker identifiers.
It SHALL preserve the original cleanup failure and SHALL NOT alter a step
verdict or count as a cancellation, fill, successful endpoint, soak proof,
capability-attestation evidence, or engine-interlock satisfaction.

#### Scenario: repeated reconciliation

- **WHEN** reconciliation is requested again for an artifact already marked
  `reconciled-absent`
- **THEN** no duplicate event or broker mutation occurs

#### Scenario: attestation remains unchanged

- **WHEN** an artifact is reconciled absent after a failed cleanup
- **THEN** existing failed conditional capability evidence remains failed and no
  new engine-start coverage is produced
