# a128 Tasks

## 0. 사람 관문

- [ ] 0.1 TypeSafe API 키 조달 — 사람이 발급해 `TYPESAFE_API_KEY` env 로 공급
      (저장소·기억·로그에 키 저장 금지). 키 전까지 1.x 구현은 mock 으로 진행 가능.
- [ ] 0.2 외부 API 호출량 기본값 승인 — 세션당 샘플링 간격 10분 × 상위 10종목 × 판단 2
      (세션당 ≤ 약 780 판단 호출 × 4세션, 비용은 키 발급 시 요금표로 확인해 여기 기록).
- [ ] 0.3 (선택) WTS 뉴스 소스용 `tossctl auth extend` — 사람 실행. 안 하면 뉴스 결손
      모드로 측정(원장에 명시됨).

## 1. 구현 (Opus 팀메이트 — 새 파일만, 생산 무접촉)

- [ ] 1.1 TypeSafe API 계약 동결 — docs.typesafe.ai/api.md·primitives/noul.md 를 읽고
      요청/응답 계약을 원문 인용으로 `analysis/typesafe-api-contract.md` 에 기록.
      추측 필드 금지(계약은 영수증에서).
- [ ] 1.2 `tools/a128-jev-shadow-probe/` — judgment client(stdlib HTTP, env 키,
      타임아웃·재시도), state 조립기(공개 시장 데이터만 — 계좌 필드 차단),
      수집 루프(rankings→stocks→prices/orderbook, rate limit 백오프, 세션 경계 준수),
      원장 writer(JSONL append, 행 2종), 라벨러(h분 뒤 /prices 재조회),
      보고서 생성기(p 구간표·구간당 n<30 보류 규칙).
- [ ] 1.3 시험 — mock TypeSafe 서버·mock 시세 픽스처로: 계좌 필드 차단, 뉴스 결손 명시,
      원장 행 스키마, 라벨 결손 사유, 보고서 보류 규칙, 키 부재 시 시작 거부.
      `go vet`·`make lint` 통과. FLM: not-applicable(새 파일만, High-risk 무접촉 —
      비례 원칙).
- [ ] 1.4 팀메이트 커밋(푸시 금지) — 허용 경로: `tools/a128-jev-shadow-probe/`,
      `openspec/changes/a128-jev-shadow-judgment-probe/`. 그 밖 0.

## 2. 리뷰·착지

- [ ] 2.1 gstack 리뷰(경량 — 비례 원칙) + Manager 검증(경로 전수·트레일러·시험 재실행).
- [ ] 2.2 Manager 푸시.

## 3. 측정 (2일 — 사용자 지시 2026-10-07)

- [ ] 3.1 1일차: KR 세션 + US 세션 각 1회 프로브 실행(에이전트 가능 — read-only).
      실행 로그·원장 보존.
- [ ] 3.2 2일차: 동일. 라벨러 완주 확인(마지막 판단 + h분까지).
- [ ] 3.3 원장·state 파일 커밋(계좌 데이터 0 확인 후).

## 4. 판정·종결

- [ ] 4.1 calibration 보고서 생성 — J1 예측력 유무(평평하면 J1 폐기 기록),
      J2 p≥0.9 구간의 적중률·표본 수, 2일 한계 명기.
- [ ] 4.2 사람 결정 기록 — 90% 임계 채택/기각/보류, 후속 change(수동 보조·엔진 veto
      gate) 착수 여부. 후속 change 범위 초안을 review.md 에 남긴다.
- [ ] 4.3 gate → archive(Opus) → Story 경로 갱신.
