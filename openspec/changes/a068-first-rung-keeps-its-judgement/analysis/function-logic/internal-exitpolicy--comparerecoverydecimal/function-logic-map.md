# Function Logic Map: `compareRecoveryDecimal`

- Source: `internal/exitpolicy/recovery.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a068-first-rung-keeps-its-judgement/base-commit.txt`

**a062는 이 함수를 편집하지 않는다.** `compareRecoveryStage` 바로 앞에 있어 diff
hunk와 줄 범위가 인접했을 뿐이며, 본문은 base commit과 동일하다. 그럼에도 게이트가
요구하는 증거를 남긴다 — 이 함수가 `compareRecoveryStage`와 함께 복구 후보 선택의
세 축(protection, high-water, stage) 중 둘을 담당하기 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a string` | 양의 십진 문자열 (재계산 후보의 값) | `SelectRecoverySnapshot` | 파싱 실패나 비양수는 오류로 전파 |
| `b string` | 양의 십진 문자열 (저장 후보의 값) | 같음 | 같음 |

**불변식**: 값을 만들어 내지 않는다. 두 문자열을 십진으로 읽어 부호만 답한다.
반올림·절사·기본값 대입이 없다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `positive(a)`가 실패 | 없음 | `0, err` — 호출자가 격리로 이어감 | 기존 `recovery_validation_test.go` |
| B2 | `positive(b)`가 실패 | 없음 | `0, err` | 같음 |
| (없음) | 둘 다 성공 | 없음 | `ar.Cmp(br), nil` | a068 2.1, 2.2, 2.3 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `positive` | 양의 십진 파싱 | 순수, 실패는 오류 반환 | AST |
| `(*big.Rat).Cmp` | 부호 비교 | 순수 | AST |

I/O 없음. 시계 없음. 전역 상태 없음.

## State mutations and fallbacks

- 아무것도 변경하지 않는다.
- fallback 없음. 읽을 수 없는 값은 오류이고, 호출자가 fail-closed로 처리한다.

## Safety conclusion

- Safe edit boundary: a062는 이 함수를 편집하지 않았다 (본문 base 동일).
- High-risk impact: 이 함수 자체는 보호선 비교의 일부이므로 High-risk 경로에 속하지만,
  a062의 변경 범위 밖이다.
- §0.9: 값을 만들지 않으므로 기준선을 낮출 경로가 없다.
