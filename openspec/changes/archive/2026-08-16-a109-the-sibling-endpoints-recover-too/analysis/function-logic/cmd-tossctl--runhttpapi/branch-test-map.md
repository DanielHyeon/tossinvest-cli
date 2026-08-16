# Branch Test Map: `runHTTPAPI`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | ctx 가 nil 인 호출 | 기존 httpapi 명령 테스트 | no | yes |
| B2 | boundary 해석 실패는 fatal | 기존 | no | yes |
| B3 | 직접 TLS 경로 | 기존 TLS 테스트 | no | yes |
| B4 | 인증서 로드 실패는 fatal | 기존 | no | yes |
| B5 | 자원 열기 실패는 fatal | 기존 | no | yes |
| B6 | 스냅샷 캐시 실패는 fatal | 기존 | no | yes |
| B7 | stream 생성 실패는 fatal | 기존 | no | yes |
| B8 | mutation route 실패는 fatal | 기존 `httpapi_mutation_test.go` | no | yes |
| B9 | router 생성 실패는 fatal | 기존 | no | yes |
| B10 | server 생성 실패는 fatal | 기존 | no | yes |
| B11 | 직접 TLS 의 TLSConfig | 기존 | no | yes |
| B12 | 직접 TLS 의 serve | 기존 | no | yes |
| B13 | serve 오류 vs ctx 종료 | 기존 | no | yes |
| B14 | `ErrServerClosed` 는 정상 종료 | 기존 | no | yes |
| B15 | Shutdown 실패 보고 | 기존 | no | yes |
| B16 | 종료 후 serve 오류 보고 | 기존 | no | yes |

이 함수의 a109 §2-fix 편집은 **인자 하나**이고 분기를 만들지 않는다. 그 인자가 실제로
쓰이는지는 호출 대상 쪽에서 잰다(`TestTheReattachWakeSurvivesABrokenAggregate`,
뮤테이션 M31 CAUGHT). 여기서 새 분기 테스트를 만들면 그것은 같은 사실의 두 번째 사본이다.
