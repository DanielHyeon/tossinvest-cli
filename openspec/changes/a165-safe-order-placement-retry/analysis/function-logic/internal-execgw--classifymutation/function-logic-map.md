# Function Logic Map: `classifyMutation`

- Source: `internal/execgw/classify.go` (21-99)
- Qualified function: `classifyMutation`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 020dba811b666e94…)
- AST evidence: `ast.json` — AST branches 10, 반환 7, 호출 21
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official 패키지 전체, journal 은 `-run 'Classify|SendTracker|Dispatch'`) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

브로커 호출 1회의 결과를 journal 분류(ACKED/NOT_SENT/REJECTED/AMBIGUOUS)로 바꾼다. a165 의 갈래표(미접수 확정 / 결과 불명 / 확정 거절)는 이 함수의 반환 지점 위에 얹는다(design D2) — 이 함수 자체는 바꾸지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| send journal.SendState | NotStarted/Partial/Complete | WithSendTracker | — |
| err | official 오류 | Client.send | nil 이면 B1 ACKED |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 23:3, 33:3, 46:3, 53:3, 74:3, 93:3, 98:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 22:2` | if err == nil { | 아래 「State mutations」 | 소스 줄 22 이하 | 실행됨 — 패키지 스위트 count=1 |
| B2 | `if at 32:2` | if reason, refused := policyRefusal(err); refused { | 아래 「State mutations」 | 소스 줄 32 이하 | 실행됨 — 패키지 스위트 count=1 |
| B3 | `switch at 44:2` | switch reason, verdict := classifyRefusalCode(err); verdict { | 아래 「State mutations」 | 소스 줄 44 이하 | 측정 불가(본문 블록 없음) |
| B4 | `case at 45:2` | case refusalCodeDefinitive: | 아래 「State mutations」 | 소스 줄 45 이하 | 실행됨 — 패키지 스위트 count=1 |
| B5 | `case at 52:2` | case refusalCodeContradictory: | 아래 「State mutations」 | 소스 줄 52 이하 | 실행됨 — 패키지 스위트 count=1 |
| B6 | `if at 66:2` | if reason, refused := ClassifyBrokerRefusal(err); refused { | 아래 「State mutations」 | 소스 줄 66 이하 | 실행됨 — 패키지 스위트 count=1 |
| B7 | `if at 69:3` | if errors.As(err, &branch) && branch.Source == trading.BranchSourcePostPrepareConfirmation { | 아래 「State mutations」 | 소스 줄 69 이하 | 미실행 — 패키지 스위트 count=0 |
| B8 | `if at 84:2` | if status, known := statusOf(err); known { | 아래 「State mutations」 | 소스 줄 84 이하 | 실행됨 — 패키지 스위트 count=1 |
| B9 | `if at 87:3` | if outcome.Detail == "" { | 아래 「State mutations」 | 소스 줄 87 이하 | 미실행 — 패키지 스위트 count=0 |
| B10 | `else at 89:10` | } else { | 아래 「State mutations」 | 소스 줄 89 이하 | 측정 불가(본문 블록 없음) |

## Calls and live bindings

429 는 B8(statusOf→429)→`ClassifyHTTPMutation` 의 default(AMBIGUOUS). 5xx 도 B8(500)→AMBIGUOUS. 상태 없는 전송 오류는 98 의 `ClassifyHTTPMutation(send,0,err)` — send==NotStarted 면 NOT_SENT, 아니면 AMBIGUOUS. 409 `*APIError` 는 B3 code 판정 없음 → B8 → AMBIGUOUS.

AST 호출 이름 전체: `ClassifyBrokerRefusal`, `brokerOrderID`, `classifyRefusalCode`, `err.Error`, `errors.As`, `journal.ClassifyHTTPMutation`, `policyRefusal`, `reasonForClass`, `statusOf`, `string`.

## State mutations and fallbacks

- 없음(순수).

## Safety conclusion

- 편집 계획 없음. 갈래 판정의 입력으로만 씀.
