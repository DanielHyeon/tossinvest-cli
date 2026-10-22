## ADDED Requirements

### Requirement: verification cleanup reconciliation records a broker-confirmed expiry

The reconciliation operation SHALL close an outstanding verification conditional
artifact (one left after a failed cleanup) as `reconciled expired` only through an explicit
expiry mode of the reconciliation operation, chosen by the operator before
approval, and only on positive evidence from official GET reads made within the
same run that the broker ended that exact identifier as `EXPIRED` without
firing. Without the expiry mode the operation SHALL behave exactly as before.
The operation SHALL send no order, cancel or modify request; its broker reader
SHALL expose read-only GET methods only.

The expiry mode SHALL keep every common precondition of the absence
reconciliation: strict record decoding, exactly one record-selected candidate,
account, profile and market binding, human approval before any order or
conditional-order list read, positive instrument controls, two complete list
reads compared as a multiset with schema validation and row checks, the Q3
freshness bound through append admission, and the locked re-read before the
append. It SHALL NOT require the Q1 retention measurement, because it infers
nothing from absence.

In addition it SHALL read the artifact by identifier twice, once before the
first list read and once after the second, and classify each response as a
statement (HTTP 200 passing schema validation), as silent (HTTP 404 whose error
code is `conditional-order-not-found`), or as a refusal (any other outcome).
The two classifications SHALL match, and two statements SHALL agree on status,
first-leg status, triggered order identifier and second-leg presence.

The operation SHALL append only when all of the following hold, and SHALL
refuse otherwise:

- at least one current official response names the artifact's identifier: a
  by-id statement, or a CLOSED-group row with that identifier; an HTTP 404 and
  an identifier absent from the lists SHALL NOT count as evidence;
- every such response reports top-level status `EXPIRED`, an empty triggered
  order identifier and no second leg, and a by-id statement echoes the exact
  identifier, the artifact's symbol and the record's market and reports a
  first-leg status from the measured expiry vocabulary;
- the symbol's complete OPEN conditional-order group is empty, its complete
  OPEN plain-order group is empty, and no CLOSED row of another identifier
  carries a triggered order identifier or status `COMPLETED`;
- every CLOSED row's status is in the expiry-mode terminal vocabulary, which
  SHALL be a reviewed in-code value sourced from a measured receipt and is
  empty, refusing every row, until that receipt exists;
- every other record artifact sharing the artifact's chain identifier is
  already terminal.

The append SHALL be one new record line that never rewrites an earlier line.
It SHALL mark the artifact with a distinct `reconciled expired` ending, carry a
basis fingerprint over the rows and by-id statements judged, the raw statuses
and response classifications of each source, and the run's timestamps, and
SHALL carry no account number in clear and no raw response body. Every reader
of the outstanding set SHALL treat that ending as terminal, and status and
report output SHALL show it apart from cancellation, fill and reconciled
absence.

The expiry mode SHALL be registered under the operation's existing
`mutating: true` annotation, so a conversational agent does not run it.

#### Scenario: the broker names the artifact EXPIRED by identifier

- **WHEN** both by-id reads return the artifact's identifier as `EXPIRED`
  with no triggered order and no second leg, and every other condition holds
- **THEN** exactly one `reconciled expired` line is appended and the artifact
  leaves the outstanding set and the cleanup plan

#### Scenario: only the CLOSED list names the artifact

- **WHEN** both by-id reads are silent (`conditional-order-not-found`) but
  both list reads hold the identifier in the CLOSED group as `EXPIRED` and
  untriggered, and every other condition holds
- **THEN** exactly one `reconciled expired` line is appended

#### Scenario: a 404 alone is not an ending

- **WHEN** both by-id reads are silent and the CLOSED group does not hold the
  identifier
- **THEN** no line is appended and the refusal states that a 404 or an absence
  is not evidence of an ending

#### Scenario: the broker reports the artifact live or fired

- **WHEN** any response naming the identifier reports an OPEN-family status,
  `COMPLETED`, a non-empty triggered order identifier or a second leg
- **THEN** no line is appended and the artifact remains outstanding

