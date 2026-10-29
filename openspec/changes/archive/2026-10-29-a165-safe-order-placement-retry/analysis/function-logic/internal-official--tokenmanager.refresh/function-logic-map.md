# Function Logic Map: `tokenManager.refresh`

- Source: `internal/official/token.go` (108-126)
- Qualified function: `tokenManager.refresh`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 4cadb04ff6d6691a…)
- AST evidence: `ast.json` — AST branches 1, 반환 2, 호출 6
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

401 뒤 토큰. B1(119) 캐시 토큰이 유효하고 거절된 것과 다르면 채택(adopted=true, 121), 아니면 교환(125, adopted=false). `Client.send` B8 이 adopted=false 면 break — 401 재전송 횟수의 근거.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| refused | 거절된 토큰 | send | — |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 121:3, 125:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 119:2` | if ct, err := m.loadCache(); err == nil && isStillValid(ct, now) && ct.AccessToken != refused { | 아래 「State mutations」 | 소스 줄 119 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`m.loadCache`, `isStillValid`, `m.exchange`.

AST 호출 이름 전체: `isStillValid`, `m.exchange`, `m.loadCache`, `m.mu.Lock`, `m.mu.Unlock`, `time.Now`.

## State mutations and fallbacks

- 토큰 캐시.

## Safety conclusion

- 편집 계획 없음 — D9 시험의 근거.
