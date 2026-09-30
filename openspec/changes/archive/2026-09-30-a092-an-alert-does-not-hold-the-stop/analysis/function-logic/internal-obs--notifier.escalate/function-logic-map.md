# Function Logic Map: `Notifier.escalate`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 뒤**, :425–447, 분기 4 · 반환 2 · 호출 6, source_sha256 `d705f78d68c1…`, 추출 커밋 `15b64676`. 편집 전 번들은 `analysis/pre-edit/r26b/internal-obs--notifier.escalate/`에 보존.
- Risk scan: `risk-pattern-report.md`
- 편집: (b910173a, 보이스 A#4 · B#4) 승격 실패 줄의 `FieldEvent` → `FieldTriggerEvent`. (15b64676, 게이트 준비 gstack 리뷰) 같은 줄의 error 칸을 `MaskAccount(err, n.AccountRef)` 로 — 원장 문구의 계좌 원문을 가림(불변식 8 (a), judge 게이트 설명 고정 문구의 짝). 분기 · 반환 불변. (같은 줄의 `FieldAccount` 는 base 관행 — 불변식 8 (b) 사람 결정 큐.)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` · `n.AccountRef` | 둘 다 있어야 승격 포함 | 조립 | B1 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:426) | 승격 미포함 | — | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseBeforeTheEvidenceDoesNotChangeTheVerdict` |
| B2 | switch (:431) | 결과 분기 | — | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` |
| B3 | case (:432) | 승격 실패 로그 | — | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` |
| B4 | case (:438) | 승격 됨 로그 | — | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `EscalateOperatingMode(…, nil)` | 승격 · 통지 없음 | 오류 반환 | AST |

## State mutations and fallbacks

- 모드 행 하나(변화 시).

## Safety conclusion

- Safe edit boundary: 로그 키만.
- High-risk impact: yes(모드 승격) — 이 편집은 판정 불변.
