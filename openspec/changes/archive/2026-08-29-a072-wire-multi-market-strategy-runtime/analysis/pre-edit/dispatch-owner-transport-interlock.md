# Dispatch owner / transport interlock pre-edit map

## Safety invariant

Once a KR or US strategy lease commits `SUBMITTING`, no replacement central owner may advance the fencing epoch while that unexpired lease can still cross broker transport. Immediately before the broker call, the Gateway must prove the same owner and exact `SUBMITTING` revision are current and unexpired.

## Existing branches

- `AcquireStrategyDispatchOwner`: validates the owner identity, allocates a token, reads the current epoch, advances it with revision CAS, commits.
- `Attempt.beginStrategyDispatchTx`: proves current owner and claim authority, spends the nonce, moves the core attempt to `DISPATCH_STARTED`, and atomically moves the lease to `SUBMITTING`.
- Gateway dispatch callback: re-reads decision, protection, entry gate, scheduler, and reservations, then calls broker transport.

## Required branch changes

- Owner acquisition must fail closed with a typed busy error when an unexpired `SUBMITTING/RESERVED` lease exists.
- Expired `SUBMITTING` work must not permanently prevent crash recovery; a replacement owner may advance after expiry.
- The last no-byte-sent boundary must re-read current owner plus exact lease identity/state/revision/disposition/expiry before creating the send tracker or calling transport.
- A deterministic race test must attempt takeover after the `SUBMITTING` commit and prove no epoch advance and exactly one broker call.

## Branch tests

- KR and US active `SUBMITTING` leases reject takeover.
- At the exact expiry boundary takeover is allowed for recovery.
- A stale, expired, replaced, or non-`SUBMITTING` lease fails the final transport proof.
- Gateway race seam proves the failed takeover cannot fence settlement after a broker call.
