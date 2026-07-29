# Function Logic Map: `normaliseEndpoint`

- Source: `internal/soak/attest.go`
- Function: `internal/soak/attest.go:normaliseEndpoint`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `attest-covers-supervised-mutations`

이 change가 추가한 leaf. 비교용 철자 정규화 — 저장에는 쓰지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| e string | `METHOD /path` | 호출자 | 빈 문자열은 빈 문자열 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — 선형 실행 | 없음 | 단일 반환 | `normaliseEndpoint` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.Join` | ast.json calls (line 328) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `strings.Fields` | ast.json calls (line 328) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `strings.ToUpper` | ast.json calls (line 328) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `strings.TrimSpace` | ast.json calls (line 328) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |

라이브 바인딩 없음 — 브로커·계좌·네트워크를 직접 호출하지 않는다.

## State mutations and fallbacks

- 없음 — 순수 함수.

## Safety conclusion

- Safe edit boundary: 신규 함수.
- High-risk impact: no
