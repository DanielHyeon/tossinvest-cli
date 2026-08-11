# Branch Test Map: `prepareRegister`

> **측정 방법**: `go test -covermode=set -coverprofile`. 분기 *조건*이 아니라 **true 결과
> 본문의 실행 여부**를 측정했다. 담당 테스트는 테스트별 개별 프로파일로 특정했다.

| Branch | Scenario | Test | true 결과 실행됨 |
|---|---|---|---|
| B1 | position 취득/capability 실패 | `TestSealedStateAndCapabilityTamperFailClosed` | **yes** (L6) |
| B2 | entry latched / phase 부적합 | — | **NO** |
| B3 | 이미 pending인 operation | — | **NO** |
| B4 | 보호가 이미 active | — | **NO** |
| B5 | 브로커가 정확한 operation 조회를 못 함 | — | **NO** |
| B6 | 보호 수량이 보유를 정확히 채우지 못함 | `TestRegisterRequiresFullAvailableCoverage` | **yes** (L21) |
| fall-through | 정상 제출 준비 | `TestRegisterResponseCrashRecoversExactlyOnce` 외 | **yes** (L32) |

## 발견된 공백

**6개 분기 중 4개(B2·B3·B4·B5)가 미실행이다.** 성격이 applyFill과 다르다 — 이쪽은
**중복 제출 방지** 계열이다.

- **B3·B4 (이미 pending / 이미 active)** — 같은 포지션에 보호주문이 두 번 나가는 것을 막는 방어선.
  프로덕션 배선 후 재시도·복구 경로에서 가장 먼저 부딪힐 분기인데 미검증이다.
- **B5 (exactOperationLookup 불가)** — 설계 문서 §8.2가 요구하는 "시장별 실계좌 능력 매트릭스"가
  실제로 판정을 막는지 확인된 적이 없다.
- **B2 (entry latched)** — 진입이 닫힌 상태에서 보호 등록을 시도하는 경우.

**a100 tasks에 4개 분기 각각의 RED 테스트를 포함한다.**
