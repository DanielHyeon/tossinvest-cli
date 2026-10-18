# a155 gstack implementation review

Current disposition: **PASS — scoped implementation and gate-compatibility repair review**. Initial aggregate gate failed the a047 import guard; repeat aggregate gate is pending. No aggregate gate success or deployment approval claimed.

Codex-host review using installed gstack review workflow and checklist, with security/testing concerns examined directly. No Claude or cross-model execution claimed. Baseline `f156b66eaef9958ef7ef3cbfee5c2c6927fee1a8`. No production edits, independent test execution, live services, telemetry or external disclosure performed by this reviewer. Implementer supplied the test and CLI receipts below. Separate independent review belongs to another reviewer.

## Gate repair follow-up

P2: import paths compared with strings.Trim rather than decoding Go string literals miss legal raw-string and escaped interpreted imports. This can bypass both inverse runtime-import rejection and entry into the leaf selector allowlist. Requested strconv.Unquote and negative fixtures exercising actual guard entry for both forms; no production-code expansion requested. Resolved: shared importMatches uses strconv.Unquote in both actual guard entry paths. Raw-string and escaped literal fixtures exercise runtime import rejection and engine matching followed by forbidden-selector rejection. Alias, dot and blank-import cases remain covered. import-guard-red.log retains the bypass reproduction; final import-guard-green.log records all three affected package race suites passing. No reviewer test execution claimed.

The exact observe.go exception permits only five named RR projection types/constants/function; the inverse observation import guard permits the exact offline CLI and rejects other runtime/tool paths. Existing runtime importer allowlist entries and descriptor restriction remain. Changes are test-contract compatibility only; existing production sources are unchanged. gate1-failed.log (exit 2) is retained; guard-packages.log is superseded by final import-guard-green.log.

## Findings and resolution history

Initial disposition was HOLD; retain these findings rather than replacing failure history with a clean first-pass claim.

1. **P1, physical ledger boundary:** JSON decoder accepted an additional newline although recovery consumes one physical line. A successful append could therefore make its own ledger unreadable. `ValidateLine` now requires exactly one final physical newline within the complete 65,536-byte bound. Behavioral failure retained in `review-red.log`; corrected validation covered by race suite.
2. **P2, FIFO open:** regular-file validation ran only after potentially blocking `open`. `openRegular` now uses `O_NONBLOCK` alongside descriptor-relative `O_NOFOLLOW`, then checks regular type and single link before use. FIFO regression demonstrates prompt refusal.
3. **P1, recovered durability:** valid suffix recovery previously populated duplicate IDs without re-establishing ledger/checkpoint durability after an append-sync failure. Matching checkpoints could also reflect rename followed by failed directory sync. Recovery now validates all evidence first, syncs ledger on every reopen, restores missing/lagging checkpoint through the existing synced-temp/rename/directory-sync path, or syncs matching checkpoint and directory. Any failure refuses Open before duplicate acknowledgement. `recovery-red.log` retains missing-checkpoint reproduction. Current tests inject recovery sync and matching-checkpoint directory sync failures, then require successful reopen and duplicate recognition. No ledger truncation or uncertain append retry added.
4. Independent review boundary findings also checked: omitted quantity remains unknown; explicit null/zero rejected; supplied quantity copied; market and geometry currencies agree; all supplied bar bases must be unadjusted. These are resolved in current source.

## Source and test assessment

Reviewed package, Linux/unsupported-platform storage, queue, CLI, tests and README against the implementation contract. Existing runtime paths remain unchanged. Pure feature/RR reuse does not grant execution authority; US unsupported states and caller-supplied provenance are explicit. Exact planned risk is separate from realized/cost knowledge. Stable IDs use normalized UTC input and fixed calculation/schema versions; payload digests detect corruption, not authenticated provenance.

Producer encodes/copies bounded records; nonblocking 256-entry queue uses drop-newest counters and does not hold its state mutex during disk I/O. Close stops acceptance and drains; writer failure latches and preserves uncertain pending work. CLI validates bounded input before opening output and returns failure for drops/failures/pending work. No broker, journal, reservation, account, network or runtime producer added.

Filesystem checks walk no-follow ancestor descriptors, refuse protected paths and foreign directory contents, lock directory inode, and require single-link regular children. Append and checkpoint boundaries preserve evidence on failure. Ledger recovery checks every envelope/digest, duplicate/conflicting IDs, checkpoint prefix/count/offset, limits and torn tails before writes. Privileged filesystem-owner attacks and nonlocal filesystems remain outside the approved contract.

