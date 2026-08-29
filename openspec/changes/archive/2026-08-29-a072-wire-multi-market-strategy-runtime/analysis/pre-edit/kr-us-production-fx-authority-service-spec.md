# Spec: KR/US production FX authority service

**Author:** Codex backend teammate
**Date:** 2026-08-04
**Status:** Approved
**Reviewers:** a072 manager contract and paired-delivery mandate
**Related specs:** a066 multi-horizon risk buckets; a072 strategy runtime account-base wave

## Context

`internal/officialfx` already owns sealed exact-decimal FX evidence, but its `HaircutPolicy` and
same-currency `IdentitySnapshot` constructors are intentionally private and no production loader can
mint them. Consequently KR identity and US USD-to-account-base authority cannot yet be collected by
the production runtime. Inventing a haircut constant or accepting caller-provided rate, digest or
freshness would bypass the a066/a072 monetary-authority boundary.

This slice adds one package-owned service which attempts KR and US in the same collection wave. KR
re-verifies the selected account through the immutable official production origin before minting a
bounded identity snapshot. US accepts a haircut only from a pinned, Ed25519-signed, canonical
`fx-risk-policy-manifest.json`, then uses the existing `ReadOfficial` boundary. Failures remain
market-local: either market may produce valid evidence while its peer is refused.

## Functional Requirements

- FR-1: `ProductionAuthorityService.Collect` MUST attempt KR account identity and US official FX
  in one invocation using one frozen UTC clock observation.
- FR-2: KR evidence MUST be minted only after a production-origin official account-list read
  proves exactly one selected account matching the configured account identity.
- FR-3: The KR identity snapshot MUST use package-owned fixed rate/haircut `1`, MUST bind the
  exact account currency, and MUST use a package-owned freshness window no greater than five minutes.
- FR-4: US policy MUST come only from the exact pinned bytes of
  `fx-risk-policy-manifest.json`; absence, unsafe metadata, digest mismatch, non-canonical JSON,
  unknown/duplicate fields or trailing data MUST refuse US.
- FR-5: The manifest MUST bind schema, exact account/base/quote/market, positive monotonic
  generation, policy ID/version, multiplier `>= 1`, canonical observed/fresh-until UTC times,
  approver, pinned key ID, exact Ed25519 algorithm and signature.
- FR-6: The service MUST verify the manifest signature over the canonical body before minting the
  private `HaircutPolicy`, and MUST NOT synthesize or default a multiplier.
- FR-7: US evidence MUST be minted through the existing `ReadOfficial` production-origin read
  with the manifest-minted policy.
- FR-8: Durable state MUST reject wall-clock rollback, generation rollback and same-generation
  manifest substitution. Exact same-generation/same-digest replay MAY be collected again while
  current.
- FR-9: KR refusal MUST NOT suppress an otherwise valid US result, and US refusal MUST NOT
  suppress an otherwise valid KR result.
- FR-10: Public configuration and result APIs MUST NOT expose raw rate, haircut, evidence digest,
  observed-at or fresh-until as caller-controlled mint fields. Results expose only opaque
  `officialfx.Evidence` per market.
- FR-11: Production market workers MUST be able to collect KR or US authority independently.
  A slow peer official read MUST NOT be started or awaited by the requesting market worker.

## Non-Functional Requirements

- NFR-1: Manifest and durable-state files MUST be regular, non-symlink files under an absolute,
  owner-only `0700` directory, with manifest mode `0400` and state/lock mode `0600` on supported
  Unix production builds.
- NFR-2: Manifest input MUST be bounded to 1 MiB and decoded with duplicate-key, unknown-field,
  trailing-value and byte-for-byte canonical checks.
- NFR-3: Concurrent collectors and restarted service instances MUST serialize trusted-time and
  generation state with a process-shared file lock; race tests MUST pass.
- NFR-4: This service MUST own no broker, journal, activation, operating-toggle or LIVE-order
  capability and MUST perform no network I/O in tests.

## Acceptance Criteria

### AC-1: Paired successful collection (FR-1, FR-2, FR-3, FR-5, FR-7)

Given an exact official account identity and a current signed USD-to-KRW policy manifest
When `Collect` runs at the frozen clock
Then KR yields current KRW identity evidence and US yields current USD-to-KRW official evidence
And both official read paths were attempted exactly once.

### AC-2: US policy failure is isolated (FR-4, FR-6, FR-9)

Given a current official KR account and an absent, unsigned, stale or wrong-scope US manifest
When `Collect` runs
Then KR yields valid identity evidence
And US yields a typed refusal without an official FX read.

### AC-3: KR official identity failure is isolated (FR-2, FR-9)

Given a failing or mismatched official account read and a valid US manifest/rate
When `Collect` runs
Then KR yields a typed refusal
And US yields valid official evidence.

### AC-4: No numeric default (FR-5, FR-6)

