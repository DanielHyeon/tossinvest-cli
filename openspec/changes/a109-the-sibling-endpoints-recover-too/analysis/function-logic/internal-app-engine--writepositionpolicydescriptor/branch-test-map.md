# Branch Test Map: `writePositionPolicyDescriptor`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 직렬화 실패 | 커버 없음 — 고정 구조체라 도달 불가에 가깝다 | no | no |
| B2 | 0750 control 디렉터리에서 staging 거부 | `TestPositionPolicyControlRejectsInsecureControlFilesystem` | no | no |
| B3 | `.position-policy-control-` 접두로 staging 생성 | a109 §1.5 이름-집합 완전성 테스트가 이 접두로 만든 실제 이름의 소속을 잰다 | no | yes(§1.5) |
| B4 | 열린 staging 검증 실패 | 커버 없음 | no | no |
| B5 | staging Stat 실패 | 커버 없음 | no | no |
| B6 | 본문 쓰기 실패 | `TestWritePositionPolicyDescriptorPreservesChmodAndWriteErrors` (`position_policy_transport_test.go:67`) | no | no |
| B7 | 닫힌 staging 검증 실패 | 커버 없음 | no | no |
| B8 | staging inode 교체 거부 | 커버 없음 | no | no |
| B9 | rename 직전 디렉터리 재검증 실패 | 커버 없음 | no | no |
| B10 | rename 실패 | 커버 없음 | no | no |
| B11 | 정상 발행이면 rollback 없음 | `TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint`(0600 정규 파일 확인) | no | no |
| B12 | rename 뒤 실패 시 최종 이름 회수 | 커버 없음 | no | no |
| B13 | published 이름·모드 검증 | `TestPositionPolicyCommandServerStartsOverALeftover`가 회수 후 재발행된 descriptor를 읽는다 | no | no |
| B14 | published inode 불일치 | 커버 없음 | no | no |
| B15 | 디렉터리 open 실패 | 커버 없음 | no | no |
| B16 | 디렉터리 sync 실패 | 커버 없음 | no | no |
| B17 | 디렉터리 close 실패 | 커버 없음 | no | no |
