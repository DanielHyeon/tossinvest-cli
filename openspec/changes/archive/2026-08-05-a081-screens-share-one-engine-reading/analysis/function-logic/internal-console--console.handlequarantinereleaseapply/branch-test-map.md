# Branch Test Map: `Console.handleQuarantineReleaseApply`

> **개정 2026-08-05 (review.md 3차, 8.4).** 초안의 C1("성공 뒤 캐시가 버려진다")은
> 존재하지 않는 data flow를 단정하고 있었다. 그 계약과 테스트를 함께 제거했다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 참: 격리 커맨더 미발견 → refuse / 거짓: 계속 | 기존 | no | yes |
| B2 | 참: capability 없음 → 403, 엔진 미도달 / 거짓: 계속 | 기존 | no | yes |
| B3 | 참: 해제 실패 → 분류된 거부 / 거짓: 성공 경로 → 303 redirect | 기존 | no | yes |
| C1 | 격리 상태는 엔진 정책 캐시를 지나지 않으므로 이 경로는 캐시를 무효화하지 않는다 | 8.4 (코드 주석 + logic map 불변식 1) | — | — |

C1은 **부재의 계약**이라 테스트로 단정하지 않는다. 화면의 격리 배지가 콘솔 자신의
journal 읽기에서 온다는 것은 a079의 기존 테스트가 이미 덮고, 여기서 확인할 것은
"무효화가 없다"는 사실 자체다. 그것을 테스트로 고정하면 부재를 단정하는 테스트가
되어, 나중에 정당한 이유로 무효화가 필요해졌을 때 그 테스트가 먼저 막는다.
근거는 코드 주석과 Function Logic Map 불변식 1에 남긴다.
