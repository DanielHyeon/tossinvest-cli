# Function Logic Map: `ReconcileDriver.alertUnmanaged`

- Source: `internal/app/engine/adoption.go`
- AST evidence: `ast.json` (구현 후 추출)
- Risk scan: `risk-pattern-report.md`

High-risk(reconciliation 알림). 사유 행렬 확장(design D2): exclude는 enabled 무관 표기, include 시도 실패는 실패를 말한다 — '꺼져 있다' 금지(P2-6). latch 로직 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (함수 서술 참조 — 이 change의 delta·design D1~D6이 계약) | — | 현재 HEAD + 위 테스트 | 테스트 실패로 관측 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if | 없음 | — | TestAFailedIncludeCycleSaysTriedNotOff(RED: 'adoption is off' 오표기)·TestIncludeDoesNotBypassExclusion·TestTheUnmanagedAlertSurvivesAdoptionBeingOff green |
| B2 | switch | 없음 | — | TestAFailedIncludeCycleSaysTriedNotOff(RED: 'adoption is off' 오표기)·TestIncludeDoesNotBypassExclusion·TestTheUnmanagedAlertSurvivesAdoptionBeingOff green |
| B3 | case | 없음 | — | TestAFailedIncludeCycleSaysTriedNotOff(RED: 'adoption is off' 오표기)·TestIncludeDoesNotBypassExclusion·TestTheUnmanagedAlertSurvivesAdoptionBeingOff green |
| B4 | case | 없음 | — | TestAFailedIncludeCycleSaysTriedNotOff(RED: 'adoption is off' 오표기)·TestIncludeDoesNotBypassExclusion·TestTheUnmanagedAlertSurvivesAdoptionBeingOff green |
| B5 | case | 없음 | — | TestAFailedIncludeCycleSaysTriedNotOff(RED: 'adoption is off' 오표기)·TestIncludeDoesNotBypassExclusion·TestTheUnmanagedAlertSurvivesAdoptionBeingOff green |
| B6 | case | 없음 | — | TestAFailedIncludeCycleSaysTriedNotOff(RED: 'adoption is off' 오표기)·TestIncludeDoesNotBypassExclusion·TestTheUnmanagedAlertSurvivesAdoptionBeingOff green |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| ast.json calls 참조 | 목적 서술 참조 | 기존 계약 무변경 | ast.json + HEAD |

## State mutations and fallbacks

- 없음(문구 선택만; latch map 갱신은 기존 그대로)

## Safety conclusion

- Safe edit boundary: why switch의 분기 확장만
- High-risk impact: yes — reconciliation/audit 경로, Pre-Edit 선언·적대적 리뷰 하에 수정
