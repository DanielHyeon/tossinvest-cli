# a155 independent implementation review

2026-10-18. Final verdict: PASS for offline implementation. Full repository gate and lifecycle remain pending; no actual capture, runtime latency, authenticated source or trading approval is established.

## Scope

Read canonical implementation/geometry contracts, research-first evidence, all new production files in internal/strategyobserve and offline replay CLI, tests and review receipts. Applied code-reviewer universal/Go guidance. Memory quick search returned no task-specific evidence used. No production edits, account/settings changes, network collection or a157 work performed.

## Initial HOLD findings and resolution

1. Currency identity: US features declaring KRW with USD geometry yielded known USD risk. Independent isolated Go overlay reproduced failure without modifying production. Calculate now checks feature and geometry currency against market-derived currency. Regression passes.
2. Optional quantity: omitted quantity rejected otherwise valid geometry. Independent overlay reproduced `invalid quantity`. Quantity now optional pointer, omitted means risk unknown while independent geometry/RR remains available; explicit zero/null invalid, positive bounded quantity copied to avoid caller mutation. Regression passes. Manager also clarified supplied-bar missing basis must reject; implementation now requires explicit unadjusted for each supplied bar while absent bars do not block independent geometry.
3. P1 recovery durability: readable complete records after append-fsync failure were accepted as duplicates without ever completing a successful sync. Final recovery validates all bytes/checkpoint first, then syncs ledger; matching checkpoint gets file+directory sync, lagging/missing checkpoint is durably rewritten before Open succeeds. Injected recovery-sync and recovery-directory-sync failures refuse recovery; successful restart precedes duplicate success. Initial failing behavioral receipt is preserved in `/tmp/tossos-a155-recovery-red.log`.

Other review-stream fixes inspected: single physical newline enforcement and nonblocking open before regular-file validation prevent embedded-record boundary ambiguity and FIFO-open blocking. Original evidence not repaired/truncated.

## Independent verification

Final direct execution: `go test -race -count=1 ./internal/strategyobserve ./tools/a155-observe-versioned-strategy-decisions` PASS; package times 1.327s and 1.035s. This includes numerical/US, missing/invalid quantity, basis/currency, immutable payload, queue/close races, blocked/failing writer, physical size, lock/path/type, checkpoint/crash and final recovery barriers.

Writer actual CLI receipt `analysis/verification/reviewed-cli.log` inspected: first durable=1, repeat duplicate=1, input SHA unchanged, one ledger record, USD price risk 70, US features and Parker RR unsupported. CLI receipt is writer-produced, distinct from independent test execution.

## Contract assessment

- Pure feature/RR dependencies retained, no broker/journal/reservation/coordinator dependency or runtime producer introduced. US hypothetical risk is separate from unsupported US features/RR. Exact rational risk and provenance retained.
- Queue copy owns encoded bytes; capacity 256, encoded limit includes newline, full drops newest. Mutex covers admission/closure/counts but never callback disk I/O. CLI closes input and waits for writer completion; errors leave failed/pending visible. This proves architectural isolation only.
- Dedicated directory inode lock and nofollow traversal/file opens isolate output. Hardlinks/nonregular files and foreign directory entries rejected. Linux-only implementation fails explicitly elsewhere. Invalid recovery preserves original evidence; checkpoint temp residue fails closed rather than being silently removed.
- Ledger identity/content conflict is rejected; lagging checkpoint is validated against its exact prefix and then advanced only after sync. Recovery/append bounded by record and byte limits. No truncation or migration of prior evidence.
- Input CLI validates all supplied lines before creating output and returns failure for invalid input, storage failure or drops. This bounded fixture replay does not promise lossless high-rate collection.

## Reviewed SHA-256

