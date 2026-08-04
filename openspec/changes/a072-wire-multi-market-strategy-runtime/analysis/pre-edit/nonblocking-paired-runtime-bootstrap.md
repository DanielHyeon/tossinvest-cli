# Non-blocking paired runtime bootstrap — pre-edit map

## Invariant

`engineRuntime` must construct and start reconciliation, exit observation and
fill detection without waiting for the KR/US strategy authority graph. Strategy
authority collection may perform official and SQLite reads, so it belongs in the
already bounded `strategy-entry` children after the runtime starts.

## Current branches

- `engineRuntime` synchronously calls
  `Context.NewPairedStrategyEntryProductionAssembly` before `NewRuntime`.
- that assembly loads schedule, candidate, route, FX, proposal, risk and account
  authority for both markets.
- a slow or unavailable strategy read therefore delays every safety loop.

## Target branches

- construct exactly two dormant refresh workers synchronously, one for KR and
  one for US;
- start both behind the existing shared barrier and enqueue both immediately;
- each child calls the coalesced paired refresh, so one authority wave serves
  both children without serializing their dispatch paths;
- public `Trigger` remains disabled while a worker is dormant;
- failures in refresh-only work remain market-local and cannot stop safety
  loops or create entry authority.

## Branch test map

| Branch | Expected proof |
|---|---|
| runtime construction | uses the non-blocking paired refresh supervisor |
| KR/US bootstrap | exactly two dormant refresh workers exist |
| public trigger | remains `DISABLED` before authority is collected |
| runtime loop set | still contains one `strategy-entry` outer loop |
| slow authority source | cannot delay `engineRuntime` construction |

