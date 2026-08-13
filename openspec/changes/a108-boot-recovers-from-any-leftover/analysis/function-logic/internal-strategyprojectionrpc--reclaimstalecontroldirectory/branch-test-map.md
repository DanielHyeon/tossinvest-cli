# Branch Test Map: `reclaimStaleControlDirectory`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (144-197)
- 테스트: `internal/strategyprojectionrpc/a108_leftover_recovery_test.go`(신규),
  `internal/strategyprojectionrpc/transport_unix_test.go`(기존)
- "RED observed"는 **편집 전 코드에서 그 테스트가 실패하는 것을 실제로 봤는가**다.
  거부 분기는 편집 전에도 GREEN이었으므로 `no (편집 전 GREEN)`이고, 그 분기가 살아
  있다는 증거는 RED이 아니라 뮤테이션 원장(`../../mutation-ledger-t1.md`)이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 잔재 디렉터리가 0700이 아니거나 symlink다 | `TestStartRefusesUnsafeLeftoverShapes` (0700 아님 · symlink 행) | no (편집 전 GREEN) | yes |
| B2 | 잔재 디렉터리의 소유 uid가 다르다 | 없음 — 같은 uid로 도는 테스트로 재현 불가(무변경 분기) | no | no |
| B3 | `os.ReadDir`이 실패한다 | 없음 — 0700 통과 후 읽기 실패는 환경 이상(무변경 분기) | no | no |
| B4 | 엔트리를 순회해 이름 집합을 만든다 | `TestStartRecoversFromDescriptorOnlyLeftover` | yes (`stale endpoint is incomplete`) | yes |
| B5 | 낯선 엔트리가 하나라도 있다 | `TestStartRefusesControlDirectoryWithUnknownEntry` | no (편집 전 GREEN) | yes |
| B6 | descriptor가 남아 있다 (S1·S3) | `TestStartRecoversFromDescriptorOnlyLeftover` | yes (`stale endpoint is incomplete`) | yes |
| B7 | descriptor가 검증을 통과하지 못한다 | `TestStartRefusesUnsafeLeftoverShapes` (0600 아님 · 반쯤 쓰인 행) | no (편집 전 GREEN) | yes |
| B8 | socket이 남아 있다 (S2·S3) | `TestStartRecoversFromDeadSocketOnlyLeftover` | yes (`stale endpoint is incomplete`) | yes |
| B9 | socket이 socket이 아니거나 0600이 아니다 | `TestStartRefusesUnsafeLeftoverShapes` (일반 파일 · 0600 아님 행) | no (편집 전 GREEN) | yes |
| B10 | socket에 hard link가 걸려 있다 | `TestStartRefusesUnsafeLeftoverShapes` (hard link 행) | no (편집 전 GREEN) | yes |
| B11 | 그 경로에서 누군가 아직 수락한다 | `TestStartRefusesLiveSocketWhoseDescriptorPIDIsDead`, `TestStartRefusesLiveSocketWithoutDescriptor`, `TestStartRefusesLiveProjectionOwnerWithoutRemovingIt` | yes — 첫째는 편집 전 코드가 **살아 있는 socket을 지우고 기동을 받아들였다** | yes |
| B12 | descriptor·socket을 차례로 제거한다 | `TestStartRecoversFromDeadEndpointWhoseDescriptorPIDIsAlive` | yes (`projection owner is still alive`) | yes |
| B13 | 제거가 실패했고 그것이 `ErrNotExist`가 아니다 | `TestStartRecoversFromEmptyControlDirectoryLeftover` (용인 쪽), 뮤테이션 M8 (엄격 쪽) | yes (`stale endpoint is incomplete`) | yes |
| B14 | 디렉터리 제거가 실패한다 | 없음 — B5가 비어 있지 않은 디렉터리를 먼저 막는다(무변경 분기) | no | no |

## 미테스트로 남긴 3개 (B2·B3·B14)

셋 다 이 change가 **바꾸지 않은** 줄이고, 셋 다 편집 전에도 무테스트였다. B2는 다른
uid의 디렉터리, B3은 읽기 불가 디렉터리, B14는 "우리 이름 둘만 있는데 rmdir이 실패"라는
상태를 요구하며, 앞의 둘은 테스트 프로세스 권한으로 만들 수 없고 B14는 B5가 선행 차단한다.
`not-applicable`이 아니라 **측정 한계**로 적는다.
