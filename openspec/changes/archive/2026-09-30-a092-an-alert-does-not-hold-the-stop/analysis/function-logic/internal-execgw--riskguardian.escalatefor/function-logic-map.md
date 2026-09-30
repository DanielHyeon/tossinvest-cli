# Function Logic Map: `RiskGuardian.escalateFor`

- Source: `internal/execgw/riskguardian.go`
- AST evidence: `ast.json` — **편집 뒤**, :640–656, 분기 3 · 반환 4 · 호출 4, source_sha256 `88cf7c183018…`, 추출 커밋 `b910173a`. 편집 전 번들은 `analysis/pre-edit/r26b/internal-execgw--riskguardian.escalatefor.json(AST)`에 보존.
- Risk scan: `risk-pattern-report.md`
- 편집: (b910173a, 보이스 B#7 잔재) 일일 손실 승격이 커밋되고 통지 기록만 실패하면 「승격됨 · 통지 기록 실패」 오류로 — 「재시작이 푸는 차단」 오보 제거. `errors.Is` 로 원 오류 보존. 진입 거절은 호출자가 이미 판정(불변).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `verdict.Reason` | DAILY_LOSS_LIMIT_REACHED 만 승격 | 체인 | B1 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:641) | 다른 거절 → 승격 없음 | — | `TestAChainRefusalIssuesNothing`, `TestOtherChainRefusalsAreNotTriggers` |
| B2 | if (:644) | 승격 오류 | — | (미실행) |
| B3 | if (:646) | 커밋됨 · 통지 기록 실패 → 「승격됨」 오류 | — | `TestA092AnUnannouncedDailyLossTighteningIsReportedAsTightened` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `journal.EscalateOperatingMode` | 일일 손실 → ENTRY_BLOCKED | 오류는 거절에 덧붙음(errors.Join) | AST |

## State mutations and fallbacks

- 모드 행 하나(변화 시).

## Safety conclusion

- Safe edit boundary: 오류 문구 갈래 하나 — 거절 판정 · 승격 호출 불변(호출자 `errors.Join(chainRefusal, …)`).
- High-risk impact: yes(Guardian) — 판정 불변, 진입은 어느 갈래든 거절.
