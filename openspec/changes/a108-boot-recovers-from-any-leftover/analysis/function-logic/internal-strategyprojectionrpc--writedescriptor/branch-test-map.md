# Branch Test Map: `writeDescriptor`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (398-453) — revision `current`
- Fix 라운드가 이 함수를 stage+rename으로 다시 썼다(design D1-2, A1 F2).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | JSON marshal 실패 | 없음 — 아래 「실패 분기를 재지 않은 이유」 | no | no |
| B2 | 임시 파일 생성 실패 | 없음 | no | no |
| B3 | 임시 파일 chmod 실패 | 없음 | no | no |
| B4 | 임시 파일 stat 실패 | 없음 | no | no |
| B5 | 본문 write 실패 | 없음 | no | no |
| B6 | sync 실패 | 없음 | no | no |
| B7 | close 실패 | 없음 | no | no |
| B8 | 닫힌 임시 파일이 바뀌었다 | 없음 | no | no |
| B9 | rename 실패 | 없음 | no | no |
| B10 | 발행된 파일이 임시 파일이 아니다 | 없음 | no | no |
| B11 | 디렉터리 열기 실패 | 없음 | no | no |
| B12 | 디렉터리 sync 실패 | 없음 | no | no |

## 이 함수가 재는 것은 분기가 아니라 성질이다

열두 분기는 전부 **실패** 분기이고, 성공 경로에는 분기가 없다. 그래서 표의 열두 줄은
비어 있고, 이 함수의 계약은 성질로 잰다.

| 성질 | Test | RED observed | GREEN observed |
|---|---|---|---|
| 최종 이름은 완성된 파일에만 붙는다 | `TestDescriptorPublicationIsAtomic` | no (Fix 라운드 신규 핀) | yes |
| 기동 뒤 임시 이름이 남지 않는다 | `TestStartPublishesBothArtifactsByRename` | no (신규 핀) | yes |
| 반쯤 쓰인 descriptor는 회수된다 | `TestStartRecoversFromEmptyDescriptorLeftover`, `TestStartRecoversFromTruncatedDescriptorLeftover` | yes (`stale descriptor is unsafe`) | yes |
| 임시 이름 잔재는 회수된다 | `TestStartRecoversFromUnpublishedStagingLeftover` | yes (`unexpected entries`) | yes |

뮤테이션 M9(제자리 `O_EXCL`)·M9b(제자리 `O_TRUNC`)가 첫 행 하나로 죽는다. 처음 원장을
돌렸을 때 M9는 **살아남았다** — 회수 테스트가 반쯤 쓰인 파일을 디스크에 직접 만들기
때문에 `writeDescriptor`가 어떻게 쓰든 통과했기 때문이다. 그 구멍을 첫 행이 닫았다.

## 실패 분기를 재지 않은 이유 (선언된 생략)

B1~B12는 파일시스템 오류(ENOSPC·EIO·권한)를 주입해야 닿는다. 주입 seam을 만들려면
production 코드에 파일 연산 간접층을 넣어야 하고, 그것은 이 change의 범위가 아니며
발행 경로에 새 실패 모드를 더한다. `not-applicable`로 선언한다 — 침묵한 생략이 아니다.

각 실패 분기가 하는 일은 같다: 임시 이름을 지우고(`defer`) 오류를 그대로 올린다.
최종 이름은 손대지 않는다. 그 불변은 코드 모양으로 읽히고, 성질 표의 첫 두 행이
"최종 이름이 부분 상태를 갖지 않는다"를 밖에서 확인한다.
