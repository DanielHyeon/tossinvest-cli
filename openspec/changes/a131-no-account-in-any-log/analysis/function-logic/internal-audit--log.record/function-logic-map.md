# Function Logic Map: `Log.Record`

- Source: `internal/audit/audit.go` (173-210)
- Qualified: `Log.Record`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` 29886e7a4d031757…) — AST branches 8 · returns 6
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 감사 줄 하나를 O_APPEND + fsync 로 쓴다. 감사 기록의 유일한 쓰기 입구(`RecordChange` · `RecordAction` 이 거침).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `l` | nil 가능 | Open | nil 이면 무동작(B1) |
| `e` | Entry | 호출자 | At·Subject 비면 채움 |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 174:2 | `if l == nil {` | 무편집 |
| B2 | if at 179:2 | `if e.At.IsZero() {` | 무편집 |
| B3 | else at 181:9 | `} else {` | 무편집 |
| B4 | if at 184:2 | `if strings.TrimSpace(e.Subject) == "" {` | 무편집 |
| B5 | if at 189:2 | `if err != nil {` | 무편집 |
| B6 | if at 198:2 | `if err != nil {` | 무편집 |
| B7 | if at 203:2 | `if _, err := f.Write(line); err != nil {` | 무편집 |
| B8 | if at 206:2 | `if err := f.Sync(); err != nil {` | 무편집 |

Exact AST return positions: 175:3, 190:3, 199:3, 204:3, 207:3, 209:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `json.Marshal` | 직렬화 | 오류면 반환 | AST calls |
| `os.OpenFile`·`Write`·`Sync` | 추가 쓰기 | 오류면 반환 | AST calls |
| `RedactEntry`(a131 신설) | Setting·Old·New·Detail·Subject 의 계좌 치환 | 순수 | design D3 |

## State mutations and fallbacks

- 파일 끝에 한 줄. a131: 직렬화 전 `e = RedactEntry(e)` — 상세(Detail)의 계좌 텍스트만 `[account]`(R-11).

## Safety conclusion

- Safe edit boundary: B1 뒤 한 줄. 이후 분기 무편집.
- High-risk impact: no — 감사 의미 보존(설계 D3-a). §0.5 추적성 유지.
