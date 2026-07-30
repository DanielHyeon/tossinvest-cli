# Function Logic Map: `observationDetail`

- Source: `internal/verifylive/us_market_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트 헬퍼**다(HEAD L231-244). 기록(record)에서 한 단계의 마지막 항목을 찾아
주어진 키의 `Detail` 문자열을 돌려준다. 이 change 구간의 앞선 커밋 `f62457c`
(`apply-us-measurement-fixes`)가 넣었고, `add-candidate-discovery`의 base가 그보다
앞서기 때문에 이 change의 diff에 잡힌다.

기록의 **내용**을 단언하려면 그것을 꺼내는 한 곳이 필요하다 — 두 테스트가 같은 방식으로
꺼내야 "US 문장이 KR을 말하지 않는다"와 "KR 문장이 여전히 KR을 말한다"가 같은 대상에
대한 두 반쪽이 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `entries` | 한 실행의 기록 항목들 | `h.entries()` | — |
| `step` | `StepOrderCancel` \| `StepOrderAmend` 등 | 호출 테스트 | 항목 없으면 `t.Fatalf` |
| `key` | 관측 키(`order.place.ok` 등) | 호출 테스트 | 없으면 `t.Fatalf` |

불변식: **부재를 빈 문자열로 돌려주지 않는다.** 없으면 `t.Fatalf`다 — 관측이 사라졌는데
`strings.Contains("")`가 통과해 테스트가 조용히 아무것도 재지 않는 상태를 막는다.
(`return ""`는 `t.Fatalf` 뒤의 도달 불가 줄이며 컴파일러를 위한 것이다.)

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L234) | `LastEntry`가 그 단계를 못 찾음 | 없음 | `t.Fatalf("no record entry for %s")` | 자체 실행 |
| B2 (range, L237) | 항목의 관측 순회 | 없음 | — | 자체 실행 |
| B3 (if, L238) | `o.Key == key` | 없음 | `o.Detail` 반환 | 자체 실행 |

꼬리: 키를 못 찾으면 `t.Fatalf("%s recorded no %s")`.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.Helper` | 실패 위치를 호출부로 | — | ast.json calls |
| `LastEntry` | 그 단계의 마지막 기록 항목 | `(Entry, bool)` | ast.json calls |
| `t.Fatalf` ×2 | 부재를 실패로 | — | ast.json calls |

## State mutations and fallbacks

- 없음. 메모리 상의 기록 slice만 읽는다. 실계좌·브로커·파일 무접촉.

## Safety conclusion

- Safe edit boundary: 신규 테스트 헬퍼 가산.
- High-risk impact: **no** — 테스트 전용 읽기 헬퍼. 다만 이 헬퍼의 **부재 처리**가
  두 기록 테스트의 값을 지탱한다: 부재를 빈 문자열로 흡수했다면 관측이 통째로 사라져도
  `Contains(KR)`이 false여서 US 테스트가 통과했을 것이다.
