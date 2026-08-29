# Function Logic Map: `TestTokenRefresh`

- Source: `internal/official/token_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change가 **고치는 대상이 아니라 부딪힌 것**이다. `refresh()`의
시그니처가 바뀌면서 호출부가 컴파일되지 않는다. **판정은 바꾸지 않는다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| httptest 서버 | 언제나 `AT2`를 준다 | 같은 함수 | — |
| `hits` | 교환 횟수 | 핸들러 | — |
| 불변식 | **`refresh`는 교환한다** (`hits == 2`) | 이 테스트의 원래 주장 | 이것이 바뀌면 단일 프로세스 의미론이 바뀐 것이다 |

편집 후 넘기는 `refused`는 방금 얻은 `"AT2"`이고, 디스크에도 같은 `AT2`뿐이다.
따라서 채택 후보가 없어 fallthrough로 가고 **교환한다** — `hits == 2`가 그대로
성립한다. 이 change의 좁음이 여기서 증명된다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `refresh` 오류이거나 토큰이 `AT2`가 아니다 | 없음 | `t.Fatalf` | 자기 자신 |
| B2 | **신규** — `adopted`가 true다 | 없음 | `t.Fatal` | 자기 자신 |
| B3 | `hits != 2` | 없음 | `t.Fatalf` | 자기 자신 + 변이 M4 |

B2를 더한 이유: 시그니처가 늘어난 반환값을 버리면 그 값이 무엇이든 테스트가
통과한다. 이 시나리오에서 채택은 **일어나서는 안 되는 일**이므로 단언한다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newTokenManager` | 대상 | — | AST calls |
| `m.token` | 최초 교환 | — | AST calls |
| `m.refresh` | 판정 대상. **인자 하나·반환 하나가 늘었다** | — | AST calls |

## State mutations and fallbacks

- 상태를 바꾸지 않는다. httptest 서버 하나와 임시 디렉터리.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: **호출부와 새 반환값 단언뿐.** 서버, `hits` 기대값,
  토큰 기대값은 글자 그대로 보존한다.
- 이 편집은 판정력을 **약화하지 않는다** — 단언이 하나 늘었다.
- High-risk impact: **no.** 테스트 전용.
