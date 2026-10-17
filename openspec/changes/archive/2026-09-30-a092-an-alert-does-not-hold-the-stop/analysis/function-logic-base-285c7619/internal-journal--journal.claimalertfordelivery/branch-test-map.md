# Branch Test Map: `Journal.ClaimAlertForDelivery`

Source: `internal/journal/outbox.go` (169-261). AST 기준 분기 11 / 이탈 10 /
defers 1 / go_statements 0.

`Test` 열은 **실제로 존재하는 테스트**만 적는다. 없으면 "없음"이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:173` `EventKey` 공백 → 오류 | **없음** — 이 함수에 공백 키를 주는 테스트가 없다 | no | no |
| B2 | `:176` `Type` 공백 → 오류 | **없음** | no | no |
| B3 | `:183` `BeginTx` 실패 | **없음** — 닫힌 DB 주입 테스트가 없다 | no | no |
| B4 | `:194` switch 진입 | 아래 B5·B8이 대표 | — | — |
| B5 | `:195` 기존 행이 있다 | `a096_claim_for_delivery_test.go` `TestClaimingADeliveredAlertInsideTheWindowOwesNothing:50` | no | yes |
| B6 | `:197` **재무장** | `TestClaimingADeliveredAlertPastTheWindowIsReArmed:91`, `a096b_round2_test.go TestReArmingClearsThePreviousAcknowledgement:75`, `a097_rearm_is_a_new_episode_test.go TestReArmingCarriesTheCauseThatReArmedIt:44`·`TestReArmingResetsTheAttemptCount:89`·`TestReArmingClearsThePreviousEpisodeTimestamps:146` | yes (a097) | yes |
| B7 | `:229` 재무장 UPDATE 실패 | **없음** | no | no |
| B8 | `:241` 조회가 `ErrNoRows`가 아닌 오류 | **없음** | no | no |
| B9 | `:249` INSERT 실패 | **없음** | no | no |
| B10 | `:253` `LastInsertId` 실패 | **없음** | no | no |
| B11 | `:256` `Commit` 실패 | **없음** | no | no |
| — `:260` | 새 행 → `owed=true` | `TestClaimingAFreshAlertOwesDelivery:27` | no | yes |

## a092가 이 함수에 대해 지는 것

편집하지 않으므로 **이 함수 자체에 대한 새 RED는 없다.** 다만 17판의 D0.3이
B5/B6의 열거를 근거로 쓰므로, 그 열거가 실제 동시성에서 성립하는지는
**§6.0 R17-3**(claim × 배달 루프 인터리빙)이 관측한다. R17-3은 이 함수의 분기를
테스트하는 것이 아니라 **두 호출자가 이 함수를 동시에 부를 때의 결과**를
테스트한다 — 다른 대상이므로 위 표의 행을 채우지 않는다.

미테스트 분기 8개(B1·B2·B3·B7·B8·B9·B10·B11)는 전부 입력 검증이거나 SQLite 실패
주입이고, **a092의 범위가 아니다**(`not-applicable`: 이 change는 이 함수를
편집하지 않는다). B1·B2는 주입 없이도 테스트할 수 있다는 점에서 나머지와 다르며,
**a092가 만들지 않는다는 사실을 여기 남긴다.**
