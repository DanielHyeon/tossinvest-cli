# Function Logic Map: `TestTheUSTierMatchesItsRecordedDerivation`

- Source: `internal/config/limits_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `us-single-name` | 500 / 1,500 / 50 / 1% / 100 / USD | design D1 | 미등록은 `t.Fatal` |
| `us-small-live` 형태 | 노출/주문 >= 3 | StockOS 전사 코퍼스 | 3 미만이면 D1 논증 재검토 요구 |
| KRW 등가 임계 | 2,000 KRW/USD | design D1 상한 논증 | 초과는 `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 티어 미등록 | 없음 | `t.Fatal` | 자기 자신 |
| B2 | 노출 != 주문×3 | 없음 | `t.Errorf` | 자기 자신 |
| B3 | `us-small-live` 형태가 3배 미만 | 없음 | `t.Errorf` ("stricter shape" 논증 붕괴 경고) | 자기 자신 |
| B4 | 일일 손실 != `us-small-live` | 없음 | `t.Errorf` | 자기 자신 |
| B5 | 전 티어 순회 | 없음 | 없음 | 자기 자신 |
| B6 | 수량이 티어마다 다름 | 없음 | `t.Errorf` | 자기 자신 |
| B7 | 비율이 티어마다 다름 | 없음 | `t.Errorf` | 자기 자신 |
| B8 | 주문 상한 × 2,000 > 승인 KRW | 없음 | `t.Errorf` (등가 논증 붕괴) | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `GuardianTierByID` | 대상 티어 | 미등록은 `ok=false` | CodeGraph + AST |
| `GuardianTiers` | 공통값 검사 | 오류 없음 | CodeGraph + AST |

## State mutations and fallbacks

- 상태 변경 없음. 신규 테스트다.
- 이 테스트가 지키는 것은 값이 아니라 **논증**이다. B8은 "누군가 이 티어를 올리면 무엇을 다시 논증해야 하는지"를 실행 가능한 형태로 남긴 것이고, B3은 논증이 기댄 전제(`us-small-live`가 더 느슨한 형태)가 나중에 깨지면 알려 준다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: `parityBreaksAt` 상수와 3배 계수. 둘 다 design D1과 함께 움직여야 한다.
- High-risk impact: yes — 사이징 수치의 유도 근거다.
