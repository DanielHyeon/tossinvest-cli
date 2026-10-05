## ADDED Requirements

### Requirement: verification cleanup reconciliation requires authoritative absence

A verification artifact that remains outstanding after a failed cleanup SHALL
remain outstanding unless a dedicated reconciliation operation obtains a fresh,
complete, account-scoped official read that narrows the absence window for the
exact record-owned artifact to the refusal rules below (the list API offers no
snapshot cut, so this is bounded evidence, not proof). A cleanup DELETE error, including HTTP 404, and a generic
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
- a conditional that fired and has left the lists is excluded by the Q1 rule
  (decided 2026-09-28, option (a)): once the fired-conditional retention of the
  CLOSED group has been measured (human, in-market, read-only), an artifact
  whose age (from its outstanding line's creation time, which SHALL be non-zero)
  is within the measured retention bound is covered by the CLOSED-trace and
  open-plain-order checks above, and an older artifact SHALL be refused; the
  bound SHALL live only as a reviewed in-code value that is absent in production
  until measured, never as configuration; until it exists the operation SHALL
  refuse;
- every row read carries the artifact's symbol, and every row read carries the
  requested market;
- the symbol queried is the byte-exact symbol string of the artifact's own
  record line (never caller-supplied), and an instrument read for that symbol
  succeeds and echoes the same symbol (positive controls against a
  blanket-empty response);
- the complete read set above, performed twice in succession, yields the same
  same (group, identifier, status, triggered order identifier) multiset both
  times, with a duplicate (group, identifier) within one read refusing by
  itself, within the Q3 freshness bound (the RED lot SHALL fix it as a named
  conservative constant approved in review; until that
  constant exists the operation SHALL refuse).

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
  including the Q1 rule (the measured retention bound exists and covers the
  artifact)
- **THEN** the tool appends one distinct `reconciled-absent` event and does not
  schedule a new cleanup mutation for that artifact

#### Scenario: the Q1 retention measurement is absent

- **WHEN** every other condition holds but the Q1 fired-conditional retention
  measurement does not exist yet
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

- **WHEN** the second read of the three groups (conditional OPEN, then plain
  OPEN, then CLOSED, in that fixed order) differs from the first as a multiset
  of (group, identifier, status, triggered order identifier), or either read
  holds a duplicate (group, identifier), or the pair exceeds the freshness bound
- **THEN** no reconciliation event is appended

#### Scenario: a positive control fails

- **WHEN** every absence condition holds but the instrument read fails, or it
  echoes a different symbol, or the queried symbol string is not byte-identical
  to the artifact record line's symbol
- **THEN** no reconciliation event is appended

#### Scenario: the eligible artifact is not unique

- **WHEN** the record has zero or more than one eligible outstanding conditional
  artifact
- **THEN** the operation refuses before any order-list or conditional-order-list read (the account-list and instrument reads these checks use are the only permitted earlier reads)

### Requirement: reconciliation binds account, profile and market before reading

The reconciliation operation SHALL, before any order-list or
conditional-order-list read and before the local append (the account-list read
and the instrument read these checks themselves use are the only official
reads permitted earlier),
require that every record line mentioning the selected artifact carries the same
masked account reference and that it equals the masked form of the reference
resolved for the current credentials (this compares the retained last digits,
not full identity); that the selected account sequence is known rather than
lazily resolved; that the credentials list exactly one account with a non-empty
display reference (several accounts, including two sharing the retained last
digits, refuse); that an explicit profile directory is given, credentials are not
supplied through the environment (neither environment credential variable is
set — one being set alone also refuses), and the record path is derived from that
profile directory rather than supplied as an override; and that the market is
given explicitly, which selects the record file. Any missing, mixed, or
mismatched value SHALL refuse without reading or appending.

#### Scenario: mixed or mismatched account reference

- **WHEN** the artifact's record lines carry different account references, or
  one that differs from the current credentials' reference
- **THEN** the operation refuses before any order-list or conditional-order-list read (the account-list and instrument reads these checks use are the only permitted earlier reads)

#### Scenario: record override

- **WHEN** the record path is supplied as an override instead of derived from the
  credentials' profile
- **THEN** the operation refuses before any order-list or conditional-order-list read (the account-list and instrument reads these checks use are the only permitted earlier reads)

#### Scenario: credentials from the environment or no explicit profile

- **WHEN** the credentials come from environment variables, or exactly one of
  the two environment credential variables is set, or no explicit profile
  directory is given
- **THEN** the operation refuses before any order-list or conditional-order-list read (the account-list and instrument reads these checks use are the only permitted earlier reads)

#### Scenario: more than one account behind the credentials

- **WHEN** the credentials list several accounts with a non-empty display
  reference, or two accounts share the retained last digits
- **THEN** the operation refuses before any order-list or conditional-order-list read (the account-list and instrument reads these checks use are the only permitted earlier reads)

#### Scenario: market mismatch

- **WHEN** an official row for the symbol reports a market other than the
  requested one
- **THEN** no reconciliation event is appended

### Requirement: reconciled absence preserves audit and attestation boundaries

`reconciled-absent` SHALL be append-only, idempotent, and attributable to a
bounded official-read basis recorded only as a versioned, domain-tagged digest.
It SHALL NOT retain a raw account identifier and SHALL NOT carry request call
records, so the reads it made cannot become successful-endpoint evidence. Its
representation of the broker identifier it reconciles follows Q2 decision (a)
(2026-09-28): the reconciliation line reuses the identifier key the reconciled
artifact's own record line already carries, and it SHALL NOT add a broker
identifier that the reconciled artifact's own record line does not already
carry.
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
