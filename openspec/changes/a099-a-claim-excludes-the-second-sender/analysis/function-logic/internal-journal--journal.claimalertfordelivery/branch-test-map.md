# Branch Test Map: `Journal.ClaimAlertForDelivery`

**GREEN 칸은 실측해서 채운다** — 덮이지 않은 것을 덮였다고 적지 않는다.
`ast.json`의 열거가 정본이다: 분기 11 · 이탈 10.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:173` event key 공백 | 기존 `outbox_test.go` | no | **yes (기존)** |
| B2 | `:176` type 공백 | 기존 | no | **yes (기존)** |
| B3 | `:183` BeginTx 실패 | 없음 — DB 오류 주입 없음 | no | **no (기존부터 없다)** |
| B4 | `:194` SELECT 결과로 갈린다 | 기존 (두 경로 다 있다) | no | **yes (기존)** |
| B5 | `:195` 기존 행을 찾았다 | **a099 R1** — `internal/journal/a099_…_test.go:41` `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` | **관측했다 (2026-08-11)** — `senders granted the right to send = 2, want 1`. 예상대로 **둘 다 `owed=true`를 받았다** | no |
| B6 | `:197` **재무장 판정 — PENDING이면 거짓이고 UPDATE가 없다** | **a099 R1 · R3** | **관측했다 (2026-08-11)** — 둘 다 `2, want 1`. PENDING이라 재무장 UPDATE가 안 돌고, **행에 발송자 표시가 하나도 안 남는다**(§3.1) | no |
| B7 | `:229` 재무장 UPDATE 실패 | 기존 (a097) | no | **yes (기존)** |
| B8 | `:241` SELECT가 ErrNoRows 아닌 오류 | 없음 | no | **no** |
| B9 | `:249` INSERT 실패 | 기존 (UNIQUE 충돌) | no | **yes (기존)** |
| B10 | `:253` LastInsertId 실패 | 없음 | no | **no** |
| B11 | `:256` commit 실패 | 없음 | no | **no** |

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

- **B3 · B8 · B10 · B11** — 전부 드라이버 오류 경로다. a099는 이 넷의 조건도 반환도
  안 바꾸므로 새 테스트를 요구하지 않는다. **`not-applicable`: 이 change는 이 넷을
  근거로 아무것도 주장하지 않는다.** 오늘 안 덮여 있다는 사실은 a099가 만든 것이 아니고
  a099가 고칠 것도 아니다 — 적어 두는 이유는 **침묵한 생략을 금지**하기 때문이다.
- B1·B2·B4·B7·B9는 기존 테스트가 덮는다. a099가 안 건드린다.

## 덮이지 않은 것을 이름으로 적는다

- **트랜잭션 격리 수준** — `BeginTx(ctx, nil)`(`:182`)은 드라이버 기본을 쓴다.
  a099의 임차 CAS가 **동시 트랜잭션 둘 사이에서** 실제로 배타적인지는 이 함수의
  분기가 아니라 SQLite의 잠금 모형이 정한다. **R3이 그것을 `-race`로 관측하고,
  관측 전에는 성립한다고 적지 않는다.**
- **`claimOwed`의 판정 자체** — 이 함수의 분기가 아니라 그 함수의 분기다.
  `internal-journal--claimowed` 번들이 진다.
