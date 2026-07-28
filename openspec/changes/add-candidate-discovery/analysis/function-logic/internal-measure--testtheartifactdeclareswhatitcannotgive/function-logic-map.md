# Function Logic Map: `TestTheArtifactDeclaresWhatItCannotGive`

- Source: `internal/measure/counterfactual_test.go`
- AST evidence: `ast.json` (revision=current, L211–250, 분기 6개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — 본문 변경: 주석 열 정렬만 (gofmt, 커밋 dcc2030) — 코드 토큰 동일 (revision=current)

**본문 코드는 base와 동일하다.** 바뀐 것은 문자열 슬라이스의 줄 끝 주석 정렬뿐이다 — 커밋 dcc2030의 `gofmt`가 CJK 폭 때문에 어긋나 있던 열을 다시 맞췄다. 주석 제거 후 토큰 열이 완전히 일치한다(base `583772c4` L211–250 대 현재 L211–250).

테스트 자체는 산출물이 **주지 못하는 것을 먼저 선언**하는지 본다: 경계 지도는 공급하고, 실거래 모집단에서만 나오는 것은 조건부이며, 수수료·세율은 미검증이고, 마감 포지션이 필요한 것은 공급하지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `measure.Grid(...)` 렌더 결과 | 문자열 | counterfactual 렌더러 | 선언이 표보다 뒤에 오면 FAIL |
| 필수 문자열 7종 | 고정 목록 | 이 테스트 | 하나라도 빠지면 FAIL |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 그리드 산출 에러 | — | `t.Fatalf` | 이 테스트 |
| B2 | 선언이 표보다 뒤 또는 부재 | — | `t.Fatalf` | 동일 |
| B3 | 필수 문자열 순회 | — | — | 동일 |
| B4 | 필수 문자열 부재 | — | `t.Errorf` | 동일 |
| B5 | `LeftTruncationNote` 부재 | — | `t.Error` | 동일 |
| B6 | `StopWidthCircularityNote` 부재 | — | `t.Error` | 동일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `measure.Grid` / `Render` | 산출물 생성 | — | ast.json calls |
| `strings.Index` / `Contains` | 선언 위치와 존재 검사 | — | 동일 |

## State mutations and fallbacks

- 테스트 — 파일도 계좌도 건드리지 않는다.

## Safety conclusion

- Safe edit boundary: 주석 정렬만 — 코드 무변경.
- High-risk impact: no (측정 산출물의 정직성 검사 — 주문·Guardian·원장 경로 무접촉).
