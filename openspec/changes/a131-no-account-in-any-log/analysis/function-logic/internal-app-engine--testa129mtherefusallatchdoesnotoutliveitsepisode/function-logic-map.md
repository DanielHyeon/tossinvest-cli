# Function Logic Map: `TestA129MTheRefusalLatchDoesNotOutliveItsEpisode`

- Source: `internal/app/engine/a129_exit_order_log_test.go` (762-786)
- Qualified: `TestA129MTheRefusalLatchDoesNotOutliveItsEpisode`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` a89249db4c0a3586…) — AST branches 2 · returns 0
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 시험 함수(a131 이 기대값만 바꿈). a129 래치 시험의 `position_id` 기대값을 치환된 꼴로 바꿈. 편집 뒤 작성 — 생산 코드가 아니라 기대값의 정정이며 시험의 분기·흐름은 그대로.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 하네스 | — | 같은 파일의 픽스처 | 실패는 t.Fatal |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 780:2 | `if len(f.submit.places) == placesBefore {` | 무편집 |
| B2 | if at 783:2 | `if got := len(a129Of(f.lines(), obs.EventOrderRejected)); got != 2 {` | 무편집 |

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
