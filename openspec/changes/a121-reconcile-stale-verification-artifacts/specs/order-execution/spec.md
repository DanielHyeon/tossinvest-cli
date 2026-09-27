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

- the complete OPEN conditional-order group for the symbol contains no
  conditional order at all (a modify issues a new identifier and invalidates the
  old one, so a live successor would appear there under a different identifier);
  no exception is made for a successor the record might prove it owns;
- the complete CLOSED conditional-order group for the symbol does not contain the
  artifact's identifier, and contains no row of any identifier whose triggered
  order identifier is non-empty or whose status is `COMPLETED` (a successor that
  fired is listed under its own identifier); an artifact found there as
  `EXPIRED` is refused, and the refusal SHALL state that an expired artifact
  cannot be reconciled through this operation at all;
- the complete OPEN plain-order group for the symbol contains no order (a fired
  child that has not filled rests there);
- a conditional that fired and has left the lists is excluded by the rule chosen
  in design Revision 1 Q1 `[비움 — Q1]`; until that rule exists the operation
  SHALL refuse;
- every row read carries the artifact's symbol, and every row read carries the
  requested market;
- the complete read set above, performed twice in succession, yields the same
  sorted (group, identifier, status, triggered order identifier) set both times,
  within the bound chosen in design Revision 1 Q3 `[비움 — Q3]`.

The operation SHALL select exactly one candidate without accepting a
caller-supplied identifier: a conditional artifact that the record's cleanup
selection would currently offer for cleanup (so its hold is released and it is
not an unresolved M0 measurement). It SHALL refuse when zero or more than one
candidate exists.

The operation SHALL hold the same execution exclusion and rate budget a live
verification holds, and SHALL re-read the record immediately before appending
and refuse if it changed since the candidate was selected.

#### Scenario: DELETE 404 remains nonterminal

- **WHEN** cleanup receives `conditional-order-not-found` from the official API
- **THEN** the artifact remains outstanding and a later resume cannot treat the
  404 itself as cleanup success

#### Scenario: every condition of authoritative absence holds

- **WHEN** every condition listed above holds for the selected candidate,
  including the rule chosen for design Revision 1 Q1
- **THEN** the tool appends one distinct `reconciled-absent` event and does not
  schedule a new cleanup mutation for that artifact

#### Scenario: Q1 is unanswered

- **WHEN** every other condition holds but no rule for design Revision 1 Q1
  exists
- **THEN** no reconciliation event is appended

#### Scenario: a successor fired under a different identifier

- **WHEN** the symbol's CLOSED group holds a row of another identifier with a
  non-empty triggered order identifier or status `COMPLETED`
- **THEN** no reconciliation event is appended

#### Scenario: a fired child rests as a plain order

- **WHEN** the symbol's OPEN plain-order group holds any order
- **THEN** no reconciliation event is appended

#### Scenario: concurrent change to the record

- **WHEN** the record changed between candidate selection and the append
- **THEN** no reconciliation event is appended

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
masked account reference and that it equals the masked form of the reference
resolved for the current credentials (this compares the retained last digits,
not full identity); that the selected account sequence is known rather than
lazily resolved; that the credentials list exactly one account with a non-empty
display reference (several accounts, including two sharing the retained last
digits, refuse); that an explicit profile directory is given, credentials are not
supplied through the environment, and the record path is derived from that
profile directory rather than supplied as an override; and that the market is
given explicitly, which selects the record file. Any missing, mixed, or
mismatched value SHALL refuse without reading or appending.

#### Scenario: mixed or mismatched account reference

- **WHEN** the artifact's record lines carry different account references, or
  one that differs from the current credentials' reference
- **THEN** the operation refuses before any official read

#### Scenario: record override

- **WHEN** the record path is supplied as an override instead of derived from the
  credentials' profile
- **THEN** the operation refuses before any official read

#### Scenario: credentials from the environment or no explicit profile

- **WHEN** the credentials come from environment variables, or no explicit
  profile directory is given
- **THEN** the operation refuses before any official read

#### Scenario: market mismatch

- **WHEN** an official row for the symbol reports a market other than the
  requested one
- **THEN** no reconciliation event is appended

### Requirement: reconciled absence preserves audit and attestation boundaries

`reconciled-absent` SHALL be append-only, idempotent, and attributable to a
bounded official-read basis recorded only as a versioned, domain-tagged digest.
It SHALL NOT retain a raw account identifier and SHALL NOT carry request call
records, so the reads it made cannot become successful-endpoint evidence. Its
representation of the broker
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
