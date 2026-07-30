## Why

`engine.NewContext` already resolves the official account before opening the journal,
but the successful response does not prime the same official client's
`X-Tossinvest-Account` sequence. Restart reconciliation therefore performs a
redundant `/api/v1/accounts` request on its first account-scoped read; a 429 on that
duplicate request aborts recovery and shuts the engine down even though account
resolution just succeeded.

## What Changes

- Define the first official account record as the only implicit default and require
  both a non-empty account number and a positive sequence before engine startup can
  bind it.
- Make a fully successful official `Accounts` read atomically retain that same
  first record's positive sequence when the client does not already have an
  explicitly configured positive sequence.
- Preserve an explicit positive `WithAccountSeq` selection for general official
  clients, while requiring engine startup to prove it matches the strict first
  record before binding the journal account.
- Add regressions for one-request priming, positive explicit-sequence
  preservation, explicit engine mismatch refusal, invalid/empty/error responses,
  recursive-lock avoidance, concurrent scoped first use, public-discovery/scoped
  contention, cached scoped-request immediacy, cancellation, discovery drift
  refusal, and account/header record identity.
- Add engine-facing coverage proving startup account resolution and restart snapshot
  collection share one account discovery.

## Capabilities

### New Capabilities

- `official-account-scope`: Defines the implicit default-account record,
  account-sequence cache transitions, concurrency, and explicit override contract
  used by every official account-scoped read and mutation.

### Modified Capabilities

- `reconciliation`: Restart recovery reuses the account identity already resolved by
  engine startup instead of rediscovering it before the first snapshot.

## Impact

- `internal/official`: account-list and lazy account-sequence cache synchronization.
- `internal/app/engine`: startup validates that the journal account number and
  official account header come from the same first record.
- The cached field scopes every official account read and mutation. Guardian,
  journal ordering, interlock clauses, mutation retry, exit behavior, and gate-flip
  behavior remain unchanged.
- No mutation retry, gate relaxation, or partial-snapshot fallback is introduced.
