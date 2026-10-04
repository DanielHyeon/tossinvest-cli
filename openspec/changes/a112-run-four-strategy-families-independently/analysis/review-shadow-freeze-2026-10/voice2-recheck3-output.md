# SHADOW re-freeze 재검 3라운드 voice2 — SubagentHandback 원문 (agent a9cea4756f3b985f0, 저장 2026-10-05)

**PASS**: both of my remaining round-2 items are CLOSED, and v3 opens no new P0 or P1. Two small fixes are worth folding into the design (notes 1 and 2 below).

I did not read or search anything under `~/.codex`, and did not open the review folder's `_codex/` either. Everything below was checked on a copy of commit 4d22d726 at `/tmp/rv2r3-4d22d726`. The brief v3 hash matches: sha256 `f7883a47…30fc`. Go code is unchanged from 2817064c (`git diff --stat` over `internal/`, `cmd/` and `docs/api` is empty), so my round-2 probe results still apply.

## Results

| Round-2 item | Verdict | Where v3 fixes it | Evidence |
|---|---|---|---|
| P1-1: shadow inputs dropped on every market-closing branch | **CLOSED** | §4 (carry range: the six closures after coordination plus success carry the inputs collected before the gate; the seven closures before coordination report "no observation") and §8 (the precondition fixtures must include a symbol proposed only by an OFF family) | My round-2 case B is now explicitly carried: a declared market, one OFF family the sole proposer for a symbol, ending in `FAMILY_GATE_CLOSED` with routed 2, gated 1, `[DORMANT]`. §4 cites this population by name, and the §8 non-empty-sample check now covers it. The carrier sits outside the authority struct (a separate `strategyShadowBatch` and an opaque `ShadowInput`, not `strategyworker.Input`). So the handoff source `proposals.forMarket(m).dispatchHandoffs()` is structurally untouched by shadow data. |
| New P1: separate package vs unexported strategyrouter helpers | **CLOSED** | §3 (one-line exported wrappers in a new file `production_shared_export.go`, existing functions unedited, a single descriptor table, read errors folded with `%v`, plus a pin that `errors.Is(shadowErr, ErrProductionRouteUnavailable)` is false) | **Sketch on the copy:** I added the wrapper file and called it from a separate package `internal/zzrv2shadowsketch`. It compiles and passes. The exported descriptor table has 4 entries, and deleting from the returned map does not affect the next call, because `productionRouteDescriptors` builds a fresh map each time (`production.go:586`). For the read error, folding with `%v` gives `errors.Is(route)=false`, while `%w` gives `true`, so the planned pin can tell the two apart. The existing `TestExternalAPIExposesNoAuthorityMintingConstructor` and `TestPackageHasNoMutationAuthorityOrRuntimeDependency` still pass with the wrapper file present. No strategyrouter test freezes the exported API (no `IsExported` census). |

## New P0/P1 opened by v3
None. Notes that do not block:

1. **Lineage-collision closure carries an incomplete batch.**
   - v3 §4 lists "계보 충돌" (lineage collision) among the carried closures. But on a collision, `coordinateMarketProposals` returns in the middle of the loop (`strategy_market_coordinator.go:113-115`, `return arbitration, refused`).
   - So the carried batch only holds the proposals seen before the collision, and lanes after that point would show a false NO_INPUT.
   - It is rare (an integrity fault) and raises no exposure.
   - Fix: either treat collision as "no observation", or finish collecting the shadow inputs before returning. Pin either choice with one test.
2. **Async shadow step must stamp the wave number captured at its start.** This is a clarification for the race v3 introduces.
   - §5/§5.1 should state that the shadow step records the wave number captured at start, taken from the same assembly as its inputs, and never the current counter at write time.
   - If it stamped at write time, a slow wave-N shadow could land as wave N+1. Then N+1's activation might have turned that lane ON, and the validator rule "SHADOW ⇒ desired=OFF" would reject the whole snapshot on every external read path. That fails closed (no exposure), but the projection would be lost.
   - Fix: one sentence in the design, plus one fake-clock test.
3. **Exported wrappers are read-only.** Nothing they expose can mint authority, so widening the strategyrouter API this way does not raise exposure.

## Repo unchanged
- **Start and end are identical.** HEAD `4d22d72663dcf3a42be928b3ecf3db817ffc6ac5`. Status: ` M docs/ROADMAP.md`, ` M …a112…/tasks.md`, `?? .reticle-setup-crash.log`, `?? …/review-shadow-freeze-2026-10/`, `?? …/shadow-2026-10/`, `?? w4.log`.
- I only read the repository (status, rev-parse, diff, archive).
- The sketch files exist only in the copy: `/tmp/rv2r3-4d22d726/internal/strategyrouter/production_shared_export.go` and `/tmp/rv2r3-4d22d726/internal/zzrv2shadowsketch/sketch_test.go`.
