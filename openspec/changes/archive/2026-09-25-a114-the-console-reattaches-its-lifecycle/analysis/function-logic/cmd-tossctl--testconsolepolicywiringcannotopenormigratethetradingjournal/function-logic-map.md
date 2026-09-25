# Function Logic Map: `TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json` — 구현 후(분기 4; base 와 분기 동일)
- Risk scan: `risk-pattern-report.md`

a114 편집: 같은 이유로 두 파일을 함께 읽는다 — 금지 문자열(journal 쓰기 능력)도 wrapper 파일까지 본다(범위가 넓어진 방향).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 읽는 소스 | `console.go` + (a114) `console_lifecycle_attach.go` | 디스크 | `readSource` 실패면 `t.Fatalf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 금지 문자열 순회 | 없음 | `t.Error` | 자기 자신 |
| B2 | 금지 문자열 발견 | 없음 | `t.Error` | 자기 자신 |
| B3 | 필수 좁은 client 문자열 순회 | 없음 | `t.Error` | 자기 자신 |
| B4 | 필수 문자열 없음 | 없음 | `t.Error` | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readSource` | 소스 텍스트 | `t.Fatalf` | AST call |

## State mutations and fallbacks

- 없음(읽기 전용 소스 핀).

## Safety conclusion

- Safe edit boundary: 읽는 소스에 wrapper 파일을 더한 한 줄. 검사 목록·금지 목록 무변경.
- High-risk impact: no — 테스트.
