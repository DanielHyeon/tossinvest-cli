# Function Logic Map: `TestReconcileRefusesANewerRecordFormat`

- Source: `internal/verifylive/reconcile_g2_test.go` (357-378)
- Qualified: `TestReconcileRefusesANewerRecordFormat`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` 69d35b7c393e8b03…) — AST branches 1 · returns 0
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 시험 함수(a131 이 기대값만 바꿈). 「더 새로운 형식」 줄을 하드코딩 2 에서 `RecordFormatVersion+1` 로 — 형식 2 가 이제 현재 형식(a131 R-2c). 편집 뒤 작성 — 생산 코드가 아니라 기대값의 정정이며 시험의 분기·흐름은 그대로.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 하네스 | — | 같은 파일의 픽스처 | 실패는 t.Fatal |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | range at 359:2 | `for name, write := range map[string]func(h *rcHarness){` | 무편집 |

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