#### Scenario: a by-id read fails otherwise

- **WHEN** a by-id read returns rate limiting, a server, authentication or
  transport error, a 404 with another or unreadable error code, or a response
  failing schema validation, or the two by-id reads disagree
- **THEN** no line is appended

#### Scenario: another order is live on the symbol

- **WHEN** the symbol's OPEN conditional-order group or OPEN plain-order group
  holds any order, or a CLOSED row of another identifier shows a fired trace
- **THEN** no line is appended

#### Scenario: the expiry vocabulary has no receipt yet

- **WHEN** every other condition holds but the expiry-mode terminal
  vocabulary is still empty
- **THEN** no line is appended

#### Scenario: an older binary reads the new line

- **WHEN** a binary without the `reconciled expired` ending reads a record
  holding that line
- **THEN** it treats the artifact as outstanding, as before the line existed

## MODIFIED Requirements

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
  `EXPIRED` is refused by this absence rule, and the refusal SHALL state that
  an expiry is an ending rather than an absence and that a broker-confirmed
  expiry is reconciled only through the explicit expiry mode (see
  "verification cleanup reconciliation records a broker-confirmed expiry");
- the complete OPEN plain-order group for the symbol contains no order (a fired
  child that has not filled rests there);
- a conditional that fired and has left the lists is excluded by the Q1 rule
  (decided 2026-09-28, option (a)): once the fired-conditional retention of the
  CLOSED group has been measured (human, in-market, read-only), an artifact
  whose age (from its outstanding line's creation time, which SHALL be non-zero)
  is within the measured retention bound is covered by the CLOSED-trace and
  open-plain-order checks above, and an older artifact SHALL be refused; the
  measurement SHALL also establish the eviction model of the CLOSED history,
  and a count-based or indeterminate model means this rule cannot be satisfied
  and the operation SHALL refuse permanently (a duration alone cannot rule out
  row-count eviction); the
  bound SHALL live only as a reviewed in-code value that is absent in production
  until measured, never as configuration; until it exists the operation SHALL
  refuse;
- every row read carries the artifact's symbol, and every row read carries the
  requested market;
- the symbol queried is the byte-exact symbol string of the artifact's own
  record line (never caller-supplied), and an instrument read for that symbol
  succeeds and echoes the same symbol (positive controls against a
  blanket-empty response);
- every list response passes schema validation: a null result payload, a
  missing or null-valued collection or pagination field, or a cursor
  inconsistent with the has-next flag refuses rather than reading as an empty
  list;
- every row read carries a status from the measured terminal vocabulary and all
  required fields; a missing or unknown status, a status inconsistent with its
  group, or any row whose second (OCO) leg is present refuses — OCO-shaped
  conditionals are not reconciled through this operation;
- the complete read set above, performed twice in succession, yields the same
  (group, identifier, status, triggered order identifier) multiset both
  times, with a duplicate (group, identifier) within one read refusing by
  itself, within the Q3 freshness bound (the RED lot SHALL fix it as a named
  conservative constant approved in review; until that
  constant exists the operation SHALL refuse). The bound SHALL cover the whole
  window from the first read's start through append admission, and the
  operation SHALL recheck freshness and the Q1 age immediately before the
  append. Before selection and again immediately before the append, every
  non-blank line of the record SHALL decode strictly; an undecodable line
  anywhere, including the final line, refuses.

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
  including the Q1 rule (the measured retention bound exists under a measured
  duration-based eviction model and covers the artifact)
- **THEN** the tool appends one distinct `reconciled-absent` event and does not
  schedule a new cleanup mutation for that artifact

#### Scenario: the Q1 retention measurement is absent

- **WHEN** every other condition holds but the Q1 fired-conditional retention
  measurement does not exist yet, or it identifies a count-based eviction
  model, or it could not determine the eviction model
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

#### Scenario: an expired artifact in absence mode points to the expiry mode

- **WHEN** reconciliation runs without the expiry mode and the artifact's
  identifier is listed in the CLOSED group as `EXPIRED`
- **THEN** no reconciliation event is appended, and the refusal states that an
  expiry is not an absence and names the expiry mode
