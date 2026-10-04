# a112 0.5 voice1 (testing) 표적 재검 — P1#1 · P1#2 SubagentHandback 원문 (agent a2f93a2390322cb6b, 좌표 2497427e, 저장 2026-10-05)

안전 규칙 위반 없음(`~/.codex` 와 `_codex*/` 는 열지 않음, 저장소 편집 0, 실험은 2497427e `git archive` 사본에서만 함, 사본은 끝나고 삭제함). 새 사냥은 하지 않았고 P0 도 보지 못함.

**판정: P1#1 반영 · P1#2 반영**

## P1#1 — 부분 ON 차등이 빈 표본이던 문제 → 반영
- 고친 자리 확인:
  - `internal/app/engine/a112_shadow_cycle_test.go:96` `kr.snapshot.ProposalSetDigest = strategyProposalSetDigest(kr.entries)`
  - `:283` `if placed := without.trace(t).placed; placed < 1 {`
- 재실행 1 — 고친 픽스처의 궤적: 리뷰 때와 같이 `:286` 에 t.Logf 를 넣고 `go test -tags tossos_testseams -count=2 -v -run TestTheShadowPinLeavesTheDispatchTraceUnchanged ./internal/app/engine` 실행.
  - 결과(2회 모두 같음): `partial ON trace with={observed:1 placed:1 leases:1 laneLatches:0 latchRecoveries:0} without={…같음}`. undeclared 도 placed:1. → ok
  - 리뷰 때(4cbcfb36)는 같은 자리가 전부 0 이었으므로 0 → 1 로 바뀐 것을 확인함.
- 재실행 2 — 전제가 빈 표본을 막는지: 사본에서 `:96` 한 줄만 지우고(픽스처 결함을 되살림) 같은 시험 실행.
  - 결과: `--- FAIL: …/partial_ON` 이 `a112_shadow_cycle_test.go:283: precondition: the baseline dispatched 0 orders — the differential would compare two empty traces` 로 실패함. 빈 표본이 이제 통과가 아니라 실패로 드러남.
- 처분 표 `disposition.md:14`(시험#1, 「편집 전 placed 0 → 1, 두 하위시험 PASS」)와 일치함.

## P1#2 — 단계 폐포 핀이 blacklist 이던 문제 → 반영
- 고친 자리 확인:
  - `a112_shadow_structure_test.go:695` `allowedWorkers := map[string]bool{"Lane.Key": true, "Lane.ShadowEligible": true, "Lane.ShadowOutcomeOver": true}`
  - 하한(셋 다 닿아야 함)은 `:701`, 주기 함수 양성 대조는 `:713`.
  - 새 시험 `a112_shadow_cycle_test.go:232` `TestASuccessfulShadowStepLeavesEveryLaneStatusUnchanged` — 8 레인 `Status()` 를 전후 비교하고, 전제로 WOULD_EMIT ≥ 1 을 둠.
- 대조군(변이 없음): 세 시험 모두 PASS.
- 변이 시험: 로트 명령 6개(SET_731S_TESTS 를 옮긴 스크립트)를 그대로 돌림.

| 변이 | 넣은 자리 | 결과 | 실패한 시험 |
|---|---|---|---|
| A01 `_ = lane.Offer()` | `\t\tkey := lane.Key()` 앞 (`strategy_lane_shadow.go:190`) | **CAUGHT** | TestASuccessfulShadowStepLeavesEveryLaneStatusUnchanged, TestTheShadowStepClosureNeverRefreshesCoordinatesDispatchesOrWrites |
| A02 `_, _ = lane.Fail("shadow", false)` | 같은 자리 | **CAUGHT** | 위 두 시험 + TestAFailingShadowStepChangesNothingButItsOwnObservation/deadline |
| 원래 변이 `_ = lane.Offer()` | 리뷰 때 자리, 관측 대입 뒤(`:195` 뒤) | **CAUGHT** | A01 과 같은 두 시험 |

- 세 변이 모두 행동 시험과 구조 핀이 함께 잡음. 리뷰 때는 원래 변이가 6/6 ok 로 생존했고, 이제는 잡힘.
- 저자 원장 `lot-0.5-response/mutation-0.5-R-run1.tsv` 의 A01 행(2 failing, 같은 이름)과 A02 행(3 failing, 같은 이름)은 내 재실행과 실패 시험까지 일치함.
  - 참고: 원장 TREE 줄은 「4cbcfb36 + own 파일들」 이라 2497427e 자체가 아님. 내 재실행은 2497427e 에서 했고 결과가 같음.
- 처분 표 `disposition.md:15` 와 일치함.

## 저장소 상태
- HEAD: 시작과 끝 모두 `2497427e00149bac92195cb38261daa225cfdcc1` 으로 같음.
- `status --short` 는 같지 않음.
  - 시작: `?? .reticle-setup-crash.log`, `?? w4.log`
  - 끝: 위 둘 + `?? openspec/changes/a112-run-four-strategy-families-independently/analysis/harness/gate_suites.sh` + `?? openspec/changes/a112-run-four-strategy-families-independently/analysis/measurements/gate-8.1-8.3-2026-10-05/`
- 늘어난 두 항목은 내가 만든 것이 아님. 나는 저장소에 쓰지 않았고 scratchpad 의 사본에만 썼음. 이름으로 보아 병행 세션의 게이트 작업 산출물로 보이며, 출처는 직접 확인하지 않았음.
