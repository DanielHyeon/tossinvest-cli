# Function Logic Map: `consoleMutationConfirmer`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=base, L449–451, 분기 0개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — 본문 **byte 동일**. 인접 삽입의 diff hunk 교차로 evidence가 요구되었다 (revision=base)

콘솔이 **제공하지 않는** 건별 승인 게이트다. 이 change의 수정 대상이 아니며 인접 삽입(발굴 seam 블록)의 diff hunk 교차로 evidence가 요구되었다 — 본문은 base와 byte 동일하다.

`verifylive.Options`는 non-nil Confirmer를 요구하고 nil을 허용적인 기본값으로 대체하지 않는다. 이 값이 그 요구를 만족시킨다. 콘솔은 `ConfirmEach`를 세우지 않으므로 아무도 이것을 부르지 않고, 나중의 편집이 부르게 되더라도 **거절한다** — 여기서 실수는 그 방향으로만 실패해야 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (입력 없음) | — | — | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 없음 | 항상 `verifylive.ErrNotATerminal`을 돌려주는 closure | `TestConsoleMutationConfirmerRefuses`, `TestConsoleWiresTheWebConfirmerAndRefusesPerMutationPrompts` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls=null |

## State mutations and fallbacks

- 상태 없음. 순수 상수 closure.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입만 존재.
- High-risk impact: yes (주문 경로 — 승인 게이트) — 이 값이 허용으로 바뀌면 건별 승인이 자동으로 만족되어 사람 승인 없이 변경이 진행될 수 있다. 현재는 항상 거절한다.
