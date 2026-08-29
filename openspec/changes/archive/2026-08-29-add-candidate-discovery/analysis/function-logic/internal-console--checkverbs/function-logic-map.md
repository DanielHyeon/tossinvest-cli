# Function Logic Map: `checkVerbs`

- Source: `internal/console/static_test.go`
- Change: `add-candidate-discovery`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

식별자가 계좌에 대한 요청을 이름으로 쓰면 실패한다. 본문은 예외 없는 `checkVerbsExcept` 호출 한 줄이다.

**이 필터가 잡는 것과 잡지 못하는 것**: `mutationVerbs`에 철자로 적힌 15개 (order·sell·buy·cancel·modify·amend·flatten·place·create·delete·update·submit·transfer·withdraw·conditional)의 **부분 문자열**만 본다. `Liquidate`·`Execute`·`Unwind`·`Square`·`Dispose`·`Sweep`은 그대로 통과한다. 반대로 `CancelPolicy`·`UpdateLabel` 같은 무해한 이름도 걸린다 — 느슨함이 양쪽으로 느슨하다. **실질 검사는 `Options` 필드의 메서드 집합**이고, 이 필터는 그 위의 보조 장치라고 `capability.VerbExemptions`의 문서가 스스로 적는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `name` | 식별자 | closure / 필드 이름 / 메서드 이름 / 패키지 var 이름 | 부분 문자열 일치면 `t.Errorf` |
| `subject` | 실패 메시지의 주어 | 호출자 | 해당 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 본문에 분기가 없다 | 단일 경로 | `TestNoCapabilityReachesTheConsoleAroundOptions`의 모든 호출 지점 + `TestEveryCapability…` B4의 미열거 필드 경로 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `checkVerbsExcept(t, subject, name, nil)` | 예외 없는 형태 | 순수 위임 | static_test.go:1341 |

## State mutations and fallbacks

- 없음(판정 전용).

## Safety conclusion

- Safe edit boundary: 신설(분기). 예외 있는 형태가 필요해지면서 예외 없는 호출자를 위해 남긴 얇은 래퍼다.
- High-risk impact: yes (주문 능력 주입 차단의 이름 필터 — 단, 위에 적은 대로 이것 하나로 보증이 되지 않는다)
