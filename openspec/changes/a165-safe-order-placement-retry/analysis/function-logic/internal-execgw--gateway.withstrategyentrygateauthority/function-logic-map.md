# Function Logic Map: `Gateway.withStrategyEntryGateAuthority`

- Source: `internal/execgw/strategy_entry_gate_authority.go` (46-81)
- Qualified function: `Gateway.withStrategyEntryGateAuthority`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 0d91fbdd3bcf5df5…)
- AST evidence: `ast.json` — AST branches 8, 반환 8, 호출 18
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official 패키지 전체, journal 은 `-run 'Classify|SendTracker|Dispatch'`) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

strategy first-leg 의 최종 진입 게이트. `g.entry.mu` 를 잡은 채(63 defer Unlock) `fn()` 을 80 에서 **한 번** 부른다 — 브로커 호출 동안 EntryGate 잠금이 잡혀 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| expected | 관찰된 게이트 권한 | ObserveStrategyEntryGate | 다르면 거절 |
| fn | submit 의 call 클로저 | submit | nil 이면 거절 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 50:3, 58:4, 67:4, 69:3, 73:4, 75:3, 78:3, 80:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 49:2` | if g == nil \|\| g.entry == nil \|\| fn == nil \|\| expected.generation == 0 \|\| expected.digest == "" { | 아래 「State mutations」 | 소스 줄 49 이하 | 미실행 — 패키지 스위트 count=0 |
| B2 | `if at 56:2` | if refresh != nil { | 아래 「State mutations」 | 소스 줄 56 이하 | 미실행 — 패키지 스위트 count=0 |
| B3 | `if at 57:3` | if err := refresh(); err != nil { | 아래 「State mutations」 | 소스 줄 57 이하 | 미실행 — 패키지 스위트 count=0 |
| B4 | `if at 64:2` | if current != expected { | 아래 「State mutations」 | 소스 줄 64 이하 | 실행됨 — 패키지 스위트 count=1 |
| B5 | `if at 66:3` | if marked := g.entry.dailyLossOnlyRejectionLocked(market, symbol); marked != nil { | 아래 「State mutations」 | 소스 줄 66 이하 | 실행됨 — 패키지 스위트 count=1 |
| B6 | `if at 71:2` | if rejected := g.entry.checkAccountEntryLocked(); rejected != nil { | 아래 「State mutations」 | 소스 줄 71 이하 | 미실행 — 패키지 스위트 count=0 |
| B7 | `if at 72:3` | if marked := g.entry.dailyLossOnlyRejectionLocked(market, symbol); marked != nil && marked.Detail == rejected.Detail { | 아래 「State mutations」 | 소스 줄 72 이하 | 미실행 — 패키지 스위트 count=0 |
| B8 | `if at 77:2` | if rejected := g.entry.checkSymbolEntryLocked(market, symbol); rejected != nil { | 아래 「State mutations」 | 소스 줄 77 이하 | 미실행 — 패키지 스위트 count=0 |

## Calls and live bindings

fn 은 반환문(80)에서 1회. 루프 없음.

AST 호출 이름 전체: `errors.New`, `fmt.Errorf`, `fn`, `g.entry.checkAccountEntryLocked`, `g.entry.checkSymbolEntryLocked`, `g.entry.dailyLossOnlyRejectionLocked`, `g.entry.mu.Lock`, `g.entry.mu.Unlock`, `refresh`, `sealStrategyEntryGateAuthority`, `strings.ToLower`, `strings.ToUpper`, `strings.TrimSpace`.

## State mutations and fallbacks

- 잠금 획득/해제만.

## Safety conclusion

- 편집 계획 없음. 재시도의 대기를 이 잠금 안에 두면 EntryGate 를 쓰는 다른 경로(Block 등)가 최대 대기만큼 멈춘다 — design D1 은 재전송마다 이 함수를 새로 통과하고 대기는 잠금 밖에 둔다.
