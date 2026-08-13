# Branch Test Map: `reclaimStaleControlDirectory`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (206-282) — revision `current`
- RED은 첫 라운드(tasks 1절)와 Fix 라운드(tasks 6.1) 두 번 있었다. 아래 표의 RED 열은
  **그 분기를 연 라운드에서** 관측한 것이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 디렉터리가 0700이 아니거나 symlink다 | `TestStartRefusesUnsafeLeftoverShapes` | no (편집 전 GREEN) | yes |
| B2 | 남이 만든 디렉터리다 | `TestOwnershipClauseRefusesAnotherUser` | no (Fix 라운드 신규 핀) | yes |
| B3 | `ReadDir` 실패 | 없음 | no | no |
| B4 | 엔트리를 훑는다 | `TestStartRefusesControlDirectoryWithUnknownEntry` | no (편집 전 GREEN) | yes |
| B5 | 이름을 셋으로 가른다 | `TestStartRecoversFromUnpublishedStagingLeftover` | yes (`unexpected entries`) | yes |
| B6 | 최종 이름이다 | `TestStartRecoversFromDescriptorOnlyLeftover` | yes (`stale endpoint is incomplete`) | yes |
| B7 | 발행 전 임시 이름이다 | `TestStartRecoversFromUnpublishedStagingLeftover` | yes (`unexpected entries`) | yes |
| B8 | 낯선 이름이다 → 전체 보존 | `TestStartRefusesControlDirectoryWithUnknownEntry` | no (편집 전 GREEN) | yes |
| B9 | socket이 남아 있다 | `TestStartRecoversFromDeadSocketOnlyLeftover` | yes (`stale endpoint is incomplete`) | yes |
| B10 | 거부: 남의 모양 / 통과: pre-chmod 0700은 우리 것 | `TestStartRefusesUnsafeLeftoverShapes` · `TestStartRecoversFromPreChmodSocketLeftover` | yes (`stale socket is unsafe` — 후자) | yes |
| B11 | 누가 아직 수락한다 → 거부 | `TestStartRefusesLiveSocketWhoseDescriptorPIDIsDead` | no (편집 전 GREEN) | yes |
| B12 | descriptor가 남아 있다 | `TestStartRecoversFromDescriptorOnlyLeftover` | yes | yes |
| B13 | 거부: **형식**이 틀림 / 통과: **내용**이 반쯤 | `TestStartRefusesUnsafeLeftoverShapes` · `TestStartRecoversFromEmptyDescriptorLeftover` · `TestStartRecoversFromTruncatedDescriptorLeftover` | yes (`stale descriptor is unsafe` — 후자 둘) | yes |
| B14 | 제거 목록을 훑는다 | `TestStartRecoversFromDeadEndpointWhoseDescriptorPIDIsAlive` | yes (`projection owner is still alive`) | yes |
| B15 | 이미 없는 파일은 실패가 아니다 | `TestCloseToleratesLeftoverAlreadyRemoved` | no (편집 전 GREEN) | yes |
| B16 | 이미 없는 디렉터리는 실패가 아니다 | 없음 — 아래 「B16은 재지 않았다」 | no | no |

B10과 B13의 행이 두 판정을 함께 적는 것은 같은 줄이 **어떤 모양은 거부하고 어떤
모양은 통과시키기** 때문이다. 한쪽만 적으면 완화가 어디까지인지 문서에서 사라진다.

## B16은 재지 않았다 (선언된 생략)

`os.Remove(controlDir)`의 `ErrNotExist` 절은 파일 제거와 rmdir **사이**에 디렉터리가
사라져야 걸린다. 그 인터리빙은 production에 seam을 뚫지 않고 결정적으로 만들 수 없고,
두 회수를 동시에 돌리는 방법도 진 쪽이 앞 단계에서 먼저 실패해 단언이 흔들린다.
뮤테이션 M17은 **살아남았고** 원장 §B3에 그렇게 적었다. A1 F6이 요구한 것은 핀이
아니라 파일/디렉터리 비대칭의 제거였다(P3) — 그것은 코드에 있다.

## 첫 라운드에서 이 표가 달라진 점

- B5·B7이 새로 생겼다(staging 분류). 첫 라운드에는 이름이 둘뿐이었다.
- B9~B11이 B12~B13보다 **앞으로** 왔다. 사망 입증이 descriptor 엄격도를 정하기 때문이다.
- B16이 새로 생겼다(rmdir의 `ErrNotExist` 절).
