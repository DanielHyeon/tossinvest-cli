# Branch Test Map: `OpenReadOnly`

`ast.json`의 열거가 정본이다: 분기 10 · 이탈 8.
**a099가 이 함수에서 바꾼 것은 식별자 하나다** (`defaultBusyTimeout` →
`DefaultBusyTimeout`, `:183`). **아래 표는 전부 인용이고 RED는 하나도 없다.**

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:156` `Path`가 비어 `DefaultPath()`로 간다 | 기존 | no | **yes (기존)** |
| B2 | `:158` 경로 해결 실패 | 없음 | no | **no** |
| B3 | `:168` `os.Stat` 실패 | 기존 (`ErrJournalMissing` 테스트) | no | **yes (기존)** |
| B5 | `:169` 파일이 없다 → `ErrJournalMissing` | 기존 | no | **yes (기존)** |
| B4 | `:173` B3의 `else` | 기존 | no | **yes (기존)** |
| B6 | `:173` 디렉터리다 → `ErrJournalMissing` | 기존 | no | **yes (기존)** |
| B7 | `:183` **`busy <= 0`이면 `DefaultBusyTimeout`** | **없음 — 기본값 쪽 미검증** | no | **no** |
| B8 | `:188` `sql.Open` 실패 | 없음 | no | **no** |
| B9 | `:197` `PingContext` 실패 | 없음 | no | **no** |
| B10 | `:203` 스키마가 이 빌드보다 새롭다 | 기존 (스키마 상한 테스트) | no | **yes (기존)** |

이탈 `:207`(정상 개방)은 분기가 아니다. 콘솔 조회 테스트가 전부 지나간다.

## a099가 이 함수에서 실제로 한 일

| 무엇 | 값 | 분기 | 이탈 |
|---|---|---|---|
| base | `defaultBusyTimeout` = 5초 | 10 | 8 |
| 지금 | `DefaultBusyTimeout` = 5초 | 10 | 8 |

**셋 다 같다.** 이 번들은 그것을 보이기 위해 존재한다.

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

- **열 전부.** a099는 이 함수의 어떤 판정도 안 바꿨다.
  **`not-applicable`: 이 change는 이 함수의 분기를 근거로 아무것도 주장하지 않는다.**

## 덮이지 않은 것을 이름으로 적는다

- **B2 · B7 · B8 · B9에 테스트가 없다.** 전부 a099 이전부터 그렇다.
  특히 **B7의 「기본값이 쓰인다」쪽**은 `Open`의 같은 분기와 짝을 이루는 공백이다.
  a099는 `AlertLease`에 대해서만 그 종류의 공백을 메웠다.
- **`DefaultBusyTimeout`을 읽는 곳이 패키지 밖에 생겼다** —
  `obs/alert_lease.go`. 이 상수를 바꾸면 **임차 길이(81초)의 유도가 따라 움직인다.**
  `TestTheDefaultLeaseOutlastsTheDeliveryBound` `a099_lease_events_test.go:242`가
  그 관계를 고정하지만, **이 번들의 어떤 분기도 그것을 안 본다.**
