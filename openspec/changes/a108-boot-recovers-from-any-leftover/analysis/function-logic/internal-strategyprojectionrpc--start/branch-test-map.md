# Branch Test Map: `Start`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (34-116) — revision `base`
- 이 change는 `Start` 본문을 **바꾸지 않았다.** 아래 RED/GREEN은 이 함수를 통과하는
  회수 시나리오에서 관측된 것이고, 함수 자체의 편집 증거가 아니다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | reader가 없다 | 없음(무변경·기존 무테스트) | no | no |
| B2 | 엔진 디렉터리가 안전하지 않다 | 없음(무변경·기존 무테스트) | no | no |
| B3 | control 디렉터리가 이미 있다 | `TestStartRecoversFromDescriptorOnlyLeftover` | yes (`stale endpoint is incomplete`) | yes |
| B4 | Mkdir 실패가 `ErrExist`가 아니다 | 없음(무변경) | no | no |
| B5 | 회수가 거부했다 → 기동 실패 | `TestStartRefusesControlDirectoryWithUnknownEntry` | no (편집 전 GREEN) | yes |
| B6 | 회수 후 재Mkdir 실패 | 없음(무변경) | no | no |
| B7 | listen 실패 | 없음(무변경) | no | no |
| B8 | chmod 실패 | 없음(무변경) | no | no |
| B9 | 토큰 생성 실패 | 없음(무변경) | no | no |
| B10 | 토큰 불일치 → 401 | `TestUnixEndpointStrictMethodBodyQueryAuthAndRouteGuards` | no (기존 GREEN) | yes |
| B11 | GET·HEAD가 아니다 → 405 | `TestUnixEndpointStrictMethodBodyQueryAuthAndRouteGuards` | no (기존 GREEN) | yes |
| B12 | 본문·쿼리가 있다 → 400 | `TestUnixEndpointStrictMethodBodyQueryAuthAndRouteGuards` | no (기존 GREEN) | yes |
| B13 | 스냅샷 검증 실패 → 503 | 없음(무변경) | no | no |
| B14 | HEAD → 헤더만 | 없음(무변경) | no | no |
| B15 | descriptor 발행 실패 | 없음(무변경) | no | no |

## 미테스트 분기에 대하여

11개가 무테스트이고 **전부 이 change가 바꾸지 않은 줄**이다. 편집 전에도 같은 상태였고
(기존 테스트 5개는 정상 경로·인증 경계·회수 두 경우만 덮는다), 이 change는 그 숫자를
줄이지도 늘리지도 않았다. 커버리지 개선은 이 change의 scope가 아니다 —
`not-applicable`이 아니라 **범위 밖**으로 적는다.
