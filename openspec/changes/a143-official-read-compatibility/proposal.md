## Why

General official charts must satisfy consumers' chronological-order contract while strict engine candle ingestion retains its distinct newest-first, close-time cursor contract. Published REST/AsyncAPI specifications and domestic alphanumeric read symbols also need a current compatibility baseline.

## What Changes

- Adapt general official charts into oldest-to-newest order without changing timestamps.
- Accept official domestic six-character uppercase alphanumeric symbols in the strict read lexical check, retaining identity and all data-validity checks.
- Refresh public REST snapshot to the verified published version and add an AsyncAPI snapshot with source/hash evidence.
- Add offline regression coverage for general charts and strict-read invariants.

## Capabilities

### New Capabilities

- `official-read-compatibility`: General chart order, domestic read-symbol compatibility and pinned public API evidence.

### Modified Capabilities

None.

## Impact

STORY-TOS-a143 / FEAT-TOS-014. Expected boundary: internal/official/candle_reads.go, strict_minute_candles.go (lexical predicate only), corresponding tests and docs/migration API snapshots/metadata. No workflow/schedule/operating setting changes. No new API clients, websocket connection, order types, mutation/expiry validation, strategy enablement, symbol-universe changes or production API calls. Existing strict raw candles, cursor semantics, officialbars conversion, journal, risk controls and shared tokens remain invariant. Intake AST: /tmp/tossos-a143-intake; revalidate hashes against worktree before implementation. Depends on reviewed a142 landing.
