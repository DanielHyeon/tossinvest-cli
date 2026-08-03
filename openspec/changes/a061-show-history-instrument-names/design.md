# a061 · Design

## Root cause

`journal.ReadOnly` correctly joins `positions.market` and `positions.symbol`, but
the schema has no instrument-name column. `Console.history` copies only those
fields into `tripRow` and `eventRow`, and the template prints only `.Symbol`.
The apparent KR/US difference is presentation: an English ticker is readable,
whereas a six-digit KR code is not.

## Decisions

### D1. Enrich at the read-model boundary

Names are current descriptive metadata, not frozen PnL evidence. The console asks
for names after the journal projection is complete and attaches them only to
request-local rows. Journal values and ordering remain untouched.

### D2. Use a narrow batch-only capability

The console receives an `InstrumentNameReader` whose only method accepts plain
`InstrumentRef` values and returns plain `InstrumentName` values. The command
adapter lazily resolves the account and builds one official client on the first
live-data screen. History and the account-scoped screens then reuse that exact
client and its OAuth token manager. The request context bounds the cold history
build and every metadata chunk. No order method or credential object crosses into
`internal/console`. Unique symbols are resolved in one logical call; the adapter
owns the official endpoint's 200-symbol chunk size.

### D3. Share one profile-scoped Open API rate budget

Optional name enrichment must not consume requests while a live verification is
starting or running. Metadata therefore takes a non-blocking kernel file lease in
the active profile's journal directory and falls back to symbols if it is busy.
CLI verification, console verification, and `verify abort` first take the existing
engine/update/verification execution flock, publish the profile run-intent marker,
then take the same rate-budget lease before broker construction and hold all three
through the operation. This gives verification priority without multiple owners
deleting one advisory marker. An abort is refused while another verifier is live;
after admission it reloads the record so its cancellation target set is current.
Failure to publish the required intent refuses the operation before any broker
construction; the older soak-only marker helper remains advisory for its existing
callers.
The lease path is derived from the active profile, never from a user-supplied
evidence-record override.

### D4. Fail open for labels, fail closed for facts

A name lookup failure never removes a journal row and never invents a name. The
symbol stays visible, and the page reports that only name enrichment failed.
HTML escaping remains the standard `html/template` behavior.

## Risks

| Risk | Mitigation |
|---|---|
| Metadata outage hides history | Preserve every journal row and raw symbol |
| A wide broker reaches the console | Narrow static capability guard and adapter method values |
| Per-row API fan-out | Deduplicate once and chunk only inside one batch seam |
| Metadata outage amplifies retries | Cache a failed attempt for one minute and keep symbols visible |
| Name markup injection | Render as template text, never `template.HTML` |
| Metadata competes with live verification | Profile-scoped cross-process lease; verification owns priority |
| Partial metadata response suppresses retries | Cache accepted names only; retry omitted keys after one minute |
