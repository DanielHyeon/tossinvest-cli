# Function Logic Map: `Runner.chainOf`

- Source: `internal/verifylive/runner.go`
- Function: `internal/verifylive/runner.go:Runner.chainOf`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-holds-what-it-awaits`

이 객체가 이미 달고 있는 사슬을 찾는다. 최신 언급이 이기도록 **현재 단계 → 이 실행이 쓴 줄 → 시작할 때의 기록** 순으로 본다. 정정이 새 식별자를 만들 때 후속 객체가 선행 객체의 사슬을 물려받는 자리이고, 등록이 멱등 재생으로 같은 식별자를 돌려받았을 때 사슬을 새로 만들지 않는 자리다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sr.artifacts | 현재 단계가 기록한 객체 | sr.created/cancelled | 여기 있으면 우선 |
| r.written | 이 실행이 이미 쓴 줄 | runner | 다음 우선 |
| r.prior | 실행 시작 시점의 기록 | capability-verify*.jsonl | 마지막 — 없으면 "" |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 758) — `for _, entries := range [][]Entry{{{Artifacts: sr.artifacts}}, r.written, r.prior} {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `if` (line 759) — `if c := ChainOf(entries, kind, id); c != "" {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ChainOf` | ast.json calls (line 759) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 없음 — 읽기만 한다.

## Safety conclusion

- Safe edit boundary: 신규 leaf. 빈 문자열을 돌려주면 호출부가 새 사슬을 만들고, 그것이 첫 등록의 정상 경로다.
- High-risk impact: no — 사슬은 정리 판정에 쓰이지 않는다.
