# Branch Test Map: `Panel`

> **정정 2026-07-29 (§8.1·8.2, 독립 리뷰 F1).** B3의 "Test" 칸이 초판에서
> `TestAnUnknownRankingTypeIsRefused` · `TestEveryPanelSourceHasItsOwnID`였다. **둘 다 이
> 분기를 덮지 않는다** — 전자는 생성자를 직접 부르므로 `Panel` 경로에 대해 아무 말도 하지
> 않고, 후자는 넘겨받은 패널의 id 중복·비어있음만 본다. 오류로 조용히 빠진 원천은 그 둘
> 어느 것도 발동시키지 않는다(재현: 매핑 없는 enum 값을 리터럴에 넣고 패키지 전체 GREEN).
> 이 칸이 거짓이면 Branch Test Map 절차 자체가 형식이 되므로, §8.1에서 실제로 덮는 가드를
> 만들고 RED를 관측한 뒤 칸을 고쳤다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `official != nil`이면 랭킹 원천이 만들어지고, nil이면 WTS만 남은 패널을 조용히 허용하지 않는다 | `TestThePanelHandsItsClockToEverySourceItBuilds` (non-nil) · `TestThePanelWithoutAnOfficialClientIsNotSilentlyWTSOnly` (nil) | no — 이 change가 바꾼 조건이 아니다 | yes |
| B2 | 순회 대상 집합에 `TOP_GAINERS`가 없고(KR·US), 그 **값**이 스냅샷의 금지 조합에 들어가지 않으며, 만들어진 원천의 id가 서로 다르고 clock 순회 대상 집합과 정확히 일치한다 | `TestNoMarketPanelBuildsTheGainersRanking` · `TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused` · `TestEveryPanelSourceHasItsOwnID` · `TestThePanelHandsItsClockToEverySourceItBuilds` | **yes** — 1.1 적용 전 앞의 둘이 실패(`the KR/US panel builds official_rankings_top_gainers`, `Panel builds TOP_GAINERS and the reader asks for duration "realtime"`). 변이(리터럴 복원)에서 셋이 다시 RED | yes |
| B2′ | 이 리터럴이 **AST가 읽는 그 리터럴**이다 — 목록이 `Panel` 밖으로 옮겨지면 위 스냅샷 가드가 초록인 채로 눈이 먼다(독립 리뷰 F4) | `TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds` | **yes** — 목록을 패키지 수준 `var`로 올리고 `Panel`에 무관한 `[]string` 리터럴을 남긴 변이에서 RED. **같은 변이에서 `TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused`는 PASS였다** — 그것이 F4가 말한 눈멂이고, 이 행이 그것을 닫는다 | yes |
| B3 | `OfficialRanking`이 오류를 내면 그 원천은 패널에서 **조용히** 빠진다 — 그리고 그 상태가 실패로 보고된다 | `TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds` (`snapshot_drift_test.go`. 리터럴이 지목한 타입 집합 == `Panel`이 실제로 만든 원천의 타입 집합) | **yes** — `rankingSourceID`에 항목이 없는 실제 enum 값 `TOSS_SECURITIES_TRADING_AMOUNT`를 리터럴에 넣은 변이에서 RED. 같은 변이가 가드 도입 **전에는** 패키지 53건 전부 통과였다 | yes |
| B4 | WTS 인기 목록은 KR 패널에만 들어가고, 원천 자신도 다른 시장을 거부한다 | `TestTheUSPanelDoesNotIncludeTheKoreanPopularityRanking` · `TestThePopularityRankingRefusesAMarketItCannotSee` | no | yes |
