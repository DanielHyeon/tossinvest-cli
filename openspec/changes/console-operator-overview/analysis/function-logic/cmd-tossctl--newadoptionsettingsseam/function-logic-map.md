# Function Logic Map: `newAdoptionSettingsSeam`

- Source: `cmd/tossctl/adoptionsettings.go`
- AST evidence: `ast.json` (revision=current, L28–34, 분기 1개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `47672c6f` — 본문 변경: 경로 해석을 `configServiceFor`로 추출하고 nil 판정·래핑만 남김 (revision=current)

편입 설정 seam의 구체 타입을 만든다. 경로 해석 로직이 `configServiceFor`로 빠져나가면서 이 함수는 **nil 판정과 래핑만** 남았다 — 그 추출이 이 branch range의 변경 내용이다.

넘어가는 것은 `Load`/`Save` 두 메서드뿐이고, `*config.Service`는 이 구조체 안에 갇힌다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root *rootOptions` | nil 허용 | root flag | `configServiceFor`가 nil이면 nil |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `svc == nil` | 없음 | `nil` (구체 포인터 nil) | `TestATypedNilSeamNeverReachesTheInterface` |
| (else) | svc 확보 | 없음 | `*consoleAdoptionSettings` | `TestTheSeamSavesAuditsAndPreservesTheFile` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `configServiceFor` | 경로 해석을 한 곳으로 모은다 | nil이 유일한 실패 신호 | L29 |

## State mutations and fallbacks

- 상태 변이 없음. 저장·감사 기록은 `Save`/`recordAdoptionSave`가 한다.

## Safety conclusion

- Safe edit boundary: nil 판정. 구체 포인터를 인터페이스로 미리 올리는 편집은 typed-nil 함정을 되살린다.
- High-risk impact: yes (손절 파라미터 경로) — 이 seam이 저장하는 편입 블록에 `default_stop_pct`가 들어 있다. 주문 능력은 넘기지 않는다.
