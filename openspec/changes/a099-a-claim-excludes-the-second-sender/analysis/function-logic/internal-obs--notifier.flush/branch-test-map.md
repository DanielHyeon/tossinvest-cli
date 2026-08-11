# Branch Test Map: `Notifier.Flush`

**GREEN 칸은 실측해서 채운다.**
`ast.json`의 열거가 정본이다: 분기 6 · 이탈 4.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:428` journal이 배선 안 됐다 | 기존 — `obs_test.go`·`a097_exclusion_is_an_event_test.go`가 `Flush`를 부르는 두 파일이다. **정확한 테스트 함수는 §2.4가 실측한다** | no | **yes (기존)** |
| B2 | `:438` `PendingAlerts`가 실패한다 | 없음 — DB 오류 주입 없음 | no | **no (기존부터 없다)** |
| B3 | `:441` **밀린 행을 돈다 — 그리고 행마다 claim한다** | **a099 R4** | **planned RED — 미관측.** 오늘 claim을 아예 안 부른다. **단 1라운드 B-P8: 같은 `Notifier`로 쓰면 `n.mu` 때문에 오늘도 통과한다 — 별도 발송자로 preclaim해야 RED다** | no |
| B4 | `:442` publisher가 배선 안 됐다 | 기존 — `a097_exclusion_is_an_event_test.go:92` `TestFlushCannotPublishBesideASend` 계열 (§2.4가 실측) | no | **yes (기존)** |
| B5 | `:451` 발송이 실패한다 — 행은 PENDING으로 남는다 | 기존 + **a099 R6** | **R6은 회귀 핀 — RED 아님** | **yes (기존)** |
| B6 | `:455` 정산이 실패해 루프를 끊는다 | 기존 | no | **yes (기존)** |

## R4가 관측하는 것

**오늘**: `Flush`가 `PendingAlerts`로 받은 행을 **전부** publish한다.
다른 발송자가 이미 잡은 행이어도 상관하지 않는다 — claim 호출이 `ast.json`의
호출 목록에 **없다**.

**a099 이후**: 못 잡은 행은 publish 없이 건너뛴다.

프로덕션 호출자가 0이므로 이 RED는 **테스트가 직접 `Flush`를 부를 때만** 보인다.
그 사실이 이 편집의 위험을 가장 낮게 만든다.

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

- **B2** — 질의 오류 경로. a099가 안 건드린다.
  **`not-applicable`: 이 change는 B2를 근거로 아무것도 주장하지 않는다.**
- **B1 · B4 · B6** — 기존 테스트가 덮는다. a099가 조건도 이탈도 안 바꾼다.

## 덮이지 않은 것을 이름으로 적는다

- **`n.mu`가 루프 전체를 덮는다**(`:434-435`). a099는 안 건드린다.
  **a098의 배달 루프가 이 함수를 부르면 그 잠금이 exit 경로와 다시 만난다** —
  19라운드 A-P3 = B-P1이 그것이다. **`not-applicable`: a092의 표면이다.**
- **`:452`가 `MarkAlertAttemptFailed`의 반환을 버린다.** a099가 소유자 비교를
  더하면 「임차를 잃었다」가 여기서 조용해진다. contract 단계에서 이 함수와
  함께 사라진다.
- **`Publish`의 기한이 이 함수에 없다** — 임차 만료 값이 그것을 모르는 채로
  정해지면 안 된다. §3.4가 측정으로 잇는다.
