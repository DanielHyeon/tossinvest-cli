# Function Logic Map: `TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json` — 구현 후(분기 3; base 와 분기 동일)
- Risk scan: `risk-pattern-report.md`

a114 편집: 읽는 소스를 `console.go` 에서 `console.go + console_lifecycle_attach.go` 로 넓혔다. lifecycle dial 이 wrapper 파일로 옮겨 갔기 때문이다 — 핀의 의도(콘솔은 좁은 engine client 로만 닿고 journal 을 열지 않는다)는 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 읽는 소스 | `console.go` + (a114) `console_lifecycle_attach.go` | 디스크 | `readSource` 실패면 `t.Fatalf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 필수 문자열 5개 순회 | 없음 | `t.Error` | 자기 자신 |
| B2 | 필수 문자열 없음 | 없음 | `t.Error` | 자기 자신 |
| B3 | `journal.Open(` 금지 | 없음 | `t.Error` | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readSource` | 소스 텍스트 | `t.Fatalf` | AST call |

## State mutations and fallbacks

- 없음(읽기 전용 소스 핀).

## Safety conclusion

- Safe edit boundary: 읽는 소스에 wrapper 파일을 더한 한 줄. 검사 목록·금지 목록 무변경.
- High-risk impact: no — 테스트.
