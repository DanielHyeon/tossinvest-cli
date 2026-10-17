# Review: a067-add-kr-us-continuation-lanes

Date: 2026-08-04 · scope: pure dormant continuation contracts

## Concurrent KR/US outcome

- `kr_short_flow_continuation_v1` and `us_short_participation_continuation_v1` are returned by one registry unit under `a067-kr-us-continuation-v1`.
- Both descriptors remain desired/effective `OFF`; neither evaluator waits for peer readiness, session state, or activation.
- KR uses signed notional flow ppm. US uses share participation and signed price-change ppm. Both use strict market schemas and checked integer arithmetic.
- The pure package has no broker, journal, exit, gateway, operating-toggle, or registry authority dependency.

## RED to GREEN record

The first focused run failed at compile time because the market, frozen-FX, campaign-plan, cap, and evaluation contracts did not exist. The RED test set already contained both KR and US schema boundaries, one-market failure isolation, 8:4:2 allocation, cap/risk admission, fill/cancel replay, invalidation/common-exit separation, property checks, fuzz targets, and static dependency closure.

After implementation, both markets became GREEN in the same package and release. No existing registry, runtime, journal, engine, or dispatch function was edited, so task 1.1's existing-function Logic Map gate is not applicable; the new-function branch maps are executable tests in `internal/continuationlane`.

## Safety review

- Allocation is immutable: `floor(Q*8/14)`, `floor(Q*4/14)`, final remainder; partial fill and cancel never move unused quantity upward.
- Proposed quantity is bounded by both planned remaining and the frozen a066 `q_final`. The private cap seal binds the exact immutable campaign-plan digest, exact plan policy digest, proposed reservation quantity, bucket-set provenance, validity window, and frozen FX snapshot; policy mismatch and same-market/same-policy cross-plan replay are refused.
- KR/US threshold config values are created through sealed constructors. Frozen FX and a066 cap attestations use package-private validated constructors plus private content seals, so external caller literals and post-construction mutation cannot pass evaluation. Same-currency plans reject every FX object and zero-quantity plans are invalid.
- Admission checks `filled + held + proposed <= immutable budget` with checked arithmetic.
- Positive-fill risk is the greater of transferred conservative reservation and exact stop-distance/fee risk. Cross-currency US risk uses the identical official frozen quote, quote-to-account direction, haircut, and final minor-unit ceiling.
- Actual overage or unknown risk applies the risk event and latches all later exposure-raising legs. Duplicate fills and cancels are idempotent; cancel releases held risk only.
- Missing FillID/CancelID uses a length-framed full non-ID preimage digest. The exact raw retry is idempotent, a distinct raw preimage remains separate evidence, and unidentified cancels cannot release held risk.
- Fill and cancel apply paths are bound to the exact plan/risk state, campaign, leg, order, source and RFC3339Nano observation time. Campaign/order/source identities are capped at 256 bytes. Foreign campaign, leg 99, oversized identity and empty/invalid provenance retain prior held/filled accounting, store non-applied evidence and latch unknown risk; only the identical scoped raw retry is a duplicate.
- Quantity-zero fills are retained as non-applied evidence, leave held/filled accounting unchanged and latch unknown risk. The same invariant holds for zero observations that reuse an existing positive FillID.
- Fill accounting computes held and filled successors before assigning either. Transferred-over-held, corrupt held/filled and `filled+risk` overflow preserve both prior values while retaining fill evidence and latching unknown risk; integer text is bounded to 256-bit/78-digit inputs before `big.Int` parsing.
- Stop provenance is created through a sealed constructor binding price, source, policy, version, digest, observed time and fresh-until. Evaluation requires exact `observed_at <= evaluated_at <= fresh_until`; stale but well-formed and post-seal-mutated candidates fail closed.
- Strict KR/US JSON decoding rejects duplicate object keys at every nesting level as well as unknown fields and multiple top-level values.
- Structural/exit invalidation without a non-empty typed code is a typed refusal, and both cancelled and expired legs are terminal.
- Invalidation can suppress an add but never creates an exit decision. `CommonExitIndependent` stays true and the existing common-exit package passes its race test unchanged.
- Registration is descriptive and dormant. There is no LIVE hostname, approval, toggle, broker, journal-writer, or exit-authority mutation.

## Verification

| Check | Result |
|---|---|
| `go test ./internal/continuationlane` | pass |
| `go test -race ./internal/continuationlane ./internal/exitpolicy` | pass |
| allocation/admission property test, 20 repetitions (100,000 generated cases) | pass |
| allocation fuzz after final hardening, 2 seconds (618,747 executions) | pass |
| fill retry/zero-quantity fuzz after final hardening, 2 seconds (276,492 executions) | pass |
| `go vet ./internal/continuationlane` | pass |
| evidence/campaign/risk/strategy/dispatch/exit/scheduler regressions | pass |
| `openspec validate a067-add-kr-us-continuation-lanes --strict` | valid |
| focused statement coverage | 77.5% |

