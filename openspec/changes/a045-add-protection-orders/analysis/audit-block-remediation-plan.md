# Independent audit BLOCK remediation map

Date: 2026-08-01
Base: `1b6da8d7ee2455fe2f937873ba69c41c7355b300`

## Trust and authorization boundary (H1, M1-M3)

| Function boundary | Current branch/risk | Required fail-closed change | RED evidence |
|---|---|---|---|
| exported parse/verify/load | caller supplies trust path, owner, digest and `now`; parsed key snapshot can outlive hard revocation | configured verifier is sealed; production has no injectable constructor; it owns a canonical policy source and clock; final authorization reloads policy, root, digest, generation and signer | caller-injected inputs no longer compile; parse-before-revoke fails |
| signed parse | matrix and a copied trust key are returned | retain exact signed bytes/key ID plus policy generation only; do not authorize | hard revoke between parse and verify |
| final verify | trusts copied key and returns whole matrix | finish caller evidence/scope work first; then reload current root, reject generation rollback/reuse, resample time, reverify signature/key window, and return only the exact matched scope and row | revoke-after-evidence, revoked/unknown/replaced root and extra-row authority tests |
| artifact read | direct parent checked only before file read | capture and revalidate parent identity, type, mode and owner after read; unsupported owner/link metadata fails closed | post-read parent replace/mode tests and Windows compile |
| matrix canonical form | valid arrays can be reordered and accounts can use multiple hyphen groupings | exact UTC timestamps, strictly sorted evidence/capability rows, duplicate refusal and digits-only 8-14 account grammar | reordered arrays, offset timestamps and hyphen aliases fail |

## Durable saga boundary (H2)

| Function boundary | Current branch/risk | Required fail-closed change | RED evidence |
|---|---|---|---|
| `Repository.Insert` | any valid state can be inserted | only a revision-1 `PLANNED` saga can establish identity | ACTIVE/REGISTERING insert rejection |
| `Repository.Update` | caller submits a complete mutable saga | remove arbitrary-state update; load stored row and apply one typed `Event` through `Transition` | forged attempt/broker lineage cases no longer expressible/accepted |
| CAS | read transaction then update; one-connection tests do not exercise contention | atomic `WHERE revision=?` CAS over two real SQLite connections; persist last event kind + canonical fingerprint and on stale/zero rows reload only for that exact event result | simultaneous conflicting events: exactly one success; same event retry: one revision; registration/replace result collision rejects |
| schema migration | v1 rows have no event identity and previously allowed non-PLANNED insert | add event identity columns, validate every row, and accept blank lineage only for exact revision-1 PLANNED | ACTIVE/REGISTERING rev1 and PLANNED rev2 legacy rows reject |
| `Transition` | response events do not bind attempt/broker identity | active/unknown/replace responses bind stored attempt ID; trigger/close bind stored broker ID; irrelevant identity fields are rejected | REGISTERING(a1)→ACTIVE(a-forged,b-forged), ACTIVE(b1)→TRIGGERED(b2) |

## Flatten boundary (H3)

| Branch | Required invariant | RED evidence |
|---|---|---|
| decision clock | `start <= observations <= decisionAt <= start+2s`; decisionAt after deadline or before an observation is IN_DOUBT | late decision, rollback and stale/pre-start observation |
| cancel-first | exact scope/broker; terminal non-triggered cancel precedes sellable observation | existing response-loss/race/order tables |
| authorization | ALLOWED returns an opaque permit bound to exact scope, quantity, issue/deadline and shared atomic one-shot state | copied permit can be consumed once; +1h and wrong scope/quantity reject |

## Safety boundary

No production signer, private key, trust writer, TOFU, API import, broker mutation, `WIRED`, LIVE toggle,
or task checkbox is introduced. Missing canonical policy provisioning remains `UNWIRED/OFF`.
