# Function Logic Map: `resolveDeclared`

- Source: `internal/console/static_test.go`
- Change: `add-candidate-discovery`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

패키지가 선언한 타입 이름의 사슬을 **더 이상 이름이 아닌 선언**까지 따라간다. 한 홉으로는 부족했다 — `type Ticker = Wide`와 `type A B; type B C` 둘 다 메서드 집합을 이름 두 개 너머에 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `expr` | 타입 표현 | 필드 선언 | `*ast.Ident`가 아니면 그대로 돌려준다 — 한정 타입·제네릭 인스턴스는 여기서 해석되지 않는다 |
| `declaredTypes` | 패키지 선언 타입 색인 | `packageTypes` | 없는 이름은 그대로 돌려준다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for {` | 없음 | 없음 | 인터페이스 seam 6 + 별칭 사슬 변이 |
| B2 | `ident, ok := expr.(*ast.Ident); !ok` | 없음 | `expr` 반환 | func 타입·구조체 등 |
| B3 | `next, ok := declaredTypes[ident.Name]; !ok || seen[ident.Name]` | 없음 | `expr` 반환 | 순환 별칭 변이 — 무한 루프 대신 이름을 그대로 돌려준다 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | 맵 조회와 타입 단언뿐 | 순수 | ast.json calls=null |

## State mutations and fallbacks

- 없음(순수 함수). `seen`이 순환을 끊는다.

## Safety conclusion

- Safe edit boundary: 신설. 이전에는 호출 지점에서 한 홉만 따라갔다.
- High-risk impact: yes (메서드 집합 검사의 전제 — 사슬 중간에서 멈추면 인터페이스가 아닌 이름을 보고 검사가 끝난다)
