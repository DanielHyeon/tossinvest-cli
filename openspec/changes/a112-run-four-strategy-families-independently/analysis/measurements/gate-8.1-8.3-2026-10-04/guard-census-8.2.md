# a112 8.2 정적 · 의존 가드 census (Explore 에이전트, 읽기 전용 · go list 실측, HEAD fc0911f9, 2026-10-04) — 원문

I did not read or search anything under ~/.codex. I edited no files and ran no git write commands and no `go test`. To measure closures I ran `go list -deps [-test] [-tags tossos_testseams]`, which is read-only apart from the Go build cache.

# Census for a112 task 8.2: dependency and static guards (HEAD fc0911f9)

**Bottom line:** guard coverage is strong for strategyworker and strategyhandoff and transitive for most lane packages. strategyproposal and strategyprojection have no capability guard. officialbars passes its own guard but its import closure does contain the official order client. For (e) there is no repository-wide guard, and none of these packages uses the existing `internal/testenv` transport guard.

Requirement source: `openspec/changes/a112-run-four-strategy-families-independently/tasks.md:582` (8.2), `design.md:94` ("8개 worker 어디에도 broker mutator, writable journal, Guardian issuer, activation writer 또는 toggle writer를 주입하지 않는다") and `tasks.md:186,199`.

## 1. Packages in scope (all exist under /mnt/D/Axipient/workspace/TossOS/internal/)

breakoutlane, continuationlane, reversallane, weeklyvaluelane, strategyflow, strategyworker, strategyevidence, strategyproposal, strategycoordinator, strategyarbiter, strategyhandoff, officialbars, strategyrouter.

- **Added:** `internal/strategyprojection`. It is new in a112 (diff from base aeeb209e) and named in design/tasks. Its production imports are standard library only.
- **Adjacent, not added:**
  - `internal/officialfx` (pre-a112, inside the worker closure).
  - `internal/strategyruntime` (pre-a112; has `dependency_test.go:12 TestPackageHasNoBrokerLiveTransportToggleOrRuntimeWriter`, which scans for tokens including "tossinvest.com"/"toss.im").
  - `internal/strategyprojectionrpc` (a108 transport with additive a112 fields).
  - Engine host `internal/app/engine/strategy_lane_runtime.go`.

Where the capabilities live in this module:
- **(a) WTS/broker mutator:** `internal/official` (`orders_write.go:182/222/250` PlaceOrder/CancelOrder/ModifyOrder), `internal/client`, `internal/hybrid`, `internal/trading`, `internal/orderintent`, `internal/protectionofficial`, `internal/verifylive`.
- **(b) writable journal:** `internal/journal`.
- **(c) Guardian issuer:** `internal/execgw` (`guardian.go`) and `internal/strategydispatch`.
- **(d) activation/toggle writer:** `internal/ops`, `internal/config`, and the activation byte encoder `strategyrouter.EncodeProductionFamilyActivation` (`strategyrouter/production_family_activation.go:281`).

## 2. Guard inventory

None of the guard files has a `//go:build` line, so all of them run untagged. The go-list-based walks run `go list` without `-tags`, so tagged test files are never walked. The parser and `os.ReadDir` scans read every .go file, so they do see tagged `*_testseam.go` production files.

