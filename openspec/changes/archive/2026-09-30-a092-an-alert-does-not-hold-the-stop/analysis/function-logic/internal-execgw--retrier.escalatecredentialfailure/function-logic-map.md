# Function Logic Map: `Retrier.escalateCredentialFailure`

- Source: `internal/execgw/retry.go`
- AST evidence: `ast.json` — **편집 뒤**, :409–425, 분기 3 · 반환 4 · 호출 5, source_sha256 `56813596206b…`, 추출 커밋 `e55102f0`. 편집 전 번들은 `analysis/pre-edit/unit5/`(없으면 단위 ④ 번들이 편집 전).
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ⑤ `e55102f0`): `ErrModeAnnouncementFailed`(커밋됨 · 통지 기록 실패) 갈래를 가려 「tightened, but its notice could not be recorded」로 — 분기 하나 추가(B#7).
- 재번호: difflib 정렬(편집 전 `b01e0cd0` 소스 대비): 새 분기 B3(:415).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 + 위 편집 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if r.Escalate == nil \|\| strings.TrimSpace(r.AccountRef) == ""` (:410) | — | — | `TestAuthClassificationStillLatchesThroughTheSentinel`, `TestAuthFailureLatchesEntryImmediately` |
| B2 | `if _, _, err := r.Escalate.EscalateOperatingMode(ctx, r.AccountRef,` (:413) | — | — | (미실행) |
| B3 | `if errors.Is(err, journal.ErrModeAnnouncementFailed)` (:415) | — | — | `TestA092AnUnannouncedCredentialTighteningIsReportedAsTightened` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `errors.Is`, `fmt.Errorf`, `r.Escalate.EscalateOperatingMode`, `strings.TrimSpace` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: 중간 — 보고 정확성.
