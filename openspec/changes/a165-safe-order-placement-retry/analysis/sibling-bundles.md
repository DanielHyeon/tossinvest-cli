# check_analysis 형제 함수 기록 (2026-10-29)

## 판정

base는 `34e12967ed7d`로 유지함. base 재고정과 `--record-landing`은 사람 결정으로 금지됨. a165 자신의 required bundle 5개는 기존 change 기록을 유지함. 형제 30개는 AST의 `source_sha256`과 `revision`을 대조하여 archive bundle을 그대로 복사함. 새 bundle 생성은 0개임.

- a170 archive에서 28개를 복사함. 이 중 `revision: current` 23개와 `revision: base` 5개임.
- a164 archive에서 `revision: current` 2개를 복사함.
- `revision: base` bundle은 현재 소스와 다르거나 현재 소스에서 함수가 사라진 경우에도 base SHA 증거가 필요하므로 재생성하지 않음.
- 최신 확인 명령: `python3 tools/logic-map/check_analysis.py --change a165-safe-order-placement-retry`
- 확인 HEAD: `6a7c6ffe7fff`. 결과: base부터 working tree까지 147 commit, required 35개, `evidence complete or diff-proven exempt` 및 종료 코드 0.

## 형제 함수 목록

도구 원문: `/tmp/a165-check-analysis-siblings.out`. 형제 함수는 30개이며, a165 자체 함수 5개와 합쳐 required 35개임. 아래 30개 모두 source SHA 대조 archive bundle을 복사해 충족함.

1. `internal/campaignstateevidence/authority_guard_test.go:allowedImport`
2. `internal/campaignstateevidence/authority_guard_test.go:TestGuardRejectsRuntimeCaptureUse`
3. `internal/console/chrome.go:Console.chromeFor`
4. `internal/console/chrome.go:holdingsSnapshot.freshness`
5. `internal/console/chrome.go:ordersSnapshot.freshness`
6. `internal/console/chrome.go:signalsView.freshness`
7. `internal/console/display_primitives_test.go:TestTheHeadingStepsAreDistinguishable`
8. `internal/console/display_primitives_test.go:declaredFontSize`
9. `internal/console/display_primitives_test.go:declaredBodyFontSize`
10. `internal/console/display_primitives_test.go:TestTheSummaryCostsNoBrokerCall`
11. `internal/console/optimization.go:Console.handleOptimization`
12. `internal/console/optimization_candidate_filters_test.go:TestCandidateFilterCardsContainExactMarketMetricMatrix`
13. `internal/console/optimization_candidate_filters_test.go:TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly`
14. `internal/console/optimization_ui_contract_test.go:TestOptimizationDOMHasNoArbitraryInputAndSubmitsOnlyOwnerOptions`
15. `internal/console/optimization_ui_contract_test.go:TestOptimizationUsesOnePresetPreviewFlowWithoutClientDraft`
16. `internal/console/optimization_view.go:optimizationCategoryViews`
17. `internal/console/pages.go:Console.handleDashboard`
18. `internal/console/pages.go:Console.handleReport`
19. `internal/console/performance_history.go:Console.handlePerformanceHistory`
20. `internal/console/performance_history_test.go:fakePerformanceReader.Dashboard`
21. `internal/console/settings_card_test.go:TestEveryCardShowsWhatChangesAndWhen`
22. `internal/console/settings_card_test.go:TestTheCurrencyConsequenceIsRenderedAndNotFolded`
23. `internal/console/settings_limits_test.go:TestThePresetControlsAskForNoTyping`
24. `internal/console/settings_limits_test.go:limitCurrent`
25. `internal/console/trading_views_test.go:TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA`
26. `internal/official/stock_reads.go:adaptStocks`
27. `internal/verifylive/record_test.go:TestRecordCarriesNoUnmaskedAccountNumber`
28. `tools/a112-mb-us-source/identity.go:verifyPrescribedBinary`
29. `tools/a112-mb-us-source/main_test.go:TestVerifyPrescribedBinaryRejectsExecutableHashMismatch`
30. `tools/a112-mb-us-source/main_test.go:fakeBuildRunner.rebuild`

## 현재 change의 경계

- a165가 새로 편집한 `Gateway.withCancelRefusal`은 자체 FLM/BTM/AST를 이 change에 기록함.
- 형제 source·문서·base는 수정하지 않음.
