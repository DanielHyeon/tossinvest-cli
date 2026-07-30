# Branch Test Map: `positionRow.Label`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 미편입·미지정 행 | `TestAnUnmanagedRowsLabelFollowsItsCheckbox` | no | yes |
| B2 | 원장 미판독 + 제외 등재 | `TestAnUnreadableJournalStaysUnknownEvenWhenExcluded` | no | yes |
| B3 | 제외 등재 행 | `TestAnExcludedRowIsLabelledAndOffersRelease` | yes | yes |
| B4 | 편입 지정 행 | `TestAnUnmanagedRowsLabelFollowsItsCheckbox` | no | yes |
| B5 | 어느 목록에도 없는 행 | `TestAnUnmanagedRowsLabelFollowsItsCheckbox` | no | yes |
| B6 | 관리 종료 포지션 | `TestThePositionsScreenJoinsBrokerAndJournal` | no | yes |
| B7 | exit 라인 있는 관리 포지션 | `TestThePositionsScreenJoinsBrokerAndJournal` | no | yes |
| B8 | 자격은 있으나 exit 미개설 | `TestThePositionsScreenJoinsBrokerAndJournal` | no | yes |
| — | 제외+편입 동시 등재는 제외가 이긴다 | `TestExclusionBeatsDesignationInTheLabel` | yes | yes |
