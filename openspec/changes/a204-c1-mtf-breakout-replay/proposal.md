# a204 — 전략 C1 을 StockOS 청산으로 오프라인 재생해 잰다 (c1-mtf-breakout-replay)

## Why

a203 은 과거 KR 약 100 세션의 공식 1분봉을 저장소 밖 캐시(`~/.local/state/tossos-a203/cache/`)에 모으고 돌파 KR(동결 v1)을 재생함.
같은 캐시로 **다른 진입 규칙 하나(C1)** 를 함께 재면 추가 브로커 조회 없이 후보를 하나 더 판정할 수 있음. C1 은 「30분 추세가 위일 때만,
3분 돌파 뒤 첫 되돌림에서 산다」는 다중 시간대 돌파이고, 청산은 이미 실거래 중인 StockOS 청산을 **그대로** 씀 — 청산을 새로 설계하면
진입과 청산 둘 다 미검증이 되므로, 청산은 검증된 쪽을 옮겨 오고 진입만 잼.

Manager 위임(2026-11-27): 「a203 의 형제 change a204 — 오프라인 재생기, StockOS 청산 원문 복사, 판정 사전등록. 생산 코드 0 · 브로커 호출 0」.

## What Changes

1. **사전등록(design §2 · §6 · §7)**: C1 진입 숫자 · 체결 가정 · 1R 정의 · 판정 규칙(PASS / FAIL / INSUFFICIENT)과 100점 E 축 배점을
   실데이터가 생기기 전에 커밋함. 이 change 안에서 실데이터 실행은 없음.
2. **재생 도구** `tools/a204-c1-mtf-breakout-replay/`(Python, 표준 라이브러리만):
   - `stockos_port/` — StockOS 순수 모듈 8개를 커밋 `fcad869e` 에서 **원문 그대로** 복사(머리 주석 1줄만 추가, import 수정 0).
   - `aee_exits.py` — StockOS `auto_exit_execution.py`(AEE) 의 청산 오케스트레이션을 줄 번호 인용과 함께 재구현.
   - `c1_entry.py` — 봉 집계(닫힌 봉만) · EMA · 거래량 z · 돌파 · 첫 되돌림 · 제외 규칙.
   - `sim_config.py` — StockOS 와 같은 이름 · 기본값 · 범위의 청산 설정 파일 검증(범위 밖 → 종료 코드 2) · digest.
   - `verdict.py` — 판정 계산. `replay.py` — CLI 와 세션 자료 어댑터(지금은 픽스처 형식만).
3. **시험**: `python3 -m unittest`(pytest 미설치) — 집계 미래 정보 차단 · 진입 경계 · STOP_FIRST · floor 단조 · 세션 조임 경계 · 설정 경계 ·
   원문 복사 sha256 · 판정 경계 · 보고서 digest.

## What Does NOT Change

- TossOS 엔진 · 주문 · 손절 · 원장 · 대사 · 콘솔 · 운영 설정 · 토글 — 무접촉. `internal/` · `cmd/` Go 편집 0, 기존 Go 함수 편집 0.
- StockOS 저장소 — 읽기만 함(복사 원본 커밋 기록).
- a203 의 사전등록 · 파일 · 캐시 — 무접촉. a204 는 a203 캐시를 **나중에 읽기만** 함.
- 브로커 · 네트워크 호출 0, 자격 증명 · `.env` · `~/.config/tossctl` 읽기 0.
- 이 설정 파일은 **시뮬레이터 설정**이며 TossOS 런타임 설정이 아님.

## Capabilities

### Modified Capabilities

- 없음 — 생산 동작 · 요구사항 변경 0. `skip_specs: true`(선례 a202 · a203).

## Impact

- 위험 등급: **Low**. 오프라인 분석 도구, High-risk 경로 · 게이트 판정 변경 0.
- 비례 원칙(docs/WORKFLOW.md): FLM — **not-applicable**(기존 Go 함수 본문 편집 0, 새 Python 파일만). 변이 원장 · 다중 적대 리뷰 —
  **not-applicable**(주문 경로 밖 오프라인 도구, High-risk · 게이트 판정 변경 0). 시험 + 경량 gstack 리뷰.
- 의존: a203 캐시의 디스크 형식이 아직 없음 → 어댑터 인터페이스 + 픽스처 형식으로 두고 열린 항목 O1 로 남김.

## 열린 항목

- **O1** a203 캐시 · 종목군 목록의 디스크 형식이 정해지면 그 형식을 읽는 어댑터를 추가(새 change 아님 — 이 change 의 tasks 4.x, 사전등록 숫자는 불변).
- **O2** 표본 세션 수가 51 미만이면 개발 절반이 비어 판정 조건 (5)를 잴 수 없음 → design §7 에서 INSUFFICIENT 로 둠(Manager 확인 항목 — 위임 원문에 없던 조건).
- **O3** 판정 실행(tasks 5.x)은 a203 본 수집 뒤. 그 전에는 archive 하지 않음.
