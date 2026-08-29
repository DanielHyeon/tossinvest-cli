# Function Logic Map: `TestTheCeilingIsTheMaxAcrossRegisteredTiers`

- Source: `internal/config/limits_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| KRW 기대 상한 | 승인 집합 다섯 값 | `risk-management` 정책 수치의 provenance | 불일치는 `t.Errorf` |
| USD 기대 상한 | 500 / 1,500 / 50 / 1% / 100 | design D1 (이 change에서 갱신) | 불일치는 `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 통화 두 건 순회 | 없음 | 없음 | 자기 자신 |
| B2 | `GuardianCeiling` 오류 | 없음 | `t.Fatalf` | `TestAnUnregisteredCurrencyFailsClosed` |
| B3 | 상한 값 불일치 | 없음 | `t.Errorf` | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `GuardianCeiling` | 통화별 필드 최대 | 미등록 통화는 error (fail-closed) | CodeGraph + AST |

## State mutations and fallbacks

- 상태 변경 없음.
- 이 change의 편집은 USD 기대치 두 필드(notional 300→500, exposure 1,000→1,500)뿐이다. KRW 블록은 한 글자도 건드리지 않았고, 그 사실 자체를 `TestRegisteringTheUSTierMovedExactlyTwoCeilings`가 따로 다시 확인한다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: 두 기대 리터럴. 상한이 움직이면 반드시 여기가 먼저 빨개진다.
- High-risk impact: yes — 이 테스트가 상한 이동을 조용히 통과시키면 §0.9 검토 없이 완화가 들어온다.
