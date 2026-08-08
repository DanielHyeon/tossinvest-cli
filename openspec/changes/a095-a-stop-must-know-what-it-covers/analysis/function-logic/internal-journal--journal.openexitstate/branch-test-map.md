# Branch Test Map: `Journal.OpenExitState`

- Source: `internal/journal/exit_state.go`

> **「진입 실측」은 측정값이다** — 네 패키지의 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a095가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:119` `if id == "" {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B2 | `:123` `if kind == "" {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B3 | `:126` `if kind != ExitPolicyRatchet && kind != ExitPolicyLadder {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B4 | `:131` `switch kind {` | — | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B5 | `:132` `case ExitPolicyRatchet:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B6 | `:133` `if policyID != "" {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B7 | `:136` `case ExitPolicyLadder:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B8 | `:137` `if policyID == "" {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B9 | `:139` `} else if _, ok := exitpolicy.CommonPolicyByID(policyID); !ok && policyID != "default_v1" {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B10 | `:139` `} else if _, ok := exitpolicy.CommonPolicyByID(policyID); !ok && policyID != "default_v1" {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B11 | `:147` `if err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B12 | `:153` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B13 | `:163` `if errors.Is(err, sql.ErrNoRows) {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B14 | `:166` `if err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B15 | `:169` `if !position.ExitEligible(decisionID.String, adoptionID.String) {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B16 | `:174` `if err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B17 | `:178` `if _, err := tx.ExecContext(ctx, `` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B18 | `:187` `if isUniqueViolation(err) {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B19 | `:192` `if err := appendExitEventTx(ctx, tx, exitEventRow{` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B20 | `:198` `if err := tx.Commit(); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |

**미진입 분기 4개**: B1, B6, B12, B13
**자체 블록 없는 분기 1개**: B4 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
