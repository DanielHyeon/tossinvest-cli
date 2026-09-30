# Function Logic Map: `Runtime.escalate`

- Source: `internal/app/engine/runtime.go`
- AST evidence: `ast.json` — **편집 뒤**, :442–482, 분기 3 · 반환 3 · 호출 11, source_sha256 `485dbc576483…`, 추출 커밋 `b910173a`. 편집 전 번들은 `analysis/pre-edit/r26b/internal-app-engine--runtime.escalate.json(AST)`에 보존.
- Risk scan: `risk-pattern-report.md`
- 편집: (b910173a, 보이스 B#7 잔재) 지속 실패 승격이 커밋되고 통지 기록만 실패하면 「승격됨 · 통지 기록 실패」로 기록 — 「재시작이 푸는 차단」 오보 제거. 승격 호출 · 알림 불변.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.opts.Escalate` · `AccountRef` | 둘 다 있어야 승격 | 조립 | B1 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:459) | 승격기 · 계좌 없음 → 반환 | — | (미실행) |
| B2 | if (:463) | 커밋됨 · 통지 기록 실패 → 「승격됨」 경고 | — | `TestA092AnUnannouncedSustainedTighteningIsReportedAsTightened` |
| B3 | if (:472) | 승격 오류 → 「재시작이 푼다」 경고 | — | `TestTheDegradationAlertGoesOutEvenWhenTheModeTransitionFails` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Escalate.EscalateOperatingMode` | 지속 실패 → ENTRY_BLOCKED | 오류 분류 | AST |

## State mutations and fallbacks

- 모드 행 하나(변화 시) · 로그.

## Safety conclusion

- Safe edit boundary: 로그 갈래 하나 — 승격 호출 · 루프 재시도 불변.
- High-risk impact: yes(감독자 승격) — 판정 불변.
