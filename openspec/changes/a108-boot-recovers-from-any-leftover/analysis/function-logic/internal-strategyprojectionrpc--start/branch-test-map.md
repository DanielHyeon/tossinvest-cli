# Branch Test Map: `Start`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (79-162) — revision `current`
- Fix 라운드가 이 함수를 편집했다(socket 발행을 `listenPrivateSocket`으로). 첫 라운드의
  "무변경" 선언은 철회됐다 — 근거는 `function-logic-map.md`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | reader가 없다 | 없음(무변경·기존 무테스트) | no | no |
| B2 | 엔진 디렉터리가 안전하지 않다 | 없음(무변경·기존 무테스트) | no | no |
| B3 | control 디렉터리가 이미 있다 | `TestStartRecoversFromDescriptorOnlyLeftover` | yes (`stale endpoint is incomplete`) | yes |
| B4 | Mkdir 실패가 `ErrExist`가 아니다 | 없음(무변경) | no | no |
| B5 | 회수가 거부했다 → 기동 실패 | `TestStartRefusesControlDirectoryWithUnknownEntry` | no (편집 전 GREEN) | yes |
| B6 | 회수 후 재Mkdir 실패 | 없음(무변경) | no | no |
| B7 | socket 발행 실패 | 없음 — 아래 「B7의 간접 측정」 | no | no |
| B8 | 토큰 생성 실패 | 없음(무변경) | no | no |
| B9 | 토큰 불일치 → 401 | `TestUnixEndpointStrictMethodBodyQueryAuthAndRouteGuards` | no (기존 GREEN) | yes |
| B10 | GET·HEAD가 아니다 → 405 | `TestUnixEndpointStrictMethodBodyQueryAuthAndRouteGuards` | no (기존 GREEN) | yes |
| B11 | 본문·쿼리가 있다 → 400 | `TestUnixEndpointStrictMethodBodyQueryAuthAndRouteGuards` | no (기존 GREEN) | yes |
| B12 | 스냅샷 검증 실패 → 503 | 없음(무변경) | no | no |
| B13 | HEAD → 헤더만 | 없음(무변경) | no | no |
| B14 | descriptor 발행 실패 | 없음 — 아래 「B7의 간접 측정」 | no | no |

## B7의 간접 측정 (발행의 **성공** 경로가 재는 것)

B7·B14는 실패 분기이고 테스트가 없다. 그러나 그 두 줄이 부르는 발행 의례 자체는
성공 경로에서 재고 있다.

- `TestStartPublishesBothArtifactsByRename` — 기동이 끝난 디렉터리에 최종 이름 둘만
  있고, listener가 기억하는 이름이 최종 socket 경로가 **아니다**(제자리 bind가 아님).
- `TestDescriptorPublicationIsAtomic` — 재발행 300회와 동시에 읽으면서 반쯤 쓰인
  descriptor가 보이지 않는다.

뮤테이션 원장 §B M9·M9b·M10이 이 두 테스트로 죽는다.

## 미테스트 분기에 대하여

9개가 무테스트다. B1·B2·B4·B6·B8·B12·B13은 **이 change가 바꾸지 않은 줄**이고 편집
전에도 같은 상태였다 — 커버리지 개선은 이 change의 scope가 아니라 **범위 밖**이다.
B7·B14는 이 change가 만든 줄이지만 실패 주입 seam이 없다(발행 실패를 만들려면
파일시스템 오류를 주입해야 한다) — 위의 간접 측정으로 대신하고 `not-applicable`로
선언한다. 침묵한 생략이 아니다.
