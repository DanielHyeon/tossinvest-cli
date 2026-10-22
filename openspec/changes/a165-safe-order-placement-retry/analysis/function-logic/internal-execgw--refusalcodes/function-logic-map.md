# Function Logic Map: `refusalCodes`

- Source: `internal/execgw/refusal_code.go` (91-106)
- Qualified function: `refusalCodes`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 4e885e54a90205d2…)
- AST evidence: `ast.json` — AST branches 3, 반환 3, 호출 6
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official 패키지 전체, journal 은 `-run 'Classify|SendTracker|Dispatch'`) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

오류 본문의 최상위 `code` 와 `error.code` 를 읽는다. 취소 409 `already-processing` 판정은 이 한 곳을 쓴다(1.3 A9) — 두 자리 모순이면 재시도·확정 분류 없음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| body | APIError.Body | classifyStatus default | JSON 아님이면 ok=false(B1) |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 94:3, 101:4, 105:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 93:2` | if err := json.Unmarshal([]byte(body), &root); err != nil \|\| root == nil { | 아래 「State mutations」 | 소스 줄 93 이하 | 실행됨 — 패키지 스위트 count=1 |
| B2 | `if at 97:2` | if raw, has := root["error"]; has && !isJSONNull(raw) { | 아래 「State mutations」 | 소스 줄 97 이하 | 실행됨 — 패키지 스위트 count=1 |
| B3 | `if at 99:3` | if err := json.Unmarshal(raw, &inner); err != nil \|\| inner == nil { | 아래 「State mutations」 | 소스 줄 99 이하 | 미실행 — 패키지 스위트 count=0 |

## Calls and live bindings

`json.Unmarshal`, `readCode`, `isJSONNull`.

AST 호출 이름 전체: `isJSONNull`, `json.Unmarshal`, `readCode`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 편집 계획 없음.
