# Function Logic Map: `main`

- Source: `cmd/tossctl/main.go` (11-22)
- Qualified: `main`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` bca4951a332de1a2…) — AST branches 1 · returns 0
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 프로세스 진입점 — 언어 결정 뒤 루트 명령 실행, 오류면 stderr 에 한 줄 쓰고 exit 1. engine run · soak run · soak attest 의 마지막 오류 줄이 engine.log · soak.log · attest 로그로 간다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `os.Args`·환경 | — | OS | — |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 18:2 | `if err := root.Execute(); err != nil {` | 편집 영향 |

반환문 없음(AST returns 0).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newRootCmd().Execute` | 명령 실행 | 오류 반환 | AST calls |
| `fmt.Fprintln` | 마지막 오류 줄 | 쓰기 오류 무시(종전) | AST calls |
| `redact.Writer`(a131) | 오류 줄 치환 | — | design D3 |

## State mutations and fallbacks

- a131: 루트 명령에 `SetErr(redact.Writer(os.Stderr))`(cobra 의 `Error:` 줄, R-1) · B1 의 마지막 오류 줄도 `redact.Writer(os.Stderr)`. exit 코드 불변.

## Safety conclusion

- Safe edit boundary: 루트 생성과 Execute 사이에 SetErr 한 줄, B1 본문 한 줄.
- High-risk impact: no