Evidence inspected: `review-red.log` preserves four behavioral failures; `review-green.log` is superseded by current `recovery-green.log`, whose two package race runs PASS. `recovery-red.log` preserves the durability failure. `recovery-cli.log` shows durable=1 followed by duplicate=1, zero drops/errors/pending, unchanged fixture hash, one ledger line, exact USD risk 70, and unsupported US features/RR. Fault and boundary tests are synthetic; no production latency, power-loss hardware, real observation or profitability claim is established. No unresolved actionable findings remain in this snapshot.

Metadata refresh: reviewed final task preparation wording and Manager acceptance. These do not assert aggregate gate/archive/push completion. Source/test files match the reviewed snapshot.

## Gate 2 evidence-only amendment

Gate 2 failed stage 5 because required formal Function Logic Map bundles were missing. The failure remains in gate2-failed.log. Reviewed two AST/FLM/BTM/risk bundles: modified A047 guard pins current source; unchanged adjacent descriptor helper correctly pins baseline f156b66e, with baseline line numbers and source hash. Both hashes verified against their declared revisions. Branch maps explicitly disclose that historical error branches were not each independently induced; passing suite references are not claimed as exhaustive branch execution. flm-check2.log records evidence complete or diff-proven exempt. Source/tests match prior reviewed manifest; checker was not weakened. Evidence-only amendment PASS; gate 3 remains pending.

## Reviewed manifest

SHA-256 of the exact newline-delimited manifest block: `bcb30156048ffaf2cbd4f0ccd5cbb2dc591b33c5a16cbb204d17f76722b0d727`.

