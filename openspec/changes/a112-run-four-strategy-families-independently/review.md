# Planning Review

## Function Logic Map: not-applicable

This change-creation diff modifies OpenSpec/PM planning artifacts only and changes no existing Go function body. The marker applies only to the current planning diff; it is not an implementation exemption. Before any existing function listed in `analysis/pre-edit-targets.md` is edited, tasks 1.1-1.3 require a fresh CodeGraph impact query plus current-base Go AST, Function Logic Map, Branch Test Map and risk-pattern report.

## Planning validation

- Exact current 6-lane/3-family gates and shared runtime/dispatch/protection boundaries are recorded in `analysis/current-main-evidence.md`.
- The proposed topology preserves eight pure evaluator instances, two market coordinators and one shared owner/Guardian/risk/dispatch/Gateway authority.
- Every new descriptor and runtime state remains OFF/UNOBSERVED; this planning change grants no activation or mutation authority.
- Strict OpenSpec validation and PM hierarchy generation/check pass. Full `make sdd-check` must be re-run after `make sdd-sync` at implementation start because the shared worktree hard-evidence index is currently stale.

## 2026-08-15 Manager proposal-freeze review

### Independent Terra planning adversary

- P0: current official lossless minute-candle authority is KR-only; generic US candle adaptation is float/time-derived and cannot satisfy M-B. Disposition: L1 and every production 8-lane wiring remain blocked until M-B PASS; L2 pure breakout core may proceed only through a fixture-only inert port after L0.
- P0: a064/a066/a070/a100 prerequisites are incomplete and A100 R0/M0 landing does not produce ProtectionReady. Disposition: L6 dispatch integration and L7 container deployment remain blocked; all exposure-raising requests must stay zero.
- P1: current `strategyrouter.Route` selects raw scores before pure evaluation. Disposition: spec now requires all-candidate sealed RouteSet and post-evaluation common calibrated arbitration.
- P1: current runtime assumes one proposal per market and two workers. Disposition: spec now requires explicit 8 `FamilyWorker`, 2 `MarketCoordinator`, owner-scope bounded queues and no market-wide single-proposal gate.
- P1: quota functions, SHADOW state, calibration singleton behavior and additive 8-lane projection were under-specified. Disposition: tasks/spec/design and pre-edit targets now freeze each contract and require RED coverage.

### Manager acceptance state

- Proposal/design/spec/tasks may be accepted only after strict OpenSpec validation and a gstack plan review report P0/P1=0.
- This planning acceptance does not authorize production code, activation, container replacement, live Broker access or order mutation.
- After A100 R0/M0 merges, L0 must re-freeze `base-commit.txt`, CodeGraph fingerprint, AST/FLM/BTM/risk targets and dependency counts before any Terra production lot starts.

### gstack plan-eng review round 1

- Date: 2026-08-15 KST. Voice: independent Terra gstack engineering reviewer. Commands: strict OpenSpec validation and diff-check PASS; review was read-only.
- P1: L0 could not own “executable” goldens under its old analysis-only boundary. Disposition: L0 now owns immutable canonical JSON/hash/validation receipts under `analysis/goldens`; L2/L3 tests must consume hash-identical fixtures.
- P1: M-B had no collector/reviewer/receipt/PASS authority. Disposition: task 0.7 now defines a read-only Terra collector, separate reviewer, Manager PASS authority, sanitized receipt schema and fail/HOLD outcome.
- P1: setup ID/correction/session semantics were ambiguous. Disposition: task/spec now fix the canonical setup-ID preimage, snapshot revision separation, terminal non-resurrection and session rollover.
- P1: quote/FX sizing units, direction, rounding and inclusive veto boundaries were incomplete. Disposition: task/spec now define quote and FX seals, versioned limits/refusal codes, exact-boundary tests and conservative rounding.
- P2: production callers could regress to legacy raw-score `Route`. Disposition: L3 owns the exact engine callers and a static guard forbids `strategyrouter.Route(` in the A112 production caller closure while preserving legacy behavior elsewhere.
- Round-2 gstack recheck initially reported P0=0/P1=0 with one L5 FLM P2. After adding the exact production bootstrap symbols, quote/FX formulas, PROPOSED→CONSUMED correction boundary, first-leg admission functions and exact RouteSet AST/import guard, the final same-reviewer recheck reported P0=0/P1=0/P2=0. Strict OpenSpec, PM hierarchy, tracked diff-check and untracked-aware EOF/whitespace checks all PASS.
- The separate Terra planning adversary also completed a final same-reviewer pass at P0=0/P1=0/P2=0 and accepted the proposal freeze.
- Proposal freeze is accepted for L0 only after A100 R0/M0 reaches main. L2 may start after Manager accepts L0. M-B/L1 and L3–L7 remain blocked by their dependency cells; no production wiring, container replacement or exposure-raising request is authorized.

## 2026-08-15 L0 current-base freeze acceptance

### Landing and ownership

- A100 R0/M0 landed through `DanielHyeon/tossinvest-cli` PR #3. The A112 frozen base is the resulting `main == origin/main == 016da6245feb60e13971388be386c2c2041469a8`.
- The Terra L0 implementer edited only `base-commit.txt` and `analysis/**`. No Go source, Go test, PM, configuration, runtime, Broker, container or activation surface was changed.
- The ownership ledger above is frozen. Any future exception requires a Manager-authored OpenSpec amendment before an implementation agent edits the file.

### Terra implementation evidence

- Current-main evidence and pre-edit targets were regenerated against `016da624…`; the prior `882a0b49` baseline was removed.
- `analysis/function-logic/inventory.json` contains exactly 38 current function records, matching exactly 38 AST/FLM/BTM/risk directories and 152 bundle artifacts. Type-only `AcquireRequest`/`Capability`, constant-only `execgw.ProfileProtection`, and the protection prerequisite row are explicitly deferred as non-function/citation evidence rather than assigned fabricated ASTs.
- Canonical goldens freeze the exact eight descriptors, four families, two markets, worker/coordinator keys, OFF/OFF/UNOBSERVED defaults, exact transition/refusal enums, calibrated arbitration, owner-scope queue, setup/correction identity, strict official evidence identity, quote/FX boundaries and conservative rounding.
- `analysis/goldens/validate-four-family-runtime.jq` returns `true`/exit 0 for the canonical golden and `false`/exit 1 for seven-descriptor, duplicate-descriptor, unknown-family and missing-descriptor mutations. The sorted SHA-256 manifest covers the four canonical JSON goldens and the jq validator.
- Dependency truth is explicit: a072 is complete 57/57; a064, a066 and a070 remain incomplete; the separate A112 paired production gate remains blocked; A100 R0/M0 is present but shipped protection remains `Wired:false`/`UNWIRED`; the official lossless US 1-minute source and M-B PASS are absent.

### Separate Terra adversarial review and fixes

- First review: P0=0/P1=4/P2=1. It found an incorrect a072 status, a nonexistent KR source citation, non-reproducible jq PASS claims, incomplete transition/refusal enums and weak SDD receipt evidence. The original implementer fixed all findings; the same reviewer verified closure.
- Manager verification then found stale `current-main-evidence.md` and missing current-base FLM bundles. The original implementer regenerated the evidence, target inventory and 38 bundles. The same reviewer found one remaining stale receiver name (`Quota.Acquire/Complete`); the implementer corrected it to `QuotaAuthority.Acquire/Complete` and refreshed the receipt hash.
- Final same-reviewer verdict: P0=0/P1=0/P2=0. No stale base, receiver, target, SHA, placeholder or unmapped function remains.

### Manager verification and decision

- Manager independently verified the exact base, 38/38 bundle inventory, all AST JSON, `check_analysis.py`, CodeGraph hard-evidence freshness, strict OpenSpec validation, jq positive/four-negative behavior, manifest hashes and `git diff --check`.
- `make sdd-check` is hard-pass; CodeGraphContext and GBrain emit advisory freshness warnings only and do not widen authority.
- L0 is **ACCEPTED**. L2 pure `internal/breakoutlane/**` fixture-only work may start. M-B/L1 and L3–L7 remain blocked by their documented dependencies, and every exposure-raising request remains zero.

## 2026-08-15 L2 pure breakout core acceptance

### Ownership and authority boundary

- The Terra L2 implementers edited only new `internal/breakoutlane/**` production and test files. SOL/Manager edited only this OpenSpec task ledger and review record.
- The package is an inert pure evaluator. Its production import closure is limited to Go standard-library value/arithmetic/hash packages and exposes no official client, network, broker, journal, Guardian, configuration/toggle writer, system clock or runtime activation authority.
- This acceptance covers tasks 3.3–3.5 only. M-B/L1 official US source and strict decoder, L3 `PROPOSED → CONSUMED` shared admission, registry/RouteSet, runtime wiring, dispatch, build and deployment remain unchecked and blocked.

### Terra RED/GREEN and first adversarial cycle

- Initial RED introduced the absent pure API and covered closed-bar state transitions, skipped/duplicate/first-touch edges, threshold boundaries, quote vetoes, exact sizing and FX rounding. The first GREEN established sealed builders, immutable evidence snapshots, deterministic setup identity and a pure evaluator.
- A separate read-only Terra adversary found mutable/raw state advancement, unsealed quote/FX liveness, zero/missing-price sizing, cost-exclusive RR, non-contiguous bars, timeout-boundary bypass, unchecked overflow, weak correction lineage and snapshot-construction gaps. The implementer added grouped REDs that reproduced each defect before repair.
- Repairs removed raw `Machine/Event/Apply` authority, made decisions opaque and sealed, required 15–512 contiguous regular-session closed one-minute bars, enforced exact KR/US timeout boundaries, used checked cost-inclusive sizing, and sealed append-only higher-revision correction lineage.
- The same adversary rechecked the frozen repair and found no remaining functional P0/P1. CodeGraph freshness was re-synchronized after source freeze rather than waived.

### gstack review and final red-team cycle

- gstack testing found missing executable boundaries for failed reclaim, minimum RR, quote age/spread/drift, both FX directions/scales, v1 thresholds and retest tolerance endpoints. gstack maintainability found duplicated v1 literals and raw state vocabulary. The implementer captured focused REDs, centralized typed v1 constants/phases and added the missing boundary tables.
- gstack security/performance found no L2 P0/P1; its source-integrity observation is retained as an L1 strict-decoder/source-seal concern and does not grant L2 production evidence authority.
- The final independent Terra red team reproduced seven P1 defects in one grouped RED: duplicate snapshots were not idempotent; stale quote/FX typed refusals were unreachable; same-currency identity FX was rejected; close outside `[low,high]` was accepted; sizing used proposed rather than worst executable entry; config digest did not bind every mutable field; and append replay permitted same-revision historical bar-content mutation.
- The implementer repaired all seven with exact-snapshot replay, structural-versus-liveness validation separation, canonical 1:1 same-currency FX, closed-bar range validation, worst-entry/cost-inclusive risk and RR, deterministic all-field `V1ConfigDigest`, and full closed-bar content lineage. The same red team rechecked the final source: P0=0, P1=0. Its sole P2 is the deliberately deferred L3 `CONSUMED` transition.
- A separate read-only Codex cross-model review independently reported **NO FINDINGS** for the seven repaired paths. Its sandbox could not create a Go build directory, so Manager and Terra test receipts, not that sandbox, are the executable test authority.

### Manager verification and decision

- Manager independently ran `go test ./internal/breakoutlane -count=1`, `go test -race ./internal/breakoutlane -count=1`, `go vet ./internal/breakoutlane`, `go test ./internal/breakoutlane -count=20` and coverage; all passed and final statement coverage is 93.4%. The Terra freeze also records `go test ./...` and the same red-team recheck records `go vet ./...` PASS.
- Strict OpenSpec validation, PM generated-tracker check, `check_analysis.py`, both canonical jq validators, the sorted golden SHA-256 manifest, GOROOT gofmt, tracked and untracked-aware whitespace checks, `make sdd-sync`, `make sdd-check` and hard index freshness all pass; CodeGraphContext/GBrain warnings remain advisory only.
- L2 is **ACCEPTED** with P0=0/P1=0. Tasks 3.3–3.5 are complete. Broader RED-contract tasks 2.2–2.5 remain unchecked because L2 intentionally does not implement `CONSUMED`, the L1 strict decoder/dual-cutoff source, restart persistence or shared first-leg admission. No deployment, activation, Broker request, order, config, engine or container action is authorized by this acceptance.

## 2026-08-15 M-B official US source measurement

### Terra collector receipt

- The Terra collector wrote only `analysis/measurements/m-b-us-source/**` and made no external API, Broker, auth/config, verify, engine/container, build/install, staging, commit or push call.
- The sealed self-excluding manifest covers the sanitized source identity, request contract, raw-payload absence, pagination absence, session/calendar contract, rate-budget contract, redaction and command records. Every listed SHA-256 verifies.
- Static current-base evidence proves `official.Client.RawMinuteCandles` preserves decimal/timestamp strings but rejects every non-KR market before `Client.get`. The generic US-capable candle/price/orderbook paths parse decimal strings to float values, parse timestamps to `time.Time`, add local fetch time or permit a `source=both` WTS fallback, so none is an allowed lossless production authority.
- Because the lossless US precondition fails before HTTP, the collector deliberately sent no authenticated request. Raw US bar/quote bytes, a terminal pagination transcript, observed USD/session-calendar join and observed rate headers therefore remain explicitly `MISSING` rather than being inferred from static OpenAPI documentation.
- The `/mnt/D` fuse mount reports the receipt directory and payloads as `0755` after successful chmod attempts. The receipt records the required 0700/0600 modes as `MISSING`; it does not claim a permission PASS.

### Separate Terra adversarial review and Manager decision

- The separate read-only reviewer independently verified every manifest hash, source/blob/base identity, the pre-HTTP US guard, float/time/WTS rejection facts, all explicit absence states, static AAPL/USD/1m/cursor/rate contract citations, redaction and no staged M-B changes. Evidence-lot verdict: P0=0/P1=0/P2=0.
- Operational M-B verdict is **HOLD**, not PASS. The exact lossless official US source, raw payloads, terminal pagination, observed calendar/currency and rate evidence are absent, and the receipt permissions are not verifiable on this mount.
- Task 0.7 remains unchecked. L1 and L3–L7 remain blocked; L2 acceptance is unaffected because it consumes inert fixtures only. No production wiring, activation, container replacement or exposure-raising request is authorized.

## 2026-08-15 M-B0 measurement-seam amendment freeze

### Scope challenge and decision

- Market closure is not the blocker for non-LIVE work. L0 and L2 are accepted, and the static M-B evidence lot is accepted; operational M-B remains HOLD because current main has no lossless US raw source.
- Two separate Terra read-only design reviewers rejected widening `RawMinuteCandles`, generic candle/quote adapters, ordinary `Client.get/send`, `AttemptTrace/RateBudget`, `AuthHeaders` or a product CLI/runtime caller. Those paths respectively broaden KR authority, admit float/time/WTS evidence, refresh/write credentials, detach body from raw rate evidence or expose account/token data.
- Manager chose the smaller isolated design: M-B0 owns only new `internal/official/a112_mbus_read*` files and has zero production callers. Existing official client/token/trace/raw-reader bodies remain read-only. A later M-B1 tool is separately owned under `tools/a112-mb-us-source/**`, is not installed or linked into `tossctl`, and cannot start until M-B0 is accepted.

### Frozen safety and authority boundary

- M-B0 revalidates one exact `*official.Client` instance under its configuration lock; it accepts no caller-minted origin token, separate provider or cross-client trace/result. It is cached-token-only with an every-component no-symlink current-UID regular 0600 cache and the existing 60-second skew, proxy-free, redirect-refusing, GET-only and single-attempt. Each request requires a caller deadline no later than 15 seconds, sends only Authorization/optional Accept and caps the body at 2 MiB. It never exchanges or refreshes OAuth, writes token cache, resolves an account, retries/falls back, calls WTS/hybrid, or invokes broker/journal/config/engine/container surfaces.
- Request identity is exact `US/AAPL/1m/count=200/adjusted=false`. Each opaque result binds the same request's descriptor, capped raw body, status, absent/null/empty/string cursor and allow-listed raw rate headers before decode. Existing KR `RawMinuteCandles` continues to reject US.
- M-B1 may only crawl four candle pages from an explicit initial cursor and exact session date, plus one orderbook and one calendar GET, in a mode-verified POSIX `/tmp` receipt. Each request has a 15-second deadline, the whole run has a 120-second deadline, and no request starts with less than 15 seconds remaining. It passes the decoded prior cursor bytes without normalization and preserves raw cursor JSON separately; terminal null missing within four pages, loop, unsafe FUSE mode, redaction failure or response ambiguity is HOLD with no fallback.
- Official candle schema has no publisher `closed/finalized` bit or server-observed timestamp. Therefore even a future Manager M-B PASS authorizes only L1 implementation start. L1 dual-cutoff/finality/correction RED→GREEN and independent acceptance remain mandatory before any US closed-bar or production evidence authority exists.

### Review disposition

- First same-scope amendment review found P0=0/P1=5/P2=2. The isolated new-file design, four-page HOLD and finality non-authority were correct, but path TOCTOU, automatic compression, duplicate/empty rate-budget continuation, cursor byte ambiguity and run/source identity were incomplete.
- Manager accepted all five P1s. The amendment now requires FD/openat/Fstat no-follow token and receipt access, disabled compression plus identity encoding/default-User-Agent suppression and limit-plus-one reads, slice-cardinality rate validation with next-page budget gate, decoded cursor bytes plus forensic raw JSON, explicit token-cache/default-client handoff, pre/post source/module/base/build/executable SHA binding, and a process-clock boundary limited to token expiry/deadlines with no serialized or inferred source/finality authority. The two P2s are closed by a 120-second overall budget with a 15-second start reserve and exact definition/test/tool symbol-reference guard exemptions.
- The same Terra reviewer rechecked the frozen amendment after the Manager fixes and returned `P0=0 / P1=0 / P2=0`. Descriptor-relative token/receipt access, exact on-wire transport behavior, rate-header cardinality and continuation budget, cursor byte identity, explicit token-cache construction, bounded deadlines, source/build identity, symbol-reference closure and clock non-authority are all closed. M-B0 implementation may start only in the exact new `internal/official/a112_mbus_read*` files; every existing official/product file remains read-only.
- A separate gstack engineering plan review then returned `P0=0 / P1=0 / P2=0` and `SHIP` for M-B0 implementation only. It confirmed that same-package new files can lock and revalidate the private `Client` authority state, clone the sealed transport, use the repository's existing Unix `openat/O_NOFOLLOW/Fstat` pattern and preserve the KR raw-reader refusal without editing any existing official body. M-B1 and every product/source/finality/activation path remain HOLD.
- M-B0 proposal P0 remains 0: unbounded-history/rate-abuse and false-finality risks are closed by the finite cap/HOLD rule and explicit non-authority rule. This verdict authorizes only the M-B0 Terra RED→GREEN lot. M-B1, M-B PASS, L1, runtime wiring, activation, dispatch and deployment remain blocked.
- M-B0 tasks were unchecked at proposal freeze. They were completed only after the following implementation, adversarial, gstack and Manager verification record. M-B1 and parent M-B task 0.7 remain unchecked. No external request, commit, deployment, activation or exposure action is authorized by M-B0 acceptance.

## 2026-08-15 M-B0 implementation acceptance

### Terra RED/GREEN and ownership

- The Terra implementer created only `internal/official/a112_mbus_read.go`, `a112_mbus_read_unix.go`, `a112_mbus_read_unsupported.go`, `a112_mbus_read_test.go` and `a112_mbus_static_test.go`. Every existing official/product file remained read-only, and no external API, OAuth exchange/cache write, account discovery, order, configuration, runtime, container, install, stage, commit or push action occurred.
- The initial corrected RED failed only for absent M-B0 APIs. Later adversarial REDs reproduced foreign token-manager/client binding, invalid initial UTF-8 cursor network entry, duplicate/invalid/surrogate cursor normalization, null-result evidence, exported/private same-package caller-closure gaps and path-insensitive AST false positives before each repair. Already-correct proxy removal, exact body/rate and clock boundaries were recorded as coverage additions rather than retroactive RED claims.
- GREEN binds the exact Client/base/transport/token-manager instance, reads the cached token through Unix descriptor-relative no-follow/Fstat checks, performs one proxy-free/identity-encoded/redirect-refusing GET, retains bounded raw body/cursor/rate evidence and exposes copy-on-read results. The repository-wide AST guard rejects exported selectors and same-package private M-B0 capabilities outside exact definitions/tests/future tool while allowing unrelated same-named packages.

### Independent reviews

- The separate Terra adversary initially returned HOLD for duplicate/invalid UTF-8 cursor decoding and an incomplete caller guard. Its repeated same-review cycles found and forced closure of unpaired surrogate replacement, private same-package helper access and a path-insensitive false positive. Final verdict is `P0=0 / P1=0 / P2=0`, ACCEPT.
- The separate gstack implementation review initially returned HOLD because frozen proxy, same-instance, exact Content-Length, full rate-header and deterministic clock boundaries lacked executable tests. After the implementer added those tests and closed the resulting cross-client/input/result/AST defects, final gstack verdict is `P0=0 / P1=0 / P2=1`, ACCEPT. The only P2 is `/mnt/D` FUSE reporting source files as 0755; landing must force every new Go file to index mode 100644 and verify the cached summary.

### Manager verification and decision

- Manager independently ran focused M-B0 tests, focused race, `-count=20`, `go vet ./internal/official`, full `internal/official`, strict OpenSpec, logic-map, PM tracker, GOROOT gofmt, diff whitespace and Windows/Darwin compile-only checks; all passed. `make sdd-sync && make sdd-check` passed with CodeGraph hard evidence current and GBrain advisory stale only.
- `go test ./... -count=1` reached the pre-existing `internal/journal` default 10-minute timeout while every other emitted package, including `internal/official`, passed. This is not recorded as a PASS. Manager reran `go test -timeout 30m ./internal/journal -count=1`, which passed in 361.656 seconds, closing the only timed-out package without misrepresenting the original command.
- M-B0 is **ACCEPTED** and tasks 0.7a–0.7a.3 are complete. Parent M-B task 0.7, M-B1, M-B PASS, L1, source/finality authority, runtime wiring, activation, dispatch and deployment remain HOLD. M-B0 has zero production callers and creates no exposure-raising authority.

## 2026-08-15 M-B1 implementation acceptance

### Terra RED/GREEN and ownership

- The Terra implementer created only the non-installed `tools/a112-mb-us-source/**` collector. It constructs `official.New(official.Credentials{}, tokenCache)` and calls only the three accepted M-B0 reads. No existing product/runtime file, external API, OAuth exchange, account/order/config/engine/container surface, installation, stage, commit or push was touched.
- Initial REDs covered the absent collector, terminal pagination, first-error stop, secure receipt and cursor continuity. Independent-review RED rounds then reproduced cancellation during finalization, the exact 120-second boundary, nested secret values, unexpected receipt entries, caller-claimed build commands, untracked-content drift, executable mismatch, invalid frozen base, inherited `GOFLAGS=-toolexec`, missing `go mod verify`, omitted actual compiled inputs, broad-prefix evil files, payload/manifest byte tamper, omitted canonical query, nested duplicate-key secret hiding, hostile Git environment, source symlink, run-directory chmod drift, 256-MiB identity truncation and 64-MiB receipt tail append. Each named counterexample failed before its corresponding repair and passed afterward.
- GREEN enforces a single overall deadline and per-request budget; exact M-B0-only reads; recursive JSON/redaction failure; pinned-FD receipt creation, complete content rehash and strict self-excluding manifest; exact canonical query; absolute sanitized Git/Go provenance; offline module verification and actual compiled-input closure; exact GOOS-bound production manifest; reproducible executable identity; and fail-closed source/receipt size limits. The tool remains uninstalled and has no production caller.

### Independent reviews

- The separate Terra implementation adversary repeatedly returned HOLD for the finalization, redaction, receipt, build-provenance, module/cache, compiled-closure, exact-manifest, directory-mode and truncation boundaries. After the final 64-MiB tail-append repair, the same reviewer returned `P0=0 / P1=0 / P2=1`, ACCEPT. Its P2 is only `/mnt/D` FUSE exposing new Go files as 0755; any later landing must force and verify Git index mode 100644.
- The separate gstack implementation review initially returned HOLD for payload-integrity manifest verification, omitted canonical query, recursive duplicate-key redaction and inherited Git provenance; its subsequent review found the exact 64-MiB receipt tail case. After repairs, the same gstack reviewer returned `P0=0 / P1=0 / P2=1`, ACCEPT with the same index-mode hygiene condition.

### Verification and Manager decision

- Terra verification passed focused M-B0/M-B1 suites, race, `-count=20`, selected hostile cases at higher repeat counts, full vet, the real prescribed offline rebuild including inherited-`GOFLAGS` attack, Linux/Windows/Darwin builds, strict OpenSpec, logic-map, PM, GOROOT gofmt and diff checks. A prior full repository run passed including `internal/journal` in 586.856 seconds. The final repair's duplicate full run passed the M-B1 tool and every emitted package except the pre-existing `internal/journal` default 10-minute SQLite/FUSE contention timeout; its named timed-out test passed alone in 0.825 seconds. The timed-out aggregate command is not represented as a PASS.
- `make sdd-sync && make sdd-check` was rerun only after the final source and OpenSpec freeze; CodeGraph hard evidence is current. Advisory GBrain freshness does not grant or remove authority.
- M-B1 implementation is **ACCEPTED** and tasks 0.7b–0.7b.1 are complete. Task 0.7b.2 has not run: no external official read-only measurement or raw receipt exists. Parent task 0.7, Manager M-B PASS task 0.7c, L1/L3–L7, source/finality authority, activation, dispatch and deployment remain HOLD.

## 2026-08-15 M-B1 first authorized measurement HOLD and provenance reopening

- The human-authorized Terra collector completed cache and POSIX-root preflight, selected explicit historical US session date `2026-08-14` with an explicitly empty initial cursor, and started the prescribed offline identity path. It stopped before the first official reader call with sanitized error `pre-request identity: Go runtime root is not absolute`; retry, fallback and remediation were deliberately not attempted. The claimed official external GET count is 0 and every order, auth exchange/update, config, engine/container, install, stage, commit and push action remained 0.
- The frozen HOLD receipt is `/tmp/a112-mb-us-receipt.Cryl62/a112-mb-us-b3b35e93fd6aee346a96e8aecb22e282`. Parent/run directories are current-owner exact0700 and the only entry is a current-owner regular exact0600 `TAINTED` file containing `HOLD\n`. There is no preflight, manifest, raw candle/orderbook/calendar payload or secret/account/order/token material. A separate read-only Terra receipt reviewer returned `P0=0 / P1=0 / P2=1`, `ACCEPT-HOLD`: current control flow proves this identity failure precedes every official reader, but the generic taint marker cannot independently bind executable identity or historical request count and is never PASS evidence.
- Separate gstack evidence review confirmed an implementation P1 rather than an operational outage. The prescribed rebuild always passes `-trimpath`, while `verifiedGoToolchain` requires `runtime.GOROOT()` to be absolute. Go 1.26.5 linker behavior omits `runtime.defaultGOROOT` under `-trimpath`; the prescribed binary therefore cannot satisfy its own preflight. The real prescribed-build tests passed because they build but do not execute that binary through identity preflight. Exporting ambient `GOROOT` is rejected as caller-controlled provenance.
- Manager reopened task 0.7b.1 and added 0.7b.1.1. The repair must require an explicit absolute `--go-binary`, verify and rehash it pre/post, use it consistently for every Go closure/module/build operation, ignore ambient locator/build controls, and add a real trimpath-built offline identity-preflight integration RED/GREEN before another external run. Task 0.7b.2, 0.7c, parent 0.7, M-B PASS and L1/L3–L7 remain unchecked/HOLD.
- The first locator repair captured the missing-config RED and made the real trimpath identity-preflight and standalone missing-cache processes pass offline. A separate Terra adversary nevertheless returned `P0=0 / P1=1 / P2=1`: Go and Git commands still reopened their verified pathnames with `exec.Command`, so a writable-parent swap-after-hash/restore-before-post-hash attack could execute different bytes. Manager added task 0.7b.1.2 requiring exact-byte private execution snapshots and clarified that tracked compiled dependencies are frozen by base/diff/content/closure hashes while untracked M-B0/tool production inputs remain exact-GOOS allow-listed. External rerun remains forbidden until the same reviewer and gstack both return P0/P1=0.
- The private executable RED/GREEN then reproduced the swap/restore attack and closed direct Go/Git pathname execution, but the real preflight exposed Go relocation: a copied Go binary cannot find its distribution, while setting `GOROOT` back to the caller-derived original tree would reopen compiler/source swap authority. Manager added task 0.7b.1.3 requiring a bounded descriptor-walked, content-manifested private Go distribution root, per-command revalidation and cleanup. Ambient or original GOROOT remains forbidden and the external run remains HOLD.
- The first secure distribution implementation measured about 2m20s for 269 MiB/15,026 files with per-file fsync, already exceeding the unchanged 120-second collector budget before module/list/build. Manager added task 0.7b.1.4: bounded parallel copy may improve latency without weakening no-follow, metadata/digest/fsync, deterministic manifest, cancellation or cleanup, but setup cannot move outside the collector clock. Actual snapshot and end-to-end preflight must fit or remain HOLD.

### M-B1 provenance repair acceptance

- The same Terra implementer captured honest REDs for the missing `goBinary` field, prescribed trimpath relocation, Go/Git swap-after-hash/restore-before-post-hash, original compiler/source swap, serial snapshot overrun and stale private-directory cleanup. GREEN requires an explicit absolute `--go-binary`; snapshots exact Go/Git bytes; descriptor-walks and manifests the selected Go distribution; uses a fixed bounded worker pool with no-follow/O_EXCL/pre-post metadata/digest/file+directory fsync; revalidates before/after commands; and removes every private tree on success/HOLD.
- The frozen implementation full tool suite passed in 197.169 seconds, with the real collector identity preflight at 64.47 seconds and hostile standalone process at 62.17 seconds inside their own 120-second contexts. Final exact-candidate short, race-short, count20, vet, Linux/Darwin/Windows trimpath builds and focused M-B0 tests passed; no private execution/GOROOT directory remained. The later cancellation/cleanup hardening full run passed in 197.169 seconds-equivalent scope, and the same-reviewer real tests independently passed under a contended host.
- The same read-only Terra adversary returned `P0=0 / P1=0 / P2=0`, ACCEPT. It verified Go/Git swap resistance, private GOROOT source/tool integrity, deterministic manifest, cancellation and cleanup, real trimpath preflight, hostile missing-cache process, exact untracked GOOS allowlist and tracked closure binding through frozen base/diff/per-input SHA.
- A separate gstack implementation review returned `P0=0 / P1=0 / P2=1`, ACCEPT. Its nonblocking P2 is deadline liveness: Git subprocesses are not context-bound, so a stalled local Git read may return HOLD after 120 seconds, but identity must complete and liveness is rechecked before any official reader admission. This does not create source, mutation or exposure authority and is retained for later hardening.
- Manager focused verification passed tool short/race-short/vet, focused M-B0, strict OpenSpec, logic-map, PM hierarchy and diff checks. Tasks 0.7b.1–0.7b.1.4 are re-accepted. Task 0.7b.2 has not been rerun; parent 0.7, Manager M-B PASS 0.7c, L1/L3–L7, source/finality authority, activation, dispatch and deployment remain HOLD.

## 2026-08-16 M-B1 authorized measurement rerun — ACCEPT-HOLD

### Single bounded run and frozen receipt

- The human explicitly authorized one read-only M-B1 rerun. The Terra collector used the required explicit Go binary, an existing current-UID 0600 cached token without OAuth exchange/write, exact session date `2026-08-14`, an explicitly empty initial cursor and a new POSIX `/tmp` receipt root. It made no order, preview, auth update/exchange, config, verify lifecycle, engine/container, install/deploy, stage, commit or push action.
- The prescribed trimpath build succeeded and the collector started exactly once. Collector attestation reports that the first and only official candle request returned HTTP 401 and exit 2, with retry, fallback, later candle pages, orderbook and calendar requests all zero. This HTTP status and exact request count are collector attestation, not independently replayable fields in the receipt; they are retained as P2 attribution rather than source authority.
- The frozen HOLD receipt is `/tmp/a112-mb-us-receipt-rerun.PQy4Em/a112-mb-us-b560a90cba0989c743890e1addc9ab10`. Parent and run are current-UID exact0700 directories. Its exact entry set is `preflight.json` and `TAINTED`, both current-UID regular exact0600. Their SHA-256 values are respectively `569bf45937fd9859081a8dd0815968d4f2af37e6b62534d13991a9232d6300fa` and `3a6e33f5ab9acbd5ca6cd34321def789cf40c24eeaa8142f398c0ac208d61574`; `TAINTED` is exactly `HOLD\n`.
- There is no manifest, raw candle/cursor/rate header, orderbook, calendar, post-request identity or success payload. Consequently terminal pagination, USD, explicit US regular-session join, raw quote, unique rate headers, redaction of raw response and source/build pre/post equality are absent and cannot support M-B PASS.

### Independent Terra and gstack evidence reviews

- The separate read-only Terra receipt adversary returned `P0=0 / P1=0 / P2=2`, `ACCEPT-HOLD`. It independently verified exact filesystem modes, entry set, payload bytes/hashes, sanitized preflight schema and source/build maps. Current workspace comparison found 0 mismatches across 25 source hashes, 74 compiled-input hashes and 217 untracked-content hashes. P2s are the receipt-unbound HTTP/request-count attestation and the absence of a sealed manifest/post identity after the first error.
- The separate read-only gstack evidence review returned `DONE_WITH_CONCERNS`, `P0=0 / P1=0 / P2=2`, `ACCEPT-HOLD`. It confirmed the one-shot non-2xx HOLD at `internal/official/a112_mbus_read.go`, immediate first-reader return at `tools/a112-mb-us-source/main.go`, fail-closed taint creation, and no production/order/config/auth/runtime authority. It also retained the existing nonblocking Git-process liveness P2: a stalled local Git command may return HOLD late but cannot reach reader admission after expiry.

### Manager verification and authority decision

- Manager independently revalidated root/run/file type, UID and exact 0700/0600 modes, the exact two-entry set, both SHA-256 values, `TAINTED` bytes, valid preflight schema, empty-cursor digest, base/current HEAD, selected Go digest and 0 mismatches across all 25/74/217 identity maps. Focused no-retry, first-error, M-B0, short tool and vet commands all passed.
- Task 0.7b.2 is complete only as a human-authorized bounded **HOLD measurement attempt** with accepted receipt integrity. It is not an operational success and does not mint M-B, source, finality, calendar, rate, activation, dispatch or deployment authority.
- Task 0.7c and parent task 0.7 remain unchecked. M-B remains **HOLD** because raw candle/quote/calendar, terminal cursor and unique rate-header evidence are missing. L1 and L3–L7 remain blocked; L2 remains inert and unaffected. No automatic rerun is authorized.

## 2026-08-16 M-B1 third authorized measurement (run 3) — ACCEPT-HOLD at the page cap, transport proven

### Root cause of the 13:10 HTTP 401 and the operational precondition it exposes

- The 13:10 401 was not a candle-endpoint or seam defect. StockOS `infra-toss-intelligence-1` (source `stockos/apps/toss-intelligence/main.go`, `officialAuth`) issues client-credentials tokens for the **same** OpenAPI credential and keeps them in process memory only, so it can never adopt the shared cache token. The broker keeps one live token per credential (a082 measurement), so each foreign issue killed the cache token; the host re-bought it roughly once a minute (soak `token_expires_at` advanced 20–54 s before every cycle start; soak cycle 14 lost `GET /api/v1/orders` to the same `authentication failed (HTTP 401)` at 13:06). The cache token was live for only about one minute of every window, which is why a no-retry collector saw 401 on its first request.
- The human stopped `infra-toss-intelligence-1` at about 15:56 KST. The cache token then stayed unchanged for more than 4.5 minutes and every subsequent official read succeeded. **New precondition for any future external M-B run**: every holder of the OpenAPI credential (StockOS `toss-intelligence`, TossOS engine/httpapi/console containers sharing `TOSSOS_CONFIG_DIR`, host `tossctl`) must be enumerated and either quiescent or sharing the cache. The durable fix is a separate OpenAPI application key for StockOS; stopping a foreign consumer is a human-authorized side effect on another product and is not an agent action. TossOS token code (`internal/official/token.go`: 60-second skew auto-exchange plus a082 adopt-before-buy) needs no change.

### Bounded runs and receipts

- The human authorized one further read-only run after auth recovery. Agent execution of the verified collector binary was hard-blocked by the session permission classifier (`/tmp` binary plus token-cache path), so the human ran it directly. The receipt root `/tmp/a112-mb-us-receipt-run3.lCsMTV` (current-UID exact 0700) contains **three** run directories: `a112-mb-us-345aed38cb863d6828200caae71ed3a5` (15:55:22 KST) and `a112-mb-us-84b898bd86c949f109fe994656a513d0` (15:56:03 KST), both human-run before the Manager's go signal, and the authorized run `a112-mb-us-9206d489c3f7a3a5ae64723d4f085b59` (16:01:28–16:02:01 KST). Total official GETs across the three runs: 12 candle, 0 orderbook, 0 calendar. All three runs are byte-identical in `preflight.json`, all four raw bodies and all four cursor files (market closed); only rate-header/meta bytes differ.
- The executable was rebuilt to the frozen SHA-256 `f2f71c3dd66ac4e78373f61a1b0a629f3fb3c6f44ca269f3e68a54e857db793f` **only** under the prescribed environment (`GOENV=off GOFLAGS= GOWORK=off GOPROXY=off GOSUMDB=off CGO_ENABLED=0`, `-mod=readonly -trimpath -buildvcs=false`); a default-environment build of the same source produced a different SHA and was correctly refused before execution.
- Authorized-run entry set: `preflight.json` (sha256 `06e98bbc45571656017516fad391bdacdee0fe5b0dd48c900e920c4ce84333b2`), `candle-01..04.{raw,cursor.raw,headers.raw,meta}.json`, `TAINTED` (`HOLD\n`, sha256 `3a6e33f5ab9acbd5ca6cd34321def789cf40c24eeaa8142f398c0ac208d61574`); every payload current-UID regular exact 0600; no `manifest.json`, `orderbook.*`, `calendar.*` or post identity. Raw body digests: `1cad7830…`, `c849cbb7…`, `3659adc1…`, `c44f1bde…`. Full hash table is in `analysis/measurements/m-b-us-source/receipt-2026-08-16-run3.json`.
- Collector result: exit 2, `a112 m-b1 measurement HOLD: candle pagination did not reach raw null within four pages` (`tools/a112-mb-us-source/main.go:170-171`).

### What the broker returned (raw bytes)

- `GET /api/v1/candles adjusted=false&count=200&interval=1m&symbol=AAPL`, HTTP 200 × 4. Page 1 `2026-08-15T09:00:00.000+09:00 → 04:47`, `nextBefore 04:46`; page 2 `04:46 → 01:27`, `nextBefore 01:26`; page 3 `01:26 → 2026-08-14T22:07`, `nextBefore 22:06`; page 4 `22:06 → 18:33`, `nextBefore 2026-08-14T18:32:00.000+09:00` (non-null). Cursor file bytes equal raw `result.nextBefore` on every page; page N+1's first bar equals page N's `nextBefore` (inclusive cursor, 3/3); 800 unique strictly descending timestamps; `before` percent-encoded exactly once (`%3A`, `%2B`).
- All OHLCV values are JSON decimal strings; `currency` is `USD` in 800/800; timestamps carry the `+09:00` offset with `.000` millis. **The 2026-08-14 US regular session (22:30 KST 08-14 → 04:59 KST 08-15) is fully covered: 390 of 390 one-minute bars, zero gaps.** All 52 non-one-minute gaps lie in extended hours (untraded minutes are omitted, not zero-filled); 90 post-close bars (07:00–09:00 KST) fall outside every documented calendar session — an L1 calendar-join decision, not a source defect.
- Rate headers: `X-Ratelimit-Limit 20`, `X-Ratelimit-Remaining 19/19/18/17`, `X-Ratelimit-Reset 1`, `Retry-After` absent — cardinality one per response; identical bytes on pages 1–2 are a one-second refill, bound per page by `meta.json` body digests.

### Why HOLD — a design assumption, not a tool or broker defect

- The OpenAPI contract defines `nextBefore null` as the last page of retained history and an absent `before` as "start from the latest bar". From an **empty** initial cursor, terminal null therefore requires walking AAPL 1-minute history to its beginning, which 4 × 200 bars cannot reach; the run passed after-hours → regular session → pre-market and stopped 800 bars back at 05:33 EDT with the cursor still non-null. Whether the endpoint terminates per trading day or serves unbounded history is unobserved. Under either hypothesis, "raw null within four pages from an empty/session-end cursor" cannot coexist with "full regular-session coverage" in one receipt.
- The assumption is encoded at `tasks.md` 0.7b (`exactly four candle pages maximum … HOLD on … cap exhaustion`), 0.7b.2 (`terminal null within the cap`), 0.7c (`page-cap HOLD … keeps task 0.7 unchecked`), `design.md` M-B1 paragraph (`Candle은 최대 4페이지 … Terminal은 raw JSON null만 허용`), the spec M-B1 paragraph and both scenarios, `review.md` 2026-08-15 lines on the finite cap/HOLD rule, and `main.go:134,170-171`. `--session-date` feeds only the calendar GET, never the candle crawl.

### Independent reviews and Manager verification

- The separate read-only receipt adversary returned `P0=0 / P1=0 / P2=6`, `ACCEPT-HOLD`. It independently reproduced the executable SHA under the prescribed environment, the Go-distribution manifest digest over 16,692 entries, `go_toolchain_sha256`, `git_binary_sha256`, `tracked_diff_sha256`, `untracked_sha256`, `frozen_input_sha256`, and found 0 mismatches across 25/74/217 identity maps; every per-page body/cursor/header digest in `meta.json` recomputed equal; strict JSON, no duplicate keys, no secret-like values (only path names in `preflight.json`); modes exact; write order and ctime/mtime consistent with `writeResult` and deferred taint. P2s: three invocations/12 GETs to log; structural cap (empty-cursor reruns are pointless); stale `analysis/measurements` (now refreshed by `receipt-2026-08-16-run3.json`); rate unit unstated vs the older "5 requests per second" note; exit code/stderr attestation-only; no session join/quote/seal.
- The separate read-only gstack evidence review returned `DONE_WITH_CONCERNS`, `P0=0 / P1=1 / P2=6`, `ACCEPT-HOLD`. Its P1 is the design criterion above (blocks PASS, not the receipt). Its evidence table: raw body, USD, decimal preservation, cursor decode/continuity, allow-listed unique rate headers, redaction and exact modes **DEMONSTRATED**; one-GET-per-call/no-retry PARTIAL (structural plus receipts, count not receipt-bound); terminal null, orderbook, calendar/session join, post identity and self-excluding manifest **ABSENT**.
- Manager independently replayed root/run/file types, UID and exact 0700/0600 modes, the 18-entry set, `TAINTED` bytes, every SHA-256 above, preflight config equality with the frozen 13:10 receipt (identity equal except the Manager's own `review.md`/`tasks.md` untracked edits), cursor continuity, per-page monotonicity, the 390/390 regular-session join against the documented KST window, USD/decimal-string types and the rate-header sequence.
- Task 0.7b.2 remains complete only as human-authorized bounded **HOLD measurement attempts**; the M-B0 seam is now proven against production for raw transport, USD, decimals, cursor and rate semantics. **Task 0.7c and parent 0.7 remain unchecked**; M-B remains **HOLD** (terminal cursor, orderbook, calendar join, post identity and manifest absent). L1 and L3–L7 remain blocked; L2 inert.

### Manager amendment proposal (requires human approval before any edit or rerun)

1. Keep the 4-page cap as a hard request bound. Redefine the candle PASS criterion as **explicit regular-session full coverage** for `--session-date`: recommended initial `--before` = the calendar's `regularMarket` end (e.g. `2026-08-15T05:00:00.000+09:00`) or empty; at least one pre-session bar observed proving the crawl crossed session start; 390 bars, no gaps, USD, decimal strings, cursor continuity. Move terminal-null observation to either (a) an optional separate explicit-cursor terminal probe (`--before` far in the past, expect `candles: []` and raw `null`, ≤1 GET) or (b) "OpenAPI-documented; the L1 strict decoder fails closed on anything but string/null". Consequence for the tool: cap exhaustion is recorded rather than HOLD so one receipt can carry candles + orderbook + calendar + seal — via normal Terra RED→GREEN→adversarial→gstack, no production code.
2. Record the credential-holder precondition (above) in `design.md` M-B and as a pre-run checklist line under 0.7b.2.
3. Optional cheap discovery before amending: one human-authorized bounded run with `--before=2026-08-14T18:32:00.000+09:00` (page-4 cursor) to learn whether null appears at a trading-day boundary; ≤4 GETs.
4. Human authorization is needed for any external run (≤6 GETs), the credential-holder check/stop, and the OpenSpec amendment. No automatic rerun. Extra human invocations should be avoided: empty-cursor reruns reproduce the identical HOLD.

### 2026-08-16 amendment approval and lot opening

- The human approved amendment option 1 (regular-session full coverage as the candle PASS criterion; terminal null recorded when seen but no longer a precondition; cap exhaustion recorded, not HOLD; credential-holder pre-run checklist). Applied to `design.md` (M-B1 paragraph + operational precondition), `specs/breakout-retest-strategy-lane/spec.md` (requirement text; scenarios "bounded pagination이 cap을 소진함", "cursor loop 또는 malformed cursor", "regular session이 완전히 덮이지 않음", amended PASS scenario), and `tasks.md` (0.7b.2 wording, new unchecked 0.7b.3 amendment lot and 0.7b.4 pre-run checklist, 0.7c wording). `openspec validate --strict` and the PM tracker check pass.
- Function Logic Map produced before the lot opened: `analysis/function-logic/tools-a112-mb-us-source--run/` (AST 36 branches, source SHA `2aa28733…` = the run-3 executable's source; B26 @ main.go:170 is the only branch the amendment changes; `check_analysis.py` reports evidence complete). Note that `run` is not in the frozen base, so the gate does not require this map; it exists because this review cites the branch as evidence.
- Lot 0.7b.3 ownership: a Terra implementer edits only `tools/a112-mb-us-source/**` (RED first), a different Terra adversary reviews read-only, then gstack, then Manager acceptance. Existing gap recorded for honesty: `main_test.go` had **no test** for the four-non-null-pages HOLD branch (B26) before this lot — the branch was first exercised live.

### 2026-08-16 lot 0.7b.3 — Terra RED/GREEN, adversary and gstack reviews

- Terra implementer (Opus) edited only `tools/a112-mb-us-source/main.go` and `main_test.go`. RED captured against the pre-amendment source (SHA `2aa28733…`): `TestRunRecordsCapExhaustionAndStillCollectsOrderbookCalendarAndSeals` failed with "cap exhaustion must not HOLD: … candle pagination did not reach raw null within four pages", `TestRunRecordsNullTerminalInCandleCrawlRecord` failed with "candle-crawl.json: no such file". GREEN: `pages` counter, `candle-crawl.json` (`schema a112-mb-us-candle-crawl:v1`, `pages`, `terminal` ∈ {`null`,`cap_exhausted`}, `last_cursor_sha256` = `sha256:` of the decoded page-4 cursor value bytes, `omitempty` on null) written through the live-guarded `store.writeJSON` after the loop, one extra `live()`, then orderbook/calendar/post identity/seal unchanged. Post-amendment `main.go` SHA `1945343d39b89e45ccc0fbe0accca9a3e3d5a157cab9e5a2ef838d3e8caecac9`; the gstack reviewer reconstructed the pre-amendment file from the current one by reverting exactly those five edits and reproduced `2aa28733…` byte-for-byte, so the lot's `main.go` change is provably only those edits. Verification: `go test ./tools/a112-mb-us-source -short` ok, full suite ok (125–133 s), `-race` ok, `go vet` clean, `gofmt -l` empty, `go test ./internal/official -run TestA112MBUS` ok (static guard undisturbed). Historical anchors elsewhere in this file (`main.go:170-171`, "AST 36 branches") describe the run-3 executable; post-lot the branch is `main.go:175-184` and the FLM AST has 38 branches.
- Manager regenerated the FLM (`analysis/function-logic/tools-a112-mb-us-source--run/`, AST 38 branches, SHA `1945343d…`) after the edit and re-based the maps; `check_analysis.py` reports evidence complete.
- Separate Terra adversary (read-only, `go test -overlay` mutations, repo untouched): proved no code path can issue a fifth candle request (single `Candle` call site inside the 4-bound loop; bound→5 mutant issued a real 5th request and the exact ordered call-list assertion killed it), all pre-existing HOLD paths preserved anchor-by-anchor, receipt record deterministic and manifest-bound, `writeJSON` path acceptable for a tool-authored constant record. Mutation table: 7/9 killed (bound 4→5, terminal mis-set both directions, record outside manifest, pages off-by-one, digest of `cfg.before`, record before loop); survivors: `sha256→md5` (no test pins the algorithm) and deleting the new post-crawl `live()` (redundant with the store's own post-write guard). Verdict HOLD on one **P1: `tasks.md` 0.7b still listed "cap exhaustion" among HOLD causes** — Manager struck it and annotated 0.7b (same pass: FLM inputs-table anchors re-based, branch-test-map B25/B28 relabelled `not-applicable` with the reason instead of claiming tests, `candle-crawl.json` added to the spec sentence and to the 0.7b.4 reviewer checklist). Remaining P2s: page-4 rate gate still HOLDs on remaining < 1 (accepted as the conservative reading of "rate gate 실패는 변함없이 HOLD"; recorded in the FLM), digest algorithm unpinned, B25 failing side untested, `analysis/measurements` manifest does not cover `receipt-2026-08-16-run3.json`, Korean comments acceptable per CLAUDE.md rule 8.
- Separate gstack review: `DONE_WITH_CONCERNS`, `P0=0 / P1=0 / P2=6` — same test-adequacy gaps (crawl-write failure taint untested, page-4/any rate-gate failing side untested, `omitempty` shape unpinned, `last_cursor_sha256` not cross-checked against `candle-04.meta.json.cursor_value_sha256`), plus a note that `cli.go` `--before` help still says "may be empty" while 0.7b.4 forbids empty-cursor reruns as a human-process rule.
- Manager decision: implementation accepted in substance; **acceptance of 0.7b.3 is withheld** until the implementer adds the regression tests both reviewers asked for (crawl-write failure → HOLD+taint; page-4 and earlier-page rate-gate failure; `digestBytes` literal SHA-256 vector and cross-check with `candle-04.meta.json`; pages-before-null/`omitempty` pin) and the same adversary rechecks. No production, `cmd/tossctl`, `internal/official` or runtime file was touched; the collector remains uninstalled.
- Implementer follow-up (main.go byte-identical, SHA `1945343d…` re-verified; `main_test.go` only): added `TestRunHoldsAndTaintsWhenCandleCrawlRecordWriteFails`, `TestRunHoldsWhenPageFourRateBudgetIsExhausted`, `TestRunHoldsWhenEarlierPageRateBudgetIsExhausted`, `TestDigestBytesPinsSHA256Vector` (literal vector `sha256:0012a3fa…675b38aa` for `c4`), `TestRunRecordsPagesBeforeNullTerminal` (pages=2, `omitempty` pinned on raw bytes) and the `last_cursor_sha256 == candle-04.meta.json.cursor_value_sha256` cross-check inside the cap test. Each is a regression guard that passes on the accepted code; the implementer proved each kills its named mutant (swallowed write error, skip-page-4 gate, gate removed, digest input change, raw-JSON-instead-of-decoded cursor). Full suite ok (123 s), `-short` ok, `-race` ok, vet/gofmt clean.
- Same adversary recheck: **ACCEPT, P0=0 / P1=0 / P2=2** — re-ran the mutation set with `-overlay`: 11 of 12 mutants now die (`sha256→md5`, constant digest, skip page-4 gate, remove gate, swallow write error, drop `omitempty` all killed by the new tests); the single survivor (deleting the post-crawl `live()`) is unobservable through `run()` because the store's own post-write guard fires first, and the map records it as `not-applicable` with that reason. P1-1 and prior P2-1/2/3/4/5/7 confirmed fixed. Remaining P2-a (FLM "Required test" cells for B25/B27/B28 out of sync with the branch-test map) and P2-b (`analysis/measurements/m-b-us-source/manifest.sha256` did not cover the two files added today) — both fixed by Manager in this pass (cells synced; manifest now pins 11/11 files, `sha256sum -c` clean).
- **Manager acceptance: lot 0.7b.3 ACCEPTED, task 0.7b.3 checked.** The amended collector records cap exhaustion instead of HOLDing and reaches orderbook/calendar/seal; the frozen executable identity for the next authorized run must be re-derived from the amended source under the prescribed environment before that run (its SHA will differ from `f2f71c3d…`). Task 0.7b.4 (pre-run checklist), 0.7c and parent 0.7 remain unchecked; M-B remains HOLD; L1/L3–L7 blocked; nothing staged or committed.

### 2026-08-16 18:49 KST — amended executable identity frozen for the next authorized run (0.7b.4 preparation, no run)

- Source under freeze: `tools/a112-mb-us-source/main.go` SHA-256 `1945343d39b89e45ccc0fbe0accca9a3e3d5a157cab9e5a2ef838d3e8caecac9` (the accepted lot 0.7b.3 file, unchanged), worktree HEAD `016da624`, selected Go `/home/daniel/Downloads/go1.26.5.linux-amd64/go/bin/go` (`go1.26.5`, binary SHA-256 `8da5fd321795754b994c64e3eb8a5a14ff47bd285559a7e876f3c79abafc67f9`, equal to run 3's `go_toolchain_sha256`).
- Built **twice** with independent fresh private `GOCACHE`s under exactly the environment `identity.go` prescribes (`env -i HOME=/nonexistent GOROOT=<distribution> GOENV=off GOFLAGS= GOWORK=off GOPROXY=off GOSUMDB=off CGO_ENABLED=0 GOCACHE=<fresh> GOMODCACHE=/home/daniel/go/pkg/mod`, command `go build -mod=readonly -trimpath -buildvcs=false -o … ./tools/a112-mb-us-source`, `Dir` = repo root). Both builds produced the same bytes: **executable SHA-256 `232c787c50685c623f6bfade2d0713850e295ff32bb83834d9f0d324fb3137ce`** — this is the frozen identity for the next authorized run; it differs from run 3's `f2f71c3d…` exactly as the acceptance note predicted. The pinned copy is `/tmp/a112-mb-us-build-amended.OSVxn5/a112-mb-us-source-pinned` (owner-only 0700 directory, 0700 file); an empty owner-only receipt root `/tmp/a112-mb-us-receipt-run4.ZTagql` (0700) was created for the run. Nothing was executed: the session permission classifier still blocks agent execution of the `/tmp` collector, even a no-argument smoke check, so the first execution will be the human's authorized run.
- Read-only credential-holder pre-check at 18:49 KST (task 0.7b.4 item 1, to be re-confirmed by the human immediately before the run): `infra-toss-intelligence-1` **Exited (2)** for ~3 h; `tossos-tossos-1` and `tossos-httpapi-1` Up (healthy) sharing the token cache; the cache file's mtime has been unchanged since 15:55:42 KST (~2 h 54 min without a re-mint, so no foreign holder is minting). Metadata only; the cache contents were not read.
- Prepared command (one invocation per authorization; `--before` = the 2026-08-14 `regularMarket` end so the crawl starts at session close and crosses session start within the 4-page cap; empty-cursor reruns remain unauthorized): `<pinned> --token-cache=/home/daniel/.config/tossctl/openapi-token.json --go-binary=/home/daniel/Downloads/go1.26.5.linux-amd64/go/bin/go --receipt-root=/tmp/a112-mb-us-receipt-run4.ZTagql --session-date=2026-08-14 --before=2026-08-15T05:00:00.000+09:00`. Human-side verification before invoking: `sha256sum` of the pinned file must print `232c787c…`; do not edit any repo file between invoking and exit (the pre/post identity re-hash covers the untracked content set, and a mid-run edit would HOLD the run as drift).

## 2026-08-16 M-B1 fourth authorized measurement (run 4) — sealed receipt, Manager replay

### Invocations and honesty note

- The human's pre-run check (pasted): `infra-toss-intelligence-1` Exited (2), `tossos-tossos-1`/`tossos-httpapi-1` Up (healthy), token-cache mtime still `15:55:42` KST, pinned executable SHA `232c787c…` confirmed. Then the human invoked the prepared command. The receipt root `/tmp/a112-mb-us-receipt-run4.ZTagql` contains **two** run directories, i.e. two human invocations (the stderr/exit lines were not pasted at the time of this entry; requested from the human):
  - **Run A** `a112-mb-us-9b318b1f04d3e420d850ff42c410b34d`: created 19:01:48.66 KST; `preflight.json` 19:02:49.98 (pre-identity + prescribed rebuild took ~61 s, versus ~30 s in run 3); candle 01–04, `candle-crawl.json`, orderbook, calendar all written 19:02:50–19:02:52; **`TAINTED` (`HOLD\n`) at 19:03:35.13** — i.e. HOLD during the post-request identity/seal phase (`main.go:196-217`), no `manifest.json`. Its `preflight.json` and all six raw bodies are byte-identical to run B's, and its pre-identity equals run B's pre- and post-identity, so the HOLD was not a persistent identity drift; the exact reason (transient snapshot error, deadline, or an operator interrupt) awaits the pasted stderr. Not used as evidence.
  - **Run B** `a112-mb-us-dbe13033da92e741f6101f98893d6951`: created 19:04:24.50; `preflight.json` 19:05:23.54; collection 19:05:23–19:05:25; **`manifest.json` sealed 19:05:59.59**, no `TAINTED`. Total elapsed ~95 s of the 120 s budget (identity phases dominate: ~59 s pre, ~34 s post). Twelve official GETs across the two runs (2 × [4 candle + 1 orderbook + 1 calendar]).
- Rule reminder recorded for the ledger: authorization covered one invocation; the second was the human's own decision after the first HOLD and is recorded as such — no agent executed or re-ran anything (the classifier still blocks agent execution of the `/tmp` binary).

### Manager read-only replay of run B (script in session scratchpad; no raw body or token material printed)

- Modes: root and run directories current-UID exact 0700; all 25 entries current-UID regular exact 0600; no `TAINTED`.
- `manifest.json` (`schema a112-mb-us-receipt:v1`, sha256 `765013e4d678dd94a6a1405627140543f2219b67d3e74b4b16d8c29149e0b563`): 24 entries, self-excluding, 0 missing / 0 extra / 0 hash mismatches against the files on disk; embedded post identity **equals** the `preflight.json` identity (no drift). `preflight.json` sha256 `51567351e20a75d7031c10f399f98036321d337921be59c012b1ae90bdf6aafe`; identity `executable_sha256 232c787c…` (= frozen), `go_toolchain_sha256 8da5fd32…`, base/HEAD `016da624…`, 223 untracked entries with 0 content drift versus the current worktree; config `session_date 2026-08-14`, `before_sha256 d3b7bd38…` = sha256 of the literal `2026-08-15T05:00:00.000+09:00`.
- `candle-crawl.json` (sha256 `000d4980…`, 172 bytes): `{"schema":"a112-mb-us-candle-crawl:v1","pages":4,"terminal":"cap_exhausted","last_cursor_sha256":"sha256:149d0507…"}`; `pages` equals the four `candle-NN.raw.json` files; `last_cursor_sha256` **equals** `candle-04.meta.json.cursor_value_sha256`; page-4 cursor non-null so `cap_exhausted` is consistent.
- Per-page `meta.json`: status 200 × 6; canonical queries `adjusted=false&before=<cursor percent-encoded once>&count=200&interval=1m&symbol=AAPL` (page 1 carries the explicit `--before`), `GET /api/v1/orderbook symbol=AAPL`, `GET /api/v1/market-calendar/US date=2026-08-14`; every `body_sha256`, `cursor_json_sha256`, `cursor_value_sha256` recomputed equal to the raw files (rate-header digests are per-header maps, left to the reviewers).
- Raw body digests: candle-01 `de0f87d3…`, 02 `ff8f7f8b…`, 03 `65ab23ba…`, 04 `f8bd3e31…`; orderbook `fec9d6e4…`; calendar `0b62946b…`.
- Candles: 200 bars per page, 800 unique strictly-descending timestamps `2026-08-15T05:00 → 2026-08-14T15:08 KST`; cursor file bytes equal raw `nextBefore` on every page; page N+1 first bar equals page N `nextBefore` (inclusive cursor, 3/3); all OHLCV fields JSON strings; `currency USD` 800/800; keys `timestamp/openPrice/highPrice/lowPrice/closePrice/volume/currency`.
- **Regular-session join with the calendar body** (`result.today.regularMarket 2026-08-14T22:30:00.000+09:00 → 2026-08-15T05:00:00.000+09:00`, end exclusive): **390 of 390 one-minute bars present, 0 missing**; 409 bars observed before session start (309 inside `preMarket 17:00–22:30`, 100 inside `dayMarket` before 17:00, earliest 15:08 KST), proving the crawl crossed session start; exactly one bar at/after the regular end (the `05:00` bar = the inclusive `--before` cursor itself). Calendar also carries `previousBusinessDay 2026-08-13` and `nextBusinessDay 2026-08-17` with the same four-session structure.
- Orderbook: HTTP 200, `{"result":{"timestamp":"2026-08-14T16:59:58.000+09:00","currency":"USD","asks":[],"bids":[]}}` — an empty book at 16:59:58 KST 08-14 (market closed; last quote is from the previous day-market end). Rate headers (cardinality one, allow-listed): candles `Limit 20 / Remaining 19,19,18,19 / Reset 1`, orderbook `Limit 15 / Remaining 14 / Reset 1`, calendar `Limit 3 / Remaining 2 / Reset 1`, `Retry-After` absent everywhere.
- Manager finding: every 0.7b.2 item and every amended 0.7b.4 reviewer item is satisfied on replay. Terminal null was not observed (recorded, not required). Independent reviews follow before any PASS.

### Independent reviews of run B (both read-only; neither touched the repo, the receipt, the token cache or the network)

- **Terra receipt adversary: `ACCEPT-WITH-P2`, `P0=0 / P1=0 / P2=7`.** Independently recomputed and confirmed: root/run 0700 and 25/25 payloads 0600 current-UID nlink 1 (both runs); manifest `765013e4…` 24 sorted self-excluding entries 0/0/0; manifest identity byte-equal to preflight; run A preflight byte-identical to run B; `executable_sha256 232c787c…` == pinned == its **own fresh prescribed rebuild** (`env -i` exact `selectedGoEnvironmentWithRoot`, byte-identical, build dir deleted, never executed); `go_toolchain 8da5fd32…`, `git_binary 2a8c18fb…`, `go_distribution_sha256 2f398a29…` recomputed over 16,692 entries, `frozen_input f077bc71…`, `untracked 7e37cc24…`, `tracked_diff 1fdae7ae…`, HEAD/base `016da624…`; identity vs current worktree 25/74/223 → 0/0/1 mismatch (only `review.md`, Manager edits; `tasks.md` unchanged since the run; untracked set 223/223 identical; every untracked compiled input in the allowlist); `before_sha256` == sha256 of the literal cursor and page-1 query carries it percent-encoded once; all 6 meta digests incl. all 24 per-header `rate_header_sha256` (NUL-joined values) equal; cursor file == raw `nextBefore` 4/4, page N+1 `before` == page N decoded cursor 3/3, first bar == prior cursor 3/3, cursor == last bar − 1 min 4/4, 5 distinct cursors (no loop); `candle-crawl.json` `000d4980…` pages 4 / `cap_exhausted` / `last_cursor_sha256 149d0507…` == `candle-04.meta.cursor_value_sha256` == sha256(`2026-08-14T15:07:00.000+09:00`); 800 unique strictly descending bars; all 7 keys strings, prices `^\d+(\.\d+)?$`, volume `^\d+$`, 0 bare numbers, 0 OHLC-inconsistent bars, USD 800/800; calendar `today 2026-08-14`, join uses `today`; **390/390 end-exclusive, 0 missing, also 390/390 under bar-end labelling**; 409 pre-session bars, 27 non-1-min steps all before 22:30; exactly 1 bar ≥ 05:00; strict JSON, no duplicate keys, no invalid UTF-8, no secret-like values (only path names in preflight/manifest); write order/ctimes consistent with the code path in both runs; run B timing 95.1 s with every request started ≥ 59 s before the deadline; `go test ./tools/a112-mb-us-source -short -count=1` ok; FLM `ast.json` source SHA == current `main.go`. Findings (all P2): (1) two invocations under a one-invocation authorization, 12 GETs, run B is the human's own rerun after run A's HOLD — 0.7c must carry the human's second invocation as its own authorization line plus both stderr/exit lines (pending); (2) bar-label convention unproven — volume signature (22:29 → 2,065; 22:30 → 2,215; 22:31 → 37,733; 04:59 → 32,707; 05:00 → 62,729) points to bar-END labelling; coverage is convention-agnostic (22:30..05:00 inclusive = 391 consecutive bars all present), but the "409 pre-session / 1 post-session" statement above holds only under bar-start (410 / 0 under bar-end) — record the join as convention-agnostic and hand the convention to L1 (`opening_range_first_bar_id`); (3) orderbook is an empty book with timestamp 16:59:58 (2 s before `dayMarket.endTime`, before the same date's regular session) — proves endpoint/USD/shape only, level element schema unobserved, no evidence the endpoint carries a US regular-session book — record as an absence state, L1 must not infer quote availability/typing; (4) schema noise: orderbook/calendar `meta.json` carry `cursor_json_sha256`/`cursor_value_sha256` = sha256("") because `digestBytes(nil)` is non-empty so `omitempty` never fires; `headers.raw.json` serialises absent `Retry-After` as `null` — code-consistent, document it (done in `receipt-schema-candle-crawl.md`); (5) request count structural not receipt-bound (carried); rate quota shared with live TossOS containers (run A orderbook Remaining 13/15); `X-Ratelimit-Reset` unit unstated; (6) repo receipt for run 4 missing and `tasks.md:37` "(no run yet)" text stale (now fixed: `receipt-2026-08-16-run4.json`, manifest 12/12); (7) run A HOLD cause unrecorded. Run A analysis: elapsed 106.5 s excludes the 120 s deadline; all 24 payloads equal run B's except the four rate-header-bearing files; TAINTED's inode reuse pattern indicates the normal defer chain (graceful HOLD, not SIGKILL); fault ≈ 19:03:30–34; plausible: `post-request identity: <cause>` (transient go/git/copy I/O or SIGINT/SIGTERM ctx cancel), `identity drift during measurement` (a transient worktree file present between 19:02:50 and ≈19:03:33 and gone by 19:05:23), or `before/after post-request identity: context canceled`; `seal`/`verifySealed` unlikely (run dir clean). Not verified: broker side / exact GET count (structural), the human's pre-check (paste only), token cache (never opened), runtime Go/Git private-snapshot behaviour, `go list` closure re-run, run A stderr, live-session endpoint behaviour, bar-label convention (signature, not proof).
- **gstack evidence review: `SHIP-IT`, `P0=0 / P1=0 / P2=7`; recommends the Manager MAY record M-B PASS on run B provided the PASS entry carries the residuals.** Independent recomputation agrees with the adversary on every hash/count above (manifest 0/0/0, identity JSON-canonical equal, executable/toolchain/git binary/base/`frozen_input`/`untracked`/`tracked_diff` equal, 25 source + 74 compiled inputs hash-equal to the worktree, untracked compiled inputs == 13/13 linux allowlist, closure has no wts/hybrid/cmd package, `before_sha256`, all meta digests incl. rate-header maps, cursors 4/4 · 3/3 · 3/3, `candle-crawl.json`, 390/390 end-exclusive with 0 non-1-minute gaps inside the window, 409 pre-session bars, 3,200 price fields all strings (3,180 with 1–4 dp, 20 integer-form), 800 integer-string volumes, USD 800/800 + orderbook). Evidence table: DEMONSTRATED — official-only source (seam calls only `official.A112MBUS{Candle,Orderbook,Calendar}` under `authorityOriginLocked()` requiring `https://openapi.tossinvest.com`, paths `/api/v1/candles`, `/api/v1/orderbook`, `/api/v1/market-calendar/US`); one GET per call / no retry / ≤4 pages / 6 GETs (single `hc.Do` on a per-call cloned transport, loop `page < 4`, two `collectSingle`, sealed success ⇒ one meta per request, `Remaining` sequence consistent, `-short` suite and `TestA112MBUS` ok offline); raw body hash-bound; USD; decimal preservation; cursor decode/continuity/no loop (`--before` seeded in `seen`, `main.go:130`); bounded cap recorded and sealed; regular-session full coverage joined with the calendar body; calendar/session join (`date=` == `--session-date` == body `today.date`); rate headers allow-listed, cardinality one, `Retry-After` null × 6; redaction (raw bodies pass key/value/bearer/JWT scans; only heuristic hits are identity path names); exact 0700/0600 uid 1000 on ext4; pre/post identity equal and build identity == running executable == frozen SHA; self-excluding manifest; no production/order/config/auth/runtime authority touched (pre == post identity, tracked diff = 5 pre-existing docs/pm files, token cache mtime unchanged, no installed binary, seam has no token/exchange/write path, no leftover private `/tmp` dirs). PARTIAL — orderbook (empty book: level field names/decimal encoding unobserved); credential-holder precondition (human paste covers StockOS + engine/httpapi; console container / host `tossctl` not explicitly enumerated; corroborated by cache mtime still `15:55:42` through both runs and 12/12 HTTP 200). ABSENT-recorded — terminal null (`cap_exhausted`, last cursor `15:07` non-null). P2s: (1) two invocations/12 GETs, run A stderr not captured, human ratification requested; (2) empty orderbook — L1 cannot cite this receipt for level shape; (3) bar timestamp labelling undecidable from the receipt (05:00 bar sits at `afterMarket` start) — finality/labelling stays L1 (`design.md` l.168); (4) variable decimal scale — L1 strict decoder must accept it; (5) shared rate quota — a market-hours run could trip the strict page-4 gate (FLM B25); (6) `writeJSON` payloads carry host paths by design — sanitized `analysis/**` must not mirror them; (7) no run-4 repo receipt yet, `manifest.sha256` at 11 files, `cli.go:18` help text carried. Not verified: `go_distribution_sha256` (relied on adversary/receipt), prescribed rebuild not re-executed, docker state (forbidden — relied on the paste + cache mtime), run A cause, linker-level dead code of `internal/official/orders_write.go` in the package closure (unreachable from the tool's three call sites), full/`-race` suite, invoked pathname (identity bound by SHA regardless).

### Manager decision — M-B PASS (2026-08-16 evening)

- Both independent reviewers returned P0 = 0 / P1 = 0 on the sealed run-B receipt and both recomputed every value the Manager replayed. Under the human-approved 2026-08-16 amendment, the candle PASS criterion is explicit regular-session full coverage joined with the calendar body (**390/390, 0 missing, convention-agnostic**, ≥ 1 pre-session bar observed), with cap exhaustion recorded in the sealed `candle-crawl.json` and terminal null recorded-not-required; every 0.7b.2 item (raw body, USD, decimal preservation, cursor decode/continuity/no-loop, allow-listed unique rate headers, redaction, exact modes, pre/post identity, prescribed build identity == running executable == frozen `232c787c…`, self-excluding manifest) is DEMONSTRATED. No missing/ambiguous field, no coverage gap, no cursor loop/malformed HOLD, no rate-header cardinality ambiguity, no unofficial/WTS/float/time-derived evidence, no unsafe mode. **The Manager records M-B PASS on `/tmp/a112-mb-us-receipt-run4.ZTagql/a112-mb-us-dbe13033da92e741f6101f98893d6951` (manifest `765013e4…`).** Sanitized repo record: `analysis/measurements/m-b-us-source/receipt-2026-08-16-run4.json` (supersedes `receipt.json` 2026-08-15 and `receipt-2026-08-16-run3.json`; `manifest.sha256` now 12/12).
- Scope of the PASS (unchanged from 0.7c): it authorizes **only L1 implementation start**. It does not accept L1, prove candle publisher finality, closed-bar authority or the bar-label convention, prove quote level schema, activate a lane, dispatch, deploy or raise exposure.
- Residuals carried into L1's inputs and this ledger (from the two reviews): (a) orderbook empty book — L1 must not infer quote level schema/typing/availability from this receipt; a live-session quote observation is a separate, human-authorized measurement if L1 needs it; (b) bar-label convention undecided (volume signature suggests bar-end) — L1 decides `opening_range_first_bar_id` explicitly and proves finality separately; (c) variable decimal scale (1–4 dp, integer-form prices) — L1 strict decoder must accept variable scale without float conversion; (d) rate quota shared with live TossOS containers and `X-Ratelimit-Reset` unit unstated — L1 treats Reset as opaque and budgets conservatively; (e) two human invocations under a one-invocation authorization: recorded honestly; **the human is asked to paste both stderr/exit lines** (run A's `a112 m-b1 measurement HOLD: <reason>` and run B's `A112 M-B1 receipt sealed …` / `exit=0`); they are appended here on receipt and do not change the PASS (run B's success is proven by the sealed manifest and `verifySealed`); (f) `cli.go` `--before` help text ("may be empty") vs the human-process rule — cosmetic, left for a later tool lot; (g) credential-holder enumeration should name the console container / host `tossctl` explicitly next time (none was running; the human shell is the only host holder).
- Task ledger: 0.7b.4 checked (pre-run checklist executed by the human, amended collector at its frozen SHA, explicit `--before` at the `regularMarket` end, two independent reviewers; the one-invocation deviation is recorded above), 0.7c checked (this record), parent 0.7 checked. **L1 may start**; L3–L7 remain gated by their own tasks; L2 inert. Nothing staged or committed; the worktree stays intentionally dirty and unmerged pending base re-freeze at landing.

## 2026-08-16 L1 lot opening — official source / evidence authority (Manager brief)

### Inputs (Full SDD order)

- Memory recall: token-war precondition, "FLM before claiming", "passing test is not evidence", "correction unit is the value" (session memory). OpenSpec: L1 row `tasks.md:10`, tasks 2.5 / 3.1 / 3.2 / 3.6 / 3.7 / 3.8, `design.md` §5 (incl. amendment, l.166 clock rule, l.168 finality), spec requirement "breakout evidence는 strict append-only snapshot으로 재생 가능하다", golden `analysis/goldens/breakout-evidence-and-sizing-v1.json` (`setup_identity`, `evidence_schema`, `official_source_identities` — consumed hash-identical, never rewritten).
- CodeGraph hard evidence + AST: `analysis/l1-evidence-pack.md` (Manager assistant, read-only, codegraph live 33,853 nodes, every row re-confirmed by Read/grep). Existing FLMs `analysis/function-logic/internal-strategyevidence--kindsupportsmarket/` and `--validatetypedpayload/` are **CURRENT** (`ast.json source_sha256 c49652af…` == current `model.go`); those are the only two existing function bodies L1 edits.

### Manager decisions on the pack's open questions (binding for the lot)

1. **Lot split.** L1a = pure evidence contract in `internal/strategyevidence` (tasks 2.5, 3.1, 3.2; offline; no `official` import). L1b = official raw KR/US bar + quote producer (task 3.6, part of 3.7/3.8; High-risk adjacency to the official client GET/token path → Pre-Edit declaration, additive files only). L1b starts only after L1a is accepted.
2. **Where the official adapters live (Q1).** New files under `internal/official/` (that is what "additive read-only official … adapter files" in the L1 row means; `RawMinuteCandles` lives there). They may not reference any `a112MBUS*` identifier (M-B0 guard `a112_mbus_static_test.go:476-498`), must not edit `candle_raw.go`, `client.go`, `token.go`, `trace.go`, `ratebudget.go`, and use the ordinary production token path (`c.get`/`send`, a082 adopt-before-buy) — cache-only reads are a measurement property, not a production one (Q7). KR keeps `official.Client.RawMinuteCandles` as its proven identity (golden `official_source_identities.KR`); US gets its own strict raw reader with the M-B-proven contract (`analysis/l1-evidence-pack.md` §4). Strict post-decode validation for both markets lives in `strategyevidence` (L1a).
3. **Storage shape (Q2).** One envelope per closed 1-minute bar (design bar identity `(market, symbol, session_id, interval, open_at)` → `SourceRecordID`; correction = higher `RevisionIdentity`, `SupersedesRevisionIdentity` = prior). No edit to `SnapshotQuery`/`SealSnapshot`/`snapshotDigest`; L1a adds a **new** kind-scoped bounded ordered read (`BarSeriesQuery` → `BarSeries`, ≤ 512 bars, latest visible revision per identity under the dual cutoff, own digest). Bar/quote evidence is written to a **dedicated store file** (e.g. `<journal>/breakout-bars.db`), never to the existing `evidence.db`, so the legacy symbol snapshot (`choose`, no kind filter) is unaffected — recorded here as an L5/L6 wiring constraint; L1 wires nothing.
4. **Issuer scope (Q3).** `IssuerIdentity = "<MARKET>:<SYMBOL>"` (uppercase, e.g. `US:AAPL`, `KR:005930`), `IssuerMappingVersion = "a112-bar-issuer-v1"`.
5. **Dual cutoff / header mapping (Q4).** `SourceEventAt` = bar `open_at`; `SourceAvailableAt` = bar close instant (`open_at + 60 s`); `ObservedAt` = the official response instant the bar was read from (client clock, response-bound, used only for staleness/future-bar rejection); `IngestedAt` = store clock (= design `recorded_at`). `EvaluationAt` cuts on `SourceAvailableAt` (design `source_observed_at` semantics), `IngestionCutoff` on `IngestedAt`.
6. **Finality rule (design l.168; the process clock is never authority).** A bar may carry `closed=true` only when a strictly later bar of the same symbol/interval was observed in an official response (**successor-observed**, `finality: "successor_observed"`) — same page, later page or later poll. The newest bar of a poll is therefore never admitted until a successor exists. Additionally rejected: `open_at_ms >= source_observed_at_ms` (future), `closed=false` (unfinished), open_at not on a 60 s boundary, bar outside the session's calendar day.
7. **Bar-label convention (residual (b) of the PASS).** `bar_label = "open_at"` (bar-start), versioned inside the payload; matches design l.142, the KR precedent and the schema comment. The pack's recommendation of a small human-approved live-session probe (one GET comparing the newest bar's timestamp against wall time) is adopted as an **L1b acceptance precondition**, not an L1a one.
8. **Session identity (Q5).** `session_id = "<CALENDAR>:<exchange-local YYYY-MM-DD>"` with `KRX` for KR (golden vector) and `US` for US; `calendar_version` = `scheduler.CalendarSnapshot.Version` (engine `CalendarDigest`); IDs use uppercase market codes, `Header.Market` stays `marketclock.Market`.
9. **Numbers.** Prices/volumes are stored as integer minor units (`uint64`, JSON integers without fraction/exponent) **plus** the raw decimal strings exactly as received; the decoder recomputes minor from raw with a declared per-currency scale (`KRW` 0, `USD` 2) using integer arithmetic only, rejects over-precise/signed/exponent strings, bare non-integer JSON numbers, and currency ≠ market currency (residual (c) variable scale).
10. **Quote (residual (a)).** L1a defines a strict `official_quote_l1:v1` payload (bid/ask/last minor + raw strings, currency, `source_observed_at_ms`, `received_at_ms`, digest = `breakoutlane.QuoteSealDigest`-compatible field set) with the same integer rules; L1b's quote producer must fail closed (typed refusal, zero envelopes) until level element schema is observed under a human-approved probe. `Reset` header stays opaque (residual (d)).
11. **Ownership guardrails.** `consumer.go` static import guard untouched; `breakoutlane` is not imported by production code (its guard permits stdlib only; conversion to `EvidenceInput` is L3); no `cmd/`, engine, router, scheduler, journal or toggle file; goldens read-only; fixtures hand-written and shape-only (never `/tmp` receipt bodies).

### L1a lot definition (Terra implementer, RED first)

- Files: new `internal/strategyevidence/breakout_bar.go` (kinds `official_closed_bar_1m`, `official_quote_l1`; payload structs; strict decoders), `breakout_series.go` (bar identity/revision helpers, `BarSeriesQuery`/`BarSeries`/`(*Store).SealBarSeries`), `breakout_bar_test.go`, `breakout_series_test.go`; edits limited to the two switch dispatches in `model.go` (`kindSupportsMarket` case list; `validateTypedPayload` kind dispatch placed after `rejectSecretFields` and before the legacy generic map so legacy kinds are byte-for-byte unaffected). Nothing else.
- REDs (task 2.5, each named): unknown field; unknown enum (`bar_label`, `finality`, `currency`, `market`); float/decimal JSON number and bare-number price; minor ≠ raw×scale mismatch, over-precise raw, signed/exponent raw; secret-like field; interval ≠ 60000; open_at not on the minute; future bar; unfinished bar (`closed=false`); duplicate identity same revision different digest → `ErrRevisionConflict` quarantine; correction append-only (`r2` visible only under a later `IngestionCutoff`; the earlier `BarSeries` digest unchanged = replay); dual-cutoff independence (`EvaluationAt` vs `IngestionCutoff`); series bounded (> 512 → refusal, no truncation); out-of-order/duplicate open_at within a series read; regular-session filter; `kindSupportsMarket` KR+US for both kinds; legacy kinds unaffected (existing fixtures' canonical bytes and digests unchanged; whole existing suite green).
- GREEN gates: `go test ./internal/strategyevidence -count=1`, `-race`, `go vet`, `gofmt -l` empty, the four lane/proposal packages' suites still green (they construct envelopes), `go build ./...`. Post-GREEN: implementer reports the mutant list it killed (bound 512→513, finality enum widened, scale table changed, dispatch order swapped, revision comparison flipped) — "passing test is not evidence".
- Then: separate Terra adversary (read-only, `-overlay` mutations), gstack review, Manager FLM/BTM refresh for the two edited bodies + new-function AST, acceptance. L1b opens only after that.

### 2026-08-16 L1a — Terra implementation report and Manager contract correction

- Terra implementer (Opus) delivered: `internal/strategyevidence/breakout_bar.go` (790 lines: kinds `official_closed_bar_1m`/`official_quote_l1`, payload structs, strict decoders, header helpers), `breakout_series.go` (217: `BarSeriesQuery`/`ClosedBarRecord`/`BarSeries`/`(*Store).SealBarSeries`, domain-separated `barSeriesDigest`), `breakout_bar_test.go` (705, 25 tests), `breakout_series_test.go` (448, 11 tests); `model.go` +12/−1 = the two permitted switch dispatches only. RED captured against unmodified code (build failure on the new symbols; the legacy-bytes test was captured pre-edit: canonical `{"blocked":true,"code":"x","score_ppm":25e1,"value":1}`, digest `7e30f2af…`, unchanged post-edit). GREEN: package `-count=1` and `-race` ok, consumers (continuationlane/reversallane/weeklyvaluelane/breakoutlane, strategyproposal under its `tossos_testseams` tag) ok, `go build ./...`, vet, gofmt (run as `$(go env GOROOT)/bin/gofmt` — gofmt is not on PATH) clean. Mutation table 6/6 killed with `sha256sum -c` restoration proof (cap 512→513, finality enum widened, USD scale changed, dispatch moved after the legacy map, revision comparison flipped, future-bar check deleted).
- Deviations reported (accepted by the Manager unless stated): D1 integer rule is by canonical value, not spelling — `canonicalJSON` runs before `validateTypedPayload`, so `1.9080e4` is already `19080`; fractions/negative exponents/signs still refused (test `TestClosedBarIntegerRuleIsAboutTheValueNotTheSpelling`); D2 `SealBarSeries` scopes by exact `source_record_id` byte range + post-decode equality instead of `market_effective_date`; D3 no separate scan cap (bound applies to the resolved bar count); D4 `calendar_version` validated but payload-only (Header must not gain fields); D5 `EvidenceID = "<SourceRecordID>:r<revision>"`; D6 `/mnt/D` fuse mount shows 755 for every file, git mode is 100644; D7 strategyproposal tests are tag-gated.
- **Manager contract correction (my error in decision 9).** The M-B receipt observed USD prices with 1–4 fractional digits (3,180 of 3,200 price fields), so `USD → price_scale 2` with "reject over-precise raw" would refuse genuine broker bars and the US lane could never run. Existing production convention is USD scale 2 (`internal/app/engine/strategy_first_leg_admission.go:110`, `internal/strategyflow/canonical_projection.go:176`), but the evidence layer must be lossless (design §5 "lossless", spec "raw decimal … 보존"). **Amended decision 9: `KRW → price_scale 0`, `USD → price_scale 4` (minor = 1e-4 USD); raw with more fractional digits than the scale is still refused (fail closed).** The conversion from evidence minor (scale 4) to the lane/proposal/admission scale-2 convention is an explicit, tested boundary in L3 (rounding direction per field decided and reviewed there; US `TickMinor` at scale 4 is 100 for the 0.01 tick), recorded as an L3 input. Quote payload uses the same scales.
- Open questions answered: (2) header↔payload binding — L1a adds one combined constructor that builds header and payload from a single typed input so a producer cannot mismatch them at write time; `SealBarSeries` keeps refusing mismatches on read; (3) keep `open_at_ms` unbound from `SourceRecordID` (defensive duplicate-minute refusal stays reachable and tested); (4) no index; (5) quote `MarketEffectiveDate` via `Market.TradingDay` is not a calendar assertion — L3 input.
- Follow-up requested from the same implementer before the adversary review: apply the amended scale table + tests, add the combined constructor + tests, re-run VERIFY and the scale mutant.
- Implementer follow-up (same agent): scale table single source `marketMoney` (KR→KRW,0; US→USD,4; the only hard-coded scale in the package per grep), realistic mixed-precision USD fixtures (4/1/2-dp raws accepted, 5-dp refused), `price_scale` mismatch table-driven; combined constructors `NewClosedBar1mEnvelope(ClosedBar1mInput)` / `NewQuoteL1Envelope(QuoteL1Input)` derive currency+scale from the market, recompute every minor from raw, write payload clocks/identity from the header they just built (header↔payload mismatch impossible through the constructor; read-side refusal kept as second defence — `TestSealBarSeriesRefusesAPayloadThatDisagreesWithItsHeader`). 87 top-level tests; VERIFY re-run green (build ./..., package `-count=1`/`-race`, consumers, strategyproposal tagged, vet, gofmt via GOROOT); scale mutants 4→3 and 4→5 KILLED with `sha256sum -c` restoration proof; `model.go` diff unchanged (two dispatch sites). Adversary and gstack reviewers dispatched (read-only, `-overlay` only).

### 2026-08-17 L1a — independent reviews

- **gstack code review (read-only, `-overlay` only): `DONE_WITH_CONCERNS`, P0=0 / P1=2 / P2=10.** VERIFY reproduced (package `-race` ok, vet, gofmt via GOROOT, consumers 299 tests / 4 pkgs, strategyproposal tagged 3, `go build ./...`, imports still only `internal/clock`, 87 top-level tests). Decisions 4/5/7/8/9/10 DEMONSTRATED; decision 6 PARTIAL. **P1-1** (`breakout_bar.go:442-523`, `:133-193`): "bar outside the session's calendar day → rejected" is ABSENT — probe: `open_at = 2026-09-01T13:30Z` inside `session_id US:2026-08-14` accepted by header helper, `NewEnvelope` and the combined constructor, and would be served by `SealBarSeries` (prefix scoping); fix: require `market.TradingDay(openAt) == session date` in the header helper and recompute it in `checkClosedBar1m`. **P1-2** (`:481-483` vs `:165-169`): payload-level check only refuses `open_at_ms >= source_observed_at_ms`, so `observed = open+1 ms … +59,999 ms` with `closed=true` passes `validateTypedPayload` (the only hook re-run on Append/replay/Valid) — an unfinished bar sealable through `NewEnvelope`; fix: require `source_observed_at_ms >= open_at_ms + 60000` in the payload check (mirrors the header rule) and cross-check `payload.SourceObservedAtMS == header.ObservedAt` on read. P2s: (1) kind/symbol SQL predicates untested (scope carried by the record-id prefix) and `:` allowed in symbols although it is the record-id separator — refuse `:` in symbol; (2) revision-identity↔payload-revision and session cross-checks on read untested; (3) overflow guards in `minorFromRawDecimal`/`appendDecimalDigit`/`canonicalIntegerValue` correct by probe (`1844674407370955.1615`@4 = 2^64−1 accepted, `.1616` refused, `1e19` ok, `2e19` refused) but untested; (4) NOT EXISTS supersedes subquery redundant with the in-memory max-revision rule (KISS); (5) no golden vector for `barSeriesDigest`, `RegularSessionOnly`/`MaxBars` not proven to be digest inputs; (6) no structural SELECT-only AST guard for `SealBarSeries`; (7) exported header helpers + raw `NewEnvelope` remain a mismatch path — unexport or document; (8) exported `Decode*Payload` use a plain decoder (no duplicate-key/trailing-value refusal on non-canonical input); (9) read side does not check Authority/SchemaVersion/IssuerIdentity/Currency equalities; (10) wire field names spelled twice (bound by round-trip tests). Mutation: control mutant killed; **10/10 reviewer mutants SURVIVED** (drop kind predicate, drop symbol predicate, drop revision/session cross-checks, drop supersedes subquery, remove overflow guards, 60 s stricter finality, drop `RegularSessionOnly` from digest, negative exponent (equivalent), drop 32-char bound). Residuals for L1b/L3: successor-observed finality enforced nowhere in L1a (producer claim; consider `successor_open_at_ms` in the payload while v1 is unshipped), calendar/gap policy is L1b's join, USD scale-4→admission scale-2 boundary in L3 (`TickMinor` 100), quote payload cannot represent an empty book (producer must refuse), `calendar_version` binding in L3, store separation and FLM/BTM refresh are Manager tasks. Repo proven unchanged (hashes `89a9e6cb…`/`985fdc02…`/`a67ab059…`).
- **Terra adversary (read-only, 38 `-overlay` mutants + 17 probes, repo proven unchanged): `HOLD`, P0=0 / P1=4 / P2=10; 30/38 mutants killed (2 equivalent).** P1-1 same as gstack P1-1 (no calendar-day binding: `open_at 2026-09-23` and `1996-01-02` accepted inside `US:2026-08-14`; with D3 an off-day flood is read before the bound); **P1-2** read path never reconciles payload with header (`breakout_series.go:134-142`): header minute 13:30 with payload `open_at_ms` 19:30 → `SealBarSeries(EvaluationAt 13:32)` returns a bar six hours past the cutoff; also KRW header with USD payload and payload session 08-13 under an 08-14 record id accepted at write time; **P1-3** the availability/finality clock rule lives only in the constructor: hand-built `Header{SourceAvailableAt: openAt, ObservedAt: openAt+1ms}` + payload observed +1 ms passes `NewEnvelope`/`Append`/`SealBarSeries(EvaluationAt = openAt+1ms)` — an unfinished bar visible a minute early (same class as gstack P1-2); **P1-4** a higher revision with the same `SourceRecordID` but a different `open_at_ms` moves the bar to another minute (identity change disguised as correction; passes the append-only test because the earlier digest is unchanged). P2s: (5) a lower-case payload symbol (`checkClosedBar1m` does not upper-case-compare) is accepted at write and refused forever on read — permanent session denial with no repair path; (6) enum tests kill only their one literal (`finality` widened with a different literal SURVIVED); (7) series digest unpinned — dropping the domain string, `PayloadDigest`, `RegularSessionOnly` or `MaxBars` from the digest all SURVIVED; (8) SQL kind/symbol/upper-bound predicates only tested as a union (previous-day fixture never exercises `textPrefixUpperBound`; a later-session and a crafted-record-id fixture needed); (9) read-side revision/session cross-checks untested; (10) NOT EXISTS supersedes subquery redundant with the in-memory max (equivalent mutant); (11) leading-zero raw `"00231.4350"` accepted → two byte-different payloads for one economic bar → conflict quarantine instead of idempotency; (12) "out-of-order bar" has no executable meaning at this layer (sorting) — needs an explicit not-applicable line or a refusal; (13) `low>high` branch dead (equivalent); (14) `regular_session`/`calendar_version` are unverifiable producer claims — L3 must not treat them as a calendar join. Verified clean under attack: decoder completeness vs task 2.5 and the golden reject list, integer bounds (`1e19` ok, `2^64` and `1e1000000` refused), raw decimal edge cases, scale table single source, quarantine/orphan-revision refusal, dual-cutoff independence, SELECT-only, ownership (model.go +11/−1 exactly two sites; imports only `internal/clock`; goldens 6/6), legacy vector `7e30f2af…` reproduced independently.
- Implementer fix round (consolidated P1×4 + P2×10, RED-first): calendar-day binding in both the header helper and `checkClosedBar1m` (`Market.TradingDay(open_at) == session date`; market-local day — US 20:00 ET post-market bar and KR 09:00 KST both accepted on their session date); read-side header↔payload binding as a ten-case switch in `scanClosedBarRecord` (market/session/interval, symbol, authority, schema version, issuer identity, `SourceAvailableAt == SourceEventAt + 60 s`, `open_at_ms`, `source_observed_at_ms`, currency, revision identity — each `EVIDENCE_IDENTITY_MISMATCH` naming the EvidenceID); the duplicate-open_at refusal remained reachable (free-form `SourceRecordID`) and its test was rebuilt so it fails for the right reason; payload finality minimum `source_observed_at_ms >= open_at_ms + 60000`; **new required payload field `successor_open_at_ms`** (multiple of 60000, `>= open_at + 60000`, `<= source_observed_at_ms`; L1a checks the structural claim only, L1b must fill it from an observed successor); symbol `:` refused everywhere and non-upper-case payload symbols refused (write-accept/read-refuse impossibility); table-driven enum tests (4–6 foreign values incl. `""` and case variants); series digest golden vector `51d80380…` recomputed from a literal length-prefixed preimage in the test + digest-changes-with-every-input test; SQL scope test with a later session, a legacy-kind row under the prefix and a forged-record-id MSFT closed bar; NOT EXISTS supersedes subquery removed (in-memory max-revision is the single authority); leading-zero raw refused; overflow boundary table (2^64−1 accepted, +1 refused, `1e19` ok, `2e19` refused, 33 chars refused); header helpers unexported; `Decode*Payload` run `canonicalJSON` first; AST SELECT-only structural test for `SealBarSeries`. Sizes: `breakout_bar.go` 1009, `breakout_series.go` 229, tests 1212/900; 106 top-level tests; `model.go` diff byte-identical to the accepted two-site diff. Mutation: 24 mutants, 0 survivors, `sha256sum -c` restoration OK — with an honesty note: the first run scored a non-compiling deletion as a kill and, once made compilable, that mutant plus two others survived because the header check masked the payload check; four tests were added and the run repeated. VERIFY green (build ./..., package `-count=1`/`-race`, consumers, strategyproposal tagged, vet, gofmt via GOROOT). Manager not-applicable line for spec "out-of-order bar": `SealBarSeries` sorts by `open_at` and refuses duplicates; ordering as an evidence property is enforced by L3's `ordered_closed_bar_ids_and_revisions` — no refusal at this layer by design.
- **Rechecks (same reviewers, read-only): adversary `ACCEPT-WITH-P2` P0=0/P1=0/P2=7 (32/37 mutants killed, 4 survivors proven equivalent + the enum-widening literal that example tests cannot kill); gstack `SHIP-IT` P0=0/P1=0/P2=5 (8/9 round-1 mutants now die, 20/20 new mutants die).** All four P1 probes refused at write AND read (off-day bar at header `breakout_bar.go:165-172` and independently at payload `:508-519`; 13:30/19:30 drift refused at `breakout_series.go:139-140`; +1 ms / +59,999 ms unfinished bar refused at write `:504-507`, exact +60,000 with `SourceAvailableAt = open_at` refused at read `:137-138`; moved-minute correction refused). Golden series digest `51d80380…` recomputed independently by the adversary from the 17-field length-prefixed preimage; legacy vector `7e30f2af…` re-verified; goldens 6/6; TradingDay edges probed (ET 00:00/04:00/09:30/15:59/19:59/20:00, DST-end neighbours, KST 00:00/09:00/23:59) — exchange-local date exactly, `time/tzdata` embedded; successor bounds accept the 15:59 ET last regular bar with a 16:00 after-hours or next-poll successor. Remaining P2s: (g1) one drift-table case is vacuous (refused before reaching the read guard) — make the table stage-aware and use a forged foreign-session row; (g2) the moved-correction test passes for an incidental reason (observed-instant case) — bind `SourceRecordID` to `(code,symbol,session,interval,SourceEventAt)` on read, which closes identity moves by construction (Manager decision: adopt; the duplicate-minute guard then becomes unreachable and is to be deleted with its test, not kept green); (g3) leading-zero raw refusal is a new fail-closed rule — recorded as contract + L1b receipt check (0 leading-zero raws in accepted pages); (g4) add `IssuerMappingVersion`/`MarketEffectiveDate == session date`/`Unit == minor` read-side equalities (cheap); (g5) 32-char raw bound is unreachable defence (≤20 integer digits + ≤4 fraction with leading zeros refused) — drop it; (a1) a hand-built `Header` via public `NewEnvelope` can still write a row that poisons a session read forever (e.g. upper-case foreign symbol payload under another symbol's header) — accepted residual: header helpers are unexported and the combined constructors are the only supported write API; L1b's producer must use them exclusively (static guard in L1b); (a2) measured scan cost: 390-bar session 42.4 ms at 1 revision/bar, 168.4 ms at 4 revisions/bar (superseded rows decoded and each row canonicalised/decoded twice through `scanEnvelope→NewEnvelope` then `DecodeClosedBar1mPayload`; no scan cap by D3) — accepted for the dedicated per-session store, recorded as an L3/L5 input (reuse the decoded payload or restore a superseding predicate before wiring into a per-minute evaluation path); (a3) enum-widening residual (`"assumed"`) recorded; (a4) overnight/extended US bars must be labelled with their own ET date by L1b, never the calendar body's `today.date` (Toss `dayMarket` start unobserved before 02:08 ET); (a5) out-of-order bar — not-applicable line recorded above; (a6) equivalents M11/M33/N16/N19 recorded so they are not mistaken for coverage; (a7) test count is 107 (implementer said 106).
- Final hygiene round (implementer): read-side record-id spine — `scanClosedBarRecord` first requires `header.SourceRecordID == closedBar1mRecordID(payload.Market, payload.Symbol, payload.SessionID, payload.IntervalMS, header.SourceEventAt)` (same builder as the header helper); this exactly subsumes the former payload-vs-query and symbol-vs-header cases, which were deleted (Manager ruling: accepted, KISS), and makes the duplicate-open_at branch unreachable — deleted with its test (two visible rows sharing a minute must share the record id = the winners map key). Stage-aware drift table (15 cases: 13 read / 1 envelope / 1 append; a forged self-consistent 08-13 row under the 08-14 prefix is killed by the record-id case), moved-correction test reworked (header clocks + payload moved consistently to 13:31 under the 13:30 record id → refused by `source_record_id`), three new read-side equalities (IssuerMappingVersion, MarketEffectiveDate == session date, Unit == minor), 32-char raw bound removed (unreachable; a 33-char raw is refused by the 64-bit rule, proven by test), measured-cost comment on `SealBarSeries` (implementer's own measurement 47.0 ms / 159.1 ms for 390 bars at 1 / 4 revisions per bar). Sizes 1016 / 242 / 1225 / 947; 106 top-level tests; VERIFY green; mutants: record-id case, three equalities and a falsified stage declaration all KILLED (compiling variants), the "stage assertion neutralised" mutant survives by construction (deleting an assertion cannot fail its own test) — reported, not counted. Hashes `ea18740b…` / `ece27d1e…` / `a67ab059…`.
- **Adversary focused confirmation: `ACCEPT`, no new P0/P1/P2.** Subsumption PROVEN — every record-id field is unambiguously delimited (`:` banned in symbols; market ∈ {KR,US}; session grammar `CAL:` + 10-char date; interval pinned 60000 on both sides) so string equality forces field-wise equality, and the SQL `e.symbol=?` predicate pins header symbol to the query; 9/9 attack constructions failed (foreign payload symbol; rewritten record id and/or symbol column; KR payload under US header; shifted session; later/earlier session prefixes; trailing byte; zero-padded minute). Duplicate-open_at branch provably unreachable (same minute ⇒ same SourceEventAt ⇒ same record id ⇒ same map key; two revisions collapse to the highest). Recorded dependency: the subsumption rests on the three grammar invariants + the symbol predicate; if a later lot admits a second interval or variable-length session date, the deleted cases must return (comment at `breakout_series.go:138-140`). Mutants Z1–Z3 KILLED; `-race`/vet/gofmt clean; `model.go` still exactly two hunks (`@@ -219 +219 @@`, `@@ -407,0 +408,10 @@`); repo unchanged.
- FLM follow-up: authoring the new-function maps exposed one untested branch side in `SealBarSeries` (the winners loop's `continue` when a lower revision is scanned after a higher one — reachable only when `revision_identity` text order inverts numeric order, `r10 < r2`). Implementer added `TestSealBarSeriesPicksTheHighestRevisionNumericallyNotTextually` (r1..r11; asserts the SQL scan order `r1,r10,r11,r2,…` in-test; full cutoff → r11 with the `continue` side taken 8 times; cutoff at r10 → r10, numeric not textual); a compiling mutant neutralising the guard degrades the winner to r9 — killed. Test files only; production hashes unchanged; 107 top-level tests.

### Manager acceptance — L1a (2026-08-17)

- Evidence chain: CodeGraph pack → Manager brief (decisions 1–11, decision 9 amended to KRW 0 / USD 4) → Terra RED→GREEN → adversary HOLD (P1×4) + gstack DONE_WITH_CONCERNS (P1×2) → consolidated fix round → adversary ACCEPT-WITH-P2 + gstack SHIP-IT → hygiene round → adversary focused ACCEPT (subsumption and unreachability proven) → FLM/BTM refresh for the two edited bodies (`kindSupportsMarket` 5 branches, `validateTypedPayload` 12 branches, SHA `a67ab059…`) → new-function AST/FLM/BTM for the eight functions the reviews cite (`checkClosedBar1m` 28, `newClosedBar1mHeader` 11, `minorFromRawDecimal` 10, `canonicalIntegerValue` 8, `NewClosedBar1mEnvelope` 6, `Store.SealBarSeries` 11, `scanClosedBarRecord` 16, `normalizeBarSeriesQuery` 7; `check_analysis.py` evidence complete) → Manager verification (`go build ./...`, package 266 tests incl. subtests, vet, gofmt via GOROOT, consumers 299/4 pkgs, strategyproposal tagged 3, `TestA112MBUS` static guard 58 untouched; `model.go` +11/−1 exactly two hunks; hashes `ea18740b…` / `ece27d1e…` / `a67ab059…`).
- **L1a ACCEPTED. Tasks 2.5, 3.1, 3.2 checked.** Ownership honoured (four new files + two dispatch sites; no `official`/`breakoutlane`/`net/http` import; goldens 6/6; consumer static guards green). Nothing staged or committed.
- Residuals carried (all recorded above): finality successor rule is a structural claim in L1a — L1b's producer must fill `successor_open_at_ms` from an actually observed successor and own the "newest bar of a poll is never admitted" rule with its own RED; calendar join / regular-session flag / gap policy are L1b's; overnight US bars labelled with their own ET date; leading-zero raw refusal to be checked against the receipt (0 leading-zero raws) in L1b; the L1b producer must use only the two combined constructors (static guard in L1b) because a hand-built `Header` via public `NewEnvelope` can still poison a session read; `SealBarSeries` measured cost (≈47/159 ms per 390-bar session at 1/4 revisions) — L3/L5 reuse the decoded record, never call it on a per-minute hot path; USD scale-4 → admission scale-2 conversion with per-field rounding is an L3 boundary (`TickMinor` 100); the record-id subsumption rests on the symbol/session/interval grammar invariants (deleted read-side cases must return if a later lot admits a second interval); enum-widening residual (example tests cannot kill a foreign literal); quote payload cannot represent an empty book — L1b's producer refuses with zero envelopes until level schema is observed under a human-approved probe.
- Gates: `openspec validate --strict` valid, PM tracker current, `git diff --check` clean, `check_analysis` complete, `make sdd-sync` (first attempt hit the 15 s `codegraphcontext stats` advisory timeout, retried clean), `make sdd-check` exit 0 (GBrain advisory WARN only).
- Next: L1b brief (official raw US reader + KR/US bar/quote producer; Pre-Edit declaration for the High-risk-adjacent official client path; live-session bar-label probe as an acceptance precondition).

## 2026-08-17 L1b lot opening — official strict minute-candle reader + closed-bar producer (Manager brief)

### Inputs (Full SDD order)

- Memory recall: token-war precondition (`tossos-shared-openapi-credential-token-war`), "FLM before claiming", "passing test is not evidence", "contract numbers from the receipt", "boot is not a button" (a poll must not carry a human's intent). OpenSpec: L1 row `tasks.md:10`, tasks 3.6 / 3.7 / 3.8, `design.md` §5 ("official read-only candle/quote adapter가 regular-session closed bars를 … append한다"; l.166 clock rule; l.168 finality), M-B PASS record `analysis/measurements/m-b-us-source/receipt-2026-08-16-run4.json`, documented contract `docs/migration/openapi.latest.json` (`Candle.timestamp` = "봉 시작 시각"; `before` = "페이지네이션 상한 (inclusive, ISO 8601) … 미지정 시 가장 최신 봉부터"; `nextBefore` `string|null`, "마지막 페이지면 null", not in `required`; `count` ≤ 200; `OrderbookEntry {price, volume}` decimal strings — documented, still unobserved).
- CodeGraph hard evidence (live index 34,053 nodes): `SealBarSeries` callers = 19 tests only; `WithAttemptObserver` production callers = `M0ReadSource.ConditionalOrderRaw/OrderRawByID` (`internal/official/m0_reads.go:24-36`) and `verifylive.m0ReadContext` — the raw-bytes-through-the-ordinary-path pattern already exists in production; `RawMinuteCandles` still has zero production callers.
- AST artifacts produced **before** this brief for every existing function whose branches it relies on (citation-only bundles, not edited): `analysis/function-logic/internal-official--unwrapanddecode/` (4 branches: `out == nil` early return without decode; std envelope decode; `Result == nil`; payload decode), `internal-official--client.send/` (9: token; ≤2-iteration 401 refresh loop; `!adopted` break; non-2xx `classifyStatus`; 2xx `unwrapAndDecode`), `internal-official--client.dorequest/` (2: transport error; `ReadAll` error — rate headers recorded before the body, `io.ReadAll` uncapped, observer emitted on all three paths), `internal-strategyevidence--store.append/` (17: empty digest; store clock before observed; identity found → idempotent iff digest **and** immutable provenance equal else quarantine + `ErrRevisionConflict`; supersedes must exist and share scope; insert), `internal-scheduler--adaptcalendarday/` (10: empty date; KR `Integrated.RegularMarket` / US `RegularMarket`; nil session → `Regular == nil` without error; exchange-local date equality; early close), `internal-clock--market.tradingday/` (1). Their FLM/BTM markdown is being completed by a documentation agent (citation-only disposition); `check_analysis` must be complete before the L1b gate.
- Ledger fact (recorded honestly): the host rebooted at 00:59 KST 2026-08-17 and `/tmp` was cleared — the frozen raw receipts `/tmp/a112-mb-us-receipt-run3.lCsMTV`, `/tmp/a112-mb-us-receipt-run4.ZTagql` and the pinned collector `/tmp/a112-mb-us-build-amended.OSVxn5` no longer exist. The M-B PASS stands: it was verified before the loss by Manager replay and two independent reviewers, and the repository holds only the sanitized receipt/manifest by design (`design.md` "Raw body는 secure /tmp에만"). Consequence: residual (e) (run-A stderr) is closed as unrecoverable; no later decision may depend on re-reading raw bodies; the L1b contract is fixed from the sanitized receipt + documented spec + the live probes below.

### Manager decisions (binding for L1b; numbering continues from the L1 brief)

12. **Lot split refinement.** L1b = strict minute-candle reader (KR **and** US) + closed-bar producer (bars half of task 3.6, the bar rows of 3.7). **L1c** = quote (orderbook top-of-book + last) — opens only after the human-approved level-schema probe (decision 10 fail-closed rule) and gets its own brief. Task 3.6 is checked only after L1c.
13. **One strict reader for both markets (amends the KR clause of decision 2).** New `internal/official/strict_minute_candles.go` (+ `_test.go`): `func (c *Client) StrictMinuteCandles(ctx context.Context, market, symbol string, count int, before string) (StrictMinutePage, error)`. `RawMinuteCandles` is **untouched** (a047's 5m identity and its US-refusal regression tests stay); the golden's KR identity is the endpoint/params/raw DTO, which this reader preserves (`/api/v1/candles`, `interval=1m`, `adjusted=false`, raw decimal/timestamp strings). Rationale: pack Q6 — the KR bar path must not get a weaker cursor/envelope contract than US; one implementation, one test suite. Ordinary production token path: the reader calls `c.get(ctx, "/api/v1/candles", q, nil)` (`out == nil` ⇒ `unwrapAndDecode` returns without a std decode — AST B1) and captures the exact bytes of the **last** attempt through a **chained** `WithAttemptObserver` (any observer already on the ctx is still called — the M0 precedent must not be shadowed). `send`'s ≤2 refresh-on-401 (AST B4–B8) is accepted as production behaviour; the reader adds no retry, never calls token/refresh/exchange, and refuses if no attempt with a 2xx body was observed. It re-implements the reference algorithm (strict object, cursor tri-state) and may not reference any `a112MBUS*` identifier (repo guard `TestA112MBUSStaticNoProductCallerOrForbiddenPath`).
14. **Strict decode contract (receipt §4 + documented spec).** Valid UTF-8; strict object at every depth (duplicate key ⇒ refuse); envelope object with `result` present and an object (other envelope keys ignored, M-B0 precedent); `result` keys exactly `{candles, nextBefore}` (unknown ⇒ refuse; `nextBefore` absent ⇒ refuse — amendment "absent/empty/non-string cursor는 HOLD"); `candles` array, `len ≤ count ≤ 200`; each candle object exactly the seven keys `timestamp/openPrice/highPrice/lowPrice/closePrice/volume/currency`, every value a JSON **string** (bare number ⇒ refuse), non-empty, decimals ≤ 30 bytes (doc `maxLength`); numeric grammar is **not** re-validated here — `strategyevidence` is the single numeric authority; timestamps must match `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{1,3})?[+-]\d{2}:\d{2}$` (numeric offset mandatory; observed `.000+09:00`, doc example without fraction) and parse; instants strictly descending and unique; every instant `<= before` when `before` is given (doc: inclusive upper bound); `before`, when given, must satisfy the same grammar **before** the request is sent; currency of every candle == market currency (`KR→KRW`, `US→USD`) else refuse; `nextBefore`: JSON string ⇒ decoded value must satisfy the timestamp grammar and be **strictly older than the oldest bar** (loop guard; observed cursor = last bar − 1 min, inclusive on the next page), `null` ⇒ terminal, anything else ⇒ refuse; body > 2 MiB ⇒ refuse (post-read cap; `doRequest` reads uncapped — accepted residual, transport-wide); empty `candles` with a string cursor is a valid page (the producer stops on it). Symbol grammar: KR `^[0-9]{6}$`, US `^[A-Z][A-Z0-9.]{0,9}$`; market ∈ {KR, US}; `count` 1..200. Query encoding: `url.Values.Encode()` (single percent-encoding; `+` → `%2B` as the doc requires; canonical order `adjusted,before,count,interval,symbol` as measured).
15. **Reader result.** `StrictMinutePage{Market, Symbol string; Candles []RawMinuteCandle /* newest first, as received */; Terminal bool; NextBefore string /* decoded, "" iff Terminal */; ReadAt time.Time /* BodyReadComplete of the used attempt — the response-bound instant */; StatusCode int; BodyDigest string /* "sha256:" of the exact bytes */; Budget RateBudget /* c.RateBudget("/api/v1/candles") after the call, advisory */}`. Exported plain struct (the producer's test fake needs to build it); a static guard in `internal/official` asserts that no non-test file other than `strict_minute_candles.go` contains the composite literal `StrictMinutePage{`.
16. **Producer package.** New `internal/officialbars/` (naming precedent `internal/officialfx`): `producer.go`, `producer_test.go`, `guard_test.go`. Import allowlist (AST guard test): stdlib + `internal/official`, `internal/scheduler`, `internal/strategyevidence`, `internal/clock` only — never `net/http`, `breakoutlane`, engine, journal, execgw, guardian, router, toggles. Constructor-only guard: production files reference `strategyevidence.NewClosedBar1mEnvelope` (and later `NewQuoteL1Envelope`) but **never** `strategyevidence.NewEnvelope` or a `strategyevidence.Header{` literal (L1a residual a1). No `a112MBUS` byte sequence anywhere in the package.
17. **Producer API and poll algorithm.**
    ```go
    type CandleReader interface { StrictMinuteCandles(ctx context.Context, market, symbol string, count int, before string) (official.StrictMinutePage, error) }
    type BarStore interface { Append(context.Context, strategyevidence.Envelope) (strategyevidence.AppendResult, error); SealBarSeries(context.Context, strategyevidence.BarSeriesQuery) (strategyevidence.BarSeries, error) }
    type BarPollInput struct { Market marketclock.Market; Symbol string; Calendar scheduler.CalendarSnapshot; PollAt time.Time; LowerBound time.Time /* zero ⇒ Today.Regular.Open */; MaxPages int /* zero ⇒ 4; 1..8 */ }
    type BarPollResult struct { SessionID, CalendarVersion string; Pages, Observed, Admitted, Unchanged, Corrections, Conflicts int; Refused []BarRefusal; NewestObserved time.Time; Terminal, Truncated, RateGated, ReachedLowerBound bool; Gaps []MinuteGap }
    func PollClosedBars(ctx context.Context, reader CandleReader, store BarStore, in BarPollInput) (BarPollResult, error)
    ```
    (a) Validate: market KR/US; symbol grammar; `Calendar.Market == Market`; `Calendar.ValidityAt(PollAt) == CalendarValid`; `Calendar.Today.Regular != nil` else typed refusal `NO_REGULAR_SESSION`; `session_id = "KRX:"|"US:" + Today.Date`; `calendar_version = Calendar.Version`. (b) Page-1 `before` = `PollAt` truncated to the second, formatted `2006-01-02T15:04:05.000-07:00` in `Asia/Seoul` for **both** markets (the measured request literal; the process clock is a query upper bound here, never authority — a slow local clock only delays admission). (c) Crawl: `count=200`; after each page enforce cross-page rules — the inclusive overlap bar (page N+1 first == page N last instant) must be **byte-identical** else refuse `OVERLAP_MISMATCH`; cursor loop (cursor not strictly older than the previous cursor) ⇒ refuse; stop when `Terminal`, an empty page, the page's oldest instant `< LowerBound` (`ReachedLowerBound`), `MaxPages` reached (`Truncated`), or the advisory `Budget.Exhausted()` before the next request (`RateGated`). **All pages are collected and validated before any append** — a refusal appends nothing. (d) Merge newest-first, unique. (e) One `SealBarSeries` read (`EvaluationAt = IngestionCutoff = PollAt`, `MaxBars 512`, `RegularSessionOnly false`) → `known[recordID] = ClosedBarRecord` (measured ≈47 ms per 390-bar session — accepted for a per-poll producer read; never inside a strategy refresh critical section). (f) For `i ≥ 1` (bars[0] is the newest observed bar and is **never** admitted — decision 6): admit only bars with `Today.Regular.Open ≤ open_at < Today.Regular.Close` (design §5 regular-session-only storage; pre/after-hours bars still serve as successors and as the "crawl passed session start" witness); `ClosedBar1mInput{Market, Symbol, SessionID, CalendarVersion, OpenAt, ObservedAt: page.ReadAt, Currency, Raw, SuccessorOpenAt: bars[i-1].open_at, RegularSession: true, SourceResponseDigest: page.BodyDigest, Revision}` where `Revision = 1` if unknown, **skip** (`Unchanged`) if known and `Raw` equal, `known.Revision+1` (`Corrections`) if known and `Raw` differs — this is what makes re-polls idempotent although the payload carries observation metadata (`source_observed_at_ms`, `successor_open_at_ms`, `source_response_digest` differ every poll; `Store.Append` idempotency compares the payload digest, AST `Store.Append` found-branch); constructor refusal ⇒ per-bar `BarRefusal{OpenAt, Reason}` recorded, poll continues (fail-closed for that bar; L3's contiguity rule refuses the session — evidence for the other bars persists); `Append` `ErrRevisionConflict` ⇒ `Conflicts++`, continue (foreign writer/quarantine); any other store error ⇒ return `STORE_ERROR` (appended rows before the error remain legitimate evidence; the result carries the counts). (g) `Gaps`: regular-window minutes missing between consecutive observed instants — informational for L5 logging; **gap policy decision:** the producer never suppresses observed bars because of a gap; refusing a session for a missing minute is L3's read-side rule (contiguity), because a late-published bar must remain appendable and visible under a later `IngestionCutoff` (dual-cutoff replay). (h) No goroutines, no clock inside (PollAt injected), no logging side effects; deterministic given pages + store.
18. **Calendar/day binding.** `Today.Date` is the exchange-local date (`adaptCalendarDay` AST: session start/end local date must equal `day.Date`); US overnight bars therefore get their ET date via `Market.TradingDay` (L1a header/payload checks) — a test polls at 01:00 KST and proves `session_id US:<ET date>`. KR uses `Integrated.RegularMarket` (09:00–15:30 KST, end-exclusive → the 15:30 closing-auction bar is outside, matching the a047 precedent `bars.go:158-159`) — recorded residual, not a defect for opening-range/breakout logic.
19. **Bar label** stays `open_at` (documented "봉 시작 시각"; live probe below confirms; the value is versioned inside the payload).
20. **Retry/token/rate policy.** Ordinary `send` (≤2 refresh on 401) accepted; the reader never touches `token.go` paths directly; 429/5xx/transport ⇒ `classifyStatus`/`ErrTransport` ⇒ the poll returns `READER_ERROR` with zero appends (pages-first rule). Rate: advisory `RateBudget` gate between pages only; the quota is shared with the live containers (receipt) — L5 sets cadence; the producer never loops.
21. **Live probes (human-approved, ≤6 GETs, read-only tossctl commands; run sheet below).** L1b acceptance precondition = the bar-label probe recorded for **both** markets (KR today 09:00–15:30 KST, US 22:30–05:00 KST); the orderbook probes feed L1c. Before any probe the human enumerates credential holders (StockOS `infra-toss-intelligence-1` must be stopped or sharing the cache — it is not in `docker ps` now; TossOS containers share the cache).
22. **Ownership / Pre-Edit declaration (High-risk adjacency: official client GET/token path).** Files created: `internal/official/strict_minute_candles.go`, `internal/official/strict_minute_candles_test.go`, `internal/official/strict_minute_candles_guard_test.go` (may be merged into the test file), `internal/officialbars/producer.go`, `internal/officialbars/producer_test.go`, `internal/officialbars/guard_test.go`. **Not edited:** `client.go`, `token.go`, `trace.go`, `ratebudget.go`, `candle_raw.go`, `candle_reads.go`, `a112_mbus_*`, anything under `internal/strategyevidence/` (L1a accepted — a defect found there is reported, not patched in this lot), engine/cmd/journal/scheduler/toggle files, goldens. Landing check: `git status --short` shows only the six new files beyond the pre-existing dirty set; new files at git mode 100644.
23. **Residuals recorded now.** `doRequest` unbounded read (post-read cap only); envelope std-decoder weakness sidestepped by decoding the observed bytes; ordinary `send` may refresh the shared token (token-war memory) — the producer runs only when L5 wires it and a human approves; the KR cursor/pagination contract is documented but not measured (M-B measured US only) — the KR live probe is bar-label only, so the first KR full-session crawl under L5 is a human-observed run before any lane reads it; unknown-key refusal makes a benign broker field addition a visible failure by design (re-measure, then widen).

### RED list for the Terra implementer (task 3.7 rows; each RED named and captured before GREEN)

Reader (`internal/official`, httptest, fake `/oauth2/token` as in `trace_test.go`): market/symbol/count/before grammar refusals (no request sent — assert zero server hits); exact query string `adjusted=false&before=<%2B-encoded>&count=200&interval=1m&symbol=AAPL`; duplicate key at envelope/result/candle depth ⇒ refuse; unknown `result` key / unknown candle key ⇒ refuse; bare-number price/volume ⇒ refuse; missing candle key ⇒ refuse; >200 candles / > count ⇒ refuse; non-descending / duplicate instants ⇒ refuse; instant > before ⇒ refuse; timestamp without offset or with `Z` ⇒ refuse; currency mismatch (US page with `KRW`) ⇒ refuse; `nextBefore` absent/empty/number/object/array ⇒ refuse, `null` ⇒ `Terminal`, string not older than the oldest bar ⇒ refuse; body > 2 MiB ⇒ refuse; 401 then 200 ⇒ success with the **second** attempt's `ReadAt`/`BodyDigest` (observer chained: an outer observer sees both attempts); 429/500 ⇒ error, no page; `Budget` reflects the response headers; `RawMinuteCandles` US refusal regression untouched (existing test still green); guard: `StrictMinutePage{` literal only in the reader file; no `a112MBUS` byte sequence in the new files (repo guard also green).
Producer (`internal/officialbars`, fake reader + real temp `strategyevidence.Store`, KR and US calendars built through `scheduler.AdaptOfficialCalendar` from hand-written `official.MarketCalendarResponse`): the newest bar of a poll is never admitted (3-bar page ⇒ 2 envelopes; each `successor_open_at_ms` == the next-newer instant); the same bar is admitted on the next poll once a successor exists; regular-window filter (pre-session and after-hours bars observed, counted, not appended; the 16:00 ET bar is the successor of 15:59); US overnight poll at 01:00 KST ⇒ `session_id US:<ET date>`, `calendar_version == Calendar.Version`; calendar invalid / `Regular == nil` / market mismatch ⇒ typed refusal, zero requests; page-1 `before` literal format asserted on the fake; multi-page crawl with the inclusive overlap deduplicated; overlap mismatch ⇒ refuse whole poll, zero appends; cursor loop ⇒ refuse; empty page ⇒ stop; `MaxPages` ⇒ `Truncated`; `LowerBound` ⇒ `ReachedLowerBound`; `Budget.Exhausted()` ⇒ `RateGated` (no further request); re-poll with identical content ⇒ `Unchanged == n`, zero new rows, zero conflicts (row count via `SealBarSeries` unchanged); correction (changed close) ⇒ revision 2 with `SupersedesRevisionIdentity r1`, and the earlier-cutoff `SealBarSeries` digest unchanged (append-only replay); foreign same-revision different-digest row pre-inserted ⇒ `Conflicts == 1`, poll continues; a bar the constructor refuses (e.g. leading-zero raw `00231.4350`) ⇒ `Refused` entry, other bars admitted; gap between observed regular minutes ⇒ `Gaps` reported and bars still admitted; store error mid-append ⇒ `STORE_ERROR` with the counts so far; determinism: same pages + fresh store twice ⇒ identical `BarPollResult` and identical `SealBarSeries` digest; guards: import allowlist, constructor-only (`NewEnvelope`/`Header{` absent), no `net/http`, no `a112MBUS`.
GREEN gates: `go build ./...`; `go test ./internal/official ./internal/officialbars -count=1` and `-race`; `go vet`; `$(go env GOROOT)/bin/gofmt -l` empty; `go test ./internal/strategyevidence -count=1` still green (untouched); `TestA112MBUS*` guard green; then the implementer's mutant table (newest-bar admitted; regular filter dropped; overlap check dropped; unknown-key refusal dropped; cursor-null/absent swapped; `Unchanged` skip dropped ⇒ conflicts appear; revision+1 → 1; before-format changed) with `sha256sum -c` restoration proof — "passing test is not evidence".

### Live probe run sheet (human runs; paste outputs; ≤6 GETs total)

Pre-check (paste): `docker ps --format '{{.Names}} {{.Status}}'` (confirm no `infra-toss-intelligence-1` running unless it shares the cache); `date +%T`.
KR (any time 09:00–15:30 KST; two GETs ~25 s apart inside one minute + one orderbook GET):
```
date +%T; tossctl quote chart 005930 --interval 1m --count 3 --output json
sleep 25; date +%T; tossctl quote chart 005930 --interval 1m --count 3 --output json
tossctl quote orderbook 005930 --output json
```
US (22:30–05:00 KST, same shape with `AAPL`).
Reading rule: with bar-start labels the newest candle's `time` equals the **current** wall-clock minute and its `volume`/`close` grows between the two GETs (the forming bar is shown, labelled by its open); with bar-end labels the newest label would be one minute **ahead** of the wall clock or the newest bar would be static. Record: label convention per market, whether the forming bar is served, and the orderbook level shape (keys, string vs number) for L1c. No other flags, no repeats beyond this sheet; a fresh explicit approval is needed for any additional GET.

### 2026-08-17 L1b — Terra implementation report and Manager correction (decision 17(c))

- Terra implementer (Opus) delivered exactly the declared files: `internal/official/strict_minute_candles.go` (520 lines, sha `5bbd3f9d…`), `strict_minute_candles_test.go` (471; the `StrictMinutePage{` guard lives here, as the brief permits), `internal/officialbars/producer.go` (408, sha `e11c0e0c…`), `producer_test.go` (909), `guard_test.go` (131); 37 new top-level tests (16 / 18 / 3). RED captured before GREEN: build failures on the new symbols for both packages; behaviour REDs against naive first cuts — reader without the `result`-key-count rule, with absent `nextBefore` treated as terminal and without the candle key-count rule (3 subtests failed: `cursor_absent`, `unknown_result_key`, `unknown_candle_key`); producer looping from index 0 with `successor = open+1m` and no overlap check (10 of 32 failed incl. `TestPollNeverAdmitsTheNewestObservedBar` `Admitted:3` want 2, `TestPollRefusesAnOverlapThatIsNotByteIdentical`; one of the ten was a stale-calendar fixture bug the implementer fixed and declared). GREEN: `go build ./...`, official 344 / officialbars 35 (both `-race`), strategyevidence 266 (untouched), breakoutlane 88, `TestA112MBUS*` 58, vet, gofmt via GOROOT. Manager re-ran: build ok, 733 tests / 4 packages, `-race` 379 / 2 packages, vet, gofmt clean, `a112MBUS` byte count 0 in all five files, `officialbars` imports = stdlib + `internal/clock`, `internal/official`, `internal/scheduler`, `internal/strategyevidence`.
- Mutation table (compiling mutants, `sha256sum -c` 5/5 restoration): 10/10 brief mutants + G1 (`NewEnvelope` reach) + G2 (`net/http` import) KILLED — with an honesty note accepted by the Manager: the first harness accumulated mutants instead of restoring them (invalid table, four false NON-COMPILING); the implementer reversed all edits, proved byte-identical restoration and re-ran per-mutant from pristine copies; only the second table is reported. The `StrictMinutePage{` guard got a `scanned >= 10` non-vacuity assertion but was not itself mutation-proved (would require editing a forbidden file) — declared. A stray `reader_full.go.bak` from the botched harness reached the repo root and was deleted after byte-comparison with its scratchpad copy (never a `.go` file; the repo AST guard never saw it).
- **Manager correction — decision 17(c) was mis-stated (my error).** I wrote "the inclusive overlap bar (page N+1 first == page N last instant) must be byte-identical"; the receipt says cursor == last bar − 1 min and page N+1's **first bar == the cursor** (inclusive), and decision 14 itself requires the cursor to be strictly older than the oldest bar — so under the measured contract page N+1 never repeats page N's last bar and the literal rule would refuse every multi-page crawl. Accepted implementer form (D1): refuse `OVERLAP_MISMATCH` when page N+1 starts **newer** than page N's oldest instant (forward walk), and require byte-equality only when the instants are equal (a defensive branch reachable through the `CandleReader` interface, tested and mutation-killed). Same lesson as decision 9 (`contract-numbers-from-the-receipt`): the overlap sentence was written from memory of the word "inclusive", not from the receipt line.
- Other deviations accepted: D2 `known` keyed by `Payload.OpenAtMS` — the `SealBarSeries` query already pins market/symbol/session/interval, and `ClosedBarRecord` exposes no `SourceRecordID` (re-spelling the record-id grammar in a second package is the duplication L1a's review flagged); D3 reader `market` must be exactly `KR`/`US` (fail-closed, producer upper-cases); D4 cursor-loop guard = new cursor strictly older than the `before` this request carried (binds page 1 too); D5 `strictMinuteMaxDepth = 16` recursion bound for the depth-strict walk (2 MiB body cap alone would not bound the stack); D6 `NO_SUCCESSFUL_ATTEMPT` unreachable through today's `send` — kept as a fail-closed guard, untested, declared; D7 guard needles assembled as `"a112"+"MBUS"` so the guard files carry no forbidden byte sequence.
- Implementer observations recorded as inputs: every stored row has `regular_session = true` (the `RegularSessionOnly` digest input is unexercised by production data — L3/L5); `ClosedBarRecord` lacks `SourceRecordID` (a later lot may add it additively). FLM-agent refinements recorded as review inputs: `send`'s bound is `attempt < 2 && code == 401`; **no test drives `c.rates.record` through `doRequest`** (rate budget only unit-tested in isolation — the shared-quota observation gap); `{"result":null}` is not `env.Result == nil` (irrelevant to the reader, which passes `out == nil`); `adaptCalendarDay`'s `default` arm is unreachable via `AdaptOfficialCalendar`; `EarlyClose` is `< 6h30m`, so full KR/US sessions sit on the boundary and are not early closes.
- Adversary and gstack reviewers dispatched (read-only, `-overlay` mutations only).

### 2026-08-17 L1b — independent reviews (round 1) and Manager rulings

- **gstack code review (read-only, 32 `-overlay` mutants, 8 killer probes incl. 2 end-to-end with a real `official.Client`): `DONE_WITH_CONCERNS`, P0=0 / P1=3 / P2=12.** P1-1 `checkOverlap` (`producer.go:344-357`) is unreachable through the real reader (reader forces cursor < oldest and every bar ≤ `before` = cursor, so page N+1 always starts older) — the multi-page fixtures use shapes the real reader would refuse and the receipt's measured contiguity is enforced nowhere; P1-2 successor rule unpinned (`SuccessorOpenAt := open+1m` SURVIVES; all fixtures adjacent minutes); P1-3 `LowerBound` never set by a test (mutant SURVIVES). P2: no `StrictReason*` constant asserted; UTF-8 rule masked (bad byte only in the cursor); deep walk/depth bound untested; gap clamps untested; `ObservedAt` vs `PollAt` indistinguishable in fixtures; `defaultMaxPages` unpinned; reader and producer never wired together (first wiring exposed the store-clock/ObservedAt coupling); duplicated symbol grammar; reader-error path drops `Pages`; unreachable `RefusalPageInvalid`/`DependencyMissing` untested; doc nits (MaxPages message, `MinuteGap` doc, "바이트" vs 7 decoded fields); page `Market`/`Symbol` never checked; 12 exported `StrictReason*` without consumers. Decisions 12–16, 18, 20, 22, 23 DEMONSTRATED; 17 PARTIAL; 19/21 ABSENT until the human live probe. Repo unchanged (hashes match).
- **Terra adversary (read-only, 61 `-overlay` mutants: 45 killed / 16 survived — 5 equivalent incl. a deliberate control, 4 masked-by-sibling pairs, 5 genuine gaps each with a killer that passes on HEAD; 24-body reader table all per decision 14): `HOLD`, P0=0 / P1=4 / P2=12.** **P1-1** `ValidityAt == CalendarValid` (`producer.go:171-173`) with `calendarMaxAge` 6 h < the 6 h 30 m sessions ⇒ no calendar snapshot lets a poll succeed after open+6h — the 15:30–16:00 ET / 15:00–15:30 KST bars (incl. the 15:59 bar that needs a successor) could never become evidence (proof `TestAdvCalendarWindowCoversTheWholeSession`: every cell at open+6h and +6h29m REFUSED); **P1-2** = gstack P1-2 (gapped successor; the receipt shows 27 non-1-minute gaps in one crawl); **P1-3** `ObservedAt := PollAt` SURVIVES (fixtures build `ReadAt == PollAt`); **P1-4** `IngestionCutoff = PollAt` for the `known` read cannot see rows the store stamped after `PollAt` (store clock in production is strictly after) ⇒ a retry with the same `PollAt` or a duplicate poll re-appends every stored bar at r1 with different observation metadata ⇒ `ErrRevisionConflict` quarantine rows and `Conflicts` (the signal reserved for a foreign writer); tests hid it by pinning the fake store clock to `PollAt`. P2: UTF-8 masked; depth bound untested; `defaultMaxPages` untested (the only quota bound on a shared credential); page identity ignored (KR/005930 page adopted for a US/AAPL poll); `seen` dedupe silently keeps the first copy / an ascending page is refused only downstream; reader-error path loses counts; a malformed (off-minute) bar poisons its older neighbour (its successor); absent `nextBefore` hard-refused although the doc lists only `candles` as required and terminal null was never observed; `[]json.RawMessage` allocation before the count bound (2 MiB ≈ 20× the measured page); `RegularSessionOnly:false` is a no-op today; `PollAt` truncation untested; guards read sources from disk under `runtime.Caller` so `-overlay` cannot mutate what they see (no efficacy claim). D1–D4, D6, D7 ACCEPTED (D2 with proof of `OpenAtMS` injectivity within one query); D5 CHALLENGED (right value, no test).

**Manager rulings (binding for the fix round; decisions continue numbering):**
24. **Calendar gate for recording (amends 17(a); resolves adversary P1-1 — my defect).** `ValidityAt` is a *trading* freshness gate; the producer records evidence and must be able to poll the whole session. New rule: refuse only `CalendarClockSkew`; accept `VALID/STALE/REFRESH_TOO_LATE`; additionally require `Market.TradingDay(PollAt) == Calendar.Today.Date` (binds the caller's snapshot to the poll instant: works at 01:00 KST for the ET session and through the after-hours), `Calendar.Market == Market`, `Today.Regular != nil`. Test at open+6h and open+6h29m; the `calendar_version` in every payload keeps the snapshot identity visible to L3.
25. **`known` read horizon (resolves adversary P1-4 — my defect).** De-duplication must see every row already stored for the session regardless of ingestion instant: read `SealBarSeries` with `EvaluationAt = IngestionCutoff = PollAt + knownReadHorizon` (`knownReadHorizon = 366 * 24h`, a documented package constant; replay cutoffs are L3's concern, not the producer's). Tests: store clock strictly ahead of `PollAt`; retry after STORE_ERROR with the same `PollAt` ⇒ `Unchanged`, zero conflicts; duplicate poll at the same instant ⇒ zero conflicts.
26. **Bar instants are minute-aligned in the reader (resolves adversary P2-7 by construction).** The measured contract has every timestamp at `:00.000`; the reader refuses any candle instant with non-zero seconds or fraction (`RefusalCandleNotOnMinute`), so an off-minute bar cannot reach the producer and poison its neighbour's successor. Cursor grammar unchanged (a cursor is a bound, not a bar).
27. **Overlap rule (resolves gstack P1-1; amends 17(c) once more).** Keep `checkOverlap` as an interface-level defensive guard with an explicit unreachability comment (L1a record-id-subsumption precedent) and this `not-applicable` line: *through `official.Client` the reader's `cursor < oldest` and `bar ≤ before` invariants make both branches unreachable; they exist for foreign `CandleReader` implementations.* Do **not** require `first == cursor` — under the measured semantics the cursor minute may have no trades (extended-hours gaps), so equality would refuse legitimate crawls; the only measured invariant is `first ≤ cursor`, which the reader already enforces. Fixtures: the primary multi-page tests use the measured shape (cursor = last − 1 min, page N+1 starts at or below the cursor); the two guard tests stay, labelled as interface-guard tests; add one end-to-end test wiring the real `official.Client` (httptest, 2 pages) to `PollClosedBars` with a real-clock store (`Options.Clock` default) and `PollAt = time.Now()` (P2-7). Replace the silent `seen` dedupe: drop only the equal-instant overlap bar after `checkOverlap`, then refuse if the merged list is not strictly descending (`RefusalPageOrder`).
28. **Reader last-attempt rule (D6/R9).** `strictMinuteLastSuccess` becomes "the LAST attempt must be 2xx" (refuse otherwise) — equivalent today, future-proof if `send` ever retries after a 2xx.
29. Cheap hygiene accepted: check `page.Market == code && page.Symbol == symbol` (`RefusalPageIdentity`); reader-error path keeps `Pages`; drop the producer's duplicated symbol regexp (the reader refuses before any request; the constructor and `SealBarSeries` refuse again) — keep a nil/empty check only; assert `StrictReason*` constants in the reader table; a proper UTF-8 test (bad byte inside `openPrice`); depth-bound test (17-deep refused, 16 accepted); `defaultMaxPages == 4` pinned; gap-clamp edge tests; `SupersedesRevisionIdentity == "r1"` asserted in the correction test; `PollAt` sub-second truncation test; explicit `LowerBound` test; gapped-successor test (every admitted bar's `successor_open_at_ms == next-newer observed instant`); `ObservedAt` provenance test (`ReadAt = PollAt − 10 min`); doc nits (MaxPages message "0 = default 4, else 1..8", `MinuteGap` doc, "7개 필드가 다르면"); comment the unreachable `RefusalPageInvalid`/`DependencyMissing`/`strictMinuteWalk default` paths. Residuals recorded, not fixed: absent `nextBefore` refused although the doc lists only `candles` as required (fail-closed until a terminal is measured; already decision 14/23); `[]json.RawMessage` allocation before the count bound within the 2 MiB cap; guards not mutation-provable via `-overlay`.
- Fix round (implementer, RED-first; both reviewers' session limit interruption at 05:10 KST handled by resuming them on the unchanged hashes): rulings 24–29 applied — calendar gate refuses only `CalendarClockSkew` and binds `Market.TradingDay(PollAt) == Today.Date` (`RefusalCalendarDay`; RED reproduced the adversary's open+6h/+6h29m refusals before the fix); `knownReadHorizon = 366*24h` for the de-duplication read (RED reproduced the same-`PollAt` retry conflicts); reader refuses non-minute-aligned candle instants (`StrictReasonCandleNotOnMinute`) and has its own `StrictReasonInvalidUTF8`; `checkOverlap` kept with the not-applicable unreachability comment, silent `seen` dedupe replaced by drop-equal-overlap + `RefusalPageOrder`, primary multi-page fixtures rewritten to the measured shape, two guard tests renamed `TestPollInterfaceGuard…`; end-to-end test through the real `official.Client` (httptest) with a default-clock store and `PollAt = time.Now()`; `strictMinuteFinalAttempt` = last attempt must be 2xx; `RefusalPageIdentity`; `Pages` preserved on reader error; producer symbol regexp removed; doc nits; the masked duplicate/trailing pair resolved by making the deep walk the single authority (`strictMinuteObject` reduced to a pure extractor). Sizes 531 / 500 / 456 / 1440 / 131; 52 top-level tests; hashes `441bed46…` / `a866693f…`; mutants N01–N18 KILLED, N19 (last-2xx vs last-attempt) equivalent by construction; round-1 mutants re-run KILLED. Manager re-verified: build, 755 tests / 4 packages, `-race` 401 / 2 packages, vet, gofmt, `a112MBUS` count 0, `git status` = the five new files only.
- **Rechecks: gstack `SHIP-IT` P0=0/P1=0/P2=6 (all 15 round-1 findings RESOLVED; 10/11 surviving mutants now killed, R9 equivalent; 8 new-rule mutants killed; no hole found in the TradingDay binding); adversary `ACCEPT-WITH-P2` P0=0/P1=0/P2=4 (all four P1 RESOLVED with its own killer tests passing on HEAD; 26/30 mutants killed — survivors NK/R25 equivalent, CTRL control, P24 recorded no-op; the former masked pairs now die individually).** New P2s: adjacent-duplicate instant in a foreign page not refused as a whole poll (`RefusalPageOrder` only tested ascending; fail-closed per bar downstream); cursor sub-minute acceptance untested (ruling 26's "cursor is a bound" unpinned); e2e ObservedAt assertion `>=` does not discriminate `ObservedAt := PollAt` and the skip window is ~22 min/day not ~14; `RefusalPageOrder` lacks the sibling's unreachability comment; producer trusts `Today.Regular` window to lie on `Today.Date` (guaranteed by `AdaptOfficialCalendar`, not for hand-built snapshots); `knownReadHorizon` comment lacks the cost/512 note; ruling 26 makes one off-minute bar refuse the whole page (availability trade-off, by design); observation: under ruling 25 a pre-existing foreign same-identity row is absorbed as a correction and `Conflicts` signals only a writer racing between the dedupe read and the append (L5 must not read `Conflicts == 0` as "no foreign writer").
- Final hygiene round (implementer): adjacent-duplicate interface-guard test (`TestPollInterfaceGuardRefusesAnAdjacentDuplicateInstant`, mutant relaxing the merged-order rule KILLED); sub-minute cursor ACCEPT case pinned (`TestStrictMinuteCandlesAcceptsASubMinuteCursor`, stricter mutant KILLED); e2e ObservedAt assertion strictly `>` PollAt (mutant `ObservedAt := PollAt` KILLED by the wiring test alone) and the skip window corrected to ~22 min/day; `RefusalPageOrder` not-applicable comment and `knownReadHorizon` cost/512 note; new fail-closed check that `Today.Regular.Open/Close` fall on `Today.Date` in the market's location (`RefusalCalendarInvalid`; RED `TestPollRefusesACalendarWindowThatIsNotOnItsOwnDay`, mutant KILLED). Sizes 531 / 517 / 482 / 1497 / 131; 55 top-level tests (18 / 34 / 3); hashes `441bed46…` (reader unchanged) / `83960410…` (producer). Manager re-verified: build, 758 tests / 4 packages, `-race` 404 / 2 packages, vet, gofmt.

### Manager verification — L1b code complete (2026-08-17); acceptance pending the human live probe (decision 21)

- Evidence chain: citation-only AST bundles for the six existing functions the brief relies on (before the brief) → Manager brief (decisions 12–23) → Terra RED→GREEN → adversary HOLD (P1×4) + gstack DONE_WITH_CONCERNS (P1×3) → Manager rulings 24–29 (two of the P1s were defects in my brief: the 6 h calendar freshness gate vs 6 h 30 m sessions, and the `IngestionCutoff = PollAt` de-duplication read; one more brief error corrected earlier: decision 17(c) overlap wording) → fix round → adversary ACCEPT-WITH-P2 + gstack SHIP-IT → hygiene round → new-function AST bundles (12: `PollClosedBars` 42 branches, `adoptPage` 2, `checkOverlap` 3, `minuteGaps` 5, `Client.StrictMinuteCandles` 11, `strictMinuteFinalAttempt` 2, `strictMinuteDecode` 10, `strictMinuteCandles` 7, `strictMinuteCandle` 10, `strictMinuteCursor` 6, `strictMinuteObject` 8, `strictMinuteWalk` 15) with FLM/BTM authored by the documentation agent.
- Ownership honoured: exactly five new files (`internal/official/strict_minute_candles.go` + test, `internal/officialbars/{producer.go, producer_test.go, guard_test.go}`); no forbidden file modified; `a112MBUS` byte count 0; `officialbars` imports = stdlib + `internal/clock`, `internal/official`, `internal/scheduler`, `internal/strategyevidence`; nothing wired into engine/cmd; nothing staged or committed.
- **What remains before L1b is ACCEPTED:** the human-run bar-label live probe for KR (09:00–15:30 KST) and US (22:30–05:00 KST) per the run sheet above (decision 21). If the probe shows bar-start labels with the forming bar served, L1b is accepted as-is (task 3.7's bar rows recorded; 3.6 stays open for L1c quote); if it shows bar-end labels, `BarLabelOpenAt` semantics and the payload constant become an L1a/L1b amendment before acceptance. The orderbook probe result opens L1c.
- Residuals carried: absent `nextBefore` refused (doc lists only `candles` as required; terminal null unmeasured); `doRequest` uncapped read (post-read 2 MiB cap); `[]json.RawMessage` allocation before the count bound; one off-minute bar refuses the whole page (visible failure by design); `Conflicts` signals only a writer racing between the dedupe read and the append (a pre-existing foreign same-identity row is absorbed as a correction — L5 must not read `Conflicts == 0` as "no foreign writer"); `RegularSessionOnly` is a no-op on this producer's output; guards not mutation-provable via `-overlay`; e2e test skips ~22 min/day near ET midnight (announced); the KR cursor/pagination contract is documented but unmeasured — first KR full-session crawl under L5 is a human-observed run; the ordinary `send` may refresh the shared token — the producer runs only when L5 wires it under human approval; `RawMinuteCandles` still spells the candles path as a literal (out of scope; follow-up so the rate-budget key cannot drift).
- Gates at code-complete (2026-08-17): `openspec validate --strict` valid; PM tracker current; `git diff --check` clean; `check_analysis` "evidence complete or diff-proven exempt" (12 new-function bundles: 121 branch rows, every AST id covered — recorded gaps: `PollClosedBars` B1 nil deps / B33 seal-read error / B32 untaken side, `strictMinuteCandle` B4 seven-keys-one-renamed, `strictMinuteWalk` B7/B15 malformed-stream paths; documentation agent's corrections: the `cursor absent` subtest lands on `strictMinuteDecode`'s key-count branch, `cursor replaced by another key` is what pins the cursor rule; test count for the two packages is 404 with `-race`, the fix-round line's 401 was the pre-hygiene count); `make sdd-sync` clean (GBrain advisory busy → previous freshness kept, WARN only); `make sdd-check` exit 0 (GBrain advisory WARN only). tasks.md unchanged: 3.6 stays open for L1c (quote), 3.7's bar rows are recorded here and get checked at L1b acceptance after the probe.

### Live probe log (human-run, decision 21)

- **KR attempt 1 — 2026-08-17 09:10:33 KST, 1 GET (`tossctl quote chart 005930 --interval 1m --count 3 --output json`; the second chart GET and the orderbook GET were not executed — the human interrupted).** Result: newest candle `2026-08-14T20:00:00+09:00` (Friday), then 19:59 and 19:58 — no bar for 2026-08-17 at 09:10. Reading: **2026-08-17 is a KRX substitute holiday** (광복절 2026-08-15 falls on Saturday → Monday substitute), so no forming bar exists and the KR bar-label probe cannot be read today; deferred to 2026-08-18 09:00–15:30 KST. Incidental evidence: the last Friday bar is labelled 20:00 (NXT after-market close) with volume 150,860 vs 44,910 / 35,153 for the two preceding minutes — a closing-print signature that is consistent with **either** convention (bar-end 19:59–20:00, or bar-start 20:00 holding the 20:00:00 closing-auction print), i.e. not disambiguating on its own — the two-GET intra-minute comparison during a live session remains the discriminating measurement. Values in this output are float-adapted `Client.Candles` output used only for the label question, never as evidence.

### 2026-08-17 Landing (human directive: commit → main merge → push → deploy)

- Human instruction after the KR probe attempt: 「커밋 로컬/메인 병합 푸시 배포까지 해줘」. Manager caveats stated: a112 is mid-flight (L1b pending probe; L1c, L3–L7 unstarted) so what lands is inert (all lanes OFF/UNOBSERVED, no engine wiring; precedent: `016da624 [A100 R0/M0]` landed mid-change); the branch was 50 commits behind `origin/main`; deployment restarts the engine and is done while both markets are closed (KRX substitute holiday, US pre-open).
- Commits on `feat/a112-four-family-runtime` on top of `016da624`: `7d628041` docs/PM, `5f73ea74` L2 breakoutlane, `a78ee6ff` L1a strategyevidence, `29f05054` M-B seam + collector, `5a44d424` L1b reader + producer (all new Go files at mode 100644; the human's untracked `save-session.sh` was left alone). Merge `d9d9f71f` of `origin/main` (`a3671987`): conflicts only in `_registry.yaml` (story list a112 + a113–a115) and the two generated trackers (regenerated); `go build ./...` and the full `go test ./...` on the merged tree exit 0. `base-commit.txt` re-frozen to `d9d9f71f` (`current-main-evidence.md` records the proof that main touched none of a112's edited/cited files, so all AST bundles stay valid).
- Not a completion claim: `make gate CHANGE=a112…` steps 1–2 (open tasks) are expected to fail until L7; landing runs steps 5–9 (`sdd-check`, `test`, `vet`, `validate`) plus openspec strict, PM check, `check_analysis`.
- **Deployment verified (2026-08-18 03:22 KST).** Both containers run `tossos@sha256:4d4c1061…` (the a112 pin built from `3cd38531`): `docker inspect` `Config.Image` is the digest, `created` 2026-08-17 11:11 KST (the human's replacement run; the compose output pasted later reads `Running` because the same digest was already in place — an idempotent re-run, not a skipped deploy), `config_files` now the canonical `~/.local/share/tossos-deploy/release-override.yaml`. Host rebooted 2026-08-18 03:17; both containers returned at 03:18, `healthy`, `RestartCount=0`, `docker top | grep -c 'engine run'` = 1. Content differential inside the running container: `grep -ac StrictMinuteCandles` 6, `official_closed_bar_1m` 2 (a109 build 0 / 0) — the a112 code is present and, as designed, inert (`officialbars.PollClosedBars` is not linked into `cmd`, no engine wiring, lanes OFF/UNOBSERVED).
- **Rollback-pin trap recurred and was repaired.** The run sheet's `cp release-override.yaml rollback-override.yaml` was executed a second time *after* release had already been switched to the new digest, so the rollback file pointed at the new image (no rollback target). The a109 image `a5673fb81ebb` was still tagged, so `rollback-override.yaml` was restored to it and `schema-compatibility-a112.txt` written (journal schema 31 both ways, no migration). Procedure fix for the next landing: the `cp` belongs strictly before the first override rewrite; re-running the sheet must skip it.

- **US bar-label probe (human-run, 2026-08-18 03:29 KST, 3 GETs) — the convention is bar-END, contradicting decision 7/19.** Local clock NTP-synced (`timedatectl`: `System clock synchronized: yes`, NTP active), so a clock-skew explanation is excluded by measurement, not assumption. GET 1 at 03:29:14 KST returned newest `2026-08-18T03:30:00+09:00` vol 251 (o 304.68 / h 304.685 / l 304.68 / c 304.68), then `03:29:00` vol 2,997 and `03:28:00` vol 2,155. GET 2 at 03:29:40 returned the **same** newest label `03:30:00` with vol **1,089**, h 304.7176, l 304.67, c 304.69 — it grew — while `03:29:00` stayed frozen at 2,997. So while the wall clock was inside the minute `[03:29:00, 03:30:00)`, the bar labelled `03:29` was already final and the bar labelled `03:30` was the one accumulating: **the timestamp is the bar's close instant, not its open instant.**
- Independent corroboration without any clock: the 2026-08-17 09:10 KR probe returned as its newest bar `2026-08-14T20:00:00+09:00` with vol 150,860 against 44,910 / 35,153 for the two before it. KRX NXT after-market closes at exactly 20:00 KST, so under bar-start a bar labelled 20:00 would begin at the close instant and could not carry the closing print; under bar-end it is the session's final minute `[19:59, 20:00)`. My earlier note calling that observation "consistent with either convention" was wrong — it points the same way as the US probe. (KR's own convention is still formally unmeasured; the KR probe stands for 2026-08-18 09:00–15:30 KST.)
- Every source the brief leaned on was wrong on this point: `docs/migration/openapi.latest.json` (`Candle.timestamp` "봉 시작 시각"), `internal/official/candle_reads.go:16`'s schema comment, `design.md:142`'s bar identity, and the a047 KR precedent (`internal/strategymarket/bars.go:158-159,189`, which computes `closedAt = minutes[0] + 5m` from a timestamp it reads as an open). Documentation and precedent lost to one 26-second measurement — the same lesson as decision 9 (USD scale) and decision 17(c) (cursor overlap), now for a *semantic* rather than a numeric constant.
- Prices observed: `304.7176` (4 dp), `304.685` (3 dp), `304.68` (2 dp) — the amended decision 9 (`USD → price_scale 4`) is confirmed against live data; scale 2 would have refused this bar.
- **Manager amendment — decision 30 (supersedes the label half of decisions 7 and 19).** The broker's `timestamp` is the bar **close**; the stored evidence keeps `bar_label = "open_at"` (L1a's payload semantics are unchanged and remain the strategy-facing contract) and the **producer converts**: `open_at = timestamp − interval` (60,000 ms), asserting minute alignment on both sides. The reader stays a transport DTO and keeps the raw broker string, with the convention documented at the field. Consequences recorded: (a) `before` is an inclusive upper bound on the *close* instant, so a request with `before = PollAt` already excludes the forming bar — the "never admit `bars[0]`" successor rule stays (decision 6 requires an *observed* successor as proof) at the cost of one poll of latency per bar, which is acceptable for a 1-minute cadence; (b) the regular-window filter now compares true open instants, so the last regular US minute is the bar labelled `05:00 KST` and the bar labelled `22:30 KST` is the last pre-market minute — the inverse of what the brief assumed; (c) `setup_id` derives from `Bars[0].ID` (`breakoutlane/machine.go:201`), so this correction had to land before any setup identity was minted — it does, since nothing is wired.
- **Latent defect recorded, out of a112 scope:** `internal/strategycandle` + `internal/strategymarket`'s a047 KRX 5-minute aggregator reads the same endpoint's timestamp as a bar open; under bar-end labelling its aggregation window and its 09:00–15:30 minute-of-day guard are shifted by one minute. It has **zero production callers** (evidence pack §1.3; CodeGraph: 5 test callers of `RawMinuteCandles`, 1 test caller of `AdaptAndSealClosedKRXFiveMinute`), so there is no live impact, and KR's convention is not yet measured. Recorded here for a separate change; a112 does not edit it.
- **L1c input (orderbook probe, same run):** live US AAPL returned exactly **one level per side** — `offers[{price 304.71, volume 160}]`, `bids[{price 304.66, volume 40}]`, `total_offer_volume 160`, `total_bid_volume 40`. That answers "the level element schema is UNOBSERVED" for the *shape* (`{price, volume}`, one L1 level per side, non-empty during a live session) but the values came through the float-adapting console path (`adaptOrderbook`/`parseDecimal`), so the raw decimal string precision is still unmeasured; L1c's own strict reader must observe the raw bytes before the quote producer leaves fail-closed.
- **Decision-30 correction landed in code (2026-08-18).** The conversion lives at exactly one point — `officialbars.adoptPage` parses the wire value as a close instant, computes `openAt = closeAt − barInterval`, and refuses `RefusalPageInvalid` if the converted value is not minute-aligned; everything downstream already worked on `observedBar.openAt`, so nothing else moved. The reader is doc-only: it keeps the broker's bytes (its `BodyDigest` covers them) and now states that `Timestamp`/`before`/`nextBefore` are close instants and that `before = now` already excludes the forming bar. The implementer kept the conversion at adoption rather than in the reader (a transport DTO must not reinterpret bytes) or at envelope construction (the window filter, successor and gap logic in between would still be reading a close as an open) — accepted. RED: re-basing the fixtures to express opens made **13 of 57 producer tests fail, every one off by exactly one minute** (`newest observed 13:34` want `13:33`, gap start, `LowerBound` not reached, the window test storing the wrong two minutes, the after-hours clamp at 385 want 386) — the measure of how far the wrong convention had spread. Fixtures re-based with justification, not silently: `usBar`/`krBar` now take the open and synthesise `Timestamp = open + 1m`; new `usBarLabelled`/`krBarLabelled` speak in broker labels for the boundary tests; `cursorBefore` changed from `open − 1m` to `open` because "oldest label − 1 min" equals the oldest bar's own open (the emitted string is unchanged). Four new tests (conversion + `successor_open_at_ms`, US closing label as last regular bar / 09:30 label as pre-market, the KR equivalent, gaps on converted opens); producer suite 34 → 38 top-level. Mutants (a) conversion dropped, (b) inverted, (c) applied twice, (d) not applied to `SuccessorOpenAt`, (e) old bar-start expectation restored — **5/5 KILLED**, with an honesty note that the first (e) was an equivalent mutant that survived and was redefined before being reported. VERIFY: `go build ./...`, official + officialbars `-count=1` and `-race` (408), strategyevidence 266 + breakoutlane 88 untouched, `TestA112MBUS` 58, vet, gofmt — all exit 0. Hashes `d32181a9…` (reader) / `8d45ca93…` (producer).
- FLM refresh for the correction: the twelve L1b bundles regenerated against the new sources (branch counts unchanged except `adoptPage` 2 → 3 for the minute-alignment refusal; `pollclosedbars`' regular-window rows rewritten to the inverted boundary and a two-clock-space invariant added — `observedBar.openAt` is converted, cursor/`before` deliberately stay in label space). The gate then required six more bundles because the correction edited test helpers that existed at the frozen base — `usBar`, `krBar`, `cursorBefore`, `TestPollStoresOnlyRegularSessionBarsButUsesTheOthersAsSuccessors` (current revision) and `brokerTimestamp`, `TestPollRefusesACalendarWindowThatIsNotOnItsOwnDay` (**base revision**, extracted from `d9d9f71f`, since only the base-side hunk intersects them). All written and honest about being test-side. `check_analysis` → "evidence complete or diff-proven exempt".
- **Still open before L1b acceptance:** the KR bar-label probe (2026-08-18 09:00–15:30 KST, `005930`, the same three commands). The code currently applies the measured US convention to both markets; if KR labels differ, `adoptPage` needs a per-market rule and the KR fixtures change again. L1c (quote) remains gated on a raw-bytes observation of the orderbook levels.

### 2026-08-18 KR bar-label probe — the convention is bar-END in both markets; L1b ACCEPTED

- **Receipt (human-run, 2026-08-18, 3 GETs, `005930`).** GET 1 at **09:06:26 KST** returned newest `2026-08-18T09:07:00+09:00` vol 44,854 (o 284,000 / h 284,000 / l 283,500 / c 283,500), then `09:06:00` vol 207,297 (o 283,500 / h 284,000 / l 283,000 / c 284,000) and `09:05:00` vol 215,174; `fetched_at 00:06:27.602Z`. GET 2 at **09:06:52 KST** returned the **same** newest label `09:07:00` with vol **84,562** and c 284,000 — it grew — while `09:06:00` stayed frozen at 207,297 and `09:05:00` at 215,174; `fetched_at 00:06:53.096Z`.
- **Reading.** The wall clock sat inside the minute `[09:06:00, 09:07:00)`. The bar that accumulated was the one labelled `09:07`; the bar labelled `09:06` was already final. Under bar-start the forming bar would have to be the one labelled `09:06` and it would have to grow (it did not), and a bar labelled `09:07` could not exist at all — its window had not begun. **KR labels a bar by its close instant, exactly as US does.** Two corroborations inside the same receipt: the accrual matches a window opened at 09:06:00 (44,854 at +26 s ≈ 1,725/s; 84,562 at +52 s ≈ 1,626/s), and the `09:07` bar's open 284,000 equals the `09:06` bar's close 284,000 (contiguous chain).
- **The probe measured the path L1b reads.** `tossctl quote chart` routes `hybrid.Client.GetChart` → `official.Client.Candles` → `GET /api/v1/candles`; the receipt's `"product_code": ""` proves the official adapter served it — `adaptCandles` leaves `ProductCode` empty by construction (`internal/official/candle_reads.go:99-105`) while the WTS chart path always fills it (`internal/client/chart.go:70-72`). Values are float-adapted and were used only for the label question, never as evidence.
- **Manager decision 31 — no per-market rule; decision 30 stands for KR too.** Both markets are now measured bar-END, so `officialbars.adoptPage`'s single market-independent conversion (`open_at = timestamp − 60,000 ms`, minute alignment asserted on both sides) is correct as landed in `d8a87196`: **no code change follows from this probe** and the KR fixtures stand as re-based. Residual unchanged: this measures KR *regular-session* minutes (the 2026-08-14 NXT `20:00` closing print corroborates the same convention), while the KR **cursor/pagination** contract is still unmeasured — the first KR full-session crawl under L5 stays a human-observed run.
- **L1b ACCEPTED (2026-08-18).** Decision 21's acceptance precondition is discharged. Scope delivered: strict reader `internal/official/strict_minute_candles.go` (+ test) and closed-bar producer `internal/officialbars/{producer.go, producer_test.go, guard_test.go}`; 38 producer + 18 reader top-level tests; the conversion mutation-proven 5/5; 18 AST bundles (12 new-function + 6 test-side, two at base revision) with `check_analysis` complete; gates green at `d8a87196`. What is landed stays **inert** — nothing wired into `cmd`/engine, every lane OFF/UNOBSERVED.
- **tasks.md stays unchanged, deliberately.** 3.6 waits for L1c (the quote half of the producer). 3.7 is a single atomic row that also demands the quote REDs, the timezone/restart/correction/cache-miss replay fixtures and broker mutation spies at zero; the bar half is recorded here and the row is **not** checked — checking it now would claim work that does not exist.
- **L1c input (KR orderbook, same run).** `tossctl quote orderbook 005930` returned **ten levels per side** with integer KRW prices (offers 284,000…288,500, bids 283,500…279,000, tick 500), `total_offer_volume 1,194,201` and `total_bid_volume 655,237` — each **exactly** the sum of its own ten levels, so the totals expose no depth beyond the returned book. Live US AAPL returned **one** level per side: **the level count is market-dependent and L1c must not assume a fixed depth** (nor read `total_*` as full-book). Integer KRW prices agree with L1a's `KRW → price_scale 0`. Still unmeasured: these values passed through `adaptOrderbook`/`parseDecimal`, so L1c's strict reader must observe the raw bytes before the quote producer leaves fail-closed.
- **Follow-up unit (out of a112 scope), now three copies of one false claim.** The "timestamp = bar open" reading survives in `docs/migration/openapi.latest.json` (the broker's own document — not ours to edit; recorded as measured-wrong), in the schema comment at `internal/official/candle_reads.go:16`, and in executable form in `internal/strategycandle` + `internal/strategymarket`'s a047 KRX 5-minute aggregator (one-minute shift in its window and its 09:00–15:30 guard; zero production callers). They are one correction unit and belong in one separate change; a112 edits none of them, and L1b's declared five-file ownership is preserved by that choice.

## 2026-08-18 L1c lot opening — official strict quote reader + top-of-book evidence producer (Manager brief)

### Correction to the 2026-08-18 probe record (decision 32)

32. **The orderbook totals are our own arithmetic, not the broker's answer — the "no hidden depth" reading of the KR probe is withdrawn.** The earlier entry claimed that `total_offer_volume` 1,194,201 and `total_bid_volume` 655,237 matching the sum of the ten printed levels showed the visible ladder was the whole book. It shows nothing of the kind: `adaptOrderbook` *computes* both fields with the `+=` accumulators at `market_reads.go:234` and `:245` (AST bundle `internal-official--adaptorderbook`, branches B1/B2, single return at 252), and the response DTO `apiOrderbook` (`:201-206`) has no total field at all — the doc comment at `:227` says so in words, and `TestAdaptOrderbookUnit` pins it with the literal comment `// TotalOffer = 8500+3400`. The equality was guaranteed by the code before the request was sent. This is the "generated evidence must be measured" failure in its purest form, and it is recorded rather than quietly deleted. What the probe *does* still support: KR returned ten levels per side and US AAPL one, so **level count is response data, never a constant**. The contrast that makes the point: the WTS adapter reads server-sent `offerVolume`/`bidVolume` (`internal/client/marketdata.go:597-606`), so on that path the same comparison would have been a real check — on the official path it can never be.

### Inputs (Full SDD order)

- **Memory recall:** "contract numbers from the receipt" (constants/enums/caps are quoted from a measurement sentence, never from memory — this brief obeys it literally below), "generated evidence must be measured" (the reason decision 32 exists), "FLM before claiming" (the five AST bundles below were produced *before* this brief, not deferred to an implementation task), "passing test is not evidence", token-war precondition (`tossos-shared-openapi-credential-token-war`), "boot is not a button".
- **OpenSpec:** L1 row `tasks.md:10` ("additive read-only official KR/US raw bar/**quote** adapter files"), task 3.6 (checked only after this lot), the quote rows of 3.7; `design.md:125` — *live quote는 ARMED 뒤 최종 spread/drift/freshness veto에만 사용하며 breakout/retest/reclaim transition의 근거가 될 수 없다* (the quote is a veto input, never a state-transition input); `design.md:172` (quote spread/freshness inside the breakout payload); decision 10 of the L1 brief (L1a defines `official_quote_l1:v1`; the producer fails closed until the level element schema is observed); decisions 12 (L1c gets its own brief) and 22 (frozen not-edited list).
- **Receipt (the numbers below are quoted from it, not remembered):** `analysis/measurements/m-b-us-source/receipt-2026-08-16-run4.json`, `observed_broker_behaviour.orderbook` — *"HTTP 200 {result:{timestamp 2026-08-14T16:59:58.000+09:00, currency USD, asks [], bids []}} — empty book, market closed; ask/bid level element schema and decimal encoding UNOBSERVED"*, and residual 1 — *"orderbook: empty book only — L1 must not infer quote level schema/typing/availability from this receipt"*. Three facts follow and one does not. Measured: the `result` object carries exactly `{timestamp, currency, asks, bids}`; `timestamp` was **present and non-null**, in the same `.000+09:00` KST-offset form as candles, for a **US** symbol; `currency` was `"USD"`. Not measured: everything inside a level, because the arrays were empty. Also never measured anywhere: `GET /api/v1/prices` raw bytes — the M-B seam has no descriptor for it (`a112_mbus_read.go:83-94` covers candles, orderbook, calendar only).
- **CodeGraph hard evidence (live index 34,752 nodes / 113,702 edges):** `NewQuoteL1Envelope` has exactly one caller and it is a test (`breakout_bar_test.go:699`) — the L1a quote contract is still unconsumed, so L1c writes its first production caller. `official.Client.Prices` has nine non-test callers including engine paths (`adoption.go:221`, `exitloop.go:743`, `tracer.go:353`, `position_policy_command.go:290`, `flatten/liquidate.go:372`, `soak.go:666`, `ops/read_operations.go:78`, `hybrid/client.go:117`, `verifylive/steps_trigger.go:743`); `official.Client.Orderbook` has three (`hybrid/client.go:133`, `verifylive/steps_trigger.go:743`, and the M-B collector). Both are load-bearing live reads — L1c adds siblings and edits neither. The official `parseDecimal` (`market_reads.go:15`) is called by all eleven `adapt*` functions in the package; it is the package-wide float boundary, not a local one.
- **AST artifacts produced before this brief** (citation-only bundles; none of these functions is edited by a112, and each carries `source_sha256 b356524b…` == the current `market_reads.go`): `analysis/function-logic/internal-official--adaptorderbook/` (2 range branches, 1 return, 12 calls; no reference to `raw.Timestamp`, no `time.Parse`, `time.Now().UTC()` at 258), `--adaptprices/` (1 range branch; same drop of `p.Timestamp`), `--parsedecimal/` (2 branches, 3 returns; both failure branches return `0` and the `strconv.ParseFloat` error is discarded), `--client.orderbook/` (1 branch, 2 returns; the request has exactly one query parameter — there is no depth/level parameter to ask with), `--client.prices/` (1 branch, 2 returns; an unknown symbol yields an empty slice and a nil error). `go test ./internal/official -count=1` green (351 tests, exit 0) on the same worktree.

### What the existing console path can and cannot see (why a probe alone cannot open this lot)

The 2026-08-18 human probe was read through `tossctl quote orderbook` → `hybrid.GetOrderBook` (`internal/hybrid/client.go:133-138`) → `official.Client.Orderbook` → `adaptOrderbook`. That path **drops the broker's `timestamp`** (`domain.OrderBook` has no such field; the AST shows the body never reads it, and `TestOrderbookIntegration` serves `"timestamp":"2026-03-25T09:30:00.123+09:00"` that no assertion can reach) and **drops `currency`**, and it converts every decimal through the fail-open `parseDecimal`. So no amount of console probing can report whether the field was present, what its precision was, or whether a level's price was `"284000"` or `284000`. `product_code: ""` remains a sound official-path fingerprint (asserted by `TestAdaptOrderbookUnit`; the WTS adapter always sets it, `marketdata.go:612-613`) — that inference stands. The rest must come from bytes.

### Manager decisions (binding for L1c; numbering continues)

33. **Scope is top-of-book, and that dissolves the depth question.** The evidence kind is `official_quote_l1` — L1 means one level. The producer reads `asks[0]` and `bids[0]` and refuses an empty side (`RefusalEmptySide`); it never counts, caps or interprets the rest of the ladder, and it stores no aggregate. The KR-10/US-1 asymmetry therefore needs no rule and no constant: a variable-length array is read at index 0 or refused. Anything beyond index 0 is out of scope for a112 (nothing in `design.md` consumes depth), and adding it would invent a cap we have not measured.
34. **One seal spans two endpoints.** `QuoteSealInput` requires `BidMinor`, `AskMinor` **and** `LastMinor` (`breakoutlane/types.go:117-130`), and no official endpoint returns all three: bid/ask come from `GET /api/v1/orderbook`, last from `GET /api/v1/prices` (`adaptPrices` is the only official source of a last price). One quote envelope therefore costs **two GETs**, issued back-to-back on the same client, and the producer never retries one half — a retry would widen the gap between the halves invisibly. Either half failing ⇒ typed refusal, zero envelopes.
35. **The broker instant is mandatory; the local clock may never stand in for it.** `validateQuote` (`breakoutlane/rules.go:32-45`) refuses on `evaluated − SourceObservedAtMS > MaxQuoteAgeMS`. If `source_observed_at_ms` were our read instant, that veto would be measuring our own clock and would pass by construction — the freshness gate would be decorative. Both responses must carry a non-null `timestamp` parsing under the existing `strictMinuteInstant` grammar (`2006-01-02T15:04:05[.000]±hh:mm`, no `Z`, no naive form); a null, absent or unparseable timestamp on either half is `RefusalNoSourceInstant` with zero envelopes. The receipt shows the field populated for a US symbol with a KST offset, so this is a contract we have seen honoured once, not a hope.
36. **Binding the two halves: conservative min/max, and no invented skew constant.** `source_observed_at_ms = min(t_orderbook, t_prices)` and `received_at_ms = max(readAt_orderbook, readAt_prices)`. Both choices move in the refusing direction: the older source instant and the later receipt instant can only make the quote look staler, never fresher. There is deliberately **no new skew-tolerance constant** — if the two reads are far apart, the older one ages the whole quote and the *existing* `MaxQuoteAgeMS` veto refuses it downstream. Inventing a millisecond bound here would be exactly the "contract numbers from memory" failure: we have measured no such number. `checkQuoteL1` already enforces `received_at_ms >= source_observed_at_ms`, which max/min satisfies by construction.
37. **`source_response_digest` covers both bodies.** The L1a payload is frozen with one digest field, so it is a composite over an ordered pair: `"sha256:" + hex(sha256("a112-quote-l1-v1\n" + <orderbook body sha256 hex> + "\n" + <prices body sha256 hex>))`. Order is fixed (orderbook first) and the prefix is versioned, so the value is reproducible from the two raw bodies and covers both. A digest over only one body would leave half the evidence unattributable.
38. **Strict readers (new files, no edits).** `internal/official/strict_quote_reads.go` (+ `_test.go`) adds `func (c *Client) StrictOrderbookTop(ctx context.Context, market, symbol string) (StrictTopOfBook, error)` and `func (c *Client) StrictLastPrice(ctx context.Context, market, symbol string) (StrictLastPrice, error)`, each one endpoint, each returning the raw decimal strings exactly as received, the broker instant, `ReadAt` (the `BodyReadComplete` of the last attempt), `StatusCode`, `BodyDigest` and the advisory `RateBudget` — the `StrictMinutePage` shape (decision 15). Strict decode is the L1b contract of decision 14 applied to these bodies: valid UTF-8, strict object at every depth, duplicate key at any depth ⇒ refuse, `result` keys exactly `{timestamp, currency, asks, bids}` for the orderbook and exactly `{symbol, lastPrice, currency, timestamp}` per row for prices, unknown key ⇒ refuse, every scalar a JSON **string** (bare number ⇒ refuse), `currency` equal to the market currency. The shape-generic helpers already in the package (`strictMinuteObject`, `strictMinuteCheckJSON`, `strictMinuteWalk`, `strictMinuteString`, `strictMinuteInstant`, `strictMinuteFinalAttempt`, `strictMinuteMarketCurrency`, `strictMinuteCheckSymbol`) are **called, not edited** — `strict_minute_candles.go` stays byte-identical (L1b accepted). The quote readers get their own error type and refuse helper so a quote refusal is never reported as a candle error. `/prices` is asked for exactly one symbol and the reader refuses a response whose row count is not 1 or whose `symbol` does not echo the request (`Client.Prices`'s silence on this, AST `--client.prices` fall-through, is the reason).
39. **The reader is the fail-closed instrument — there is no schema toggle.** Decision 10 required the producer to fail closed until the level element schema is observed. A reader with exactly one accepted shape *is* that: if `asks[0]` is not an object with exactly `{price, volume}` string values, it refuses and appends nothing. So L1c may be implemented now, and the human-approved live run becomes **acceptance evidence** rather than a precondition for writing code. This removes the flag that decision 10's wording would otherwise imply; a flag would have been a second place for the truth to live.
40. **Producer.** New `internal/officialbars/quote.go` (+ `quote_test.go`) — same package as the bar producer, whose import allowlist and constructor guard (decision 16) already anticipate `NewQuoteL1Envelope`. `func PollQuoteL1(ctx, reader QuoteReader, store QuoteStore, in QuotePollInput) (QuotePollResult, error)`; identity is the observation instant (`SourceRecordID = "<CODE>:<SYMBOL>:quote_l1:<source_observed_at_ms>"`, `newQuoteL1Header`), so **revision is always 1**: a re-read landing on the same instant with an identical digest is `Unchanged` (idempotent, no append), with a different digest it is a `Conflict` (quarantine, never overwritten) — the bar producer's rule, reached here by identity rather than by session. Prices/volumes are never converted to float in this package; `NewQuoteL1Envelope` recomputes minors from the raw strings by integer arithmetic at the market scale (KRW 0, USD 4), which also refuses over-precise US decimals — consistent with the receipt's *"variable decimal scale (1-4 dp, integer-form prices)"*.
41. **Rate and cadence.** Two GETs per quote against a quota shared with the live containers and with the neighbouring product's credential. The producer performs exactly one pair per call and never loops; cadence is L5's decision. `RateBudget` stays advisory and `X-RateLimit-Reset` stays opaque (M-B residual 4).
42. **Acceptance precondition (human-approved, read-only, ≤4 GETs).** One run per market **while that market is open**, so the ladder is not empty: KR 09:00–15:30 KST, US 22:30–05:00 KST. The run exercises the new readers (not `tossctl quote`, which cannot see what we need) and reports shape only — key names, JSON types, digit counts, level counts, the two instants and their difference — never prices as evidence and never a raw body in the repository. A fresh explicit human approval is required for each run; no agent reruns it. If a level's element schema differs from `{price, volume}` strings, the reader refuses, the run is still a success as a measurement, and this brief is amended before any code changes.
43. **Ownership / Pre-Edit declaration (High-risk adjacency: official client GET/token path).** Created: `internal/official/strict_quote_reads.go`, `internal/official/strict_quote_reads_test.go`, `internal/officialbars/quote.go`, `internal/officialbars/quote_test.go`. **Not edited:** `client.go`, `token.go`, `trace.go`, `ratebudget.go`, `candle_raw.go`, `candle_reads.go`, `market_reads.go`, `strict_minute_candles.go`, `producer.go`, `a112_mbus_*`, anything under `internal/strategyevidence/`, `internal/breakoutlane/`, `internal/hybrid/`, `internal/client/`, any `cmd/`, engine, router, scheduler, journal, toggle or container file. No production caller is wired: as with L1b, the producer has zero non-test callers when the lot lands.
44. **Residuals recorded now.** (a) The level element schema and its decimal encoding remain UNOBSERVED until decision 42's run — the readers encode a hypothesis taken from `openapi.latest.json` (`OrderbookEntry {price, volume}` decimal strings) and refuse anything else; that hypothesis is *documented*, not measured, and this brief says so rather than letting the code imply otherwise. (b) `/api/v1/prices` has never been observed at the byte level at all. (c) The nullability of both `timestamp` fields is documented as nullable and observed non-null exactly once (US, closed market); a null on either half is refused, so a broker that nulls it during an open session makes the quote producer silent by design — visible as refusals, never as fabricated freshness. (d) Two GETs per quote on a shared quota. (e) `Reset` header remains opaque. (f) The correction unit of decision 32 also implicates `docs/migration/openapi.latest.json`'s own comment lineage; that document is the broker's and is not ours to edit — the standing follow-up change covers our two copies.

### Brief amendment (2026-08-27, human-approved): decision 45

45. **The acceptance run needs a runner, and it is a tool, not a command.** Decision 43's created list is extended by
    exactly one directory: `tools/a112-l1c-quote-probe/` (`main.go`, `report.go`, `report_test.go`). Decision 42's run
    must exercise the new readers; decision 43 forbids `cmd/`, so `tossctl` cannot carry it; and the precedent for
    exactly this problem is a repository tool that is deliberately not installed into `tossctl`
    (`tools/a112-mb-us-source/main.go`). Constraints, enforced by what the code can reach rather than by a promise: it
    calls **only** `StrictOrderbookTop` and `StrictLastPrice` — no other client method appears in its source, and a
    static test asserts that; exactly **two GETs per invocation, one market per invocation**; `--market` and `--symbol`
    are required with no defaults, so it cannot run by accident; it prints **shape only** — acceptance or the typed
    refusal, the integer/fraction digit counts of every decimal, the two broker instants and their difference, the two
    read instants and their difference, both status codes and both body digests — and **never a price, never a volume,
    never a raw body**, with a test that fails if a fixture's price text appears in the rendered report. It writes no
    file and appends no evidence: a probe that wrote to the store would be a production caller, and decision 43 says
    this lot lands with none.
    **One line of decision 42 is deliberately not delivered:** the level count. Reporting it would require the reader
    to count the ladder, which decision 33 forbids. The 2026-08-18 console measurement (KR ten levels, US one) is the
    answer to that question and is recorded above; re-measuring it would mean weakening the reader to satisfy a report.
    A fresh human approval is still required per run, and no agent runs it.

### RED list for the Terra implementer (quote rows of task 3.7; each RED named and captured before GREEN)

Readers (`internal/official`, httptest + fake `/oauth2/token`, the `trace_test.go` pattern): market/symbol grammar refusals with zero server hits; exact query strings `symbol=005930` and `symbols=AAPL`; duplicate key at envelope/result/level depth ⇒ refuse; unknown `result` key, unknown level key, missing level key ⇒ refuse; bare-number price/volume/lastPrice ⇒ refuse; empty `asks` or `bids` ⇒ refuse (the M-B closed-market body is the fixture shape); `timestamp` null / absent / `Z`-suffixed / offset-less ⇒ refuse; `currency` ≠ market currency ⇒ refuse; `/prices` returning zero rows, two rows, or a row whose `symbol` does not echo ⇒ refuse; non-2xx last attempt ⇒ refuse (decision 28's rule, reused); `BodyDigest` equals `sha256` of the exact served bytes; the KR ten-level and US one-level bodies both accepted with only index 0 read.
Producer (`internal/officialbars`, fake readers + real temp `strategyevidence.Store`): bid > ask ⇒ refuse before any append (and `checkQuoteL1` would refuse anyway — assert both); either half erroring ⇒ zero appends and no partial state; `source_observed_at_ms` == min and `received_at_ms` == max, proven by a fixture where the two halves disagree; composite digest equals the declared preimage; same instant + same digest ⇒ `Unchanged`, no second append; same instant + different digest ⇒ `Conflict`/quarantine, first revision still readable; USD five-decimal raw ⇒ refuse (scale 4); KRW fractional raw ⇒ refuse (scale 0); no `float64` anywhere in the package's production files (AST guard); import allowlist guard extended to the new file.
GREEN gates: `go build ./...`; `go test ./internal/official ./internal/officialbars -count=1` and `-race`; `go vet`; `$(go env GOROOT)/bin/gofmt -l` empty; `go test ./internal/strategyevidence ./internal/breakoutlane -count=1` still green (untouched); `TestA112MBUS*` guards green; then the implementer's mutant table with `sha256sum -c` restoration proof (min↔max swapped; empty-side check dropped; timestamp-null refusal dropped; digest preimage reordered; row-echo check dropped; conflict path downgraded to overwrite) — "passing test is not evidence", and the revert baseline is the pre-mutation file hash, not `git checkout`.

## 2026-08-27 L1c implementation — official strict quote readers + top-of-book producer (Manager record)

### Ownership and authority boundary

- **Created, exactly decision 43's list and nothing else:** `internal/official/strict_quote_reads.go` (423 lines),
  `internal/official/strict_quote_reads_test.go` (357), `internal/officialbars/quote.go` (247),
  `internal/officialbars/quote_test.go` (508). `git status` at code-complete carries those four untracked files plus a
  one-line change to `base-commit.txt` (next section) — no other file in the repository moved.
- **Not edited, as declared:** `client.go`, `token.go`, `trace.go`, `ratebudget.go`, `candle_raw.go`, `candle_reads.go`,
  `market_reads.go`, `strict_minute_candles.go`, `producer.go`, `a112_mbus_*`, anything under `internal/strategyevidence`,
  `internal/breakoutlane`, `internal/hybrid`, `internal/client`, any `cmd/`, engine, router, scheduler, journal, toggle or
  container file. The shape helpers (`strictMinuteObject`, `strictMinuteCheckJSON`, `strictMinuteWalk`,
  `strictMinuteString`, `strictMinuteInstant`, `strictMinuteFinalAttempt`, `strictMinuteMarketCurrency`,
  `strictMinuteCheckSymbol`) and the caps (`strictMinuteMaxBody`, `strictMinuteMaxDecimal`) are **called**;
  `strict_minute_candles.go` is byte-identical to its L1b state.
- **Pre-Edit (High-risk adjacency: official client GET/token path):** no existing function body was edited, so no
  existing symbol needed a Pre-Edit gate. The new readers reach the broker only through the unchanged `c.get` with a
  per-call attempt observer that **chains** an outer observer instead of replacing it (the `StrictMinuteCandles`
  pattern, so an M-B style outer observation is never silently dropped).
- **No production caller is wired.** As with L1b, `PollQuoteL1` and both readers have zero non-test callers when this
  lot lands; nothing about the running containers changes.

### Baseline re-freeze (procedure, not a new decision)

The lot opened with `check_analysis` **failing** on five functions that belong to **a117**
(`internal/strategymarket/bars.go:aggregateClosedKRXFiveMinute` plus four test helpers in `bars_test.go`,
`strategycandle/adapter_test.go`, `strategyengine/lane_test.go`): a117 landed on `main` after a112's frozen base
`d9d9f71f`, so a112's diff window had swallowed another change's edits. Following the 2026-08-17 landing precedent
recorded in `analysis/current-main-evidence.md`, `base-commit.txt` was re-frozen to `a8c3d067` (branch and `main` HEAD).
Proven before re-freezing: the intervening commits cite no bundle source — `73a1e85c` is openspec-only, and `a8c3d067`
touches `internal/official/candle_reads.go` (comment), `internal/strategymarket/*`, `internal/strategycandle/adapter_test.go`,
`internal/strategyengine/lane_test.go` and PM files, none of which is the `file` of any bundle in
`analysis/function-logic/`. After the re-freeze `check_analysis` reports "evidence complete or diff-proven exempt" and
the L1c diff contains only this lot's own files.

### What was built

**Readers (`internal/official`).** `(*Client).StrictOrderbookTop(ctx, market, symbol) (StrictTopOfBook, error)` and
`(*Client).StrictLastPrice(ctx, market, symbol) (StrictLastPrice, error)`, each exactly one GET, each returning the raw
decimal strings as received, the broker instant (raw text *and* parsed), `ReadAt` (`BodyReadComplete` of the last
attempt), `StatusCode`, `BodyDigest` over the exact served bytes and the advisory `Budget` — the `StrictMinutePage`
shape of decision 15. `StrictQuoteError{Reason, Detail}` is a distinct type with its own reason vocabulary
(`EMPTY_SIDE`, `NO_SOURCE_INSTANT`, `SYMBOL_ECHO_MISMATCH`, `ROW_COUNT_INVALID`, …); `strictQuoteAdopt` re-types a
shape-helper refusal into a quote refusal so the grammar keeps **one** authority while a quote failure is never
reported as a candle failure (`TestStrictQuoteRefusalsAreNeverCandleRefusals` pins both halves of that sentence).

**Producer (`internal/officialbars`).** `PollQuoteL1(ctx, reader QuoteReader, store QuoteStore, in QuotePollInput)`
reads the orderbook half first, then the prices half, checks that both answer for the requested market/symbol and agree
on currency, refuses a missing instant on either half, binds them conservatively, composes the digest, builds the
envelope through `strategyevidence.NewQuoteL1Envelope` and appends once. `QuotePollResult.Outcome` is exactly one of
`ADMITTED` / `UNCHANGED` / `CONFLICT`.

### Decisions honoured

- **33 (top of book only).** `strictQuoteSideTop` reads index 0 and refuses an empty side; no level count, cap or
  aggregate exists anywhere in either file. The KR ten-level and US one-level bodies are both accepted, and the same
  test asserts only index 0 reached the result.
- **34 (one seal, two endpoints, no retry).** The reader interface has exactly two methods; `PollQuoteL1` calls each at
  most once. `TestPollQuoteL1AppendsNothingWhenEitherHalfFails` asserts zero appends, the reader's own error preserved
  through `errors.Is`, **and** the call ledger (`orderbook` alone when the first half fails; never a repeat).
- **35 (the broker instant is mandatory).** `strictQuoteInstant` demands a JSON string parsing under the existing
  `strictMinuteInstant` grammar; null, absent, `Z`-suffixed, offset-less and impossible dates are all refused, and the
  producer refuses a zero instant or a zero read instant on either half. The local clock is never substituted — neither
  file calls `time.Now`.
- **36 (min/max, no skew constant).** `earlier(...)`/`later(...)`; a fixture where the halves disagree in both
  directions proves observed = older source instant and received = later read instant. No millisecond tolerance
  constant was introduced.
- **37 (composite digest).** `sha256("a112-quote-l1-v1\n" + <orderbook hex> + "\n" + <prices hex>)`, orderbook first.
  The test re-derives the preimage from the written rule rather than calling the production helper, and a malformed
  half-digest is a typed refusal (`QUOTE_DIGEST_INVALID`) before any append.
- **38 (strict readers, new files).** `result` keys exactly `{timestamp, currency, asks, bids}`; level keys exactly
  `{price, volume}`; price rows exactly `{symbol, lastPrice, currency, timestamp}`; every scalar a JSON string (a bare
  number is refused); duplicate keys refused at envelope, result and level depth; `currency` must equal the market
  currency; `/prices` is asked for exactly one symbol and a response with zero rows, two rows or a row that does not
  echo the symbol is refused (the fall-through silence of `Client.Prices` is exactly why).
- **39 (the reader is the fail-closed instrument).** No schema flag exists. One accepted shape, refusal otherwise.
- **40 (producer, revision always 1).** `SourceRecordID` is `<CODE>:<SYMBOL>:quote_l1:<source_observed_at_ms>` via
  `newQuoteL1Header`; `quoteRevision = 1`. **Implementation note:** the producer does not pre-read the store. Identity
  collisions are judged by `Store.Append` itself — `AppendResult.Idempotent` ⇒ `UNCHANGED`, `ErrRevisionConflict` ⇒
  `CONFLICT` with the first revision left readable and the attempt quarantined. That is the same rule the brief states,
  reached without adding a read API to `internal/strategyevidence` (which decision 43 forbids editing). Its one
  behavioural consequence is recorded as residual (g) below.
- **41 (rate and cadence).** Exactly one pair of GETs per call, no loop, no cadence; `Budget` stays advisory.
- **44 (residuals).** Carried forward and extended below.

### RED → GREEN

1. **RED (readers).** `strict_quote_reads_test.go` written first; `go vet ./internal/official` failed with
   `undefined: StrictQuoteError`, `undefined: StrictQuoteLevel`, `*Client has no field or method StrictOrderbookTop`.
2. **GREEN (readers).** `go test ./internal/official -run 'StrictQuote|StrictOrderbookTop|StrictLastPrice' -count=1`
   → **61** tests/subtests, exit 0.
3. **RED (producer).** `quote_test.go` written next; `go vet ./internal/officialbars` failed with `undefined: QuoteStore`.
4. **GREEN (producer).** `go test ./internal/officialbars -run 'TestPollQuoteL1|TestProductionNeverTouchesFloatingPoint'
   -v -count=1` → **29** tests/subtests, exit 0. One of them is not in the brief's RED list and was added during the
   self-review pass: a broker instant **after** our read instant (clock skew) is refused rather than clamped, because
   clamping would make a quote that was "observed after it was received" look ordinary and would loosen decision 35's
   freshness gate by exactly that much.
5. **Package gates.** `go build ./...` clean; `go test ./internal/official ./internal/officialbars -count=1` and the same
   with `-race` → **497** tests/subtests, exit 0; `go test ./internal/strategyevidence ./internal/breakoutlane
   ./internal/strategymarket ./internal/strategycandle -count=1` → **380**, exit 0 (untouched packages still green);
   `$(go env GOROOT)/bin/gofmt -l` empty on all four new files; `go vet` clean.
6. **Whole repository, after the commit.** `go test ./... -count=1` → **93 packages ok, 0 FAIL, exit 0**;
   `make vet` exit 0; `$(go env GOROOT)/bin/gofmt -l internal cmd tools` empty;
   `openspec validate --strict` valid; `check_analysis` "evidence complete or diff-proven exempt";
   `python3 tools/pm/generate_master_tracker.py --check` current; `make sdd-sync` clean (GBrain advisory busy →
   previous freshness kept, WARN only; the first attempt died on the codegraphcontext 15s advisory timeout and was
   retried); `make sdd-check` exit 0 with "CodeGraph hard-evidence index matches the worktree".
7. **Static guards.** The package-wide import allowlist and combined-constructor guards already glob `*.go`, so they
   cover `quote.go` without editing `guard_test.go`; the new float guard
   (`TestProductionNeverTouchesFloatingPoint`) lives in the created test file, walks every production file's AST and
   fails on `float64`, `float32`, `ParseFloat` or `FormatFloat` — and asserts it inspected a non-zero number of
   identifiers, so "the guard checked nothing" cannot pass as "the guard found nothing"
   (`generated-evidence-must-be-measured`).

### Mutation table (six mutants, `sha256` restoration proof)

Each mutant was applied to a production file, the named test run, the file restored from its pre-mutation bytes, and the
restored file's SHA-256 compared with the baseline hash — not `git checkout`
(`mutation-revert-needs-the-right-baseline`).

| Mutant | File | Test that must fail | Result |
| --- | --- | --- | --- |
| `earlier`/`later` swapped | `quote.go` | `TestPollQuoteL1BindsTheHalvesConservatively` | KILLED |
| empty-side refusal returns a zero level | `strict_quote_reads.go` | `TestStrictOrderbookTopRefusesBodiesOutsideTheContract` | KILLED |
| missing source instant accepted | `strict_quote_reads.go` | `TestStrictOrderbookTopRefusesBodiesOutsideTheContract` | KILLED |
| digest preimage reordered | `quote.go` | `TestPollQuoteL1SealsTwoReadsIntoOneRow` | KILLED |
| row-echo check disabled | `strict_quote_reads.go` | `TestStrictLastPriceRefusesBodiesOutsideTheContract` | KILLED |
| conflict reported as admitted | `quote.go` | `TestPollQuoteL1QuarantinesADifferentBodyAtTheSameInstant` | KILLED |

All six killed; both files restored to their pre-mutation hashes.

### Function Logic Map: not-applicable

L1c adds only **new files containing new leaf functions**. No function that existed at the frozen base `a8c3d067`
changed, which `check_analysis` computes directly (`changed_existing_functions` requires a bundle only for a function
present in the base file whose body a hunk touches). The five *citation-only* bundles for the existing functions this
lot's brief reasons about (`adaptOrderbook`, `adaptPrices`, `parseDecimal`, `Client.Orderbook`, `Client.Prices`) were
produced **before** the brief at `73a1e85c` and are unchanged. The exemption is declared here rather than left silent.

### Gate practice learned here: `go test ./...` must run *after* committing a new file in `internal/official`

The first full-suite run failed in `tools/a112-mb-us-source`
(`TestTrimpathBuiltCollectorCompletesOfflineIdentityPreflightBeforeReader`,
`TestTrimpathBuiltCollectorProcessHoldsBeforeExternalReadOnMissingCache`) with
`a112 m-b1 measurement HOLD: pre-request identity: compiled input is not approved`. That is the M-B collector's
executable-identity gate doing exactly its job: `validateCompiledInputForGOOSWith` (`identity.go:347-361`) accepts a
compiled input only if it is **git-tracked** or on the frozen untracked allowlist, and the collector compiles the whole
of `internal/official`. While `strict_quote_reads.go` sat untracked, every M-B identity preflight refused. After the
commit the same package is green (76 tests, exit 0). Nothing in the collector or the allowlist was changed. The lesson
for later lots: adding a file to `internal/official` makes `go test ./...` fail until that file is committed, and the
failure text points at identity, not at the new file — so read it as "untracked", not as "broken".

### Residuals

Carrying (a)–(f) from decision 44 unchanged, and adding:

- **(g) A same-instant re-read with different bytes quarantines rather than de-duplicating.** Because the producer does
  not pre-read the store, two reads that land on the same broker instant with different bodies produce a `CONFLICT`
  (nothing overwritten, the attempt quarantined, the first revision readable). Two reads that land on the same instant
  with *identical* bodies **and** identical read instants are `UNCHANGED`. In production the read instants will differ,
  so a repeated identical broker snapshot inside one instant granularity would be counted as a conflict, not as an
  idempotent no-op. The direction is safe (nothing is overwritten and nothing is fabricated) and the signal is visible,
  but if L5 chooses a cadence finer than the broker's timestamp granularity it must expect conflict rows. Deciding that
  is L5's, not this lot's.
- **(h) "Index 0 is the best level" is measured, not contracted.** `asks[0]`/`bids[0]` were the best ask and best bid in
  the 2026-08-18 KR probe (asks ascending 284,000→288,500, bids descending 283,500→279,000) and trivially so in the
  one-level US book; `openapi.latest.json`'s *schema descriptions* agree ("낮은 가격순" / "높은 가격순") while its
  *examples* show asks worst-first. The reader does not verify the ordering, because verifying it would require parsing
  every level's decimal inside the transport layer — and decimals have exactly one authority
  (`strategyevidence`). If the broker ever served worst-first, the quote would read as a **wider** spread, which the
  existing spread veto refuses; it cannot make a quote look tighter than it is. Decision 42's run should report the
  first two prices per side so this stays measured.
- **(i) There is no way to run decision 42's probe yet.** The run must exercise the new readers, and decision 43
  forbids touching `cmd/`. The M-B precedent for exactly this problem is a dedicated tool
  (`tools/a112-mb-us-source/main.go`). Adding `tools/a112-l1c-quote-probe/` is therefore the obvious shape, but it is a
  **fifth created file** that decision 43's list does not carry, so it needs a one-line brief amendment (or an explicit
  human decision to run it as an untracked one-shot) before it is written. No agent runs it either way — decision 42
  requires a fresh human approval per run.
- **(j) The new evidence kind reaches one unfiltered reader.** `SealBarSeries` filters on `evidence_kind`
  (`breakout_series.go:73`), so quote rows can never contaminate a bar series. `SealSnapshot` does **not** filter by
  kind — it returns every envelope for `(market, symbol, issuer)`. It has zero production callers today (grep over
  `internal/` and `cmd/`: the only hit is its own definition), so nothing changes now; but whoever adopts it after L5
  wires the quote producer will start receiving `official_quote_l1` items next to bars and must dispatch on
  `Header.Kind`. Recorded here so that arrives as a known property rather than a surprise.

### Runner built under decision 45 (2026-08-27)

`tools/a112-l1c-quote-probe/` — `main.go` (77 lines), `report.go` (152), `report_test.go` (196). RED first
(`undefined: observation`), then GREEN: **15 tests**, exit 0.

- `run()` refuses before any request unless `--market` and `--symbol` are both given; it then resolves the ordinary
  Open API paths, builds the ordinary client, and calls `StrictOrderbookTop` then `StrictLastPrice` in the producer's
  order. A failed half is not retried and stops the run.
- `renderReport` prints digit counts, instants, differences, status codes and digests, plus the seal-binding decision
  the producer would make and a scale verdict derived from the digit counts (four fraction digits fit USD scale 4; a
  fifth would be refused). It never prints a price, a volume or a body.
- Three static guards, each mutation-proven: `TestReportNeverCarriesAValue` (mutant that prints `row.raw` → KILLED),
  `TestProbeReachesOnlyTheTwoStrictReaders` (mutant adding `client.Accounts(ctx)` → KILLED; the same guard's text check
  also fired for real during development, on a *comment* that spelled the evidence package, and the comment was
  reworded rather than the guard weakened), `TestScaleVerdictFollowsTheMarketScale` (mutant loosening the scale
  comparison by one → KILLED). All three files restored to their pre-mutation hashes.
- The probe writes no file and appends no evidence, so this lot still lands with **zero production callers** of
  `PollQuoteL1`.

Run sheet (human, one fresh approval per run, market must be open):

```
go run ./tools/a112-l1c-quote-probe --market US --symbol AAPL      # 22:30–05:00 KST
go run ./tools/a112-l1c-quote-probe --market KR --symbol 005930    # 09:00–15:30 KST
```

### What remains before L1c is ACCEPTED

1. The decision-42 acceptance run, one per market while that market is open (KR 09:00–15:30 KST, US 22:30–05:00 KST),
   reporting shape only — key names, JSON types, digit counts, level counts, the two instants and their difference. No
   prices as evidence, no raw body in the repository.
2. ~~The runner decision in residual (i)~~ — **closed** by decision 45 (2026-08-27). The runner exists and has
   been run (2026-08-28, decision 46), but against two *closed* markets, so item 1 above is still open.
3. Then, and only then, task 3.6 is checked (decision 12) and this section is closed as ACCEPTED. `tasks.md` is
   deliberately unchanged by this landing, exactly as at L1b: 3.7 is a single atomic row that also demands the
   timezone/restart/correction/cache-miss replay fixtures and broker mutation spies at zero, which this lot does not
   deliver.

### Brief amendment (2026-08-28, human instruction): decision 46

46. **Decision 42's "no agent reruns it" clause is lifted by the user; its read-only ceiling is not.** The user
    instructed the agent to run the probe without a per-run approval ("승인없이 너가 돌려", 2026-08-28). That clause of
    decision 42 is therefore overridden for this change, and this record says so rather than pretending the original
    clause was honoured. Everything else in decision 42 stands unchanged and was in fact obeyed: the runner reaches only
    the two strict readers, each invocation is exactly two GETs of one market, nothing is written, and no price, volume
    or raw body reaches the repository. Four GETs total were spent (two markets × two endpoints).

### 2026-08-28 L1c probe runs — both markets CLOSED, so this is measurement, not acceptance

Run at 05:29–05:31 KST by the agent under decision 46. **Neither market was open**: US regular hours had ended at 05:00
KST (the book was in the after-hours session) and KR opens at 09:00 KST (the book was yesterday's close). Decision 42's
acceptance precondition is an **open** market, so these runs do **not** close acceptance and task 3.6 stays unchecked.

| | US / AAPL | KR / 005930 |
|---|---|---|
| `/orderbook` | HTTP 200, `sha256:43660d9f…38e433` | HTTP 200, `sha256:1cdc8b06…eed3b` |
| `/prices` | HTTP 200, `sha256:200ba802…d52fa8` | HTTP 200, `sha256:4f99162d…1af2a` |
| broker instant, orderbook | `2026-08-28T05:28:58.000+09:00` | `2026-08-27T20:00:00.000+09:00` |
| broker instant, prices | `2026-08-28T05:28:30.000+09:00` | `2026-08-27T19:59:59.000+09:00` |
| broker instant difference | **28s** | 1s |
| read instant difference | 73.6ms | 72.3ms |
| ask.price digits (int/frac) | 3 / **2** | 6 / 0 |
| bid.price digits (int/frac) | 3 / **1** | 6 / 0 |
| ask.volume / bid.volume digits | 2/0, 2/0 | 6/0, 5/0 |
| last digits | 3 / 2 | 6 / 0 |
| scale verdict | fits USD scale 4 | fits KRW scale 0 |

**What these runs settle.**

- **Residual (b) is closed.** `/api/v1/prices` had never been observed at the byte level in either market. It has now
  been observed four times over two markets and returned the exact key set the strict reader demands each time — any
  other key set, a duplicate key, a non-string scalar or a row count ≠ 1 would have been a typed refusal, not a 200.
- **Residual (a) is closed for index 0.** The level element schema hypothesis taken from `openapi.latest.json`
  (`{price, volume}` as decimal strings) is now *measured* rather than documented, in both markets. It remains
  unmeasured for indices ≥ 1, which the reader does not read (decision 33).
- **Residual (c) is closed.** Both `timestamp` fields were non-null in all four bodies, in both markets.
- **Decimals are trimmed, not zero-padded.** The single US orderbook body carried `ask.price` with two fraction digits
  and `bid.price` with one. Any consumer assuming a fixed decimal width for a market is wrong. The producer does not
  assume one — its scale rule is an upper bound, which both markets satisfied.
- **Decision 36's conservative binding did visible work.** On the US pair the two halves were **28 seconds** apart and
  the binding took the *older* broker instant as `source_observed_at` and the *later* read as `received_at`, widening
  the recorded age instead of narrowing it. In both samples the last-price half happened to win both ends; that is an
  artefact of these two samples (it is read second, and here it also carried the earlier instant), not a rule.

**What they do not settle.** Open-session behaviour: a US regular-session price may carry more fraction digits than the
two seen after hours (the scale-4 ceiling is untested against a value that approaches it), and level counts are still
the 2026-08-18 console figures. Residual (h) is *unreachable through this runner by construction* — it asks for the
first two prices per side, and decision 33 confines the reader to index 0, so no amount of probe running can close it;
closing it needs either a reader change or a separate one-off observation, and neither is this lot's.

- **(k) NEW — nothing in the quote path bounds a quote's age, and the KR run measured it.** The KR bodies were stamped
  `2026-08-27T20:00:00+09:00` and were read at 2026-08-28 05:30 KST — **9h30m old** — and both strict readers returned
  success. `PollQuoteL1`'s instant checks assert *presence* only (`quote.go:137-143`, "… is missing"), and its refusal
  set is reader-error / identity / currency / instant-missing / digest / envelope / store-error with **no age rule**;
  the envelope refuses `received_at` *before* `observed_at` (clock skew) but not `received_at` long *after* it. So the
  producer would have admitted a nine-and-a-half-hour-old closed-market book as an `ADMITTED` quote carrying a fresh
  `received_at`. This is harmless today — the lot lands with zero production callers — but it is now a *measured* gap,
  not a suspected one: whoever wires the producer in L5 must add a freshness gate first, or a strategy can act on a
  book from a market that has been shut for nine hours.

## L3 — canonical contracts and RouteSet

### Brief amendment (2026-08-28): decision 47

47. **`RouteSet` is a new file in `internal/strategyrouter`, and `router.go` stays untouched.** Task 4.3.1 requires a
    sealed RouteSet authority that emits *every* eligible family candidate, and task 4.3.2 requires that legacy `Route`
    and its callers outside the exact production closure remain **behaviourally unchanged**. Those two requirements
    together forbid editing `Route`'s body. The ownership ledger grants L3 exactly one file in this package
    (`production.go`), so the created list is extended by exactly two files: `internal/strategyrouter/routeset.go` and
    `internal/strategyrouter/routeset_test.go`. `router.go` is not edited — not one line — which is what makes "Route is
    unchanged" a compile-level fact rather than a promise. `RouteSet` re-derives the same prelude validation from the
    same exported inputs rather than calling `Route`, because calling it would re-impose the single-winner selection
    this lot exists to remove.

### Decision 48 — what the breakout binding may take from where

48. **`internal/breakoutlane` is L2's and is not edited; the production envelope the composition needs is built in
    `strategyflow`.** The other three families each expose a `BuildProduction<Market>ProposalAuthority` in their own
    lane package, so `strategyproposal` hands them a manifest scope and receives a sealed request. `breakoutlane` has
    no such builder and no exported accessor on `EvidenceSnapshot`, and it is not L3's to change. Therefore
    `strategyflow` owns `BreakoutRequest`, which carries the pure core's `EvidenceInput` alongside the lineage envelope
    (account, candidate, campaign, leg ordinal, planned ceiling, risk-budget digest), the price provenance envelope and
    the execution-policy inputs — every one of which the pure core does not and must not produce. `LaneRelease` for
    this family is a `strategyflow` constant (`BreakoutRelease`), because the frozen golden's descriptor record has no
    `release` member and inventing one inside L2's package would be an unauthorised edit.
    **q_candidate versus q_final:** `breakoutlane.Decision` exposes both. `Propose` is documented as the cap-free
    q_candidate composition, so the proposal adapter reads `CandidateQuantity()`; `Evaluate` reads `FinalQuantity()`,
    the same value capped by the manifest's `FinalCap`. This adds no cap and relaxes none: `q_final = min(...)` remains
    a066's, and the golden's `q_final_must_not_exceed_q_candidate` is preserved because `FinalCap` can only lower.

### Decision 49 — why the breakout lane fails closed instead of building an input

49. **The pure core requires four derived per-bar/per-snapshot members that no evidence record stores and no code
    produces. L3 closes the lane rather than inventing them.** This was found while wiring task 4.4 and is recorded
    here because `internal/strategyproposal/production.go:42-46` cites it as its refusal reason.

    **What the core demands (measured).** `breakoutlane.ClosedBarInput` (`types.go:104-106`) carries `RVOLPPM`,
    `UpperWickRangePPM` and `VolumeExpanded`; `EvidenceInput` (`types.go:167-177`) carries `ATRMinor`, and
    `validStructuralEvidenceInput` refuses outright when `ATRMinor == 0` (`machine.go:141`). The machine reads all
    four and none is decorative: RVOL and wick together gate the breakout admission (`machine.go:57`), RVOL sets the
    1.2/2.0/2.5 counterfactual flags (`:60-62`), `VolumeExpanded` gates the retest-failure path (`:96`), and ATR sets
    both the close buffer (`:54`) and the retest tolerance (`:102`). All four are inside the sealed digests
    (`arithmetic.go:83`, `machine.go:198`), so none may be quietly defaulted to zero — a zero changes the digest.

    **What L1 stores (measured).** `strategyevidence.ClosedBar1mPayload` (`breakout_bar.go:51-75`) holds
    `OpenMinor/HighMinor/LowMinor/CloseMinor/Volume` plus identity, finality, revision and provenance members — and no
    RVOL, wick, ATR or volume-expansion member. `grep -rni 'atr|rvol|wick' internal/strategyevidence/` returns only
    `driftAtRead` and `TestOfficialAuthorityMarketMatrix`: substring collisions, **zero real matches**.

    **Who produces them (measured).** Nobody. Outside `internal/breakoutlane` the only assignments to any of the four
    in the whole tree are in L3's own fixture (`internal/strategyflow/breakout_binding_test.go:55,117`); inside the
    package every assignment is a test literal (`evaluator_test.go:17`, `final_redteam_test.go:110`). No test, spec or
    config pins a formula for any of them.

    **What is derivable and what is not.** `UpperWickRangePPM` is arithmetic over stored O/H/L/C and could be derived
    *once the formula is chosen* — but nothing fixes it, and choosing one here would invent the semantics of a frozen
    threshold (`v1UpperWickRangeMaxPPM = 350_000`, `types.go:27`). `RVOLPPM` and `VolumeExpanded` need a volume
    baseline and `ATRMinor` needs a window; neither the baseline nor the period appears in `V1Config` (`types.go:83`),
    in `design.md`, or in `openspec/specs/`. `BarSeriesQuery` takes exactly one `SessionID`
    (`breakout_series.go:23-32`), so any cross-session baseline is N separate reads at the store's own measured cost
    of 47ms (one revision) / 159ms (four revisions) per 390-bar session (`breakout_series.go:51-52`). None of that is
    a detail `buildLaneInput` can settle.

    **Why inventing a value would break the design, not merely approximate it.** `design.md:174` states that the
    rolling-window/ATR/RVOL cache "는 성능 최적화일 뿐 권위가 아니다. cache miss/restart 때 append-only evidence로
    동일 결과를 재생해야 한다." A number computed at composition time from an unrecorded rule is not replayable from
    append-only evidence, so a fabricated ATR or RVOL would violate the design's own authority rule and would do so
    *invisibly*, inside a sealed digest.

    **Therefore** `buildLaneInput` returns `ErrBreakoutEvidenceUnavailable` for `breakoutlane.KRLaneID`/`USLaneID`
    (`production.go:328-330`), and the sentinel names all four members rather than three. This fails closed: the
    scope is skipped, no proposal is emitted, and nothing downstream sees a partially-invented snapshot. Defining and
    storing the four members is an evidence-schema change — L1's, not L3's composition — so it is recorded as a
    blocker here instead of being worked around. Tasks 4.1/4.2/4.3/4.3.1/4.3.2 are unaffected and land complete; the
    disposition of 4.4's breakout branch and of 4.5's breakout half remains open pending that lot.

### Decision 50 — the frozen 8-set breaks three paired-lane assertions outside L3's ledger

50. **L3's ownership ledger is extended by exactly three test files, because growing `strategyflow.Descriptors()`
    from six to eight is a compile-clean change that only a *tagged* test run can catch, and it breaks assertions
    those three files hard-code.** The ledger grants L3 `internal/strategyflow/**`,
    `internal/strategyrouter/production.go`, `internal/strategyproposal/**`,
    `internal/app/engine/strategy_route_authority.go`, `strategy_proposal_authority.go` and their tests. It is
    extended to also permit:

    - `internal/app/engine/strategy_first_leg_admission_test.go`
    - `internal/execgw/riskguardian_account_base_testseam_test.go`
    - `internal/journal/strategyflow_projection_test.go`

    **Why they break (measured).** Each iterates `strategyflow.Descriptors()` and then asserts a literal count:
    `len(descriptors) != 6` (`riskguardian_account_base_testseam_test.go:100`,
    `strategyflow_projection_test.go:18`) and `markets["KR"] != 3 || markets["US"] != 3`
    (`strategy_first_leg_admission_test.go:54`, plus the same pair in the other two). Every *subtest* passes,
    including both new breakout lanes — only the count assertions fail. The edits are therefore mechanical: 6→8,
    3→4, and the names that say `AllSix`/`ThreeLanes` are corrected, since a test named for six lanes while running
    eight is a false claim whether or not it is green.

    **Why this was not visible earlier — and what that says about the gate.** `make test` is `go test ./...`
    (`Makefile:36`) and CI runs that target (`.github/workflows/ci.yml:30-31`). Neither passes
    `-tags tossos_testseams`, so **no test behind that build tag has ever run in CI**. Measured on this tree:
    `internal/strategyproposal` runs **0** tests untagged and 6 tagged; `internal/strategyflow` 84 → 91;
    `internal/app/engine` 602 → 647. The whole of `internal/strategyproposal`'s test suite — the package L3's
    production wiring lives in — is invisible to the repository's own gate. This lot's first untagged run was
    green while eight tagged tests were failing, which is the "passing test is not evidence" failure in its
    ordinary, undramatic form: the suite that passed was not the suite that mattered.

    **A separate defect this exposed, now fixed.** `ProductionBatchAuthorityForTest`
    (`internal/strategyproposal/production_testseam.go`) sealed its map with the bare symbol while task 4.3.1
    changed the production key to `batchKey(symbol, laneID)`. `For` then found nothing and every engine caller saw
    `NO_ACCEPTED_PROPOSAL`. Seven of the eight tagged failures were that one key. The seam now calls the same
    `batchKey` the production path calls, so the two cannot drift apart again by construction. The file is inside
    the ledger's `internal/strategyproposal/**`, so no extension was needed for it.

### L3 landing state (2026-08-28)

**Ticked: 4.1, 4.2, 4.3.1, 4.3.2.** The 8-descriptor matrix, the sealed `LaneInput` union with both breakout
variants and their two registries, the all-candidate `RouteSet` replacing pre-evaluation selection, and the
AST/import-resolution guard over `strategy_route_authority.go`, `strategy_proposal_authority.go`,
`strategy_entry_supervisor.go` plus the `strategy_*coordinator*.go` glob.

**Not ticked, and why — three different reasons, none of them "ran out of time".**

- **4.3** — the descriptor table and the exact candidate validation *are* four families per market
  (`productionRouteDescriptors`, `validProductionRouteCandidates`), and a legacy three-family manifest is refused
  by the count equality. What is missing is the rest of the task's own sentence: **family/scoring/calibration
  seals**. Measured: `productionRouteCandidate` carries a pre-existing `Score` (`production.go:59`) that legacy
  `Route` uses for single-winner selection (`router.go:276-278`), `RouteDecision` does not carry it at all, and
  `grep -rn "Calibrat"` over `internal/strategyrouter` returns nothing. `RouteSet`'s seal is a decision-*set*
  seal (`routeset.go:153-176`), not a calibration seal. Ticking 4.3 on the two-thirds that are done would claim
  a seal that does not exist.
- **4.4** — scope validation is extended (`validScopes` now keys on (symbol, laneID)), but `buildLaneInput`'s
  breakout construction is blocked by decision 49 and fails closed instead.
- **4.5** — not started. Note that its "all 8 descriptors" is reachable for route and proposal lineage but not
  for a breakout *production* input, for the same reason as 4.4.

**Measured verification.**

| Command | Result |
|---|---|
| `make lint` (gofmt + `go vet ./...`) | clean |
| `go test ./...` (untagged, the CI target) | all packages ok; `internal/journal` 361s |
| `go test -tags tossos_testseams ./internal/{strategyflow,strategyproposal,strategyrouter,app/engine,execgw,riskbucket}/...` | ok |
| `go test -tags tossos_testseams` over `internal/journal`'s strategyflow projection tests | ok |
| `python3 tools/logic-map/check_analysis.py --change a112-…` | evidence complete |

**What the Function Logic Maps record that a green suite does not.** 44 bundles were regenerated from measured
coverage profiles rather than authored dispositions, and they record **46 branch arms that no measured profile
ever enters** — concentrated in `buildLaneInput` (18), `LoadProductionAuthorityBatch` (14),
`strategyRouteAuthorityLoader.collectMarket` (7) and `evaluateWith` (4). Those are refusal arms on the
production proposal path. They are written down as gaps, not smoothed over: `go test` does not instrument
`_test.go` files, so test-function bundles carry an arm classification and the run that exercised them instead
of a fabricated coverage number.

### L3 review finding C4–C6 closed by mutation-verified regression tests (2026-08-29)

The independent review of commit `3343e23d` reported three CRITICAL test-debt findings. Each was reported as
CRITICAL because a **mutation of production code survived the suite** — not because the production code was
wrong. All three are now closed by untagged tests, and each closure is proved the same way it was opened: by
mutating production code and observing a named test fail.

**Why untagged.** `make test` and CI run `go test ./...` with no `-tags tossos_testseams`
(`Makefile:36`, `.github/workflows/ci.yml:30-31`). `internal/strategyproposal`'s only test file carries that
tag, so the package ran **0 tests in CI**. Measured before and after, with `rtk proxy` so the runner's summary
does not swallow `-v`:

| package (untagged) | before | after |
|---|---|---|
| `internal/strategyproposal` | **0** | 7 |
| `internal/strategyflow` | 84 | 89 |

**New files, all without a build tag.**

- `internal/strategyproposal/breakout_fail_closed_test.go` — the breakout refusal, plus a differential case
  proving a non-breakout lane with the same empty scope is refused for a *different* reason. Without the
  second case the first proves only "some error", not "the breakout lock".
- `internal/strategyproposal/batch_authority_test.go` — `For` refusing a symbol with two lanes, `For`
  returning the single lane, lane-order determinism over 64 rounds (Go randomises map iteration, so dropping
  the sort fails this with probability ≈ 1), and the `\x00` key separator preventing `AAP` from reaching
  `AAPL`.
- `internal/strategyflow/route_decision_selection_test.go` — `selectRouteDecision` with **more than one
  decision**, which is the entire point of the change: the requested lane found at position 3, a non-canonical
  decision skipped rather than terminating the scan, the no-match fallback keeping the first canonical
  decision, and both empty cases.

**Mutation receipts.** Each mutation was applied to production source, run, then restored from a pristine copy
and confirmed by `git diff --quiet` *and* by re-counting the restored symbol.

| # | mutation | result | killed by |
|---|---|---|---|
| M1 | delete the breakout guard in `buildLaneInput` | KILLED | `TestBuildLaneInputRefusesBreakoutLanesWithTheBreakoutReason` |
| M2 | `For` returns `values[0]` whenever the symbol has any lane | KILLED | `TestForRefusesASymbolThatHasMoreThanOneLane` |
| M3 | drop `sort.Strings(keys)` in `LanesFor` | **discarded — false kill** | build failure, not a test |
| M3′ | replace the sort with `sort.SearchStrings` so the build still holds | KILLED | `TestLanesForReturnsEveryLaneInTheSameOrderEveryTime` |
| M4 | drop the `\x00` from `LanesFor`'s prefix | KILLED | `TestLanesForDoesNotLeakIntoASymbolThatMerelySharesAPrefix` |
| M5 | `selectRouteDecision` scans only `Decisions[:1]` | KILLED | `…FindsTheRequestedLaneEvenWhenItIsNotFirst`, `…SkipsUnknownDecisionsAndKeepsLooking` |

M3 is recorded rather than deleted. `sort` appears exactly once in `production.go`, so replacing the call with
`_ = keys` left the import unused and the package stopped compiling. A build failure looks identical to a kill
in the exit status and proves nothing about the test. M3′ is the real receipt.

**Tasks deliberately left unticked.** These tests close review findings, not tasks. 4.4 asks for strict
breakout snapshot/config *construction* (blocked by decision 49) and 4.5 for paired 8-descriptor integration
tests; neither is delivered here.

**Verification.** Real `gofmt` from `$(go env GOROOT)/bin` — 0 findings; `make lint` clean; `make test`
(untagged, the CI target) — no failures; `-tags tossos_testseams` for both packages — ok; `make sdd-sync`
all indexes current; `make sdd-check` OK. `make gate CHANGE=a112-…` still fails with **47 incomplete tasks**,
unchanged by this commit — it is the change-completion gate and cannot pass mid-change.

### Pre-existing tagged-only failure found while running the full tagged suite (2026-08-29)

`TestActualEngineRecoveryStillFailsClosedOnASnapshot429` (`cmd/tossctl/engine_account_seq_recovery_test.go:219`)
fails at `HEAD`, and fails identically with this commit's three new files moved out of the tree, so it is not
caused by them. The assertion arrived on 2026-07-30 in `93165f96`, a month before a112 L3, and sits on many
branches — and it is behind `tossos_testseams`, so CI has never run it.

**What actually fails matters.** The safety assertions pass: recovery returns `ErrRecoveryIncomplete`,
`report.Complete` is false, the partial snapshot is discarded, and `/api/v1/accounts` is called exactly once.
The failing assertion is the call count — `/api/v1/orders calls = 21, want 1`, over 300s. So recovery **does**
fail closed; what drifted is the retry policy against a 429 on a read endpoint. No order-placing path is
implicated. It is recorded here as an open finding, unowned by a112, rather than silently carried.

### Decision 51 — task 4.3's three seals, and why one of them was deleted rather than kept

51. **The descriptor table and the exact candidate validation were already four families per market; what task 4.3
    still owed was the family, scoring and calibration seals. All three now live in
    `internal/strategyrouter/production.go`, and a legacy manifest is refused as activation authority by a declared
    contract instead of by an emergent byte accident.**

    **Anchor correction, adopted.** The independent review's C3 cited `production.go:389`
    (`return len(seen) == len(want)`). Measured against the frozen blob — `git show
    c8d9c0d5:internal/strategyrouter/production.go`, function at 374-390, three AST branches — the arm that kills a
    legacy three-family manifest is **B1 at `:376`, `len(values) != len(want)`**; `:389` is a tail return that is
    always true once the loop completes. Every claim below is anchored on B1, not on the tail.

    **No legacy tolerance.** `design.md:208` lists `missing`, `partial 3-of-4` and `legacy 3-lane ON manifest`
    together as refusal targets, `:210` excludes implicit migration and `:254` says a legacy manifest stays
    `OFF/refusal`. So 4.3's own sentence and C3 do not conflict, and **no migration fallback was added**. What the
    seals add is that the refusal is now *declared*: `productionRouteSchema` moved to `strategy-lane-authority:v2`
    (`production.go:34`), so a v1 manifest is refused by name rather than only by the canonical-JSON byte round-trip.

    **What each seal is.**

    - **Family seal.** `productionLaneDescriptor` gained `Family` and the table (`production.go:563`) is now the one
      declaration of the family/lane binding. `validProductionRouteCandidates` checks
      `descriptor.Family != value.Family` (`:547`); the manifest carries `family` per candidate (`:84`).
      `ProductionRouteAuthority.Seals().Family` is a domain-separated SHA-256 over the market and the
      family/horizon/lane/version rows (`:222`).
    - **Scoring seal.** The raw `int64 Score` is **gone from the manifest**. Candidates carry
      `score_ppm` (`:88`), bounded to the approved `0..1,000,000` (`design.md`: "integer `score_ppm`(0..1,000,000)"),
      and `LoadProductionRouteAuthorityBatch` deliberately leaves `Candidate.Score` unset. `Seals().Scoring` (`:237`)
      hashes the approved score version together with the per-family ppm.
    - **Calibration seal.** The manifest body carries `arbitration_score_version` and `calibration_digest`
      (`:121-122`); `verifyProductionRouteManifest` refuses the whole market when either is absent (`:487`).
      `Seals().Calibration` (`:250`) binds the two. `SealsValid()` (`:196`) recomputes all three from the calibration
      and score rows the accessors return, so a value with a seal but drifted materials — the exact defect class of
      decision 50 — is rejected by construction.

    **The scoring seal is load-bearing, not decorative.** With no raw score in the manifest, legacy `Route` sees four
    eligible candidates all scoring zero and returns `RefusalAmbiguous`. "No pre-evaluation selection" therefore stops
    being a promise about call sites and becomes a property of the data: there is nothing left to select *by*.
    `TestProductionRouteCandidatesCarryNoRawArbitrationScore` asserts both halves.

    **Mutation receipts.** Each mutation was applied to production source, built, run, then restored from a pristine
    copy taken before the mutation and confirmed byte-identical. A mutation that broke the build was discarded rather
    than counted (the M3 lesson from the previous lot); none of these did.

    | # | mutation | result | killed by |
    |---|---|---|---|
    | M6′ | delete `descriptor.Family != value.Family` | KILLED | `TestProductionRouteCandidatesRejectFamilyDriftAndPartialFamilyCoverage` |
    | M7 | delete the `len(families) == len(want)` tail clause | **SURVIVED** | nothing — see below |
    | M8′ | delete the `score_ppm` range check | KILLED | `TestProductionRouteCandidatesRejectAScorePPMAboveTheApprovedRange` |
    | M9 | delete the calibration requirement in `verifyProductionRouteManifest` | KILLED | `TestProductionRouteManifestRefusesAMissingCalibrationSeal` |
    | M10 | hand a raw score back with `Score: int64(value.ScorePPM)` | KILLED | `TestProductionRouteCandidatesCarryNoRawArbitrationScore` |
    | M11 | roll the schema back to `strategy-lane-authority:v1` | KILLED | `TestProductionRouteManifestRefusesTheLegacySchemaVersion` |
    | M12 | make the scoring seal ignore `score_ppm` | KILLED | `TestProductionRouteAuthorityCarriesThreeIndependentSeals` |
    | M13 | return the sealed score slice instead of a copy | KILLED | `TestProductionRouteAuthorityCarriesThreeIndependentSeals` |
    | M14 | `SealsValid()` returns true unconditionally | KILLED | `TestProductionRouteAuthorityCarriesThreeIndependentSeals` |
    | M17 | bind the KR breakout lane to `FamilyContinuation` in the table | KILLED | `TestProductionRouteDescriptorsCoverFourFamiliesPerMarket` + 6 others |

    **M7 survived, and the code was deleted rather than the test strengthened.** The first draft also counted
    distinct families and returned `len(families) == len(want)`. No input can make that false: the count equality at
    B1 fixes the list length, lane ids are deduplicated, and each lane must carry the table's family — so the family
    count always equals the lane count. An unfalsifiable condition is not a defence, it is an unkillable claim of
    coverage. The clause is gone; the property it pretended to check — *the table itself* holds four distinct
    families — is now asserted where it can actually fail, in
    `TestProductionRouteDescriptorsCoverFourFamiliesPerMarket`, and M17 is its receipt.

    **Ownership ledger, extended by exactly one file.** `internal/strategyrouter/production_test.go` had to be
    edited: it builds `productionRouteCandidate` with **positional** struct literals (`:253` at the time of the edit),
    so any new sealed field breaks its compilation, and its fixture manifest is by definition a *legacy* manifest once
    the seals are required. There is no way to add a required seal and leave the fixture that omits it valid — that is
    the feature. Literals are now written with named fields so the next sealed field cannot silently shift a value
    into the wrong slot.

    **What this does not do — recorded, not smoothed over.**

    - The seals are **carried, not yet consumed**. Nothing downstream calls `SealsValid()` or `Calibration()`. The
      consumer is task 5.4's coordinator ("validates proposal seals/freshness … calibrated arbitration"), and
      `design.md:239` puts `ARBITRATION_UNCALIBRATED` there. Wiring the check into the existing market worker would
      hide arbitration inside the component `design.md` explicitly says not to hide it in, and would require editing
      `internal/strategyrouter/production_testseam.go`, which is outside this lot's ledger.
    - **Refusal diagnosability is not closed.** Every manifest fault still collapses into the single sentinel
      `ErrProductionRouteUnavailable`, so an operator sees `ROUTE_AUTHORITY_INVALID` without knowing which seal was
      missing. Typed refusal reasons were **not** implemented: `verifyProductionRouteManifest` is one boolean
      expression, and attributing a reason means either splitting a high-risk validator mid-lot or keeping a second
      copy of each checked value — the failure mode the "correction unit is the value" rule exists to prevent. The
      `:v2` schema name closes the *class* (a legacy manifest is legible as such); per-field reasons remain open.
    - Untouched on purpose, per the C1 re-classification: `Descriptors()`/`registry.go`, activation digests, desired
      state and the eight-element `four-family-runtime-v1.json` golden. `internal/strategyrouter/router.go` is not
      edited — not one line.

### Decision 52 — review finding C2 survives task 4.3, and the correction belongs to one producer

52. **C2 is real, it is not removed by 4.3's seals, and the minimal canonical fix is one guard in
    `strategyProposalAuthorityLoader.collectMarket` — not three edits at the three gates.**

    **Measured, from the AST rather than by eye.** `ProductionBatchAuthority.For`
    (`internal/strategyproposal/production.go:100-106`, one branch, B1 at `102:2`) returns `(zero, false)` when a
    symbol has anything other than exactly one lane, and **carries no reason**. In the engine,
    `strategyProposalAuthorityLoader.collectMarket` consumed that as `refused++; continue` — the symbol simply left
    the list. Three readers then share one shape:

    | reader | anchor | branch |
    |---|---|---|
    | `strategyProposalAuthorityPair.ResultAuthority` | `strategy_proposal_authority.go:98` | B1 at `95:3` (pre-edit AST) |
    | `strategyAccountAuthorityLoader.collectMarket` | `strategy_account_first_leg_authority.go:155` | B1 at `155:2` |
    | `strategyProjectionFromAssembly` | `strategy_runtime_projection.go:105` | B7 at `105:3` |

    All three test `len(entries) != 1`, and all three read the **same** `strategyProposalMarketAuthority.entries`. So
    with two proposing symbols the market is blocked; drop one of them for being *ambiguous* — a fail-closed decision —
    and the count becomes 1 and the **other** symbol is released. One symbol's refusal to choose becomes another
    symbol's permission to trade.

    **Does 4.3 remove it? No.** The seals validate and carry manifest data; they create no arbiter. The calibrated
    arbiter that would resolve two families for one symbol is task 5.4, which does not exist. The hypothesis that a
    calibration arbiter dissolves C2 is right about the *end state* and wrong about *this lot*.

    **The correction unit is the value, so it is applied once.** The value is the entry list. The guard sits in
    `collectMarket` (`strategy_proposal_authority.go:207`) **before** the first `entries = append`, and returns the
    new typed reason `AMBIGUOUS_FAMILY_PROPOSAL` (`:40`) for the whole market. `ProductionBatchAuthority.Ambiguous`
    (`internal/strategyproposal/production.go:112`) is what lets the caller tell "no proposal" from "too many". The
    two gate files outside this lot's ledger — `strategy_account_first_leg_authority.go` and
    `strategy_runtime_projection.go` — were **read only** and are unchanged; they need no edit because they consume
    the corrected list.

    The change is one-directional: it only adds a refusal. No input that was refused before is admitted now.

    | # | mutation | result | killed by |
    |---|---|---|---|
    | M15 | delete the market-level ambiguity guard | KILLED | `TestProposalCollectMarketClosesTheMarketBeforeBuildingEntries` |
    | M15b | call `Ambiguous` but do not return | KILLED | `TestProposalCollectMarketClosesTheMarketBeforeBuildingEntries` |
    | M16 | `Ambiguous` uses `> 2` | KILLED | `TestAmbiguousSeparatesNoProposalFromTooManyProposals` |
    | M18 | `Ambiguous` counts the market instead of the symbol | KILLED | `TestAmbiguousSeparatesNoProposalFromTooManyProposals` |

    **What the coverage says, and it is not flattering.** Measured with count-mode profiles:
    `strategyProposalAuthorityLoader.collectMarket` is entered **4x under the tagged engine suite and 0x under the
    untagged one**, and the new refusal arm (B12 at `217:3`) is entered by **no** suite. The runtime path has no CI
    coverage at all. That is why the engine-side test is a structural one: it parses the function and requires the
    ambiguity refusal to *end before* the first `entries = append`, so it fails whether the guard is deleted or merely
    demoted to a counter. Both mutations above confirm it. The full-fidelity behavioural test lives where it can run
    untagged — `internal/strategyproposal`, where a batch with two lanes for one symbol can actually be built.

    **Residual, and whose it is.** The same composition exists one stage earlier:
    `strategyRouteAuthorityLoader.collectMarket` also does `refused++; continue` when `RouteSet` refuses a *signed*
    scope, and an owner-reconstruction integrity fault there shrinks the route list the same way. `design.md`'s fault
    table already assigns that case — "journal/Gateway/fence/owner integrity fault → 모든 신규 entry fail-closed" — and
    task **5.6** is its named deliverable. It is not corrected here because `RefusalDisabled` (every lane OFF) is the
    *normal* path through that same arm, so separating integrity faults from ordinary dormancy is a behaviour change
    that belongs with the coordinator, not tacked onto this lot.

### Decision 53 — the RouteSet-widened proposal admission is inside the intended scope, but its singleton is un-arbitrated

53. **Replacing `Route` with `RouteSet` widened proposal admission from "equals the single raw-score winner" to
    "is a member of the eligible set". That is what task 4.3.1 asked for, it cannot widen an owned symbol, and with
    OFF defaults it admits nothing at all — but the family that reaches dispatch is now decided by manifest
    composition rather than by an arbiter, and closing that is task 5.4's.**

    **What changed, measured against the frozen base.** At `a8c3d067`,
    `internal/strategyproposal/production.go:241` called `strategyrouter.Route(target.Router)` and admitted a scope
    only if `routed.Decision` matched it exactly. Now (`:289`) it calls `RouteSet` and defers to
    `routeSetAdmitsScope` (`:319-333`, three AST branches: B1 range `320:2`, B2 if `321:3`, B3 if `327:3`), which
    admits any scope whose full identity appears in the eligible set. The result map keys on
    `batchKey(symbol, laneID)`, so two families for one symbol no longer overwrite each other.

    **Three bounds, each measured from `RouteSet`'s own AST (`routeset.go`, 23 branches).**

    1. *An owned symbol cannot widen.* B15 at `104:2` (`len(active) == 1`) returns at `110:3` with **exactly one**
       decision carrying `ExistingOwner: true`. `routeSetAdmitsScope`'s B3 then refuses when
       `decision.CampaignID != scope.CampaignID`. So no second family can enter a symbol that already has an owner —
       the duplicate-entry risk `design.md` names first is untouched.
    2. *OFF admits nothing.* B20 at `124:3` skips any candidate that is not `Eligible && Desired==ON &&
       Effective==ON`; B21 at `130:2` then returns `RefusalDisabled`, and the proposal loader requires
       `routed.Code == RefusalNone` before it consults the set at all. With every descriptor shipped OFF the widened
       admission is unreachable.
    3. *Reaching it requires a signed v2 manifest.* Turning a lane ON means signing a manifest whose candidates are
       `eligible/desired/effective = ON` — which, after decision 51, must also name an approved
       `arbitration_score_version` and `calibration_digest`. So the calibration seal makes the *un-calibrated* variant
       of this widening unsignable, even though it does not arbitrate.

    **What genuinely widened, and the honest name for it.** A symbol whose highest-raw-score family had no proposal
    scope used to produce nothing; now a lower-ranked family with a scope is admitted and, if it is the only one,
    passes the `len(entries) == 1` gate and is dispatched. Removing raw-score preselection is exactly what
    `design.md` requires ("Raw `Candidate.Score` 또는 registry 순서는 cross-family 선택 권위가 될 수 없다"), so the
    *removal* is not a defect. The defect-shaped remainder is that the arbiter meant to replace it is not here, and
    `design.md` is explicit that even a lone proposal must be refused without an approved calibration
    ("singleton proposal도 approved common score version/calibration digest가 없으면 `ARBITRATION_UNCALIBRATED`로
    거부한다"). Today an un-arbitrated singleton is accepted.

    **Disposition.** Not a second C2 — C2 was one symbol's refusal releasing another symbol, which is a composition
    bug; this is a missing component, already scoped, gated behind a human-signed ON manifest and behind a100
    protection. It is recorded as an open obligation on **task 5.4**: the coordinator must refuse a singleton whose
    route authority fails `SealsValid()` or names no approved calibration. Decision 51's seals are the inputs that
    obligation now has to work with.

### L3 landing state (2026-08-29) — 4.3 ticked

**Newly ticked: 4.3.** The descriptor table and exact candidate validation were already four families per market;
this lot added the three seals the task sentence also asks for and made the legacy refusal a declared contract
(`strategy-lane-authority:v2`). Every seal has a mutation receipt in decision 51.

**Still not ticked, unchanged reasons.**

- **4.4** — `buildLaneInput`'s breakout construction is still blocked by decision 49 and fails closed.
- **4.5** — not started; its "all 8 descriptors" is reachable for route and proposal lineage but not for a breakout
  *production* input, for the same reason as 4.4.

**Untagged test counts (top-level `Test` functions, `go test -list`).**

| package | before | after |
|---|---|---|
| `internal/strategyrouter` | 40 | 46 |
| `internal/strategyproposal` | 7 | 8 |
| `internal/app/engine` | 386 | 387 |

All six new router tests, the new `internal/strategyproposal/ambiguous_symbol_test.go` and the new
`internal/app/engine/strategy_proposal_ambiguity_test.go` carry **no build tag**, so CI runs them.

**Measured verification.**

| Command | Result |
|---|---|
| `$(go env GOROOT)/bin/gofmt -l .` | 0 findings |
| `make lint` (`go vet ./...`) | clean |
| `go vet -tags tossos_testseams` over strategyrouter/strategyproposal/strategyflow/app-engine | clean (exit 0) |
| `make test` (untagged, the CI target) | no failures; `tools/a112-mb-us-source` 309s dominates the run |
| `go test -count=1 -tags tossos_testseams ./internal/{strategyflow,strategyproposal,strategyrouter,app/engine}/...` | ok (engine 82s) |
| `openspec validate … --strict --no-interactive` | valid |
| `python3 tools/logic-map/check_analysis.py --change a112-…` | evidence complete |

**Function Logic Map bundles refreshed or created: 16.** Eleven were stale because the file hash moved; five are
new (`verifyProductionRouteManifest`, `LoadProductionRouteAuthorityBatch`,
`ProductionRouteAuthority.OwnerDigest`, and two production-route test functions). Every branch row carries a
count read out of a count-mode coverage profile, and rows whose arm no profile entered say so.

`ProductionRouteAuthority.OwnerDigest` is pinned `revision: base` on purpose: its body is unchanged, and the
checker required it only because a pure insertion immediately after a function counts as intersecting it.

**The uncomfortable number, recorded rather than smoothed over.**
`strategyProposalAuthorityLoader.collectMarket` is entered **0x by the untagged (CI) engine suite** and 4x by the
tagged one, and its new ambiguity refusal arm is entered by **no** suite at all. The C2 correction is therefore
guarded in CI by a structural test over the function's own AST, plus a full-fidelity behavioural test one package
down in `internal/strategyproposal` where an ambiguous batch can actually be constructed without a build tag.

### Decision 54 — C1 is a deploy-gate item, not a code defect, and C3's mechanism is subsumed by it (2026-08-29)

The independent review's two remaining CRITICALs were re-derived from the code. **Both prescriptions change.**
Every anchor below was re-read from `git show c8d9c0d5:<path>` blobs, because task 4.3 moved the line numbers in
`internal/strategyrouter/production.go` while this was being measured.

**The digest chain, and what a mismatch actually does.** `strategy_schedule_authority.go:288-292` hashes
`{Domain, RouterID, RouterRelease, Lanes[]}` with SHA-256, where each `Lanes` element is
`{Market, Horizon, LaneID, LaneVersion, Release}` — **lane identity only; `Desired`/`Effective` are not in it.**
Four places compare the result: `scheduler/desired.go:292`, `scheduler/production_activation.go:188`,
`strategyrouter/production.go:329`, `strategyproposal/production.go:518`. A mismatch is neither a panic nor an
exception: `Restore` yields `ResumeDesiredMismatch` → `Ready=false` → `strategy_candidate_authority.go:150` and
`strategy_route_authority.go:147` return `SCHEDULE_NOT_READY` → `strategy_entry_supervisor.go:425` returns nil.
**The strategy entry loop keeps cycling and produces zero candidates. The safety loops are unaffected.**

**C3 is subsumed.** The digest gate is evaluated *before* manifest validation, so the lane-count check C3 named is
never reached. Repairing desired-state alone would still have `production.go:329` reject an old manifest
irrespective of candidate count. C3 is therefore not tracked as an independent item.

**Anchor correction.** The reported `production.go:389` (`return len(seen) == len(want)`) is the right function but
the wrong arm. Of that function's three branches (374-390 at `c8d9c0d5`), **B1 at `:376`
(`len(values) != len(want)`)** is what kills a legacy manifest; `:389` is an always-true tail return. Decision 51
is anchored on B1.

**There is no lane toggle in the code.** `Descriptors()` (`internal/strategyflow/registry.go`) is
`return append([]Descriptor(nil), pairedDescriptors[:]...)` — one statement, zero branches, and it reads no
toggle. A whole-tree `grep -rn "TOSSOS_STRATEGY"` returns only signed-manifest digest pins and public keys.
Lane ON/OFF is expressed **only** inside the signed manifest's `eligible/desired/effective`, and all eight
`pairedDescriptors` elements are `Desired: StateOff, Effective: StateOff`. **"Eight present, all OFF" is this
design's canonical spelling of OFF**, and task 4.1's frozen golden (`four-family-runtime-v1.json`, enforced by
`breakout_binding_test.go:151`'s `len != 8`) is what fixes it there.

**C1 is not a burden a112 created.** The activation manifest's lifetime is 24 hours
(`production_activation.go:24`, `productionActivationMaximumLife = 24 * time.Hour`) and `build_digest` is bound to
the commit (`protection_wiring.go:24`). **Deploying a112 at all already invalidates the manifest.** Re-issuing is
a standing requirement of every deploy, not this change's regression.

**Current real-account exposure is zero, and the reason is corrected.** `~/.config/tossctl` holds no
`scheduler-desired-*`, `strategy-activation-*`, `*lane-authority*` or `strategy-proposal-*` file; its three
strategy-related directories contain only socket `endpoint.json`. With no file, `DefaultDesiredState()`
(`AutoStart=false`) applies and `Restore` returns `ResumeAutoStartOff`, so **the digest comparison is never
reached at all.** "Blocked because the digest differs" was the wrong description; "never gets as far as comparing"
is the right one.

**Disposition.**

- **C1 is reclassified from CRITICAL to a deploy-gate item.** L3's ownership ledger contains neither
  `strategy_schedule_authority.go` nor `strategy_runtime_projection.go`, so there was never an L3 edit that could
  address it. **No L3 code changes.**
- **L5** exposes `ConfigDigest`/`BuildDigest` additively on the read-only projection. Today the value is filled
  into the snapshot but leaves through no operator surface at all
  (`grep -rn "ConfigDigest" internal/httpapi/ cmd/tossctl/` → 0 hits), so **the number a human must write into the
  file cannot be read.** Proof obligation: projected value == `strategyRuntimeConfigDigest()`.
- **L7** turns the re-issue order (KR/US × 5 steps) into a deploy manifest.
- **No signer is built.** a072 spec OS-1 fixes signing as a human authority outside this repository.

**C3's prescription is rejected; only its diagnosability residue is accepted.** `design.md:208` lists `missing`,
`partial 3-of-4` and `legacy 3-lane ON manifest` side by side in one sentence as things to refuse; `:210` excludes
implicit migration because it widens the approved scope; `:254` states that a legacy manifest stays OFF/refusal.
C3's narrow reading requires tolerating `missing`, which contradicts `:208` directly. §0.2's "OFF = upstream" does
not govern here: the evidence that clause points at is the 650 upstream-inheritance tests, and
`internal/strategyrouter/router.go` is product code first created in TossOS commit `7b57fd24` (2026-08-04).
**Runtime decision semantics do not change** — breakout-OFF is expressed by signing the fourth candidate
`eligible:false`. C3 read "behavioural compatibility" and "file compatibility" as one thing. The residue is
per-field refusal attribution, and it is **not closed**: decision 51 raised the schema to
`strategy-lane-authority:v2`, which closes the *class* of the refusal, but `verifyProductionRouteManifest` remains
a single boolean. Carried forward as an open item.

**Manager rulings on decision 51's two requests.**

1. **Ledger extension by one file — `internal/strategyrouter/production_test.go` — approved.** It is the test for
   a file already inside the ledger, it builds `productionRouteCandidate` with positional literals so a new field
   breaks it mechanically, and its fixture manifest is by definition a legacy manifest once seals are required.
2. **The `strategy-lane-authority:v2` bump — approved.** Verified fail-closed: the schema is checked inside the
   refusing conjunction at `production.go:480`. Verified zero live impact: no signed route manifest exists on
   disk. And `design.md:210` requires a re-signed four-family manifest regardless.

**Independent Manager verification of `e12bca03`** (re-run, not taken on report):

| Check | Result |
|---|---|
| `git diff c8d9c0d5 e12bca03 -- internal/strategyrouter/router.go` | empty — Route is untouched at compile level |
| `grep -rn "strategyrouter\.Route("` outside tests | 0 callers |
| C2 guard diff | +14 lines, insertion only, returns `fail(...)` before the first `entries` append |
| `go test -count=1` over strategyrouter/strategyproposal/strategyflow/app-engine | ok |
| the same four packages with `-tags tossos_testseams` | ok |

### The `cmd/tossctl` rate-limit failure has an owner, and it is not a112 (2026-08-29)

The failure recorded above as unowned is now diagnosed. **It is a stale expectation, not a production defect.**

**21 is not a retry policy; it is a division.** `MaxRateLimitWait` 5m ÷ `RateLimitBackoff` 15s + 1. The 300-second
wall time comes from the test passing `clock.System()`. Chain:
`engine_account_seq_recovery_test.go:194,205` → `runtime_wiring.go:175` → `reconcile/recovery.go:238` → `:298` →
`:371-404` `stableSnapshot` → `Collector.Collect` → `snapshot.go:271` `ScanOrders` → `GET /api/v1/orders` →
`official/errors.go:54` → `ErrRateLimited` → `recovery.go:382` `waitOutRateLimit` → `ratelimit.go:108`
`clk.Sleep(ctx, 15s)` → `recovery.go:385-386` "Deliberately no attempt++" → `continue`.

**The contract was replaced deliberately.** The test file is a single commit, `93165f96` (2026-07-30), never
updated. The retry contract changed 14 days later in `1c76a580` (2026-08-13,
`feat(reconcile): 복구가 rate limit에 죽지 않는다 [a102 §1]`), whose message cites a real 2026-08-13T02:03:30.545Z
incident and states the pre-change behaviour outright: a 429 on the boot account read ended restart recovery as
immediately incomplete. `a102/specs/reconciliation/spec.md:14-16` requires the retry, the finite total wait, and a
closed entry gate throughout. a102 is 34/34 complete. It hid because a102's own verification command
(`a102/tasks.md:124`) carries no build tag.

**Safety judgement — no shared mutation path.** The only code that loops is `stableSnapshot`, and inside it only
`Collector.Collect` (three GETs). The execution gateway has no mutation-retry entry point at all
(`internal/execgw/retry.go:6-14`, `:104-106`). Collector and gateway share one `official.Client` object but never
overlap in time.

**The cost that is real, and why it is still an improvement.** `runtime.go:288-297` runs `Recover` to completion
**before any loop**, and `exit` — the stop-loss observer — is one of those loops (`cmd/tossctl/engine.go:679-710`).
A sustained 429 at boot therefore delays the stop-loss loop by up to five minutes. This is **not** a weakening
under safety invariant 4: before `1c76a580` the same 429 ended recovery immediately, the engine did not start, and
the stop-loss loop did not run *at all*. The bound is finite and exceeding it fails closed.

**Owner: not a112.** Fixing the assertion belongs with the lot that changed the contract (a102, not yet archived,
and already holding the relevant logic-map bundles) or a small new change. Hard-coding `want 21` is refused: 21 is
derived from two constants, so it rots when they move, and it burns 300 seconds of CI on every run.

### Decision 55 — L5 5.0: 운영자가 적어야 하는 숫자를 읽을 수 있게 한다 (2026-08-29)

결정 54가 L5에 남긴 잔여를 종결한다. 그 결정의 근거는 명령 하나였다 —
`grep -rn "ConfigDigest" internal/httpapi/ cmd/tossctl/` → **0 hits**. 엔진은 두 digest를 계산해
스냅샷 안에 채우면서도 어떤 운영 표면으로도 내보내지 않았고, 그래서 **사람이 활성화 매니페스트에
적어야 할 숫자를 읽을 방법이 없었다.**

**어디에 뒀는가 — envelope이지 시장 레코드가 아니다.** 두 값은 시장별 사실이 아니라 엔진 프로세스
하나의 사실이다. 그리고 운영자가 이 숫자를 찾는 순간은 **두 시장이 다 UNKNOWN일 때**다. 시장
레코드 안에 뒀다면 정확히 그때 사라진다. `SchemaVersion`/`GeneratedAt`과 같은 층에 둔 이유다.

**어디서 붙이는가 — store가 아니라 `Context.Read`.** store의 내용은 전략 assembly refresh가
채우는데, 그 refresh는 아직 한 번도 안 돌았거나 실패했을 수 있다. 운영자가 이 숫자를 가장 필요로
하는 상태가 바로 그 상태다. `Read`는 REST·SSE·콘솔·Unix transport가 모두 지나는 단 하나의
출구라, 여기 붙이면 표면마다 사본이 생기지 않는다.

**콘솔이 스스로 계산하면 안 되는 이유.** 두 값 모두 같은 저장소의 상수(`RouterID`,
`RouterRelease`, `Descriptors()`, `version.Current()`)로 만들어서 콘솔 바이너리도 똑같이 계산할 수
있다. 하지만 콘솔과 엔진의 build가 다르면 그렇게 만든 숫자는 **엔진이 거절할 매니페스트**를 낳는다.
값의 권위는 엔진 프로세스 하나이고, 엔진이 아닌 곳이 만든 스냅샷은 값이 아니라 `not_observed`다.

**증명 의무.** 결정 54가 요구한 그대로 — 투영된 값 == `strategyRuntimeConfigDigest()`.
`TestStrategyRuntimeReadExposesThisProcessConfigAndBuildDigest`가 dormant 스냅샷에서 이를 잰다.

| 확인 | 결과 |
|---|---|
| `Context.Read`의 기존 테스트 수 | **0** — 이 lot이 이 함수를 처음 실행한다 |
| RED 실측 | `Validate`(4 하위 케이스), `Clone`, 엔진 노출, 콘솔 두 건 |
| M1 — `Clone`에서 `Runtime` 복사 제거 | KILLED (REST 응답의 `runtime.configDigest`가 null) |
| M2 — `WithRuntimeIdentity`의 config/build 인자 맞바꿈 | KILLED (엔진 테스트 2건) |
| `go vet ./...`, `gofmt -l` | 0 |
| `check_analysis.py --change a112…` | evidence complete |

**⚠ 이 문단의 최초 판단은 틀렸고 결정 56이 뒤집었다.** 처음에는 구버전 콘솔이 새 엔진을 못
읽는 것을 "공개하고 받아들이는 대가"로 적었다. 그것은 이 change 자신의 승인된 spec 위반이다 —
아래 결정 56을 보라.

**SchemaVersion은 올리지 않았다.** `tossos.strategy-runtime-projection/v1` 그대로다. 올리면
`Validate`가 버전 불일치로 **양쪽 방향 모두**를 하드 실패시킨다 — 지금은 한 방향만 저하된다.
필드는 순수 가산이고 짝이 비어 있는 것이 유효한 상태이므로 버전을 올릴 계약 변화가 아니다.

**남긴 것.** 결정 54의 L7 잔여(KR/US × 5단계 재발급 순서표)는 그대로 열려 있다. 이 lot은 사람이
**읽을** 값을 만들었을 뿐, 무엇을 어떤 순서로 서명할지는 L7이 쓴다.

**base-commit 재고정.** a112의 `base-commit.txt`를 `a8c3d067`(a117) → `aeeb209e`(a118)로 옮겼다.
a118이 a112 위가 아니라 옆에 착륙했기 때문에, 옮기지 않으면 a112의 "수정된 함수" 집합에 a118의
테스트 편집이 섞여 들어와 이 lot이 남의 변경에 대한 증거를 요구받는다.

### Decision 56 — 독립 리뷰가 5.0의 P1 둘을 잡았다: 접두사와 wire 관용 (2026-08-29)

교차 모델(Codex) read-only 적대 리뷰. **P0 0건, P1 2건, P2 2건.** 둘 다 실제 결함이고 둘 다 고쳤다.

**P1-1 — 접두사를 벗겨서 기능 자체가 무효였다.**

투영이 `projectionDigest()`로 `sha256:`를 벗긴 64자리를 내보냈다. 그런데 매니페스트 검증은
`body.ConfigVersion != binding.ConfigVersion`과 `body.BuildDigest != binding.BuildDigest`의
**정확한 문자열 비교**이고(`internal/scheduler/production_activation.go`의
`validateProductionActivationManifest`), binding 쪽 값은 `strategy_schedule_authority.go`의
`prepare`가 넣는 `sha256:<64hex>`다. `internal/strategyrouter/production.go`의
`body.ConfigVersion != config.SchedulerConfigVersion`도 같다.

즉 **화면의 숫자를 그대로 옮겨 적으면 `ErrManifestMismatch`가 난다.** 이 lot이 없애려던 바로
그 실패를 이 lot이 새로 만들 뻔했다.

기전은 익숙하다. 이 패키지의 다른 digest 필드(evidence, activation)가 벗긴 64자리를 쓰고
`validDigest`가 그 형식만 통과시키므로, 새 필드도 그 관례에 맞춘 것이다. **보여 주는 값과
받아 적는 값의 차이**를 못 본 것이 잘못이다. 이제 `runtimeIdentityPrefix` 주석이 그 구분을 적는다.

**그리고 내 테스트가 결함을 그대로 굳혔다.** 기대값을 production과 같은 `projectionDigest()`로
만들었기 때문에 둘이 같이 움직여 통과했다 — a118에서 배수 3/5/7로 같은 오류를 저지른 것과
같은 부류다([[falsification-must-vary-the-right-axis]]). 정정은 **소비자가 요구하는 모양을
직접 적는 것**이다: `sha256:` 접두사 + 소문자 64자리 hex를 형식으로 단언하고, 그 다음에 신원을
비교한다. 뮤테이션 M4(원래 결함 재도입)가 두 테스트에 죽는다.

**P1-2 — 엄격 디코딩은 이 change 자신의 spec 위반이다.**

`strategyprojectionrpc.Client.Read`의 `DisallowUnknownFields()`는 새 엔진이 필드를 하나 더
실어 보내는 순간 구버전 콘솔의 읽기를 통째로 실패시킨다. a112의 승인된 spec
`runtime lineage와 health는 lane 단위로 결정적으로 관측된다`는 정반대를 요구한다 —
"older readers가 additive unknown fields를 무시할 수 있어야 한다 (SHALL)", 그리고
`legacy projection reader` 시나리오 "additive lane/coordinator fields를 무시해도 read가
실패하지 않는다".

결정 55에서 이것을 "공개하고 받아들이는 대가"로 적은 것은 **틀렸다.** 권위 순서는 안전 불변식 →
승인된 OpenSpec → 현재 HEAD·실행 테스트다. 스스로 승인한 spec이 금지한 실패를 배포 관행으로
정당화할 수 없다. 그리고 a112는 `lanes[8]`·`coordinators[2]`를 추가할 change이므로 이 충돌은
어차피 터질 예정이었다.

**엄격함이 실제로 지키던 것은 없다.** 응답은 0700 디렉터리 안 0600 소켓 뒤에서 bearer 토큰으로
인증된 엔진 자신이 보내고, 크기는 `MaxProjectionBytes`, 형식은 디코더, 값 개수는 trailing 검사,
의미는 `strategyprojection.Validate`가 각각 막는다. 모르는 필드 거절이 유일하게 만든 결과가
"화면 전체가 죽는 것"이었다. 디스크 파일을 읽는 `readDescriptor`의 엄격함은 **유지한다** —
위협 모델이 다르다.

**Manager 재정 — 원장 확장 2건.**

1. **`internal/strategyprojectionrpc/transport.go` — 승인.** L5 원장에는 없지만 이 lot이 바꾼
   projection이 건너는 transport 자체이고, 위반한 spec 요구가 이 change 자신의 것이다.
2. **`internal/strategyprojectionrpc/transport_unix_test.go` — 승인.**
   `TestUnixClientRejectsUnknownSchemaFieldsAndOversizedResponse`의 두 팔 중 unknown-field 팔이
   spec이 금지하는 동작을 고정하고 있었다. 크기 팔은 그대로 두고 이름을
   `TestUnixClientRejectsOversizedResponse`로 좁혔다. 잃은 커버리지는 없다 —
   `a112_additive_fields_test.go`가 양쪽으로 잰다. 사라진 이름은 base 리비전 FLM 번들로
   남겼다(`revision: base`).

**잔여 window는 여전히 있고, 이번에는 고칠 수 없다.** 이미 설치된 구버전 바이너리에는 엄격한
디코더가 박혀 있다. 이 정정이 사는 것은 **다음번부터**다. 지금 배포의 완화책은 하나뿐이고
현재 배포가 이미 그렇게 한다 — `compose.yaml`의 두 서비스가 같은 이미지에서 뜨고 한 번에 교체된다.

**P2 둘.**

- `runtime` 생략·`null`·`{}`가 Go에서 모두 같은 nil 쌍으로 디코딩되어 OpenAPI의 `required`를
  강제하지 못한다. **고치지 않는다.** 이 모델의 모든 필드가 같은 성질이고(생략된 `lane`도
  zero value가 된다), 문서의 `required`는 "우리가 무엇을 내보내는가"를 적은 것이며 Go는 이 필드를
  항상 직렬화한다. 받는 쪽 판정은 `Validate`가 의미 단위로 한다.
- 투영→Unix client→집계→SSE를 identity를 실은 채로 관통하는 테스트가 없다. `Client.Read`
  경로는 `TestClientIgnoresAdditiveFieldsFromANewerEngine`이 덮지만 SSE parity는 여전히
  nil 쌍 스냅샷을 쓴다. **열린 잔여로 남긴다** — L5의 8-worker projection이 `lanes[8]`을 실을 때
  같은 자리에서 함께 세우는 것이 맞다.

| 뮤테이션 (모두 실측) | 결과 |
|---|---|
| M1 — `Clone`에서 `Runtime` 복사 제거 | KILLED |
| M2 — config/build 인자 맞바꿈 | KILLED |
| M3 — 콘솔이 부재 자리에 값을 지어냄 | KILLED |
| M4 — `sha256:` 접두사 재차 제거 (P1-1 원래 결함) | KILLED |
| M5 — `DisallowUnknownFields()` 복원 (P1-2 원래 결함) | KILLED |

## 2026-08-29 L5 태스크 5.4.1 — 보정 중재 코어 (Manager 결정 57)

사용자 지시: `L5의 ConfigDigest 노출 해소 후 5.4 보정 중재자 진행`. ConfigDigest 노출은
결정 55~56으로 종결(`5c5efec7`)했고, 이어서 5.4를 두 선택지로 물어 사용자가 **1번(중재 코어만)**
을 골랐다. 코디네이터 런타임(큐·cadence·bounded handoff)은 5.1/5.1.1/5.2/5.3/5.5의 몫으로 남기고,
5.4를 5.4.1(코어, 이번 lot)과 5.4.2(런타임 이식)로 나눴다.

**무엇을 만들었나.** `internal/strategyarbiter` — 소유자 범위 하나에서 봉인된 제안 중 최대
하나를 고르는 순수 함수 `Arbitrate`. 거절은 11종의 typed refusal이다.

| 코드 | 언제 |
|---|---|
| `ARBITRATION_INVALID_REQUEST` | 신원 누락, 또는 제안들이 서로 다른 자격 집합에 묶임 |
| `ARBITRATION_NO_ELIGIBLE_PROPOSAL` | 견줄 제안 0건 |
| `ARBITRATION_INVALID_SEAL` | 봉인 뒤 값이 바뀜 |
| `ARBITRATION_SCOPE_MISMATCH` | 권한 열쇠 또는 제안 계보가 기대 범위 밖 |
| `ARBITRATION_STALE_PROPOSAL` | 후보 유효 기한 경과 |
| `ARBITRATION_DUPLICATE_LANE` | 같은 레인이 두 번 |
| `ARBITRATION_INELIGIBLE_LANE` | 봉인된 자격 집합에 없는 레인 |
| `ARBITRATION_UNCALIBRATED` | 채점 봉인 불일치·버전 불일치·상한 초과 |
| `ARBITRATION_UNKNOWN_FAMILY` | 가족 점수 행에 정확히 하나로 안 붙음 |
| `ARBITRATION_OWNER_CONFLICT` | 활성 소유자가 있는데 그 소유자 하나로 정리 안 됨 |
| `ARBITRATION_SCORE_TIE` | 최고점 동률 |

**설계 결정 셋.**

1. **채점 기준을 호출자에게서 받지 않는다.** 처음에는 `Proposal{Calibration, Scores, SealsValid bool}`
   로 잡았다가 버렸다. `SealsValid` 를 호출자가 `true` 로 적을 수 있으면 그 확인은 없는 것과 같다.
   지금은 `Proposal{Result, Authority}` 이고 중재자가 직접 `Authority.SealsValid()` 를 부른다.
2. **자격 집합도 유도한다.** `RouteSetResult` 를 인자로 받지 않고 `strategyrouter.RouteSet(authority.Request())`
   를 중재자가 부른다. 그리고 모든 제안의 자격 집합 `SetDigest()` 가 같아야 한다 —
   서로 다른 자격 집합에서 온 제안을 한 저울에 올리지 않는다.
3. **상한은 유도하고 복사하지 않는다.** `strategyrouter.ScorePPMMax = productionRouteScorePPMMax`.
   `1_000_000` 을 중재자에 다시 적으면 언젠가 한쪽만 고쳐진다([[contract-numbers-from-the-receipt]]의 재발 방지).

**배선.** `strategyProposalAuthorityLoader.collectMarket` 의 C2 사전 스캔(`batch.Ambiguous`)을
지우고 종목마다 `batch.LanesFor` → `arbitrateProposalScope` → 선택으로 바꿨다.
`StrategyProposalAmbiguousFamily` 상수는 생산자가 없어져 삭제했고
`StrategyProposalArbitrationRefused` + 스냅샷의 `ArbitrationRefusal` 필드가 대신 들어갔다.

**보수성.** 중재가 거절하면 그 종목만 빼지 않고 **시장 전체를 닫는다.** C2가 지적한 기전이
그대로 살아 있기 때문이다 — 하나를 빼면 목록이 둘에서 하나로 줄어 아래 세 소비자가 공유하는
`len(entries) != 1` 관문이 오히려 만족되고 상관없는 다른 종목이 대신 풀린다. 5.4 이후 새로
통과할 수 있게 된 경우는 오직 하나다: **봉인된 매니페스트가 승인한 점수로 유일 승자가 정해질 때.**

**실계좌 노출 0 (실측).** `~/.config/tossctl/` 에 `strategy-lane-authority-{KR,US}.json` 도
`strategy-proposal-input-{KR,US}.json` 도 없고, `TOSSOS_STRATEGY_LANE_*` / `TOSSOS_STRATEGY_PROPOSAL_*`
env는 저장소·배포 override 어디에도 없다. 매니페스트가 없으면 route가 never ready고 제안 자체가
만들어지지 않는다. 8개 레인이 꺼져 있다는 기억이 아니라 파일 부재로 확인했다.

**뮤테이션이 죽은 코드 둘을 잡아냈다.** 이것이 이번 lot에서 가장 중요한 부분이다.

- **M2 SURVIVED** — `calibration.ScoreVersion == "" || CalibrationDigest == ""` 조기 검사.
  `SealsValid()` 가 `productionRouteIdentity` 로 빈 값을 이미 거절하므로(`production.go:713`)
  이 검사는 **절대 실행되지 않는다.** 테스트를 더하는 대신 검사를 지웠다.
- **`selectHighestScore` 의 `if best < 0`** — 네 스위트 전부에서 진입 0회로 **측정**됐다.
  `Arbitrate` 가 빈 목록을 먼저 거절하므로 도달 불가다. 지웠다. 빈 목록이 들어와도 `ties == 0`
  이라 `ties != 1` 에서 닫히고 색인은 쓰이지 않는다.

같은 판단을 두 곳에 적으면 한쪽은 죽은 코드가 되고, 죽은 코드는 지켜 주는 것이 없다.

**뮤테이션 22종, 전부 KILLED**(위 둘은 제거로 처리). 살아남았다가 테스트를 더해 죽인 것 넷:

| # | 왜 살아남았나 | 무엇을 더했나 |
|---|---|---|
| M10 | 열거 밖 가족 이름을 쓰는 점수 행이 어느 테스트에도 없었다 | `TestAScoreRowNamingAnUnapprovedFamilyIsUnknown` |
| M11 | 한 레인에 점수 행이 둘 붙은 표가 없었다 | `TestALaneMatchingTwoScoreRowsIsUnknown` |
| M14/15/16/17 | 범위를 바꾸면 권한 열쇠와 계보가 **함께** 틀려, 계보 검사가 단독으로 발화한 적이 없었다 | `TestAProposalWhoseLineageLeavesTheScopeIsRefused` — 권한은 맞고 계보만 다른 요청 |
| M18 | 권한 열쇠 검사가 계보 검사와 늘 겹쳤다 | `TestAProposalMeasuredAgainstAnotherSymbolsAuthorityIsRefused` |
| M-E2 | **승자가 우연히 목록의 0번이었다.** `kr_short_absorption_reversal_v1 < kr_short_flow_continuation_v1` 이라 `lanes[0]` 이 정답과 같았다 | 승자를 세 레인의 **가운데**(breakout)로 옮겼다 — `lanes[0]` 과 `lanes[len-1]` 둘 다 죽는다 |
| M-E3 | 기대 종목을 승인 후보에서 읽는지 경로 권한에서 읽는지가 픽스처에서 늘 같았다 | `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused` — 승인 후보 005930, 권한과 계보는 000660 |

M-E2는 [[falsification-must-vary-the-right-axis]]의 재발이다. 크기(점수)만 바꾸고 **위치**(정렬 순서)를
안 바꿨더니, 늘 첫 번째를 집는 구현과 구분되지 않았다.

**삭제한 테스트.** `strategy_proposal_ambiguity_test.go` 의 AST 구조 검사를 지웠다. 그것은
`batch.Ambiguous` 호출이 첫 `append` 보다 앞에 있다는 *구조* 를 봤고, 그 구조는 사라졌다.
구조로 갈음한 이유는 그때 `ProductionBatchAuthorityForTest` 의 map key가 종목이라 한 종목에
제안 둘을 담을 수 없었기 때문이다. 새 seam `ProductionBatchAuthorityMultiLaneForTest` 가 그것을
가능하게 했고, 이제 같은 성질을 **행동**으로 잰다(M-E1이 KILLED로 그것을 증명한다).
사라진 이름은 base 리비전 FLM 번들로 남겼다.

**픽스처를 고쳤다 — 느슨하게가 아니라 엄격하게.** `proposalRoutePair` 가 만들던 경로 권한에는
보정도 가족 점수도 없었다. 5.4 이후 그런 권한은 제안이 하나뿐인 종목도 거절시킨다. 이것은
승인된 spec이 요구하는 동작이다("Singleton proposal도 approved score/calibration authority가 없으면
production dispatch에 전달해서는 안 된다"). 즉 픽스처가 **spec이 금지하는 상태**를 정상으로
흉내 내고 있었다. 매니페스트가 싣는 것과 같은 네 가족 점수표를 붙였고, 그 뒤에야 통과한다.

**측정 도구를 만들었다.** 커버리지 블록↔AST 분기 좌표 매퍼가 없어 새로 썼다. 첫 판은 "같은 줄에서
조건 뒤에 시작하는 블록"을 골랐는데, 여러 줄 조건에서는 여는 중괄호가 아래 줄에 있어 엉뚱한 블록을
집었다. "분기 위치 바로 다음에 시작하는 블록"으로 고쳤다. `switch` 는 자기 본문 블록이 없어
첫 `case` 의 수가 표시된다 — 표에 그대로 적는다.

**남긴 잔여.**

- 5.4.2 — 명시적 `MarketCoordinator` 타입, 큐 intake, coalescing, bounded handoff.
- `collectMarket` 의 B2/B3/B4/B7(route/FX 미준비, loader 미구성, 중복 종목)은 이번에도 진입 0회로
  측정됐다. 이 lot이 만든 분기가 아니라 기존 가드이며, 이전 번들에서도 같았다.
- 서로 다른 채점 버전의 제안이 한 범위에 섞이는 경우는 **오늘 생산에서 만들어질 수 없다** —
  종목당 경로 권한이 하나뿐이라 보정도 하나다. spec 시나리오가 요구하므로 기전은 만들었고
  단위 테스트로 증명했지만, 생산 도달성은 5.1의 봉인된 envelope가 워커별 보정을 실을 때 생긴다.
- 태스크 5.5의 의존성 폐쇄 증명은 열려 있다. 이번 lot은 **중재자 패키지 하나**에 대해
  직접 import 허용목록을 테스트로 강제했다(`TestTheArbiterCannotReachAnyMutationCapability`,
  M0으로 반증 확인). `go list -deps ./internal/strategyarbiter` 는 journal·execgw·broker를 포함하지 않는다.

## 2026-08-29 L5 태스크 5.4.1 독립 적대적 리뷰 응답 (Manager 결정 58)

리뷰 결과 P0 0건, **P1 2건**, P2 1건. **두 P1 모두 코드로 확인했고 둘 다 고쳤다.**
리뷰는 C2 회귀와 KR/US 공유 상태는 없다고 확인했다.

### P1-1 — 봉인된 제안과 경로 결정의 증거·설정 결속이 없었다 (확인·수정)

`Propose` 는 그때의 경로 결정에서 `EvidenceDigest`/`ConfigDigest` 를 계보에 그대로
박는다(`flow.go:67-68`). 내 중재자는 레인 삼항만 자격 집합과 맞춰 보고 그 두 값은 보지
않았다. 그래서 같은 `(계좌, 시장, 종목, 세대, 레인)` 이면서 **다른 증거로 만들어진** 옛
제안이 현재 권한을 타고 통과할 수 있었다. 후보 유효 기한 안이면 신선도도 못 막는다.

오늘 생산에서는 도달 불가다 — `collectMarket` 이 매 사이클 같은 권한으로 배치를 새로
만들고 제안을 사이클 넘어 보관하지 않는다. **그러나 5.1의 coalescing queue 가 바로 그
보관을 만든다.** 그때 이 구멍이 열린다.

`selectHighestScore` 가 결정의 두 다이제스트와 계보의 두 값을 비교하도록 고쳤다.
뮤테이션 M20(결속 제거)·M21(증거만)·M22(설정만) 모두 KILLED.

**리뷰가 지적한 더 중요한 것: 내 픽스처가 생산에서 있을 수 없는 상태였다.** 경로 픽스처는
`"lane-evidence"`/`"lane-config"` 를, 제안 계보는 `sha256:6…`/`sha256:8…` 를 쓰고 있었다.
생산에서 그 둘은 **같은 값**이다. 두 픽스처가 서로 어긋난 채로 모든 테스트가 초록이었고,
그게 이 검사의 부재를 가렸다. 값을 손으로 맞추지 않고 `arbitrationLineageDigests` 로
**실제 계보에서 읽게** 했다 — 두 곳에 적으면 언젠가 다시 어긋난다.

### P1-2 — 거절 코드를 지어냈다. 동결 골든에 이미 이름이 있었다 (확인·수정)

`analysis/goldens/four-family-runtime-v1.json:65` 의 `refusal_enums.arbitration` 은
정확히 여섯 개다.

`ARBITRATION_UNCALIBRATED` · `ARBITRATION_TIE` · `ARBITRATION_MULTIPLE_OWNER` ·
`ARBITRATION_STALE_OWNER` · `ARBITRATION_STALE_ENVELOPE` · `ARBITRATION_SEAL_MISMATCH`

같은 파일 6~7행이 "내용을 바꾸려면 Manager 가 쓴 OpenSpec amendment 와 새 manifest/receipt
가 필요하다"고 적어 두었다. 나는 열한 개를 **지어냈고**, 그중 넷은 이름이 다르고
(`SCORE_TIE`/`STALE_PROPOSAL`/`INVALID_SEAL`) 둘(`MULTIPLE_OWNER`, `STALE_OWNER`)은 아예
없이 `INVALID_REQUEST` 로 뭉갰다.

**이것은 [[contract-numbers-from-the-receipt]] 의 다섯 번째 재발이다.** 앞의 넷과 같은
기전이다 — design.md 산문과 spec.md 산문을 읽고 이름을 **유도**했다. 동결 골든을 찾아본
적이 없다. 심지어 "이 enum 을 고정한 소비자가 없다"고 스스로 확인까지 했는데, 그 확인은
`ARBITRATION_` 을 코드에서 grep 한 것이었지 **계약 파일을 찾은 것이 아니었다.**

수정:

- `Refusal` 을 골든의 여섯 개로 좁혔다. 값은 골든에서 그대로 옮겼다.
- `RouteSet` 이 묵은 리비전과 소유자 중복을 같은 코드로 뭉치므로, `ownerSnapshotFault` 가
  스냅샷의 소유자 목록·리비전·신선도 창을 **직접 보고** 두 코드를 갈라낸다.
  돌아온 코드를 되짚어 추측하지 않는다.
- 잃어버린 진단 정보는 계약 밖 `Outcome.Detail`(과 엔진 스냅샷의 `ArbitrationDetail`)이
  들고 간다. 여섯 코드는 여러 원인을 한 이름으로 묶기 때문에, 그것만 남기면 운영자가
  `SEAL_MISMATCH` 하나로 일곱 가지 원인을 보게 된다.
- **재발 방지를 코드에 심었다.** `TestTheRefusalContractIsExactlyTheFrozenGolden` 이
  골든 JSON 을 파일에서 읽어 구현이 내보내는 집합과 정확히 대조한다. 매핑표가 이름 붙인
  코드와 singleton 거절 코드까지 확인한다. 사람이 옮겨 적을 자리가 없으므로 같은 실수가
  두 번 일어날 수 없다. 뮤테이션 M26(코드 하나를 골든 밖 이름으로 되돌림) KILLED.

골든 자체는 **건드리지 않았다.** 그 파일의 변경은 Manager 의 OpenSpec amendment 를
요구하며, 구현자가 계약을 자기 구현에 맞추는 것은 그 규칙을 뒤집는 일이다.

### P2 — 정상 owner scope 둘도 여전히 시장 관문에 막힌다 (확인·유지)

맞다. `TestThreeFamiliesOnOneSymbol…` 이 `entries==2` 를 기대하지만 바로 다음
`ResultAuthority` 의 `len(entries) != 1` 이 둘 다 막는다. spec.md:47 은 두 scope 가 독립
handoff 를 기다려야 한다고 쓴다. **이번 lot 에서 고치지 않는다** — 그 관문은 태스크 5.2
가 대체할 대상이고, 5.4.2 로 명시해 두었다. 현재 상태는 이전보다 **더** 보수적이므로
안전 방향이며, 리뷰도 그렇게 판정했다.

### 잔여 (P2와 별개)

- 선택 뒤 `Outcome` 이 버려져 dispatch 측이 "어떤 score/version/calibration 으로
  중재됐는지" 재검증할 수 없다. 골든의 envelope 필수 필드에 `arbitration_score_version`·
  `calibration_digest` 가 있으므로 5.1의 봉인 envelope 가 그것을 실을 자리다. 열어 둔다.

### 최종 실측

| | |
|---|---|
| 무태그 스위트 | exit 0 — 104 packages, 96 ok, FAIL 0 |
| 태그 스위트 | exit 0 — 104 packages, 97 ok, FAIL 0 |
| 뮤테이션 | 30종 전부 KILLED |
| FLM 증거 | evidence complete; 중재자 5개 함수의 분기 36개 **전부** 진입 측정됨 |

`collectMarket` 의 B2·B3·B4·B7(route/FX 미준비, loader 미구성, 중복 종목)은 이번에도 진입
0회다. 이 lot 이 만든 분기가 아니라 기존 가드이며 이전 번들에서도 같았다.

## L5 5.4.2 — 조정자가 받고 접고 줄 세운다 (Manager 기록)

### 무엇이 바뀌었나

중재를 `collectMarket` 안의 반복문에서 꺼내 시장 조정자로 옮겼다. 새 패키지
`internal/strategycoordinator` 가 큐 흡수·접기(coalescing)·소유자 범위 결정 순서·유계 인계를 맡고,
엔진 쪽 `coordinateMarketProposals` 가 그 둘 사이에서 값을 옮기고 계보 신원을 레인 권한으로 되돌린다.
`arbitrateProposalScope` 는 트리에서 사라졌다(`rg` 0건, 대조군 3건으로 도구 동작 확인).

### P0 — 생산 조정자가 매 주기 36GB 를 잡았다

`NewMarketCoordinator` 가 용량을 `1<<30` 으로 넘겼고 그 값이 그대로
`make(map[laneSlot]entry, capacity)` 에 들어갔다. 커널이 세 번 OOM 을 냈다:

```
09:51:46  Killed process (engine.test)  anon-rss 36,825,432kB
10:22:18  Killed process (engine.test)  anon-rss 36,389,960kB
10:26:33  Killed process (engine.test)  anon-rss 36,608,084kB
oom-kill: constraint=CONSTRAINT_NONE ... global_oom
```

이것은 테스트만의 문제가 아니다. `coordinateMarketProposals` 는 **제안 수집 주기마다** 이 생성자를
부른다. 같은 파일이 `Capacity = MaxOwnerScopes * LanesPerMarket`(=256)을 "서버가 정하며 호출자가
못 바꾼다"고 선언해 두고 생성자만 10억을 들고 있었다 — [[correction-unit-must-be-the-value]] 와 같은
기전이다. 선언된 값과 쓰이는 값이 갈라졌다.

수정: 생성자가 `Capacity` 를 쓰게 하고, 미리 잡는 칸 수는 상한이 아니라 `LanesPerMarket` 로 낮췄다.
상한은 넘침을 **판정**하는 수이지 미리 **잡을** 수가 아니다.

### 이 lot 의 진짜 실패는 실행하지 않은 것이었다

`TestTheProductionCoordinatorUsesTheServerOwnedCapacity` 는 **이미 있었다.** 그 단언이 바로 이
버그를 겨눈다. 그런데 맵 선할당이 단언에 닿기 전에 프로세스를 죽여서 아무도 그것을 보지 못했다.

더 근본적으로, 이 lot 을 GREEN 이라고 부른 실행은 전부 **무태그**였다.
`coordinator_test.go`·`fixture_test.go`·`a112_coordinator_test.go` 는 `tossos_testseams` 뒤에 있고,
무태그 `go test ./internal/strategycoordinator/` 의 `ok ... 0.003s` 는 골든 대조 두 건만 돈 결과다.
[[passing-test-is-not-evidence]] 그대로다 — 통과는 증거가 아니고, 특히 **돌지 않은** 통과는 아무것도 아니다.
a118 이 `make test-seams` 를 게이트에 넣어 둔 것이 이 lot 을 막았을 경로다.

### 뮤테이션 (pristine 사본 복원, SHA-256 대조 확인)

| # | mutation | result | killed by |
|---|---|---|---|
| M-C1 | 조정자 용량을 `1<<30` 으로 되돌림 | KILLED | `TestTheProductionCoordinatorUsesTheServerOwnedCapacity` |
| M-E1 | `if outcome.Overflow` → `if false` | KILLED | `TestAMarketWithMoreScopesThanTheQueueHolds…` |
| M-E2 | 넘침 결과의 `QueueDropCount` 대입 삭제 | KILLED | 같은 테스트 |
| M-E3 | `entries()` 의 `return nil, false` → `continue` | 처음 SURVIVED → 테스트 추가 후 KILLED | `TestASelectionWithNoLaneToComeBackToClosesInsteadOfShrinkingTheList` |

M-E3 가 살아남은 것은 우연이 아니라 측정과 일치했다 — 그 가드의 진입 수가 0 이었다. 가드를 지우는 대신
그것을 죽이는 테스트를 넣었다. 그 뮤테이션이 만드는 결함은 "종목이 조용히 사라지고 목록만 짧아지는 것"이라
안전 방향으로도 비울 수 없는 자리다.

### Function Logic Map

- 재생성: `collectMarket`(13→14 분기), 신규 `coordinateMarketProposals`, 신규 `strategyMarketArbitration.entries`.
- 위치만 밀린 6개(`Arbitrate`·`continueExistingOwner`·`familyScore`·`ownerSnapshotFault`·`selectHighestScore`·
  `ResultAuthority`)는 본문이 **바이트 단위로 동일함을 대조**한 뒤 해시와 좌표만 옮겼다. 측정값은 코드가
  같으므로 그대로 유효하다.
- `arbitrateProposalScope` 번들은 삭제했다. 그 함수는 트리에 없고, 동결 base 에도 없었다(파일 자체가
  이 change 가 만든 것). `revision: base` 로 남기면 base 에 존재하지 않는 것을 존재한다고 적는 셈이다.
- **귀속 완전성을 측정으로 증명했다**: 모든 분기에서 테스트별 진입 수의 합 = 스위트 전체 진입 수.
  집합 밖 테스트가 어느 arm 이든 들어갔다면 이 등식이 깨진다. 깨진 행 0.

### 남은 커버리지 구멍 (열어 둠)

- `coordinateMarketProposals` B5 (계보 신원 충돌)와 `collectMarket` B9 (`arbitration.collision`) 진입 0회.
  둘은 같은 사건의 안팎이며, 생산 경로로 만들려면 한 종목의 두 레인이 같은 봉인 계보 신원을 들고 와야 한다.
  통과가 아니라 구멍으로 적는다.
- `collectMarket` B2·B3·B4·B7(FX 미준비, loader 미구성, 빈/중복 종목)은 이번에도 진입 0회다.
  이 lot 이 만든 분기가 아니라 기존 가드이며 이전 번들에서도 같았다.

### 실측

| | |
|---|---|
| engine 태그 스위트 | PASS, 73.9% of statements — **표본 하나다.** 5.5-fix2 의 3회 재측정(2912·2909·2909 of 3941)이 이 값을 73.8~73.9% 로 정정했다; 이 줄은 그때 무엇을 적었는지의 기록으로 남긴다 |
| engine 무태그 스위트 | PASS, 63.9% of statements |
| strategycoordinator 태그 스위트 | PASS (4GB cgroup 캡 안에서) |
| gofmt (`$(go env GOROOT)/bin/gofmt`) | 위반 0 |
| check_analysis | evidence complete |

무거운 실행은 전부 `systemd-run --user --scope -p MemoryMax=… -p MemorySwapMax=0` 안에서 돌렸다.
묶지 않으면 터질 때 데스크톱이 같이 죽는다.

## L5 5.4.2 — 독립 적대적 리뷰 2차 (2026-08-30)

구현 세션과 분리된 두 리뷰어(정확성 적대적 · 안전 불변식)를 같은 diff에 붙였다.
리뷰어는 파일을 편집하지 않았고, 자기 주장을 `go test -overlay` 뮤테이션으로 스스로 검증했다.

### 안전 불변식 감사 결과

| 질문 | 판정 | 근거 |
|---|---|---|
| 실계좌 노출이 지금 0인가 | **예** | 운영 이미지(`2026-08-17` 빌드) 안 `tossctl` 에 `PROPOSAL_QUEUE_OVERFLOW`·`strategycoordinator`·`ARBITRATION_REFUSED` 문자열 0개, 양성 대조(`ROUTE_NOT_READY` 등) 1개씩. 브랜치는 main 에 안 붙었다(`main..HEAD` 9 커밋) |
| 배포돼도 경로가 열리는가 | **아니오** | 컨테이너 env 는 다섯 개뿐이고 `TOSSOS_STRATEGY_*` 가 없다 → `strategy_route_authority.go` 의 키 길이 검사가 발화해 `Ready=false`, `collectMarket` 이 새 코드보다 43줄 위에서 반환 |
| 토글 OFF 동치성 | **예** | OFF 상태 조기 반환 다섯 개가 전부 이 lot 이 만진 첫 줄보다 위에 있다. 스냅샷 구조체는 값 비교되는 곳이 없고 JSON 태그도 없다 |
| 손절·보호 경로에 얹히는가 | **아니오** | 이 경로가 낼 수 있는 주문은 `Side: "buy"` 하나다. 감축·청산은 Guardian 의 `ReductionIssuer` 가 별도 감시 루프로 낸다. 조정자는 주기마다 새로 만들어져 주기를 넘겨 지연시킬 상태가 없다 |
| 감사 추적성 | **아니오 — 깨져 있다** | 아래 P1-2 |

**토글이 없다는 사실을 그대로 적는다.** 이 코드에는 기능 토글이 없다. `collectMarket` 자체는
지금도 5초마다 불린다(휴면 worker 의 `RefreshesAuthority`). "OFF"의 정체는 서명된 매니페스트
파일과 네 개 env 가 배포에 없다는 것뿐이며, 운영자가 그것을 넣는 순간 같은 프로세스 안에서
조정자가 산다. 관문이 하나 더 있는 것이 아니다.

### 고친 것

| # | 발견 | 처리 |
|---|---|---|
| P1-1 | 큐 상한 `MaxOwnerScopes=64`(=256 칸)가 상류 매니페스트 상한 10,000 보다 39배 작다. "상류에 상한이 없어 유도할 영수증이 없다"는 주석은 **사실이 아니었다** — 영수증은 `strategyproposal/production.go` 의 `validScopes` 에 있었다. 65 종목이면 정상 매니페스트가 매 주기 시장을 닫는다 | `MaxOwnerScopes` 삭제, `Capacity = 10_000`, 영수증을 `MaxManifestScopes` 로 이름 붙여 내보내고 `receipt_contract_test.go` 가 두 패키지를 함께 읽어 같음을 강제 |
| P1-2 | 새 진단(`PROPOSAL_QUEUE_OVERFLOW`·`ArbitrationRefusal`·`ArbitrationDetail`·`QueueDropCount`)이 엔진 패키지를 못 벗어난다. 운영자 표면은 이 닫힘을 `EVIDENCE_STALE` 하나로 뭉갠다. 그런데 주석은 "이 숫자가 조용한 유실이 없음을 증언한다"고 적고 있었다 | **투영 배선은 태스크 7.3 이므로 여기서 하지 않는다.** 대신 거짓 주석을 지우고, 이 수가 지금 배선에서 **언제나 0** 이라는 사실까지 함께 적었다. 골든에 없는 투영 enum 을 지어내지 않는다 |
| P2-3 | `INTERNAL_FAILURE` 로 닫는 두 경로(계보 충돌·되돌릴 자리 없음)가 진단을 전부 버렸다. 같은 코드가 이 함수 안에서만 다섯 가지 일에 붙는다 | 두 경로에 구별되는 detail 상수와 digest·counts 를 실었다 |
| P2-4 | `Arbitrate` 가 결함을 읽고 잠금을 놓은 뒤 선택을 만든다. 그 사이 `Submit` 이 낸 결함은 이 호출에 안 보여, **이미 닫힌 조정자가 선택을 내놓는다** | 잠금을 메서드 전체로 넓혔다. 안에서 부르는 중재자는 순수 함수이고 I/O 가 없다 |
| P2-8 | 이 lot 이 `MaxManifestScopes` 를 만들면서 같은 파일 11줄 위에 벌거벗은 `10_000` 을 남겼다. 세는 것이 다른데(종목 vs 종목×레인) 눈에는 중복으로 보인다 | `productionMaxTargets` 로 따로 이름 붙이고, 왜 값이 같아도 합치면 안 되는지 적었다 |
| P3 | `len(entries)==0` 가지만 `QueueDropCount` 를 안 실었다 | 형제 넷과 맞췄다 |
| P3 | 닫힘 가지의 `RefusedCount = refused + 1` 이 규모에서 거짓말이다. 실측: 10,001 종목이 경로에 올라 하나도 못 나간 주기가 "1건 거절" | 닫힘 가지는 `RoutedCount` 를 그대로 싣는다 — 시장이 닫히면 전부 거절이다 |

### 살아남던 뮤테이션과 그것을 죽인 시험

리뷰어가 **직접 돌려서** 살아남는 것을 확인한 여섯 개다. 표의 시험은 이번에 새로 넣었고,
넣은 뒤 같은 뮤테이션을 다시 돌려 죽는 것을 확인했다.

| ID | 뮤테이션 | 전 | 후 | 죽인 시험 |
|---|---|---|---|---|
| M-G | `Capacity` 10,000 → 256 | — | KILLED | `TestTheQueueCapacityIsReadFromTheManifestScopeCeiling` |
| M-H | `existing.key != key` 절 삭제 | SURVIVED | KILLED | `TestTheSameLineageWithOnlyADifferentSnapshotDigestClosesTheScope` |
| M-I | `ProposalFamily` 가 늘 `""` | SURVIVED (engine 스위트까지) | KILLED | `TestAnAdmittedEnvelopeCarriesItsFamilyInTheDedupKey` |
| M-J | `Key.Parts()` 값 순서 뒤섞기 | SURVIVED | KILLED | `TestPartsMatchesTheGoldenFieldOrderPositionally` |
| M-K | `scopeOrderLess` 세대 갈래 → `false` | SURVIVED | KILLED | `TestScopeOrderFallsThroughToPositionGeneration` |
| M-L | 범위 안 레인 정렬 삭제 | SURVIVED | KILLED | `TestTwoFaultyLanesInOneScopeAlwaysReportTheSameRefusal` |

M-H 가 살아남던 기전은 이 저장소가 이미 기록한 것과 같다 — 기존 시험이 스냅샷과
후보 유효기한 **두 축**을 함께 흔들었고, 유효기한은 계보 신원에 들어가므로 신원 절만으로도
통과했다. 축 하나만 흔드는 시험을 새로 넣었다.

### 죽이지 못한 것 (정직하게 적는다)

- **M-F (`Arbitrate` 의 이른 잠금 해제) — SURVIVED.** 고쳤지만 **시험으로 증명하지 못했다.**
  실패 이유는 시험이 약해서가 아니라 결함이 밖에서 관측 불가능하기 때문이다. 올바른 판본에서도
  `Arbitrate` 가 경쟁에서 이기면 "닫히지 않은 결과"가 정당하게 나오므로, 그 관측만으로는 두
  판본을 가를 수 없다. 가르려면 임계 구역 안에 주입 지점이 있어야 한다.
  `TestAMarketClosedWhileArbitrateIsRunningStillReportsClosed` 는 겹침을 실제로 만들고
  race detector 로 자료 경쟁이 없음을 지키지만, 이 뮤테이션은 못 죽인다. 시험 주석에 그대로 적었다.
  확정적 증명은 8 worker/2 coordinator 하네스를 세우는 **태스크 5.7** 의 몫이다.
- **M-M (`laneOrderLess` 를 `Family` 비교만으로 축소) — SURVIVED.** 한 시장의 네 레인이 가족과
  1:1 이라 `Family` 만으로도 전순서가 되기 때문이다. 그래도 `LaneID`·`LaneVersion` 갈래를 남긴다 —
  뒤 매니페스트가 한 가족에 두 레인을 붙이면 `Family` 만으로는 전순서가 아니게 되고
  `sort.Slice` 가 불안정해져 결정성이 깨진다. 방어이며 죽은 코드가 아니다. 시험은 못 만든다.
- **`Submit` 의 계보-범위 검사** 는 중재자가 같은 판정을 같은 detail 로 내므로 블랙박스로 죽일 수 없다.
  남긴다 — 중재까지 가기 전에 잘못된 범위로 만든 열쇠가 큐에 앉는 것을 막는다.

### 구조적으로 도달 불가인 표면 — 그대로 적는다

리뷰어가 뮤테이션으로 확인했다(접힘 가지에 `panic` 을 넣어도 engine 스위트가 통과했다).

- **접힘(coalescing) 은 지금 배선에서 절대 일어나지 않는다.** `collectMarket` 이 종목 중복을 먼저
  거절하고 매니페스트가 (종목, 레인)마다 하나만 실으므로 같은 레인 칸이 두 번 오지 않는다.
- **넘침도 일어나지 않는다.** 큐 상한이 매니페스트 상한과 같아졌으므로(P1-1) 정상 매니페스트는
  정확히 들어맞고, 넘침은 상류 검증이 뚫렸을 때만 발화하는 방어다. 경계는 정확하다 —
  `validScopes` 가 `>` 로 거절하므로 10,000 개짜리 합법 매니페스트는 넘치지 않는다.
- 따라서 **`QueueDropCount` 는 지금 언제나 0** 이고, 넘침 시험은 매니페스트 검증을 우회하는
  주입 배치로만 그 가지에 닿는다. 이 셋을 "조용한 유실이 없음의 증거"로 인용하지 않는다.
  여러 worker 가 같은 칸을 두고 다투게 되는 것은 태스크 5.2·5.7 이다.

### 미룬 것 — 태스크 5.4.3 으로 명시

리뷰어가 **작동하는 시연**과 함께 낸 P1 하나를 이 lot 에서 고치지 않는다.

`batch.LanesFor(symbol)` 이 빈 목록을 돌려주는 일은 "이 종목은 원래 제안이 없다"와
"이 종목의 제안이 고장으로 사라졌다"를 구별하지 못한다. `LoadProductionAuthorityBatch`
안에는 조용히 빈 목록을 만드는 `continue` 가 일곱 개 있다. 시연 결과는 다음과 같다 —
두 종목이 경로에 올랐고 005930 의 제안만 사라졌는데, 시장이 `READY` 로 보고되고 목록이
둘에서 하나로 줄었으며, 그 결과 `len(entries) != 1` 관문이 오히려 **만족되어** 000660 이
풀렸다. 005930 이 제안을 유지했다면 풀리지 않았을 종목이다.

미루는 근거: 이 `refused++; continue` 는 5.4.2 **이전부터** 같은 자리에 있었고, 고치려면
`internal/strategyproposal` 의 배치 로더가 종목별 부재 사유를 형으로 들고 나오게 바꿔야 한다 —
5.4.2 의 범위("중재자를 조정자 런타임에 얹는다") 밖이다. 다만 이 lot 이 그 루프를 다시 쓰면서
같은 불변식을 주석으로 네 번 다시 선언했으므로, **선언만 하고 달성하지 않은 채로 두지 않는다.**
태스크 5.4.3 으로 적었다. 침묵한 생략이 아니다.

### 이번 회차 실측

| | |
|---|---|
| `make test` (무태그 전체) | PASS |
| `make test-seams` (태그 전체) | PASS |
| `make vet` | PASS |
| engine 태그 스위트 (커버리지) | PASS, 73.8% of statements |
| engine 무태그 스위트 (커버리지) | PASS, 63.6% of statements |
| `strategycoordinator`·`strategyarbiter`·`strategyproposal` `-race` | PASS |
| gofmt (`$(go env GOROOT)/bin/gofmt`) | 위반 0 |
| `check_analysis` | evidence complete |
| `make sdd-check` | PASS |
| `openspec validate --all --strict` | 59/59 |

FLM 번들 아홉 개를 다시 만들었다(engine 넷 + strategyproposal 다섯). `strategyproposal` 쪽 다섯은
본문이 안 바뀌고 const 블록이 늘어 줄만 +15 밀린 것이라 측정은 그대로 두고 위치·해시만 재기준화했으며,
그 재기준화가 온전한지는 **md 가 인용한 모든 `줄:칸` 이 그 번들 ast.json 안에 실제로 있는지**를
기계로 확인했다(미확인 0)  
> **정정(5.5-fix3):** 그 확인은 좌표의 **존재**만 봤다 — 역할은 보지 않았고, 게이트도 보지 않았다(`check_analysis.py` 의 좌표 검사는 소스 범위와 분기 개수뿐). 역할 대조는 `role_check.py` 로 들어왔다.. engine 넷은 코드가 바뀌었으므로 커버리지를 다시 측정해 표를 다시 만들었다.

## L5 5.4.3 — 사라진 제안과 없던 제안을 가른다 (2026-08-30)

5.4.2 적대적 리뷰가 시연과 함께 낸 P1 을 고친다. 그때는 범위 밖이라 태스크로 미뤘고, 그 태스크가 이것이다.

### 무엇이 문제였나

`batch.LanesFor(symbol)` 이 빈 목록을 돌려주는 일이 두 가지를 한 이름으로 뭉쳤다 —
"이 종목은 원래 제안이 없다"와 "이 종목의 제안이 고장으로 사라졌다". 뭉치면 고장 하나가
시장의 목록을 줄이고, **줄어든 목록이 `len(entries) != 1` 관문을 오히려 만족시켜** 상관없는
다른 종목을 풀어 준다. 고장이 시스템을 더 관대하게 만드는 방향이다.

### 고장의 기준을 어디에 그었나 — 그리고 처음에 틀렸다

처음 그은 경계는 "매니페스트와 자격 집합이 둘 다 받아들인 뒤에 제안을 못 만들면 고장"이었다.
`buildLaneInput` 오류 전부와 `!proposal.ValidProposal()` 전부가 고장이 되었다.

**그 경계는 틀렸고, 스스로 쓴 시험이 그것을 잡았다.** `TestAnEvaluationRefusalIsAbsenceNotFault`
가 목표가를 진입가 아래로 내리자 `LANE_INPUT_UNAVAILABLE` 고장으로 기록됐다 — 보호적이지 않은
목표가는 **매일 일어나는 정상 거절**이고 동결 골든이 `quote_fx_sizing` 으로 이름까지 붙여 둔
것이다. 그대로 냈으면 스프레드가 한 번 넓어질 때마다 그 시장이 통째로 닫혔다.
5.4.2 의 큐 상한 P1 과 정확히 같은 모양의 고장이다 — **정상 입력을 시스템이 거부하게 만드는 것.**

고쳐 그은 기준은 "제안이 안 나왔다"가 아니라 **"약속받은 봉인된 입력을 얻지 못했다"** 다.

| 건너뛰기 | 판정 | 근거 |
|---|---|---|
| 매니페스트 스코프의 종목이 이번 주기 대상이 아님 | 정상 부재 | 매니페스트는 상위집합이다 |
| `RouteSet` 을 쓸 수 없음 | **고장** | 자격이 없는 것이 아니라 자격이 있는지조차 말할 수 없다 |
| 후보 생애가 매니페스트와 다름 | 정상 부재 | 묵은 매니페스트는 흔하다. 고장으로 세면 시장이 멈춘다 |
| 자격 집합이 그 레인·캠페인을 안 받음 | 정상 부재 | 라우터가 그렇게 말한 것이다 |
| 봉인된 증거 스냅샷 재생 실패 | **고장** | 약속받은 봉인 입력이 없다 |
| `buildLaneInput` 오류 | 정상 부재 | 이 오류 공간은 돌파 증거 미구축부터 가격 관계·주간 예약까지 **시장 상태와 설정**이 대부분이다 |
| `!ValidProposal()` 이고 `Code != RefusalNone` | 정상 부재 | 평가가 이유를 들고 거절한 것이다 |
| `!ValidProposal()` 이고 `Code == RefusalNone` | **고장** | 이유 없이 봉인만 안 섰다. 방어적 검사 |

### 뮤테이션

| ID | 뮤테이션 | 판정 | 죽인 시험 |
|---|---|---|---|
| N-A | 고장을 아예 기록하지 않는다(변경 전 행동) | KILLED | `TestAProposalLostAfterAdmissionIsRecordedAsATypedFault` |
| N-B | 후보 생애 불일치를 고장으로 센다 | KILLED | `TestAScopeTheCurrentCandidateDoesNotMatchIsAbsenceNotFault` |
| N-C | 엔진이 고장을 보고도 안 닫는다 | KILLED | `TestALostProposalClosesTheMarketInsteadOfReleasingTheOtherSymbol` |
| N-D | 사라진 자리를 스냅샷에 안 적는다 | KILLED | 같은 시험 |
| N-F | 경로 권한 불가를 고장으로 안 센다 | KILLED | `TestAnUnusableRouteAuthorityIsAFault` |
| N-H | `buildLaneInput` 오류를 고장으로 센다 | KILLED | `TestABreakoutLaneWithNoEvidenceYetIsAbsenceNotFault` |
| N-J | 증거 재생 실패를 고장으로 안 센다 | KILLED | `TestAProposalLostAfterAdmissionIsRecordedAsATypedFault` |
| N-I | 평가 거절까지 고장으로 센다 | **SURVIVED** | 아래 |

**N-I 가 살아남는 이유를 탐침으로 확인했다.** `!proposal.ValidProposal()` 가지 맨 앞에
`panic` 을 넣고 두 패키지 스위트를 돌렸더니 **아무 시험도 그 가지에 닿지 않았다.**
`buildLaneInput` 이 먼저 거절하기 때문이다. 즉 두 팔을 가를 시험을 지금은 만들 수 없다.
그래도 가지를 지운다는 선택은 하지 않는다 — 생산에서 `Propose` 는 호가·FX·사이징 사유로
분명히 거절하고, 지우면 봉인 안 선 제안이 `values` 에 담겨 조정자까지 간다.
그 분류가 옳다는 근거는 시험이 아니라 **한 층 위에서 같은 판정을 시험으로 못 박았다는 것**이다.

### 실측

| | |
|---|---|
| `make test` (무태그 전체) | PASS |
| `make test-seams` (태그 전체) | PASS |
| `make lint` (gofmt + vet 양쪽 태그) | PASS |
| proposal 태그 스위트 (커버리지) | PASS, 30.1% |
| proposal 무태그 스위트 | PASS, 9.7% |
| engine 태그 스위트 | PASS, 73.8% |
| engine 무태그 스위트 | PASS, 63.6% |
| `check_analysis` | evidence complete |

FLM 번들 열하나를 갱신했다. 그중 둘은 **새로 만든 것**이고 계획에 없었다 —
`ProductionBatchAuthority.Len`(리시버 구조체에 필드가 늘어서 "수정된 기존 함수"가 되었다)과
`productionFixture`(픽스처를 두 조각으로 쪼개면서 본문이 바뀌었다). 이 저장소가 이미
기록한 대로 Logic Map 은 늘 계획보다 많다. 인용한 모든 `줄:칸` 이 해당 ast.json 안에
실제로 있는지 기계로 확인했다(미확인 0)  
> **정정(5.5-fix3):** 그 확인은 좌표의 **존재**만 봤다 — 역할은 보지 않았고, 게이트도 보지 않았다(`check_analysis.py` 의 좌표 검사는 소스 범위와 분기 개수뿐). 역할 대조는 `role_check.py` 로 들어왔다..

## L5 5.5 — 조정자와 공유 dispatch 사이의 경계에 이름을 붙인다 (2026-08-30)

### 무엇이 문제였나

조정자가 고른 것은 `strategyProposalMarketAuthority.entries` 라는 **목록**으로 나가고, 그 목록을
받는 생산 코드가 **여섯 함수**였다(`strategy_account_first_leg_authority.go` 한 파일에 서로 다른
소비자가 둘 있다 — `:151` 과 `:210`). 여섯 곳이 각자 `len(entries) != 1` 을 다시 쓰고 각자
`entries[0]` 을 다시 집었다. 동반 조건은 서로 달랐다 — worker 승격은 `snapshot.Ready` 를 보고,
`ResultAuthority` 와 읽기 전용 projection 은 보지 않았다.

같은 규칙의 사본이 다섯이면 하나를 고쳐도 넷은 옛 규칙을 유지하고, 그 차이는 아무도 보고하지
않는다. 이 저장소는 같은 기전을 이미 기록했다 — 정정의 단위는 file:line 이 아니라 **값**이다.

그리고 그 규칙은 아무 이름도 갖고 있지 않았다. 종목이 둘인 시장은 조정이 계약대로 끝나도
아무 말 없이 아무것도 내보내지 않는다. 운영자에게는 "오늘 후보가 없었다"와 구별되지 않는다.

### 무엇을 했나 — 그리고 무엇을 하지 않았나

새 파일 `internal/app/engine/strategy_dispatch_handoff.go` 하나가 그 판단을 갖는다.

```go
// 5.5 당시의 모습이다. 아래 5.5-fix 가 이 둘을 모두 없앴다 —
// 상수는 strategyhandoff.Capacity 로, 타입은 strategyhandoff.Handoff 로 갔다.
const strategyMarketHandoffCapacity = 1
func (authority strategyProposalMarketAuthority) dispatchHandoff() strategyDispatchHandoff
```

거절은 세 이름이다: `HANDOFF_MARKET_CLOSED`, `HANDOFF_NO_SELECTION`, `HANDOFF_OVER_CAPACITY`.
중재 거절 이름(`strategyarbiter.Refusal`)을 빌려 쓰지 않는다 — 중재가 실패한 것과 중재 결과를
넘기다 막힌 것은 다른 사건이고, 같은 이름을 쓰면 운영자가 둘을 구별할 수 없다.

**상한 1 은 이 태스크가 고른 값이 아니라 지금 코드에서 읽은 값이다.** 동결 골든의
`queue.market_wide_single_proposal_assumption_forbidden = true` 와
`queue.selected_limit` 원문 `"at most one selected proposal per owner scope"` 는 이 가정이 결국 사라져야 한다고 말한다.
그것을 실제로 없애는 것은 태스크 5.2 이고, 없애는 순간 종목 둘인 시장이 주문을 **두 건**
내게 된다 — 노출을 올리는 변경이다. 의존성 매트릭스는 A100 ProtectionReady 가
`UNWIRED_INCOMPLETE` 인 동안 `exposure-raising dispatch` 를 금지한다. 그래서 5.5 는 상한에
**이름을 붙였고 풀지는 않았다.** ~~이제 5.2 는 상수 하나를 고치는 일이 된다.~~
**[정정 — 아래 5.5-fix 참조]** 이 문장은 거짓이었고, 거짓인 채로 코드 주석에도
들어갔다. **[재정정 2026-08-31: "커밋 메시지에도"는 과장이었다. `8a161fc8` 의
메시지는 "그것을 실제로 푸는 것은 태스크 5.2 다"라고만 적었다 — 둘이지 셋이 아니다.]** 첫 판본은 상한을 넘지 않으면 무조건 `entries[0]` 을 건넸으므로,
상수만 2 로 올리면 종목 둘인 시장이 **승인되면서 두 번째 선택을 조용히 버렸다**.

바꾼 자리는 넷이다: worker 승격, dispatch 주기, `ResultAuthority`, 읽기 전용 projection.
`strategy_account_first_leg_authority.go` 의 사본 다섯 개는 **L6 소유**라 남겼다 —
조용히 남긴 것이 아니라 `singleProposalAssumptionCensus` 표에 태스크 이름과 함께 적었고,
여섯 번째 사본이 생기면 그 표가 깨진다.

### compile/dependency 수준 증명

| 시험 | 무엇을 고정하나 |
|---|---|
| `TestTheSingleProposalAssumptionLivesOnlyWhereTheCensusSaysItDoes` | 그 가정이 적혀 있어도 되는 파일과 그 개수. 측정값이다 — 처음 손으로 쓴 4 를 도구가 5 로 정정했다 |
| ~~`TestTheCoordinatorAndHandoffSeamsHoldNoMutationCapability`~~ | ~~seam 두 파일이 execgw·journal·orderintent 를 들여오지 않는다~~ **[5.5-fix 에서 `TestTheSeamFilesStayWithinTheirDeclaredClosure` 로 개명. 옛 이름은 거짓이었다 — 함수형 필드로 세탁한 능력은 못 잡는다]** |
| `TestTheWorkerBuilderOnlyObservesThroughTheGateway` | worker 승격 본문에서 `gateway.` 로 시작하는 호출은 전부 `Observe` 로 시작한다 |
| `TestExactlyOneProductionCallSiteTurnsAHandoffIntoADispatch` | `.dispatch(` 생산 호출 자리는 정확히 하나 **[5.5-fix 정정: 넘어가는 값의 조건이 `handoff.result` 라는 철자에서 "경계가 묶어 준 이름"으로 바뀌었다]** |
| ~~`TestNoProductionSiteReadsAHandoffWithoutAskingWhetherItWasAdmitted`~~ | ~~handoff 값을 읽는 함수는 반드시 `Admitted()` 를 부른다~~ **[삭제됨 — 5.5-fix. 이 시험은 "불렀는지"만 보았고 세 가지 철자로 우회됐다]** |
| `TestEveryMarketAuthorityCarryingEntriesAlsoReportsReady` | handoff 에 새로 생긴 준비 상태 검사가 **거부하게 될 정상 입력이 없다**는 근거 |

### 뮤테이션 — 하나가 살아남았고, 그것이 시험 하나를 더 만들었다

| ID | 뮤테이션 | 판정 | 죽인 시험 |
|---|---|---|---|
| M1 | handoff 의 준비 상태 검사를 지운다 | KILLED | `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached` |
| M2 | 상한을 하나 올린다 | KILLED | `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff` |
| M3 | 선택 없음에 상한 초과 이름을 쓴다 | KILLED | `TestAMarketThatSelectedNothingSaysSoInsteadOfBorrowingTheCapacityName` |
| M4 | 거절할 때 `pending` 을 비운다 | KILLED | `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff` |
| M6 | dispatch 호출 자리를 하나 더 만든다 | KILLED | `TestExactlyOneProductionCallSiteTurnsAHandoffIntoADispatch` |
| M7 | worker 승격 본문에서 `PlaceClaimedStrategy` 를 부른다 | KILLED | `TestTheWorkerBuilderOnlyObservesThroughTheGateway` |
| M8 | seam 파일에 execgw import 를 넣는다 | KILLED | `TestTheSeamFilesStayWithinTheirDeclaredClosure`(당시 이름 `…HoldNoMutationCapability`) |
| M9 | projection 에 옛 사본을 되살린다 | KILLED | `TestTheSingleProposalAssumptionLivesOnlyWhereTheCensusSaysItDoes` |
| M10 | 항목을 담은 채 `Ready:false` 인 시장 권한을 만든다 | KILLED | `TestEveryMarketAuthorityCarryingEntriesAlsoReportsReady` |
| M5 | worker 승격에서 `!handoff.Admitted()` 를 지운다 | **처음 SURVIVED** → 아래 |
| M11 | `ResultAuthority` 가 거절을 묻지 않고 result 를 읽는다 | KILLED | `TestNoProductionSiteReadsAHandoffWithoutAskingWhetherItWasAdmitted` |
| M12 | projection 이 거절을 묻지 않고 result 를 읽는다 | KILLED | 같은 시험 |

**M5 가 살아남은 이유가 이 태스크에서 가장 중요한 발견이다.** worker 승격에서 handoff 거절
검사를 지워도 그때 있던 시험이 전부 통과했다 — 거절된 handoff 의 `result` 가 영값이라 바로
아래 `!result.ValidProposal()` 이 대신 걸러 주기 때문이다. 즉 그 자리의 안전을 지키고 있던
것은 handoff 가 아니라 **우연**이었다. 거절일 때도 값을 채우도록 누가 바꾸면 그 우연은
사라지고, 어떤 **동작** 시험도 그것을 잡지 못한다. 그래서 동작이 아니라 쓰여 있는 것을 보는
`TestNoProductionSiteReadsAHandoffWithoutAskingWhetherItWasAdmitted` 를 추가했고, M5·M11·M12 가
그 시험 하나에 죽는다.

원복은 전부 SHA-256 대조로 확인했다 — `git checkout` 은 커밋 전 GREEN 까지 지운다.

### 실측 — 그리고 커버리지 공백을 숨기지 않는다

| | |
|---|---|
| `make test` (무태그 전체) | PASS |
| `make test-seams` (태그 전체) | PASS |
| `make lint` (gofmt + vet 양쪽 태그) | PASS |
| engine 태그 스위트 (커버리지) | PASS, 73.9% — **표본 하나다.** 5.5-fix2 의 3회 재측정(2912·2909·2909 of 3941)이 이 값을 73.8~73.9% 로 정정했다; 이 줄은 그때 무엇을 적었는지의 기록으로 남긴다 |
| engine 무태그 스위트 | PASS, **63.4~63.5%** — 같은 바이너리·같은 트리에서 4회 측정해 2502/2502/2498/2499 of 3943 이 나왔다. 이 수는 실행마다 흔들린다(스케줄링 의존 시험이 무태그 쪽에 있다). 표본 하나를 안정된 값으로 적었던 것을 정정한다 |
| `check_analysis` | evidence complete |
| 인용 좌표 검증 | 미확인 0 — **정정(5.5-fix3):** 이 "미확인 0" 은 좌표가 **존재하는지**만 본 결과다. 좌표가 무엇을 가리키는지는 그때 아무도 보지 않았고, 게이트도 보지 않았다. 역할 대조는 `tools/logic-map/role_check.py` 가 들어온 지금부터 성립하며, 그것을 이 저장소 401개 번들에 처음 돌렸을 때 잘린 호출 표 3개가 나왔다 |

분기 귀속은 시험별 프로파일의 합이 스위트 전체와 정확히 같다(MISMATCH 0). 그 측정이 드러낸 것:

- **`runProductionStrategyMarketCycle` 의 다섯 분기 전부 진입 0.** 두 스위트 어느 쪽도 이
  함수에 들어오지 않는다. 즉 공유 dispatch 로 가는 유일한 자리의 근거는 실행이 아니라
  AST 가드뿐이다. 이 공백은 이 태스크가 만든 것이 아니고(이 함수는 `*Context` 와 살아 있는
  journal·gateway 를 요구한다) 주인은 태스크 5.7 이다.
- **`strategyProjectionFromAssembly` 의 여덟 분기 전부 진입 0.** 트리 안에
  `StrategyEntryProductionAssembly` 를 만드는 시험이 하나도 없다. 주인은 태스크 7.3 이다.
- worker 승격의 B3·B5·B6(봉인 깨진 제안, 진입 게이트 관측 실패, digest/만료) 진입 0. 주인은 5.7.

진입 0 은 통과가 아니다. 적지 않으면 구조 가드가 동작 증거처럼 읽힌다.

### 남긴 것

- 상한 자체(`market_wide_single_proposal_assumption_forbidden`)를 푸는 일 — 태스크 5.2.
- handoff 거절을 운영자 화면에 비추는 일 — 태스크 7.3. 지금 `pending` 과 거절 코드는 값으로만
  존재하고 어떤 표면에도 나가지 않는다.
- `strategy_account_first_leg_authority.go` 의 사본 다섯 — L6(태스크 6.2).
- 독립 적대적 리뷰는 아래 5.5-fix 에서 받았다. P1 세 건이 나왔고 전부 고쳤다.

## L5 5.5-fix — 적대 리뷰가 찾은 P1 세 건 (2026-08-30)

세 리뷰어가 read-only 로 검토했고 P0 은 없었다. 배포된 동작에 결함도 없었다 — 진리표상
갈라지는 유일한 행 `(Ready=false, len=1)` 은 도달 불가이고(항목을 채우는 유일한 return 이
`Ready: true` 를 함께 박는다), 주문 경로는 시장당 1건 그대로였다.

문제는 전부 **내가 쓴 문장이 내가 만든 검사보다 컸다**는 데 있었다.

### P1-1 — 건네줄 수 있는 것보다 많이 승인했다

첫 판본은 상한을 넘지 않으면 무조건 `entries[0]` 을 건넸다.

```go
if pending > strategyMarketHandoffCapacity { return ...OverCapacity }
return strategyDispatchHandoff{result: authority.entries[0].authority.Proposal(), ...}
```

`strategyMarketHandoffCapacity` 를 2 로 올리면 종목 둘인 시장이 **승인되면서 두 번째 선택을
거절 이름도 계수기도 없이 버린다.** 이 change 가 없애겠다고 선언한 바로 그 "조용한 유실"을
seam 안에 새로 만든 것이었다. 그리고 내 주석·`review.md`·커밋 메시지가 셋 다
"이제 5.2 는 상수 하나를 고치는 일이 된다"고 적어 **미래의 유지보수자를 그 함정으로
초대**하고 있었다. 리뷰어가 실제로 상수를 올려 보았고, 트리 안 어떤 시험도 유실을 잡지 못했다.

**고친 방법 — 승인과 전달을 한 서명으로 묶었다.**

```go
func (handoff Handoff) Single() (strategyflow.Result, bool)  // 실린 것이 하나가 아니면 값을 안 준다
```

상한만 올리면 그날 `Single` 이 값을 내주지 않고, dispatch 는 하나를 내보내는 대신
**아무것도** 내보내지 않는다. 조용히 하나를 버리는 것보다 낫고, 그것이 다중 dispatch 를
구현하라는 신호다. `Admit` 이 만들 수 없는 상태이므로 시험이 그 상태를 직접 만들어 고정한다
(`TestARaisedCapacityHandsOffNothingRatherThanDroppingTheSecondSelection`) — 만들 수 없다고
검사를 빼면, 상수를 올린 사람이 그 순간 아무 시험도 안 깨지는 것을 보게 된다.

### P1-2 — 가드가 "물었는지"만 보고 "막았는지"는 안 봤다

`TestNoProductionSiteReadsAHandoffWithoutAskingWhetherItWasAdmitted` 는 `Admitted()` 가
**불렸는지**만 확인했다. 세 가지 우회가 전부 살아남았고 두 스위트가 모두 초록이었다:
`_ = handoff.Admitted()` 로 호출만 남기고 관문 삭제, `var` 선언 바인딩(`*ast.AssignStmt` 만
보므로 안 보임), 그리고 `p.dispatchHandoff().result` 처럼 바인딩 없이 바로 읽기. 게다가
구조체 리터럴을 `handoff` 로 이름 짓고 `result` 필드를 달면 경계를 거치지 않은 값이
dispatch 까지 갔다. 나는 그 시험이 M5·M11·M12 를 죽인다고 적었지만, 죽는 것은 *호출 삭제*
한 형태뿐이었다.

**고친 방법 — 가드를 지우고 타입으로 바꿨다.** 제안이 경계 밖으로 나가는 길은 `Single()`
하나뿐이고 그 서명이 bool 을 함께 준다. 답을 안 받는 철자는 `_` 하나뿐이며(한 값짜리 문맥은
컴파일이 안 된다) 그 하나를 `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer` 가 막는다.
받아 놓고 안 쓰면 컴파일러가 "declared and not used" 로 막는다. 새 접근자를 달아 우회하는
문은 `internal/strategyhandoff/escape_test.go` 가 잠근다(Result 를 내주는 메서드는 반드시
두 값을 돌려주고, `Handoff` 는 공개 필드를 갖지 않는다).

### P1-3 — 능력 폐쇄를 틀린 영수증에서 읽었다

금지 import 목록을 `strategy_dispatch_cycle.go` 의 import 에서 베껴 왔다. 그건 **그 파일의**
영수증이지 이 모듈의 변경 표면이 아니다. `internal/official`(PlaceOrder/CancelOrder/
ModifyOrder + 조건부 주문 3종)이 통째로 빠져 있었고, 더 나쁜 건 **import 없이도 뚫린다**는
점이었다 — seam 파일에 `officialBroker` 를 받는 함수를 넣으면 `b.off.CancelConditionalOrder`
로 보호 손절을 취소하면서 "변경 능력 없음" 판정을 받았다.

**고친 방법 — 판단을 패키지 밖으로 옮겼다.** 같은 패키지 안에서는 이 약속을 증명할 수 없다.
새 `internal/strategyhandoff` 는 `internal/strategyflow` 하나만 들여오고, 그 사실을 두 시험이
지킨다: 직접 import **허용 목록**(금지 목록은 빠뜨린 것을 조용히 통과시킨다)과 `go list -deps`
전이 폐쇄 + 양성 대조. 엔진에 남은 것은 어댑터 한 함수뿐이고, 그 파일의 검사도 허용 목록으로
바꿨으며 같은 패키지 능력은 **손으로 적은 이름이 아니라 계산한 타입 집합**으로 막는다
(필드 타입을 따라 고정점까지; 양성 대조로 `officialBroker` 와 `Context` 가 잡히는지 확인한다).

### 무엇이 바뀌었고 무엇이 안 바뀌었나

동작은 그대로다. `Capacity` 는 여전히 1 이고, 시장당 dispatch 는 여전히 최대 1건이며,
`entries` 를 채우는 경로도 그대로다. 바뀐 것은 **틀릴 수 있는 방법**이다.

| | 전 | 후 |
|---|---|---|
| 상한을 올리면 | 두 번째 선택이 조용히 사라진다 | 아무것도 안 나간다(fail-closed) |
| 거절을 안 보고 값 읽기 | 세 가지 철자로 우회 가능 | `_` 하나뿐이고 그것을 막는다 |
| seam 의 변경 능력 | 파일 단위 금지 목록, in-package 우회 가능 | 패키지 밖 import 폐쇄(전이+양성대조) |
| 단일 제안 가정이 적힌 곳 | seam 1 + first-leg 5 | first-leg 5 (dispatch 경로에서 사라짐) |

### 반증 실측 — 12/12 KILLED

| ID | 뮤테이션 | 판정 | 죽인 시험 |
|---|---|---|---|
| P1a-M1 | seam 파일이 `officialBroker` 를 손에 쥔다(**리뷰어가 시연한 그 우회**) | KILLED | `TestTheSeamFilesStayWithinTheirDeclaredClosure`(당시 이름 `…HoldNoMutationCapability`) — 5.5 판본에서는 **통과했다** |
| M2 | seam 파일이 `internal/journal` 을 들여온다 | KILLED | 같은 시험(허용 목록) |
| M3 | `dispatch` 를 메서드 값으로 꺼내 둔다 | KILLED | `TestExactlyOneProductionCallSiteTurnsAHandoffIntoADispatch` |
| M4 | 경계가 묶어 주지 않은 이름을 dispatch 에 넘긴다 | KILLED | 같은 시험 |
| M5 | worker 승격 관문을 지우고 `_ = handedOff` | KILLED | `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer` — 5.5 판본에서는 **SURVIVED** |
| M6 | dispatch 주기 관문을 지우고 `_ = handedOff` | KILLED | 같은 시험 |
| M7 | `ResultAuthority` 관문을 지우고 `_ = handedOff` | KILLED | 같은 시험 |
| M8 | projection 관문을 지우고 `_ = handedOff` | KILLED | 같은 시험 |
| M9 | 어댑터가 준비 상태를 안 본다 | KILLED | `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached` |
| M10 | 어댑터가 첫 항목만 싣는다 | KILLED | `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff`, `TestARefusedHandoffLeavesTheWorkerDormant` |
| M11 | `Single` 이 개수 검사를 그만둔다 | KILLED | `TestARaisedCapacityHandsOffNothingRatherThanDroppingTheSecondSelection` |
| M12 | `Admit` 이 상한 검사를 그만둔다 | KILLED | 세 시험 |

M1 과 M5 가 이 로트의 요점이다. **둘 다 5.5 판본에서는 통과했고**, 그것이 리뷰가 P1 로 든 이유다.

원복은 `try/finally` 안에서 하고 12개 전부 sha256 대조했다(`unrestored_files: []`).
앞선 배치는 하니스 타임아웃이 **중간에 죽이면서 뮤테이션 하나를 트리에 남겼고**, 이어진
배치가 그 오염된 baseline 위에서 쟀다 — `PATTERN-ABSENT` 한 줄이 그 유일한 신호였다.
그 판을 버리고 깨끗한 트리에서 다시 쟀다.

### 이 검사들이 증명하지 못하는 것

- `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer` 는 답이 **어떤 if 조건 안에 나오는지**만
  본다. 그 조건이 값을 쓰기 전에 반드시 돌아간다는 **지배 관계**는 증명하지 않는다. 진짜
  지배를 증명하려면 `go/types` 와 제어 흐름 그래프가 필요하다. 여기서 막는 것은 실제로
  일어난 실패 방식이다.
- 엔진 안에 남은 어댑터 파일의 검사는 **파일 규율**이지 컴파일러 보장이 아니다. 컴파일러
  보장이 필요한 자리 — 무엇이 건너갈지 고르는 판단 — 를 그래서 패키지 밖으로 옮겼다.
- `HANDOFF_NO_SELECTION` 은 여전히 생산에서 도달 불가이고(`Ready==true ⟹ len≥1`),
  `HANDOFF_MARKET_CLOSED` 는 여전히 여덟 가지 사유를 한 이름으로 뭉친다. 리뷰가 P1/P2 로
  든 것이고 이 로트에서 **고치지 않았다** — 사용자가 P1 셋만 지시했다. 남긴 것에 적는다.

### 실측

| | |
|---|---|
| `make test` (무태그 전체) | PASS |
| `make test-seams` (태그 전체) | PASS |
| `make lint` (gofmt + vet 양쪽 태그) | PASS |
| `make sdd-check` | PASS |
| `make validate` | PASS |
| engine 태그 스위트 | PASS, 73.9% — **표본 하나였다**. 아래 5.5-fix2 에서 3회 재측정해 2912·2909·2909 of 3941(73.8~73.9%)로 정정한다 |
| engine 무태그 스위트 | PASS, 63.5~63.6% (2498~2502 of 3936, 5회 측정) — 5.5-fix2 재측정은 2498~2501, 즉 63.5% 한 자리다 |
| `internal/strategyhandoff` | PASS, 11 시험 |
| `check_analysis` | evidence complete |
| 인용 좌표 검증(이 로트가 만진 4 파일) | 미확인 0 — **정정(5.5-fix3):** 이 "미확인 0" 은 좌표가 **존재하는지**만 본 결과다. 좌표가 무엇을 가리키는지는 그때 아무도 보지 않았고, 게이트도 보지 않았다. 역할 대조는 `tools/logic-map/role_check.py` 가 들어온 지금부터 성립하며, 그것을 이 저장소 401개 번들에 처음 돌렸을 때 잘린 호출 표 3개가 나왔다 |
| 뮤테이션 | 12/12 KILLED, `unrestored_files: []` |
| 분기 귀속 등식 | MISMATCH 0 |

### 남긴 것 (P2 — 이 로트에서 고치지 않았다)

- `HANDOFF_NO_SELECTION` 도달 불가, `HANDOFF_MARKET_CLOSED` 가 여덟 사유를 뭉치는 것.
- `Refusal()`·`Pending()` 을 읽는 생산 코드가 아직 없다 — 태스크 7.3.
- ~~이 change 의 다른 여덟 번들이 `Pinned revision: current` 라고 하면서 `source_sha256` 이
  뒤처져 있다.~~ **[정정 2026-08-31: 거꾸로였다. 130 개 번들을 세어 보니 stale 은 14 개이고
  그 14 개는 **전부 base revision 을 선언한다**. `current` 라면서 stale 인 번들은 0 개다.
  세지 않고 적었다.]**
- 골든을 한국어로 의역한 자리들 — 이 로트가 만진 문서에서는 원문으로 바꿨다.

## L5 5.5-fix2 — 재검토가 P1 다섯을 더 찾았고, 뿌리는 하나였다 (2026-08-31)

계약대로 같은 리뷰어 셋이 `eb072255` 를 read-only 로 재검토했다. **P0 없음.** 동작은
안전했다 — 진리표 8행이 fix 전후 바이트 동일, 시장당 주문 최대 1건, 5.5-fix 의 P1-1
(상한 함정)은 `Capacity=2` 에서 주문 0 으로 **닫힌 것이 확인됐다**.

그런데 P1-2(가드)와 P1-3(능력 폐쇄)은 닫히지 않았다. 리뷰어 B 가 다섯 우회를 시연했고
**전부 두 스위트가 초록이었다**. 리뷰어 A·C 도 그중 하나(경계 탈출)를 각자 독립으로 찾았다.

### 뚫린 다섯과, 그 다섯이 같은 병이라는 것

| | 우회 | 왜 통과했나 |
|---|---|---|
| B-1 | `if !handedOff { }` — 관문을 지우고 조건만 남긴다 | 검사가 "답이 어떤 if 조건에 나오는가"만 물었다. 게다가 더 강한 `switch` 형태를 **거부**했다 |
| B-2 | `var result, _ = …Single()` | 검사가 `*ast.AssignStmt` 만 보았다. **같은 지적을 두 번 받고도 안 고쳤다** |
| B-3 | 묶인 이름을 `entries` 에서 다시 덮어쓴다 | 이름은 맞고 값이 원본이다. 이름 검사로는 볼 수 없다 |
| B-4 | `func Escape(h Handoff) Result` | 검사가 **메서드**만 보았다. `Into(dst *Result)`·`Any() any`·`map[int]Result`도 통과 |
| B-5 | `run func() error` 에 `*officialBroker` 를 클로저로 담는다 | 고정점 계산이 `TypeSpec` 의 타입 표현식만 본다. import 0, 이름 0 으로 보호 손절 취소 |

B-1 과 B-3 을 합치면 한 편집으로 `Capacity`·거절 이름 셋이 모두 우회되고 상한 초과 시장이
마지막 가족의 제안을 dispatch 한다.

**뿌리는 하나다: 값·역할을 물어야 할 자리에 이름·철자를 물었다.** 실제로 닫혀 있던 것들은
전부 타입이나 패키지 경계가 막은 것이었다(`p.dispatchHandoff().result` 는 컴파일이 안 된다).

### 무엇을 했나

가드를 더 쓰지 않았다. **검사가 필요 없게 만들거나, 열거가 아닌 사실로 바꿨다.**

1. **dispatch 경로에서 답을 없앴다.** `Handoff.Deliver(func(strategyflow.Result) error) error`.
   몸통이 도는지는 경계가 정한다 — 이 함수에 무시할 boolean 이 없다. B-1·B-2 가 여기서 끝난다.
2. **`entries` 식별자를 소비자 넷의 본문에서 금지했다.** Go 에서 비공개 필드를 읽는 철자는
   그 이름 하나뿐이고 `reflect` 로도 값을 꺼낼 수 없다(`Field(i).Interface()` 는 패닉한다).
   그래서 이 금지는 열거가 아니라 **완전한 사실**이다. 양성 대조로 어댑터가 실제로
   `entries` 를 읽는 것을 확인한다 — 안 읽으면 스캔이 고장난 것이다. B-3 이 여기서 끝난다.
3. **패키지의 공개 표면 전체를 이름과 서명까지 고정했다**(`exportedSurface`). 모양을 세지
   않는다 — 새 탈출구는 예외 없이 새 공개 이름을 요구하므로 표와 어긋난다. B-4 가 끝난다.
4. **B-5 는 고치지 않고 주장을 철회했다.** 같은 패키지 안에서는 이 성질을 시험으로 배제할
   수 없다. 시험 이름을 `TestTheCoordinatorAndHandoffSeamsHoldNoMutationCapability` 에서
   `TestTheSeamFilesStayWithinTheirDeclaredClosure` 로 바꿨다 — 옛 이름은 거짓이었다.
   무엇이 결정적이고(파일 단위 import 허용 목록) 무엇이 부분적인지(계산된 능력 타입)를
   주석에 적었다. 정말 필요한 자리는 이미 패키지 밖으로 옮겨져 있다.
5. **`Refusal()` 과 `Single()` 이 한 술어(`refusalNow`)를 쓴다.** 5.5-fix 는 둘이 갈라져서,
   상한을 올리면 주문은 안 나가는데 이름은 "승인"이라고 보고했다. 그 상태에 `OverCarried`
   라는 이름을 줬다 — 무명 fail-closed 는 이 change 가 없애려던 것과 같은 모양이다.
6. **좌표 검사기를 역할까지 보게 고쳤다.** 앞선 판본은 "그 좌표가 AST 어딘가에 있는가"만
   물었고, 그래서 projection 번들의 `Exact AST return positions: 155:2` 를 통과시켰다 —
   155:2 는 반환이 아니라 **함수의 end 위치**로 존재했다. 존재 검사가 역할 검사 자리에
   서 있었던 것이고, 그것이 위의 다섯과 정확히 같은 병이다. 새 검사기는 반환·분기·호출
   표를 각각의 역할로 대조하고 `파일.go:줄` 인용까지 본다. 돌리자마자 리뷰어 C 가 손으로
   찾은 두 좌표를 **그대로** 재현했고(`155:2`, `a112_dispatch_handoff_test.go:112`),
   고친 뒤에는 이 로트가 새로 만든 오류 하나(B4 분기 위치)도 잡았다.

### 반증 실측 — 13 중 12 KILLED

| ID | 뮤테이션 | 판정 | 죽인 시험 |
|---|---|---|---|
| P1b-B1 | `Deliver` 를 `Single()` 로 되돌리고 관문을 빈 `if` 로 | KILLED | `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer`(묶는 자리 3→4) |
| P1b-B2 | `var result, _ = …Single()` | KILLED | 같은 시험 |
| P1b-B3 | `result` 를 `entries` 마지막 값으로 덮어쓴다 | KILLED | `TestSeamConsumersCannotReadTheRawEntryListAgain` |
| P1b-B4 | `func Escape(h Handoff) Result` | KILLED | `TestThePackageExposesExactlyTheSurfaceTheSeamNeeds` |
| P1b-B5 | `func() error` 로 능력 세탁 | **SURVIVED** | — 위 4번에서 주장을 철회했다. 숨기지 않는다 |
| P1b-R1 | seam 파일이 `officialBroker` 를 쥔다 | KILLED | `TestTheSeamFilesStayWithinTheirDeclaredClosure` |
| P1b-R2 | seam 파일이 `internal/journal` 을 들여온다 | KILLED | 같은 시험 |
| P1b-R3 | 어댑터가 첫 항목만 싣는다 | KILLED | `TestAMarketWithTwoSelectedScopesNamesWhyNothingWasHandedOff` |
| P1b-R4 | 어댑터가 준비 상태를 안 본다 | KILLED | `TestAClosedMarketHandsOffNothingEvenWhenAnEntryIsStillAttached` |
| P1b-R5 | `Refusal()` 이 `Single()` 과 갈라진다 | KILLED | `TestARaisedCapacityHandsOffNothing…` |
| P1b-R6 | `Deliver` 가 거절을 무시한다 | KILLED | `TestDeliverRunsTheBodyOnlyWhenSomethingCrossedTheSeam` |
| P1b-R7 | 시장 권한을 위치 인자 리터럴로 만든다 | KILLED | `TestEveryMarketAuthorityCarryingEntriesAlsoReportsReady` |
| P1b-B4b | `Single` 이 bool 을 뗀다 | KILLED | 컴파일 실패 + 표면 표 |

원복은 `try/finally` 안에서 하고 sha256 으로 대조했다(`unrestored_files: []`, `leftover_files: []`).

### 리뷰어 C 가 잡은 증거 오류 — 전부 정정

| 주장 | 실측 |
|---|---|
| 인용 좌표 미확인 0 | **거짓**. 둘 틀렸다. 검사기가 역할을 안 봤다(위 6번) |
| 태그 2912/3941, "안정적" | **2912·2909·2909** — 태그도 흔들린다. 표본을 안정값으로 적었다 |
| 무태그 2498~2502 (63.5~63.6%) | 재측정 **2498·2498·2499·2500·2501** = 63.5% 한 자리 |
| "다른 여덟 번들이 current 인데 stale" | **0 건**. stale 14 개는 전부 base 선언. 세지 않고 적었다 |
| "주석·review.md·커밋 메시지 셋 다" | **둘**. `8a161fc8` 메시지에는 그 문장이 없다 |
| 5.5 절의 코드 블록·시험 표 두 행 | 지금 트리에 없는 이름들. 취소선과 함께 표시 |
| `tasks.md` "only inside that package" | 여섯 중 둘은 아직 engine 안(다섯 자리). 문장을 정밀화 |
| 가드 시험 안의 골든 한국어 의역 | 원문 병기로 교체 |
| `reflect.TypeOf` 가 "값을 만들지 않는다" | 영값을 **만든다**. 실질 주장(채워서 넣는 시험이 없다)은 유지 |
| 뮤테이션 ID 충돌(12 아래 14 개) | 이 로트는 `P1b-` 접두사로 분리 |

### 실측

| | |
|---|---|
| `make test` / `make test-seams` / `make lint` / `make sdd-sync` | PASS |
| `make sdd-check` / `openspec validate` | a117 중복 스텁(병행 세션)에서만 실패. 그 스텁을 치우면 tracker current, 59/59 |
| engine 태그 스위트 | PASS, **73.8~73.9%** (2912·2909·2909 of 3941, 3회) |
| engine 무태그 스위트 | PASS, **63.5%** (2498·2498·2499·2500·2501 of 3936, 5회) |
| `internal/strategyhandoff` | PASS, 17 시험 — **정정(5.5-fix3):** 그때 최상위 시험은 **16**개였다(17 은 셈 착오). fix3 이후 20개(handoff 15 · escape 2 · dependency-closure 3) |
| `check_analysis` | evidence complete |
| 역할 대조 좌표 검증(4 번들) | 문제 0 / 반환 18 · 분기 40 · 호출 54 · 파일:줄 11 |
| 뮤테이션 | 13 중 12 KILLED, 1 은 철회한 주장 |

`make test` 가 한 번 네 개 실패했는데 원인은 코드가 아니라 **트리**였다: 적대 리뷰어를
`.claude/worktrees/` 안에서 돌렸더니 저장소를 훑는 정적 검사 넷이 코드베이스 사본 셋을
더 보았다. 워크트리를 지우고 다시 돌려 통과했다.

### 남긴 것

- **B-5(같은 패키지 능력 세탁)**: 시험으로 배제할 수 없다. 남은 어댑터 한 함수마저
  `internal/strategyhandoff` 로 옮기면 import 폐쇄가 결정적이 되지만, 그것은
  `strategyProposalMarketAuthority` 를 패키지 밖으로 노출하는 설계 변경이라 L6/5.2 의 몫이다.
- `HANDOFF_NO_SELECTION` 도달 불가, `HANDOFF_MARKET_CLOSED` 가 여덟 사유를 뭉치는 것 —
  `Admit(ready bool, …)` 이 bool 만 받으므로 개명은 이제 서명 변경을 요구한다(7.3).
- `Refusal()`·`Pending()` 을 읽는 생산 코드가 아직 없다(7.3).
- 이 로트도 재검토를 받아야 한다.

## 5.5 현재 상태 — 아래 fix 절들보다 이것을 먼저 읽는다

fix3~fix13 은 **적대 리뷰 13 라운드의 시간순 기록**이다. 나중 절이 앞 절을 정정하는
곳이 많으므로, 지금 무엇이 서 있는지는 그것들을 diff 해서 재구성하지 말고 여기서 읽는다.
자세한 것은 `analysis/function-logic/internal-app-engine--productionstrategyfirstlegauthorityloader.collectstrategyfirstlegauthority/branch-test-map.md`
의 B3 절과 tasks.md 의 6.2 선결조건에 있다.

**무엇이 오늘 주문을 막는가.** `strategy_account_first_leg_authority.go` 가 건너온
값을 믿지 않고 자기 권한 쌍에서 제안을 다시 꺼내 봉인된 identity 두 개를 대조한다 —
개수 관문 `:217`, 재유도 `:221`–`:222`, 비교 `:223`–`:225`.

**무엇이 그것을 고정하는가.** 셋이 함께여야 닫힌다.

| | 어디 |
|---|---|
| 행동 — 이 가드가 돌고 거절한다 | 시험 함수 넷(하위 여섯), `strategy_first_leg_identity_backstop_test.go` |
| 구조 — 이 가드가 봉인된 identity 를 **그대로** 비교한다 | 단언 하나, `strategy_first_leg_backstop_shape_test.go` |
| 해시 완전성 — 그 identity 가 **모든 필드**를 담는다 | 32 + 8 케이스, `strategyflow/execution_terms_identity_fields_test.go` 와 `weeklyvaluelane/execution_policy_identity_fields_test.go` |

행동만으로는 끝나지 않는다는 것이 이 13 라운드의 결론이다. identity 가 담는 스칼라가
32 개이고, 라운드마다 축 하나씩을 시험으로 사면 다음 축이 항상 남았다.

**아직 열려 있는 것.** 셋 다 이름만 적혀 있고 닫히지 않았다.

1. `dispatchHandoff` 수신자 위조. 어떤 철자 census 로도 못 닫는다고 fix4 가 물렸고
   그대로다. 13 라운드가 굳힌 것은 **backstop 이지 seam 이 아니다.** 진짜 해법
   (`Admit` 이 조정자만 만들 수 있는 봉인 값을 받는 것)은 6.2 의 차단 선결조건이다.
2. `revision: base` 번들 15 개가 소스 해시에 묶이지 않는다(13 개는 `base-commit.txt`
   와도 불일치). "133/133 이 열거한다"는 표 모양의 참이지 좌표의 참이 아니다.
3. `strategyflow` 에 exported 표면 census 가 없다. 이제 봉인된 `Result` 를 만드는
   함수가 태그 아래 둘이다.

**사람이 확인해야 하는 것.** 리뷰로는 못 정한다.

- 1 번을 연 채로 5.5 를 내보내는 것을 받아들이는지.
- ~~CI(`.github/workflows/ci.yml`)가 `make sdd-test` 도 `check_analysis` 도 안 돌린다.~~
  **2026-09-02 에 사람이 "배선"을 골라 절반을 닫았다.** `sdd-checks` job 이 새로 생겨
  `make sdd-check-ci` 를 부른다 — `sdd-check` 에서 **러너로 옮길 수 없다고 측정된 둘**
  (`sdd-doctor` 는 로컬 설치 CLI 를, `check_index_freshness.py` 는 codegraph 실행 파일과
  gitignore 된 `.sdd/index-state.json` 을 본다. 외부 도구를 지운 PATH + depth-1 클론에서
  각각 exit 1)만 뺀 부분집합이다. 이제 CI 가 파이썬 시험 180개 — 그중 `tools/logic-map`
  77개가 여덟 라운드 분량의 `role_check.py`·`check_analysis.py` 열거 강제다 — 와
  `go test ./tools/logic-map` 을 돈다. 목록이 두 곳으로 갈라지는 것은
  `tools/sdd/test_ci_runs_portable_sdd_checks.py` 가 막는다(뮤테이션 5종으로 반증 확인).
  **안 배선한 것은 `check_analysis --change <id>` 자체다.** 그 검사는 번들을 유도 당시
  소스에 묶으므로 change 하나의 완료 게이트에서만 참이다 — 측정하니 활성 change 31 개 중
  통과가 **1 개(a112)**, 나머지 30 개는 AST 소스 해시 stale 15 · 넓어진 수정 집합의 FLM
  누락 11 · base-commit 누락/무효 4 였다. 저장소 전체로 켜면 첫날부터 빨갛고, PR 이
  건드린 change 만 골라 켜도 위 2 번과 같은 뿌리의 빚으로 남의 세션이 막힌다.
- ~~외부 `a117` 스텁이 `make sdd-check` 를 깨고 있다.~~ **2026-09-02 에 사람이 소유를
  넘겨 처리했다** — `a119-…` 로 renumber 하고 `STORY-TOS-a119`(FEAT-TOS-001)·registry·
  generated tracker 를 만들었다. Story 의 acceptance 는 그 세션의 proposal.md 에서 옮겨
  적은 것이고 내용의 주인은 여전히 그 세션이다. `make sdd-check` 는 이제 통과한다.
  남은 것은 그 스텁에 `specs/` 델타가 없어 `make validate --all` 이 1 건 실패하는 것.
- `make gate` 는 이 change 에서 한 번도 돈 적이 없다(완료 게이트, 2/10 에서 멈춘다).

**이 기록에서 가장 옮겨 쓸 만한 것 — 반증 방법이 틀린 두 사례.**

1. `go test -overlay` 는 **디스크를 읽는 파서에 안 보인다.** 소스를 읽는 시험을
   overlay 로 반증하면 전부 GREEN 이 나오고, 그것은 "가드가 강하다"가 아니라
   **"안 쟀다"** 이다. 파일을 실제로 바꾸고 해시로 복원해야 한다.
2. 리네임 뮤테이션이 `*.go` 를 통째로 훑으면 **시험 파일까지 고쳐** 자기를 잡아야 할
   시험을 수리한다. 뮤테이션의 범위는 생산 코드로 한정해야 한다.

둘 다 "통과했다"를 근거로 쓰기 전에 **반증이 실제로 그 축을 흔들었는지** 확인해야
한다는 같은 규칙의 사례다.

## 5.5-fix3 — 3라운드 적대 재검토와 그 수정

### 리뷰어 셋이 같은 구멍을 세 가지 철자로 뚫었다

서로를 못 보는 리뷰어 셋에게 로트 `8f8b1178` 을 read-only 로 맡겼다. **P0 0건.** A 는
돈 경로 재작성이 행동상 동일함을 반환 집합 7행·CAS 4행 표로 확인했고
(`Claimed || (State != "FLAT" && State != "CLOSED")`, `&&` 가 더 강하게 묶는다;
`Proposal()` 은 순수 접근자라 단락 평가에 부작용이 없다), `deliverable == 1` 이므로
`refusalNow` 가 옛 술어와 동치임을 보였다.

**P1 은 셋 다 같은 것이었다.** 각자 `rawSelection()`, 새 파일의 `relay()`,
Deliver 몸통 안의 `rawTailProposal()` 로 경계를 우회했고 셋 다 두 스위트가 초록이었다
(A 630 passed / B 1323 of 1393, baseline 과 동일 / C 새 실패 0). 뿌리는 2라운드와 같다 —
`guard_test.go:599` 가 **이름 소속**을 물었고, 같은 파일 `:667` 이 "이름을 보는 어떤 검사도
그것을 보지 못한다"고 스스로 적어 두었다. `entries` 금지는 **네 함수의 본문**만 셌으므로
헬퍼 하나를 한 다리 건너 두면 사라진다. 게다가 그 금지의 양성 대조가 반례였다 —
소비자 넷은 전원 `dispatchHandoff()` 를 통해 이미 한 다리 건너 `entries` 를 읽고 있었다.

### 고친 방법: 검사를 지우고 타입을 바꿨다

| 무엇 | 어떻게 |
|---|---|
| 봉투 타입 | `strategyhandoff.Delivered` — 값 필드가 비공개. 밖에서는 영값만 만들 수 있다 |
| dispatch 서명 | `dispatch(ctx, strategyhandoff.Delivered)` — 경계 밖 `Result` 는 **컴파일 실패** |
| 지운 검사 | `TestSeamConsumersCannotReadTheRawEntryListAgain`, `seamConsumerFuncs`, `identMentions` 의 `entries` 금지, 호출 인자의 이름 소속 검사(`fromSeam`), `deliveredParamName` |
| 더한 검사 | `TestOnlyTheSeamsEnvelopeCanReachTheSharedDispatch`(서명 고정), `TestExactlyOneProductionSiteAdmitsIntoTheSeam`(패키지 전체의 `Admit` 세기), `strategyhandoff` 의 `TestOnlyTheEngineImportsThisSeam`(이 경계를 들여오는 패키지 고정, 106 패키지 훑어 importer 1 — *4차 정정: 그 106 은 `go list <module>/...` 의 우주이고 `testdata` 를 건너뛴다. 5.5-fix4 에서 `-deps` 폐포 셋 2595 항목으로 바꿨다*) |
| 적재 미달 이름 | `refusalNow` 가 `<`/`>` 로 갈라 미달은 `HANDOFF_NO_SELECTION`. 다섯째 거절 어휘를 만들지 않는다 |
| census 연산자 | `singleProposalAssumptionSites` 가 `value.Op` 를 읽는다. 앞 판본은 `len(x.entries) - 1` 도 "1과의 비교"로 셌다 |

봉투가 증명하지 **못하는** 것을 코드와 문서에 함께 적었다: 봉투는 "dispatch 된 값이
`Admit` 을 거쳤다"만 증명하고, 그 `Admit` 을 부른 것이 조정자였는지는 증명하지 않는다.
~~그 나머지는 위의 두 세기가 맡고, 각각 **자기 범위 안에서 완전**하다(하나는 엔진 패키지
전체의 호출, 하나는 모듈 전체의 import).~~

**4차 정정 — 이 문장은 거짓이었고, 셋 다 그것을 뚫었다.** (a) 엔진 세기는 함수 본문
밖을 못 봤다. (b) import 세기의 우주 `go list <module>/...` 는 `testdata` 디렉터리를
건너뛰는데 컴파일러는 안 건너뛴다. (c) **그리고 둘 다 고쳐도 닫히지 않는다** —
`dispatchHandoff` 는 평범한 패키지 사설 구조체의 메서드라 엔진 안 아무 함수나 수신자를
리터럴로 만들어 부를 수 있고, 그때 `Admit` 도 import 도 나타나지 않는다. **주조 권한이
이 패키지 안에 진짜로 있으므로 철자를 세는 검사로는 영원히 못 닫는다.**

### 뮤테이션

| ID | 편집 | 결과 |
|---|---|---|
| M1 | 경계를 지나지 않은 `Result` 를 dispatch 에 넘긴다(리뷰어 셋의 공격 그대로) | KILLED — **컴파일 실패** — *4차 정정: 참이지만 **오늘의 철자에 대해서만** 참이었다. `Delivered` 의 필드를 `Value` 로 한 단어 바꾸면 셋이 전부 되살아나는데 그 비공개성을 고정하는 시험이 없었다(A P1-1). 5.5-fix4 에서 표면 표가 타입의 모양을 적게 했다.* |
| M2 | 엔진이 스스로 `Admit` 을 불러 봉투를 만들어 넘긴다 | ~~KILLED~~ — **거짓이었다.** 세기가 `*ast.FuncDecl` 안만 훑어서 패키지 수준 `var f = …Admit` 이 빠져나갔다(A P1-2a, B P1-1). 5.5-fix4 에서 파일 전체로 올렸다 |
| M3 | 같은 봉투로 dispatch 를 3회 부른다 | **SURVIVED** |
| M4 | 분기 표의 좌표를 같은 함수의 반환 좌표로 바꾼다(실제 a112 번들) | KILLED — `role_check` |
| M5 | `len(entries) - 1` 산술식을 넣는다 | 세지 않음(의도대로) |

원복은 `try/finally` 안에서 하고 sha256 으로 대조했다.

**M3 은 남긴 것이 아니라 주장을 바꾼 것이다.** *(4차 정정: 아래 문장이 "쟀다"고 적은
귀속은 재지 않은 것이었다 — 그 시험은 `err != nil` 과 주문 1건만 단언하고 막은 주체는
`t.Logf` 에만 있었다. B 가 CAS 관문을 지우고 돌려 보니 시험은 그대로 통과했고 원장의
replay-identity 가 대신 막았다. 5.5-fix4 에서 이름으로 단언하게 고쳤고, at-most-once 가
**과잉 결정**되어 있다는 것이 그 과정에서 드러났다.)* 소스 검사는 철자를 셀 뿐 실행을 세지
않으므로 "한 주기에 한 번"은 이 가드의 성질이 아니다. 실제로 막는 것은 원장이고, 그것을
재서 확인했다 — `TestTheSameEnvelopeCannotPlaceASecondOrder` 에서 두 번째 dispatch 는
`AUTHORITY_COLLECTION_FAILED: production position campaign CAS changed` 로 막히고 Gateway
주문은 1건에서 멈춘다. 위조 봉투는 Gateway **관측 0건**으로 첫 줄에서 걸린다
(`TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall`) — 가정하지 않고 쟀다.

### 게이트가 자기 증거를 못 잡던 것 — 기전을 고쳤다

C 가 `check_analysis.py` 에 오류 넷을 심어 전부 통과시켰다. 그 좌표 검사는 소스 **범위**와
분기 **개수**만 보고, 좌표가 무엇을 가리키는지는 보지 않았다. `tools/logic-map/role_check.py`
가 그 역할을 대조하고 `check_analysis.py` 에 배선됐다.

배선 **전에** 저장소의 번들 401개에 돌려 쟀다(*4차 정정: 그것은 **비아카이브 부분집합**
이었다. 저장소는 3014개이고, 아카이브 2610개에서 오탐 셋이 나왔다 — 대비율 `4.5:1`,
장 시작 `09:00`, 주소 `127.0.0.1:0`. 개수만 늘리고 **종류**를 안 바꾼 측정이었다*):
**맞는 문서에서 터진 것 0**, 진짜 결함
3건 — 열거형 호출 표 셋이 정확히 40행에서 멈춰 있었다(실제 64·46·91). 셋 다 복구했다.
오탐을 없애는 과정에서 규칙을 두 번 좁혔다: 표의 "AST kind" 칸은 산문이 들어오고
(`case KindKRNetFlow → KR only`), 손으로 쓴 호출 분석 표(*4차 정정: 204 가 아니라 **226개**;
204 는 그중 최빈 헤더 철자의 수였다*)는 한 줄이 호출 넷을 묶는다.
그래서 **좌표만** 대조하고, 완전성을 주장하는 표(`| Callee expression | …`)만 1:1로 본다.
`test_role_check.py` 는 역할 차원마다 오류를 하나씩 심고, 침묵해야 할 경우도 같은 수만큼
못 박는다(13 시험).

### 실측

| | |
|---|---|
| `make lint` | PASS (gofmt 경로 확인함 — 빈 출력이 "검사 0"이 아님) |
| `make test` | PASS, 98 패키지 ok, FAIL 0 |
| `make test-seams` | PASS, 99 패키지 ok, FAIL 0 |
| engine 태그 커버리지 | **73.9%** — 2914·2915·2914 of 3942 (3회) |
| engine 무태그 커버리지 | **63.4~63.5%** — 2498·2499·2499 of 3937 (3회) |
| `check_analysis` (a112) | evidence complete or diff-proven exempt |
| `role_check` (저장소 404 번들) | findings 0 |
| `tools/logic-map` 단위 시험 | 46 PASS |
| `make sdd-test` | scripts 15 · logic-map 46 · sdd 29 · sdd-history 22 PASS; pm 만 1 실패 — 그 실패도 아래 a117 스텁이 원인이다(`duplicate numbered change a117`) |
| `make sdd-sync` | PASS (GBrain advisory busy — 이전 신선도 유지) |
| `make sdd-check` | **FAIL — 원인 둘, 이 행은 순서가 틀렸다.** *4차 정정: 먼저 걸리는 것은 이 로트 자신의 stale CodeGraph 지문이다(`sdd-sync` 를 부모 커밋에서 돌리고 그 뒤 커밋했다). `generate_master_tracker --check` 는 그 앞에서 멈추므로 a117 스텁까지 가지도 않는다. 두 번째 원인이 병행 Codex 세션의 untracked 중복 `a117-codex-session-handoff-and-gbrain-startup` 스텁이고, 그 세션이 번호를 다시 매겨야 한다(빈 번호 104·105·106·119·120)* |
| `make gate CHANGE=a112…` | **not-applicable** — 이것은 change 전체의 완료 게이트이고 a112 는 L1~L8 에 미완료 44건이 남아 진행 중이다. 2/10 단계에서 멈춘다. 로트 단위 검증은 위 항목들이다 |
| `openspec validate a112… --strict` | **valid** |
| `make validate` (`--all --strict`) | 59 passed / 1 failed — 실패는 위와 같은 a117 스텁 하나뿐 |

### 남긴 것

- **`Refusal()`·`Pending()` 의 생산 독자는 여전히 0.** 네 `HANDOFF_*` 이름이 운영자
  화면에서는 `RefusalEvidenceStale` 하나로 접히므로, 지금 상태는 소스에서만 이름이 있고
  운영에서는 **틀린 이름으로** 보고된다. 배선하지 않기로 했다 — 배선하면 동결 골든의
  `refusal_enums` 에 없는 거절 어휘를 영수증 없이 늘리게 되고, 그것은 5.2/7.3 의 몫이다.
  범위를 "소스에서 명명했다"로 좁혀 여기 남긴다.
- **B-5(같은 패키지 능력 세탁)**: 2라운드에서 철회한 그대로. 어댑터 `dispatchHandoff` 가
  아직 엔진 안에 있어서 import 폐쇄가 결정적이지 않다.
- `ErrNoDelivery` 는 시장을 잠글 수 있으나 현재 도달 불가이고 문서화되지 않았다(7.3).
- 이 로트도 재검토를 받아야 한다 — 세 라운드 연속 같은 뿌리에서 P1 이 나왔다.

## 5.5-fix4 — 4라운드 적대 재검토: 셋 다 REJECT, 뿌리는 여전히 하나 (2026-08-31)

로트 `b73b35e1` 을 서로 못 보는 리뷰어 셋에게 read-only 로 맡겼다. **P0 0건, 셋 다
REJECT.** 세 축(봉투 / 돈 경로 / 증거 게이트)에서 독립으로 왔는데 결론이 겹쳤다:
**이 로트가 고치겠다고 만든 검사 둘이, 각자 자기가 고치려던 결함을 그대로 재현했다.**

### 한 단어로 죽는 로트였다

A 가 가장 싼 공격을 찾았다. `Delivered` 의 필드를 `result` → `Value` 로 **이름만**
바꾸면 3라운드의 우회 셋이 전부 되살아나고 두 스위트가 그대로 초록이다. 로트 전체가
그 필드 하나의 비공개성 위에 서 있는데 **그것을 고정하는 시험이 없었다.**
`escape_test.go` 는 타입을 `"type"` 이라는 한 단어로만 적었고, 필드를 보는 유일한
검사는 `spec.Name.Name != "Handoff"` 로 걸러 `Delivered` 를 보지 않았다. 검사 셋을
지우고 타입 하나로 바꿨는데, 그 타입을 지키는 것이 아무것도 없었다.

### 그리고 철자로는 영원히 안 닫히는 것

A P1-2(b) 가 진짜 뿌리다. `strategyProposalMarketAuthority` 는 필드 셋짜리 패키지
사설 구조체이고 `dispatchHandoff` 는 그 위의 메서드다. 그래서 엔진 안 아무 함수나
수신자를 리터럴로 만들어 자기가 고른 entries 로 봉투를 찍을 수 있고, 그때 `Admit`
토큰도 import 도 나타나지 않는다. A 가 `OVER_CAPACITY` 로 거절된 시장을 우회해 봉투를
받아내는 코드를 컴파일해 보였다. **주조 권한이 이 패키지 안에 진짜로 있으므로 철자를
세는 검사로는 닫히지 않는다.** 4라운드 동안 그 세기의 범위만 넓혀 왔다.

### 오늘 안전한 이유는 이 경계가 아니다

A 가 만든 네 우회 중 **주문을 낸 것은 하나도 없다.** 전부
`strategy_account_first_leg_authority.go` 에서 막힌다 — 개수 관문 `:217`
(`len(proposal.entries) != 1`), 재유도 `:221`–`:222`
(`proposalAuthority := proposal.entries[0].authority` → `proposalAuthority.Proposal()`),
identity 대조 `:223`–`:225`
(`result.Lineage.Identity != accepted.result.Lineage.Identity ||
result.ExecutionTerms.Identity() != accepted.result.ExecutionTerms.Identity()` →
`production proposal identity changed`). 앞 판본은 이 자리를 `:217/:222` 로 적었는데,
그 두 줄은 관문과 재유도이고 **실제로 위조를 거절하는 비교는 `:223` 이다.**
그 다섯 줄은 `singleProposalAssumptionCensus` 에 **L6 6.2 가 지울 빚**으로 등록돼 있다.

**그런데 그 backstop 에는 시험이 없었다.** `grep -rn "production proposal identity
changed" --include=*.go .` 는 생산 코드 한 줄만 돌려준다 — 소비자가 없다. census 는
`entries` 의 **모양**을 세므로 `:223`–`:225` 만 지우면 5 가 그대로 유지되고 초록이다.
즉 5.5 가 "오늘의 안전은 여기서 온다"고 적은 줄은, 지워도 게이트가 조용한 줄이었다.
Q1 권장안을 진행하면서 이 로트가 그것을 닫았다:
`TestFirstLegAuthorityRefusesAProposalItDidNotAuthorize`
(`strategy_first_leg_identity_backstop_test.go`, `tossos_testseams`) 는 KR 제안으로
만든 `accepted` 를, KR 자리에 US 제안이 들어 있는 권한 쌍에 건네고 위 문구를 요구한다.
**반증:** `:223`–`:225` 를 지우면 오류가
`production risk authority scope changed` 로 바뀌어 빨개진다(sha256 원복 확인). 이것이
identity 대조가 범위 대조보다 먼저 온다는 근거이기도 하다 — 순서를 말로 적지 않고
뮤테이션으로 재서 적는다.

즉 오늘의 backstop 은 이 로트가 만든 경계가 아니고, 6.2 가 그 다섯 줄을 지울 때
대신 설 기계는 위 세 구멍이 뚫려 있었다. **6.2 는 진짜 출처 증명 —
`Admit` 이 조정자만 만들 수 있는 봉인 값을 요구하는 것 — 이 서기 전에는 그 다섯 줄을
지울 수 없다.** 이 차단 선결조건은 여기뿐 아니라 **tasks.md 의 6.2 본문에** 적었다.
6.2 를 여는 사람이 읽는 곳은 5.5 의 문단이 아니라 자기 태스크이기 때문이다.

### 고친 것과 그 반증

| # | 리뷰어 | 결함 | 고침 | 반증(뮤테이션) |
|---|---|---|---|---|
| 1 | A P1-1 | 봉투 필드의 비공개성을 아무것도 고정 안 함 | 표면 표가 `types.ExprString` 으로 **모양**을 적는다; 필드 검사를 공개 타입 전부로 | `result`→`Value` → 두 검사가 **각각** 빨강 |
| 2 | A P1-2a · B P1-1 | `Admit` 세기가 `*ast.FuncDecl` 안만 훑음 | `admitSites` 가 **선언 전부**를 훑고 자리 이름을 붙인다 | `var mintHandoff = …Admit` → 빨강, 자리를 이름으로 부름 |
| 3 | A P1-3 | import 우주가 `testdata` 를 건너뜀 | `-deps` 폐포 셋(출하·시험·태그) 2595 항목 | `engine/testdata/seamrelay` → 빨강 |
| 4 | B P1-2 | 막은 주체가 `t.Logf` 에만 있음 | 이름으로 단언 | CAS 관문 삭제 → 빨강(원장 replay 가 대신 막는 것이 드러남) |
| 5 | C P1-1 | `branch-test-map` 좌표 38개가 옛 줄번호 | 좌표 재기준화 + 검사기가 그 파일도 본다 | 실제 번들에 stale 앵커 주입 → 게이트 exit 1 |
| 6 | C P1-2 | 행 아무 데서나 숫자쌍을 좌표로 읽음 | **칸 하나가 통째로 좌표일 때만** 읽는다 | 오탐 셋을 시험에 심음; 맞는 행에 시각 삽입 → 조용, 좌표 틀리면 → 빨강 |
| 7 | C P1-4 | `sdd-check` 실패 원인 오귀속 | 지문 재sync + 순서대로 다시 적음 | — |

**세기의 범위는 이제 말로 적지 않고 세어서 적는다.** `TestAdmitCensusSeesEverySpellingItClaims`
가 네 철자(평범한 호출·패키지 수준 함수 값·dot import·합성 리터럴)를 fixture 로 담아
확인한다. 앞 판본은 같은 문장을 주석으로만 적었고 **그중 하나가 거짓이었으며, 그
거짓이 구멍을 가렸다.**

### 좌표 규칙은 모양을 세어서 정했다

`4.5:1`(대비율) · `09:00`(장 시작) · `127.0.0.1:0`(주소)가 좌표로 읽히고 있었다. 3014
번들에서 좌표를 담은 칸의 모양을 전수로 셌다: FLM 은 `N:N` 780칸, branch-test-map 은
`if at N:N` 356 · `range at N:N` 51 · `N:N` 22 · `case at N:N` 22 · `switch at N:N` 6 ·
`else at N:N` 2 · `branchless happy path at N:N` 2 와 그 뒤의 ` — 산문`. 그 모양만 읽는다.
**오탐 셋은 지어낸 값이 아니라 저장소에서 나온 값이고, 그대로 시험에 심었다.**

### 남긴 것 — 침묵한 생략이 되지 않도록

- **`dispatchHandoff` 수신자 위조는 안 닫혔다.** 철자 세기로는 못 닫는다(위 참조).
  진짜 해법은 `Admit` 이 조정자만 만들 수 있는 봉인 값을 받는 것이고, 그것은
  coordinator 패키지를 건드려 5.5 범위를 넘는다. **사람이 권장안을 골랐다(5.5-fix5)** —
  주장을 물리고, 오늘의 backstop 을 이름으로 적고, 진짜 출처 증명을 6.2 의 차단
  선결조건으로 건다. 아래 §5.5-fix5 참조.
- **완전성 표지를 단 호출 표 122개 중 39개는 여전히 기계 검사를 안 받는다.** 행에
  좌표 대신 줄 번호만 적어서 규칙이 조용히 면제한다. `| Callee expression |` 이라는
  **표지를 문서 저자가 고르므로, 이것은 이 change 가 지운 "이름이 있는지 보는 검사"와
  같은 병이다.** ~~39개를 이름 붙은 빚으로 여기 남긴다.~~ **5.5-fix5 에서 닫았다** —
  빚으로 남기지 않고 39 개 + 표가 아예 없던 11 개를 모두 열거로 바꾸고, 면제 자체를
  오류로 만들었다. 아래 §5.5-fix5 참조.
- `Refusal()`/`Pending()` 은 여전히 비테스트 호출자가 0이다(B P2-1 이 독립 확인).
  동결 골든에는 `HANDOFF_` 이름이 **하나도** 없다 — 네 이름 모두 골든 미수록이다.
- `ast.json` 은 검사되지 않는 신탁이다(C P2-1). 문서와 ast.json 을 **같이** 자르면
  통과한다. 소스에서 다시 유도하는 것이 근본 해법이고 여러 구멍을 한꺼번에 죽인다.
- `strategy_entry_supervisor.go:450` 의 CAS 술어는 `:232` 의 사본이고, 둘의 순서
  (바깥이 느슨 → 안이 엄격)를 고정하는 것이 없다(B P2-2, 이 로트가 만든 것 아님).
- 리스 소진을 삼키면 campaign 이 claimed 로 남아 그 종목의 CAS 가 이후 계속
  실패할 수 있다(B P2-3, 선행 결함 `8022f578`). 주문은 안 나가지만 운영자 신호가 없다.
- 엔진 패키지의 `-race` 1회는 1160초다. `-count=5` 는 약 97분이라 기본 10분 예산으로는
  끝나지 않는다. **경합은 0이다** — 다섯 번 다 확인했다. 돌릴 수 있는 형태는
  `go test -race -count=1 -timeout 40m ./internal/app/engine/`.

### 실측 (5.5-fix4)

| | |
|---|---|
| `make lint` | PASS |
| `make test` | 98 ok / FAIL 0 |
| `make test-seams` | 99 ok / FAIL 0 |
| `tools/logic-map` 단위 시험 | **54 PASS** (46 → 8 추가) |
| `role_check` 저장소 전수 | **3014 번들, findings 0** (비아카이브 404 + 아카이브 2610) |
| `check_analysis` (a112) | evidence complete or diff-proven exempt |
| 게이트 종단 주입 | stale 앵커 주입 → exit 1, 원복 sha256 일치 |
| engine 태그 커버리지 | **74.0%** — 2917 of 3942 (3회 모두 동일) |
| engine 무태그 커버리지 | **63.5%** — 2501 of 3937 (3회 모두 동일) |
| `make sdd-sync` → `sdd-check` | 지문 재sync 후 "CodeGraph hard-evidence index matches the worktree" 통과; **남은 실패는 병행 세션의 a117 중복 스텁 하나뿐** |
| `openspec validate a112 --strict` | valid |
| `make gate` | **not-applicable** — change 전체의 완료 게이트이고 a112 는 미완료 태스크가 남아 2/10 단계에서 멈춘다. C 가 독립 확인함 |

**커버리지가 이 로트에서 미세하게 올랐다**(2914→2917, 2498→2501). 새 시험이 생산
코드를 더 덮은 것이 아니라 census/봉투 검사가 더 많은 선언을 훑기 때문이고, 세 번
모두 같은 값이 나왔다 — 앞 로트에서 표본마다 흔들리던 것과 다르다.

## 5.5-fix5 — 사람이 고른 두 결정을 실행한 로트

4차 재검토가 남긴 두 구멍은 제가 판단할 수 없어 사람에게 물었고, 답이 왔다.
Q1 은 권장안(주장을 물리고 6.2 의 차단 선결조건으로 건다), Q2 는 "39개가 기계
검사를 받도록 고친다"였다. 이 절은 그 둘을 실행한 결과다.

### Q1 — 출처 증명 주장을 물리고, 오늘의 backstop 을 시험으로 못 박았다

권장안을 적으면서 **먼저 확인해야 할 것이 하나 남아 있었다.** "오늘 우회가 주문으로
이어지지 않는 이유는 `strategy_account_first_leg_authority.go` 의 재유도·identity
대조"라고 4차에 적었는데, 그 줄에 **시험이 있는지는 안 봤다.**

```
$ grep -rn "production proposal identity changed" --include=*.go .
./internal/app/engine/strategy_account_first_leg_authority.go:224
```

소비자가 없다. 즉 **오늘의 안전이 여기서 온다고 적어 둔 줄은, 지워도 두 스위트가
초록인 줄이었다.** census(`singleProposalAssumptionCensus`)가 지킨다고 생각하기 쉽지만
census 는 `entries` 의 **모양**을 세므로 identity 비교만 지우면 5 가 그대로 유지된다.
네 라운드 동안 반복된 것과 같은 자리 바꿈이다 — 세는 것과 막는 것은 다르다.

좌표도 틀려 있었다. 4차는 이 자리를 `:217/:222` 로 적었는데 그 둘은 개수 관문과
재유도이고, **실제로 위조를 거절하는 비교는 `:223`–`:225`** 다. 사본 두 곳을 모두
고쳤다 — 정정의 단위는 file:line 이 아니라 값이다.

그래서 권장안은 문서 세 곳을 고치는 일이 아니라 **시험 하나를 새로 만드는 일**이 됐다.

| | |
|---|---|
| 새 시험 | `TestFirstLegAuthorityRefusesAProposalItDidNotAuthorize` (`internal/app/engine/strategy_first_leg_identity_backstop_test.go`, `tossos_testseams`) |
| 무엇을 하나 | KR 제안으로 만든 `accepted` 를, KR 자리에 US 제안이 들어 있는 권한 쌍에 건넨다 — 주조된 봉투가 나르는 결과가 권한 쌍에 없는 상황과 같은 모양 |
| 무엇을 요구하나 | 오류에 `production proposal identity changed` 가 들어 있을 것 |
| **반증** | `:223`–`:225` 삭제 → 오류가 `production risk authority scope changed` 로 바뀌어 **빨강**. sha256 원복 확인(`1d710710…`) |

그 반증은 덤으로 **identity 대조가 위험 범위 대조보다 먼저 온다**는 것을 잰 값이다.
시험이 두 축(시장·identity)을 함께 바꾸므로 순서를 말로 적으면 미검증 주장이 된다 —
반증은 결함이 사는 축을 바꿔야 한다.

**새 파일에 넣었다.** 처음에는 기존 시험에서 fixture 를 추출했는데 `check_analysis` 가
바로 잡았다: 기존 함수 `TestProductionFirstLegAuthorityLoaderPairedKRUS` 의 본문을
바꿨으므로 FLM 번들을 요구한다. 행동 변화가 0 인 추출 리팩터링에 번들을 다는 대신
기존 함수를 **원문 그대로 되돌리고** 새 파일에 fixture 를 따로 뒀다
— 새 코드는 새 파일에 둔다. 되돌림은 `git show HEAD:<path>` 로 그 파일만 했다;
`git checkout` 은 커밋 안 된 다른 작업까지 지우므로 쓰지 않는다.

차단 선결조건은 review 가 아니라 **tasks.md 의 6.2 본문**에 적었다. 6.2 를 여는 사람이
읽는 곳은 5.5 의 문단이 아니라 자기 태스크이기 때문이다. 거기에 두 가드가 **서로 다른
반쪽**을 잡는다는 것과(census 는 모양, 새 시험은 행동), 그 다섯 줄을 지울 수 있는
조건(`Admit` 이 조정자만 만들 수 있는 봉인 값을 받을 때)을 함께 적었다.

**안 닫힌 것은 그대로다.** `dispatchHandoff` 수신자 위조는 여전히 컴파일된다. 권장안은
그것을 닫는 안이 아니라 **주장을 사실에 맞추고 진짜 해법을 6.2 의 선결조건으로 거는**
안이다. 5.5 는 이 구멍을 닫지 않았다고 읽는 것이 맞다.

### Q2 — 면제를 빚으로 남기지 않고 없앴다

4차가 센 값: 완전성 표지를 단 호출 표 122 개 중 83 개만 1:1 대조되고, **39 개는 행에
좌표 대신 줄 번호만 적어 검사기가 표 전체를 건너뛰었다.** 표지를 저자가 고르므로
"감사를 받을지"를 문서 저자가 정하고 있었다.

**문서 쪽 — 50 개 번들을 열거로 바꿨다.**

| | |
|---|---|
| 빠져나가던 표 | 39 개 (전부 a112) |
| 열거 표가 아예 없던 번들 | 11 개 (같은 change) |
| 다시 쓴 번들 | **50 개** (파일 변경 48 개, 2 개는 이미 목표 모양) |
| 손으로 쓴 행 → 기계 행 | 215 → 952 |
| change 전체 | **133/133 번들**, 호출 행 **2300**; 좌표를 적은 표 **127** 개, 호출이 0 이라 "없다"를 적은 표 **6** 개 (5차 리뷰가 128/5 를 정정) |

손으로 쓴 분석은 **지우지 않았다.** 열거 표 아래에 "완전성 주장이 아니다"라는 제목을
달아 남기고, 머리글을 `Callee expression` → `Callee (hand-written note)` 로 낮춰
표지를 못 쓰게 했다. 그 산문의 줄 번호가 위 표와 어긋나면 위 표가 정본이라고 각
번들에 적었다(실제로 `internal-official--adaptprices` 에서 `173:10` vs `173:15` 로
어긋난다 — 손으로 읽은 값이 하나 틀렸던 것이고, 그것이 이 규칙이 있어야 하는 이유다).

**검사기 쪽 — 세 가지가 달라졌고, 각각 반증했다.**

| # | 바뀐 것 | 뮤테이션 | 결과 |
|---|---|---|---|
| 1 | 표지를 달고 좌표를 뺀 행은 **오류**다(면제 아님) | 행 하나의 `221:9` → `221` | exit 1, "1 of 2 row(s) carry no coordinate cell" |
| 2 | 섹션의 **첫 연속 표 한 벌**만 읽는다 | 열거 두 행 삭제(아래 주석 표 3 행은 그대로) | exit 1, "enumerates 4 call(s) but ast.json has 6" — 주석 행이 개수를 채우지 못한다 |
| 3 | 강제는 **change 단위**다(`call_enumeration_in_use`) | 한 번들의 표지를 `\| Callee \| Why called \|` 로 교체 | exit 1, "this change enumerates call positions elsewhere, so … must open with a … table" |
| 4 | 호출이 0 인 함수만 좌표 없는 표를 적을 수 있다 | 호출 2 개인 번들을 "(no call expressions…)" 로 위장 | exit 1, "cites no call position, while ast.json has 2" |
| 5 | 잘린 표는 그대로 잡힌다(회귀) | 열거 한 행 삭제 | exit 1, "enumerates 1 call(s) but ast.json has 2" |

다섯 모두 sha256 원복을 확인했다. 단위 시험도 같은 방식으로 반증했다 — role_check 에
네 가지 되돌림 뮤테이션을 심으니 **각각 정확히 시험 하나씩만** 빨개졌다(겹쳐서
가려 주는 시험이 없다는 뜻이다).

**남는 선택은 change 단위다 — 그리고 그것을 검사기에 적어 뒀다.** 열거 표를 한 번도
쓰지 않는 change 는 여전히 아무것도 요구받지 않는다. 그것이 다른 16 개 활성 change 의
손으로 쓴 번들 271 개를 정상으로 유지하는 이유다 —
fail-closed 규칙이 거절할 **정상 입력**을 먼저 세고 나서 켰다. 대신 어떤 change 의 첫 번들이 표지를
다는 순간 나머지 전부가 함께 요구된다. 그 절벽은 의도된 값이고, 놀라지 않도록
`role_check.py` 주석에 적었다.

### 실측 (5.5-fix5)

| | |
|---|---|
| `make lint` | PASS |
| `make test` | 98 ok / FAIL 0 |
| `make test-seams` | 99 ok / FAIL 0 |
| `tools/logic-map` 단위 시험 | **62 PASS** (54 → 8 추가) |
| `role_check` 저장소 전수 | **3014 번들 / 93 change, 강제가 켜진 change 1 개, findings 0** |
| `check_analysis` (a112) | evidence complete or diff-proven exempt |
| 게이트 종단 주입 | 5 종 뮤테이션 전부 exit 1, 원복 sha256 일치 |
| engine 태그 커버리지 | 2915·2916·2916 of 3942 (**73.9~74.0%**) |
| engine 무태그 커버리지 | 2498·2499·2501 of 3937 (**63.4~63.5%**) |
| `openspec validate a112 --strict` | valid |
| `make gate` | **not-applicable** — change 완료 게이트이고 a112 는 미완료 태스크가 남아 2/10 에서 멈춘다 |

커버리지는 4차의 2917(3회 동일)보다 **한두 문 낮고 회차마다 흔들린다.** 새 시험이
생산 코드를 더 덮었는데 값이 내려간 것이므로, 이 패키지의 동시성 시험이 만드는 잡음
범위 안이라고 읽는 것이 맞다 — 올랐다고 적지 않는다.

## 5.5-fix6 — 5차 적대 리뷰가 낸 P1 셋을 고친 로트

5차 리뷰는 P0 0 · P1 3 · P2 5 로 REQUEST CHANGES 였다. 셋 다 실재했고 재현했다.
**같은 병이 이번 라운드의 고침 안에 다시 있었다** — 그 사실이 이 절의 요점이다.

### P1-1 — 새 시험이 결함이 사는 축을 안 바꿨다

5.5-fix5 가 만든 `TestFirstLegAuthorityRefusesAProposalItDidNotAuthorize` 는
KR 자리에 **US** entries 를 넣어 위조한다. 그러면 identity 만이 아니라 시장·종목·
계좌까지 함께 달라진다. 그래서 `:223` 의 술어가 identity 가 아니라 **market 만**
비교하도록 약해져도 이 시험은 초록이다.

재현했다. `go test -overlay` 로 술어만
`string(result.Lineage.Market) != string(accepted.result.Lineage.Market)` 로 바꾸고
(저장소 파일 sha256 `1d710710…` 앞뒤 동일):

| | |
|---|---|
| `…AProposalItDidNotAuthorize` | **PASS** — 뮤턴트를 못 잡는다 |
| `…ASiblingCampaignOnTheSameSymbol` (새로 추가) | **FAIL**, 그리고 실패값이 `err=<nil>` |

`err=<nil>` 이 핵심이다. 뮤턴트에서는 **권한 쌍에 없는 캠페인으로 1차 진입이 그대로
나간다.** 4차가 "우회 넷 중 주문을 낸 것은 없다"고 적은 근거가, 술어 한 줄만 약해지면
사라진다는 뜻이다.

고침은 `TestFirstLegAuthorityRefusesASiblingCampaignOnTheSameSymbol` 이다. 계좌·시장·
종목·수량·가격을 전부 고정하고 `campaignID` 하나만 바꾼, 똑같이 봉인된 KR 제안을
넣는다. 시험이 **스스로 축이 하나뿐인지 확인**한다 — 형제 제안을 만드는 쪽이 나중에
다른 축까지 바꾸면 그 자리에서 터진다.

a112 의 목표 상태는 한 종목에 네 가족이므로 "같은 시장·같은 종목·다른 identity" 가
오히려 지배적인 모양이다. 5.5-fix5 는 그 모양을 한 번도 안 밟았다.

기억에 적어 둔 규칙(`반증은 결함이 사는 축을 바꿔야 한다`)을 **문장으로 인용하면서
적용하지 않았다.** 규칙을 아는 것과 쓰는 것은 다르다.

### P1-2 — 완전성 표지를 단 열을 검사기가 한 번도 안 봤다

`ENUMERATED_CALLS` 는 `| Callee expression |` 로 표를 고르고, 그 다음 검사기는 각
행에서 **좌표만** 읽었다. 완전성을 주장하는 열이 정작 대조되지 않았다.

리뷰어의 시연을 그대로 재현했다 — `adaptPrices` 의 두 행을 좌표는 그대로 두고 철자만
`os.Exit`·`exec.Command` 로 바꾸면 `role_errors` 가 `[]` 를 돌려줬다. 즉 완전성을
주장하는 표 아래에서 `adaptPrices` 가 `os.Exit` 을 부른다고 적을 수 있었다.

이것이 다섯 라운드 동안 고쳐 온 그 병이다. 좌표만 보는 검사는 "그 자리에 호출이
있다"까지만 말하고, 그 행이 그 호출을 **가리킨다**는 것은 말하지 않는다.

이제 좌표가 맞은 뒤 철자를 대조한다. 철자가 없는 노드는 기존 문서 관례인 `(unnamed)` 로 읽는다 — 표지를 단 133 개
안에서는 15 개지만 **저장소 전체로는 189 번들에 281 개**다(6차 리뷰가 정정). 두
번째 change 가 표지를 다는 순간 이 탈출구는 15 가 아니라 281 이 된다. **133 개 번들 전부가 그대로 통과했다** — 생성한 표의 철자가
실제로 맞았다는 뜻이고, 그것은 이 검사를 켜기 전에는 근거 없이 믿던 것이다.

### P1-3 — 이 로트가 추가한 시험 8 개가 직접 실행에서 안 돌았다

`if __name__ == "__main__": unittest.main()` 가 **파일 중간**에 있었고, 5.5-fix5 가
추가한 두 클래스는 그 아래에 붙었다. `python3 test_role_check.py` 는 21 개만 돌고
OK 를 냈다. 그 8 개는 네 가지 검사기 변경 **전부**의 유일한 가드였다.

| | |
|---|---|
| 직접 실행 (고치기 전) | Ran **21** … OK |
| discover (고치기 전) | Ran **29** … OK |
| 직접 실행 (고친 뒤) | Ran **34** … OK — fix5 의 29 개에 fix6 이 5 개를 더했다 |

회귀 반증: 스크래치패드 사본에서 5.5-fix5 가 없앤 침묵 면제(`if blank: return []`)를
되살렸더니 — 고치기 전 직접 실행은 **OK**, 고친 뒤 직접 실행은 **FAILED (failures=1)**.
39 개 면제가 돌아오는 것을 자연스러운 로컬 명령이 조용히 통과시키고 있었다.

`.github/workflows/ci.yml` 은 `make test`·`make lint`·`make build` 만 돈다. 이 8 개는
`make sdd-check`/`make gate` 에만 있고 CI 는 안 쓴다 — 리뷰어가 짚은 대로다. 이번
로트에서 CI 워크플로는 건드리지 않았다(범위 밖). **미해결로 남긴다.**

### P2 다섯

| # | 무엇 | 고침 |
|---|---|---|
| P2-1 | 새 시험의 **파일명**이 세 곳에서 틀렸다(`…authority_test.go`, 실재하지만 그 시험이 없는 파일) — 그중 하나가 6.2 의 선결조건 문장 | 세 사본 전부 `strategy_first_leg_identity_backstop_test.go` 로. 남은 사본 0 |
| P2-2 | B3 의 branch-test-map 이 아직 "planned … not run" 이었다 | 실측 RED/GREEN 과 두 시험 이름을 기록. 뮤테이션 두 종과 명령·sha256 도 함께 |
| P2-3 | 강제 판정이 표지를 **첫 표의 첫 줄**에서만 찾아, 열거 위에 `|` 한 줄만 얹으면 표지를 남긴 채 감사가 꺼졌다 | 섹션의 아무 줄에서나 표지를 찾는다. 133 번들 전부를 오염시켜도 이제 `True`. "열거가 첫 표여야 한다"는 별도 규칙으로 남아 그 번들이 따로 터진다 |
| P2-4 | `_first_table` 주석이 실제보다 세게 적혀 있었다 | 리뷰어도 해로운 입력을 못 만들었고, 주석 정확성 문제로 남긴다. **미해결** |
| P2-5 | 측정표의 127/6 을 128/5 로 적었다 | 정정. 호출 0 인 번들 6 개를 이름으로 확인 |

### "133/133" 의 한계 — 리뷰어가 잰 것을 그대로 적는다

리뷰어는 생성기를 믿지 않고 독립 `go/parser` 워커로 재유도했다: **118 개 대조, 불일치
0.** 나머지 **15 개는 `revision: base`** 라서 `check_analysis` 의 소스 해시 검사가
묶지 않으며, 그중 13 개는 이 change 의 `base-commit.txt`(`aeeb209e`)와도 안 맞는다.

그러니 정확히 말하면 **118 개는 소스로 재유도되어 맞았고, 15 개는 아무것도 검사하지
않는 신탁과 일치한다.** "133/133 이 열거한다"는 표 모양에 대한 참이지 좌표가 옳다는
뜻이 아니다. 구조적인 것이고 이 로트가 만든 것은 아니지만, 이 로트의 주장이 그것을
물려받으므로 여기 적는다. **미해결로 남긴다.**

### 실측 (5.5-fix6)

| | |
|---|---|
| `test_role_check.py` | **34 PASS** (직접 실행·discover 동일) — fix4 의 54 와 fix5 의 62 는 `tools/logic-map` **전체** discover 값이고 단위가 다르다. 같은 기준으로 재면 **67** 이다 |
| 새 규칙 반증 | 철자 대조 되돌림 → 해당 클래스 2 개만 FAIL · 강제 판정 되돌림 → 해당 시험 1 개만 FAIL |
| 새 Go 시험 반증 | market-only 술어 뮤턴트에서 `…ASiblingCampaignOnTheSameSymbol` 만 FAIL(`err=<nil>`), 기존 시험은 PASS |

이 표에 게이트 전 항목(`make lint`·`make test`·`make test-seams`·전수 스윕·`check_analysis`·`openspec validate`)이 빠져 있던 것을 6차가 지적했다. 돌리긴 했으나 §0 계약상 영수증은 이 파일에 있어야 하므로, 인도된 트리 기준 전 항목을 **5.5-fix7 의 실측표**에 적는다.

## 5.5-fix7 — 6차 적대 리뷰가 낸 P1 하나와 P2 여섯

6차는 P0 0 · P1 1 · P2 6 으로 다시 REQUEST CHANGES 였다. P1 은 **제 고침 한 단계
아래에 같은 병이 있었다**는 것이었고, 재현했다.

### P1 — `:223` 은 논리합 둘이고, 세 시험 중 둘이 같은 반쪽만 밟았다

```go
if result.Lineage.Identity != accepted.result.Lineage.Identity ||
   result.ExecutionTerms.Identity() != accepted.result.ExecutionTerms.Identity() {
```

`internal/strategyflow/types.go:315` 에서 `executionTermsIdentity` 는 마지막에
`terms.lineageIdentity` 를 해시에 넣는다. 그러므로 **lineage 가 다르면 terms 도
반드시 다르다** — 왼쪽 항은 오른쪽에 포섭되고, 수량과 entry·stop·target 세 가격을
지키는 것은 **오른쪽 항뿐**이다.

fix6 이 만든 형제 시험은 `campaignID` 를 바꾼다. 그것은 축 하나가 아니라 lineage
identity·terms identity·CampaignID **셋을 함께** 움직인다. 그래서 두 항이 같이 참이
되고, 시험은 어느 항이 일하는지 구별하지 못한다.

네 뮤턴트를 `go test -overlay` 로 넣어 쟀다(저장소 파일 sha256 `1d710710…` 앞뒤 동일):

| `:223` 의 술어 | …ItDidNotAuthorize | …ASiblingCampaign | …RewrittenExecutionTerms |
|---|---|---|---|
| 원본(논리합 둘) | PASS | PASS | PASS |
| `Lineage.Market` 만 | PASS | **FAIL** | **FAIL** |
| **`Lineage.Identity` 만**(오른쪽 삭제) | PASS | PASS | **FAIL** |
| **`Lineage.CampaignID` 만** | PASS | PASS | **FAIL** |
| `ExecutionTerms.Identity()` 만(왼쪽 삭제) | PASS | PASS | PASS |

마지막 줄이 **왼쪽 항이 포섭된다는 잰 값**이다 — 동등 변이다. 그리고 셋째·넷째 줄이
6차의 지적이다: `||` 반쪽을 지우는 흔한 실수 하나로 **수량과 세 가격의 유일한 검사가
사라지는데** fix6 의 두 시험은 초록이었다. 손절·사이징은 프로젝트가 High-risk 로
못 박은 경로다.

고침은 `TestFirstLegAuthorityRefusesRewrittenExecutionTermsUnderTheSameLineage` 다.
lineage 는 가격을 담지 않으므로(`authority_testseam.go` 의 lineage 에 entry·stop·target
이 없고 `PlannedCeiling` 은 주문 수량이 아니다), `campaignID` 를 고정한 채 손절 95→80 ·
익절 120→130 만 바꾸면 **왼쪽 항은 거짓이고 오른쪽 항만 참**이 된다. 시험은 그것도
스스로 확인한다 — lineage identity 가 같고 terms identity 가 다른지 먼저 본다.

**한계는 리뷰어가 적은 대로 적는다.** `AcceptedResultForAuthorityTest` 는
`tossos_testseams` 빌드 태그라 이 입력 자체는 생산 바이너리에서 만들 수 없고,
생산 경로에서 가격이 움직이면 lane 증거 digest 가 lineage 를 함께 움직일 수 있다.
생산에서 도달 가능한 입력은 **증명하지 못했다.** 이 항목이 서는 근거는 그것이
아니라, 지키는 항이 어떤 시험으로도 안 덮여 있었다는 것이다.

### P2 여섯

| # | 무엇 | 고침 |
|---|---|---|
| P2-1 | 강제 판정이 여전히 **철자 존재 검사**였다. 공백 둘·앞 들여쓰기·굵게 — 화면에서 똑같이 보이는 세 가지로 133 개가 통째로 빠져나갔다 | 판정을 **내용**으로 옮겼다: 어느 표든 그 함수의 호출 좌표를 빠짐없이 같은 순서로 적고 있으면 이 change 는 열거를 쓰는 것이다. 표지 정규화(공백·강조·들여쓰기)도 함께. 다섯 공격 전부 강제가 켜진 채 남는다 |
| P2-2 | `_call_names` 는 dict 아닌 노드를 `(unnamed)` 로 채우고 `_coords` 는 건너뛰어, **맞는 문서**가 어긋난 목록으로 터졌다 | 같은 방식으로 건너뛴다. 리뷰어의 입력으로 findings 0 확인 |
| P2-3 | "철자 없는 노드 15 개(저장소에)" — 15 는 표지를 단 133 개 안의 수다 | 저장소 전체 **281 개 / 189 번들**로 정정(코드 주석과 review 양쪽) |
| P2-4 | 고친 뒤 직접 실행을 29 로 적었는데 인도된 트리는 34였고, 34 를 fix4 의 54·fix5 의 62 와 같은 라벨에 뒀다(단위가 다르다) | 값과 단위를 함께 적었다. 같은 기준(`tools/logic-map` 전체 discover)이면 fix6 은 67, 지금은 **75** |
| P2-5 | fix6 실측표가 게이트 전 항목을 빠뜨렸다 | fix6 에 그 사실을 적고, 인도된 트리 기준 전 항목을 아래 표에 적는다. branch-test-map 의 절 제목(fix5→fix5·fix6)과 중복된 꼬리말도 정리 |
| P2-6 | tasks 6.2 의 "두 번째가 무엇을 비교하는지 증명한다" | 어느 반쪽을 덮고 어느 반쪽을 안 덮는지로 다시 썼다 |

### 이 라운드가 만든 검사기 변경과 그 반증

| 바뀐 것 | 되돌렸을 때 |
|---|---|
| 강제 판정을 표의 **내용**으로 | `TheMandateIsAPropertyOfTheTableNotOfItsHeading` 2 개만 FAIL |
| 표지 철자 **정규화** | `AHeadingThatLooksTheSameIsTreatedTheSame` 3 개만 FAIL |
| 섹션의 **모든 표**를 판정에 넣음 | `TheMandateIsNotTurnedOffByALineAboveTheEnumeration` 1 개만 FAIL |
| `_call_names` 를 `_coords` 와 같게 | `TheNodeListsMustAgreeAboutMalformedNodes` 1 개만 FAIL |

**남는 잔여는 하나이고 숨길 수 없는 것이다.** 열거를 정말로 그만두면(좌표를 지우면)
강제가 꺼진다. 그때는 완전성 주장도 함께 사라지므로 감사만 조용히 꺼지는 일은 없다.
저장소 3014 번들로 재면 이 판정에 걸리는 change 는 여전히 a112 하나다 — 다른 16 개
change 의 손으로 쓴 표 중 우연히 1:1 인 것은 없다.

### 실측 (5.5-fix7, 인도된 트리)

| | |
|---|---|
| `make lint` | PASS |
| `make test` | **98 ok / 0 FAIL** |
| `make test-seams` | **99 ok / 0 FAIL** |
| `test_role_check.py` | **42 PASS** (직접 실행 = discover) |
| `tools/logic-map` 전체 discover | **75 PASS** |
| 저장소 전수 스윕 | **3014 번들 / 93 change / 강제 1 / findings 0** |
| `check_analysis --change a112` | evidence complete or diff-proven exempt |
| `openspec validate --strict` | valid |
| gofmt | 0 |
| 술어 뮤턴트 4 종 | 위 표대로, 저장소 sha256 `1d710710…` 불변 |
| `make sdd-sync` | all indexes current |
| `make sdd-check` | 외부 a117 스텁에서만 실패(제 것 아님) |
| `make gate` | **not-applicable** — change 완료 게이트, a112 는 미완료 태스크가 남아 2/10 에서 멈춘다 |

## 5.5-fix8 — 7차 적대 리뷰가 낸 P1 하나와 P2 넷

7차는 P0 0 · P1 1 · P2 4. **세 라운드 연속으로 같은 병이 한 단계씩 내려갔다** —
5차는 "블록이 있느냐"만 봤고, 6차는 논리합의 포섭된 반쪽만 봤고, 7차는 그 반쪽
안에서 **가격 두 개를 함께** 움직였다.

### P1 — 세 번째 시험이 stop 과 target 을 동시에 옮겨서, 한 필드 누락은 못 잡았다

6.2 가 실제로 할 편집은 논리항 삭제가 아니라 **비교를 필드별로 펼치면서 하나를
빠뜨리는 것**이다. 리뷰어가 `:223` 을
`Lineage.Identity || Entry || EffectiveStop || Target || Quantity` 로 펼친 뒤 하나씩
빼 보니, fix7 의 세 시험이 **전부 초록**이었고 각각 그 가격만 바꾼 제안으로 주문이
나갔다 — `EffectiveStop` 을 뺐을 때는 손절 95→80 짜리 1차 진입이 나가면서 엔진
패키지가 두 스위트 모두 초록이었다.

고침은 세 번째 시험을 **축 하나씩 세 하위 시험**으로 나눈 것이다. 각 케이스는
`assertOnlyOnePriceMoved` 로 **자기 축 말고는 아무것도 안 움직였다**는 것을 스스로
확인한다 — lineage identity 가 같고, 수량이 같고, 나머지 두 가격이 같아야 한다.

| 뺀 필드 | 빨개지는 하위 시험 |
|---|---|
| `Entry()` | `…/entry 만` **만** |
| `EffectiveStop()` | `…/stop 만` **만** |
| `Target()` | `…/target 만` **만** |

**수량은 덮지 못했다고 적는다.** 시험 seam 에서 `PlannedCeiling` 이 수량과 묶여
있어, lineage 를 고정한 채 수량만 바꾼 제안을 만들 수 없다. B3 의 branch-test-map
에 그 칸을 "덮이지 않음"으로 적었다 — 침묵한 생략은 금지다.

### P2 넷

| # | 무엇 | 고침 |
|---|---|---|
| P2-1 | "남는 회피는 열거를 그만두는 것뿐이고 숨길 수 없다"가 **거짓**이었다. 열거 표를 통째로 다른 `##` 절로 옮기면 표지도 행도 그대로 남은 채 강제가 꺼졌다 — 판정을 절 이름에 걸었으니 그것도 이름 검사다 | 판정이 **문서 전체**의 표를 본다. 그리고 `check_analysis` 가 번들의 두 산문 파일을 이어서 넘긴다. 이동 공격은 이제 `True` 로 남는다. 남는 회피를 정확히 적었다: 그 change 의 어느 산문에도 1:1 표가 없어야 하고, 그 길은 (a) 좌표 지우기 (b) 번들 밖으로 옮기기 둘뿐이며 둘 다 133 개를 다 고쳐야 한다 |
| P2-2 | `_call_names` 와 `_coords` 가 여전히 한 단계 더 어긋났다 — `at` 이 없거나 dict 가 아닌 노드에서 갈렸고, **맞는 문서**가 터졌다 | 두 목록을 `_node_pairs` 하나에서 만든다. 구조상 어긋날 수가 없다 |
| P2-3 | tasks 6.2 가 `:136` 에서 "all three" 라 해 놓고 `:138` 에서 "both tests above" 로 끝났다 — 구현자가 마지막에 읽는 문장이 옛 값이었다 | "all three tests above" |
| P2-4 | `role_check.py` 모듈 주석이 아직 **표지가 강제를 켠다**고 적고 있었다(네 문장) | 다시 썼다. 표지는 스위치가 아니라 "어느 표를 대조할지" 고르는 것뿐이고, 판정은 내용으로 한다고 적었다. 뚫린 세 라운드도 함께 |

**리뷰어가 자기 6차 권고를 정정한 것도 적어 둔다.** "`ast.json` 만으로 무조건
강제하라"를 재 보니 **3014 중 2881 개가 16 개 change 에서 터진다.** change 단위
절벽은 오늘 피할 수 없는 값이고, 그래서 내용 기준 판정이 맞는 선택이었다.

### 이 라운드가 만든 검사기 변경과 그 반증

| 바뀐 것 | 되돌렸을 때 |
|---|---|
| 판정이 문서 전체를 봄 | `test_moving_the_table_to_another_section_does_not_turn_it_off` 1 개만 FAIL |
| 좌표·철자를 `_node_pairs` 하나에서 | `test_a_malformed_node_does_not_shift_the_name_list` 의 하위 2 개만 FAIL |

### 실측 (5.5-fix8, 인도된 트리)

| | |
|---|---|
| `make lint` | PASS |
| `make test` | **98 ok / 0 FAIL** |
| `make test-seams` | **99 ok / 0 FAIL** |
| `test_role_check.py` | **43 PASS** (직접 실행 = discover) |
| `tools/logic-map` 전체 discover | **76 PASS** |
| 저장소 전수 스윕(두 파일 기준) | **3014 번들 / 93 change / 강제 1 / findings 0** |
| `check_analysis --change a112` | evidence complete or diff-proven exempt |
| `openspec validate --strict` | valid |
| gofmt | 0 |
| 필드 누락 확장 3 종 | 각각 해당 하위 시험만 RED, 저장소 sha256 `1d710710…` 불변 |
| `make sdd-sync` | all indexes current |
| `make sdd-check` | 외부 a117 스텁에서만 실패(제 것 아님). 리뷰어는 mutating 이라 안 돌렸다 — 정당한 판단이고, 그래서 이 두 줄은 제 실행값이다 |
| `make gate` | **not-applicable** — change 완료 게이트, a112 는 미완료 태스크가 남아 2/10 에서 멈춘다 |

## 5.5-fix9 — 8차 적대 리뷰(P1 0)의 P2 넷

8차는 **P0 0 · P1 0 · P2 4** 로 APPROVE WITH FINDINGS 였다. 네 라운드 만에 처음으로
P1 이 없다 — 축 공격이 통하지 않았고, 리뷰어가 주문을 내는 입력을 만들지 못했다.
그래도 P2 넷을 다 닫았다. 둘은 "남는 회피는 X 뿐"이라는 문장이 **또** 틀린 것이었다.

### P2-1 — `PriceProvenance` 는 여덟 필드인데 시험은 숫자 하나만 움직일 수 있었다

`assertOnlyOnePriceMoved` 의 `==` 는 여덟 필드를 다 비교한다(Go 는 패키지 밖에서도
비공개 필드를 포함한 구조체 비교를 허용한다). 그런데 `AcceptedResultForAuthorityTest`
는 `source`·`version`·`digest` 를 역할별 상수로, `currency`·`minorScale` 을 시장에서,
`unitVersion` 을 하드코딩으로, `asOf` 를 `observedAt` 에서 만든다 — 그리고 `asOf` 의
원천인 `observedAt` 은 lineage 도 해시한다. **lineage 를 고정하면 움직일 수 있는
필드가 `priceMinor` 하나뿐이었다.**

그래서 `:223` 을 세 가격의 `priceMinor` 비교로 좁혀도 여섯 하위 시험이 전부 초록이고
엔진 패키지가 두 스위트 다 초록이었다.

**이것은 지어낸 위협이 아니다.** `internal/continuationlane/execution_terms.go` 는 같은
손절 숫자에 대해 신선한 후보와 저장된 권한에서 서로 다른 출처를 만든다
(`saved-effective-stop`/`stop-state-v1`). 그 패키지는 위조된 손절 출처를 이미 위협으로
다루고 시험도 있다(`execution_terms_test.go` 가 `Version: "forged"` 를 심는다).
1차 진입 backstop 에는 그 시험이 없었다.

고침은 둘이다.

1. `internal/strategyflow/authority_stop_provenance_testseam.go` (**새 파일**,
   `tossos_testseams`): `ResultWithRestatedStopProvenanceForTest` 가 이미 봉인된 제안의
   손절 **출처 셋만** 다시 적고 다시 봉인한다. 숫자와 lineage 는 그대로다.
2. `TestFirstLegAuthorityRefusesARestatedStopProvenanceAtTheSamePrice`. 시험이 스스로
   세 가격의 `priceMinor` 가 모두 같은지 확인한다 — 그 확인이 없으면 이 시험도 숫자
   비교로 통과하는 시험이 된다.

| 뮤테이션 | 결과 |
|---|---|
| 세 가격의 `priceMinor` 만 비교 | RED — 새 시험 **만** 실패 |

**여전히 안 덮인 것을 적는다.** 출처 축은 stop 하나로만 밟는다. entry·target 의 출처와
주문 수량은 덮이지 않았고, B3 의 표에 그렇게 적었다.

### P2-2 — "남는 회피"가 네 라운드 연속 코드보다 한 칸 넓었다

fix8 은 "(b) 번들 **밖** 다른 파일로 옮기기"라고 적었다. 그런데 `_bundle_text` 는
파일 **두 개를 열거**해서 읽고 있었고, `risk-pattern-report.md` 는 번들 필수 파일 넷
중 셋째다. 리뷰어가 133 개 전부의 열거를 거기로 옮기니 표는 머리글도 행도 그대로 살아
있는 채 강제가 꺼졌다.

| 열거가 있는 곳 | 강제 |
|---|---|
| 기준선 | True |
| 같은 파일의 다른 `##` 절 | True |
| `branch-test-map.md` | True |
| **`risk-pattern-report.md`** (고치기 전) | **False** |
| `risk-pattern-report.md` (고친 뒤) | **True** |

고침은 한 줄이다 — `_bundle_text` 가 번들 디렉터리의 `*.md` 를 **전부** 읽는다.
파일을 열거하지 않으므로 "번들 밖" 이라는 문장이 이제 **구조상** 참이다.

절 → 철자 → `##` 절 → 파일 두 개. 네 번 모두 같은 방향으로 틀렸다: 코드가 실제로
세는 **범위**를 확인하지 않고 문장을 썼다.

### P2-3 · P2-4

- **P2-3** 은 제 인계 메시지의 슬립이었다. 강제가 걸리는 a112 번들은 **133** 개이고,
  127 은 좌표를 적은 표의 수다(호출이 0 인 6 개도 표지 표로 열어야 하고 열고 있다).
  `review.md` 는 127 을 옳은 자리에만 쓰고 있었다 — 리뷰어가 확인했다.
- **P2-4** B3 의 가격 세 줄이 "가격 전체가 고정된 것처럼" 읽혔다. 세 줄을
  `숫자(priceMinor)` 로 한정하고, stop 출처 줄과 안 덮인 축 줄을 더했다.

### 이 라운드가 만든 변경과 그 반증

| 바뀐 것 | 되돌렸을 때 |
|---|---|
| `_bundle_text` 가 번들의 모든 `.md` 를 읽음 | `test_every_md_in_the_bundle_directory_is_read` 1 개만 FAIL |
| 손절 출처 축 시험 + 새 seam | `priceMinor` 만 비교하는 뮤턴트에서 그 시험 1 개만 FAIL |

### 실측 (5.5-fix9, 인도된 트리)

| | |
|---|---|
| `make lint` | PASS |
| `make test` | **98 ok / 0 FAIL** |
| `make test-seams` | **99 ok / 0 FAIL** |
| `tools/logic-map` 전체 discover | **77 PASS** |
| 저장소 전수 스윕(번들의 모든 `.md`) | **3014 번들 / 93 change / 강제 1(a112, 133 번들) / findings 0** |
| `check_analysis --change a112` | evidence complete or diff-proven exempt |
| `openspec validate --strict` | valid |
| gofmt | 0 |
| `make sdd-sync` | all indexes current |
| `make sdd-check` | 외부 a117 스텁에서만 실패(제 것 아님) |
| `make gate` | **not-applicable** — change 완료 게이트, a112 는 미완료 태스크가 남아 2/10 에서 멈춘다 |

## 5.5-fix10 — 9차가 낸 P1 하나, 그리고 **전략을 바꾼 라운드**

9차는 P0 0 · P1 1 · P2 3. P1 은 stop 출처 세 필드를 **함께** 옮긴 시험이라
`digest` 하나만 보는 술어에 안 걸린다는 것이었고, 리뷰어가 제 seam 으로 세 줄 만에
`stop.digest` 만 바꾼 제안을 만들어 **주문을 냈다**.

### 왜 여덟 번째 하위 시험을 만들지 않았나

`executionTermsIdentity` 가 담는 스칼라는 **32 개**다 — 계좌·시장·종목·캠페인·
legOrdinal·수량·정책·lineage 여덟에, 가격 셋 × 필드 여덟. 5~9 차는 라운드마다
축 하나를 시험으로 샀고 매번 다음 축이 남았다. 32 개를 엔진 fixture 로 사려면
대부분에 대해 **그 축만 움직이는 seam 을 새로 만들어야 한다** — fix9 가 그것을
하나 만드느라 새 exported 함수를 하나 늘렸다. 이 방식은 종료하지 않는다.

의무를 둘로 나누면 각각 끝난다. 리뷰어의 권고이고, 그대로 했다.

| 무엇을 증명하나 | 어디서 | 왜 거기서 끝나나 |
|---|---|---|
| identity 가 **모든 필드**를 담는다 | `internal/strategyflow/execution_terms_identity_fields_test.go` (신규) | 필드를 가진 타입 안이라 fixture 도 seam 도 필요 없다. 32 개를 하나씩 바꿔 해시가 달라지는지 보고, **개수는 `reflect` 로 타입에서 읽는다** — 필드가 늘면 시험이 터져 표를 늘리게 만든다 |
| 이 가드가 **그 identity 를 그대로** 비교한다 | `internal/app/engine/strategy_first_leg_backstop_shape_test.go` (신규) | 단언 **하나**. `:223` 의 조건이 정확히 두 identity 비교의 논리합인지 AST 구조로 본다 |

**이것은 이 change 가 다섯 라운드에 걸쳐 지운 철자 census 가 아니다.** 그것들은
"토큰 X 가 범위 어딘가에 있는가"를 물었고, X 가 다른 곳에 나타나면 뚫렸다 —
완전성이 세는 **범위**의 함수였다. 여기서는 **그 비교의 피연산자가 이 식인가**를
묻는다. 가드 그 자체인 AST 노드에 대한 구조 등식이라 다른 곳의 등장으로는 만족시킬
수 없다. 그리고 혼자 서지 않는다: 행동 시험 넷(하위 여섯)이 "이 가드가 돌고 거절한다"를
증명하고, 32 필드 표가 "비교되는 값이 모든 필드를 담는다"를 증명한다.
**행동 + 구조 + 해시 완전성** 셋이 함께라야 닫힌다.

### 반증

32 필드 표 — `types.go` 의 해시에서 필드를 빼 본다(overlay, 원본 불변):

| 뺀 것 | 결과 |
|---|---|
| 가격의 `digest` | RED — `entry.digest`·`stop.digest`·`target.digest` 셋만 |
| `terms.quantity` | RED — `quantity` 하나만 |

구조 단언 — **소스를 실제로 바꿔야 한다.** `parser.ParseFile` 은 디스크를 읽으므로
`go test -overlay` 가 안 보인다. 처음에 overlay 로 재서 여섯 개가 전부 GREEN 이
나왔고, 그것은 가드가 약하다는 뜻이 아니라 **제 반증 방법이 틀렸다는 뜻**이었다.
백업 파일에서 복원하고 sha256(`1d710710…`)을 매 회 확인하며 다시 쟀다.

| `:223` 을 바꾼 것 | 구조 단언 |
|---|---|
| `Lineage.Market` 만 · `Lineage.Identity` 만 · `CampaignID` 만 | RED |
| `ExecutionTerms.Identity()` 만 (**행동상 동등 변이**) | RED |
| 세 `priceMinor` 만 · 거기에 `EffectiveStop().Source()` 를 더한 것 | RED |

넷째 줄이 요점이다. 포섭된 항을 지운 것이라 **행동으로는 어떤 시험도 구별할 수
없다.** 구조는 구별한다.

### P2 셋

| # | 무엇 | 고침 |
|---|---|---|
| P2-1 | "남는 회피" 문장이 **다섯 번째로** 코드보다 한 칸 넓었다. `_bundle_text` 가 `*.md` 를 globbing 했고 `.md` 도 목록이다 — `notes.txt`·`notes.mdx`·`notes.markdown` 로 강제가 꺼졌다 | 이름으로도 확장자로도 안 거른다. 번들 디렉터리에서 **읽히는 파일 전부**를 넘긴다(디코딩 실패만 건너뛴다). 네 공격 모두 `True`. 시험도 그 세 확장자를 포함하게 넓혔다 |
| P2-2 | `strategyflow` 의 exported 표면에 census 가 없다 — 이제 봉인된 `Result` 를 만드는 함수가 **둘**이다 | **부채로 이름만 적는다.** `strategyhandoff` 가 4차 전에 있던 상태와 같다. 이 change 의 파일 소유 밖이고, census 를 하나 더 만드는 것은 이 라운드가 지운 검사와 같은 종류를 늘리는 일이다 |
| P2-3 | B3 의 stop 출처 줄이 "출처가 고정된 것처럼" 읽혔다 | `세 필드를 함께 바꾼 경우만` 으로 한정하고, `digest` 하나는 그 시험이 못 잡는다고 적었다. 나머지 축은 구조 단언과 32 필드 표가 덮는다고 표에 명시 |

### 실측 (5.5-fix10, 인도된 트리)

| | |
|---|---|
| `make lint` | PASS |
| `make test` | **98 ok / 0 FAIL** |
| `make test-seams` | **99 ok / 0 FAIL** |
| `tools/logic-map` 전체 discover | **77 PASS** |
| 저장소 전수 스윕(읽히는 파일 전부) | **3014 번들 / 93 change / 강제 1(a112, 133 번들) / findings 0** |
| `check_analysis --change a112` | evidence complete or diff-proven exempt |
| `openspec validate --strict` | valid |
| gofmt | 0 |
| 뮤테이션 원복 | 백업+sha256 확인(구조 단언은 디스크를 읽으므로 overlay 로는 반증되지 않는다) |
| `make sdd-sync` | all indexes current |
| `make sdd-check` | 외부 a117 스텁에서만 실패(제 것 아님) |
| `make gate` | **not-applicable** — change 완료 게이트, a112 는 미완료 태스크가 남아 2/10 에서 멈춘다 |

## 5.5-fix11 — 10차가 낸 P1 둘: 전략을 끝내려고 만든 두 산출물 안에 같은 병이 있었다

10차는 P0 0 · P1 2 · P2 3. 전략 분할 자체는 옳다고 확인됐고, **양쪽 반쪽이 각각
자기 범위를 안 못 박은 것**이 P1 이었다.

### P1-1 — 조건은 구조로 보면서 본문은 존재로 봤다

`refusesWith` 는 "이 블록 **어딘가에** 거절 문구가 있는가"를 물었다. 그것은 이
change 가 다섯 라운드에 걸쳐 지운 존재 검사다. 리뷰어가 조건을 **한 글자도 안 바꾸고**
본문 안에 분기를 넣었다:

```go
if <조건 그대로> {
    if result.Lineage.PlannedCeiling != accepted.result.Lineage.PlannedCeiling {
        result = accepted.result          // 건너온 값을 그대로 믿는다
    } else {
        return execgw.QFinalCampaignFirstLegIssuance{}, errors.New("production proposal identity changed")
    }
}
```

구조 단언 초록, 행동 시험 넷(하위 여섯) 초록, 두 스위트 초록, 그리고 **위조된 수량 9 ·
손절 80 짜리 1차 진입**. fixture 가 전부 ceiling 8 이라 그 탈출구를 아무도 안 밟았다.

고침은 본문도 구조로 보는 것이다 — `requireOnlyTheRefusal`: 문장 **하나**, 반환 둘,
첫째가 빈 발주, 둘째가 그 문구의 `errors.New`. 실제 파일에 그 뮤턴트를 넣어 재니
`가드 본문이 두 값을 돌려주는 return 이 아니다: *ast.IfStmt` 로 **RED**(sha256
`1d710710…` 복원 확인). 이 요구는 유한하다 — 더해도 다음 단계가 생기지 않는다.

`refusesWith` 는 남겼지만 **가드를 찾는 데만** 쓴다. 찾는 데 존재 검사를 쓰는 것과
증명하는 데 쓰는 것은 다르고, 그 구분을 코드 주석에 적었다.

### P1-2 — 32 필드 census 가 세 타입 중 둘만 읽었다

`scalars = terms.NumField() - 1 - 3 + 3*price.NumField()` 는 `ExecutionPolicy` 를
스칼라 **하나로 하드코딩**했다. 그 타입은 아홉 필드이고 `executionTermsIdentity` 는
`policy.identity` 하나만 해시한다 — **여덟 필드가 세는 범위 밖에 있었다.**

| overlay | 앞 판본 | 지금 |
|---|---|---|
| `ExecutionPolicy` 에 필드 추가 | 초록 | **RED** — `위임 9(want 8)` |
| `PriceProvenance` 에 필드 추가(대조군) | RED | RED |

고침은 둘이다. (a) 산술이 **세 타입 모두**의 `NumField()` 를 읽는다. (b) 여덟 필드가
`policy.identity` 로 **위임**되는지를 `TestBreakoutPolicyIdentityChangesWithEveryFieldItCovers`
가 잰다 — `capSnapshotID` 를 정책 해시에서 빼면 그 케이스만 RED.

**위임이 확인되는 범위를 적는다.** continuation·reversal 은 나머지 여덟을 비운 채
`ExecutionPolicy{identity: …}` 를 만들므로 잃을 값이 없다. **weekly 는 다르다** —
`weeklyPolicy` 가 아홉을 다 채우면서 identity 는 lane 이 계산한 것을 받는다. 그 유도는
이 패키지가 확인하지 않는다. **확인 안 된 위임으로 남긴다.**

### P2 셋

| # | 무엇 | 고침 |
|---|---|---|
| P2-1 | "행동 시험 **일곱**" — 실제는 함수 넷, 하위 시험 여섯. 같은 로트 안에서 셋은 맞고 둘은 틀렸다 | 두 사본 다 "넷(하위 여섯)" |
| P2-2 | fix10 실측표가 `sdd-sync`·`sdd-check`·`make gate` 세 줄을 빠뜨렸다 — 8차 P2-5 와 같은 회귀 | 세 줄 복원 |
| P2-3 | B3 가 구조 단언이 **잡는 것**만 적었다 | 못 잡던 것(본문 분기)과 지금도 안 보는 것(가드 앞뒤)을 같은 표에 적었다 |

### 실측 (5.5-fix11, 인도된 트리)

| | |
|---|---|
| `make lint` | PASS |
| `make test` | **98 ok / 0 FAIL** |
| `make test-seams` | **99 ok / 0 FAIL** |
| `tools/logic-map` 전체 discover | **77 PASS** |
| 저장소 전수 스윕 | **3014 번들 / 93 change / 강제 1(a112, 133 번들) / findings 0** |
| `check_analysis --change a112` | evidence complete or diff-proven exempt |
| `openspec validate --strict` | valid |
| gofmt | 0 |
| `make sdd-sync` | all indexes current |
| `make sdd-check` | 외부 a117 스텁에서만 실패(제 것 아님) |
| `make gate` | **not-applicable** — change 완료 게이트, a112 는 미완료 태스크가 남아 2/10 에서 멈춘다 |

## 5.5-fix12 — 11차: 탈출구를 한 줄 위로 옮기면 그대로 통했다

11차는 P0 0 · P1 1 · P2 3.

### P1 — 10차 뮤턴트를 가드 **밖으로 한 줄** 옮기면 새 본문 검사가 안 본다

```go
if result.Lineage.PlannedCeiling != accepted.result.Lineage.PlannedCeiling {
    result = accepted.result
}
<가드 노드 한 바이트도 안 바뀜>
```

대입이 비교보다 먼저이므로 가드가 볼 때는 identity 가 이미 같고, 가드는 옳게 아무것도
안 한다. 구조 단언 초록 · 행동 시험 넷(하위 여섯) 초록 · 두 스위트 초록 · **위조된
수량 9 · 손절 80 짜리 1차 진입**.

**그리고 B3 에 적어 둔 문장이 틀렸다.** "가드 앞은 다른 가드(B2·B4)와 행동 시험이
맡는다"고 적었는데 재 보니 **아무도 안 맡고 있었다.** 구멍을 이름만 적고 주인은
확인하지 않은 것이다 — 이 로트가 "세는 범위를 확인하라"를 배운 것과 같은 실수의
다른 얼굴이다.

고침은 두 술어이고 유한하다.

| 요구 | 무엇을 막나 |
|---|---|
| 가드는 `result := proposalAuthority.Proposal()` **바로 다음** 최상위 문장 | 위로 옮긴 탈출구, 그리고 감싸기 |
| `result` 는 함수 전체에서 **한 번만** 대입 | 가드 뒤에서 다시 대입하는 변형 |

실제 파일에 넣어 재고 sha256 `1d710710…` 로 매번 복원 확인:

| 뮤턴트 | 결과 |
|---|---|
| 탈출구를 가드 위로(11차의 H) | RED — `가드 바로 앞 문장이 …가 아니다` |
| 가드를 바깥 `if` 로 감싸기 | RED — `가드 2 개` |
| 가드 뒤에서 재대입 | RED — `result 가 2 번 대입된다` |

### P2 셋

**P2-1 — 감싸기는 *우연히* 막히고 있었다.** `refusesWith` 가 재귀해서 감싼 `if` 도
후보로 잡히는 것이 그 방어였다. `refusesWith` 를 직속 문장만 보도록 "정리"하면
조용히 사라진다. 그 상태를 실제로 만들어(생산 파일 + 시험 파일 둘 다 백업·복원)
재 보니 **최상위 문장 요구가 독립적으로** 잡았다: `가드가 함수 최상위 문장이 아니다`.
의존을 주석에도 적었다.

**P2-2 — weekly 위임은 확인 가능했고 성립한다.** 리뷰어가 두 홉을 추적했다:
`executionPolicy` 의 `Identity` 는 `{seal, DecisionDigest, CalendarDigest,
CapSnapshotID}` 에서 나오고 `seal` 은 나머지 다섯 값을 담는다. "확인 안 됨"으로
남기는 대신 그 패키지에 여덟 케이스를 뒀다
(`internal/weeklyvaluelane/execution_policy_identity_fields_test.go`).
preimage 에서 `exitCosts` 를 빼면 `ExitCostsMinor` 케이스만 RED.
값 다섯은 preimage 를 바꾸고 **다시 봉인해서** 넣는다 — 봉인 안 한 preimage 는
`valid()` 가 거절하므로 그 경로가 실제 경로다.

**P2-3 — `NumField()-1` 은 세기만 했다.** 이제 `identity` 필드가 실재하는지 확인하고,
나머지 여덟을 **이름으로** 표와 대조한다. 다만 정직하게 적는다: **리네임을 실제로
막는 것은 컴파일이다.** 생산 파일에서만 `capSnapshotID` 를 리네임해 보니 시험이
그 필드를 직접 쓰므로 빌드가 깨진다. 이름 루프는 보조다.

**반증 하나가 무효였다.** 처음에 `*.go` 전체를 리네임했더니 시험 파일까지 함께 바뀌어
초록이 나왔다 — 뮤테이션이 자기가 잡아야 할 시험도 같이 고친 것이다. 생산 파일만
바꾸도록 고쳐 다시 쟀다.

### 실측 (5.5-fix12, 인도된 트리)

| | |
|---|---|
| `make lint` | PASS |
| `make test` | **98 ok / 0 FAIL** |
| `make test-seams` | **99 ok / 0 FAIL** |
| `tools/logic-map` 전체 discover | **77 PASS** |
| 저장소 전수 스윕 | **3014 번들 / 93 change / 강제 1(a112, 133 번들) / findings 0** |
| `check_analysis --change a112` | evidence complete or diff-proven exempt |
| `openspec validate --strict` | valid |
| gofmt | 0 |
| `make sdd-sync` | all indexes current |
| `make sdd-check` | 외부 a117 스텁에서만 실패(제 것 아님) |
| `make gate` | **not-applicable** — change 완료 게이트, a112 는 2/10 에서 멈춘다 |

## 5.5-fix13 — 12차(P1 0)의 P2 둘

12차는 **P0 0 · P1 0 · P2 2** 로 APPROVE WITH FINDINGS. 두 번째로 P1 이 없는 라운드다.

### P2-1 — 사슬의 **첫 고리**가 안 잡혀 있었다

`statementText` 는 렌더된 문자열 비교이고, 그것은 존재 검사가 한 단계 올라간 것이다 —
토큰이 아니라 **문장 하나의 철자**를 본다. 재유도는 두 고리인데 고리 2 만 잡았다.

리뷰어가 넣은 것은 **6.2 가 실제로 쓸 편집**이다: 네 가족 중 조정자가 고른 항목을
loop 로 찾아 `proposalAuthority` 를 다시 고른다. 고리 2 와 가드는 한 바이트도 안
바뀌고 구조 단언·행동 시험 여섯·census 가 전부 초록이다. 그러면 가드가 **자기 참조**가
된다 — accepted 와 맞는 항목을 골라 놓고 그것을 accepted 와 비교한다.

**오늘은 해가 없다.** `:217` 이 `len(entries) != 1` 로 항목을 하나로 묶고, census 가
그 관문을 잡고 있다. **그 둘이 정확히 6.2 가 바꿀 것**이므로 지금 닫았다.

이제 세 이름(`proposal`·`proposalAuthority`·`result`)이 각각 한 번만 대입되고,
두 고리가 가드 바로 앞에 이 순서로 있어야 한다. 실제 파일 뮤테이션(sha256
`1d710710…` 복원 확인):

| 뮤턴트 | 결과 |
|---|---|
| 고리 1 뒤 loop 로 재선택(6.2 의 모양) | RED — 사슬과 대입 수 **둘 다** |
| `proposal` 재대입 | RED |

### P2-2 — `&result` 는 대입으로 안 세고 있었다

클로저 안의 `result = …` 와 안쪽 스코프의 `result :=` 는 이미 세고 있었다.
빠진 것은 포인터 넘김이다 — `&result` 는 `*ast.UnaryExpr` 라서 대입만 세면 안 보인다.

리뷰어는 이것을 **무해하다고 재고** P2 로 냈다: 가드가 뒤따르는 모든 문장을 지배하므로
가드 뒤의 채택은 identity 가 이미 같을 때만 돌고, identity 동일은 봉인된 값 전체의
동일이라 채택이 무의미하다. 그래도 세는 편이 한 줄이라 세기로 했고, 그 이유를 주석에
적었다. `_ = &result` 를 넣으면 RED.

### 리뷰어가 이 로트의 **방법 정정 두 개**를 근거로 인정했다

`-overlay` 가 디스크를 읽는 파서에 안 보인다는 것(10차)과, 리네임 뮤테이션이 시험
파일까지 고쳐 자기를 잡아야 할 시험을 수리했다는 것(11차). 둘 다 **틀린 반증**이었고
둘 다 기록에 남겼다. 반증 방법이 틀리면 결과는 "가드가 약하다"가 아니라 "안 쟀다"이다.

### 실측 (5.5-fix13, 인도된 트리)

| | |
|---|---|
| `make lint` | PASS |
| `make test` | **98 ok / 0 FAIL** |
| `make test-seams` | **99 ok / 0 FAIL** |
| `tools/logic-map` 전체 discover | **77 PASS** |
| 저장소 전수 스윕 | **3014 번들 / 93 change / 강제 1(a112, 133 번들) / findings 0** |
| `check_analysis --change a112` | evidence complete or diff-proven exempt |
| `openspec validate --strict` | valid |
| gofmt | 0 |
| `make sdd-sync` | all indexes current |
| `make sdd-check` | 외부 a117 스텁에서만 실패(제 것 아님) |
| `make gate` | **not-applicable** — change 완료 게이트, a112 는 2/10 에서 멈춘다 |

## 2026-09-02 L5 태스크 5.1.1 — FamilyWorker 도입 (구현 기록)

사람이 고른 범위: **타입 도입만.** 여덟 worker 와 폐포 증명을 새 패키지에 세우고
엔진 배선은 건드리지 않는다. 설계가 정한 순서가 그것이고(`design.md:255` — 여덟
worker 와 두 조정자를 먼저 dormant/shadow 로 세운다), 그래야 기존 High-risk 함수
내부를 열지 않으므로 Pre-Edit 선언이 필요 없는 Normal 위험 로트가 된다.

### 로트를 열기 전에 잰 것

| 물음 | 답 | 근거 |
|---|---|---|
| `MarketCoordinator` 는 있는가 | **있다** — 5.4.2 가 랜딩했다 | `internal/strategycoordinator/coordinator.go:151` |
| `FamilyWorker` 는 있는가 | **없다** | `grep -rn FamilyWorker --include=*.go .` 무출력 |
| 5.1 의 넷 중 무엇이 남았는가 | 타입 키·불변 봉투·병합 큐는 이미 있고, **versioned worker runtime policy 만** 없다 | `Key:57` · `Envelope:104` · `Submit:203` / `RuntimePolicy` 유일 히트는 무관한 journal 시험 |
| 5.5 가 넘긴 주장은 오늘 참인가 | **거짓이다** | `productionStrategyWorker:364` 가 `Cycle` 을 `c *Context` 클로저로 만든다. AST 산출물이 `Context.runProductionStrategyMarketCycle` 의 호출로 `c.Journal.CurrentPositionCampaignCAS`(`:446`)와 `fresh.dispatch.dispatch`(`:453`)를 열거한다 — 손으로 읽은 것이 아니다 |
| 계약 값은 어디서 읽는가 | 동결 골든 | `worker_key_fields` 네 개 · `worker_count` 8 · `descriptors` 여덟 줄(전부 OFF/OFF/UNOBSERVED) |

### 설계 판단 하나 — 왜 레인 권한 객체를 받지 않는가

`internal/strategyproposal` 은 `internal/journal` 을 직접 들여온다. worker 가
`strategyproposal` 의 레인 권한을 인자로 받으면 그 순간 **쓰기 가능한 원장이
worker 폐포 안에 들어온다** — 스펙이 이름으로 금지한 넷 중 하나다. 그래서 worker 는
봉인된 값(`strategyarbiter.Proposal` = `strategyflow.Result` + 봉인된 경로 권한)만
받고, 그 조합의 전이 폐포는 `go list -deps` 로 재어 14 개 내부 패키지이며 그 안에
journal·execgw·official·app 이 없다.

### 남긴 것

`Run` 이 하는 판정은 **하나**다: 이 제안이 이 worker 의 레인인가. 범위·봉인 재확인·
용량은 조정자가 이미 판정하므로 다시 세지 않는다 — 조정자 `Submit` 이 가족을 읽는
자리에 같은 이유의 주석이 있다("같은 판정을 두 곳에 두면 운영자가 보는 진단이 갈린다").

결과 종류는 셋을 **따로** 둔다(`EMITTED`/`DORMANT`/`REFUSED`). "봉투 없음" 하나로
뭉치면 잠든 worker 와 거절당한 제안이 같은 값이 되고, 그것이 5.4.3 이 고친 혼동
(사라진 제안과 없던 제안)과 같은 모양이다.

거절 코드는 새로 만들지 않았다. 골든의 `refusal_enums.arbitration` 여섯 개가 계약이고
`RefusalSealMismatch` 가 "봉인으로 신원을 다시 세울 수 없다"를 이미 덮는다. 새로 만든
두 문자열은 `Detail` 이며 계약이 아니다.

### 반증

16 뮤턴트 + 폐포 반증 2 종, 전부 CAUGHT, 매번 해시로 원복 확인.

**1 차에서 넷이 살아남았다** — 시장·레인 ID·레인 버전·지평 비교를 **하나씩** 지워도
열두 시험이 초록이었다. 우연이 아니다: 가족 유도(`strategyarbiter.familyScore`)가
이미 (horizon, lane_id, lane_version) 세 축으로 점수 행을 찾고, 레인 ID 는 시장마다
다르다. 그래서 한 축을 지워도 남은 축이 같은 입력을 막는다.

이것을 "동등 변이"로 넘기지 않았다. 넷을 **함께** 지우면 US worker 가 KR 제안을
받는다 — 각 비교는 정답에 필요하고 개별 관측만 안 되는 것이다. 축마다 시험을 하나씩
사는 일이 끝나지 않는다는 것은 이 change 가 5.5 에서 13 라운드에 걸쳐 얻은 결론이고,
답도 같다: 행동은 "가드가 돌고 거절한다"를, 구조(`worker_shape_test.go`)는 "가드가
이 다섯을 비교한다"를 증명한다. 구조 단언을 넣은 뒤 넷이 전부 죽었고, 삭제가 아닌
**바꿔치기**(엉뚱한 피연산자·순서 뒤집기·`&&`→`||`)와 골든 목록 훼손(하나 빼기·순서
바꾸기·하나를 ON 으로 태어나게·열쇠 필드 이름 바꾸기)도 함께 죽는다.

폐포 반증은 `internal/journal` 과 `internal/execgw` 를 실제로 import 해 본 것이고,
직접 허용목록과 전이 걸음 **양쪽**에서 걸린다.

### 여전히 열려 있는 것

- 생산은 아직 `StrategyMarketWorker` 를 돌린다. "`StrategyMarketWorker` 뒤에 숨기지
  않는다"의 나머지 절반은 5.2 의 문장 그대로이고, 그전에 dispatch handoff 를 spy 로
  막는 것이 설계 순서다.
- 5.1 의 남은 4 분의 1(서버 소유 versioned worker runtime policy)은 손대지 않았다.
- 새 패키지는 아직 생산 호출자가 0 이다. 골든 대조 시험이 여덟을 생산 상수에 묶어
  두므로 조용히 썩지는 않지만, 배선 전까지는 쓰이지 않는 코드다.

### Function Logic Map

**not-applicable — 새 패키지의 새 파일뿐이고 기존 함수 내부를 편집하지 않았다.**
`check_analysis.py` 자신의 판정과 같다(`:143`–`:145`: 새 파일에 새로 생긴 함수는
비교할 base 로직이 없으므로 "수정된 기존 함수"로 요구하지 않는다). 이 기록이 열거하는
기존 함수 내부 사실(위 표의 `:446`·`:453`)은 손으로 읽은 것이 아니라 이미 있는 AST
산출물에서 읽었다.

### 태스크를 나눴다 (2026-09-02, 사용자 결정)

위 "여전히 열려 있는 것"이 한 태스크 안에 랜딩한 절반과 안 한 절반을 같이 담고 있어서,
사용자가 나누라고 했다. 원래 5.1.1 문장의 **"rather than" 앞**이 5.1.1(타입 도입,
이제 `[x]`), **뒤**가 새 태스크 5.1.2(배선, `[ ]`)다.

나누면서 **의무는 하나도 안 빠졌다.** 원문의 모든 절이 둘 중 하나의 소유이고, 닫은
쪽에는 "이 태스크가 주장하지 않는 것" 문단을 남겨 생산 호출자가 0 이라는 사실이
`[x]` 뒤에 숨지 않게 했다. 5.5 본문의 소유자 표시도 함께 갈랐다 — 거기 적힌
"owned by 5.1.1"은 이제 새 타입의 폐포(5.1.1, 랜딩)와 생산 경로의 폐포(5.1.2, 열림)
둘로 나뉜다. 안 갈랐으면 닫힌 태스크가 열린 의무를 가리키게 된다.

`[x]` 로 닫은 근거는 이 절 위의 실측 그대로다 — 새로 돌린 것은 없고, 나눈 것은
태스크 문장이지 코드가 아니다.

### 런타임 정책과 레인 고장 상태 (2026-09-02, 5.1 닫힘 · 5.3 을 5.3.1/5.3.2 로 나눔)

`internal/strategyworker` 에 `RuntimePolicy`(policy.go)와 `Lane`(lane.go)이 섰다.
5.1 의 남은 4분의 1(서버 소유 versioned worker runtime policy)과 5.3 의 고장 절반을
같은 로트로 묶었다 — 정책만 떼면 5.1.1 본문이 이미 지적한 "소비자 없는 매니페스트"가
그대로 재현되기 때문이다. 레인이 임계값과 backoff 두 값을 실제로 쓴다.

**숫자는 하나도 고르지 않았다.** 영수증은 전부
`internal/app/engine/strategy_entry_supervisor.go` — 이 정책이 언젠가 대신할 런타임이다.

| 값 | 영수증 | 어떻게 묶었나 |
|---|---|---|
| cadence 5s | `PollInterval: DefaultStrategyCycleLimit` 세 자리(`:346`, `:391`, `:419`) | 엔진 디렉터리의 모든 비시험 파일에서 `PollInterval` 대입을 **열거**하고 전부 그 상수인지 본다 |
| queue depth 1 | 생산이 `QueueDepth` 를 한 번도 지정하지 않는다(`:510`–`:512` 가 기본값을 채운다) | 모든 `StrategyEntrySupervisorOptions` 리터럴에서 그 키의 **부재**를 요구한다 |
| deadline 30s | `CycleLimit: MaximumStrategyCycleLimit`(`:318`, `:352`) | 상수를 파싱해 평가 |
| backoff 5s/30s | `DefaultStrategyRestartStep`, `MaximumStrategyRestartBackoff` | 상수 + `strategyRestartBackoff` 본문의 구조 대조 |
| 실패 임계값 1 | `strategyMarketRuntime` 열두 필드 중 실패를 세는 것이 없다 → 첫 실패에 잠근다 | 열두 이름을 **전부** 세어 동결 목록과 견준다 |

임계값을 1 보다 크게 잡는 것은 완화다 — 오늘이라면 잠겼을 레인이 한 번 더 진입을
시도하게 된다. 그것은 서명된 매니페스트가 정할 일이라 여기서는 오늘 동작을 그대로 뒀다.
설계 고장표가 요구하는 "보통 오류는 세고 비정상은 즉시 잠근다"는 **모양**은 구현했고,
그래서 임계값이 1 이어도 비정상 경로는 따로 존재한다.

엔진을 import 하지 않고 소스를 파싱한 이유: import 하면 journal 이 이 패키지의 **시험**
폐포에 들어와 5.1.1 이 세운 `-deps-test` 걸음이 깨진다. 한 파일이 아니라 디렉터리를
읽는 이유: 파일 하나를 고르면 "완전하다"는 주장이 고른 사람의 것이 된다.

**뮤테이션 21 개, 전부 잡힘.** 원복은 해시 백업에서 했다 — 이 로트는 커밋 전이라
`git checkout` 을 쓰면 GREEN 까지 지워지고 `git diff --quiet` 가 그것을 초록으로 통과시킨다.
두 개가 처음 형태에서 살아남았고 둘 다 이유가 다르다.

- **닿지 않은 뮤테이션.** "레인 여덟을 공유한다"로 만든 첫 변이가 `defer` 로 캐시를
  채우면서 반환은 매번 새 집합을 주고 있었다. 살아남은 것이 아니라 **대상에 닿지
  않은 것**이다. 제대로 만든 변이(패키지 수준 `var sharedLanes = productionLanes()`)는
  잡힌다.
- **진짜 구멍 하나.** `attempt >= steps` 를 `attempt > steps` 로 바꿔도 초록이었다.
  생산 천장 30초가 계단 5초의 정확한 배수라 두 비교가 갈리는 유일한 시도에서 답이
  같기 때문이다. 결함이 사는 차원은 시도의 **크기**가 아니라 **가분성**이고, 이것은
  이 change 가 이미 적어 둔 규칙(반증은 결함이 사는 축을 바꿔야 한다)을 인용만 하고
  적용하지 않은 자리였다. 7초 계단·30초 천장 사다리를 커밋된 시험으로 심었고,
  반증 칸(4번째 시도는 28초가 아니라 30초)은 구현에서 다시 유도하지 않고 손으로 적었다 —
  유도하면 그 시험은 구현을 다시 쓴 것이 되어 아무것도 반증하지 못한다.

**구조 단언은 안 만들었고 그 이유를 적는다.** `Fail` 의 가드는 세 항(`이미 잠김`,
`비정상`, `임계값 도달`)인데 셋이 서로를 가려 주지 않는다 — 하나씩 지우면 각각 다른
시험이 빨개진다(M1·M2·M3). 5.5 와 5.1.1 에서 구조 단언이 필요했던 이유는 축들이 서로를
가려 행동으로 개별 관측이 안 됐기 때문이고, 여기서는 그 조건이 성립하지 않는다.

**닫지 않은 것, 이름을 적는다.** latch 는 프로세스 메모리에 있다. 재시작이 지우고,
durable 기록을 쓰거나 읽는 코드는 없다 — `Fault` 는 entry 를 되돌릴 방법이 없는 관측
값이고, 그 방법을 주는 순간 "이 패키지에는 latch writer 가 없다"는 폐포 증명이 거짓이 된다.
cadence·queue depth·deadline 세 값은 정책이 **들고만** 있고 강제하지 않는다. 그 둘이 5.3.2 다.
그리고 생산 호출자는 여전히 0 이다 — `ProductionLanes()` 를 부르는 곳은 5.1.2 다.

허용 목록에 표준 라이브러리 셋(`strconv`·`strings`·`time`)을 이름으로 열었다.
"표준이면 다 허용"으로 두면 `os/exec` 와 `net/http` 가 함께 열려 목록이 지키는 것이
없어진다. `strategycoordinator` 가 같은 방식으로 `sort`·`sync`·`time` 을 열어 둔 선례를 따랐다.

### 레인이 자기 시계를 갖는다 (2026-09-02, 5.3.2 · 5.3 의 두 번째 분할)

로트 범위는 5.3.2 의 세 절 중 둘이다 — 단일 비행 카덴스와 단조 마감 시한.
셋째(durable latch/recovery)는 5.3.3 으로 떼어 열어 두었다. 취향의 문제가 아니다:
`internal/strategyworker` 는 **쓰기 능력이 없다는 것이 증명된** 패키지이고
(`dependency_closure_test.go` 가 `-deps`·`-deps-test` 를 걷는다), 지속 latch 기록은
쓰는 쪽이 있어야 한다. 여기서 만들면 5.1.1 이 세운 성질이 사라진다.
`design.md:223` 도 "lane health/latch projection" 을 `internal/app/engine` 에 준다.

#### 영수증

| 만든 것 | 영수증 | 자리 |
|---|---|---|
| 감시견 세 갈래 | `invokeBoundedStrategyCycle` | `strategy_entry_supervisor.go:879`–`:899` |
| 마감 시한 초과 = abnormal·abandoned | 같은 함수의 셋째 갈래 | `:897` |
| 취소는 실패가 아니다 | `runMarket` 이 `if cancelled { return }` 을 오류 처리보다 앞에 둔다 | `:807` |
| 카덴스 기준점 = 사이클 시작 | `runStrategyPoller` 가 enqueue 직후에 잔다 | `:719`–`:727` |
| 투입 결과 세 철자 | `StrategyTriggerResult` 상수 | `:204`–`:207` |
| panic → 비정상 실패 | `invokeStrategyCycle` | `:901` |

인용 전에 두 함수의 `tools/logic-map` 번들을 먼저 만들었다
(`internal-app-engine--invokeboundedstrategycycle`,
`internal-app-engine--strategyentrysupervisor.runmarket`). 두 번들 다
`revision: base` 이며 이 change 는 두 함수를 편집하지 않는다 — 인용만 한다.

#### 번들이 **측정으로** 찾아낸 것: 엔진 스위트의 빈칸 일곱

커버리지를 주장하지 않고 쟀다. 시험마다 따로 `-coverprofile` 을 돌리고, 패키지
전체도 한 번 돌려 블록 `count` 를 읽었다. 어느 시험도 돌리지 않는 블록이 일곱이다:

- `invokeBoundedStrategyCycle` `894-896` — 마감 시한과 취소가 **함께** 오는 경합
- `runMarket` `787-790`, `821-824` — 잠금 자체가 실패하는 두 경로
- `runMarket` `795-796`, `829-830` — 재시작 대기가 계약 밖 지연으로 실패하는 두 경로
- `runMarket` `800-801` — 꺼졌거나 잠긴 worker 가 사이클을 건너뛰는 경로
- `runMarket` `813-814` — 권한 갱신 전용 worker 의 오류가 잠그지 않는 경로

공통점이 있다: **고장 처리 자체가 고장 나는 경로**다. 이 로트는 그것을 메우지
않는다(그 함수들을 편집하지 않으므로). 태스크 5.7 의 자리로 적어 두었다.

#### 두 정본이 갈리는 자리 하나 — 고르지 않고 적었다

엔진은 마감 시한 초과를 `abnormal=true` 로 돌려준다(`:897`). 설계 고장표
(`design.md:198`)는 deadline 을 "보통 오류 — 세고 다시 시도" 줄에 둔다.
**두 정본이 갈린다.** 이 로트는 엔진을 따랐고 이유는 둘이다: (1) 생산 임계값이
1 이라 오늘은 두 해석의 결과가 같다, (2) 갈리는 방향(임계값 > 1)에서 엔진 쪽이
더 보수적이다 — 설계대로면 마감 시한을 넘긴 레인이 한 번 더 진입을 시도한다.
`TestTheEngineWatchdogStillDecidesWhatThisLaneCopied` 가 엔진의 결론을 문자열로
못 박으므로, 엔진이 마음을 바꾸면 사람이 그 자리에서 보게 된다.
**임계값을 1 보다 크게 올리기 전에 사람이 이 갈림을 정해야 한다.**

#### 골든이 요구하는데 엔진에 없는 것 하나

`queue.overflow` 는 "typed refusal **and bounded drop counter**" 다. 엔진은
`StrategyTriggerFull` 이라는 typed refusal 은 갖고 있지만 **버린 수를 세는 곳이
없다**(`enqueueStrategyPoll` `:731`–`:747`, `Trigger` `:625`–`:641` 어디에도).
`Lane.Dropped()` 가 그 빈칸이다.

#### 왜 큐를 여기서 다시 만들지 않았나

골든 `queue` 절의 `dedup_key_fields`·`same_key`·`ordering` 은 이미
`strategycoordinator` 가 구현한다(`Key`, `laneSlot`, `Submit`, `Drops`).
레인의 투입 칸은 **봉투 큐가 아니라 방아쇠 칸**이다 — 엔진의
`worker.queue chan struct{}` 와 같은 것이고, 깊이만 정책이 정한다. 같은 판정을
두 곳에 두면 운영자가 보는 진단이 갈린다는 것이 이 패키지가 이미 적어 둔 규칙이다.

#### 반증 — 27 개, 27 개 잡힘. 두 개는 첫 형태가 틀렸다

- **M18 이 진짜 구멍이었다.** 감시견이 마감 시한(30초) 대신 카덴스(5초)만큼 자게
  바꿔도 초록이었다. 가짜 시계를 30초 밀면 5초짜리 잠도 함께 깨기 때문이다.
  결함이 사는 차원은 "깨어나는가"가 아니라 "**언제** 깨어나는가"이고, 그 축은
  마감 시한보다 **짧게** 밀어야만 갈린다. `TestTheWatchdogSleepsForTheDeadlineAnd
  NotSomeOtherPolicyValue` 를 커밋된 시험으로 심었고, 그 시험은 두 값이 같아지면
  스스로 `t.Fatalf` 로 무의미해졌다고 말한다.
- **M26 은 대상에 안 닿았다.** "여덟 레인을 한 벌로 공유" 변이를 처음엔 없는
  함수 이름으로 썼고 빌드가 깨졌다. 빌드 실패는 잡은 것이 아니다. 컴파일되는
  패키지 수준 캐시로 다시 쓰니 기존 신선도 시험이 잡는다.

구조 단언을 **썼다** — 5.1/5.3.1 로트에서는 안 썼고 그 이유도 적었는데, 여기서는
조건이 다르다. 감시견 세 갈래의 **순서**는 행동으로 관측되지 않는다: 채널 셋이
동시에 준비되면 Go 는 무작위로 고르므로 순서를 바꿔도 반드시 빨개지는 단일 시험이
없다. 그래서 엔진 쪽과 이 패키지 쪽 select 를 둘 다 AST 로 못 박았고, 갈래 재배열
변이(M16)가 그 시험에만 걸린다.

#### 골든의 세 이름이 이제 행동을 요구한다

`policy_test.go` 의 짝 시험에서 `cadence`·`bounded_queue`·`monotonic_deadline`
셋은 예전에 `policy.Cadence() > 0` 처럼 **들고 있는 수**를 물었다. 아무도 그 수를
강제하지 않아도 초록인 검사다. 이제 셋 다 레인을 실제로 돌려서 그 수가 행동을
바꾸는지 묻는다(주기 전 거절, 깊이 초과 거절+계수, 마감 시한 초과 폐기).

#### 폐포 허용 목록이 넓어졌다

`internal/clock` 과 stdlib 넷(`context`·`errors`·`fmt`·`sync`)이 늘었다.
`internal/clock` 은 자체 폐포가 `context`·`time`·`time/tzdata` 뿐이고 저장소의
단일 주입 시간 출처다. 레인이 시계를 **드는** 이유는 감시견이 시간을 읽는 것이
아니라 **자야** 하기 때문이고, 그래서 `Fail(now, …)` 의 `now` 인자를 없앴다 —
출처가 둘이면 같은 사이클의 backoff 기한과 마감 시한이 다른 시간축에 놓인다.
`fmt` 가 전이로 `os` 를 끌고 오는 것은 사실이며, 이 목록이 막는 것은 **직접**
들여오는 능력이고 모듈 안의 능력 패키지는 `-deps`/`-deps-test` 걸음이 따로 본다 —
그 사실을 허용 목록 주석에 적었다.

#### 닫지 않은 것

- latch 는 여전히 프로세스 메모리다. 5.3.3.
- 생산 호출자는 여전히 0 이다. `ProductionLanes(clk)` 를 부르는 생산 코드가 없다.
- 레인의 투입 칸은 봉투를 담지 않는다. 담게 하려면 조정자의 열쇠·접기 판정을
  옮겨 와야 하고, 그것은 5.2 가 정할 일이다.
- 엔진 스위트의 빈칸 일곱(위)은 이 로트가 메우지 않았다. 5.7.

### 여덟과 둘을 함께 세워 본다 (2026-09-02, 5.7)

`internal/strategyworker/rehearsal_test.go` 가 여덟 레인과 두 조정자를 한 무대에
세우고 여덟 goroutine 이 한 신호에 함께 출발한다. `design.md:255` 가 교체 전에
요구한 리허설이고, 교체 자체는 아니다 — 여기 서는 레인은 시험이 켠 사본이다.

#### 이 저장소는 `-race` 를 한 번도 돌린 적이 없었다

`Makefile` 과 `.github/workflows/ci.yml` 전체에 그 낱말이 없었다. "race 시험을
더하라"는 태스크를, 어떤 게이트도 검출기로 돌리지 않는 시험으로 만족시킬 수는
없다 — a118 이 태그 테스트에서 배운 것과 같은 모양이다. 그래서 이 로트는 시험만
더하지 않고 **도는 곳**을 만들었다: `make test-race`, `make gate` 9/11 단계,
CI job 한 단계, 그리고 그 셋 중 하나라도 빠지면 실패하는
`tools/sdd/test_race_detector_actually_runs.py`.

**측정으로 고른 범위.** `go test -race ./...` 는 10분에 끊겼다(완주 못 함).
이 런타임이 사는 일곱 패키지는 11.9초다. 그래서 일곱만 돈다.
**남은 34개는 여전히 검출기 밖이다** — 생산 코드에 goroutine·채널·`sync.`·
`atomic.` 이 있는 패키지가 41개이고 그중 34개가 안 돈다(`internal/journal`,
`internal/app/engine` 포함). 타깃 주석에 그 수와 이유를 적었다.

#### 검출기가 실제로 산 것 — 반증으로 쟀다

| 변이 | `-race` 없이 | `-race` 로 |
|---|---|---|
| N2 첫 실패 이유를 패키지 수준 슬롯에 **쓴다**(각 레인 필드에도 복사) | **초록** | 빨강 — `DATA RACE` + 격리 시험 |
| N2b 읽는 쪽이 그 공유 슬롯을 **본다** | 빨강 | 빨강 |

즉 결정적 교차는 시험이 혼자 잡고, 경합성 교차는 검출기가 있어야 잡는다.
이것이 이 저장소가 이미 적어 둔 함정(a099 R1 은 핸들 하나를 공유해 세 라운드
동안 순차만 증명했다)의 반대편 증거다.

나머지 반증 6개는 전부 잡혔다: 감시견 goroutine 미취소(N1), 세지 않는 드롭(N3),
도착 순서로 정렬하는 조정자(N4), 실패로 세는 취소(N5), 남의 시장 범위를 받는
조정자(N6), 세지 않는 폐기(N7). 게이트 배선 반증 5개도 전부 잡혔다.

#### 5.7 이 5.3.2 의 구멍을 찾았다 — 여기서 고쳤다

`Step` 이 `Cycle` 만 돌려주던 판본에서는 **보통 오류가 레인에 들어갈 길이
없었다.** 실패는 panic 과 마감 시한뿐이고 둘 다 비정상이라 임계값을 기다리지
않는다. 그러면 설계 고장표(`design.md:198`)의 "deadline/ordinary error — 세고
다시 시도" 줄에 닿는 입력이 존재하지 않고, `FailureThreshold` 는 어떤 생산
드라이버에게도 죽은 값이 된다. 엔진 영수증도 이쪽이다 —
`StrategyCycle = func(context.Context) error`. `Step` 은 이제 `(Cycle, error)` 다.

리허설을 안 만들었다면 이 구멍은 5.1.2 가 실제 사이클을 꽂는 날에 발견됐을
것이고, 그날은 이것보다 나쁜 날이다.

#### seam 을 spy 로 막으라는 지시는 이 패키지에서 실행할 수 없다

첫 판본은 `strategyhandoff` 를 들여와 진짜 handoff 를 spy 로 감쌌고,
`strategyhandoff/dependency_closure_test.go` 의 importer census 가 그것을 잡았다 —
그 seam 을 들여올 수 있는 패키지는 엔진 하나다(들여오는 것이 `Admit` 을 부를 수
있는 유일한 길이라서). 그래서 리허설은 seam 에 **닿을 수조차 없고**, 그것이
"아무것도 건너가지 않았다"의 더 강한 형태다. 진짜 spy 는 엔진 쪽이고 5.1.2/5.2 다.
이 사실을 산문이 아니라 `dispatchSpy` 주석에 적어 두었다.

#### 재기동이 잠금을 잊는다는 것을 재어서 못 박았다

`TestALaneBuiltWithoutADurableRecordIsBornUnlatched`(5.7 당시 이름 `TestRestartForgetsALatchedLaneWhichIsExactlyWhatTask533MustFix`)는 두 단언을
함께 한다: 새 레인 여덟이 건강하고, **원래 레인은 여전히 잠겨 있다.** 뒤엣것이
없으면 "잠근 적이 없다"와 구별되지 않는다. 이것이 5.3.3 의 자리다.

#### 깜빡임 하나를 고쳤다

`TestAnAbandonedCycleKeepsItsOwnGoroutineUntilTheStepReturns` 가 스위트 전체를
돌 때 한 번 실패했다(`4 → 4`). 기준선을 `runtime.NumGoroutine()` 로 한 번만
읽었는데, 앞 시험이 끝내는 중인 goroutine 이 섞여 기준선이 부풀어 있었다.
기준선을 "값이 연속 다섯 번 같을 때까지" 기다려 읽도록 바꿨고, 패키지를 다섯 번
연속 돌려 확인했다. 깜빡이는 시험은 다음 사람에게 무시당하므로 남겨 둘 수 없다.

#### 닫지 않은 것

- 이 리허설은 **이 타입들**이 동시에 돌 때 서로를 안 밟는다는 것만 증명한다.
  오늘의 엔진은 증명하지 않는다 — 엔진 쪽 빈칸 일곱은 여전히 두 FLM 번들의
  branch-test-map 에 좌표로만 있다.
- 34개 패키지가 경합 검출기 밖이다.
- 생산 드라이버가 없으므로 "동시 worker 여덟"은 시험이 만든 goroutine 여덟이다.

## 2026-09-03 L5 태스크 5.6.1 — 무엇이 어디까지 번지는가를 처음으로 쟀다 (구현 기록)

한 줄: **전략 고장이 엔진을 세우는 경로 넷은 이 저장소에서 한 번도 실행된 적이
없었고, 그 넷이 실행되면 손절을 놓는 loop 도 함께 죽는다.** 5.6 은 그것을
재는 로트이고, 생산 코드는 한 줄도 바꾸지 않았다.

### 왜 5.6 이 "보존"으로 끝날 수 없었나

태스크 문장은 "중앙 무결성 처리를 **보존**한다"이다. 보존은 대상이 있어야 하고,
대상은 5.1.2/5.2 의 교체다. 교체 전에는 "보존했다"를 말할 수 없다. 그래서
5.1/5.3 이 그랬듯 둘로 갈랐다 — 5.6.1 은 **오늘 서 있는 런타임을 재는 계기**를
만들고, 5.6.2 는 여덟 lane 위에서 같은 세 절을 다시 증명한다. 원문의 모든 절이
둘 중 하나의 소유다.

계기를 먼저 만든 이유는 측정에서 나왔다. 이 change 의 두 FLM 번들이 남긴
`count=0` 블록 일곱은 무작위 잔여가 아니다. **그중 넷은 전략 고장이 `Run` 의
반환에 닿는 유일한 경로다.** supervised loop 가 반환하면 Runtime 은 나머지 loop 를
전부 취소한다(runtime.go 의 "부분 생존 금지") — fill detection·reconcile·
exit observation 포함. 즉 그 넷은 "진입 쪽 고장이 손절을 함께 끄는가"를 정하는
자리이고, 아무도 그것을 돌려 본 적이 없었다.

### 잰 결과 (일곱 전부 `count=1`)

| 블록 | 무엇 | 채운 시험 |
|---|---|---|
| `787-790` | 만료 잠금 자체가 실패 → 중앙 | `TestTheFourEscalations…`/"권한 만료의 잠금도…" |
| `795-796` | 만료 뒤 재시작 대기 실패(비취소) → 중앙 | 같은 시험/"만료 뒤의 재시작 대기도…" |
| `821-824` | 보통 오류의 잠금 자체가 실패 → 중앙 | 같은 시험/"관측 시각이 없으면…"·"latch revision 이 소진되면…" |
| `829-830` | 잠근 뒤 재시작 대기 실패(비취소) → 중앙 | 같은 시험/"재시작 기한이 계약 밖이면…" |
| `800-801` | 잠긴 시장이 큐에 남은 요청을 버린다 | `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue` |
| `813-814` | 권한 갱신 전용 worker 가 오류를 삼킨다 | `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError` 외 1 |
| `894-896` | 마감 시한이 울렸는데 상위가 이미 취소됨 | `TestTheWatchdogRechecksCancellationInsteadOfTrustingItsOwnTimer` |

**넷의 공통점이 답이다.** 셋 다 평가가 실패한 것이 아니라 **감독자 자신의 장부가
깨진 것**이다: 관측 시각이 없다(`latchMarket` B4, `928:2`), latch revision 이
소진됐다(B5, `932:2`), 재시작 기한이 30s 계약 밖이다(`waitMarketRestart` B3,
`845:2`). 보통 오류·panic·마감 시한은 그 목록에 **없다.** 그것이 "lane/market
고장은 국소로 남는다"의 실질이고, 이제 산문이 아니라 다섯 칸의 값이다.

### 오늘 생산이 실제로 도는 구성은 `813-814` 다

`NewRefreshingPairedStrategyEntrySupervisor` 가 만드는 두 worker 는
`Effective=false, RefreshesAuthority=true` 이고, `cmd/tossctl/engine.go:647` 이
Runtime 에 넣는 것은 그 감독자다. 그래서 `refreshOnly` 가 참이고, 사이클 오류는
잠금 없이 삼켜져 다음 poll 이 다시 시도한다. 즉 **오늘 저널·Gateway 고장은 시장을
잠그지도, 엔진을 세우지도 않는다 — 그냥 그 사이클이 주문을 내지 않는다.**

### 갈라진 두 권위 — 사람이 정할 것

`runMarket` 의 판정 순서는 `refreshOnly`(813:4)가 `isCentralStrategyIntegrity`
(816:4)보다 **앞**이다. 그래서 오늘 생산이 도는 유일한 구성에서는 중앙 무결성
오류조차 삼켜진다.

- `design.md:198` 의 고장표: "journal/Gateway/fence/owner integrity fault → 모든
  신규 entry fail-closed".
- 같은 절의 "lane context 와 safety context 를 분리한다" + spec 의 "lane worker 가
  safety loop 를 취소해서는 안 된다 (MUST NOT)".

순서를 뒤집으면 뒤엣것이 깨진다. 그리고 "fail-closed" 의 수단으로 **프로세스 정지**
를 고르는 것은 보수적인 선택이 아니다 — 엔진이 서면 손절을 놓는 주체가 사라진다.
이 저장소에는 진입만 닫는 수단이 이미 있다(`execgw.EntryGate.Block`, reconcile·
alert·filldetect·flatten 이 전부 그것을 쓴다). 그러므로 이 로트는 **순서를 바꾸지
않고 값으로 고정**했다(`TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo`,
변이 M10 이 잡힌다). 바꾸려면 사람이 정한다.

### 한 축 한 시험은 여기서도 끝나지 않았다

`a112_fault_classification_test.go` 는 실제 저널·실제 Guardian·spy Gateway 로 세
고장을 심고 매번 (a) Gateway place 호출 0, (b) 반환 오류가 중앙으로 분류되지
않음을 함께 본다. Gateway 오류를 중앙으로 승격시키는 변이 둘은 잡혔다.
**저널 리스 발급을 승격시키는 변이(M13)는 살아남았다** — 닫힌 저널이 그 줄에
닿기 전에 소유권 fence 획득에서 먼저 실패하기 때문이다(측정: `sql: database is
closed`). 5.5·5.1.1 과 같은 결론이고 답도 같다: **세는 범위를 올린다.**

`a112_central_integrity_census_test.go` 는 패키지의 **모든 비시험 파일**을 파싱해
sentinel 이 나타나는 자리 여덟을 전부 열거하고 얼린다(그중 만드는 자리는 셋,
전부 `Run` 안). 그리고 `StrategyCentralIntegrityFailure` 의 **생산 호출자가 0**
임을 확인한다. M13 은 이 시험이 잡는다. 어느 것이 "만들기"이고 어느 것이 "읽기"
인지 시험이 판정하지 않는 이유는, 그 판정 자체가 우회 지점이 되기 때문이다 —
열거는 판정하지 않으므로 우회할 것이 없고, 역할은 얼린 목록 옆 주석이 진다.

### 어디에도 적혀 있지 않던 균형 하나

`latchMarket` 의 fault 건네주기 `default` 팔(`964-965`)은 `count=0` 이고
**도달 불가**다: 스트림 용량 2, 잠긴 시장은 `evaluationState`(857:2)가 거부하므로
시장당 잠금 한 번, 시장은 정확히 둘. 2 = 2.

이 등식은 어디에도 없었다. 그리고 5.1.2 가 둘을 여덟으로 바꾼다. 그러면 세 번째
lane 의 잠금이 fault 를 못 건네고, 그 팔이 잠금을 오류로 되돌리고, `runMarket` 이
중앙으로 올리고, **엔진과 safety loop 가 함께 선다** — "lane 고장은 국소로 남는다"
를 지키려던 코드에서. `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch`
가 그 등식을 값으로 못 박고, 용량을 1 로 만드는 변이(M7)가 잡힌다.

### `894-896` 은 "안 돈 것"이 아니라 "관측할 수 없던 것"이었다

실제 취소에서는 `ctx.Done()` 갈래와 마감 시한 갈래가 둘 다 준비되고, Go 는 무작위로
고르며, **두 갈래의 반환값이 같다.** `Done()` 이 절대 준비되지 않으면서 `Err()` 는
취소를 보고하는 context 를 넣으면 마감 시한 갈래만 열려 결정적으로 돈다. 그때
이 갈래가 하는 일이 드러난다: 자기 타이머를 믿지 않고 취소를 다시 확인해 **취소를
`abnormal` 로 승격시키지 않는다.** 생산 임계값이 1 이므로 그 승격은 곧 "정상
종료가 시장을 잠근다"이다. 대조군과 함께 두 칸을 재고, 재확인을 지우는 변이(M6)가
잡힌다.

### 반증

`strategy_entry_supervisor.go` 에 10, `strategy_dispatch_cycle.go` 에 3.
해시 백업에서 복원했고(커밋 전 로트라 `git checkout` 금지), 복원 뒤 SHA-256 을
대조했다.

| 변이 | 결과 |
|---|---|
| M1 관측 시각 가드 삭제 | CAUGHT |
| M2 latch revision 소진 가드 삭제 | CAUGHT |
| M3 재시작 기한 상한 완화 | CAUGHT |
| M4 `refreshOnly` 판정 반전 | CAUGHT |
| M5 잠긴 시장 건너뛰기 무력화 | CAUGHT |
| M6 watchdog 의 취소 재확인 삭제 | CAUGHT |
| M7 fault 스트림 용량 2 → 1 | CAUGHT (`TestTheFaultStreamHolds…`, `TestEveryWorkerCanHandOff…`) |
| M8 중앙 무결성 판정 무력화 | CAUGHT (기존 `TestCentralIntegrityFailureEscapes…`) |
| M9 중앙 신호 전달 삭제 | CAUGHT |
| M10 중앙 판정을 `refreshOnly` 앞으로 | CAUGHT (`TestARefreshOnlyWorkerSwallows…`) |
| M11 진입 게이트 오류를 중앙으로 승격 | CAUGHT |
| M12 보호 관측 오류를 중앙으로 승격 | CAUGHT |
| M13 저널 리스 발급 오류를 중앙으로 승격 | 행동 시험에서 **SURVIVED** → census 가 CAUGHT |

### 닫지 않은 것

- **5.6.2 가 열려 있다.** 여기서 잰 것은 전부 "시장 둘, 시장당 소비자 goroutine
  하나"의 성질이다. 교체가 그 셋을 다 바꾼다.
- **순서 결정(refreshOnly 가 중앙보다 앞)은 사람의 것이다.** 이 로트는 고정만 했다.
- 도달 불가로 확인한 블록 셋(`917-919`, `964-965`, `waitMarketRestart` 의
  `837-839`·`841-843`)은 메우지 않았다. 각 번들의 branch-test-map 에 왜 열리지
  않는지를 적었다 — "시험을 안 썼다"와 구별하기 위해서다.
- 이 로트는 생산 코드를 바꾸지 않았다. `git diff -- '*.go'` 가 비어 있다.

## 2026-09-03 L5 태스크 5.1.2.1 — 여덟이 생산 런타임 안에 섰다 (구현 기록)

### 무엇을 옮겼나

전략군 하나를 평가하는 일이 더 이상 `*Context` 클로저 안에 있지 않다.

오늘 생산이 "worker" 라고 부르는 것은 `StrategyMarketWorker` 이고 그 `Cycle` 은
`runProductionStrategyMarketCycle` — `*Context` 를 담은 클로저다. AST 산출물이
그 안에서 `c.Journal.CurrentPositionCampaignCAS`(`462:15`)와
`fresh.dispatch.dispatch`(`469:12`)를 열거한다. 전략군 평가가 원장을 쓰고 주문을
낼 수 있는 자리에서 일어난다는 뜻이다.

이제 그 함수는 새로 고침 뒤·handoff 앞에서 이 시장의 **네 레인**을 한 번씩
돌린다(`441:2`). 레인 안에서 도는 것은 `strategyFamilyLaneStep` 하나이고,
그 함수는 패키지 수준이며 인자가 `*strategyworker.Lane` 뿐이다. 무엇을 만질 수
있는지가 주석이 아니라 **인자의 타입**으로 정해지고, 그 타입이 사는 패키지는
`-deps`/`-deps-test` 를 훑어 broker mutator·writable journal·Guardian issuer 가
자기 폐포에 없다는 것을 시험으로 지킨다.

`CurrentPositionCampaignCAS` 와 `dispatch` 는 **그대로 그 함수에 있다.** 그것이
맞다 — 스펙은 시장 하나에 변경 권한을 하나만 두라고 요구하고, 이 함수가 그
하나다. 옮긴 것은 *전략군의* 사이클이지 시장의 변경 권한이 아니다.

### 왜 태스크를 갈랐나 — 고른 것이 아니라 막힌 것이다

원문 5.1.2 는 "두 시장 worker 계약을 여덟 FamilyWorker 로 **교체**한다"였다.
교체의 나머지 절반(여덟이 진입의 **관문**이 되는 것)은 이 로트에서 할 수 없다.

두 권위가 그것을 막는다. 둘 다 이 로트보다 먼저 있었다.

1. 동결 골든 `analysis/goldens/four-family-runtime-v1.json` 이 여덟 서술자를
   전부 `desired: OFF, effective: OFF, runtime: UNOBSERVED` 로 얼렸다.
2. 스펙: "Legacy 3-family approval 은 4-family activation 으로 자동 승격되어서는
   안 되며 (MUST NOT)".

잠든 `FamilyWorker.Run` 은 아무것도 보기 전에 DORMANT 를 돌려준다. 그러므로
오늘 여덟을 관문으로 세우면 생산 진입이 0 이 된다 — 그리고 이 저장소의 불변식은
"토글 OFF 는 upstream 동작과 동일해야 한다"이다. 승격은 서명된 활성화
매니페스트가 하는 일이고 그것은 8 절과 5.1.2.2 다.

이 결론을 우회할 길도 확인했다: 세 legacy 가족만 기존 승인에서 effective 를
유도하는 것. `strategyworker` 는 켜진 worker 를 패키지 밖에서 만들 수 없게
`newWorker` 를 내보내지 않는다(그것이 "server-owned" 를 주석이 아니게 하는
유일한 장치다). 그 문을 여는 것이 곧 매니페스트 작업이므로, 5.1.2 안에서 조용히
여는 것은 고위험 경로의 범위 확대다.

### 증명의 모양 — 왜 census 이고 행동 시험이 아닌가

여덟이 전부 DORMANT 이므로 **어떤 실행도** 레인이 켜졌다면 무엇을 만질 수
있었는지 보여 주지 못한다. 5.5·5.1.1·5.6.1 이 세 번 같은 곳에 닿았다: 축마다
시험 하나는 끝나지 않는다. 답도 세 번 같다 — 세는 **범위**를 올린다.

| 시험 | 무엇을 세나 | 범위 |
|---|---|---|
| `TestOnlyThePackageLevelStepEverRunsInsideALane` | `Lane.RunBounded` 에 넘기는 값 전부 | 엔진 패키지의 모든 비시험 파일 |
| `TestTheMarketCycleRunsItsLanesAndTheRefreshDoesNot` | `evaluate` 를 부르는 자리 전부 | 같음 |
| `TestTheFamilyLaneStepCarriesNothingButItsLane` | 그 값의 수신자·인자·이름 | 그 함수 하나 |
| `TestTheLaneScopeReadsTheApprovedSymbolAndNotTheProposalsOwnLineage` | 범위 네 필드가 어디서 오는지 | 그 함수 하나 |

행동 쪽은 여섯: 여덟/넷 기수, 프로세스 수명, 잠듦(빈 입력), **자기 것이라고
인정한 제안에도 잠듦**, 버린 투입이 사이클을 열지 않음, 이름만 바꾼 계보를
아무도 안 받음.

"자기 것이라고 인정한 제안" 이 중요한 이유: 빈 입력으로 DORMANT 를 받는 것은
켜진 레인도 마찬가지다. 첫 판본은 dispatch 주기 fixture 를 썼는데 그 경로의
경로 권한에 가족 점수 행이 없어 **여덟 중 아무도** 제안을 자기 것으로 알아보지
못했다. 그러면 "봉투 0 건" 은 배선이 조용해서가 아니라 배선이 닿지 않아서다.
`TestExactlyOneLaneOwnsEachSealedProposal` 이 그 전제를 따로 지킨다.

### 무시할 수 있는 답을 또 하나 없앴다

첫 판본의 `evaluate` 는 관측과 봉투를 함께 돌려줬다. 호출자가 둘 중 하나를
버릴 수 있고, 버린 것은 아무도 못 본다. 같은 함수가 5.5-fix2 에서 `Single()` 의
boolean 으로 이미 그 실패를 겪었다. 지금 `evaluate` 는 **아무것도 돌려주지
않는다** — 결과는 런타임 안에 남고 `observations()` 가 읽는다.

### 5.6.1 의 census 를 먼저 고쳐야 했다

`a112_central_integrity_census_test.go` 는 sentinel 이 나타나는 자리를 **파일
절대 줄**로 얼렸다. 이 로트가 `runProductionStrategyMarketCycle` 에 열여섯 줄을
넣자 `Run` 의 세 자리가 669/674/713 → 685/690/729 로 밀려 시험이 빨개졌다.
개수도 감싼 함수도 그대로인데.

그 상태에서 통과시키는 방법은 기대 목록을 고치는 것뿐인데, **그 편집은 이
census 가 잡으려는 편집과 구분되지 않는다.** 그래서 좌표를 감싼 선언의 시작
으로부터의 거리로 바꿨다(`Run+2`, `Run+7`, `Run+46`). 남의 편집에는 흔들리지
않고, 그 선언 **안에서** 자리가 옮기거나 개수가 바뀌는 것은 그대로 잡는다.
아래 M12(새 mint)와 C1(무관한 편집)이 두 방향을 각각 확인한다.

### 반증 — 12 개 뮤테이션 + 대조군 1

해시 백업에서 복원하고 복원 후 SHA-256 을 대조했다(`git checkout` 은 커밋 안 된
로트를 지우므로 쓰지 않았다).

| # | 뮤테이션 | 결과 | 잡은 시험 |
|---|---|---|---|
| M1 | 단계 함수가 `*Context` 인자를 하나 더 받는다 | CAUGHT | step shape · RunBounded census |
| M2 | 시장 주기가 레인을 안 돌린다 | CAUGHT | evaluate call-site census |
| M3 | 새로 고침 잠금 **안**에 두 번째 레인 단계를 연다 | CAUGHT | 같음 |
| M4 | 레인이 넷만 선다 | CAUGHT | 기수 |
| M5 | 레인이 켜진 채로 태어난다 | CAUGHT | 잠듦 |
| M6 | 새로 고침마다 레인을 다시 만든다 | CAUGHT | 프로세스 수명 |
| M7 | 모든 레인이 모든 제안을 자기 것이라 한다 | CAUGHT | 소유자 수 · 위조 계보 |
| M8 | 소유 판정이 봉인을 안 본다 | CAUGHT | 위조 계보 |
| M9 | 범위의 종목을 제안의 계보에서 읽는다 | CAUGHT | 범위 출처 |
| M10 | 칸이 찬 투입도 사이클을 연다 | CAUGHT (2차) | 버린 투입 |
| M11 | 낼 것 없는 레인을 건너뛴다 | CAUGHT | 잠듦(넷이 돌았는가) |
| M12 | 엔진에 중앙 무결성 sentinel 을 만드는 자리를 하나 연다 | CAUGHT | census |
| C1 | (대조군) `Run` 위쪽에 무관한 주석 한 줄 | SURVIVED (의도) | census 가 남의 편집에 안 흔들린다 |

**M10 은 1 차에서 살아남았다.** 첫 판본의 큐 시험은 버린 **수**만 봤고, 그 수는
`Offer` 가 올린다 — 그래서 "칸이 찼으면 사이클을 열지 않는다"는 갈래를 지워도
초록이었다. 그때 실제로 문을 닫고 있던 것은 그 갈래가 아니라 **카덴스**였다.
시험을 다시 썼다: 카덴스를 지나 보낸 뒤에 칸을 채우면, 갈래가 없을 때 버려진
투입 하나가 그대로 사이클을 연다. 이제 그 자리를 `Start` 로 본다.
(같은 뿌리의 교훈이 이 change 에 이미 있다 — 살아남은 뮤테이션은 "동등 변이"가
아니라 "우연이 지키는 안전"일 수 있다.)

### 이 로트가 주장하지 않는 것

- 여덟은 아직 진입의 관문이 **아니다**. 5.1.2.2.
- 레인이 보는 것은 조정을 **지난** 제안이다. 조정 앞으로 옮기는 것도 5.1.2.2.
- 레인 고장은 아직 감독자의 fault 스트림으로 가지 않는다. 그래서 5.6.1 이 찾은
  `cap(faults) == 2 == 시장 수` 균형은 이 로트에서 건드려지지 않았다. 그것을
  옮기는 순간 그 등식이 깨진다.
- 레인마다 자기 goroutine·cadence 를 갖는 것은 5.2 다. 지금은 시장 주기가
  네 레인을 차례로 돌린다.

## 2026-09-03 L5 태스크 5.2.1 — 원격 권한 수집이 공유 잠금 밖으로 나갔다 (구현 기록)

### 편집 전에 잰 것

| 사실 | 근거 |
|---|---|
| `refreshPairedStrategyEntryProductionAssembly` 가 `c.strategyRefreshMu` 를 들고 `NewPairedStrategyEntryProductionAssembly` 전체를 돈다 | 편집 전 `ast.json`: `c.strategyRefreshMu.Lock` `481:2`, `defer …Unlock` `482:2`, `c.NewPairedStrategyEntryProductionAssembly` `487:16` — 셋 다 한 본문 안 |
| 그 파도가 원격을 탄다 | `NewPaired…` 본문이 `newStrategyScheduleAuthorityLoader(…, c.official, …)`(→ `TypedMarketCalendar`)와 `newStrategyFXAuthorityLoader(…, c.official, …)`(→ `officialfx.ProductionAuthorityService`)를 만들고 각각 `collect` 를 부른다 |
| 경쟁하는 goroutine 은 정확히 둘 | CodeGraph: `refreshPaired…` 의 호출자는 `runProductionStrategyMarketCycle` 하나뿐이고, 그것은 시장마다 하나씩 돈다 |
| 1초 창이 파도의 **시작** 시각을 잰다 | 편집 전 `ast.json`: `:=` at `483:2`(now 읽기)가 `487:16`(파도)보다 앞이고 `=` at `491:2`(`strategyRefreshAt = now`)가 뒤 |

두 번째 사실의 귀결이 이 로트의 핵심이다. 파도가 1초보다 오래 걸리면 그것이 끝나는
순간 캐시는 이미 그만큼 묵었고, 잠금을 물려받은 두 번째 시장은 창 밖이라 자기 파도를
처음부터 다시 돈다. **원격이 느릴수록 합치기가 정확히 꺼진다.**

### RED (편집 전 소스에서 실제로 실행)

```
--- FAIL: TestTheRemoteAuthorityWaveNeverRunsUnderTheSharedAssemblyMutex
    Context.refreshPairedStrategyEntryProductionAssembly 가 공유 assembly mutex 를 들고 원격 권한 파도를 돌린다.
    그 함수의 호출 목록: [c.NewPairedStrategyEntryProductionAssembly c.strategyRefreshMu.Lock
                        c.strategyRefreshMu.Unlock clk.Now clk.Now().UTC errors.New now.Before now.Sub]
--- FAIL: TestExactlyOneFunctionRunsTheRemoteAuthorityWave
     got: Context.refreshPairedStrategyEntryProductionAssembly
    want: Context.collectStrategyRefreshWave
```

행동 절반은 RED 를 만들 수 없다. 이 패키지에서 파도를 느리게 만드는 방법이 없기
때문이다(원격은 실제 official client 뒤에 있고 주입된 시계로 늦출 수 없다). 그 절반의
이빨은 RED 가 아니라 아래 배터리가 지운다. 이것을 침묵으로 넘기지 않고 BTM 의 B3 칸에
"빨갛다가 아니라 없다"로 적었다.

### 반증: 16개 CAUGHT, 대조군 2개 SURVIVED(의도)

| 변이 | 심은 것 | 결과 |
|---|---|---|
| M1 | 진행 중인 파도 확인을 지운다(모두가 지도자) | CAUGHT |
| M2 | 발표가 파도를 자리에서 안 내린다 | CAUGHT |
| M3 | 실패한 파도도 캐시에 넣는다 | CAUGHT |
| M4 | 창을 `<` 에서 `<=` 로 | CAUGHT |
| M5 | 기다리는 쪽이 `ctx` 를 안 본다 | CAUGHT (거품 교착 감지) |
| M6 | 기다린 쪽에 오류를 안 준다 | CAUGHT |
| M7 | 지도자가 파도를 만들고 발표하지 않는다 | **1차 SURVIVED** → 아래 |
| M8 | 파도를 잠금 안으로 다시 넣는다 | CAUGHT (호출 목록 셈) |
| M9 | 패닉 발표 defer 를 지운다 | CAUGHT (거품 교착 감지) |
| M10 | 발표가 파도에 값을 안 싣는다 | CAUGHT |
| M10b | 발표가 `done` 을 먼저 닫고 값을 나중에 쓴다 | **`-race` 없이 SURVIVED / `-race` 로 CAUGHT** |
| M10c | 발표가 `done` 을 안 닫는다 | CAUGHT (거품 교착 감지) |
| M11 | 합류한 쪽도 지도자라고 답한다 | CAUGHT |
| M12 | 기다리는 대신 빈 값을 돌려준다 | CAUGHT |
| M13 | 실패한 파도를 비우지 않는다 | CAUGHT |
| M15 | 잠금 안에서 헬퍼 한 다리를 거쳐 원격을 부른다 | CAUGHT (얼린 호출 목록) |
| C1 | 무관한 주석 한 줄 | SURVIVED (의도) |
| C2 | 수신자 철자를 `c` 에서 `owner` 로 바꿔 파도를 부른다 | SURVIVED (의도 — 아래) |

원복은 해시 백업으로 하고 SHA-256 으로 확인했다(`git checkout` 은 커밋 전 GREEN 까지
지운다 — 이 change 가 5.3.1 에서 이미 배운 것).

**M7 이 살아남은 이유가 이 로트에서 가장 중요한 발견이다.** 위 시험이 전부 *시험이
지도자*였다 — `joinStrategyRefreshWave` 로 지도자 자리를 잡고, 생산 함수는 언제나
기다리는 쪽으로만 들어갔다. 그래서 "지도자가 파도를 발표한다"를 재는 실행이 하나도
없었다. 이름을 얼리는 셈도 못 잡는다: M7 은 수집을 원래 함수에 그대로 두므로 자리 셈이
초록이다. M7 의 실제 결과는 조용하지 않다 — 파도가 자리에 남고 아무도 닫지 않으므로
**그 뒤의 모든 주기가** 죽은 파도에 합류해 자기 ctx 가 죽을 때까지 기다린다. 두 시장의
진입이 함께 멈춘다. `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` 이 생산 경로를
지도자로 세워 그 구멍을 막는다.

**M15 와 C2 는 이 셈의 두 우회로를 각각 막았다는 확인이다.** M15 는 잠금 안에서
헬퍼 하나를 부르고 그 헬퍼가 원격을 타는 모양이다 — 잠금을 만지는 함수 **목록**만
얼렸다면 초록이었을 것이고, 그 목록이 5.5 의 적대 리뷰가 실제로 뚫은 자리다. 얼린
**호출 목록**이 그것을 잡는다. C2 는 반대 방향이다: 첫 판본의 셈은 호출 철자를
`c.NewPairedStrategyEntryProductionAssembly` 로 통째로 비교했고, 수신자 이름을 하나
바꾸는 편집이면 셈이 눈이 멀었다. 지금은 메서드 이름만 보므로 C2 는 여전히 보인다
(그래서 SURVIVED 가 맞는 답이다). 맨 이름을 빼는 이유도 잰 것이다: 이 패키지에는 같은
이름의 패키지 수준 값 전용 생성자가 따로 있고 그것은 원격을 타지 않는다.

**M10b 가 검출기의 값을 잰 두 번째 사례다.** 5.7 의 N2 와 같은 모양이 이번엔
`internal/app/engine` 에서 나왔다: `-race` 없이 5회 반복해도 초록이고, `-race` 로는
`WARNING: DATA RACE` 로 즉시 빨갛다.

### 검출기 배선

| 잰 것 | 값 |
|---|---|
| `go test -race -tags tossos_testseams ./internal/app/engine` | **14분 46초** |
| 기존 `RACE_PACKAGES` 일곱 개 전부 | 11.9초 (5.7 측정) |
| 이 로트의 동시성 시험만 `-race` 로 | **5.7초** |
| `make test-race` 전체 (배선 뒤) | 13.9초 |

그래서 패키지를 통째로 넣지 않고 이름으로 골랐다. 이름 고르기의 실패 방식은 a118 이
이미 겪은 것 — 나중에 늘린 시험이 목록 밖으로 조용히 남는 것. `tools/sdd/
test_race_detector_actually_runs.py` 에 두 시험을 더해, `a112_refresh_singleflight_test.go`
의 `Test` 함수 전부가 목록에 있어야 하고 recipe 에 그 줄이 있어야 한다.

### 덤: 5.6.1 의 커버리지 블록 번호를 다시 쟀다

5.6.1 이 BTM 에 적어 둔 `block A-B` 29개가 5.1.2.1(+16)과 이 로트(+3) 뒤 19줄 밀린
채였고 아무도 옮기지 않았다. 산술로 옮기지 않고 프로파일
(`go test -count=1 -tags tossos_testseams -coverprofile ./internal/app/engine/`,
2026-09-03, 77.6% of statements)을 다시 떠서 대조했다: 28개가 정확히 +19 자리에 있었고
기록된 `count` 도 전부 일치했다. 남은 하나는 실제 블록이 `804-806` 인데 `785-786`
(= 804-805)로 적혀 있었다 — 5.6.1 이 적을 때 한 줄 어긋난 것이다. 옮겨 적지 않고 잰
값을 넣고, 그 사실을 해당 BTM 에 적었다.

### 이 로트가 주장하지 않는 것

- 시장 단위 단일 제안 준비 상태는 그대로다. 그것은 5.2.2 이고 5.1.2.2 와 같은 사람
  결정에 걸려 있다.
- 7.5 의 "no remote I/O under strategy refresh mutex" 성능·운용 시험은 여기 없다.
  이 로트는 구조 셈과 동시성 행동을 쟀고, 부하 아래 지연은 재지 않았다.
- `internal/app/engine` 의 나머지 동시성(감독자 루프·투영 저장소·dispatch 주기)은
  여전히 검출기 밖이다. 통째로 넣으면 게이트가 15분 늘어난다.

## 2026-09-03 L5 태스크 5.3.3 — Pre-Edit 선언 (High-risk: 원장 스키마)

`docs/WORKFLOW.md:408` 이 요구하는 선언이다. 이 로트는 **원장 스키마**를 건드리므로
High-risk 다.

```text
Pre-Edit Gate:
- change id / task id: a112-run-four-strategy-families-independently / L5 5.3.3
- 대상 심볼(패키지.함수):
    internal/journal            SchemaVersion(31→32), migrations(추가 한 줄),
                                새 파일 strategy_lane_latch.go(신규 함수)
    internal/strategyworker     ProductionLanes(기존, 이 change 가 만든 새 함수),
                                newLane(기존, 같음), 새 값 타입 LatchRestore
    internal/app/engine         newStrategyLaneRuntime/productionStrategyLanes/
                                evaluate(전부 5.1.2.1 이 만든 새 함수),
                                runProductionStrategyMarketCycle(frozen base 함수 — FLM 필수),
                                NewPairedStrategyEntryProductionAssembly(frozen base 함수 — FLM 필수)
  (선언 뒤 정정) 계획에는 `strategyScheduleAuthorityPair.Snapshot` 편집이 있었다.
  실제로는 그 공개 스냅숏을 건드리지 않고, 서명된 일정 **권위**를 assembly 의
  비공개 칸으로 그대로 들고 갔다 — 스칼라 사본을 만들면 활성화 세대가 권위에서
  한 다리 멀어지고, 그 한 다리가 복구 조건이 서명과 갈라지는 자리가 된다.
  대신 편집된 frozen-base 함수는 `NewPairedStrategyEntryProductionAssembly` 이고
  그 FLM 을 재생성했다. 그리고 계획에 없던 frozen-base 함수 하나가 더 편집됐다:
  `internal/journal/schema_test.go` 의 `TestSchemaTablesAndColumns` — 얼린 테이블
  열거표다. 새 테이블 둘이 생기면 그 시험이 빨개지는 것이 설계이고, 실제로
  빨개져서 두 줄을 사람 손으로 선언했다. 그 FLM 도 새로 만들었다.
- 기존 동작 파악 근거:
    · 마이그레이션 규칙 4개와 down-migration 부재는 internal/journal/schema.go:8-22 에
      적혀 있고, v25(strategy_dispatch_runtime_v25.sql)가 같은 영역의 선례다 —
      STRICT 테이블 · CHECK 제약 · 불변/단조 트리거 · DELETE 금지.
    · 레인 latch 의 현재 의미: internal/strategyworker/lane.go 의 latched/latchRevision/
      firstFailure 는 프로세스 메모리이고, Succeed() 는 계수기만 지우며 latch 는 못
      지운다. rehearsal_test.go 의
      TestALaneBuiltWithoutADurableRecordIsBornUnlatched(당시 이름
      TestRestartForgetsALatchedLaneWhichIsExactlyWhatTask533MustFix)가 그 구멍을
      이미 값으로 못 박아 두었다.
    · 복구 증거의 출처: scheduler.Activation 은 "opaque capability issued only after an
      exact manifest verification"(desired.go:236)이고 Generation() 은 ed25519 서명
      매니페스트의 manifest.Generation(production_activation.go:159)이다. 즉 사람이
      서명한 파일을 바꾸지 않으면 오르지 않는다.
    · 시장·가족 enum 은 골든 산문이 아니라 코드에서 읽었다:
      strategyrouter/types.go:9(KR/US), production.go:57-60(네 가족).
- upstream 상속 테스트 영향: no. 새 테이블 둘과 새 함수뿐이고, 기존 테이블·트리거·
  쿼리를 하나도 바꾸지 않는다. `SchemaVersion` 상승은 전방 전용이며 구버전 바이너리는
  `ErrSchemaTooNew` 로 **거절**한다(조용히 오해하지 않는다, schema.go:19-22).
- 실패 테스트 선행 작성: yes.
- 안전 불변식 §0 위반 여부 검토: 통과. 근거를 값으로 적는다 —
    · 이 로트는 주문·손절·익절·사이징 경로를 건드리지 않는다. latch 는 **신규 진입만**
      막고 exit/fill/reconcile/protection 은 5.6.1 이 잰 대로 별개 loop 다.
    · 오늘 여덟 레인은 전부 DORMANT 라 `lane.Run` 이 아무것도 보기 전에 DORMANT 를
      돌려준다. 사이클이 오류를 내는 유일한 길은 패닉과 마감 시한인데 dormant 사이클은
      즉시 끝난다. **그래서 오늘 생산에서 이 테이블에 쓰이는 행은 0이다.**
    · 되돌릴 수 없는 것 하나: 이 마이그레이션이 돈 DB 를 main 바이너리로 되돌리면
      엔진이 `ErrSchemaTooNew` 로 뜨지 않는다. a112 는 8.6 기준 배포 BLOCKED 이고
      병합 순서는 사람이 정한다. 이 로트는 배포하지 않는다.

**이 로트가 없애는 운영 수단 하나를 먼저 적는다(fail-closed 가 무엇을 거부하는가).**
오늘 잠긴 레인을 여는 방법은 **엔진 재시작**이다. 이 로트 뒤에는 재시작이 그것을
열지 않는다 — 열려면 generation 이 더 큰 **서명된 활성화 매니페스트**가 필요하다.
그것이 요구된 동작이지만(design.md 의 고장표 "recovery evidence 필요"), 재시작으로
고치던 운영자에게는 수단이 하나 사라지는 것이다. 그래서 복구 조건은 DB 트리거로
못 박고, 무엇이 필요한지를 오류 문구가 말하게 한다.

## 2026-09-03 L5 태스크 5.3.3 — 잠금이 프로세스보다 오래 산다 (구현 기록)

### 무엇을 어디에 뒀고, 왜 거기인가

| 조각 | 자리 | 이 자리인 근거(읽은 것) |
|---|---|---|
| durable 기록 | 원장, schema **v32**, append-only 두 테이블 | 저널 고장의 분류가 이미 있다 — 신규 진입은 막고 엔진은 안 세운다(5.6.1 의 `a112_fault_classification_test.go`). 별도 저장소는 그 분류를 새로 정해야 하고 트랜잭션·백업·불변 트리거를 다시 만들어야 한다 |
| writer | `internal/app/engine` | `internal/strategyworker` 는 자기 폐포에 쓸 수 있는 저널이 없음을 `-deps` 로 증명한 패키지다. `design.md:223` 이 "lane health/latch projection" 을 엔진에 배정 |
| 레인이 받는 것 | 값 `LatchRestore` | 값은 능력이 아니다. 폐포 증명이 그대로 남는다 |
| 복구 증거 | `scheduler.Activation.Generation()` | "opaque capability issued only after an exact manifest verification. Callers cannot forge it with a bool or public struct literal"(`desired.go:236`)이고 값은 ed25519 서명 매니페스트의 `manifest.Generation`(`production_activation.go:159`) |
| 복구 판정 | SQLite 트리거 | 코드에 두면 다른 호출자가 다른 판정을 쓴다. `strategy_lane_latch_recovery_needs_newer_activation` 이 "엄격히 더 큰 세대"를 강제한다 |
| 첫 원인 보존 | SQLite 트리거 + Go 의 열린-잠금 조회 | `strategy_lane_latch_first_cause_wins`. 레인의 메모리 규칙("두 번째 실패는 새 latch 판을 안 찍는다")이 프로세스 경계를 건너서도 사는 유일한 자리 |

### 복구는 잠금을 푸는 것이 아니다

푸는 길이 존재하면 그 길은 증거 없이도 불릴 수 있다. 그래서 `strategyworker` 에는
`latched` 를 **거짓으로 만드는 자리가 하나도 없고**, 그 사실을 패키지 전체를 훑는
셈이 지킨다 — 얼린 목록은 `latch = true`, `restoreLatch = true` 둘뿐이다. 엔진은
원장이 복구를 받아들인 **뒤에만** 그 레인을 기록 없이 새로 세운다(`rebornLane`).

### RED (편집 전 소스에서 실제로 실행)

```
--- FAIL: TestALatchedLaneComesBackLatchedAfterTheProcessRestarts
    재시작이 잠긴 레인을 열었다 — 잠근 이유는 그대로인데 잠금만 사라졌다
```

컴파일 오류가 아니라 행동이다. 이 시험은 편집 전의 API 만 쓴다.

### 가장 우회하기 쉬운 자리를 따로 얼렸다

복구 조건은 "더 큰 수"가 아니라 "더 큰 **서명된** 세대"다. 그 인자에 주기 계수기나
시각을 넣으면 트리거는 그대로 통과하고 잠금은 저절로 열린다 — 그리고 **두 수 다 그냥
커지므로 어떤 행동 시험도 차이를 못 본다.** 그래서 인자로 넘어가는 식 자체를 센다:
`runProductionStrategyMarketCycle: fresh.schedule.forMarket(market).restore.Activation.Generation()`
하나만 얼려 두었다.

### 오늘 생산에 쓰는 행은 0이고, 그것도 값이다

여덟 레인은 전부 DORMANT 라 `Run` 이 아무것도 보기 전에 돌아오고 오류가 없다 →
잠기지 않는다 → 행이 생기지 않는다. `TestTheProductionStepNeverLatchesSoTheLedgerStaysEmpty`
가 두 시장에 다섯 주기를 돌리고 원장이 비어 있는지 본다.

### 이 로트가 없애는 운영 수단

오늘 잠긴 레인을 여는 방법은 **엔진 재시작**이다. 이 뒤로는 재시작이 열지 않는다.
요구된 동작이지만(design.md 고장표의 "recovery evidence 필요"), 재시작으로 고치던
사람에게는 수단이 하나 사라지는 것이므로 거절 문구가 무엇이 필요한지 말한다
(`journal.ErrStrategyLaneLatchRecoveryEvidence`). 부수적으로 이것은 안전상 이득이기도
하다 — 엔진 정지는 손절을 놓는 주체를 없앤다.

### 반증: 14개 CAUGHT, 대조군 1개 SURVIVED(의도)

| 변이 | 심은 것 | 결과 |
|---|---|---|
| N1 | 기록을 레인에 붙이지 않는다 | CAUGHT |
| N2 | 붙지 않은 기록의 오류를 삼킨다 | CAUGHT |
| N2b | 붙지 않은 기록을 아예 안 알아본다 | CAUGHT |
| N3 | 잠긴 레인을 원장에 남기지 않는다 | CAUGHT |
| N4 | 남기기 실패를 조용히 넘긴다 | CAUGHT |
| N5 | 거절된 복구를 성공처럼 다룬다 | CAUGHT |
| N6 | 원장에 묻지 않고 레인을 다시 세운다 | CAUGHT |
| N7 | "더 큰 세대" 트리거를 지운다 | CAUGHT |
| N8 | 열린 잠금 조회를 지운다(첫 원인 보존이 트리거에만 남는다) | CAUGHT |
| N9 | `restoreLatch` 가 `latched = false` 를 쓴다 | CAUGHT |
| N10 | 되살릴 때 첫 원인 문구를 바꾼다 | CAUGHT |
| N11 | 복구 세대에 서명과 무관한 0 을 넘긴다 | CAUGHT |
| N12 | 레인을 세우면서 기록을 안 읽는다 | CAUGHT |
| C1 | 무관한 주석 한 줄 | SURVIVED (의도) |

원복은 해시 백업으로 하고 SHA-256 으로 확인했다.

**그리고 배터리가 실제 설계 결함 **둘**을 찾아냈다 — 그것이 이 로트에서 가장 값진 결과다.**

**결함 하나: 문 없는 fail-closed.** 첫 판본은 이 빌드에 없는 레인을 가리키는 durable
기록을 만나면 오류를 내고 그 기록을 **버렸다.** 버린 순간 그것을 닫을 방법이 사라진다 —
`RecoverStrategyLaneLatch` 는 원장 순번을 받는데 그 순번을 아무도 안 들고 있기 때문이다.
결과는 진입이 영원히 멈추고 빠져나갈 길이 없는 상태다. 기억에 "fail-closed 는 무엇을
거부하는지 말해야 한다"로 남은 모양의 사촌이다: 거부는 맞는데 **되돌릴 문이 없다.**
지금은 붙지 않은 기록도 들고 있다가 복구를 함께 시도하므로, 더 큰 서명 활성화 세대가
오면 닫힌다. 그 문이 실제로 열리는지는 시험이 값으로 잰다
(`TestADurableLatchThatNamesNoLaneInThisBuildStopsTheCycleLoudlyAndCanBeClosed`).

**결함 둘: 판정이 둘이었다.**
첫 판본은 복구 판정이 **둘**이었다. Go 쪽이 `activationGeneration > latch.ActivationGeneration`
인 잠금만 골라 원장에 물었고, 원장의 트리거가 다시 같은 판정을 했다. 반증이 그 대가를
값으로 보여 줬다:

- 부등호를 `>=` 로 바꾼 변이 → **SURVIVED.** 트리거가 대신 막았다.
- 트리거를 지운 변이 → 저널 시험만 빨갛고 **엔진 시험은 초록.** 부등호가 대신 막았다.

즉 **각자가 상대의 시험을 통과시킨다.** 둘 중 하나가 조용히 사라져도 아무 색도 변하지
않고, 남은 하나가 우연히 지키는 상태가 된다 — 기억에 "살아남은 뮤테이션 = 우연이
지키는 안전"으로 남은 모양 그대로다. 그래서 Go 쪽 부등호를 **지웠다.** 이제 이 시장의
열린 잠금 전부를 원장에 물어보고 답을 따르며, 판정은 기록이 사는 곳 하나에만 있다.
그 뒤 트리거를 지운 변이는 엔진 시험도 빨갛다(위 표의 N7).

### 이 로트가 주장하지 않는 것

- 여덟은 여전히 진입의 관문이 아니다(5.1.2.2).
- 복구는 세대가 오른 **다음 주기**에 일어난다. 세대를 올리는 서명 매니페스트 자체는
  8 절이다.
- 독립 적대 리뷰(8.5)는 아직이다. 원장 스키마는 `docs/WORKFLOW.md:403` 기준 High-risk 라
  구현자와 다른 사람의 리뷰를 요구하며, 이 로트는 그것을 대신하지 않는다.


## 2026-09-03 L5/L8 로트 — 서명된 4-가족 활성화와 관문 교체 (Manager 결정 59)

### 사람이 고른 것

HANDOFF.md 의 사람 결정 (7) 을 물었고, 세 선택지 중 **1번**을 골랐다:
**8.7.1(매니페스트) + 5.1.2.2(관문)를 한 로트로, 5.2.2 는 그다음.**

고르지 않은 둘과 그 대가를 함께 적는다. (2) 매니페스트만 별도 로트 — 로트는 가장
작지만 소비자 없는 매니페스트가 한 번 랜딩하고, 그것은 이 change 가 5.1 에서 이미
피한 모양이다("소비자 없는 매니페스트가 되지 않도록 5.3 의 고장 절반을 같은 로트에
붙였다"). (3) 셋을 한 로트로 — 서명 검증·관문 교체·진입 상한 해제를 한 번에 리뷰해야
하고, 그 규모는 5.5 가 적대 리뷰 13 라운드를 쓴 규모다.

### 착수 전에 잰 것 넷 — 계획을 바꾼 측정

**(1) 생산에는 서명된 전략 매니페스트가 하나도 배포돼 있지 않다.**
`~/.config/tossctl/` 에 `strategy-lane-authority-{KR,US}.json`,
`strategy-activation-{KR,US}.json` 이 **없다**(`find ~/.config/tossctl -maxdepth 2
-name 'strategy*'` → 0 건). 즉 오늘 생산의 전략 진입 파이프라인은 배선이 아니라
**입력 부재**로 정지해 있다. 이 사실이 이 로트의 토글 불변식 논증의 바닥이다 —
매니페스트가 없으면 승격 집합은 공집합이고, 공집합은 오늘의 값이다.

**(2) "정확히 네 가족 per market" 서명 매니페스트는 이미 있다 — 다만 종목 단위다.**
`internal/strategyrouter` 의 `strategy-lane-authority-<MARKET>.json` 은 ed25519
서명·digest 핀·소유자/권한 검사를 갖추고, `validProductionRouteCandidates` 가
`productionRouteDescriptors(market)`(정확히 4개)와 **개수까지** 대조한다
(`len(values) != len(want)` → 거절, `len(seen) == len(want)` → 통과). 그래서
legacy 3-lane 매니페스트는 구조적으로 아무것도 승격할 수 없다(태스크 4.3).
**그런데 그 매니페스트의 `Desired`/`Effective` 는 `scope`(종목·세대)마다 있다.**
여덟 `FamilyWorker` 는 종목이 없는 (시장, 가족, 레인, 버전) 열쇠이므로, 종목별 행을
worker 하나의 상태로 접으려면 "모든 scope 에서 ON 이면 ON" 같은 **새 규칙을 지어내야**
한다. 지어낸 규칙은 계약이 아니다. 그리고 8.7 이 요구하는 build·ProtectionReady
digest 는 그 매니페스트에 없다. 그래서 **별도 매니페스트**로 간다 —
design.md:210 의 "새 runtime activation 은 exact 4-family-aware signed manifest 가
필요하다"와 8.7 의 "**separate** human-approved operating activation" 이 같은 말이다.

**(3) 승격은 저장된 상태가 아니라 주기마다 건네는 값이어야 한다.**
활성화에는 만료가 있다(`scheduler` 의 v1 매니페스트는 최대 24시간). 레인은
프로세스 수명이다(latch 와 연속 실패 계수기가 기억이라서 — 5.1.2.1). 승격을 레인
생성 시점에 굽는 설계는 **묵은 ON** 을 만든다. 묵은 OFF 는 안전한 방향이지만 묵은
ON 은 아니다. 그래서 활성화는 `Run` 의 **인자**이고, 신선도는 파도가 준다.
따라서 `FamilyWorker` 의 `desired`/`effective` **필드는 없어진다** — 그 두 값은
활성화의 함수다. 동결 골든이 여덟을 `desired: OFF, effective: OFF` 로 얼린 것은
"활성화가 없을 때의 값"이고, 그것이 `Desired(FamilyActivation{}) == OFF` 로 시험된다.

**(4) 관문의 값은 on/off 가 아니라 가족 단위 고장 격리다 — 두 권위의 겉보기 충돌이
여기서 풀린다.** 5.1.2.1 은 "오늘 여덟을 관문으로 세우면 생산 진입이 0 이 되고 그것은
토글 OFF 불변식(§0-2, OFF = upstream 동작)에 어긋난다"고 적었다. 스펙은 반대로
"required activation authority 가 missing 이면 broker exposure-raising request 는
0건이어야 한다 (SHALL)"고 적었다. 둘을 동시에 만족시키는 유일한 모양은 **관문을
활성화된 런타임 안에만 세우는 것**이다: 활성화가 없으면 기존 시장 단위 경로가 그대로
돌고(= upstream 동작), 활성화가 있으면 그 시장의 넷이 각자 판정한다. design 이
partial 3-of-4 를 시장 전체 OFF 로 못 박았으므로 활성화된 시장의 넷은 전부 ON 이고,
그때 관문이 실제로 막는 것은 **잠긴 레인 하나**다 — 그 가족의 제안만 멈추고 이웃 셋은
계속한다. 그것이 이 change 전체가 사려는 것이다("시장 장애 격리이지 전략군 장애
격리가 아니다", design.md:3).

### Pre-Edit Gate (High-risk)

```text
- change id / task id: a112-run-four-strategy-families-independently / 8.7.1 + 5.1.2.2
- 대상 심볼(신규):
  - strategyrouter.FamilyActivation / LoadProductionFamilyActivation /
    ProductionFamilyActivationFileName / FamilyActivationForTest (tossos_testseams)
- 대상 심볼(기존 함수 내부 수정):
  - strategyworker.FamilyWorker.Run / .Desired / .Effective / newWorker /
    ProductionWorkers / dormant
  - strategyworker.Lane.Run
  - engine.strategyFamilyLaneStep / strategyLaneRuntime.evaluate / .runLane
  - engine.Context.runProductionStrategyMarketCycle
  - engine.coordinateMarketProposals
  - engine.strategyProposalAuthorityLoader.collectMarket
- 기존 동작 파악 근거 (측정):
  - FamilyWorker.Run 의 생산 호출자는 Lane.Run 하나 (codegraph callers + grep)
  - Lane.Run 의 생산 호출자는 strategyFamilyLaneStep 하나
    (internal/app/engine/strategy_lane_runtime.go:162)
  - coordinateMarketProposals 의 호출자는 collectMarket 하나
    (codegraph callers → strategy_proposal_authority.go:211)
  - ProductionLanesFrom 의 호출자는 restoreLatches 와 rebornLane 둘
    (codegraph impact → strategy_lane_latch.go:66, :193)
  - 기존 시험: strategyworker/{cycle,lane,restore,rehearsal}_test.go,
    golden_contract_test.go, worker_shape_test.go;
    engine/a112_lane_dependency_test.go
- upstream 상속 테스트 영향: no. 전략 진입 경로는 TossOS 제품 추가이며 upstream
  650 스위트에 없다. 회귀 방지는 make test(무태그) + make test-seams 양쪽.
- 실패 테스트 선행 작성: yes
- 안전 불변식 §0 위반 여부 검토: 통과
  - §0-2 (OFF = upstream): 활성화가 없으면 승격 집합은 공집합이고 관문은 세우지
    않는다. 생산에는 서명 매니페스트가 아예 없다(위 측정 (1)).
  - §0-3 (손절 즉시성): 이 로트는 entry 경로만 만진다. safety loop 는 안 만진다.
  - §0-7 (운영 토글 flip 은 사람이): 이 로트는 flip 수단을 **만들지 않는다.**
    승격의 유일한 출처는 사람이 서명한 파일이고, 시험용 주조 seam 은
    tossos_testseams 태그 아래에만 있다(scheduler.ActivationForTest 선례).
```

### 이 로트가 주장하지 않을 것 (착수 시점에 미리 적는다)

- 시장 단위 단일 제안 상한(`strategyhandoff.Capacity = 1`)은 그대로다 — 5.2.2.
- 기존 시장 단위 경로를 **지우지** 않는다. 활성화가 없을 때의 갈래로 남는다.
  지우는 것은 5.2.2 와 6 절이다.
- 8.7.2(활성화가 없거나 만료·폐기됐을 때의 운영 rollback 자세)는 이 로트가 아니다.
- 독립 적대 리뷰(8.5)는 이 로트가 대신하지 않는다.

### 랜딩한 것 — 8.7.1 + 5.1.2.2

**새 것 하나: 서명된 4-가족 활성화** (`internal/strategyrouter/production_family_activation.go`).
`strategy-family-activation-<MARKET>.json` 은 ed25519 서명 · digest 핀 · 소유자/권한
검사(0400, 소유 UID, symlink 거부, 열기 전후 동일 inode) · 64 KiB 상한 · **바이트가
정확히 이 구조체의 정규 직렬화와 같아야 한다**는 한 등식을 요구한다. 그 한 등식이
unknown field · 중복 키 · 뒤에 붙은 JSON · 필드 순서와 공백을 바꾼 사본을 **함께**
거절한다 — 같은 패키지의 `decodeProductionRouteManifest` 가 같은 이유로 같은 모양이다.

서술자는 **정확히 그 시장의 넷**이어야 하고, 표를 여기서 다시 적지 않고
`productionRouteDescriptors(market)` 를 그대로 쓴다. 옮겨 적으면 두 표가 갈리고,
갈리는 순간 한쪽 매니페스트로 켜진 레인이 다른 쪽에서는 미지의 레인이 된다.

`FamilyActivation` 은 불투명하다. 필드가 전부 비공개이므로 이 패키지 밖에서는
**영값만** 만들 수 있고, **영값이 안전한 값이다**(아무것도 승격하지 않는다). 그래서
"켜진 worker 를 아무나 만들 수 있는가"를 셈 시험으로 지킬 필요가 없다 — 승격을 얻는
유일한 길이 서명 검증을 통과하는 것이다. 시험용 주조 문은 `tossos_testseams` 태그
아래에만 있다(선례: `internal/scheduler/activation_testseam.go`).

**`FamilyWorker` 의 desired/effective 필드가 없어졌다.** 두 값은 활성화의 함수다
(`Desired(activation)` / `Effective(activation)`). 골든이 여덟을
`desired: OFF, effective: OFF` 로 얼린 것은 "활성화가 없을 때의 값"이고,
`golden_contract_test.go` 가 이제 `Desired(FamilyActivation{})` 로 그것을 시험한다.
필드가 아니라 인자인 이유는 만료다 — 레인은 프로세스 수명이고 활성화에는 24시간
상한이 있으므로, 승격을 값 안에 구우면 **묵은 ON** 이 생긴다. 묵은 OFF 는 안전한
방향이지만 묵은 ON 은 아니다. `newWorker` 에서 상태 인자가 사라졌으므로 시험조차
상태를 직접 주조할 수 없다.

**관문**(`internal/app/engine/strategy_family_activation.go`)은 `coordinateMarketProposals`
안, 조정자 `Submit` **앞**에 선다. 뒤에 세우면 중재가 이미 한 범위의 승자를 골라
버렸고, 그 승자의 레인이 잠겨 있으면 그 범위는 이웃 가족이 이길 수 있었는데도 통째로
닫힌다. 관문이 하는 일은 그 가족의 레인에게 묻는 것 하나이고, 레인이 낸 봉투를 그대로
조정자에 넣는다.

**레인을 제안 수집 앞으로 옮겼다.** 순서가 안전이다: 관문은 레인의 잠금을 읽어
판정하고 그 잠금은 원장에서 태어난다(5.3.3). 뒤에 세우면 재시작 뒤 **첫 주기**에
durably 잠긴 레인이 열린 것으로 읽히고 그 가족의 제안이 조정자에 닿는다. 그 창은 한
주기뿐이라 어떤 행동 시험도 우연히 잡지 못하므로 순서를 구조로 못 박았다
(`TestTheLanesAreBuiltBeforeTheProposalsAreCollected`).

**다섯 digest 는 두 단계에서 결속한다.** 보정·달력·빌드는 제안 수집 단계에 존재하므로
거기서, 위험 번들과 ProtectionReady 는 그 단계에 **없으므로**(둘 다 제안 뒤에 수집된다)
`buildProductionStrategyMarketWorker` 에서 결속한다. 없는 사실을 결속하면 그 결속은
어떤 정상 입력으로도 참이 될 수 없고, 그것이 이 change 가 이미 한 번 만들었다 고친
문 없는 fail-closed 다. 검증은 한 번, 결속은 사실이 사는 자리에서.

### 반증 — 58개, 그리고 그것이 **코드를 네 줄 지웠다**

반증이 찾은 것 중 가장 무거운 것부터.

**판정이 셋이었다 (M2 · M7 · M27 전부 SURVIVED).** "서술자는 정확히 넷" 규칙을
세 검사가 나눠 지키고 있었다 — 입력 개수(`len(body.Descriptors) == 4`) · 중복 거절 ·
완전성(`len(state) == 4`). 셋 중 **아무 둘이면 충분**하므로 셋을 각각 지운 변이가
전부 살아남았다. 각자가 상대의 시험을 통과시킨다(5.3.3 이 원장에서 만난 것과 같은
모양). 개수 검사가 나머지 둘의 재진술이므로 그것을 **지웠다**. 남은 둘은 서로 다른
성질이고 각각 다른 입력에서만 짐을 진다: 중복 거절은 `[c,r,w,b,b]`(다섯 중 넷이 다
있음)를, 완전성은 `[c,r,w]`를. 지운 뒤 두 변이 다 CAUGHT.

**시장 결속을 우연이 지키고 있었다 (M1 SURVIVED).** `body.Market != config.Market` 를
지워도 색이 안 바뀌었다 — 기존 사례는 몸통만 US 로 바꾸고 서술자는 KR 로 두었기
때문에 **서술자 표 대조**가 대신 잡았다. 그래서 몸통·서술자·digest 가 전부 US 로
온전하고 어긋난 것이 **파일 이름이 말하는 시장** 하나뿐인 사례를 만들었다
(`TestTheOtherMarketsWholeManifestInThisMarketsFilePromotesNothing`). 막지 못하면
KR 파일이 US 레인을 켠다.

**닿지 않는 방어 셋을 지웠다.** `Verified()` 의 `validMarket`(M28), 검증의
`len(want) == 0`(M29), 관문의 `loader.lanes == nil`(E10) — 셋 다 지워도 아무 색이
바뀌지 않았다. 지키는 것이 없는 방어다. `len(want) == 0` 은 지우는 대신 **닫아 두는
등식**을 시험으로 만들었다: 파일 이름이 있는 시장의 집합 == 서술자 표가 있는 시장의
집합(`TestEveryMarketWithAnActivationFileNameHasADescriptorTable`). 세 번째 시장에
이름만 붙이면 서술자 0 개가 "완전"해지고 그때 이 시험이 실패한다.

**닫힌 시장이 관문의 활성화를 버렸다 (E6 SURVIVED).** 준비된 시장만 재고 있었다.
그 상태의 대가는 관측이다 — 관문은 섰는데 그 주기의 레인 관측은 승격을 못 보고
DORMANT 를 적는다. 즉 시장이 고장 난 바로 그 주기에 운영자의 레인 화면이 어두워진다.
`fail` 클로저가 활성화를 함께 싣게 고쳤다(반환값마다 `carry(...)` 를 부르는 첫 판본은
이 change 가 반복해서 고쳐 온 **무시할 수 있는 답**이었다).

**동등 변이 하나는 이유를 적어 둔다.** M31(`productionRouteDescriptors(body.Market)` →
`(config.Market)`)은 SURVIVED 이고 결함이 아니다. 바로 위에서
`body.Market == config.Market` 가 강제되고 그 강제 자체가 M1 로 CAUGHT 이므로 두 식은
**증명 가능하게** 같다. "동등 변이"를 근거 없이 넘기지 않기 위해 그 증명을 적는다.

라운드별: 매니페스트 28개(21 CAUGHT · 5 SURVIVED · 2 anchor 실패) → 고친 뒤 6개
(4 CAUGHT) → 3차 5개(3 CAUGHT · 1 동등 · 1 anchor) · 관문 10개(6 CAUGHT · E7 은 의도한
대조군 · 2 SURVIVED · 1 anchor) → 고친 뒤 4개(4 CAUGHT) · 뒤 단계 결속 5개(5 CAUGHT).
E17(활성화가 없어도 결속을 요구한다)이 CAUGHT 인 것이 토글 OFF 방향이 실제로 짐을
진다는 증거다.

### 열거표 넷이 걸렸고 넷 다 판단해서 고쳤다

1. `TestOnlyThePackageLevelStepEverRunsInsideALane` — 얼린 한 줄의 철자가
   `strategyFamilyLaneStep(lane, promotion)` 로 바뀐다. 여전히 한 줄, 여전히 패키지
   수준.
2. `TestTheFamilyLaneStepCarriesNothingButItsLane` → `…AndTheSignedPromotion`.
   **개수 검사를 타입 목록으로 바꿨다.** 개수 검사는 정당한 인자가 늘 때 반드시
   걸리고, 걸렸을 때 통과시키는 유일한 방법이 숫자를 고치는 것이라 감시 대상과 같은
   손짓이 된다. 이제 새 인자는 자기 타입을 목록에 적어야 하고 그 타입이 능력이면
   같은 시험의 계산이 잡는다.
3. `TestTheSingleProposalAssumptionLivesOnlyWhereTheCensusSaysItDoes` — 새 파일이
   `routes.entries[0]` 을 써서 걸렸다. 그 철자는 "제안이 하나"라는 가정과 같으므로
   **코드를 고쳤다**: 첫 항목을 기준으로 삼는 대신 집합으로 "항목 전부가 같은 하나에
   동의한다"를 세게 했다. census 는 손대지 않았다.
4. `TestTheSeamFilesStayWithinTheirDeclaredClosure` — 조정자 seam 파일의 허용 목록에
   `internal/strategyworker` 를 더했다. 그 한 줄이 능력을 늘리지 않는다는 것은 여기서
   주장하지 않고 **그 패키지의 `dependency_closure_test.go` 가 `-deps`/`-deps-test` 로
   증명한다**(양성 대조로 걸음이 얕지 않다는 것까지 잰다).

그리고 `else if` 를 두 `if` 로 풀었다. Go AST 는 `} else if` 를 **같은 좌표의 else 와
if 두 노드**로 내므로, 좌표로 분기를 세는 이 change 의 열거표에서 두 줄이 구별되지
않는다.

### 이 로트가 주장하지 않는 것

- 시장 단위 단일 제안 상한(`strategyhandoff.Capacity = 1`)은 그대로다 — 5.2.2.
- 기존 시장 단위 경로를 지우지 않았다. 활성화가 없을 때의 갈래로 남는다.
- 8.7.2(활성화가 없거나 만료·폐기됐을 때의 운영 rollback 자세)는 이 로트가 아니다.
- 독립 적대 리뷰(8.5)는 아직이다. 진입 경로는 High-risk 이므로 구현자와 다른 사람의
  리뷰를 요구하며 이 로트는 그것을 대신하지 않는다.
- 서명된 매니페스트를 **실제로 만들어 배포하는 절차**는 이 로트에 없다. 생산에는
  그 파일이 하나도 없고(측정), 그래서 오늘 생산 동작 변화는 0 이다.

## 2026-09-04 — 8.5 독립 적대 리뷰와 그것이 연 수정 [결정 60]

### 리뷰 실행 방식과 범위

`docs/WORKFLOW.md` 는 "구현을 만든 컨텍스트와 그것을 검증하는 컨텍스트는 항상 별도
세션"을 요구한다. 리뷰어 **일곱**(testing · security · maintainability · performance ·
api-contract · red-team · 적대 chaos)과 **다른 모델**(codex, `model_reasoning_effort=high`)이
각각 신선한 컨텍스트에서 돌았다.

**범위를 좁힌 것을 명시한다.** 리뷰 대상은 리뷰된 적 없는 로트
`06be6ca9^..515f85a8` 의 `internal/` 이다. 브랜치 전체 diff 는 6,225 파일이라 그대로
주면 리뷰어가 어디에도 닿지 못한다. 앞선 커밋들은 각자의 라운드를 review.md 에 갖고
있다. 이 축소는 침묵한 생략이 아니라 기록된 선택이다.

### 판정: 이대로 main 불가

**P0-1 · 관문 거절이 노출을 제조한다.** 레드팀이 이 저장소 픽스처로 실측했다:
매니페스트 없음 = entries 2 → `HANDOFF_OVER_CAPACITY` → 주문 0. 활성화 + 레인 잠금 =
entries 1 → **주문 나감**. 고장이 전혀 없는 부분 매니페스트(4 중 1만 ON)에서도 = 1 →
**주문 나감**. 같은 기전을 이 change 가 이미 두 번 고쳤고(5.4.2·5.4.3) 그 경고 주석이
관문 바로 위에 있다. 내 시험이 못 잡은 이유는 **한 종목·두 가족**만 세워 범위 *안*
격리만 쟀기 때문이다 — 익스플로잇은 두 종목이다.

**P0-2 · 위험 번들 결속이 충족 불가능하다.** `r.snapshot.BundleDigest` 는
`sha256JSON({Scope, Policy, Entries})` 이고 `RiskSnapshotScope` 는 `Symbol` 과
`AsOf time.Time` 을 품는다. 엔진은 `scope.AsOf.Equal(loader.observedAt)` 를 요구하고
그 값은 파도의 벽시계다. 서명된 상수가 같아질 수 없으므로, 매니페스트를 배포하는
순간 두 시장이 영원히 dormant 가 된다. 시험이 초록이던 이유는
`buildWorkerUnderActivation` 이 **시험 시점에 살아 있는 digest 를 읽어** 매니페스트에
넣었기 때문이다. 파일을 미리 서명하는 운영자는 그럴 수 없다.

**P0-3 · 그 결속이 주문 경로에 없다.** 판정 결과인 `StrategyMarketWorker.Effective` 를
읽는 곳은 화면(`strategy_runtime_projection.go:107`)과 supervisor 검증뿐이다. dormant
worker 도 `Cycle` 을 들고 계속 돌며 `dispatchHandoff().Deliver` 로 주문에 닿고,
`strategyDispatchCycle.dispatch` 는 자기 승인 사슬을 다시 유도하며 가족 활성화를 보지
않는다. P0-2 와 겹치면 **화면은 멈추고 주문은 계속 나간다.**

**배포 불가 사유(추가).** 18 필드 매니페스트에 문서·생성기·골든이 전무하다. 형제 둘은
`docs/operations.md` 에 절을 갖는다. 정규화 규칙도 형제와 반대다(바이트 일치 vs
pretty-print 허용). 약 30 가지 원인이 sentinel 하나로 뭉치고 엔진이 그것마저 버린다.

### 리뷰어가 틀린 것 (받아 적지 않고 대조한 것)

- **env 신뢰 앵커**: 이 로트가 만든 구멍이 아니다. `strategy_schedule_authority.go:252`
  도 같은 방식으로 env 에서 KEY_ID/PUBLIC_KEY 를 읽는다 — 저장소 전체의 기존 자세다.
- **재생 공격**: `productionRouteDigest(data) != config.ManifestDigest` 핀이 파일만의
  롤백을 막는다. env 핀까지 되돌리는 것은 운영자의 의도적 행위이고 수명 24h 가 경계다.
  잔여는 세대 래칫이 없다는 것뿐이며 8.7.2 로 넘긴다.
- **TrustedKey 길이 삭제 시 패닉**: 엔진이 호출 전에 길이를 검증하므로 생산 경로에서는
  패닉에 닿지 않는다. 심각도가 과대평가됐다.

### 방법에 대한 결론

반증 58 개와 자기 적대 리뷰 13 라운드가 이 셋을 하나도 못 잡았다. 전부 **내가 고른 축**
안에서 움직였기 때문이다. 신선한 컨텍스트 일곱이 프레임 자체를 깼고 그중 하나는
익스플로잇을 실측으로 재현했다. High-risk 로트에서 반증 횟수는 별도 검증 컨텍스트를
대체하지 못한다.

### 8.8.1 구현 기록 (이 커밋)

관문 거절이 소유자 범위를 통째로 지우면 시장을 닫는다(`FAMILY_GATE_CLOSED`,
`RefusedCount = RoutedCount`). 범위 **안에서** 가족 하나가 막히고 이웃이 이기는 것은
그대로 둔다 — `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading` 이 그것을 지키고
이번 로트에서도 초록이다. `arbitration.gated` 는 이제 스냅샷으로 나가
(`GatedCount`·`GatedOutcomes`) 쓰이기만 하던 필드가 읽히게 됐다.

**반증**: 새 가드를 `if false && …` 로 무력화하니
`TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve` 가 "시장이 열려 있다
(선택=1 건)"으로 빨개졌다. 원복은 해시 검증으로 했다
(`9e757ca147d0d66035c11bc9511d5d2f65c12b3585ff1b17e9d64e3a60642ec8`).

**FLM 재측정**: 편집한 두 함수와 같은 파일의 형제 둘, 총 네 묶음을 다시 만들었다.
분기 처리 상태는 커버리지 프로파일로 다시 쟀고(태그 78.3% / 무태그 69.5%, 테스트별
27 개), 귀속 완전성 등식이 모든 분기에서 성립했다. `collectMarket` 은 새 분기가
**가운데**에 들어가 옛 B10~B15 가 B11~B16 으로 밀렸으므로 조건을 소스와 하나씩
대조했다. B10 진입 7 회 중 다섯이 US 시장이라는 fixture 성질도 그대로 적었다.

### 8.8.2 구현 기록 (2026-09-04)

**무엇을 바꿨나.** 매니페스트 본문의 두 필드를 서명 가능한 값으로 교체했다.

| 앞 판본 | 왜 안 됐나 | 새 판본 |
|---|---|---|
| `risk_bundle_digest` = per-cycle 스냅샷 봉인 | `RiskSnapshotScope` 가 `Symbol` 과 `AsOf`(파도 벽시계)를 품어 파도마다·종목마다 바뀐다 | `risk_policy_digest` = 운영자가 이미 핀하는 `TOSSOS_RISK_BUCKET_<M>_MANIFEST_SHA256` |
| `protection_ready_digest` = 살아 있는 readiness 신원 | 스냅샷마다 바뀌므로 등식이 성립할 수 없다 | `protection_ready_min_generation` (uint64, 0 금지) |

넷(경로·보정·달력·빌드·위험 정책)은 제안 수집 단계가 등식으로 결속한다 — 넷 다
그 단계에 존재하고 활성화 수명 동안 변하지 않는다. ProtectionReady 하한만
`strategyDispatchCycle.dispatch` 가 결속한다.

**계약 편차를 명시한다.** 8.7 원문은 "current … ProtectionReady **digests**" 다.
등식으로 읽으면 어떤 정상 입력으로도 참이 될 수 없으므로 하한으로 바꿨다. 세대가
단조 증가하므로 하한은 안전 방향으로만 어긋난다. 이것은 침묵한 생략이 아니라
기록된 편차다.

**들어낸 것.** `buildProductionStrategyMarketWorker` 의 결속(분기 둘)을 삭제했다.
그 함수가 만드는 `Effective` 는 화면(`strategy_runtime_projection.go:107`)과 승격
판정만 움직이고, 주문은 refresh worker 의 사이클이 `dispatchHandoff().Deliver` 로
내보내며 그 경로는 이 서술자를 읽지 않는다.

**시험이 값을 시스템에서 읽지 않는다.** 앞 판본의
`buildWorkerUnderActivation` 은 살아 있는 digest 를 읽어 매니페스트에 넣었고,
그래서 충족 불가라는 사실 자체를 가렸다. 그 헬퍼와 시험을 지우고
`TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor` 로 바꿨다 —
하한을 숫자로 적고 배선의 보호 세대(9)와 견주며, 세 경우(활성화 없음 / 하한과
같음 / 하한보다 낮음)를 함께 세운다.

**반증 둘, 서로 다른 행.** 가드를 `if false && …` 로 지우면 "낮으면 거절" 행이
빨개지고(주문 1건, want 0), `if true || …` 로 항상 거절하면 "같으면 나간다" 행이
빨개진다(주문 0건, want 1). 한 방향만 빨개지는 시험은 "항상 거절" 판본을
통과시키므로, 두 방향이 다 필요하다. 원복은 해시 검증으로 했다.

**FLM 재측정**: 편집이 끌어온 번들 17 개(형제 함수와 시험 함수 포함, 새 스캐폴드
둘)를 다시 만들었다. 커버리지 재측정(태그 78.3% / 무태그 69.4%, 테스트별 69 개)에서
귀속 완전성 등식이 처음에는 `dispatch` 의 여섯 arm 에서 깨졌다 — 내 59 개 집합 밖의
시험이 그 arm 에 들어갔기 때문이다. dispatch 를 모는 파일 넷의 시험 전부를 집합에
넣어 등식을 복구했다. **등식이 깨진 것을 무시하지 않고 집합을 넓힌 것이 요점이다.**
`dispatch` 는 새 분기 둘이 가운데에 들어가 옛 B6~B14 가 B8~B16 으로 밀렸으므로
조건을 소스와 하나씩 대조했고, `buildProductionStrategyMarketWorker` 는 분기 둘이
사라져 그 행을 지우고 옛 B8 을 새 B6 으로 옮겼다.

**알고도 남긴 것 하나.** `strategyEntrySupervisor.runMarket` 의 FLM 분기표는 16 개
중 15 개만 싣는다. 이 로트가 만든 것이 아니고(내 편집은 그 함수에 분기를 더하지
않았다) 완전성은 BTM 이 지키므로 그대로 두었다. 8.8.4 에서 정리한다.

## 2026-09-04 태스크 8.8.3 — 서명을 빼고, 사람이 만들 수 있게 한다 (구현 기록) [a112 결정 61]

### 사람 결정 61

사람이 "서명이 필요 없도록 하면 안 되나" 라고 물었고, 세 선택지를 평문으로 제시한 뒤
**2번**을 골랐다: ed25519 를 빼고 신뢰 앵커를 외부 digest pin 하나로 둔다.

제시한 근거와 그 반대편을 여기 남긴다 — 나중에 이 결정을 되돌아볼 사람이 무엇이
교환됐는지 알아야 한다.

- **서명이 실제로 더 사는 것은 하나다.** digest pin 은 서명과 **독립적으로** 파일
  치환을 이미 막는다(`:285` 의 `productionRouteDigest(data) != config.ManifestDigest`
  는 서명 검증과 별개 문이었다). 서명이 추가로 사는 것은 승인자와 배포자의 분리다.
- **그 성질은 개인키가 배포 호스트 밖에 있을 때만 참이다.** 키가 config 옆에 있으면
  env 를 쓸 수 있는 것은 무엇이든 파일과 핀을 함께 바꿀 수 있으므로 서명은
  아무것도 더 지키지 않고 키 분실·교체 비용만 남는다. 사람이 키를 호스트 밖에 둘
  생각이 없다고 답했다.
- **포기한 것을 명시한다.** 이 저장소에서 자율로 도는 에이전트가 config 와 env 를
  둘 다 쓸 수 있다면 4-가족 런타임을 켤 수 있다. 앞 판본에서는 그것이 개인키 없이
  **구조적으로** 불가능했고, 이제는 정책이다. 이것이 결정 61 이 실제로 교환한 값이다.
- **선례가 이미 있다.** 후보 임계값 활성화는 사람 승인 파일이면서 서명이 없다
  (`internal/app/engine/strategy_candidate_authority.go:193` 의
  `<PREFIX>_<MARKET>_ACTIVATION_SHA256`). 저장소가 두 관례를 다 갖고 있었고 이
  매니페스트는 그중 가벼운 쪽으로 옮겼다.
- **spec 은 바뀌지 않는다.** `spec.md` 에서 "signed" 는 SHADOW manifest 에만 붙는다.
  활성화 서명 요구는 `design.md:210` 한 줄이었고 그 줄을 개정했다.

### 정정한 전제 하나

사람의 물음에는 "서명을 빼면 도구가 필요 없다" 가 함축돼 있었고 그것은 **틀렸다.**
정규 바이트 등식(`decodeProductionFamilyActivation` 의 `bytes.Equal(canonical, data)`)
은 서명과 무관한 별개의 문이고 그대로 남는다. Go `encoding/json` 출력은 구조체 선언
순서·무들여쓰기·끝 개행 없음이므로 에디터로 쓴 JSON 은 통과할 수 없다. 그래서 도구는
여전히 필요하고, 달라진 것은 **그 도구가 비밀을 갖지 않는다**는 것뿐이다.

### 무엇이 랜딩했나

- `internal/strategyrouter/production_family_activation.go`: `signature`·`key_id`·
  `signature_algorithm` 필드와 `productionFamilyActivationManifest` 래퍼가 없어졌다.
  `FamilyActivationConfig` 에서 `TrustedKeyID`·`TrustedKey` 가 없어졌다. domain 상수가
  `…/ed25519/v1` → `…/sha256-pin/v1` 이 됐다(서명이 없는데 도메인이 ed25519 를
  말하면 그 문자열이 거짓말이 된다).
- 새 exported 표면 둘: `FamilyActivationDocument`(사람이 채우는 값)와
  `EncodeProductionFamilyActivation`(정규 바이트). **서술자 넷은 문서에 없다** — 켤
  가족만 받고 lane id·버전·수평선은 검증기가 쓰는 바로 그 표에서 유도한다. 운영자가
  손으로 적으면 표가 바뀔 때 조용히 옛 이름을 넣게 되고 그 거절은 배포 시각에야 보인다.
- `internal/app/engine/strategy_family_activation.go`: 배포가 관리할 env 가 시장마다
  셋에서 **하나**로 줄었다(key id·공개키 제거).
- `tools/a112-family-activation/`: 비밀 없는 생성기. `0400` 으로 쓰고 마지막 줄에
  env 핀을 그대로 낸다.
- `docs/operations.md`: 형제 절들과 같은 모양의 운영 절 하나. 다섯 결속 값의 출처와
  "핀을 안 넣으면 파일이 있어도 아무것도 안 켜진다" 를 적었다.
- `internal/strategyrouter/testdata/strategy-family-activation-KR.json`: **커밋된 골든.**

### 골든이 이 로트의 핵심이고, 그 이유

기존 로더 시험은 전부 `json.Marshal(body)` 로 바이트를 만들고 검증기도 같은 호출을
한다. 그래서 "도구가 낼 수 있는 바이트" 와 "검증기가 받는 바이트" 가 갈려도 전부
초록이다. 그리고 승격 경로 시험은 파일을 아예 안 지나간다(`FamilyActivationForTest`
가 구조체를 직접 주조한다). 즉 **파일 → 핀 → 승격 경로를 끝까지 도는 실행이 하나도
없었다.**

골든은 도구가 낸 커밋된 바이트이고, 그 SHA-256 이 시험에 **손으로 적힌 리터럴**이다
(`goldenFamilyActivationKRDigest`). 시험은 그것을 재생성하지 않는다. 직렬화가 바뀌면
골든이 빨개진다 — 반증 M4 가 그것을 확인했다.

절차 기록: RED 은 골든 시험 셋이 "testdata 없음" 으로 실패한 것이고(능력 부재),
digest 리터럴은 GREEN 단계에서 도구 출력을 옮겨 적었다. 그것은 기대값 수정이 아니라
새 산출물이지만, **커밋된 산출물의 해시를 그 산출물에서 읽었다**는 사실은 남는다 —
막는 것은 바이트가 커밋돼 있고 시험이 재생성하지 않는다는 것 하나다.

### 반증 7개, 전부 CAUGHT (둘은 두 번째 판에서)

| # | 변이 | 결과 |
|---|---|---|
| M1 | digest 핀 검사 삭제 | CAUGHT |
| M2 | 정규 바이트 등식 삭제 | CAUGHT |
| M3 | 서술자 정렬 삭제 (같은 입력이 다른 바이트) | **틀린 기록이었다 — 아래 정정 절 참조** |
| M4 | domain 상수를 ed25519 로 되돌림 (도구·검증기 드리프트) | CAUGHT |
| M5 | 핀을 자기 자신과 비교 (자기 참조) | CAUGHT |
| M6 | 생산 엔진 파일이 인코더를 호출 | CAUGHT |
| M7 | 도구가 인코더를 안 부름 (이름 변경 시뮬레이션) | **첫 판 SURVIVED** |

**M1 의 첫 시도는 무효였다.** `sed` 구분자로 `|` 를 쓰면서 패턴에 `||` 가 있어 치환이
일어나지 않았고, 그 상태의 "SURVIVED" 는 "변이가 대상에 닿지 않았다" 였다. 기억에
"뮤테이션은 대상에 닿아야 한다" 로 남은 것과 같은 모양이라, 다시 할 때 치환 개수를
`assert` 로 확인하고 돌렸다 — 닿자마자 CAUGHT.

**M7 이 진짜 결함을 찾았다.** 가드의 대조군이 심볼 **합집합**을 세고 있어서, 인코더
참조가 사라져도 `FamilyActivationDocument` 가 대신 세어져 통과했다. 즉 인코더 이름이
바뀌면 가드가 조용히 아무것도 안 세게 된다. 셈 단위를 (파일 × 심볼) 로 올려서 고쳤고,
다시 돌리니 CAUGHT. 5.5·5.1.1·5.6.1 이 도달한 결론과 같다 — 한 축짜리 셈은 종료하지
않으므로 세는 범위를 올린다.

### 이 로트가 주장하지 않는 것

- **US 골든이 없다.** KR 하나만 커밋했다. 두 시장의 경로는 같은 함수이고 시장 결속은
  별도 시험이 재므로 골든을 둘 두면 같은 것을 두 번 재게 된다. 다만 US 매니페스트를
  실제로 만들어 본 실행은 없다.
- **생산 동작 변화 0.** 생산에 매니페스트가 0 건이므로(측정: `~/.config/tossctl`) 이
  편집으로 켜지거나 꺼지는 것이 없다. 신뢰 앵커가 하나 줄었을 뿐 남은 앵커는 그대로다.
- **8.7.2 는 여전히 열려 있다.** 만료·폐기 시의 rollback 자세는 이 로트에 없다.
- **키 관리 절차가 없어진 것이지 승인 절차가 없어진 것이 아니다.** env 핀을 배포하는
  것이 여전히 사람의 행위다.

### Pre-Edit Gate

```text
- change id / task id: a112-run-four-strategy-families-independently / 8.8.3 [결정 61]
- 대상 심볼: strategyrouter.LoadProductionFamilyActivation ·
  .decodeProductionFamilyActivation · .validateProductionFamilyActivation ·
  engine.strategyProposalAuthorityLoader.loadFamilyActivation
- 기존 동작 파악 근거: CodeGraph 호출자 12 (생산 1 = loadFamilyActivation, 시험 11),
  production_family_activation_test.go 13 시험
- upstream 상속 테스트 영향: no (a112 가 만든 신규 파일)
- 실패 테스트 선행 작성: yes (골든 시험 셋, "testdata 없음" 으로 RED)
- 안전 불변식 §0 위반 여부 검토: 통과 (생산 매니페스트 0 건 → 동작 변화 0)
- Function Logic Map: **not-applicable** — 넷 다 frozen base `aeeb209e` 에 부재.
  `git cat-file -e aeeb209e:internal/strategyrouter/production_family_activation.go`
  가 "exists on disk, but not in aeeb209e" 로 확인. 8.7.1 이 base 이후에 만든
  신규 leaf 함수다.
```

### 검증 — 전부 돌았다 (2026-09-04, Go 툴체인 복구 후)

| 명령 | 결과 |
|---|---|
| `make lint` | exit 0 (`go vet ./...` + `go vet -tags tossos_testseams ./...`) |
| `gofmt -l internal/ tools/ cmd/` | 무출력, exit 0 |
| `make test` (무태그) | exit 0 · **99 ok / 0 FAIL** |
| `make test-seams` (태그) | exit 0 · **100 ok / 0 FAIL** |
| `make test-race` | exit 0 |
| `make sdd-sync` | exit 0 (`all indexes current`; GBrain advisory busy 는 무해) |
| `make sdd-check` | exit 0 |
| `check_analysis.py --change a112` | exit 0 · `evidence complete or diff-proven exempt` |
| `python3 -m unittest discover -s tools/logic-map` | 77 OK |
| `openspec validate a112 --strict --no-interactive` | valid |
| 반증 7종 | 위 표, 전부 CAUGHT |

`make gate CHANGE=…` 는 not-applicable — change **완료** 게이트라 미완료 태스크가
있으면 2/10 에서 멈춘다(HANDOFF.md 5 절).

### 환경 사고 둘, 그리고 그중 하나가 만든 가짜 초록

**(1) Go 툴체인이 세션 중간에 사라졌다.** `rtk proxy go …` 가 30여 회 성공한 뒤
`go: No such file or directory` 가 됐다. 원인을 뒤져 보니 Go 는 공식 위치가 아니라
`~/Downloads/go1.26.5.linux-amd64/go` 에 풀려 있었고(2026-07-23), `~/.bashrc`·
`~/.profile` 어디에도 PATH 설정이 없었다 — 즉 쉘에서는 원래부터 안 잡혔고 초반에
돌던 것은 프로세스가 물려받은 별도 환경 덕이었다. 사람이 `/usr/local/go` 로 옮기고
`~/.bashrc` 에 PATH 를 넣은 뒤 위 표를 전부 다시 돌렸다.

**(2) 그 사이에 내가 가짜 초록을 하나 만들었다 — 기록해 둔다.**

```bash
$(go env GOROOT)/bin/gofmt -l internal/ tools/ 2>&1 | head -5 && echo "GOFMT CLEAN"
```

`go` 가 PATH 에 없어 `go env GOROOT` 가 빈 문자열을 냈고 → `/bin/gofmt` 실행 →
없는 파일 → 에러가 **파이프로** 흘러가고 → `head` 가 exit 0 → `&&` 통과 → echo.
재현해서 확인했다. **gofmt 는 그때 한 번도 안 돌았다.** 기억
`missing-tool-reports-clean` 이 적어 둔 그대로다. 판정 명령을 파이프로 감싸면
파이프라인 종료 코드가 마지막 명령의 것이 되어 앞의 실패가 사라진다 — 판정에는
파이프를 쓰지 않고 `exit=$?` 를 직접 본다.

**같은 구조인데 성립하는 것과 아닌 것을 갈라 둔다.** `BUILD OK`·`VET OK` 도 같은
`| head && echo` 모양이었으나, 그때는 `rtk proxy go` 가 살아 있었고 출력이 **비어
있었다**는 별개 증거가 있으므로 성립한다. `go test` 결과는 실제 패키지명과
소요시간이 찍혔으므로 성립한다. 성립하지 않은 것은 도구 자체가 없었던 gofmt 하나다.

**(3) 무태그 통과가 태그 빌드의 증거가 아님을 이 로트가 실제로 겪었다.**
`internal/app/engine/a112_family_gate_test.go` 가 지운 상수 둘
(`strategyFamilyActivationKeyIDEnv`·`strategyFamilyActivationPublicKeyEnv`)을 계속
참조하고 있었다. `tossos_testseams` 태그 아래에서만 컴파일되므로 `make test` 는
그것을 못 봤고, grep 으로 찾아 고쳤다. 이후 `go vet -tags tossos_testseams` 와
`make test-seams` 100 ok 로 컴파일까지 확인했다. a118 이 게이트에 태그 스위트를
넣어 둔 덕에 이 결함이 커밋 전에 잡혔다.

## 2026-09-04 태스크 8.8.3 — 독립 적대 리뷰와 그것이 연 수정

위 구현 기록을 쓴 뒤 다른 에이전트에게 read-only 적대 리뷰를 시켰다(범위: 미커밋
working tree 전체, 8.8.3 로트). 판정 **HOLD**, `P0=0 / P1=3 / P2=4`.

세 P1 을 전부 직접 재확인한 뒤 고쳤다. 리뷰어 말을 그대로 받지 않고 각각을 명령으로
다시 잰 기록을 남긴다.

### P1-1 — 운영 문서가 **없는 env 이름**을 부른다 (확인함, 고침)

`docs/operations.md` 의 생성 명령이 `$TOSSOS_STRATEGY_ROUTE_KR_MANIFEST_SHA256` 을
넘겼다. 그 이름은 저장소에 없다. 엔진이 실제로 핀하는 값은
`strategy_route_authority.go:19` 의 `TOSSOS_STRATEGY_LANE_KR_MANIFEST_SHA256` 이다.

```
rg 'TOSSOS_STRATEGY_ROUTE'  →  단 한 건, docs/operations.md:279 (문서 자신)
rg 'TOSSOS_STRATEGY_LANE'   →  internal/app/engine/strategy_route_authority.go:19-22
```

고장 방식이 나쁘다. 셸이 없는 변수를 빈 문자열로 펼치고, 도구는 값의 **의미**를
검사하지 않으므로(그 판정을 두 곳에 두지 않기로 한 설계다) 빈 `route_manifest_digest`
를 실은 매니페스트를 성공적으로 만들어 낸다. 운영자는 그 digest 를 핀하고, 런타임은
결속 불일치로 거절하는데 그 거절이 오늘 "매니페스트를 배포 안 함"과 **같은 오류**로
보인다(구별은 8.8.4). 즉 문서대로 따라 한 사람이 아무 진단도 없이 안 켜지는 시스템을
얻는다. fail-closed 라 안전 구멍은 아니지만, **문서가 작동하는 매니페스트를 만들지
못한다**는 것은 이 태스크의 산출물 넷 중 하나가 결함이라는 뜻이다.

고친 것: 이름을 바로잡고, 다섯 값의 출처를 산문이 아니라 **표**로 적었다.
`calibration_digest` 는 projection 에 나오지 않아 경로 권한 파일
`strategy-lane-authority-<MARKET>.json` 의 같은 이름 필드에서 읽어야 하는데
(`production.go:122` 로 확인), 앞 판본은 "경로 권한이 말하는 보정 digest" 라고만 적어
어디를 열어야 하는지 말하지 않았다.

### P1-2 — 문서의 절차가 **이틀째부터 실패한다** (확인함, 고침)

도구가 `os.WriteFile(out, data, 0o400)` 였고 문서의 `-out` 은 살아 있는 경로였다.
0400 파일은 소유자도 다시 열어 쓸 수 없다. 매니페스트 수명 상한이 24시간이므로
재발급은 **매일** 있는 일인데, 두 번째 실행이 이렇게 끝난다.

```
run1 exit=0   (커밋된 골든과 바이트 동일, digest 도 핀과 같음)
run2 exit=1   open …/tool-out-KR.json: permission denied
```

`chmod` 를 안내해 덮어쓰게 만드는 것이 아니라 **살아 있는 매니페스트를 건드리지
않는 쪽**으로 고쳤다. 도구는 `O_EXCL` 로 만들고, 이미 있으면 "덮어쓰지 않는다" 로
실패한다 — 실패 자리가 읽히는 중인 파일을 자르기 **전**으로 옮겨졌고, 오류가 원인을
말한다("permission denied" 는 원인을 말하지 않았다). 문서에는 재발급 4 단계(새 경로
→ 핀 교체 → 옛 파일 치환 → 재시작)와, 2·3 사이의 어긋난 창이 어느 쪽으로 기울어도
OFF 라는 것, 그리고 engine lock 을 잡으므로 **두 시장이 닫힌 창(KST 05:00~09:00)**
에서 한다는 것을 적었다.

```
고친 뒤: run2 exit=1  "매니페스트 파일을 만들지 못했다 — 이미 있으면 덮어쓰지 않는다: … file exists"
         cmp(살아 있는 파일, 골든) exit=0   ← 건드리지 않았다
```

### P1-3 — **M3 = CAUGHT 는 틀린 기록이었다** (확인함, 고침)

가장 무겁다. 위 표의 M3 행은 정렬 삭제가 잡힌다고 적었으나 잡히지 않는다. 기전은
단순하다: **저작 경로를 실행하는 시험이 하나도 없었다.**

```
rg EncodeProductionFamilyActivation → 참조 3곳
  · 정의 (production_family_activation.go)
  · tools/a112-family-activation/main.go        ← 유일한 호출
  · 인코더 가드 시험                            ← AST 식별자로 셀 뿐, 부르지 않는다
tools/a112-family-activation/ 에 _test.go 없음
```

골든 시험은 **커밋된 바이트를 로더가 받아들이는지**만 잰다. 그것은 로더 직렬화기의
드리프트 검출기일 뿐, `body()` 의 서술자 유도와 정렬은 어느 시험도 실행하지 않는다.
정렬을 지우고 스위트를 돌리면 전부 초록이다(리뷰어가 복사본에서 확인했고, 나도
치환 개수를 세어 확인한 뒤 재현했다 — 같은 입력 5회가 서로 다른 digest 셋을 냈다).

실무 결과는 활성화가 아니라 **감사 가능성**이다. 도구가 방금 쓴 바이트의 digest 를
출력하므로 무작위 순서의 매니페스트도 켜지기는 한다. 잃는 것은 "배포된 매니페스트의
digest 를 그 입력으로부터 다시 유도하는" 능력, 즉 사후에 "이 핀이 무엇을 승인했는가"
를 답하는 길이다.

고친 것 — 저작 경로에 시험 둘을 심었다. 둘은 **서로 다른 결함**을 잡는다.

| 시험 | 무엇을 재나 |
|---|---|
| `TestTheAuthoringEncoderStillProducesTheCommittedGoldenBytes` | 도구가 지금도 커밋된 골든 바이트를 낸다 (저작 ↔ 골든 드리프트) |
| `TestTheAuthoringEncoderOrdersDescriptorsTheSameWayEveryRun` | 같은 문서가 64회 모두 같은 바이트 (map 순회 의존) |

기대값은 실행 중인 시스템에서 읽지 않는다. 문서 값을 손으로 적고 결과를 **커밋된
파일**과 겨룬다. 반대로 하면(도구를 돌려 그 출력을 기대값으로) 무엇이든 통과한다 —
8.7.1 이 그렇게 충족 불가능한 결속을 초록으로 통과시켰다.

64회인 이유: 네 서술자의 순열은 24가지라 1회로는 **24분의 1 확률로 우연히 초록**이
된다. 실제로 M3 재실행에서 골든 등식 시험은 그 우연으로 통과했고 순서 시험만
빨개졌다 — 시험을 둘로 나눈 근거가 측정으로 나왔다.

### 반증 셋을 더 돌렸다

| # | 변이 | 결과 |
|---|---|---|
| M3(재실행) | 서술자 정렬 삭제 | **CAUGHT** — `…OrdersDescriptorsTheSameWayEveryRun` (골든 등식은 24분의 1로 통과) |
| M8 | `body()` 가 `ApprovedAt` 에 `IssuedAt` 을 싣는다 (필드 오배선) | CAUGHT — `…StillProducesTheCommittedGoldenBytes` |
| M9 | 가드가 점 import 를 안 본다 | CAUGHT — 점 import 하위 시험만 |
| M10 | 가드가 빈 import 를 만나면 조기 반환 | CAUGHT — 빈-import-우선 하위 시험만 |

M8 은 골든 등식이 **저작 경로**를 재는지 확인한다(정렬만 잡으면 그 시험이 필요
없을 수 있으므로). M9·M10 은 각자 자기 행만 빨갛게 만든다 — 축별로 갈라져 있다.

### P2 넷도 처리했다

- **P2-1 (가드 우회 둘).** `resolvedFamilyActivationRouterName` 이 import spec 을
  **하나만** 보고 `.`·`_` 를 만나면 "안 부른다" 로 답했다. 그래서 점 import(맨
  이름으로 부른다)와 `_` spec 을 진짜 별칭보다 먼저 적는 파일이 그냥 통과했다.
  리뷰어가 셋 다 컴파일해 보였다. 세는 범위가 입력의 함수였던 것이고, M7·5.5 와
  같은 기전이다. spec 전부를 보고 이름 **집합**과 점 여부를 함께 내도록 고쳤다.
  반증값은 일회성 실험으로 두지 않고 `TestTheEncoderGuardCountsEveryWayToNameThePackage`
  의 네 표본으로 **커밋했다**. 짝이 되는
  `…CountsNothingWhenThePackageIsNotCalled` 가 "무엇이든 센다" 판본을 배제한다.
- **P2-2 (사라진 서명을 계속 주장하는 문장들).** 코드 주석 9곳, 시험 이름 4개를
  고쳤다. 이름 넷 중 둘은 `Makefile` 의 `RACE_ENGINE_TESTS` 목록에 있어서, 조용히
  게이트에서 빠질 수 있는 자리다 — 이름을 바꾼 뒤 **실제로 도는 개수**를 셌다:
  `-run` 목록 18개, `-v` 실행 결과 최상위 `--- PASS` **18개**, 둘 다 이름이 바뀐
  시험을 포함한다. (경로 권한·위험 정책 매니페스트를 "서명된" 이라 부르는 문장은
  **참이므로 그대로 뒀다** — 그 둘은 여전히 ed25519 + digest pin 이다.)
- **P2-3 (기록이 빠뜨린 교환 하나).** 위 결정 61 절은 잃은 것을 하나만 적었다
  (config+env 를 다 쓸 수 있는 에이전트가 런타임을 켤 수 있다). 리뷰어가 두 번째를
  찾았고 그것이 맞다: **승인자의 부인방지**다. `key_id` 와 서명이 매니페스트를 특정
  승인 키에 묶었는데, 이제 `actor` 는 인증되지 않은 바이트 안의 자유 문자열이고
  도구도 런타임도 검사하지 않는다. 사후에 "누가 이 활성화를 승인했는가" 는 이제
  전적으로 **핀을 넣은 배포 이력**이 답한다. `docs/operations.md` 에 그대로 적었다.
  (리뷰어가 함께 확인한 것: 서명이 결속하던 **필드**는 하나도 안 잃었다. 제거된
  검사는 `body.SignatureAlgorithm`·`body.KeyID` 둘뿐이고, 나머지는 digest 핀 +
  `bytes.Equal(canonical, data)` 가 그대로 덮는다.)
- **P2-4 (문서 빈칸).** `-revoked` 를 적었다 — 다만 **오늘 그 이유가 운영자에게
  보이지 않는다**는 것을 함께 적었다. 로더는 폐기를 별도 sentinel 로 구별하지만
  `familyGateFor` 가 모든 오류를 빈 gate 하나로 접는다(`strategy_family_activation.go:108`
  에서 확인). 그래서 급히 끄는 정석은 `-revoked` 매니페스트가 아니라 **핀을 빼고
  재시작**이라고 적었다. 도구를 service UID 로 돌려야 한다는 것도 적었다.

### 리뷰어가 CLEAN 으로 확인한 것 (침묵과 구별하려고 남긴다)

골든 무결성(파일 SHA-256 == 손으로 적은 리터럴, 시험이 재생성하지 않음), 도구 ↔ 골든
왕복 바이트 동일, 골든의 반증 가능성(필드 선언 순서를 바꾸면 골든 시험 **하나만**
빨개진다), 손으로 만든 바이트의 로더 우회 없음(수용 집합 == 이 빌드 marshaller 로
왕복하는 바이트, 도구가 못 내는 유일한 형태는 `desired=ON, effective=OFF` 로 안전
방향), 시각 왕복(RFC3339Nano · UTC 강제), 서술자 정렬의 전순서성(시장마다 가족 넷이
서로 다름 — `production_test.go:315` 가 고정), **생산 동작 변화 0**(`~/.config/tossctl`
에 `strategy-*` 0건, 배포 override 에 `FAMILY_ACTIVATION` 0건, 그리고 로더가
`readProductionRouteFile` **앞**에서 핀 형식으로 단락하므로 새 I/O 도 없다),
보호 하한이 주문 경로에 결속됨(`strategy_dispatch_cycle.go:109`), 가드 워크의 종료성.

### 수정 뒤 검증 — 전부 다시 돌렸다

| 명령 | 결과 |
|---|---|
| `make lint` | exit 0 |
| `gofmt -l internal/ tools/ cmd/` | 0 바이트, exit 0 |
| `make test` (무태그) | exit 0 · **99 ok / 0 FAIL** |
| `make test-seams` (태그) | exit 0 · **100 ok / 0 FAIL** |
| `make test-race` | exit 0 · 8 패키지 ok · 엔진 목록 18개 중 **18개 실행 확인** |
| 도구 왕복 | `cmp(도구 출력, 커밋된 골든)` exit 0 · digest == 핀 |
| 도구 재실행 | exit 1, 원인을 말하고 살아 있는 파일을 안 건드림 |
| 반증 M3·M8·M9·M10 | 전부 CAUGHT, 각자 다른 행 |

### 이 리뷰가 바꾸지 않은 것

리뷰어도 **P0 은 0**이라고 판정했고 나도 재확인했다. 생산 동작 변화는 여전히 0 이다
— 생산에 매니페스트가 없고 핀도 없다. 세 P1 은 전부 "산출물이 결함" 이지 "런타임이
위험" 이 아니었다. US 골든이 없다는 것과 8.7.2 가 열려 있다는 것도 그대로다.
