# KR/US Production Scheduler Activation Verifier

**Author:** Codex
**Date:** 2026-08-04
**Status:** Approved by the frozen `a072-wire-multi-market-strategy-runtime` paired-delivery contract
**Reviewers:** TossOS owner, independent high-risk reviewer before release
**Change:** `a072-wire-multi-market-strategy-runtime`
**Implementation scope:** new code in `internal/scheduler`; existing scheduler state and `Restore` logic remain unchanged

## Context

`scheduler.Restore` already refuses autostart unless an opaque `Activation` is minted by an
implementation of the package-private `ActivationVerifier` contract. Production currently supplies no
verifier, so both KR and US necessarily return `MANIFEST_UNAVAILABLE`. This is the correct dormant
baseline but leaves the approved production runtime graph without a read-only path for consuming an
existing human-approved activation.

The runtime requires independent KR and US activation authority. A combined bool, one shared manifest,
or a caller-populated activation DTO would couple failure domains and could synthesize permission. The
production verifier therefore consumes two separately pinned, owner-only, canonical, signed files and
returns two independent `RestoreResult` values from one paired operation. It never writes desired state,
activation, settings, journal, or broker state.

## Functional Requirements

- FR-1: The implementation MUST provide one production verifier per market whose constructor accepts
  only an absolute config directory, exact market, immutable file digest, and trusted Ed25519 key
  ID/public key. The paired restore accepts the clock and freezes exactly one observation for both
  markets.
- FR-2: The verifier MUST read only the canonical market file
  `strategy-activation-KR.json` or `strategy-activation-US.json`, owned by the current process owner and
  mode `0400`, with a bounded file size.
- FR-3: The verifier MUST reject unknown or duplicate JSON fields, trailing JSON values, non-canonical
  RFC3339Nano UTC timestamps, non-canonical base64 signatures, and any mismatch in schema, domain,
  algorithm, key ID, file digest, market, scheduler version, desired revision, calendar version, session,
  config version, build digest, actor, or approval time.
- FR-4: The signature MUST cover every authority field plus `issued_at`, `expires_at`, `generation`,
  and `revoked`; verification MUST use canonical JSON bytes produced from the signature-free body.
- FR-5: A manifest MUST be rejected when revoked, not yet issued, expired, issued before its bound
  human approval, valid for longer than 24 hours, or when its generation is zero.
- FR-6: Successful verification MUST be usable only through the existing package-private
  `ActivationVerifier` interface and existing `Restore`; no public constructor for `Activation` MAY be
  added.
- FR-7: A paired production restore MUST start KR and US verification concurrently from one frozen
  time observation and MUST return independent market-keyed results. KR failure MUST NOT cancel, replace,
  or weaken US verification, and the inverse MUST also hold.
- FR-8: The paired restore MUST require exactly one KR request and one US request. Duplicate, missing,
  combined, or cross-market requests MUST return paired fail-closed results without verifying either
  manifest.
- FR-9: No function in this implementation MAY create or modify scheduler desired state, activation
  files, operating toggles, journal rows, orders, broker state, or LIVE approval.

## Non-Functional Requirements

- **NFR-1 Security:** File ownership, mode, digest, signature, exact binding, time, and revocation checks
  MUST all pass before `Restore` can receive successful verifier authority.
- **NFR-2 Isolation:** Each market verification MUST have its own goroutine, context observation, file,
  digest pin, and verifier. Peer failure MUST NOT be propagated as that market's error.
- **NFR-3 Determinism:** The same file bytes, trust pin, binding, and frozen time MUST produce the same
  classification.
- **NFR-4 Compatibility:** Existing `DesiredState`, `Activation`, `Restore`, and default dormant behavior
  MUST remain byte/API compatible.
- **NFR-5 Boundedness:** Each manifest MUST be at most 64 KiB and validity MUST be at most 24 hours.

## Acceptance Criteria

### AC-1: Paired exact activation (FR-1, FR-2, FR-3, FR-4, FR-5, FR-6, NFR-1)

Given valid, separately signed KR and US manifests with exact pins and exact current bindings,
when paired restore runs, then both results are `Restored=true` with
  `EXACT_MANIFEST` and non-nil opaque activations.

### AC-2: Concurrent independent verification (FR-7, NFR-2)

Given KR verification blocks while US is valid, when paired restore runs, then
  US verification starts without waiting for KR; releasing KR later yields two independent results.

### AC-3: Market-local refusal (FR-7, FR-8)

Given one market has a bad signature, wrong digest, stale time, revocation, or
  cross-market body, when paired restore runs, then only that market refuses and the exact valid peer may
  restore.

