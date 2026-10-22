# Function Logic Map: `unwrapAndDecode`

- Source: `internal/official/client.go` (213-228)
- Qualified function: `unwrapAndDecode`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 6d145916a6797f3a…)
- AST evidence: `ast.json` — AST branches 4, 반환 5, 호출 5
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official 패키지 전체, journal 은 `-run 'Classify|SendTracker|Dispatch'`) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

2xx 본문의 envelope 해석. 실패는 B2(218)·B3(221)·B4(224) 모두 **감싼** `ErrServer`(219·222·225) — 2xx 로 접수된 응답이다. a165 (b) 에서 제외(동일성 판정이 거름).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| body | 2xx 본문 | send | — |
| out | nil 가능 | 호출자 | B1 nil 이면 버림 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 215:3, 219:3, 222:3, 225:3, 227:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 214:2` | if out == nil { | 아래 「State mutations」 | 소스 줄 214 이하 | 실행됨 — 패키지 스위트 count=1 |
| B2 | `if at 218:2` | if err := json.Unmarshal(body, &env); err != nil { | 아래 「State mutations」 | 소스 줄 218 이하 | 실행됨 — 패키지 스위트 count=1 |
| B3 | `if at 221:2` | if env.Result == nil { | 아래 「State mutations」 | 소스 줄 221 이하 | 미실행 — 패키지 스위트 count=0 |
| B4 | `if at 224:2` | if err := json.Unmarshal(env.Result, out); err != nil { | 아래 「State mutations」 | 소스 줄 224 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`json.Unmarshal`, `fmt.Errorf`.

AST 호출 이름 전체: `fmt.Errorf`, `json.Unmarshal`.

## State mutations and fallbacks

- out 채움.

## Safety conclusion

- 편집 계획 없음.
