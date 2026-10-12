# Function Logic Map: `newEngineRunCmd`

- Source: `cmd/tossctl/engine.go` (133-172)
- Qualified: `newEngineRunCmd`
- AST evidence: `ast.json` (편집 전(base), `source_sha256` de6bcd1329628b5e…) — AST branches 0 · returns 2
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** `tossctl engine run` cobra 명령을 만든다. RunE 가 `runEngineRun` 을 부른다 — 이 프로세스의 stdout/stderr 가 engine.log 다(`spawnDetachedEngine` · `engine-autostart.sh`).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root` | rootOptions | newRootCmd | — |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | branchless happy path at 133:1 | `func newEngineRunCmd(root *rootOptions) *cobra.Command {` | 분기 없음 — RunE 클로저 본문 편집 |

Exact AST return positions: 168:4, 171:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `runEngineRun` | 엔진 실행 | 오류 반환 | AST calls |
| `redact.Writer`(a131) | cmd 의 out/err 를 치환 쓰기로 감쌈 | 쓰기 오류 전파 | design D3 |

## State mutations and fallbacks

- a131: RunE 안에서 `cmd.SetOut(redact.Writer(cmd.OutOrStdout()))` · `cmd.SetErr(…)` 뒤 `runEngineRun` — 엔진 로거도 이 errOut 을 받는다.

## Safety conclusion

- Safe edit boundary: RunE 클로저에 감싸기 두 줄. runEngineRun 본문 무편집.
- High-risk impact: no(직접) — 엔진 프로세스의 출력 바이트만.
