## 0. Manager Orchestration and Independent Review Protocol

이 change의 production 코드와 테스트는 Terra 구현 에이전트가 작성한다. SOL/Manager는 OpenSpec·분석·리뷰 기록과 최종 완료 판정만 소유한다. 각 구현 로트는 다른 Terra 적대 리뷰어가 read-only로 검토하며, 구현자가 P0/P1을 수정한 뒤 같은 리뷰어가 재검토해야 한다. 로트 밖 파일은 Manager가 명시적으로 재할당하기 전까지 편집하지 않는다.

| Lot | Terra implementer ownership | Depends on | Separate adversary focus | Acceptance gate |
|---|---|---|---|---|
| L0 — current-base freeze | `analysis/**` including immutable `analysis/goldens/*.json`, their hash manifest/validation receipt, and `base-commit.txt`; production/test files are read-only | A100 R0/M0 main landing | stale graph/AST, missing impact/branch/risk rows, dependency matrix truth, non-machine-readable goldens | Manager accepts fresh base, exact target map and validated non-code goldens before any production edit |
| M-B0 — unused official US measurement seam | new `internal/official/a112_mbus_read.go`, `a112_mbus_read_unix.go`, `a112_mbus_read_unsupported.go`, `a112_mbus_read_test.go`, `a112_mbus_static_test.go` only; every existing official/client/token/trace/raw reader and every product caller is read-only | L0 + accepted M-B HOLD evidence | authority-origin binding, FD-safe cached-token GET, one-attempt raw body/rate/cursor preservation, compression/proxy/redirect/account/WTS/mutation exclusion, zero product callers | RED/GREEN/race/vet/static guard; separate reviewer P0/P1=0; M-B remains HOLD and no source/finality authority is minted |
| M-B1 — bounded one-shot collector and receipt | new `tools/a112-mb-us-source/**`, sanitized secure-`/tmp` receipt, and `analysis/measurements/m-b-us-source/**` only; no `cmd/tossctl`, engine, strategy or runtime file | accepted M-B0 | exact AAPL/US/1m/count=200/unadjusted request, bounded cursor continuity, calendar/quote/rate evidence, 0700/0600, redaction and no retry/fallback | human-authorized read-only run; separate receipt reviewer; Manager alone may record M-B PASS |
| L1 — official source/evidence authority | `internal/strategyevidence/**`, additive read-only official KR/US raw bar/quote adapter files and their tests | L0 + Manager-recorded M-B PASS | raw bytes/decimal/time/calendar/currency/pagination/rate semantics, strict dual-cutoff decoder, append-only correction/replay | production US source identity frozen; no WTS/float/fixture fallback; finalized/closed authority independently proved; reviewer P0/P1=0 |
| L2 — pure breakout core | new `internal/breakoutlane/**` and fixture/property tests only; official/runtime/router files are read-only | L0; L1 contract may be represented by an inert test-only port, never production authority | pure state machine, integer arithmetic, deterministic replay, first-touch/averaging-down refusal, dependency closure | RED captured, GREEN/race/vet, no broker/journal/toggle/clock/imported production source; reviewer P0/P1=0 |
| L3 — canonical contracts and RouteSet | `internal/strategyflow/**`, `internal/strategyrouter/production.go`, `internal/strategyproposal/**`, `internal/app/engine/strategy_route_authority.go`, `strategy_proposal_authority.go` and their tests; `router.go` selection body remains read-only until L4 | L1 + L2 | exact 8-set atomicity, tagged union/registry, manifest migration, all-candidate RouteSet, removal of market-wide preselection/one-proposal caller assumptions, family/calibration seals | existing six fixtures preserved, no pre-evaluator cross-family selection, reviewer P0/P1=0 |
| L4 — family quota and post-evaluation arbiter | ~~`internal/strategyrouter/{router,quota,scheduler}.go`~~ → quota 는 **`internal/scheduler.BudgetCoordinator`**(7.1 설계 입력, 2026-10-01 a070 처분 감사 §quota — `QuotaAuthority` 편집 대상 지정 철회), new coordinator queue/arbiter files and tests | L3 | one physical quota, family-scoped capabilities, owner-scope dedup/order/overflow, calibrated selection after pure proposals | property/race/pressure tests GREEN, raw candidate score never wins cross-family arbitration, reviewer P0/P1=0 |
| L5 — 8-worker runtime and read-only projection | `internal/app/engine/strategy_entry_supervisor.go` (`NewStrategyEntrySupervisor` and worker construction), new family-worker/coordinator runtime files, `strategy_runtime_projection.go` (`Context.Read`, `strategyProjectionFromAssembly`), `internal/strategyprojection/**`, additive HTTP/console read models and tests | L4 | explicit 8 workers/2 coordinators, failure isolation, OFF/UNOBSERVED vs SHADOW, additive projection compatibility | fault/race/leak/API/console tests GREEN, no mutation or activation writer dependency, reviewer P0/P1=0 |
| L6 — shared dispatch/risk integration | `internal/app/engine/strategy_risk_authority.go` (`strategyRiskAuthorityLoader.collectMarket`), `strategy_account_first_leg_authority.go` (`productionStrategyFirstLegAuthorityLoader.collectStrategyFirstLegAuthority`), `strategy_first_leg_admission.go`, and the lineage-only seam in `strategy_dispatch_cycle.go`; scheduler/execgw budget and required journal admission files may be added only by a Manager amendment | L5 + a064/a066/a070/a072 complete + A100 ProtectionReady | q_final monotonicity, one owner/reservation/lease, validation order, first-leg-only, crash/retry and reduce-only preservation | missing prerequisite yields broker exposure request 0; dispatch body remains reuse-first/citation-only unless separately approved; reviewer P0/P1=0 |
| L7 — post-edit evidence/gates and dormant release | `analysis/**`, verification/review receipts and deployment manifests only; production/test files read-only | L1–L6 accepted | AST SHA drift, unmapped branches, skipped gates, dependency truth, additive schema, dormant rollout | final gstack review, repository gates, Manager source/test/spec reconciliation; deploy remains conditional |

- [x] 0.1 Land A100 R0/M0 on `main`, then have L0 recapture the A112 base and prove that no A100 product/protection task is falsely treated as complete.
- [x] 0.2 Freeze the table above as the file-ownership ledger; every exception requires a Manager-authored OpenSpec amendment before the edit.
- [x] 0.3 For each L1–L6 lot, record `Terra RED → Terra GREEN → separate Terra adversarial review → implementer fix → same reviewer recheck → Manager acceptance` in `review.md` with exact commands and file ownership.
  **닫음(2026-10-04, Manager 판정 — 기록 의무 이행):** 로트별 리뷰 사슬(RED → GREEN → 별도 적대 리뷰 → 수리 → 같은 리뷰어 재확인 → Manager 수락)이 review.md 에 있다 — L0 :42 · L1a :311(브리프) · :339(구현 보고) · :348(독립 리뷰) · :358(수락) · L1b :366 · :416 · :425 · :441 · :472(ACCEPTED) · L1c :483 · :542 · :753 · :774 · L2 :70(수락, P0/P1=0) · L3 :827 · :903 · :940 · :980 · :1242 · L4(코디네이터 · 중재 — 5.4.x · 7.1/7.2) :1540 · :1645 · :1731 · :1816 · :1933 · :5895 · L5 :1412 · :2008~:3451(5.5 와 fix 13 라운드) · :3452 · :3790 · :3925 · :4053 · :4166 · :4229 · :4335 · :5250~:5522 · :5658 · :5737 · :5821 · :5930 · :5996 · :6015 · L6 :5523 · :5605 · :5643 · :5851 · :5873 · :6070 · :6251 · :6282. 명령 · 소유 파일은 각 절과 `analysis/measurements/lot-*/`(RED · 변이 · verify 로그). **L1c 수락은 장중 사람 프로브 대기 — 기록상 측정 1회 존재(:774, 두 시장 마감 중 실행), 수락 아님**(사람 큐 항목, 이 체크와 독립).
- [x] 0.4 Forbid a reviewer from editing the lot under review; reviewers report P0/P1/P2 and only P0/P1=0 may advance the dependency chain.
  **닫음(2026-10-04, Manager 판정 — 이행 완료):** 리뷰어는 읽기 전용으로 돌았다 — L1b 「Terra adversary (read-only, 61 `-overlay` mutants)」(review.md :428), 5.5 적대 리뷰 13 라운드(:2126~:3451), 8.5 4판(`analysis/review-8.5-2026-10/` — 프롬프트가 편집 금지 · 리뷰어 출력마다 저장소 무변경 실측, codex r2 는 아카이브 사본에서 workspace-write). P0/P1 이 0 이 아니면 전진하지 않았다 — 각 로트 수락 줄의 「reviewer P0/P1=0」 과 8.5 HOLD → 응답 로트 6f5b0df6 → SHIP(:6416 절).
- [ ] 0.5 After L7, run one final gstack pre-landing review over the complete diff; fix and re-review every P0/P1 before Manager completion verification.
- [x] 0.6 Treat build, dormant deployment and operating activation as three distinct gates. Missing A100 ProtectionReady permits build/shadow tests only and blocks container replacement and every exposure-raising dispatch.
  **닫음(2026-10-04, Manager 판정 — 증거 인용):** 세 관문이 따로 집행된다. ① 빌드: `measurements/gate-8.1-8.3-2026-10-04/build-only-8.6.log`(go build 무태그 · 태그 · cmd/tossctl exit 0). ② 휴면 배포: 사람 정책 — 이미지 빌드 · 컨테이너 교체는 사람이 `make image CHANGE=…`(docs/operations.md :415) · 운영 반영 승인(.claude/CLAUDE.md 안전 7 · docs/WORKFLOW.md :24); A100 ProtectionReady 미완(`openspec/changes/a100-wire-fill-to-broker-protection/tasks.md` 미완료 86, 2026-10-04 실측)이라 8.6 은 BLOCKED. ③ 운영 활성화: 서명된 4-가족 활성화 없이는 레인 effective ON 불가(8.7.1 · 8.7.2, 생산 핀 0) + ProtectionReady 증명 없이는 상승 0 — execgw `TestEachProtectionAttestationFailureStopsBuysAndKeepsReductionsFlowing` · engine `TestAnExpiredProtectionAttestationStopsTheStrategyFirstLegBeforeTheBroker`(6.6, review.md :6282).
- [x] 0.7 Complete M-B0 and M-B1 before L1. A different Terra reviewer is read-only for each implementation/evidence lot, and SOL/Manager alone records M-B PASS in `review.md`. M-B0 or M-B1 completion MUST NOT itself mark task 0.7 complete.
  - [x] 0.7a M-B0: add an A112-only, unused official measurement seam in the exact new `internal/official/a112_mbus_read*` files. It MUST retain the existing `RawMinuteCandles` US rejection and MUST NOT edit or reuse the ordinary retrying `Client.get/send`, `AttemptTrace/RateBudget`, `AuthHeaders`, token exchange/refresh/cache writer, generic float/time adapters, WTS/hybrid, account discovery or any order/config/runtime surface. Its factory accepts one exact `*official.Client`, revalidates that same instance's sealed authority origin/transport under its configuration lock, and MUST NOT accept a caller-supplied origin token or a separate client/read provider. It clones that instance's transport with `Proxy=nil` and `DisableCompression=true`, refuses redirects and non-GET, sends `Accept-Encoding: identity`, explicitly suppresses the default Go `User-Agent`, and rejects non-empty/non-identity `Content-Encoding`. On Unix it walks/opens the cached-token path descriptor-relatively with no-follow semantics, `O_RDONLY|O_CLOEXEC`, then `Fstat`s the opened FD as current-UID regular 0600; unsupported platforms HOLD before request. It validates the token with the existing 60-second skew without exchange/write; requires a caller deadline no later than 15 seconds; performs exactly one fixed request per call; the on-wire application header allowlist is `Authorization`, optional `Accept`, `Accept-Encoding: identity` (plus protocol-owned `Host` only); prechecks `Content-Length` and reads through a 2-MiB-plus-one limiter; preserves the same request's raw body, exact descriptor, raw `nextBefore` JSON plus its decoded UTF-8 value bytes, and allow-listed raw headers `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, `Retry-After` before decode; and exposes only opaque copy-on-read evidence. An injected/process clock may be used only for cached-token expiry and request/overall deadline enforcement; it MUST NOT be serialized as official evidence or used to infer candle finality/source observation time. Existing product references to the new M-B0 exported symbols MUST remain zero and a resolved Go-AST selector/reference guard MUST exempt only their definitions, M-B0 tests and the future exact `tools/a112-mb-us-source` collector; it MUST NOT ban ordinary imports of the pre-existing `internal/official` package.
  - [x] 0.7a.1 Capture named REDs for exact candle `US/AAPL/1m/count=200/adjusted=false`, exact raw AAPL orderbook and explicit-date US calendar GETs, same-instance binding versus nil/configured/cross-client/wrapped clients, decoded cursor value continuity plus forensic raw JSON for escapes/NUL/space/Unicode, `nextBefore` absent/null/empty/string/number/object/array, raw decimal/timestamp/USD preservation, same-request body/rate binding, one data GET per call, on-wire header allowlist/default-User-Agent suppression as observed by `httptest`, missing/>15-second deadline, injected-clock token/deadline boundaries and absence of serialized/inferred wall-clock evidence, `Content-Length` 2-MiB boundary/overflow and chunked limit-plus-one, gzip/br/unexpected encoding, malformed envelope, 302/proxy/401/403/429/5xx, cache miss/expiry/path-swap/symlink-ancestor/symlink-leaf/mode/owner, non-GET, OAuth/account/order/config/WTS calls at zero, production-symbol-reference closure and unchanged `RawMinuteCandles` US refusal. RED tests use `httptest`/private test seams only and make no external request.
  - [x] 0.7a.2 Parse every allow-listed header as a slice, not `Header.Get`: successful evidence requires exactly one valid Limit/Remaining/Reset value and no Retry-After; 429 diagnostic permits exactly one valid Retry-After but mints no evidence. Duplicate, missing, malformed or negative values HOLD. A subsequent candle page is allowed only when the immediately prior candle response from the same endpoint has `remaining >= 1`; otherwise the next network request count is zero. Orderbook and calendar are each single-shot and never retried; any first error aborts the remaining plan.
  - [x] 0.7a.3 Run `go test ./internal/official -run '^TestA112MBUS' -count=1`, the same focused race suite, `go vet ./internal/official`, focused `-count=20`, static import/reference checks, gofmt, diff-check and SDD/logic-map gates. M-B0 acceptance proves only a bounded raw transport capability; M-B remains HOLD, raw evidence is not production evidence, and no L1/strategy/runtime caller is permitted.
  - [x] 0.7b M-B1: only after Manager accepts M-B0, add the non-installed one-shot tool under `tools/a112-mb-us-source/**`. It accepts required explicit `--token-cache`, `--go-binary`, secure `/tmp` receipt root, exact `--session-date=YYYY-MM-DD` and initial `--before`; constructs a default-origin `official.New(official.Credentials{}, tokenCache)` with no config/account resolver and relies only on M-B0's direct cache-only seam, so missing/invalid cache cannot fall through to OAuth. It fixes symbol/market/interval/count/adjusted in code; opens the root descriptor-relatively/no-follow, `Fstat`s current-UID directory 0700, and uses `openat(O_EXCL|O_NOFOLLOW|O_CLOEXEC)` plus `Fstat`/directory-FD fsync for 0600 payloads. A sentinel proves this capability before network; unsupported platforms HOLD. It calls the M-B0 seam with exactly four candle pages maximum (amended 2026-08-16: exhausting the four-page cap is recorded in the receipt as `cap_exhausted`, not HOLD — see 0.7b.3) plus one orderbook and one calendar request, each with a 15-second deadline and a 120-second overall deadline; exactly 120 seconds is expired, it starts no request unless 15 seconds remain, and it rechecks cancellation/deadline before and after identity, every receipt write, seal and success. It passes only the decoded prior cursor value bytes without trim/casefold/normalization and applies URL percent-encoding once, while preserving raw cursor JSON and exact canonical query separately. It never retries/backoffs/falls back and returns HOLD on cursor loop, signal, rate gate, unsafe filesystem, recursive duplicate JSON key, secret-like key/value, redaction failure or any response error. Before seal and before returning success it revalidates the pinned root/run descriptors as current-UID exact-0700 directories; every payload is a current-UID regular exact-0600 file of at most 64 MiB whose complete bytes, same-FD device/inode/size/read length and SHA-256 match a strictly parsed, complete, deterministic, self-excluding manifest. Any unexpected entry, mode/owner drift, post-write byte drift, manifest drift or limit overflow taints the run and HOLDs. It MUST NOT be installed, linked into `tossctl`, or imported by production code.
  - [x] 0.7b.1 Before the first request, M-B1 seals SHA-256 for the approved exact M-B0/tool production-input manifest, approved tests, `go.mod`/`go.sum`, the required caller-supplied absolute `--go-binary`, verified absolute Git binary, frozen base, exact tracked diff/untracked content set, actual compiled Go/Cgo/embed closure, prescribed build environment and executable bytes. It MUST NOT discover build authority from `runtime.GOROOT`, ambient `GOROOT`, `PATH` or inherited `GO*`; the exact no-follow regular Go executable selected by `--go-binary` is resolved, hashed and revalidated pre/post, then used for `go env`, `go mod verify`, `go list` and the prescribed rebuild. Every Git command uses the vetted absolute binary, canonical root `Dir` and a sanitized environment with inherited `GIT_*` ignored; approved source reads stay inside that canonical root, open the leaf with `O_NOFOLLOW`, require a regular file, bind the same FD's device/inode/size/read length, and reject inputs larger than 256 MiB instead of truncating. The prescribed rebuild inherits no caller `GO*`, `CC`, `CXX` or `CGO` controls, pins `GOENV=off`, empty `GOFLAGS`, `GOWORK=off`, `GOPROXY=off`, `GOSUMDB=off`, `CGO_ENABLED=0`, a fresh private `GOCACHE` and verified offline `GOMODCACHE`, runs `go mod verify`, derives the actual closure with machine-readable `go list -json -deps`, rejects every compiled input outside the exact GOOS-bound manifest, and must reproduce the running executable SHA-256. It re-hashes the same identities after collection; any drift is HOLD. Ad-hoc `go run`, a caller-claimed build command or an identity-unknown binary cannot support PASS. Raw bodies stay only in secure `/tmp`; `analysis/**` receives sanitized receipt metadata, hashes, modes and absence states, never a raw mirror.
    - [x] 0.7b.1.1 Capture an offline RED with the exact prescribed `-trimpath` binary proving `runtime.GOROOT()` is empty/non-absolute and current identity preflight cannot reach the reader. GREEN must require the explicit absolute `--go-binary`, reject missing/relative/symlinked/non-regular/world-writable/drifted binaries before reader call 1, prove the selected binary alone serves every Go env/module/list/build step, ignore hostile ambient `GOROOT/PATH/GO*`, and execute the real trimpath-built collector through successful identity preflight with a recording reader still at zero external calls.
    - [x] 0.7b.1.2 Capture REDs for Go and Git pathname swap-after-hash/restore-before-post-hash attacks. GREEN must execute only an exact-byte private snapshot copied from the already opened no-follow regular FD into a freshly created current-UID 0700 `/tmp` directory and O_EXCL 0500/0555 regular file, fsync file and directory, bind the snapshot digest to the selected source binary digest, reject unexpected entries/mode/owner/content drift, and keep the snapshot capability private for the whole identity/build phase. No verified command may reopen the caller-controlled original pathname after verification. The complete machine-derived tracked dependency closure is frozen by base/diff/content/compiled SHA evidence; untracked M-B0/tool production inputs remain restricted to the exact GOOS allowlist rather than treating arbitrary tracked or untracked additions as approved.
    - [x] 0.7b.1.3 Because a relocated Go executable resolves its distribution through GOROOT, capture REDs proving a private Go-binary copy with an unbound original GOROOT can execute swapped compiler/source bytes or fail relocation. GREEN must descriptor-walk the selected `<go-root>/bin/go` distribution without symlinks, bind root and every copied regular input by pre/post device/inode/size/read length and SHA, reject special files, cap the snapshot at 512 MiB and 50,000 entries, and create a deterministic manifest in a fresh current-UID 0700 private root. The private root must contain only O_EXCL owner-only directories/files, be file/directory-fsynced, reverified before and after every Go command, and supply the only internally set `GOROOT`; the original toolchain pathname and ambient GOROOT remain forbidden after snapshot. Cleanup must remove the private tree on every success/HOLD path, and real trimpath preflight plus source/tool swap-restore and leftover-directory tests must pass offline.
    - [x] 0.7b.1.4 The secure Go-distribution snapshot and prescribed identity/rebuild must finish inside the existing 120-second collector deadline on the measurement host; setup outside that deadline is forbidden. A fixed finite worker pool may parallelize only a pre-enumerated descriptor-validated inventory while preserving per-file no-follow/pre-post metadata/digest/fsync, deterministic directory creation/fsync, sorted manifest, first-error cancellation and complete cleanup. Record real-host snapshot time and end-to-end offline identity-preflight time; timeout is HOLD, not a reason to relax durability or move work before the clock.
  - [x] 0.7b.2 Run the human-authorized read-only M-B1 probe from the verified POSIX receipt root. A Terra collector owns only the tool, sanitized `/tmp` receipts and `analysis/measurements/m-b-us-source/**`; a different Terra reviewer independently verifies source/binary identity, raw body hashes, decoded cursor continuity plus forensic raw JSON, explicit regular-session full coverage (amended 2026-08-16: terminal null is recorded when seen but is no longer a PASS precondition), USD, explicit US regular-session calendar join, quote body, unique allow-listed rate headers, redaction, exact modes and self-excluding manifest. No order, config, auth update/exchange, engine, container, install, stage, commit or push action is allowed.
  - [x] 0.7b.3 (amendment 2026-08-16, human-approved after run 3 ACCEPT-HOLD) Terra amendment lot on `tools/a112-mb-us-source/**` only: capture a RED proving four non-null candle pages currently HOLD before orderbook/calendar/seal, then GREEN so that cap exhaustion is recorded in the receipt (`cap_exhausted` with the last cursor, in the sealed manifest set) and the run continues to orderbook, calendar, post identity and seal with zero further candle requests; keep the 4-page hard bound, the null-only terminal typing, cursor loop/malformed HOLD, the rate gate and every other HOLD path byte-identical in intent; refresh `analysis/function-logic/tools-a112-mb-us-source--run` AST/maps and this lot's sanitized receipt schema; a different Terra adversary and a gstack review must return P0/P1=0 before Manager acceptance. No production, `cmd/tossctl`, official-package or runtime file is touched.
  - [x] 0.7b.4 (amendment 2026-08-16) Pre-run checklist for every external M-B run: the human enumerates every holder of the OpenAPI credential (StockOS `infra-toss-intelligence-1`, TossOS engine/httpapi/console containers, host `tossctl`) and records that each shares the token cache or is stopped; the run uses the amended collector at its frozen executable SHA rebuilt under the prescribed environment, an explicit initial `--before` at or after the session's `regularMarket` end (empty-cursor reruns are not authorized), exactly one invocation per authorization; then a different Terra receipt adversary and gstack evidence reviewer verify regular-session full coverage against the calendar body, the sealed `candle-crawl.json` record (`pages`, `terminal`, `last_cursor_sha256` == sha256 of the last decoded cursor), plus every 0.7b.2 item. Prepared 2026-08-16 18:49 KST: amended executable rebuilt twice under the prescribed environment from `main.go` SHA `1945343d…`, both builds identical, **frozen executable SHA-256 `232c787c50685c623f6bfade2d0713850e295ff32bb83834d9f0d324fb3137ce`**, pinned at `/tmp/a112-mb-us-build-amended.OSVxn5/a112-mb-us-source-pinned`, receipt root `/tmp/a112-mb-us-receipt-run4.ZTagql`, `--before=2026-08-15T05:00:00.000+09:00`. Executed by the human 2026-08-16 19:01–19:06 KST after the pasted pre-check (StockOS holder Exited, engine/httpapi sharing the cache, cache mtime unchanged, SHA confirmed): **two** invocations (run A HOLD-tainted in the post-identity phase, run B sealed) — a recorded deviation from "exactly one invocation per authorization", the second being the human's own rerun; Terra receipt adversary `ACCEPT-WITH-P2` and gstack `SHIP-IT`, both P0/P1 = 0; regular-session 390/390 joined with the calendar body, sealed `candle-crawl.json` consistent (see review.md run 4 section and `analysis/measurements/m-b-us-source/receipt-2026-08-16-run4.json`).
  - [x] 0.7c Manager alone records M-B PASS after the independent M-B1 receipt review. Any missing/ambiguous field, regular-session coverage gap, cursor loop/malformed HOLD, rate ambiguity, unofficial/WTS/float/time-derived evidence, unsafe mode or reviewer P0/P1 keeps task 0.7 unchecked and blocks L1/L3–L7. M-B PASS authorizes only L1 implementation start; it does not accept L1, prove candle publisher finality/closed-bar authority, activate a lane, dispatch, deploy or raise exposure. **M-B PASS recorded 2026-08-16 (review.md "Manager decision — M-B PASS") on the sealed run-4 B receipt (manifest `765013e4…`, executable `232c787c…`); residuals carried to L1: empty-orderbook (no quote level schema), bar-label convention undecided, variable decimal scale, shared rate quota / Reset unit opaque, run A stderr and the human's second-invocation lines pending in the ledger.**
