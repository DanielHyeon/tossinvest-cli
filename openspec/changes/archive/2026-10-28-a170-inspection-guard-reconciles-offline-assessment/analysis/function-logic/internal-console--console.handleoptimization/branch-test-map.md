# Branch Test Map: `Console.handleOptimization`

- Source: `internal/console/optimization.go` (66-127) — 편집 뒤. AST branches 14.
- 측정: 시험별 `go test ./internal/console/ -run '^<Test>$' -count=1 -coverprofile` 로 분기 본문 블록 count>0 확인(2026-10-25, a177 close). a177 편집(85행, 직선)은 `TestPerformanceWindowIsSingleSourceAcrossScreens` 가 고정.

| Branch | Anchor | Scenario | Test | a177 |
|---|---|---|---|---|
| B1 | `67:2` | POST 등 → 405 | 시험 없음(측정상 본문 미실행) — a177 무관 기존 공백 | 무변경 |
| B2 | `74:2` | category 없음 → 개요 | `TestOptimizationStaleEvidenceIsExplicitAndFailClosed` · `TestTheOldPathsStillRender` | 무변경 |
| B3 | `81:2` | 모르는 category → 경고 | `TestOptimizationUnknownCategoryFallsBackToOverviewWithoutMutation` | 무변경 |
| B4 | `87:2` | legacy seam 배선 | `TestOptimizationShowsExactlyThreePoliciesAndExternalBehavior` | 무변경 |
| B5 | `89:3` | legacy Load 오류 | 시험 없음(측정상 본문 미실행) — a177 무관 기존 공백 | 무변경 |
| B6 | `91:10` | legacy Load 성공 | `TestOptimizationShowsExactlyThreePoliciesAndExternalBehavior` | 무변경 |
| B7 | `95:2` | lifecycle 배선 | `TestOptimizationLoadingModelIsServerBlockingAndReadFailureIsFailClosed` · `TestOptimizationStaleEvidenceIsExplicitAndFailClosed` | 무변경 |
| B8 | `97:3` | lifecycle Read 오류 | `TestOptimizationLoadingModelIsServerBlockingAndReadFailureIsFailClosed` | 무변경 |
| B9 | `99:10` | lifecycle Read 성공 | `TestOptimizationStaleEvidenceIsExplicitAndFailClosed` · `TestOptimizationUnknownCategoryFallsBackToOverviewWithoutMutation` | 무변경 |
| B10 | `108:2` | 공통 정책 순회 | `TestOptimizationShowsExactlyThreePoliciesAndExternalBehavior` · `TestPerformanceWindowIsSingleSourceAcrossScreens` | 무변경 |
| B11 | `110:3` | Desired 빈 값 → legacy 대체 | `TestOptimizationShowsExactlyThreePoliciesAndExternalBehavior` | 무변경 |
| B12 | `117:2` | protection seam 배선 | `TestExitProtectionCurrentRowUsesOnlyOpaqueActionAndCheckbox` | 무변경 |
| B13 | `120:3` | protection List 오류 | 시험 없음(측정상 본문 미실행) — a177 무관 기존 공백 | 무변경 |
| B14 | `122:10` | protection List 성공 | `TestExitProtectionCurrentRowUsesOnlyOpaqueActionAndCheckbox` | 무변경 |
