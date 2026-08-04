# Production strategy-runtime projection wiring — pre-edit map

## Function Logic Map

### `engine.NewContext`

1. Assemble the official-only account, journal, Gateway, interlock, and safety loops.
2. Own a read-only in-memory projection initialized to the exact paired KR/US dormant snapshot.
3. The projection exposes no order, activation, toggle, or journal mutation method.

### `Context.NewPairedStrategyEntryProductionAssembly`

1. Freeze one observation time and collect the paired KR/US authority graph.
2. Construct independent market workers and the shared supervisor.
3. Project only verified observations; missing facts remain typed UNKNOWN/OFF.
4. Atomically replace the engine-owned read projection after validation.

### `runEngineRun`

1. Acquire the engine exclusion and assemble the verified runtime.
2. Start an authenticated read-only Unix projection server owned by the engine process.
3. Run and drain the safety runtime; close the projection server during shutdown.

### `runHTTPAPI` / `runConsole`

1. Resolve the engine directory without starting or mutating the engine.
2. If the authenticated projection descriptor exists, dial it read-only.
3. Inject the same client into REST, console, and the full SSE snapshot.
4. Missing/unreadable runtime remains an honest dormant/unavailable presentation.

## Branch Test Map

| Branch | Expected result | Test |
|---|---|---|
| Engine has no configured KR/US authority | Valid paired dormant snapshot | store/projection unit test |
| KR ready, US unavailable | KR facts preserved; US typed UNKNOWN/OFF | projection builder test |
| Engine Unix endpoint exists | Console, REST, and SSE read the same snapshot | command construction/integration test |
| Descriptor absent or invalid | No mutation or engine start; surface remains dormant | command branch test |
| Projection replacement is concurrent | Readers receive immutable validated old/new snapshots only | store race test |

## Safety invariants

- The projection is observation-only and cannot activate either market.
- No live broker mutation or operational toggle is introduced.
- KR and US identities are never inferred across map keys.
- SSE and point reads use the same engine-owned source.
