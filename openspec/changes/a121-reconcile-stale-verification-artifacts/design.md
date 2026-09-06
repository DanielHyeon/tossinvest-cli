# a121 design — read-backed artifact reconciliation

## Context

The current cleanup contract deliberately records a failed conditional DELETE
and leaves the record outstanding. That is correct for any 404 by itself: the
broker may have cancelled, expired, replaced, triggered, or otherwise removed
the object. A generic human observation of no active orders also lacks the
record's object-type and identity proof. The repair must add a different audit
fact, never reinterpret either input.

## Decision

Introduce a dedicated reconciliation operation that only reads official
conditional-order pages and appends one local record event when it can prove a
single record-owned conditional is absent. It does not call an order mutation.

The operation must:

1. select only an outstanding conditional artifact from the chosen verification
   record and its resolved profile/account context;
2. read every bounded page for both official `OPEN` and `CLOSED` groups using
   the recorded symbol and exact opaque identifier comparison;
3. reject repeated/missing cursors, incomplete pages, wrong profile/account,
   unsupported object type, stale snapshot, ambiguous identity, or any read
   failure without writing a terminal event;
4. append `reconciled-absent` only when the exact identifier is absent from the
   complete fresh authoritative result; and
5. make the event idempotent, retain a digest/basis for the read without raw
   account or broker identifiers, and leave the original failed cleanup entry
   intact.

`Outstanding` may stop presenting only an artifact with that explicit event.
The projection must label it reconciled absent, never cancelled or filled.
`PendingCleanup` and resume planning must not schedule a cancellation for that
artifact after reconciliation.

## Safety and attestation boundary

The operation is an account read plus local audit append. It must never be
counted in successful endpoint, verified capability, soak-attestation, or
engine-start evidence. It cannot change a failed `cleanup`, `conditional-*`,
or other measurement verdict. A DELETE 404 without the successful official read
remains outstanding.

The CLI must make the record/profile target explicit, redact operator output,
and refuse broad account scans or arbitrary identifiers. It must not offer an
order-mutation flag, hidden approval bypass, retry of cleanup, `--redo`, or a
new record path.

## Compatibility and rollback

Existing records without a reconciliation event retain their current behavior.
The event is append-only and ignored by older binaries as an unknown record
kind only if their parser already preserves forward-compatible entries; if that
is not true, schema/version handling must fail closed before write. Rollback
means keeping the event and using the prior binary only when it safely reads the
version; it never removes or rewrites evidence.

## Verification approach

Create RED tests for every fail-closed boundary before implementation. Use fake
official readers and isolated records only. Tests must prove no broker mutation
method is reachable, preserve the failed DELETE event, and prove the new event
does not influence endpoint or attestation success.
