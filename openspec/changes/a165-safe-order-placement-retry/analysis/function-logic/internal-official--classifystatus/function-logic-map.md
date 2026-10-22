# Function Logic Map: `classifyStatus`

- Source: `internal/official/errors.go` (39-61)
- Qualified function: `classifyStatus`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 5f9cc4baf967a604…)
- AST evidence: `ast.json` — AST branches 6, 반환 5, 호출 5
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official 패키지 전체, journal 은 `-run 'Classify|SendTracker|Dispatch'`) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

비 2xx status 를 official 오류로. B5(56) `code >= 500` 은 **sentinel `ErrServer` 그 자체**(57)를 돌려준다 — a165 (b) 의 「5xx 를 실제로 받음」은 이 동일성(`err == official.ErrServer`)으로 판정한다(1.3 A2).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| code | HTTP status | doRequest | — |
| body | 응답 본문 | doRequest | 401/403 은 본문 대신 코드만 실음 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 51:4, 53:3, 55:3, 57:3, 59:3

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `switch at 40:2` | switch { | 아래 「State mutations」 | 소스 줄 40 이하 | 측정 불가(본문 블록 없음) |
| B2 | `case at 41:2` | case code == 401 \|\| code == 403: | 아래 「State mutations」 | 소스 줄 41 이하 | 실행됨 — 패키지 스위트 count=1 |
| B3 | `if at 50:3` | if reIPWord.Match(bytes.ToLower(body)) { | 아래 「State mutations」 | 소스 줄 50 이하 | 실행됨 — 패키지 스위트 count=1 |
| B4 | `case at 54:2` | case code == 429: | 아래 「State mutations」 | 소스 줄 54 이하 | 실행됨 — 패키지 스위트 count=1 |
| B5 | `case at 56:2` | case code >= 500: | 아래 「State mutations」 | 소스 줄 56 이하 | 실행됨 — 패키지 스위트 count=1 |
| B6 | `case at 58:2` | default: | 아래 「State mutations」 | 소스 줄 58 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`reIPWord.Match`, `fmt.Errorf`(401/403), `&APIError`(default).

AST 호출 이름 전체: `bytes.ToLower`, `fmt.Errorf`, `reIPWord.Match`, `string`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 편집 계획 없음 — 동일성 반환을 시험으로 고정(orders_raw_test.go:201 이 이미 `err == ErrServer` 를 단언).
