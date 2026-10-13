# Function Logic Map: `TestTheTransitionLogLineIsCountable`

- Source: `internal/obs/mode_test.go` (89-133)
- Qualified: `TestTheTransitionLogLineIsCountable`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` e9cfcc50feb5cda6…) — AST branches 7 · returns 0
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 시험 함수(a131 이 기대값만 바꿈). 운영 모드 전이 줄이 예전엔 `account` 필드를 셌음 — a131 뒤에는 그 필드가 **없음**을 단언(사용자 결정). 편집 뒤 작성 — 생산 코드가 아니라 기대값의 정정이며 시험의 분기·흐름은 그대로.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 하네스 | — | 같은 파일의 픽스처 | 실패는 t.Fatal |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 95:2 | `if err := n.AnnounceOperatingMode(context.Background(), journal.ModeEntryBlocked,` | 무편집 |
| B2 | if at 108:2 | `if err := json.Unmarshal([]byte(first), &line); err != nil {` | 무편집 |
| B3 | if at 118:2 | `if _, ok := line[obs.FieldAccount]; ok {` | 무편집 |
| B4 | range at 121:2 | `for field, value := range want {` | 무편집 |
| B5 | if at 123:3 | `if !ok {` | 무편집 |
| B6 | if at 126:3 | `if got != value {` | 무편집 |
| B7 | if at 130:2 | `if reason, _ := line[obs.FieldReason].(string); !strings.Contains(reason, "approved-by: sre-1") {` | 무편집 |

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
