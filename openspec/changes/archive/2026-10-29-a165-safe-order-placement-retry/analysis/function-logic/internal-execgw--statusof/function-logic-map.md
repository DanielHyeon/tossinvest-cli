# Function Logic Map: `statusOf`

- Source: `internal/execgw/classify.go` (125-143)
- Qualified function: `statusOf`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 020dba811b666e94…)
- AST evidence: `ast.json` — AST branches 6, 반환 5, 호출 5
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

official 오류에서 HTTP status 를 복원. 429→`ErrRateLimited`(B4), 5xx→`ErrServer`(B5), 401/403→B3.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| err | official 오류 | Client.send | 모르면 (0,false) |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 128:3, 133:3, 137:3, 139:3, 141:3

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 127:2` | if errors.As(err, &apiErr) { | 아래 「State mutations」 | 소스 줄 127 이하 | 실행됨 — 패키지 스위트 count=1 |
| B2 | `switch at 130:2` | switch { | 아래 「State mutations」 | 소스 줄 130 이하 | 측정 불가(본문 블록 없음) |
| B3 | `case at 131:2` | case errors.Is(err, official.ErrAuth), errors.Is(err, official.ErrIPNotAllowed): | 아래 「State mutations」 | 소스 줄 131 이하 | 실행됨 — 패키지 스위트 count=1 |
| B4 | `case at 134:2` | case errors.Is(err, official.ErrRateLimited): | 아래 「State mutations」 | 소스 줄 134 이하 | 실행됨 — 패키지 스위트 count=1 |
| B5 | `case at 138:2` | case errors.Is(err, official.ErrServer): | 아래 「State mutations」 | 소스 줄 138 이하 | 실행됨 — 패키지 스위트 count=1 |
| B6 | `case at 140:2` | default: | 아래 「State mutations」 | 소스 줄 140 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

errors.As/Is 만.

AST 호출 이름 전체: `errors.As`, `errors.Is`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 편집 계획 없음.
