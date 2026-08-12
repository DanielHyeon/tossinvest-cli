# Branch Test Map: `Journal.MarkAlertDelivered`

**GREEN 칸은 실측해서 채운다.**
`ast.json`의 열거가 정본이다: 분기 1 · 이탈 2.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:344` `ExecContext`가 실패한다 | 없음 — DB 오류 주입 없음 | no | **no (기존부터 없다)** |
| 이탈 `:347` | **PENDING 행을 DELIVERED로 옮긴다 — 정확히 한 번** | 기존 `outbox_test.go` + **a099 R13** | **R13은 planned RED — 미관측.** 1라운드 A-P4·B-P5: 소유자 문자열은 fencing token이 아니므로 **같은 문자열로 재취득한 ABA를 못 잡는다** — 토큰 단위 관측은 **R14**가 따로 진다 | no |

## ⚠ 이 CAS는 이중 **발송**을 막지 않는다 — 그것이 a099의 근거다

`WHERE id = ? AND state = ?`(`:342-343`)는 진짜 CAS다. **그러나 순서가 늦다.**

| 순간 | 무슨 일 |
|---|---|
| 1 | 발송자 둘이 같은 PENDING 행에 `owed=true`를 받는다 |
| 2 | **둘 다 publish한다** — 운영자 전화기에 푸시 둘 |
| 3 | 하나가 이 CAS로 DELIVERED를 쓴다 |
| 4 | 다른 하나가 0행을 받아 `ErrAlertNotFound`를 본다 |

**4단계의 오류가 2단계를 되돌리지 않는다.** 2026-08-08의 `no such alert` 줄이
그 4단계였다(a096 round 1, blocker 1의 기록).

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

- **B1** — 드라이버 오류 경로. a099가 조건도 반환도 안 바꾼다.
  **`not-applicable`: 이 change는 B1을 근거로 아무것도 주장하지 않는다.**
  덮여 있지 않다는 사실은 a099가 만든 것이 아니다.

## 덮이지 않은 것을 이름으로 적는다

- **소유자 비교가 만드는 새 실패**: 임차를 잃은 발송자가 publish에 **성공한 뒤**
  0행을 받는 경로. 그 발송자의 푸시는 이미 나갔고 행은 PENDING으로 남는다.
  **R13이 이것을 관측하되 「막는다」고 적지 않는다** — at-least-once의 대가다
  (design D8).
- `requireOneRow`(`outbox.go:465-474`)의 분기는 그 함수의 것이다. a099는 안 건드린다.
