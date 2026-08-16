# Branch Test Map: `TestReleaseErrorsAreMappedToOperatorLanguage`

이 함수 자체가 테스트이므로 「그 분기를 무엇이 재는가」는 **뮤테이션**이 답한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 다섯 sentinel 전부를 돈다 | 자기 자신(표 반복) | no | yes |
| B2 | 상태 코드 매핑 | 자기 자신 | no | yes |
| B3 | 화면 문구 매핑(미배선 행은 **상수**로) | 뮤테이션 M20(문구 한 곳만 되돌리기) | no (a079 부터 초록) | yes |