- [x] 0.8 Allow L2 pure breakout work after L0 even while M-B is blocked, but its source port must be fixture-only and incapable of becoming production authority; L3–L6 remain blocked until their dependency cells are complete.

## 1. SDD Baseline and Dependency Gates

- [x] 1.1 Rebase the implementation session on the approved change base, run `make sdd-sync`, record fresh CodeGraph definitions/callers/callees/impact for every target in `analysis/pre-edit-targets.md`, and update `base-commit.txt` only through an explicit change review.
- [x] 1.2 Audit a064, a066, a070, a072 and a100 task/gate status plus M-B official US evidence feasibility into an executable dependency matrix; prove that missing evidence replay/source authority, q_final/owner/exit-bypass, router/runtime or ProtectionReady keeps exposure-raising requests at zero.

  **정정(2026-10-01, a070 처분 ② 종결 — Manager 지시).** 동결 골든 `analysis/goldens/dependency-matrix.json`(manifest.sha256 로 고정 — L0 동결이라 파일은 고치지 않고 여기와 HANDOFF 에 정정을 기록한다)의
  「a070 router/scheduler · INCOMPLETE · blocks L6 · L7」 행은 더 이상 사실이 아니다: a070 은 `--skip-specs` 로 아카이브됐고(`ee7b8cb8`, 처분 ② 부분 대체), 코어 타입 ·
  owner/market 봉인은 a072 · a112 생산 기반으로 살아 있으며 `Route()` 는 a112 RouteSet 이 대체했다(처분 감사 `0d3e2849` —
  `openspec/changes/archive/2026-10-01-a070-add-multi-market-horizon-router/analysis/disposition-audit.md`). 따라서 a070 은 L6 · L7 의 차단 선행이 아니다. 남은 실재 공백
  (스케줄러 durable record · migration · `QuotaAuthority` 는 생산 호출자 0 인 섬)은 a112 7.1 의 설계 입력(아래 7.1)과 ROADMAP 「a070 이월」이 소유한다.
- [x] 1.3 Generate and complete current-base Go AST, Function Logic Map, Branch Test Map and risk-pattern report before editing each existing function listed as `FLM-required`, including every fail-closed, retry, owner and safety-bypass branch.
- [x] 1.4 Freeze the exact 8-descriptor matrix, family/worker keys, state/refusal enums, arbitration envelope, KR/US official source identities, breakout evidence schema, threshold table, queue/deadline/backoff policy and compatibility rules as immutable canonical JSON under `analysis/goldens/`, with sorted-path SHA256 manifest and a validation receipt showing `jq -e` schema assertions plus duplicate/unknown/missing-member negative cases. L2/L3 tests consume hash-identical copies and MUST NOT reinterpret or rewrite the L0 goldens.
- [x] 1.4.1 Freeze exact family enum `{CONTINUATION, REVERSAL, WEEKLY_VALUE, BREAKOUT_RETEST}`, integer `score_ppm` range `0..1_000_000`, approved score/calibration version and digest semantics, and require production refusal even for a singleton proposal when calibration authority is absent.
- [x] 1.4.2 Freeze coordinator owner-scope queue contract: dedup key `(account,market,symbol,position_generation,family,lane_id,lane_version,snapshot_digest)`, deterministic order, positive finite capacity, latest-per-key coalescing, typed overflow/drop counters and no market-wide single-proposal assumption.
- [x] 1.4.3 Freeze additive operator schema: retain all legacy market-level fields, add deterministic fixed-order `lanes[8]` and `coordinators[2]`, keep the surface read-only and require older clients to tolerate unknown additive fields.
- [x] 1.4.4 Freeze setup identity as `sha256:` lowercase hex over the UTF-8, domain-separated, NUL-joined canonical `(market,symbol,session_id,calendar_version,opening_range_first_bar_id,opening_range_last_bar_id,lane_id,lane_version,config_digest)` preimage with no trailing delimiter and a known digest vector. Bar revision belongs to snapshot/evidence digest, not setup ID. A pre-terminal correction replays a new immutable snapshot for the same setup; a post-terminal correction cannot resurrect or mint another proposal; a new regular session creates a new setup ID.
- [x] 1.4.5 Freeze quote/FX sizing inputs and inclusive boundaries: positive `max_quote_age_ms`, `max_spread_ppm`, `max_entry_drift_ppm`; accept only age/spread/absolute drift `<=` limit; otherwise stable `QUOTE_STALE`, `SPREAD_TOO_WIDE`, `ENTRY_DRIFT_EXCEEDED`. Quote seal includes bid/ask/last integer minor units, currency, source/received timestamps and digest. FX seal names account/instrument currencies, direction, integer scale, as-of/fresh-until and digest; every conversion rounds risk/notional capacity down and per-share cost/risk up so quantity never rounds upward.
- [x] 1.5 Obtain human approval of proposal/design/spec deltas before production code changes; record unresolved review findings without weakening OFF defaults or safety prerequisites.

## 2. RED Contract and Property Tests

- [x] 2.1 Add failing registry tests requiring exactly continuation/reversal/weekly-value/breakout-retest × KR/US descriptors, exact OFF/OFF/UNOBSERVED defaults and rejection of partial/duplicate/unknown/mismatched bindings.

  **2.1 종결(2026-10-01 대조 감사 — 미착지).** 이름 결속: 여덟 서술자 — strategyflow `TestPairedRegistryCoversAllFourFamiliesInBothMarkets` · `TestPairedRegistryCoversKRUSContinuationReversalWeeklyAndBreakout`,
  strategyrouter `TestProductionRouteDescriptorsCoverFourFamiliesPerMarket`, strategyworker `TestProductionWorkersAreExactlyTheEightTheGoldenFroze` · `TestEveryProductionWorkerKeyIsDistinct`;
  OFF/OFF/UNOBSERVED — 같은 시험들 + `TestDescriptorsShipKRAndUSTogetherDefaultOFF` · `TestEveryProductionWorkerIsBornDormantAndEmitsNothing`; partial/duplicate/unknown/mismatched —
  `TestValidateDescriptorsRejectsPartialDuplicateUnknownAndMismatched` · `TestProductionRouteCandidatesRejectLegacyThreeFamilyAndPartialSets` · `TestProductionRouteCandidatesRejectFamilyDriftAndPartialFamilyCoverage`.
  빈칸(불일치 축이 Desired 하나뿐) → `a112_descriptor_axes_test.go` `TestEveryDescriptorFieldOtherThanTheKeyIsPartOfTheBinding`(열쇠 아닌 필드 전부를 반사로 열거, 변이 X1 CAUGHT).
- [x] 2.2 Add failing breakout state-machine table/property tests for every valid forward edge, skipped edge, terminal non-resurrection, duplicate/reordered bar, first-touch refusal and deterministic replay.

  **2.2 종결(2026-10-01 breakout 덮개 2차 — 미착지).** 「every valid forward edge」 = Manager 판정 B1(골든 `allowed_transitions` 는 검증 집합): 생산자 있는 여덟은
  `a112_transition_producer_census_test.go` `TestTheObservedBreakoutEdgesPlusTheReservedSixAreTheGoldenSet`(관측 변 = 여덟, ∪ 예약 여섯 = 골든 열넷)과
  `TestTheBreakoutTransitionProducersAreExactlyTheCensus`(패키지 생산 파일 전체의 phase 생산 자리 수 · 우회 철자 0 — 생산자를 더하면 뒤집힌다);
  skipped edge — `TestAdversarialSnapshotEvaluatorRejectsRawTransitionBypass` · `TestPublicSurfaceCannotAssertEventsOrForgeMachine`(전이 입구는 봉인 스냅숏뿐);
  terminal non-resurrection — `TestTerminalCorrectionsRetainTerminalAuthority` · `TestACorrectionCannotResurrectAFailedSetupIntoALowerLeg` · `TestNoAveragingDownLegAfterAFailedSetup`;
  duplicate/reordered bar — `TestDuplicateOrReorderedBarsAreRefusedBeforeTheSnapshotSeals` · `TestSnapshotRejectsSkippedOpeningAndPostBreakoutSequence`;
  first-touch — `TestAdversarialFirstTouchMissingRangeAndBadBarCannotPropose`; deterministic replay — `TestFinalRedTeamDuplicateSnapshotIsIdempotent`.
  변이 `lot-bk/mutation-bk2.tsv` BK2-05~08 · 27 · 28 CAUGHT.

- [x] 2.2.1 Add failing setup-identity tests for pre-terminal correction, correction after each terminal state, the PROPOSED-before-CONSUMED window and regular-session rollover, proving stable same-session setup ID, immutable snapshot revision and no proposal seal/first-leg resurrection.

  **2.2.1 종결(2026-10-01 — 미착지).** `a112_setup_identity_and_bars_test.go` `TestTheSetupIDIsStableAcrossSameSessionCorrectionsAndChangesWithTheSession`(종단 전 정정 ·
  PROPOSED 뒤 · INVALIDATED 뒤 같은 setup, 세션 교체 = 새 setup · 이전 종단 비상속 — 봉 ID 가 같아도) · `TestACorrectionAfterConsumedNeitherResurrectsNorReissues`
  (CONSUMED 는 생산자가 없어 패키지 안에서 봉인한 prior 로 소비 계약만 — 생산은 6.4) · `TestAProposedSetupNeverReSizesIntoALowerLegOrARetreatedStop`(PROPOSED-before-CONSUMED 창) ·
  기존 `TestAdversarialCorrectionReplaysPreTerminalAndPreservesProposed` · `TestTerminalCorrectionsRetainTerminalAuthority`(TIMED_OUT 뒤). 변이 BK2-01~04 · 29 · 30 CAUGHT
  (BK2-30 「CONSUMED 비종단」 은 이 로트 전에는 잡는 시험이 없었다 — 원장의 유일한 실패 이름이 새 시험).

- [x] 2.3 Add failing threshold boundary tests for 15-minute range, 1-tick versus 0.10 ATR buffer, 0.10..0.25 ATR retest tolerance, KR 8/US 10 timeout, 1.5 RVOL, 1.2/2.0/2.5 counterfactuals and 0.35 wick veto using integer PPM arithmetic.

  **2.3 종결(2026-10-01 — 미착지).** 1.5 입장 · 2.0 · 2.5 반사실 경계 `TestRVOLAdmissionAndCounterfactualBoundaries`(BK2-24~26 CAUGHT); 범위 · buffer · 허용폭 · 시한 · wick
  기존 `TestGstackRepairFrozenVocabularyAndV1Thresholds` · `TestGstackRepairRetestQualifiesToleranceEndpoints` · `TestTimeoutExactBoundaryKRAndUS`.
  **1.2 반사실 = Manager 판정 (b)**(design.md 「1.2 반사실의 의미」 — 원문 인용 + (a) 의 공허 논증): 입장 못 한 봉 중 close buffer · wick 통과 · 1.2 <= RVOL 이면
  `RVOLAt1200000` 기록(`evaluateFresh` 새 B7, 기록 전용). `a112_rvol_counterfactual_test.go` `TestTheOnePointTwoCounterfactualRecordsABarThatOnlyTheLowerThresholdWouldAdmit`
  (1.2 정확 · 1.5 바로 아래 기록, 1.2 바로 아래 · wick 초과 · buffer 미달 미기록, 각 경우 RVOL 1.0 쌍둥이와 결정 동일) · `TestTheAdmittedBreakoutPathIsUnchangedByTheCounterfactual`.
  RED `lot-2.3/red-2.3.log`, 변이 `lot-2.3/mutation-2.3.tsv` CF-01~10 CAUGHT(기록 억제 · 문턱 교환 1.5/2.0 · +1 · 배타 · 조건 누락 둘 · 입장 깃발 오기록 · 결정 변경 · 입장 경로 교환).

- [x] 2.4 Add failing sizing property tests proving cost-inclusive `risk_per_share`, overflow-safe floor, `0 <= q_final <= q_candidate`, non-protective stop/target refusal and no averaging-down or stop retreat.

  **2.4 종결(2026-10-01 — 미착지).** `a112_sizing_oracle_test.go` `TestSizingMatchesTheExactRationalOracle`(결정적 난수 2만 건 · math/big 신탁: 비용 포함 risk ·
  넘침은 감싸지 않고 거절 · 수락마다 q_candidate = 정확한 유리수 바닥(코드와 다른 경로) · 0 <= q_final <= q_candidate <= …; 갈래 분포 하한 50 · 128 비트 곱 수락 365 건) ·
  `TestMulDivMatchesTheExactQuotientOrRefuses`; 비보호 stop/target — `TestANonProtectiveStopNeverProposes` · `TestGstackRepairFailedReclaimAndRiskRewardBoundary`;
  물타기 · 손절 후퇴 금지 — `a112_no_averaging_down_test.go` 네 시험(대조군이 prior 없이 실제로 제안함을 보인 뒤 prior 가 막음을 잰다).
  변이 BK2-09~18 CAUGHT, BK2-12(entry 0 절 단독 삭제) 동등 — stop==0 · stop>=entry 가 대신 거절.

- [x] 2.4.1 Add exact-boundary quote/FX property tests for age/spread/drift at and one unit beyond each limit, both FX directions/scales, stale/mismatched currency seals, overflow and conservative rounding; every accepted result must be no larger than the exact rational floor.

  **2.4.1 종결(2026-10-01 — 미착지).** `TestQuoteVetoMatchesTheGoldenFormulasAtAndBeyondEachLimit`(골든 spread · drift 식을 big.Int 로 — 한계와 같으면 수락, 한 단위 넘으면
  SPREAD_TOO_WIDE / ENTRY_DRIFT_EXCEEDED, ask<entry 포함, drift 넘침 = SIZING_OVERFLOW; spread 는 <= 2e6 이라 넘침 갈래는 도달 불가 — 시험이 그 상한을 단언; 나이 포함 경계 ·
  source>received · received>evaluated) · 기존 `TestGstackRepairQuoteAndFXExactBoundaries`(두 방향 · scale) · `TestAnInverseFXSealVerifiesTheCallersDigest` ·
  `TestAdversarialQuoteAndFXSealTimeOrderDigestCurrencyDirectionScale` · `TestAdversarialQuoteFXAndSizingRefusals`; 정확한 유리수 바닥은 2.4 의 신탁. 변이 BK2-19~23 CAUGHT.

