# KR/US Production Schedule Authority Loader

**Author:** Codex
**Date:** 2026-08-04
**Status:** Approved by the frozen `a072-wire-multi-market-strategy-runtime` paired-delivery contract
**Reviewers:** TossOS owner, independent high-risk reviewer before release
**Change:** `a072-wire-multi-market-strategy-runtime`

## Context

The scheduler now has an exact signed-manifest verifier, but the engine does not yet load per-market
desired state, fetch official market calendars, derive the current build/config binding, or invoke the
paired restore. `NewPairedStrategyEntryProductionAssembly` therefore still receives only scalar
automation observations and correctly constructs two dormant workers.

This slice adds the engine-owned read-only authority loader that closes the schedule portion of that
gap for KR and US together. It does not create worker cycles and cannot submit an order. Later production
assembly consumes the loader's private authority only after candidate, risk, protection, and dispatch
dependencies are all present.

## Functional Requirements

- FR-1: The engine MUST load exactly `scheduler-desired-KR.json` and `scheduler-desired-US.json` from the
  resolved profile config directory at one frozen engine-clock observation.
- FR-2: The engine MUST derive the current strategy config digest from the compiled router ID/release and
  all six compiled KR/US lane descriptors; a desired file MUST NOT self-assert the current config.
- FR-3: For each enabled/autostart market, the engine MUST read that market's immutable activation digest,
  key ID, and canonical Ed25519 public key from three market-specific deployment environment pins.
- FR-4: The engine MUST fetch KR and US typed official calendars concurrently using each exchange's IANA
  local date, adapt them with `scheduler.AdaptOfficialCalendar`, and bind their exact versions into the
  current scheduler binding.
- FR-5: The engine MUST invoke `scheduler.RestorePairedProduction` once with one KR and one US request and
  the same frozen clock used to load desired state and adapt calendars.
- FR-6: A missing, malformed, stale, or refused market-local desired file, pin, calendar, or activation
  MUST close only that market schedule authority. It MUST NOT cancel or replace an exact peer result.
- FR-7: When a market desired state is OFF or autostart is false, the loader MUST NOT fetch its calendar
  or require deployment pins, and it MUST return a dormant market result.
- FR-8: The loader MUST expose only a read-only snapshot publicly. Opaque activation and calendar values
  MUST remain private to `internal/app/engine` until the complete worker is constructed.
- FR-9: The loader MUST NOT write desired state, manifests, config, journal, operating toggles, activation,
  broker state, or orders and MUST NOT turn either worker effective.

## Non-Functional Requirements

- NFR-1 Isolation: KR and US desired/calendar/pin collection MUST execute in independent goroutines; a
  stalled market is bounded by the caller context and does not prevent the peer operation from starting.
- NFR-2 Security: Base64 public keys MUST use strict canonical standard encoding and decode to exactly 32
  bytes. Digest and key ID validation MUST remain in the scheduler verifier.
- NFR-3 Determinism: The compiled config digest MUST be stable across process restarts for identical router
  and lane descriptors and MUST change if any identity/release field changes.
- NFR-4 Compatibility: With no new files or environment pins, production behavior MUST remain two paired
  dormant/OFF markets and make zero strategy calendar calls.
- NFR-5 Mutation safety: This slice MUST perform zero broker requests other than the two official calendar
  GETs required for explicitly enabled/autostart markets.

## Acceptance Criteria

### AC-1: Paired exact schedule authority (FR-1, FR-2, FR-3, FR-4, FR-5)

Given exact KR and US desired files, deployment pins, signed activations, and official calendars, when the
engine loader runs, then both private authorities restore against one frozen clock and the public snapshot
reports two independently ready markets with exact calendar and manifest digests.

### AC-2: Concurrent calendar reads (FR-4, FR-6, NFR-1)

Given a blocked KR calendar endpoint and a valid US endpoint, when the loader runs, then both requests start
without serial waiting and US collection completes independently before KR is released.

### AC-3: Market-local refusal (FR-6)

Given KR has a bad activation signature or calendar and US is exact, when the loader runs, then KR is not
ready and US remains ready; the inverse scenario yields the symmetric result.

### AC-4: Dormant zero-read baseline (FR-7, FR-9, NFR-4, NFR-5)

Given absent desired files or explicit OFF/autostart-false desired state, when the production engine is
assembled, then both markets remain dormant and no strategy calendar, activation, journal, or broker
mutation call occurs.

### AC-5: Compiled config drift (FR-2, NFR-3)

Given a signed manifest bound to a different compiled config digest, when restore runs, then only that
market returns a manifest mismatch and no opaque activation escapes.

### AC-6: Public capability boundary (FR-8, FR-9)

Given an external caller receives the loader snapshot, when its exported fields and methods are inspected,
then it can observe state/reason/digests only and cannot access Activation, Calendar, Gateway, Journal,
Guardian, trigger, cycle, writer, toggle, broker, or order capabilities.

## Edge Cases

- EC-1: The profile directory is absent or a desired file is malformed — market-local unavailable result.
- EC-2: One or more of a market's three deployment pins is missing or malformed — that market remains
  unavailable; exact peer pins are still processed.
- EC-3: The official endpoint returns a holiday, early close, malformed timezone/date, stale fetch, or
  cross-market payload — adapter/decision refuses without peer contamination.
- EC-4: Engine clock is zero or context is cancelled — no activation is returned.
- EC-5: Calendar version differs from desired/manifest after fetch — restore mismatch.
- EC-6: Environment contains China/combined-market pins — ignored; only the six exact KR/US names are read.

## API Contracts

N/A — this slice adds no `GET /...` server route. It invokes existing official calendar GETs through the
sealed engine client and exposes one in-process read-only snapshot.

```ts
interface StrategyScheduleMarketSnapshot {
  market: "KR" | "US";
  desiredEnabled: boolean;
  desiredAutostart: boolean;
  ready: boolean;
  reason: ResumeReason;
  calendarVersion: string;
  activationManifestDigest: string;
  configDigest: string;
}

interface PairedStrategyScheduleSnapshot {
  observedAt: RFC3339NanoUTC;
  KR: StrategyScheduleMarketSnapshot;
  US: StrategyScheduleMarketSnapshot;
}
```

## Data Models

| Entity | Field | Type | Constraints |
|---|---|---|---|
| desired file | path | string | exact market filename under resolved config dir |
| deployment pins | digest/key ID/public key | strings | market-specific, all-or-nothing, read-only |
| current binding | build/config/calendar | strings | engine-derived; never desired-derived config/build |
| private authority | activation/calendar/desired | opaque values | engine package only |
| public snapshot | ready/reason/digests | scalars | no authority or mutation method |

No database or existing config schema changes are introduced.

## Out of Scope

- OS-1: Writing desired files, deployment environment, or signed activation manifests.
- OS-2: Candidate/evidence loading, lane evaluation, risk buckets, first-leg issuance, dispatch, or orders.
- OS-3: Making either production worker effective before all remaining paired dependencies pass.
- OS-4: China, a combined KR+US authority, or KR-stable-before-US scheduling.
