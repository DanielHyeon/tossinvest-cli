# Function Logic Map: `ChainOf`

- Source: `internal/verifylive/record.go`
- Function: `internal/verifylive/record.go:ChainOf`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-holds-what-it-awaits`

한 객체가 어느 **측정 사슬**에 속하는지 기록에서 찾는다. 가장 나중에 사슬을 명시한 줄이 이긴다. 필요한 이유는 실측이다 — 조건주문 정정은 새 식별자를 발급하고 옛 식별자를 즉시 404로 만든다(measurements.md M19·M40, 양 시장 동일). 그 뒤로는 **브로커가 둘이 한 보호였다고 말해줄 수 없고** 기록만이 말할 수 있다. 이 change 이전에는 그 연결이 산문 `Note`에만 있었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| entries []Entry | 훑을 기록 | prior·written·현재 단계 | 빈 사슬이면 "" |
| kind, id string | 찾을 객체 | 호출부 | 못 찾으면 "" — 호출부가 새 사슬을 만든다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 473) — `for _, e := range entries {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `range` (line 474) — `for _, a := range e.Artifacts {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B3 | `if` (line 475) — `if a.Kind == kind && a.ID == id && a.ChainID != "" {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 없음 | 호출 없음 | — | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 없음 — 문자열 하나를 돌려준다.

## Safety conclusion

- Safe edit boundary: 신규 leaf. 빈 문자열이 '이 기록은 그룹을 말하지 않는다'는 뜻이고 그것이 2026-07-30 이전 모든 줄이다.
- High-risk impact: no — 사슬은 정리 판정에 쓰이지 않는다(review.md A4). 재구성용 기록이다.
