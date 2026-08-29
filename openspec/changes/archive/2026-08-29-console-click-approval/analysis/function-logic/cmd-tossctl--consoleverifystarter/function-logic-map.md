# Function Logic Map: `consoleVerifyStarter`

- Source: `cmd/tossctl/console.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `console-click-approval`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| root/recordPath | CLI 옵션 | cmd | 브로커 생성 실패는 에러로 종결 |
| confirm | 콘솔의 BatchConfirmer | console 패키지 | 승인은 콘솔 화면에서만 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (cmd/tossctl/console.go:365, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalRecordNamesTheClickChannel(console harness가 같은 배선을 재현) |
| B2 | `if err != nil {` (cmd/tossctl/console.go:369, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalRecordNamesTheClickChannel(console harness가 같은 배선을 재현) |
| B3 | `if err != nil {` (cmd/tossctl/console.go:374, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalRecordNamesTheClickChannel(console harness가 같은 배선을 재현) |
| B4 | `if err != nil {` (cmd/tossctl/console.go:395, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalRecordNamesTheClickChannel(console harness가 같은 배선을 재현) |
| B5 | `if runErr != nil && (errors.Is(runErr, context.Canceled) \|\| errors.Is(runErr, context.DeadlineExceeded)) {` (cmd/tossctl/console.go:404, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalRecordNamesTheClickChannel(console harness가 같은 배선을 재현) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| verifylive.New | runner 구성 — ApprovalChannelConsoleClick를 넘긴다 | 구성 오류는 에러 반환 | AST callees |
| verifylive.OpenRecorder | 증거 기록 열기 | 에러 반환 | AST callees |
| holdVerifyRunLock | soak 일시정지 마커 | advisory | AST callees |

## State mutations and fallbacks

- runner를 만들고 실행한다(무변경). 추가된 것은 승인 채널 1개 필드다.
- consoleMutationConfirmer는 그대로 ErrNotATerminal을 반환한다 — 콘솔에는 단계별 확인이 없다.

## Safety conclusion

- Safe edit boundary: 이 배선만이 클릭 채널을 선언한다. CLI 경로가 같은 값을 넘기면 TTY run의 증거가 거짓이 된다.
- High-risk impact: no — 배선. 승인 게이트는 콘솔 핸들러와 runner에 있다.
