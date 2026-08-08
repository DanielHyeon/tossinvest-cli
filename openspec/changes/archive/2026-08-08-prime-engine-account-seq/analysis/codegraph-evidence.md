# CodeGraph Evidence

## Definitions and flow

- `internal/app/engine.NewContext` constructs one `official.Client` through
  `NewOrderPath`, then calls `resolveAccountRef` before journal, Gateway,
  Guardian, and interlock publication.
- Pre-edit, `resolveAccountRef` skipped records with blank `DisplayName`, while
  lazy sequence resolution always used `accounts[0].ID`. That selection mismatch
  could bind the journal to one account number and official headers to another.
- `Context.SnapshotCollector` gives the same `official.Client` to
  `execgw.OfficialOrders`, holdings, and buying-power readers.
- `official.Client.getAcct` calls `ensureAccountSeq`. Pre-edit, the earlier
  public `Accounts` call did not populate `Client.accountSeq`, so recovery's
  first order scan performed a redundant `/api/v1/accounts` call.
- `reconcile.Recovery.stableSnapshot` rejects any collector error as
  `ErrRecoveryIncomplete`; `Runtime.Run` performs recovery before loops and exits
  on that error. This fail-closed behavior is correct.

## Existing retry evidence

- `verifylive.readRetry` retries read-only 429 responses with bounded 15s/30s
  waits, but it is verification-runner policy and not part of engine recovery.
- The observed engine error is specifically
  `lazy account-seq resolution: official: rate limited`; the immediately prior
  startup account resolution already had the needed answer.
- Therefore this hotfix first defines one strict first-record default, then
  removes the duplicate request. It does not import the verification retry
  policy, retry mutations, or accept partial snapshots.

## Impact boundary

- Direct implementation targets: `Client.Accounts`,
  `Client.ensureAccountSeq`, and engine `resolveAccountRef`.
- Downstream consumers: every official account-scoped read/write benefits from a
  successfully primed client, but explicit `WithAccountSeq` and lazy-only paths
  remain unchanged.
- Engine evidence must show one `/api/v1/accounts` call across `NewContext` plus
  the first restart snapshot and the expected sequence header on orders,
  holdings, and buying power.

## Post-edit verification

- `make sdd-sync` completed the CodeGraph sync (`6 changed files`, `147 nodes`)
  and CodeGraph reports the index up to date.
- Post-edit CodeGraph context still identifies
  `Context.Recovery → reconcile.Recovery.Run → stableSnapshot` as the startup
  fail-closed path. Its impact query sees the edited official and engine
  definitions; cross-file name resolution remains best-effort, so the concrete
  call composition is locked by the CLI assembly test rather than inferred from
  that advisory edge set.
- CodeGraphContext is installed and its stats probe succeeds (4,612 files,
  9,199 functions), but its incremental `update --quiet` did not terminate
  within 120 seconds. Per `docs/WORKFLOW.md` it remains advisory; no production
  fact or gate decision relies on it.
- The actual `NewContext → Context.Recovery → Recovery.Run` regression observes
  one account discovery, two complete snapshots, and the same sequence header
  on orders, holdings, and buying power. A real orders-endpoint 429 still returns
  `ErrRecoveryIncomplete` with an empty discarded snapshot.
- The final CodeGraph impact query confirms `ensureAccountSeq` feeds every
  account-scoped read and mutation. Its selected-sequence atomic fast path keeps
  already scoped exit/cancel/place requests out of public account-list mutex
  contention; only unresolved first discovery is serialized.
- Later public account discovery that contradicts an implicit selection now
  fails closed, and deterministic race tests observe actual mutex waiters before
  releasing the first discovery.
