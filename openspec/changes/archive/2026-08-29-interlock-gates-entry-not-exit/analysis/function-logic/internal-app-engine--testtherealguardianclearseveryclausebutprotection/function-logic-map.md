# Function Logic Map: `TestTheRealGuardianClearsEveryClauseButProtection`

- Source: `internal/app/engine/guardian_test.go`
- AST evidence: `ast.json` (revision: base — 이 change에서 삭제된 함수)
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

삭제됨 — TestTheRealGuardianClearsEveryClause로 대체. 조항 6 거부를 '앞 절이 모두 통과했다'의 증거로 쓰던 단언이었고, 그 거부가 없어졌으므로 같은 증거를 '기동이 성공한다'로 다시 쓴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 픽스처 | 함수 본문이 세운다 | 테스트 자신 | 단언 실패 시 t.Error/t.Fatal |

## Branches and early returns

분기는 전부 단언 가드다. 각 가드는 실패 시 이 테스트 자신이 보고한다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` @ internal/app/engine/guardian_test.go:101 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ internal/app/engine/guardian_test.go:104 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `if` @ internal/app/engine/guardian_test.go:107 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B4 | `if` @ internal/app/engine/guardian_test.go:114 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B5 | `range` @ internal/app/engine/guardian_test.go:117 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B6 | `if` @ internal/app/engine/guardian_test.go:127 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 테스트 대상 API | 이 테스트가 무엇을 보는지 | 단언으로 처리 | AST |

## State mutations and fallbacks

- 격리된 임시 디렉터리와 임시 journal 외에는 없다.

## Safety conclusion

- Safe edit boundary: 테스트 함수. 프로덕션 동작을 바꾸지 않는다.
- High-risk impact: no
