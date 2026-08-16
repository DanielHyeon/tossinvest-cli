# Branch Test Map: `validatePrivateDirectory`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 빈 경로 | 커버 없음 — 모든 호출자가 구성된 경로를 넘긴다 | no | no |
| B2 | symlink control 디렉터리 | `TestPositionPolicyControlRejectsInsecureControlFilesystem` (`position_policy_transport_test.go:227`, "symlinked control directory") | no | no |
| B3 | 디렉터리 부재 | `TestAlertControlServerStartsOverALeftover`의 첫 기동 경로가 Mkdir 후에만 검증에 닿는 순서를 잰다(부재 자체의 직접 커버는 없음) | no | no |
| B4 | 디렉터리 자리의 정규 파일 | 커버 없음 | no | no |
| B5 | 남의 uid 소유 | 커버 없음 — 비root 테스트는 소유자를 바꿀 수 없다 | no | no |
| B6 | 0750 control 디렉터리 거부 | `TestPositionPolicyControlRejectsInsecureControlFilesystem` ("wrong existing control directory mode") | no | no |
| B7 | 0770 엔진 디렉터리 거부 | `TestPositionPolicyControlRejectsInsecureControlFilesystem` ("group writable engine directory") | no | no |