| ID | Location (path:line, function) | What it actually checks | Scope |
|---|---|---|---|
| G1 | `breakoutlane/guard_test.go:87 TestResolvedProductionClosureAllowlist` | Allow-list of 7 stdlib imports (crypto/sha256, encoding/hex, errors, math/bits, strconv, strings, unicode/utf8); also bans FuncDecls Apply/NewMachine/NewQuote/NewFX | Direct, production files only. Effectively transitive, since no module package can be imported. |
| G2 | `continuationlane/dependency_test.go:12 TestPurePackageHasNoBrokerJournalExitOrToggleAuthorityDependency` | Deny substrings /broker,/journal,/exit,/gateway,/operating,/toggle,/registry | Direct, production only. Does not match /official, /execgw, /client, /hybrid, /trading, /config, /ops. |
| G3 | `reversallane/dependency_test.go:12 TestPurePackageHasNoBrokerJournalExitRegistryOrToggleAuthority` and `:36 TestProductionFilesContainNoMutationOrExitDecisionTypes` | G2's list plus /strategyengine; then a token scan for PlaceOrder(, CancelOrder(, type JournalWriter, type ToggleWriter, MintRiskCap, … | Direct, production only |
| G4 | `weeklyvaluelane/dependency_test.go:12 TestPurePackageHasNoSourceAPIBrokerJournalExitOrToggleAuthority` | Deny net/http,/broker,/journal,/exit,/gateway,/operating,/toggle,/runtime,/opendart,/edgar | Direct, production only |
| G5 | `strategyflow/dependency_test.go:11 TestPureFlowAndLaneDependencyClosureHasNoMutationAuthority` | `go list -json -deps .` with deny list /internal/journal, execgw, trading, official, config, app/engine, console, httpapi | Transitive, production only (no -test). Covers the 4 lanes, strategyrouter and strategyevidence. Does not list client, hybrid, ops, orderintent. |
| G6a | `strategyworker/dependency_closure_test.go:31 TestTheWorkerImportsNothingOutsideItsAllowedClosure` | Allow-list: breakoutlane, clock, continuationlane, reversallane, strategyarbiter, strategycoordinator, strategyrouter, weeklyvaluelane, plus 7 named stdlib packages | Direct, production only |
| G6b | `strategyworker/dependency_closure_test.go:100 TestTheWorkerTransitivelyReachesNoMutationCapability` (subtests -deps, -deps-test) | Deny official, hybrid, client, trading, orderintent, execgw, journal, ops, protectionofficial, verifylive, config, app; positive control requires strategyflow, candidate, domain | Transitive. Production and the worker's own tests, untagged. The production closure includes breakoutlane, continuationlane, reversallane, weeklyvaluelane, strategyflow, strategyarbiter, strategycoordinator, strategyrouter, strategyevidence, officialfx and clock, so it is the strongest guard for those. |
| G7a | `strategyhandoff/dependency_closure_test.go:25 TestTheHandoffSeamImportsNothingOutsideItsAllowedClosure` | Allow-list: errors, strings, strategyflow | Direct, production only |
| G7b | `strategyhandoff/dependency_closure_test.go:71 TestTheHandoffSeamTransitivelyReachesNoMutationCapability` | Same deny list as G6b; positive control | Transitive, -deps and -deps-test, untagged |
| G7c | `strategyhandoff/dependency_closure_test.go:179 TestOnlyTheEngineImportsThisSeam` | Census of who imports strategyhandoff: only internal/app/engine may | Uses three `go list` universes (cmd/tossctl -deps; `./...` -deps -test; and the same with `-tags tossos_testseams`) |
| G8 | `strategycoordinator/dependency_closure_test.go:14 TestTheCoordinatorCannotReachAnyMutationCapability` | Allow-list: strategyarbiter, strategyrouter, sort, sync, time | Direct, production only |
| G9 | `strategyarbiter/dependency_closure_test.go:13 TestTheArbiterCannotReachAnyMutationCapability` | Allow-list: strategyflow, strategyrouter, time | Direct, production only |
| G10a | `strategyevidence/dormant_scope_test.go:103 TestStrategyEvidenceImportClosureReachesNoMutationPath` | Its own recursive parser walk over module-internal production files (all build tags; stdlib matched by name only). Deny fragments net/http,/broker,/dispatch,/execgw,/guardian,/operating,/runtime,/toggle,/journal,/order. Positive control on internal/obs. | Transitive, production only. Misses /official, /client, /hybrid, /trading, /config, /ops. |
| G10b | `strategyevidence/dormant_scope_test.go:232 TestDormantSnapshotReadIsSelectOnlyAcrossTheWholeCallChain`; `consumer_static_test.go:12 TestDormantSnapshotReadPortIsStructurallySelectOnly`; `breakout_series_test.go:788 TestSealBarSeriesIsStructurallySelectOnly`; `readonly_test.go:13 TestOpenReadOnlyReplaysButCannotCreateSnapshotOrAppend`, `:61 TestOpenReadOnlyRefusesWrongModeAndSymlink` | The read paths are SELECT-only (AST: no Exec/Begin/INSERT/UPDATE/…), and the read-only handle is checked at runtime | Note: the package also owns a writable append-only evidence store (`store.go:108-370`, ExecContext INSERT). That is by design and is not the journal. |
| G11 | `officialbars/guard_test.go:57 TestProductionImportsStayInsideTheAllowlist` | Non-stdlib allow-list {clock, official, scheduler, strategyevidence}; denies net/http | Direct, production only. It explicitly allows internal/official. |
| G12a | `strategyrouter/dependency_test.go:13 TestPackageHasNoMutationAuthorityOrRuntimeDependency` | Token scan for PlaceOrder(, CancelOrder(, JournalWriter, BrokerClient, ToggleWriter, ActivationWriter, CampaignWriter, OwnerWriter; import deny /internal/scheduler, /internal/journal, /internal/app, /internal/httpapi | Direct, production only, including tagged testseam files |
| G12b | `strategyrouter/production_family_activation_encoder_guard_test.go:45 TestOnlyTheAuthoringToolCanBuildActivationBytes`, `:188`, `:238` (falsification controls) | Repo-wide walk of non-test .go files. Only `production_family_activation.go` and `tools/a112-family-activation/main.go` may reference EncodeProductionFamilyActivation / FamilyActivationDocument (alias-resolved, per symbol). | Repo-wide, production only |
| G13 | `app/engine/a112_lane_runtime_test.go:331 TestTheFamilyLaneStepCarriesNothingButItsLaneAndTheSignedPromotion` and `:193 TestOnlyThePackageLevelStepEverRunsInsideALane` | AST pin: `strategyFamilyLaneStep` takes exactly (*strategyworker.Lane, strategyrouter.FamilyActivation) and names no capability identifier | Engine-side host boundary |
| G14 | `app/engine/deps_test.go:28 TestEngineDependencyGraphExcludesWTSMutators` | The engine's transitive closure excludes internal/client, internal/hybrid, internal/app | Production only |
| G15 | `testenv/static_test.go:165 TestNoProductionCodeHardcodesATestTransport`, `:192 TestTestenvItselfIsNotImportedByProductionCode`; `testenv/testenv_test.go:28,69,97,121,140,161,188` | Self-tests of the hostname-guard transport | Opt-in library (see section 4) |

Not capability guards: strategyproposal `a112_breakout_wall_test.go:60 TestNoNonTestCodeBuildsABreakoutLaneInputAroundTheWall` and strategyprojection `a112_no_metric_emitter_test.go` (metric imports only).

## 3. Matrix

Notation: d = direct imports, T = transitive closure, P = production files only, +t = also covers test imports. Cells say "absent (measured)" where nothing guards the property but I checked that the package is absent.

| Package | (a) WTS/broker mutator | (b) writable journal | (c) Guardian issuer | (d) activation/toggle writer | (e) tests→live host |
|---|---|---|---|---|---|
| breakoutlane | G1 (d≈T, P); G6b, G5 (T, P) | G1; G6b, G5 | G1; G6b, G5 | G1; G6b (ops, config); G12b | GAP. Test imports are stdlib only; no net/http in the test closure (measured). |
| continuationlane | G2 partial (d, P; no /official, /execgw); G6b, G5 (T, P) | G2 (d); G6b, G5 | G6b, G5 only (G2's "/gateway" does not match "execgw") | G2 /toggle (d); G6b; G12b | GAP. Tagged `production_proposal_test.go` is not walked. |
| reversallane | G3 (d + tokens, P); G6b, G5 | G3; G6b, G5 | G6b, G5 | G3 (/toggle, ToggleWriter); G6b; G12b | GAP |
| weeklyvaluelane | G4 (d, P); G6b, G5 | G4; G6b, G5 | G6b, G5 | G4; G6b; G12b | GAP. G4 denies net/http in production files only. |
| strategyflow | G5 (T, P; no client/hybrid); G6b (T, P) | G5; G6b | G5; G6b | G5 (config); G6b; G12b | GAP. Tagged `production_integration_test.go` and the testseam files are not walked by `go list`. |
| strategyworker | G6a (d, P); G6b (T, P+t, untagged) | G6a; G6b | G6a; G6b | G6a; G6b; G12b; G13 | Partial: G6b -deps-test excludes Toss client packages from the test binary, but net/http is not denied. The 5 tagged test files are not walked; their tagged closure is clean (measured). |
| strategyevidence | G10a (T, P; misses /official, /client, /hybrid, /trading); G6b (T, P) | G10a (/journal); G6b; G10b (read-path SELECT-only) | G10a (/execgw, /guardian); G6b | G10a (/toggle); G6b (ops, config) | Partial: G10a denies net/http in the production closure only. Tests: GAP. |
| strategyproposal | GAP. Nothing in a guard; absent from closure (measured). | GAP. Imports internal/journal directly (`production.go:334-337` uses `journal.OpenReadOnly`); no guard pins the read-only use. | GAP (absent, measured) | GAP (absent from closure, measured); G12b only | GAP |
| strategycoordinator | G8 (d, P); G6b (T, P) | G8; G6b. Tests: the tagged external test `receipt_contract_test.go:9` imports strategyproposal, so internal/journal is in the tagged test closure (measured) and nothing catches it. | G8; G6b | G8; G6b; G12b | GAP |
| strategyarbiter | G9 (d, P); G6b, G5 (T, P) | G9; G6b | G9; G6b | G9; G6b; G12b | GAP |
| strategyhandoff | G7a (d, P); G7b (T, P+t); G7c | G7a; G7b | G7a; G7b | G7a; G7b; G12b | Partial, like strategyworker (G7b -deps-test) |
| officialbars | Not a missing guard: the closure contains internal/official (PlaceOrder/Cancel/Modify), trading, orderintent and config (measured), and G11 allow-lists official. Mitigation is narrow read interfaces at `producer.go:30` and `quote.go:37-38`. | G11 catches a direct import only; nothing transitive | G11 direct only | Closure contains internal/config; G11 direct only; G12b | GAP. Tests link net/http, httptest and the official client; the only HTTP test points it at httptest (`producer_test.go:1412-1431`, WithBaseURL(server.URL)), but nothing enforces that. |
| strategyrouter | G12a (tokens + d, P; no official/execgw/client deny); G6b, G5 (T, P) | G12a (/internal/journal, d, P); G6b; G5. Production opens the journal DB through its own sqlite with `mode=ro&_pragma=query_only(1)` and a ReadOnly tx (`production.go:620-626`); no test pins the read-only DSN. Tagged external test `a127_real_journal_test.go:21` calls `journal.Open` (writable) in the test closure. | G6b, G5 only | The package defines the activation encoder; G12b restricts callers repo-wide. G12a bans ToggleWriter/ActivationWriter tokens. File writes only in tagged `production_route_manifest_testseam.go:74`. | GAP (no net/http in test closure, measured) |
| strategyprojection | GAP (stdlib-only production, measured) | GAP | GAP | GAP | GAP |

Measured closures (no committed guard):
- `-deps -test` untagged: forbidden packages appear only in strategyproposal (journal) and officialbars (config, official, orderintent, trading).
- With `-tags tossos_testseams`: also strategycoordinator (journal) and strategyrouter.test (journal).
- net/http appears only in the officialbars test closure.

## 4. (e) Tests reaching live hostnames

- **No repository-wide enforcement.** `internal/testenv/testenv.go:41-45` defines real hosts (tossinvest.com, toss.im, tossbank.com). `Guard`/`GuardedClient` (`:113`, `:134`) block every request to them, with a louder error for POST/PUT/PATCH/DELETE, and `InstallGuard` (`:149`) swaps `http.DefaultTransport`. It is opt-in: no package has a `TestMain` calling `InstallGuard`, and the `TestMain`s that exist are in `cmd/tossctl/help_convention_test.go`, `internal/console/fake_broker_test.go` and `internal/verifylive/{fake_broker,static}_test.go`.
- **Who imports testenv:** only cmd/tossctl tests, `internal/app/engine/engine_test.go`, console, verifylive and testenv itself. **None of the 14 in-scope packages imports it.**
- `testenv/static_test.go:165,192` only stop production code from shipping httptest/testenv. They say nothing about tests reaching live hosts.
- **Per-package proxies today:**
  - G6b/G7b `-deps-test` keep internal/official, client and hybrid out of the strategyworker and strategyhandoff test binaries, but they do not deny net/http and do not walk tagged tests.
  - G4, G10a and G11 deny net/http in production files only.
  - `internal/app/engine/wts_isolation_test.go:123 TestEngineMutationMatrixNeverReachesWTS` covers the engine, which is out of scope.
- **Coverage of in-scope packages:** none has an enforced test-side hostname guard. 13 of 14 have no net/http in their test closures (measured, untagged and tagged) but no test asserts it. officialbars is the one package whose tests can dial out.

## 5. Commands that run every guard above

None of these needs `-tags tossos_testseams`. G7c runs the tagged universe internally. Adding the tag also runs fine, but the `go list` walks stay untagged.
```
go test ./internal/breakoutlane -run '^(TestResolvedProductionClosureAllowlist|TestPublicSurfaceCannotAssertEventsOrForgeMachine)$'
go test ./internal/continuationlane -run '^TestPurePackageHasNoBrokerJournalExitOrToggleAuthorityDependency$'
go test ./internal/reversallane -run '^(TestPurePackageHasNoBrokerJournalExitRegistryOrToggleAuthority|TestProductionFilesContainNoMutationOrExitDecisionTypes)$'
go test ./internal/weeklyvaluelane -run '^TestPurePackageHasNoSourceAPIBrokerJournalExitOrToggleAuthority$'
go test ./internal/strategyflow -run '^TestPureFlowAndLaneDependencyClosureHasNoMutationAuthority$'
go test ./internal/strategyworker -run '^(TestTheWorkerImportsNothingOutsideItsAllowedClosure|TestTheWorkerTransitivelyReachesNoMutationCapability)$'
go test ./internal/strategyevidence -run '^(TestStrategyEvidenceImportClosureReachesNoMutationPath|TestDormantSnapshotReadIsSelectOnlyAcrossTheWholeCallChain|TestDormantSnapshotReadPortIsStructurallySelectOnly|TestSealBarSeriesIsStructurallySelectOnly|TestOpenReadOnlyReplaysButCannotCreateSnapshotOrAppend|TestOpenReadOnlyRefusesWrongModeAndSymlink)$'
go test ./internal/strategycoordinator -run '^TestTheCoordinatorCannotReachAnyMutationCapability$'
go test ./internal/strategyarbiter -run '^TestTheArbiterCannotReachAnyMutationCapability$'
go test ./internal/strategyhandoff -run '^(TestTheHandoffSeamImportsNothingOutsideItsAllowedClosure|TestTheHandoffSeamTransitivelyReachesNoMutationCapability|TestOnlyTheEngineImportsThisSeam)$'
go test ./internal/officialbars -run '^TestProductionImportsStayInsideTheAllowlist$'
go test ./internal/strategyrouter -run '^(TestPackageHasNoMutationAuthorityOrRuntimeDependency|TestOnlyTheAuthoringToolCanBuildActivationBytes|TestTheEncoderGuardCountsEveryWayToNameThePackage|TestTheEncoderGuardCountsNothingWhenThePackageIsNotCalled|TestExternalAPIExposesNoAuthorityMintingConstructor)$'
go test ./internal/app/engine -run '^(TestTheFamilyLaneStepCarriesNothingButItsLaneAndTheSignedPromotion|TestOnlyThePackageLevelStepEverRunsInsideALane|TestEngineDependencyGraphExcludesWTSMutators)$'
go test ./internal/testenv -run '^(TestNoProductionCodeHardcodesATestTransport|TestTestenvItselfIsNotImportedByProductionCode|TestIsRealHost|TestGuard.*|TestInstallGuardCoversTheDefaultTransport)$'
```
Optional context: `go test ./internal/strategyruntime -run '^TestPackageHasNoBrokerLiveTransportToggleOrRuntimeWriter$'`.

## 6. Gaps, each with a suggestion

1. **strategyproposal, (a)–(d): no guard; it imports internal/journal.** Add a `dependency_closure_test.go` with a direct allow-list, a `-deps`/`-deps-test` deny list like G6b minus journal, and an AST pin that only `journal.OpenReadOnly` / read-only types are referenced (no `journal.Open`/writer methods).
2. **strategyprojection, (a)–(e): no guard.** Add a stdlib-only direct allow-list like G1.
3. **officialbars (a)/(d): the closure contains the official order client, trading, orderintent and config.** Either record a reviewed exception that only the `StrictMinuteCandles`/`StrictOrderbookTop`/`StrictLastPrice` interfaces are used (AST pin that no `*official.Client` method other than those three is named), or move the types to a read-only subpackage and deny internal/official. Also add a transitive deny for journal, execgw and ops.
4. **(e) everywhere: no test-side hostname guard.** Add a `TestMain` with `testenv.InstallGuard` to officialbars at minimum. Better: one guard (like G7c's universes) that runs `go list -deps -test` with and without `-tags tossos_testseams` over the 14 packages and asserts net/http is absent except officialbars, where it requires testenv.
5. **Walks skip tagged test files.** G6b, G7b and G5 call `go list` without `-tags tossos_testseams`, so tagged tests in strategyworker, strategycoordinator, strategyrouter, strategyflow and the lanes are never walked; today the coordinator and router tagged tests pull in internal/journal. Add a tagged `-deps -test` subtest, and decide whether journal is acceptable in test binaries (it is a test-only real-journal fixture).
6. **Test imports of the lane, arbiter, coordinator, flow and router packages are not walked by any guard.** G6b's `-deps-test` covers only strategyworker's own tests. Extend G6b's walk to `-test` over each package in the closure.
7. **strategyflow deny list (G5) is narrower than G6b** (missing client, hybrid, ops, orderintent, protectionofficial, verifylive). Unify on G6b's list.
8. **G2, G3, G4, G10a and G12a use substring deny lists that miss /official, /execgw, /client, /hybrid, /trading, /config and /ops.** Today they are backstopped only by G6b being transitive. Switch them to allow-lists, or share G6b's list.
9. **strategyrouter (b): read-only journal DSN not pinned.** Add an AST/string pin on `production.go:620-626` (`mode=ro`, `query_only(1)`, `ReadOnly: true`) and a deny on ExecContext/INSERT in production files, like G10b.
10. **strategyproposal (b): same issue.** Pin the use of `journal.OpenReadOnly` (`production.go:337`) and forbid `journal.Open`.
11. **(c) Guardian issuer is never named directly outside G10a's /guardian fragment.** Add internal/strategydispatch, alongside internal/execgw, to the shared deny list.

