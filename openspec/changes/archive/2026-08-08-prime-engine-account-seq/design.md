## Context

The engine constructs one `official.Client`, calls `Accounts` to resolve the
human account number, and later gives that same client to restart reconciliation.
`Accounts` currently returns the sequence in `domain.Account.ID` but does not
retain it. The first account-scoped order read therefore invokes lazy resolution,
causing a second `/api/v1/accounts` call immediately after the successful first
one. The observed second call returned `official.ErrRateLimited`, so the
fail-closed partial-snapshot contract stopped startup.

The journal, Guardian, entry gate, snapshot ordering, and partial-snapshot
rejection are correct and must not be relaxed.

## Goals / Non-Goals

**Goals:**

- Reuse a successful, valid first account record on the same official client.
- Bind the engine's human account reference and request sequence to that exact
  record or refuse startup.
- Keep account-sequence selection synchronized under concurrent first use.
- Preserve a positive `WithAccountSeq` as the explicit source of truth for the
  general official client, and refuse engine startup if it differs from the
  strict first record.
- Preserve lazy account resolution for callers that never call `Accounts`.
- Prove the actual engine startup/recovery composition makes one account-list
  request.

**Non-Goals:**

- Retrying mutations or weakening restart recovery.
- Accepting a partial snapshot after 429.
- Adding configurable multi-account selection; the implicit default remains the
  first record.
- Adding a general-purpose rate-limit retry policy.

## Decisions

### D1. One strict implicit default record

The first account record is the sole implicit default. It is usable only when its
human account number is non-empty and its sequence is positive. Engine startup
will not skip an unusable first record and silently bind a later account, because
the official client's lazy account-scoped behavior has always selected the first
record. Refusal is safer than splitting journal identity and request routing.

Configurable account selection is a separate feature and would have to pass one
selected record through both identities explicitly.

### D2. Cache inside the official client

`Client.Accounts` and `Client.ensureAccountSeq` will share a locked account-list
helper. Only a fully successful response with a positive first sequence primes an
empty cache. Error, empty, zero, and negative responses leave it unresolved.

This is preferable to returning an extra sequence through engine-only APIs:
the client already owns the `X-Tossinvest-Account` header and its lazy cache, and
other callers should receive the same no-duplicate-request behavior.

An explicit positive `WithAccountSeq` remains authoritative and is never
overwritten. `WithAccountSeq(0)` keeps its existing meaning: unresolved, not an
explicit selection. A negative configured value is invalid: it is not replaced
from discovery and can never be emitted as a header.

After the engine's public account discovery, startup atomically reads the
client's selected sequence and requires it to equal the positive sequence in the
first account record. Discovery publication happens under the mutex, while the
completed selection read is lock-free. This preserves explicit client selection without
allowing the engine's journal account number and official request header to name
different records. Configurable multi-account selection remains out of scope.

### D3. Hold the existing account mutex across first discovery

Both public `Accounts` discovery and lazy discovery use the same mutex. This
keeps the existing “at most one unresolved account request at a time” property and
avoids a race where startup discovery and an account-scoped read both observe
zero and send duplicate requests.

The public method calls a lock-assuming helper; lazy resolution must call the
same helper directly while holding the mutex, never re-enter the public method.
The network call was already made while holding this mutex in
`ensureAccountSeq`; extending the same serialization to `Accounts` introduces no
new lock ordering with token or HTTP state. Concurrent scoped first-use shares
one discovery. When public `Accounts` is already in flight, a following scoped
call waits for and reuses its sequence. The reverse order is not promised:
`Accounts` returns the full list, which this hotfix deliberately does not cache,
so a public list call that begins after lazy sequence-only discovery performs its
own read. Deadline, contention, and error-unlock tests guard this decision under
`-race`.

### D4. Do not add a 429 retry in this hotfix

The observed error is on a redundant request whose answer was already obtained.
Retrying it would spend two or three more calls and delay a deterministic engine
assembly defect. Genuine 429 responses from required snapshot endpoints continue
to fail closed; a separate measured policy can add bounded read retry without
mixing it into this root-cause fix.

## Risks / Trade-offs

- [An unresolved `Accounts` call holds the discovery mutex during HTTP I/O] →
  unresolved first use is serialized, but the selected sequence has an atomic
  fast path so already scoped reads and risk-reducing mutations never wait
  behind unrelated public account-list I/O. Later public discovery that differs
  from an implicit selection is rejected instead of silently splitting identity.
- [An unusable first record followed by a valid record now refuses engine startup]
  → this closes a live-account routing ambiguity; no later record is selected
  without an explicit account-selection contract.
- [The first account remains the implicit default] → require account number and
  positive sequence from the same record; an explicit positive sequence remains
  authoritative for non-engine callers, while engine startup additionally
  requires exact equality.
- [A later snapshot endpoint can still return 429] → recovery remains fail-closed;
  this hotfix eliminates only the proven duplicate account discovery.

## Migration Plan

Build and stage a new local candidate after the hotfix gate passes. The operator
performs the existing reviewed installation/restart. Rollback is the fixed
`tossctl.rollback` sibling; no config or journal migration is required.

## Open Questions

None.
