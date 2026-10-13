# Function Logic Map: `Recorder.Append`

- Source: `internal/soak/record.go` (201-235)
- Qualified: `Recorder.Append`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` 9cd4b2d86043c177…) — AST branches 8 · returns 7
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** soak 주기 하나를 한 줄로 기록 파일에 추가·동기화한다(`soak.go:389` 하나가 부름).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c` | Cycle | Runner | FormatVersion·Kind 비면 채움 |
| `r.f` | 열린 파일 또는 nil | OpenRecorder | nil 이면 오류 |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 202:2 | `if c.FormatVersion == 0 {` | 무편집 |
| B2 | if at 205:2 | `if c.Kind == "" {` | 무편집 |
| B3 | if at 209:2 | `if err != nil {` | 무편집 |
| B4 | if at 214:2 | `if err != nil {` | 편집 영향 |
| B5 | if at 221:2 | `if r.f == nil {` | 무편집 |
| B6 | if at 225:2 | `if err := reopenIfReplaced(r); err != nil {` | 편집 영향 |
| B7 | if at 228:2 | `if _, err := r.f.Write(line); err != nil {` | 무편집 |
| B8 | if at 231:2 | `if err := r.f.Sync(); err != nil {` | 무편집 |

Exact AST return positions: 210:3, 215:3, 222:3, 226:3, 229:3, 232:3, 234:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `json.Marshal` | 직렬화 | 오류면 반환 | AST calls |
| `redact.SealRecordLine`(a131 신설) | account_ref 봉인 + 문자열 치환 | 키 파일 생성 실패면 오류 반환 | design D4 |
| `r.f.Write`·`Sync` | 추가 쓰기 | 오류면 반환 | AST calls |

## State mutations and fallbacks

- 파일 끝에 한 줄. a131: 직렬화 직후 메모리 키로 봉인(redact.SealRecordLineWithKey) · 쓰기 전 경로와 fd 의 inode 대조, 다르면 다시 엶(R-4a).

## Safety conclusion

- Safe edit boundary: B3 뒤 봉인 호출·오류 반환, B4 뒤 inode 대조·재열기. 쓰기·동기화 무편집.
- High-risk impact: yes(간접) — 인터록 증거 기록. 봉인/열기 왕복으로 결속 값 불변.
