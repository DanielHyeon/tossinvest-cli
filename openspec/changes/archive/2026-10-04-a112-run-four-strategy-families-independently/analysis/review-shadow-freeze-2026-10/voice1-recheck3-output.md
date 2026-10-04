# SHADOW re-freeze 재검 3라운드 voice1 — SubagentHandback 원문 (agent a8ef03efb4591dbc6, 저장 2026-10-05)

**Overall: FAIL** (voice 1, re-check round 3). P1-1, P1-3/N1, P1-4, P1-5 and the unsafe/reflect note are CLOSED. N2 is PARTIAL. v3 opens one new P1, N3. No P0.

Safety: I read nothing under `~/.codex`. The repo-local `analysis/.../_codex/` directory was not opened. I edited nothing in the repo. Experiments ran in a scratchpad copy of 4d22d726. Brief v3 sha256 matches (f7883a47…30fc). Between 2817064c and 4d22d726 only `design.md` changed (one line), so every code coordinate from round 2 still holds.

| Round-2 item | Verdict | v3 coordinate | Evidence |
|---|---|---|---|
| P1-1 (OFF-only scopes close the market, so SHADOW sees nothing) | **CLOSED** | §4 "carry range" | Success plus the six closures after coordination now carry the pre-gate inputs. My round-2 repro cases (an OFF family proposing alone → FAMILY_GATE_CLOSED; a declared all-OFF market → FAMILY_GATE_CLOSED) are all closures after coordination, so they are now carried. The seven closures before coordination report "no observation", which is honest. Ungraded note: the lineage-collision closure returns in the middle of the loop (strategy_market_coordinator.go:113-115), so the batch it carries covers only the routes processed so far. Pin or document that. |
| P1-3 / N1 (carried inputs inside the order-path struct, untyped after the accessor) | **CLOSED** | §4 (separate value beside the authority, opaque `strategyworker.ShadowInput` accepted only by `ShadowVerdict`); §2② census by use (`types.Info.Uses` plus expression types, with a positive control) | `strategyProposalMarketAuthority` and `strategyMarketArbitration` no longer hold the inputs, so `dispatchHandoffs`, `dispatchHandoff` and `authorityForOwnerScope` cannot reach them. Because `ShadowInput` is not `Input`, `admit` and `Submit` cannot take it. Ungraded wording note: the census works per function, so "zero in coordinateMarketProposals' coordination path" cannot be expressed. That function must be allowlisted because it is where the inputs are collected. Better: one call to a helper, pinned at statement level. The §8 differential is the real guard there. |
| P1-4 (shadow latency before dispatch) | **CLOSED** | §5 (starts in the cycle closure after `runProductionStrategyMarketCycle` returns; async, single-flight; constant deadline; separate ctx; AST pins ①②; deadline-overrun fault pin) | The order path no longer waits on the shadow step, and the last-statement pin (`a112_market_delivery_structure_test.go:169-175`) stays intact. The new risk this placement creates is N3. |
| P1-5 (helpers unreachable from a separate package) | **CLOSED** | §3 (`production_shared_export.go` with neutral wrappers, existing functions untouched, one descriptor table, read errors folded with `%v`, `errors.Is(shadowErr, ErrProductionRouteUnavailable)==false`, wrapper layer inside the AST pin) | Attempted attack: exporting the table could let strategyshadow mutate the activation validator's shared map. It cannot: `productionRouteDescriptors` returns a **new map literal on every call** (production.go:586-603), so exporting it shares no mutable state. The field type `productionLaneDescriptor` has exported fields (:573-577). Exporting the file reader adds only a read. |
| N2 (SHADOW not cleared within the process) | **PARTIAL** | §5.1 (freshness: shadow counts only when its wave equals the latest evaluate wave; pins (a) revoked, (b) expired, (c) closed, (d) failed) | Closed when new waves keep coming. Open: the rule is relative to waves only. If no new evaluate wave arrives, the last SHADOW stays "fresh" indefinitely, even past the manifest's `expires_at` or after revocation. That happens when the refresh fails (`runProductionStrategyMarketCycle` returns before `evaluate`, strategy_entry_supervisor.go:515-518) or when the market is latched (`evaluationState` blocks cycles, :1035). That contradicts :89 "only while a valid manifest exists". Fix: carry the manifest's `expiresAt` with the observation, drop it at projection time against `now`, or clear shadow observations on any cycle error. Add pin (e): no new wave plus manifest expiry → UNOBSERVED. Also the wave number must be the one captured with the batch, not the one current when the result is recorded (see N3). |
| unsafe/reflect note | **CLOSED** | §2 (strategyshadow's dependency closure must not include `unsafe` or `reflect`; walked with `testenv` ListDeps in four modes) | — |

## New P1 opened by v3

**N3. §5 never says where the post-cycle shadow step gets its same-wave data. The obvious source lets SHADOW latch the other market.**
- The shadow step runs after `runProductionStrategyMarketCycle` returns. That function keeps `fresh` local (strategy_entry_supervisor.go:515) and, because of the last-statement pin, returns only the delivery error. So the closure must get the shadow batch, the activation and the evaluate wave from somewhere else, and §5 does not say where.
- The natural call, `refreshPairedStrategyEntryProductionAssembly`, returns the cached assembly only within 1s of the refresh start (strategy_refresh_wave.go:73). After that the caller becomes the **leader of a new remote refresh wave**, and the leader's ctx drives `NewPairedStrategyEntryProductionAssembly(ctx, clk)` (strategy_entry_supervisor.go:592-594).
- The shadow ctx has its own 2s deadline. If a shadow-led wave is cancelled by that deadline, the error becomes `wave.err`. The other market's real cycle that joined the wave receives **the same error** (`awaitStrategyRefreshWave`, strategy_refresh_wave.go:149-155; `publishStrategyRefreshWave`, :88-101), its cycle fails, and `latchMarket` runs (:1093-1124).
- That path also breaks "no new remote I/O" and puts the shadow batch on a different wave from evaluate, so the freshness rule discards it and the projection stays empty.
- Census ③ may block the refresh call if "gateway" covers gateway reads, but the design names neither the hazard nor the source.
- Required:
  - The shadow step reads only an already published snapshot, e.g. the cached `c.strategyRefresh` or a value the closure is given, stamped with the wave it came from.
  - It must never call `refreshPaired…`/`joinStrategyRefreshWave` or become a wave leader. Put those calls in census ③'s forbidden set.
  - Add a fault pin: the shadow deadline expires while the other market's cycle is in progress → that market is not latched.
  - The shadow goroutine reads and writes lane runtime memory concurrently with the next `evaluate` and with `recoverMarketLanes` swapping lanes. Name it explicitly in the `make test-race` list.

## Repo unchanged (start = end)

```
 M docs/ROADMAP.md
 M openspec/changes/a112-run-four-strategy-families-independently/tasks.md
?? .reticle-setup-crash.log
?? openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/
?? openspec/changes/a112-run-four-strategy-families-independently/analysis/shadow-2026-10/
?? w4.log
HEAD 4d22d72663dcf3a42be928b3ecf3db817ffc6ac5
```
