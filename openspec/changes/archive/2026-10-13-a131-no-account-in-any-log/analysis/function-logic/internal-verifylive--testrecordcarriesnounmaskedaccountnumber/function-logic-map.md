# Function Logic Map: `TestRecordCarriesNoUnmaskedAccountNumber`

- Source: `internal/verifylive/record_test.go` (156-177)
- Qualified: `TestRecordCarriesNoUnmaskedAccountNumber`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` fe4b744224c3e17f…) — AST branches 5 · returns 0
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 시험 함수(a131 이 기대값만 바꿈). 기록에 마스크도 평문으로 없고, 읽는 쪽이 열면 마스크임을 단언(마스크도 계좌 정보 — design D0). 편집 뒤 작성 — 생산 코드가 아니라 기대값의 정정이며 시험의 분기·흐름은 그대로.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 하네스 | — | 같은 파일의 픽스처 | 실패는 t.Fatal |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 159:2 | `if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {` | 무편집 |
| B2 | if at 163:2 | `if err != nil {` | 무편집 |
| B3 | if at 166:2 | `if strings.Contains(string(data), "123-45-678901") {` | 무편집 |
| B4 | if at 170:2 | `if strings.Contains(string(data), "8901") {` | 무편집 |
| B5 | if at 174:2 | `if err != nil \|\| len(entries) == 0 \|\| !strings.HasSuffix(entries[0].AccountRef, "8901") {` | 무편집 |

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