- [x] 2.5 Add failing strict evidence tests for unknown fields/enums, float/minor-unit mismatch, secret-like fields, unbounded/duplicate/future/unfinished bars, append-only correction revision and dual-cutoff snapshot replay. (L1a 2026-08-16/17: RED-first in `breakout_bar_test.go`/`breakout_series_test.go`; "out-of-order bar" is enforced by L3's ordered bar ids — recorded not-applicable at this layer in review.md.)
- [x] 2.6 Add failing arbitration tests for unique highest calibrated score, exact tie, incomparable calibration, stale seal, active-owner priority, multiple-owner corruption and at-most-one dispatch handoff per owner scope.

  **2.6 종결(2026-10-01 대조 감사 — 미착지).** 이름 결속(strategyarbiter `arbiter_selection_test.go` 외): 유일 최고 `TestThreeFamiliesOnOneSymbolYieldTheSingleHighestScore` ·
  `TestATieBelowTheTopStillLeavesAUniqueWinner`; 동점 `TestATieAtTheTopIsRefusedRatherThanBrokenArbitrarily`; 비교불가 보정 `TestProposalsUnderDifferentScoreVersionsAreIncomparable` ·
  `TestAScoreAboveTheApprovedCeilingIsRefused`; 활성 소유자 우선 `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore` · strategyrouter `TestRouteSetPreservesTheActiveOwnerAloneBeforeAnyComparison`;
  다중 소유자 `TestTwoActiveOwnersAreItsOwnRefusal`; 범위당 handoff 1 — strategyhandoff `TestEachOwnerScopeCrossesTheSeamOnItsOwn` · `TestTwoSelectedScopesAreRefusedByNameInsteadOfSilently` ·
  strategyworker `TestOneOwnerScopeHandsAtMostOneThingToTheSeam` · engine `TestTheSameOwnerScopeSealedTwiceRefusesTheActivatedMarket`. **「stale seal」 은 해석 매핑이다**(스펙 · 설계 · 골든에
  정의 없음 — 발명 계약 아님): stale envelope `TestAStaleProposalClosesTheWholeScope`(arbiter_selection_test.go:143) · stale owner `TestAStaleOwnerRevisionIsItsOwnRefusal`(:269) ·
  `TestAnOwnerSnapshotOutsideItsFreshnessWindowIsAStaleOwner`(:308) · 봉인 뒤 변조 `TestAProposalMutatedAfterSealingIsRefused`(:165) · 같은 레인 다른 스냅숏
  strategycoordinator `TestTheSameLaneWithADifferentSnapshotClosesTheScope`(coordinator_test.go:93).
- [x] 2.6.1 Add failing tests proving the production RouteSet returns every eligible family candidate without raw-score preselection, singleton uncalibrated proposals refuse, multiple symbols in one market arbitrate independently and queue overflow cannot silently drop the active-owner scope.

  **2.6.1 종결(2026-10-01 대조 감사 — 미착지).** 이름 결속: RouteSet 전수 · 원점수 미사전선택 `TestRouteSetEmitsEveryEligibleCandidateWithoutRawScorePreselection` ·
  `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently` · `TestProductionRouteCandidatesCarryNoRawArbitrationScore`; 보정 없는 단일 거절
  `TestASingletonProposalWithoutApprovedScoreAuthorityIsRefused` · engine `TestAnUncalibratedMarketRefusesEvenASingleProposal`; 한 시장 여러 종목 독립 선택
  `TestEachOwnerScopeInOneMarketGetsItsOwnSelection` · `TestSelectionOrderDoesNotDependOnSubmissionOrder`. 빈칸(넘침 fixture 에 활성 소유자 범위 0) →
  strategycoordinator `TestOverflowNeverSilentlyDropsAnActiveOwnerScope`(두 순서, 변이 X2 · X3 CAUGHT).
- [x] 2.7 Add failing worker-isolation tests for independent cadence/queue/deadline/latch, panic recovery, bounded retry/backoff, coalescing/drop accounting and no peer worker state mutation.

  **2.7 종결(2026-10-01 대조 감사 — 미착지).** 이름 결속: 칸 `TestEveryProductionLaneKeepsItsOwnSlotAndFlight` · `TestTheInboundSlotIsBoundedAndCountsWhatItDropped`; 마감 시한
  `TestTheLaneFollowsTheDesignFaultTableForEveryKindItCanExpress` · engine `TestAHungLaneDoesNotDelayItsPeersInTheSameWave`; 잠금 `TestAFaultOnOneLaneChangesNothingOnItsPeers` ·
  `TestOneLatchedLaneLeavesItsSevenPeersOpenAcrossARestart`; panic `TestAPanickingStepIsAnAbnormalFailureRatherThanACrash` · `TestAPanicOutsideALaneStepStillReachesTheMarketCycle`;
  backoff `TestALaneInBackoffWaitsItsRestartDeadline` · `TestTheBackoffLadderSaturatesAtItsCeiling`; 접힘/버림 `TestQueuePressureNeverBlocksAndNeverLosesCount` ·
  `TestTheSameDedupKeyArrivingAgainCoalescesAndCountsADrop`; 이웃 상태 무변경 `TestAFaultOnOneLaneChangesNothingOnItsPeers` · `TestProductionLanesHandsOutFreshLanesEveryTime`.
  빈칸(교차 cadence 창 · 「Flight」 미단언) → `a112_cross_lane_isolation_test.go` `TestOneLaneInsideItsCadenceWindowDoesNotGateItsPeer` · `TestOneLaneInFlightDoesNotMakeItsPeerInFlight`(변이 X4 CAUGHT).
- [x] 2.8 Add failing safety tests proving every lane-local/market-local failure and low-priority budget exhaustion leaves fill, reconcile, protection, exit and emergency reduction cadence callable and reserved.

  **2.8 종결(2026-10-01 대조 감사 — 미착지).** 이름 결속(liveness): `TestEightSimultaneousLaneFaultsLeaveTheSafetyLoopsRunning` · `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive` ·
  `TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets`. 빈칸 셋을 메움: ① 안전 생애 **다섯**(fill · reconcile · protection · exit · emergency reduction)의 cadence 를 진입 포화
  (`TestSafetyLoopsKeepTheirCadenceWhileEveryEntryQueueIsSaturated` — 셋 → 다섯) · **실패 종류 전부**(시장 오류 · 시장 panic · 레인 멈춤 · 레인 step panic · 잠긴 레인 —
  `TestSafetyLoopsKeepTheirCadenceThroughEveryEntryFailureKind`) 아래 정확히 10/10.5 로; ② 저우선 · 전략 고갈 뒤 **안전 등급 넷 전부** 허용 + 예비 불변
  (`TestEverySafetyClassStaysCallableAfterStrategyCapacityIsExhausted`, 등급은 isSafetyClass · PollClass 상수 census 로); 변이 X5 · X6 CAUGHT. 한계(명명): 안전 loop 는 자리 표시
  — 생산 loop 는 브로커가 필요하다; 중앙 결함은 설계상 안전 loop 를 세운다(`TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt`).

## 3. Breakout Evidence and Pure Lane Core

- [x] 3.1 Add additive breakout evidence kind/types and a strict canonical decoder in `internal/strategyevidence` without destructive SQLite schema changes or generic weekly/flow fallback. (L1a accepted 2026-08-17: kinds `official_closed_bar_1m`/`official_quote_l1`, strict decoders, combined constructors; `model.go` edited at exactly the two dispatch sites; no schema change — see review.md L1a sections.)
- [x] 3.2 Implement official-calendar closed 1-minute bar identity, append-only correction revisions and immutable point-in-time breakout snapshot assembly with bounded ordered inputs. (L1a accepted 2026-08-17: bar identity `(market,symbol,session_id,60000,open_at)` as `SourceRecordID`, revisions `r<n>` with supersedes chain, dual-cutoff `SealBarSeries` ≤ 512 bars with domain-separated digest golden `51d80380…`; the calendar *join/gap policy* and the producer's successor-observed guard remain L1b — task 3.6.)
- [x] 3.3 Implement `internal/breakoutlane` shared pure transition core, typed states/refusals and KR/US v1 descriptors/adapters with no broker, writable journal, Guardian, toggle or system-clock dependency.
- [x] 3.4 Implement versioned opening-range, breakout-buffer, retest/reclaim, timeout, RVOL/wick and quote veto rules with exact integer minor/PPM arithmetic and complete transition provenance.
- [x] 3.5 Implement cost/FX/fee-aware q_candidate sizing, protective stop/target validation and a single first-leg proposal seal; enforce idempotency by setup/snapshot/config identity.
- [x] 3.6 Implement the read-only official KR/US bar/quote evidence producer outside strategy refresh critical sections, using the M-B-proven lossless source, shared snapshots and scheduler capabilities rather than per-lane duplicate polling; do not fallback to WTS, float-adapted domain candles or test fixtures as production authority.
  **이연 — not-applicable in a112(2026-10-04, Manager 판정; 미구현을 완료로 읽지 말 것):** 생산자는 `internal/officialbars` 로 섰지만 생산 importer 0(8.2 census) · L1c 수락 사람 프로브 대기(review.md :753 · :774) · breakout 생산 배선은 결정 49 의 벽 뒤(`strategyproposal/a112_breakout_wall_test.go:32,60` · `breakout_fail_closed_test.go:19,34`). 배선 · 수락은 벽 해제 로트(B · C 선행 — tasks 6.4 · docs/ROADMAP.md breakout 행)로.
- [x] 3.7 Turn all breakout RED tests GREEN and prove deterministic replay across timezone, restart, correction and cache-miss fixtures with broker mutation spies at zero.
  **이연 — not-applicable in a112(2026-10-04, Manager 판정; 미구현을 완료로 읽지 말 것):** 순수 코어의 RED → GREEN · 결정적 재생은 L2 수락(review.md :70)과 breakout 덮개(:6144 · :6174)로 섰다. 생산 경로의 시간대 · 재시작 · 정정 · 캐시 미스 재생은 생산자(3.6) 배선과 결정 49 벽 해제 뒤에만 잴 대상이 있다 — 같은 해제 로트로.
- [x] 3.8 Add candidate-origin tests proving Toss rank/volume/flow is read-only discovery evidence, cannot substitute for official closed bars/session/tradability and exposes no Toss manual-condition-order mutation path.
  **닫음(2026-10-04, Manager 판정 — 감사 후 최소 보강):** 처분 표 `measurements/gate-8.1-8.3-2026-10-04/disposition-3.8-4.5.md`(감사 `audit-3.8-4.5.md`) · a · b · c(순수 코어 · 흐름 층) · e 를 새 시험으로 닫음(review.md 「3.8 · 4.5 감사 보강 로트」, 변이 10/10). **이연(not-applicable in a112):** d 거래 가능 여부 · b/c 생산 절반 — breakout 입력에 그 자리가 없고 생산 경로가 3.6(생산자 미배선) + 결정 49 벽 뒤.

## 4. Canonical Registry and Production Assembly

- [x] 4.1 Expand `pairedDescriptors`, exact validation/canonicalization and lane matching from 6 to 8 descriptors in one change, preserving existing six IDs/versions and OFF defaults.
- [x] 4.2 Extend the sealed `LaneInput` union, constructors, evaluation/proposal registries and result conversion with explicit KR/US breakout variants; reject tag/route mismatches without fallback.
- [x] 4.3 Extend production route descriptors and exact candidate validation from three to four families per market, adding family/scoring/calibration seals and rejecting legacy/partial manifests as activation authority.
- [x] 4.3.1 Replace production use of pre-evaluation `strategyrouter.Route` selection with a sealed `RouteSet` authority that validates and emits all eligible family candidates; only pure proposal evaluation followed by the coordinator's common calibrated arbiter may select across families.
- [x] 4.3.2 Add a Go-AST/import-resolution guard over exact production caller files `internal/app/engine/strategy_route_authority.go`, `strategy_proposal_authority.go`, `strategy_entry_supervisor.go` and every new `strategy_*coordinator*.go`: resolve the `github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter` import alias and forbid any selector call to its `Route` symbol while requiring at least one resolved `RouteSet` call. Legacy `Route` and callers outside this exact closure remain behaviorally unchanged.
- [x] 4.4 Extend production proposal scope validation and `buildLaneInput` with strict breakout snapshot/config construction while preserving continuation/reversal/weekly-value fixtures byte-for-byte where contracts are unchanged.
  **이연 — not-applicable in a112(2026-10-04, Manager 판정; 미구현을 완료로 읽지 말 것):** breakout 의 buildLaneInput 은 결정 49 로 `ErrBreakoutEvidenceUnavailable`(production.go :439-441) — 벽 핀 `a112_breakout_wall_test.go:32,60` · `breakout_fail_closed_test.go:19,34`. 엄격한 스냅숏 · 설정 구성은 ATR · RVOL · wick · VolumeExpanded 증거 생산 뒤에만 의미가 있고, 해제는 B(SetupID 계보) · C(소비된 setup 원장) 선행(tasks 6.4 · review.md :6251). 나머지 세 가족의 생산 구성은 그대로이며 실 적재기로 잰다(3.8 · 4.5 로트 `a112_paired_lane_batch_test.go`).
- [x] 4.5 Add paired KR/US production integration tests covering all 8 descriptors, exact route/proposal lineage, matrix migration refusal and no implicit desired/effective/LIVE activation.
  **닫음(2026-10-04, Manager 판정 — 감사 후 최소 보강 + (A)(i)):** 처분 표 `measurements/gate-8.1-8.3-2026-10-04/disposition-3.8-4.5.md` · a(벽 아래: 경로 8/8 · 흐름 8/8 · 제안 실 적재 6/6) · b · c · d · e 를 새 시험으로 닫음(변이 10/10). **이연:** 벽 너머 4.5-a(breakout 제안 · dispatch 생산 적재 — 결정 49, 해제는 B · C 선행) · MigrateLegacy 근거 제외(생산 호출자 0).

## 5. Independent Worker and Arbitration Runtime

- [x] 5.0 Expose the engine process's own `ConfigDigest`/`BuildDigest` additively on the read-only projection envelope so REST, SSE, console and the private Unix transport all show the two numbers a signed activation manifest requires; the value must survive a market-level latch overlay, a snapshot that did not come from a running engine must report absence rather than a number, and no consumer may substitute its own build [a112 결정 54, 결정 55].
- [x] 5.1 Introduce typed family/worker keys, immutable proposal envelopes, bounded coalescing queue and versioned worker runtime policy without exposing mutation capabilities. **(Closed 2026-09-02; the first three quarters landed with 5.4.2 and 5.1.1, the fourth with the runtime-policy lot below.)**

  **Where the four quarters live.** Typed family/worker key: `strategyworker.Key` — the four fields the golden froze in `worker_key_fields`, deliberately not the queue's eight-field `queue.dedup_key_fields`, because a key carrying the symbol would make the worker count grow with the universe instead of staying at eight. Immutable proposal envelope and bounded coalescing queue: `internal/strategycoordinator` (5.4.2), whose capacity is read off `strategyproposal.MaxManifestScopes` rather than chosen. Versioned worker runtime policy: `strategyworker.RuntimePolicy` in `internal/strategyworker/policy.go` — unexported fields and one `ProductionRuntimePolicy()` that all eight lanes carry, so nothing outside the package can mint a policy. That is what "server-owned" has to mean if it is to be more than a comment.

  **Its six values are read, not chosen, and the receipt is `internal/app/engine/strategy_entry_supervisor.go` — the runtime this one replaces.** Cadence 5s: the receipt is not the constant's name but the assignment — all three production worker descriptors write `PollInterval: DefaultStrategyCycleLimit` (`:346`, `:391`, `:419`), and `TestTheCadenceReceiptIsEveryProductionPollIntervalAssignment` enumerates every such assignment in every non-test engine file rather than the one file that happens to hold them today. Queue depth 1: production never sets `QueueDepth`, so the constructor fills `DefaultStrategyQueueDepth` (`:510`–`:512`); the test therefore asserts the *absence* of a choice at every `StrategyEntrySupervisorOptions` literal. Cycle deadline 30s and backoff 5s/30s: `MaximumStrategyCycleLimit`, `DefaultStrategyRestartStep`, `MaximumStrategyRestartBackoff`. All five are parsed out of the engine source and evaluated — a form the evaluator does not recognise is a failure, not a silent zero — because importing the engine would put a writable journal in this package's test closure and break the `-deps-test` walk 5.1.1 installed.

  **The sixth value, the failure threshold, is 1, and that is a refusal to loosen.** Today's engine keeps no consecutive-failure state at all: `strategyMarketRuntime` has twelve fields and none of them counts failures before the latch, so any cycle error calls `latchMarket` at once. A threshold above 1 would let a lane try again where today it is already latched — a loosening, which is a signed manifest's decision and not an implementer's. `TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` enumerates all twelve field names rather than asking whether some counter-shaped name exists, so a differently-spelled counter added to the engine still reddens it.

  **What this task does not claim.** Three of the six values are carried, not enforced: no lane schedules itself on the cadence, keeps an inbound queue, or watches the cycle deadline — that is 5.3.2, which landed 2026-09-02 and made all three enforced. The lane enforces the failure threshold and the two backoff values. And nothing in production reads any of it: `ProductionLanes()` has no production caller until 5.1.2.
- [x] 5.1.1 Introduce the explicit `FamilyWorker` and `MarketCoordinator` production types — the values themselves, provably unable to mutate anything, standing dormant with no production caller. **(Landed 2026-09-02, `4724c3d0`.)**

  **Split note, 2026-09-02.** The original 5.1.1 read "Introduce explicit `FamilyWorker` and `MarketCoordinator` production types rather than hiding four family instances behind the existing two-market `StrategyMarketWorker` contract." It is now two tasks: this one is the clause before "rather than", and 5.1.2 is the clause after it, carried over verbatim. The split moved no obligation out of the change — every word of the original text is owned by one of the two, and 5.1.2 is still open.

  `MarketCoordinator` arrived with 5.4.2 (`internal/strategycoordinator/coordinator.go:151`). `FamilyWorker` is new in `internal/strategyworker`: the four-field worker key the golden froze (`worker_key_fields`, distinct from the queue's eight-field `queue.dedup_key_fields`), the eight production workers in the golden's `descriptors` order, and one pure cycle that turns a sealed proposal into a `strategycoordinator.Envelope` or a typed refusal. Each worker is born `desired=OFF`, `effective=OFF`, `runtime=UNOBSERVED` and every one of the eight returns `DORMANT` for an input a switched-on worker emits — the control case is asserted so the eight dormant assertions cannot pass because the input was bad.

  **This closes what 5.5 handed over, for the new type.** The spec's `Worker dependency closure 에는 broker mutator, writable journal, Guardian issuer, activation/toggle writer 가 없어야 한다` is now a test, not a comment: a direct-import allow-list plus a transitive `go list -deps`/`-deps-test` walk with a positive control, falsified by importing `internal/journal` and `internal/execgw` (both fire on both walks). The reason the worker takes sealed values rather than a `strategyproposal` lane authority is that `internal/strategyproposal` imports `internal/journal`, so accepting the authority object would have put a writable journal inside the closure.

  Sixteen mutants and two closure falsifications, all caught, restored by hash. Four of the sixteen — deleting the market, lane-id, lane-version or horizon comparison — survived the behavioural suite because the family derivation and the market-prefixed lane ids shadow each other one at a time; deleting any two lets a US worker admit a KR proposal. Per-axis behavioural tests do not terminate here for the same reason they did not in 5.5, so the guard is split the same way: behaviour proves it runs and refuses, and `worker_shape_test.go` pins the five conjuncts structurally.

  **What this task does not claim.** The eight have no production caller — they are unreached code until 5.1.2. `make test` builds them and the golden-contract test binds them to the frozen `descriptors`, so they cannot drift from the contract while they wait, but nothing in production calls them. Task 5.1's remaining quarter (server-owned versioned worker runtime policy) was untouched by this lot and closed by the later runtime-policy lot; its other three quarters (typed key, immutable envelope, bounded coalescing queue) already existed in `strategycoordinator`.
- [x] 5.1.2.1 Stand the eight `FamilyWorker` instances inside the production runtime with a family cycle that is not a closure over `*Context`. **(Landed 2026-09-03.)**

  **Split note, 2026-09-03.** The original 5.1.2 read "Replace the existing two-market `StrategyMarketWorker` contract with those eight `FamilyWorker` instances, so that no family's cycle is a closure over `*Context`." This task owns the purpose clause — "so that no family's cycle is a closure over `*Context`" — and 5.1.2.2 owns the replacement of the *gate*. Every clause of the original is owned by exactly one of the two.

  **The split is forced by two authorities that predate it, not chosen.** `analysis/goldens/four-family-runtime-v1.json` freezes all eight descriptors at `desired: OFF, effective: OFF, runtime: UNOBSERVED`, and the spec requires that "Legacy 3-family approval은 4-family activation으로 자동 승격되어서는 안 되며 (MUST NOT)". A dormant `FamilyWorker.Run` returns `DORMANT` before it looks at anything, so making the eight the gate today would take production entry to zero — the toggle-OFF invariant says the new runtime off must equal upstream behaviour. Promotion needs a signed activation manifest, which is section 8 and 5.1.2.2.

  **What landed.** `internal/app/engine/strategy_lane_runtime.go`: a process-lived `strategyLaneRuntime` holding `strategyworker.ProductionLanes(clk)` (8), reached through `Context.productionStrategyLanes` the way `strategyDispatchOwner` already is — process-lived because a lane's latch and failure counter are memory, and rebuilding them each refresh reopens a lane that latched for a reason that still holds. `runProductionStrategyMarketCycle` now runs this market's four lanes after the refresh and outside `c.strategyRefreshMu`, feeding each the sealed proposal it says is its own (`Lane.Owns`, one new delegating method — the judgement stays in `FamilyWorker.owns`).

  **The boundary this moved, stated as what a test can read.** The thing a lane runs is `strategyFamilyLaneStep`, a package-level function whose only parameter is `*strategyworker.Lane`. What it can touch is fixed by that parameter's type, not by a comment: `strategyworker`'s `dependency_closure_test.go` walks `-deps`/`-deps-test` and proves that package's closure holds no broker mutator, writable journal or Guardian issuer. The market cycle still holds `c.Journal.CurrentPositionCampaignCAS` and `fresh.dispatch.dispatch`, and that is correct — the spec requires exactly one mutation authority per market and that function is it. What moved is the *family's* cycle, not the market's.

  **Why the proof is a census and not a behavioural test.** All eight are dormant, so no execution can show what a lane could have touched. `TestOnlyThePackageLevelStepEverRunsInsideALane` parses every non-test file in the engine package, enumerates every value handed to `Lane.RunBounded`, and freezes the list at one entry. That is the same answer 5.5, 5.1.1 and 5.6.1 reached — one test per axis does not terminate, so raise the counting scope. `TestTheMarketCycleRunsItsLanesAndTheRefreshDoesNot` freezes the call sites of `evaluate` the same way, which is what keeps the lane stage out of the shared refresh mutex.

  **`evaluate` returns nothing.** An earlier draft returned the observations and the emitted envelopes; a caller can drop either and nobody sees it. Same defect and same fix as `Deliver` in this very function (5.5-fix2): remove the ignorable answer rather than test for its use. Results live in the runtime and `observations()` reads them.

  **One census of this change's own making had to be repaired first.** `a112_central_integrity_census_test.go` (5.6.1) froze absolute line numbers, so a sixteen-line insertion into an unrelated function above `Run` broke it, and the only way to pass would have been to edit the expected list — indistinguishable from the edit the census exists to catch. It now records each occurrence as an offset from its own enclosing declaration, which is stable under other people's edits and still catches a new mint (mutant M12) and a moved one.

- [x] 5.1.2.2 Replace the existing two-market `StrategyMarketWorker` gate with those eight, once a signed four-family activation manifest can promote them. **(Landed 2026-09-03.)**

  **Why it is open.** Production runs `StrategyMarketWorker`, whose `Cycle` is a closure over `*Context`; the regenerated AST artifact for `Context.runProductionStrategyMarketCycle` still enumerates `c.Journal.CurrentPositionCampaignCAS` at `:462` and `fresh.dispatch.dispatch` at `:469`. That is the market's one mutation authority and it stays; what is still missing is that the eight decide whether anything reaches it. Today they cannot, because they are all `effective: OFF` and nothing in `strategyworker` can mint an ON worker from outside the package — deliberately, since "server-owned" would otherwise be a comment.

  **Two things 5.1.2.1 measured that this task must carry.** The lanes see the *arbitrated* proposal for their scope, not the pre-arbitration fan-out — `coordinateMarketProposals` still submits to the coordinator directly, so moving the gate means moving the lanes upstream of arbitration. And the fault-stream balance 5.6.1 found (`cap(faults) == 2 == markets`) is untouched here because lane faults stay inside `strategyworker.Lane`; routing them to the supervisor's stream is what would break it.

  **이 태스크의 선결 조건은 8.7.1 이고 같은 로트에서 함께 랜딩한다 [a112 결정 59].** 사람이 2026-09-03 에 세 선택지 중 "매니페스트 + 5.1.2.2 한 로트, 5.2.2 는 그다음" 을 골랐다. 이유는 두 방향이다: 매니페스트만 먼저 랜딩하면 소비자가 없고(5.1 이 이미 피한 모양), 셋을 한 로트로 묶으면 서명 검증·관문 교체·진입 상한 해제를 한 번에 리뷰해야 한다(5.5 가 적대 리뷰 13 라운드 걸린 규모).

  **Done.** A lane that is not effective stops its family's proposal from reaching the coordinator, the promotion comes from a signed manifest rather than from legacy approval, and `TestEveryLaneStaysDormantOnAProposalItActuallyOwns` is replaced by a test that shows both states rather than deleted.

  **관문이 어디에 서는가, 그리고 왜 거기인가.** `coordinateMarketProposals` 안, 조정자 `Submit` **앞**이다. 뒤에 세우면 중재가 이미 한 범위의 승자를 골라 버렸고, 그 승자의 레인이 잠겨 있으면 그 범위는 이웃 가족이 이길 수 있었는데도 통째로 닫힌다. 5.1.2.1 이 "관문을 옮기려면 레인을 중재 앞으로 옮겨야 한다" 고 적어 둔 그 자리다. 관문이 하는 일은 그 가족의 레인에게 묻는 것 하나이고(`Lane.Owns` → `Lane.Run`), 레인이 낸 봉투를 그대로 조정자에 넣는다.

  **두 권위의 겉보기 충돌을 여기서 푼다.** 5.1.2.1 은 "오늘 여덟을 관문으로 세우면 생산 진입이 0 이 되고 그것은 토글 OFF 불변식(§0-2)에 어긋난다" 고 적었고, 스펙은 반대로 "required activation authority 가 missing 이면 broker exposure-raising request 는 0건이어야 한다 (SHALL)" 고 적었다. 둘을 함께 만족시키는 유일한 모양이 **관문을 활성화된 런타임 안에만 세우는 것**이다: 활성화가 없으면 기존 시장 단위 경로가 그대로 돌고(= upstream 동작 보존), 있으면 그 시장의 넷이 각자 판정한다. design 이 partial 3-of-4 를 시장 전체 OFF 로 못 박았으므로 활성화된 시장의 서술자는 언제나 넷이고, 그래서 관문이 실제로 막는 것은 **잠긴 레인 하나**다 — 그 가족의 제안만 멈추고 이웃 셋은 계속한다. 그것이 이 change 가 사려던 것이다(`design.md:3`, "시장 장애 격리이지 전략군 장애 격리가 아니다").

  **순서가 안전이다: 레인을 제안 수집 앞으로 옮겼다.** 관문은 레인의 잠금을 읽어 판정하고 그 잠금은 원장에서 태어난다(5.3.3). 뒤에 세우면 재시작 뒤 **첫 주기**에 durably 잠긴 레인이 열린 것으로 읽히고 그 가족의 제안이 조정자에 닿는다. 그 창은 한 주기뿐이라 어떤 행동 시험도 우연히 잡지 못하므로 순서를 구조로 못 박았다(`TestTheLanesAreBuiltBeforeTheProposalsAreCollected`).

  **세 상태를 함께 세운다.** 원래 시험을 지우지 않고 교체했다. 활성화 없음 → 관문이 서지 않고 조정 결과가 관문 이전과 같은 값이다. 활성화 + 승격 → 그 가족의 제안이 조정자에 닿는다. 활성화 + 일부만 승격 → **관문이 순위를 바꾼다**(점수 1위의 레인이 안 켜졌으면 2위가 이긴다). 활성화 + 잠김 → 그 가족만 멈추고 2위가 이긴다. 마지막 둘이 "관문이 통과 도장이 아니다" 의 증거다 — 시장이 닫혔는지만 재면 승자를 잠갔을 때 아무도 못 이기는 구현도 통과한다.

  **봉투가 둘 있고 그 둘이 같은 값인지를 잰다.** 관문이 서지 않은 갈래는 엔진이 만든 봉투를, 선 갈래는 레인이 만든 봉투를 넣는다. 두 사본이 갈릴 수 있으므로 구조를 단언하지 않고 **값을 견준다**(`TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`) — 필드가 하나 늘 때 구조 단언은 통과하고 이 등식은 실패한다.

  **시험 seam 을 하나 만들었고, 생산 경로가 지도자가 되는 실행을 함께 두었다.** `loader.loadActivation` 은 같은 로더의 `load` 필드와 같은 관례다. seam 만 있으면 "seam 을 건너뛴다" 변이가 모든 시험을 통과하므로, `TestTheProductionActivationLoaderRunsAndFindsNoManifest` 가 그 필드를 nil 로 둔 채 실제 구현을 돌린다. 그 시험은 결속을 **전부 유효하게 채운다** — 하나라도 비우면 "파일이 없다" 대신 "설정이 어긋났다" 를 재게 되고 파일 부재라는 축은 한 번도 안 재진다.

  **이 태스크가 주장하지 않는 것.** 시장 단위 단일 제안 상한은 그대로다(5.2.2). 기존 시장 단위 경로를 지우지 않았다 — 활성화가 없을 때의 갈래로 남고, 지우는 것은 5.2.2 와 6 절이다. 레인 고장은 아직 감독자 fault 스트림에 안 간다(5.6.1 의 `cap(faults)==2==시장 수` 등식이 그대로다). 그리고 오늘 생산에는 서명된 매니페스트가 없으므로 **생산 동작 변화는 0** 이다.
- [x] 5.2.1 Keep remote evidence refresh outside the shared assembly mutex. **(Landed 2026-09-03.)**

  **Split note, 2026-09-03.** 원문 5.2 는 두 절이다: (a) "market-level single-proposal readiness 를 시장마다 네 개의 독립 감독 lane worker 로 교체"와 (b) "remote evidence refresh 를 공유 assembly mutex 밖에 둔다". 이 태스크가 (b) 를, 5.2.2 가 (a) 를 가진다. 원문의 모든 절은 둘 중 정확히 하나가 가진다. 가른 이유는 (a) 가 5.1.2.2 와 같은 사람 결정(서명된 4-family 활성화 매니페스트)에 걸려 있는데 (b) 는 아무것에도 안 걸려 있고, 오늘 재서 확인한 실제 결함이기 때문이다.

  **잰 결함 둘.** `refreshPairedStrategyEntryProductionAssembly` 는 `c.strategyRefreshMu` 를 들고 `NewPairedStrategyEntryProductionAssembly` 전체를 돌았다. 그 함수는 official 달력(`TypedMarketCalendar`)과 official FX 를 타고 후보 DB·저널·evidence DB 를 읽는다. CodeGraph 로 이 함수의 호출자는 `runProductionStrategyMarketCycle` 하나이고 그것이 시장마다 하나씩이므로, 경쟁하는 goroutine 은 정확히 둘이다 — KR 의 느린 official 응답이 US 주기 전체를 세웠고, 그 주기는 5.1.2.1 이후 레인 평가까지 들고 있다. 둘째, `strategyRefreshAt` 에 넣는 값이 파도의 **완료** 시각이 아니라 **시작** 시각이라 1초 창은 파도가 1초보다 오래 걸리면 아무도 못 태운다: 3초짜리 파도가 끝나는 순간 캐시는 이미 3초 묵었고, 잠금을 물려받은 두 번째 시장은 창 밖이라 자기 파도를 처음부터 다시 돌았다. **원격이 느릴수록 — 합치는 것이 가장 필요한 때 — 합치기가 정확히 꺼졌다.**

  **고친 모양.** 잠금은 상태 전이만 지킨다. `joinStrategyRefreshWave` 가 잠금 안에서 세 답 중 하나를 준다(신선한 캐시 / 도는 파도에 합류 / 지도자). 파도는 잠금 밖에서 돌고, 기다리는 시장은 mutex 가 아니라 채널에서 기다린다 — 그래서 자기 주기가 취소되면 빠져나온다(오늘 `Lock()` 에 걸린 시장은 ctx 를 못 본다). **1초 창의 의미는 바꾸지 않았다**: 여전히 시작 시각을 잰다. 완료 시각으로 옮기면 캐시 수명이 파도 길이만큼 늘어나는 완화가 되고, 이 태스크는 그것을 요구하지 않는다. 합치기는 창이 아니라 파도가 한다.

  **RED 은 셈이다.** 이 패키지에서 파도를 느리게 만들 방법이 없다(원격은 실제 official client 뒤에 있고 시계로 못 늦춘다). 그래서 편집 **전** 소스에서 실제로 실패한 것은 두 구조 시험이다: 잠금을 만지는 함수와 원격 파도를 부르는 함수가 서로소여야 한다는 것, 그리고 파도를 도는 자리가 하나여야 한다는 것. 잠금을 만지는 함수의 **호출 목록까지** 얼린다 — 하나만 얼리면 잠금 안에서 헬퍼를 부르고 그 헬퍼가 원격을 부르는 한 다리 우회가 남고, 그것이 5.5 의 적대 리뷰가 실제로 뚫은 방법이다.

  **행동 절반은 `testing/synctest` 로 잰다.** sleep 이 하나도 없다. 그리고 이 도구가 여기서 특히 맞는 이유를 쟀다: **뮤텍스에 걸린 goroutine 은 "확실히 멈춰 선" 것으로 치지 않는다**(Go 1.26 측정 — `sync.Mutex.Lock` 에 걸린 goroutine 이 있으면 `synctest.Wait()` 가 영원히 안 돌아온다). 즉 이 시험들은 "채널에서 기다린다"와 "잠금 뒤에 줄 서 있다"를 구별한다.

  **반증 16개 전부 CAUGHT, 대조군 2개 SURVIVED(의도).** 두 개는 두 번째 판이 필요했다. `M7`(지도자가 파도를 만들고 발표하지 않는다)은 첫 배터리에서 **살아남았다** — 위 시험이 전부 *시험이 지도자*라 생산 경로가 지도자가 되는 실행이 하나도 없었기 때문이다. 이름을 얼리는 셈으로도 못 잡는다(그 변이는 수집을 원래 자리에 그대로 둔다). `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` 이 그 구멍을 막는다. `M10b`(발표가 `done` 을 먼저 닫고 값을 나중에 쓴다)는 **`-race` 없이는 초록이고 `-race` 로만 빨갛다** — 5.7 의 N2 와 같은 모양이 이번엔 `internal/app/engine` 에서 나왔다.

  **그래서 검출기를 이 패키지에 배선했다.** 2026-09-03 측정: `go test -race -tags tossos_testseams ./internal/app/engine` 은 **14분 46초**(기존 일곱 패키지 전부가 11.9초)라 통째로는 못 들어온다. 그래서 이 로트가 만든 동시성 시험만 이름으로 골라 돈다 — **5.7초**. 이름 고르기의 실패 방식(나중에 늘린 시험이 목록 밖으로 남는 것, a118 이 겪은 것)은 `tools/sdd/test_race_detector_actually_runs.py` 가 그 파일의 `Test` 함수 전부와 목록을 대조해 막는다.

  **덤으로 고친 것 하나.** 5.6.1 이 BTM 에 적어 둔 커버리지 블록 번호 29개가 5.1.2.1(+16)과 이 로트(+3) 뒤 19줄 밀린 채였다. 산술로 옮기지 않고 프로파일을 다시 떠서 대조했다: 28개가 정확히 +19 자리에 있었고 `count` 도 전부 일치했다. 남은 하나는 실제 블록이 `804-806` 인데 `785-786`(= 804-805)로 적혀 있었다 — 옮겨 적을 때 한 줄 어긋난 것이고, 잰 값으로 바꿨다.

  **이 태스크가 주장하지 않는 것.** 시장 단위 단일 제안 준비 상태는 그대로다(5.2.2). 7.5 의 "no remote I/O under strategy refresh mutex" 성능·운용 시험은 여기 없다 — 이 로트가 넣은 것은 구조 셈과 동시성 행동 시험이고, 부하 아래 지연을 재는 것은 7.5 다. 그리고 `internal/app/engine` 의 나머지 동시성은 여전히 검출기 밖이다.
- [x] 5.2.2.1 Move the handoff capacity from the market to the owner scope, in activated markets only. **(Landed 2026-09-30.)**

  **Split note, 2026-09-30 (Manager 판정 — 분할 (가)).** 원문 5.2.2 는 「Replace market-level single-proposal readiness with four independently supervised lane workers per market」과 아래 Done 네 문장이다. 상한을 경계에서 들어내도 하류의 네 권한(결과 권한 `ResultAuthority` · 위험 · 계좌 `collectMarket` B1 · 1차 레그 `collectStrategyFirstLegAuthority` B2)이 여전히 시장당 제안 하나를 요구하므로, 두 소유자 범위 시장이 실제로 거래하려면 그 넷을 소유자 범위 단위로 옮겨야 한다. 그중 1차 레그의 다섯 줄(`strategy_account_first_leg_authority.go` :217 · :221–:225)은 결정 (1) 에 따라 **L6 6.2 봉인 전까지 유일한 방어**라 봉인 전에는 손대지 않는다. 그래서 셋으로 갈랐다. 원문의 문장마다 소유자는 정확히 하나다:

  | 원문 문장 | 소유 |
  |---|---|
  | 「소유자 범위마다 최대 하나가 bounded handoff 를 기다린다」 | **5.2.2.1** |
  | 「상한을 실제로 올리는 편집은 매니페스트가 선 뒤에만 한다」 | **5.2.2.1**(서명 활성화된 시장에서만 상한 상향) |
  | 「시장의 준비 상태가 "이 시장에 제안이 정확히 하나"가 아니라 네 레인 각자의 준비 상태에서 나오고」 + 제목의 「Replace market-level single-proposal readiness」 + 들어냄의 결과인 「두 소유자 범위 시장이 거래한다」 | **5.2.2.2** |
  | 「각 레인이 자기 cadence·큐·마감·health·latch 로 독립 감독되며」 + 제목의 「four independently supervised lane workers」 | **5.6.2.2**(레인 고장이 감독자 fault 스트림으로 가는 편집과 같은 편집 — 5.6.2.1 이월 표) |

  **무엇이 바뀌었나.** `strategyhandoff.AdmitEachOwnerScope` — 선택마다 `Admit` 을 한 번씩 불러 handoff 를 **여러 개** 돌려준다. 값이 나가는 문(`Single` · `Deliver`)과 `Capacity=1` 은 그대로라서, 같은 1 이 서명 활성화된 시장에서는 **소유자 범위당** 1 이 된다(HANDOFF 의 「상수와 Single 서명을 함께 바꾼다」 예고 대신 개수를 늘린 이유 — review). 시장 단위 판정은 쪼개지 않는다: 닫힘 · 선택 없음은 handoff 하나, 같은 소유자 범위(계좌 · 시장 · 종목 · 포지션 세대, `strategyrouter.OwnerKey` 와 같은 정규화)에 둘이 실리면 시장 전체 `OverCapacity`. 엔진은 `dispatchHandoffs` 가 **서명 활성화가 있을 때만** 이 문을 쓰고(없으면 오늘의 `dispatchHandoff` 하나), 주문 경로 `runProductionStrategyMarketCycle` 은 `deliverEachStrategyHandoff` 로 조정자 순서대로 건네며 **첫 오류에서 멈춘다**(보수 방향). 공유 dispatch 호출 자리는 여전히 그 함수 하나다.

  **오늘-동등성.** 생산에 서명 매니페스트 0건이라 모든 시장이 활성화 없는 갈래 — 토글 OFF = upstream 동작 불변. 활성화된 두 소유자 범위 시장도 경계는 지나지만 1차 레그 개수 관문(B2 `len(proposal.entries) != 1`, 문구 `paired production authority is incomplete for market`)이 거절해 주문 0 이다 — `TestTwoOwnerScopesStillPlaceNothingBecauseTheFirstLegGuardRefuses` 가 이를 못 박고, 같은 조립에서 범위 하나면 Gateway 스파이까지 닿는 대조로 거절을 개수 조건에 귀속시킨다.

  **RED · 반증.** `analysis/measurements/lot-5.6.2-5.2.2/red-5.2.2.1.log`(넷 FAIL — 범위당 handoff 부재, 대조 하나 GREEN). 변이 `analysis/harness/a112_lot_mutate.py --set 5.2.2.1` — review 절.

  **리뷰 수리(2026-09-30, 4목소리 — codex BLOCK · A APPROVE · B BLOCK · C APPROVE).** 처분 전표와 수리 내용은 review 「2026-09-30 태스크 5.6.2.1 · 5.2.2.1 적대 리뷰 처분 · 수리 로트」 절. 요지: 소유자 범위에 horizon · 레인이 섞이면 잡는 시험, 전달 몸통의 의미 무변경 이동 + 스파이 구동 + 식별자 해소 못, 경계 이름 허용 목록(「자동으로 새 문을 본다」 철회), 패키지 내부 주조 census, `.Single()` 전수 세기, 같은 범위 중복의 엔진 핀, 두 순서의 동등성 핀, 5.6.2.1 변이 원장 재실행 · 커밋.

  **이 태스크가 주장하지 않는 것.** 두 소유자 범위 시장의 거래(5.2.2.2). 레인별 독립 감독(5.6.2.2). 시장 준비 상태의 레인 유도 — `buildProductionStrategyMarketWorker` · `ResultAuthority` · projection 은 여전히 시장 단위 `dispatchHandoff().Single()` 을 읽는다(5.2.2.2).
- [x] 5.2.2.2 Move the downstream authorities to owner-scope units so a two-owner-scope market trades. **(Owns the title's readiness clause and 「두 소유자 범위 시장이 거래한다」 — split 2026-09-30.)**

  **착수 조건 — L6 6.2 봉인 완료.** 결정 (1)(HANDOFF 「결정 (1)(5)(6) 기록」): 1차 레그 권한의 다섯 줄(`strategy_account_first_leg_authority.go` :217 `len(proposal.entries) != 1` · :221–:225 identity 대조)은 봉인 전 방어이고, 6.2 가 그 자리를 봉인으로 대체한 뒤에만 바꾼다. 6.2 봉인 전에 이 태스크를 시작하지 않는다.

  **착수 조건 둘째 — 같은 범위 중복 핀이 서 있을 것**(2026-09-30 리뷰 보이스 C #2). B2 를 걷어 내면 「같은 소유자 범위의 봉인된 제안 둘」 경로가 곧바로 주문으로 이어진다. 그 경로를 엔진 쪽에서 막는 핀 `TestTheSameOwnerScopeSealedTwiceRefusesTheActivatedMarket`(활성화 시장 → handoff 하나 · OverCapacity)이 착수 시점에 초록이어야 하고, 이 태스크의 편집 뒤에도 초록이어야 한다.

  **기준선의 성격**(보이스 A #1). 5.2.2.1 의 오늘-동등성 핀은 **생산 모양이 아니다** — 위험 · 계좌 권한이 범위 하나짜리 fixture 라 두 범위여도 Ready 로 남는 의도적 최악 조건이고, 「B2 개수 조건이 유일한 방어」는 fixture 순서에서만 참이다(조정자 순서에서는 :221 · :228 이 대신 막는다). 생산에서는 결과 권한 · 계좌 B1 · 위험 권한 재수집도 두 범위를 거절한다. 이 태스크는 두 순서의 핀을 **의도적으로** 뒤집되 생산 모양의 두 범위 시험을 새로 세운다.

  **옮길 것.** 결과 권한(`ResultAuthority` — `strategy_proposal_authority.go` 의 `dispatchHandoff().Single()`), 위험 권한, 계좌 권한(`collectMarket` B1 `len(proposal.entries) != 1`), 1차 레그 권한(B2 · identity 대조), 그리고 worker 승격(`buildProductionStrategyMarketWorker`)과 projection 의 시장 단위 `Single()` 읽기. 5.2.2.1 의 오늘-동등성 핀(`TestTwoOwnerScopesStillPlaceNothingBecauseTheFirstLegGuardRefuses`)이 「무엇이 바뀌는가」의 기준선이다 — 이 태스크가 그 핀을 **의도적으로** 뒤집는다. `deliverEachStrategyHandoff` 의 「첫 오류에서 멈춤」을 소유자 범위별 고장 격리로 바꿀지도 여기서 정한다 — **결정 항목: 굶음**(보이스 A (T)): 조정자 사전순으로 앞선 범위가 매 주기 실패하면 뒤 범위는 매 주기 굶는다(안전 방향이지만 liveness 결함).

  **6.2.0 의 효력 범위(2026-10-01, 6.2 리뷰 보이스 A #4 — 인용할 때 과대 읽지 말 것).** 6.2.0 의 소유자 범위 선택은 개수 관문이 1 인 동안 **거절 문구만** 바꾸고 수락 집합은 편집 전과 같다. 봉인의 보호 효과는 이 태스크가 개수 관문을 걷는 순간부터 생긴다 — 그래서 이 태스크의 편집 뒤에 위조 축 행동 시험(`a112_first_leg_owner_scope_seal_test.go`)이 **두 범위 쌍에서** 다시 초록이어야 한다.

  **편집 로트(2026-10-01 — 리뷰 전, 미착지).** Manager 판정 J1~J5 · B(스키마 핀은 별도 change). 하류 권한(결과 · 위험 · 계좌 · 1차 레그 · worker 승격 · projection)을
  `dispatchHandoffs` 목록 · 소유자 범위 키로 옮기고 1차 레그의 시장 단위 개수 관문을 지웠다. 범위 권한이 없으면 그 범위만 타입 거절(`strategyScopeRefusal`), 전달 몸통은
  그 타입만 건너뛰고 나머지 오류에서 멈춤. 통화 재유도(A#3) · 범위별 계좌 적재(A#5) · digest 단일 출처(A#6). Done 시험 `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`
  — **범위별 발급은 파도 순차**(같은 파도 둘째 범위는 공유 버킷 사용량 CAS `BUCKET_USAGE_STALE` 로 거절 — (e) 보호의 설계), 위조 다섯 축은 두 범위 쌍에서 재실행.
  생산 동작 변화 0(서명 활성화 0). 변이 21 CAUGHT · M20 예상 SURVIVED. 잔여 R1~R5(review 「5.2.2.2 편집 로트」) — R1 스키마 핀 27 은 **두 자리**(riskbucket · strategyrouter route — 2026-10-01 a127 조사) **레인 활성화의 경성 선행**(ROADMAP, a127 이 함께 수리).
  착지 `80ae96a5`. **리뷰 라운드(2026-10-01): A · codex BLOCK, B APPROVE** — 처분 · 수리는 review 「5.2.2.2 리뷰 라운드(80ae96a5)」 절: 활성화 없는 시장의 개수 관문 복원
  (codex #1), J4 = (A) 적재 단계 원인 분류(riskbucket 범위 국소 sentinel + 엔진 원인 운반 — A #1 · codex #2), 승격 근거 범위 권한 · 최소 만료 · As census · M20 격추 · 첫 파도 두 레그
  단언 · 문서 정정. 수리 변이 21 CAUGHT(M14 예상 생존) — 착지 `face8d0d`. codex 재확인 #1: 계좌 경계 P1 → 판정 (A)(계좌 매니페스트는 시장 단위 — 계좌 실패는 언제나 결함), 2차 수리 변이 5/5 CAUGHT — 착지 `bf269eb5`. **codex 재확인 #2 APPROVE → 체크(2026-10-01).** 잔여 R1~R5 · 새 명명 잔여는 review 「5.2.2.2」 절들.

  **이월(6.2 봉인 리뷰, 2026-10-01 — 이 태스크가 개수 관문을 걷을 때 함께):** (a) 발급 통화를 봉투(`accepted.currency`)가 아니라 `result.Lineage.Market` 에서 다시 유도(보이스 A #3 — 오늘은 Guardian 이 fail-closed 로 막음), (b) 계좌 권한도 선택된 소유자 범위 단위로 다시 유도(보이스 A #5 — 계좌 권한은 오늘 `entries[0]` 종목으로 적재되고 선택 범위와 대조되지 않음), (c) 제안 집합 digest 식을 한 곳으로(`collectMarket` 인라인과 `strategyProposalSetDigest` 사본 — 보이스 A #6, 갈라지면 fail-closed). (d) 변이 하네스 대조군의 JSON 추가 실행이 종료 코드도 보게(6.2 codex 재확인 #2 P2), (e) 봉인 시험 파일의 옛 주석 한 줄(「entries[0] 이면 identity 거절」) 삭제((T)). 그리고 codex 의 말 그대로: 6.2.0 의 APPROVE 는 **개수 관문 제거의 안전성을 승인한 것이 아니다.**

  **이월(2026-10-01, 5.2.2.1 리뷰 codex 4차 — review 끝 절).** (1) 게이트 · 하네스가 종료 코드가 아니라 이름 붙은 시험의 실행을 확인(`init` 조기 종료로 시험 이진이 `ok` 가 되는 경로 — 동결 시험을 포함한 모든 시험의 일반 부류), (2) strategyhandoff 소스 동결에 비`.go` 빌드 입력 포함 또는 금지, (3) digest 재고정 커밋에 독립 리뷰 기록 결속. 6.2 와 공유.

  **Done.** 서명 활성화된 두 소유자 범위 시장이 범위마다 주문을 낸다(소유자 범위마다 최대 하나), 시장 준비 상태가 레인 · 범위의 준비 상태에서 나오고, 활성화 없는 시장은 여전히 시장 단위다.
- [x] 5.3.1 Implement the lane-local health/failure counters, bounded retry/backoff and the entry-only latch. **(Landed 2026-09-02.)**

  **Split note, 2026-09-02.** The original 5.3 read "Implement lane-local single-flight cadence, monotonic deadline, health/failure counters, bounded retry/backoff and durable entry-only latch/recovery conditions." It is now three tasks: this one owns "health/failure counters, bounded retry/backoff and [the] entry-only latch", 5.3.2 owns "single-flight cadence, monotonic deadline", and 5.3.3 owns the word **durable** together with "recovery conditions". Every clause of the original is owned by exactly one of the three; 5.3.2 landed 2026-09-02 and 5.3.3 is open. The seam is placed where it is because this lot's latch is in process memory: it survives nothing.

  `strategyworker.Lane` in `internal/strategyworker/lane.go` is the state machine, one per production worker, held apart from `FamilyWorker` because a worker is a value that remembers nothing and a failure count is a memory — folding them together would make every copy of a worker silently copy its fault state too. `ProductionLanes()` builds eight fresh lanes on every call; a package-level set would hand a once-latched lane to every caller for the life of the process, and a mutant that does exactly that is caught. No lane holds a reference to any peer, so the golden's `peer_lane_state_mutation_forbidden` is structurally true here and a fault on one lane is measured to leave the other seven `HEALTHY`.

  Health is three values, not two: `HEALTHY`, `DEGRADED` (failures are stacking but entry is still open) and `LATCHED`. A lane about to latch must not read as healthy. The latch itself follows the design's fault table rather than the engine's single behaviour — an abnormal cycle (panic, unexpected return) latches at once, an ordinary error counts and retries — while the *production* threshold stays at 1 so today's behaviour is unchanged (see 5.1). `Succeed()` clears the counter and deliberately cannot clear the latch: an entry-only latch needs recovery evidence, and this package has no way to make any. A second failure on a latched lane mints no second latch record, so the operator keeps the *first* cause rather than the last.

  Twenty-one mutants, all caught, restored from a hashed backup rather than `git checkout` (the lot is uncommitted; `checkout` would delete the GREEN with the mutant). Two of the twenty-one survived their first form and both are recorded because the reason matters. One did not reach the thing under test — the "shared lanes" mutant still returned a fresh set — which is the trap this change already wrote down. The other was real: `attempt >= steps` mutated to `attempt > steps` stayed green because the production ceiling (30s) is an exact multiple of the step (5s), so the two comparisons agree at the only attempt that separates them. The defect lives in divisibility, not in size, so a 7s-step/30s-ceiling ladder is now a committed test with the falsifying cell (attempt 4 is 30s, not 28s) written out by hand rather than re-derived from the implementation.

  **What this task does not claim.** The latch is in memory. A restart clears it, and nothing here writes or reads a durable latch record — `Fault` is a read-only observation with no way to restore entry, and giving it one would put a latch writer inside a closure that is proven not to have one. Nothing calls `Fail`/`Succeed` in production; the cycle that would call them is 5.1.2/5.2.
- [x] 5.3.2 Implement lane-local single-flight cadence and monotonic deadline. **(Landed 2026-09-02.)**

  **Second split note, 2026-09-02.** 5.3.2 as written owned three clauses: "single-flight cadence", "monotonic deadline", and the word **durable** together with "recovery conditions". This task now owns the first two; the third is 5.3.3 and is open. The seam is not a matter of taste. `internal/strategyworker` is the package whose import closure is *proven* to contain no writable journal, no broker mutator and no toggle writer — a durable latch record needs a writer, so building one here would delete the property 5.1.1 installed. `design.md:223` already assigns "lane health/latch projection" to `internal/app/engine`.

  `strategyworker.Lane` now owns a clock (`internal/clock`, the repository's single injected time source) and five gates, in this order: latch → in-flight → backoff → cadence → empty slot. Order is contract. The latch comes first so a latched lane does not silently *eat* the trigger it is holding; the engine reads its queue first and refuses afterwards (`runMarket:779` → `:800`), which is safe there only because the engine's trigger carries nothing and this one does.

  **Single-flight is now state, not wiring.** The engine gets single-flight for free from having exactly one consumer goroutine per market — a property that is true of the *wiring*, so it disappears the moment a second driver exists, and 5.2 introduces per-family drivers. The lane holds `inFlight` itself, and a mutant that removes it reddens.

  **The cadence clock starts at the cycle's start, not its end.** The receipt is `runStrategyPoller` (`:719`), which sleeps immediately after attempting the enqueue rather than after the cycle returns. Measuring from completion would let one slow cycle push out every period behind it.

  **The deadline is the engine's watchdog, transcribed rather than called.** Calling it would mean importing `internal/app/engine`, which drags the journal and the gateway into this package's closure. So `invokeBounded` is a copy, and two AST tests keep the copy from drifting: one pins the engine's three select arms and their exact returns, the other pins this package's. Arm *order* is invisible to behavioural tests — when several channels are ready Go picks at random — which is why it is pinned structurally.

  **One divergence between two authorities, recorded rather than resolved.** The engine returns `abnormal=true` for a deadline (`:897`); `design.md:198` puts "deadline/ordinary error" on the *count-and-retry* row. This lot followed the engine, because with the production threshold at 1 the two readings produce identical behaviour today, and where they diverge (threshold > 1) the engine's is the more conservative. The AST receipt test is the place a human will see it if the engine ever changes its mind.

  **What the engine does not have and this lane does: a drop counter.** `enqueueStrategyPoll` returns `StrategyTriggerFull` and nothing anywhere counts the refused trigger. The golden's `queue.overflow` requires "typed refusal **and** bounded drop counter". `Lane.Dropped()` is that counter, saturating like the others.

  **Falsification: 27 mutants, 27 caught.** Two needed a second form. `M18` (the watchdog sleeps for the *cadence* instead of the deadline) survived its first battery, because advancing a fake clock by 30s also wakes a 5s sleeper — the defect lives in *when* the watchdog wakes, not *whether*, so the committed test now advances by less than the deadline and asserts nothing fired. `M26` (all eight lanes share one set) first failed to compile, which is not a catch; rewritten as a compiling package-level cache it is caught by the existing freshness test.

  **What this task does not claim.** The latch is still process memory (5.3.3). The lane's inbound slot holds *triggers*, not envelopes — the golden's `queue.same_key`/`dedup_key_fields` coalescing lives in `strategycoordinator` and is deliberately not re-implemented here, because two copies of one judgement give the operator two diagnoses. Nothing in production calls any of it; the driver is 5.1.2/5.2.
- [x] 5.3.3 Implement the durable entry latch and its recovery conditions. **(Landed 2026-09-03.)**

  **Why it is open and where it must live.** `Lane.Latched()` is a bool in process memory: a restart clears it, so a lane that latched for a reason that still holds comes back open. Closing that needs a record that outlives the process and a writer for it — and this package is *proven* to have no writer (`dependency_closure_test.go` walks `-deps` and `-deps-test`). `Fault` is deliberately read-only and carries no way to restore entry; giving it one would put a latch writer inside the closure that is proven not to have one. `design.md:223` assigns "lane health/latch projection" to `internal/app/engine`, and `design.md:199` says the recovery condition is *evidence*, not a successful cycle — which is why `Succeed()` clears the counter and cannot clear the latch.

  **Done.** A lane is born from a durable record rather than born unlatched; the record names the first cause, not the last; restoring entry requires recovery evidence that some component outside this package's closure produces; and a restart of the engine leaves a latched lane latched. Note today's engine has no durable latch either (`strategyMarketRuntime.latched` is a struct field), so this is new behaviour, not a transcription — it needs its own receipt, and that receipt is the signed manifest, not the current code.

  **어디에 살고, 왜 거기인가.** 기록은 원장(schema v32, 두 append-only 테이블)이다. 고른 것이 아니라 잰 것이다: 저널 고장의 분류는 이미 정해져 있다 — 신규 진입은 막지만 엔진은 세우지 않는다는 것을 5.6.1 의 `a112_fault_classification_test.go` 가 값으로 확인해 두었다. 별도 저장소를 세우면 그 분류를 처음부터 다시 정해야 하고 트랜잭션·백업·불변 트리거도 다시 만들어야 한다. writer 는 `internal/app/engine` 에 있다 — `internal/strategyworker` 는 자기 폐포에 쓸 수 있는 저널이 없다는 것을 `-deps` 로 증명한 패키지이고, `design.md:223` 이 "lane health/latch projection" 을 엔진에 배정한 것과 같은 이유다. 그 패키지가 받는 것은 **값**(`LatchRestore`)뿐이고 값은 능력이 아니다.

  **복구 증거가 무엇인지도 고른 것이 아니라 읽은 것이다.** `scheduler.Activation` 은 "an opaque capability issued only after an exact manifest verification. Callers cannot forge it with a bool or public struct literal"(`desired.go:236`)이고 `Generation()` 이 돌려주는 값은 ed25519 서명 매니페스트 파일의 `manifest.Generation`(`production_activation.go:159`)이다. 사람이 서명된 파일을 바꾸지 않으면 그 수는 오르지 않는다. 그래서 복구 조건은 "**엄격히 더 큰** 서명 활성화 세대"이고, 되돌린 매니페스트는 잠금을 열지 못한다. `design.md:199` 의 "복구 조건은 성공한 사이클이 아니라 증거"가 이것으로 값이 됐다.

  **판정은 Go 가 아니라 SQLite 트리거가 한다.** `strategy_lane_latch_recovery_needs_newer_activation` 이 그 문장이고, 같은 세대를 되돌려 주는 복구는 DB 가 거절한다. 코드에 두면 다른 호출자가 다른 판정을 쓸 수 있다. 첫 원인이 마지막 원인에 덮이지 않는 것도 트리거(`strategy_lane_latch_first_cause_wins`)이고, 잠금과 복구는 둘 다 UPDATE·DELETE 가 금지된 **기록**이다.

  **복구는 잠금을 푸는 것이 아니라 레인을 다시 태어나게 하는 것이다.** 푸는 길이 존재하면 그 길은 증거 없이도 불릴 수 있다. `strategyworker` 에는 `latched` 를 거짓으로 만드는 자리가 **하나도 없고**, 그 사실은 패키지 전체를 훑는 셈 시험이 지킨다(얼린 목록은 `latch = true`, `restoreLatch = true` 둘뿐). 엔진은 원장이 복구를 받아들인 **뒤에만** 그 레인을 기록 없이 새로 세운다.

  **RED 은 컴파일 오류가 아니라 행동이다.** `TestALatchedLaneComesBackLatchedAfterTheProcessRestarts` 는 오늘의 API 만 써서 편집 전에 실제로 빨갰다 — "재시작이 잠긴 레인을 열었다".

  **가장 우회하기 쉬운 자리를 따로 얼렸다.** 복구 조건은 "더 큰 수"가 아니라 "더 큰 **서명된** 세대"다. 그 인자에 주기 계수기나 시각을 넣으면 트리거는 그대로 통과하고 잠금은 저절로 열리는데, 두 수 다 그냥 커지므로 **어떤 행동 시험도 차이를 못 본다.** 그래서 인자로 넘어가는 식 자체를 센다(`TestTheRecoveryGenerationComesFromTheVerifiedActivationAndNothingElse`).

  **오늘 생산에 쓰는 행은 0이고 그것도 값으로 확인했다.** 여덟 레인은 전부 DORMANT 라 `Run` 이 아무것도 보기 전에 돌아오고 오류가 없다 → 잠기지 않는다 → 행이 생기지 않는다. `TestTheProductionStepNeverLatchesSoTheLedgerStaysEmpty` 가 다섯 주기를 두 시장에 돌리고 원장이 비어 있는지 본다.

  **이 로트가 없애는 운영 수단 하나.** 오늘 잠긴 레인을 여는 방법은 **엔진 재시작**이다. 이 뒤로는 재시작이 그것을 열지 않는다 — 열려면 generation 이 더 큰 서명된 활성화 매니페스트가 있어야 한다. 요구된 동작이지만 재시작으로 고치던 사람에게는 수단이 하나 사라지는 것이므로, 거절 문구가 무엇이 필요한지 말한다(`journal.ErrStrategyLaneLatchRecoveryEvidence`). 그리고 재시작을 복구 수단으로 쓰지 않게 되는 것은 안전상 이득이기도 하다 — 기억에 남은 대로 엔진 정지는 손절을 놓는 주체를 없앤다.

  **원장 스키마는 High-risk 라 Pre-Edit 선언을 review.md 에 먼저 적었다.** 되돌릴 수 없는 것 하나를 함께 적는다: 이 마이그레이션이 돈 DB 를 구버전 바이너리로 되돌리면 엔진이 `ErrSchemaTooNew` 로 뜨지 않는다(조용히 오해하지 않고 거절한다). a112 는 8.6 기준 배포 BLOCKED 이고 병합 순서는 사람이 정한다.

  **5.7 이 남긴 시험은 지운 것이 아니라 뜻이 바뀌었다.** `TestRestartForgetsALatchedLaneWhichIsExactlyWhatTask533MustFix` 의 단언은 그대로 옳다 — 기록 없이 세운 레인은 열려 있다. 이제 그것은 결함이 아니라 계약이므로 이름을 `TestALaneBuiltWithoutADurableRecordIsBornUnlatched` 로 바꾸고, 기록에서 태어나는 쪽을 `TestAProductionLaneBornFromADurableRecordIsLatched` 가 맡는다. 둘이 함께 있어야 "잠금의 출처는 기록 하나뿐"이 된다.

  **반증이 설계 결함 둘을 찾아냈고 둘 다 고쳤다.** 하나는 **문 없는 fail-closed**: 이 빌드에 없는 레인을 가리키는 기록을 만나면 오류를 내고 그 기록을 버렸는데, 버린 순간 그것을 닫을 방법이 사라진다(복구는 원장 순번을 받는다). 진입이 영원히 멈추고 빠져나갈 길이 없는 상태다. 지금은 붙지 않은 기록도 들고 있다가 복구를 함께 시도한다. 다른 하나는 **판정이 둘**: 복구 조건을 Go 부등호와 SQLite 트리거 양쪽에 뒀더니 부등호를 `>=` 로 바꾼 변이가 살아남고(트리거가 막는다) 트리거를 지운 변이는 엔진 시험에서 살아남았다(부등호가 막는다) — 각자가 상대의 시험을 통과시킨다. Go 쪽을 지워 판정을 하나로 만들었고, 그 뒤 트리거를 지운 변이는 엔진 시험도 빨갛다. 반증 14개 CAUGHT, 대조군 1개 SURVIVED.

  **이 태스크가 주장하지 않는 것.** 여덟은 여전히 진입의 관문이 아니다(5.1.2.2). 복구는 세대가 오른 **다음 주기**에 일어나며, 그 세대를 올리는 매니페스트 자체는 8 절이다. 그리고 `internal/app/engine` 의 나머지 동시성은 여전히 검출기 밖이다 — 이 로트가 공유 런타임에 가변 상태를 더했으므로 새 동시성 시험을 `RACE_ENGINE_TESTS` 에 더했고, 목록의 완전성은 이제 `RACE_ENGINE_FILES` 가 가리키는 파일 전부와 대조한다.
- [x] 5.4.1 Implement the pure calibrated arbiter — validate proposal seals/freshness/owner scope, group exact owner scopes, perform active-owner-first calibrated arbitration over the signed `score_ppm` table, and return typed refusal outcomes; wire it into the per-market proposal path so a refused scope closes that market rather than dropping one symbol, which removes the `AMBIGUOUS_FAMILY_PROPOSAL` fail-closed guard at the root [review C2, a112 결정 57].
- [x] 5.4.2 Host that arbiter in the explicit KR/US `MarketCoordinator` runtime from 5.1.1 — queue intake, coalescing, deterministic owner-scope ordering and bounded handoff — instead of the current in-line call from `strategyProposalAuthorityLoader.collectMarket`.
- [x] 5.4.3 Make a routed symbol that produced no proposal distinguishable from one whose proposal was lost: `batch.LanesFor` returning empty currently means both, so any of the seven silent `continue` paths in `LoadProductionAuthorityBatch` (evidence replay error, lane-input build error, route-set invalid, scope not admitted, invalid proposal seal) shortens the market's entry list, and a shorter list *satisfies* the shared `len(entries) != 1` gate — so a fault on symbol A releases symbol B, which would not have been released otherwise. Carry a typed per-symbol absence reason out of the batch loader and close the market on the fault kinds, leaving legitimate absence as a refusal. Predates 5.4.2 and was found by the 5.4.2 adversarial review with a working demonstration [5.4.2 review P1-1].
- [x] 5.5 Connect each coordinator through one bounded handoff to the existing shared `strategyDispatchCycle`. The handoff is `strategyhandoff.Handoff`, a value in its own package outside `internal/app/engine`, carrying at most `strategyhandoff.Capacity` selection with a typed refusal. It hands the selection out through two doors: `Single() (strategyflow.Result, bool)` for the three read-only consumers, and `Deliver(func(strategyhandoff.Delivered) error) error` for the one order-emitting consumer. The market-wide `len(entries) != 1` assumption that **six** consumer functions each spelled separately is gone from the dispatch path and now lives inside that package (the other two consumers, `strategy_account_first_leg_authority.go` `:151` and `:210`, still spell it in five places inside `internal/app/engine` and are owned by L6 task 6.2 — the census test pins that count), and the golden's `queue.market_wide_single_proposal_assumption_forbidden` stays owned by 5.2 — this task named the bound, it did not lift it. **Not proven, and owned by 5.1.1 (the new type's own closure — landed 2026-09-02) and 5.1.2 (the production path's — still open):** that a worker's `Cycle` cannot reach broker mutation — the production worker's cycle is a closure over `*Context`, which holds Journal/Gateway/Guardian, so that half of the original task text is `not-applicable` until 5.1.2 puts the production path onto the per-family worker types.

  Three adversarial review rounds, all recorded in `review.md`. Round 1 found three P1s (a capacity that admitted more than it delivered, a presence-check guard that did not gate, and a capability ban derived from the wrong receipt). Round 2 confirmed the first closed and broke the other two again with five demonstrated bypasses, all with both suites green. Round 3 confirmed the money-path rewrite behaviour-identical and found **the same hole a third time**: three reviewers, independently and without seeing each other, bypassed the seam with `rawSelection()`, a new file's `relay()`, and `rawTailProposal()` — all three green on both suites. One root cause across all three rounds: **name/spelling checks standing in for value/role checks.** Round 2's own fix text claimed the `entries` ban was "complete rather than enumerative"; that is true of the token and false of the capability, because the ban counted the token inside four *function bodies*, so one helper a call away defeated it — and the ban's positive control was itself a counterexample, since all four consumers already read `entries` one call away through `dispatchHandoff()`.

  Round 3's fix stops writing checks and changes the type. `strategyhandoff.Delivered` wraps the crossing value in an unexported field, so outside the seam package only a zero envelope can be built; `strategyDispatchCycle.dispatch` takes that envelope, and all three demonstrated bypasses now **fail to compile** (mutation M1). What the envelope does *not* prove is written down with it: it proves the dispatched value came through `Admit`, not that the coordinator called `Admit`. That remainder is held by two counts, ~~each complete in its own scope~~ — `TestExactlyOneProductionSiteAdmitsIntoTheSeam` counts `Admit` across every engine production file (not four function bodies), and `strategyhandoff`'s `TestOnlyTheEngineImportsThisSeam` pins which packages may import the seam at all, since calling `Admit` requires importing it. ~~Mutation M2 (engine builds its own envelope) is caught by the first.~~ **Mutation M3 survives:** dispatching the same envelope three times passes every source check, because a spelling census cannot count executions. At-most-once is therefore not a property of these guards — it is held by the journal's position-campaign CAS.

  **Round 4 rejected all of the previous paragraph's completeness claims, and every correction below is measured in `review.md` §5.5-fix4.** Three reviewers, again independent, again all REJECT with P0 = 0. (a) Nothing pinned the envelope's opacity: renaming `Delivered`'s one unexported field restored all three round-3 bypasses with both suites green — the surface table recorded types as the bare word `"type"` and the field walk was hardcoded to `Handoff`. It now records `types.ExprString` of the declared type and walks every exported type. (b) The `Admit` census descended only into `*ast.FuncDecl`, so a package-level `var f = strategyhandoff.Admit` escaped it — while the census's own comment named that spelling as covered. The walk is now over every declaration, and `TestAdmitCensusSeesEverySpellingItClaims` measures the claimed scope against a fixture instead of asserting it in prose. (c) The import guard's universe, `go list <module>/...`, skips `testdata` directories that the compiler does not; it is now the linker's `-deps` closure. (d) `TestTheSameEnvelopeCannotPlaceASecondOrder` only logged the blocker's name, so deleting the CAS gate left it green — the journal's replay-identity check blocked instead. It now asserts the name, and at-most-once is recorded as **over-determined** rather than CAS-only.

  **What round 4 could not close, and what that means for L6 6.2.** `dispatchHandoff` is a method on an ordinary package-private struct, so any engine function can build the receiver from entries of its own choosing and mint an envelope without writing `Admit` or importing the seam. **No spelling census can close this**; the honest fix is for `Admit` to require a sealed value only the coordinator can construct, which reaches outside this task's file ownership. Today none of round 4's four compiling bypasses places an order, because `strategy_account_first_leg_authority.go` re-derives the proposal from the loader's own authority pair and compares identities — the count gate at `:217`, the re-derivation at `:221`–`:222` and the comparison at `:223`–`:225`, which are the five sites the census records as L6 6.2's debt. Those lines carried **no test** until this lot added `TestFirstLegAuthorityRefusesAProposalItDidNotAuthorize`; the census counts `entries` shapes, so deleting only the comparison left it green. Both guards, and the condition under which 6.2 may remove the lines, are written into task 6.2 itself. **6.2 must not delete them until real provenance stands in their place.**

  Two claims from earlier rounds are **withdrawn**, not softened. (a) "The seam that chooses what crosses holds no mutation capability" — same-package capability laundering through a function-typed field cannot be excluded by any test inside `internal/app/engine`, so the guard was renamed to `TestTheSeamFilesStayWithinTheirDeclaredClosure` and its limits written down; what actually moved out of the package is the crossing *door*, not the whole judgement, since the adapter `dispatchHandoff` still reads `entries` here. (b) "The ban is complete rather than enumerative" — disproven three times over; the ban and its test are deleted rather than strengthened. Also corrected in this round: `refusalNow` called an under-carried handoff `HANDOFF_OVER_CARRIED`, which named the empty case as its opposite; under-carry now answers `HANDOFF_NO_SELECTION`, adding no fifth refusal word the frozen golden does not carry.

  Round 3 also found that the gate could not have caught its own evidence errors: `make gate` runs `check_analysis.py` as one of its ten steps, whose coordinate check compares the source range and the branch *count* and nothing else, so a coordinate cited in a branch row that is actually a return, and a call table stopped early, both passed. `tools/logic-map/role_check.py` now compares each cited coordinate against the role `ast.json` gives it and is wired into `check_analysis.py`. Measured against 401 bundles before wiring: zero findings on correct maps, three real ones — enumerated call tables in this change stopped at exactly 40 rows while the AST had 64, 46 and 91. Those three are repaired. `test_role_check.py` plants one error per role dimension plus the silence cases that would otherwise make the check fire on honest hand-written analysis tables.

  **Round 4 corrected that measurement and the checker it validated.** "All 401 bundles in the repository" was the non-archived subset; the repository holds 3014, and over the other 2610 the checker fired on three *correct* documents — reading a WCAG contrast ratio `4.5:1`, a market-open time `09:00` and a socket address `127.0.0.1:0` as source coordinates, because it took the first digit pair anywhere in the row. The sweep varied bundle count but not bundle kind, and prose variety is the dimension the defect lived in. A coordinate is now read only from a cell that is entirely a coordinate, with the accepted cell shapes counted across all 3014 bundles rather than assumed; the three real false positives are planted in `test_role_check.py` as silence cases. Round 4 also found the checker was handed only `function-logic-map.md` while `branch-test-map.md` restates the same branch coordinates — this lot had rebased four bundles' `ast.json` and maps and left 38 anchor coordinates on old line numbers, and the gate stayed silent. Those are repaired and the checker now reads both files. Re-measured after both changes: 3014 bundles, zero findings.

  **Round 5 closed that last exemption rather than recording it.** Round 4 left 39 of the 122 tables wearing the completeness header outside the 1:1 comparison: they wrote bare line numbers, and a row without a coordinate made the checker skip the whole table — so whether a table was audited was decided by the document's author, which is the existence-check disease this change removed everywhere else. All 39 now carry the enumeration derived from `ast.json`, as do the 11 bundles in this change that carried no enumeration at all: 133 of 133 bundles, 2300 call rows. The hand-written analysis is not deleted — it is kept below each enumeration under a heading that says it is not a completeness claim, with its header demoted from `Callee expression` so it cannot wear the marker.

  The checker changed in three ways, each falsified by mutation against the repaired documents: a row under the completeness header that carries no coordinate is now an **error** rather than an exemption; only the **first contiguous table** of the section is read, so a note table below cannot pad a truncated enumeration; and the requirement is now **per change** — a change that enumerates anywhere must enumerate everywhere, measured by `call_enumeration_in_use` rather than declared. Re-measured across all 3014 bundles in 93 changes: enumeration required in 1 change, findings 0. What is left is a change-level choice: a change that never enumerates is still asked for nothing, and that is what keeps the other 16 active changes' 271 hand-written bundles legal. The cliff — the first bundle to adopt the header pulls in all its siblings — is deliberate and written into `role_check.py` so it is not a surprise.
  **Round 5 rejected the round-5 fixes themselves: P0 = 0, P1 = 3, all three reproduced and all three now fixed (`review.md` §5.5-fix6).** (a) The new backstop test was an existence check on an error string. It forges by swapping the market, so identity, market, symbol and account all change together; rewriting the predicate at `:223` to compare only `Lineage.Market` keeps it green — and under that mutation a first leg **issues** for a campaign the authority pair never carried (`err=<nil>`, measured through `go test -overlay` with the repository file's sha256 unchanged). `TestFirstLegAuthorityRefusesASiblingCampaignOnTheSameSymbol` now varies identity alone — same account, market, symbol, quantity and price, a different `campaignID` — which is the shape a112's own target state makes dominant, and it is red under that mutation. This is the rule the change already wrote down (a falsification must vary the axis the defect lives in) quoted and not applied. (b) `role_check.py` keyed tables on `| Callee expression |` and then compared **only coordinates**, never the callee column the marker claims completeness for: `adaptPrices` could be documented as calling `os.Exit` with the gate silent. The spelling is now compared (nodes without text read as `(unnamed)`, the existing convention); all 133 bundles passed unchanged, which is the first evidence that the generated spellings were right. (c) `unittest.main()` sat mid-file, so `python3 test_role_check.py` ran 21 of 29 tests and the 8 this change added — the only guards on all four checker rules — never ran in the natural local form; restoring the silent exemption in a scratchpad copy was reported **OK** before the fix and **FAILED** after. Also corrected: the new test's filename was wrong in three places including 6.2's own prerequisite; B3's branch-test-map row still read "planned … not run" and now carries the measured RED/GREEN; the enumeration mandate keyed on the *first* table's first line, so one pipe line above an enumeration turned auditing off with the marker still standing (it now scans the whole section, and 133 poisoned bundles still measure `True`); and the fix-5 table's 128/5 was really 127/6.

  **Round 6 rejected the round-5 fixes' own fix, P0 = 0, P1 = 1 (`review.md` §5.5-fix7).** The sibling test varied `campaignID`, which is not one axis of `:223` but three of its terms; because the terms identity commits to the lineage identity, the left disjunct is redundant and the right one alone guards quantity and the three prices. Measured with four one-line predicate mutants (`go test -overlay`, repository sha256 unchanged): a `Lineage.Identity`-only or `CampaignID`-only predicate left both existing tests green while a stop-price substitution reached issuance, and an `ExecutionTerms.Identity()`-only predicate is an equivalent mutant. The third test closes it. Six P2s were also fixed, the largest being that the enumeration mandate was still a header-spelling existence check — one extra space, one leading space or bold markup turned auditing off for all 133 bundles while rendering identically. The mandate is now a property of the table's content (a table whose coordinates are the complete ordered call list), the header match is normalised, and every table in the section counts; the only remaining opt-out is to stop enumerating, which removes the completeness claim with it. Repo sweep unchanged at 3014 bundles / 93 changes / 1 enforced / 0 findings.

  **Round 7 rejected round 6's fix in turn, P0 = 0, P1 = 1 (`review.md` §5.5-fix8).** The execution-terms test moved stop and target together, so any predicate that still noticed either one kept it green — and the edit 6.2 will actually make is not deleting a disjunct but expanding the comparison field by field and forgetting one. Measured: expanding `:223` to `Lineage.Identity || Entry || EffectiveStop || Target || Quantity` and omitting any single price left all three tests green while that price alone reached issuance (with `EffectiveStop` omitted, a stop of 80 against an accepted 95, engine package green on both suites). The test is now three single-axis subtests, each asserting via `assertOnlyOnePriceMoved` that lineage, quantity and the other two prices are unchanged; each field omission reddens exactly its own subtest. **Order quantity remains uncovered** — the test seam ties quantity to `PlannedCeiling`, so a quantity-only twin under a fixed lineage cannot be built; B3's map records that cell as uncovered rather than leaving it silent. Four P2s also fixed, the largest being that the mandate was keyed on which `##` section owned the table, so moving all 133 enumerations to another heading turned auditing off with header and rows intact; it now scans the whole bundle prose, both files. Repo sweep unchanged at 3014 / 93 / 1 enforced / 0 findings.

  **Round 8 was the first with no P1 (`review.md` §5.5-fix9): the axis attack that broke rounds 5-7 found no path to an order.** Its four P2s are closed. The largest: `PriceProvenance` has eight fields and `executionTermsIdentity` hashes all eight, but the test seam could only move `priceMinor` while holding the lineage constant — so narrowing `:223` to three `priceMinor` comparisons left all six subtests green. A same-number provenance swap is real in production (`continuationlane/execution_terms.go` mints `saved-effective-stop` for the same stop value, and that package already tests forged provenance). A new `tossos_testseams` file, `internal/strategyflow/authority_stop_provenance_testseam.go`, restates the stop provenance without touching the number, and `TestFirstLegAuthorityRefusesARestatedStopProvenanceAtTheSamePrice` is the only test red under that mutant. **Entry/target provenance and order quantity remain uncovered**, recorded as such in B3's table. Also: the mandate's file list was two enumerated names, so moving all 133 enumerations into `risk-pattern-report.md` — a required bundle file — turned auditing off with the tables intact; `_bundle_text` now reads every `.md` in the bundle directory, which makes the 'outside the bundle' claim true by construction after four rounds of it being one scope too wide.

  **Round 9 found the per-axis strategy cannot terminate, and round 10 changed it (`review.md` §5.5-fix10).** The stop-provenance test moved `source`, `version` and `digest` together, so a predicate noticing only `Source()` stayed green while a `digest`-only forgery issued an order. `executionTermsIdentity` hashes **32 scalars**, most of which the test seam cannot move alone, so buying one axis per round does not end. The obligation is now split, and each half terminates: `internal/strategyflow/execution_terms_identity_fields_test.go` varies all 32 one at a time and reads the field count off the type with `reflect`, so adding a field breaks it; `internal/app/engine/strategy_first_leg_backstop_shape_test.go` is a single AST assertion that `:223`'s condition is exactly the disjunction of the two sealed-identity comparisons. That is structural equality on the guard's own node, not a spelling census over a scope — and it kills every mutant from rounds 6-9 including `ExecutionTerms.Identity()`-only, which no behavioural test can distinguish. **6.2 must keep the four behavioural tests, the 32-field table and the shape assertion green**; if 6.2 needs to change what `:223` compares, the shape assertion is the line to update deliberately, and the 32-field table says what the identity already covers so the comparison need not be expanded field by field. Note for whoever runs these: the shape assertion reads the file from disk, so `go test -overlay` cannot falsify it — mutate the file and restore from a hashed backup.

  **Round 10 found the same disease inside both halves of round 10's own fix (`review.md` §5.5-fix11), P0 = 0, P1 = 2, both closed.** The shape assertion pinned the condition structurally but the body by existence — asking only whether the refusal literal appeared somewhere in the block — so an inserted branch that adopted the crossing value when `PlannedCeiling` differed kept the assertion and all six behavioural subtests green and issued a forged quantity 9 at stop 80. The body is now pinned structurally: one statement, a two-value return, `errors.New` with that literal. And the 32-field census read `NumField()` for `ExecutionTerms` and `PriceProvenance` while hardcoding `ExecutionPolicy` as one scalar, so its eight other declared fields were outside the count; the arithmetic now reads all three types and `TestBreakoutPolicyIdentityChangesWithEveryFieldItCovers` measures that those eight are delegated through `policy.identity`. **The weekly lane's delegation is unverified in this package** — `weeklyPolicy` fills all nine fields with an identity the lane computed — and is recorded as such rather than left silent.

  **Round 11 (`review.md` §5.5-fix12), P0 = 0, P1 = 1, closed.** Lifting round 10's escape one line above the guard left the guard node byte-identical, so the new body check never saw it: adoption happened before the comparison, every test stayed green, and a forged quantity 9 at stop 80 issued. The B3 note written in round 10 claimed B2/B4 and the behavioural tests held that gap; measured, none of them did — the gap had been named with the wrong owner. The shape assertion now also requires the guard to be the top-level statement immediately after `result := proposalAuthority.Proposal()` and `result` to be assigned exactly once in the function, which kills the lifted escape, the wrapper and any later reassignment. Round 11 also traced the weekly delegation and found it sound, so instead of leaving it named as unverified there is now an eight-case table in `internal/weeklyvaluelane` pinning it; dropping a field from `weeklyExecutionPreimageSeal`'s preimage reddens exactly that field's case.

  **Round 12 (`review.md` §5.5-fix13) had no P1; its two P2s are closed, and one of them is aimed squarely at 6.2.** The adjacency predicate pinned only the second link of the re-derivation, so inserting the loop 6.2 will actually write — pick the entry among four families whose identity matches `accepted`, then compare it to `accepted` — left the guard node byte-identical and every test green, making the guard self-referential. Two things mask it today: `:217`'s `len(entries) != 1` and the census that pins it, **both of which 6.2 changes**. The assertion now pins the whole chain — `proposal`, `proposalAuthority` and `result` each assigned exactly once (pointer writes counted), with both re-derivation links immediately preceding the guard in order. **6.2 must therefore update this assertion deliberately when it admits four families**: the chain is what makes the comparison meaningful, so relaxing `:217` without re-establishing how the single entry is chosen is the failure this predicate exists to catch.

  **Left open, named rather than closed.** CI (`.github/workflows/ci.yml`) runs `make test`, `make lint`, `make build` — the logic-map suite lives only in `make sdd-check`/`make gate`, so those 8 guards still never run in CI; changing the workflow is outside this task. And "133 of 133 bundles enumerate" is a claim about table shape, not about coordinate truth: an independent `go/parser` re-derivation confirmed 118 of them against source with zero mismatches, while the other 15 are `revision: base` bundles that `check_analysis`'s source-hash check does not bind at all (13 of those 15 do not match this change's `base-commit.txt`). That gap is structural and pre-existing, but this change's claim inherits it.

- [x] 5.6.1 Measure and pin the fault scope of the runtime that exists today: which faults block new entry, which stay market-local, and which stop the engine. **(Landed 2026-09-03.)**

  **Split note, 2026-09-03.** The original 5.6 read "Preserve central integrity handling: journal/Gateway/fence/multiple-owner faults block all new entry while lane/market faults remain local and all safety loops retain independent contexts." *Preserve* has an object, and the object is the swap: the three properties have to survive 5.1.2/5.2 replacing two market workers with eight lane workers. That half cannot be done before the swap exists. So this task owns the measuring instrument on today's runtime and 5.6.2 owns re-proving the same three clauses on the eight-lane one. Every clause of the original is owned by one of the two.

  **Why the instrument had to come first.** Six of `runMarket`'s blocks and one of `invokeBoundedStrategyCycle`'s were `count=0` in a whole-package coverage profile — recorded in this change's two FLM bundles and previously assigned to 5.7, which did not take them (5.7's rehearsal measures the new types, not the engine). Those seven are not a random remainder: **four of them are the only paths by which a strategy fault reaches `Run`'s return.** A supervised loop returning makes the `Runtime` cancel every other loop — fill detection, reconcile, exit observation — so those four blocks are the ones that decide whether an entry-side fault can take the stop-loss down with it. This repository had never executed any of them.

  All seven are now `count=1`, each by a named test (`internal/app/engine/a112_fault_scope_test.go`, `a112_watchdog_cancellation_test.go`). No production line changed.

  **What the measurement says.** The four escalations are exactly the supervisor's own broken bookkeeping — no observation time (`latchMarket` B4, `928:2`), latch revision exhausted (B5, `932:2`), restart delay outside the bounded contract (`waitMarketRestart` B3, `845:2`). Evaluation failure — ordinary error, panic, deadline — is not on that list, and that is the substance of "lane/market faults remain local". The other two blocks are the two skip paths, and one of them (`813:4`, the refresh-only swallow) is **the configuration production actually runs**: `NewRefreshingPairedStrategyEntrySupervisor` builds both workers `Effective=false, RefreshesAuthority=true`, so every cycle error there is swallowed and retried on the next poll.

  **The first clause is satisfied by refusal, not by latching or stopping.** `a112_fault_classification_test.go` injects a Gateway protection failure, a Gateway entry-gate refusal and an unusable journal into the real dispatch cycle and shows each time that the Gateway place call count is zero *and* the returned error is not classified central. Two mutants that promote a Gateway error to central are caught; a third, promoting the journal lease issuance, **survived** — the closed journal fails earlier, at owner-fence acquisition, so no behavioural test reaches that line. That is the same "one test per axis does not terminate" result as 5.5 and 5.1.1, and the same answer: raise the counting scope. `a112_central_integrity_census_test.go` parses every non-test file in the package and freezes the complete list of places the central sentinel appears (eight, of which three actually mint it, all inside `Run`), plus the fact that `StrategyCentralIntegrityFailure` has **zero production callers**. The surviving mutant is caught by that census.

  **One balance that was nowhere written down.** `latchMarket`'s fault-handoff `default` arm (`964-965`) is `count=0` and is *unreachable*: the stream holds 2, a latched market is refused by `evaluationState` so each market latches at most once, and there are exactly two markets. 2 = 2. Nothing bound those two numbers together, and 5.1.2 turns the second one into eight — at which point the third lane's latch would fail its handoff, escalate to central, and stop the engine along with every safety loop. `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch` now binds them; a mutant that sets the capacity to 1 is caught.

  Ten mutants on `strategy_entry_supervisor.go` (all caught) plus three on `strategy_dispatch_cycle.go` (two caught behaviourally, the third by the census), restored from a hashed backup rather than `git checkout`.

  **What this task does not claim.** It changes no production behaviour — the lot is tests and analysis only. It does not decide the open ordering question below; it pins today's answer so the question cannot be settled silently.
- [x] 5.6.2.1 Wire the central-integrity fail-closed to the entry gate on today's runtime (human decision (6), 2026-09-30). **(Landed 2026-09-30.)**

  **Split note, 2026-09-30 (Manager 판정 Q3).** 원문 5.6.2(아래)는 「여덟 레인 런타임에서 세 절을 다시 증명」이다. 사람 결정 (6) 은 그 앞에 한 가지를 정했다: fail-closed 의 수단은 프로세스 정지가 아니라 `execgw.EntryGate.Block`. 이 태스크는 그 배선 하나를 오늘 런타임 위에 세운다. **원문 5.6.2 의 절은 전부 5.6.2.2 가 가진다**(재증명이 본문이므로) — 이 태스크는 결정 (6) 의 배선만 가진다.

  **무엇이 바뀌었나.** `runMarket` B12(권한 갱신 전용 worker 갈래 — 오늘 생산이 도는 유일한 구성) 안에 새 분기 B13: 중앙 무결성 오류면 `blockEntryOnCentralIntegrity` 가 `ReasonStrategyCentralIntegrity`(`strategy_central_integrity`, 커밋 `3260f4eb`)로 신규 진입을 닫고 `continue` — 루프는 산다. 판정 순서(refreshOnly 가 중앙 판정보다 앞)는 그대로다. 게이트가 없는 조립에서만 기존의 프로세스 전체 fail-closed 로 올리고, 생산 생성자(`NewRefreshingPairedStrategyEntrySupervisor`)는 게이트 없는 Context 를 거절하며 엔진 자신의 게이트를 넘긴다(같은 게이트를 이미 요구하는 `Recovery` 가 앞에 있어 생산 기동 동작 변화 0). effective worker 의 중앙 고장(B14)과 감독자 장부 고장 넷의 엔진 정지는 census 가 얼린 계약이라 바꾸지 않았다(Manager 판정 Q1 = 안 1).

  **RED · 반증.** `analysis/measurements/lot-5.6.2-5.2.2/red-5.6.2.1.log`: 게이트 잠김 · 엔진 불정지 · 두 시장 계속(한 시험), 게이트 없는 조립의 삼킴 금지, 생산 생성자의 게이트 요구 · 역할(엔진 자신의 게이트) — 넷 FAIL, 대조(보통 오류는 게이트를 잠그지 않음)는 GREEN. 변이 E01~E11 11/11 CAUGHT(`analysis/harness/a112_lot_mutate.py --set 5.6.2.1`, HEAD archive + 이 로트 파일 사본, 무변이 대조군 GREEN). 편집 전 번들 `analysis/measurements/lot-5.6.2-5.2.2/pre-edit/`, 편집 뒤 재측정 `coverage-post-5.6.2.1-engine.json`.

  **이 태스크가 주장하지 않는 것.** 여덟 레인 재증명(fault 스트림 용량 = 레인 수 유도 · 레인 고장 여덟 동시에도 fill/reconcile/exit 생존) — 5.6.2.2. effective worker 활성화 시 중앙 고장의 처분(B14) — 이월(review).
- [x] 5.6.2.2 Re-prove the same three clauses on the eight-lane runtime once 5.1.2/5.2 have swapped it in. **(Owns every clause of the original 5.6.2 — split 2026-09-30; runs after 5.2.2.1. Also owns 5.2.2's 「each lane independently supervised」 — split 2026-09-30.)**

  **Why it is open.** Every property 5.6.1 measured is a property of *two market workers driven by one consumer goroutine each*. The swap changes the number, the drivers and the fault sources. Concretely, three things measured here are known to need re-deriving: the fault-stream capacity equals the worker count (2 today, 8 after); the refresh-only swallow at `813:4` is the production configuration today and will not be after; and "each market latches at most once" is what makes the handoff `default` arm unreachable.

  **Done.** The seven blocks stay executed against whatever then runs the production cycle, the census still returns zero production callers of `StrategyCentralIntegrityFailure`, and a lane fault — including eight simultaneous ones — still leaves fill detection, reconcile and exit observation running.

  **선납(2026-10-01, 5.2.2.1 리뷰 수리 3차).** 5.6.2.2 의 행동 커버 요구 중 **주기 전달 부분**은 `a112_market_cycle_delivery_test.go`(`TestTheProductionCycleHandsEveryOwnerScopeToTheDispatch` · 활성화 없음 대조)가 선납했다 — `runProductionStrategyMarketCycle` 을 통째로 돌려(권한 새로 고침은 1초 캐시 주입, 레인 런타임은 생산 생성자) dispatch 가 handoff 를 전부 받는지 센다. 레인 고장 · 여덟 동시 고장 · 안전 루프 생존은 여전히 이 태스크의 몫이다.

  **종결(2026-10-01).** 재유도: 감독자 worker 는 여전히 시장 둘(레인은 시장 주기 안의 런타임, 레인 고장은 레인 잠금에서 끝남) — fault 스트림 2 = 2 · handoff `default`
  도달 불가 · refresh-only 가 유일한 생산 구성. 일곱 블록 전부 count ≥ 1(현재 좌표), census 0, 여덟 동시 레인 고장 · 레인 잠금 기록 실패 모두 안전 loop 생존
  (`a112_eight_lane_fault_test.go`), 변이 6/6 CAUGHT. review 「5.6.2.2」 절.
- [x] 5.7 Add race, goroutine-leak, queue pressure, fake-clock and fault-injection tests for 8 concurrent workers and 2 coordinators, including simultaneous same-symbol proposals and shutdown/restart. **(Landed 2026-09-02.)**

  `internal/strategyworker/rehearsal_test.go` stands all eight lanes and both coordinators up together and drives them from eight goroutines behind one gate. This is the rehearsal `design.md:255` asks for before the swap, not the swap: the lanes here are test-turned-ON copies, and production callers remain zero.

  **The race detector had never run in this repository.** Neither the `Makefile` nor `.github/workflows/ci.yml` contained the string `-race` before this lot. A task that asks for "race tests" is not satisfied by tests that no gate runs under the detector — that is the a118 lesson applied a third time. `make test-race` now runs it, `make gate` runs that as step 9 of 11, CI runs it, and `tools/sdd/test_race_detector_actually_runs.py` fails if any of those three wirings is removed (five mutants, five caught).

  **Measured, not assumed:** `go test -race ./...` did not finish inside ten minutes, while the seven packages this runtime lives in take 11.9 seconds. So the target names those seven. **Thirty-four other packages whose production code uses goroutines, channels, `sync.` or `atomic.` remain outside the detector** — including `internal/journal` and `internal/app/engine`. That gap is named in the target's comment rather than left silent.

  **What the detector actually bought, measured.** A mutant that moves a lane's first-failure reason into one package-level slot is **green without `-race`** and red with it. Its deterministic sibling — readers seeing the shared slot — is caught by the test alone. So the isolation test has teeth for deterministic cross-talk and needs the detector for the racy kind; both are recorded in `review.md`.

  **5.7 found a hole in 5.3.2 and it is fixed here.** `Step` returned only a `Cycle`, so an *ordinary* error had no way into the lane: the sole failure paths were panic and deadline, both abnormal, both latching without waiting for the threshold. The design's fault table row "deadline/ordinary error → count and retry" was therefore unreachable, and `FailureThreshold` was a dead value for any production driver. `Step` now returns `(Cycle, error)`, which is also what the engine's receipt says (`StrategyCycle = func(context.Context) error`).

  **`design.md:255`'s "spy the shared dispatch handoff" cannot be done from this package, and that is not a gap.** `strategyhandoff`'s own importer census pins the set of packages that may import the seam to the engine alone, because importing it is the only way to call `Admit` — the first draft of this file imported it and the census caught that. So the rehearsal cannot reach the seam at all, which is a stronger statement than a spy counting zero. The real spy belongs in the engine and is 5.1.2/5.2's.

  **Restart forgetting a latch is now a committed assertion, not prose.** `TestRestartForgetsALatchedLaneWhichIsExactlyWhatTask533MustFix` measures both halves — the fresh lane is healthy *and* the original is still latched — so the 5.3.3 gap cannot quietly be believed fixed. **(5.3.3 이 그 구멍을 닫은 뒤 이름이 `TestALaneBuiltWithoutADurableRecordIsBornUnlatched` 로 바뀌었다. 단언은 그대로다: 기록 없이 세운 레인은 열려 있다 — 이제 결함이 아니라 계약이다.)**

  **What this task does not claim.** It proves these types do not tread on each other when run concurrently. It proves nothing about today's engine: the seven engine blocks no test executes are still listed, with coordinates, in the two FLM bundles' branch-test maps. There is still no production driver, so "8 concurrent workers" here means eight test-driven goroutines, not eight production goroutines.

## 6. Shared Risk Owner and Dispatch Integration

- [x] 6.1 Bind family to the server-owned a066 strategy risk bucket identity without creating a second Guardian, account-wide cap, journal, owner key, dispatch owner or Gateway.

  **착지 `40ec5aff`(2026-10-01, 별도 리뷰 불요 — Manager).** Manager 판정 (C): 「risk_id 가 family 를 함의하며 적재기가 강제」 — `validProductionRiskPolicyContents` 가 레인 family 를
  `strategyrouter.ProductionLaneFamily`(정본 표 유도, 완전성 8 = 4 × 2 시험)로 해소해 한 risk_id 를 두 family 가 공유하거나 family 미해소 레인이면 정책 거절(결함 등급).
  (A) 매니페스트 family 필드는 v2 때 묶음, (B) 버킷 값 family 접두는 용량 복제라 기각 — design 정정 사슬. 두 번째 Guardian · account cap · 원장 · owner key · dispatch owner ·
  Gateway 는 만들지 않았다(새 상태 0 — 검증만). **weekly family 활성화의 경성 선행: riskbucket horizon 매핑 결정**(WEEKLY 는 SHORT/MEDIUM 에 매핑되지 않아 적재기가
  「unsupported horizon」 으로 거절 — 의도된 현재 상태로 핀 `TestAWeeklyLaneIsRefusedForItsHorizonUntilAMappingIsDecided`; ROADMAP).
- [x] 6.2 Prove q_final remains the minimum of q_candidate and every Guardian/horizon/market/family/sector/symbol cap and that concurrent family admission creates one owner/decision/reservation transaction only.

  **본문 종결(2026-10-01 — 미착지 · 리뷰 전).** 생산 경로(활성 두 범위 · family 결속 6.1) 위에서 잼(`a112_qfinal_family_test.go`): ① min 결속 — family 버킷이 가장 작은 cap 이면
  발급 수량 = 그 cap(`MaximumQuantity` 독립 계산 · 원장 q_final · 스파이 수량) < q_candidate ② 한 family 버킷 고갈 — 그 family 만 거절, continuation 은 **두 순서 모두** 발급
  (스펙과 갈려 정지 보고 → Manager 판정 (A): admit 이 precheck 의 `QFinalRefusal{BUCKET_CAP_EXHAUSTED}` 를 타입 · 코드로 범위 거절에 실음 — 문구 아님) ③ 공유 차원(horizon) 고갈 —
  두 범위 각자 범위 거절 · 발급 0 · 주기 생존 ④ 같은 범위 · 다른 family 가 두 조립에서 한 원장으로 경합 — 원장 owner 충돌로 주문 1 · 결정 1 · 예약 한 세트. 경계: 버킷 고갈이
  아닌 precheck 거절(SYMBOL_NOT_ALLOWED · EXISTING_GUARDIAN_CAP)과 발급 단계 STALE 은 결함(주기 멈춤) 그대로, R3(진입 관문 관측)는 넓히지 않음. 변이 U01~U06 6/6 CAUGHT.

  **Blocking prerequisite inherited from 5.5 — read before touching `strategy_account_first_leg_authority.go`.** 5.5's handoff seam proves that a dispatched value passed through `strategyhandoff.Admit`; it does **not** prove the coordinator called `Admit`, because `dispatchHandoff` is a method on a package-private struct that any engine function can build from entries of its own choosing. Round 4 compiled four such bypasses. None of them reaches the broker, and the reason is in this file, not in the seam: `collectStrategyFirstLegAuthority` refuses the crossing value, re-derives the proposal from its own authority pair and compares identities at `:217` (the count gate), `:221`–`:222` (re-derivation) and `:223`–`:225` (the identity comparison, whose refusal is `production proposal identity changed`). Those lines are the five sites `singleProposalAssumptionCensus` records as 6.2's debt.

  Two guards now hold that debt, and they hold different halves. `TestTheSingleProposalAssumptionLivesOnlyWhereTheCensusSaysItDoes` fails if the five `entries`/`len(entries)` sites move or disappear — but it counts shapes, so deleting only the identity comparison at `:223`–`:225` leaves it green. `TestFirstLegAuthorityRefusesAProposalItDidNotAuthorize` (`strategy_first_leg_identity_backstop_test.go`, `tossos_testseams`) closes that: it hands the loader a `strategyFirstLegAccepted` built from the KR proposal while the loader's own pair carries the US one, and requires the exact refusal above. Deleting `:223`–`:225` turns it red — measured, the refusal becomes `production risk authority scope changed`. Before 5.5's fix lot there was no test on those lines at all; deleting them was silent.

  **That test alone is not enough, and round 5 measured why.** It forges by putting the US entries in the KR slot, so market, symbol and account change along with identity — a predicate rewritten to compare only `Lineage.Market` keeps it green. Under exactly that mutation (`go test -overlay`, repository file unchanged) a first leg **issues** for a campaign the authority pair never carried: the loader returns `err=<nil>`. `TestFirstLegAuthorityRefusesASiblingCampaignOnTheSameSymbol` in the same file holds the axis that matters: same account, same market, same symbol, same quantity and price, a different `campaignID` — the shape a112's own target state (four families on one symbol) makes dominant. It is red under that mutation. **Round 6 then measured that neither test covers the half that matters most.** `:223` is a disjunction, and `executionTermsIdentity` hashes `lineageIdentity` (`internal/strategyflow/types.go:315`), so the left disjunct is subsumed by the right and the right one alone guards the order quantity and the entry/stop/target prices. Both tests above move `campaignID`, which moves both disjuncts together — deleting the right half keeps them green while removing the only check on size and stop price. `TestFirstLegAuthorityRefusesRewrittenExecutionTermsUnderTheSameLineage` holds `campaignID` and the whole lineage constant and rewrites stop/target, so it is red under exactly that deletion and under a `CampaignID`-only predicate. 6.2 must keep **all three** green: the first proves the block exists, the second and third split the two disjuncts — lineage and prices/size respectively. A `ExecutionTerms.Identity()`-only predicate is an equivalent mutant (all three stay green), which is the measured proof that the left disjunct is redundant.

  ~~**6.2 may remove these lines only once `Admit` requires a value only the coordinator can construct**~~ **정정(2026-10-01, Manager 판정 · 사용자 거부권 보고됨):** 봉인 기제 = `Admit` 입력 타입이 아니라 마지막 권한의 **범위-단위 재유도(의미 봉인)**. 근거: 조정자는 신뢰 주조자가 아니다(엔진이 생성 · 공급 · 구동 → 조정자 토큰은 「어떤 중재가 돌았다」만 증명, 새 조정자에 패자 가족의 진짜 봉투만 넣으면 유효한 토큰). 수용 · 선결 구조는 결정 (1) 그대로, 기제만 정정. 그래서 다섯 줄은 **삭제되지 않고 일반화로 대체**된다 — `proposal.entries[0]` 선택이 `authorityForOwnerScope`(accepted 계보의 OwnerKey 로 조립 권한 쌍에서 유일 항목 선택, 0 또는 복수면 거절)로 바뀌고 identity 가드는 그대로 남는다. 위 세 시험은 초록 그대로이고 위조 축 다섯이 행동 시험으로 더해졌다.
- [x] 6.2.0 Seal the dispatch forgery path by owner-scope re-derivation at the first-leg authority, and census strategyflow's sealed-Result minting (decision (1) prerequisite · decision (5) share). **(Landed 2026-10-01 — 이 태스크가 6.2 의 차단 선결조건을 닫는다; 6.2 본문(q_final 최소값 · 동시 가족 입장 단일 트랜잭션)은 열린 채 남는다.)**

  **무엇이 바뀌었나.** `collectStrategyFirstLegAuthority`: 준비 조건과 시장 단위 개수 관문을 나누고, 그 사이에 소유자 범위 선택(`authorityForOwnerScope`, 새 파일 `strategy_first_leg_owner_scope.go`) — 범위로 고르고 identity 로 대조(자기 참조 함정 회피). 선택 실패 거절 문구는 identity 거절 문구를 머리로 담는다. 개수 관문(`len(entries) != 1`)은 5.2.2.2 몫으로 남는다 — 이 태스크는 판정을 더 엄격하게만 한다(생산 동작 변화 0). A-lite: 활성화 시장에서 dispatch 목록의 제안 집합 digest 가 조립의 것과 다르면 시장 단위 MarketClosed(2차 방어, 봉인 아님). strategyflow: 공개 표면 golden · 봉인 쓰기 census · 봉인 함수 셋의 AST 정본 digest 동결 + review 기록 결속. 이름 붙은 봉인 시험의 실행 증거(`go test -json` pass 사건) 하네스 `analysis/harness/verify_named_tests.py`(5.2.2.1 리뷰 이월 #1 — 이 로트 한정).

  **증거.** RED `analysis/measurements/lot-6.2-seal/red-6.2-seal.log`(셋 FAIL), 변이 `analysis/measurements/lot-6.2-seal/mutation-6.2-seal.tsv`(S01~S09 · F01~F05 CAUGHT, N04 GREEN-AS-EXPECTED), review 「2026-10-01 6.2 봉인 로트」 절.

  **리뷰 수리(2026-10-01, codex BLOCK · A BLOCK · B BLOCK — review 「6.2 봉인 로트 적대 리뷰 처분 · 수리」 절).** 공유 배열 구멍(loader 가 조립 slice 를 그대로 듦 → dispatch 쪽 제자리 교체로 같은 계보 · 다른 손절 쌍둥이가 **발급됨**, 실측 RED)을 구성 때 떼어 내기로 닫음. strategyflow 언급 census · 빌드 제약 모델 · 재생성 멈춤, 하위 시험 인지 하네스 · 변이 대조군 pass 사건, 선택 함수 직접 시험 · 축별 시험, 거짓 증거 문장 정정. 6.2.0 의 효력 범위는 위 5.2.2.2 절.
- [x] 6.3 Preserve the current dispatch validation order and final authority rechecks; add only the proposal family/arbitration lineage required by the lease preimage and reject any digest/version drift before transport.

  **6.3 종결(2026-10-01 — 미착지).** Manager 판정 (A)+(C): lease 스키마 무변경 — family 는 lease LaneID 가 정본 표로 함의, 중재 계보는 **발급 시점**(1차 레그
  권한 · admission 커밋 전)에 대조됨을 시험으로(`a112_dispatch_lineage_test.go`: 순서 AST 동결 · lease 레인 → 가족 단사 · 같은 범위 다른 계보의 발급 시점 거절
  (예약 · lease 행 불변 · 게이트웨이 0) · 최종 검사의 재검증 drift 거절). transport 전 계보 재대조는 없음 — ROADMAP 「a112 이월」 (B) 행(활성화 로트 선행).
  잔여 (c): 재검증 drift 판정을 순수 함수로 의미 무변경 이동(영수증) + 축별 시험. 변이 13(동등 표기 A2 · A3 — 철자 핀만).
- [x] 6.4 Enforce breakout first-leg-only production authority and add broker spies proving duplicate evaluation/restart/correction cannot create a second first-leg or any scale-in mutation.

  **6.4 입력(2026-10-01 B1 · breakout 덮개 2차).** (a) 골든 변 `PROPOSED → CONSUMED` 의 생산자가 이 태스크다 — v1 평가기에는 없고(B1 실측),
  `a112_transition_producer_census_test.go` 가 `phaseConsumed` 생산 자리 0 을 고정하므로 이 태스크가 생산자를 더할 때 census 와 review B1 표를 같이 고친다;
  소비 계약(CONSUMED prior 는 정정 뒤 보존)은 `TestACorrectionAfterConsumedNeitherResurrectsNorReissues` 가 미리 잰다. (b) 레인이 막지 **않는** 둘째 첫 레그 경로
  (실측): config 재버전(새 setup ID — 스펙대로 소급 재해석 없음)은 같은 세션 · 같은 종목에 새 제안을 낸다(prior 를 넘겨도 다른 setup 이라 fresh — 프로브에서 후보 99 → 19, 새 ProposalID);
  세션 교체도 새 setup 이다. 레인 판정은 setup 단위라 「(종목, 세션) 또는 포지션 단위 첫 레그 하나」 는 이 태스크의 권한이 세워야 한다.

  **6.4 종결(2026-10-01 Manager 판정 A + CONSUMED (i) — 미착지).** 생산 변경 0. 첫 레그 전용 권한은 공유 journal admission 이 진다 — 캠페인이 열린 동안 같은 종목의
  둘째 첫 레그는 세 층에서 각각 거절: ① 전달 몸통(`strategy_market_handoff_delivery.go` CAS 건너뛰기) ② 1차 레그 권한 loader(`production position campaign CAS changed`)
  ③ journal(`insertFirstLegCampaignTx` · `idx_position_campaign_active_scope` · 위험 버킷 owner). 시험(seam 주입 — 생산 breakout 입력이 벽에 막혀 있음):
  engine `a112_breakout_first_leg_only_test.go` `TestARedeliveredBreakoutFirstLegReachesTheBrokerOnce`(중복 평가 멱등 — 재전달 세 파도 뒤 브로커 1 · 캠페인 1 · 결속 1 · 레그 1) ·
  `TestASecondBreakoutFirstLegWhileTheClaimIsActiveNeverReachesTheBroker`(층 ① — 몸통 도달 · dispatch 0) · `TestTheFirstLegAuthorityRefusesASecondBreakoutLegEvenPastTheDeliverySkip`(층 ②) ·
  `TestNoScaleInOrExtraBreakoutLegWhileTheCampaignIsOpen`(개입 금지 — 네 파도 · 다른 캠페인 제안 섞어도 추가 노출 레그 0); 층 ③ 단독은 journal
  `TestFirstLegAtomicAdmissionCompetingSameScopeHasOneWinner` · `TestFirstLegAtomicAdmissionExactReplayUsesOriginalJournalToken` + 프로브 `lot-6.4/probe-journal-layer.log`(층 ①② 끔 →
  `risk bucket owner conflict`, 브로커 1 유지). 벽 핀(조건 ③): strategyproposal `a112_breakout_wall_test.go` `TestEveryProductionBreakoutLaneInputIsRefusedAtTheWall`(서술자 두 출처로
  breakout 레인 전부) · `TestNoNonTestCodeBuildsABreakoutLaneInputAroundTheWall`(우회 생산자 0). 변이 `lot-6.4/mutation-6.4.tsv` W1~W4 · L1 · L1b · L2 CAUGHT.
  CONSUMED (i): design.md 「v1 CONSUMED 기록」 — 레인 phase CONSUMED 생산자 0(census 불변), 골든 불변.

  **잔여(재생 수순 그대로).** 손절-종결 → 재시작 → 새 매니페스트 CampaignID → 같은 setup 둘째 첫 레그: 첫 레그 캠페인이 손절로 CLOSED 되고 claim 이 풀린다 → 재시작으로 레인 prior 를 잃는다(생산에서 `BreakoutRequest.Prior` 를 채우는 곳 0) → 같은 setup 이 봉 하나만 더해져도 새 스냅숏 digest · 새 ProposalID 로 다시 PROPOSED → 서명 제안 입력 매니페스트가 새 CampaignID 를 실으면 journal 은 FLAT/CLOSED · claim 없음 · 새 캠페인 PK 로 받아들인다 → 같은 setup 의 둘째 첫 레그. 스펙(breakout-retest-strategy-lane 「breakout v1 production 권위는 first-leg 하나로 제한된다」): "Proposal replay, duplicate bar delivery, correction 또는 restart가 동일 setup/bar에서 두 번째 first-leg 권위를 만들면 안 되며 (MUST NOT)".
  오늘 이 수순은 벽(`ErrBreakoutEvidenceUnavailable`) 뒤에 있어 도달 불가다. **면제 불가: 「ErrBreakoutEvidenceUnavailable 해제 로트는 B(SetupID 계보 결속) 또는 C(consumed-setup 원장 기록) 착지 전 해제 불가」.**
- [x] 6.5 Add crash/retry tests across coordinator handoff, owner/q_final admission, lease claim, SUBMITTING and exact outcome reconciliation without releasing or duplicating capacity incorrectly.
- [x] 6.6 Add prerequisite regression tests proving a066 incomplete owner/exit gate or a100 missing/mismatched/expired protection attestation yields exposure-raising broker request zero while reduce-only paths continue.

  **6.5 종결(2026-10-01 Manager 판정 A65 — 미착지).** 생산 변경 0. engine `a112_dispatch_crash_restart_test.go`
  `TestADispatchCrashAtEveryStepNeitherDuplicatesTheOrderNorReleasesTheCapacity` — 네 crash 지점마다 원장을 닫고 다시 열어 새 Guardian · 새 dispatch owner · 새 파도로
  같은 제안을 재전달: ① handoff 뒤 · admission 앞 → 재시작 뒤 브로커 1 · 한 세트(결속 1 · 캠페인 1 · owner 1 · HELD 5 · lease CLAIMED 1); ② admission 커밋 뒤 · lease
  앞(lease INSERT RAISE 트리거) → 브로커 0 · HELD 5 · lease 0; ③ claim 뒤 · SUBMITTING 앞 → 브로커 0 · lease CLAIMED · HELD 5; ④ SUBMITTING 중(전송 시작 뒤) → 추가
  브로커 0 · lease SUBMITTING · HELD 5, 그리고 lease 수명 동안 새 owner 거절(`ErrStrategyDispatchOwnerBusy` — 그 시장 전략 진입 전부 정지), 인수 유예 뒤 owner 는 서지만
  복구 분류는 ATTESTED_OUTCOME_REQUIRED 하나. 모든 지점에서 브로커 총합 ≤ 1 · 풀린 용량 0 · 중복 0. **승인한 단언 집합과의 차이(실측):** 제안은 「①–③ 브로커 1」 이었으나
  ② · ③ 은 재시작 뒤 사슬이 이어지지 않아 0 이다 — 아래 잔여. 변이 `lot-6.5-6.6/mutation-6.5-6.6.tsv` C1 CAUGHT.

  **6.6 종결(2026-10-01 Manager 판정 A66 — 미착지).** (h) execgw `a112_protection_prerequisite_test.go` `TestEachProtectionAttestationFailureStopsBuysAndKeepsReductionsFlowing`
  — 없음(DefaultSnapshot · `missing_evidence`) · 불일치(다른 계좌 · 다른 tool digest · `attestation_scope_mismatch`) · 만료(지금 정확히 · 1분 전 · `attestation_expired`)
  각각 같은 시험 안에서 매수 브로커 0 + 매도 · 취소 · 축소 정정 브로커 도달, 양성 대조군(신선한 WIRED → 매수 1). 엔진 수준 만료 1: engine
  `a112_protection_expired_strategy_test.go` `TestAnExpiredProtectionAttestationStopsTheStrategyFirstLegBeforeTheBroker`(dispatch 주기의 보호 관측을 실제 게이트웨이로 —
  만료 정확히 · 1분 전 → 송신 0, 대조군 신선 → 1). `failProtection` 이 세 모양을 못 가르는 현황은 게이트웨이 층에서 가름으로 해소(생산 변경 불요). (g) = (i)+(ii)
  (Manager 판정 — 의존성 게이트 발명 금지): (i) a066 미완 세계(q_final 표식 · a066 admission 없음 → 버킷 불일치 거절, 손실 잠금) 상승 0 과 reduce-only 지속의 같은-시험 짝 —
  execgw `TestA066LossLockAndBucketFailureNeverBlockRiskReducingPaths`(+ `TestA066StrategyLastMomentQFinalBarrierRefusesALockTakenAfterTheInitialCheck`);
  (ii) 서명 4-가족 활성화 없음 → 여덟 레인 worker DORMANT(정상 입력에도 봉투 0, 대조군 켜진 worker 는 봉투) — strategyworker `TestEveryProductionWorkerIsBornDormantAndEmitsNothing`,
  서술자 기본 OFF — strategyrouter `TestDescriptorsShipKRAndUSTogetherDefaultOFF`. (주의: 활성화 없는 시장의 기존 단일 경로 조정은 그대로 돈다 —
  `TestWithoutAVerifiedActivationCoordinationIsUnchanged`; 그 경로의 상승 0 은 (i) 의 a066 관문 · 버킷 불일치와 서명 제안 입력 부재가 진다.) 변이 P1~P4 CAUGHT.

  **잔여(6.5 — 면제 불가 활성화 선행).** 근거(④ 의 위험): SUBMITTING 중 crash 면 브로커에 실주문이 존재할 수 있는데 이 빌드에는 그 결과를 알아낼 경로가 0 이다 — `internal/journal/strategy_dispatch_runtime.go:178` 「No constructor for that authority exists in this build」(ATTESTED_OUTCOME_REQUIRED), `DiscoverStrategyDispatchRecovery` · `RecoverClaimedStrategyDispatchLease` 생산 호출 0 · `internal/strategyruntime`(모델 복구) 생산 import 0 (grep 영수증 `lot-6.5-6.6/recovery-callers-grep.log`). UNKNOWN_BROKER_STATE 가족의 실체이며 활성화 상태에서는 liveness 가 아니라 안전 문제다. 같은 경로 부재로 ② admission 커밋 뒤 · lease 앞, ③ claim 뒤 · SUBMITTING 앞 crash 도 재시작 뒤 이어지지 않는다(그 종목 claim · HELD 용량이 풀리지 않음 — ② 는 lease 가 없어 복구 열거에도 안 보인다).
  **면제 불가: 「활성화 로트는 outcome reconciliation 착지 전 해제 불가」.**

## 7. Scheduler Observability and Operator Surfaces

- [x] 7.1 Extend scheduler capability scope with family while retaining one physical endpoint/reset-generation quota, commitment set, absolute issuance cap, observation-cycle authority and safety reserve.

  **설계 입력(2026-10-01, a070 처분 ② 종결 — Manager 지시).** 7.1 의 API 예산 subscope 는 `strategyrouter.QuotaAuthority` 가 아니라 정본 구현 **`internal/scheduler.BudgetCoordinator`**
  위에 얹는다. 근거: `QuotaAuthority` 는 정본 요구보다 약하다 — capability 토큰이 요청 필드 + 스냅숏 digest 의 결정적 sha256(정본은 암호학적 난수), reset generation 진행 ·
  observation cycle · `SafetyReserve` 규칙(50% 올림 · 최소 5)이 없고 reserve 를 호출자가 넣는다(a070 처분 감사 `0d3e2849` §quota). 거기에 family 를 얹으면 정본보다 약한 두 번째
  예산 권한이 생산에 선다. L4 원장의 `quota.go` 편집 대상 지정은 이 입력으로 교체했다.
- [x] 7.2 Add replay/concurrency tests proving a capability cannot cross family and 8 simultaneous acquires cannot issue beyond the shared physical allowance.

  **7.1 · 7.2 종결(2026-10-01 — 미착지).** Manager 판정 Q1~Q4: `internal/scheduler.BudgetCoordinator` 위에 additive `TryAcquireStrategy` · `CompleteStrategy` +
  `StrategyScope{Market, Horizon, Family}`(정규화 검증, family → horizon 표 정합은 비강제 — 필요해지면 6.1 식 정본 표 유도). 토큰 · commitment 에 범위 digest, 완료 때 대조(양방향 교차
  replay 금지), 용량은 endpoint 하나의 commitment 집합 공유(복제 0). 거절은 골든 `BUDGET_DEFERRED` 하나 + Detail. 생산 호출자 0 을 핀(`TestTheStrategyBudgetAPIHasNoProductionCallerYet`)
  — 레인 evidence polling 배선은 7.5 또는 활성화 로트가 핀을 뒤집으며 한다. 시험(`internal/scheduler/a112_strategy_scope_test.go`): continuation → breakout replay 거부 · 양방향 교차 ·
  공유 집합 비복제 · 여덟 레인 마지막 자리 경쟁(하나만 commit, 나머지 BUDGET_DEFERRED, 예비 유지, 안전 등급 통과) · 잘못된 범위/안전 등급. 변이 T01~T10 10/10 CAUGHT.
- [x] 7.3 Project read-only family/worker desired/effective/runtime, cycle generation, queue/drop, health/latch, first refusal, evidence/config/calibration and arbitration lineage to the existing console/API model.

  **7.3 종결(2026-10-01 — 미착지).** Manager 판정: 계약 모양(envelope additive `lanes[8]` · `coordinators[2]`, SchemaVersion v1 그대로, R4 를 `coordinators[].selected[]` 로
  흡수) · Q1=(B) cycle generation = 엔진 시장별 물결 번호 중 레인이 마지막으로 관측된 번호(0 = 미관측, 프로세스 수명) · Q2 기한은 기존 읽기 접근자만 · Q3 runtime 은
  worker 값(UNOBSERVED) 그대로 · (A) first refusal(`lanes[].refusal` — REFUSED 일 때만 골든 중재 코드) · config/calibration 계보(`selected[].configDigest` ·
  `scoreVersion` · `calibrationDigest`). `internal/strategyprojection` 은 import 0 잎 그대로(레인 표 · 어휘를 골든 · 생산 상수와 시험으로 대조), 엔진 `Read` 가
  레인을 지금 상태로 읽기만 해서 덧씌움, 조정자 자식은 조립 발행 때(실패 갈래에서도). OpenAPI 세 스키마. 콘솔 template 무편집(화면 노출은 후속 콘솔 로트).
  계좌 원문 0(lineageIdentity 는 SHA-256). 변이 `lot-7.3/mutation-7.3.tsv`(동등 표기 둘 외 CAUGHT), 읽기 전용 불변 변이(Offer · Fail) CAUGHT. review 「7.3」 절.
- [ ] 7.3.1 Distinguish `OFF/OFF/UNOBSERVED` from explicit read-only `SHADOW`: SHADOW may evaluate and project counterfactuals but cannot mint desired/effective/activation, own dispatch capability or survive restart without a server-owned signed shadow manifest.

  **7.3.1 상태(2026-10-04 Manager 판정 — 열림, 사용자 결정 대기).** spec four-family-strategy-runtime :88-93 이 SHADOW 를 SHALL/MUST NOT 으로 구속한다(허용 절은
  「서명 shadow 매니페스트가 있을 때만」, 금지 둘은 ON 승격 · dispatch 소유 · 재시작 자동 복구). 오늘은 SHADOW 상태가 없어(`strategyrouter.RuntimeState` = {UNOBSERVED},
  골든 runtime 전부 UNOBSERVED) 빈 표본으로 성립한다 — **R2(이 로트)가 그것을 핀 통과로 바꿨다:** engine `a112_shadow_absent_restart_test.go`
  `TestARestartAfterAnObservedPromotionComesBackOffOffUnobservedWithNothingWritten`(앞 프로세스가 검증 활성화 아래 KR 네 레인 ON 관측 → 같은 원장으로 재시작 → 여덟 다
  OFF/OFF/UNOBSERVED, 원장 dispatch lease · 레인 잠금 · 복구 기록 0) · `TestTheRuntimeVocabularyIsExactlyUnobservedUntilAShadowLotExtendsIt`(router `RuntimeState` · projection
  `LaneRuntime` 상수 = 정확히 {UNOBSERVED}); 기존 결속 `TestEveryProductionWorkerIsBornDormantAndEmitsNothing` · `TestDescriptorsShipKRAndUSTogetherDefaultOFF` ·
  `TestDormantAndUnavailableSnapshotsCarryEightUnobservedLanesAndTwoCoordinators` · 활성화 파일 생산 작성자 0 `TestOnlyTheAuthoringToolCanBuildActivationBytes`. 변이
  `lot-7.3.1-R2/mutation-7.3.1-R2.tsv` 4/4 CAUGHT. **남은 결정(R3 — 게이트 차단, Manager 가 사용자에게 보고):** SHADOW 를 a112 에서 만들 것인가 / 활성화 로트로 미룰 것인가 /
  spec 델타의 SHADOW 허용 절 · 시나리오(:89 둘째 절 · :91-93)와 design :289 · :291 배포 계획을 후속 change 로 옮겨 정정할 것인가(Manager 권고 — spec/design 수정은 사용자 승인 뒤
  Manager 몫). SHADOW 로트의 선행은 ROADMAP 행(신뢰 앵커 · 골든 개정).
- [x] 7.4 Bound metrics cardinality by fixed market/family/lane/version/reason labels; keep symbol/setup/candidate identifiers in logs/journal queries rather than metric labels.

  **7.4 종결(2026-10-01 — 미착지).** 실측: 생산 메트릭 방출기 0(묶을 label 없음). Manager 판정 (A): 금지 명시 가드
  `internal/strategyprojection/a112_no_metric_emitter_test.go` — internal · cmd 비시험 Go 의 메트릭 API import 금지, 허용 목록(이름) 빈 채,
  머리말에 label 계약(고정 다섯 · symbol/setup/candidate 금지 · 골든 인용). 첫 방출기는 이 시험을 뒤집으며 계약을 세워야 한다. 변이 Q01 · Q02 CAUGHT.
- [x] 7.5 Add performance/operability tests for bounded evidence fan-out, no remote I/O under strategy refresh mutex, independent lane latency, status snapshot consistency and safety cadence under all entry queues saturated.

  **7.5 종결(2026-10-01 — 미착지).** Manager 판정 D1=(A) · D2 · C1~C3. 레인 지연 독립: 한 시장의 레인 넷을 시장 주기 안에서 동시 실행 + join(멈춘 레인은
  자기 마감 시한에 버려지고 이웃은 정상 지연, 시장 지연 = 최댓값 ≤ 마감 시한 1 회 — synctest 등식), 레인 goroutine panic 은 join 뒤 시장 주기에서 다시 던짐,
  버려진 step goroutine 수명 명명(레인당 최대 하나). 상태 행 일관: `Lane.Status()` 한 잠금 + 투영 단일화(AST 핀 · -race 불변식). fan-out(조립 하나 · 물결 0 ·
  제안 하나 → 레인 하나) · 멈춘 원격 물결 중 Read 비차단 · 모든 진입 큐 포화 하 안전 loop cadence. 변이 R01~R13 13/13 CAUGHT, `make test-race` 목록 · 엔진 줄 태그 갱신. 시험 seam 은 태그 빌드에만(생산 바이너리 seam 0 — 5.1.2.1 핀이 첫 구현의 무태그 필드를 잡아 핀 강화와 함께 수리).

## 8. Verification Rollout and Review

- [ ] 8.1 Run focused unit/property/integration tests and race tests for breakout evidence/core, strategyflow, strategyrouter, strategyproposal, scheduler and engine worker/coordinator packages; attach RED-to-GREEN evidence to every Branch Test Map row.
- [x] 8.2 Run dependency/static guards proving lane/evidence/worker packages contain no WTS or broker mutator, writable journal, Guardian issuer, activation/toggle writer and tests cannot POST to a live hostname.
  **닫음(2026-10-04, Manager 판정 결정 1 — a112 안에서 닫는다):** census `measurements/gate-8.1-8.3-2026-10-04/guard-census-8.2.md`(기존 15 가드 GREEN, `guards-8.2.log`) + 8.2 가드 로트(review.md 「8.2 가드 로트」, `measurements/lot-8.2-G/` 변이 13/13). 남는 한계(명기): 공식 클라이언트 기본 Transport 는 DefaultTransport 가드 밖 — officialbars 시험의 official.New 는 WithBaseURL + WithHTTPClient 강제로 막음; 시험 이진의 journal 은 router · coordinator 의 임시 실원장 픽스처에만(이름 예외).
- [ ] 8.3 Run `openspec validate a112-run-four-strategy-families-independently --strict --no-interactive`, PM tracker generation/check, `make sdd-check`, `make test`, `make vet`, `make validate` and `make gate CHANGE=a112-run-four-strategy-families-independently`.
- [ ] 8.4 Refresh all edited-function AST/FLM/BTM/risk reports after GREEN implementation and confirm every branch/risk row maps to an automated test or an explicit reviewed non-code control.

  **8.4 준비(2026-10-04 Manager 판정 — base 무관 로트, 미착지 체크 아님).** `check_analysis` 173 → 150(= 창 요약 2 + 타 change 소관 missing 148, stale 0). stale 15 번들(전부
  a127 82080177 이 낡게 함): 줄 이동 10 은 `shift_same_file_bundles.py`, 몸통 편집 · 구조 동일 2 는 `rebase_bundle.py`, 구조 변경 3 은 a127 아카이브 번들을 옮김(아카이브 좌표 ·
  호출 표 생성). 재고정 후보에서 요구될 a112 함수 `TestProductionWorkersAreExactlyTheEightTheGoldenFroze` 경량 번들 추가(편집 전 FLM 은 7e124a73 에서). 분류 영수증 ·
  창 영수증 · 재고정 모의 `measurements/gateprep-2026-10-04/`, 도구 `harness/gateprep_window_receipt.py` · `render_gateprep_bundles.py`. **남은 것:** base 재고정(사람 승인 —
  권고 1e25b3a3, 그 창 16 함수 전부 지금 FRESH → 승인 뒤 영수증 생성만) · 7.3.1 R3 · 8.5.
  **재고정 집행(2026-10-04, 사용자 승인 — Manager 전달):** base aeeb209e → 1e25b3a3. 귀속 실측 · 새 창 영수증(required 20, FRESH 20) · 게이트 모의 rc=0 — `measurements/repin-1e25b3a3/`, review.md 「base 재고정」. base-commit.txt 단독 커밋. 영수증은 SHADOW 로트 착지 뒤 최종 갱신.
- [x] 8.5 Complete independent adversarial review for owner uniqueness, score calibration, q_final monotonicity, evidence correction/replay, queue/failure isolation, API quota sharing, OFF defaults and safety-loop independence; resolve all P0/P1 findings.

  **8.5 명시 대상 추가(2026-10-01 Manager 판정).** 2.3 (b) 의 `breakoutlane.evaluateFresh` 편집(1.2 반사실 기록 갈래 B7 — High-risk 함수) — 착지 시점 독립 적대 리뷰는
  비례 원칙으로 생략(기록 전용 · `decisionSeal` 무포함 · 입장 경로 바이트 동일 · 쌍둥이 비교 · 판정 변경 변이 CAUGHT)했으므로 이 리뷰에서 덮는다. 증거: review.md 「2.3 (b)」 절,
  `measurements/lot-2.3/`, `function-logic/internal-breakoutlane--evaluatefresh/`. 리터럴 `1_200_000` 두 자리는 변이 핀(CF-03/04 · CF-10)으로 수용(Manager).
  **8.5 명시 대상 추가(2026-10-04 Manager 판정 — 8.8.4 로트 B).** `strategyrouter` 활성화 적재의 필드명 래핑(LoadProductionFamilyActivation 분기 9 → 10 · validate · decode ·
  body · familyActivationRemaining · 새 failedFields)과 `strategyProposalAuthorityLoader.collectMarket` 의 관문 계산 이동. 증거: review.md 「8.8.4 로트 B」, `measurements/lot-8.8.4-B/`.
  **8.5 실행(2026-10-04) — 4판(codex r2 · 보이스 1 · 2 · 3), Manager 최종 판정 HOLD → 응답 로트 완료 · 재검증 시 SHIP.** P0 전판 0. 리뷰 기록 착지 178cc196
  (`analysis/review-8.5-2026-10/` — 합본 `combined-8.5.md`, 설계 브리프 `design-brief-B2-P1.md`). 응답 로트(생산 4 · 시험/증거 8)는 review.md 「8.5 응답 로트」 ·
  `measurements/lot-8.5-R/`. ROADMAP 등재 2: 4 적재기 취소 응답((d), 「핀 선언 전 착지」 면제 불가 선행 · F1-x 후보) · 운영자 표면 행에 P2-g 합류.
  **닫음(2026-10-04, Manager 최종 판정 SHIP):** 응답 로트 6f5b0df6(변이 23/23 + 엔진 태그 전체 5/5 CAUGHT, check_analysis 148 = 148) — 근거 `combined-8.5.md` · review.md
  「8.5 응답 로트」 · 이 판정. 6f5b0df6 별도 리뷰 not-applicable(4판 처방의 구현 — 생존 변이 · 판별 핀이 수락 기준, 새 적대 표면 없음). (d) 는 a112 범위 밖 이연
  (not-applicable in a112 — 기존 4 적재기 공통 결함, 오늘 노출 0; ROADMAP 행이 활성화 핀 선언 전 착지를 강제).
- [x] 8.6 If and only if current A100 ProtectionReady and all dependency gates are complete, build/deploy in dormant OFF/UNOBSERVED mode and verify lane/automation/autostart/LIVE approval remain unchanged. Otherwise perform build-only/shadow-fixture verification, record deployment as BLOCKED, and prove exposure-raising broker requests remain zero.
  **닫음 — 배포 BLOCKED 기록(2026-10-04, Manager 판정; else 갈래 그대로):** A100 ProtectionReady 미완(`a100-wire-fill-to-broker-protection` 미완료 86) · 이미지 빌드 · 컨테이너 교체는 사람 정책 → **이 change 의 배포 = BLOCKED**, 이미지 빌드 · 교체 0. build-only 검증 `measurements/gate-8.1-8.3-2026-10-04/build-only-8.6.log`(go build 무태그 · 태그 · cmd/tossctl exit 0). 상승 0 증명(shadow-fixture, 이름 결속 pass 사건 7/7, 같은 로그): 활성화 없음 → dispatch 전달 0 · 생산 활성화 적재기 매니페스트 없음 · 활성화 없는 조정 불변 · ON 을 말하는 서명 매니페스트로도 레인 DORMANT · 만료 ProtectionReady → 브로커 0 · 생산 worker 휴면 · 보호 증명 실패 → 매수 0.
- [x] 8.7.1 Build the mechanism that *requires* a separate human-approved operating activation: no lane may read effective ON without a verified signed four-family manifest binding the current calibration, market calendar, risk, build and ProtectionReady digests. **(Landed 2026-09-03.)**

  **Split note, 2026-09-03 [a112 결정 59].** 원문 8.7 은 두 절이다: (a) "before any lane effective ON" 을 **요구하는 수단**과 (b) "otherwise rollback entry workers to OFF while retaining shared safety/lineage state" 라는 **운영 자세**. 이 태스크가 (a) 를, 8.7.2 가 (b) 를 가진다. 원문의 모든 절은 둘 중 정확히 하나가 가진다.

  **왜 지금 갈라 여기서 하나.** 5.1.2.2 와 5.2.2 가 **같은** 이 수단에 걸려 있다(사람 결정 (7), HANDOFF.md). 사람이 2026-09-03 에 "매니페스트 + 5.1.2.2 를 한 로트, 5.2.2 는 그다음" 을 골랐다. 매니페스트만 따로 랜딩하면 소비자 없는 매니페스트가 되고, 그것은 이 change 가 5.1 에서 이미 피한 모양이다.

  **어디에 살고, 왜 거기인가.** `internal/strategyrouter` 다. 고른 것이 아니라 두 제약이 정한 것이다: `design.md:221` 이 "exact four-per-market manifest" 를 그 패키지에 배정했고, 여덟 worker 가 사는 `internal/strategyworker` 는 이미 그 패키지를 자기 **허용 폐포**에 들여오고 있다(`dependency_closure_test.go`). 새 패키지에 두면 그 폐포를 넓혀야 하고, `internal/scheduler` 에 두면 desired-state **writer** 가 폐포에 들어와 스펙의 "worker dependency closure 에 activation/toggle writer 가 없어야 한다 (MUST NOT)" 를 깬다. 파일·digest·시각·식별자 검사는 같은 패키지의 감사받은 helper 를 그대로 쓴다(`readProductionRouteFile`·`productionRouteDigest`·`productionRouteTime`·`productionRouteIdentity`).

  **별도 매니페스트인 이유는 잰 것이다.** 같은 패키지의 `strategy-lane-authority-<MARKET>.json` 도 이미 "시장마다 정확히 네 가족"을 서명으로 못 박는다(태스크 4.3, `validProductionRouteCandidates` 가 `len(values) != len(want)` 와 `len(seen) == len(want)` 로 개수까지 대조한다). 그러나 그 매니페스트의 `Desired`/`Effective` 는 **scope(종목·세대)마다** 있고, 여덟 `FamilyWorker` 의 열쇠에는 종목이 없다(골든 `worker_key_fields` 네 필드에 종목이 없다). 종목별 행을 worker 하나의 상태로 접으려면 "모든 scope 에서 ON 이면 ON" 같은 규칙을 **지어내야** 하고 지어낸 규칙은 계약이 아니다. 그리고 8.7 이 이름 부른 build·ProtectionReady digest 가 그 매니페스트에 없다. `design.md:210` 의 "새 runtime activation 은 exact 4-family-aware signed manifest 가 필요하다" 와 8.7 의 "**separate** human-approved operating activation" 이 같은 말이다.

  **불투명함을 셈이 아니라 서명이 지킨다.** `FamilyActivation` 의 필드는 전부 비공개이므로 이 패키지 밖에서는 **영값만** 만들 수 있고, **영값이 안전한 값이다** — 아무것도 승격하지 않는다. 그래서 "켜진 worker 를 아무나 만들 수 있는가" 를 열거표로 지킬 필요가 없다: 승격을 얻는 유일한 길이 ed25519 검증을 통과하는 것이다. 5.5 가 세 라운드 동안 배운 것("이름·철자 검사는 반드시 뚫린다")의 정반대 방향이다.

  **다섯 digest 는 두 단계에서 결속한다.** 보정·경로 매니페스트·달력·빌드는 제안 수집 단계에 존재하므로 거기서 결속하고, 위험 번들과 ProtectionReady 는 그 단계에 **없으므로**(둘 다 제안 뒤에 수집된다) `buildProductionStrategyMarketWorker` 에서 결속한다. 없는 사실을 결속하면 그 결속은 어떤 정상 입력으로도 참이 될 수 없고, 그것이 이 change 가 이미 한 번 만들었다 고친 문 없는 fail-closed 다. 검증은 한 번, 결속은 사실이 사는 자리에서.

  **반증이 코드를 네 줄 지웠다.** 가장 무거운 것: "서술자는 정확히 넷" 규칙을 세 검사가 나눠 지키고 있었고 셋 중 **아무 둘이면 충분**하므로 셋을 각각 지운 변이가 전부 살아남았다 — 각자가 상대의 시험을 통과시킨다. 개수 검사가 나머지 둘의 재진술이라 그것을 지웠다. 그리고 닿지 않는 방어 셋(`Verified` 의 `validMarket`, 검증의 `len(want) == 0`, 관문의 `loader.lanes == nil`)을 지웠다. `len(want) == 0` 은 지우는 대신 **닫아 두는 등식**을 시험으로 만들었다: 파일 이름이 있는 시장의 집합 == 서술자 표가 있는 시장의 집합. 상세는 review.md 의 2026-09-03 절.

  **이 태스크가 주장하지 않는 것.** 서명된 매니페스트를 실제로 만들어 배포하는 절차는 여기 없다. 생산에는 그 파일이 하나도 없고(측정: `find ~/.config/tossctl -maxdepth 2 -name 'strategy*'` → 0 건), 그래서 오늘 생산 동작 변화는 0 이다. 활성화가 없거나 만료·폐기됐을 때의 운영 rollback 자세는 8.7.2 다.

- [x] 8.7.2 Rollback entry workers to OFF when that activation is missing, expired or revoked, while retaining shared safety and lineage state. **(Landed 2026-09-27, [a112 결정 62].)**

  **닫은 구멍.** 8.7.1 의 관문은 로드 오류를 종류 없이 "관문 없음" 으로 접었고, 관문이 없으면 기존 시장 단위 경로가
  돌았다. 그래서 사람이 매니페스트에서 끈 가족이 활성화가 만료·폐기되는 순간 기존 경로로 **되살아났다** — RED 가
  네 행(만료·폐기·파일 없음·핀 뒤 바뀐 바이트)에서 그 넓힘을 값으로 보였다.

  **판별자 = 배포 핀의 존재 (결정 62).** 빈 핀 → `ErrProductionFamilyActivationUndeclared` → 기존 경로(= 오늘 생산, 핀 0 건
  측정 → 생산 동작 변화 0). 핀이 있으면 무엇이 틀렸든 관문이 `rolledBack` 으로 서고, 네 레인은 영값 활성화로 DORMANT(잠긴
  레인은 LATCHED)를 내고, 범위가 전부 지워져 그 시장은 `FAMILY_GATE_CLOSED` — **신규 진입만** 닫힌다. 판별 규칙은
  strategyrouter 한 곳(적재기 첫 줄), 만료 판정도 한 곳(`familyActivationRemaining`).

  **lease 와 최종 검사.** 주문 lease TTL 은 가족 활성화의 남은 수명으로 min 결합만 된다(`FamilyActivation.LeaseCeiling`,
  q_final admission 앞). Codex·적대 리뷰가 연 P1 뒤 dispatch 주기가 실시계(`now`, 생산 조립이 `clk.Now`)를 갖고, 게이트웨이가
  브로커 바이트 전에 부르는 최종 검사가 가족 만료를 다시 본다. 불변식은 이 형태다: lease 행은 명목상 만료를 최대 δ(파도→발급)
  넘을 수 있으나 SUBMITTING 은 만료 뒤 최종 검사를 통과할 수 없다. 시계가 없는데 가족 활성화가 검증돼 있으면 거절한다 —
  생산에서 그 조합에 닿는 조립은 0(측정).

  **보존.** 되돌림 경로(`installed`·`admit`·`familyGateFor`·`loadFamilyActivation`)의 모든 호출은 허용 목록 census 가 읽기로
  못 박는다. 잠긴 레인은 되돌림 동안·그 뒤에도 잠긴 채다. lease·`PlaceClaimedStrategy`·가족 활성화 참조는 손절·청산·대사·
  체결 감지·보호 9 패키지에서 0 이다.

  **이 태스크가 주장하지 않는 것.** 재시작을 넘는 세대 래칫(원장 기록 필요, 선례 5.3.3), 한 번 활성화된 시장을 재시작 넘어
  계속 닫는 것(v34 필요), 프로세스 안 파일 폐기의 최종 검사 반영(다음 파도부터), 미선언/검증을 화면에서 가르는 것(8.8.4 인접).
  상세·측정·리뷰 네 보이스·뒤집은 기존 시험 둘은 review.md 의 2026-09-27 절.

## 8.8 8.5 독립 적대 리뷰가 연 수정 (2026-09-04, 결정 60)

8.5 를 리뷰어 7 + 다른 모델 1 로 실행했다(범위: `06be6ca9^..515f85a8`). 판정은
**이대로 main 불가**다. 아래는 그 리뷰가 연 태스크이며, 8.7.1 이 landed 로 남아
있어도 그 landing 은 여기 셋이 닫히기 전까지 **배포 가능하지 않다**.

리뷰 기록·근거·리뷰어가 틀린 것까지 review.md 의 2026-09-04 절에 있다.

- [x] 8.8.1 관문 거절이 시장의 소유자 범위 수를 줄여 downstream 의 "정확히 하나"
  관문을 **만족시키는** 경로를 닫는다. 경로에 오른 범위가 레인을 갖고도 봉투를
  하나도 못 내면 그 시장을 닫는다(5.4.2·5.4.3 과 같은 처리). 같은 범위 안에서
  가족 하나가 막히고 이웃 가족이 이기는 것은 그대로 둔다 — 그것이 이 change 가
  사려던 격리다. 멈춘 결과 종류를 스냅샷에 실어 `gated` 가 읽히게 한다.

  **(Landed 2026-09-04.)** `FAMILY_GATE_CLOSED` 로 시장을 닫고 `RefusedCount` 에
  `RoutedCount` 를 그대로 넣는다 — 5.4.2·5.4.3 과 같은 처리다. 세는 단위는 **제안이
  아니라 소유자 범위**다: 범위의 레인이 전부 관문에 막혔을 때만 닫는다. 일부만 막힌
  범위는 이웃 가족이 같은 범위에서 이기므로 목록 길이가 그대로이고, 그것은 이 change 가
  사려던 격리이지 소멸이 아니다. `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading`
  (역전형을 잠가 두 범위가 다 봉투를 냄)과
  `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`
  (지속형을 잠가 `000660` 범위가 소멸)가 그 둘을 갈라 세운다. `arbitration.gated` 는
  이제 `GatedCount`·`GatedOutcomes` 로 스냅샷에 나간다. 상세는 review.md 2026-09-04 절.

- [x] 8.8.2 매니페스트가 결속하는 다섯 값을 **서명 가능한 값**으로 바꾼다.
  `risk_bundle_digest` 는 per-cycle 스냅샷 봉인(`Symbol`·`AsOf` 포함)이라 어떤
  정상 입력으로도 참이 될 수 없었다. 운영자가 이미 관리하는
  `TOSSOS_RISK_BUCKET_<M>_MANIFEST_SHA256` 을 쓴다. ProtectionReady 는 살아 있는
  상태라 등식으로 결속할 수 없으므로 **하한 세대**로 바꾸고, 그 판정을 주문 경로에
  둔다. 지금 판정은 `buildProductionStrategyMarketWorker` 안에 있어 화면만 바꾼다.

  **(Landed 2026-09-04.)** 매니페스트 본문이 `risk_bundle_digest` 대신
  `risk_policy_digest`(운영자가 이미 핀하는 `TOSSOS_RISK_BUCKET_<M>_MANIFEST_SHA256`)
  를, `protection_ready_digest` 대신 `protection_ready_min_generation`(uint64, 0 금지)
  을 싣는다. 앞의 넷은 제안 수집 단계가 등식으로 결속하고 — 넷 다 그 단계에
  존재하며 활성화 수명 동안 변하지 않는다 — ProtectionReady 하한만
  `strategyDispatchCycle.dispatch` 가 결속한다. 그 자리가 보호 세대를 들고 있으면서
  주문을 거절할 수 있는 유일한 자리다. `buildProductionStrategyMarketWorker` 의
  결속은 들어냈다(분기 둘 삭제): 값이 충족 불가였고, 그 함수가 만드는 `Effective`
  는 화면과 승격만 움직여 주문을 하나도 막지 못했다.

  하한을 **등식이 아니라 하한**으로 둔 것은 ProtectionReady 가 살아 있는 상태이기
  때문이다. 세대는 단조 증가하므로 하한은 안전 방향으로만 어긋난다. 이 편차는
  8.7 의 "current … ProtectionReady digests" 를 문자 그대로 따르지 않으므로 여기
  적는다 — 문자 그대로 따르면 어떤 정상 입력으로도 참이 될 수 없다.

  시험은 값을 시스템에서 읽지 않는다. `TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor`
  가 하한을 숫자로 적고 배선의 보호 세대(9)와 견주며, 세 경우(활성화 없음 / 같음 /
  낮음)를 함께 세워 "항상 거절" 판본을 배제한다. 반증 둘이 서로 다른 행을 빨갛게
  만든다. 상세는 review.md 2026-09-04 절.


- [x] 8.8.3 이 매니페스트를 사람이 만들 수 있게 한다: 정규 바이트를 내는
  `tools/` 도구, `docs/operations.md` 의 형제 절과 같은 문서, 그리고 검증되는 골든
  픽스처 하나. 시험은 **매니페스트 바이트를 먼저 만들고 나서** 엔진을 돌린다 —
  결함이 사는 축이 값이 아니라 순서다.

  **개정 2026-09-04 [a112 결정 61] — 서명을 뺀다.** 원문은 "정규 바이트를 내고
  **서명하는**" 이었다. 사람이 세 선택지 중 2번을 골랐다: ed25519 를 빼고 신뢰
  앵커를 외부 digest pin 하나로 둔다. 근거와 이 개정이 바꾸지 않는 것은
  `design.md` §8 의 같은 날짜 개정 블록에 있다.

  **그래서 이 태스크가 가져가는 것이 하나 늘었다** — 서명 제거 자체가 여기 속한다.
  8.7.1 은 `[x]` 로 두되 그 본문의 서명 문장은 이 태스크가 정정한다. 이유는 순서다:
  도구·문서·골든이 서명 유무에 따라 다른 물건이 되므로, 서명을 먼저 빼지 않으면
  이 태스크가 만들 물건이 만들자마자 틀린 것이 된다.

  **골든 픽스처가 이 로트에서 특히 중요한 이유.** 오늘 로더 시험은 검증기 자신의
  직렬화기로 바이트를 만든다(`json.Marshal(body)` → 시험, 그리고 검증기도 같은
  호출). 그래서 "사람이 만들 수 있는 바이트" 와 "검증기가 받는 바이트" 가 갈려도
  전부 초록이다. 그리고 승격 경로 시험은 파일을 아예 안 지나간다
  (`FamilyActivationForTest` 가 구조체를 직접 주조한다). 골든은 **커밋된 바이트**를
  읽고 그 SHA-256 을 시험에 **손으로 적은 리터럴**로 대조해야 한다 — 실행 중인
  시스템에서 읽어 비교하면 무엇이든 통과한다(8.7.1 이 그렇게 충족 불가능한 결속을
  초록으로 통과시켰다).

  **(Landed 2026-09-04.)** ed25519 가 없어졌다: `signature`·`key_id`·
  `signature_algorithm` 필드와 `TrustedKeyID`·`TrustedKey` config 가 사라지고,
  배포가 관리할 값이 시장마다 셋에서 **하나**(`…_MANIFEST_SHA256` 핀)로 줄었다.
  domain 상수도 `…/ed25519/v1` → `…/sha256-pin/v1` 로 바꿨다 — 서명이 없는데
  도메인이 ed25519 를 말하면 그 문자열이 거짓말이 된다. 남은 문 셋(현재 UID 소유
  `0400` 정규 파일 · 핀과 바이트 동일 · 이 빌드의 정규 직렬화와 한 바이트도 다르지
  않음)은 그대로다.

  산출물 넷: `tools/a112-family-activation`(비밀 없는 생성기, `0400`·`O_EXCL` 로
  쓰고 env 핀 줄을 낸다), `docs/operations.md` 의 운영 절, 커밋된 골든
  `internal/strategyrouter/testdata/strategy-family-activation-KR.json`, 그리고 새
  exported 표면 둘(`FamilyActivationDocument`·`EncodeProductionFamilyActivation`).
  **서술자 넷은 문서에 없다** — 켤 가족만 받고 lane id·버전·수평선은 검증기가 쓰는
  바로 그 표에서 유도한다. 운영자가 손으로 적으면 표가 바뀔 때 조용히 옛 이름을
  넣게 되고 그 거절은 배포 시각에야 보인다.

  **독립 적대 리뷰가 P1 셋을 열었고 전부 고쳤다** (판정 HOLD → 수정 완료).
  (1) 문서가 없는 env 이름(`TOSSOS_STRATEGY_ROUTE_…`)을 불러 **작동하는 매니페스트를
  만들 수 없었다** — 실제 이름은 `TOSSOS_STRATEGY_LANE_…` 이다. (2) 도구가 `0400`
  파일을 덮어쓰려 해 **재발급이 이틀째부터 실패**했다 — `O_EXCL` 로 바꾸고 재발급
  4 단계를 문서에 적었다. (3) **`EncodeProductionFamilyActivation` 을 부르는 시험이
  하나도 없어서** 서술자 정렬·유도가 어디서도 실행되지 않았다. 그래서 review.md 의
  M3 = CAUGHT 는 **틀린 기록**이었고, 그 행을 정정한 뒤 저작 경로 시험 둘을 심었다
  (골든 바이트 등식 · 64회 순서 안정성). 한 회로는 순열 24 가지 중 하나를 우연히
  맞혀 초록이 되므로 둘로 갈랐다 — 실제로 M3 재실행이 그 우연을 보여 줬다.

  P2 넷도 처리했다: 가드가 점 import·빈-import-우선 두 우회를 못 보던 것(세는 범위가
  입력의 함수였다 — M7 과 같은 기전, 반증값을 커밋된 시험 넷으로 심었다), 사라진
  서명을 계속 주장하던 주석 9곳·시험 이름 4개(둘은 `Makefile` 의 race 목록에 있어
  이름 변경 뒤 **실제로 도는 개수 18**을 세어 확인했다), 기록이 빠뜨린 교환 하나
  (**승인자 부인방지**를 잃었다 — `actor` 는 인증되지 않은 자유 문자열이고, "누가
  승인했는가" 는 이제 배포 이력이 답한다), 그리고 `-revoked`·값 출처·service UID
  문서 빈칸. 상세는 review.md 의 2026-09-04 두 절.

  **생산 동작 변화 0.** 생산에 매니페스트도 핀도 없고(측정), 로더는 핀 형식 검사에서
  파일을 열기 **전에** 단락하므로 새 I/O 도 없다. 신뢰 앵커가 하나 줄었을 뿐 남은
  앵커는 그대로다. 만료·폐기 시의 rollback 자세는 8.7.2 가, 오류 구별은 8.8.4 가
  가져간다.

- [x] 8.8.4 P1 정리: 오류를 필드 이름과 함께 감싸 스냅샷에 내기(만료·폐기를
  "배포 안 함"과 구별), 닫힘 갈래 12 중 7 이 영값 활성화를 싣는 것, 골든의
  desired/effective 단언 무효화, 엔진 race 목록에 관문 시험 등록,
  `promotion` 인자를 검증된 활성화로 도는 시험, `*_testseam.go` 빌드 태그 가드,
  그리고 레인을 제안 수집 앞으로 옮기며 생긴 원장 오류의 양시장 전파.

  **8.8.4 진행(2026-10-04 Manager 판정 — 두 로트 A → B).** 로트 A(생산 .go 0): 항목 3 `TestTheGoldenOffIsEachWorkersDefaultAndOnlyItsSignedActivationFlipsIt`
  (공허한 상수-대-상수 절 교체) · 항목 4 관문 시험 두 파일 race 등재(8.7.2 기록 정정) · 항목 5 `TestAVerifiedPromotionLetsTheOwningLaneEmitThroughTheLaneRuntime`
  (+ 무태그 생산 `laneStepFor` 경유 시험 실재 확인) · 항목 6 `tools/sdd/test_testseam_build_tags.py` · 항목 7(b) 전파 불변 = 4f49a8eb 의 의도된 결정 문서화 ·
  runMarket FLM 표 ast 기준 수리 — review.md 「8.8.4 로트 A」. 남은 것 = **로트 B**: 항목 1(strategyrouter sentinel 필드명 래핑, errors.Is 보존) · 항목 2((ii) gate
  활성화 carry + 폐포 전수 표 시험) — 생산 편집, 8.5 재리뷰 명시 대상. 잔여(로트 B 에서 등재): 활성화 실패 사유 · `SwallowedCycleErrors` 의 운영자 표면.

  **8.8.4 종결 — 로트 B(2026-10-04, Q-B1=(c) · Q-B2 계산/판정 분리 — 미착지).** 항목 1: `strategyrouter/production_family_activation.go` 의 맨 sentinel
  반환 18 곳(측정 — 지시의 "10" 정정) 전부 필드명 `%w` 래핑, errors.Is 보존(`==` 비교 0). 복합 결속 셋(설정 결속 · 몸통 결속 · 수명)과 서술자 필드 검사는 분기 하나
  그대로 `failedFields` 로 어긋난 필드 **전부**를 한 메시지에(단락 평가 의존 항 0 확인). 읽기 결함과 핀 불일치는 갈래를 갈라(Load B5/B6) 결함은 읽기 함수 오류를 사슬에,
  불일치만 `manifest_digest`. 시험 `a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`(실패 모양 40 — 단일 · 다중 동시 ·
  결함/불일치 분리) · `TestTheDocumentPathNamesItsRefusalToo`. 항목 2: `collectMarket` 의 관문 **계산만** ROUTE_NOT_READY 가드 직후로 — 13 닫힘 중 첫째만 영값(사유 명명),
  열둘이 관문 활성화; kind 순서 · FAMILY_GATE_CLOSED 자리 불변. 시험 `a112_proposal_closure_carriage_test.go` `TestTheThirteenProposalClosuresKeepTheirOrderAndTheGateIsComputedRightAfterRouteReadiness`
  (AST census — kind 순서 + 관문 자리 + gate.activation 을 싣는 리터럴 수) · `TestEveryReachableProposalClosureCarriesTheGatesActivationExceptRouteNotReady`(닿는 닫힘 여덟 × 관문
  세 모양 — 사유 · 실은 활성화 · 적재 호출 수). 거짓 전칭 주석 정정. Pre-Edit `lot-8.8.4-B/pre-edit/`(여섯 함수) · RED `red-8.8.4-B.log` · 재번호 `renumber.txt`(difflib) · 편집 뒤 번들
  일곱(`harness/render_884b_bundles.py`) · 변이 `mutation-8.8.4-B.tsv` 13/13 CAUGHT · 이름 결속 22/22. **잔여(활성화 로트 선행 — ROADMAP):** 활성화 실패 사유 · `SwallowedCycleErrors`
  의 운영자 표면(snapshot/projection 필드) 0 — 사유는 오류 사슬에만 있다; 공유 읽기 함수 `readProductionRouteFile` 이 OS 원인을 자기 sentinel 하나로 접는다(파일 없음 · 권한 ·
  소유자 구별 불가).
