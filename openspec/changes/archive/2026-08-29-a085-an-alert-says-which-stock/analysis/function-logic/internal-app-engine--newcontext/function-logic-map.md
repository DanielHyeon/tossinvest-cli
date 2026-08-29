# Function Logic Map: `NewContext`

- Source: `internal/app/engine/engine.go` (lines 418–588)
- AST evidence: `ast.json` (`source_sha256: 7c5cc656bd813ed96cd31f8bf8579755efd4a447b5da33e0e67d48be1a3a7d3b`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 알림 문구와 표시 전용 배선. 주문·손절·사이징·원장 판정 경로를 바꾸지 않는다.

## What it does

조립 배선. a085는 알림 표면이 공유하는 InstrumentNames registry 하나를 만들어 넘긴다. 다른 배선은 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Names` registry | nil 허용 | 대사 루프가 holdings 스냅샷에서 채운다 | nil이면 코드만 렌더 — a085 이전 동작 |
| 브로커 `name` | 빈 문자열 허용 | `GET /api/v1/holdings` 기존 필드 | 비면 코드만 렌더. 추정하지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (420) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (427) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (431) `if` — if clk == nil | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (445) `if` — if opts.Publisher != nil | 본문 참조 | — | 아래 Branch Test Map |
| B5 | (453) `if` — if err := recordGateSettings(auditLog, gate, cfg.Engine.Adoption, notification | 본문 참조 | — | 아래 Branch Test Map |
| B6 | (468) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B7 | (475) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B8 | (484) `if` — if err := bindApplyHooks(jrn); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B9 | (503) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B10 | (518) `if` — if gate.Enabled && guardian == nil && !opts.disableProductionGuardian | 본문 참조 | — | 아래 Branch Test Map |
| B11 | (520) `if` — if factory == nil | 본문 참조 | — | 아래 Branch Test Map |
| B12 | (526) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B13 | (546) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B14 | (550) `if` — if !automation.Verified | 본문 참조 | — | 아래 Branch Test Map |
| B15 | (555) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `NewOrderPath` | (419) path, err := NewOrderPath(opts) | 호출부 계약 유지 | AST `calls` |
| `openAuditLog` | (426) auditLog, err := openAuditLog(opts) | 호출부 계약 유지 | AST `calls` |
| `clock.System` | (432) clk = clock.System() | 호출부 계약 유지 | AST `calls` |
| `newAutomationStatus` | (436) status := newAutomationStatus(gate, paths, opts.protectionReadiness()) | 호출부 계약 유지 | AST `calls` |
| `opts.protectionReadiness` | (436) status := newAutomationStatus(gate, paths, opts.protectionReadiness()) | 호출부 계약 유지 | AST `calls` |
| `resolveNotificationPublisher` | (444) publisher, notifications := resolveNotificationPublisher(cfg.Engine.Notifications, opts.Getenv) | 호출부 계약 유지 | AST `calls` |
| `recordGateSettings` | (453) if err := recordGateSettings(auditLog, gate, cfg.Engine.Adoption, notifications, | 호출부 계약 유지 | AST `calls` |
| `fmt.Errorf` | (458) return nil, fmt.Errorf("engine: recording the automation-gate audit trail: %w", err) | 호출부 계약 유지 | AST `calls` |
| `resolveAccountRef` | (467) accountRef, err := resolveAccountRef(ctx, off) | 호출부 계약 유지 | AST `calls` |
| `refuseStartup` | (469) return nil, refuseStartup(auditLog, gate, fmt.Errorf("%w: %w", ErrAccountUnresolved, err)) | 호출부 계약 유지 | AST `calls` |
| `openEngineJournal` | (474) jrn, err := openEngineJournal(ctx, opts, clk) | 호출부 계약 유지 | AST `calls` |
| `bindApplyHooks` | (484) if err := bindApplyHooks(jrn); err != nil { | 호출부 계약 유지 | AST `calls` |
| `jrn.Close` | (485) _ = jrn.Close() | 호출부 계약 유지 | AST `calls` |
| `buildGateway` | (490) wiring, err := buildGateway(ctx, gatewayInputs{ | 호출부 계약 유지 | AST `calls` |
| `protectionManifestPin` | (497) manifestPin: protectionManifestPin(opts), | 호출부 계약 유지 | AST `calls` |
| `factory` | (523) guardian, err = factory( | 호출부 계약 유지 | AST `calls` |
| `runInterlock` | (539) automation, err := runInterlock(status, gate, auditLog, opts.Logger, gateFacts{ | 호출부 계약 유지 | AST `calls` |
| `clk.Now` | (544) now:       clk.Now(), | 호출부 계약 유지 | AST `calls` |
| `strategyprojection.NewStore` | (554) projection, err := strategyprojection.NewStore(strategyprojection.DormantSnapshot(clk.Now().UTC())) | 호출부 계약 유지 | AST `calls` |
| `strategyprojection.DormantSnapshot` | (554) projection, err := strategyprojection.NewStore(strategyprojection.DormantSnapshot(clk.Now().UTC())) | 호출부 계약 유지 | AST `calls` |
| `UTC` | (554) projection, err := strategyprojection.NewStore(strategyprojection.DormantSnapshot(clk.Now().UTC())) | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 알림 제목·본문 문자열만. 이벤트 타입·Key·Fields·발송 조건·latch·원장은 바뀌지 않는다.

## Safety conclusion

- Safe edit boundary: 사람이 읽는 문자열과 표시 전용 필드. 기계 판독 표면(payload·로그 키·원장 cause)은 영문·원문 그대로다.
- High-risk impact: no — 판정도 주문도 이 경로에 없다. §0.4는 불변이다: 이름은 이미 대금을 지불한 응답에서 온다.
