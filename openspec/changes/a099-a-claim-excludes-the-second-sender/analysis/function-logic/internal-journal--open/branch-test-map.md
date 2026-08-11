# Branch Test Map: `Open`

**GREEN 칸은 실측해서 채운다.**
`ast.json`의 열거가 정본이다: 분기 14 · 이탈 8 · 호출 24 · defer 0.

**a099가 더하는 것은 분기 하나**(B7 바로 뒤, `AlertLease` 기본값)이고,
그 하나만 RED를 갖는다. 나머지 열넷은 **인용**이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:110` `Path`가 비어 `DefaultPath()`로 간다 | 기존 (`DefaultPath` 테스트) | no | **yes (기존)** |
| B2 | `:112` 경로 해결이 실패한다 | **없음** | no | **no — 안 덮였다** |
| B3 | `:121` `FSProber`가 nil이다 | 기존 (FS 검사 테스트) | no | **yes (기존)** |
| B4 | `:127` 파일시스템이 거절된다 | 기존 | no | **yes (기존)** |
| B5 | `:131` `MkdirAll`이 실패한다 | **없음** | no | **no** |
| B6 | `:136` `Clock`이 nil이면 `clock.System()` | 기존 — 테스트가 전부 주입한다 | no | **부분 — 기본값 쪽 미검증** |
| B7 | `:140` `BusyTimeout`이 0이면 기본값 | `durability_test.go:903` `openTestJournalWithBusy`(정의) — 부르는 자리는 `:616`·`:617`뿐이고 둘 다 `100ms`를 **준다**. **주입 쪽만** | no | **부분 — 기본값 쪽 미검증** |
| B7+ | **신설** — `AlertLease`가 0이면 `DefaultAlertLease` (§4.1b) | **a099 R20** — `DefaultAlertLease > bound(기본 설정)` | **yes — §4.1b 전에는 상수가 없어 컴파일이 안 된다** | (§4.1b 뒤) |
| B8 | `:145` `sql.Open`이 실패한다 | **없음** | no | **no** |
| B9 | `:155` `PingContext`가 실패한다 | **없음** | no | **no** |
| B10 | `:165` 손상된 DB | 기존 (integrity 테스트) | no | **yes (기존)** |
| B11 | `:170` `migrationOverride`가 있다 | 기존 (migration 테스트) | no | **yes (기존)** |
| B12 | `:173` migration이 실패한다 | 기존 — `migration_v5_test.go` | no | **yes (기존)** |
| B13 | `:179` 세 파일 경로를 훑는다 | 기존 | no | **yes (기존)** |
| B14 | `:180` 파일이 있으면 `0600`으로 | 기존 | no | **yes (기존)** |
| — | 이탈 `:184` 정상 개방 | 기존 (거의 모든 테스트) | no | **yes (기존)** |

> **R20의 RED는 「단언이 틀린다」가 아니라 「컴파일이 안 된다」다.**
> `DefaultAlertLease`가 §4.1b 전에는 존재하지 않는다. 그 실패 출력을 §3.1이 붙인다.
> **컴파일 실패를 RED라고 부르는 것은 정직하다** — 그 시점에 그 단언은 쓸 수 없다.

> **B6·B7의 「기본값이 쓰인다」쪽이 오늘 안 덮여 있다.**
> a099는 lease에 대해서만 그 구멍을 메운다. `Clock`·`BusyTimeout` 쪽은
> **`not-applicable`: a099 밖이다.** 이름을 적는 이유는 침묵한 생략이 금지이기 때문이다.
