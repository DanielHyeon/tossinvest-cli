# EntryGate ABA generation pre-edit map

## Safety invariant

An entry-gate authority observed before dispatch cannot become valid again after an allowed → blocked → allowed ABA transition. Every effective account- or symbol-gate mutation advances one monotonic process-lifetime revision, and the authority digest binds that revision.

## Existing branches

- `Block`/`Clear` mutate account latches.
- `BlockSymbol`/`ClearSymbol`/`ClearSymbolReason` mutate scoped latches.
- `RebuildReconcileProjection` replaces reconcile-family projections.
- `ObserveStrategyEntryGate` refreshes durable reconciliation, checks admission, and currently returns a constant generation.

## Required branch changes

- Store a non-zero monotonic revision in `EntryGate`.
- Advance it only when effective gate state changes; repeated identical block/clear operations do not create false transitions.
- Bind revision into the sealed authority digest and expose an exact-authority comparison immediately before transport.
- Test account and KR/US symbol ABA transitions, including repeated no-op mutations.

## Restart scope

The authority is consumed only by a lease owned by the same engine process and owner epoch. Restart recovery advances the dispatch owner fence and cannot reuse the prior lease. The monotonic requirement therefore applies over the gate/owner lifetime, while the durable owner epoch supplies the restart fence.
