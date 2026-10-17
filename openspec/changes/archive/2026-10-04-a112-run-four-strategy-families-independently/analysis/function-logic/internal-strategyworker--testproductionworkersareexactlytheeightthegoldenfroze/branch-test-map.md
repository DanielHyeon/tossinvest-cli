# Branch Test Map: `TestProductionWorkersAreExactlyTheEightTheGoldenFroze (시험)`

- Source SHA-256: `38997843a0a5276323cef0187e88945fc8368b1c0e245572007599a3718bd892`; AST branch locations are authoritative.
- Revision: **modified (a112 gate-prep, 2026-10-01).** 8.8.4 로트 A(f473d815): 공허한 desired/effective 절(영값 활성화 — `lookup` 이 첫 줄에서 OFF 를 돌려줘 상수-대-상수)을 빼고 Runtime 대조만 남겼다. 그 행동 단언은 태그 파일 `a112_golden_desired_effective_test.go` `TestTheGoldenOffIsEachWorkersDefaultAndOnlyItsSignedActivationFlipsIt` 로 옮겼다.
- 편집 전 번들: `analysis/measurements/gateprep-2026-10-04/pre-edit/internal-strategyworker--testproductionworkersareexactlytheeightthegoldenfroze/`. 변이 원장 `analysis/measurements/lot-8.8.4-A/mutation-8.8.4-A.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 83:2 — 골든 자기 모순(worker_count ≠ descriptors 수) → Fatal | 이 시험 자신(시험 코드) — `go test ./internal/strategyworker` 무태그 · 태그 GREEN(`lot-8.8.4-A/verify-8.8.4-A.log`) | no — 시험 코드 | yes |
| B2 | if at 87:2 — 생산 worker 수 ≠ 골든 → Fatal | 이 시험 자신(시험 코드) — `go test ./internal/strategyworker` 무태그 · 태그 GREEN(`lot-8.8.4-A/verify-8.8.4-A.log`) | no — 시험 코드 | yes |
| B3 | range at 92:2 — 골든 서술자 순회 | 이 시험 자신(시험 코드) — `go test ./internal/strategyworker` 무태그 · 태그 GREEN(`lot-8.8.4-A/verify-8.8.4-A.log`) | no — 시험 코드 | yes |
| B4 | if at 95:3 — 열쇠(시장 · 가족 · 레인 · 버전) 드리프트 → Error | 이 시험 자신(시험 코드) — `go test ./internal/strategyworker` 무태그 · 태그 GREEN(`lot-8.8.4-A/verify-8.8.4-A.log`) | no — 시험 코드 | yes |
| B5 | if at 101:3 — horizon 드리프트 → Error | 이 시험 자신(시험 코드) — `go test ./internal/strategyworker` 무태그 · 태그 GREEN(`lot-8.8.4-A/verify-8.8.4-A.log`) | no — 시험 코드 | yes |
| B6 | if at 108:3 — runtime 드리프트 → Error(로트 A — desired/effective 절 제거 뒤 남은 대조) | 이 시험 자신(시험 코드) — `go test ./internal/strategyworker` 무태그 · 태그 GREEN(`lot-8.8.4-A/verify-8.8.4-A.log`) | no — 시험 코드; 옮겨 간 desired/effective 행동 단언은 변이 A2 CAUGHT(`analysis/measurements/lot-8.8.4-A/mutation-8.8.4-A.tsv`) | yes |
