# Branch Test Map: `TestTheCommandsRefuseWhenNoEngineIsRunning`

이 함수 자체가 테스트이므로 「그 분기를 무엇이 재는가」는 **뮤테이션**이 답한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 임시 디렉터리 생성 실패 | 자기 자신(`t.Fatal`) | no | yes |
| B2 | 0700 강제 실패 | 자기 자신 | no | yes |
| B3 | 두 명령 모두 검사한다 | 자기 자신(반복) | no | yes |
| B4 | 엔진 없이 성공하면 실패 | 자기 자신 | no | yes |
| B5 | **거부의 정체가 sentinel 이다** | 뮤테이션 M27b (sentinel → 날것의 오류) | yes (M27b 로 FAIL 재현) | yes |
