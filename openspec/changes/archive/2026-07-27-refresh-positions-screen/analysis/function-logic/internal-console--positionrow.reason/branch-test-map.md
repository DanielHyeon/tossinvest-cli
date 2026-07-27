# Branch Test Map: `positionRow.Reason`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | switch 진입(모든 행에서 평가) | 아래 전 분기 테스트가 공유 | yes | yes |
| B2 | `Unknown() \|\| !InJournal` — 행 문장 없음: 미판독은 공지 1회, 원장 부재 보유 2종목도 공지 1회 | `TestTheUnknownStateNoticeAppearsOncePerPage`(RED: 공지 3회 반복) + `TestTheJournalAbsenceNoticeAppearsOncePerPage`(RED: 공지 0회·구 문장 행 반복) | yes | yes |
| B3 | 자격 기록 없는 원장 포지션의 행별 사유 유지 | `TestAnUnmanagedHoldingIsLabelledExactlyOnce`("편입 기록(adoption)도 없는 포지션이다") | yes (신규 공지 문자열 부재로 실패) | yes |
| B4 | 자격 있음·exit 미개설 문장 유지 | `TestTheJournalAbsenceNoticeAppearsOncePerPage` 행별 사유 단언 + 기존 자격 표시 케이스 | — (기존 동작 유지 가드) | yes |
| B5 | exit 라인 존재 시 행 사유 없음(default) | `TestThePositionsScreenShowsTheExitLineOfAManagedPosition` | — (기존 동작 유지 가드) | yes |

RED 실행: `go test ./internal/console/ -run 'TestTheJournalAbsenceNoticeAppearsOncePerPage|TestTheUnknownStateNoticeAppearsOncePerPage|TestThePositionsScreenAsksTheBrowserToReloadAtTheCacheTTL|TestTheVerificationScreensKeepTheirTwoSecondReload|TestAnUnmanagedHoldingIsLabelledExactlyOnce'` — 4 failed / 1 passed (2026-07-27, 구현 전).
GREEN 실행: `go test ./internal/console/ -race -count=1` — 115 passed (2026-07-27, 구현 후).
