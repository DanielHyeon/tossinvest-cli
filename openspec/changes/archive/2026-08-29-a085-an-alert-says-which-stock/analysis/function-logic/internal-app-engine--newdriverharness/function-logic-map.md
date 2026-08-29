# Function Logic Map: `newDriverHarness`

- Source: `internal/app/engine/reconcileloop_test.go` (lines 121–186)
- AST evidence: `ast.json` (`source_sha256: 2456d7022ac6ce4a6de30e65d7c57d0b10b7a67ded5c6c404c38faa40e550106`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 테스트 코드.

## What it does

드라이버 테스트 하네스. a085는 InstrumentNames를 하네스와 옵션에 배선한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Names` registry | nil 허용 | 대사 루프가 holdings 스냅샷에서 채운다 | nil이면 코드만 렌더 — a085 이전 동작 |
| 브로커 `name` | 빈 문자열 허용 | `GET /api/v1/holdings` 기존 필드 | 비면 코드만 렌더. 추정하지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (129) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (133) `if` — if err := j.SetApplyHooks(journal.ApplyHooks | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (177) `if` — if mutate != nil | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (181) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.Helper` | (122) t.Helper() | 호출부 계약 유지 | AST `calls` |
| `clock.NewFake` | (123) clk := clock.NewFake(reconcileLoopNow) | 호출부 계약 유지 | AST `calls` |
| `journal.Open` | (124) j, err := journal.Open(context.Background(), journal.Options{ | 호출부 계약 유지 | AST `calls` |
| `context.Background` | (124) j, err := journal.Open(context.Background(), journal.Options{ | 호출부 계약 유지 | AST `calls` |
| `filepath.Join` | (125) Path:     filepath.Join(t.TempDir(), "journal.db"), | 호출부 계약 유지 | AST `calls` |
| `t.TempDir` | (125) Path:     filepath.Join(t.TempDir(), "journal.db"), | 호출부 계약 유지 | AST `calls` |
| `journal.FixedFSProber` | (127) FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}), | 호출부 계약 유지 | AST `calls` |
| `t.Fatalf` | (130) t.Fatalf("journal.Open: %v", err) | 호출부 계약 유지 | AST `calls` |
| `t.Cleanup` | (132) t.Cleanup(func() { _ = j.Close() }) | 호출부 계약 유지 | AST `calls` |
| `j.Close` | (132) t.Cleanup(func() { _ = j.Close() }) | 호출부 계약 유지 | AST `calls` |
| `j.SetApplyHooks` | (133) if err := j.SetApplyHooks(journal.ApplyHooks{ | 호출부 계약 유지 | AST `calls` |
| `execgw.NewEntryGate` | (139) gate := execgw.NewEntryGate(clk, nil) | 호출부 계약 유지 | AST `calls` |
| `mutate` | (178) mutate(&opts) | 호출부 계약 유지 | AST `calls` |
| `engine.NewReconcileDriver` | (180) driver, err := engine.NewReconcileDriver(opts) | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 알림 제목·본문 문자열만. 이벤트 타입·Key·Fields·발송 조건·latch·원장은 바뀌지 않는다.

## Safety conclusion

- Safe edit boundary: 사람이 읽는 문자열과 표시 전용 필드. 기계 판독 표면(payload·로그 키·원장 cause)은 영문·원문 그대로다.
- High-risk impact: no — 판정도 주문도 이 경로에 없다. §0.4는 불변이다: 이름은 이미 대금을 지불한 응답에서 온다.
