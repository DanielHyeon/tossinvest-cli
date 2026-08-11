# Branch Test Map: `Detector.pollLocked`

> **측정 방법**: `go test -covermode=set -coverprofile ./internal/filldetect` (패키지 전체,
> 87.3% of statements). 분기의 *조건*이 아니라 **true 결과 본문 행의 실행 여부**를 측정했다.
> 조건 statement가 covered인 것은 조건이 평가됐다는 뜻일 뿐 그 분기를 탔다는 뜻이 아니다.
> 측정일 2026-08-11, source SHA-256 `5441296826821097…`(`ast.json`).

| Branch | Scenario | Test | true 결과 실행됨 | 비고 |
|---|---|---|---|---|
| B1 | 구성 검증 실패 | `internal/filldetect` 패키지 | **yes** (L286) | |
| B2 | 주문 조회 실패 → outage | 동상 | **yes** (L295) | |
| B3 | 계좌 포지션 조회 실패 → outage | 동상 | **no** (L302) | |
| B4 | 매수여력 sweep 실패 → outage | 동상 | **no** (L310) | |
| B5 | 스냅샷 루프 본문 | 동상 | **yes** (L317) | 직렬 처리 |
| B6 | **`Ledger.Apply` 실패 → outage + 잔여 스냅샷 폐기** | — | **no** (L319) | **a100이 도달 가능하게 만들려던 분기** |
| B7 | `FailClosed` → 심볼 차단 후 continue | 동상 | **yes** (L325) | |
| B8 | `Delta > 0` → 체결 계수 + 신선도 표본 | 동상 | **yes** (L330) | SLO 오염 경로 |
| B9 | `CommittedAt` zero → 현재 시각 대체 | 동상 | **no** (L333) | |
| B10 | 음수 latency → 0으로 바닥 | 동상 | **no** (L339) | 시계 skew |

**측정 결과: 10개 중 5개가 true 결과를 한 번도 실행하지 않았다** (B3, B4, B6, B9, B10).

**B6이 이 change에 결정적이다.** a100의 원안(D8, `filldetect.Ledger` 데코레이터)은 보호
왕복을 `Ledger.Apply` 안에 넣는다. 그 배치는 (1) 보호 실패를 `Apply` 에러로 만들 수 있고 —
그러면 **한 번도 실행된 적 없는 B6이 프로덕션에서 처음 실행되며 같은 사이클의 남은 체결
반영을 버린다** — (2) 실패 없이 성공하더라도 왕복 시간이 같은 사이클 뒤 스냅샷의
`CommittedAt`을 밀어 B8의 신선도 표본을 오염시킨다.

원안 tasks 4.6은 (1)만 격리했고 (2)는 다루지 않았다. **두 경로 모두 회피하는 유일한 배치는
보호를 이 함수 밖으로 내보내는 것이다** — 재설계된 D8이 그렇게 한다. 따라서 a100은 B3·B4·B6·
B9·B10을 새로 도달 가능하게 만들지 않으며, 이 다섯은 **a100의 RED 대상이 아니다**
(`internal/filldetect`를 편집하지 않으므로 그 패키지의 커버리지 개선은 별도 change의 일이다).
