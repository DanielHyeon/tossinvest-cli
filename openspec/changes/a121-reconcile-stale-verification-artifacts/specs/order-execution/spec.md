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

Because the official list carries no snapshot or read-cut token, absence of the
artifact's identifier SHALL NOT by itself be accepted. The operation SHALL append
only when all of the following hold for the artifact's symbol, and SHALL refuse
otherwise:

- the complete OPEN group for the symbol contains no conditional order at all
  (a modify issues a new identifier and invalidates the old one, so a live
  successor would appear there under a different identifier);
- the complete CLOSED group for the symbol does not contain the artifact's
  identifier (an identifier found there is a terminal state, not absence);
- a conditional that fired and left the lists is excluded by the rule chosen in
  design Revision 1 Q1 `[비움 — Q1]`; until that rule exists the operation SHALL
  refuse;
- the complete OPEN+CLOSED read set, performed twice in succession, yields the
  same (group, identifier, status) set both times, within the bound chosen in
  design Revision 1 Q3 `[비움 — Q3]`.

The operation SHALL select exactly one outstanding conditional artifact from the
record without accepting a caller-supplied identifier, and SHALL refuse when zero
or more than one is eligible or when the only candidate is held for a later step.

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

#### Scenario: a live successor exists under a new identifier

- **WHEN** the artifact's identifier is absent but the symbol's OPEN group holds
  any conditional order
- **THEN** no reconciliation event is appended and the artifact remains
  outstanding

#### Scenario: the identifier is found in the CLOSED group

- **WHEN** the artifact's identifier appears in the CLOSED group
- **THEN** no reconciliation event is appended

#### Scenario: the two reads disagree

- **WHEN** the second OPEN+CLOSED read differs from the first in any group,
  identifier, or status, or the pair exceeds the freshness bound
- **THEN** no reconciliation event is appended

#### Scenario: the eligible artifact is not unique

- **WHEN** the record has zero or more than one eligible outstanding conditional
  artifact
- **THEN** the operation refuses before any official read

### Requirement: reconciliation binds account, profile and market before reading

Before any official read or local append, the reconciliation operation SHALL
require that every record line mentioning the selected artifact carries the same
masked account reference and that it equals the reference resolved for the
current credentials; that the selected account sequence is known rather than
lazily resolved; that the record path is derived from the same profile root as
the credentials rather than supplied as an override; and that the requested
market names the record file and matches the market of every official row read
for the symbol. Any missing, mixed, or mismatched value SHALL refuse without
reading or appending.

#### Scenario: mixed or mismatched account reference

- **WHEN** the artifact's record lines carry different account references, or
  one that differs from the current credentials' reference
- **THEN** the operation refuses before any official read

#### Scenario: record override

- **WHEN** the record path is supplied as an override instead of derived from the
  credentials' profile
- **THEN** the operation refuses before any official read

#### Scenario: market mismatch

- **WHEN** an official row for the symbol reports a market other than the
  requested one
- **THEN** no reconciliation event is appended

### Requirement: reconciled absence preserves audit and attestation boundaries

`reconciled-absent` SHALL be append-only, idempotent, and attributable to a
bounded official-read basis recorded only as a versioned, domain-tagged digest.
It SHALL NOT retain a raw account identifier. Its representation of the broker
identifier it reconciles is chosen in design Revision 1 Q2 `[비움 — Q2]`; in
either choice it SHALL NOT add a broker identifier that the reconciled artifact's
own record line does not already carry.
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