```text
d096ad98da3d956371cbdf8e6972d1cae00f89d0916fd09ea3b22f2d4135540c  internal/strategyengine/a155_wiring_guard_test.go
3c2d6f1e56f95aa30be693de2027a5c19e64ee5071c4d666548e1f4d3e40977a  internal/strategyengine/wiring_test.go
35d3c8b2c987cd1b25ccd20211fd61e6a805f564572486433e2598dfe65b4089  internal/strategyobserve/observe.go
fbca23654afeae079d1d003ca9ace52d1b96ee280c3b97e1b112c7a71f8fa94c  internal/strategyobserve/observe_test.go
5776f5cc20c1f38947afa8adf228c45a65c26da044caf7a6d738274faa37359b  internal/strategyobserve/queue.go
b3ce05454bc28fbd2ed3612ccfc71370627601b54c8c015b7ca21352a4d12844  internal/strategyobserve/store_linux.go
32f02c9de3b565791387efbc1333bbd4bd3905cfe91706075c0fe1f6517b522e  internal/strategyobserve/store_linux_test.go
39d6a19b695136de175a81a9da0b18ced408cb2e4cf9a2208c44cbd6ce434181  internal/strategyobserve/store_other.go
44632251ccf6dea2be8046ddbd4829a8ac248a1f9cff5f0753a8117565f94137  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/function-logic/internal-strategyengine--assertconsoleonlycallsdormantdescriptor/ast.json
d0fcf5f448f1bdd9213cd94830d24d6182e1212af592da1fbfa31c9b76d89f5e  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/function-logic/internal-strategyengine--assertconsoleonlycallsdormantdescriptor/branch-test-map.md
6f1355f91f991037c3e1c9be4a1b3aab7a31243cd18fed3835e8b3e55dfd5f3d  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/function-logic/internal-strategyengine--assertconsoleonlycallsdormantdescriptor/function-logic-map.md
3531368135def2f5d64224e8d9e12743137162f8b7cf64435f315c641c87b01f  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/function-logic/internal-strategyengine--assertconsoleonlycallsdormantdescriptor/risk-pattern-report.md
4208ebb3ac4df66bcf073f4589e470ef6b7575d1b3ee83e827a623c54d9af15a  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/function-logic/internal-strategyengine--testa047shipsnoruntimeorderorexitwiring/ast.json
e7a7e09b6457cc6940b286232395d9fb63340a9b25046c88491e70e530c7d8b0  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/function-logic/internal-strategyengine--testa047shipsnoruntimeorderorexitwiring/branch-test-map.md
cc430394899dde531711c3301a2a58364b66133d4ae0e2faef7176de3402a7ff  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/function-logic/internal-strategyengine--testa047shipsnoruntimeorderorexitwiring/function-logic-map.md
3531368135def2f5d64224e8d9e12743137162f8b7cf64435f315c641c87b01f  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/function-logic/internal-strategyengine--testa047shipsnoruntimeorderorexitwiring/risk-pattern-report.md
765798684515b74a5220c96ea25cbcc29120ecb2b0274f82cc3f68be36884f1c  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/implementation-evidence.md
0d89c0ae8f7654bba4ec6369e5248bcb371d4ea2d9e6945fb10e1a8d57b17974  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/pre-edit.md
73e549af50248bfeca25d1e82c63b01f2bdaef652249b9401678fe31d1fbf537  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/research-first.md
2dbadf4269d0534907d05f039f858dcf70aac65f8d7e0f866e6f3e736b72a636  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/reviews/independent-review.md
c1c641a1fa5a2a464143e43d8eace4e01954c824a5e47fb1e723bfd24d31f98d  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/reviews/manager-verification.md
0e22ea09eedd1e3431641fe38c06f486a9902b735b5a63016badb641f3be7f82  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/cli.log
0e080f254cf5fdb6aa3c9d51322e0f2c5eec323922bd6ebdf67f24e50948741d  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/compile-red.log
15dbc2bde369fd88a7f11cddac73fba90905b46227aa65546d109dcde4b57710  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/early-compile-failure.log
2c1d913569de17acd6b99fc493bf645f383ffaf4ce1fff2650c851bc2cc2527e  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/early-test-failure.log
097a532b21659577814aa41c0d8de804ac4e160b5af80d81bc3b1d12490571c3  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/flm-check.log
097a532b21659577814aa41c0d8de804ac4e160b5af80d81bc3b1d12490571c3  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/flm-check2.log
5ef8a50a8ffc955988e2184442041d7d8e816e0484b6bb778d5daea3e9ebe947  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/focused-race.log
e4520b4d2425f0fb24e83fabad55de597fc3506eacc8d84cd9a07f73d81cc97d  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/gate1-failed.log
ee787e7e8cd8d4381de9dceb83766e4a3537f2b047785f830d584acf234c514a  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/gate2-failed.log
8ab7d943abe8f471e1929fa4bbc339b4d8343a1b23e5fbcec163ae8eb5fa679e  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/guard-packages.log
bc6edf226c86be5c32be8053ea975d0d68d6e04f096cee17088d1b484ed7df53  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/import-guard-green.log
f9c7f640b24d117f068a4e64af7521b67c4009d2ecef500e0c980f24b2934de8  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/import-guard-red.log
4744deeedadd1a483e23e9738776a034c70390c66514e5ca7497e422163cc27a  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/pre-edit-sync.log
2741c5cdd744fa20f9ca46b5180fd181451a790c8881239d679aa137f2103b87  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/recovery-cli.log
6eecf01f2687b6804d3840ad4755e7eba26237e820b0ff084aada29b88dbd613  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/recovery-green.log
8b16dc67d650ed8f3a99589876397db8046509c82371e0c4d5e254d418020a6d  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/recovery-red.log
25aa6b9453290fb5b7f374f1ebbde6d87fdce9d5785279d6854862c7539a21ba  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/review-green.log
bb74ecffbecb227164474c2957840560abe455a0ab8fba7f061449b3772cadc4  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/review-red.log
2266958216f3fbffb38a5a287bdae1885aa14882629b79b2510c06e89e5fde68  openspec/changes/a155-observe-versioned-strategy-decisions/analysis/verification/reviewed-cli.log
1b92bb5eac91075831d8fff896697449d2dac3d03a4df022a2deeba0c170533b  openspec/changes/a155-observe-versioned-strategy-decisions/base-commit.txt
c85658e381424ed4d7fcfcf765abb640a2c115105d4a13e6e98fe671ac03e8f1  openspec/changes/a155-observe-versioned-strategy-decisions/design.md
5d6f84bef695c7449b416f8ecd11c74c8d83c834213928d453466c11562f197a  openspec/changes/a155-observe-versioned-strategy-decisions/proposal.md
687ad8498d5fd8063d71b288ae6742c98aafa71498c0c392dbca76f535a20251  openspec/changes/a155-observe-versioned-strategy-decisions/review.md
060f57eebd911bd140c075b48a12eff491a906c6472a7abbc47ff412d74d8849  openspec/changes/a155-observe-versioned-strategy-decisions/specs/strategy-runtime/spec.md
aa2929380ffb68c431254dc55850c1fbce62fa54063b71ef33abfa963d6f6101  openspec/changes/a155-observe-versioned-strategy-decisions/tasks.md
3c25c8d41df510403d54725e54a8aad1b58f3f62369eae8ca7be38f4340f6791  tools/a155-observe-versioned-strategy-decisions/README.md
49f37b043b97a36c50bc9db2c1e8032c3f752e3bc195014a6d1495c7dae2cede  tools/a155-observe-versioned-strategy-decisions/main.go
d7396a14d4038584996fd8b5e37d0d5c153b495b2d4102dd207e010738a4d71d  tools/a155-observe-versioned-strategy-decisions/main_test.go
765886e02d0d6afb864e18859465fa59c35e1eab262e10604dbbbd803ceb9c2b  tools/a155-observe-versioned-strategy-decisions/testdata/synthetic.jsonl
```
