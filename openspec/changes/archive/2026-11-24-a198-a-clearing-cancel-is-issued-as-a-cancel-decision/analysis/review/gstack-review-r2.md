# gstack review r2 (2026-11-24, 보완 — change 전체 `e1eaf720c..fa8771ed3`, 1회)

- 실행: Skill `review`(gstack, SESSION_ID 36613-1795459896-ac4b5678), worktree `~/tossos-work/a198-impl`. diff 는 `rtk proxy git diff e1eaf720c..HEAD`(전체 37,850 줄, 대부분 OpenSpec 증거) · Go 만 2,748 줄 — 리뷰어들은 Go diff 를 읽음. 자동 수리 없음(코드 판정 편집은 Manager 결정 대기), 4분류: **A 런타임 오류(blocker)** · **B 코드 편집 필요(Manager 결정 대기)** · **C 시험 보강(바로 반영)** · **D 문서·기록(바로 반영 또는 기록만)**.
- 범위 신호: `SCOPE_STATE_MACHINE=unknown` · backend · tests · docs · config, migrations·API·frontend·auth 없음.
- 패스별 실행:

| 패스 | 실행 | 결과 |
|---|---|---|
| testing | 실행 | 1(C) |
| maintainability | 실행 | 2(D) |
| security(시험 파일 요약 모드) | 실행 | NO FINDINGS |
| performance | 실행 | 1(D, 기존 잔여) |
| simplification | 실행 | NO FINDINGS |
| **state-machines**(지난번 미실행) | 실행 — `gstack-state-matrix check` → `STATE_MATRIX_CHECK: ok machines=3`; D7(DISPATCH_STARTED→FAILED_CONFIRMED)·D9(ACKED→CONFIRMED)·D10(IN_DOUBT→CONFIRMED/FAILED_CONFIRMED) 모두 `transitionTable`/`ValidateTransition` 과 코드의 from 검사가 일치, 종결 상태 이동·전이 행 누락 경로 없음 | 미시험 셀 2(C) |
| **red-team**(지난번 미실행, diff > 200 줄로 활성) | 실행(시험 파일 요약 모드) | 3(B 2 · D 1) |
| Claude 적대 패스 | 실행(시험 파일 요약 모드) | 5(B 2 · D 3), 「Recommendation: merge as-is … no path puts a stop-loss over a live order」 |
| **codex 바깥 목소리**(지난번 미실행) | 실행 — `codex exec -s read-only … </dev/null` rc 0, 원문 `gstack-r2-codex.md` | P1 1 · P2 1(둘 다 B), `VERDICT: FINDINGS` |
| **smoke**(Step 4.7, 지난번 미실행) | 실행 — 성공 경로 1 + 가장 위험한 실패 경로: `go test ./internal/app/engine -run 'A198AnInDoubtBuyCancelCloses|A198AFailedCancelClose|A198TheSettlerLeaves' -v` → 3 PASS(`analysis/verify/r3-smoke.txt`) | 통과 |
| data-migration · api-contract · design | not-applicable — 마이그레이션·API 계약·UI 변경 없음 | — |
| Gate scan | `GATE_SUMMARY: listed=9 eligible=2 inspected=2 unread=0 tagged={}` | 아래 |

## 발견과 처분

