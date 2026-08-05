# Branch Test Map: `Console.handlePositionPolicyApply`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 참: 커맨더 미배선 → 501 / 거짓: 계속 | 기존 | no | yes |
| B2 | 참: capability 없음 → 403, 엔진 미도달 / 거짓: 계속 | 기존, 4.3 | no | yes |
| B3 | 참: Apply 실패 → 분류된 거부, **캐시 유지** / 거짓: 성공 경로 | 4.3 | no | yes |
| C1 | 성공 뒤 캐시가 버려져 다음 렌더가 엔진을 다시 읽는다 | 4.1 | yes | yes |

C1의 RED는 무효화 호출을 지운 변이 검증(6.3)에서 관측했다 —
`runtime 4→4, list 4→4`.
