# Branch Test Map: `New`

a081이 더하는 갈래는 B7 하나다. B1–B6은 이 change가 손대지 않으며 기존 생성자
테스트가 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 참: StartVerify 미배선 → `ErrNoVerifyWiring` / 거짓: 계속 | 기존 | no | yes |
| B2 | 참: 시계 미주입 → 시스템 UTC / 거짓: 주입된 시계 | 기존 | no | yes |
| B3 | 참: remote runtime 실패 전파 / 거짓: 계속 | 기존 | no | yes |
| B4 | 참: 출력 미주입 → Discard / 거짓: 주입된 출력 | 기존 | no | yes |
| B5 | 참: Binary 미주입 → binstamp.Self / 거짓: 주입값 | 기존 | no | yes |
| B6 | 참: boot note 있음 → 보관 / 거짓: 없음 | 기존 | no | yes |
| B7 | 참: 커맨더 배선 → **간격 둘을 받는** 캐시 생성, 이후 렌더가 각 간격을 공유한다 / 거짓: 미배선 → 캐시 nil, 엔진 읽기 0회 | 2.2, 5.2, 8.2 | yes | yes |

B7의 참 갈래 RED는 `TestRedrawingTheLineDoesNotAskTheEngineAgainWithinTheInterval`이
캐시 없이 렌더마다 읽기를 관측한 것이고, 두 간격이 하나로 묶이는 변이는
`TestTheCacheIntervalsAreTheirOwnConstants`와
`TestAReconcileBlockIsNotHeldForTheLifecycleInterval`이 잡는다. 거짓 갈래는
커맨더를 배선하지 않는 기존 700여 건이 무수정으로 통과하는 것으로 확인한다.
