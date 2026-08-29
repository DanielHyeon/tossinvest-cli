# Function Logic Map: `Steps`

- Source: `internal/verifylive/verifylive.go`
- Function: `internal/verifylive/verifylive.go:Steps`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-plans-the-object-it-mutates`

절차 자체를 데이터로 돌려준다. 이 change는 `conditional-modify`·`conditional-cancel` 두 항목에 `ActsOnConditional: true`를 선언한다 — 두 단계 본문이 이미 하고 있던 것을 카탈로그가 말하게 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (없음) | 인자 없는 생성자 | 이 파일 | 반환값은 매 호출 새 slice다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — 선형 실행 | 없음 | 단일 반환 | `Steps` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 없음 | 호출 없음 | — | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 없음 — 리터럴을 만들어 돌려준다. 호출자가 받은 slice를 바꿔도 다음 호출에 영향이 없다.

## Safety conclusion

- Safe edit boundary: 두 항목의 필드 추가뿐이다. 순서·ID·Mutations·DependsOn은 무변경이며, 계획 digest 고정 테스트가 그것을 증명한다.
- High-risk impact: yes — 카탈로그가 승인 목록의 원천이다. 다만 이 편집은 선언 2건이고 새 mutation은 없다.
