# Function Logic Map: `TestRegisteringTheUSTierMovedExactlyTwoCeilings`

- Source: `internal/config/limits_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| USD 상한 | notional 500, exposure 1,500 | design D1 | 다르면 `t.Errorf` |
| 불변 USD 필드 | quantity 100, ratio 0.01, daily loss 50 | 전 티어 공통값 + D1의 미이동 결정 | 움직였으면 `t.Errorf` |
| KRW 상한 | `kr-small-live` 전체 | 승인 집합 | 다르면 `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | USD 상한 조회 오류 | 없음 | `t.Fatalf` | `TestAnUnregisteredCurrencyFailsClosed` |
| B2 | notional != 500 | 없음 | `t.Errorf` | 자기 자신 |
| B3 | exposure != 1,500 | 없음 | `t.Errorf` | 자기 자신 |
| B4 | quantity != 공유 cap | 없음 | `t.Errorf` | 자기 자신 |
| B5 | ratio != 0.01 | 없음 | `t.Errorf` | 자기 자신 |
| B6 | daily loss != 50 | 없음 | `t.Errorf` (사유 문구에 1,333 KRW/USD 근거) | 자기 자신 |
| B7 | KRW 상한 조회 오류 | 없음 | `t.Fatalf` | 자기 자신 |
| B8 | KRW 상한 != 승인 집합 | 없음 | `t.Errorf` | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `GuardianCeiling` | 통화별 필드 최대 | 미등록 통화는 error | CodeGraph + AST |
| `stockOSTiers` (map) | KRW 기대치를 전사 코퍼스에서 가져온다 | 없음 | 같은 파일 |

## State mutations and fallbacks

- 상태 변경 없음. 신규 테스트다.
- 존재 이유: 상한 표(`TestTheCeilingIsTheMaxAcrossRegisteredTiers`)는 값이 맞는지만 말하고 **무엇이 움직였는지**는 말하지 않는다. 이 change가 승인받은 것은 "USD 두 필드"이므로, 셋째 필드나 KRW로 샌 완화는 표가 갱신되는 순간 함께 통과해 버린다. B4~B6과 B8이 그 경로를 막는다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: 없음에 가깝다 — 이 테스트의 리터럴을 고치는 것은 승인 범위를 고치는 것이다.
- High-risk impact: yes — §0.9 완화 범위의 기계적 경계다.
