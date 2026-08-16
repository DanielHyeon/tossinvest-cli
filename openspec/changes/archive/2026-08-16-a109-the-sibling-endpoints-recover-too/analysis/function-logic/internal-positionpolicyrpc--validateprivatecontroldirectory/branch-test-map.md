# Branch Test Map: `ValidatePrivateControlDirectory`

RED/GREEN 열은 **a109가 이 함수를 편집하지 않으므로** 전부 no다. 여기 적는 것은
"a109 이전에 참이고 이후에도 참이어야 하는 것"의 근거다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 빈 control 디렉터리 경로 | 커버 없음 — 세 호출자 모두 `filepath.Dir(...)` 결과를 넘기므로 빈 문자열이 도달하지 않는다(도달 불가에 가까운 방어) | no | no |
| B2 | group-writable 엔진 디렉터리 아래에서 기동 | 성립 방향은 `TestTheAlertSocketIsPrivateToThisUser`가 0700/0600을 실측한다. 거부 방향은 같은 `validatePrivateDirectory`를 쓰는 자매 경로의 `TestPositionPolicyControlRejectsInsecureControlFilesystem`이 잰다 | no | no |
