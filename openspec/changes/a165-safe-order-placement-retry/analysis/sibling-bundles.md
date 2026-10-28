# check_analysis 형제 함수 기록 (2026-10-29)

## 판정

base는 `34e12967ed7d`로 유지함. base 재고정과 `--record-landing`은 사람 결정으로 금지됨. 최초 기록(HEAD `27a012ce0d06`)은 a165 밖의 **139 commit**을 함께 보아 총 35개 함수를 요구했음. D5 후속 5개 a165 commit 뒤 최신 `check_analysis.py`는 **144 commit**, required 35개를 보고함. a165 자신의 required bundle은 현재 change에 모두 유지하고, archive 단계에서 Manager가 source SHA 대조 뒤 bundle을 복사하는 a170 방식으로 형제 bundle을 처리함.

## 형제 함수 목록

도구 원문: `/tmp/a165-check-analysis-postcommit-20261029.txt`. 원문이 표시한 missing sibling 함수는 30개이며, a165 자체 함수 5개와 합쳐 required 35개임.

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

- 최신 `check_analysis.py`는 a165 자체 누락 0이고, source SHA 기반 archive 복사 전에는 30개 sibling 누락 때문에 의도적으로 PASS가 될 수 없음.
- a165가 새로 편집한 `Gateway.withCancelRefusal`은 자체 FLM/BTM/AST를 이 change에 기록함.
- 형제 source·문서·base는 수정하지 않음.
