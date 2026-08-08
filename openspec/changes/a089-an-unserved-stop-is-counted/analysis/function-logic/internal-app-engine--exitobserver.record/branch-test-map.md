# Branch Test Map: `ExitObserver.record`

AST 기준 분기 14 / 이탈 5. 기존 테스트는 `internal/app/engine/exitloop_test.go`,
`a084b_rejudge_bound_test.go`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:1095` `quote.FetchedAt`가 zero → 관측 출처 `cycle` | 간접 (harness 기본 경로) | no | yes |
| B2 | `:1097` `FetchedAt` 존재 → 출처 `quote_fetched_at` | 간접 | no | yes |
| B3 | `:1117` 주문 가능 + (`CancelPendingFirst` 또는 `isFullExit`) → 심볼 정리 진입 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` `:861` | no | yes |
| B4 | `:1118` 재판정 중 익절은 보류(`ArmSuppressedReJudge`) | `a084b_rejudge_bound_test.go` | no | yes |
| B5 | `:1140` 평시 → `clearTheSymbol` 호출 | `:861` | no | yes |
| B6 | `:1142` `clearTheSymbol` 오류 → `return err` `:1143` | **없음** | no | no |
| B7 | `:1145` 정리 실패 → **`noteDelay`** + 발의 보류 | `TestAnUncancellableEntryWithholdsTheLiquidationAndAlertsPastTheBound` `:883` | no | yes |
| B8 | `:1149` 정리 성공 → **`clearDelay`** `:1150` | `:914-919` (취소 성공 후 청산이 나간다) | no | yes |
| B9 | `:1156` 주문 가능 → intent id 발급 | 다수 | no | yes |
| B10 | `:1158` decision id에서 intent id를 못 얻으면 새로 발급 | 간접 | no | yes |
| B11 | `:1170` `RecordExitJudgementResult` 오류 | 간접 | no | yes |
| B12 | `:1171` `ErrProposalPending` → `return nil` `:1175` | **없음** | no | no |
| B13 | `:1177` `ErrExitSnapshotQuarantined` → 알림 후 `return err` `:1188` | a074 계열 | no | yes |
| B14 | `:1190` `ArmOutcome != Armed` → `return nil` `:1191` | **없음** | no | no |

## 지연 시계의 두 원인이 여기서 갈린다

B7과 B8은 **같은 조건의 양변**이다 — `clearTheSymbol`의 성공 여부.

| 미제출 원인 | `cleared` | 시계 | 근거 |
|---|---|---|---|
| working order를 못 치웠다 (B7) | false | **시작·누적** | `:883` 테스트가 31초 후 critical 확인 (`:900`,`:906`,`:910`) |
| 브로커가 거부했다 (`submit` B10) | **true** | B8이 매 주기 초기화 | 거부된 주문은 살아 있지 않다 |

**`:1150`을 단독 제거하면 B7이 시작한 시계의 유일한 해제점이 사라진다.**
해제를 `submit`의 `StateConfirmed`로 옮기는 설계라면 `:914-919`가 그 대체의 회귀 테스트다.

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R1 | 살아 있는 주문 없이 보호 제출이 거부됨 | B8이 시계를 지우지 **않는다** (1라운드 C1의 회귀) |
| R2 | B7이 시작한 시계가 `StateConfirmed`에서 해제된다 | `:883`·`:914-919`가 그대로 통과 |
| R3 | B12(`ErrProposalPending`) 시 미제출로 **계수하지 않는다** | 접수 성공 후에도 `pending_action`이 남으므로 |
| R4 | B14(`ArmOutcome != Armed`) 시 미제출로 **계수하지 않는다** | 선행 제안이 주문을 가질 수 있으므로 |

## 미측정

`arm_suppressed_reason`이 `exit_events` **671행 전부 공백**, `effective_source`에 `saved`
**0건** → B4·B7·B14는 이 배포에서 한 번도 발화한 적이 없다 `[미측정]`.
