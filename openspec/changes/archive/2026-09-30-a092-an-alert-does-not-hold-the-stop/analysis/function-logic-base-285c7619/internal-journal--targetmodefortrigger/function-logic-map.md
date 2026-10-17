# Function Logic Map: `TargetModeForTrigger`

- Source: `internal/journal/operating_mode.go` (537-549)
- AST evidence: `ast.json` — branches 3, returns 2, calls 1, assignments 0,
  defers 0, go_statements 0
- Risk scan: `risk-pattern-report.md`

**4라운드 H4의 자리다.** 네 문서가 이 함수의 case들에 대해 **전칭 주장**("위 네 트리거의
목표는 전부 `ModeEntryBlocked`다")을 근거로 썼는데 산출물이 없었다.
주장 자체는 참이지만 **방법이 금지된 것**이었다 — 3라운드 H2(b)의 네 번째 재발이다.
a092는 이 함수를 편집하지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `trigger` | 열거된 6개 또는 그 밖 | 호출자 | 그 밖이면 B3이 `("", false)` |

## Branches and early returns

| Branch | 위치 | 조건 | Return | 목표 모드 |
|---|---|---|---|---|
| B1 | `:538` | `switch strings.TrimSpace(trigger)` | — | — |
| B2 | `:539` | **6개 트리거를 한 case에 모은다** — `ModeTriggerDailyLossLimit` · `ModeTriggerCredentialRejected` · `ModeTriggerCriticalAlertUndelivered` · **`ModeTriggerExitObservationOutage`** · `ModeTriggerReconcileCycleFailure` · `ModeTriggerFillDetectionFailure` | `:545` | **`ModeEntryBlocked`, true** |
| B3 | `:546` | `default` | `:547` | `("", false)` |

### 전칭 주장이 성립하는 이유는 case가 하나이기 때문이다

**AST branches 3 · returns 2.** 값을 돌려주는 case는 **B2 하나**이고 그것이 6개 트리거를
전부 담는다. 그러므로 "모든 자동 트리거의 목표가 `ENTRY_BLOCKED`"는 case별로 확인할
필요가 없다 — **분기 구조 자체가 그것을 말한다.** `HALT_ALL`로 가는 case는 존재하지
않으며, 그것은 `TransitionOperatingMode` B8(`:373`)의
`ErrHaltAllIsNeverAutomatic`과 정합한다.

주석(`:531-535`)도 같은 말을 하고, 함수가 상수 대신 함수로 존재하는 이유를
"트리거 추가가 한 표의 눈에 보이는 편집이 되게, 그리고 테스트가 전체 집합에 대해
`HALT_ALL`에 닿는 것이 없음을 단언할 수 있게"라고 적었다.

## Calls and live bindings

| Callee | 위치 | 계약 | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | `:538` | 즉시 | AST calls |

**네트워크도 원장도 없다.** 순수 함수다.

## State mutations and fallbacks

- 없음. AST assignments 0, defers 0, go_statements 0.

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: no (순수 매핑) — 다만 **H1 헤드라인 두 개가 이 함수 위에 서 있다.**
- **a092가 여기서 얻는 사실**: `ModeTriggerExitObservationOutage`의 목표가
  `ENTRY_BLOCKED`이고, 계정이 이미 그 모드이므로 `TransitionOperatingMode` B15가
  announce 전에 반환한다 ⇒ **오늘 두절 사이클은 34초이지 68초가 아니다.**
  그 사슬의 첫 고리가 이 함수이고, 이제 열거로 뒷받침된다.
