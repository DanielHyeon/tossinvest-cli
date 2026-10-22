# Function Logic Map: `newMatcher`

- Source: `internal/execgw/indoubt.go` (652-679)
- Qualified function: `newMatcher`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` d220e7df90d725bf…)
- AST evidence: `ast.json` — AST branches 3, 반환 2, 호출 11
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official 패키지 전체, journal 은 `-run 'Classify|SendTracker|Dispatch'`) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

intent·attempt 로 매처. B3(673) 디스패치 시각 ±SubmitWindow(기본 5분) 창. a165 조회는 첫 전송 시각을 stamp 로 준다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| intent | journal.Intent | Prepare | 수량 파싱 실패 B1 |
| rec | AttemptRecord | — | 시각 없으면 창 미사용 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 655:3, 678:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 654:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 654 이하 | 미실행 — 패키지 스위트 count=0 |
| B2 | `if at 670:2` | if stamp == "" { | 아래 「State mutations」 | 소스 줄 670 이하 | 실행됨 — 패키지 스위트 count=1 |
| B3 | `if at 673:2` | if dispatched, err := time.Parse(time.RFC3339, stamp); err == nil { | 아래 「State mutations」 | 소스 줄 673 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`strconv.ParseFloat`, `parseOptionalDecimal`, `time.Parse`.

AST 호출 이름 전체: `dispatched.Add`, `fmt.Errorf`, `parseOptionalDecimal`, `strconv.ParseFloat`, `strings.ToUpper`, `strings.TrimSpace`, `time.Parse`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 편집 계획 없음.
