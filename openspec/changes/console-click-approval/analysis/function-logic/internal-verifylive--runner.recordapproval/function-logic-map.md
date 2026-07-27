# Function Logic Map: `Runner.recordApproval`

- Source: `internal/verifylive/runner.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `console-click-approval`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| plan | 승인 대상 계획 | approveBatch | digest·요청 수·단계 목록으로 기록 |
| batch.Resumed | bool | NewBatch | approval.resumed로 기록 |
| r.approvalChannel | 채널 상수 | New | approval.model의 detail이 된다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — 단일 경로 (internal/verifylive/runner.go) | 아래 State mutations 참조 | 정상 반환 | TestTheApprovalChannelIsRecordedAsGiven, TestTheApprovalRecordNamesTheClickChannel(console) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Recorder.Append | 증거 기록에 승인 엔트리를 남긴다 | 에러는 호출자가 '기록 불가 → 전송 없음'으로 처리 | AST callees |
| maskedAccount | 계좌 마스킹 | 없음 | AST callees |

## State mutations and fallbacks

- 증거 파일에 1줄 append. 승인·거부 모두 기록된다(무변경).
- detail이 하드코딩 문자열에서 채널 값으로 바뀌었다 — 기록의 다른 필드·키·구조는 무변경.

## Safety conclusion

- Safe edit boundary: 승인 사실·계획 digest·요청 수는 계속 기록되어야 한다. 채널은 추가 정보이지 대체가 아니다.
- High-risk impact: no — 기록. 다만 이 기록이 '사람이 승인했다'의 증거이므로 거짓 문구를 쓰지 않는 것이 요구사항이다.
