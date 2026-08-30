# Function Logic Map: `conditionalRecord`

- Source: `internal/verifylive/cleanup_test.go`
- Function: `internal/verifylive/cleanup_test.go:conditionalRecord`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-reopens-conditional-chain`

이 change의 테스트. 아래 분기·호출은 ast.json에서 읽었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 fixture | 테스트가 만든 값 | 이 파일 | t.Fatalf/t.Errorf로 실패를 보고한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — 선형 실행 | 없음 | 단일 반환 | `conditionalRecord` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `append` | ast.json calls (line 280) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |

라이브 바인딩 없음 — 브로커·계좌·네트워크를 직접 호출하지 않는다.

## State mutations and fallbacks

- 테스트 지역 상태만 변경한다. 프로덕션 상태·브로커·계좌를 건드리지 않는다.

## Safety conclusion

- Safe edit boundary: 테스트 함수이므로 프로덕션 동작 경계가 없다.
- High-risk impact: no — 테스트 코드. httptest/fake 브로커만 쓰며 실계좌에 접근하지 않는다.
