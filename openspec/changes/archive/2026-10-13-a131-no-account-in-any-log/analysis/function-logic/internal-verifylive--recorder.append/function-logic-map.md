# Function Logic Map: `Recorder.Append`

- Source: `internal/verifylive/record.go` (349-383)
- Qualified: `Recorder.Append`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` 9a289d21048578d8…) — AST branches 8 · returns 7
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
| B1 | if at 350:2 | `if e.FormatVersion == 0 {` | 무편집 |
| B2 | if at 353:2 | `if e.Kind == "" {` | 무편집 |
| B3 | if at 357:2 | `if err != nil {` | 무편집 |
| B4 | if at 362:2 | `if err != nil {` | 편집 영향 |
| B5 | if at 369:2 | `if r.f == nil {` | 무편집 |
| B6 | if at 373:2 | `if err := reopenIfReplaced(r); err != nil {` | 편집 영향 |
| B7 | if at 376:2 | `if _, err := r.f.Write(line); err != nil {` | 무편집 |
| B8 | if at 379:2 | `if err := r.f.Sync(); err != nil {` | 무편집 |

Exact AST return positions: 358:3, 363:3, 370:3, 374:3, 377:3, 380:3, 382:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `json.Marshal` | 직렬화 | 오류면 반환 | AST calls |
| `redact.SealRecordLine`(a131 신설) | account_ref 봉인 + 문자열 치환 | 키 생성 실패면 오류 | design D4 |
| `r.f.Write`·`Sync` | 추가 쓰기 | 오류면 반환 | AST calls |

## State mutations and fallbacks

- 파일 끝에 한 줄. a131: 직렬화 직후 메모리 키로 봉인(redact.SealRecordLineWithKey) · 쓰기 전 경로와 fd 의 inode 대조, 다르면 다시 엶(R-4a).

## Safety conclusion

- Safe edit boundary: B3 뒤 봉인 호출·오류 반환, B4 뒤 inode 대조·재열기. 쓰기·동기화 무편집.
- High-risk impact: yes(간접) — 감독 증거 기록. 마스크 값 왕복 불변.
