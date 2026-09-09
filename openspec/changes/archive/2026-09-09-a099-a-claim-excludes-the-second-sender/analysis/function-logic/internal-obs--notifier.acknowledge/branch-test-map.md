# Branch Test Map: `Notifier.Acknowledge`

**GREEN 칸은 실측해서 채운다.**
`ast.json`의 열거가 정본이다: 분기 10 · 이탈 6 · 호출 12 · defer 1.

**a099는 이 함수를 편집하지 않는다.** RED 칸이 전부 `no`인 것은 그래서다 —
이 산출물은 **design C6의 인용을 뒷받침하는 증거**이지 변경 계획이 아니다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:477` 운영자 이름이 공백이다 | `obs_test.go:461` `TestAcknowledgeRequiresAnIdentity` | no | **yes (기존)** |
| B2 | `:480` `n.Journal`이 nil이다 | **없음** | no | **no — 안 덮였다** |
| B3 | `:481` 원장 없이 게이트만 푼다 | **없음** | no | **no — 안 덮였다** |
| B4 | `:490` `ids`가 비어 밀린 것 전부를 승인한다 | `obs_test.go:427` `TestRecoveredDeliveryDoesNotReleaseTheGateByItself` | no | **yes (기존)** |
| B5 | `:492` `PendingAlerts`가 오류를 준다 | **없음 — DB 오류 주입 없음** | no | **no** |
| B6 | `:495` 밀린 행들의 id를 모은다 | `obs_test.go:427` | no | **yes (기존)** |
| B7 | `:499` id마다 승인한다 | `obs_test.go:485`·`:492` `TestAcknowledgeWhileStillPendingKeepsTheBlock` | no | **yes (기존)** |
| B8 | `:500` 승인이 `ErrAlertNotFound` **아닌** 오류를 준다 | **없음** | no | **no** |
| B9 | `:507` `UndeliveredCount`가 오류를 준다 | **없음** | no | **no** |
| B10 | `:510` **양방향** — 참(남은 수 0 → `Gate.Clear` `:511`)과 거짓(0이 아니면 **안 푼다**)을 각각 다른 테스트가 덮는다 | 참: `obs_test.go:427` `TestRecoveredDeliveryDoesNotReleaseTheGateByItself` · 거짓: `obs_test.go:470` `TestAcknowledgeWhileStillPendingKeepsTheBlock` | no | **yes (기존, 양방향)** |
| — | 이탈 `:513` 정상 종료 | 위 셋 | no | **yes (기존)** |
| — | (동시성) 발송 중 승인이 들어와도 `n.mu`가 가른다 | `a096_one_send_per_condition_test.go:397` | no | **yes (기존)** |

**B10의 두 방향이 다 덮여 있다는 것이 이 산출물의 결론이다.**
`TestRecoveredDeliveryDoesNotReleaseTheGateByItself`는 *"전달이 복구돼도 게이트는
스스로 안 풀린다"*를 단언하고, `TestAcknowledgeWhileStillPendingKeepsTheBlock`은
*"승인해도 남은 것이 있으면 안 풀린다"*를 단언한다.

> **⛔ 이 둘이 a099 3판의 C6을 반증한다.** 3판은 *"미전달 수가 0이 되면 사람 개입
> 없이 풀린다"*를 규범으로 적었는데, 그것이 참이면 **위 첫 테스트가 깨진다.**
> 사용자 결정 5-1이 그 규범을 되돌렸고, **그래서 이 표의 RED 칸이 전부 `no`다** —
> a099는 이 함수에서 아무것도 바꾸지 않는다.

**안 덮인 다섯(B2·B3·B5·B8·B9)은 a099가 안 건드리므로 `not-applicable`이다.**
그래도 이름을 적는 이유는 침묵한 생략이 금지이기 때문이고, 다음에 이 함수를 여는
change가 그 다섯을 물려받는다.
