# Review — a071-wire-kr-us-protection-readiness

- Date: 2026-08-04
- Stage: independent security review BLOCK; production remains read-only `UNWIRED`, task 3.5 reopened
- Voices: Manager safety review, independent operations/security review, authority-boundary self-review

## Findings and disposition

- Readiness is market-scoped `WIRED|UNWIRED` plus typed refusal; no combined KR+US state exists.
- Signed attestation binds pinned trust roots, key ID/algorithm allowlist, rotation/revocation, monotonic
  serial, maximum lifetime and durable trusted-time floor.
- Scope includes exact broker client-key echo, lookup/uniqueness, pending/terminal/cancel query,
  dedup/idempotency and replace semantics. Missing capability is `UNWIRED`, not guessed support.
- Submit/cancel unknown and orphan orders use exact identity reconciliation only; no symbol/time inference,
  blind resubmit or inferred cancellation is allowed.
- The isolated core verifies a canonical JSON envelope signed with Ed25519. The pinned policy has an explicit
  algorithm allowlist, key lifecycle windows, revocation timestamps and bounded rotation overlap.
- Monotonic serial is durable at `(account, profile, market)` scope, so rotation to a different key ID cannot
  reset the counter. Trusted time has a sealed durable floor and assessment is a pure old-state/new-state step.
- **Resolved independent-review HIGH:** every valid non-rollback trusted-time observation advances the durable
  floor even when evidence is missing or invalid. A later rollback cannot hide inside an unattested interval.
- **Resolved independent-review HIGH:** corrupt durable state is returned as the exact non-committable preimage;
  assessment never clones and re-seals missing or modified serials into a repaired state.
- File bytes, resolved path, owner, exact `0600` mode, regular/symlink status and size are sealed modeled inputs;
  duplicate and unknown JSON fields are distinct typed refusals.
- Runtime scope binds exact account/profile/market/order/session/quantity/trigger/replace/tool/build/evidence and
  the complete broker client-key/lookup/uniqueness/query/dedup contract.
- A market becomes `WIRED` only when both attestation and an exact sealed supervisor binding validate. Snapshot
  fields and all authority-producing constructors are private; the only public constructor returns paired
  `UNWIRED` defaults.
- **Resolved independent-review HIGH:** the public `Options.ProtectionOverrideForTest` scalar forge and exported
  WIRED/UNWIRED test pointers were removed. `ProfileProtection` remains `UNWIRED` compatibility reporting only;
  engine status overrides are never forwarded to Gateway execution authority. Static and reflection tests reject
  any non-test declaration or exported scalar readiness field.
- Gateway admission uses a market-scoped sealed adapter, and the exact checkpoint generation/identity is
  revalidated immediately inside `DispatchVerified` before broker transport. Missing provider, attestation,
  supervisor evidence, scope mismatch, expiry, corruption or drift refuses only the exposure-raising mutation in
  that market; a peer-market seal remains independently usable.
- Reduce-only SELL, CANCEL and exposure-reducing AMEND succeed without reading the readiness provider. Static
  source isolation keeps stop, emergency liquidation, reconciliation and fill paths free of this dependency.
- The dormant lifecycle core derives stable operation identity from exact account, position, market, generation,
  revision and operation kind. Submit recovery uses that exact key; cancel and replacement recovery use the exact
  broker order ID. A scoped `NOT_FOUND` result must repeat every identity field before a same-key retry is even
  considered.
- Entry is closed while unprotected, pending, unknown, reconciling or terminal. It opens only for an unlatched
  ACTIVE order whose broker claim plus other sell claims exactly equals holdings. Registration and replacement
  therefore require full available coverage; partial fills reduce holdings and broker claim in the same pure
  transition and re-evaluate the equality.
- Unknown replacement and cancellation retain the pre-existing ACTIVE observation. Replacement is a single
  continuous-coverage command, refuses trigger retreat and never models cancel-then-place. Unknown submit without
  attested idempotency becomes no-resubmit reconciliation.
- Duplicate fills are stable-ID no-ops; conflicting duplicate fills latch reconciliation. Unowned orphan orders
  are never adopted, canceled or guessed, and a conflicting re-observation preserves the first evidence.
- The KR and US position maps share only a canonical container seal. A KR recovery latch does not alter US ACTIVE
  state, and exact fill/exit handling remains immediate in both markets.
- KR and US are assembled concurrently from one paired manifest into two exact market contracts. Neither market's
  readiness is a prerequisite for constructing or verifying the other; a market-local artifact drift closes only
  that market, while corrupt shared policy/state/trusted-time inputs close both.