| # | 출처 | 발견 | 분류 | 처분 |
|---|---|---|---|---|
| G1 | codex P1 | 취소 실패 줄(`noteClearCancelFailure` → `logger.Warn`)이 관측 루프 안 동기 쓰기 — 로그 싱크가 막히면 뒤 포지션의 손절 판정까지 멈춤(a129 줄은 주기 끝으로 미룸) | B | **Manager 결정 대기** — D4 「같은 싱크에 직접 한 줄」(Manager 결정 2026-11-23)의 재검토 사안. 구현 리뷰 r1 의 「실패 줄 쓰기가 루프 안 동기」 잔여와 같은 축 |
| G2 | codex P2 | `obs.Logger.Warn` 이 쓰기 오류를 돌려주지 않아 ENOSPC·EIO 로 실패한 줄도 표시가 서서 복구 뒤에도 같은 실패가 다시 안 적힘 | B | **Manager 결정 대기**(로그만 — 판정 무영향). 패닉은 D4a 가 처리함 |
| G3 | Claude 적대 #1 | `settledCancelRefusal` 이 HTTP 409 를 요구하지 않음(모호 결과의 아무 status + already-* 본문이 확정 거절) | B | **Manager 결정 대기** — gstack r1 A4/S1 과 같은 건(설계 D7 은 code 판정) |
| G4 | red-team #1 | `ConfirmedCancelOf` 가 취소 intent 의 거래일 = 주문 거래일을 요구 — 전일 주문뿐 아니라 **미국 주문이 ET 자정을 건너는 KST 주간 세션**에서도 접수된 취소를 못 찾아 D8·매도 대기가 서지 않음(재취소 → 409 already-canceled → D7 확정 거절 → 종결 기록까지 보류·매 주기 취소) | B | **Manager 결정 대기** — 기존 「전일 결속」 잔여(design §6)의 범위가 넓어짐(기존 원장 함수 편집 필요). 안전 방향(손절 보류)이며 초과 매도 아님 |
| G5 | red-team #3 | `CloseInDoubtCancelOnTargetClose` 가 `BEGIN IMMEDIATE` 로 쓰기 잠금을 먼저 잡고 CTE 판독 둘을 함 — 대상이 살아 있는 IN_DOUBT 취소마다 5 초마다 잠금, 그 사이 손절 Prepare·flatten 이 busy_timeout 대기 | B | **Manager 결정 대기**(판독을 잠금 밖으로 — 한 트랜잭션 결정 Q1 과 맞물림). 영향 크기 미측정 |
| G6 | Claude 적대 #5 | D10 「닫음」 로그 줄에 종목·대상 번호가 없어 운영자가 풀린 손절과 잇기 어려움 | B | **Manager 결정 대기**(로그 내용 — 원문 오류를 빼는 기존 규칙과 같이 결정) |
| G7 | testing | `closeCancel` 의 원장 오류 갈래(ctx 아님 — nil 반환·한 번 경고·회복 뒤 닫힘) 시험 없음 | C | **반영**: `TestA198AFailedCancelCloseDoesNotStopThePlaceResolution`(트리거로 CANCEL 상태 쓰기 거절 → 스윕 2 회 오류 0·발주 해소 진입·경고 줄 1 → 트리거 제거 뒤 CONFIRMED) PASS |
| G8 | state-machines | 미시험 셀: RECORDED·DISPATCH_STARTED 엔진 취소 × 스윕 | C | **반영**: `TestA198TheSettlerLeavesFlattenAndAckedCancelsAlone` 에 두 행(대상 종결) 추가 — 무변 PASS |
| G9 | state-machines | 미시험 셀: 번호 없는 ACKED CANCEL × 기동 | C | **반영**: `TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed` 표에 `numberless cancel`(ACKED·명명 알림 1) — 번들 재생성 |
| G10 | maintainability #1 | in_doubt_settle.go 머리·`sweep` 문서가 「PLACE 만」 그대로 | D | **반영**(주석만) — 번들 재생성(본문 무변, 2 줄 밀림) |
| G11 | maintainability #2 | `ReductionIssuer`·`ExitObserverOptions.Issuer` 문서가 취소의 `execgw.Issuer`(D2)를, `clearTheSymbol` 문서가 D8 을 말하지 않음 | D | **반영**(같은 줄 수로 주석만 — clearTheSymbol 분기·호출 무변 확인, 파일 해시만 갱신) |
| G12 | performance | D7 확정 거절이 종결 기록까지 매 주기 재취소(결정·attempt·브로커 호출) | D | 기록만 — 기존 잔여(gstack r1 A1, Manager 결정: 이 change 밖) |
| G13 | Claude 적대 #2 | D7 code 의 실물 응답 미측정 | D | 기록만 — 기존 잔여(d7-d8-evidence §1) |
| G14 | Claude 적대 #3 | D8 접수된 매수 취소 → 손절이 반대 미체결로 매 주기 거절 | D | 기록만 — 기존 잔여(design §2-5) |
| G15 | Claude 적대 #4 | `clearCancelLogged` 표시 맵이 확정 때만 지워짐 | D | 기록만 — 기존(r1 A7) |
| G16 | red-team #2 | D10 은 대상이 살아 있으면 무변 — 5xx 가 브로커에 안 닿은 가장 흔한 IN_DOUBT 는 대상 종결·재시작까지 보류 | D | 기록만 — design §6 「이전과 같은 경우」(대상이 살아 있는 동안의 보류)에 이미 있음. 「r2 회귀를 닫음」 표현은 「종결 기록 뒤 보류」에 한정됨을 review.md 에 명시 |

## Gate edits

`Gate edits: 2 listed-eligible, 2 read, 1 finding — 0 kept, 0 restored, 1 open` — `internal/reconcile/a094_acked_boot_test.go` 의 두 hunk(gate_id `ed0ccf19294c`·`cfb1d2ecb2ca`, untagged)는 D9 로 계획된 뒤집기(취소 행 ACKED → CONFIRMED·알림 0, AMEND·번호 없는 PLACE 행 추가). 추천 처분 B(유지 — 근거: design §4-29·D9, FLM 번들 `internal-reconcile--testa094anackedcancelornumberlessplaceisonlynamed`, 검증됨). 사람의 답이 없어 기록은 **open**(Manager 결정 대기).

## 결론

- `Pre-Landing Review: 6 issues (0 critical, 6 informational)` — 미해결 = B 6건(G1~G6, 전부 Manager 결정 대기). A(런타임 오류 blocker) 0. C 3건·D 2건 반영, D 5건 기록.
- COMPLETED: 모든 패스 실행(state-machines·red-team·codex·smoke 포함), CONVERGED: 아니오(시험·주석 반영 1 회 — 2회차 상한으로 재리뷰 안 함).