Given a signed manifest whose multiplier is absent, zero, below one or non-canonical
When `Collect` runs
Then US is refused before the official FX read
And no package fallback multiplier is used.

### AC-5: Durable rollback resistance (FR-8, NFR-3)

Given a successfully collected generation and trusted-time floor
When a restarted service observes a lower generation, substituted same-generation digest or earlier
clock
Then US is refused before the official FX read
And KR official identity collection remains independently available.

### AC-6: Concurrent exact replay (FR-8, NFR-3)

Given one valid pinned manifest
When concurrent service instances collect the same generation and digest
Then every successful US result binds that exact policy and durable state remains canonical.

### AC-7: Caller cannot mint authority (FR-10)

Given code outside `internal/officialfx`
When it constructs public production config or reads a collection result
Then it cannot provide raw rate/haircut/digest/freshness inputs or construct a non-zero evidence
capability.

### AC-8: Peer latency is isolated (FR-11)

Given the US official FX read is blocked
When the KR worker calls `CollectKR`
Then KR identity authority completes without waiting for US; the inverse holds for `CollectUS`
while the KR account-identity read is blocked.

## Edge Cases and Error Scenarios

- EC-1: Nil context/client/service or zero/non-UTC clock -> affected market typed unavailable.
- EC-2: Multiple or missing matching accounts, selected-account drift, configured HTTP origin -> KR refusal.
- EC-3: Manifest symlink, wrong owner/mode, oversized/empty file -> US refusal.
- EC-4: Duplicate/unknown fields, non-canonical time/decimal/base64 or bad signature -> US refusal.
- EC-5: Quote equals base or market is not exact `US` -> US scope refusal; no identity inference.
- EC-6: Manifest/rate validity windows do not intersect -> US refusal from the existing sealing path.
- EC-7: State deletion after bootstrap or corrupt state/lock marker -> US state-corrupt refusal.
- EC-8: Context cancellation in one official read does not fabricate peer evidence.

## API Contracts

```go
type ProductionAuthorityConfig struct {
    ConfigDir       string
    AccountID       string
    AccountCurrency string
    ManifestDigest string            // immutable trust pin, not an evidence mint field
    TrustedKeyID   string
    TrustedKey     ed25519.PublicKey // immutable trust root
    Now            func() time.Time
}

func NewProductionAuthorityService(
    ProductionAuthorityConfig,
    *official.Client,
) *ProductionAuthorityService

func (*ProductionAuthorityService) Collect(context.Context) Collection
func (*ProductionAuthorityService) CollectKR(context.Context) (Evidence, error)
func (*ProductionAuthorityService) CollectUS(context.Context) (Evidence, error)
func (Collection) KR() (Evidence, error)
func (Collection) US() (Evidence, error)
```

N/A -- this is an internal Go capability and adds no HTTP method or route; in particular it MUST NOT
expose `POST /fx-authority`. Equivalent language-neutral
result shape:

```typescript
interface PairedFXCollection {
  KR: OpaqueEvidence | TypedRefusal;
  US: OpaqueEvidence | TypedRefusal;
}
```

No public constructor accepts a rate, haircut, evidence digest, evidence observation or freshness
time. `Collection` returns evidence by value and does not expose policy material.

## Data Models

| Entity | Field | Type | Constraints |
| --- | --- | --- | --- |
| Manifest body | schema_version | string | exact `fx-risk-policy/v1` |
| Manifest body | account_id | string | exact configured account |
| Manifest body | account_currency | string | exact canonical configured base |
| Manifest body | market | string | exact `US` |
| Manifest body | quote_currency | string | canonical and different from base |
| Manifest body | generation | uint64 | positive, durable monotonic |
| Manifest body | policy_id/version | string | bounded canonical identities |
| Manifest body | multiplier | decimal string | canonical positive and `>= 1`; mandatory |
| Manifest body | observed_at/fresh_until | RFC3339Nano UTC | observed <= now <= fresh, non-empty |
| Manifest body | approver | string | bounded, non-empty |
| Manifest body | key_id | string | exact pinned key |
| Manifest body | signature_algorithm | string | exact `Ed25519` |
| Manifest envelope | signature | base64 string | 64-byte Ed25519 signature of canonical body |
| Durable state | schema_version | string | exact `fx-risk-policy-state/v1` |
| Durable state | trusted_time_floor | RFC3339Nano UTC | monotonic |
| Durable state | generation | uint64 | monotonic |
| Durable state | manifest_digest | string | exact digest for generation |

## Out of Scope

- OS-1: Broker calls, journal writes, dispatch leases, activation/approval creation and operating toggles.
- OS-2: Caller-supplied/test-only public policy constructors or default haircut values.
- OS-3: China-market evidence and non-USD US quote currencies in this a072 production profile.
- OS-4: Key rotation/revocation infrastructure beyond one explicitly pinned Ed25519 trust root.
- OS-5: Reusing protection-readiness activation attestations as FX-risk policy authority.
