# Tasks: add-candidate-discovery

각 `[T]`는 RED → GREEN → REFACTOR → VERIFY를 거친다. 이 change는 읽기 전용이므로 실계좌
mutating 단계가 없다 — 검증은 fixture와 주입 clock으로 한다.

## 1. 기반 — 관측과 영속

- [ ] 1.1 [T] `Observation`·`Candidate` 타입 + 파생값 표시. 원천이 준 값과 계산값을 구분.
- [ ] 1.2 [T] 후보 저장소(SQLite, 원장과 **별개 파일·별개 락**). ext4 강제·fuseblk 거부는
  원장 규칙을 따른다. RED: 재시작 후 이전 프로세스의 관측을 읽지 못함 → GREEN.
- [ ] 1.3 [T] 저장소가 엔진 원장 락을 잡지 않음을 테스트로 고정 — 스캔이 도는 동안
  `tossctl engine run`의 flock 획득이 막히지 않는다.
- [ ] 1.4 [T] 후보 수명(활성/냉각/만료)과 **재진입 시 `first_seen_at` 유지**.
  RED: 냉각 중 재진입이 `first_seen_at`을 갱신 → GREEN(유지). 만료 후는 새 후보.

## 2. 원천

- [ ] 2.1 [T] 공식 Open API 어댑터 — `Rankings`(`MARKET_TRADING_AMOUNT`,
  `MARKET_TRADING_VOLUME`, `TOP_GAINERS`), `MarketInvestorTrading`, 지표 캔들.
  **공식 원천만으로 후보가 산출됨**을 테스트로 고정.
- [ ] 2.2 [T] WTS 어댑터 — 인기 순위, 투자자 순매수 순위, 테마, AI 시그널, 스크리너.
  가산 원천이며 없어도 산출이 멈추지 않는다.
- [ ] 2.3 [T] 원천 강등 — 일부 실패 시 남은 원천으로 계속하고 `sources`·`completeness`·
  `degraded`를 후보와 스캔 결과에 기록. RED: WTS 전부 실패 시 후보 0건 → GREEN.
- [ ] 2.4 [T] 원천 병합 — 같은 `(market, symbol)`을 올린 여러 원천이 후보 **하나**가 되고
  각 원천이 그 하나를 지지한 것으로 기록된다.

## 3. 시간축 지표

- [ ] 3.1 [T] 주입 clock 배선. 테스트가 시간 경과를 만든다.
- [ ] 3.2 [T] 거래대금·거래량 변화율 — **경과 시간으로 정규화**. RED: 관측당 차분을 쓰면
  1분 간격과 10분 간격이 같은 값 → GREEN(1분 쪽이 크다).
- [ ] 3.3 [T] 최초 발견가 대비 확장률.
- [ ] 3.4 [T] 일중 range position(고점 근접도) — 캔들 기반, 캔들 부재 시 미계산으로 표시.

## 4. 추격 위험 veto

- [ ] 4.1 [T] `seen_late` — 최초 관측 시점에 이미 임계를 크게 초과.
- [ ] 4.2 [T] `extended` — 최초 발견가 대비 확장률 상한 초과.
- [ ] 4.3 [T] `near_high` — 일중 고점 근접도 상한 초과.
- [ ] 4.4 [T] veto가 점수와 **분리**됨을 고정. RED: 가속도 최상위가 `near_high`를 상쇄해
  통과 → GREEN(거부권 유지).
- [ ] 4.5 [T] veto 후보도 저장·보고된다. 늦게 본 비율을 셀 수 있어야 한다.

## 5. 사람이 보는 표면

- [ ] 5.1 [T] `tossctl candidate scan` — 1회 스캔. `mutating: false`. 출력에 임계별 통과 수,
  강등·백오프 표시.
- [ ] 5.2 [T] `tossctl candidate watch` — 반복 스캔. 간격 하한 강제.
- [ ] 5.3 [T] 실계좌 검증 runlock이 살아 있으면 `watch` 시작 거부.
- [ ] 5.4 [T] 429 백오프 + 백오프 사실을 스캔 결과에 기록.
- [ ] 5.5 [T] 콘솔 `/candidates` — 후보 목록, 최초 발견 시각, 지표, veto 사유, 원천·완전성.
  **읽기 전용 화면**. 확인 문자열 타이핑 같은 마찰을 넣지 않는다.

## 6. 격리와 게이트

- [ ] 6.1 [T] 의존성 격리 테스트 — `internal/candidate`의 import 그래프에 주문 실행 경로가
  없음을 고정. 후보에서 주문으로 가는 코드가 컴파일되지 않아야 한다.
- [ ] 6.2 Function Logic Map + `check_analysis.py`
- [ ] 6.3 PM registry allowlist + fixture 등록
- [ ] 6.4 `docs/ROADMAP.md` Phase 3 갱신 — T3.1을 이 change로 분리하고 착수 조건을 고친다
- [ ] 6.5 `make sdd-sync && make sdd-check && make gate CHANGE=add-candidate-discovery`
- [ ] 6.6 독립 리뷰(gstack) — 특히 격리 테스트가 실제로 컴파일 차단을 하는지

## 착수 전 사용자 확정 필요 (design.md 미결정)

1. 임계 기본값(순위 상위 N, 가속도 배수, `near_high` 상한)
2. 스캔 간격 기본값
3. KR 먼저인지 KR·US 동시인지
