# Branch Test Map: `buildCandidatePanel`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Open API 경로 해석 실패 | **커버 없음** | no | no |
| B2 | 자격증명 읽기 실패 | **커버 없음** | no | no |
| B3 | 자격증명 없음 | **커버 없음** | no | no |
| B4 | WTS 세션 유무 | **커버 없음** | no | no |
| B5 | config 읽기 실패 시 WTS 없이 진행 | **커버 없음** | no | no |
| (파일 이동) | verdict를 읽는 파일이 `official.New`를 명명하지 않는다 | `TestOnlyTheListedFilesCanNameTheChaseVerdict` · `TestNoFileThatReadsTheVerdictCanAlsoPlaceAnOrder` · `TestTheOrderVerbDetectorSeesTheFormsThatUsedToWalkPast`(양성 대조군: `official.New`가 실제로 탐지된다) | yes (F5가 세 구멍을 실행으로 확인) | yes |

**정직한 커버리지 기록**: 이 함수의 다섯 분기 어느 것도 테스트가 없다 —
`rg 'buildCandidatePanel' cmd/tossctl/*_test.go`가 0건이다. 이 change 이전에도 그랬고,
이 change의 편집(파일 이동 + `clock.System()` 인자)은 분기를 하나도 만들지 않았다.
**커버된 것은 이 함수가 어디 있는가**이며 그것이 이 change가 바꾼 것이다: 세 정적 테스트가
`official.New`를 명명하는 파일과 chase verdict를 읽는 파일이 겹치지 않는지를 검사한다.
