# Function Logic Map: `attestationWatch.eventFor`

> **편집 후 지도 — 2단계 GREEN 뒤(2026-11-23).** 편집 전 원본 `../../pre-edit-ast/internal-app-engine--attestationwatch.eventfor.ast.json`, 재번호 `renumber.txt`(옛/새 분기를 (종류, 줄 문장)으로 difflib 정렬 — 생성기 `../../renumber.py.txt`).

- Source: `internal/app/engine/attestation_watch.go` (359-439)
- AST evidence: `ast.json` — branches **17** / returns **1** / calls **26**, source SHA-256 `587249217d6328264ddd0ce53b3614496265a12c223003f3e20d17dd6bc89272`(파일 단위, 편집 후).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/attestation_watch.go --func attestationWatch.eventFor`.
- Risk scan: `risk-pattern-report.md`.
- a166 census: `attestation_watch.go`·`a135_attestation_watch_internal_test.go` 0건(1.4).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `v`·`key`·`w.critical` | judge·stageKey·설정 | — | `w.critical` false 면 critical 단계도 normal 타입(B1 무변경) |
| 불변식 | 거절 상태 알림은 제목과 본문 첫 문장에서 「지금 엔진을 재시작하면 기동이 거부된다」 단정(Q3), 계좌·경로·원문 오류 없음 | engine-safety delta | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@361) | critical 단계 && 알림 켜짐 | critical 타입 | — | `TestA196OverdueIsCriticalOncePerDay` · `TestA196NotificationsOffDowngradesOverdue` |
| B2 (if@366) | 만료 시각 있음 | `expires` | — | 날수·만료 시험 |
| B3 (switch@377) | 단계 문구 | — | — | — |
| B4 (case@378) | notice·imminent | — | — | — |
| B5 (if@379) **새** | 기한 = 감독 증거 | 제목 「감독하 실주문 검증 만료 D-N — 그 뒤 엔진 재시작이 거부될 수 있다」, 본문 기한 시각·「다른 유효한 감독 증거가 없으면」·만료 시각·`tossctl verify run` | — | `TestA196ProofDeadlineGovernsTheDayCount` |
| B6 (else@383) **새** | 기한 = 만료 | 종전 제목 + 결과·갱신 할 일 | — | `TestA196WithoutSupervisedProofCountsTheExpiry` |
| B7 (case@388) **새** | proof_overdue | 지남 문구(design D7) | — | `TestA196OverdueIsCriticalOncePerDay` |
| B8 (case@392) | expired | 제목·본문 첫 문장 = 거부 단정 | — | `TestA196RefusalTellsTheConsequenceFirst`「expired」 |
| B9 (case@395) | scope_lost | 같음 + 누락 endpoint + `tossctl verify run` | — | 같은 시험「scope_lost」·`TestA135RenewalResetsAndScopeLossIsImmediate` |
| B10 (case@399) | incomplete | 같음 | — | 같은 시험 |
| B11 (case@402) | account_mismatch | 같음 | — | 같은 시험 |
| B12 (case@406) | 그 밖(unreadable) | 같음 | — | 같은 시험 · `TestA135UnreadableAlertCarriesNoAccountOrPath` |
| B13 (if@414) | 만료 시각 | `expires_at` | — | 누출 시험 |
| B14 (if@417) | 날수 단계 | `days_left` | — | `TestA196NewSupervisedProofRestartsTheCount` |
| B15 (if@420) | 누락 있음 | `missing_endpoints` | — | 누출 시험 |
| B16 (if@423) | ProofAt | `supervised_proof_at` | — | 누출 시험 |
| B17 (if@426) **새** | 기한 있음 | `deadline_at`·`deadline_kind` | — | `TestA196ProofDeadlineGovernsTheDayCount` |
| (끝) | — | **`Key: "attestation-watch:" + key`**(종전 `+ ":" + w.identity`) | return@431 | `TestA196RestartedWatchReusesTheRejectionKey` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Sprintf`·`strings.Join`·`Format` | 문구 | 순수 | AST calls 26 |
| (하류) `obs.Notifier.Notify` — critical 은 Key 로 outbox 행 | 같은 Key = 같은 행(재발급·재시작) | 무편집 | 1단계 cited-ast |

## State mutations and fallbacks

- 지역 값만. `w.identity` 는 더 읽지 않음(Key 에서 뺌).

## Safety conclusion

- Safe edit boundary: 문구·필드·Key. 타입 선택(B1) 무변경(D6), 필드 이름 기존 4 유지 + 새 2.
- High-risk impact: no.
