# Dormant implementation evidence

Date: 2026-07-31

## Approved boundary

This implementation now contains the complete testable protection core: signed capability
verification, durable create/replace/cancel attempts, response-loss recovery, reconciliation,
the typed a041 protection-line bridge, cancel-first flatten authorization, and an official-only
adapter. It still has no production activation capability: `Activation` has no exported minter,
`cmd/` and `internal/app` do not construct the controller or adapter, and the shipped
`execgw.ProfileProtection = ProtectionUnwired` constant is unchanged. Consequently no code in
the shipped application can dispatch these broker mutations.

## CodeGraph hard evidence

The worktree index was refreshed with `make sdd-sync` before editing.

- `Attestation.Verify` impact reaches the legacy execution attestation writer/reader,
  `internal/soak.BuildAttestation`, console display, and engine interlock tests (23 symbols in
  the pre-edit CodeGraph impact). To preserve all of those callers, the protection matrix is
  a separate file and type; no existing `attest.Load`, `Save`, `Verify`, or `Attestation`
  function/type was edited.
- `ProfileProtection` resolves to the constant in `internal/execgw/protection.go`, is forwarded
  by the engine interlock, and gates `Gateway.checkProtection` before exposure-raising broker
  dispatch. It was not edited. The new dormant static test also requires it to remain
  `UNWIRED` and rejects any `cmd/` or `internal/app/` import of `internal/protection`.
- Existing conditional mutations remain the concrete official methods in
  `internal/official/conditional_writes.go`. The new `internal/protectionofficial` adapter calls
  only that official client, requires exact integer/capability/readback identity, scans bounded
  OPEN+CLOSED pages for recovery, and treats cancel disappearance or unknown lifecycle as
  `IN_DOUBT`. It imports neither WTS/hybrid transports nor `internal/trading`.
- The new protection attempt journal is additive v13 and is written before dispatch. Protection
  quantity converges from holding minus the single open-sell/local claim and refuses oversell.
  Cancel-first flatten issues only an exact-scope, exact-broker, exact-quantity, two-second,
  one-shot authorization after terminal non-triggered cancel and authoritative sellable reads.
  The authorization itself has no sell method and the dormant application has no consumer.

## Go AST / Function Logic Map decision

Function Logic Map: not-applicable — the diff adds new Go files and changes no existing Go
function body. Consequently there is no pre-existing branch, early return, mutation, fallback,
or side effect whose map could be updated. New functions are covered directly by table-driven,
adversarial, repository, static-wiring, and race tests.

The future tasks that edit official conditional writes, engine gateway/reservations,
reconciliation, or flatten remain unchecked and will require their own per-symbol AST,
Function Logic Map, Branch Test Map, risk report, and human-attested evidence before editing.

### Security follow-up mapping

The independent review required changes to functions introduced by the first dormant commit.
Before those edits, AST, Function Logic Map, Branch Test Map, and risk reports were created under
`analysis/function-logic/` for attestation parsing/validation/scope checks, saga validation and
transition, sell-claim and flatten decisions, reconciliation comparison, and repository update.
The maps record the RED cases for external evidence recomputation, file/parent integrity, strict
account grammar, state-specific saga fields, immutable persisted identity, typed exact scope,
duplicate broker IDs, ordered two-second flatten observations, and int64 boundaries. They replace
the original new-functions-only exemption for this follow-up diff; the initial commit's statement
remains an accurate record of its own pre-edit gate.

## Dormant TDD coverage

- strict version, unknown/trailing JSON, issue/expiry window, exact account/profile/market/
  session/type/order/trigger/quantity scope, both verifier tool builds and SHA-256 evidence
  digests;
- symlink, non-regular file, mode, owner, open-file identity, and size checks;
- PLANNED/REGISTERING/ACTIVE/REPLACING/IN_DOUBT transitions, response-loss quarantine,
  monotonic trigger replacement, and old/new broker identity retention;
- one-share protection and aggregate sell-claim oversell refusal;
- first-fill 1-second arm and 2-second ACTIVE boundary behavior;
- missing/duplicate/orphan/quantity/trigger discrepancy classification;
- 2-second flatten decision with terminal cancel and authoritative sellable quantity, returning
  `IN_DOUBT` for response loss, trigger race, lateness, or insufficient quantity;
