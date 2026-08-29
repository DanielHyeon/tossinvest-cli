# Function Logic Map: `New`

- Source: `internal/verifylive/runner.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `console-click-approval`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| o.ApprovalChannel | 빈 문자열 또는 채널 상수 | 호출자(cmd/tossctl) | 빈 값은 ApprovalChannelTyped로 채워진다 — §0.2 zero-value 안전 |
| 기존 필수 입력 | Broker·Recorder·Confirm·ConfirmBatch·AccountRef | 호출자 | 누락은 에러로 거부(무변경) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if o.Broker == nil {` (internal/verifylive/runner.go:154, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalChannelIsRecordedAsGiven |
| B2 | `if o.Recorder == nil {` (internal/verifylive/runner.go:157, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalChannelIsRecordedAsGiven |
| B3 | `if o.Confirm == nil {` (internal/verifylive/runner.go:160, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalChannelIsRecordedAsGiven |
| B4 | `if o.ConfirmBatch == nil {` (internal/verifylive/runner.go:163, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalChannelIsRecordedAsGiven |
| B5 | `if strings.TrimSpace(o.AccountRef) == "" {` (internal/verifylive/runner.go:167, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalChannelIsRecordedAsGiven |
| B6 | `if err != nil {` (internal/verifylive/runner.go:171, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalChannelIsRecordedAsGiven |
| B7 | `if r.approvalChannel == "" {` (internal/verifylive/runner.go:195, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalChannelIsRecordedAsGiven |
| B8 | `if r.out == nil {` (internal/verifylive/runner.go:198, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalChannelIsRecordedAsGiven |
| B9 | `if r.now == nil {` (internal/verifylive/runner.go:201, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalChannelIsRecordedAsGiven |
| B10 | `if r.sleep == nil {` (internal/verifylive/runner.go:204, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalChannelIsRecordedAsGiven |
| B11 | `if r.maxSellQuantity <= 0 {` (internal/verifylive/runner.go:207, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalChannelIsRecordedAsGiven |
| B12 | `if r.ttlWait <= 0 {` (internal/verifylive/runner.go:210, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalChannelIsRecordedAsGiven |
| B13 | `if r.process.InstanceID == "" {` (internal/verifylive/runner.go:213, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalChannelIsRecordedAsGiven |
| B14 | `for _, id := range o.Redo {` (internal/verifylive/runner.go:216, range) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalChannelIsRecordedAsGiven |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| strings.TrimSpace | 채널 문자열 정규화 | 없음 | AST callees |
| ValidateOffset | 오프셋 검증(무변경) | 에러 반환 | AST callees |

## State mutations and fallbacks

- Runner.approvalChannel를 1회 설정한다. 이후 recordApproval·writeBanner만 읽는다.
- 다른 필드 초기화·검증 순서는 무변경.

## Safety conclusion

- Safe edit boundary: 기본값이 '타이핑'이어야 한다. 기본값이 클릭이면 CLI run의 증거가 거짓이 된다.
- High-risk impact: no — 구성만 한다. 승인 게이트 자체(ConfirmBatch 필수)는 무변경.
