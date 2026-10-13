# Function Logic Map: `newRootCmd`

- Source: `cmd/tossctl/root.go` (52-202)
- Qualified: `newRootCmd`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` 6146622f5e142a62…) — AST branches 7 · returns 3
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 루트 명령 — 전역 플래그와 하위 명령 등록.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 없음 | — | — | — |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 62:4 | `if err != nil {` | 무편집 |
| B2 | if at 71:4 | `if cachePath, err := resolveUpdateCachePath(opts); err == nil {` | 무편집 |
| B3 | if at 79:4 | `if status, err := loadConfigStatus(opts); err == nil {` | 무편집 |
| B4 | if at 85:4 | `if format == output.FormatTable {` | 무편집 |
| B5 | if at 87:5 | `if credFile, _, err := resolveOpenAPIPaths(opts); err == nil {` | 무편집 |
| B6 | if at 88:6 | `if creds, err := official.LoadCredentials(os.Getenv, credFile); err == nil && creds != nil {` | 무편집 |
| B7 | if at 96:5 | `if shouldHintOnboarding(state, tui.IsInteractive(os.Stdin, os.Stdout), cmd.Name()) {` | 무편집 |

Exact AST return positions: 63:5, 101:4, 201:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `cmd.AddCommand` | 하위 명령 등록 | — | AST calls |
| `newScrubAccountRecordsCmd`(a131 신설) | 정리 명령 | — | design D5 |

## State mutations and fallbacks

- a131: AddCommand 목록에 정리 명령 하나 추가.

## Safety conclusion

- Safe edit boundary: AddCommand 인자 하나 추가 — 분기 무편집.
- High-risk impact: no
