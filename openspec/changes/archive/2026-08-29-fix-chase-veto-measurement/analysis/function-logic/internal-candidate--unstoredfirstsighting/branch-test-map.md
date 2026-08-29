# Branch Test Map: `unstoredFirstSighting`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 관측을 두 번 훑는다 | `TestTheUnqualifiedReasonIsNotOneOfTheStoresOwnGaps` | — (기존 동작) | yes |
| B2 | 순위 없는 행은 후보가 아니다 | 동상('nothing ranked' — 가격만 있는 관측) | — (기존 동작) | yes |
| B3 | 가장 이른 순위 행 | 동상 | — (기존 동작) | yes |
| B4 | 창 밖 행은 이 생명의 최초가 아니다 | 동상('a ranked row with no stored column') | — (기존 동작) | yes |
| B5 | 창 안 가장 이른 행 | 동상 | — (기존 동작) | yes |
| B6 | 네 사유가 서로 다르다 | `TestTheUnqualifiedReasonIsNotOneOfTheStoresOwnGaps`(중복 사유를 실패로 만든다) | — (기존 동작) | yes |
| B7 | 관측이 아예 없다 | 동상 | — (기존 동작) | yes |
| B8 | 순위를 싣는 소스가 없다 | 동상 | — (기존 동작) | yes |
| B9 | 창 안에 행이 있는데 칼럼이 없다 | `TestASessionStartDoesNotStampThePanelAsSeenLate`(보류된 첫 tick이 정확히 이 상태) | yes (보류 도입으로 이 경로가 세션 첫 tick의 정상 상태가 됐다) | yes |
| (신규 1줄) | 거부된 sighting도 절단 사실을 싣는다 | **직접 커버 없음** | no | no |

**정직한 커버리지 기록**: 새로 더한 `out.Truncation` 대입을 **직접** 확인하는 테스트는
없다. `MeasureFirstSighting`의 저장-위치 경로에서는 `TestATruncatedReadingsPositionIsNotAPercentile`이
`got.Truncation.Yes()`를 단언하지만, `unstoredFirstSighting` 경로의 같은 필드를 읽는
단언은 어디에도 없다. 이 필드는 이 경로에서 아무것도 결정하지 않으므로(넷 다 이미
미측정) 위험은 표시 누락에 그친다.
