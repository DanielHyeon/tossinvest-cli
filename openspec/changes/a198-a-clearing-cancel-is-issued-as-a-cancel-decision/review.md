# a198-a-clearing-cancel-is-issued-as-a-cancel-decision Review

## 계약 작성 (2026-11-23)
- 작성: Manager(Fable). 발견 경위: a197 1단계 probe(브랜치 `feat/a197-external-opposite-buy-cancel`) — 실제 게이트웨이로 길 치우기의 취소가 원장에서 거절됨을 재현. 분기 주장의 AST 는 그 조사 산출물 인용이며 change 안 사본·HEAD 재확인은 tasks 1.3.
- `openspec validate a198-a-clearing-cancel-is-issued-as-a-cancel-decision --strict`: valid (rc 0, 2026-11-23 HEAD 064f59fd9).
- proposal-freeze 리뷰: 미실시(tasks 1.7~1.8).
- Function Logic Map: 편집 대상 기존 함수의 번들은 tasks 1.6 에서 만든다.

## 1단계 계약과 증거 (2026-11-23, 팀메이트 Opus)
- base `e1eaf720c`(1.1). CodeGraph 1.6.0 hard evidence·CodeGraphContext not-applicable·증거 조정(1.2) — `analysis/code-context/`.
- 주장 대조(1.3) — `analysis/claim-check.md`, AST 36 개 `analysis/cited-ast/`, probe 재실행 `analysis/probe/`(a197 원본 바이트 동일 + (b) 확장, 버린 worktree 미커밋 파일). 결정을 바꾸지 않는 정정 1: proposal 「Why」 표의 「키만 없애서는…」 줄(결정 행에 종류 필드 없음 · 매도 주문은 키만 없으면 방향 비교 통과 · checkIdempotencyKey 는 키가 있을 때만).
- 편집 표면(1.4) `analysis/edit-surface.md`, rate limit(1.5) `analysis/rate-limit.md`, pre-edit FLM/BTM 번들 2(1.6) `analysis/function-logic/`(clearTheSymbol · Context.ExitObserver). `check_analysis.py --change …` rc 0(required 0).
- `openspec validate a198-a-clearing-cancel-is-issued-as-a-cancel-decision --strict`: valid (rc 0, 2026-11-23, 이 브랜치 정정 뒤).
- Manager 결정 필요(문서 미편집): ① D4 `warnEpisode` 는 운영 `opts.Log` nil 로 무기록 ② D2 안 — (가-B) 관측기 내부 `execgw.Issuer` 구성 권고(편집 함수 1~2, 하네스 0), 단 D3·조립 Scenario 공집합화 ③ a087 current 번들이 `Context.ExitObserver` sha 를 고정 — 착지 순서 ④ 취소 사전 GET 과 정본 「브로커 조회를 더하지 않는다」 ⑤ 접수된 매수 취소 뒤 재취소(§2-1).

## proposal-freeze 리뷰 r1 (2026-11-23)
- 보이스: codex-cli 0.165.0 `codex exec -s read-only </dev/null`(다른 계열, 적대적 Eng). 원문 `analysis/review/proposal-freeze-r1-codex.md`. 판정 「freeze 불가 — P1 4, P2 2, P0 미입증」.

| # | 발견 | 대조 | 처분 |
|---|---|---|---|
| P1-1 | D4 대로면 운영에서 원인 기록 0(opts.Log nil); 전체 Log 연결은 계좌 노출 위험 | 1.3 claim-check D4 와 같은 결론, 코드 재확인 | **Manager 결정 대기**(싱크 선택) |
| P1-2 | CANCEL 이 ACKED 에 남으면(MarkAcked dispatch.go:181 → Settle :199 사이 crash·쓰기 실패) 기동 복구(acked_boot.go:51 「CANCEL/AMEND … stays ACKED」)·세션 해소(PLACE 만)·운영자 해소(resolution.go:143 UNRESOLVED_IN_DOUBT 만) 어느 것도 못 닫고 clearTheSymbol B3 이 손절을 막음, 계수 제외(onlyEngineCancelsInFlight B5) | 코드 재확인 — 사실. 1단계 산출물에 없던 발견 | **Manager 결정 대기**(복구 계약 = 새 범위) |
| P1-3 | 전일 주문의 당일 취소: ConfirmedCancelOf 는 주문 거래일, 게이트웨이는 취소 거래일 기록 → 확정 취소 재전송·withPending=false 시 대기 방벽 누락 | claim-check (c) 말미와 같은 사실 | **Manager 결정 대기**(결속 계약 = 새 가드) |
| P1-4 | 취소마다 사전 GET(service.go:218 → precheck.go:98) — 정본 「조회 0건」과 문언 충돌, 예산·지연 계상 부족 | rate-limit.md 와 같은 사실 | **Manager 결정 대기**(정본 해석) |
| P2-5 | D3 조립 실패는 엔진 런타임 전체(무충돌 손절 포함)를 막음 — design §3 「새 거절 없음」에 빠짐 | engine.go:684 반환 확인 | **Manager 결정 대기**(D2 안과 함께) |
| P2-6 | 결정 종류는 원장에 없어 §4 「종류 CANCEL 을 원장으로 단언」 불가(AMEND 변이 못 잡음); 결정은 대상 주문번호를 결속하지 않음 | claim-check P5 와 같은 사실 | **Manager 결정 대기**(시험 전략 변경) |

- 사실 오류 정정 반영: 없음(모든 발견이 결정·범위·시험 전략 사안). 리뷰가 제안한 시험 표(원문 하단)는 미반영 — 새 가드·범위 확장 금지 규칙.
- r2: 미실시 — r1 발견이 전부 Manager 결정 대기라 반영한 수정이 없음(재리뷰할 변경 없음).
