# Market-local final schedule validation pre-edit map

## Safety invariant

KR and US are admitted and dispatched simultaneously but independently. A final KR authority read must not wait for or fail because the US source is blocked, and vice versa.

## Existing branch

The final dispatch closure calls the paired schedule collector, which starts both market reads and waits for both before selecting the target market.

## Required branch

Add a target-market collector that prepares and restores only the requested market at the same frozen observation time, and use it solely for the final no-byte-sent revalidation. Initial supervisor assembly remains paired and concurrent.

## Branch tests

- KR final validation completes while the US provider blocks.
- US final validation completes while the KR provider blocks.
- Target-market revision, calendar, manifest, signed generation, and expiry changes still fail closed.