Full-repository `make sdd-sync`, `make sdd-check`, and `make gate CHANGE=a067-add-kr-us-continuation-lanes` remain an integration-stage task. This review does not claim activation or runtime wiring.

## 4.2 회귀 실측 (2026-09-25, HEAD d72bc401)

측정 위치: 저장소 밖 격리 워크트리 `/mnt/D/Axipient/workspace/TossOS-worktrees/archive-batch1`
(`git checkout --detach d72bc401`, 실행 전후 `git status --short` 0줄). 주 워크트리의 남의
미커밋 편집을 타지 않게 하기 위함. 측정 뒤 HEAD 는 `317b3f8d` 로 움직였지만
`d72bc401..317b3f8d` 의 변경은 `openspec/changes/a095-…/review.md` 한 파일뿐이라 Go 코드는 동일함.
모든 로그는 `/tmp/claude-1000/a067a068-42/` 아래.

대상 30 패키지(`pkgs.txt`): 레인 4(continuationlane·reversallane·breakoutlane·weeklyvaluelane),
evidence 2(strategyevidence·optimizationevidence), campaign 1(positioncampaign),
risk 3(risk·riskbucket·riskcalc), strategy 17(strategy~strategyworker, strategydispatch 포함),
exit 2(exitpolicy·exitquarantine), scheduler 1. 레인 패키지를 import 하는 소비자
(`go list` Imports/TestImports 로 잰 것: strategyflow·strategyproposal·strategyrouter·strategyworker)는
모두 이 30 안에 있음. 통합 소비자 `internal/app/engine` 은 따로 돌림.

| 명령 | rc | 시간 | 결과 | 로그 |
|---|---|---|---|---|
| `go test -count=1 <30 패키지>` | 0 | 61s | ok 28 · no test files 2(strategy·strategyaccount) · FAIL 0 | `01-test.log` |
| `go test -count=1 -tags tossos_testseams <30 패키지>` | 0 | 100s | ok 29 · no test files 1 · FAIL 0 | `02-test-seams.log` |
| `go test -count=1 -tags tossos_testseams ./internal/app/engine` | 0 | 96s | ok 1 · FAIL 0 | `03-app-engine.log` |
| `go test -race -count=1 -tags tossos_testseams <12 패키지>` (`race-pkgs.txt`: continuationlane·reversallane·exitpolicy·positioncampaign·riskbucket·strategyevidence·strategyworker·strategyflow·strategyproposal·strategyrouter·strategydispatch·scheduler) | 0 | 23s | ok 12 · FAIL 0 | `04-race.log` |
| `go vet <30 패키지> ./internal/app/engine` (무태그 / `-tags tossos_testseams`) | 0 / 0 | 5s / 4s | 출력 0줄 / 0줄 | `05-vet.log` · `05b-vet-seams.log` |
| `FuzzEightFourTwoConservesQuantity` 2s | 0 | 2.1s | 마지막 보고 execs 147,985 | `06-fuzz-*.log` |
| `FuzzRiskFillRetryIsIdempotent` 2s | 0 | 2.1s | 마지막 보고 execs 103,333 | 〃 |
| `FuzzTwoFourEightConservesQuantity` 2s | 0 | 2.1s | 마지막 보고 execs 136,927 | 〃 |
| `FuzzFillRetryIsIdempotent` 2s | 0 | 2.1s | 마지막 보고 execs 114,894 | 〃 |

실패 0: 31 패키지(30 + app/engine) 무태그·seams·race·vet·fuzz 전부 rc 0. 원인 귀속할 실패 없음.

### 라이브 무변이 확인

(a) base `c57915dd` 이후 두 change 디렉터리를 만진 커밋은 `38c616be`(문서·PM 만)·`75cc736f`
(continuationlane 만)·`4809d837`(reversallane 만)이고, openspec 밖에서 만진 코드는 레인 두
패키지뿐. 두 패키지는 base 에 없던 새 패키지(`git ls-tree` 0 파일)이고 이후 커밋 6개
(acd2d1a6·590a0bc0·a0a23c57·542ed067·f737e4f8·8022f578)가 더 만졌으므로 창 전체
`git diff c57915dd d72bc401 -- internal/continuationlane internal/reversallane`(39 파일,
+5,783 / −0, `08-lane-diff.patch`)를 셈. 변경 줄 중 `host|https?://|enabled|approve|toggle`
(대소문자 무시) 적중 17줄(`08-pattern-hits.txt`):
- `host`·URL·`approve`: **0**
- `toggle` 5: 전부 금지 의존 가드 시험 이름·금지 문자열 목록(dependency_test.go 4)과
  "runtime-toggle authority 없음" 주석 1(continuationlane/types.go)
- `enabled` 12: 평가 입력 구조체 필드 `EvaluationContext.Enabled` 선언 2, `if !context.Enabled`
  fail-closed 가드 2, 값 리터럴 `Enabled: true` 8(시험 픽스처·testseam 6, 순수 값 조립
  `production_proposal.go` 2 — 후속 커밋 acd2d1a6/8022f578 이 넣은 메모리 안 값이며 저장·토글 쓰기 아님)
