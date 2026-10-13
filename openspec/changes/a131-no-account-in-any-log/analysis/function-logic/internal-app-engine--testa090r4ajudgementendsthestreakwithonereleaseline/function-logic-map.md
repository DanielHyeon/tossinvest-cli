# Function Logic Map: `TestA090R4AJudgementEndsTheStreakWithOneReleaseLine`

- Source: `internal/app/engine/a090_unobserved_position_test.go` (742-769)
- Qualified: `TestA090R4AJudgementEndsTheStreakWithOneReleaseLine`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` 89841eae7562f6ca…) — AST branches 5 · returns 0
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 시험 함수(a131 이 기대값만 바꿈). a090 미관측 로그 줄의 `position_id` 기대값을 `redact.String(id)`(= `pos-[redacted]`)로 바꿈 — 줄의 포지션 id 는 계좌 유도 식별자라 로그에서 치환됨(design D0). 알림 payload(Fields)의 기대값은 원문 그대로(로그가 아님). 편집 뒤 작성 — 생산 코드가 아니라 기대값의 정정이며 시험의 분기·흐름은 그대로.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 하네스 | — | 같은 파일의 픽스처 | 실패는 t.Fatal |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 749:2 | `if cycle.Unobserved != 0 \|\| cycle.Judged != 2 {` | 무편집 |
| B2 | if at 753:2 | `if len(lines) != 2 {` | 무편집 |
| B3 | if at 757:2 | `if release["position_id"] != redact.String(target.ID) {` | 무편집 |
| B4 | if at 760:2 | `if secs, ok := release["unobserved_seconds"].(float64); !ok \|\| secs < 15 {` | 무편집 |
| B5 | if at 766:2 | `if got := len(f.lines()); got != 3 {` | 무편집 |

반환문 없음(AST returns 0).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 하네스·대상 함수 | 시험 | — | AST calls |
| `redact.String`/`redact.OpenAccount`(일부) | 치환·봉인 기대값 계산 | 순수 | a131 |

## State mutations and fallbacks

- 시험 상태만.

## Safety conclusion

- Safe edit boundary: 기대값 식만 바꿈 — 분기·흐름 불변.
- High-risk impact: no — 시험 기대값.
