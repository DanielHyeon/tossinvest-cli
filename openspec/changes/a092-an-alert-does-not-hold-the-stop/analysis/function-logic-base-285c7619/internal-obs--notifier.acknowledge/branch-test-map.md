# Branch Test Map: `Notifier.Acknowledge`

Source: `internal/obs/notifier.go` (476-514). AST 기준 분기 10 / 이탈 6 /
defers 1 / go_statements 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:477` 조작자 신원 없음 → 거부 | `internal/obs/obs_test.go TestAcknowledgeRequiresAnIdentity:461` | no | yes |
| B2 | `:480` journal이 nil → 게이트만 풀고 끝 | **없음** | no | no |
| B3 | `:481` 그 경로에서 게이트가 있다 | **없음** | no | no |
| B4 | `:490` `ids` 없이 호출 → 미전달 전부 | `obs_test.go TestRecoveredDeliveryDoesNotReleaseTheGateByItself:451`, `a096_one_send_per_condition_test.go TestAcknowledgeCannotClearTheGateMidSend:397` | no | yes |
| B5 | `:492` `PendingAlerts` 실패 | **없음** — 닫힌 DB 주입 테스트가 없다 | no | no |
| B6 | `:495` 조회 결과를 `ids`로 모은다 | B4와 같은 테스트 | no | yes |
| B7 | `:499` id마다 승인 | `obs_test.go TestAcknowledgeWhileStillPendingKeepsTheBlock:470`(`:485`·`:492`가 id를 명시) | no | yes |
| B8 | `:500` 승인 실패가 `ErrAlertNotFound`가 아니다 → 오류 | **없음** | no | no |
| B9 | `:507` `UndeliveredCount` 실패 | **없음** | no | no |
| B10 | `:510` **남은 게 0일 때만 게이트 해제** | `TestAcknowledgeWhileStillPendingKeepsTheBlock:470`(0이 아니면 안 푼다), `TestRecoveredDeliveryDoesNotReleaseTheGateByItself:427`(배달만으로는 안 풀린다) | no | yes |

## a092가 이 함수에 대해 지는 것

본문을 편집하지 않으므로 분기에 대한 새 RED는 없다. **a092가 지는 것은
이 함수가 프로덕션에서 도달 가능한가**이고, 그것은 위 표의 어느 행도 아니다.

- **§6.0 R17-11** — 배달 루프가 걸어 둔 래치를 `Acknowledge`가 푼다.
  B10을 17판의 배선 위에서 다시 관측한다.

**`TestAcknowledgeCannotClearTheGateMidSend:397`이 17판의 위험을 이미 적어 뒀다.**
그 테스트는 `n.mu`가 해제와 발송을 가른다는 것에 의존한다. 17판이
`claimAndDeliver`에서 잠금을 좁힐 때 **이 함수의 잠금은 유지해야** 하고,
배달 루프도 정산 구간에서 같은 배제를 받아야 한다. 그러지 않으면
"세고 나서 푸는 사이에 재무장이 끼어든다"가 되살아난다 — **게이트가 존재하는
이유가 무너지는 유일한 경로다.** §8 GREEN의 조건으로 tasks.md에 남긴다.

미테스트 분기 5개(B2·B3·B5·B8·B9)는 nil 배선이거나 SQLite 실패 주입이고,
a092는 이 함수의 본문을 편집하지 않으므로 여기서 만들지 않는다
(`not-applicable`: 본문 미편집). B2·B3은 주입 없이도 만들 수 있다는 점에서
나머지와 다르며, **a092가 만들지 않는다는 사실을 남긴다.**
