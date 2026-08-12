# Branch Test Map: `Open`

**GREEN 칸은 실측해서 채운다.**
`ast.json`의 열거가 정본이다: 분기 15 · 이탈 8 · defer 0.

**a099가 더한 것은 분기 하나**(B9 `:163`, `AlertLease` 기본값)이고,
그 하나만 RED를 갖는다. 나머지 열넷은 **인용**이다.
§5.6 갱신: 분기 하나가 늘어 이후 ID가 하나씩 밀렸고, 줄 번호도 전부 다시 뽑았다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:129` `Path`가 비어 `DefaultPath()`로 간다 | 기존 (`DefaultPath` 테스트) | no | **yes (기존)** |
| B2 | `:131` 경로 해결이 실패한다 | **없음** | no | **no — 안 덮였다** |
| B3 | `:140` `FSProber`가 nil이다 | 기존 (FS 검사 테스트) | no | **yes (기존)** |
| B4 | `:146` 파일시스템이 거절된다 | 기존 | no | **yes (기존)** |
| B5 | `:150` `MkdirAll`이 실패한다 | **없음** | no | **no** |
| B6 | `:155` `Clock`이 nil이면 `clock.System()` | 기존 — 테스트가 전부 주입한다 | no | **부분 — 기본값 쪽 미검증** |
| B7 | `:159` `BusyTimeout`이 0 이하면 기본값 | `durability_test.go`의 `openTestJournalWithBusy`를 부르는 두 자리가 둘 다 `100ms`를 **준다**. **주입 쪽만** | no | **부분 — 기본값 쪽 미검증** |
| B9 | **`:163` `AlertLease`가 0 이하면 `DefaultAlertLease` — a099가 더한 분기** | `TestOpenFillsTheDefaultLease` `a099_lease_lifecycle_test.go:44` · `TestTheDefaultLeaseOutlastsTheDeliveryBound` `a099_lease_events_test.go:242` | **yes — §4.1 전에는 `DefaultAlertLease`가 없어 컴파일이 안 된다** | **yes** |
| B8 | `:168` `sql.Open`이 실패한다 | **없음** | no | **no** |
| B10 | `:178` `PingContext`가 실패한다 | **없음** | no | **no** |
| B11 | `:188` 손상된 DB | 기존 (integrity 테스트) | no | **yes (기존)** |
| B12 | `:193` `migrationOverride`가 있다 | 기존 (migration 테스트) · a099는 `TestAFailedV31MigrationLeavesTheJournalAtV30` `a099_regression_pins_test.go:257` | **yes (R26)** | **yes** |
| B13 | `:196` migration이 실패한다 | 기존 — `migration_v5_test.go` · `TestAFailedV31MigrationLeavesTheJournalAtV30` `a099_regression_pins_test.go:257` | **yes (R26)** | **yes** |
| B14 | `:202` 세 파일 경로를 훑는다 | 기존 | no | **yes (기존)** |
| B15 | `:203` 파일이 있으면 `0600`으로 | 기존 | no | **yes (기존)** |

이탈 `:207`(정상 개방)은 분기가 아니다. 사실상 모든 저널 테스트가 지나간다.

> **B9의 ID가 B7 다음이 아니라 B8 앞이다.** AST는 소스 순서로 번호를 매기고
> `AlertLease` 기본값 블록(`:162-165`)이 `sql.Open`(`:167`)보다 위에 있다.
> 표를 소스 순서로 정렬한 이유가 그것이다 — **ID를 읽는 순서가 코드를 읽는 순서다.**

> **R20의 RED는 「단언이 틀린다」가 아니라 「컴파일이 안 된다」였다.**
> `DefaultAlertLease`가 §4.1 전에는 존재하지 않는다.
> **컴파일 실패를 RED라고 부르는 것은 정직하다** — 그 시점에 그 단언은 쓸 수 없다.

> **B6·B7의 「기본값이 쓰인다」쪽이 오늘도 안 덮여 있다.**
> a099는 lease에 대해서만 그 구멍을 메웠다(B9). `Clock`·`BusyTimeout` 쪽은
> **`not-applicable`: a099 밖이다.** 이름을 적는 이유는 침묵한 생략이 금지이기 때문이다.

## 덮이지 않은 것을 이름으로 적는다

- **B2·B5·B8·B10에 테스트가 없다.** 경로 해결·디렉터리 생성·드라이버 개방·핑 실패다.
  전부 a099 이전부터 그렇고, **`not-applicable`: 이 change는 넷을 근거로
  아무것도 주장하지 않는다.**
- **B9는 「0 이하면 기본값」만 본다.** 음수 lease를 준 경우를 따로 보는 테스트는 없고,
  `lease <= 0` 한 조건이 둘을 같이 처리한다. `TestOpenFillsTheDefaultLease`는
  **0만** 준다.