- **Security review C1/C2/M7 disposition:** the repository has no production caller from a journal-committed fill
  into an exact journal-derived stop/expiry and durable protection Plan/Register lifecycle. Therefore both KR/US
  production assemblies now claim `Wired=false`; engine construction contains no official protection gateway,
  controller minter, arbitrary `GatewayFactory`, or protection DB. Exact activation/dispatch revalidation cannot be
  inferred or defaulted; task 3.5 remains open until the complete fill lifecycle is independently verified.
- **Security review H3 disposition:** cached verdicts are re-evaluated per market at attestation expiry, key
  revocation and key overlap boundaries with exact instant fail-close, even when only the peer market fingerprint
  changed. Revalidation reads both verdicts from the original sealed snapshot before resealing the result.
- **Security review H4 disposition:** an owner-only cross-process flock serializes load/assess/store. File existence
  is not bootstrap authority: an exact marker is written and fsynced only after valid state is loaded or durably
  stored. Empty lock plus absent state is first bootstrap; marker plus absent state is corruption. A peer's newer
  serial invalidates stale in-process cache before any decision is returned.
- **Security review H5 disposition:** production has no protection SQLite startup dependency. A colliding/failing
  `protection.db` path cannot stop journal, Gateway, exit/fill/reconcile safety runtime assembly; entry remains
  closed. Manifest contracts are built locally and atomically published only after both markets validate, so a
  malformed second market cannot leak a partial contract that fails engine construction.
- **Security review M6 disposition:** manifest, attestation, evidence and state reads walk every absolute parent
  component from a pinned root dirfd using `openat(O_NOFOLLOW)` and validate the final owner-only directory by
  `fstat`; file reads use the same parent fd and descriptor `fstat` before/after reading.

## Verification

- Strict OpenSpec validation: PASS.
- `go test ./internal/protectionreadiness -count=1`: PASS.
- `go test -race ./internal/protectionreadiness -count=1`: PASS.
- `go vet ./internal/protectionreadiness`: PASS.
- Affected packages (`protectionreadiness`, `protection`, `execgw`, `reconcile`, `app/engine`) targeted and full
  tests: PASS; `cmd/tossctl`: PASS.
- Affected-package `go test -race`: PASS (`execgw` 264.135s, `reconcile` 101.866s, `app/engine` 288.622s).
- Affected packages plus `cmd/tossctl` `go vet`: PASS; strict OpenSpec: PASS; `git diff --check`: PASS.
- Post-edit Function Logic Maps are complete for every a071-owned changed function. The repository-wide checker
  reports only concurrent a066/a073 worktree functions outside this wave.
- `make sdd-sync` refreshed the CodeGraph fingerprint (11 changed files). Its advisory CodeGraphContext phase
  stalled for more than two minutes and was interrupted as permitted; no runtime or broker state was touched.
- `FuzzArbitraryAttestationNeverWires` and `FuzzSerialMustStrictlyIncrease` (3s each): PASS.
- Statement coverage: 87.6%.
- Static dependency/API tests exclude live transport, runtime mutation packages and exported trust/evidence/
  supervisor/state minting constructors.
- Existing protection controller, gateway, engine and journal integration tests remain pending by design.
- Isolated lifecycle unit/crash-matrix tests, race and vet: PASS.
- Lifecycle fuzz (`FuzzOperationIdentitySeparation`, `FuzzDuplicateFillNeverDoubleDecrements`, 3s each): PASS.
- Lifecycle statement coverage: 83.4%.
- Post-edit Go AST evidence: `lifecycle.go` SHA-256
  `de50441bc89c79ec5cfeb8308a837db1cffede7e3ab52c661eccb9515d1688e5`; `state.go` SHA-256
  `df5c5459c6d2add80bcfdcadb03f4af1dbcb19cce1424687c2d855a40fd66cbe`.
- Test-only fake official broker has no socket or hostname. Static dependency/API tests reject live transport,
  runtime protection/execgw/engine/journal/broker imports, toggle/lane/LIVE approval authority and exported
  authority-minting functions.
- `Mutations` counts one atomic pure durable-state transition; `ExternalMutations` remains zero. A separate
  `StateCommitAllowed` bit prevents persistence of corrupt/untrusted-time results.

## Verdict

The isolated cores and Gateway decision boundary remain available for review, but production supervisor/official
gateway assembly is intentionally withdrawn pending the complete journal-committed fill lifecycle. KR and US remain
concurrent read-only lanes and default independently to `UNWIRED`. This change creates no lane, activation,
automation or LIVE authority and makes no official order call. Task 3.5 and final approval remain blocked.

## Addendum — 2026-08-11 (scope change, not a re-review)

The 2026-08-04 verdict above is unchanged and was not re-run. What changed is ownership, not the finding.

