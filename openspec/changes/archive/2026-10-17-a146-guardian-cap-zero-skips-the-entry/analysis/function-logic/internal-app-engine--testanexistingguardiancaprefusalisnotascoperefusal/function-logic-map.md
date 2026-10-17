# Function Logic Map: `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal`

- Source: `internal/app/engine/a112_qfinal_family_test.go` (248-265)
- Qualified: `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal`
- AST evidence: `ast.json`(현재) — 현재 AST branches 3 · returns 0
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 84abf4ec7ba8d3a8…) — 편집 전 갈래 수 2
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 — 버킷 고갈이 아닌 q_final 거절(예산 1 KRW → Guardian 상한 0)은 범위 거절이 아님.

**a146 편집.** 기대를 새 동작으로: 범위 거절이 아님은 그대로, 대신 주기가 멈추지 않고(오류 없음) 두 범위가 다 평가되며 브로커 0.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if placed := fixture.placedSymbols(); len(placed) != 0 {` (if 253) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| B2 | `if err != nil \|\| a112ScopeRefusalOf(err) != nil {` (if 256) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| B3 | `if len(skips) != 2 \|\| !errors.As(skips[0], &qFinal) \|\| qFinal.Code != riskbucket.RefusalExistingGuardianCap \|\|` (if 260) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |

Exact AST return positions: (없음)

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 15종 — 대표: `Error`, `a112ScopeRefusalOf`, `append`, `context.Background`, `dispatchStrategyMarketHandoffs`, `errors.As`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a146 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a146 편집」에 적은 한 자리.
- High-risk impact: yes(전략 첫 다리 진입 분류 — 주문을 늘리지 않음: 건너뛴 진입은 브로커 0, 매도 · 손절 · 청산 경로 무관)
