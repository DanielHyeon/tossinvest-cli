# Function Logic Map: `consoleGateLimitsSeam`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=current, L305–311, 분기 1개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — base에는 이 함수가 없었다. 현재 본문을 diff hunk가 덮으므로 evidence가 요구된다 (revision=current)

개요 화면에 Guardian 자동매매 게이트의 **천장 숫자만** 건네는 seam의 생성 지점이다.

- **넘어가는 것**: `console.GateLimitsReader` — 메서드 하나(`GateLimits`), 반환은 float64 다섯 개와 통화 문자열.
- **넘어가지 않는 것**: `*config.Service`(Init과 외과적 writer를 가진다), `config.AutomationGate` 타입 자체 (internal/console은 이 타입을 이름으로도 부를 수 없다 — `TestTheConsoleDecidesNothingAboutTheGate`). 즉 콘솔은 한도를 **읽을 수만** 있고 브라우저에서 위험 한도를 옮길 수 없다.
- **미배선 처리**: config 파일 경로를 해석할 수 없으면 seam이 nil이고 개요는 한도를 0이 아니라 "seam 미배선"으로 렌더한다. nil 판정을 **구체 타입에서** 하는 이유는 `consoleSettingsSeam`과 같다 — 인터페이스 안의 typed-nil은 배선된 것처럼 보인다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root *rootOptions` | nil 허용 | `--config-dir` 또는 `config.DefaultPaths()` | 해석 실패 → `configServiceFor`가 nil → seam nil |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `svc == nil` (config 경로 해석 실패) | 없음 | `nil` — 개요는 seam 미배선으로 렌더 | `TestTheConsoleComesUpWithoutTheLimitsSeam` (콘솔이 그래도 뜬다) |
| (else) | svc 확보 | 없음 | `consoleGateLimits{svc}` (값 타입) | `TestTheConsoleIsHandedTheLimitsAsNumbersAndNoWayToWriteThem` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `configServiceFor` | 편입 설정 화면과 **같은 파일**을 읽게 만든다 | nil 반환이 유일한 실패 신호 | adoptionsettings.go L41; `TestTheSeamSavesAuditsAndPreservesTheFile` |

## State mutations and fallbacks

- 상태 변이 없음. 값 타입 `consoleGateLimits`를 돌려주므로 typed-nil 문제가 구조적으로 없다.
- 쓰기 경로가 없다는 것이 설계다: 게이트를 켜거나 한도를 옮기는 것은 브라우저 밖의 §0.7 사람 결정이다.

## Safety conclusion

- Safe edit boundary: nil 판정과 반환 타입. 두 번째 메서드를 추가하는 것은 콘솔에 위험 한도 편집 능력을 주는 변경이다.
- High-risk impact: yes (Guardian 경로) — 읽기 전용이지만 Guardian 천장을 다루는 이음매다. 여기에 Save 계열 메서드를 하나 붙이는 편집이면 콘솔이 브라우저에서 위험 한도를 옮길 수 있게 된다. 주문 능력은 넘기지 않는다.