- `internal/strategyobserve/observe.go`: `35d3c8b2c987cd1b25ccd20211fd61e6a805f564572486433e2598dfe65b4089`
- `internal/strategyobserve/observe_test.go`: `fbca23654afeae079d1d003ca9ace52d1b96ee280c3b97e1b112c7a71f8fa94c`
- `internal/strategyobserve/queue.go`: `5776f5cc20c1f38947afa8adf228c45a65c26da044caf7a6d738274faa37359b`
- `internal/strategyobserve/store_linux.go`: `b3ce05454bc28fbd2ed3612ccfc71370627601b54c8c015b7ca21352a4d12844`
- `internal/strategyobserve/store_linux_test.go`: `32f02c9de3b565791387efbc1333bbd4bd3905cfe91706075c0fe1f6517b522e`
- `internal/strategyobserve/store_other.go`: `39d6a19b695136de175a81a9da0b18ced408cb2e4cf9a2208c44cbd6ce434181`
- `tools/a155-observe-versioned-strategy-decisions/README.md`: `3c25c8d41df510403d54725e54a8aad1b58f3f62369eae8ca7be38f4340f6791`
- `tools/a155-observe-versioned-strategy-decisions/main.go`: `49f37b043b97a36c50bc9db2c1e8032c3f752e3bc195014a6d1495c7dae2cede`
- `tools/a155-observe-versioned-strategy-decisions/main_test.go`: `d7396a14d4038584996fd8b5e37d0d5c153b495b2d4102dd207e010738a4d71d`
- `tools/a155-observe-versioned-strategy-decisions/testdata/synthetic.jsonl`: `765886e02d0d6afb864e18859465fa59c35e1eab262e10604dbbbd803ceb9c2b`

## Gate 1 architecture guard repair — final PASS

First full gate failed the a047 import census; `analysis/verification/gate1-failed.log` remains failure evidence. Reviewed fresh pre-edit AST/query references and test branch map in pre-edit.md, existing wiring_test.go change, and a155_wiring_guard_test.go.

Narrow exception is exact observe.go with five named pure RR references. Alias-aware selector inspection rejects other engine members and dot/blank imports. Inverse census rejects strategyobserve import from runtime, other tools and subpackages, permitting only the exact offline CLI (and its own package directory). Prior importer rules and console descriptor restriction remain. Owning engine directory is now scanned for inverse imports instead of skipped wholesale.

Independent review found a raw-string import bypass in initial repair: quote trimming did not interpret valid Go import literals. Writer preserved behavioral RED in import-guard-red.log; shared strconv.Unquote now normalizes both outer engine import matching and inverse observation import matching. Raw/escaped imports and forbidden selector cases are negative fixtures. Final source reread confirms both uses. Reviewer isolated raw-import overlay ran after the concurrent repair landed and passed; it is GREEN evidence, not an independent RED claim.

Final writer race receipt import-guard-green.log inspected: strategyengine 5.091s, strategyobserve 1.337s, offline CLI 1.044s, all PASS. No runtime production edits in repair. Independent verdict PASS for repaired guard and previously reviewed offline implementation; repeat aggregate gate still required.

Repair hashes:
- `internal/strategyengine/wiring_test.go`: `3c2d6f1e56f95aa30be693de2027a5c19e64ee5071c4d666548e1f4d3e40977a`
- `internal/strategyengine/a155_wiring_guard_test.go`: `d096ad98da3d956371cbdf8e6972d1cae00f89d0916fd09ea3b22f2d4135540c`

## Gate 2 evidence-only review — PASS

Second aggregate stop was missing formal function-logic evidence, not a new runtime failure. Reviewed both AST/FLM/BTM bundles and flm-check2.log. Main guard AST SHA matches current wiring_test.go (3c2d6f1e…); unchanged descriptor helper AST pins base file SHA 6298b6ee03571ea1a53a062dc5ff37d7ea6bff076f606eb6daba4a4ffa8f26d2 and base lines 62–82. Independent extraction confirms helper body byte-identical between f156b66e base and current source. Source/test hashes remain identical to prior reviewed repair.

Branch maps enumerate current guard branches and unchanged helper branches, retain existing descriptor constraints and link focused/negative receipts. They explicitly do not claim every historical error branch was separately induced. flm-check2.log reports two required functions and evidence complete/diff-proven exempt. No checker bypass or source change observed; no redundant test rerun performed. Prior gate failures remain failures. Evidence-only acceptance PASS; aggregate rerun still required.