- additive SQLite repository round-trip and optimistic-revision conflict;
- static absence of engine/command wiring, production HTTP/official/trading adapters, and
  paper/shadow/canary protection paths.

The security follow-up additionally covers external evidence byte absence/tampering/extras,
canonical capability-digest binding, exact evidence basename, parent/file link and permission
checks, strict protection-only account normalization, cross-account/profile/market/symbol input,
duplicate broker identity, invalid saga state-field combinations, persisted identity/state jumps,
observation ordering and identity, and overflow-safe maximum-int64 sell claims.

## Intentionally incomplete

Production policy/trust-root/envelope provisioning, an exported activation minter, app/cmd
construction, an execution-gateway readiness flip, a runtime toggle, and all LIVE handoff work
remain intentionally absent. Signed readiness can be evaluated but cannot activate the controller.
The console therefore defaults to `OFF / 지원 확인 전 사용 불가`; it accepts no free symbol,
trigger, quantity, reason, or typed confirmation, and weakening is limited to a server-issued
opaque capability plus checkbox and three-second delay.

After the `4679fe2` independent review, restart trust was tightened further: durable rows never open
entry by themselves. A current controller instance must exactly re-confirm every ACTIVE broker row;
all other states and all recovery/I/O errors keep the latch closed. Broker calls now carry bounded
one-/two-second operation budgets, and cancel/recovery compare a typed durable target rather than an
identifier alone. These runtime contracts remain dormant because the shipped application still has
no activation construction path.

The subsequent restart-inventory review removes the remaining empty-repository exception: every new
controller instance starts closed, including a zero-row database, until a bounded complete broker
inventory read proves no orphan and exact coverage for all durable ACTIVE sagas. Replacement history
is no longer correlated by client ID alone. Canonical durable mutation bodies and acknowledged
target/result IDs form a checked acyclic chain; only exact terminal, untriggered predecessors from
that chain may be filtered from recovery, reconciliation, or cancel fallback. The response-loss path
durably acknowledges the recovered result ID before activating the saga, so repeated restart and
cancel operations reconstruct the same identity without memory-only state. No UI, app wiring,
activation capability, execution profile, or broker transport boundary changed.

The final crash-window follow-up treats mutation-attempt and saga writes as two independently
durable commits. DISPATCHED/IN_DOUBT uses only a unique canonical inventory match; ACKNOWLEDGED uses
its persisted result ID; both paths complete the saga without relying on process memory. PLANNED is
not considered dispatched, and ambiguous cardinality stays closed. Recovered unknown results are
acknowledged before saga activation so another SIGKILL is idempotently recoverable. Exact expiry is
now mandatory in the current broker target and every create/replace response, cancel, recovery, and
reconciliation comparison. The official cancel adapter performs an exact detail check before DELETE
when detail is available. These changes remain confined to the dormant packages and evidence; no UI,
runtime wiring, approval surface, or official execution profile changed.

The SLA-authority follow-up keeps no caller-controlled clock at a safety boundary. Registration
compares the compatibility fill argument to the durable PLANNED saga timestamp and computes arm and
ACTIVE deadlines only from that persisted value. Flatten captures its own operation start and shares
one absolute two-second context across cancel and sellable reads. The official adapter also promotes
exact current detail from best-effort to a mandatory pre-DELETE condition: any unavailable,
ambiguous, or mismatched preflight produces zero cancel calls, while post-delete disappearance still
fails closed. No schema expansion was required because PLANNED `UpdatedAt` is written with the fill
and is not mutated until registration begins. No app, UI, toggle, profile, or activation wiring was
changed.

The pagination-completeness follow-up makes the post-delete fallback authoritative only after both
OPEN and CLOSED scans independently reach a consistent terminal page. Target discovery never
short-circuits completeness. Page-limit exhaustion, empty or non-progressing next cursors, cursor
cycles, contradictory pagination metadata, and partial lifecycle scans remain ambiguous and cannot
produce a flatten permit. This is confined to the dormant official adapter and its evidence; broker
preflight, post-delete terminal checks, deadlines, UI, toggles, wiring, and readiness remain unchanged.
