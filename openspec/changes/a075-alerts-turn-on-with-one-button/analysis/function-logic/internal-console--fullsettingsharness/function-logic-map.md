# Function Logic Map: `fullSettingsHarness`

- Source: `internal/console/settings_cadence_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a075-alerts-turn-on-with-one-button/base-commit.txt`
- 위험 등급: Normal (test)

## Inputs and invariants

카드 표준 검사(`TestEveryCardShowsWhatChangesAndWhen`,
`TestEveryCardEitherSavesOrSaysWhyNot`)는 **모든 seam이 배선된** 콘솔과 **아무것도
배선되지 않은** 콘솔 양쪽을 돈다. 새 카드가 앞쪽에서 검사되려면 그 harness가 알림
seam을 알아야 한다 — 모르면 카드가 계속 미배선 상태로 렌더되고 새 카드는 표준 검사를
사실상 통과하지 못한 채 통과한다.

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `tweak ...func(*Options)` | 0개 이상 | 호출자 | 뒤에 오는 tweak이 앞을 덮는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| — | 분기 없음 | 없음 | — | 행복 경로 1건 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newHarness` | 콘솔 인스턴스 | 없음 | AST |
| 각 fake 생성자 | seam 주입 | 없음 | AST |

## State mutations and fallbacks

- 테스트 지역 `Options`만 만든다.
- a075가 더한 줄은 `o.Notifications = &fakeNotifications{}` 하나이며 기존 주입의 순서를 바꾸지 않는다.
- `tweak`이 뒤에 적용되므로 개별 테스트가 nil로 되돌려 미배선 경로를 볼 수 있다 — `TestAnUnwiredAlertSeamSaysWhyInsteadOfOfferingAButton`이 그렇게 한다.

## Safety conclusion

- Safe edit boundary: 주입 한 줄.
- High-risk impact: **no** — 테스트 하네스다.
- 이 편집이 없으면 카드 표준 검사가 새 카드에 대해 아무것도 확인하지 않는다.
