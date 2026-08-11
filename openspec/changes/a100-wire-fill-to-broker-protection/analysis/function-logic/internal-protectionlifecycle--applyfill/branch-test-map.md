# Branch Test Map: `applyFill`

> **측정 방법**: `go test -covermode=set -coverprofile`. 분기 *조건*이 아니라 **true 결과
> 본문의 실행 여부**를 측정했다. 담당 테스트는 테스트별 개별 프로파일로 특정했다.

| Branch | Scenario | Test | true 결과 실행됨 |
|---|---|---|---|
| B1 | position 취득 실패(비-InvalidObservation) | — | **NO** |
| B2 | state seal 무효 | — | **NO** |
| B3 | fill 식별자 무효 / broker order id 불일치 | — | **NO** |
| B4 | 같은 FillID, 다른 내용 → latch + 거부 | `TestDuplicateAndPartialFillConvergeOnce` | **yes** (L252) |
| B5 | 같은 FillID, 같은 내용 → 멱등 Duplicate | `TestDuplicateAndPartialFillConvergeOnce` | **yes** (L249) |
| B6 | 수량 0 또는 claim 초과 | — | **NO** |
| B7 | 잔량 0 → `Terminal` 전이 | — | **NO** |
| fall-through | 정상 부분체결 반영 | `TestDuplicateAndPartialFillConvergeOnce` | **yes** (L270) |

## 발견된 공백 — a100의 가장 큰 리스크

**7개 분기 중 5개(B1·B2·B3·B6·B7)가 한 번도 실행되지 않는다.**

이 함수는 a100이 프로덕션에 배선하려는 **바로 그 함수**다. 지금까지 호출자가 없었으므로
거부 경로가 검증되지 않은 채 남아 있었다. 특히 위험한 둘:

- **B3 (broker order id 불일치)** — 잘못된 체결을 남의 포지션에 귀속시키는 것을 막는 유일한 방어선.
- **B7 (잔량 0 → Terminal)** — 완전 체결로 보호가 종료되는 전이. 미실행이라는 것은
  **보호주문이 다 채워졌을 때 상태가 올바르게 닫히는지 확인된 적이 없다**는 뜻이다.

**a100 tasks는 이 5개 분기 각각에 RED 테스트를 먼저 세운다.** 배선보다 먼저다 —
호출자가 없는 동안 숨어 있던 결함을 배선과 동시에 프로덕션으로 내보내지 않기 위해서다.
