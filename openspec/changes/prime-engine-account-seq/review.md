# Proposal Freeze Review

## 2026-07-30 — Initial independent review

**Verdict:** NO-GO

The review identified the following release-blocking gaps:

1. Engine account resolution could skip a blank first account number while the
   official client's lazy sequence resolver still selected the first record.
   That could bind the journal and Guardian to one account while sending official
   requests with another account's sequence.
2. The official-client contract did not yet specify cache behavior for partial,
   empty, zero, negative, explicit-sequence, or error responses.
3. The lock-sharing design lacked explicit recursive-lock, concurrent-first-use,
   in-flight public/scoped contention, cancellation, and unlock regressions.
4. The engine regression needed to execute the real
   `NewContext → Context.Recovery → Recovery.Run` composition and distinguish the
   redundant account lookup from a genuine snapshot-endpoint 429.
5. The mutation-side blast radius of the shared account header was not explicit.

## Resolution

- The first official account record is now the sole implicit default. Engine
  startup requires its account number to be nonblank and its sequence positive;
  it never skips to a later record.
- A new `official-account-scope` capability specifies fully-successful priming,
  invalid/error non-priming, explicit nonzero preservation, zero-as-unresolved,
  serialization, cancellation, and account/header identity.
- The design requires a public locked `Accounts` path and a shared
  lock-assuming helper so lazy resolution cannot recursively acquire the mutex.
- RED tasks now require deadline-bounded race regressions and the actual engine
  startup/recovery composition, including a genuine snapshot 429 that remains
  fail-closed.
- The proposal now records that the shared sequence scopes official mutations as
  well as reads, while explicitly forbidding mutation retry or gate relaxation.

The revised artifacts require a fresh independent GO/NO-GO verdict before
production code is edited.

## 2026-07-30 — Second independent review

**Verdict:** NO-GO

The second review confirmed the original gaps were closed, then identified:

1. A positive explicit sequence could differ from the first record and split the
   engine's journal account number from its official request header.
2. “nonzero explicit” accidentally treated a negative sequence as authoritative.
3. The concurrency wording promised one request even when lazy sequence-only
   discovery preceded a public full-list call, although this change does not
   cache account lists.

## Second resolution

- Engine startup now requires the selected client sequence to be positive and
  exactly equal to the first account record before opening the journal.
- Only a positive explicit sequence is authoritative. A configured negative
  sequence is rejected without a request or header.
- The one-request guarantee is scoped to concurrent account-scoped first use and
  to a public `Accounts` request already in flight before a scoped caller.
  Reverse-order account-list caching is explicitly out of scope.
- Function and branch maps now describe lazy error/empty tests as new work.

A third independent freeze verdict is required before production edits.

## 2026-07-30 — Third proposal-freeze review

**Verdict:** GO

The reviewer confirmed there were no remaining P0/P1 proposal blockers after
the second resolution. In particular, the strict first-record identity,
positive exact-match engine selection, scoped concurrency promise, invalid
sequence refusal, actual recovery composition, and mutation-header blast radius
are frozen.

## Implementation review input

The implementation:

- serializes public and lazy account discovery under the existing client mutex;
- primes only zero from a fully decoded positive first record;
- rejects configured/discovered negative or zero sequences before emitting a
  header;
- compares the engine client's selected sequence with the strict first record
  before opening the journal;
- leaves recovery, snapshot rejection, Guardian, order mutation, retry, and gate
  code unchanged.

Verification available to the independent implementation reviewer:

- focused official and actual CLI assembly/recovery regressions: PASS;
- `go test -race ./internal/official`: PASS;
- tagged CLI assembly/recovery race regressions: PASS;
- `go test ./... -count=1 -timeout=180s`: PASS;
- strict OpenSpec validation and post-edit Function Logic Map check: PASS;
- actual orders-endpoint 429 remains `ErrRecoveryIncomplete` with an empty
  discarded snapshot;
- no real endpoint, order mutation, installed binary replacement, or engine
  restart was invoked.

Final implementation GO/NO-GO remains pending.

## 2026-07-30 — First implementation review

**Verdict:** NO-GO

Production logic had no P0/P1 finding. Two P1 test-adequacy gaps remained:

1. The public/scoped contention test released the public request immediately
   after spawning the scoped goroutine, so scheduling could let public discovery
   finish before the scoped call entered.
2. The explicit-negative regression counted only the final scoped endpoint, so
   it did not prove the invalid explicit value was rejected without an account
   discovery request.

## Implementation review resolution

- Added a scoped-call entry barrier, asserted the scoped call cannot complete
  while public discovery is blocked, and asserted the account-list request count
  remains exactly one before release.
- Made `/api/v1/accounts` a valid success path in the explicit-negative test and
  asserted its call count remains zero, independently of the scoped endpoint
  count.

A fresh implementation re-review and race run are required.

## 2026-07-30 — Adversarial implementation review

**Verdict:** NO-GO

The broader review found that the first resolution was still insufficient:

1. The scoped-call channel closed before the goroutine actually attempted
   `Client.mu`, so scheduler delay could still make the contention test pass
   without exercising the intended wait.
2. Holding the discovery mutex during a slow public `Accounts` request also
   delayed already selected account-scoped reads and mutations. That could add
   account-list latency to a risk-reducing cancel or exit.
3. An implicit selected sequence and a later contradictory public first record
   were not distinguished from an explicit override.
4. Malformed-2xx non-priming, matching explicit engine startup, and retry after
   invalid or empty discovery were not fully pinned.

## Adversarial resolution

- Replaced the selected sequence with an atomic value and added a pre-lock fast
  path. Only unresolved first discovery enters the mutex; already selected
  account-scoped reads and mutations do not wait for public account-list I/O.
- Added immutable positive-explicit provenance. A later public response that is
  empty, invalid, or differs from an implicit selection is rejected without
  silently changing the selected header scope.
- Replaced the fake entry channel with a bounded runtime-stack observation of
  actual `sync.Mutex.Lock → Client.ensureAccountSeq` waiters before releasing
  discovery. Failure cleanup always releases blocked test handlers.
- Added malformed response retry, empty/zero/negative retry, full implicit drift
  variants, and matching explicit sequence 7 through actual engine recovery.
- Refreshed all affected AST evidence and Function Logic Maps, including the
  modified `WithAccountSeq` option and the atomic `SelectedAccountSeq` path.

## 2026-07-30 — Final implementation review

**Verdict:** GO

The independent proposal/implementation reviewer and testing specialist report
no remaining P0/P1 findings. The maintainability specialist's final stale-wording
finding was corrected. The Codex adversarial pass found the invalid-discovery
retry gap; that gap and the analogous empty-discovery retry were added, then
confirmed by the independent reviewers.

Final verification evidence before the delivery gate:

- `go test -race ./internal/official`: PASS;
- deterministic concurrency, invalid/empty retry, drift, and cached fast-path
  subset repeated 50 times under `-race`: PASS;
- actual tagged CLI engine assembly/recovery subset repeated 20 times under
  `-race`: PASS;
- `go test ./... -count=1 -timeout=180s`: PASS;
- strict OpenSpec validation, current Function Logic Map hashes, CodeGraph hard
  fingerprint, `make sdd-check`, and `git diff --check`: PASS;
- no real broker endpoint, live order, gate flip, current binary replacement, or
  engine restart was invoked.
