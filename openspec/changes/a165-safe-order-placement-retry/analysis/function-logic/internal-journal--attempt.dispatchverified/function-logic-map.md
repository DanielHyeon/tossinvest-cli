# Function Logic Map: `Attempt.DispatchVerified`

- Source: `internal/journal/dispatch.go` (132-223)
- Qualified function: `Attempt.DispatchVerified`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 42165f21f0707b4d…)
- AST evidence: `ast.json` — AST branches 19, 반환 11, 호출 19
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

dispatch 1회 → 정착. ACK 이면 B11 MarkAcked(181) 뒤 B12·B13 verify(184–185) 실패 시 MarkInDoubt(191) — A1 의 목록 대조가 들어갈 자리. 클래스 REJECTED 는 B6 FAILED_CONFIRMED.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| dispatch | DispatchFunc | submit | nil 이면 B1 |
| verify | ExistenceCheck | roundTripFor | nil 이면 생략 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 134:3, 141:3, 154:4, 161:4, 176:5, 182:4, 192:6, 200:4, 207:4, 216:4, 222:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 133:2` | if dispatch == nil { | 아래 「State mutations」 | 소스 줄 133 이하 | 미실행 — 패키지 스위트 count=0 |
| B2 | `if at 140:2` | if err := a.MarkDispatchStarted(ctx); err != nil { | 아래 「State mutations」 | 소스 줄 140 이하 | 실행됨 — 패키지 스위트 count=1 |
| B3 | `switch at 150:2` | switch outcome.Class { | 아래 「State mutations」 | 소스 줄 150 이하 | 측정 불가(본문 블록 없음) |
| B4 | `case at 151:2` | case DispatchNotSent: | 아래 「State mutations」 | 소스 줄 151 이하 | 미실행 — 패키지 스위트 count=0 |
| B5 | `if at 153:3` | if err := a.Settle(ctx, StateNotDispatched, res.ReasonCode, outcome.Detail); err != nil { | 아래 「State mutations」 | 소스 줄 153 이하 | 미실행 — 패키지 스위트 count=0 |
| B6 | `case at 158:2` | case DispatchRejected: | 아래 「State mutations」 | 소스 줄 158 이하 | 미실행 — 패키지 스위트 count=0 |
| B7 | `if at 160:3` | if err := a.Settle(ctx, StateFailedConfirmed, res.ReasonCode, outcome.Detail); err != nil { | 아래 「State mutations」 | 소스 줄 160 이하 | 미실행 — 패키지 스위트 count=0 |
| B8 | `case at 165:2` | case DispatchAcked: | 아래 「State mutations」 | 소스 줄 165 이하 | 실행됨 — 패키지 스위트 count=1 |
| B9 | `if at 170:3` | if strings.TrimSpace(outcome.BrokerOrderID) == "" { | 아래 「State mutations」 | 소스 줄 170 이하 | 미실행 — 패키지 스위트 count=0 |
| B10 | `if at 174:4` | if err := a.MarkInDoubt(ctx, res.ReasonCode, | 아래 「State mutations」 | 소스 줄 174 이하 | 측정 불가(본문 블록 없음) |
| B11 | `if at 181:3` | if err := a.MarkAcked(ctx, outcome.BrokerOrderID); err != nil { | 아래 「State mutations」 | 소스 줄 181 이하 | 미실행 — 패키지 스위트 count=0 |
| B12 | `if at 184:3` | if verify != nil { | 아래 「State mutations」 | 소스 줄 184 이하 | 실행됨 — 패키지 스위트 count=1 |
| B13 | `if at 185:4` | if verifyErr := verify(ctx, outcome.BrokerOrderID); verifyErr != nil { | 아래 「State mutations」 | 소스 줄 185 이하 | 실행됨 — 패키지 스위트 count=1 |
| B14 | `if at 191:5` | if err := a.MarkInDoubt(ctx, res.ReasonCode, res.Detail); err != nil { | 아래 「State mutations」 | 소스 줄 191 이하 | 미실행 — 패키지 스위트 count=0 |
| B15 | `if at 199:3` | if err := a.Settle(ctx, StateConfirmed, res.ReasonCode, outcome.Detail); err != nil { | 아래 「State mutations」 | 소스 줄 199 이하 | 미실행 — 패키지 스위트 count=0 |
| B16 | `case at 204:2` | case DispatchAmbiguous: | 아래 「State mutations」 | 소스 줄 204 이하 | 미실행 — 패키지 스위트 count=0 |
| B17 | `if at 206:3` | if err := a.MarkInDoubt(ctx, res.ReasonCode, outcome.Detail); err != nil { | 아래 「State mutations」 | 소스 줄 206 이하 | 미실행 — 패키지 스위트 count=0 |
| B18 | `case at 211:2` | default: | 아래 「State mutations」 | 소스 줄 211 이하 | 실행됨 — 패키지 스위트 count=1 |
| B19 | `if at 215:3` | if err := a.MarkInDoubt(ctx, res.ReasonCode, detail); err != nil { | 아래 「State mutations」 | 소스 줄 215 이하 | 미실행 — 패키지 스위트 count=0 |

## Calls and live bindings

`MarkDispatchStarted`, `Settle`, `MarkAcked`, `MarkInDoubt`.

AST 호출 이름 전체: `a.MarkAcked`, `a.MarkDispatchStarted`, `a.MarkInDoubt`, `a.Settle`, `dispatch`, `firstNonEmpty`, `fmt.Errorf`, `fmt.Sprintf`, `strings.TrimSpace`, `verify`.

## State mutations and fallbacks

- attempt 상태 전이(journal).

## Safety conclusion

- 편집 계획 없음.
