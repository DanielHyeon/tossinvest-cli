# Branch Test Map: `wtsPopularityReader`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil 클라이언트는 nil 인터페이스가 된다 | **직접 커버 없음** — 계약의 반대편만 `internal/candidatesrc`의 `TestThePanelWithoutAnOfficialClientIsNotSilentlyWTSOnly`가 잡는다(nil reader → 소스 부재) | no | no |

**정직한 커버리지 기록**: `cmd/tossctl` 안에서 이 함수를 직접 부르는 테스트는 없다.
계약의 반대편(`Panel`이 nil reader를 어떻게 다루는가)만 테스트되어 있고, typed-nil 방지
자체는 구조로 선다.
