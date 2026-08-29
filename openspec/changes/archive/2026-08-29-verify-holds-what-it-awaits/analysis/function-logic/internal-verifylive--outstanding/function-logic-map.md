# Function Logic Map: `Outstanding`

- Source: `internal/verifylive/record.go`
- Function: `internal/verifylive/record.go:Outstanding`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-holds-what-it-awaits`

기록이 말하는 **아직 살아 있는 객체**를 돌려준다. 이 도구가 깨끗하게 끝났는지 판정하는 함수이고 화면·요약·노출 상한·정리가 전부 이것을 읽는다. 이 change의 편집은 본문을 `outstandingLines`로 옮기고 껍데기만 남긴 것이다 — 규칙(마지막 비취소 언급이 이긴다, 취소는 단조)은 한 줄도 안 바뀌었다. 옮긴 이유는 정리 규칙이 **그 줄의 index**를 필요로 하기 때문이고, 그 index를 여기서 버리면 `HeldUntil`을 읽은 줄과 판정을 재는 줄이 달라진다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| entries []Entry | append-only 기록 | capability-verify*.jsonl | 빈 기록이면 nil |
| outstandingLines(entries) | index를 단 같은 결과 | record.go | 동일 순서·동일 내용 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 460) — `for _, l := range outstandingLines(entries) {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `outstandingLines` | ast.json calls (line 460) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `append` | ast.json calls (line 461) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 없음 — 새 slice를 만든다. nil-ness도 보존한다(빈 결과는 nil).

## Safety conclusion

- Safe edit boundary: 본문 이동뿐. 반환 값·순서·nil 여부가 모두 같고 기존 호출부 전부가 무변경 테스트로 덮여 있다.
- High-risk impact: yes — 노출 상한과 정리 대상이 이 함수의 결과를 센다.
