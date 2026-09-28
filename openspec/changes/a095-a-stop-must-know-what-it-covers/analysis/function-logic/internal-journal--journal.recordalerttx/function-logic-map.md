# Function Logic Map: `Journal.recordAlertTx`

- Source: `internal/journal/outbox.go` (`276`–`364`)
- Qualified: `Journal.recordAlertTx`
- AST evidence: `ast.json` (`source_sha256` ad74f6c9b295467e…)
- Risk scan: `risk-pattern-report.md`
- 분기 8 · return 7 · 호출 14

**역할.** critical 알림을 outbox에 기록하고 전송 의무를 판정한다. 같은 event key의 행이 있으면 `claimOwed`에 묻고, 재무장(rearm)일 때만 그 행의 제목 · 본문 · payload를 새 발생으로 바꾼다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 기존 행의 `state` | PENDING · DELIVERED · ACKNOWLEDGED · 기타 | 원장 `alert_outbox` | B3 창 — `claimOwed`가 owed · rearm을 답함 |
| `rearm` | `claimOwed`의 둘째 답 | 아래 번들 | B4 — 참일 때만 본문 교체 UPDATE(B5) |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:280` `if err != nil {` | `RFC3339`, `Scan`, `j.clk.Now`, `tx.QueryRowContext` | :281 | 아니오 |
| B2 | switch | `:292` `switch {` | — | — | — |
| B3 | case | `:293` `case err == nil:` | `claimOwed` | — | 예 |
| B4 | if | `:295` `if rearm {` | — | — | 예 |
| B5 | if | `:334` `if _, uerr := tx.ExecContext(ctx,` | `fmt.Errorf`, `tx.ExecContext` | :343, :346 | — |
| B6 | case | `:347` `case !errors.Is(err, sql.ErrNoRows):` | `errors.Is`, `fmt.Errorf`, `tx.ExecContext` | :348 | 아니오 |
| B7 | if | `:355` `if err != nil {` | `fmt.Errorf`, `res.LastInsertId` | :356 | 아니오 |
| B8 | if | `:359` `if err != nil {` | `fmt.Errorf` | :360, :363 | 아니오 |

## Calls and live bindings

`tx.QueryRowContext`(기존 행) · `claimOwed`(B3 창) · `tx.ExecContext`(B5 — 재무장 UPDATE, 또는 B6 창 — 새 행 INSERT).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

재무장 시 기존 행을 PENDING으로 되돌리고 제목 · 본문 · payload · 시도 · 승인 · 임차를 새 발생으로 교체. 새 key면 INSERT.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** r3 N3(편입 성공 뒤 남는 PENDING 「무보호」 행)의 사실: 본문 교체는 B4(`rearm`)가 참일 때만 일어나고, `claimOwed`는 PENDING 행에 rearm을 주지 않는다(그 번들 B2). 따라서 **PENDING 행은 같은 key의 새 발생으로 본문이 바뀌지 않는다.** 사실이 해소되어 그 행을 정산하는 연산도 이 함수에 없다. N3의 처리는 `design.md` D1 「사실이 해소된 뒤의 행」과 Q8.
- **High-risk impact**: yes — critical 알림의 durable 기록 자리다.
