# Function Logic Map: `mergeEngine`

- Source: `internal/config/engine.go` (317-350, 편집 뒤)
- Qualified: `mergeEngine`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 9d3ab3d2a37da777…, 편집 전 분기 수 5) — 현재 branches 5 · returns 2
- Risk scan: `risk-pattern-report.md`

**역할.** config.json 의 engine 블록을 기본값 위에 덮어씀.

**a136 편집.** 끝에 `cfg.AutomationGate.MaxDailyBuyNotional = gate.MaxDailyBuyNotional` 한 줄(분기 무변경). FLM 전에 편집됨(P1-3 기록).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `raw` | nil 허용 | 파싱 | nil 이면 무변경 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if raw == nil {` (if 318) | — | 분기 본문 참조 | `TestA136DailyBuyCeilingIsParsedFromTheGateBlock` · `TestAutomationGateParsed` |
| B2 | `if raw.Autostart != nil {` (if 321) | — | 분기 본문 참조 | `TestA136DailyBuyCeilingIsParsedFromTheGateBlock` · `TestAutomationGateParsed` |
| B3 | `if raw.ExitPolicy != nil {` (if 325) | — | 분기 본문 참조 | `TestA136DailyBuyCeilingIsParsedFromTheGateBlock` · `TestAutomationGateParsed` |
| B4 | `if raw.AutomationGate == nil {` (if 335) | — | 분기 본문 참조 | `TestA136DailyBuyCeilingIsParsedFromTheGateBlock` · `TestAutomationGateParsed` |
| B5 | `if gate.Enabled != nil {` (if 339) | — | 분기 본문 참조 | `TestA136DailyBuyCeilingIsParsedFromTheGateBlock` · `TestAutomationGateParsed` |

Exact AST return positions: 319 · 336

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `mergeAdoption` · `mergeNotifications` | 하위 블록 | — | AST |

## State mutations and fallbacks

- cfg 필드 대입.

## Safety conclusion

- Safe edit boundary: 대입 한 줄.
- High-risk impact: no(설정 읽기)
