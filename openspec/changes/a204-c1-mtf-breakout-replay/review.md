# a204 review

## 리뷰 등급과 생략 사유 (2026-11-27)

- **Function Logic Map: not-applicable** — 기존 Go 함수 본문 편집 0(새 Python 파일만). `tools/logic-map/check_analysis.py --change a204-c1-mtf-breakout-replay` → required 0.
- **변이 원장: not-applicable** — docs/WORKFLOW.md 「비례 원칙」: 주문 경로 밖 오프라인 분석 도구, High-risk 경로 · 게이트 판정 변경 0.
  (리뷰가 지적한 「시험이 못 박지 못한 규칙」은 사본에서 하나씩 바꿔 보는 확인으로 대신함 — 아래 표.)
- **다중 적대 리뷰: not-applicable** — 같은 비례 원칙. gstack `review` 경량 1회(전문 리뷰어 1 + 적대 리뷰어 1, 같은 하네스).
- **proposal-freeze 리뷰:** 경량 — `openspec validate --strict` + 작성자 셀프 검토 + Manager 검증(WORKFLOW 「리뷰 게이트」 도구 등급).
- **바깥 모델(Codex) 리뷰: 안 돌림** — 인증은 됐으나 경량 리뷰 범위로 생략. 다른 계열 보이스는 Manager 검증이 맡음(미충족 커버리지로 기록).

## Pre-Edit 선언

- 대상: 새 파일만 — `tools/a204-c1-mtf-breakout-replay/**`, `openspec/changes/a204-…/**`, PM Story · registry · Feature · 생성 tracker.
- TossOS 생산 코드(`internal/` · `cmd/`) · 기존 함수 · 설정 · 토글 · 엔진 무접촉. StockOS 저장소 읽기만(`git show` 로 복사).
- 불변식: 브로커 · 네트워크 호출 0, 자격 증명 읽기 0(`tests/test_static.py` 가 import 허용 목록 · 금지 문자열로 묶음).
- RED: `analysis/verify/red.txt`(구현 모듈 없음 → import 오류 5). GREEN: `analysis/verify/green.txt`.
- rollback: 커밋 revert.

## gstack review 1회차 (2026-11-27)

보이스: 시험 + 유지보수 전문 리뷰어 1, 적대 리뷰어 1(「Think like an attacker and a chaos engineer」). 범위: 도구 자체 코드 + design.md. `stockos_port/` 는 원문 복사라 내용 리뷰 제외(sha256 시험이 원본 일치만 확인).
Gate scan: `listed=9 eligible=1` — 유일한 표시는 `stockos_port/exit_strategy.py` 새 파일(복사본, 완화된 게이트 아님) → keep.

