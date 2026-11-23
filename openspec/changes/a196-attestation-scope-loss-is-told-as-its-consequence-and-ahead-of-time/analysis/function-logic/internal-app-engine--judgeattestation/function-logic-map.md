# Function Logic Map: `judgeAttestation`

> **편집 후 지도 — 2단계 GREEN 뒤(2026-11-23).** 편집 전 원본 `../../pre-edit-ast/internal-app-engine--judgeattestation.ast.json`, 재번호 `renumber.txt`(옛/새 분기를 (종류, 줄 문장)으로 difflib 정렬 — 생성기 `../../renumber.py.txt`).

- Source: `internal/app/engine/attestation_watch.go` (98-142)
- AST evidence: `ast.json` — branches **13** / returns **4** / calls **11**, source SHA-256 `b3b8bb818e6e8837eb836cbd1be2a741004f7c114ba4282a13c44c4a99941bb8`(파일 단위, 편집 후).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/attestation_watch.go --func judgeAttestation`.
- Risk scan: `risk-pattern-report.md`.
- a166 census: `attestation_watch.go`·`a135_attestation_watch_internal_test.go` 0건(1.4).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `att`·`loadErr`·`now`·`accountRef` | 편집 전과 같음 | `attest.Load`·감시 시계·엔진 계좌 | B1 → unreadable |
| 불변식 | 판정 술어 = 인터록 술어(B1~B8 무변경) | engine-safety | — |
| 불변식(a196) | 기한 = min(만료, 실린 증거 중 가장 이른 것 + (만료 − 발급)) — `restartDeadline`(새 함수). Verify 통과면 만료는 아직 미래이므로 남은 시간 ≤ 0 은 감독 증거의 기한뿐 | design D1·D3 | B9 → proof_overdue |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@99) | `loadErr != nil` | 없음 | return@100 unreadable | `TestA135UnusableAttestationsAreCritical` |
| B2 (if@104) | Verify 거절 | — | — | 아래 B4~B8 |
| B3 (switch@105) | 오류 종류 | — | — | — |
| B4 (case@106) | `ErrExpired` | expired | return@119 | `TestA135StageBoundaries`「만료 순간」 |
| B5 (case@108) | `ErrEndpointNotAttested` | scope_lost + Missing(정렬) | return@119 | `TestA135ScopeLossIsCritical`·`TestA196ScopeLossOncePerDayAcrossReissues` |
| B6 (case@111) | `ErrAccountMismatch` | account_mismatch | return@119 | `TestA135UnusableAttestationsAreCritical` |
| B7 (case@113) | `ErrIncomplete` | incomplete | return@119 | 같은 시험 |
| B8 (case@115) | 그 밖 | unreadable | return@119 | 없음(도달 불가, 편집 전과 같음) |
| (L123-124) | Verify 통과 | `v.Deadline, v.DeadlineKind = restartDeadline(att, v.ProofAt)`, `remaining := v.Deadline.Sub(now)` | — | `TestA196ProofDeadlineGovernsTheDayCount`·`TestA196WithoutSupervisedProofCountsTheExpiry` |
| B9 (if@125) **새** | `remaining <= 0` | `v.Stage = proof_overdue` | return@129 | `TestA196DeadlineBoundaries`(기한 정각·V+1ns·하루 반)·`TestA196OverdueIsCriticalOncePerDay`·`TestA135StageBoundaries`「기한 정각」 |
| (L132) | remaining > 0 | `v.DaysLeft = daysLeftCeil(remaining)` — 양수에만 호출 | — | — |
| B10 (switch@133) | 날수 단계 | — | — | — |
| B11 (case@134) | `DaysLeft > 7` | healthy | return@141 | StageBoundaries「D-8」·`TestA196ProofDeadlineGovernsTheDayCount`(7d1s) |
| B12 (case@136) | `DaysLeft > 2` | notice | return@141 | StageBoundaries·`TestA196DeadlineBoundaries`「2일 1초」 |
| B13 (case@138) | 그 밖 | imminent | return@141 | `TestA196DeadlineBoundaries`「정확히 2일」~「V−1ns」 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `binding`(call@102) | 실린 증거 중 가장 이른 것 | 순수, 무편집 | AST |
| `att.Verify`(call@104) | 인터록 술어 | 무편집 | cited-ast |
| `restartDeadline`(call@123) **새 함수** | 기한과 종류 | 순수. 증거 0 · 유효 기간 ≤ 0 → 만료 | 새 leaf(번들 대상 아님) |
| `daysLeftCeil`(call@132) | 하루 단위 올림 | 양수 전제를 지킴(B9 가 먼저 거름) — 무편집 | AST |

## State mutations and fallbacks

- 지역 `v` 만. 새 필드 `Deadline`·`DeadlineKind` 는 Verify 통과 판정에서만 채움(거절 단계는 0 — 날짜로 셈).
- 편집 전 대비: 분기 1 추가(B9), 이후 분기 재번호(renumber.txt).

## Safety conclusion

- Safe edit boundary: Verify 통과 뒤만 바뀜 — B1~B8 문장 불변(renumber same).
- High-risk impact: no — 관측 전용.
