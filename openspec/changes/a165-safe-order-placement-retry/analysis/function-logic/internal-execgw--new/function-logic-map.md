# Function Logic Map: `New`

- Source: `internal/execgw/gateway.go` (151-193)
- Qualified function: `New`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` c304e32a62e42759…; GREEN 뒤 AST 재추출 2026-10-29)
- AST evidence: `ast.json` — AST branches 8, 반환 4, 호출 7
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

게이트웨이 생성자. a165 는 `Options.CancelRefusal` 을 `g.cancelRefusal` 로 옮기는 대입 한 줄만 더한다(분기 추가 없음).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| opts | Options | 배선 | B2–B4 필수 필드 없으면 오류 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 154:3, 156:3, 158:3, 192:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `switch at 152:2` | switch { | 아래 「State mutations」 | 소스 줄 149 이하 | 측정 불가(본문 블록 없음) |
| B2 | `case at 153:2` | case opts.Journal == nil: | 아래 「State mutations」 | 소스 줄 150 이하 | 미실행 — 패키지 스위트 count=0 |
| B3 | `case at 155:2` | case opts.Trading == nil: | 아래 「State mutations」 | 소스 줄 152 이하 | 미실행 — 패키지 스위트 count=0 |
| B4 | `case at 157:2` | case strings.TrimSpace(opts.AccountRef) == "": | 아래 「State mutations」 | 소스 줄 154 이하 | 미실행 — 패키지 스위트 count=0 |
| B5 | `if at 161:2` | if protectionCheckForTest == nil && !opts.forceReadinessAdapterForTest { | 아래 「State mutations」 | 소스 줄 158 이하 | 실행됨 — 패키지 스위트 count=1 |
| B6 | `if at 183:2` | if g.clk == nil { | 아래 「State mutations」 | 소스 줄 179 이하 | 미실행 — 패키지 스위트 count=0 |
| B7 | `if at 186:2` | if g.source == "" { | 아래 「State mutations」 | 소스 줄 182 이하 | 미실행 — 패키지 스위트 count=0 |
| B8 | `if at 189:2` | if g.newID == nil { | 아래 「State mutations」 | 소스 줄 185 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`strings.TrimSpace`, `clock.System`.

AST 호출 이름 전체: `clock.System`, `errors.New`, `strings.TrimSpace`.

## State mutations and fallbacks

- Gateway 구성.

## Safety conclusion

- High-risk 경로의 구성만 — 대입 1줄.