| # | 출처 | 지적 | 처리 |
|---|---|---|---|
| A1 | 적대 | sample 날짜 중복이 hold-out 을 줄이고 거래를 두 번 셈(INSUFFICIENT → PASS 가능) | 수리: 어댑터가 중복 거절 + `verdict.evaluate` 가 중복 제거. 시험 `test_duplicate_sample_dates_do_not_shrink_holdout` · `test_rejects_bad_manifest_values` |
| A2 | 적대 | 직전 세션 봉이 없으면 「거시 거절」로 숨어 결손률에 안 잡힘 | 수리: d−1 1분봉 없음 = 결손(design §5). a203 수집 범위 요구를 O1 계약에 적음. 시험 `test_missing_previous_session_is_a_gap_not_macro_reject` |
| A3 | 적대 | 종목군 검증 없음(세션 누락 · 중복 종목 · 문자열) | 수리: 표본 세션마다 중복 없는 종목 코드 목록 요구, 코드 형식 `[0-9A-Za-z]{1,12}`(경로 이탈 차단) |
| A4 | 적대 | NaN/Infinity → 틱 걷기 무한 루프, 분 중복이 거래량 부풀림, OHLC 불일치 | 수리: NaN · Infinity 거절, 양수 · OHLC 일관 · 분 오름차순 · 6칸 행 요구. 시험 `test_rejects_bad_bar_rows` |
| A5 | 적대 | 15:20~15:29 동시호가 봉을 연속매매로 체결 | 수리: 그 구간 봉 버림(15:30 종가 봉만). 시험 `test_auction_minutes_are_dropped` |
| A6 · A8 · A10 · A14 | 적대 | 진입 · 청산 즉시 체결 낙관, EOD 강제 청산, 특수일, min-gap 고정 | 한계로 적음(design §10 · 보고서 `limitations`). EOD 건수를 지표에 추가 |
| A7 | 적대 | R% 가 작은 거래가 평균 R 을 지배 | 판정 규칙은 사전등록대로 R 유지, 거래당 평균 net% 를 나란히 보고(Manager 확인 항목) |
| A9 | 적대 | 가격대 경계 「한 틱 아래」가 윗 가격대 틱을 씀(20,000 → 19,950) | 수리: `krTickMinor(가격 − 1)` — 매도 bid · 구조 손절 둘 다. 시험 `test_sell_tick_below_band_boundary` · `test_stop_one_tick_below_band_boundary` |
| A11 | 적대 | 빈 3분 칸이 창을 늘림 | 설계에 「있는 봉을 셈」으로 명시 |
| A12 | 적대 | 표본 지문이 자료 바이트를 안 묶음 | 수리: `input_fingerprint`(읽은 모든 파일 sha256). 시험 `test_input_fingerprint_binds_bar_bytes` |
| A13 | 적대 · 전문 | 형식 오류 · 출력 디렉터리 없음이 rc 1 traceback, NaN JSON 출력 | 수리: 경계에서 InputError, `OSError` 도 종료 코드 2, `allow_nan=False`. 시험 `test_missing_output_directory_is_exit_2` |
| S1~S6 | 전문 | 시험이 못 박지 못한 규칙(거시 마지막 15 · 엄격 >, 재무장, 띠 경계, 손절 폴백, ATR 이전 종가, 부분익절 표시 · 사다리 · floor 올림 · 확인 초, c5 음수 · 55% 경계) | 수리: 시험 추가 + 띠 비교를 정수 곱셈으로(부동소수 경계 제거). 아래 표로 확인 |
| S7 | 전문 | digest 가 1 과 1.0 을 다르게 셈 | 수리: 실수 필드는 float 로 정규화 |
| S9 | 전문 | 워밍업 전 세션 상한 2 | A2 수리로 d−1 결손 처리, 상한 2 는 15 봉에 충분(26 봉) — 유지 |

**수리 확인(사본에서 한 곳씩 바꿔 시험이 실패하는지)** — 모두 FAILED(시험이 잡음):
띠 상한 `<=`→`<`, 띠 하한 `>=`→`>`, 시한 뒤 재무장 위치, 손절 폴백 0.97→0.98, ATR TR 을 고가−저가로, 거시 `>`→`>=`, 거시 전체 봉, 부분익절 표시 삭제,
사다리 상태 보존 삭제, floor 올림 삭제, floor 이탈 확인 삭제, c5 음수 갈래 삭제, 55%→50%, sample 중복 제거 삭제, 한 틱 아래 → 같은 가격대 틱.
1회차 앞(구현 직후)에도 같은 방법으로 봉 안 순서(고가 먼저) · 미래 정보(`end`→`start`) · ratchet 하향 · TP2 해제 · 첫 되돌림 · 세션 경계 14:51 · 표본 ≤ 50 을 확인함(모두 FAILED).

2회차: 별도 리뷰어 재실행 없음(경량 상한) — 수리는 위 시험 · 사본 확인으로 검증. 수렴(무편집 회차)은 기록하지 않음.

## 검증 기록

- 시험: `cd tools/a204-c1-mtf-breakout-replay && python3 -m unittest discover -s tests -t . -v` → 95 OK(`analysis/verify/green.txt`). StockOS 저장소 대조 시험도 실행됨(건너뜀 0).
- 종단 실행: `analysis/verify/fixture-run.txt` — 합성 자료(실데이터 아님) · rc 0 / 덮어쓰기 rc 2 / 설정 오류 rc 2 · 102 세션 × 30 종목 합성 21.3 초.
- `openspec validate a204-c1-mtf-breakout-replay --strict` 통과. PM `generate_master_tracker.py --check` 통과.
- 실행 안 함: `make gate` · `make test` · `make sdd-check`(전체 게이트는 명시 요청 시에만 — Go 변경 0), Codex 바깥 리뷰.