- 운영 토글·승인·hostname 을 **쓰는** 변경: 0

(b) dormant/OFF 정적 가드 13개를 격리 워크트리에서 `-tags tossos_testseams -v` 로 따로 돌림
(`09-dormant-guards.log`): PASS 13 · FAIL 0 · "no tests to run" 0, 패키지 5개 rc 전부 0.
continuationlane 3(순수 의존 폐포·KR/US 동시 기본 OFF·OFF/실패 비전파), reversallane 3(같은 셋의
reversal 판 + invalidation 이 common exit 권한을 안 가짐), strategyworker 2(모든 생산 워커 dormant
탄생·import 폐포), strategyrouter 2(0 활성화·활성화 파일 없음 → 승격 0), app/engine 3(생산 활성화
로더가 매니페스트 0건·활성화 없으면 조정 불변·레인이 자기 제안에서도 dormant).

### a067 전용 — 부분 체결

`go test -count=1 -race -v -run '^TestPartialCancelDoesNotReallocateAndA066CapStillBinds$' ./internal/continuationlane`
rc 0, `--- PASS` 1(`07b-a067-partial.log`).

### 4.3 을 이번에 하지 않은 사유

게이트 정책 결정 대기. Manager 가 격리 워크트리(54004f44)에서 잰 것: 5단계
`check_analysis.py` 가 `revision: current` 번들 0 → base `c57915dd` 뒤 착지한 391 커밋이 바꾼
남의 기존 함수 338개의 Function Logic Map 을 요구함(`/tmp/claude-1000/ca-a067-add-kr-us-continuation-lanes.log`).
이 change 가 편집한 기존 함수는 0 이므로 레인 코드의 결함이 아니라 창 정의의 문제임. 두 레인은
이 측정 시점에도 기본 OFF(위 (b)).

## 종결 시퀀스 (2026-09-27, a122 선례 · Manager 2단계 승인)

### Function Logic Map: not-applicable

**사유.** 이 change 의 Go 작업은 새 패키지 `internal/continuationlane/` 의 새 파일 11개(75cc736f)이고,
그 커밋이 수정한 **기존** 함수는 0 이다 — `check_analysis.changed_existing_functions(75cc736f^, 75cc736f)`
실측 0(2026-09-27, HEAD 4df72ed8). 새 함수는 Function Logic Map 대상이 아니므로 번들 0 이 맞다.

### base 재고정 영수증 (0de27fab)

- 잰 순간: HEAD 4df72ed8, 모집단 = `git log --full-history -- openspec/changes/a067-…` 커밋 4.
  - 38c616be(2026-08-03, 명세) go 0
  - 75cc736f(2026-08-04, 레인 코어) go 11 — 전부 새 파일, 기존 함수 수정 0
  - 448dfeb1(2026-08-04, 통합 병합) combined diff go 9 = `cmd/tossctl/console.go` · `internal/app/engine/{engine,interlock}.go`
    · `internal/config/engine.go` · `internal/console/*` 5 — a074~a079 충돌 해소, continuationlane 0
  - 51345e12(2026-09-25, 4.2 회귀 기록) go 0
- 옛 base c57915dd 에서 5단계 rc 1 · required 340 · 창에 착지 커밋 451 — 전부 다른 change 의 기존 함수
  (`/tmp/claude-1000/a122-lot/audit/ca-a067-add-kr-us-continuation-lanes.log`).
- 재고정 c57915dd → 4df72ed8, `base-commit.txt` 만 커밋(0de27fab). 재고정 뒤 check_analysis:
  required 0, 남은 사유는 이 마커뿐.

### 4.3 두 레인 OFF 기록 (코드 인용만, LIVE·운영 접근 0)

- `internal/continuationlane/registry.go:20-21` — KR·US 서술자 둘 다 `DesiredState: StateOff, EffectiveState: StateOff`.
- `registry.go:33-35` — `ValidateRegistry` 는 OFF 가 아닌 서술자를 거절한다.
- `internal/continuationlane/dependency_test.go:12-34` — 패키지 비시험 파일이 `/broker` · `/journal` · `/exit` ·
  `/gateway` · `/operating` · `/toggle` · `/registry` 를 import 하면 실패한다.
- **정정:** "runtime 배선 0" 은 이 패키지 안에서만 참이다. a112 가 뒤에 두 레인을 생산 경로에 DORMANT 로 세웠다 —
  `internal/strategyworker/worker.go:136-138` 이 레인 ID 를 워커 키로 쓰고, `worker.go:193-195` 의 `Run` 은
  `Effective(activation) != StateOn` 이면 `OutcomeDormant` 로 끝난다. `worker.go:182-183`: activation 은
  `strategyrouter.FamilyActivation`(필드 전부 비공개)라 패키지 밖에서 영값만 만들 수 있고, 영값은 아무것도 켜지 않는다.
  그래서 두 레인은 여전히 OFF 다.
