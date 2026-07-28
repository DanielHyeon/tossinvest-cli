# Branch Test Map: `Collect`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 시장 없는 스캔 | **커버 없음** | no | no |
| B2 | instant 없는 스캔 | **커버 없음** | no | no |
| B3 | 패널 id 검사 | `TestTwoSourcesCannotShareAnID` | — (기존 동작) | yes |
| B4 | id 충돌은 읽기 전에 거부 | 동상 | — (기존 동작) | yes |
| B5 | heard 구성 | `TestASourceThatWasNotAskedDoesNotVouchForTheCandidatesItRaised` | — (기존 동작) | yes |
| B6 | not-asked도 heard | `TestASourceTheSchedulePassedOverIsNotASourceThatIsGone` | — (기존 동작) | yes |
| B7 | 패널과 not-asked에 동시에 있는 소스 | **커버 없음** | no | no |
| B8 | 패널 전체 순회 | `scan_test.go` 전반 | — (기존 동작) | yes |
| B9 | 429·일반 오류 | `TestARankingFailureIsALossAndNotADegradedFallback` · `TestEverySourceFailingIsAnError` | — (기존 동작) | yes |
| B10 | mis-wired fallback | `TestARankingCannotClaimToHaveFallenBack` | — (기존 동작) | yes |
| B11 | 요청·도착이 결과에 실린다 | `TestATruncatedReadingReachesTheVerdictAsTruncated` · `TestAWholeReadingOfTheSameLengthIsMeasured` · `TestTheScanReportSaysWhatEachSourceAskedForAndWhatArrived` | yes (F3: 이 배선이 없으면 리포트가 빈 블록) | yes |
| B12 | 빈 읽기는 커버리지가 아니다 | `TestAnEmptyReadingIsNotEvidenceOfAbsence` | — (기존 동작) | yes |
| B13 | 행마다 `RankRequested`가 관측으로 복사된다 | `TestATruncatedReadingReachesTheVerdictAsTruncated`(F3의 mutation 대상) | yes (0으로 바꾸면 실패) | yes |
| B14 | 심볼이 빈 행 | **커버 없음** | no | no |
| B15 | 첫 가격 행 고정 | `TestTwoSourcesRaisingOneSymbolMakeOneCandidate` | — (기존 동작) | yes |
| B16 | 첫 순위 행 고정 | `TestASessionStartDoesNotStampThePanelAsSeenLate` | — (기존 동작) | yes |
| B17 | 아무 소스도 답하지 않음 | `TestEverySourceFailingIsAnError` · `TestAPanelWithNoSourcesInItIsStillAnError` | — (기존 동작) | yes |
| B18 | mis-wire가 '무응답'을 이긴다 | `TestARankingCannotClaimToHaveFallenBack` | — (기존 동작) | yes |
| B19 | 관측 기록 실패 | **커버 없음** | no | no |
| B20 | 심볼 정렬 | `scan_test.go` 전반 | — (기존 동작) | yes |
| B21 | 승격 pass | `scan_test.go` 전반 | — (기존 동작) | yes |
| B22 | 한 심볼의 승격 거부 | `TestOneRejectedSymbolDoesNotAbortTheMarket` | — (기존 동작) | yes |
| B23 | NoteSources 실패 | **커버 없음** | no | no |
| B24 | recordFirsts 실패 | **커버 없음** | no | no |
| B25 | coolAbsent 실패 | **커버 없음** | no | no |
| B26 | 완전한 결과 + mis-wire 오류 | **커버 없음** | no | no |

**정직한 커버리지 기록**. 이 change가 만든 분기(B11)와 그것이 지나는 본문(B13)은 커버된다.
커버되지 않는 것은 전부 **기존 분기**이고, 목록으로 적어 둔다:

- B1·B2 — `Collect`의 인자 검증(빈 시장, zero instant)을 넘기는 테스트가 없다.
- B7 — 패널과 not-asked에 같은 소스가 있는 모순.
- B14 — 빈 심볼 행 skip.
- B19·B23·B24·B25 — 저장 실패 전파 네 곳.
- B26 — mis-wire와 성공 읽기가 함께 있는 fixture가 없다(B18은 전 소스 실패 쪽만 잡는다).

그리고 B11의 **한계**가 테스트되지 않는다 — 행이 0개인 읽기가 `Readings`에 나타나지
않는다는 사실은 issues.md I9에 적혀 있지만 그것을 고정하는 테스트는 없다.
