# Branch Test Map: `PreparePrivateSocket`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 잔재 없음 → nil | `TestTheAlertSocketIsPrivateToThisUser`(첫 기동 경로) | no | no |
| B2 | socket 아닌 것(정규 파일)은 지우지 않고 거부 | 자매 함수의 동형 핀 `TestPrepareRuntimeSocketNeverDeletesNonSocket` (`internal/positionpolicyrpc/runtime_test.go:67`)이 같은 조건절을 잰다. 이 함수 자체의 직접 커버는 없다 | no | no |
| B3 | 남의 uid 소유 socket 거부 | 커버 없음 — 비root 테스트는 파일 소유자를 바꿀 수 없다(a108 `ownedByEffectiveUser` 주석의 같은 사유) | no | no |
| 분기 밖 종단(`os.Remove`) | 산 주인의 socket을 지우고 두 번째 서버가 올라선다 | a109 §1.2 RED가 이 동작을 고정하고, GREEN에서 기동 경로가 이 함수를 부르지 않게 되어 거부로 바뀐다 | yes(§1.2) | yes(§1.4) |
