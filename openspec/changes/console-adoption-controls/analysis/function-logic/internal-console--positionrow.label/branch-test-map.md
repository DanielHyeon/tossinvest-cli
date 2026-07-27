# Branch Test Map: `positionRow.Label`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | switch 진입(전 분기 공통) | 아래 전부 | yes | yes |
| B2 | 원장 미판독 행은 지정 여부와 무관하게 불명 | TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead | no | yes |
| B3 | 지정(체크)된 미편입 행 라벨 = 관리 편입 | TestAnUnmanagedRowsLabelFollowsItsCheckbox | yes | yes |
| B4 | 미지정 미편입 행 라벨 = 관리 외(미편입) | TestAnUnmanagedHoldingIsLabelledExactlyOnce · TestAnUnmanagedRowsLabelFollowsItsCheckbox(전반부) | yes | yes |
| B5 | exit 완결 = 관리 종료 | 직접 pin 없음(기존 분기·이번 변경 무접촉 — review.md P4 기록) | no | no |
| B6 | exit 보유 = 엔진 관리 | TestAnAdoptedHoldingRendersAsManagedWithItsBasis | no | yes |
| B7 | 자격 있으나 exit 미개설 = 엔진 관리(대기) | 직접 pin 없음(기존 분기·이번 변경 무접촉 — review.md P4 기록) | no | no |

RED 관측: `go test ./internal/console/ -run "TestTheStatusColumnHeaderSaysAdoption|TestAnUnmanagedRowsLabelFollowsItsCheckbox"`
— 구현 전 2건 실패(헤더 관리 편입 부재·지정 행 관리 외 라벨 잔존), 구현 후 전체
128건 GREEN.
