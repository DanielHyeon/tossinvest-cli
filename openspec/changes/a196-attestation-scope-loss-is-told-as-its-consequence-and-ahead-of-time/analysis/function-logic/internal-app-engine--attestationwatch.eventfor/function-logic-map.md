# Function Logic Map: `attestationWatch.eventFor`

> **편집 후 지도 — 구현 리뷰 r1 처분 반영 뒤(2026-11-23).** 편집 전 원본 `../../pre-edit-ast/internal-app-engine--attestationwatch.eventfor.ast.json`,
> 재번호 `renumber.txt`(difflib — 생성기 `../../renumber.py.txt`).

- Source: `internal/app/engine/attestation_watch.go` (364-463)
- AST evidence: `ast.json` — branches **22** / returns **1** / calls **29**, source SHA-256 `292dab380cb7273b0a42307786fb72e3511e08859e38c5e8dae785cdd4932bf3`(파일 단위).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/attestation_watch.go --func attestationWatch.eventFor`.
- Risk scan: `risk-pattern-report.md`. a166 census 0건.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `v`·`key`·`w.critical` | judge·stageKey·설정 | — | `w.critical` false 면 critical 단계도 normal 타입(B1) |
| 불변식 | 거절 상태 알림은 제목과 본문 첫 문장에서 「지금 엔진을 재시작하면 기동이 거부된다」 단정(Q3), 계좌·경로·원문 오류 없음 | engine-safety delta | — |
| 불변식(리뷰 P1-2) | 범위 상실의 원인·할 일은 빠진 endpoint 종류로 — 전부 POST 일 때만 「감독하 실주문 검증 증거 없이 쓰였다」·`tossctl verify run` | Manager 처분 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@366) | critical 단계 && 알림 켜짐 | critical 타입 | — | `TestA196OverdueIsCriticalOncePerDay` · `TestA196NotificationsOffDowngradesOverdue` |
| B2 (if@371) | 만료 시각 있음 | `expires` | — | 날수·만료 시험 |
| B3 (switch@384) | 단계 문구 | — | — | — |
| B4 (case@385) | notice·imminent | — | — | — |
| B5 (if@386) | 기한 = 감독 증거 | 「감독하 실주문 검증 만료 D-N」, 본문 「D-N(<기한> KST)에 만료된다. 다른 유효한 감독 증거가 없으면 … 재시작이 거부된다. 재시작이 거부되면 그때부터 자동 손절이 돌지 않는다. 지금 실행 중인 엔진은 계속 돈다」 + `tossctl verify run` | — | `TestA196ProofDeadlineGovernsTheDayCount` · `TestA196ImminentProofNoticeTellsTheProofDeadline`(D-2·D-1) |
| B6 (else@390) | 기한 = 만료 | 만료 D-N 문구 + 갱신 할 일 | — | `TestA196WithoutSupervisedProofCountsTheExpiry` |
| B7 (case@395) | proof_overdue | 지남 문구(같은 기준) | — | `TestA196OverdueIsCriticalOncePerDay` |
| B8 (case@399) | expired | 거부 단정 | — | `TestA196RefusalTellsTheConsequenceFirst` |
| B9 (case@402) | scope_lost | 누락 나열 | — | 아래 |
| B10 (range@406) **새** | 누락 순회 | POST 개수 셈(`mutationEndpoint` — 새 함수) | — | `TestA196ScopeLossWordingFollowsTheMissingKind` |
| B11 (if@407) **새** | POST 인가 | `mutations++` | — | 같은 시험 |
| B12 (if@411) **새** | 누락 전부 POST | 「주문 실행 검증이 attestation 에 없다」·원인·`tossctl verify run` | — | 같은 시험「POST 둘」「POST 하나」 |
| B13 (else@415) **새** | GET 이 섞임 | 「엔진이 쓰는 endpoint 가 attestation 에 없다」, 원인 단정 없음, 재발급 할 일 | — | 같은 시험「GET 하나」「POST+GET」 |
| B14 (if@418) **새** | 그중 POST 있음 | `tossctl verify run` 덧붙임 | — | 같은 시험「POST+GET」 |
| B15 (case@423) | incomplete | 거부 단정 | — | `TestA196RefusalTellsTheConsequenceFirst` |
| B16 (case@426) | account_mismatch | 거부 단정 | — | 같은 시험 |
| B17 (case@430) | 그 밖(unreadable) | 거부 단정 | — | 같은 시험 · `TestA135UnreadableAlertCarriesNoAccountOrPath` |
| B18 (if@438) | 만료 시각 | `expires_at` | — | 누출 시험 |
| B19 (if@441) | 날수 단계 | `days_left` | — | `TestA196NewSupervisedProofRestartsTheCount` |
| B20 (if@444) | 누락 있음 | `missing_endpoints` | — | 누출 시험 |
| B21 (if@447) | ProofAt | `supervised_proof_at` | — | 누출 시험 |
| B22 (if@450) | 기한 있음 | `deadline_at`·`deadline_kind` | — | `TestA196ProofDeadlineGovernsTheDayCount` |
| (끝) | — | `Key: "attestation-watch:" + key`(발급 신원 없음) | return@455 | `TestA196RestartedWatchReusesTheRejectionKey` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Sprintf`·`strings.Join`·`Format`·`mutationEndpoint`(새 함수 — `POST ` 접두) | 문구 | 순수 | AST calls 29 |
| (하류) `obs.Notifier.Notify` — critical 은 Key 로 outbox 행 | 같은 Key = 같은 행 | 무편집 | 1단계 cited-ast |

## State mutations and fallbacks

- 지역 값만. `w.identity` 는 읽지 않음.
- 누락이 비어 있는 범위 상실(현재 도달 불가 — Verify 가 누락이 있을 때만 그 오류)은 「전부 POST」 갈래로 감(0 == 0).

## Safety conclusion

- Safe edit boundary: 문구·필드·Key. 타입 선택(B1) 무변경(D6).
- High-risk impact: no.
