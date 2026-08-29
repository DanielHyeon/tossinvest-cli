# Function Logic Map: `Batch.Summary`

- Source: `internal/verifylive/confirm.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `console-click-approval`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| b.Plan | 계획(요청 0건 이상) | Runner.Plan — 카탈로그에서 구조적으로 파생 | 빈 계획도 렌더된다(승인은 runner가 묻지 않는다) |
| b.Resumed | bool | record의 StepCount>0 | false면 '이 실행', true면 '남은 부분'으로 문구만 달라진다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if b.Resumed {` (internal/verifylive/confirm.go:217, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestSummaryCarriesTheListWithoutTheTypedInstruction, TestPromptIsTheSummaryPlusTheTypedTail |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Plan.WriteLines | 계획 줄을 렌더한다 — 목록의 유일한 원천 | 없음(순수 문자열) | AST callees |

## State mutations and fallbacks

- 없음 — 순수 함수다. 수신자·인자·전역 상태를 바꾸지 않고 문자열만 만든다.
- 확인 문자열(Nonce)과 타이핑 지시는 의도적으로 제외한다 — 콘솔은 클릭으로 승인하므로 화면이 그 문자열을 보여주면 거짓 안내가 된다.

## Safety conclusion

- Safe edit boundary: 목록의 완전성. WriteLines를 우회해 요약을 직접 그리면 배치 모델의 근거가 깨진다.
- High-risk impact: no — 계좌·주문·게이트에 닿지 않는 렌더링. 다만 승인 화면의 내용을 결정하므로 계약(operator-console: 검증 배치 승인의 형식)에 묶여 있다.