Task 3.5 is superseded by `a100-wire-fill-to-broker-protection`. The blocking precondition this review named
("no production caller from a journal-committed fill into an exact journal-derived stop/expiry and durable
protection Plan/Register lifecycle") is now a100's subject. This change's set is frozen at task 3.4; the C1/C2/M7
disposition — both production assemblies claim `Wired=false`, engine construction contains no official protection
gateway, controller minter, arbitrary `GatewayFactory` or protection DB — therefore stands as the shipped state of
this change rather than as a temporary hold.

Two facts a100 measured against this change's code, recorded here because they bind whoever reviews a100:

- `NewProductionProvider` (`internal/protectionreadiness/production.go:293`) refuses to configure at all unless
  **every** supplied assembly has `Wired: true`. Today's production `UNWIRED` is therefore an unconfigured
  provider, not a per-market verdict; `Current` returns `pairedRefusalSnapshot(RefusalInvalid)` once a manifest
  pin is present. Flipping the assembly bool activates a configuration path that has never run in production.
- `assembly.ComponentDigest` must equal the manifest's `supervisor_digest` (`production.go:187`), and that digest
  is derived from the build digest, which includes VCS revision and build settings. The signed manifest is
  therefore build-bound: a rebuild invalidates it. a100 must state how that is operated.

Rescinding this addendum restores task 3.5 to this change.

## 5.2 실측 — Manager 실행 (2026-09-25, HEAD d6644e17, 격리 워크트리 TossOS-worktrees/archive-batch1)

작성자(Opus 팀메이트, 5.1 을 0233d776 으로 닫은 뒤 API 한도로 중단)와 분리된 검증 패스. 코드 변경 0. 로그 `/tmp/claude-1000/a071-52/`.

| 항목 | 명령 | rc | 결과 |
|---|---|---:|---|
| 표적 패키지 | `go test -count=1` attest·protection·protectionlifecycle·protectionofficial·protectionreadiness·execgw·app/engine (7) | 0 | ok 7 (`01-test.log`) |
| seams | 같은 7 + `-tags tossos_testseams` | 0 | ok 7 (`02-seams.log`) |
| race(표적) | 같은 7 `-race -tags tossos_testseams` | 1 | protection 계열 5 ok · **app/engine·execgw 는 `test timed out after 10m0s`(DATA RACE 0)** — 두 패키지 전체 -race 는 저장소 정본 밖(`make test-race` 는 이름 목록만 돈다) (`06-race.log`) |
| race(정본) | `make test-race` | 0 | ok 8, DATA RACE 0 (`11-make-test-race.log`) |
| journal crash/restart | `-run 'Crash\|Restart\|Recover'` app/engine·execgw·journal, seams 태그 | 0 | ok 3 (`07-crash-restart.log`) |
| vet | `make vet` | 0 | (`08-vet-all.log`) |
| 전체 test | `make test` | 0 | ok 99 (`09-make-test.log`) |
| OpenSpec | `openspec validate a071-… --strict --no-interactive` | 0 | (`05-validate.log`) |
| PM | `generate_master_tracker.py --check` | 0 | current (`10-pm-check.log`) |
| 5단계 | `check_analysis.py --change a071-…` | 1 | **남의 함수 374 요구**(weeklyvaluelane 등, 창 정책 — a122 5.6/5.7·후속 a123 초안) (`04-ca.log`) |

5.2 는 위 표로 닫는다. 5.3(`make gate`)은 5단계가 창 정책으로 성립하지 않아 미실행 — 정책 결정 뒤. 출하 상태(Addendum C1/C2/M7) 재확인은 Opus 팀메이트 로트에 남긴다(이 패스는 시험 실행만).

## 종결 시퀀스 (2026-09-27, a122·a077·a079 선례 · Manager 1차 로트 배정)

### 5.1 — 0233d776 재결속의 현재 HEAD 유효성

- HEAD 9494e0e6 에서 check_analysis: stale 0 · 해시 불일치 0 · 번들 형식 오류 0
  (`/tmp/claude-1000/audit19/ca-a071-wire-kr-us-protection-readiness.log` — 남은 것은 옛 base 창의 required 뿐).
  그래서 refresh 대상 번들은 **0** 이고, 번들 40(`revision: current` 35 · `base` 5, 파일 160)은 바이트 불변이다
  (sha256 전후 대조 일치, `/tmp/claude-1000/a071-lot/`).
- 자기 Go 커밋 둘의 기존 함수 수정: 171739a4 19 · 6aec9791 19, 합집합 **33** — 33 전부 이 change 의 번들이
  덮는다(`changed_existing_functions(c^, c)` 와 번들 `ast.json` 의 (file, function) 대조, 누락 0).

### base 재고정 영수증 (77e36cca)

- 잰 순간: HEAD 9494e0e6, 모집단 = `git log --full-history -- openspec/changes/a071-…` 커밋 13.
  옛 base 775c37cb 뒤 go 를 가진 커밋: 171739a4 · 6aec9791(자기 작업, 위 33 전부 번들로 덮임) ·
  8022f578(a072/a073 squash, 이 디렉터리 파일 1) · 4c6927ea(a100) · 448dfeb1(통합 병합) — 뒤의 셋은 자기 작업 아님.
  0233d776 이후 자기 Go 커밋 0.
- 옛 base 775c37cb 에서 5단계 rc 1 · required 396 · missing 372 · 창에 착지 커밋 457 — missing 은 전부 다른 change 의 기존 함수.
- 재고정 775c37cb → 9494e0e6, `base-commit.txt` 만 커밋(77e36cca). 재고정 뒤 check_analysis rc 0
  (required 0, evidence complete or diff-proven exempt). 번들 40 은 창 밖이 되어도 매 판정에서 해시가 대조된다.
  `--record-landing` 은 쓰지 않았다(Manager 결정).

### 3.5

a100 으로 이관 — tasks.md 「## 6. Supersession — task 3.5 → a100」(195dd972, 2026-08-11)과
`openspec/changes/a100-wire-fill-to-broker-protection/proposal.md` 「## Supersession — a071 task 3.5 (분할)」.
변경 집합은 3.4 까지로 확정되고 a100 의 변경분은 a100 의 게이트가 본다. 되돌림 조건(a100 취소·배선 제외)은 tasks.md §6 그대로다.

### 출하 상태(Addendum C1/C2/M7) 재확인 — 5.2 절이 이 로트에 남긴 것 (HEAD 9494e0e6, 읽기만)

- 두 생산 assembly 가 `Wired: false` — `internal/app/engine/protection_wiring.go:41-42`(KR·US, component 에 "fill-lifecycle-unwired").
- 엔진 구성에 공식 보호 gateway 없음 — `internal/protectionofficial` 을 import 하는 비시험 파일 0(자기 패키지 제외).
- `internal/protection` 을 import 하는 app 코드는 `internal/app/engine/gateway.go` 하나이고, 쓰는 것은 readiness 어댑터
  `protection.NewPairedReadinessAdapter`(`gateway.go:292`) 뿐이다 — controller minter · `GatewayFactory` · `protection.db`
  는 `internal/app`·`cmd` 비시험 코드에 0.
- `Wired: true` 비시험 출현 셋(`internal/console/protection_liveness.go:58` · `orders.go:312` · `holdings.go:220`)은 콘솔의
  다른 타입(liveness·화면 view)이며 `protectionreadiness.SupervisorAssembly` 가 아니다.
- 결론: Addendum 의 출하 상태가 현재 HEAD 에서 그대로다. 이 확인은 저자(Opus 팀메이트)의 사실 대조이지 독립 적대 리뷰가 아니다.

### 5.3 의 "adversarial independent review" 충족 판정 (Manager, 2026-09-27)

별도 적대 리뷰는 요구하지 않는다. 근거:
1. 독립 적대 리뷰는 이미 있다(2026-08-04). 그 유일한 차단 사유는 3.5 였고, 3.5 는 문서화된 supersession
   (a100 proposal 「Supersession — a071 task 3.5 (분할)」, tasks.md §6 195dd972)으로 해소됐으며 08-11 Addendum 이
   "verdict stands as the shipped state" 라 명시한다.
2. 이 종결 로트의 편집은 Go 0 줄이다(base-commit.txt · review.md · tasks.md, 번들은 무변 검증뿐) — 새 리뷰 대상 코드가 없다.
3. 출하 상태는 위 C1/C2/M7 재확인으로 불활성이 입증됐다(Wired:false 양쪽, protectionofficial 비시험 importer 0,
   internal/protection 사용은 gateway.go:292 의 NewPairedReadinessAdapter 뿐).
4. Manager 검증 배터리가 종결 편집 자체를 아카이브 전에 따로 검증한다.
비례 원칙(2026-09-27 사용자 지시): 이미 독립 리뷰를 통과 상태로 가진 무변경·비활성 코드에 새 다중 보이스 리뷰를 돌리는 것은 낭비다.

### 완료 게이트

- gate PASS 4fd400ea, gate3.log, 11/11 (격리 worktree `TossOS-worktrees/a071-gate`, 2026-09-27; 로그 `/tmp/claude-1000/a071-lot/gate3.log`. gate2 는 같은 커밋에서 6단계 codegraph status 프로브 15s 타임아웃 — 일시적).
