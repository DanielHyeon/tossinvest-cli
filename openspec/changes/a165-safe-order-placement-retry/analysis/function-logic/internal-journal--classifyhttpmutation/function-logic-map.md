# Function Logic Map: `ClassifyHTTPMutation`

- Source: `internal/journal/dispatch.go` (301-345)
- Qualified function: `ClassifyHTTPMutation`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` f77e0bab6044ecb1…)
- AST evidence: `ast.json` — AST branches 7, 반환 6, 호출 10
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

journal 의 비관적 상태 분류. 429·5xx·그 밖 상태는 default(B7) AMBIGUOUS, 400/401/403/404/405/415/422 는 B5 REJECTED, 전송 오류는 B2(NotStarted→NOT_SENT) 아니면 AMBIGUOUS.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| send | SendState | tracker | — |
| statusCode | HTTP | statusOf | 0 이면 B6 |
| err | 전송 오류 | — | — |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 304:4, 311:3, 322:3, 324:3, 331:3, 338:3

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 302:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 302 이하 | 실행됨 — 패키지 스위트 count=1 |
| B2 | `if at 303:3` | if send == SendNotStarted { | 아래 「State mutations」 | 소스 줄 303 이하 | 실행됨 — 패키지 스위트 count=1 |
| B3 | `switch at 320:2` | switch { | 아래 「State mutations」 | 소스 줄 320 이하 | 측정 불가(본문 블록 없음) |
| B4 | `case at 321:2` | case statusCode >= 200 && statusCode < 300: | 아래 「State mutations」 | 소스 줄 321 이하 | 실행됨 — 패키지 스위트 count=1 |
| B5 | `case at 323:2` | case isDefinitiveRejection(statusCode): | 아래 「State mutations」 | 소스 줄 323 이하 | 실행됨 — 패키지 스위트 count=1 |
| B6 | `case at 330:2` | case statusCode == 0: | 아래 「State mutations」 | 소스 줄 330 이하 | 실행됨 — 패키지 스위트 count=1 |
| B7 | `case at 337:2` | default: | 아래 「State mutations」 | 소스 줄 337 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

isDefinitiveRejection 만.

AST 호출 이름 전체: `err.Error`, `errors.New`, `fmt.Errorf`, `fmt.Sprintf`, `isDefinitiveRejection`, `send.String`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 편집 계획 없음.
