# SHADOW re-freeze 재검 voice2 — SubagentHandback 원문 (agent a9cea4756f3b985f0, 저장 2026-10-04)

**FAIL**: P1-1 is only PARTIAL, and v2 opens one new P1. P1-2, P1-3 and P1-4 are CLOSED.

I did not read or search anything under `~/.codex`. I also did not open the `_codex/` folder in the review directory. Everything below is measured on the 2817064c copy at `/tmp/rv2r-2817064c`. I checked the v2 brief hash: sha256 `7f1e840c…37bd`, same as stated.

## My first-round P1s

| ID | Verdict | Where v2 addresses it | Evidence |
|---|---|---|---|
| P1-1 (shadow inputs come only from arbitration winners, so OFF lanes always see nothing) | **PARTIAL** | v2 §4 (collect inputs before the gate, carried in a read-only `strategyShadowInputs` type), §2 census, §8 "at least one WOULD_EMIT" precondition | **Closed:** inputs are now taken before the gate, the carrier type is fenced off from `dispatchHandoffs` / `ResultAuthority` / `entries`, and §8 forbids passing on an empty sample.<br>**Still open:** v2 §4 says none of the 13 market-closing branches carries shadow inputs. In a declared market with some families ON, an OFF family that is the only proposer for one symbol erases that scope. That closes the whole market with `FAMILY_GATE_CLOSED`. So the shadow for that wave is "no observation" for all four lanes — including the OFF lane whose own proposal is the counterfactual being asked about.<br>**Probe rerun at 2817064c** (Go code is unchanged from c7219640: `git diff --stat` over internal/, cmd/, docs/api is empty): case B gives `reason=FAMILY_GATE_CLOSED`, routed 2, gatedCount 1, `gated=[DORMANT]`, entries 0. The inputs gathered before the gate were 2, and §4 throws both away.<br>§8's precondition passes on a fixture like case A (an ON lane co-proposes the same symbol, `ready=true`), so this blind spot is never tested.<br>**Fix:** on the closures that happen after coordination (`FAMILY_GATE_CLOSED`, collision, overflow, arbitration refused, unresolved, no accepted scope), carry the inputs already collected before the gate. Or record the blind spot in spec/design as an accepted narrowing and pin case B's expected projection. |
| P1-2 (amendment narrowed the restart scenario to "no pin") | **CLOSED** | amendment v2 spec :91-92 (now "no valid shadow manifest: no pin, or pin present but file missing, digest mismatch, expired, revoked, or binding mismatch"), v2 §10 ①②③ | The scenario covers both the pin-absent and the pin-present-but-unusable restart. Restart with a still-valid pin is pinned as: all UNOBSERVED before the first wave, SHADOW only from re-reading the manifest, previous wave/outcome memory not restored. `openspec` validity of the c7219640 version was confirmed in round 1. |
| P1-3 (sharing code would mean editing the High-risk activation loader, with no FLM) | **CLOSED** | v2 §3 (activation loader not edited, copy plus AST pins on both sides, separate sentinels with mutually exclusive `errors.Is` pins, cross-decode rejection pinned in both directions), §2 separate package, §13.2 AST baseline for "unedited" | The paths that could flip decision 62 (a shared sentinel, or a refactored classification) are gone. Only one non-zero `FamilyActivation{…}` literal exists outside test files (`production_family_activation.go:493`), plus the tagged test seam `:56`. That matches census ①. |
| P1-4 (tasks 7.3.1 acceptance still said "signed") | **CLOSED in the working tree only, not landed** | `tasks.md:563` in the working tree | Now reads: "valid server-owned shadow manifest bound by its deployment digest pin `TOSSOS_STRATEGY_FAMILY_SHADOW_<MARKET>_MANIFEST_SHA256` (개정 … 결정 63)". It is not in 2817064c (the file shows ` M`), so it closes when committed. |

## New P1 opened by v2

**Separate package (§2) contradicts helper reuse (§3).** v2 puts the shadow loader in a new package, `internal/strategyshadow`. §3 says it reuses `readProductionRouteFile` ("already shared") and keeps the lane descriptor table in one place. The code does not allow that from another package:
- These are all unexported in `strategyrouter`:
  - `readProductionRouteFile` and `productionRouteOwnerUID` — `production_owner_unix.go:10,12`, with a `!unix` stub in `production_owner_other.go:7-8`.
  - `productionRouteDigest`, `productionRouteTime`, `productionRouteIdentity`, `productionRouteDigestValid` — `production.go:768`, `:751`, `:756`, `:760`.
  - `productionRouteDescriptors`, the lane ID / horizon / version table the "exactly four lanes" check needs — `production.go:586`.
- The only exported accessor is `ProductionLaneFamily(market, laneID)` (`:581`). It returns the family only: no lane list, horizon or version.
- `readProductionRouteFile` is not a "neutral error" either: it returns `ErrProductionRouteUnavailable`, a strategyrouter sentinel (`production_owner_unix.go:15,19,27,31,35`).

So the implementation has to pick one of two paths, and the design names neither:
- **(a) Export these helpers from strategyrouter.** That widens a High-risk package's API, and those helpers are not on the FLM list.
- **(b) Copy them into strategyshadow.** That gives a second descriptor table — the activation loader's own comment forbids re-copying that table (`production_family_activation.go:537-539`) — and copies of the build-tag variants. Both would then need table-equality and both-sides pins.

Whichever is chosen, the shared read error must be wrapped with `%v`, not `%w`, as the activation loader does. Otherwise shadow errors also satisfy `ErrProductionRouteUnavailable`.

## P2s that v2 made worse

None. My first-round P2-1 to P2-5, P2-8 and P2-9 are resolved or improved by v2 (§6, the :38 fix, decision 63-v2, §10, §9, §7). I did not re-check P2-10 (stale ROADMAP text), since the ROADMAP is now being edited in the working tree.

## Repo unchanged
- **Start:** HEAD `2817064cc1f78c4d0d6abc04211de906a7c3efc5`. Status: ` M docs/ROADMAP.md`, ` M …a112…/tasks.md`, `?? .reticle-setup-crash.log`, `?? …/review-shadow-freeze-2026-10/`, `?? …/shadow-2026-10/`, `?? w4.log`.
- **End:** identical, same HEAD and same status.
- I only read the repository (status, rev-parse, diff, archive). All writes went to `/tmp/rv2r-2817064c` (probe test `internal/app/engine/zz_rv2_shadow_input_probe_test.go`).
- Correction from round 1: the a100 files that showed up as modified at the end of round 1 are gone from the working tree now. They appear in the 2817064c diffstat, so they were committed or reconciled elsewhere, not by me.
