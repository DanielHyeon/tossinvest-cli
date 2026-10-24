# Function Logic Map: `buildFlattenWiring`

- Source: `cmd/tossctl/flatten.go` (199-272)
- Qualified function: `buildFlattenWiring`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 3a2b20f4c145d7f1…)
- AST evidence: `ast.json` — AST branches 8, 반환 6, 호출 15
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

CLI `tossctl flatten` 배선(execgw.New 233, `flatten.Saga` 247 — 생산 유일한 Saga 생성 자리). a165 는 `CancelRefusal: execgw.FlattenCancelRefusal()` 를 더한다 — 취소 409 재시도(누적 5s·2회), 소진 시 오늘 경로.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| root, opts | CLI 옵션 | cobra | B1·B2·B5·B7·B8 오류 반환 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 202:3, 207:3, 217:3, 228:3, 243:3, 265:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 201:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 201 이하 | 미측정 — 이 패키지 coverprofile 없음 |
| B2 | `if at 206:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 206 이하 | 미측정 — 이 패키지 coverprofile 없음 |
| B3 | `range at 210:2` | for _, a := range accounts { | 아래 「State mutations」 | 소스 줄 210 이하 | 미측정 — 이 패키지 coverprofile 없음 |
| B4 | `if at 211:3` | if strings.TrimSpace(a.DisplayName) != "" { | 아래 「State mutations」 | 소스 줄 211 이하 | 미측정 — 이 패키지 coverprofile 없음 |
| B5 | `if at 216:2` | if accountRef == "" { | 아래 「State mutations」 | 소스 줄 216 이하 | 미측정 — 이 패키지 coverprofile 없음 |
| B6 | `if at 221:2` | if root.configDir != "" { | 아래 「State mutations」 | 소스 줄 221 이하 | 미측정 — 이 패키지 coverprofile 없음 |
| B7 | `if at 227:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 227 이하 | 미측정 — 이 패키지 coverprofile 없음 |
| B8 | `if at 241:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 241 이하 | 미측정 — 이 패키지 coverprofile 없음 |

## Calls and live bindings

`execgw.New`, `journal.Open` 등.

AST 호출 이름 전체: `clock.System`, `engine.NewOrderPath`, `errors.New`, `execgw.New`, `execgw.NewEntryGate`, `filepath.Join`, `fmt.Errorf`, `j.Close`, `journal.Open`, `obs.NewLogger`, `orderPath.Official.Accounts`, `os.Getenv`, `strings.TrimSpace`.

## State mutations and fallbacks

- CLI 배선.

## Safety conclusion

- 필드 1개 추가.
