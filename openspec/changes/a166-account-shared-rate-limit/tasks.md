> 정의만 — 착수 전(2026-10-22). 아래 task 는 골격이며 아무것도 시작하지 않았다.

## 0. 착수 준비

- [ ] 0.1 단일 writer 배정, `python3 tools/sdd/capture_change_base.py --change a166-account-shared-rate-limit` 로 base 고정.
- [ ] 0.2 2026-10-22 「엔진 + 수동 조회 동시 429」 관측의 로그 영수증 확보(로컬 기록만 — 계좌 API 재호출 금지). 없으면 「영수증 없음」으로 기록.
- [ ] 0.3 공식 API 를 치는 모든 전송 자리 전수(CodeGraph `--filter` 포함 + grep) — `doRequest` 밖 자리(`token.go:144`·`a112_mbus_read.go:153`·프로브) 처분.
- [ ] 0.4 search-first(외부 레지스트리·MCP 포함) — 들이는 것은 사람 승인 + `dependency-auditor`.
- [ ] 0.5 FLM/BTM: 삽입 자리 기존 함수와 엔진 폴링 429 처리 함수 — High-risk, 면제 없음.

## 1. 계약 (design freeze)

- [ ] 1.1 design.md — 선택지 1~4 중 하나를 고른 근거, 계좌 식별 키, 손절·flatten 우선순위 증명, 크래시·교착·시계 역행, 배포·rollback.
- [ ] 1.2 spec delta — order-execution 「Retry Matrix 산출물」 MODIFIED.
- [ ] 1.3 proposal-freeze 리뷰(적대적 Eng 보이스 포함) → `review.md`.

## 2. 구현 (RED → GREEN → REFACTOR → VERIFY)

- [ ] 2.1 Pre-Edit 선언.
- [ ] 2.2 RED: 다중 프로세스 합산 예산, 엔진 GET 429 갈래(Retry-After 있음 ≤30s / 초과 / 없음), 보호 청산 우선순위.
- [ ] 2.3 GREEN: 최소 구현 + 정본 retry matrix 와 코드 기본값을 같은 커밋에서 갱신.

## 3. 검증

- [ ] 3.1 변이 원장 — 공유 키·간격·잠금·우선순위·429 갈래, 양성 대조군 포함.
- [ ] 3.2 다중 적대 리뷰 + gstack review/security.
- [ ] 3.3 영향 범위 시험(`-race`·다중 프로세스 포함)과 미실행 범위 기록.
- [ ] 3.4 Manager 독립 검증.

## 4. 닫기

- [ ] 4.1 완료 게이트(사용자 명시 요청 시) → archive → Story 경로·PM sync.
- [ ] 4.2 검증된 교훈만 Hindsight retain.
