# Function Logic Map: `Recorder.Append`

- Source: `internal/verifylive/record.go` (335-360)
- Qualified: `Recorder.Append`
- AST evidence: `ast.json` (편집 전(base), `source_sha256` 9b44a5d8afd7a4dc…) — AST branches 6 · returns 5
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 검증 단계 줄 하나를 기록 파일에 추가·동기화한다(runner · abort · cleanup 이 부름).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `e` | Entry(계좌는 마스크) | Runner | FormatVersion·Kind 비면 채움 |
| `r.f` | 열린 파일 또는 nil | OpenRecorder | nil 이면 오류 |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 336:2 | `if e.FormatVersion == 0 {` | 무편집 |
| B2 | if at 339:2 | `if e.Kind == "" {` | 무편집 |
| B3 | if at 343:2 | `if err != nil {` | 편집 영향 |
| B4 | if at 350:2 | `if r.f == nil {` | 무편집 |
| B5 | if at 353:2 | `if _, err := r.f.Write(line); err != nil {` | 무편집 |
| B6 | if at 356:2 | `if err := r.f.Sync(); err != nil {` | 무편집 |

Exact AST return positions: 344:3, 351:3, 354:3, 357:3, 359:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `json.Marshal` | 직렬화 | 오류면 반환 | AST calls |
| `redact.SealRecordLine`(a131 신설) | account_ref 봉인 + 문자열 치환 | 키 생성 실패면 오류 | design D4 |
| `r.f.Write`·`Sync` | 추가 쓰기 | 오류면 반환 | AST calls |

## State mutations and fallbacks

- 파일 끝에 한 줄. a131: 직렬화 직후 봉인.

## Safety conclusion

- Safe edit boundary: B3 뒤 봉인 호출과 오류 반환 하나 추가.
- High-risk impact: yes(간접) — 감독 증거 기록. 마스크 값 왕복 불변.
