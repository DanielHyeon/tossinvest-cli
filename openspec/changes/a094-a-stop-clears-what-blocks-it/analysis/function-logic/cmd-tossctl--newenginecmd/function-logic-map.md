# Function Logic Map: `newEngineCmd`

- Source: `cmd/tossctl/engine.go` (`114`–`131`)
- Qualified: `newEngineCmd`
- AST evidence: `ast.json` (`source_sha256` 6844d37952bf3b9f…) — 구현 로트(2026-09-30) 편집 뒤. 편집 전 AST 는 `analysis/implementation/pre-edit/`
- Risk scan: `risk-pattern-report.md`
- 분기 0

**역할.** `tossctl engine` 명령 트리를 조립한다. a094: `engine attempt-resolve`(park 해동, mutating) 한 줄.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root` | 루트 옵션 | main |  |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/<pkg>/ -count=1 -covermode=set` 프로파일(2026-09-30, `analysis/harness/flm_tables.py`)에서 그 줄로 시작하는 블록의 count.

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|

## Calls and live bindings

`newEngineAttemptResolveCmd` 외 종전 AddCommand.

## State mutations and fallbacks

없음(조립).

## Safety conclusion

- 붙지 않은 명령은 없는 명령과 같다 — park 원인 알림 본문이 이 명령을 가리키므로 부착을 시험으로 고정. High-risk: no(조립).
