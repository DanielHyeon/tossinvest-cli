# Branch Test Map: `publishPrivateDescriptor`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 0700 아닌 control 디렉터리에서 발행 거부 | 커버 없음 | no | no |
| B2 | `.endpoint-` 접두로 staging 생성 | a109 §1.5의 이름-집합 완전성 테스트가 이 접두로 만든 **실제 이름**이 회수의 아는-이름 집합에 속함을 잰다 | no | yes(§1.5) |
| B3 | 열린 staging 파일 검증 실패 | 커버 없음 | no | no |
| B4 | staging Stat 실패 | 커버 없음 | no | no |
| B5 | 본문 쓰기 실패 | `TestWritePositionPolicyDescriptorPreservesChmodAndWriteErrors`가 피호출 함수의 세 실패(chmod·write·short write)를 잰다 | no | no |
| B6 | 닫힌 staging 검증 실패 | 커버 없음 | no | no |
| B7 | staging inode 교체 거부 | 커버 없음(읽기 쪽 동형은 `TestOpenPrivateDescriptorRejectsInodeReplacementDuringOpen`) | no | no |
| B8 | rename 실패 | 커버 없음 | no | no |
| B9 | 정상 발행이면 rollback 없음 | `TestTheAlertSocketIsPrivateToThisUser`(0600 descriptor 존재) | no | no |
| B10 | rename 뒤 실패 시 최종 이름 회수 | 커버 없음 | no | no |
| B11 | published inode 불일치 | 커버 없음 | no | no |
| B12 | published 검증 실패 | 커버 없음 | no | no |
| B13 | 디렉터리 open 실패 | 커버 없음 | no | no |
| B14 | 디렉터리 sync 실패 | 커버 없음 | no | no |
