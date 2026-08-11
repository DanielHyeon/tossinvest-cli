# Branch Test Map: `Notifier.Flush`

**GREEN 칸은 실측해서 채운다.**
`ast.json`의 열거가 정본이다: 분기 6 · 이탈 4.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:428` journal이 배선 안 됐다 | 기존 — `Flush`를 부르는 테스트는 **넷뿐이다**(§2.5 실측): `obs_test.go:440` `TestRecoveredDeliveryDoesNotReleaseTheGateByItself` · `obs_test.go:590` `TestCriticalAlertSurvivesAProcessRestart` · `a097_exclusion_is_an_event_test.go:109` `TestFlushCannotPublishBesideASend` · `a099_…_test.go:210` `TestFlushDoesNotPublishARowAnotherSenderHolds`(R4). **넷 다 journal을 배선한다 — B1의 참 쪽은 안 덮인다** | no | **아니오 — 거짓 쪽만 덮인다** |
| B2 | `:438` `PendingAlerts`가 실패한다 | 없음 — DB 오류 주입 없음 | no | **no (기존부터 없다)** |
| B3 | `:441` **밀린 행을 돈다 — 그리고 행마다 claim한다** | **a099 R4** — `a099_…_test.go:197` `TestFlushDoesNotPublishARowAnotherSenderHolds` | **관측했다 (2026-08-11)** — `publishes = 2, want 1`(§3.1). 1라운드 B-P8이 경고한 born-GREEN은 **피했다**: 발송자를 둘로 갈라 다른 뮤텍스를 쓰게 하고, 게이트한 publisher가 첫 발송을 붙잡아 `Flush`가 **반드시** 그 안에서 도착하게 했다 | no |
| B4 | `:442` publisher가 배선 안 됐다 | 기존 — `a097_exclusion_is_an_event_test.go:92` `TestFlushCannotPublishBesideASend`(§2.5 실측: 함수 정의 `:92`, `Flush` 호출 `:109`) | no | **yes (기존)** |
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
