# Function Logic Map: `TestTheConsoleDecidesNothingAboutTheGate`

- Source: `internal/console/engineproc_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `packageFiles(t)` | 패키지의 모든 .go | 디스크 | 파일이 빠지면 검사가 침묵한다 |
| `mayNameTheBlock` | 정확한 파일명 2개 | 이 테스트 | 접두·접미 규칙을 쓰지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 파일 순회 | 없음 | 없음 | 자기 자신 |
| B2 | 면제 목록에 없는 파일이면 금지어 확장 | `banned`에 게이트 블록 이름 추가 | 없음 | `TestTheGateEditingExemptionIsNotIdle` |
| B3 | 금지어 순회 | 없음 | 없음 | 자기 자신 |
| B4 | 코드가 금지어를 담음 | 없음 | `t.Errorf` | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `packageFiles` | 패키지 전 파일 | 실패는 t.Fatal | CodeGraph + AST |
| `nonCommentLines` | 주석 제외 — 설명은 금지 대상이 아니다 | 없음 | CodeGraph + AST |

## State mutations and fallbacks

- 이 change가 바꾼 것: 금지어 목록을 **둘로 쪼갰다**. `Interlock`·`ProtectionReady`(기동 결정)는 전 파일에서 유지, `AutomationGate`·`automation_gate`(설정 블록)는 편집을 담당하는 두 파일에만 허용.
- 면제는 **정확 파일명 비교**다. 접두 일치를 쓰면 `settings_limits_helper.go`가 논증 없이 면제를 물려받는다 — `/orders` allowlist가 같은 이유로 정확 비교인 것과 같은 규율이다.
- 템플릿 파일은 면제 목록에 넣지 않았다. 대신 문안에서 `automation_gate` 표기를 없앴다 — 면제는 적을수록 좋다.

## Safety conclusion

- Safe edit boundary: 금지어 두 목록과 면제 파일 두 개. 면제를 늘리는 변경은 그 파일이 왜 블록을 이름해야 하는지 논증해야 한다.
- High-risk impact: yes(가드) — 이 검사가 느슨해지면 콘솔이 기동 인터록을 스스로 판정하는 코드를 들여도 아무도 모른다. `Interlock`·`ProtectionReady`를 전 파일에서 유지한 이유다.
