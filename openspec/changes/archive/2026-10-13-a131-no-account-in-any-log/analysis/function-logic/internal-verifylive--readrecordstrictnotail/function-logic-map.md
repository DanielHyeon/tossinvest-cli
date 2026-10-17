# Function Logic Map: `readRecordStrictNoTail`

- Source: `internal/verifylive/reconcile_record.go` (30-44)
- Qualified: `readRecordStrictNoTail`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` 84baeb1173ccb0d7…) — AST branches 3 · returns 4
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 대사 전 기록을 한 번 읽어 엄격 해독한다 — 반환한 원본 바이트는 동시성 지문 입력.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | 파일 경로 | Reconcile | 못 읽으면 refuse |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 32:2 | `if err != nil {` | 무편집 |
| B2 | if at 36:2 | `if err != nil {` | 무편집 |
| B3 | if at 40:2 | `if err := openRecordEntries(entries, path); err != nil {` | 편집 영향 |

Exact AST return positions: 33:3, 37:3, 41:3, 43:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readRecordRaw` | 원본 읽기 | refuse | AST calls |
| `decodeRecordStrict` | 엄격 해독 | refuse | AST calls |
| `openEntryAccount`(a131) | 봉인 열기 | 빈 값이면 계좌 결속이 거절(fail-closed) | design D4 |

## State mutations and fallbacks

- a131: 해독한 항목의 계좌를 엶 — 열 수 없으면 refuse(RefuseRecordUndecodable). **원본 바이트(raw)는 바꾸지 않음** — 지문 대조 불변.

## Safety conclusion

- Safe edit boundary: 마지막 return 앞 열기 루프. 분기 B1·B2 무편집.
- High-risk impact: yes(간접) — 대사 계좌 결속 입력.
