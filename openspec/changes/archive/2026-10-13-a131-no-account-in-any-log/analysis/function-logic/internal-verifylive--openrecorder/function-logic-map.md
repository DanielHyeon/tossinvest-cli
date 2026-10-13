# Function Logic Map: `OpenRecorder`

- Source: `internal/verifylive/record.go` (327-343)
- Qualified: `OpenRecorder`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` 9a289d21048578d8…) — AST branches 3 · returns 4
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 기록 파일을 추가 쓰기용으로 연다(디렉터리 생성 · O_APPEND · 0600). verify 실행가 브로커 호출 전에 부름. 편집 전 AST 는 작업 트리(구조체 필드 3줄만 더한 상태 — 함수 본문은 base 와 같음, 줄 +3)에서 뽑음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | 기록 경로 | 호출자 | 디렉터리·파일 생성 실패면 오류 |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 329:2 | `if err := os.MkdirAll(dir, 0o700); err != nil {` | 무편집 |
| B2 | if at 333:2 | `if err != nil {` | 무편집 |
| B3 | if at 338:2 | `if err != nil {` | 편집 영향 |

Exact AST return positions: 330:3, 334:3, 340:3, 342:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `os.MkdirAll`·`os.OpenFile` | 디렉터리·파일 | 오류 반환 | AST calls |
| `redact.PrepareSealKey`(a131 신설) | 봉인 키 읽기/생성 + 시험 봉인 — 봉인 줄이 있는 기록에 키가 없으면 거절 | 오류 반환(파일 닫음) | design D4 R-2b · R-3 |

## State mutations and fallbacks

- Recorder 반환. a131: 파일을 연 뒤 키를 준비해 Recorder 에 둠 — 실패하면 파일을 닫고 오류(측정이 시작되지 않음).

## Safety conclusion

- Safe edit boundary: 마지막 return 앞에 키 준비와 그 오류 반환 하나. B1·B2 무편집.
- High-risk impact: yes(간접) — verify 의 실주문보다 먼저 키 실패를 드러냄(R-3).
