# gstack review r1 (2026-11-24, 1회차·마지막)

- 실행: Skill `review`(gstack), worktree `~/tossos-work/a198-impl`, base `main`(merge-base `e1eaf720c`), HEAD `381f229de`. PR type: mixed(application + docs). 범위: `internal/` 코드 diff(6 파일, 1168 줄) — openspec·docs 는 의도 참고.
- Gate scan: `GATE_SUMMARY: listed=2 eligible=0 inspected=0 unread=0 tagged={} unmatched=120` → **Gate edits: none detected (2 listed, 0 read)**.
- 전문가 배정: Dispatching 4 specialists: testing, maintainability, security, simplification. Skipped: performance·data-migration·api-contract·design(scope — 이 diff 에 성능 경로·마이그레이션·API 계약·UI 없음), state-machines(실행 안 함 — 누락 커버리지), red-team(실행 안 함 — 누락 커버리지, 적대 패스로 대신하지 않음). Codex 바깥 적대 패스: 실행 안 함(이 change 의 freeze 리뷰 4회를 이미 codex 로 받음 — 누락 커버리지로 기록). Claude 적대 패스: 실행.
- QA(Step 4.7) 탐색 프로브: 실행 안 함 — 선택 시험 묶음(analysis/verify/r2-*.txt)이 그 역할, 별도 smoke 미실행(누락 커버리지).

## 결과 원문 요지

| 출처 | 결과 |
|---|---|
| maintainability | NO FINDINGS |
| simplification | NO FINDINGS(advisory 0) |
| security(시험 파일은 요약 모드) | INFORMATIONAL c5 — `settledCancelRefusal` 이 HTTP status 를 보지 않아 모호 결과의 아무 APIError 본문(5xx·프록시 페이지 포함)이 세 code 를 실으면 확정 거절로 닫힘. 손절 영향은 cleared=false 라 제한적 |
| testing | INFORMATIONAL c7 — `ConfirmedCancelOf` 오류 갈래 시험 없음(R13 생존) · INFORMATIONAL c5 — 실패 줄 이벤트 이름 셋 중 하나만 단언 |
| 적대(Claude, 시험 파일 요약 모드) | 8 건 — 아래 표 |

## 처분

| # | 발견 | 처분 |
|---|---|---|
| A1 | 확정 거절(already-*·400·404·422) 뒤 백오프 없음 — 종결 기록이 올 때까지 5 초마다 결정·attempt·취소 POST(Critical 예산) 반복, 다른 종목 손절의 429 위험 | **Manager 결정 대기**(새 정책) — design §6 「확정 거절의 매 주기 재시도」 잔여에 심각도 보강 필요 |
| A2 | 실패 줄의 이벤트가 ACKED·UNRESOLVED·상태 없는 기록 attempt 를 `order.not_dispatched`(「안 나감」)로 적음 — submit 의 규칙과 다름 | **반영(명백한 결함)**: `noteClearCancelFailure` 에 submit 과 같은 갈래(FAILED_CONFIRMED → rejected · UNRESOLVED_IN_DOUBT → unresolved · NOT_DISPATCHED/상태 없음+미기록 → not_dispatched · 그 밖 기록 attempt → in_doubt), `recorded` 인자 추가(호출 4 자리) |
| T2 | 이벤트 이름 단언 부족 | **반영**: IN_DOUBT 줄 `order.in_doubt`, 발행 실패·해석 실패 줄 `order.not_dispatched` 단언 |
| A3 | D8 — 접수된 매수 취소를 종결 기록 전에 치움으로 봐 늦은 매수 체결이 CLOSING 중 RECONCILE 로 | 공시된 잔여(design §2-5) — 기록만 |
| A4/S1 | `settledCancelRefusal` 이 409 를 요구하지 않음 | **Manager 결정 대기** — 설계 D7 은 「status 로 걸지 않음(code 판정)」, 정본 a165 문단은 already-processing 판정에 「status 와 code 일치」 — 409 한정은 결정 변경 |
| A5 | 정정된 원주문이 `already-canceled` 로 답하면 already-modified 제외가 무력 | 실물 증거 필요 — 기록만(design §6 「실물 미측정」) |
| A6 | 재분류 결과의 Detail 이 「HTTP 409 does not prove…」 그대로, 사유 코드가 a165 엔진 확정(`broker_rejected`)과 다름 | 사유 코드는 Manager 결정(`dispatch_rejected_by_broker`) — 기록만. Detail 문구는 기록만 |
| A7 | 표시 맵이 확정 취소 때만 지워짐(작은 누수) · lookup 실패 표시가 성공 뒤에도 남아 같은 실패가 다시 안 적힘 | 기록만(관측 줄 — 판정 무영향) |
| A8 | 취소가 Guardian 경로를 거치지 않음 — `ReductionIssuer` 문서와 어긋남 | 설계 D2(Manager 결정) — 정본 「Guardian 위험 감소 경로 경유」 는 발의(청산 주문)에 대한 것으로 읽음, 기록만 |
| T1 | `ConfirmedCancelOf` 오류 갈래 시험 없음 | 주입 불가(design §6 시험 공백) — 리뷰가 제안한 `recorded_at='x'` 등 문자열 열 변조는 판독을 실패시키지 않음(문자열 열), 정수 열은 STRICT 거절 |

- Pre-Landing Review: 2 issues fixed(A2·T2), 0 unresolved defects in scope; INCOMPLETE — state-machines·red-team·codex·Step 4.7 smoke 미실행.
- Gate edits: none detected (2 listed, 0 read; patterns: RH-1,RH-3,RH-4,RH-12,RH-13,RH-14,RH-15,RH-16; unlisted idioms: none)
