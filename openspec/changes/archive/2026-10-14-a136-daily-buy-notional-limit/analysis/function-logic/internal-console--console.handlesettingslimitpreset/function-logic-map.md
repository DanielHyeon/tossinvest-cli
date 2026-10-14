# Function Logic Map: `handleSettingsLimitPreset`

- Source: `internal/console/settings_limits.go` (251-280, 편집 뒤)
- Qualified: `Console.handleSettingsLimitPreset`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 021a96457d9b0b17…, 편집 전 분기 수 3) — 현재 branches 4 · returns 3
- Risk scan: `risk-pattern-report.md`

**역할.** 프리셋 한 번 클릭 저장.

**a136 편집.** 저장 뒤 파일을 읽어 하루 매수 상한의 결과(통화 전환 시 지움)를 문장에 덧붙임 — 프리셋은 그 값을 쓰지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if c.opts.Limits == nil {` (if 252) | — | 분기 본문 참조 | `TestA136APresetLeavesTheDailyBuyCeilingAlone` · `TestApplyingAPresetWritesAllFiveAtOnce` |
| B2 | `if !ok {` (if 259) | — | 분기 본문 참조 | `TestA136APresetLeavesTheDailyBuyCeilingAlone` · `TestApplyingAPresetWritesAllFiveAtOnce` |
| B3 | `if err := c.opts.Limits.Save(tier.Limits); err != nil {` (if 266) | — | 분기 본문 참조 | `TestA136APresetLeavesTheDailyBuyCeilingAlone` · `TestApplyingAPresetWritesAllFiveAtOnce` |
| B4 | `if gate, err := c.opts.Limits.Load(); err == nil {` (if 275) | — | 분기 본문 참조 | `TestA136APresetLeavesTheDailyBuyCeilingAlone` · `TestApplyingAPresetWritesAllFiveAtOnce` |

Exact AST return positions: 255 · 264 · 268

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | AST |

## State mutations and fallbacks

- 설정 파일/audit(해당 시) — 기존과 같음.

## Safety conclusion

- Safe edit boundary: 칸 · 문장 하나 추가 — 기존 분기 무변경.
- High-risk impact: no(문장)
