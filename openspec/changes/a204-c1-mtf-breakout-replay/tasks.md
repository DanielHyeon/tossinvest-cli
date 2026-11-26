# a204 tasks

> 위임 2026-11-27(Manager). base `e2921fc00`. 브로커 호출 0 · 생산 코드 0. 판정 실행(5.x)은 a203 본 수집 뒤.

## 1. 계약 (실데이터 전)

- [x] 1.1 Story `STORY-TOS-a204` → FEAT-TOS-013 등록 · tracker 재생성
- [x] 1.2 proposal · design(C1 숫자 §2, 체결 §3, 청산 §4, 1R §6, 판정 §7, 설정 §8) · tasks · `.openspec.yaml`(`skip_specs: true`) — 사전등록 커밋
- [x] 1.3 StockOS 커밋 `fcad869e` 기록, 원문 8개 복사 + `MANIFEST.json`(sha256)

## 2. 구현 (RED → GREEN)

- [x] 2.1 시험 먼저(`tests/`) — RED 원문 `analysis/verify/red.txt`
- [x] 2.2 `c1_entry.py` — 집계 · EMA · z · 돌파 · 첫 되돌림 · 제외 · ATR%
- [x] 2.3 `aee_exits.py` — AEE 재구현(design §4.2) · 두 모드(§4.3)
- [x] 2.4 `sim_config.py` + `config.default.json` — 검증 · digest
- [x] 2.5 `verdict.py` · `replay.py` — 판정 · CLI · 픽스처 어댑터
- [x] 2.6 GREEN 원문 `analysis/verify/green.txt`, 픽스처 종단 실행 영수증 `analysis/verify/fixture-run.txt`

## 3. 리뷰

- [x] 3.1 gstack `review` (경량, 최대 2회차) — 발견 · 수리 `review.md`. FLM · 변이 원장 · 다중 적대 리뷰 `not-applicable` 사유 기록

## 4. a203 연결 (a203 캐시 형식 확정 뒤 — 열린 항목 O1)

- [ ] 4.1 a203 캐시 · U-A 목록을 읽는 어댑터(인터페이스 design §5) + 실제 캐시 한 세션 대조 시험. 사전등록 숫자 불변

## 5. 판정 실행 (a203 본 수집 뒤, Manager 승인)

- [ ] 5.1 a203 캐시(읽기 전용)로 실행, 보고서 지문만 커밋 · 원본 보고서는 저장소 밖
- [ ] 5.2 결과 보고(PASS 는 Story 제안 근거로만) · Manager 독립 검증 · archive
