# Signed protection attestation threat model

Date: 2026-08-01

## Scope and trust boundaries

- Assets: protection capability claims, runtime scope, evidence digests, verifier tool builds,
  Ed25519 public keys, key lifecycle metadata, and the fail-closed `UNWIRED/OFF` decision.
- External authority: an offline/operator-controlled signer and pre-provisioned public trust root.
- Runtime process: read-only verifier. It has no private key, trust-root writer, TOFU, key download,
  or fallback authority.
- Filesystem boundary: attestation and trust root use separate directories. Immutable startup
  policy pins the trust root's absolute path, owner UID and canonical SHA-256 independently from
  the runtime-owned attestation.
- Authorization boundary: signature success is necessary but insufficient; time, scope and evidence
  checks remain mandatory.

## Prioritized STRIDE findings

| ID | STRIDE | Threat / attack vector | DREAD | Mitigation owner and verification |
|---|---|---|---:|---|
| S1 | Spoofing | Same-UID attacker writes a digest-consistent unsigned matrix. | 8.2 | Security: Ed25519 signature rooted in a pre-provisioned read-only keyset; forged-signature RED test. |
| T1 | Tampering | Swap payload, account/profile/market/tool build/evidence descriptor, signature or key ID. | 8.0 | Security: domain-separated, length-delimited signing input binds header and canonical payload; swap/cross-scope RED tests. |
| T2 | Tampering | Alternate JSON/base64 encoding, duplicate/unknown fields or parser ambiguity. | 7.8 | Security: strict decoder plus byte-for-byte canonical JSON and raw base64url round-trip; malformed encoding RED tests. |
| E1 | Elevation | Unknown role/algorithm/key is treated as an attestation signer. | 8.0 | Security: exact role/algorithm, unique key IDs, 32-byte keys, no fallback; unknown/duplicate/key-size RED tests. |
| S2 | Spoofing | Replay a signed capability after scope, evidence, matrix or key validity changes. | 7.8 | Security: matrix issue/expiry, exact runtime scope/tool build/evidence binding, and full key-window containment/current-time checks. |
| T3 | Tampering | Restore a compromised/revoked key or exploit ambiguous revocation timing. | 8.4 | Operations + Security: explicit hard revocation invalidates every signature from the key; revoked-key RED test. |
| D1 | DoS | Oversized files or keysets exhaust parser resources. | 6.0 | Security: 1 MiB bounded reads and bounded key count; oversize RED test. |
| T4 | Tampering | Symlink/hardlink/rename/permission/owner race replaces trust material during read. | 8.0 | Security: exact basename, lstat/open/fstat/post-read lstat identity, owner/mode/link-count checks; filesystem RED tests. |
| T5 | Tampering | Attacker rolls the canonical policy and trust root back together, reuses one generation with another digest, or revokes during evidence hashing. | 8.6 | Security: evidence/scope first, then final sealed-policy/root/key/time reload; policy/root generation equality and per-verifier monotonic generation/digest latch; revoke-after-evidence plus rollback RED tests. |
| T6 | Tampering | Direct parent is replaced or made writable after its pre-read check. | 8.0 | Security: parent identity, owner, type and mode are re-read and revalidated after bounded artifact read; deterministic parent-race tests. |
| E2 | Elevation | A valid signature over extra rows leaks broader authority to a downstream caller. | 8.1 | Security: final value contains only a deep-copied exact scope and the single matched capability row; extra-row authority test. |
| R1 | Repudiation | Caller submits a full replacement saga, mismatched attempt/broker result, or a different event that happens to produce overlapping row fields and later claims an idempotent retry. | 8.0 | Domain + Storage: event-specific methods, stored-row transition, persisted event kind + canonical fingerprint, exact result predicate and atomic revision CAS over independent connections. Legacy blank lineage is allowed only for revision-1 PLANNED. |
| T7 | Tampering | Replay an ALLOWED flatten decision later or against another scope/quantity. | 9.0 | Domain: verifier-clock decision time and opaque atomic one-shot permit bound to scope, broker, quantity and deadline; copy/+1h/mismatch tests. |

The generic repository-wide secret scanner produces many Twilio-key regex false positives from
lockfile integrity strings, AST hashes and long test identifiers. A focused scan of every changed Go
source and test file reports zero high/critical findings after ambiguous test identifiers were split
with underscores. No production or test private signing key is stored: tests generate ephemeral keys
in `_test.go` only.

## Rotation and revocation contract

- Rotation is an explicit overlapping validity window with distinct ACTIVE key IDs.
- A signed payload must be issued and expire wholly inside the selected key window, and the selected
  key must also be valid at verification time.
- REVOKED is a hard revocation: the key is never accepted, including attestations issued before
  `revoked_at`. The timestamp and non-empty reason are audit metadata, not a grace period.
- Removing an old key is allowed only after every attestation signed by it has expired; otherwise
  verification fails unknown-key and keeps protection `UNWIRED/OFF`.

## Residual boundary

Filesystem ownership checks do not defend against the pinned provisioning owner by themselves.
The immutable startup policy therefore also pins the canonical root SHA-256; replacing the keyset
under the same UID cannot introduce attacker keys. The trust directory must be non-symlink,
traversable by the non-owner runtime, and group/other non-writable, while the `0444` file is regular
and single-link.
Root-owned `/etc` layouts and distinct-UID rootless mounts are both valid. This code intentionally
supplies no trust-root mutation path and no production key, so absence or provisioning failure is
safe unavailability rather than implicit trust.

The canonical policy itself is also a read-only, exact-basename artifact with a separately pinned
source identity inside the sealed verifier. Every parse and final authorization reloads it, and the
verifier keeps a monotonic in-memory generation/digest observation. This prevents a complete
policy+root rollback during the lifetime of a verifier. Persistence across process restarts remains
an operational provisioning responsibility; because no production verifier constructor or trust
provisioning exists in this dormant change, restart cannot silently enable authority.
