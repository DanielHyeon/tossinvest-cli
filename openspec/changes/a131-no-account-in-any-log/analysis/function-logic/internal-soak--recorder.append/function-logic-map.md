# Function Logic Map: `Recorder.Append`

- Source: `internal/soak/record.go` (187-212)
- Qualified: `Recorder.Append`
- AST evidence: `ast.json` (편집 전(base), `source_sha256` 583ceb9012f979bd…) — AST branches 6 · returns 5
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
| B1 | if at 188:2 | `if c.FormatVersion == 0 {` | 무편집 |
| B2 | if at 191:2 | `if c.Kind == "" {` | 무편집 |
| B3 | if at 195:2 | `if err != nil {` | 편집 영향 |
| B4 | if at 202:2 | `if r.f == nil {` | 무편집 |
| B5 | if at 205:2 | `if _, err := r.f.Write(line); err != nil {` | 무편집 |
| B6 | if at 208:2 | `if err := r.f.Sync(); err != nil {` | 무편집 |

Exact AST return positions: 196:3, 203:3, 206:3, 209:3, 211:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `json.Marshal` | 직렬화 | 오류면 반환 | AST calls |
| `redact.SealRecordLine`(a131 신설) | account_ref 봉인 + 문자열 치환 | 키 파일 생성 실패면 오류 반환 | design D4 |
| `r.f.Write`·`Sync` | 추가 쓰기 | 오류면 반환 | AST calls |

## State mutations and fallbacks

- 파일 끝에 한 줄. a131: 직렬화 직후 봉인 — 기록 줄에 계좌 원문 없음.

## Safety conclusion

- Safe edit boundary: B3(직렬화 오류) 뒤 봉인 호출과 그 오류 반환 하나 추가. 잠금·쓰기·동기화 무편집.
- High-risk impact: yes(간접) — 인터록 증거 기록. 봉인/열기 왕복으로 결속 값 불변.
