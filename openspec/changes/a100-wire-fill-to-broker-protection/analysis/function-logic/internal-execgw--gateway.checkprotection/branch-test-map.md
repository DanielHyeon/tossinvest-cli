# Branch Test Map: `Gateway.checkProtection`

> **측정 방법**: `go test -covermode=set -coverprofile` (패키지 전체). 분기의 *조건*이 아니라
> **true 결과 본문의 실행 여부**를 측정했다. 조건 statement가 covered인 것은 조건이 평가됐다는
> 뜻일 뿐 그 분기를 탔다는 뜻이 아니므로, 본문 행을 따로 측정했다.

| Branch | Scenario | Test | true 결과 실행됨 | 비고 |
|---|---|---|---|---|
| B1 | 노출 감소 mutation은 보호를 묻지 않는다 | `internal/execgw` 패키지 (338 tests) | **yes** (L91) | reduce-only 비대칭의 근거 |
| B2 | test seam 위임 | 동상 | **yes** (L94) | 프로덕션 경로 아님 |
| B3 | readiness provider 부재 → 거부 | 동상 | **yes** (L97) | |
| B4 | 비정규 수량 → 거부 | 동상 | **yes** (L101) | |
| B5 | readiness 거부 전파 | 동상 | **yes** (L107) | |
| fall-through | 통과 | 동상 | **yes** (L109) | seam 경유로만 도달 가능 |

**측정 결과: 6개 경로 전부 실행된다.** 담당 테스트 개별 특정은 하지 않았다(패키지 338 tests).
a100은 이 함수의 분기를 늘리지 않으므로 RED/GREEN 대상이 아니다.