### AC-4: Tamper and time refusal (FR-3, FR-4, FR-5)

Given any signed authority field is changed after signing, when verification
  runs, then `Restore` returns `MANIFEST_MISMATCH`, `MANIFEST_EXPIRED`, or `MANIFEST_REVOKED` and no
  activation is returned.

### AC-5: Exact pair requirement (FR-8)

Given duplicate KR, duplicate US, a missing peer, or a combined market request, when
  paired restore is constructed, then both results fail closed and neither verifier is called.

### AC-6: Dormant compatibility (FR-9, NFR-4)

Given the production config is absent or both desired states are OFF, when the
  runtime is assembled, then the existing two dormant workers remain OFF and no file or mutation is
  produced.

## Edge Cases

- EC-1: Current-owner identity is unavailable, directory is relative, file is symlinked, file owner
  differs, mode is not exactly `0400`, or file exceeds 64 KiB — refuse unavailable/mismatch.
- EC-2: The clock is zero, moves before approval/issuance, equals expiry, or observes a manifest valid
  for more than 24 hours — refuse.
- EC-3: Context is already cancelled or is cancelled after verification — return verification failure;
  never return activation from a cancelled request.
- EC-4: JSON contains duplicate fields, unknown fields, trailing values, non-UTC/non-canonical time,
  invalid base64, wrong signature length, or an unsupported algorithm — refuse mismatch.
- EC-5: Generation is zero or revoked is true — refuse mismatch or revoked respectively.
- EC-6: One market panics or stalls — the peer attempt remains independent; outer worker watchdog owns
  the stall bound and panic classification.

## API Contracts

N/A — this change adds no `GET /...`, mutation route, or other HTTP endpoint. The contracts below are
in-process, read-only Go boundaries expressed in TypeScript-style notation for review.

```ts
type Market = "KR" | "US";

interface ProductionActivationConfig {
  configDir: string;          // absolute; read-only
  market: Market;
  manifestDigest: `sha256:${string}`;
  trustedKeyId: string;
  trustedKey: Ed25519PublicKey;
}

interface ActivationManifestBody {
  schemaVersion: "strategy-activation-manifest:v1";
  domain: "TossOS/strategy-scheduler-activation/ed25519/v1";
  signatureAlgorithm: "Ed25519";
  keyId: string;
  generation: number;
  schedulerVersion: "scheduler-v1";
  desiredRevision: number;
  calendarVersion: string;
  market: Market;
  session: "regular";
  configVersion: string;
  buildDigest: string;
  actor: string;
  approvedAt: RFC3339NanoUTC;
  issuedAt: RFC3339NanoUTC;
  expiresAt: RFC3339NanoUTC;
  revoked: boolean;
}

interface ActivationManifest extends ActivationManifestBody {
  signature: CanonicalBase64;
}

interface PairedRestoreRequest {
  market: Market;
  desired: DesiredState;
  current: CurrentBinding;
  verifier: ProductionActivationVerifier;
}

type PairedRestoreClock = () => Date; // called exactly once per paired restore

interface PairedRestoreResult {
  KR: RestoreResult;
  US: RestoreResult;
}
```

Errors are surfaced through existing `RestoreResult.Reason` and `Err` values. No new network endpoint or
mutation command is introduced.

## Data Models

| Entity | Field | Type | Constraints |
|---|---|---|---|
| production config | config dir | path | absolute, existing owner boundary |
| production config | market | enum | exactly KR or US |
| production config | manifest digest | string | `sha256:` + 64 lowercase hex |
| production config | key | Ed25519 public key | exactly 32 bytes, bounded key ID |
| manifest | generation | uint64 | greater than zero |
| manifest | binding | ActivationBinding | exact equality on every field |
| manifest | issued/expiry | timestamp | canonical UTC, approval ≤ issue ≤ now < expiry, duration ≤24h |
| manifest | revoked | bool | must be false |
| manifest | signature | base64 | canonical standard base64, 64-byte Ed25519 signature |

No database schema is added. Files are inputs owned by an external human approval process; TossOS only
reads them.

## Out of Scope

- OS-1: Writing, signing, approving, rotating, or revoking activation manifests — human/operations authority.
- OS-2: Enabling scheduler desired state, autostart, automation gate, lane state, or LIVE trading.
- OS-3: Candidate/evidence collection, lane evaluation, risk-bucket snapshot loading, first-leg issuance,
  dispatch, Gateway transport, and protection lifecycle; these are later paired slices.
- OS-4: A combined KR+US activation manifest or one market serving as prerequisite/fallback for the other.
